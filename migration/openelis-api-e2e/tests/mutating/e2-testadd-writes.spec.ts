/**
 * e2 — TestAdd, the widest write in the wave.
 *
 * One submission creates ONE TEST PER SAMPLE TYPE it names, and each of those
 * drags a fan of rows behind it: two shared localizations, a test, a
 * sampletype_test, a panel_item per panel, a test_result per result option, a
 * result_limits row per range, a LOINC terminology mapping and the PRIMARY
 * test_result_component the new editor scopes everything by.
 *
 * Three side effects are worth naming because the screen does not: assigning to
 * an inactive test section or panel turns it back ON, and the sample type
 * follows the new test's OWN active flag — so creating an inactive test
 * DEACTIVATES a live sample type.
 *
 * The audit is far narrower than the write. Measured: one 'I' for the test, one
 * 'I' per result limit, and a 'U' for the sample type only when its flag really
 * moved. Nothing else is recorded, including from tables flagged
 * keep_history='Y'.
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

/** Every row the create leaves behind, removed in FK order. */
function dropTests(namePrefix: string) {
  const where = `name LIKE '${namePrefix}%'`;
  const ids = `SELECT id FROM clinlims.test WHERE ${where}`;
  const locs = `SELECT name_localization_id AS lid FROM clinlims.test WHERE ${where}
                UNION SELECT reporting_name_localization_id FROM clinlims.test WHERE ${where}`;
  const locIds = query(locs)
    .map((r) => r[0])
    .filter((v) => v !== "");
  exec(`UPDATE clinlims.test SET default_test_result_id = NULL WHERE ${where}`);
  // result_limits carries the FK onto test_result_component, so it goes first.
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

const PREFIX = "E2TA";

test.describe("e2 — TestAdd", () => {
  let restore: (() => void)[] = [];

  test.beforeEach(() => {
    restore = [];
  });

  test.afterEach(async ({ request }) => {
    dropTests(PREFIX);
    for (const r of restore) r();

    // Then one throwaway create, purely to RESYNC Java's caches.
    //
    // postTestAdd ends by calling refreshList on SAMPLE_TYPE_ACTIVE/_INACTIVE,
    // PANELS*, and TEST_SECTION_ACTIVE/_INACTIVE/_BY_NAME. Those lists are
    // in-memory and refresh ONLY on an application write, so a spec that turns
    // a section on through the API and then turns it back off with SQL leaves
    // Java serving a section it no longer has — which is a readonly gate
    // failure two suites away, not a failure here. One more write through the
    // endpoint rebuilds every list from the restored rows.
    //
    // Its own arguments are chosen to change nothing: an already-active sample
    // type, an already-active section, no panels.
    await post(request, "rest/TestAdd", {
      jsonWad: JSON.stringify({
        testNameEnglish: `${PREFIX}Resync`,
        testNameFrench: `${PREFIX}Resync`,
        testReportNameEnglish: `${PREFIX}Resync`,
        testReportNameFrench: `${PREFIX}Resync`,
        testSection: "36",
        dictionaryReference: "",
        panels: [],
        uom: "",
        loinc: "",
        resultType: "1",
        sampleTypes: [{ typeId: "1", tests: [{ id: "0" }] }],
        active: "Y",
        orderable: "Y",
        notifyResults: "N",
        inLabOnly: "N",
        antimicrobialResistance: "N",
      }),
      loinc: "",
    });
    dropTests(PREFIX);

    exec(
      `DELETE FROM clinlims.history WHERE timestamp > now() - interval '2 minutes'`,
    );
  });

  test("GET TestAdd — eight lists, and no loinc key at all", async ({
    request,
  }) => {
    const f = await readJson(await request.get("rest/TestAdd"), "TestAdd");

    expect(f.formName, "formName").toBe("testAddForm");
    expect(f.jsonWad, "jsonWad starts empty").toBe("");
    // form.setLoinc(new Test().getLoinc()) stores null, and the form is
    // NON_NULL — so the declared field does NOT appear on the blank form.
    expect("loinc" in f, "loinc key absent").toBe(false);

    expect(Object.keys(f).sort(), "the whole document").toEqual(
      [
        "ageRangeList",
        "cancelAction",
        "cancelMethod",
        "dictionaryList",
        "formMethod",
        "formName",
        "groupedDictionaryList",
        "jsonWad",
        "labUnitList",
        "panelList",
        "resultTypeList",
        "sampleTypeList",
        "submitOnCancel",
        "uomList",
      ].sort(),
    );

    // Six of the seven result types: createLocalizedResultTypeList branches on
    // the DESCRIPTION and has no branch for "Titer", so id 3 is dropped.
    expect(f.resultTypeList, "result types, Titer missing").toEqual([
      { id: "1", value: "Free text" },
      { id: "2", value: "Select List" },
      { id: "4", value: "Numeric" },
      { id: "5", value: "Alphanumeric" },
      { id: "6", value: "Multiselect" },
      { id: "7", value: "Cascading multiselect" },
    ]);

    // The age ranges are INVERTED against every other pair here: the id is the
    // age in months and the value is the localized name. "Infinity" sorts last.
    expect(f.ageRangeList, "predefined age ranges").toEqual([
      { id: "1", value: "Newborn" },
      { id: "12", value: "Infant" },
      { id: "60", value: "Young child" },
      { id: "168", value: "Child" },
      { id: "Infinity", value: "Adult" },
    ]);

    // sampleTypeList and labUnitList are CONCATENATIONS — the active list, then
    // the inactive one, with no re-sort.
    //
    // Asserted as SETS, and only as sets. Two separate things put the order and
    // even the split out of reach. The source queries order by a sort_order
    // with heavy ties (three sample types at 0, thirteen sections at
    // 2147483647) and add no tiebreak, so the tie order is the query plan's and
    // moves as rows are updated. And Java serves both halves from a
    // DisplayListService CACHE that only application writes refresh — so a row
    // this suite flips back with its own SQL cleanup stays on whichever side
    // Java last saw it, while the Go port reads live. What both stacks do hold
    // is that every candidate row appears exactly once.
    const allTypes = query(
      `SELECT id::text FROM clinlims.type_of_sample WHERE domain = 'H'`,
    ).map((r) => r[0]);
    expect(
      f.sampleTypeList.map((p: any) => p.id).sort(),
      "every human sample type, once",
    ).toEqual([...allTypes].sort());

    const allSections = query(`SELECT id::text FROM clinlims.test_section`).map(
      (r) => r[0],
    );
    expect(
      f.labUnitList.map((p: any) => p.id).sort(),
      "every test section, once",
    ).toEqual([...allSections].sort());

    // uomList is EVERY unit, active or not — getAll(), and the two
    // UNIT_OF_MEASURE_ACTIVE / _INACTIVE list types share this one builder.
    expect(
      f.uomList.length,
      "every unit of measure",
    ).toBe(Number(query(`SELECT count(*)::text FROM clinlims.unit_of_measure`)[0][0]));

    // dictionaryList is seven categories concatenated then sorted by the
    // LOWERCASED value. Assert the sort rather than the contents: Java serves
    // it from a DisplayListService cache that only application writes refresh,
    // so a row deleted by a spec's own SQL cleanup lingers in Java's copy while
    // the Go port reads live.
    const values = f.dictionaryList.map((p: any) => p.value.toLowerCase());
    expect(
      [...values].every((v, i) => i === 0 || values[i - 1] <= v),
      "dictionaryList is sorted by the lowercased value",
    ).toBe(true);

    // groupedDictionaryList is sorted by SIZE, ascending. Equal sizes keep the
    // iteration order of the HashSet the groups were deduplicated in, which is
    // an accident of String.hashCode — so the sizes and the group contents are
    // the contract, and the order among equal sizes is not.
    const sizes = f.groupedDictionaryList.map((g: any[]) => g.length);
    expect(
      sizes.every((n: number, i: number) => i === 0 || sizes[i - 1] <= n),
      "groups are sorted by size",
    ).toBe(true);
    const expectedGroups = new Set(
      query(
        `SELECT string_agg(value, ',' ORDER BY sort_order NULLS FIRST, id)
           FROM clinlims.test_result
          WHERE tst_rslt_type IN ('D','M','C') AND COALESCE(btrim(value), '') <> ''
          GROUP BY test_id::text`,
      ).map((r) => r[0]),
    );
    for (const group of f.groupedDictionaryList) {
      expect(
        expectedGroups.has(group.map((p: any) => p.id).join(",")),
        `group ${group.map((p: any) => p.id).join(",")} is one a test actually offers`,
      ).toBe(true);
    }
  });

  test("POST TestAdd — a numeric test, and everything it drags with it", async ({
    request,
  }) => {
    const wad = {
      testNameEnglish: `${PREFIX}Num`,
      testNameFrench: `${PREFIX}NumFr`,
      testReportNameEnglish: `${PREFIX}NumRep`,
      testReportNameFrench: `${PREFIX}NumRepFr`,
      testSection: "36",
      dictionaryReference: "",
      panels: [{ id: "1" }],
      uom: "1",
      loinc: "99999-1",
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
        {
          gender: true,
          highAgeRange: "10",
          lowNormal: "5",
          highNormal: "50",
          lowNormalFemale: "6",
          highNormalFemale: "60",
        },
      ],
    };

    const res = await post(request, "rest/TestAdd", {
      jsonWad: JSON.stringify(wad),
      loinc: "",
    });
    expect(res.status(), "create").toBe(200);

    // The POST echoes the BOUND form and refills none of the eight lists, so
    // they all vanish from the response. loinc IS present here, because the
    // body carried it.
    const body = await res.json();
    expect(Object.keys(body).sort(), "the POST document").toEqual(
      [
        "cancelAction",
        "cancelMethod",
        "formMethod",
        "formName",
        "jsonWad",
        "loinc",
        "submitOnCancel",
      ].sort(),
    );
    expect(body.jsonWad, "jsonWad comes back verbatim").toBe(
      JSON.stringify(wad),
    );

    const [row] = query(
      `SELECT id::text, name, description, local_code, COALESCE(loinc, ''), is_active,
              orderable::text, notify_results::text, in_lab_only::text,
              antimicrobial_resistance::text, is_reportable, test_section_id::text,
              uom_id::text, sort_order::text, domain, normalized_description,
              (guid IS NOT NULL)::text, COALESCE(default_test_result_id::text, '')
         FROM clinlims.test WHERE name = '${PREFIX}Num'`,
    );
    expect(row, "the test exists").toBeTruthy();
    const [
      testId,
      name,
      description,
      localCode,
      loinc,
      isActive,
      orderable,
      notifyResults,
      inLabOnly,
      amr,
      isReportable,
      sectionId,
      uomId,
      sortOrder,
      domain,
      normalized,
      hasGuid,
      defaultResult,
    ] = row;

    expect(name, "name is the English name").toBe(`${PREFIX}Num`);
    // The description is the name with the sample type in brackets — the source
    // of every doubled test name the rest of this wave reads.
    expect(description, "description carries the sample type").toBe(
      `${PREFIX}Num(Urines)`,
    );
    expect(localCode, "local_code is the same name again").toBe(`${PREFIX}Num`);
    expect(loinc, "loinc as submitted").toBe("99999-1");
    expect(isActive, "is_active is the raw Y/N string").toBe("Y");
    expect(orderable, "orderable").toBe("true");
    expect(notifyResults, "notifyResults").toBe("false");
    expect(inLabOnly, "inLabOnly").toBe("false");
    expect(amr, "antimicrobialResistance").toBe("false");
    // is_reportable is hard-coded 'N'; the form never offers it.
    expect(isReportable, "is_reportable is always N").toBe("N");
    expect(sectionId, "test section").toBe("36");
    expect(uomId, "unit of measure").toBe("1");
    // The new test takes the index of the "0" entry in its sample type's list.
    expect(sortOrder, "sort order is the index marked 0").toBe("0");
    expect(domain, "domain defaults to CLINICAL").toBe("CLINICAL");
    expect(hasGuid, "a guid was generated").toBe("true");
    // normalized_description comes from a BEFORE INSERT trigger, not the app:
    // the name and the bracketed sample type, accents stripped, lowercased.
    expect(normalized, "the trigger normalised the description").toBe(
      `${PREFIX}num`.toLowerCase() + "urines",
    );
    // Only a dictionary option can be a default, so a numeric test has none.
    expect(defaultResult, "no default result for a numeric test").toBe("");

    // Two localizations, with the LocalizationType labels as descriptions.
    const locs = query(
      `SELECT l.description, lv.locale, lv.value
         FROM clinlims.localization l
         JOIN clinlims.localization_value lv ON lv.localization_id = l.id
        WHERE l.id IN (SELECT name_localization_id FROM clinlims.test WHERE id = ${Number(testId)}
                       UNION SELECT reporting_name_localization_id FROM clinlims.test WHERE id = ${Number(testId)})
        ORDER BY l.description, lv.locale`,
    );
    expect(locs, "two localizations, en and fr each").toEqual([
      ["test name", "en", `${PREFIX}Num`],
      ["test name", "fr", `${PREFIX}NumFr`],
      ["test report name", "en", `${PREFIX}NumRep`],
      ["test report name", "fr", `${PREFIX}NumRepFr`],
    ]);

    expect(
      query(
        `SELECT sample_type_id::text FROM clinlims.sampletype_test WHERE test_id = ${Number(testId)}`,
      ).map((r) => r[0]),
      "one sampletype_test",
    ).toEqual(["1"]);

    expect(
      query(
        `SELECT panel_id::text, COALESCE(sort_order::text, '') FROM clinlims.panel_item WHERE test_id = ${Number(testId)}`,
      ),
      "one panel_item, with no sort order",
    ).toEqual([["1", ""]]);

    // One test_result for a numeric type, at sort order 1.
    expect(
      query(
        `SELECT tst_rslt_type, sort_order::text, is_active::text,
                COALESCE(value, ''), is_quantifiable::text, significant_digits::text,
                is_normal::text, (component_id IS NOT NULL)::text
           FROM clinlims.test_result WHERE test_id = ${Number(testId)}`,
      ),
      "the single numeric result",
    ).toEqual([["N", "1", "true", "", "false", "2", "false", "true"]]);

    // Two result_limits: the M row, and the F row extractLimits splits off
    // because `gender` was true. Both carry the SAME global bounds — lowValid,
    // highValid, the reporting range and the criticals are stamped onto every
    // row, not read per entry.
    expect(
      query(
        `SELECT test_result_type_id::text, COALESCE(gender, ''), min_age::text, max_age::text,
                low_normal::text, high_normal::text, low_valid::text, high_valid::text,
                low_reporting_range::text, high_reporting_range::text,
                low_critical::text, high_critical::text,
                COALESCE(normal_dictionary_id::text, ''), always_validate::text,
                (component_id IS NOT NULL)::text
           FROM clinlims.result_limits WHERE test_id = ${Number(testId)} ORDER BY id`,
      ),
      "an M row and an F row",
    ).toEqual([
      ["4", "M", "0", "10", "5", "50", "1", "100", "2", "90", "3", "80", "", "false", "true"],
      ["4", "F", "0", "10", "6", "60", "1", "100", "2", "90", "3", "80", "", "false", "true"],
    ]);

    // A non-blank loinc syncs a terminology mapping; the relationship is always
    // SAME_AS on a fresh test.
    expect(
      query(
        `SELECT source, code, relationship, is_active
           FROM clinlims.test_terminology_mapping WHERE test_id = ${Number(testId)}`,
      ),
      "the LOINC mapping",
    ).toEqual([["LOINC", "99999-1", "SAME_AS", "Y"]]);

    // The PRIMARY component, whose label is the test NAME and whose result type
    // and significant digits come from the newest result.
    expect(
      query(
        `SELECT code, label, display_order::text, COALESCE(result_type, ''),
                COALESCE(uom_id::text, ''), COALESCE(significant_digits::text, ''),
                allow_multiple_readings::text, is_active
           FROM clinlims.test_result_component WHERE test_id = ${Number(testId)}`,
      ),
      "the PRIMARY component",
    ).toEqual([["PRIMARY", `${PREFIX}Num`, "0", "N", "1", "2", "false", "Y"]]);

    // The audit: one 'I' for the test and one per limit, and nothing else. The
    // localizations, the join row, the panel item, the result, the mapping and
    // the component are all silent — and TEST_RESULT, SAMPLETYPE_TEST and
    // PANEL_ITEM are all flagged keep_history='Y'.
    const limitIds = query(
      `SELECT id::text FROM clinlims.result_limits WHERE test_id = ${Number(testId)} ORDER BY id`,
    ).map((r) => r[0]);
    expect(
      query(
        `SELECT rt.name, h.reference_id, h.activity, (h.changes IS NULL)::text
           FROM clinlims.history h JOIN clinlims.reference_tables rt ON rt.id = h.reference_table
          WHERE h.timestamp > now() - interval '2 minutes' ORDER BY h.id`,
      ),
      "the whole audit trail for the create",
    ).toEqual([
      ["TEST", testId, "I", "true"],
      ["RESULT_LIMITS", limitIds[0], "I", "true"],
      ["RESULT_LIMITS", limitIds[1], "I", "true"],
    ]);
  });

  test("POST TestAdd — a dictionary test sets a default result and deactivates its sample type", async ({
    request,
  }) => {
    const [[typeActive]] = query(
      `SELECT is_active::text FROM clinlims.type_of_sample WHERE id = 2`,
    );
    const [[typeUpdated]] = query(
      `SELECT lastupdated::text FROM clinlims.type_of_sample WHERE id = 2`,
    );
    restore.push(() => {
      exec(
        `UPDATE clinlims.type_of_sample SET is_active = ${typeActive}, lastupdated = '${typeUpdated}' WHERE id = 2`,
      );
    });

    const wad = {
      testNameEnglish: `${PREFIX}Dict`,
      testNameFrench: `${PREFIX}DictFr`,
      testReportNameEnglish: `${PREFIX}DictRep`,
      testReportNameFrench: `${PREFIX}DictRepFr`,
      testSection: "36",
      dictionaryReference: "3",
      panels: [],
      uom: "",
      loinc: "",
      resultType: "2",
      sampleTypes: [{ typeId: "2", tests: [{ id: "0" }] }],
      active: "N",
      orderable: "N",
      notifyResults: "Y",
      inLabOnly: "Y",
      antimicrobialResistance: "Y",
      dictionary: [
        { id: "1", qualified: "Y" },
        { id: "2", qualified: "N" },
      ],
      defaultTestResult: "2",
    };

    const res = await post(request, "rest/TestAdd", {
      jsonWad: JSON.stringify(wad),
      loinc: "",
    });
    expect(res.status(), "create").toBe(200);

    const [[testId, isActive, orderable, notify, inLab, amr, uomId, defaultResult]] =
      query(
        `SELECT id::text, is_active, orderable::text, notify_results::text,
                in_lab_only::text, antimicrobial_resistance::text,
                COALESCE(uom_id::text, ''), COALESCE(default_test_result_id::text, '')
           FROM clinlims.test WHERE name = '${PREFIX}Dict'`,
      );
    expect(isActive, "created inactive").toBe("N");
    expect(orderable, "orderable follows the Y test, not the string").toBe("false");
    expect(notify, "notifyResults").toBe("true");
    expect(inLab, "inLabOnly").toBe("true");
    expect(amr, "antimicrobialResistance").toBe("true");
    expect(uomId, "a blank uom leaves the column NULL").toBe("");

    // One test_result per dictionary option, at 10, 20, 30… and `value` holds
    // the DICTIONARY id rather than a result string.
    const results = query(
      `SELECT id::text, tst_rslt_type, sort_order::text, value, is_quantifiable::text,
              COALESCE(significant_digits::text, '')
         FROM clinlims.test_result WHERE test_id = ${Number(testId)} ORDER BY sort_order`,
    );
    expect(
      results.map((r) => r.slice(1)),
      "two options, ten apart",
    ).toEqual([
      ["D", "10", "1", "true", ""],
      ["D", "20", "2", "false", ""],
    ]);

    // addTests only mutates the in-memory test when an option is the default —
    // but it is a MANAGED entity, so the flush writes the column anyway.
    expect(defaultResult, "the default option is written back onto the test").toBe(
      results[1][0],
    );

    // A dictionary test gets ONE limit, and every numeric column keeps the
    // ENTITY default. Both criticals default POSITIVE infinity, which is not
    // what the column default says for low_critical.
    expect(
      query(
        `SELECT test_result_type_id::text, COALESCE(gender, ''), min_age::text, max_age::text,
                low_normal::text, high_normal::text, low_valid::text, high_valid::text,
                low_reporting_range::text, high_reporting_range::text,
                low_critical::text, high_critical::text, normal_dictionary_id::text
           FROM clinlims.result_limits WHERE test_id = ${Number(testId)}`,
      ),
      "the dictionary reference limit",
    ).toEqual([
      [
        "2",
        "",
        "0",
        "Infinity",
        "-Infinity",
        "Infinity",
        "-Infinity",
        "Infinity",
        "-Infinity",
        "Infinity",
        "Infinity",
        "Infinity",
        "3",
      ],
    ]);

    // A blank loinc writes NO terminology row: syncLegacyLoinc returns early.
    expect(
      query(
        `SELECT count(*)::text FROM clinlims.test_terminology_mapping WHERE test_id = ${Number(testId)}`,
      )[0][0],
      "no LOINC mapping for a blank code",
    ).toBe("0");

    // The component takes its result type and digits from the NEWEST result by
    // id, which for a dictionary variant is the LAST option submitted.
    expect(
      query(
        `SELECT result_type, COALESCE(uom_id::text, ''), COALESCE(significant_digits::text, '')
           FROM clinlims.test_result_component WHERE test_id = ${Number(testId)}`,
      ),
      "the PRIMARY component",
    ).toEqual([["D", "", ""]]);

    // active="N" turned the SAMPLE TYPE off. Nothing on the screen says so.
    expect(
      query(`SELECT is_active::text FROM clinlims.type_of_sample WHERE id = 2`)[0][0],
      "the sample type followed the test's active flag",
    ).toBe("false");

    // And that change IS audited, with the value it replaced.
    expect(
      query(
        `SELECT h.reference_id, h.activity, replace(encode(h.changes, 'escape'), chr(10), '<NL>')
           FROM clinlims.history h JOIN clinlims.reference_tables rt ON rt.id = h.reference_table
          WHERE rt.name = 'TYPE_OF_SAMPLE' AND h.timestamp > now() - interval '2 minutes'`,
      ),
      "the sample type update carries the OLD flag",
    ).toEqual([["2", "U", "<isActive>true</isActive><NL>"]]);
  });

  test("POST TestAdd — a Titer test gets no results at all", async ({
    request,
  }) => {
    // "T" is in neither branch of createTestResults: not "AR", not "N", not
    // "DMC". The test is created with an empty result set and no limits.
    const res = await post(request, "rest/TestAdd", {
      jsonWad: JSON.stringify({
        testNameEnglish: `${PREFIX}Titer`,
        testNameFrench: `${PREFIX}TiterFr`,
        testReportNameEnglish: `${PREFIX}TiterRep`,
        testReportNameFrench: `${PREFIX}TiterRepFr`,
        testSection: "36",
        dictionaryReference: "",
        panels: [],
        uom: "",
        loinc: "",
        resultType: "3",
        sampleTypes: [{ typeId: "1", tests: [{ id: "0" }] }],
        active: "Y",
        orderable: "Y",
        notifyResults: "N",
        inLabOnly: "N",
        antimicrobialResistance: "N",
      }),
      loinc: "",
    });
    expect(res.status(), "create").toBe(200);

    const [[testId]] = query(
      `SELECT id::text FROM clinlims.test WHERE name = '${PREFIX}Titer'`,
    );
    expect(
      query(
        `SELECT count(*)::text FROM clinlims.test_result WHERE test_id = ${Number(testId)}`,
      )[0][0],
      "no test_result rows",
    ).toBe("0");
    expect(
      query(
        `SELECT count(*)::text FROM clinlims.result_limits WHERE test_id = ${Number(testId)}`,
      )[0][0],
      "and no limits",
    ).toBe("0");
    // The component is still written, with a NULL result type — there is no
    // result to read one from.
    expect(
      query(
        `SELECT code, COALESCE(result_type, '') FROM clinlims.test_result_component WHERE test_id = ${Number(testId)}`,
      ),
      "a PRIMARY with no result type",
    ).toEqual([["PRIMARY", ""]]);
  });

  test("POST TestAdd — one submission, one test per sample type, and the siblings re-sort", async ({
    request,
  }) => {
    // Two existing tests under sample type 1, whose sort orders the submission
    // rewrites to their positions in the list.
    const siblings = query(
      `SELECT t.id::text, COALESCE(t.sort_order::text, '')
         FROM clinlims.test t JOIN clinlims.sampletype_test st ON st.test_id = t.id
        WHERE st.sample_type_id = 1 ORDER BY t.id LIMIT 2`,
    );
    expect(siblings.length, "the fixture has two tests to reorder").toBe(2);
    restore.push(() => {
      for (const [id, sortOrder] of siblings) {
        exec(
          `UPDATE clinlims.test SET sort_order = ${sortOrder === "" ? "NULL" : Number(sortOrder)} WHERE id = ${Number(id)}`,
        );
      }
    });

    const res = await post(request, "rest/TestAdd", {
      jsonWad: JSON.stringify({
        testNameEnglish: `${PREFIX}Multi`,
        testNameFrench: `${PREFIX}MultiFr`,
        testReportNameEnglish: `${PREFIX}MultiRep`,
        testReportNameFrench: `${PREFIX}MultiRepFr`,
        testSection: "36",
        dictionaryReference: "",
        panels: [],
        uom: "",
        loinc: "",
        resultType: "1",
        sampleTypes: [
          // The new test sits SECOND under sample type 1 …
          { typeId: "1", tests: [{ id: siblings[0][0] }, { id: "0" }, { id: siblings[1][0] }] },
          // … and an id naming no sample type is skipped whole.
          { typeId: "999999", tests: [{ id: "0" }] },
          { typeId: "3", tests: [{ id: "0" }] },
        ],
        active: "Y",
        orderable: "Y",
        notifyResults: "N",
        inLabOnly: "N",
        antimicrobialResistance: "N",
      }),
      loinc: "",
    });
    expect(res.status(), "create").toBe(200);

    // TWO tests, not three: the unknown sample type produced nothing.
    const created = query(
      `SELECT id::text, description, sort_order::text FROM clinlims.test
        WHERE name = '${PREFIX}Multi' ORDER BY id`,
    );
    expect(created.length, "one test per RESOLVABLE sample type").toBe(2);
    expect(
      created.map((r) => r[1]),
      "each description names its own sample type",
    ).toEqual([`${PREFIX}Multi(Urines)`, `${PREFIX}Multi(Plasma)`]);
    expect(created[0][2], "the first sits at index 1, where the 0 was").toBe("1");
    expect(created[1][2], "the second at index 0").toBe("0");

    // The two localizations are written ONCE and shared by both tests.
    const locIds = query(
      `SELECT DISTINCT name_localization_id::text, reporting_name_localization_id::text
         FROM clinlims.test WHERE name = '${PREFIX}Multi'`,
    );
    expect(locIds.length, "both tests point at the same pair").toBe(1);

    // The siblings took the indexes the list gave them.
    expect(
      query(
        `SELECT sort_order::text FROM clinlims.test WHERE id = ${Number(siblings[0][0])}`,
      )[0][0],
      "the first sibling moved to index 0",
    ).toBe("0");
    expect(
      query(
        `SELECT sort_order::text FROM clinlims.test WHERE id = ${Number(siblings[1][0])}`,
      )[0][0],
      "the third entry moved to index 2",
    ).toBe("2");

    // Both tests are text-only ("R"), so each got exactly one result row.
    expect(
      query(
        `SELECT count(*)::text FROM clinlims.test_result
          WHERE test_id IN (SELECT id FROM clinlims.test WHERE name = '${PREFIX}Multi')`,
      )[0][0],
      "one result each",
    ).toBe("2");
  });

  test("POST TestAdd — naming an inactive test section and panel turns them back on", async ({
    request,
  }) => {
    const [[sectionId, sectionActive]] = query(
      `SELECT id::text, is_active FROM clinlims.test_section WHERE is_active = 'N' ORDER BY id LIMIT 1`,
    );
    const [[panelId, panelActive]] = query(
      `SELECT id::text, is_active FROM clinlims.panel ORDER BY id LIMIT 1`,
    );
    exec(`UPDATE clinlims.panel SET is_active = 'N' WHERE id = ${Number(panelId)}`);
    restore.push(() => {
      exec(
        `UPDATE clinlims.test_section SET is_active = '${sectionActive}' WHERE id = ${Number(sectionId)}`,
      );
      exec(
        `UPDATE clinlims.panel SET is_active = '${panelActive}' WHERE id = ${Number(panelId)}`,
      );
    });

    const res = await post(request, "rest/TestAdd", {
      jsonWad: JSON.stringify({
        testNameEnglish: `${PREFIX}Wake`,
        testNameFrench: `${PREFIX}WakeFr`,
        testReportNameEnglish: `${PREFIX}WakeRep`,
        testReportNameFrench: `${PREFIX}WakeRepFr`,
        testSection: sectionId,
        dictionaryReference: "",
        panels: [{ id: panelId }],
        uom: "",
        loinc: "",
        resultType: "5",
        sampleTypes: [{ typeId: "1", tests: [{ id: "0" }] }],
        active: "Y",
        orderable: "Y",
        notifyResults: "N",
        inLabOnly: "N",
        antimicrobialResistance: "N",
      }),
      loinc: "",
    });
    expect(res.status(), "create").toBe(200);

    expect(
      query(
        `SELECT is_active FROM clinlims.test_section WHERE id = ${Number(sectionId)}`,
      )[0][0],
      "the test section was woken up",
    ).toBe("Y");
    expect(
      query(`SELECT is_active FROM clinlims.panel WHERE id = ${Number(panelId)}`)[0][0],
      "and so was the panel",
    ).toBe("Y");

    // Result type 5 is "A" — a text-only variant, so one row like the numeric
    // case, but with no significant digits.
    expect(
      query(
        `SELECT tst_rslt_type, sort_order::text, COALESCE(significant_digits::text, '')
           FROM clinlims.test_result
          WHERE test_id IN (SELECT id FROM clinlims.test WHERE name = '${PREFIX}Wake')`,
      ),
      "one alphanumeric result",
    ).toEqual([["A", "1", ""]]);
  });
});
