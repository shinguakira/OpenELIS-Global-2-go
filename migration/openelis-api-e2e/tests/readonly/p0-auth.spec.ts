// P0 Foundations — authentication & authorization parity oracle.
//
// Source of truth: migration/auth-adoption-plan.md. Every assertion below was
// derived from the Java source (SecurityConfig, CustomUserDetailsService,
// CustomFormAuthenticationSuccessHandler, CustomAuthenticationFailureHandler,
// LoginPageController#getSesssionDetails, ModuleAuthenticationInterceptor) and
// then confirmed against the live stack — no assertion is written from memory.
//
// This spec is the parity oracle for `internal/auth/` in the Go port, so it runs
// under BOTH the `api-readonly` (Java) and `go-parity` (Go) projects. It never
// imports BASE_URL directly: the target comes from the project's own baseURL, so
// retargeting is config, not an edit.
//
// Fixture: src/test/resources/fixtures/auth-e2e.sql (ids 9900-9999), loaded by
// ./src/test/resources/load-test-fixtures.sh. The dev DB otherwise ships only
// `admin` (is_admin='Y'), which short-circuits every module/role check —
// authorization asserted with admin alone proves nothing.
//
// INVERSION (Constitution V.6): every test here fails against a Go binary
// without `internal/auth/` — the boundary tests because the data is served
// anonymously, the login/session tests because the routes 404. A test that
// passes both before and after the port is not testing anything.
import {
  test,
  expect,
  request as apiRequest,
  type APIRequestContext,
  type APIResponse,
} from "@playwright/test";
import {
  E2E_PASS,
  E2E_WRONG_PASS,
  E2E_USERS,
  E2E_UNKNOWN_USER,
  ADMIN_USER,
  ADMIN_PASS,
} from "../../fixtures/env";
import {
  LOGIN_PATH,
  LOGIN_USER_FIELD,
  LOGIN_PASS_FIELD,
  SESSION_PATH,
} from "../../fixtures/contract";
import { unmaskCsrf, maskCsrf, differentAsciiToken } from "../../fixtures/csrf";
import { query } from "../../fixtures/db";
import { readJson, expectNonEmptyString } from "../../fixtures/assert";

// ── harness ────────────────────────────────────────────────────────────────
// The target under test is whichever project is running (Java: api-readonly,
// Go: go-parity). Reading it from the project config is what makes one spec
// serve as the oracle for both.
function targetBaseURL(): string {
  const b = test.info().project.use.baseURL;
  if (!b) throw new Error("project has no baseURL — cannot pick a target");
  return b;
}

/** A context with an EMPTY cookie jar: every test drives its own session. */
async function anonCtx(): Promise<APIRequestContext> {
  return apiRequest.newContext({
    baseURL: targetBaseURL(),
    ignoreHTTPSErrors: true,
    storageState: { cookies: [], origins: [] },
  });
}

function loginWith(
  ctx: APIRequestContext,
  loginName: string,
  password: string,
): Promise<APIResponse> {
  return ctx.post(LOGIN_PATH, {
    form: { [LOGIN_USER_FIELD]: loginName, [LOGIN_PASS_FIELD]: password },
  });
}

/** The JSESSIONID cookie currently held by a context, or undefined. */
async function sessionCookie(ctx: APIRequestContext) {
  const { cookies } = await ctx.storageState();
  return cookies.find((c) => c.name === "JSESSIONID");
}

/** Assert a login failed with exactly Java's apiCall JSON shape and key. */
async function expectLoginError(res: APIResponse, key: string, label: string) {
  expect(res.status(), label + " status").toBe(401);
  expect(
    (res.headers()["content-type"] ?? "").toLowerCase(),
    label + " content-type",
  ).toContain("application/json");
  const body = await res.json();
  // Exactly one key: a port must not leak which of user/password was wrong via
  // an extra field.
  expect(Object.keys(body), label + " body keys").toEqual(["error"]);
  expect(body.error, label + " error key").toBe(key);
}

