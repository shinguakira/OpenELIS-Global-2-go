// §3 Type A — static / computed / read-only config endpoints.
// These are the migration pilot set (branches: migration/a1-pilot-server-time,
// migration/a2-static-reads). This spec captures the JAVA BASELINE behavior each
// of these endpoints must reproduce once ported to Go. No DB joins, no clinical
// data — the safest possible endpoints, used to prove the migration mechanism.
//
// Retarget at a Go port:  OE_BASE_URL=https://go-host/... npm test -- 03-type-a
import { test, expect } from "@playwright/test";
import { isAuthedResponse } from "../../fixtures/contract";

// a1 — the single "first sample migration" endpoint.
const PILOT = "rest/server-time";

// a2 — the rest of Type A (param-less static / config reads).
const TYPE_A = [
  "rest/configuration-properties",
  "rest/open-configuration-properties",
  "rest/math-functions",
  "rest/analysis-status-types",
  "rest/sample-status-types",
  "rest/sample-item-status-types",
  "rest/supportedlocales/",
  "rest/supportedlocales/active",
  "rest/supportedlocales/fallback",
  "rest/menu",
];

test.describe("Type A — pilot (a1)", () => {
  test(`${PILOT} returns 200 and an authed JSON body`, async ({ request }) => {
    const res = await request.get(PILOT);
    expect(res.status(), `${PILOT} status`).toBe(200);
    const body = await res.text();
    expect(
      isAuthedResponse(res.status(), body),
      `${PILOT} should be an authed response, not the login page`,
    ).toBe(true);
    // server-time is the golden the Go port must match in shape.
    expect(body.length, `${PILOT} non-empty`).toBeGreaterThan(0);
  });
});

test.describe("Type A — static reads (a2)", () => {
  for (const path of TYPE_A) {
    test(`${path} returns 200 authed`, async ({ request }) => {
      const res = await request.get(path);
      expect(res.status(), `${path} status`).toBe(200);
      const body = await res.text();
      expect(
        isAuthedResponse(res.status(), body),
        `${path} should be an authed response, not the login page`,
      ).toBe(true);
    });
  }
});
