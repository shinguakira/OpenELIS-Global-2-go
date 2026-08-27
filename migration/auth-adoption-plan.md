# Authentication & Authorization — Go Migration Plan

Status: **All three phases implemented and live-verified on both stacks.**
Phase 1 (authentication), Phase 2 (authorization — module system + the
`hasRole('ADMIN')` gates), Phase 3 (CSRF). Branch `migration/p0-auth`, forked
from `migration-base` (resolves § 9.5).
Companion to [OpenELIS-Go-Migration-Plan.md](OpenELIS-Go-Migration-Plan.md),
[tech-stack-diff.md](tech-stack-diff.md), [branch-naming.md](branch-naming.md).

Every factual claim in this document was verified against the Java source **and**
the live running stack (authenticated curl against
`https://<host>/api/OpenELIS-Global`, plus direct `psql` against the dev
database). Where something could not be verified, it says so explicitly — and
where building the port proved a claim wrong, § 0.2 records the correction
rather than quietly editing the original.

---

## 0.1 What shipped, and how it was verified

**Oracles:** `openelis-api-e2e/tests/readonly/p0-auth.spec.ts` (authentication,
34 tests) and `p0-authz.spec.ts` (authorization, 19 tests). Each runs against
**both** targets from one file (`api-readonly` → Java, `go-parity` → Go).
Nothing in either is target-specific except one explicitly skipped test for an
endpoint that is not ported yet.

| Run | Result |
| --- | --- |
| Java (`api-readonly`, p0-auth) | **34 passed** |
| Java (`api-readonly`, p0-authz) | **19 passed** |
| Go (`go-parity`, p0-auth) | **33 passed, 1 skipped** (`rest/open-configuration-properties` is not ported — deferred to the config branch) |
| Go (`go-parity`, p0-authz) | **19 passed** |
| Go (`go-parity`, ALL ported units a1/a2/b1/b2 + p0) | **76 passed, 2 skipped** |
| Java (`api-readonly`, full suite incl. the 500-endpoint auth sweep) | **532 passed** |
| **Inversion** — p0-auth against the pre-auth binary (`migration-base`) | **27 of 33 failed** |
| **Inversion** — p0-authz against the Phase-1-only binary (`eebef8418`) | **7 of 19 failed** |

The inversion runs are the ones that matter (Constitution V.6).

For p0-auth: 27 failed, 5 passed, 1 skipped. The 5 that passed are exactly the
ones that must pass on both sides by construction — `rest/supportedlocales/active`
(legitimately anonymous) and the four *authenticated*-half boundary assertions,
whose anonymous halves all failed.

For p0-authz: **every test that asserts a DENIAL failed**, and only those. The 12
that passed are two pure-DB premise checks plus ten allow-assertions — which
cannot be inversion-sensitive, since a service with no authorization allows
everything. One of them (`is_admin='Y' bypasses the module check`) passes on the
Phase-1 binary for the *wrong reason*: there was no module check to bypass. It is
meaningful only paired with the 401 test next to it, which does fail. Stated
here rather than counted as coverage.

**Fixture:** `src/test/resources/fixtures/auth-e2e.sql`, wired into
`load-test-fixtures.sh` for both profiles. Reserved ids 9900–9999, real bcrypt
cost-12 `$2a$` hashes, one shared password held in `fixtures/env.ts`. It seeds
the cast § 7 asked for — non-admin-with-one-role, no-roles, locked, disabled,
expired, no-active-system-user, long-timeout — because the dev DB ships only
`admin`, whose `is_admin='Y'` bypasses every module check.

Phase 2 needed three more, and each closes a specific hole where a WRONG port
would still have passed every other test:

| user | `is_admin` | roles | closes |
| --- | --- | --- | --- |
| `e2e_testmgmt` | N | Test Management | separates the module check from the ADMIN gate — it HAS the `TestCatalog` module, so it gets past the interceptor and is refused by `@PreAuthorize` (500, not 401) |
| `e2e_globaladmin` | N | Global Administrator | a port implementing the ADMIN gate as `is_admin='Y'` alone |
| `e2e_isadmin` | Y | (none) | a port implementing it as "holds the Global Administrator role"; also the only user that proves the MODULE bypass is `is_admin` — with an empty module set, a mapped path would otherwise be 401 |

**Go:** `internal/auth/{valueholder,daoimpl,service,session,csrf,middleware,form,controller/rest}`,
laid out per § 3 and the constitution's five layers. Default-deny is enforced in
`common/web.Register` (§ 3.2) via an injected `web.Protector`; `RegisterOpen` is
the opt-out and every call site cites the Java rule that justifies it. The
unconfigured default **refuses** rather than serving — fail closed.

Decisions resolved:

- **§ 9.1 session store** — in-memory, behind the `session.Store` interface, so
  a shared backend is a drop-in later. Go stays loopback-bound (§ 8 option (a)).
