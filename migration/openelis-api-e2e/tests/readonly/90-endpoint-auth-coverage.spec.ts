// Full-surface auth + reachability coverage, generated over EVERY param-less GET
// endpoint (see tools/extract-endpoints.mjs). Two invariants per endpoint:
//   1. anon  → blocked (auth is enforced on every protected endpoint)
//   2. authed → an app response (endpoint is reachable)
// The "blocked" convention is defined by the contract layer, so this runs
// unchanged against a Go port (set OE_UNAUTH_MODE=status if it returns 401 JSON).
import { test, expect, request as apiRequest } from "@playwright/test";
import { BASE_URL } from "../../fixtures/env";
import { isBlocked } from "../../fixtures/contract";
import { GET_ENDPOINTS } from "../../fixtures/endpoints.generated";

// public whitelist — anon gets real content, not the blocked response
const PUBLIC =
  /open-configuration-properties|site-branding|supportedlocales\/active|^health/;

// GET endpoints that MUTATE (legacy pattern) — never hit these authenticated in a
// read-only sweep. Anon is always safe (auth blocks the action).
const UNSAFE_AUTHED =
  /Delete|Deactivate|Remove|Purge|Cancel|trigger|[Ee]xport|Import|optimize|reset|generate/;

// Endpoints that return the blocked/login shell even when authenticated (they
// require params). Tracked so the divergence is visible, not hidden.
const AUTHED_QUIRKS = new Set(["rest/ProviderMenu", "rest/SearchProviderMenu"]);

function anon() {
  return apiRequest.newContext({
    baseURL: BASE_URL,
    ignoreHTTPSErrors: true,
    storageState: { cookies: [], origins: [] },
  });
}

test.describe("auth boundary — unauthenticated GET is blocked", () => {
  for (const ep of GET_ENDPOINTS) {
    if (PUBLIC.test(ep)) continue;
    test(`anon ${ep}`, async () => {
      const ctx = await anon();
      const res = await ctx.get(ep);
      const body = await res.text();
      await ctx.dispose();
      expect(
        isBlocked(res.status(), body),
        `anon ${ep} should be blocked (auth not enforced?)`,
      ).toBe(true);
    });
  }
});

test.describe("reachability — authenticated GET returns an app response", () => {
  for (const ep of GET_ENDPOINTS) {
    if (UNSAFE_AUTHED.test(ep) || AUTHED_QUIRKS.has(ep)) continue;
    test(`authed ${ep}`, async ({ request }) => {
      const res = await request.get(ep);
      const body = await res.text();
      expect(
        isBlocked(res.status(), body),
        `authed ${ep} was blocked (status ${res.status()}) — auth/session broken`,
      ).toBe(false);
    });
  }
});
