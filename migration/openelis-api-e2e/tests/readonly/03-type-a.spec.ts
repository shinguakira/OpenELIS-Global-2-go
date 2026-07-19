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

// True when tz is a valid IANA zone id (e.g. "Etc/UTC", "Asia/Tokyo") — the
// format Java's ZoneId.systemDefault().getId() always returns. IANA region ids
// are "Area/Location" (contain "/"); Java also emits the slashless specials
// "UTC"/"GMT". Abbreviations like "JST"/"EST"/"PST" (and Go's "Local") have no
// "/" and are not UTC/GMT, so they are rejected — this is exactly the check that
// catches a Go port emitting an abbreviation instead of an IANA id.
// (Intl.DateTimeFormat is NOT used: this Node's Intl leniently accepts "JST",
// and supportedValuesOf() wrongly rejects the container's "Etc/UTC".)
// See a1-server-time-migration.md §5.
function isValidIanaZone(tz: unknown): boolean {
  if (typeof tz !== "string" || tz.length === 0) return false;
  return tz.includes("/") || tz === "UTC" || tz === "GMT";
}

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

  // Response-contract parity: SystemRestController#getServerTime returns
  // {date: yyyy-MM-dd, time: HH:mm, timezone: <IANA zone id>}. The Go port must
  // reproduce this shape — especially the timezone as an IANA id, not a Go zone
  // abbreviation. This is the assertion that guarantees timezone compatibility.
  test(`${PILOT} shape + IANA timezone (Java/Go parity)`, async ({ request }) => {
    const res = await request.get(PILOT);
    expect(res.status(), `${PILOT} status`).toBe(200);
    expect(
      (res.headers()["content-type"] ?? "").toLowerCase(),
      `${PILOT} content-type`,
    ).toContain("application/json");

    const body = await res.json();
    expect(body, `${PILOT} keys`).toHaveProperty("date");
    expect(body, `${PILOT} keys`).toHaveProperty("time");
    expect(body, `${PILOT} keys`).toHaveProperty("timezone");

    expect(body.date, `${PILOT} date = yyyy-MM-dd`).toMatch(/^\d{4}-\d{2}-\d{2}$/);
    expect(body.time, `${PILOT} time = HH:mm`).toMatch(/^([01]\d|2[0-3]):[0-5]\d$/);

    // The compatibility guarantee: timezone MUST be a valid IANA id (Java's
    // ZoneId.getId()), NOT an abbreviation like "JST". A Go port returning
    // "JST"/"EST"/etc. fails here instead of passing silently.
    expect(
      isValidIanaZone(body.timezone),
      `${PILOT} timezone "${body.timezone}" must be an IANA id (e.g. Etc/UTC), not an abbreviation`,
    ).toBe(true);
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
