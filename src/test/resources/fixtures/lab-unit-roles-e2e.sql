-- =============================================================================
-- A user restricted to ONE lab unit — E2E fixture
-- =============================================================================
-- Java filters two things by the caller's lab units, and the dev dataset made
-- neither observable:
--
--   AccessionValidationRestController:173
--     userService.getUserTestSections(getSysUserId(request), ROLE_VALIDATION)
--   AccessionValidationRestController:208
--     userService.filterAnalysisResultsByLabUnitRoles(getSysUserId(request), ...)
--
-- The only user with lab units in the stock data is `admin`, and its single
-- lab_unit_role_map row says `AllLabUnits` — the branch that returns EVERY
-- active test section and filters nothing. So a port that skipped both calls
-- matched byte-for-byte, and the filtering was untested.
--
-- That is not a cosmetic gap: without it a user authorised for one lab unit can
-- read validation-pending results from every other one.
--
-- ---- WHAT THIS SEEDS -------------------------------------------------------
-- e2e_labunit — a non-admin holding the Validation role on exactly ONE test
-- section, chosen at load time as a section that actually HAS a
-- validation-pending analysis so the filtered list is non-empty (a filter that
-- returns nothing proves nothing).
--
-- The four tables involved:
--   user_lab_unit_roles (system_user_id)              the parent row
--   lab_unit_role_map   (id, lab_unit)                one row per lab unit
--   lab_roles           (lab_unit_role_map_id, role)  the roles ON that unit
--   lab_unit_roles      (system_user_id, map_id)      the join
--
-- `lab_unit` holds a test_section ID as text, except for the literal
-- 'AllLabUnits' sentinel.
--
-- ---- ID POLICY -------------------------------------------------------------
-- 9850-9859. auth-e2e.sql owns 9900-9999 and DELETEs that whole range on load,
-- so this file stays clear of it and can load in either order.
--
-- The password hash is the same one auth-e2e.sql uses, so the plaintext is the
-- suite's shared E2E_PASS.
--
-- IDEMPOTENT: safe to re-run.
-- =============================================================================

DO $BODY$
DECLARE
    validation_role NUMERIC;
    target_section  TEXT;
    map_id          NUMERIC;
BEGIN
    -- ---- cleanup, children first -------------------------------------------
    DELETE FROM clinlims.lab_roles
     WHERE lab_unit_role_map_id IN (
        SELECT lab_unit_role_map_id FROM clinlims.lab_unit_roles WHERE system_user_id = 9850);
    DELETE FROM clinlims.lab_unit_role_map
     WHERE lab_unit_role_map_id IN (
        SELECT lab_unit_role_map_id FROM clinlims.lab_unit_roles WHERE system_user_id = 9850);
    DELETE FROM clinlims.lab_unit_roles     WHERE system_user_id = 9850;
    DELETE FROM clinlims.user_lab_unit_roles WHERE system_user_id = 9850;
    DELETE FROM clinlims.system_user_role   WHERE system_user_id = 9850;
    DELETE FROM clinlims.system_user        WHERE id = 9850;
    DELETE FROM clinlims.login_user         WHERE id = 9850;

    SELECT id INTO validation_role FROM clinlims.system_role
     WHERE trim(name) = 'Validation' LIMIT 1;

    -- A section that HAS something awaiting validation. Restricting the user to
    -- a section with nothing in it would make the filtered list empty for the
    -- wrong reason, and the test could not tell a working filter from a broken
    -- query.
    SELECT a.test_sect_id::text INTO target_section
      FROM clinlims.analysis a
      JOIN clinlims.status_of_sample st ON st.id = a.status_id
     WHERE st.status_type = 'ANALYSIS' AND st.name = 'Technical Acceptance'
       AND a.test_sect_id IS NOT NULL
     ORDER BY a.test_sect_id LIMIT 1;

    IF validation_role IS NULL OR target_section IS NULL THEN
        RAISE NOTICE 'lab-unit-roles-e2e: prerequisites missing; nothing seeded.';
        RETURN;
    END IF;

    -- ---- the user -----------------------------------------------------------
    INSERT INTO clinlims.login_user
        (id, login_name, password, password_expired_dt,
         account_locked, account_disabled, is_admin, user_time_out)
    VALUES
        (9850, 'e2e_labunit', '$2a$12$fXxEjo/QbHU7NVgpjrwfLOJFtBtJoF7tZ3xsv580buiMg5OaDkjpG',
         DATE '2031-07-10', 'N', 'N', 'N', '20');

    INSERT INTO clinlims.system_user
        (id, login_name, first_name, last_name, is_active, is_employee)
    VALUES (9850, 'e2e_labunit', 'E2E', 'LabUnit', 'Y', 'Y');

    -- The global role grant is what lets the user REACH the endpoint; the lab
    -- unit rows below are what bound WHAT THEY SEE. Both are needed, and they
    -- are separate mechanisms — a port that implements only the first passes
    -- every authorization test and still over-discloses.
    INSERT INTO clinlims.system_user_role (system_user_id, role_id)
    VALUES (9850, validation_role);

    -- ---- the lab unit -------------------------------------------------------
    INSERT INTO clinlims.user_lab_unit_roles (system_user_id, last_updated)
    VALUES (9850, now());

    map_id := (SELECT COALESCE(max(lab_unit_role_map_id), 0) + 1 FROM clinlims.lab_unit_role_map);
    INSERT INTO clinlims.lab_unit_role_map (lab_unit_role_map_id, lab_unit)
    VALUES (map_id, target_section);

    -- getUserTestSections keeps a lab unit only when its role set CONTAINS the
    -- requested role id, so the role has to be attached here as well as
    -- globally above.
    INSERT INTO clinlims.lab_roles (lab_unit_role_map_id, role)
    VALUES (map_id, validation_role::text);

    INSERT INTO clinlims.lab_unit_roles (system_user_id, lab_unit_role_map_id)
    VALUES (9850, map_id);

    RAISE NOTICE 'lab-unit-roles-e2e: seeded e2e_labunit restricted to test section % (role %)',
        target_section, validation_role;
END $BODY$;
