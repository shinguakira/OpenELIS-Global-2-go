// P0 Foundations — AUTHORIZATION parity oracle (companion to p0-auth.spec.ts,
// which covers authentication).
//
// Java gates a handful of the already-ported b1 endpoints with TWO INDEPENDENT
// mechanisms. They are evaluated in a fixed order and they use DIFFERENT
// definitions of "admin":
//
//   1. ModuleAuthenticationInterceptor — a HandlerInterceptor on /**. Looks the
//      /rest-stripped path up in system_module_url; if a row exists, the user's
//      permitted-module set must contain that module. Unmapped /rest paths are
//      AUTO-ALLOWED. Bypassed by login_user.is_admin='Y' ALONE.
//      Denial: 401 { "status": 401, "message": "Not Authorized" }
//
//   2. @PreAuthorize("hasRole('ADMIN')") — class-level on three ported
//      controllers. Granted by is_admin='Y' OR the Global Administrator role.
//      Denial: HTTP 500 (see below).
//
// The 500 is Java's, not a mistake in this spec: an AccessDeniedException raised
// by method security never reaches SecurityConfig's accessDeniedHandler, so it
// surfaces as an unhandled error instead of a 403. Migration policy pins Java's
// observable behavior and raises bugs separately (auth-adoption-plan.md §2.8
// sets that precedent for denial shapes specifically), so the port reproduces
// it and this spec asserts it. Verified live on both stacks.
//
// WHY THE USER CAST IS THIS SPECIFIC: the stock `admin` account holds the
// Global Administrator role AND is_admin='Y' AND every module, so it satisfies
// both mechanisms for every reason at once. Asserted with `admin` alone, a port
// that implemented either gate wrongly — or collapsed the two into one — would
// still pass. Each fixture user below exists to close one of those holes:
//
//   user              is_admin  roles                  TestCatalog module?
//   e2e_reception     N         Reception              no
//   e2e_testmgmt      N         Test Management        YES
//   e2e_globaladmin   N         Global Administrator   YES
//   e2e_isadmin       Y         (none)                 no (bypassed)
//
// Fixture: src/test/resources/fixtures/auth-e2e.sql.
import {
  test,
  expect,
  request as apiRequest,
  type APIRequestContext,
} from "@playwright/test";
import {
  E2E_PASS,
  E2E_USERS,
  E2E_AUTHZ_USERS,
  ADMIN_USER,
  ADMIN_PASS,
} from "../../fixtures/env";
import {
  LOGIN_PATH,
  LOGIN_USER_FIELD,
  LOGIN_PASS_FIELD,
  SESSION_PATH,
} from "../../fixtures/contract";
import { query } from "../../fixtures/db";
import { readJson } from "../../fixtures/assert";

// Endpoints under test. All are already ported (b1) and all were served to any
// authenticated caller by the Go service until P0 Phase 2 landed.
const TEST_CATALOG = "rest/TestCatalog"; // module-mapped AND admin-gated
const ADMIN_ONLY = [
  "rest/dictionary-categories",
  "rest/test-catalog/lab-units",
  "rest/test-catalog/panels",
  "rest/test-catalog/sample-types",
];
// Unmapped in system_module_url and carrying no @PreAuthorize: the control.
const UNGATED = "rest/organization-list";

function targetBaseURL(): string {
  const b = test.info().project.use.baseURL;
  if (!b) throw new Error("project has no baseURL — cannot pick a target");
  return b;
}

/** Log in as `user` and return the authenticated context. */
async function loginAs(user: string, pass: string): Promise<APIRequestContext> {
  const ctx = await apiRequest.newContext({
    baseURL: targetBaseURL(),
    ignoreHTTPSErrors: true,
    storageState: { cookies: [], origins: [] },
  });
  await ctx.get(SESSION_PATH);
  const res = await ctx.post(LOGIN_PATH, {
    form: { [LOGIN_USER_FIELD]: user, [LOGIN_PASS_FIELD]: pass },
  });
  expect(res.status(), `login as ${user}`).toBe(200);
  return ctx;
}

const asFixtureUser = (u: string) => loginAs(u, E2E_PASS);

/** The interceptor's denial: 401 with that exact body, spacing included. */
async function expectModuleDenied(ctx: APIRequestContext, path: string) {
  const res = await ctx.get(path);
  expect(res.status(), `${path} module denial status`).toBe(401);
  expect(
    (res.headers()["content-type"] ?? "").toLowerCase(),
    `${path} module denial content-type`,
  ).toContain("application/json");
  expect(await res.json(), `${path} module denial body`).toEqual({
    status: 401,
    message: "Not Authorized",
  });
}