- **§ 9.2 module system** — **PORTED.** The premise for deferring it was false:
  § 9.2 assumed "none of a1–c3's endpoints are mapped in `system_module_url`",
  but `/TestCatalog` **is** mapped (see § 0.2.9). Wired into
  `middleware.Guard`, i.e. applied to every protected route the way Java
  registers its interceptor on `/**` — so a future ported endpoint that has a
  mapping is checked without anyone remembering to. Also shipped: trimmed role
  loading and `middleware.RequireRole` (bodiless 403, for the c1 Reception gate)
  and `middleware.RequireAdmin` (§ 0.2.10).
- **§ 9.3 `/rest` stripping** — resolved as § 9.3 recommended: `TrimPrefix`, with
  the divergence from Java's `split("/rest")[1]` documented at the call site.
  Unreachable in practice — no ported or mapped `url_path` is exactly `/rest`
  or contains `/rest` twice.
- **§ 9.4 unported chains** — still unported, still out of scope.

Not addressed here, and still open: § 8.1 (session sharing during strangler
coexistence) is unchanged — Java and Go still do not share sessions.

## 0.2 Corrections to this plan, found by live verification

Everything below was discovered by probing the running Java stack while building
the oracle. The original sections are left intact; read these as amendments.

1. **§ 2.6 is wrong that CSRF is off the critical path.** It reasons "every
   c1/c2/c3 endpoint is a GET, so CSRF is not on the critical path for the
   current waves — but it is mandatory before any write wave (e1/f1)."
   `POST /Logout` is a write, it is part of the session lifecycle P0 itself
   ships, and Spring's `CsrfFilter` sits in front of the `LogoutFilter`.
   **Verified:** logout without a valid token silently NO-OPS — 302 →
   `<context>/Home?access=denied`, session intact. CSRF had to ship in Phase 1,
   not Phase 3.

2. **After logout the client keeps a dead `JSESSIONID`, and presenting it
   redirects — even on `/session`.** Spring's `LogoutFilter` is not configured
   with `deleteCookies("JSESSIONID")`, so the cookie survives; then
   `sessionManagement().invalidSessionUrl("/LoginPage")` intercepts *before* the
   `permitAll` rule. **Verified:** `GET /session` with a logged-out id returns
   302 → `<context>/LoginPage`, not `{"authenticated":false}`. § 2.5's
   description of `/session` as always answering holds only for a client that
   presents no id or a live one. A port that clears the cookie on logout looks
   more correct and is wrong.

3. **§ 2.3 lists the account-state → key mappings but not their PRECEDENCE**,
   which is observable. `AbstractUserDetailsAuthenticationProvider` runs
   pre-checks (locked, then disabled) → the bcrypt comparison → post-checks
   (credentials-expired). **Verified live:** a locked account with a *wrong*
   password still returns `error.lockedCredentials`, while an expired-password
   account with a *wrong* password returns `error.invalidcredentials`. Both are
   pinned by name in the spec; reordering them fails a test.

4. **Credentials-expired is not a date comparison.** § 2.3 says
   "`password_expired_dt` in the past". The real rule is
   `LoginUserDAOImpl.getPasswordExpiredDayNo() <= 0`, where that value comes from
   `SELECT floor(current_date - password_expired_dt) * -1` — **the database's**
   `current_date`. Equivalent in effect, but computing it from the Go process's
   clock would diverge whenever app server and Postgres disagree about the day
   (the same failure mode b2 hit on site codes). The port keeps it in SQL.

5. **The no-OE-user branch escapes the `apiCall=true` JSON contract entirely.**
   § 2.3 notes login "fails at session setup with `login.error.noOeUser`" without
   saying what a client sees. **Verified:** 302 → `<context>/LoginPage`, no JSON,
   even with `?apiCall=true` — because the credential check already SUCCEEDED and
   the failure happens inside the success handler.

6. **The denial contract is best asserted unfollowed.**
   [openelis-api-e2e.md](openelis-api-e2e.md) § 0 records the baseline as "HTTP
   200 with the login HTML page"; § 2.8 here records 302 → `/LoginPage`. Both are
   right — the first is the *followed* redirect. The oracle asserts the raw
   `302` + `Location`, because "the body looks like Tomcat's login JSP" is not a
   contract a Go port can meet, while "302 to `/LoginPage`, carrying no data" is.

7. **The § 6.1 padding trap hides from the test harness too, not just from
   `psql`.** § 6.1 explains that `'[' || name || ']'` trims and `SELECT name`
   does not — true, but `psql -tA`'s own output and the suite's `query()` helper
   BOTH trim the trailing whitespace, so a spec that asserts on raw trailing
   spaces silently proves nothing. The oracle asserts
   `octet_length(name) = 30` against `length(trim(name))` instead.

8. **Exact authed `/session` key set**, for a user with no lab-unit roles
   (verified live, and now pinned):
   `authenticated, loginMethod, sessionId, userId, loginName, firstName,
   lastName, roles, csrf`. `userLabRolesMap` and `loginLabUnit` are dropped by
   `Include.NON_NULL`; the lab-unit-roles subsystem is not ported. A user with
   zero grants gets `"roles":[]` — an empty collection, not a dropped key.

