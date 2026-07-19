// §2 — b1: dictionary + test-catalog reference reads (strict parity gate).
//
// Reference-data endpoints (taxonomy Type B, Wave 1). These are the highest-
// leverage reads — everything downstream joins to them — so the port must
// reproduce their exact shape and, where Java guarantees one, their ORDER.
//
// Design (same as 03-type-a): assert the real contract, not "HTTP 200 + not the
// login page". Ordering is asserted ONLY where the Java source guarantees it:
//   - lab-units → options.sort(compareToIgnoreCase)  → case-insensitive name asc
//   - panels    → SQL ORDER BY panelName (active only) → name asc (case-insensitive,
//                 to stay robust to Postgres column collation)
//   - sample-types → ascending numeric sort_order, a hidden column NOT present in
//                    the {id,name} shape → order cannot be verified, NOT asserted
//   - dictionary-categories, uom → getAll(), no ORDER BY → order NOT asserted
// (See the b1 source investigation in the migration notes.)
import { test, expect } from "@playwright/test";
import { count, BASELINE } from "../../fixtures/db";
import {
  readJson,
  readJsonArray,
  expectExactKeys,
  expectKeysWithin,
  expectUnique,
  expectAscending,
  expectNonEmptyString,
} from "../../fixtures/assert";

const DICT_CATEGORIES = "rest/dictionary-categories";
const UOM = "rest/uom";
const LAB_UNITS = "rest/test-catalog/lab-units";
const SAMPLE_TYPES = "rest/test-catalog/sample-types";
const PANELS = "rest/test-catalog/panels";
const TEST_CATALOG = "rest/TestCatalog";

// A reference option row is exactly { id: string, name: string }.
function assertIdName(row: any, label: string): void {
  expectExactKeys(row, ["id", "name"], label);
  expectNonEmptyString(row.id, `${label} id`);
  expectNonEmptyString(row.name, `${label} name`);
}

test.describe("b1 — dictionary + test-catalog reference reads", () => {
  test("dictionary-categories: DictionaryCategory rows with unique ids", async ({ request }) => {
    const body = await readJsonArray(await request.get(DICT_CATEGORIES), DICT_CATEGORIES);
    for (const row of body) {
      // lastupdated is null for some rows and Jackson (NON_NULL) omits it, so it
      // is an allowed-but-optional key; the other four are always present.
      expectKeysWithin(
        row,
        ["lastupdated", "id", "description", "localAbbreviation", "categoryName"],
        ["id", "description", "localAbbreviation", "categoryName"],
        `${DICT_CATEGORIES} row`,
      );
      expectNonEmptyString(row.id, `${DICT_CATEGORIES} id`);
      expect(typeof row.categoryName, `${DICT_CATEGORIES} categoryName`).toBe("string");
      expect(typeof row.description, `${DICT_CATEGORIES} description`).toBe("string");
      if ("lastupdated" in row)
        expect(typeof row.lastupdated, `${DICT_CATEGORIES} lastupdated`).toBe("number");
    }
    expectUnique(body.map((r: any) => r.id), `${DICT_CATEGORIES} id`);
    // order: getAll() with no ORDER BY → DB-natural, NOT asserted.
  });

  test("uom: IdValuePair rows with unique ids", async ({ request }) => {
    const body = await readJsonArray(await request.get(UOM), UOM);
    for (const row of body) {
      expectExactKeys(row, ["id", "value"], `${UOM} row`);
      expectNonEmptyString(row.id, `${UOM} id`);
      expect(typeof row.value, `${UOM} value`).toBe("string");
    }
    expectUnique(body.map((r: any) => r.id), `${UOM} id`);
    // order: getAll() with no ORDER BY → DB-natural, NOT asserted.
  });

  test("test-catalog/lab-units: {id,name}, unique, case-insensitive name order", async ({ request }) => {
    const body = await readJsonArray(await request.get(LAB_UNITS), LAB_UNITS);
    for (const row of body) assertIdName(row, `${LAB_UNITS} row`);
    expectUnique(body.map((r: any) => r.id), `${LAB_UNITS} id`);
    // Java: options.sort((a,b) -> a.name.compareToIgnoreCase(b.name)).
    expectAscending(body.map((r: any) => r.name), `${LAB_UNITS} name`, { caseInsensitive: true });
  });

  test("test-catalog/sample-types: {id,name}, unique ids", async ({ request }) => {
    const body = await readJsonArray(await request.get(SAMPLE_TYPES), SAMPLE_TYPES);
    for (const row of body) assertIdName(row, `${SAMPLE_TYPES} row`);
    expectUnique(body.map((r: any) => r.id), `${SAMPLE_TYPES} id`);
    // order = ascending numeric sort_order (hidden column, not in {id,name}) →
    // cannot be verified from the response, so NOT asserted.
  });

  test("test-catalog/panels: {id,name}, unique, name-ordered", async ({ request }) => {
    const body = await readJsonArray(await request.get(PANELS), PANELS);
    for (const row of body) assertIdName(row, `${PANELS} row`);
    expectUnique(body.map((r: any) => r.id), `${PANELS} id`);
    // SQL: ORDER BY panelName (active panels only).
    expectAscending(body.map((r: any) => r.name), `${PANELS} name`, { caseInsensitive: true });
  });

  test("TestCatalog: form shape with section list + catalog rows", async ({ request }) => {
    const body = await readJson(await request.get(TEST_CATALOG), TEST_CATALOG);
    expect(body.formName, `${TEST_CATALOG} formName`).toBe("testCatalogForm");

    // testSectionList: distinct section-name strings, ascending (case-sensitive,
    // derived from the compound testUnit sort in the controller).
    expect(Array.isArray(body.testSectionList), `${TEST_CATALOG} testSectionList is array`).toBe(true);
    expect(body.testSectionList.length, `${TEST_CATALOG} testSectionList non-empty`).toBeGreaterThan(0);
    for (const s of body.testSectionList) expect(typeof s, `${TEST_CATALOG} section name`).toBe("string");
    expectUnique(body.testSectionList, `${TEST_CATALOG} testSectionList`);
    expectAscending(body.testSectionList, `${TEST_CATALOG} testSectionList`);

    // testCatalogList: non-empty; each row carries a string id and a localization
    // object (the localized test-name graph the frontend renders).
    expect(Array.isArray(body.testCatalogList), `${TEST_CATALOG} testCatalogList is array`).toBe(true);
    expect(body.testCatalogList.length, `${TEST_CATALOG} testCatalogList non-empty`).toBeGreaterThan(0);
    for (const t of body.testCatalogList.slice(0, 50)) {
      expectNonEmptyString(t.id, `${TEST_CATALOG} row id`);
      expect(
        t.localization !== null && typeof t.localization === "object",
        `${TEST_CATALOG} row localization is an object`,
      ).toBe(true);
    }
  });

  test("DB oracle: seeded reference data present", async () => {
    // Java-baseline invariants; the Go port must match these row counts.
    expect(count("test"), "test rows").toBe(BASELINE.test);
    expect(count("dictionary"), "dictionary rows").toBe(BASELINE.dictionary);
    expect(count("type_of_sample"), "sample types").toBe(BASELINE.typeOfSample);
  });
});