/** The @PreAuthorize denial: Spring's unhandled-error body, HTTP 500. */
async function expectAdminDenied(ctx: APIRequestContext, path: string) {
  const res = await ctx.get(path);
  expect(res.status(), `${path} admin-gate denial status`).toBe(500);
  const body = await res.json();
  // timestamp is volatile — pin its TYPE and the rest exactly.
  expect(typeof body.timestamp, `${path} denial timestamp is a number`).toBe(
    "number",
  );
  const { timestamp, ...rest } = body;
  expect(rest, `${path} admin-gate denial body`).toEqual({
    status: 500,
    error: "Internal Server Error",
  });
}

/** An allowed read: 200 JSON with actual content. */
async function expectAllowed(ctx: APIRequestContext, path: string) {
  const body = await readJson(await ctx.get(path), path);
  expect(
    Array.isArray(body) ? body.length : Object.keys(body).length,
    `${path} returns real content`,
  ).toBeGreaterThan(0);
}

// ───────────────────────────────────────────────────────────────────────────
// 0. The premises, from the database — so the tests below mean what they claim
// ───────────────────────────────────────────────────────────────────────────
test.describe("p0-authz: database premises", () => {
  test("/TestCatalog IS mapped in system_module_url and the others are not", async () => {
    const mapped = query(
      "SELECT m.name FROM clinlims.system_module_url u" +
        " JOIN clinlims.system_module m ON m.id = u.system_module_id" +
        " WHERE u.url_path = '/TestCatalog';",
    ).map((r) => r[0]);
    expect(mapped, "/TestCatalog maps to the TestCatalog module").toEqual([
      "TestCatalog",
    ]);

    // The auto-allow premise: the control endpoint has no mapping at all, which
    // is why any authenticated user may read it.
    const unmapped = query(
      "SELECT count(*) FROM clinlims.system_module_url" +
        " WHERE url_path = '/organization-list';",
    );
    expect(unmapped[0][0], "/organization-list has no module mapping").toBe("0");
  });

  test("the fixture users differ in exactly the way the tests assume", async () => {
    const modulesFor = (loginName: string) =>
      query(
        "SELECT m.name FROM clinlims.system_user u" +
          " JOIN clinlims.system_user_role sur ON sur.system_user_id = u.id" +
          " JOIN clinlims.system_role_module srm ON srm.system_role_id = sur.role_id" +
          " JOIN clinlims.system_module m ON m.id = srm.system_module_id" +
          ` WHERE u.login_name = '${loginName}' AND m.name = 'TestCatalog';`,
      ).length;

    expect(modulesFor(E2E_USERS.reception), "reception lacks TestCatalog").toBe(
      0,
    );
    expect(
      modulesFor(E2E_AUTHZ_USERS.testMgmt),
      "testMgmt HAS TestCatalog",
    ).toBeGreaterThan(0);

    // is_admin is the module bypass; the role is what grants hasRole('ADMIN').
    const [[isAdminFlag]] = query(
      `SELECT is_admin FROM clinlims.login_user WHERE login_name = '${E2E_AUTHZ_USERS.isAdmin}';`,
    );
    expect(isAdminFlag, "isAdmin user has is_admin='Y'").toBe("Y");
    const roleCount = query(
      "SELECT count(*) FROM clinlims.system_user_role WHERE system_user_id =" +
        ` (SELECT id FROM clinlims.system_user WHERE login_name = '${E2E_AUTHZ_USERS.isAdmin}');`,
    );
    expect(roleCount[0][0], "isAdmin user holds NO roles").toBe("0");

    const [[gaAdminFlag]] = query(
      `SELECT is_admin FROM clinlims.login_user WHERE login_name = '${E2E_AUTHZ_USERS.globalAdmin}';`,
    );
    expect(gaAdminFlag, "globalAdmin user has is_admin='N'").toBe("N");
  });
});

