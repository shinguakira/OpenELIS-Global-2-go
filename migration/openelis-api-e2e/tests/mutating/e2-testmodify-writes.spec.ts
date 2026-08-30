/**
 * e2 — TestModifyEntry.
 *
 * The GET is TestAdd's eight lists plus a catalogue that is built ONLY when a
 * filter is given; the POST is a delete-then-insert rewrite of everything
 * hanging off one test.
 *
 * Two things the screen does not say. `test.name` is not among the columns the
 * update writes and moves anyway — Hibernate maps it to a DERIVED getter that
 * reads the localization — while `description` and `local_code` are never
 * rewritten, so a renamed test keeps describing itself by its old name. And a
 * NUMERIC modify does not deactivate the results it replaces, so each one
 * leaves another active row behind; only the dictionary variants clear up after
 * themselves.
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

async function post(
  request: APIRequestContext,
  path: string,
  data: Record<string, unknown>,
) {
  const body = await readJson(await request.get(SESSION_PATH), SESSION_PATH);
  return request.post(path, {
    headers: {
      [CSRF_HEADER]: body[CSRF_SESSION_FIELD],
      "Content-Type": "application/json",
    },
    data,
  });
}

const PREFIX = "E2TM";

/**
 * Everything a create leaves behind, removed in FK order.
 *
 * Keyed on local_code, NOT on name: the modify rewrites `name` out from under a
 * name-keyed cleanup and leaves the rows orphaned.
 */
function dropTests() {
  const where = `local_code LIKE '${PREFIX}%'`;
  const ids = `SELECT id FROM clinlims.test WHERE ${where}`;
  const locIds = query(
    `SELECT name_localization_id::text FROM clinlims.test WHERE ${where}
     UNION SELECT reporting_name_localization_id::text FROM clinlims.test WHERE ${where}`,
  )
    .map((r) => r[0])
    .filter((v) => v !== "");
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

/** A numeric test to modify, created through the endpoint that owns creates. */
function createWad(suffix: string, over: Record<string, unknown> = {}) {
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
    ...over,
  };
}

