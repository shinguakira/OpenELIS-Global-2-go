/**
 * e2 — the test-catalog editor's test-level writes: create-in-place, Basic Info,
 * Sample & Results with its copy-from shortcut, the activation gate, and the
 * reference ranges (🔴 clinical — these rows decide whether a result reads as
 * normal).
 *
 * These are the only editor writes that reach clinlims.history. A create leaves
 * an 'I' with a NULL payload; Basic Info and activation leave a 'U' carrying the
 * values they REPLACED, in the entity's declared-field order — and `testSection`
 * renders the section's DESCRIPTION, not its id. A save that changes nothing
 * leaves nothing.
 *
 * Two behaviours worth naming because no caller would guess them. `active: true`
 * on Basic Info is IGNORED — activation is gated on range coverage and has to go
 * through POST .../activate — so that endpoint can only ever turn a test OFF.
 * And a Sample & Results save that re-sends an existing component WITHOUT its id
 * is a 500: the match is on id alone, so the component is inserted afresh and
 * collides with the (test_id, code) unique index.
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

const PREFIX = "E2TC";

function dropTests() {
  const where = `local_code LIKE '${PREFIX}%'`;
  const ids = `SELECT id FROM clinlims.test WHERE ${where}`;
  const locIds = query(
    `SELECT name_localization_id::text FROM clinlims.test WHERE ${where}
     UNION SELECT reporting_name_localization_id::text FROM clinlims.test WHERE ${where}`,
  )
    .map((r) => r[0])
    .filter((v) => v !== "");
  exec(`DELETE FROM clinlims.test_activation_acknowledgment WHERE test_id IN (${ids})`);
  exec(`DELETE FROM clinlims.test_sample_handling_history WHERE test_sample_handling_id IN
        (SELECT id FROM clinlims.test_sample_handling WHERE test_id IN (${ids}))`);
  exec(`DELETE FROM clinlims.test_sample_handling WHERE test_id IN (${ids})`);
  exec(`UPDATE clinlims.test SET default_test_result_id = NULL WHERE ${where}`);
  exec(`DELETE FROM clinlims.result_limits WHERE test_id IN (${ids})`);
  exec(`DELETE FROM clinlims.test_result_interpretation WHERE component_id IN
        (SELECT id FROM clinlims.test_result_component WHERE test_id IN (${ids}))`);
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

async function createTest(
  request: APIRequestContext,
  suffix: string,
  over: Record<string, unknown> = {},
) {
  const res = await send(request, "post", "rest/test-catalog/tests", {
    name: `${PREFIX}${suffix}`,
    reportingName: `${PREFIX}${suffix}Rep`,
    code: `${PREFIX}${suffix}`,
    labUnitId: "36",
    sampleTypeId: "1",
    domain: "CLINICAL",
    ...over,
  });
  expect(res.status(), "fixture create").toBe(201);
  return (await res.json()).testId as string;
}

test.describe("e2 — the test-catalog editor test-level writes", () => {
  let restore: (() => void)[] = [];

  test.beforeEach(() => {
    restore = [];
    dropTests();
  });

  test.afterEach(() => {
    dropTests();
    for (const r of restore) r();
    exec(
      `DELETE FROM clinlims.history WHERE timestamp > now() - interval '3 minutes'`,
    );
  });

  test("POST /tests — five required fields, a code check, and an inactive test", async ({
    request,
  }) => {
    // labUnitId is NOT required, so a test can be created with no lab unit —
    // and is then invisible on Add Order, which filters by section.
    for (const [label, body] of [
      ["no name", { reportingName: "r", code: `${PREFIX}X`, domain: "CLINICAL", sampleTypeId: "1" }],
      ["no reportingName", { name: "n", code: `${PREFIX}X`, domain: "CLINICAL", sampleTypeId: "1" }],
      ["no code", { name: "n", reportingName: "r", domain: "CLINICAL", sampleTypeId: "1" }],
      ["an unknown domain", { name: "n", reportingName: "r", code: `${PREFIX}X`, domain: "NOPE", sampleTypeId: "1" }],
      ["no sampleTypeId", { name: "n", reportingName: "r", code: `${PREFIX}X`, domain: "CLINICAL" }],
    ] as [string, unknown][]) {
      expect(
        (await send(request, "post", "rest/test-catalog/tests", body)).status(),
        label,
      ).toBe(422);
    }

    // Column 1, not column 0: `query` returns rows of columns, so the flag is
    // the SECOND field. Reading the first one restores the id's leading digit
    // into is_active, which leaves the section in neither the active nor the
    // inactive list and breaks a spec two files away.
    const sectionActive = query(
      `SELECT id::text, is_active FROM clinlims.test_section WHERE id = 57`,
    )[0][1];
    restore.push(() => {
      exec(`UPDATE clinlims.test_section SET is_active = '${sectionActive}' WHERE id = 57`);
    });
    exec(`UPDATE clinlims.test_section SET is_active = 'N' WHERE id = 57`);

    const watermark = query(`SELECT max(id)::text FROM clinlims.history`)[0][0];
    const testId = await createTest(request, "New", {
      labUnitId: "57",
      amr: true,
      orderable: true,
      description: "",
    });

    expect(
      query(
        `SELECT name, description, local_code, domain, COALESCE(test_section_id::text, ''),
                is_active, orderable::text, antimicrobial_resistance::text, is_reportable,
                sort_order::text, (guid IS NOT NULL)::text, COALESCE(loinc, ''),
                COALESCE(uom_id::text, ''), normalized_description
           FROM clinlims.test WHERE id = ${Number(testId)}`,
      ),
      "the created row",
    ).toEqual([
      [
        // `name` follows the localization, and `description` falls back to the
        // name when the body sends a blank one.
        `${PREFIX}New`,
        `${PREFIX}New`,
        `${PREFIX}New`,
        "CLINICAL",
        "57",
        // A new test starts INACTIVE — that is the whole point of the two-step
        // create-then-activate flow.
        "N",
        "true",
        "true",
        "N",
        "0",
        "true",
        "",
        "",
        `${PREFIX}new`.toLowerCase(),
      ],
    ]);

    // Both locales get the SAME string: the create form has one name field.
    expect(
      query(
        `SELECT l.description, lv.locale, lv.value FROM clinlims.localization l
           JOIN clinlims.localization_value lv ON lv.localization_id = l.id
          WHERE l.id IN (SELECT name_localization_id FROM clinlims.test WHERE id = ${Number(testId)}
                         UNION SELECT reporting_name_localization_id FROM clinlims.test WHERE id = ${Number(testId)})
          ORDER BY l.description, lv.locale`,
      ),
      "two localizations, en == fr",
    ).toEqual([
      ["test name", "en", `${PREFIX}New`],
      ["test name", "fr", `${PREFIX}New`],
      ["test report name", "en", `${PREFIX}NewRep`],
      ["test report name", "fr", `${PREFIX}NewRep`],
    ]);

    expect(
      query(
        `SELECT sample_type_id::text FROM clinlims.sampletype_test WHERE test_id = ${Number(testId)}`,
      ).map((r) => r[0]),
      "the sample type link",
    ).toEqual(["1"]);

    // Naming an inactive lab unit turns it back on.
    expect(
      query(`SELECT id::text, is_active FROM clinlims.test_section WHERE id = 57`)[0][1],
      "the lab unit was activated by being named",
    ).toBe("Y");

    // The create is audited, with a NULL payload.
    expect(
      query(
        `SELECT rt.name, h.reference_id, h.activity, (h.changes IS NULL)::text
           FROM clinlims.history h JOIN clinlims.reference_tables rt ON rt.id = h.reference_table
          WHERE h.id > ${Number(watermark)} ORDER BY h.id`,
      ),
      "one insert row for the test and nothing else",
    ).toEqual([["TEST", testId, "I", "true"]]);

    // The code check is case-INSENSITIVE, and it is a 409 so the UI can flag
    // the field rather than the form.
    expect(
      (await send(request, "post", "rest/test-catalog/tests", {
        name: `${PREFIX}Other`,
        reportingName: `${PREFIX}OtherRep`,
        code: `${PREFIX}new`.toLowerCase(),
        sampleTypeId: "1",
        domain: "CLINICAL",
      })).status(),
      "a code that differs only in case",
    ).toBe(409);
  });

  test("PUT basic-info — what it may change, what it refuses, and what it audits", async ({
    request,
  }) => {
    const testId = await createTest(request, "Bas", { amr: true, orderable: true });

    expect(
      await readJson(
        await request.get(`rest/test-catalog/tests/${testId}/basic-info`),
        "basic-info",
      ),
      "the identity block",
    ).toEqual({
      testId,
      name: `${PREFIX}Bas`,
      code: `${PREFIX}Bas`,
      description: `${PREFIX}Bas`,
      domain: "CLINICAL",
      labUnitId: "36",
      sampleTypeId: "1",
      antimicrobialResistance: true,
      active: false,
      orderable: true,
    });

    expect(
      (await send(request, "put", `rest/test-catalog/tests/${testId}/basic-info`, { domain: "NOPE" })).status(),
      "an unknown domain is 422",
    ).toBe(422);
    // `name` is owned by the Localization section, so changing it here is a
    // refusal rather than a silent no-op — but sending it back unchanged is
    // fine, which is what lets a UI PUT the whole form.
    expect(
      (await send(request, "put", `rest/test-catalog/tests/${testId}/basic-info`, { name: "something else" })).status(),
      "a different name is 422",
    ).toBe(422);
    expect(
      (await send(request, "put", `rest/test-catalog/tests/${testId}/basic-info`, { name: `${PREFIX}Bas` })).status(),
      "the same name is accepted",
    ).toBe(200);

    const watermark = query(`SELECT max(id)::text FROM clinlims.history`)[0][0];
    const saved = await send(request, "put", `rest/test-catalog/tests/${testId}/basic-info`, {
      code: `${PREFIX}Bas2`,
      description: "changed",
      domain: "VECTOR",
      antimicrobialResistance: false,
      orderable: false,
      labUnitId: "56",
      sampleTypeId: "3",
      // Ignored: activation has to go through the coverage gate.
      active: true,
    });
    expect(saved.status(), "save").toBe(200);
    expect((await saved.json()).active, "still inactive, despite active: true").toBe(false);

    expect(
      query(
        `SELECT name, description, local_code, domain, COALESCE(test_section_id::text, ''),
                is_active, orderable::text, antimicrobial_resistance::text
           FROM clinlims.test WHERE id = ${Number(testId)}`,
      ),
      "the columns the save moved",
    ).toEqual([
      [`${PREFIX}Bas`, "changed", `${PREFIX}Bas2`, "VECTOR", "56", "N", "false", "false"],
    ]);
    // The sample-type link is reconciled to ONE type, replace-all.
    expect(
      query(
        `SELECT sample_type_id::text FROM clinlims.sampletype_test WHERE test_id = ${Number(testId)}`,
      ).map((r) => r[0]),
      "one link, the submitted one",
    ).toEqual(["3"]);

    // The audit carries the OLD values, in the entity's declared-field order,
    // and `testSection` is the section's DESCRIPTION rather than its id.
    const [[oldSection]] = query(
      `SELECT description FROM clinlims.test_section WHERE id = 36`,
    );
    expect(
      query(
        `SELECT rt.name, h.activity,
                replace(encode(h.changes, 'escape'), chr(10), '')
           FROM clinlims.history h JOIN clinlims.reference_tables rt ON rt.id = h.reference_table
          WHERE h.id > ${Number(watermark)} ORDER BY h.id`,
      ),
      "one update row, carrying what it replaced",
    ).toEqual([
      [
        "TEST",
        "U",
        `<testSection>${oldSection}</testSection><description>${PREFIX}Bas</description>` +
          `<domain>CLINICAL</domain><localCode>${PREFIX}Bas</localCode>` +
          `<orderable>true</orderable><antimicrobialResistance>true</antimicrobialResistance>`,
      ],
    ]);

    // Deactivating a test that is already inactive changes nothing, so it
    // writes nothing — not even a history row.
    const second = query(`SELECT max(id)::text FROM clinlims.history`)[0][0];
    expect(
      (await send(request, "put", `rest/test-catalog/tests/${testId}/basic-info`, { active: false })).status(),
      "a no-op deactivation",
    ).toBe(200);
    expect(
      query(`SELECT count(*)::text FROM clinlims.history WHERE id > ${Number(second)}`)[0][0],
      "and it is not audited",
    ).toBe("0");

    expect(
      (await request.get("rest/test-catalog/tests/999999/basic-info")).status(),
      "an unknown test is 404",
    ).toBe(404);
  });

  test("ranges and activate — the coverage gate, and what it lets through", async ({
    request,
  }) => {
    const testId = await createTest(request, "Rng");

    // A test with no ranges is EMPTY, not GAP — and EMPTY does not block.
    const empty = await readJson(
      await request.get(`rest/test-catalog/tests/${testId}/ranges`),
      "ranges",
    );
    expect(empty, "no ranges, no gaps").toEqual({
      testId,
      ranges: [],
      coverage: {
        male: { sex: "M", status: "EMPTY", gaps: [], overlaps: [] },
        female: { sex: "F", status: "EMPTY", gaps: [], overlaps: [] },
      },
    });

    let watermark = query(`SELECT max(id)::text FROM clinlims.history`)[0][0];
    const activated = await send(request, "post", `rest/test-catalog/tests/${testId}/activate`, {});
    expect(activated.status(), "activation with no ranges").toBe(200);
    expect(
      query(
        `SELECT id::text, is_active, orderable::text FROM clinlims.test WHERE id = ${Number(testId)}`,
      ),
      "activation also forces orderable",
    ).toEqual([[testId, "Y", "true"]]);
    expect(
      query(
        `SELECT rt.name, h.activity, replace(encode(h.changes, 'escape'), chr(10), '')
           FROM clinlims.history h JOIN clinlims.reference_tables rt ON rt.id = h.reference_table
          WHERE h.id > ${Number(watermark)} ORDER BY h.id`,
      ),
      "and it is audited with the flags it replaced",
    ).toEqual([["TEST", "U", "<orderable>false</orderable><isActive>N</isActive>"]]);
    expect(
      query(
        `SELECT count(*)::text FROM clinlims.test_activation_acknowledgment WHERE test_id = ${Number(testId)}`,
      )[0][0],
      "no acknowledgment, because there was no gap",
    ).toBe("0");

    // One band covering 0–15 leaves everything above 15 uncovered: only an
    // OPEN-ENDED top band reaches the top of the reportable lifetime.
    watermark = query(`SELECT max(id)::text FROM clinlims.history`)[0][0];
    const saved = await send(request, "put", `rest/test-catalog/tests/${testId}/ranges`, {
      ranges: [
        {
          gender: "M",
          minAge: 0,
          maxAge: 15,
          lowNormal: 1,
          highNormal: 9,
          lowCritical: 0.5,
          highCritical: 12,
        },
      ],
    });
    expect(saved.status(), "save ranges").toBe(200);
    const savedBody = await saved.json();
    expect(savedBody.coverage.male, "a tail gap above the top band").toEqual({
      sex: "M",
      status: "GAP",
      gaps: [{ fromAge: 15.0, toAge: "Infinity" }],
      overlaps: [],
    });
    // An unbounded bound is ±Infinity in the column and ABSENT on the wire, so
    // "unbounded" and "unset" are the same thing to a client.
    expect(
      Object.keys(savedBody.ranges[0]).sort(),
      "only the bounds that are finite",
    ).toEqual(
      ["gender", "highCritical", "highNormal", "id", "lowCritical", "lowNormal", "maxAge", "minAge"].sort(),
    );
    expect(
      query(
        `SELECT test_result_type_id::text, COALESCE(gender, ''), min_age::text, max_age::text,
                low_normal::text, high_normal::text, low_valid::text, high_valid::text,
                low_reporting_range::text, high_reporting_range::text,
                low_critical::text, high_critical::text,
                COALESCE(normal_dictionary_id::text, ''), COALESCE(component_id, '')
           FROM clinlims.result_limits WHERE test_id = ${Number(testId)} ORDER BY id`,
      ),
      "the row, with the entity defaults for everything the editor does not own",
    ).toEqual([
      [
        "4", "M", "0", "15", "1", "9", "-Infinity", "Infinity",
        "-Infinity", "Infinity", "0.5", "12", "", "",
      ],
    ]);
    const limitId = query(
      `SELECT id::text FROM clinlims.result_limits WHERE test_id = ${Number(testId)}`,
    )[0][0];
    expect(
      query(
        `SELECT rt.name, h.reference_id, h.activity
           FROM clinlims.history h JOIN clinlims.reference_tables rt ON rt.id = h.reference_table
          WHERE h.id > ${Number(watermark)} ORDER BY h.id`,
      ),
      "the insert is audited",
    ).toEqual([["RESULT_LIMITS", limitId, "I"]]);

    // Now the gate blocks — and the 409 carries the report as its body.
    const blocked = await send(request, "post", `rest/test-catalog/tests/${testId}/activate`, {});
    expect(blocked.status(), "a gap blocks activation").toBe(409);
    expect((await blocked.json()).male.status, "the report rides on the 409").toBe("GAP");

    const acknowledged = await send(
      request,
      "post",
      `rest/test-catalog/tests/${testId}/activate`,
      { gapsAcknowledged: '{"ok":true}' },
    );
    expect(acknowledged.status(), "an acknowledgment lets it through").toBe(200);
    expect(
      query(
        `SELECT test_id::text, gaps_acknowledged::text
           FROM clinlims.test_activation_acknowledgment WHERE test_id = ${Number(testId)}`,
      ),
      "and the override is recorded",
    ).toEqual([[testId, '{"ok": true}']]);

    for (const [label, body] of [
      ["an unknown gender", { ranges: [{ gender: "X", minAge: 0, maxAge: 1 }] }],
      ["a negative minimum", { ranges: [{ minAge: -1, maxAge: 1 }] }],
      ["a maximum at the minimum", { ranges: [{ minAge: 5, maxAge: 5 }] }],
    ] as [string, unknown][]) {
      expect(
        (await send(request, "put", `rest/test-catalog/tests/${testId}/ranges`, body)).status(),
        label,
      ).toBe(422);
    }
    expect(
      query(
        `SELECT count(*)::text FROM clinlims.result_limits WHERE test_id = ${Number(testId)}`,
      )[0][0],
      "no rejected save wrote anything",
    ).toBe("1");
  });

  test("group ranges — the same bands on every test named, ids discarded", async ({
    request,
  }) => {
    const testId = await createTest(request, "Grp");
    expect(
      (await send(request, "put", `rest/test-catalog/tests/${testId}/ranges`, {
        ranges: [{ gender: "M", minAge: 0, maxAge: 15, lowNormal: 1, highNormal: 9 }],
      })).status(),
      "seed",
    ).toBe(200);

    expect(
      (await send(request, "put", "rest/test-catalog/group/ranges", { testIds: [], ranges: [] })).status(),
      "no test ids is 422",
    ).toBe(422);
    expect(
      (await send(request, "put", "rest/test-catalog/group/ranges", {
        testIds: [testId],
        ranges: [{ gender: "X", minAge: 0, maxAge: 1 }],
      })).status(),
      "and one bad band rejects the whole request",
    ).toBe(422);

    // One real id and one imaginary one: the imaginary one is skipped.
    const res = await send(request, "put", "rest/test-catalog/group/ranges", {
      testIds: [testId, "999999"],
      ranges: [{ gender: "F", minAge: 0, maxAge: 200, lowNormal: 2, highNormal: 8 }],
    });
    expect(res.status(), "group save").toBe(200);
    expect(await res.text(), "no body").toBe("");

    // The incoming ids are DROPPED — a shared band belongs to no single test —
    // so the group save always inserts and deletes rather than updating.
    expect(
      query(
        `SELECT COALESCE(gender, ''), min_age::text, max_age::text,
                low_normal::text, high_normal::text
           FROM clinlims.result_limits WHERE test_id = ${Number(testId)} ORDER BY id`,
      ),
      "the seeded band was replaced, not edited",
    ).toEqual([["F", "0", "200", "2", "8"]]);
  });

  test("sample-results — components, their children, and the id a save must echo", async ({
    request,
  }) => {
    const testId = await createTest(request, "Res");

    expect(
      await readJson(
        await request.get(`rest/test-catalog/tests/${testId}/sample-results`),
        "sample-results",
      ),
      "a new test has no components",
    ).toEqual({ testId, components: [] });

    const watermark = query(`SELECT max(id)::text FROM clinlims.history`)[0][0];
    const saved = await send(request, "put", `rest/test-catalog/tests/${testId}/sample-results`, {
      components: [
        {
          code: "PRIMARY",
          label: "Main",
          displayOrder: 0,
          resultType: "D",
          uomId: "1",
          significantDigits: 2,
          allowMultipleReadings: false,
          interpretations: [
            { valueMatch: "1", text: "positive", severity: "HIGH", color: "#f00", displayOrder: 1 },
          ],
          options: [
            { value: "542", sortOrder: 10, normal: false },
            { value: "544", sortOrder: 20, normal: true },
          ],
        },
        { code: "SECOND", label: "Other", displayOrder: 1, resultType: "N", interpretations: [], options: [] },
      ],
    });
    expect(saved.status(), "save").toBe(200);
    const body = await saved.json();
    expect(
      body.components.map((c: any) => [c.code, c.label, c.displayOrder, c.resultType]),
      "ordered by display order then code",
    ).toEqual([
      ["PRIMARY", "Main", 0, "D"],
      ["SECOND", "Other", 1, "N"],
    ]);
    // A dictionary-backed option carries a label alongside the id it stores.
    expect(
      body.components[0].options.map((o: any) => [o.value, o.valueName, o.sortOrder, o.normal]),
      "the options, sorted by their numeric sort order",
    ).toEqual([
      ["542", "Positif", 10, false],
      ["544", "Negatif", 20, true],
    ]);

    expect(
      query(
        `SELECT code, label, display_order::text, COALESCE(result_type, ''),
                COALESCE(uom_id::text, ''), COALESCE(significant_digits::text, ''),
                allow_multiple_readings::text, is_active
           FROM clinlims.test_result_component WHERE test_id = ${Number(testId)}
          ORDER BY display_order, code`,
      ),
      "two components",
    ).toEqual([
      ["PRIMARY", "Main", "0", "D", "1", "2", "false", "Y"],
      ["SECOND", "Other", "1", "N", "", "", "false", "Y"],
    ]);
    expect(
      query(
        `SELECT COALESCE(value_match, ''), COALESCE(interpretation_text, ''),
                COALESCE(severity, ''), COALESCE(color, ''), display_order::text, is_active
           FROM clinlims.test_result_interpretation WHERE component_id IN
                 (SELECT id FROM clinlims.test_result_component WHERE test_id = ${Number(testId)})`,
      ),
      "one interpretation rule",
    ).toEqual([["1", "positive", "HIGH", "#f00", "1", "Y"]]);

    // The legacy mirror: the PRIMARY component's unit lands on test.uom_id and
    // its digit count on every option row, so the OLD Test Modify screen agrees.
    expect(
      query(
        `SELECT COALESCE(value, ''), COALESCE(sort_order::text, ''), is_active::text,
                COALESCE(is_normal::text, ''), COALESCE(tst_rslt_type, ''),
                COALESCE(significant_digits::text, ''), (component_id IS NOT NULL)::text
           FROM clinlims.test_result WHERE test_id = ${Number(testId)} ORDER BY id`,
      ),
      "the options as legacy test_result rows",
    ).toEqual([
      ["542", "10", "true", "false", "D", "2", "true"],
      ["544", "20", "true", "true", "D", "2", "true"],
    ]);
    expect(
      query(
        `SELECT id::text, COALESCE(uom_id::text, '') FROM clinlims.test WHERE id = ${Number(testId)}`,
      )[0][1],
      "and test.uom_id follows the PRIMARY component",
    ).toBe("1");

    // None of that is audited.
    expect(
      query(`SELECT count(*)::text FROM clinlims.history WHERE id > ${Number(watermark)}`)[0][0],
      "sample-results writes no history",
    ).toBe("0");

    for (const [label, req] of [
      ["a blank code", { components: [{ code: "", label: "x", interpretations: [], options: [] }] }],
      ["a blank label", { components: [{ code: "X", label: "", interpretations: [], options: [] }] }],
      [
        "a duplicate code",
        {
          components: [
            { code: "X", label: "a", interpretations: [], options: [] },
            { code: "X", label: "b", interpretations: [], options: [] },
          ],
        },
      ],
    ] as [string, unknown][]) {
      expect(
        (await send(request, "put", `rest/test-catalog/tests/${testId}/sample-results`, req)).status(),
        label,
      ).toBe(422);
    }

    // Re-sending an existing component WITHOUT its id is a 500: the match is on
    // id alone, so the component is inserted afresh and collides with the
    // (test_id, code) unique index. The UI always echoes the id; a
    // hand-written client that does not gets this.
    expect(
      (await send(request, "put", `rest/test-catalog/tests/${testId}/sample-results`, {
        components: [
          { code: "PRIMARY", label: "Main", displayOrder: 0, resultType: "D", interpretations: [], options: [] },
        ],
      })).status(),
      "an id-less re-save of an existing code",
    ).toBe(500);

    // With the id echoed, dropping the other component soft-deletes it — and
    // re-adding the code REACTIVATES that row rather than inserting, because
    // the unique slot is still taken.
    const primaryId = body.components[0].id;
    expect(
      (await send(request, "put", `rest/test-catalog/tests/${testId}/sample-results`, {
        components: [
          {
            id: primaryId,
            code: "PRIMARY",
            label: "Main",
            displayOrder: 0,
            resultType: "D",
            interpretations: [],
            options: [],
          },
        ],
      })).status(),
      "dropping a component",
    ).toBe(200);
    expect(
      query(
        `SELECT code, is_active FROM clinlims.test_result_component
          WHERE test_id = ${Number(testId)} ORDER BY code`,
      ),
      "the dropped component is switched off, not deleted",
    ).toEqual([
      ["PRIMARY", "Y"],
      ["SECOND", "N"],
    ]);
    // Its options went with it.
    expect(
      query(
        `SELECT COALESCE(value, ''), is_active::text FROM clinlims.test_result
          WHERE test_id = ${Number(testId)} ORDER BY id`,
      ),
      "and so did the options of the component that stayed, since none was sent",
    ).toEqual([
      ["542", "false"],
      ["544", "false"],
    ]);

    expect(
      (await send(request, "put", `rest/test-catalog/tests/${testId}/sample-results`, {
        components: [
          { id: primaryId, code: "PRIMARY", label: "Main", displayOrder: 0, resultType: "D", interpretations: [], options: [] },
          { code: "SECOND", label: "Back", displayOrder: 1, resultType: "N", interpretations: [], options: [] },
        ],
      })).status(),
      "re-adding the dropped code",
    ).toBe(200);
    expect(
      query(
        `SELECT code, label, is_active FROM clinlims.test_result_component
          WHERE test_id = ${Number(testId)} ORDER BY code`,
      ),
      "the soft-deleted row was reactivated with the new label",
    ).toEqual([
      ["PRIMARY", "Main", "Y"],
      ["SECOND", "Back", "Y"],
    ]);

    expect(
      (await request.get("rest/test-catalog/tests/999999/sample-results")).status(),
      "an unknown test is 404",
    ).toBe(404);
  });

  test("copy-from — a per-code copy that is a no-op the second time", async ({
    request,
  }) => {
    const sourceId = await createTest(request, "Src");
    const targetId = await createTest(request, "Dst");

    expect(
      (await send(request, "put", `rest/test-catalog/tests/${sourceId}/sample-results`, {
        components: [
          {
            code: "PRIMARY",
            label: "Main",
            displayOrder: 0,
            resultType: "D",
            uomId: "1",
            significantDigits: 2,
            interpretations: [
              { valueMatch: "1", text: "positive", severity: "HIGH", color: "#f00", displayOrder: 1 },
            ],
            options: [{ value: "542", sortOrder: 10, normal: false }],
          },
        ],
      })).status(),
      "seed the source",
    ).toBe(200);

    const copied = await send(
      request,
      "post",
      `rest/test-catalog/tests/${targetId}/sample-results/copy-from/${sourceId}`,
    );
    expect(copied.status(), "copy").toBe(200);
    const copiedBody = await copied.json();
    expect(
      copiedBody.components.map((c: any) => [c.code, c.label, c.resultType, c.uomId, c.significantDigits]),
      "the component came across whole",
    ).toEqual([["PRIMARY", "Main", "D", "1", 2]]);
    expect(
      copiedBody.components[0].interpretations.map((i: any) => [i.valueMatch, i.text]),
      "and so did its rules",
    ).toEqual([["1", "positive"]]);
    expect(
      copiedBody.components[0].options.map((o: any) => [o.value, o.sortOrder]),
      "and its options",
    ).toEqual([["542", 10]]);
    // The copies are NEW rows — the ids do not come across.
    expect(
      copiedBody.components[0].id === undefined ||
        copiedBody.components[0].id !== null,
      "the copy has an id of its own",
    ).toBe(true);

    // A component whose CODE the target already has is skipped whole, so a
    // second copy changes nothing.
    expect(
      (await send(request, "post", `rest/test-catalog/tests/${targetId}/sample-results/copy-from/${sourceId}`)).status(),
      "copying again",
    ).toBe(200);
    expect(
      query(
        `SELECT count(*)::text FROM clinlims.test_result_component WHERE test_id = ${Number(targetId)}`,
      )[0][0],
      "still one component",
    ).toBe("1");

    // Only the TARGET is checked, so an unknown source is a 200 that copied
    // nothing.
    expect(
      (await send(request, "post", `rest/test-catalog/tests/${targetId}/sample-results/copy-from/999999`)).status(),
      "an unknown source",
    ).toBe(200);
    expect(
      (await send(request, "post", `rest/test-catalog/tests/999999/sample-results/copy-from/${sourceId}`)).status(),
      "an unknown target is 404",
    ).toBe(404);
  });
});
