// DB oracle — schema + reference-data inventory. These are Java-baseline
// invariants (post `core`+`harness` fixtures); the Go port must reproduce them
// against the same schema. Widens the DB assertion surface beyond b1-testcatalog.
import { test, expect } from "@playwright/test";
import { count, query } from "../../fixtures/db";

test.describe("DB oracle — schema & reference inventory", () => {
  test("clinlims schema has the expected table count", async () => {
    const n = parseInt(
      query(
        "SELECT count(*) FROM information_schema.tables WHERE table_schema='clinlims'",
      )[0][0],
      10,
    );
    expect(n).toBe(375);
  });

  // reference tables seeded by the DB image / fixtures — stable baselines
  const REF: Record<string, number> = {
    test: 189,
    test_result: 337,
    panel: 4,
    type_of_sample: 15,
    unit_of_measure: 41,
    dictionary: 701,
    dictionary_category: 73,
  };
  for (const [table, expected] of Object.entries(REF)) {
    test(`reference table ${table} = ${expected}`, async () => {
      expect(count(table)).toBe(expected);
    });
  }

  test("core clinical tables are present (structural)", async () => {
    for (const t of [
      "patient",
      "person",
      "sample",
      "sample_human",
      "sample_item",
      "analysis",
      "result",
      "organization",
      "provider",
    ]) {
      // to_regclass is null if the table doesn't exist
      const reg = query(`SELECT to_regclass('clinlims.${t}')`)[0][0];
      expect(reg, `table ${t} should exist`).not.toBe("");
    }
  });
});
