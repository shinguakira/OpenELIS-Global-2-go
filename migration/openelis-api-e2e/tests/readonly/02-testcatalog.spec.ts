// §2 Reference data — Test Catalog reads + DB baseline (the DB oracle).
import { test, expect } from "@playwright/test";
import { count, BASELINE } from "../../fixtures/db";

test.describe("test catalog (reference data)", () => {
  test("TestCatalog endpoint returns 200 JSON", async ({ request }) => {
    const res = await request.get("rest/TestCatalog");
    expect(res.status()).toBe(200);
  });

  test("configuration-properties (display lists) returns 200", async ({
    request,
  }) => {
    const res = await request.get("rest/configuration-properties");
    expect(res.status()).toBe(200);
  });

  test("DB oracle: seeded reference data present", async () => {
    // These are Java-baseline invariants; the Go port must match them.
    expect(count("test"), "test rows").toBe(BASELINE.test);
    expect(count("dictionary"), "dictionary rows").toBe(BASELINE.dictionary);
    expect(count("type_of_sample"), "sample types").toBe(
      BASELINE.typeOfSample,
    );
  });
});