// ───────────────────────────────────────────────────────────────────────────
// 1. Anonymous bootstrap — GET /session before any login
// ───────────────────────────────────────────────────────────────────────────
test.describe("p0-auth: anonymous session bootstrap", () => {
  test("GET /session is reachable without auth and reports authenticated:false", async () => {
    const ctx = await anonCtx();
    const res = await ctx.get(SESSION_PATH);
    const body = await readJson(res, "anon /session");

    // `authenticated` is `userSessionData != null` (UserModuleServiceImpl
    // .isSessionExpired), not "has a SecurityContext".
    expect(body.authenticated, "anon authenticated").toBe(false);
    // sessionId is emitted even when unauthenticated — verified live on Java.
    expectNonEmptyString(body.sessionId, "anon sessionId");

    // Jackson is configured Include.NON_NULL globally (AppConfig:182), so an
    // anonymous session emits ONLY these two keys. Anything else is a leak:
    // roles/userId/csrf must not be reachable before login.
    expect(Object.keys(body).sort(), "anon /session keys").toEqual([
      "authenticated",
      "sessionId",
    ]);

    await ctx.dispose();
  });

  test("anonymous sessionId matches the JSESSIONID cookie the server issued", async () => {
    const ctx = await anonCtx();
    const body = await readJson(await ctx.get(SESSION_PATH), "anon /session");
    const cookie = await sessionCookie(ctx);

    expect(cookie, "JSESSIONID issued on first touch").toBeTruthy();
    // Java: request.getSession().getId(). Pins that the reported id IS the
    // transport id, not an unrelated handle.
    expect(cookie!.value, "sessionId == JSESSIONID").toBe(body.sessionId);
    // web.xml <cookie-config>: http-only=true, secure=true — unconditional.
    expect(cookie!.httpOnly, "JSESSIONID HttpOnly").toBe(true);
    expect(cookie!.secure, "JSESSIONID Secure").toBe(true);

    await ctx.dispose();
  });
});

// ───────────────────────────────────────────────────────────────────────────
// 2. Login success — POST /ValidateLogin?apiCall=true
// ───────────────────────────────────────────────────────────────────────────
test.describe("p0-auth: login success contract", () => {
  const SUCCESS_CASES = [
    { label: "admin", user: ADMIN_USER, pass: ADMIN_PASS },
    { label: "non-admin with one role", user: E2E_USERS.reception, pass: E2E_PASS },
    { label: "non-admin with no roles", user: E2E_USERS.noRoles, pass: E2E_PASS },
  ];

  for (const c of SUCCESS_CASES) {
    test(`valid credentials (${c.label}) return exactly {"success":true}`, async () => {
      const ctx = await anonCtx();
      const res = await loginWith(ctx, c.user, c.pass);

      expect(res.status(), "login status").toBe(200);
      expect(
        (res.headers()["content-type"] ?? "").toLowerCase(),
        "login content-type",
      ).toContain("application/json");
      // CustomFormAuthenticationSuccessHandler.handleApiLogin writes
      // `new JSONObject().put("success", true)` — nothing else, no redirect.
      expect(JSON.parse(await res.text()), "login body").toEqual({
        success: true,
      });

      await ctx.dispose();
    });
  }

  test("login rotates the session id (sessionFixation().migrateSession())", async () => {
    const ctx = await anonCtx();
    const before = (await (await ctx.get(SESSION_PATH)).json()).sessionId;
    expectNonEmptyString(before, "pre-login sessionId");

    await loginWith(ctx, E2E_USERS.reception, E2E_PASS);

    const after = (await (await ctx.get(SESSION_PATH)).json()).sessionId;
    expectNonEmptyString(after, "post-login sessionId");
    expect(after, "session id must change on login").not.toBe(before);
    // …and the cookie must have been re-issued to match, or the client keeps
    // presenting a dead id.
    expect((await sessionCookie(ctx))!.value, "cookie follows rotation").toBe(
      after,
    );

    await ctx.dispose();
  });
});