9. **§ 2.7's central claim is FALSE for the already-ported surface, and § 9.2's
   own warning had already materialised.** § 2.7 concludes "for c1, the module
   check is a no-op: authentication alone is the gate", and § 9.2 offers to defer
   the module system because "none of a1–c3's endpoints are mapped in
   `system_module_url`" — while noting that deferring "means Go is *more
   permissive* than Java for any endpoint someone later adds a mapping for, and
   nothing would flag that."

   `/TestCatalog` **is** one of the 382 rows in `system_module_url`, mapped to
   the `TestCatalog` module (held only by the Global Administrator and Test
   Management roles). It has been ported since b1. **Verified live:** a non-admin
   authenticated user gets `401 { "status": 401, "message": "Not Authorized" }`
   from Java and, before this work, the full catalog from Go. The queries in
   § 2.7 only covered `%patient%` paths, which is why it was missed.

   Re-checked exhaustively this time: of the 21 ported a/b routes, `/TestCatalog`
   is the only one with a mapping. All others fall to the auto-allow rule.

10. **Three ported b1 controllers carry a class-level
    `@PreAuthorize("hasRole('ADMIN')")`, and § 2.7(b) lists none of them.**
    § 2.7(b) enumerates the programmatic role checks as `PatientMergeRestController`
    (c1), `SampleEdit` and `StorageLocation` — all future waves. It misses the
    three that are **already ported**:

    | Java controller | ported route(s) |
    | --- | --- |
    | `DictionaryMenuRestController` | `rest/dictionary-categories` |
    | `TestCatalogRestController` | `rest/TestCatalog` |
    | `TestCatalogEditorRestController` | `rest/test-catalog/{lab-units,panels,sample-types}` |

    `rest/dictionary-categories` in particular looks like ordinary reference
    data and is admin-only in Java. Five ported endpoints were open to any
    authenticated user in Go.

11. **A `@PreAuthorize` denial is HTTP 500, not 403 — § 2.8's table has no row
    for it.** Verified live and deterministic, body byte-identical apart from
    the timestamp:

    ```
    HTTP/1.1 500   Content-Type: application/json;charset=UTF-8
    {"timestamp":1787803536083,"status":500,"error":"Internal Server Error"}
    ```

    The `AccessDeniedException` raised by Spring's method security never reaches
    `SecurityConfig`'s `accessDeniedHandler` (which is wired for the filter
    chain), so it surfaces as an unhandled error. This is a genuine Java bug.
    Reproduced rather than corrected, on the precedent § 2.8 sets explicitly for
    the 401-vs-403 split ("Reproduce it; do not correct it") — a divergence here
    would also change how a client that branches on 4xx-vs-5xx behaves. **Worth
    raising with the maintainers separately**, alongside § 10's other findings.

12. **"Admin" means two different things, and the difference is observable.**
    - The MODULE check's bypass is `login_user.is_admin='Y'` **alone**
      (`UserModuleServiceImpl.isUserAdmin`). The Global Administrator role does
      not bypass it.
    - Spring's `hasRole('ADMIN')` is granted by `is_admin='Y'` **OR** the Global
      Administrator role (`CustomUserDetailsService.addAuthoritiesForRole`).

    The stock `admin` account has both, so it cannot distinguish them — which is
    why the fixture seeds `e2e_globaladmin` (role, `is_admin='N'`) and
    `e2e_isadmin` (`is_admin='Y'`, zero roles). Three outcomes on ONE endpoint,
    all verified on both stacks:

    | user | has TestCatalog module | hasRole('ADMIN') | `rest/TestCatalog` |
    | --- | --- | --- | --- |
    | `e2e_reception` | no | no | **401** |
    | `e2e_testmgmt` | yes | no | **500** |
    | `e2e_isadmin` | bypassed | yes | **200** |

13. **`isRestFullPath()` tests the UN-stripped path.** The interceptor keeps two
    different forms of the path: the DB lookup uses the `/rest`-stripped one
    (`URLUtil.getReourcePathFromRequest`), while `isRestFullPath()` reads the
    interceptor's own `path` field, which `preHandle` set to
    `requestURI - contextPath` with no stripping. Feeding it the stripped path
    makes the auto-allow rule fail for every unmapped endpoint
    (`/organization-list` does not start with `/rest`), denying the entire
    ported surface. Caught here before it shipped.

---

## 0. Why this exists now

The migration plan already mandates this work. `OpenELIS-Go-Migration-Plan.md`
§7 lists **P0 Foundations** as "domain, db, tx, audit, site/system config,
**auth skeleton**" — phase zero, before any feature — and its canonical tree
(§ "Go project layout") lists `internal/auth/  # login, role/rolemodule,
privilege checks`. `tech-stack-diff.md`'s Security row prescribes the shape:
*"custom middleware + `context.Context` principal — port role/rolemodule
privilege checks explicitly."*

**`internal/auth/` does not exist.** Waves a1/a2/b1/b2 got away with that
because they serve non-PHI reference data (locales, math functions, test
catalog, organizations). **c1 is the first wave that serves PHI** — names,
birth dates, national IDs, addresses, phone numbers, email — and the Go
service currently returns all of it to any unauthenticated caller:

