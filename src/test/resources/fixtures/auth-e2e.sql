-- ============================================================================
-- auth-e2e.sql — authentication/authorization parity fixture
--
-- Purpose: the dev DB ships exactly ONE login_user (`admin`, is_admin='Y'),
-- which bypasses every module/role check
-- (ModuleAuthenticationInterceptor.hasPermission: `... || isUserAdmin(...)`).
-- Testing authorization with admin alone proves nothing, and the account-state
-- branches of the login contract (locked / disabled / expired-credentials /
-- no-OE-user) are unreachable without dedicated rows.
--
-- Seeds a fixed, reserved id range (9900-9999) so the fixture is idempotent and
-- cannot collide with application-generated ids (system_user_seq is at 110,
-- login_user_seq at 2 — this fixture deliberately does NOT advance either).
--
-- Every user shares the SAME password so the specs need exactly one secret; the
-- plaintext lives in migration/openelis-api-e2e/fixtures/env.ts (OE_E2E_PASS),
-- never inline in a spec. The hashes below are real bcrypt cost-12 `$2a$`
-- digests, matching Java's plain BCryptPasswordEncoder (SecurityConfig
-- .passwordEncoder()) — NOT a DelegatingPasswordEncoder, so there is no {id}
-- prefix and no plaintext fallback.
--
-- Consumed by: migration/openelis-api-e2e/tests/readonly/p0-auth.spec.ts
-- ============================================================================

SET search_path TO clinlims;

-- Idempotent teardown: role grants first (FK to both system_user and
-- system_role), then the two user tables.
DELETE FROM system_user_role WHERE system_user_id BETWEEN 9900 AND 9999;
DELETE FROM system_user      WHERE id             BETWEEN 9900 AND 9999;
DELETE FROM login_user       WHERE id             BETWEEN 9900 AND 9999;

-- ---------------------------------------------------------------------------
-- login_user
--
-- password_expired_dt drives LoginUserDAOImpl.getPasswordExpiredDayNo():
--     floor(current_date - password_expired_dt) * -1
-- i.e. days REMAINING. CustomUserDetailsService treats `<= 0` as
-- credentials-expired. The far-future date also keeps the value above
-- login.user.expired.reminder.day (15), so a successful login does NOT insert
-- a "password expiring" notification row as a side effect
-- (CustomFormAuthenticationSuccessHandler.passwordExpiringSoon).
-- ---------------------------------------------------------------------------
INSERT INTO login_user (id, login_name, password, password_expired_dt,
                        account_locked, account_disabled, is_admin, user_time_out)
VALUES
  -- Non-admin with exactly one narrow role (Reception): the positive case for
  -- role-gated endpoints, and the subject that proves admin-bypass is not what
  -- is being tested.
  (9901, 'e2e_reception', '$2a$12$fXxEjo/QbHU7NVgpjrwfLOJFtBtJoF7tZ3xsv580buiMg5OaDkjpG',
   DATE '2031-07-10', 'N', 'N', 'N', '20'),
  -- Authenticates fine, holds ZERO roles: the denial case for role gates, and
  -- the case that proves /session reports an empty roles[] rather than failing.
  (9902, 'e2e_noroles',   '$2a$12$fXxEjo/QbHU7NVgpjrwfLOJFtBtJoF7tZ3xsv580buiMg5OaDkjpG',
   DATE '2031-07-10', 'N', 'N', 'N', '20'),
  -- account_locked='Y' -> LockedException -> error.lockedCredentials.
  (9903, 'e2e_locked',    '$2a$12$fXxEjo/QbHU7NVgpjrwfLOJFtBtJoF7tZ3xsv580buiMg5OaDkjpG',
   DATE '2031-07-10', 'Y', 'N', 'N', '20'),
  -- account_disabled='Y' -> DisabledException -> error.disabledCredentials.
  (9904, 'e2e_disabled',  '$2a$12$fXxEjo/QbHU7NVgpjrwfLOJFtBtJoF7tZ3xsv580buiMg5OaDkjpG',
   DATE '2031-07-10', 'N', 'Y', 'N', '20'),
  -- password_expired_dt in the past -> passwordExpiredDayNo <= 0 ->
  -- CredentialsExpiredException -> error.expiredCredentials. Note this is a
  -- POST-authentication check in Spring, so it only surfaces when the password
  -- is CORRECT (see the ordering assertions in the spec).
  (9905, 'e2e_expired',   '$2a$12$fXxEjo/QbHU7NVgpjrwfLOJFtBtJoF7tZ3xsv580buiMg5OaDkjpG',
   DATE '2020-01-01', 'N', 'N', 'N', '20'),
  -- Credentials are valid, but no ACTIVE system_user carries this login_name,
  -- so LoginUserDAOImpl.getSystemUserId() returns 0 and session setup blows up
  -- inside the success handler. Pins the fact that this failure escapes the
  -- apiCall=true JSON contract entirely.
  (9906, 'e2e_noouser',   '$2a$12$fXxEjo/QbHU7NVgpjrwfLOJFtBtJoF7tZ3xsv580buiMg5OaDkjpG',
   DATE '2031-07-10', 'N', 'N', 'N', '20'),
  -- Distinct user_time_out (999 min, the varchar(3) maximum) so the per-user
  -- session TTL is observable and cannot be confused with the 20-minute default
  -- (CustomFormAuthenticationSuccessHandler.DEFAULT_SESSION_TIMEOUT_IN_MINUTES).
  (9907, 'e2e_longtimeout', '$2a$12$fXxEjo/QbHU7NVgpjrwfLOJFtBtJoF7tZ3xsv580buiMg5OaDkjpG',
   DATE '2031-07-10', 'N', 'N', 'N', '999');