// ───────────────────────────────────────────────────────────────────────────
// 3. Login failure matrix — five distinct keys, and the CHECK ORDER
// ───────────────────────────────────────────────────────────────────────────
test.describe("p0-auth: login failure matrix", () => {
  test("unknown user and wrong password are indistinguishable", async () => {
    const a = await anonCtx();
    const unknown = await loginWith(a, E2E_UNKNOWN_USER, E2E_PASS);
    await expectLoginError(unknown, "error.invalidcredentials", "unknown user");
    const unknownBody = await unknown.json();
    await a.dispose();

    const b = await anonCtx();
    const wrongPw = await loginWith(b, E2E_USERS.reception, E2E_WRONG_PASS);
    await expectLoginError(
      wrongPw,
      "error.invalidcredentials",
      "wrong password",
    );
    const wrongPwBody = await wrongPw.json();
    await b.dispose();

    // The whole point: different keys here would introduce user enumeration.
    // Compare the full bodies, not just the key.
    expect(unknownBody, "unknown-user vs wrong-password body").toEqual(
      wrongPwBody,
    );
  });

  test("account_locked='Y' yields error.lockedCredentials", async () => {
    const ctx = await anonCtx();
    await expectLoginError(
      await loginWith(ctx, E2E_USERS.locked, E2E_PASS),
      "error.lockedCredentials",
      "locked",
    );
    await ctx.dispose();
  });

  test("account_disabled='Y' yields error.disabledCredentials", async () => {
    const ctx = await anonCtx();
    await expectLoginError(
      await loginWith(ctx, E2E_USERS.disabled, E2E_PASS),
      "error.disabledCredentials",
      "disabled",
    );
    await ctx.dispose();
  });

  test("past password_expired_dt yields error.expiredCredentials", async () => {
    const ctx = await anonCtx();
    await expectLoginError(
      await loginWith(ctx, E2E_USERS.expired, E2E_PASS),
      "error.expiredCredentials",
      "expired",
    );
    await ctx.dispose();
  });

  // The two tests below pin Spring's check ORDER, which is what a port is most
  // likely to get wrong — by verifying the password first and reporting "bad
  // credentials" for a locked account, or the reverse for an expired one.
  //
  // AbstractUserDetailsAuthenticationProvider runs, in this order:
  //   preAuthenticationChecks        (locked → disabled → account-expired)
  //   additionalAuthenticationChecks (the bcrypt comparison)
  //   postAuthenticationChecks       (credentials-expired)
  test("locked beats a wrong password (account state is checked FIRST)", async () => {
    const ctx = await anonCtx();
    await expectLoginError(
      await loginWith(ctx, E2E_USERS.locked, E2E_WRONG_PASS),
      "error.lockedCredentials",
      "locked + wrong password",
    );
    await ctx.dispose();
  });

  test("a wrong password beats expired credentials (checked LAST)", async () => {
    const ctx = await anonCtx();
    await expectLoginError(
      await loginWith(ctx, E2E_USERS.expired, E2E_WRONG_PASS),
      "error.invalidcredentials",
      "expired + wrong password",
    );
    await ctx.dispose();
  });

  test("a failed login leaves the session unauthenticated", async () => {
    const ctx = await anonCtx();
    await loginWith(ctx, E2E_USERS.reception, E2E_WRONG_PASS);
    const body = await readJson(
      await ctx.get(SESSION_PATH),
      "/session after failed login",
    );
    expect(body.authenticated, "still anonymous").toBe(false);
    expect(Object.keys(body).sort(), "no identity leaked").toEqual([
      "authenticated",
      "sessionId",
    ]);
    await ctx.dispose();
  });
});

