// §0 Cross-cutting: auth boundary, public whitelist, content-type.
//
// Baseline contract (Java, verified live): an unauthenticated protected REST call
// returns the login HTML page with HTTP 200 — not 401/403 JSON. The "blocked"
// convention is centralized in fixtures/contract.ts (OE_UNAUTH_MODE), so this test
// retargets at a Go port that returns 401 JSON via config, not an edit.
import { test, expect, request as apiRequest } from "@playwright/test";
import { BASE_URL } from "../../fixtures/env";
import { isBlocked, isAuthedResponse } from "../../fixtures/contract";

function anon() {
  return apiRequest.newContext({
    baseURL: BASE_URL,
    ignoreHTTPSErrors: true,
    storageState: { cookies: [], origins: [] },
  });
}

test.describe("cross-cutting: auth boundary", () => {
  test("unauthenticated protected endpoint is blocked, no data leaked", async () => {
    const ctx = await anon();
    const s = await (await ctx.get("session")).json();
    expect(s.authenticated, "fresh session is unauthenticated").toBe(false);

    const res = await ctx.get("rest/users"); // admin-gated data endpoint
    const body = await res.text();
    expect(isBlocked(res.status(), body), "unauth request is blocked").toBe(
      true,
    );
    expect(body).not.toContain('"id"'); // no user records leaked
    await ctx.dispose();
  });

  test("public whitelist returns real content without auth", async () => {
    const ctx = await anon();
    const res = await ctx.get("rest/open-configuration-properties");
    expect(res.status()).toBe(200);
    const body = await res.text();
    expect(
      isAuthedResponse(res.status(), body),
      "whitelist is not blocked",
    ).toBe(true);
    await ctx.dispose();
  });

  test("authenticated request returns JSON data", async ({ request }) => {
    const res = await request.get("rest/users");
    expect(res.status()).toBe(200);
    expect(res.headers()["content-type"] || "").toContain("application/json");
    const body = await res.text();
    expect(body.startsWith("["), "authed users is a JSON array").toBe(true);
  });
});
