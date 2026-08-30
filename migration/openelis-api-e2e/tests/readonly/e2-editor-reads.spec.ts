/**
 * e2 group 7 — the ten test-catalog reads that pair with no write.
 *
 * They share one thing: the AUGMENTED test name, `Albumin(Urines)` — the
 * localized name with the first sample type in brackets. Several shipped tests
 * already carry their specimen in their own name, so the result reads doubled;
 * that is the wire value.
 *
 * Three things here are quirks, not bugs to tidy:
 *
 *  - `siblings` reuses the list page's row type and fills only three of its
 *    fields, so every sibling answers `active: false` however active it is.
 *  - `coverageIncomplete` is hardcoded false everywhere; the decoration was left
 *    for a later milestone and never wired.
 *  - the two read-only controllers answer 404 with Spring's RFC 7807
 *    ProblemDetail, while the editor's own 404s have no body at all. Two 404
 *    shapes under one path prefix.
 *
 * Read-only: nothing here writes, so it lives in tests/readonly/.
 */
import { test, expect } from "@playwright/test";
import { readJson } from "../../fixtures/assert";
import { query } from "../../fixtures/db";

test.describe("e2 — the test-catalog reads that pair with no write", () => {
  test("GET /tests — filters, paging, and a sort on the name the response does not show", async ({
    request,
  }) => {
    const all = Number(query(`SELECT count(*)::text FROM clinlims.test`)[0][0]);

    const page1 = await readJson(
      await request.get("rest/test-catalog/tests?page=1&pageSize=3"),
      "tests page 1",
    );
    expect(page1.page, "page").toBe(1);
    expect(page1.pageSize, "pageSize").toBe(3);
    // total is the count AFTER filtering and BEFORE paging.
    expect(page1.total, "total is the unpaged match count").toBe(all);
    expect(page1.rows.length, "one page of rows").toBe(3);

    expect(
      Object.keys(page1.rows[0]).sort(),
      "the row shape",
    ).toEqual(
      ["active", "amr", "code", "coverageIncomplete", "domain", "name", "sampleType", "testId"].sort(),
    );
    expect(
      page1.rows.every((r: any) => r.coverageIncomplete === false),
      "coverageIncomplete is hardcoded false",
    ).toBe(true);

    // Sorted by the RAW localized name, case-insensitively — NOT by the
    // augmented name the rows carry. `Albumin(Urines)` sorts under "albumin".
    const stems = page1.rows.map((r: any) => stemOf(r.name).toLowerCase());
    expect(
      stems.every((s: string, i: number) => i === 0 || stems[i - 1] <= s),
      "sorted by the unaugmented name",
    ).toBe(true);

    // Page 2 continues where page 1 stopped, and does not repeat it.
    const page2 = await readJson(
      await request.get("rest/test-catalog/tests?page=2&pageSize=3"),
      "tests page 2",
    );
    const ids1 = page1.rows.map((r: any) => r.testId);
    const ids2 = page2.rows.map((r: any) => r.testId);
    expect(ids2.filter((id: string) => ids1.includes(id)), "no overlap").toEqual([]);
    expect(
      stemOf(page2.rows[0].name).toLowerCase() >=
        stemOf(page1.rows[page1.rows.length - 1].name).toLowerCase(),
      "page 2 continues the order",
    ).toBe(true);

    // Both page and pageSize are clamped UP to 1, so a zero of either is a
    // 200 rather than an error.
    const clamped = await readJson(
      await request.get("rest/test-catalog/tests?page=0&pageSize=0"),
      "tests clamped",
    );
    expect([clamped.page, clamped.pageSize], "clamped to one").toEqual([1, 1]);
    expect(clamped.rows.length, "one row").toBe(1);

    // A page past the end echoes the page it was asked for and returns nothing.
    const past = await readJson(
      await request.get("rest/test-catalog/tests?page=9999&pageSize=3"),
      "tests past the end",
    );
    expect([past.page, past.total, past.rows.length], "empty, page echoed").toEqual([
      9999,
      all,
      0,
    ]);

    // The five filters, each against the database.
    const activeCount = Number(
      query(`SELECT count(*)::text FROM clinlims.test WHERE is_active = 'Y'`)[0][0],
    );
    expect(
      (await readJson(
        await request.get("rest/test-catalog/tests?status=active&pageSize=1"),
        "active",
      )).total,
      "status=active",
    ).toBe(activeCount);
    expect(
      (await readJson(
        await request.get("rest/test-catalog/tests?status=inactive&pageSize=1"),
        "inactive",
      )).total,
      "status=inactive",
    ).toBe(all - activeCount);
    // Any other status — including the default "all" — filters nothing.
    expect(
      (await readJson(
        await request.get("rest/test-catalog/tests?status=all&pageSize=1"),
        "all",
      )).total,
      "status=all is not a filter",
    ).toBe(all);

    const amrCount = Number(
      query(
        `SELECT count(*)::text FROM clinlims.test WHERE COALESCE(antimicrobial_resistance, false)`,
      )[0][0],
    );
    expect(
      (await readJson(
        await request.get("rest/test-catalog/tests?amr=true&pageSize=1"),
        "amr",
      )).total,
      "amr=true",
    ).toBe(amrCount);

    const clinicalCount = Number(
      query(`SELECT count(*)::text FROM clinlims.test WHERE domain = 'CLINICAL'`)[0][0],
    );
    expect(
      (await readJson(
        await request.get("rest/test-catalog/tests?domain=CLINICAL&pageSize=1"),
        "domain",
      )).total,
      "domain=CLINICAL",
    ).toBe(clinicalCount);

    const sampleTypeCount = Number(
      query(
        `SELECT count(DISTINCT test_id)::text FROM clinlims.sampletype_test WHERE sample_type_id = 1`,
      )[0][0],
    );
    expect(
      (await readJson(
        await request.get("rest/test-catalog/tests?sampleType=1&pageSize=1"),
        "sampleType",
      )).total,
      "sampleType=1",
    ).toBe(sampleTypeCount);

    // search matches the RAW name, so a specimen that only appears in the
    // augmented one does not match.
    const bySearch = await readJson(
      await request.get("rest/test-catalog/tests?search=alb&pageSize=25"),
      "search",
    );
    expect(bySearch.total > 0, "search finds the analyte").toBe(true);
    expect(
      bySearch.rows.every((r: any) => stemOf(r.name).toLowerCase().includes("alb")),
      "every hit contains the term in its unaugmented name",
    ).toBe(true);
    expect(
      (await readJson(
        await request.get("rest/test-catalog/tests?search=Urines&pageSize=1"),
        "search by specimen",
      )).total,
      "the specimen is not searchable, only the analyte",
    ).toBe(0);
  });

  test("GET /tests/{id} — the envelope, and the nine sections it advertises", async ({
    request,
  }) => {
    const envelope = await readJson(
      await request.get("rest/test-catalog/tests/6"),
      "envelope",
    );
    expect(envelope, "the whole document").toEqual({
      testId: "6",
      name: "Albumin(Urines)",
      code: "Albumin-Urines",
      domain: "CLINICAL",
      applicableSections: [
        "basic-info",
        "sample-results",
        "methods",
        "ranges",
        "storage",
        "panels",
        "terminology",
        "analyzers",
        "display-order",
      ],
    });

    const missing = await request.get("rest/test-catalog/tests/999999");
    expect(missing.status(), "an unknown test").toBe(404);
    expect(await missing.text(), "with no body").toBe("");
  });

  test("GET /localization — the two ids the editor writes through", async ({
    request,
  }) => {
    const [[nameId, reportingId]] = query(
      `SELECT name_localization_id::text, reporting_name_localization_id::text
         FROM clinlims.test WHERE id = 6`,
    );
    expect(
      await readJson(
        await request.get("rest/test-catalog/tests/6/localization"),
        "localization",
      ),
      "name first, then reportingName",
    ).toEqual({
      testId: "6",
      fields: [
        { field: "name", localizationId: nameId },
        { field: "reportingName", localizationId: reportingId },
      ],
    });

    expect(
      (await request.get("rest/test-catalog/tests/999999/localization")).status(),
      "an unknown test",
    ).toBe(404);
  });

  test("GET /loinc-integrity — the two ways a result mis-routes", async ({
    request,
  }) => {
    // A LOINC two active tests share: the resolver takes whichever comes first.
    const shared = query(
      `SELECT loinc, string_agg(id::text, ',' ORDER BY id) FROM clinlims.test
        WHERE is_active = 'Y' AND COALESCE(loinc, '') <> ''
        GROUP BY loinc HAVING count(*) > 1 ORDER BY loinc LIMIT 1`,
    );
    expect(shared.length, "the fixture has a shared LOINC").toBe(1);
    const [sharedLoinc, sharedIds] = shared[0];
    const [first, ...others] = sharedIds.split(",");

    const dup = await readJson(
      await request.get(`rest/test-catalog/tests/${first}/loinc-integrity`),
      "duplicate loinc",
    );
    expect(dup.loinc, "the code").toBe(sharedLoinc);
    expect(dup.active, "active").toBe(true);
    expect(dup.noLoinc, "it has one, so no warning for the absence").toBe(false);
    expect(
      dup.duplicates.map((d: any) => d.testId).sort(),
      "every OTHER active test on that code, self excluded",
    ).toEqual([...others].sort());

    // A test that should receive results and cannot be matched at all.
    const noLoinc = query(
      `SELECT id::text FROM clinlims.test
        WHERE COALESCE(loinc, '') = '' AND is_active = 'Y' AND orderable = true
        ORDER BY id LIMIT 1`,
    );
    if (noLoinc.length) {
      const body = await readJson(
        await request.get(`rest/test-catalog/tests/${noLoinc[0][0]}/loinc-integrity`),
        "no loinc",
      );
      expect("loinc" in body, "the key is absent, not null").toBe(false);
      expect([body.active, body.noLoinc, body.duplicates], "warned").toEqual([
        true,
        true,
        [],
      ]);
    }

    // An INACTIVE test with no LOINC is not warned about: it receives nothing.
    const inactive = query(
      `SELECT id::text FROM clinlims.test
        WHERE is_active <> 'Y' AND COALESCE(loinc, '') = '' ORDER BY id LIMIT 1`,
    );
    if (inactive.length) {
      const body = await readJson(
        await request.get(`rest/test-catalog/tests/${inactive[0][0]}/loinc-integrity`),
        "inactive",
      );
      expect([body.active, body.noLoinc], "inactive, not warned").toEqual([false, false]);
    }

    expect(
      (await request.get("rest/test-catalog/tests/999999/loinc-integrity")).status(),
      "an unknown test",
    ).toBe(404);
  });

  test("GET /dictionary — a prefix match on the abbreviation AND the entry", async ({
    request,
  }) => {
    // Blank returns nothing, deliberately, so the control does not dump the
    // whole dictionary on focus.
    expect(
      await readJson(await request.get("rest/test-catalog/dictionary"), "no param"),
      "no search",
    ).toEqual([]);
    expect(
      await readJson(
        await request.get("rest/test-catalog/dictionary?search="),
        "blank",
      ),
      "blank search",
    ).toEqual([]);

    const expected = query(
      `SELECT d.id::text, d.dict_entry FROM clinlims.dictionary d
        WHERE d.is_active = 'Y'
          AND ( (d.local_abbrev IS NOT NULL
                 AND upper(d.local_abbrev || ': ' || d.dict_entry) LIKE upper('pos%'))
             OR (d.local_abbrev IS NULL AND upper(d.dict_entry) LIKE upper('pos%')) )
        ORDER BY d.dict_entry ASC LIMIT 50`,
    );
    const hits = await readJson(
      await request.get("rest/test-catalog/dictionary?search=pos"),
      "search=pos",
    );
    expect(
      hits.map((h: any) => [h.id, h.name]),
      "prefix match, ordered by the entry, capped at 50",
    ).toEqual(expected);

    // The match runs over `ABBREV: entry` as ONE string, so an entry whose text
    // does not start with the term still matches through its abbreviation.
    const viaAbbrev = hits.filter(
      (h: any) => !h.name.toLowerCase().startsWith("pos"),
    );
    expect(
      viaAbbrev.length > 0,
      "at least one hit matched on the abbreviation, not the entry",
    ).toBe(true);

    // Case-insensitive.
    expect(
      (await readJson(
        await request.get("rest/test-catalog/dictionary?search=POS"),
        "upper",
      )).length,
      "the search is case-insensitive",
    ).toBe(hits.length);
  });

  test("GET /siblings — grouped by analyte, and every row says inactive", async ({
    request,
  }) => {
    const siblings = await readJson(
      await request.get("rest/test-catalog/tests/6/siblings"),
      "siblings",
    );
    expect(siblings.length >= 1, "a test is its own sibling").toBe(true);
    expect(
      siblings.some((s: any) => s.testId === "6"),
      "the list includes self",
    ).toBe(true);
    // One stem for the whole set — that is what makes them siblings.
    const stems = new Set(siblings.map((s: any) => stemOf(s.name).toLowerCase()));
    expect(stems.size, "one analyte").toBe(1);

    // Test 6 is active, and its own sibling row says otherwise: the rows reuse
    // the list page's type and this handler fills only three of its fields, so
    // the primitive booleans serialise as false.
    expect(
      query(`SELECT id::text, is_active FROM clinlims.test WHERE id = 6`)[0][1],
      "test 6 is active in the database",
    ).toBe("Y");
    expect(
      siblings.every((s: any) => s.active === false && s.amr === false),
      "and every sibling row reports active: false",
    ).toBe(true);

    // An unknown test is an EMPTY LIST here, not a 404 — the only endpoint in
    // this group that answers that way.
    const unknown = await request.get("rest/test-catalog/tests/999999/siblings");
    expect(unknown.status(), "an unknown test is 200").toBe(200);
    expect(await unknown.json(), "with an empty list").toEqual([]);
  });

  test("GET /group/summary — whatever ids were selected, silently minus the bad ones", async ({
    request,
  }) => {
    const summary = await readJson(
      await request.get("rest/test-catalog/group/summary?ids=6,10,999999"),
      "group summary",
    );
    expect(
      summary.map((s: any) => s.testId),
      "the unknown id is dropped with no indication",
    ).toEqual(["6", "10"]);
    expect(
      Object.keys(summary[0]).sort(),
      "the row shape",
    ).toEqual(["active", "code", "name", "sampleType", "testId"].sort());
    expect(summary[0].name, "the name is augmented").toBe("Albumin(Urines)");

    // Blank entries and surrounding spaces are trimmed away.
    expect(
      (await readJson(
        await request.get("rest/test-catalog/group/summary?ids=6,,%2010%20"),
        "group summary trimmed",
      )).map((s: any) => s.testId),
      "blanks skipped, spaces trimmed",
    ).toEqual(["6", "10"]);
  });

  test("GET /analyzers — read-only, and empty on this deployment", async ({
    request,
  }) => {
    const rows = Number(
      query(`SELECT count(*)::text FROM clinlims.analyzer_test_map WHERE test_id = 6`)[0][0],
    );
    expect(
      await readJson(
        await request.get("rest/test-catalog/tests/6/analyzers"),
        "analyzers",
      ),
      "the mapping table decides this, and it has no row for test 6",
    ).toEqual({ testId: "6", analyzers: [] });
    expect(rows, "no mapping in the fixture").toBe(0);

    expect(
      (await request.get("rest/test-catalog/tests/999999/analyzers")).status(),
      "an unknown test",
    ).toBe(404);
  });

  test("GET /{testId}/reflex-calc — the rules a test triggers", async ({ request }) => {
    const withRules = query(
      `SELECT test_id::text FROM clinlims.test_reflex ORDER BY id LIMIT 1`,
    );
    expect(withRules.length, "the fixture has a reflex rule").toBe(1);
    const testId = withRules[0][0];

    const view = await readJson(
      await request.get(`rest/test-catalog/${testId}/reflex-calc`),
      "reflex-calc",
    );
    expect(Object.keys(view).sort(), "the three cross-link lists").toEqual([
      "calculatedBy",
      "feedsInto",
      "reflexRules",
    ]);

    const expected = query(
      `SELECT r.id::text,
              COALESCE(tr.value, COALESCE(r.non_dictionary_value, '')) AS trigger,
              COALESCE(NULLIF(alv.value, ''), added.description) AS added_name
         FROM clinlims.test_reflex r
         LEFT JOIN clinlims.test added ON added.id = r.add_test_id
         LEFT JOIN clinlims.localization_value alv
                ON alv.localization_id = added.name_localization_id AND alv.locale = 'en'
         LEFT JOIN clinlims.test_result tr ON tr.id = r.tst_rslt_id
        WHERE r.test_id = ${Number(testId)} ORDER BY r.id`,
    );
    expect(
      view.reflexRules.map((r: any) => [r.id, r.triggerCondition, r.reflexTests]),
      "one row per rule",
    ).toEqual(expected);
    // The rule name falls back to the added test's name when there is no
    // internal note, so an unnamed rule is labelled by what it adds.
    expect(
      view.reflexRules.every((r: any) => r.ruleName === r.reflexTests),
      "no internal note in the fixture, so the name is the added test",
    ).toBe(true);
    // reflexTests is the added test's plain localized name — NOT augmented,
    // unlike every other name in this group.
    expect(
      view.reflexRules.every((r: any) => !r.reflexTests.includes("(")),
      "the added test is not augmented with its specimen",
    ).toBe(true);

    const none = await readJson(
      await request.get("rest/test-catalog/6/reflex-calc"),
      "no rules",
    );
    expect(none, "a test nothing reflexes off").toEqual({
      reflexRules: [],
      calculatedBy: [],
      feedsInto: [],
    });

    // This controller throws ResponseStatusException, so its 404 carries
    // Spring's ProblemDetail — a different envelope from the editor's own.
    const unknown = await request.get("rest/test-catalog/999999/reflex-calc");
    expect(unknown.status(), "an unknown test").toBe(404);
    const problem = await unknown.json();
    expect(problem.status, "an RFC 7807 body, not an empty one").toBe(404);
    expect(
      problem.type,
      "and the unresolved message key Spring emits for it",
    ).toBe("problemDetail.type.org.springframework.web.server.ResponseStatusException");
  });

  test("GET /{testId}/storage/history — the entity, jsonb columns and all", async ({
    request,
  }) => {
    const withHistory = query(
      `SELECT s.test_id::text FROM clinlims.test_sample_handling s
         JOIN clinlims.test_sample_handling_history h ON h.test_sample_handling_id = s.id
        ORDER BY s.id LIMIT 1`,
    );
    expect(withHistory.length, "the fixture has a storage trail").toBe(1);
    const testId = withHistory[0][0];

    const history = await readJson(
      await request.get(`rest/test-catalog/${testId}/storage/history`),
      "storage history",
    );
    expect(history.length >= 1, "at least one entry").toBe(true);

    const [[handlingId, changeType, newValues]] = query(
      `SELECT h.test_sample_handling_id, h.change_type, h.new_values::text
         FROM clinlims.test_sample_handling_history h
         JOIN clinlims.test_sample_handling s ON s.id = h.test_sample_handling_id
        WHERE s.test_id = ${Number(testId)}
        ORDER BY h.changed_at DESC, h.id DESC LIMIT 1`,
    );
    const newest = history[0];
    expect(newest.testSampleHandlingId, "keyed to the handling row").toBe(handlingId);
    expect(newest.changeType, "the change type").toBe(changeType);
    // The controller returns the ENTITY, which types the jsonb columns as
    // String — so they arrive as STRINGS of JSON, not as objects.
    expect(typeof newest.newValues, "newValues is a string, not an object").toBe(
      "string",
    );
    expect(JSON.parse(newest.newValues), "and it parses to the stored snapshot").toEqual(
      JSON.parse(newValues),
    );
    // previousValues is null on an INSERT and absent rather than null.
    if (changeType === "INSERT") {
      expect("previousValues" in newest, "no previous state on an insert").toBe(false);
    }
    // Newest first.
    const stamps = history.map((h: any) => h.changedAt);
    expect(
      stamps.every((v: number, i: number) => i === 0 || stamps[i - 1] >= v),
      "ordered newest first",
    ).toBe(true);

    // A test with no storage config is an EMPTY LIST; only a missing TEST is a
    // 404, and that one carries a ProblemDetail.
    const noConfig = query(
      `SELECT id::text FROM clinlims.test
        WHERE id NOT IN (SELECT test_id FROM clinlims.test_sample_handling)
        ORDER BY id LIMIT 1`,
    );
    const empty = await request.get(
      `rest/test-catalog/${noConfig[0][0]}/storage/history`,
    );
    expect(empty.status(), "no config is not a 404").toBe(200);
    expect(await empty.json(), "just an empty list").toEqual([]);

    const unknown = await request.get("rest/test-catalog/999999/storage/history");
    expect(unknown.status(), "an unknown test").toBe(404);
    expect((await unknown.json()).status, "with a ProblemDetail body").toBe(404);
  });
});

/** The augmented name without its trailing "(SampleType)". */
function stemOf(name: string): string {
  const paren = name.lastIndexOf("(");
  return (paren > 0 ? name.slice(0, paren) : name).trim();
}