// ───────────────────────────────────────────────────────────────────────────
// 4. Authenticated /session — identity, roles, and the padded-role trap
// ───────────────────────────────────────────────────────────────────────────
test.describe("p0-auth: authenticated session contract", () => {
  test("reports the full bootstrap DTO for a non-admin user", async () => {
    const ctx = await anonCtx();
    await loginWith(ctx, E2E_USERS.reception, E2E_PASS);
    const s = await readJson(await ctx.get(SESSION_PATH), "authed /session");

    expect(s.authenticated).toBe(true);
    expect(s.loginMethod, "form login").toBe("FORM");
    expect(s.loginName).toBe(E2E_USERS.reception);
    expect(s.firstName).toBe("E2E");
    expect(s.lastName).toBe("Reception");
    expectNonEmptyString(s.sessionId, "authed sessionId");
    expectNonEmptyString(s.csrf, "authed csrf");

    // EXACT key set, verified live against Java for a user with no lab-unit
    // roles. `userLabRolesMap` and `loginLabUnit` are absent — Jackson's
    // Include.NON_NULL drops them — so a port that emits them as null, or
    // invents a field, fails here rather than drifting.
    expect(Object.keys(s).sort(), "authed /session keys").toEqual(
      [
        "authenticated",
        "csrf",
        "firstName",
        "lastName",
        "loginMethod",
        "loginName",
        "roles",
        "sessionId",
        "userId",
      ].sort(),
    );

    // userId is system_user.id resolved by login_name with is_active='Y' —
    // a STRING join, not a FK (LoginUserDAOImpl.getSystemUserId).
    const [[expectedId]] = query(
      "SELECT id FROM clinlims.system_user WHERE login_name = 'e2e_reception' AND is_active = 'Y';",
    );
    expect(String(s.userId), "userId from the active system_user").toBe(
      String(expectedId),
    );

    await ctx.dispose();
  });

  test("roles are TRIMMED even though system_role.name is blank-padded", async () => {
    // Prove the column really is padded, so the trim assertion below is a real
    // test and not a coincidence. system_role.name is character(30): a Go DAO
    // scanning it without trimming gets "Reception" + 21 spaces, and every
    // `role == "Reception"` comparison then silently fails.
    //
    // The padding is asserted via octet_length rather than by matching trailing
    // spaces in the value: psql's own output and the query() helper both trim
    // the tail, which is exactly why this trap survives casual inspection.
    const [[storedLen, trimmedLen, trimmedName]] = query(
      "SELECT octet_length(name), length(trim(name)), trim(name)" +
        " FROM clinlims.system_role WHERE id = 4;",
    );
    expect(trimmedName, "role 4 is Reception").toBe("Reception");
    expect(Number(storedLen), "system_role.name is character(30)").toBe(30);
    expect(
      Number(storedLen) - Number(trimmedLen),
      "the stored value carries blank padding",
    ).toBeGreaterThan(0);

    const ctx = await anonCtx();
    await loginWith(ctx, E2E_USERS.reception, E2E_PASS);
    const s = await readJson(await ctx.get(SESSION_PATH), "authed /session");

    // Exactly one role, exactly trimmed. Not `toContain` — an extra role would
    // mean the grant query is wrong.
    expect(s.roles, "roles for a single-role user").toEqual(["Reception"]);
    for (const r of s.roles)
      expect(r, `role "${r}" has no padding`).toBe(r.trim());

    await ctx.dispose();
  });

  test("roles match the system_user_role grants in the database", async () => {
    const ctx = await anonCtx();
    await loginWith(ctx, ADMIN_USER, ADMIN_PASS);
    const s = await readJson(await ctx.get(SESSION_PATH), "admin /session");

    const expected = query(
      "SELECT trim(r.name) FROM clinlims.system_user_role sur" +
        " JOIN clinlims.system_role r ON r.id = sur.role_id" +
        " WHERE sur.system_user_id = (SELECT id FROM clinlims.system_user" +
        ` WHERE login_name = '${ADMIN_USER}' AND is_active = 'Y')` +
        " ORDER BY 1;",
    ).map((row) => row[0]);

    expect(expected.length, "DB oracle returned grants").toBeGreaterThan(0);
    expect([...s.roles].sort(), "roles == DB grants").toEqual(expected);

    await ctx.dispose();
  });

  test("a user with zero role grants gets an empty roles list, not a failure", async () => {
    const ctx = await anonCtx();
    await loginWith(ctx, E2E_USERS.noRoles, E2E_PASS);
    const s = await readJson(await ctx.get(SESSION_PATH), "no-roles /session");

    expect(s.authenticated).toBe(true);
    expect(s.loginName).toBe(E2E_USERS.noRoles);
    expect(s.roles, "no grants → empty roles").toEqual([]);

    // Cross-check the premise: this user really has no grants.
    const grants = query(
      "SELECT count(*) FROM clinlims.system_user_role WHERE system_user_id =" +
        " (SELECT id FROM clinlims.system_user WHERE login_name = 'e2e_noroles');",
    );
    expect(grants[0][0], "fixture user has no role grants").toBe("0");

    await ctx.dispose();
  });
});

