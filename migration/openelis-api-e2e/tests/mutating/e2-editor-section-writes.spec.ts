/**
 * e2 — the test-catalog editor's section saves: storage, terminology, sample-type
 * display order, and panel membership.
 *
 * None of the four is audited into clinlims.history, even though PANEL,
 * PANEL_ITEM and SAMPLETYPE_TEST all carry keep_history='Y'. Storage keeps an
 * audit of its own instead — a JSON snapshot row written only when the business
 * state actually changed, beside a `version` counter that is bumped on every
 * save whether it changed or not.
 *
 * POST /panels is here too, and it is BROKEN in Java: the insert leaves
 * panel.name_localization_id null and the column is NOT NULL, so every
 * non-blank name is a 500. Recorded in java-defects-found.md and reproduced
 * rather than repaired.
 */
import { test, expect } from "@playwright/test";
import type { APIRequestContext } from "@playwright/test";
import { readJson } from "../../fixtures/assert";
import { query, exec } from "../../fixtures/db";
import {
  SESSION_PATH,
  CSRF_SESSION_FIELD,
  CSRF_HEADER,
} from "../../fixtures/contract";

async function send(
  request: APIRequestContext,
  method: "post" | "put",
  path: string,
  data?: unknown,
) {
  const body = await readJson(await request.get(SESSION_PATH), SESSION_PATH);
  const opts: Record<string, unknown> = {
    headers: {
      [CSRF_HEADER]: body[CSRF_SESSION_FIELD],
      "Content-Type": "application/json",
    },
  };
  if (data !== undefined) opts.data = data;
  return request[method](path, opts);
}

const PREFIX = "E2ED";

/** Everything the fixture create leaves behind, removed in FK order. */
function dropTests() {
  const where = `local_code LIKE '${PREFIX}%'`;
  const ids = `SELECT id FROM clinlims.test WHERE ${where}`;
  const locIds = query(
    `SELECT name_localization_id::text FROM clinlims.test WHERE ${where}
     UNION SELECT reporting_name_localization_id::text FROM clinlims.test WHERE ${where}`,
  )
    .map((r) => r[0])
    .filter((v) => v !== "");
  exec(`DELETE FROM clinlims.test_sample_handling_history WHERE test_sample_handling_id IN
        (SELECT id FROM clinlims.test_sample_handling WHERE test_id IN (${ids}))`);
  exec(`DELETE FROM clinlims.test_sample_handling WHERE test_id IN (${ids})`);
  exec(`UPDATE clinlims.test SET default_test_result_id = NULL WHERE ${where}`);
  exec(`DELETE FROM clinlims.result_limits WHERE test_id IN (${ids})`);
  exec(`DELETE FROM clinlims.test_result WHERE test_id IN (${ids})`);
  exec(`DELETE FROM clinlims.test_result_component WHERE test_id IN (${ids})`);
  exec(`DELETE FROM clinlims.test_terminology_mapping WHERE test_id IN (${ids})`);
  exec(`DELETE FROM clinlims.panel_item WHERE test_id IN (${ids})`);
  exec(`DELETE FROM clinlims.sampletype_test WHERE test_id IN (${ids})`);
  exec(`DELETE FROM clinlims.test WHERE ${where}`);
  for (const lid of locIds) {
    exec(`DELETE FROM clinlims.localization_value WHERE localization_id = ${Number(lid)}`);
    exec(`DELETE FROM clinlims.localization WHERE id = ${Number(lid)}`);
  }
}