```
$ curl -s "http://127.0.0.1:18090/rest/patientByLabNumer?accessionNumber=E2E001"
{"id":"1000","nationalId":"E2E-PAT-001","person":{"lastName":"TEST-Smith", ...
```

The c1 e2e spec has a PHI-boundary test that **genuinely fails against Go**; it
is currently `test.skip`ped under `go-parity` with an inline reason, and still
runs against Java. That skip is the marker for this work. **Deleting it is the
definition of done for Phase 1 below.**

Containment today: the Go service is bound to `127.0.0.1` in
`migration/openelis-go/docker-compose.go.yml`, and `cmd/openelis/main.go` logs
a warning at boot. That is a mitigation, not a fix.

---

## 1. Scope

**In scope (this plan):** everything needed for a ported read endpoint to
enforce the same access decision Java makes — session authentication, the
principal on `context.Context`, role lookups, the module-permission check, CSRF
for state-changing verbs, and the `/session` bootstrap contract.

**Out of scope, deliberately:** SAML, OAuth2/OIDC, client-certificate auth, and
HTTP Basic. All four exist in Java as separate filter chains but are
**disabled by default** (`@ConditionalOnProperty` gates on
`org.itech.login.saml` / `.oauth` / `.certificate`; Basic is
`matchIfMissing=true` so it IS on, but only matches requests carrying an
`Authorization: Basic` header, which the React frontend never sends). Porting
them is its own unit of work and nothing in waves a1–c3 needs them. § 9 records
what a later port must not forget.

**Also out of scope:** writing to any auth table. This plan is read-only
against `login_user`, `system_user`, `system_user_role`, `system_role`,
`system_role_module`, `system_module`, `system_module_url`,
`system_module_param`. No user creation, no password change, no lockout
administration.

---

## 2. The contract to reproduce (verified)

### 2.1 Which Java filter chain governs `/rest/**`

`SecurityConfig.java`'s `defaultSecurityConfigurationFilterChain`
(`@Order(Ordered.LOWEST_PRECEDENCE)`). It has **no `securityMatcher`** — it is
the catch-all. Rules that matter:

| Aspect | Value |
| --- | --- |
| Authorization | `anyRequest().authenticated()` — **default-deny, no URL-level role rules** |
| Always permitted | `LOGIN_PAGES` (`/LoginPage`, `/ValidateLogin`, `/session`), `RESOURCE_PAGES` (static), FORWARD/INCLUDE/ERROR dispatches |
| Login | `formLogin`, `loginProcessingUrl("/ValidateLogin")`, `usernameParameter("loginName")`, `passwordParameter("password")` |
| CSRF | **enabled**, ignoring only `/ValidateLogin` |
| Session | `sessionFixation().migrateSession()`, `invalidSessionUrl("/LoginPage")`, `SessionCreationPolicy.IF_REQUIRED` |
| Concurrency | **none** — no `maximumSessions`, no `SessionRegistry` |

Unauthenticated endpoints (`OPEN_PAGES`) that a port must keep open:
`/rest/open-configuration-properties`, `/rest/site-branding/**`,
`/rest/supportedlocales/active`, `/health/**`, `/pluginServlet/**`,
`/ChangePasswordLogin`, `/docs/UserManual`.

> Note: `/rest/supportedlocales/active` is already ported (a2) and is
> **legitimately anonymous**. `/rest/supportedlocales` (no `/active`) is **not**
> in `OPEN_PAGES` and therefore requires auth in Java — the Go port currently
> serves both anonymously. Closing that is part of Phase 1.

### 2.2 Login

`POST /ValidateLogin` (form-urlencoded `loginName` + `password`), handled by the
**stock** `UsernamePasswordAuthenticationFilter`. The response shape is decided
by a **custom** success handler, `CustomFormAuthenticationSuccessHandler`:

- `?apiCall=true` → **HTTP 200**, `application/json`, body exactly
  `{"success":true}`. No redirect. (This is what the React client and the e2e
  suite use.)
- browser → `SavedRequestAware` redirect.

Failure (`CustomAuthenticationFailureHandler`) with `apiCall=true` → **HTTP
401**, body `{"error":"<i18n-key>"}`, key ∈ `error.invalidcredentials`,
`error.expiredCredentials`, `error.disabledCredentials`,
`error.lockedCredentials`, `error.generic`. Note **unknown user and wrong
password produce the SAME key** — do not leak which.

### 2.3 Credentials

`login_user` — verified live against the dev DB:

```
login_name | password              | account_locked | account_disabled | is_admin | user_time_out | password_expired_dt
admin      | $2a$12$… (60 chars)   | N              | N                | Y        | 220           | 2031-07-10
```

- **BCrypt, cost 12**, `$2a$` prefix. Plain `BCryptPasswordEncoder` — **no
  `DelegatingPasswordEncoder`, no `{id}` prefixes, no legacy/plaintext
  fallback.** Go: `golang.org/x/crypto/bcrypt`, `CompareHashAndPassword`.
  It reads the cost from the hash, so cost 12 needs no configuration.
