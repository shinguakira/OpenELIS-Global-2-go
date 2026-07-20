// a1 — migration unit: GET rest/server-time (Type A pilot).
// The first endpoint ported end-to-end. This spec captures the Java baseline the
// Go port must reproduce (response shape + IANA timezone). It also runs against
// the Go service via the go-parity project. See a1-server-time-migration.md.
//
// Retarget at a Go port:  OE_BASE_URL=https://go-host/... npm test -- a1-server-time
import { test, expect } from "@playwright/test";
import { isAuthedResponse } from "../../fixtures/contract";

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

test.describe("a1 — server-time (Type A pilot)", () => {
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