// ───────────────────────────────────────────────────────────────────────────
// 5. CSRF — Spring's XOR masking (auth-adoption-plan.md §2.6 / §6.2)
// ───────────────────────────────────────────────────────────────────────────
test.describe("p0-auth: csrf token masking", () => {
  test("csrf differs on every read but un-masks to one stable token", async () => {
    const ctx = await anonCtx();
    await loginWith(ctx, E2E_USERS.reception, E2E_PASS);

    const first = (await (await ctx.get(SESSION_PATH)).json()).csrf;
    const second = (await (await ctx.get(SESSION_PATH)).json()).csrf;
    expectNonEmptyString(first, "csrf read 1");
    expectNonEmptyString(second, "csrf read 2");

    // XorCsrfTokenRequestAttributeHandler re-masks with fresh random bytes on
    // every read. A port that returns a stored token verbatim fails here.
    expect(second, "csrf is re-masked per read").not.toBe(first);

    // …and a port that just returns fresh RANDOM garbage fails here: both
    // values must un-mask to the SAME underlying session token.
    const a = unmaskCsrf(first);
    const b = unmaskCsrf(second);
    expect(a, "csrf read 1 un-masks").not.toBeNull();
    expect(b, "csrf read 2 un-masks").not.toBeNull();
    expect(b, "both masks carry the same token").toBe(a);
    expect(a!.length, "un-masked token is non-empty").toBeGreaterThan(0);

    await ctx.dispose();
  });

  test("a different session gets a different underlying csrf token", async () => {
    const one = await anonCtx();
    await loginWith(one, E2E_USERS.reception, E2E_PASS);
    const tokenOne = unmaskCsrf(
      (await (await one.get(SESSION_PATH)).json()).csrf,
    );
    await one.dispose();

    const two = await anonCtx();
    await loginWith(two, E2E_USERS.reception, E2E_PASS);
    const tokenTwo = unmaskCsrf(
      (await (await two.get(SESSION_PATH)).json()).csrf,
    );
    await two.dispose();

    // Per-session, not per-user and not global.
    expect(tokenTwo, "csrf token is per-session").not.toBe(tokenOne);
  });
});