test.describe("e2 — TestModifyEntry", () => {
  let restore: (() => void)[] = [];

  test.beforeEach(() => {
    restore = [];
    dropTests();
  });

  test.afterEach(async ({ request }) => {
    dropTests();
    for (const r of restore) r();
    // One throwaway create to RESYNC Java's cached lists, for the reason
    // e2-testadd-writes explains: both endpoints refresh the section lists from
    // the database, and a spec that restores a flag with SQL afterwards leaves
    // Java serving what it last saw.
    await post(request, "rest/TestAdd", {
      jsonWad: JSON.stringify(createWad("Resync")),
      loinc: "",
    });
    dropTests();
    exec(
      `DELETE FROM clinlims.history WHERE timestamp > now() - interval '2 minutes'`,
    );
  });

  test("GET TestModifyEntry — no filter, no catalogue", async ({ request }) => {
    const f = await readJson(
      await request.get("rest/TestModifyEntry"),
      "TestModifyEntry",
    );

    expect(f.formName, "formName").toBe("testModifyEntryForm");
    expect(f.testId, "testId default").toBe("");
    expect(f.nameEnglish, "nameEnglish default").toBe("");
    expect(f.jsonWad, "jsonWad default").toBe("");
    expect("loinc" in f, "loinc is null on the blank form, so absent").toBe(false);

    // The catalogue is skipped entirely without a filter — a guard on the
    // initial page load, not an empty result.
    expect(f.testCatBeanList, "no catalogue without a filter").toEqual([]);

    // labUnitList is the ACTIVE sections ALONE. TestAdd's copy concatenates the
    // inactive ones onto the same list; this screen does not, so a test cannot
    // be moved onto a disabled section from here.
    const activeSections = Number(
      query(
        `SELECT count(*)::text FROM clinlims.test_section WHERE is_active = 'Y'`,
      )[0][0],
    );
    expect(f.labUnitList.length, "active test sections only").toBe(activeSections);

    // The rest of the lists are TestAdd's, unchanged.
    expect(f.resultTypeList, "six result types, Titer dropped").toEqual([
      { id: "1", value: "Free text" },
      { id: "2", value: "Select List" },
      { id: "4", value: "Numeric" },
      { id: "5", value: "Alphanumeric" },
      { id: "6", value: "Multiselect" },
      { id: "7", value: "Cascading multiselect" },
    ]);
    expect(f.ageRangeList[4], "Infinity sorts last").toEqual({
      id: "Infinity",
      value: "Adult",
    });
  });

  test("GET TestModifyEntry?sampleType — the catalogue, and its two shapes", async ({
    request,
  }) => {
    const f = await readJson(
      await request.get("rest/TestModifyEntry?sampleType=1"),
      "TestModifyEntry?sampleType=1",
    );

    const expected = query(
      `SELECT DISTINCT t.id::text FROM clinlims.test t
         JOIN clinlims.sampletype_test st ON st.test_id = t.id
        WHERE st.sample_type_id = 1`,
    ).map((r) => r[0]);
    expect(
      f.testCatBeanList.map((b: any) => b.id).sort(),
      "every test under the sample type, once",
    ).toEqual([...expected].sort());

    const numeric = f.testCatBeanList.find((b: any) => b.hasLimitValues);
    expect(numeric, "the fixture has a numeric test here").toBeTruthy();
    // A numeric bean carries the limit block and no dictionary block.
    expect(numeric.hasDictionaryValues, "not a dictionary test").toBe(false);
    expect("dictionaryValues" in numeric, "and no dictionary keys").toBe(false);
    expect(Object.keys(numeric.resultLimits[0]).sort(), "limit bean keys").toEqual(
      [
        "ageRange",
        "criticalRange",
        "gender",
        "normalRange",
        "reportingRange",
        "validRange",
      ],
    );
    // Every one is a rendered STRING, and the two defaults are literals: a
    // blank gender is "n/a" and an unbounded age is "Any Age".
    expect(numeric.resultLimits[0].gender, "no gender renders n/a").toBe("n/a");
    expect(numeric.resultLimits[0].ageRange, "an unbounded age").toBe("Any Age");
    expect(
      numeric.resultLimits[0].reportingRange,
      "-Infinity..Infinity renders as a phrase, not a range",
    ).toBe("Any value");

    const dict = f.testCatBeanList.find((b: any) => b.hasDictionaryValues);
    expect(dict, "and a dictionary test").toBeTruthy();
    expect(dict.hasLimitValues, "which has no limit block").toBe(false);
    expect(
      dict.dictionaryValues.length,
      "the value and id lists stay index-aligned",
    ).toBe(dict.dictionaryIds.length);
    expect(dict.uom, "a test with no unit renders n/a").toBe("n/a");
    expect(dict.significantDigits, "and so does its digit count").toBe("n/a");
    // referenceValue is "n/a" when no limit carries a dictionary normal, and
    // referenceId is then dropped entirely rather than sent as null.
    if (dict.referenceValue === "n/a") {
      expect("referenceId" in dict, "no reference id for an n/a reference").toBe(
        false,
      );
    }

    // The catalogue is sorted by test unit, then sample type, then panel, then
    // sort order — all three strings with String.compareTo.
    const keys = f.testCatBeanList.map(
      (b: any) => `${b.testUnit} ${b.sampleType} ${b.panel}`,
    );
    expect(
      keys.every((k: string, i: number) => i === 0 || keys[i - 1] <= k),
      "sorted by unit, sample type, panel",
    ).toBe(true);
  });

  test("GET TestModifyEntry?testSection — the other filter", async ({
    request,
  }) => {
    const f = await readJson(
      await request.get("rest/TestModifyEntry?testSection=36"),
      "TestModifyEntry?testSection=36",
    );
    const expected = query(
      `SELECT id::text FROM clinlims.test WHERE test_section_id = 36`,
    ).map((r) => r[0]);
    expect(
      f.testCatBeanList.map((b: any) => b.id).sort(),
      "every test in the section",
    ).toEqual([...expected].sort());
    // Both filters at once applies only the FIRST — `if (sampleType) … else if`.
    const both = await readJson(
      await request.get("rest/TestModifyEntry?sampleType=1&testSection=36"),
      "both filters",
    );
    const bySampleType = query(
      `SELECT DISTINCT t.id::text FROM clinlims.test t
         JOIN clinlims.sampletype_test st ON st.test_id = t.id
        WHERE st.sample_type_id = 1`,
    ).map((r) => r[0]);
    expect(
      both.testCatBeanList.map((b: any) => b.id).sort(),
      "sampleType wins",
    ).toEqual([...bySampleType].sort());
  });

  test("POST TestModifyEntry — a numeric modify rewrites eight columns and leaves the old result active", async ({
    request,
  }) => {
    expect(
      (await post(request, "rest/TestAdd", {
        jsonWad: JSON.stringify(createWad("Num", { loinc: "99999-7" })),
        loinc: "",
      })).status(),
      "fixture create",
    ).toBe(200);
    const [[testId, beforeResultID]] = query(
      `SELECT t.id::text, tr.id::text FROM clinlims.test t
         JOIN clinlims.test_result tr ON tr.test_id = t.id
        WHERE t.local_code = '${PREFIX}Num'`,
    );

    const res = await post(request, "rest/TestModifyEntry", {
      jsonWad: JSON.stringify({
        testId,
        testNameEnglish: `${PREFIX}Num2`,
        testNameFrench: `${PREFIX}Num2Fr`,
        testReportNameEnglish: `${PREFIX}Num2Rep`,
        testReportNameFrench: `${PREFIX}Num2RepFr`,
        testSection: "56",
        dictionaryReference: "",
        panels: [{ id: "2" }],
        uom: "5",
        loinc: "99999-8",
        resultType: "4",
        sampleTypes: [{ typeId: "3", tests: [{ id: "0" }] }],
        active: "N",
        orderable: "N",
        notifyResults: "Y",
        inLabOnly: "Y",
        antimicrobialResistance: "Y",
        lowValid: "10",
        highValid: "200",
        lowReportingRange: "20",
        highReportingRange: "180",
        lowCritical: "30",
        highCritical: "160",
        significantDigits: "3",
        resultLimits: [
          {
            gender: true,
            highAgeRange: "20",
            lowNormal: "7",
            highNormal: "70",
            lowNormalFemale: "8",
            highNormalFemale: "80",
          },
        ],
      }),
      loinc: "",
    });
    expect(res.status(), "modify").toBe(200);

    // The POST echoes the bound form and refills none of the lists.
    expect(Object.keys(await res.json()).sort(), "the POST document").toEqual(
      [
        "cancelAction",
        "cancelMethod",
        "formMethod",
        "formName",
        "jsonWad",
        "loinc",
        "nameEnglish",
        "nameFrench",
        "reportNameEnglish",
        "reportNameFrench",
        "submitOnCancel",
        "testId",
      ].sort(),
    );

    expect(
      query(
        `SELECT name, description, local_code, COALESCE(loinc, ''), is_active,
                orderable::text, notify_results::text, in_lab_only::text,
                antimicrobial_resistance::text, test_section_id::text, uom_id::text,
                is_reportable
           FROM clinlims.test WHERE id = ${Number(testId)}`,
      ),
      "the eight columns the update writes, plus a name it never names",
    ).toEqual([
      [
        // name followed the localization — Hibernate maps the column to a
        // DERIVED getter that reads it.
        `${PREFIX}Num2`,
        // description and local_code are NOT rewritten: the test still
        // describes itself by the name it was created with.
        `${PREFIX}Num(Urines)`,
        `${PREFIX}Num`,
        "99999-8",
        "N",
        "false",
        "true",
        "true",
        "true",
        "56",
        "5",
        "N",
      ],
    ]);

    // The two localizations are edited IN PLACE — same rows, new values.
    expect(
      query(
        `SELECT lv.locale, lv.value FROM clinlims.localization_value lv
          WHERE lv.localization_id = (SELECT name_localization_id FROM clinlims.test WHERE id = ${Number(testId)})
          ORDER BY lv.locale`,
      ),
      "the name localization was rewritten, not replaced",
    ).toEqual([
      ["en", `${PREFIX}Num2`],
      ["fr", `${PREFIX}Num2Fr`],
    ]);

    // The join rows are replaced wholesale.
    expect(
      query(
        `SELECT sample_type_id::text FROM clinlims.sampletype_test WHERE test_id = ${Number(testId)}`,
      ).map((r) => r[0]),
      "one sampletype_test, the submitted one",
    ).toEqual(["3"]);
    expect(
      query(
        `SELECT panel_id::text FROM clinlims.panel_item WHERE test_id = ${Number(testId)}`,
      ).map((r) => r[0]),
      "one panel_item, the submitted one",
    ).toEqual(["2"]);

    // TWO active results: the modify only ever INSERTS, and nothing deactivates
    // the old row for a numeric type. Every save leaves another one behind.
    expect(
      query(
        `SELECT id::text, tst_rslt_type, sort_order::text, is_active::text,
                COALESCE(significant_digits::text, '')
           FROM clinlims.test_result WHERE test_id = ${Number(testId)} ORDER BY id`,
      ),
      "the replaced result is still active beside its replacement",
    ).toEqual([
      [beforeResultID, "N", "1", "true", "2"],
      [String(Number(beforeResultID) + 1), "N", "1", "true", "3"],
    ]);

    // The limits are deleted and re-inserted, with the gendered entry split.
    expect(
      query(
        `SELECT COALESCE(gender, ''), min_age::text, max_age::text,
                low_normal::text, high_normal::text, low_valid::text, high_valid::text,
                low_critical::text, high_critical::text
           FROM clinlims.result_limits WHERE test_id = ${Number(testId)} ORDER BY id`,
      ),
      "an M row and an F row, both carrying the global bounds",
    ).toEqual([
      ["M", "0", "20", "7", "70", "10", "200", "30", "160"],
      ["F", "0", "20", "8", "80", "10", "200", "30", "160"],
    ]);

    // The old LOINC mapping is deactivated rather than deleted, and a new one
    // is inserted beside it.
    expect(
      query(
        `SELECT code, is_active FROM clinlims.test_terminology_mapping
          WHERE test_id = ${Number(testId)} ORDER BY code`,
      ),
      "the superseded mapping is kept, switched off",
    ).toEqual([
      ["99999-7", "N"],
      ["99999-8", "Y"],
    ]);

    // The component is UPDATED, not replaced — and it keeps the label it was
    // created with, so it still shows the test's old name.
    expect(
      query(
        `SELECT code, label, COALESCE(result_type, ''), COALESCE(uom_id::text, ''),
                COALESCE(significant_digits::text, '')
           FROM clinlims.test_result_component WHERE test_id = ${Number(testId)}`,
      ),
      "one component, re-pointed at the new unit and digits",
    ).toEqual([["PRIMARY", `${PREFIX}Num`, "N", "5", "3"]]);

    // Three history rows for the whole endpoint. The test updates, the
    // localization edits, the join-row churn, the result insert and the
    // component sync are all silent.
    const audit = query(
      `SELECT rt.name, h.activity
         FROM clinlims.history h JOIN clinlims.reference_tables rt ON rt.id = h.reference_table
        WHERE h.timestamp > now() - interval '2 minutes' ORDER BY h.id`,
    );
    expect(
      audit.slice(-3),
      "one delete and two inserts, all on RESULT_LIMITS",
    ).toEqual([
      ["RESULT_LIMITS", "D"],
      ["RESULT_LIMITS", "I"],
      ["RESULT_LIMITS", "I"],
    ]);

    // The delete payload carries the WHOLE row it removed, in the entity's
    // declared-field order, with Java's own double rendering.
    const [[payload]] = query(
      `SELECT replace(encode(h.changes, 'escape'), chr(10), '')
         FROM clinlims.history h JOIN clinlims.reference_tables rt ON rt.id = h.reference_table
        WHERE rt.name = 'RESULT_LIMITS' AND h.activity = 'D'
          AND h.timestamp > now() - interval '2 minutes'`,
    );
    expect(payload, "the deleted limit, field by field").toContain(
      `<testId>${testId}</testId><resultTypeId>4</resultTypeId><minAge>0.0</minAge>` +
        `<highCritical>80.0</highCritical><lowCritical>3.0</lowCritical><maxAge>10.0</maxAge>` +
        `<lowNormal>5.0</lowNormal><highNormal>50.0</highNormal><lowValid>1.0</lowValid>` +
        `<highValid>100.0</highValid><lowReportingRange>2.0</lowReportingRange>` +
        `<highReportingRange>90.0</highReportingRange><alwaysValidate>false</alwaysValidate>`,
    );
    // gender is absent because it was null — getChanges emits only the fields
    // that differ from a blank object.
    expect(payload, "no gender element").not.toContain("<gender>");
  });

  test("POST TestModifyEntry — a dictionary modify deactivates what it replaces", async ({
    request,
  }) => {
    expect(
      (await post(request, "rest/TestAdd", {
        jsonWad: JSON.stringify(
          createWad("Dict", {
            resultType: "2",
            dictionaryReference: "3",
            panels: [],
            uom: "",
            dictionary: [
              { id: "1", qualified: "Y" },
              { id: "2", qualified: "N" },
            ],
            defaultTestResult: "1",
          }),
        ),
        loinc: "",
      })).status(),
      "fixture create",
    ).toBe(200);
    const [[testId]] = query(
      `SELECT id::text FROM clinlims.test WHERE local_code = '${PREFIX}Dict'`,
    );

    // Sample type 30 has no sampletype_panel row, which is the only way to
    // reach the insert branch — the write happens once per sample type, ever.
    const panelsBefore = query(
      `SELECT id::text FROM clinlims.sampletype_panel WHERE sample_type_id = 30`,
    ).map((r) => r[0]);
    restore.push(() => {
      exec(
        `DELETE FROM clinlims.sampletype_panel
          WHERE sample_type_id = 30${panelsBefore.length ? ` AND id NOT IN (${panelsBefore.map((p) => `'${p}'`).join(",")})` : ""}`,
      );
    });

    const res = await post(request, "rest/TestModifyEntry", {
      jsonWad: JSON.stringify({
        testId,
        testNameEnglish: `${PREFIX}Dict`,
        testNameFrench: `${PREFIX}Dict2Fr`,
        testReportNameEnglish: `${PREFIX}Dict2Rep`,
        testReportNameFrench: `${PREFIX}Dict2RepFr`,
        testSection: "36",
        dictionaryReference: "4",
        panels: [{ id: "1" }],
        uom: "",
        loinc: "",
        resultType: "2",
        sampleTypes: [{ typeId: "30", tests: [{ id: "0" }] }],
        active: "Y",
        orderable: "Y",
        notifyResults: "N",
        inLabOnly: "N",
        antimicrobialResistance: "N",
        dictionary: [
          { id: "3", qualified: "N" },
          { id: "4", qualified: "Y" },
        ],
        defaultTestResult: "4",
      }),
      loinc: "",
    });
    expect(res.status(), "modify").toBe(200);

    // The old options are switched OFF, not deleted, and the new ones start at
    // 10 again.
    const results = query(
      `SELECT id::text, sort_order::text, is_active::text, value, is_quantifiable::text
         FROM clinlims.test_result WHERE test_id = ${Number(testId)} ORDER BY id`,
    );
    expect(
      results.map((r) => r.slice(1)),
      "two retired options and two live ones",
    ).toEqual([
      ["10", "false", "1", "true"],
      ["20", "false", "2", "false"],
      ["10", "true", "3", "false"],
      ["20", "true", "4", "true"],
    ]);

    // The default follows the submitted option, and it is the LAST insert here.
    expect(
      query(
        `SELECT COALESCE(default_test_result_id::text, '') FROM clinlims.test WHERE id = ${Number(testId)}`,
      )[0][0],
      "the default points at the new option",
    ).toBe(results[3][0]);

    // One limit, carrying the new dictionary reference.
    expect(
      query(
        `SELECT COALESCE(normal_dictionary_id::text, '') FROM clinlims.result_limits
          WHERE test_id = ${Number(testId)}`,
      ).map((r) => r[0]),
      "the reference was replaced",
    ).toEqual(["4"]);

    // A blank loinc leaves the terminology table alone.
    expect(
      query(
        `SELECT count(*)::text FROM clinlims.test_terminology_mapping WHERE test_id = ${Number(testId)}`,
      )[0][0],
      "no mapping for a blank code",
    ).toBe("0");

    // The sample type had no panel at all, so attaching one wrote the join row
    // the screen never mentions.
    expect(
      query(
        `SELECT panel_id::text FROM clinlims.sampletype_panel WHERE sample_type_id = 30`,
      ).map((r) => r[0]),
      "a sampletype_panel row appeared",
    ).toEqual([...panelsBefore.map(() => expect.anything()), "1"].slice(-1));

    // The component takes the newest active result's type and digits.
    expect(
      query(
        `SELECT COALESCE(result_type, ''), COALESCE(uom_id::text, ''),
                COALESCE(significant_digits::text, '')
           FROM clinlims.test_result_component WHERE test_id = ${Number(testId)}`,
      ),
      "the PRIMARY component, resynced",
    ).toEqual([["D", "", ""]]);
  });

  test("POST TestModifyEntry — an unresolvable sample type writes nothing but the deletes", async ({
    request,
  }) => {
    expect(
      (await post(request, "rest/TestAdd", {
        jsonWad: JSON.stringify(createWad("Skip")),
        loinc: "",
      })).status(),
      "fixture create",
    ).toBe(200);
    const [[testId]] = query(
      `SELECT id::text FROM clinlims.test WHERE local_code = '${PREFIX}Skip'`,
    );

    // createTestSets skips a set whose sample type does not resolve, so there
    // are no sets left — but the three deletes run BEFORE the loop and are not
    // undone. The test comes back stripped of its joins and its limits.
    const res = await post(request, "rest/TestModifyEntry", {
      jsonWad: JSON.stringify({
        testId,
        testNameEnglish: `${PREFIX}Skip`,
        testNameFrench: `${PREFIX}SkipFr`,
        testReportNameEnglish: `${PREFIX}SkipRep`,
        testReportNameFrench: `${PREFIX}SkipRepFr`,
        testSection: "36",
        dictionaryReference: "",
        panels: [],
        uom: "",
        loinc: "",
        resultType: "5",
        sampleTypes: [{ typeId: "999999", tests: [{ id: "0" }] }],
        active: "Y",
        orderable: "Y",
        notifyResults: "N",
        inLabOnly: "N",
        antimicrobialResistance: "N",
      }),
      loinc: "",
    });
    expect(res.status(), "modify").toBe(200);

    expect(
      query(
        `SELECT count(*)::text FROM clinlims.sampletype_test WHERE test_id = ${Number(testId)}`,
      )[0][0],
      "the sample type link is gone and nothing replaced it",
    ).toBe("0");
    expect(
      query(
        `SELECT count(*)::text FROM clinlims.panel_item WHERE test_id = ${Number(testId)}`,
      )[0][0],
      "so is the panel link",
    ).toBe("0");
    expect(
      query(
        `SELECT count(*)::text FROM clinlims.result_limits WHERE test_id = ${Number(testId)}`,
      )[0][0],
      "and so are the limits",
    ).toBe("0");
    // The test row itself is untouched — every write to it lives inside the
    // loop that never ran.
    expect(
      query(
        `SELECT is_active, test_section_id::text FROM clinlims.test WHERE id = ${Number(testId)}`,
      ),
      "the test kept the values the create gave it",
    ).toEqual([["Y", "36"]]);
  });
});