/** A numeric test under sample type 1 and panel 1, made through TestAdd. */
function createWad(suffix: string) {
  return {
    testNameEnglish: `${PREFIX}${suffix}`,
    testNameFrench: `${PREFIX}${suffix}Fr`,
    testReportNameEnglish: `${PREFIX}${suffix}Rep`,
    testReportNameFrench: `${PREFIX}${suffix}RepFr`,
    testSection: "36",
    dictionaryReference: "",
    panels: [{ id: "1" }],
    uom: "1",
    loinc: "",
    resultType: "4",
    sampleTypes: [{ typeId: "1", tests: [{ id: "0" }] }],
    active: "Y",
    orderable: "Y",
    notifyResults: "N",
    inLabOnly: "N",
    antimicrobialResistance: "N",
    lowValid: "1",
    highValid: "100",
    lowReportingRange: "2",
    highReportingRange: "90",
    lowCritical: "3",
    highCritical: "80",
    significantDigits: "2",
    resultLimits: [
      { gender: false, highAgeRange: "10", lowNormal: "5", highNormal: "50" },
    ],
  };
}

async function makeTest(request: APIRequestContext, suffix: string) {
  const res = await send(request, "post", "rest/TestAdd", {
    jsonWad: JSON.stringify(createWad(suffix)),
    loinc: "",
  });
  expect(res.status(), "fixture create").toBe(200);
  const [[testId]] = query(
    `SELECT id::text FROM clinlims.test WHERE local_code = '${PREFIX}${suffix}'`,
  );
  return testId;
}

const FULL_STORAGE = {
  storageCondition: "REFRIGERATED",
  storageConditionCustom: "",
  storageDuration: 48,
  storageDurationUnit: "HOURS",
  stabilityNotes: "note",
  protectFromLight: true,
  doNotFreeze: true,
  doNotRefrigerate: false,
  disposalMethod: "INCINERATE",
  disposalTimeframe: 7,
  disposalUnit: "DAYS",
  specialInstructions: "care",
  overrideRestricted: false,
};