- Account state → login outcome:
  - `account_disabled='Y'` → `error.disabledCredentials`
  - `account_locked='Y'` → `error.lockedCredentials`
  - `password_expired_dt` in the past → `error.expiredCredentials`
  - `login_user` row absent → `error.invalidcredentials`
- **There is no failed-attempt counting and no automatic lockout.**
  `account_locked` is a manually administered flag. Do not invent lockout in
  the port — that would be a behavior change, not a migration.
- `login_user` joins to `system_user` **by `login_name` string**, not a FK, and
  requires `system_user.is_active='Y'`. If no active `system_user` exists, Java
  resolves `systemUserId = 0` and login fails at session setup with
  `login.error.noOeUser`.

### 2.4 Session

- Transport: **`JSESSIONID` cookie**. The frontend uses
  `credentials:"include"` everywhere; there is **no bearer token anywhere in
  this application**.
- Timeout: **per user** — `login_user.user_time_out` minutes × 60 seconds,
  default 20 minutes when null. (admin = 220 min in the dev DB.)
- The app's real "am I logged in?" test is **the presence of the
  `userSessionData` session attribute**, not the Spring `SecurityContext`
  (`UserModuleServiceImpl.isSessionExpired`). A session can hold a valid
  SecurityContext and still report `authenticated:false`.
- `UserSessionData` holds: `elisUserName`, `userTimeOut`, `systemUserId`,
  `loginName`, `isAdmin`, `loginLabUnit`.

### 2.5 `GET /session` — the bootstrap contract

Live response, authenticated (values elided):

```json
{"authenticated":true,"loginMethod":"FORM","sessionId":"…","userId":"1",
 "loginName":"admin","firstName":"Open","lastName":"ELIS",
 "roles":["Validation","Reception",…],
 "userLabRolesMap":{"AllLabUnits":["Reception","Results","Reports","Validation"]},
 "csrf":"…"}
```

- `authenticated` is `userSessionData != null`.
- `sessionId` is emitted **even when unauthenticated**.
- `roles` are `system_role.name` values, **trimmed** (see the padding trap in
  § 6.1).
- `csrf` is the token the client must echo back.

### 2.6 CSRF — the single most important implementation detail

Verified live: reading `GET /session` **twice in the same session returns two
different `csrf` values**:

```
read 1: "csrf":"z-38KHwjpBXmPmspDPzii7CDh-TnEKfYsPS2X…"
read 2: "csrf":"LK3euI_EPNFgdiUi-28SlZQsK4eTU0Us-TBHy…"
```

That is Spring Security 6's `XorCsrfTokenRequestAttributeHandler`: the
server-side token is stable, but every value handed to a client is
**XOR-masked with fresh random bytes and base64url-encoded**, then un-masked on
validation. A Go implementation that stores a token and does `==` against the
submitted value **will reject every request**.

- Repository: `HttpSessionCsrfTokenRepository` (session attribute), default
  parameter `_csrf`, default header `X-CSRF-TOKEN`.
- Clients send `X-CSRF-Token` (case-insensitively the same header).
- Applies to **state-changing verbs only**. Every c1/c2/c3 endpoint is a GET,
  so **CSRF is not on the critical path for the current waves** — but it is
  mandatory before any write wave (e1/f1).

