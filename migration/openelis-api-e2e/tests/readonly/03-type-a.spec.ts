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

// ── a2 — static + first single-table DB reads ────────────────────────────────
// Scope: migration/a2-static-reads (5 endpoints). See a2-static-reads-migration.md.
//
// Design: each test asserts the endpoint's REAL contract, not "HTTP 200 + not the
// login page". Asserting the actual JSON shape/values already proves you weren't
// bounced to the login HTML (it has none of these shapes), so the old not-login
// check is dropped as dead weight. A pass here means the endpoint genuinely
// behaves — the same assertions are the Go parity gate (run via --project=go-parity
// once ported). Expected values below are the frozen Java baseline (captured live).
//
//  1. math-functions            → exact 14-operator catalog (compiled-in constant)
//  2. sample-item-status-types  → exact 3-item list (compiled-in constant)
//  3. supportedlocales          → full supported_locale list, exact DTO shape
//  4. supportedlocales/active   → active rows only, ascending sortOrder
//  5. supportedlocales/fallback → the single fallback locale (object, not array) or 404
//
// The three supportedlocales views are cross-checked against each other, so the
// filtering (active), selection (fallback) and per-row serialization must all
// agree — a seed-independent behavioral guarantee, plus strict type checks that
// catch serialization drift (e.g. Go emitting id as a number instead of "1").

const MATH = "rest/math-functions";
const SAMPLE_ITEM_STATUS = "rest/sample-item-status-types";
const LOCALES = "rest/supportedlocales"; // NOTE: no trailing slash — Spring 6 404s "…/"
const LOCALES_ACTIVE = "rest/supportedlocales/active";
const LOCALES_FALLBACK = "rest/supportedlocales/fallback";

// IdValuePair endpoints serialize as [{ id, value }]. These lists are hardcoded
// in Java (Operation.mathFunctions() / DisplayListController), so the port must
// reproduce them byte-for-byte — exact equality is the right, non-brittle check.
const MATH_FUNCTIONS = [
  { id: "+", value: "Plus" },
  { id: "-", value: "Minus" },
  { id: "/", value: "Divided By" },
  { id: "*", value: "Multiplied By" },
  { id: "(", value: "Open Bracket" },
  { id: ")", value: "Close Bracket" },
  { id: "==", value: "Equals" },
  { id: "!=", value: "Does Not Equal" },
  { id: ">=", value: "Is Greater Than Or Equal" },
  { id: "<=", value: "Is Less Than Or Equal" },
  { id: "IS_IN_NORMAL_RANGE", value: "Is With In Normal Range" },
  { id: "IS_OUTSIDE_NORMAL_RANGE", value: "Is Out Side Normal Range" },
  { id: "&&", value: "And" },
  { id: "||", value: "Or" },
];

const SAMPLE_ITEM_STATUS_TYPES = [
  { id: "", value: "All" },
  { id: "active", value: "Active" },
  { id: "disposed", value: "Disposed" },
];

// The exact key set of a SupportedLocaleDTO (sorted for comparison).
const LOCALE_KEYS = [
  "active",
  "displayName",
  "fallback",
  "id",
  "localeCode",
  "sortOrder",
];

async function readJson(res: import("@playwright/test").APIResponse, label: string) {
  expect(res.status(), `${label} status`).toBe(200);
  expect(
    (res.headers()["content-type"] ?? "").toLowerCase(),
    `${label} content-type`,
  ).toContain("application/json");
  return res.json();
}

// A SupportedLocaleDTO row must have EXACTLY these six keys with these types.
// The type checks are the serialization contract: id is a string ("1", not 1),
// active/fallback are real booleans, sortOrder is an integer. A Go port that
// diverges on any of these fails here instead of passing silently.
function assertLocaleShape(row: any, label: string) {
  expect(Object.keys(row).sort(), `${label} keys`).toEqual(LOCALE_KEYS);
  expect(typeof row.id, `${label} id is string`).toBe("string");
  expect(row.id.length, `${label} id non-empty`).toBeGreaterThan(0);
  expect(typeof row.localeCode, `${label} localeCode is string`).toBe("string");
  expect(row.localeCode.length, `${label} localeCode non-empty`).toBeGreaterThan(0);
  expect(typeof row.displayName, `${label} displayName is string`).toBe("string");
  expect(typeof row.active, `${label} active is boolean`).toBe("boolean");
  expect(typeof row.fallback, `${label} fallback is boolean`).toBe("boolean");
  expect(Number.isInteger(row.sortOrder), `${label} sortOrder is integer`).toBe(true);
}

