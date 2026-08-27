// Shared environment/config for the OpenELIS API parity suite.
// NOTE: trailing slash is required — Playwright resolves request paths with
// `new URL(path, baseURL)`, so paths must be relative (no leading slash) to keep
// the /api/OpenELIS-Global prefix.
export const BASE_URL =
  process.env.OE_BASE_URL || "https://localhost/api/OpenELIS-Global/";

// The Go port under test, for side-by-side parity: the SAME checks that run
// against Java (BASE_URL) also run against this base. Trailing slash required.
export const GO_BASE_URL = process.env.OE_GO_URL || "http://localhost:8090/";

export const ADMIN_USER = process.env.OE_ADMIN_USER || "admin";
export const ADMIN_PASS = process.env.OE_ADMIN_PASS || "adminADMIN!";

// Container name of the PostgreSQL DB (for the DB oracle).
export const DB_CONTAINER =
  process.env.OE_DB_CONTAINER || "openelisglobal-database";

// How to reach `docker` from the test runner.
//  - On Linux/CI where docker is on PATH:        OE_DOCKER="docker"
//  - On this Windows dev host (docker in WSL):   default → wsl.exe wrapper
export const DOCKER_MODE = process.env.OE_DOCKER || "wsl"; // "wsl" | "docker"
export const WSL_DISTRO = process.env.OE_WSL_DISTRO || "Ubuntu-24.04";

export const AUTH_STATE = "playwright/.auth/admin.json";

// Separate cookie jar for the Go port: the SAME auth.setup.ts handshake runs
// against both targets, but the two services do not share sessions (both keep
// them in process memory), so their jars must not overwrite each other.
export const GO_AUTH_STATE = "playwright/.auth/go-admin.json";

// ── Auth/authz parity users (src/test/resources/fixtures/auth-e2e.sql) ───────
// The dev DB ships exactly one login_user, `admin`, with is_admin='Y' — which
// short-circuits every module/role check in Java
// (ModuleAuthenticationInterceptor.hasPermission: `... || isUserAdmin(...)`).
// Authorization asserted with admin alone therefore proves nothing, so the
// fixture seeds a non-admin cast covering each account state. All of them share
// one password; it lives here (never inline in a spec) per the suite's rules.
export const E2E_PASS = process.env.OE_E2E_PASS || "e2eAUTH!pass1";

// A wrong password that is NOT any user's — for the bad-credentials branch and
// the user-enumeration check (unknown user and wrong password must not be
// distinguishable).
export const E2E_WRONG_PASS = process.env.OE_E2E_WRONG_PASS || "not-the-password";

export const E2E_USERS = {
  /** non-admin, exactly one role: Reception */
  reception: process.env.OE_E2E_USER_RECEPTION || "e2e_reception",
  /** authenticates, holds zero roles */
  noRoles: process.env.OE_E2E_USER_NOROLES || "e2e_noroles",
  /** account_locked='Y' */
  locked: process.env.OE_E2E_USER_LOCKED || "e2e_locked",
  /** account_disabled='Y' */
  disabled: process.env.OE_E2E_USER_DISABLED || "e2e_disabled",
  /** password_expired_dt in the past */
  expired: process.env.OE_E2E_USER_EXPIRED || "e2e_expired",
  /** valid credentials but no ACTIVE system_user carries the login name */
  noOeUser: process.env.OE_E2E_USER_NOOEUSER || "e2e_noouser",
  /** user_time_out=999 (vs the 20-minute default) */
  longTimeout: process.env.OE_E2E_USER_LONGTIMEOUT || "e2e_longtimeout",
} as const;

/** A login name guaranteed absent from login_user — for the unknown-user branch. */
export const E2E_UNKNOWN_USER =
  process.env.OE_E2E_UNKNOWN_USER || "e2e_no_such_user";

// ── Authorization matrix users (Phase 2) ────────────────────────────────────
// Java gates a few ported b1 endpoints with TWO independent mechanisms that use
// DIFFERENT definitions of "admin". These four users pull them apart; with only
// `admin` (which happens to hold both) a port could collapse them and still
// pass. See tests/readonly/p0-authz.spec.ts.
export const E2E_AUTHZ_USERS = {
  /** Test Management role: HAS the TestCatalog module, is NOT an ADMIN. */
  testMgmt: process.env.OE_E2E_USER_TESTMGMT || "e2e_testmgmt",
  /** Global Administrator ROLE but login_user.is_admin='N'. */
  globalAdmin: process.env.OE_E2E_USER_GLOBALADMIN || "e2e_globaladmin",
  /** login_user.is_admin='Y' with ZERO roles (so an empty module set). */
  isAdmin: process.env.OE_E2E_USER_ISADMIN || "e2e_isadmin",
} as const;