Mask/unmask algorithm (Spring's, to reproduce):
`masked = base64url( randomBytes(n) || (tokenBytes XOR randomBytes) )` where
`n = len(tokenBytes)`; unmask splits the halves and XORs back.

### 2.7 Authorization — and why c1 needs less than it looks

Two independent mechanisms:

**(a) `ModuleAuthenticationInterceptor`** — registered on `/**`, excluding
`OPEN_PAGES`/`LOGIN_PAGES`/`RESOURCE_PAGES`/`AUTH_OPEN_PAGES`. Note `/rest/**`
is **not** excluded, so it does run for REST.

Its URL→permission lookup strips the `/rest` prefix
(`URLUtil.getReourcePathFromRequest`), then looks for a `system_module_url` row
whose `url_path` **exactly equals** the stripped path. Then:

- **row found** → allow iff the user's permitted-module set contains that
  module's name.
- **no row found** → **if the path starts with `/rest`, ALLOW** (auto-allow for
  any authenticated user). Otherwise deny.

The auto-allow is intentional and documented in-source: *"REST endpoints
without SystemModuleUrl DB entries are auto-allowed for any authenticated
user. Admin-only controllers are protected via `@PreAuthorize`."*

**Verified against the live DB — none of c1's four paths are mapped:**

```sql
SELECT u.url_path, m.name FROM system_module_url u
  JOIN system_module m ON m.id=u.system_module_id
 WHERE u.url_path ILIKE '%patient%';
-- 15 rows, all for OTHER endpoints:
-- /PatientConfiguration, /PatientResults, /SamplePatientEntry, /PatientType, …
-- NOT /patientByLabNumer, /patient/merge/details, /patient-id-documents, /patient-photos
```

**So for c1, the module check is a no-op: authentication alone is the gate.**
That is a large simplification — Phase 1 can ship real parity for c1 without
implementing the module system at all.

**(b) Programmatic role checks** in controllers. The one that touches c1:
`PatientMergeRestController.hasMergePermission` requires the **`Reception`**
role and returns a **bodiless 403** otherwise. Others exist for later waves
(`SampleEdit` cancel-results requires `Validator`/`Validation`/`Biologist`;
`StorageLocation` requires `Global Administrator`).

**Mode:** `permissions.agent` = `Role` (from `SystemConfiguration.properties`;
verified there is **no `site_information` override** in the dev DB, so the file
default applies). In Role mode the permitted-module set is
`system_user_role → system_role_module → system_module.name`.
`system_user_module` is **not consulted** (and is empty: 0 rows).

### 2.8 Denial responses — three different shapes

| Source | Condition | Status | Body |
| --- | --- | --- | --- |
| Spring Security entry point | unauthenticated, any path | **302** → `/LoginPage` | — |
| Spring `accessDeniedHandler` | authenticated but denied, path starts `/rest`/`/Provider`/`/api/OpenELIS-Global/rest` | **403** | `{"status":403,"message":"Access denied"}` or `…"CSRF token missing or invalid"` |
| `ModuleAuthenticationInterceptor` | module check failed, path starts `/rest`/`/Provider`/`/dbImage`/`/logging` | **401** | `{"status":401,"message":"Not Authorized"}` |
| `ModuleAuthenticationInterceptor` | module check failed, other paths | 302 | `→ /Home?access=denied` |

The 401-vs-403 split is not a typo: the interceptor uses 401 for an
*authorization* failure, which is semantically wrong but is the observable
contract. **Reproduce it; do not correct it.** (Migration policy: pin Java's
behavior, raise bugs separately.)

---

## 3. Go design

```
internal/auth/
  valueholder/   login_user.go        LoginUser, SystemUser
                 role.go              Role, SystemModule
  daoimpl/       login_dao_impl.go    credential + system_user lookups
                 role_dao_impl.go     roles, role→module, module_url
  service/       auth_service.go      Authenticate(), BuildPrincipal()
                 authz_service.go     module resolution + role checks
  session/       store.go             Store interface + in-memory impl
                 cookie.go            JSESSIONID issue/read
  csrf/          xor.go               Spring-compatible mask/unmask
  middleware/    require_auth.go      default-deny, principal → context
                 require_csrf.go      state-changing verbs only
  controller/rest/
                 login.go             POST ValidateLogin, GET session, POST Logout
  form/          session_form.go      the /session DTO
```

Layering follows the constitution (I valueholder → II daoimpl → III service →
IV controller → V form), same as the b2 refactor and c1's implementation.

### 3.1 The principal on the context

Per `tech-stack-diff.md`'s prescription:

```go
type Principal struct {
    SystemUserID int64
    LoginName    string
    FirstName    string
    LastName     string
    IsAdmin      bool
    Roles        map[string]bool // TRIMMED role names — see § 6.1
    Modules      map[string]bool // permitted module names (Role mode)
    LoginLabUnit int64
}

func FromContext(ctx context.Context) (*Principal, bool)
```

Handlers read it from `r.Context()`. Nothing else gets a session handle.

### 3.2 Where middleware hooks in

`internal/common/web.Register` is already the single choke point — **every**
ported route in the codebase goes through it (verified: 26 call sites, no
route bypasses it). Auth slots in there:

```go
func Register(mux *http.ServeMux, method, restPath string, h http.HandlerFunc) {
    wrapped := RequireAuth(RequireCSRF(h))   // default-deny by construction
    mux.HandleFunc(method+" /api/OpenELIS-Global/"+restPath, wrapped)
    mux.HandleFunc(method+" /"+restPath, wrapped)
}
```

**Default-deny is the design invariant**: a new route is protected unless it is
explicitly registered as open. Provide `web.RegisterOpen(...)` for the
`OPEN_PAGES` set, and require a code comment naming the Java justification at
each call site. This is the opposite of opt-in protection and is the single
most important structural decision here — it makes "forgot to add auth" a
non-event.

### 3.3 Session store

Java uses the servlet container's in-memory `HttpSession`. Phase 1 mirrors
that with an in-memory store behind an interface:

```go
type Store interface {
    New(p *Principal, ttl time.Duration) (id string, err error)
    Get(id string) (*Principal, bool)
    Refresh(id string) // sliding expiry, matching maxInactiveInterval
    Delete(id string)
}
```

Cookie: name `JSESSIONID` (must match — the frontend and e2e suite send
whatever the server set), `HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/`.
`web.xml` sets `http-only` and `secure` on the Java side; match both.

Session id: 128 bits from `crypto/rand`, base64url. **Not** a UUID — Java's
container ids are opaque and nothing parses them.

Rotate the id on login (Java does `sessionFixation().migrateSession()`).

**Known limitation, stated up front:** an in-memory store means sessions die on
restart and cannot be shared across replicas. Java has the same limitation
today (no session replication configured), so this is parity, not a
regression. If the strangler proxy ever routes the same user to both Java and
Go, **the two will not share a session** — see § 8.

---

## 4. Phased plan

Each phase is independently shippable and independently verifiable.

### Phase 1 — Authentication (unblocks c1) — **DONE** (see § 0.1)
**Goal: delete the `test.skip` in the c1 PHI-boundary test.**

1. `internal/auth/valueholder` + `daoimpl`: load `login_user` by `login_name`;
   join `system_user` by `login_name` with `is_active='Y'`.
2. bcrypt verification; account state → the five error keys.
3. Session store + `JSESSIONID` cookie + per-user TTL from `user_time_out`.
4. `POST /ValidateLogin` returning exactly `{"success":true}` / 401
   `{"error":"…"}`.
5. `GET /session` returning the full bootstrap DTO (roles included).
6. `POST /Logout` invalidating the session.
7. `RequireAuth` middleware wired into `web.Register`; `RegisterOpen` for the
   `OPEN_PAGES` set.
8. Fix the a2 over-exposure: `/rest/supportedlocales` (without `/active`)
   must require auth.

**Exit criteria**
- The c1 PHI-boundary test passes against Go **with the skip removed**.
- `go-parity` login flow uses the *same* `auth.setup.ts` as Java — no
  Go-specific auth path in the suite.
- An anonymous `GET /rest/patientByLabNumer?...` returns non-200 and leaks
  no PHI.
- a1/a2/b1/b2/c1 parity still green **through an authenticated session**.

### Phase 2 — Authorization — **DONE** (see § 0.1; the module system was NOT
deferrable after all — § 0.2.9)
1. Role loading (`system_user_role` → `system_role`), trimmed.
2. `RequireRole("Reception")` for `merge/details`, producing a **bodiless
   403** to match Java.