// ───────────────────────────────────────────────────────────────────────────
// 1. The module check (ModuleAuthenticationInterceptor)
// ───────────────────────────────────────────────────────────────────────────
test.describe("p0-authz: module permission check", () => {
  test("a user without the mapped module is refused with 401", async () => {
    const ctx = await asFixtureUser(E2E_USERS.reception);
    await expectModuleDenied(ctx, TEST_CATALOG);
    await ctx.dispose();
  });

  test("an UNMAPPED /rest path is auto-allowed for any authenticated user", async () => {
    // The auto-allow rule, reproduced deliberately from Java's in-source note:
    // "REST endpoints without SystemModuleUrl DB entries are auto-allowed for
    // any authenticated user." It is why most ported endpoints need no module
    // data — and why rest/TestCatalog was the one that slipped through.
    const ctx = await asFixtureUser(E2E_USERS.reception);
    await expectAllowed(ctx, UNGATED);
    await ctx.dispose();
  });

  test("a user with ZERO roles is still auto-allowed on unmapped paths", async () => {
    // Distinguishes "auto-allow" from "allowed because the user happens to hold
    // some module": this user's permitted-module set is empty.
    const ctx = await asFixtureUser(E2E_USERS.noRoles);
    await expectAllowed(ctx, UNGATED);
    await ctx.dispose();
  });

  test("is_admin='Y' bypasses the module check even with no roles at all", async () => {
    // The ONLY user that isolates this: an empty module set plus a MAPPED path
    // would be a 401 without the bypass. Proves the bypass is is_admin, not
    // "holds the module" and not "has the Global Administrator role".
    const ctx = await asFixtureUser(E2E_AUTHZ_USERS.isAdmin);
    await expectAllowed(ctx, TEST_CATALOG);
    await ctx.dispose();
  });
});

// ───────────────────────────────────────────────────────────────────────────
// 2. The @PreAuthorize("hasRole('ADMIN')") gate
// ───────────────────────────────────────────────────────────────────────────
test.describe("p0-authz: admin-only endpoints", () => {
  for (const path of ADMIN_ONLY) {
    test(`${path} refuses a non-admin`, async () => {
      const ctx = await asFixtureUser(E2E_USERS.reception);
      await expectAdminDenied(ctx, path);
      await ctx.dispose();
    });

    test(`${path} allows the admin`, async () => {
      const ctx = await loginAs(ADMIN_USER, ADMIN_PASS);
      await expectAllowed(ctx, path);
      await ctx.dispose();
    });
  }

  test("the Global Administrator ROLE grants access without is_admin='Y'", async () => {
    // Closes the hole where a port implements the gate as `is_admin == 'Y'`.
    // Spring's hasRole('ADMIN') is satisfied by EITHER, and the stock `admin`
    // account has both — so only this user can tell the two apart.
    const ctx = await asFixtureUser(E2E_AUTHZ_USERS.globalAdmin);
    for (const path of ADMIN_ONLY) await expectAllowed(ctx, path);
    await ctx.dispose();
  });

  test("is_admin='Y' alone grants access without any role", async () => {
    // The mirror of the above: closes the hole where a port implements the gate
    // as "holds the Global Administrator role".
    const ctx = await asFixtureUser(E2E_AUTHZ_USERS.isAdmin);
    for (const path of ADMIN_ONLY) await expectAllowed(ctx, path);
    await ctx.dispose();
  });
});

// ───────────────────────────────────────────────────────────────────────────
// 3. The ORDER of the two mechanisms — the sharpest test in this file
// ───────────────────────────────────────────────────────────────────────────
test.describe("p0-authz: the two gates are independent and ordered", () => {
  test("rest/TestCatalog: module denial (401) precedes the admin gate (500)", async () => {
    // Same endpoint, two non-admin users, two DIFFERENT refusals:
    //   e2e_reception — no TestCatalog module -> the interceptor refuses first
    //   e2e_testmgmt  — HAS the module        -> gets past it, then the
    //                                            @PreAuthorize gate refuses
    // A port that runs only one of the two, or runs them in the other order,
    // produces the same status for both users and fails here. This is the test
    // that the stock `admin` account can never substitute for.
    const reception = await asFixtureUser(E2E_USERS.reception);
    await expectModuleDenied(reception, TEST_CATALOG);
    await reception.dispose();

    const testMgmt = await asFixtureUser(E2E_AUTHZ_USERS.testMgmt);
    await expectAdminDenied(testMgmt, TEST_CATALOG);
    await testMgmt.dispose();
  });

  test("the module holder is still refused on the plain admin-only endpoints", async () => {
    // Sanity for the above: e2e_testmgmt's 500 on rest/TestCatalog is the admin
    // gate, not something specific to that path — it gets the same 500 on the
    // endpoints that have no module mapping at all.
    const ctx = await asFixtureUser(E2E_AUTHZ_USERS.testMgmt);
    for (const path of ADMIN_ONLY) await expectAdminDenied(ctx, path);
    await ctx.dispose();
  });
});