test.describe("Type A — static reads (a2)", () => {
  test("math-functions returns the exact 14-operator catalog", async ({ request }) => {
    const body = await readJson(await request.get(MATH), MATH);
    // Deep equality incl. order — the compiled-in operator table is a frozen
    // contract; any added/removed/renamed operator or reorder fails.
    expect(body, `${MATH} exact contract`).toEqual(MATH_FUNCTIONS);
  });

  test("sample-item-status-types returns the exact 3-item list", async ({ request }) => {
    const body = await readJson(await request.get(SAMPLE_ITEM_STATUS), SAMPLE_ITEM_STATUS);
    expect(body, `${SAMPLE_ITEM_STATUS} exact contract`).toEqual(SAMPLE_ITEM_STATUS_TYPES);
  });

  test("supportedlocales returns the full locale list with the exact DTO shape", async ({ request }) => {
    const body = await readJson(await request.get(LOCALES), LOCALES);
    expect(Array.isArray(body), `${LOCALES} is an array`).toBe(true);
    expect(body.length, `${LOCALES} non-empty`).toBeGreaterThan(0);
    for (const row of body) assertLocaleShape(row, `${LOCALES} row`);

    // id and localeCode are identifiers — must be unique across the list.
    const ids = body.map((r: any) => r.id);
    const codes = body.map((r: any) => r.localeCode);
    expect(new Set(ids).size, `${LOCALES} id unique`).toBe(ids.length);
    expect(new Set(codes).size, `${LOCALES} localeCode unique`).toBe(codes.length);
  });

  test("supportedlocales/active = the active rows of the full list, ascending sortOrder", async ({ request }) => {
    const base = await readJson(await request.get(LOCALES), LOCALES);
    const active = await readJson(await request.get(LOCALES_ACTIVE), LOCALES_ACTIVE);
    expect(Array.isArray(active), `${LOCALES_ACTIVE} is an array`).toBe(true);

    const byId = new Map(base.map((r: any) => [r.id, r]));

    for (const row of active) {
      assertLocaleShape(row, `${LOCALES_ACTIVE} row`);
      // The filtering contract: /active must contain ONLY active=true rows.
      expect(row.active, `${LOCALES_ACTIVE} contains only active=true rows`).toBe(true);
      // Cross-view serialization parity: the same locale must serialize
      // identically in /active and in the full list.
      expect(row, `${LOCALES_ACTIVE} row ${row.id} matches full-list row`).toEqual(
        byId.get(row.id),
      );
    }

    // /active is EXACTLY the active subset of the full list (same id set).
    const baseActiveIds = base.filter((r: any) => r.active).map((r: any) => r.id).sort();
    expect(active.map((r: any) => r.id).sort(), `${LOCALES_ACTIVE} = active subset of full list`).toEqual(
      baseActiveIds,
    );

    // The ordering contract: ORDER BY sort_order ASC → non-decreasing sortOrder.
    const orders = active.map((r: any) => r.sortOrder);
    for (let i = 1; i < orders.length; i++) {
      expect(
        orders[i] >= orders[i - 1],
        `${LOCALES_ACTIVE} sortOrder ascending (index ${i}: ${orders[i - 1]} → ${orders[i]})`,
      ).toBe(true);
    }
  });

  test("supportedlocales/fallback returns the single fallback locale (object, not array)", async ({ request }) => {
    const base = await readJson(await request.get(LOCALES), LOCALES);
    const baseFallback = base.filter((r: any) => r.fallback);
    const res = await request.get(LOCALES_FALLBACK);

    // Contract: fallback is a single row selection. If the seed has no
    // fallback=true row, Java returns 404; otherwise a single object (NOT a list).
    if (baseFallback.length === 0) {
      expect(res.status(), `${LOCALES_FALLBACK} 404 when no fallback row`).toBe(404);
      return;
    }

    const fb = await readJson(res, LOCALES_FALLBACK);
    expect(Array.isArray(fb), `${LOCALES_FALLBACK} is a single object, not an array`).toBe(false);
    assertLocaleShape(fb, LOCALES_FALLBACK);
    expect(fb.fallback, `${LOCALES_FALLBACK} .fallback === true`).toBe(true);

    // It must BE one of the full list's fallback rows, serialized identically.
    const match = base.find((r: any) => r.id === fb.id);
    expect(match, `${LOCALES_FALLBACK} row ${fb.id} exists in the full list`).toBeTruthy();
    expect(fb, `${LOCALES_FALLBACK} matches its full-list row`).toEqual(match);
    expect(
      baseFallback.some((r: any) => r.id === fb.id),
      `${LOCALES_FALLBACK} is a fallback=true row`,
    ).toBe(true);
  });
});