3. Module resolution: `system_module_url` exact-match on the `/rest`-stripped
   path, `system_module_param` filtering, permitted-module set from
   `system_role_module`.
4. The **auto-allow** rule for unmapped `/rest` paths — reproduce exactly,
   including the deny for unmapped non-`/rest` paths.
5. Denial shapes: 401 `{"status":401,"message":"Not Authorized"}` from the
   module check; 403 `{"status":403,"message":"Access denied"}` from the
   access-denied path.

**Exit criteria:** a low-privilege user is denied where Java denies, with the
same status and body. Requires seeding a non-admin test user (§ 7).

### Phase 3 — CSRF (gate for every write wave) — **DONE**, and it had to ship
in Phase 1, not Phase 3 — see § 0.2.1
1. Spring-compatible XOR mask/unmask.
2. Token minted per session, exposed via `/session` as `csrf`.
3. `RequireCSRF` on state-changing verbs only, reading `X-CSRF-Token`;
   `/ValidateLogin` exempt.
4. 403 `{"status":403,"message":"CSRF token missing or invalid"}`.

**Exit criteria:** the React app can log in against Go and perform a write
without changing a line of frontend code.

---

## 5. Testing strategy

Follows the suite's existing rules ([openelis-api-e2e.md](openelis-api-e2e.md)):
no mocking, real live stack, DB oracles, inversion-tested.

- **Reuse `tests/auth.setup.ts` unchanged.** It already performs the real
  handshake (touch `session` → POST `ValidateLogin` → verify). If it works
  against Go without modification, the login contract is right. If it needs a
  Go-specific branch, the port is wrong.
- **New `e2e/auth-parity.spec.ts`** (naming per branch policy — ask before
  adding, per CLAUDE.md), asserting against **both** servers:
  - valid login → 200 `{"success":true}`, `JSESSIONID` set, `HttpOnly`
  - wrong password / unknown user → **401 with the SAME error key** (proves
    no user enumeration)
  - disabled / locked / expired accounts → their distinct keys
  - `/session` anonymous vs authenticated shapes
  - anonymous access to a protected endpoint → non-200, no PHI in the body
  - logout → subsequent request unauthenticated
  - **`Reception`-gated `merge/details` → bodiless 403 for a user without it**
- **Inversion-test every one** (Constitution V.6): each must fail against the
  pre-auth binary. A PHI-boundary test that passes both before and after is
  not testing anything.
- **Never hardcode a password in a spec.** Read from `fixtures/env.ts`, which
  already carries `ADMIN_USER`/`ADMIN_PASS`.

---

## 6. Verified traps

### 6.1 `system_role.name` is `character(30)` — blank-padded
```sql
SELECT id, name FROM system_role WHERE id=4;   -- "Reception" + 21 spaces
SELECT '[' || name || ']' FROM system_role;    -- "[Reception]"  (bpchar→text trims)
```
The raw column value **is padded**; the `||` cast trims it, which is why casual
inspection hides the problem. Go scanning that column gets the padded string,
so `role == "Reception"` **silently fails**. Java trims aggressively at every
comparison site. **Trim on read, once, in the DAO** — and cover it with a test
that would fail if the trim were removed.

### 6.2 The CSRF token is XOR-masked and differs on every read
See § 2.6, verified live. Store-and-compare will reject everything.

