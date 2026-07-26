// a2 — migration unit: static + single-table DB reads (Type A/B). 7 endpoints.
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
//  3. supportedlocales          → exactly the seeded rows (set-equality; getAll is unordered)
//  4. supportedlocales/active   → exactly the active rows, ordered by sortOrder
//  5. supportedlocales/fallback → exactly the fallback row (a single object)
//  6. analysis-status-types     → status_of_sample: DB ids + localized labels
//  7. sample-status-types       → status_of_sample: DB ids + localized labels
//
// supported_locale is stable reference data, so all three assert the EXACT seeded
// rows (values, types, count, and — where the query guarantees it — order). Exact
// equality also catches serialization drift (e.g. id as a number instead of "1").
import { test, expect } from "@playwright/test";
import { readJson } from "../../fixtures/assert";

const MATH = "rest/math-functions";
const SAMPLE_ITEM_STATUS = "rest/sample-item-status-types";
const LOCALES = "rest/supportedlocales"; // NOTE: no trailing slash — Spring 6 404s "…/"
const LOCALES_ACTIVE = "rest/supportedlocales/active";
const LOCALES_FALLBACK = "rest/supportedlocales/fallback";
const ANALYSIS_STATUS = "rest/analysis-status-types";
const SAMPLE_STATUS = "rest/sample-status-types";

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

// analysis-/sample-status-types read the status_of_sample table: id = the DB
// primary key (Java matches the internal status name → row), value = the localized
// (English) label — which differs from the DB `name` column (e.g. "Not Tested" →
// "Not started"). Stable reference data, so pin the exact rows. Captured live.
const ANALYSIS_STATUS_TYPES = [
  { id: "0", value: "" },
  { id: "4", value: "Not started" },
  { id: "14", value: "Canceled" },
  { id: "15", value: "Accepted by technician" },
  { id: "16", value: "Not accepted by technician" },
  { id: "7", value: "Not accepted by biologist" },
];
const SAMPLE_STATUS_TYPES = [
  { id: "0", value: "" },
  { id: "1", value: "No tests have been run for this sample" },
  { id: "2", value: "Some tests have been run on this sample" },
];

// supported_locale is stable reference data (no CRUD in normal operation), so we
// pin the EXACT seeded rows. Deep equality (toEqual) enforces values, the full key
// set, types (id "1" is a string, not 1) and count in one shot. Captured live from
// the Java baseline == the clinlims seed.
const EN = { id: "1", localeCode: "en", displayName: "English", active: true, fallback: true, sortOrder: 1 };
const FR = { id: "2", localeCode: "fr", displayName: "Francais", active: true, fallback: false, sortOrder: 2 };
const LOCALES_ALL = [EN, FR]; // getAll: no ORDER BY → compared order-independently (by id)
const LOCALES_ACTIVE_ORDERED = [EN, FR]; // is_active rows, ORDER BY sort_order ASC
const LOCALE_FALLBACK = EN; // the single is_fallback row

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

  test("analysis-status-types returns the exact status list (DB ids + localized labels)", async ({ request }) => {
    const body = await readJson(await request.get(ANALYSIS_STATUS), ANALYSIS_STATUS);
    expect(body, `${ANALYSIS_STATUS} exact contract`).toEqual(ANALYSIS_STATUS_TYPES);
  });

  test("sample-status-types returns the exact status list (DB ids + localized labels)", async ({ request }) => {
    const body = await readJson(await request.get(SAMPLE_STATUS), SAMPLE_STATUS);
    expect(body, `${SAMPLE_STATUS} exact contract`).toEqual(SAMPLE_STATUS_TYPES);
  });

  test("supportedlocales returns exactly the seeded locales", async ({ request }) => {
    const body = await readJson(await request.get(LOCALES), LOCALES);
    // getAll() has no ORDER BY → order is DB-natural, so compare as a set (by id).
    // Exact values/keys/types/count: catches id-as-number, wrong displayName, an
    // extra/missing row or key, etc.
    const byId = [...body].sort((a: any, b: any) => a.id.localeCompare(b.id));
    expect(byId, `${LOCALES} exact seeded rows`).toEqual(LOCALES_ALL);
  });

  test("supportedlocales/active returns exactly the active locales, ordered by sortOrder", async ({ request }) => {
    const active = await readJson(await request.get(LOCALES_ACTIVE), LOCALES_ACTIVE);
    // WHERE is_active ORDER BY sort_order ASC — order IS guaranteed, so assert the
    // exact ordered list (no inactive row may leak in).
    expect(active, `${LOCALES_ACTIVE} exact ordered`).toEqual(LOCALES_ACTIVE_ORDERED);
  });

  test("supportedlocales/fallback returns exactly the fallback locale as a single object", async ({ request }) => {
    const fb = await readJson(await request.get(LOCALES_FALLBACK), LOCALES_FALLBACK);
    // A single object, NOT an array; exactly the is_fallback row. (The seed always
    // has a fallback, so the 404-when-none code path isn't exercised here.)
    expect(Array.isArray(fb), `${LOCALES_FALLBACK} is a single object, not an array`).toBe(false);
    expect(fb, `${LOCALES_FALLBACK} exact`).toEqual(LOCALE_FALLBACK);
  });
});