test.describe("e2 — the test-catalog editor section saves", () => {
  let restore: (() => void)[] = [];

  test.beforeEach(() => {
    restore = [];
    dropTests();
  });

  test.afterEach(async ({ request }) => {
    dropTests();
    for (const r of restore) r();
    // Resync Java's cached lists — TestAdd is what created the fixture, and its
    // POST refreshes them from the database. See e2-testadd-writes.
    await send(request, "post", "rest/TestAdd", {
      jsonWad: JSON.stringify(createWad("Resync")),
      loinc: "",
    });
    dropTests();
    exec(
      `DELETE FROM clinlims.history WHERE timestamp > now() - interval '3 minutes'`,
    );
  });

  test("storage — a blank section, a save, and a re-save that changes nothing", async ({
    request,
  }) => {
    const testId = await makeTest(request, "Stor");

    // No config yet is NOT a 404: the section renders blank. The document is
    // the test id and four explicitly-false flags, and nothing else — the rest
    // of the fields are absent, not null.
    // The audit watermark: the fixture create left its own history rows, so
    // "storage wrote none" has to be measured from HERE, not from a time window.
    const [[watermark]] = query(`SELECT max(id)::text FROM clinlims.history`);

    const blank = await readJson(
      await request.get(`rest/test-catalog/tests/${testId}/storage`),
      "blank storage",
    );
    expect(blank, "the empty storage document").toEqual({
      testId,
      protectFromLight: false,
      doNotFreeze: false,
      doNotRefrigerate: false,
      overrideRestricted: false,
    });

    const saved = await send(
      request,
      "put",
      `rest/test-catalog/tests/${testId}/storage`,
      FULL_STORAGE,
    );
    expect(saved.status(), "save").toBe(200);
    // storageConditionCustom was sent BLANK and comes back absent — a blank
    // string is stored as NULL, and NULL is not serialised.
    expect(await saved.json(), "the echoed document").toEqual({
      testId,
      storageCondition: "REFRIGERATED",
      storageDuration: 48,
      storageDurationUnit: "HOURS",
      stabilityNotes: "note",
      protectFromLight: true,
      doNotFreeze: true,
      doNotRefrigerate: false,
      disposalMethod: "INCINERATE",
      disposalTimeframe: 7,
      disposalUnit: "DAYS",
      specialInstructions: "care",
      overrideRestricted: false,
    });

    expect(
      query(
        `SELECT COALESCE(storage_condition, ''), COALESCE(storage_condition_custom, ''),
                COALESCE(storage_duration::text, ''), COALESCE(storage_duration_unit, ''),
                COALESCE(stability_notes, ''), protect_from_light::text, do_not_freeze::text,
                do_not_refrigerate::text, COALESCE(disposal_method, ''),
                COALESCE(disposal_timeframe::text, ''), COALESCE(disposal_unit, ''),
                COALESCE(special_instructions, ''), override_restricted::text,
                version::text, is_active
           FROM clinlims.test_sample_handling WHERE test_id = ${Number(testId)}`,
      ),
      "one row, version 1, active",
    ).toEqual([
      [
        "REFRIGERATED", "", "48", "HOURS", "note", "true", "true", "false",
        "INCINERATE", "7", "DAYS", "care", "false", "1", "Y",
      ],
    ]);

    // The snapshot trail: an INSERT row with no previous values.
    const history = query(
      `SELECT change_type, COALESCE(previous_values::text, 'NULL'), new_values::text
         FROM clinlims.test_sample_handling_history
        WHERE test_sample_handling_id IN
              (SELECT id FROM clinlims.test_sample_handling WHERE test_id = ${Number(testId)})
        ORDER BY changed_at`,
    );
    expect(history.length, "one snapshot").toBe(1);
    expect(history[0][0], "change type").toBe("INSERT");
    expect(history[0][1], "no previous state").toBe("NULL");
    expect(
      JSON.parse(history[0][2]),
      "the snapshot carries the thirteen business fields, blanks as null",
    ).toEqual({
      storageCondition: "REFRIGERATED",
      storageConditionCustom: null,
      storageDuration: 48,
      storageDurationUnit: "HOURS",
      stabilityNotes: "note",
      protectFromLight: true,
      doNotFreeze: true,
      doNotRefrigerate: false,
      disposalMethod: "INCINERATE",
      disposalTimeframe: 7,
      disposalUnit: "DAYS",
      specialInstructions: "care",
      overrideRestricted: false,
    });

    // The same body again. version STILL moves — it counts saves, not changes —
    // but no second snapshot is written.
    expect(
      (await send(request, "put", `rest/test-catalog/tests/${testId}/storage`, FULL_STORAGE)).status(),
      "re-save",
    ).toBe(200);
    expect(
      query(
        `SELECT version::text FROM clinlims.test_sample_handling WHERE test_id = ${Number(testId)}`,
      )[0][0],
      "version counts every save",
    ).toBe("2");
    expect(
      query(
        `SELECT count(*)::text FROM clinlims.test_sample_handling_history
          WHERE test_sample_handling_id IN
                (SELECT id FROM clinlims.test_sample_handling WHERE test_id = ${Number(testId)})`,
      )[0][0],
      "but the snapshot trail did not grow",
    ).toBe("1");

    // Nothing reached clinlims.history at any point — and test_sample_handling
    // is not in reference_tables at all, so there is no keep_history flag to
    // consult either way.
    expect(
      query(
        `SELECT count(*)::text FROM clinlims.history WHERE id > ${Number(watermark)}`,
      )[0][0],
      "storage is not audited into history",
    ).toBe("0");

    expect(
      (await send(request, "put", "rest/test-catalog/tests/999999/storage", FULL_STORAGE)).status(),
      "an unknown test is 404",
    ).toBe(404);
  });

  test("group storage — a whole-document replace across several tests", async ({
    request,
  }) => {
    const testId = await makeTest(request, "Grp");
    expect(
      (await send(request, "put", `rest/test-catalog/tests/${testId}/storage`, FULL_STORAGE)).status(),
      "seed",
    ).toBe(200);

    // The guard is on the REQUEST, not the tests.
    expect(
      (await send(request, "put", "rest/test-catalog/group/storage", { testIds: [], storage: null })).status(),
      "an empty request is 422",
    ).toBe(422);
    expect(
      (await send(request, "put", "rest/test-catalog/group/storage", {
        testIds: [testId],
        storage: null,
      })).status(),
      "and so is a missing storage document",
    ).toBe(422);

    // One real id and one imaginary one: the imaginary one is skipped in
    // silence rather than failing the request.
    const res = await send(request, "put", "rest/test-catalog/group/storage", {
      testIds: [testId, "999999"],
      storage: { storageCondition: "FROZEN", protectFromLight: false },
    });
    expect(res.status(), "group save").toBe(200);
    expect(await res.text(), "the response has no body").toBe("");

    // The save is a REPLACE. Every field the group document omitted is now
    // null, including the duration and the notes that were there a moment ago.
    expect(
      query(
        `SELECT COALESCE(storage_condition, ''), COALESCE(storage_duration::text, ''),
                COALESCE(stability_notes, ''), COALESCE(disposal_method, ''),
                protect_from_light::text, do_not_freeze::text, version::text
           FROM clinlims.test_sample_handling WHERE test_id = ${Number(testId)}`,
      ),
      "everything unnamed was cleared",
    ).toEqual([["FROZEN", "", "", "", "false", "false", "2"]]);

    // The state changed, so this one DID leave a snapshot — an UPDATE with the
    // prior values.
    const history = query(
      `SELECT change_type FROM clinlims.test_sample_handling_history
        WHERE test_sample_handling_id IN
              (SELECT id FROM clinlims.test_sample_handling WHERE test_id = ${Number(testId)})
        ORDER BY changed_at`,
    ).map((r) => r[0]);
    expect(history, "insert then update").toEqual(["INSERT", "UPDATE"]);
  });

  test("terminology — an upsert by (source, code), and the legacy loinc column it drives", async ({
    request,
  }) => {
    const testId = await makeTest(request, "Term");

    expect(
      await readJson(
        await request.get(`rest/test-catalog/tests/${testId}/terminology`),
        "empty terminology",
      ),
      "a test with no mappings",
    ).toEqual({ testId, mappings: [] });

    const saved = await send(
      request,
      "put",
      `rest/test-catalog/tests/${testId}/terminology`,
      {
        mappings: [
          { source: "LOINC", code: "12345-6", relationship: "SAME_AS" },
          { source: "SNOMED", code: "999", relationship: "" },
        ],
      },
    );
    expect(saved.status(), "save").toBe(200);
    const savedBody = await saved.json();
    // As a SET, deliberately. getAllMatching has no ORDER BY, so the response
    // carries whatever order the heap scan produced — insertion order on a
    // freshly-appended table, and something else once Postgres reuses a slot a
    // previous run freed. Pinning the sequence here passed for a while and then
    // failed once in a full-suite run; the order is not a contract either stack
    // holds.
    expect(
      savedBody.mappings
        .map((m: any) => [m.source, m.code, m.relationship ?? null])
        .sort(),
      "a blank relationship is stored as NULL and comes back absent",
    ).toEqual(
      [
        ["LOINC", "12345-6", "SAME_AS"],
        ["SNOMED", "999", null],
      ].sort(),
    );

    expect(
      query(
        `SELECT source, code, COALESCE(relationship, ''), is_active
           FROM clinlims.test_terminology_mapping WHERE test_id = ${Number(testId)}
          ORDER BY source, code`,
      ),
      "both rows active",
    ).toEqual([
      ["LOINC", "12345-6", "SAME_AS", "Y"],
      ["SNOMED", "999", "", "Y"],
    ]);
    // The legacy column the rest of the application reads is re-derived from
    // the SAME_AS LOINC mapping.
    expect(
      // The id rides along so the row survives the oracle: a single all-blank
      // column comes back as an empty string, which `query` trims to no rows.
      query(
        `SELECT id::text, COALESCE(loinc, '') FROM clinlims.test WHERE id = ${Number(testId)}`,
      )[0][1],
      "test.loinc follows the mapping",
    ).toBe("12345-6");

    // Drop the LOINC mapping and change the SNOMED one.
    const second = await send(
      request,
      "put",
      `rest/test-catalog/tests/${testId}/terminology`,
      { mappings: [{ source: "SNOMED", code: "999", relationship: "BROADER_THAN" }] },
    );
    expect(second.status(), "second save").toBe(200);
    expect(
      query(
        `SELECT source, code, COALESCE(relationship, ''), is_active
           FROM clinlims.test_terminology_mapping WHERE test_id = ${Number(testId)}
          ORDER BY source, code`,
      ),
      "the dropped mapping is switched off, not deleted",
    ).toEqual([
      ["LOINC", "12345-6", "SAME_AS", "N"],
      ["SNOMED", "999", "BROADER_THAN", "Y"],
    ]);
    // And with no LOINC mapping left, the legacy column is CLEARED.
    expect(
      // The id rides along so the row survives the oracle: a single all-blank
      // column comes back as an empty string, which `query` trims to no rows.
      query(
        `SELECT id::text, COALESCE(loinc, '') FROM clinlims.test WHERE id = ${Number(testId)}`,
      )[0][1],
      "test.loinc was cleared by dropping the mapping",
    ).toBe("");

    // Four rejections, each 422 and each writing nothing.
    for (const [label, body] of [
      ["an unknown source", { mappings: [{ source: "NOPE", code: "1" }] }],
      ["a blank code", { mappings: [{ source: "LOINC", code: "" }] }],
      ["an unknown relationship", { mappings: [{ source: "LOINC", code: "1", relationship: "WAT" }] }],
      [
        "a duplicate (source, code)",
        { mappings: [{ source: "LOINC", code: "1" }, { source: "LOINC", code: "1" }] },
      ],
    ] as [string, unknown][]) {
      expect(
        (await send(request, "put", `rest/test-catalog/tests/${testId}/terminology`, body)).status(),
        label,
      ).toBe(422);
    }
    expect(
      query(
        `SELECT count(*)::text FROM clinlims.test_terminology_mapping WHERE test_id = ${Number(testId)}`,
      )[0][0],
      "no rejected request wrote anything",
    ).toBe("2");

    expect(
      (await send(request, "put", "rest/test-catalog/tests/999999/terminology", { mappings: [] })).status(),
      "an unknown test is 404",
    ).toBe(404);
  });

  test("sample-type display order — writes only the junction rows it already has", async ({
    request,
  }) => {
    const testId = await makeTest(request, "Ord");

    const before = await readJson(
      await request.get("rest/test-catalog/sample-types/1/test-order"),
      "test order",
    );
    expect(before.sampleTypeId, "the sample type").toBe("1");
    const newRow = before.tests.find((t: any) => t.testId === testId);
    expect(newRow, "the new test is listed").toBeTruthy();
    expect(
      "displayOrder" in newRow,
      "with no display order, so the key is absent",
    ).toBe(false);
    // Ordered by displayOrder with nulls LAST, then name case-insensitively.
    const orders = before.tests.map((t: any) =>
      t.displayOrder === undefined ? 2147483647 : t.displayOrder,
    );
    expect(
      orders.every((o: number, i: number) => i === 0 || orders[i - 1] <= o),
      "sorted by display order, nulls last",
    ).toBe(true);

    const saved = await send(
      request,
      "put",
      "rest/test-catalog/sample-types/1/test-order",
      { items: [{ testId, displayOrder: 42 }, { testId: "999999", displayOrder: 1 }] },
    );
    expect(saved.status(), "save").toBe(200);
    const after = await saved.json();
    expect(
      after.tests.find((t: any) => t.testId === testId).displayOrder,
      "the order was written",
    ).toBe(42);
    expect(
      query(
        `SELECT COALESCE(display_order::text, '') FROM clinlims.sampletype_test
          WHERE test_id = ${Number(testId)}`,
      ),
      "and it reached the junction row",
    ).toEqual([["42"]]);
    // A test id that is not under this sample type is not inserted — the loop
    // walks the EXISTING rows and skips what it does not find.
    expect(
      query(
        `SELECT count(*)::text FROM clinlims.sampletype_test
          WHERE sample_type_id = 1 AND test_id = 999999`,
      )[0][0],
      "the unknown id created nothing",
    ).toBe("0");

    expect(
      (await send(request, "put", "rest/test-catalog/sample-types/999999/test-order", { items: [] })).status(),
      "an unknown sample type is 404",
    ).toBe(404);
  });

  test("panel membership — replace-all, and a position fallback that counts items", async ({
    request,
  }) => {
    const testId = await makeTest(request, "Pan");

    // TestAdd put the test in panel 1.
    expect(
      await readJson(
        await request.get(`rest/test-catalog/tests/${testId}/panels`),
        "memberships",
      ),
      "one membership, with no position",
    ).toEqual({
      testId,
      memberships: [{ panelId: "1", panelName: "Bilan Biochimique" }],
    });

    // Panel 3 with an explicit position, then panel 2 with none. The fallback
    // counter starts at 1 and increments once per ITEM, so panel 2 — the second
    // item — gets 2, not 1.
    const saved = await send(
      request,
      "put",
      `rest/test-catalog/tests/${testId}/panels`,
      { memberships: [{ panelId: "3", position: 5 }, { panelId: "2", position: null }] },
    );
    expect(saved.status(), "save").toBe(200);
    expect(
      (await saved.json()).memberships,
      "sorted by panel name, and panel 1 is gone",
    ).toEqual([
      { panelId: "2", panelName: "NFS", position: 2 },
      { panelId: "3", panelName: "Typage lymphocytaire", position: 5 },
    ]);
    expect(
      query(
        `SELECT panel_id::text, COALESCE(sort_order::text, '') FROM clinlims.panel_item
          WHERE test_id = ${Number(testId)} ORDER BY panel_id`,
      ),
      "the membership set was replaced outright",
    ).toEqual([
      ["2", "2"],
      ["3", "5"],
    ]);

    // An unknown panel is rejected up front, so the rest of the request is not
    // applied either.
    expect(
      (await send(request, "put", `rest/test-catalog/tests/${testId}/panels`, {
        memberships: [{ panelId: "2", position: 1 }, { panelId: "999999" }],
      })).status(),
      "an unknown panel is 422",
    ).toBe(422);
    expect(
      query(
        `SELECT count(*)::text FROM clinlims.panel_item WHERE test_id = ${Number(testId)}`,
      )[0][0],
      "and nothing was written",
    ).toBe("2");

    // The panel-side preview reads the same rows back.
    const preview = await readJson(
      await request.get("rest/test-catalog/panels/3/test-order"),
      "panel test order",
    );
    expect(preview.panelId, "the panel").toBe("3");
    expect(
      preview.tests.find((t: any) => t.testId === testId).position,
      "the position this test holds in it",
    ).toBe(5);

    expect(
      (await send(request, "put", "rest/test-catalog/tests/999999/panels", { memberships: [] })).status(),
      "an unknown test is 404",
    ).toBe(404);
    expect(
      (await request.get("rest/test-catalog/panels/999999/test-order")).status(),
      "an unknown panel is 404",
    ).toBe(404);
  });

  test("POST /panels — a blank name is 422, and every other name is a 500", async ({
    request,
  }) => {
    expect(
      (await send(request, "post", "rest/test-catalog/panels", { name: "  " })).status(),
      "a blank name is rejected cleanly",
    ).toBe(422);

    const before = query(`SELECT count(*)::text FROM clinlims.panel`)[0][0];

    // panel.name_localization_id is NOT NULL and createPanel never writes a
    // localization, so the insert cannot succeed. This is a Java defect the
    // port reproduces: see migration/java-defects-found.md.
    const res = await send(request, "post", "rest/test-catalog/panels", {
      name: `${PREFIX}Panel`,
    });
    expect(res.status(), "a valid name still fails").toBe(500);
    expect(
      query(`SELECT count(*)::text FROM clinlims.panel`)[0][0],
      "and the failed insert rolled back",
    ).toBe(before);
    expect(
      query(
        `SELECT count(*)::text FROM clinlims.panel WHERE name = '${PREFIX}Panel'`,
      )[0][0],
      "no panel by that name exists",
    ).toBe("0");
  });
});