### 6.3 `authenticated` is not "has a SecurityContext"
It is `session.userSessionData != null`. Keep those concepts merged in Go (one
principal in one session) and this trap disappears — but the `/session`
contract must still report the same boolean Java would.

### 6.4 The interceptor's `path` field is not thread-safe
`ModuleAuthenticationInterceptor` stores the request path in a **mutable
instance field on a singleton**, then reads it in `isRestFullPath()`. Under
concurrency this races and can return an HTML redirect to a REST caller. **Go
must carry the path per-request** — this is a bug not to reproduce, because it
is nondeterministic rather than a contract. (Distinct from the deliberate
bug-pinning policy: you cannot pin a race.)

### 6.5 `/rest` stripping is `split("/rest")[1]`
Consequences to reproduce or consciously reject: a path of exactly `/rest`
throws `ArrayIndexOutOfBoundsException`; `/rest/a/rest/b` yields `/a`;
`/restore/x` is also stripped because the test is `startsWith("/rest")`.
Recommend: implement `strings.TrimPrefix(path, "/rest")` and **document the
divergence** rather than reproducing an exception.

### 6.6 Unknown user and wrong password return the same key
`error.invalidcredentials` for both. Preserve — collapsing them differently
would introduce user enumeration.

### 6.7 Sessions are not shared between Java and Go
Both are in-memory and per-process. During strangler coexistence a user
authenticated against Java is **not** authenticated against Go. See § 8.

---

## 7. Test data needed

The dev DB has 17 roles but only 10 `system_user_role` grants, and `admin` is
`is_admin='Y'` — which **bypasses every module check**
(`hasPermission(...) || isUserAdmin(...)`). Testing authorization with admin
alone proves nothing.

Phase 2 needs a fixture seeding at least:
- a **non-admin** user with a known password and exactly one narrow role,
- a user with **no** roles (to prove denial),
- a user **with** `Reception` and one **without** (for `merge/details`).

Follow the existing convention — an idempotent SQL fixture under
`src/test/resources/fixtures/` loadable by `load-test-fixtures.sh`, in a
reserved id range, exactly like `patient-media-e2e.sql`. Hashes must be real
bcrypt (cost 12) generated for the fixture, and the password must live in
`fixtures/env.ts`, never inline in a spec.

---

## 8. Deployment / strangler implications

1. **Session sharing is the blocker for a real cutover.** Today nginx routes
   whole paths to Go. The moment an authenticated user's request goes to Go,
   their Java session is meaningless. Options, in increasing order of work:
   (a) keep Go loopback-only and parity-test only (status quo);
   (b) shared session store (Redis / DB-backed) written by both — invasive on
   the Java side; (c) route by user, not by path — impractical.
   **Recommendation: (a) until Java is retired for a whole path group.**
   This should be decided before Phase 1 ships, because it determines whether
   the session store needs to be pluggable from day one.
2. **Un-bind from loopback only after Phase 1 passes.**
   `docker-compose.go.yml` currently pins `127.0.0.1:8090:8090` precisely
   because of the PHI exposure.
3. **HTTPS.** Java forces `CONFIDENTIAL` transport for `/*` via `web.xml`. Go
   sits behind the same nginx; the `Secure` cookie flag depends on that
   termination. Do not set `Secure` conditionally on a dev flag that could ship.

---

## 9. Open decisions (need a human)

1. **Session store**: in-memory (parity, simple) vs shared/persistent (enables
   real cutover). Blocks the § 8.1 decision.
2. **Do we port the module system at all (Phase 2)?** None of a1–c3's
   endpoints are mapped in `system_module_url`, so auto-allow covers them
   entirely. It is genuinely deferrable until a wave needs it — but deferring
   means Go is *more permissive* than Java for any endpoint someone later adds
   a mapping for, and nothing would flag that.
3. **`/rest` stripping** — reproduce the `split` quirks or implement clean
   `TrimPrefix` and document the divergence? Recommend the latter.
4. **The unported chains** (SAML / OAuth / cert / Basic). Fine to skip now;
   needs an explicit "not supported" answer rather than silence, because a
   deployment with `org.itech.login.saml=true` would behave differently
   between Java and Go.
5. **Branch/track.** Auth is P0 Foundations, not a wave — it does not fit the
   `migration/<type><seq>` convention. Proposed `migration/p0-auth`, forked
   from `migration-base` per `branch-naming.md`. Needs confirmation.

---

## 10. Explicitly NOT doing

Per migration policy — pin Java's behavior, never "fix" it in the port:

- **No lockout / attempt counting.** Java has none.
- **No password-complexity enforcement** on this path (read-only plan).
- **Not correcting the 401-vs-403 semantics** of the module interceptor.
- **Not correcting** `UserSessionData.userTimeOut` being stored as
  minutes×3600 while the real session timeout is minutes×60 (a real Java bug;
  only matters if something reads that field — nothing in the ported waves
  does).
- **Not touching** the committed default-admin bcrypt hash in
  `src/main/resources/adminPassword.txt`, nor the Jasypt `TextEncryptor`
  defaulting to the password `"dev"`. Both are real findings, both belong in a
  security report to the maintainers, **neither is a migration task.**