-- ---------------------------------------------------------------------------
-- system_user — joined to login_user by login_name STRING (not a FK), and only
-- when is_active='Y' (LoginUserDAOImpl.getSystemUserId). e2e_noouser's row is
-- deliberately is_active='N' so that join finds nothing.
-- ---------------------------------------------------------------------------
INSERT INTO system_user (id, login_name, first_name, last_name, is_active, is_employee)
VALUES
  (9901, 'e2e_reception',   'E2E', 'Reception', 'Y', 'Y'),
  (9902, 'e2e_noroles',     'E2E', 'NoRoles',   'Y', 'Y'),
  (9903, 'e2e_locked',      'E2E', 'Locked',    'Y', 'Y'),
  (9904, 'e2e_disabled',    'E2E', 'Disabled',  'Y', 'Y'),
  (9905, 'e2e_expired',     'E2E', 'Expired',   'Y', 'Y'),
  (9906, 'e2e_noouser',     'E2E', 'NoOeUser',  'N', 'Y'),
  (9907, 'e2e_longtimeout', 'E2E', 'LongTimeout', 'Y', 'Y');

-- ---------------------------------------------------------------------------
-- system_user_role — role 4 is 'Reception' (system_role.name is character(30),
-- i.e. blank-padded on read; every comparison site in Java trims it).
-- e2e_noroles intentionally gets nothing.
-- ---------------------------------------------------------------------------
INSERT INTO system_user_role (system_user_id, role_id) VALUES
  (9901, 4),
  (9903, 4),
  (9904, 4),
  (9905, 4),
  (9907, 4);

-- Verification (visible in the loader output).
SELECT lu.login_name, lu.account_locked, lu.account_disabled, lu.is_admin,
       lu.user_time_out,
       floor(current_date - lu.password_expired_dt) * -1 AS password_expired_day_no,
       COALESCE(su.is_active, '-')                       AS system_user_active,
       COALESCE(string_agg(trim(sr.name), ',' ORDER BY sr.id), '(none)') AS roles
  FROM login_user lu
  LEFT JOIN system_user su
         ON su.login_name = lu.login_name AND su.is_active = 'Y'
  LEFT JOIN system_user_role sur ON sur.system_user_id = su.id
  LEFT JOIN system_role sr       ON sr.id = sur.role_id
 WHERE lu.id BETWEEN 9900 AND 9999
 GROUP BY lu.id, lu.login_name, lu.account_locked, lu.account_disabled,
          lu.is_admin, lu.user_time_out, lu.password_expired_dt, su.is_active
 ORDER BY lu.id;