// ───────────────────────────────────────────────────────────────────────────
// 6. The auth boundary — default-deny, and the public whitelist
// ───────────────────────────────────────────────────────────────────────────
//
// Java's Spring Security entry point answers an unauthenticated protected
// request with 302 → /LoginPage. The suite asserts the RAW response
// (maxRedirects: 0) rather than the followed one: following the redirect lands
// on Tomcat's login JSP, which a Go port has no reason to serve, so "the body
// looks like the login HTML" is not a language-neutral contract — "302 to
// /LoginPage, carrying no data" is.
test.describe("p0-auth: default-deny boundary", () => {
  // Endpoints already ported to Go (a1/a2/b1/b2) that are NOT in Java's
  // OPEN_PAGES and therefore require authentication. These are the tests that
  // actually catch "the Go service serves real data to anyone".
  const PROTECTED = [
    "rest/server-time", // a1
    "rest/supportedlocales", // a2 — note: NOT the /active variant
    "rest/dictionary-categories", // b1
    "rest/organization-list", // b2
  ];

  for (const path of PROTECTED) {
    test(`anonymous GET ${path} is refused and returns no data`, async () => {
      const ctx = await anonCtx();
      const res = await ctx.get(path, { maxRedirects: 0 });

      expect(res.status(), `anon ${path} status`).toBe(302);
      expect(
        res.headers()["location"] ?? "",
        `anon ${path} redirect target`,
      ).toMatch(/\/LoginPage$/);

      // Java sends Content-Length: 0 on this redirect — measured. Asserting
      // "bodiless" rather than "doesn't look like JSON" matters: an earlier
      // version of this check only rejected a JSON-shaped body, and so did not
      // notice that Go's http.Redirect helper appends its own
      // `<a href="…">Found</a>.` HTML. The stricter c1 PHI-boundary assertion
      // caught it; this one is tightened to match so the gap cannot come back
      // on a non-PHI endpoint.
      const body = await res.text();
      expect(body, `anon ${path} redirect is bodiless`).toBe("");

      await ctx.dispose();
    });

    test(`authenticated GET ${path} returns JSON`, async () => {
      const ctx = await anonCtx();
      await loginWith(ctx, ADMIN_USER, ADMIN_PASS);
      const res = await ctx.get(path);
      expect(res.status(), `authed ${path} status`).toBe(200);
      expect(
        (res.headers()["content-type"] ?? "").toLowerCase(),
        `authed ${path} content-type`,
      ).toContain("application/json");
      await ctx.dispose();
    });
  }

  // OPEN_PAGES (SecurityConfig:105-107). `/rest/supportedlocales/active` is on
  // the list; its siblings `/rest/supportedlocales` and `.../fallback`
  // deliberately are not — asserted together with the PROTECTED list above,
  // the pair proves the whitelist matches the EXACT path and does not leak by
  // prefix. That specific over-exposure (Go served all three anonymously) is
  // what P0 closed.
  async function expectPublicJSON(path: string) {
    const ctx = await anonCtx();
    const res = await ctx.get(path, { maxRedirects: 0 });
    expect(res.status(), `public ${path} status`).toBe(200);
    expect(
      (res.headers()["content-type"] ?? "").toLowerCase(),
      `public ${path} content-type`,
    ).toContain("application/json");
    const body = await res.json();
    // Real content, not an empty husk — a whitelist that returns [] would pass
    // a status-only check.
    expect(
      Array.isArray(body) ? body.length : Object.keys(body).length,
      `public ${path} is non-empty`,
    ).toBeGreaterThan(0);
    await ctx.dispose();
  }

  test("rest/supportedlocales/active stays anonymous", async () => {
    await expectPublicJSON("rest/supportedlocales/active");
  });

  test("rest/open-configuration-properties stays anonymous", async () => {
    // COVERAGE GAP, stated rather than hidden: this endpoint is in Java's
    // OPEN_PAGES but is not ported yet — branch-naming.md defers
    // open-configuration-properties and configuration-properties to their own
    // config branch. Against Go it 404s because the route does not exist, which
    // says nothing about auth. Delete this skip when that branch lands.
    test.skip(
      test.info().project.name === "go-parity",
      "rest/open-configuration-properties is not ported yet (deferred to the config branch)",
    );
    await expectPublicJSON("rest/open-configuration-properties");
  });
});

