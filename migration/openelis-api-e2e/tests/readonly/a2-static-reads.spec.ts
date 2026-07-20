// a2 — migration unit: static + first single-table DB reads (Type A). 5 endpoints.
//
// Design: each test asserts the endpoint's REAL contract, not "HTTP 200 + not the
// login page". Asserting the actual JSON shape/values already proves you weren't
// bounced to the login HTML (it has none of these shapes), so the old not-login
// check is dropped as dead weight. A pass here means the endpoint genuinely
// behaves — the same assertions are the Go parity gate (run via --project=go-parity
// once ported). Expected values below are the frozen Java baseline (captured live).
// See a2-static-reads-migration.md.
//
// Retarget at a Go port:  OE_BASE_URL=https://go-host/... npm test -- a2-static-reads
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
import { test, expect } from "@playwright/test";
import { readJson } from "../../fixtures/assert";

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

test.describe("a2 — static + locale reads (Type A)", () => {
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