// ───────────────────────────────────────────────────────────────────────────
// 7. Logout — which is also the CSRF ENFORCEMENT test
// ───────────────────────────────────────────────────────────────────────────
//
// POST /Logout is the only state-changing verb on the ported surface, so it is
// where CSRF enforcement is observable. Verified live on Java:
//   * no / wrong token → 302 → /Home?access=denied, session SURVIVES
//     (the accessDeniedHandler's non-/rest branch — /Logout is not under /rest,
//     /Provider or /api/OpenELIS-Global/rest, so it redirects instead of
//     emitting the 403 JSON those paths get)
//   * valid token      → 302 → /LoginPage, session invalidated
test.describe("p0-auth: logout and csrf enforcement", () => {
  /** Log in and return [ctx, maskedCsrf]. */
  async function authedCtxWithCsrf(): Promise<[APIRequestContext, string]> {
    const ctx = await anonCtx();
    await loginWith(ctx, E2E_USERS.reception, E2E_PASS);
    const s = await (await ctx.get(SESSION_PATH)).json();
    expect(s.authenticated, "authenticated before logout").toBe(true);
    expectNonEmptyString(s.csrf, "csrf for logout");
    return [ctx, s.csrf];
  }

  test("POST Logout without a csrf token is refused and the session survives", async () => {
    const [ctx] = await authedCtxWithCsrf();

    const res = await ctx.post("Logout", { maxRedirects: 0 });
    expect(res.status(), "csrf-less logout status").toBe(302);
    expect(
      res.headers()["location"] ?? "",
      "csrf-less logout redirect target",
    ).toMatch(/\/Home\?access=denied$/);

    // The point of the test: the request must have had NO effect.
    expect(
      (await (await ctx.get(SESSION_PATH)).json()).authenticated,
      "session survives a csrf-less logout",
    ).toBe(true);

    await ctx.dispose();
  });

  test("POST Logout with a forged csrf token is refused", async () => {
    const [ctx, csrf] = await authedCtxWithCsrf();

    // A WELL-FORMED mask of a token that simply is not this session's. Built
    // rather than corrupted, and the distinction matters:
    //
    // Corrupting a character of the real masked value leaves the server
    // un-masking to an arbitrary byte, which Spring then runs through
    // Utf8.decode — so whether Java answers with its access-denied redirect
    // (302) or throws and answers 500 depends on whether that random byte
    // happens to be valid UTF-8. Measured live at roughly 60/40. That is a coin
    // toss, not a contract, and an earlier version of this test asserted one
    // side of it and flaked.
    //
    // Forging a valid mask of a different ASCII token removes the
    // nondeterminism AND tests the stronger property: the server must actually
    // un-mask and COMPARE. A port that checks only that the header is present,
    // or that compares the masked strings directly, passes the previous test
    // and fails this one.
    const real = unmaskCsrf(csrf);
    expect(real, "the session's csrf un-masks").not.toBeNull();
    const forged = maskCsrf(differentAsciiToken(real!));
    expect(unmaskCsrf(forged), "forged token carries a different value").not.toBe(
      real,
    );

    const res = await ctx.post("Logout", {
      maxRedirects: 0,
      headers: { "X-CSRF-TOKEN": forged },
    });
    expect(res.status(), "forged-csrf logout status").toBe(302);
    expect(
      res.headers()["location"] ?? "",
      "forged-csrf logout redirect target",
    ).toMatch(/\/Home\?access=denied$/);
    expect(
      (await (await ctx.get(SESSION_PATH)).json()).authenticated,
      "session survives a forged-csrf logout",
    ).toBe(true);

    await ctx.dispose();
  });

  test("POST Logout with a valid csrf token invalidates the session", async () => {
    const [ctx, csrf] = await authedCtxWithCsrf();

    // The token is sent MASKED, exactly as /session handed it over. The server
    // must un-mask before comparing — see fixtures/csrf.ts.
    const res = await ctx.post("Logout", {
      maxRedirects: 0,
      headers: { "X-CSRF-TOKEN": csrf },
    });
    expect(res.status(), "logout status").toBe(302);
    expect(res.headers()["location"] ?? "", "logout redirect target").toMatch(
      /\/LoginPage$/,
    );

    // The client still holds the now-dead JSESSIONID. Presenting it does NOT
    // yield `{"authenticated":false}` — sessionManagement().invalidSessionUrl
    // ("/LoginPage") intercepts first, so even /session (a permitAll path)
    // redirects. Pinning the redirect, not a JSON body, is the real contract.
    const session = await ctx.get(SESSION_PATH, { maxRedirects: 0 });
    expect(session.status(), "dead session id on /session").toBe(302);
    expect(
      session.headers()["location"] ?? "",
      "invalidSessionUrl target",
    ).toMatch(/\/LoginPage$/);

    const data = await ctx.get("rest/organization-list", { maxRedirects: 0 });
    expect(data.status(), "protected data re-blocked after logout").toBe(302);
    expect(
      data.headers()["location"] ?? "",
      "protected redirect target after logout",
    ).toMatch(/\/LoginPage$/);

    await ctx.dispose();
  });

  test("a client that drops the dead cookie is simply anonymous again", async () => {
    // Complements the test above: without the stale id there is nothing to
    // invalidate, so /session answers normally.
    const ctx = await anonCtx();
    const body = await readJson(await ctx.get(SESSION_PATH), "fresh /session");
    expect(body.authenticated).toBe(false);
    await ctx.dispose();
  });
});
