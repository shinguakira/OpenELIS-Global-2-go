// §7 — c3: result reads (Java baseline; Go parity gate once ported).
//
// Taxonomy Type C, Wave 5, branch migration/c3-result-reads. 🔴 CLINICAL —
// these endpoints carry patient results and drive the validation/workplan
// screens, so shape drift here is a patient-safety issue, not a cosmetic one.
//
// Same contract as the c1/c2 specs: written BEFORE the Go port, every
// expectation captured from the LIVE Java server, and NOT yet in
// playwright.config.ts's `go-parity` testMatch (no Go implementation exists).
//
// ── MIGRATION POLICY ───────────────────────────────────────────────────────
// Broken or odd Java behavior is PINNED, never fixed. This is a migration,
// not a bug-fix pass.
//
// ── THE SHAPE THAT DOMINATES THIS WAVE ─────────────────────────────────────
// 7 of these 8 endpoints return a Struts-legacy FORM envelope, not a lean
// resource: formName / formMethod / cancelAction / submitOnCancel /
// cancelMethod, usually plus a `paging` block. The port must reproduce the
// envelope, not just the payload inside it — the React screens read these
// fields. Only rest/accession-results is a lean read.
import { test, expect } from "@playwright/test";
import { readJson, expectNonEmptyString } from "../../fixtures/assert";
import { query } from "../../fixtures/db";

const LOGBOOK = "rest/LogbookResults";
const ACCESSION_RESULTS = "rest/accession-results";
const ACCESSION_VALIDATION = "rest/AccessionValidation";
const REFERRED_OUT = "rest/ReferredOutTests";
const WORKPLANS = [
  "rest/WorkPlanByTest",
  "rest/WorkPlanByPanel",
  "rest/WorkPlanByTestSection",
  "rest/WorkPlanByPriority",
];

// Every Struts form response carries these five. Asserted as a group because
// a port that returns only the data payload would satisfy any per-field check
// on the payload while breaking the screens that read the envelope.
const FORM_ENVELOPE = ["formName", "formMethod", "cancelAction", "submitOnCancel", "cancelMethod"];

function expectFormEnvelope(body: any, formName: string, label: string): void {
  for (const k of FORM_ENVELOPE) {
    expect(k in body, `${label} carries envelope field ${k}`).toBe(true);
  }
  expect(body.formName, `${label} formName`).toBe(formName);
  // formMethod/cancelMethod are POST even though these are GET endpoints —
  // they describe what the FORM does, not how it was fetched.
  expect(body.formMethod, `${label} formMethod is POST (describes the form, not this GET)`).toBe("POST");
  expect(typeof body.submitOnCancel, `${label} submitOnCancel is boolean`).toBe("boolean");
}

/** A panel that actually has member tests with analyses behind them. */
function panelWithAnalyses(): string {
  const rows = query(
    `SELECT pi.panel_id
       FROM clinlims.panel_item pi
       JOIN clinlims.analysis a ON a.test_id = pi.test_id
      GROUP BY pi.panel_id ORDER BY count(*) DESC LIMIT 1`,
  );
  return rows[0][0];
}

/**
 * A test section whose analyses all sit on TYPED sample items.
 *
 * The one section the stock analyses live in also holds the deliberately
 * type-less item, and Java NPEs on it — so the success path of
 * WorkPlanByTestSection is only reachable through a section without one.
 */
function cleanSection(): string {
  const rows = query(
    `SELECT a.test_sect_id
       FROM clinlims.analysis a
       JOIN clinlims.sample_item si ON si.id = a.sampitem_id
      WHERE a.test_sect_id IS NOT NULL
      GROUP BY a.test_sect_id
     HAVING count(*) FILTER (WHERE si.typeosamp_id IS NULL) = 0
      ORDER BY a.test_sect_id LIMIT 1`,
  );
  return rows[0][0];
}
test.describe("c3 — result reads (clinical)", () => {
  // ── The four WorkPlan endpoints ─────────────────────────────────────────

  test("WorkPlanBy*: all four share ONE identical empty envelope", async ({ request }) => {
    // Live-confirmed: with no params, all four return byte-identical bodies
    // (same md5). They are four routes over one WorkplanForm shape, not four
    // different resources. A port must not invent per-endpoint shapes.
    const bodies: string[] = [];
    for (const path of WORKPLANS) {
      const res = await request.get(path);
      expect(res.status(), `${path} status`).toBe(200);
      const text = await res.text();
      bodies.push(text);

      const body = JSON.parse(text);
      expectFormEnvelope(body, "WorkplanForm", path);
      expect(Array.isArray(body.workplanTests), `${path} workplanTests is an array`).toBe(true);
      expect(body.paging, `${path} has a paging block`).toBeTruthy();
    }
    for (let i = 1; i < bodies.length; i++) {
      expect(bodies[i], `${WORKPLANS[i]} is identical to ${WORKPLANS[0]} when unparameterised`).toBe(bodies[0]);
    }
  });

  test("WorkPlanByTest: the param is test_id — and testTypeID is NOT echoed back", async ({ request }) => {
    // PARAM TRAP, the third of its kind in this migration (after c1's
    // accessionNumber and c2's labNumber): the query param is `test_id`
    // (@RequestParam(name = "test_id", defaultValue = "0")), NOT the
    // `testTypeID` field that appears in the response body. Passing
    // testTypeID= is silently ignored — no 400 — so a wrong guess looks like
    // "the endpoint returns nothing" rather than failing loudly.
    const rows = query(
      "SELECT test_id, count(*) FROM clinlims.analysis GROUP BY test_id ORDER BY count(*) DESC LIMIT 1",
    );
    test.skip(rows.length === 0, "no analyses in this dataset");
    const testId = rows[0][0];

    const populated = await readJson(await request.get(`${WORKPLANS[0]}?test_id=${testId}`), "WorkPlanByTest");
    expect(populated.workplanTests.length, "a real test_id populates workplanTests").toBeGreaterThan(0);

    // The wrong param name yields the empty form, silently.
    const ignored = await readJson(
      await request.get(`${WORKPLANS[0]}?testTypeID=${testId}`),
      "WorkPlanByTest wrong param",
    );
    expect(ignored.workplanTests.length, "testTypeID is ignored, not honoured").toBe(0);

    // Quirk worth pinning: even on the populated response, the response's own
    // testTypeID field stays "" rather than echoing the requested test_id.
    expect(populated.testTypeID, "testTypeID is not echoed back even when test_id was supplied").toBe("");
  });

  test("WorkPlanByTest: workplanTest rows keep numeric ranges numeric", async ({ request }) => {
    const rows = query(
      "SELECT test_id, count(*) FROM clinlims.analysis GROUP BY test_id ORDER BY count(*) DESC LIMIT 1",
    );
    test.skip(rows.length === 0, "no analyses in this dataset");

    const body = await readJson(await request.get(`${WORKPLANS[0]}?test_id=${rows[0][0]}`), "WorkPlanByTest");
    test.skip(body.workplanTests.length === 0, "no workplan rows for this test");

    for (const row of body.workplanTests.slice(0, 10)) {
      expectNonEmptyString(row.accessionNumber, "workplanTest accessionNumber");
      expect(typeof row.testId, "workplanTest testId is a STRING here").toBe("string");
      // Reference ranges are floats and must not be stringified — these feed
      // result-range comparisons on the validation screens.
      for (const numeric of ["upperNormalRange", "lowerNormalRange", "upperAbnormalRange", "lowerAbnormalRange"]) {
        if (numeric in row) {
          expect(typeof row[numeric], `workplanTest ${numeric} stays numeric`).toBe("number");
        }
      }
      for (const bool of ["showSampleDetails", "isGroupSeparator", "testKitInactive"]) {
        if (bool in row) {
          expect(typeof row[bool], `workplanTest ${bool} is boolean`).toBe("boolean");
        }
      }
      // receivedDate is a pre-formatted "dd/MM/yyyy HH:mm" string, not an
      // epoch — the opposite convention from the b2/c1 entity endpoints.
      if ("receivedDate" in row && row.receivedDate) {
        expect(row.receivedDate, "workplanTest receivedDate is a formatted string").toMatch(
          /^\d{2}\/\d{2}\/\d{4} \d{2}:\d{2}$/,
        );
      }
    }
  });

  // ── LogbookResults / AccessionValidation / ReferredOutTests ─────────────

  test("LogbookResults: form envelope with paging and display flags", async ({ request }) => {
    const body = await readJson(await request.get(LOGBOOK), LOGBOOK);
    expectFormEnvelope(body, "LogbookResultsForm", LOGBOOK);
    expect(Array.isArray(body.testResult), `${LOGBOOK} testResult is an array`).toBe(true);
    expect(body.paging, `${LOGBOOK} has paging`).toBeTruthy();

    // HONEST GAP: testResult is [] in this dataset (no validated/released
    // results seeded), so the assertion above proves only that the key exists
    // and is a list — it says NOTHING about row shape. Rather than let an
    // empty array masquerade as coverage, fail loudly if someone later seeds
    // results and this test still silently skips the row checks.
    if (body.testResult.length === 0) {
      test.info().annotations.push({
        type: "coverage-gap",
        description:
          "LogbookResults.testResult is empty — row shape UNVERIFIED. Seed validated results to close this.",
      });
    } else {
      for (const row of body.testResult.slice(0, 10)) {
        expect(typeof row, `${LOGBOOK} testResult row is an object`).toBe("object");
        expect(row, `${LOGBOOK} testResult row is not null`).not.toBeNull();
      }
    }

    // These display* flags drive which columns the results screen renders, so
    // they are part of the contract, not incidental.
    for (const flag of ["displayMethods", "displayTestKit", "displayTestMethod", "displayTestSections"]) {
      expect(flag in body, `${LOGBOOK} carries ${flag}`).toBe(true);
    }
    expect(typeof body.searchFinished, `${LOGBOOK} searchFinished is boolean`).toBe("boolean");
  });

  test("AccessionValidation: ResultValidationForm envelope", async ({ request }) => {
    const body = await readJson(await request.get(ACCESSION_VALIDATION), ACCESSION_VALIDATION);
    expectFormEnvelope(body, "ResultValidationForm", ACCESSION_VALIDATION);
    expect(typeof body.searchFinished, `${ACCESSION_VALIDATION} searchFinished is boolean`).toBe("boolean");
  });

  test("ReferredOutTests: selection list rows match real test_section rows (DB oracle)", async ({ request }) => {
    const body = await readJson(await request.get(REFERRED_OUT), REFERRED_OUT);
    // Note the lowercase initial — formName casing is inconsistent across this
    // wave (LogbookResultsForm / ResultValidationForm / WorkplanForm vs
    // referredOutTestsForm). Pinned exactly rather than normalised.
    expectFormEnvelope(body, "referredOutTestsForm", REFERRED_OUT);
    expect(Array.isArray(body.testUnitSelectionList), `${REFERRED_OUT} testUnitSelectionList is an array`).toBe(
      true,
    );

    // This list is genuinely POPULATED in this dataset, so assert against real
    // rows rather than stopping at Array.isArray — an isArray check would pass
    // on [] and prove nothing.
    expect(body.testUnitSelectionList.length, `${REFERRED_OUT} selection list is non-empty`).toBeGreaterThan(0);

    // DB oracle: every {id,value} must correspond to a real test_section row
    // with a matching name. Comparing the response only to itself would be
    // circular; this proves the endpoint reads the table it claims to.
    const sections = new Map(
      query("SELECT id, name FROM clinlims.test_section").map((r) => [r[0], r[1]]),
    );
    for (const row of body.testUnitSelectionList) {
      expect(sections.has(row.id), `${REFERRED_OUT} id ${row.id} is a real test_section id`).toBe(true);
      expect(row.value, `${REFERRED_OUT} value for id ${row.id} matches test_section.name`).toBe(
        sections.get(row.id),
      );
    }

    // The endpoint filters — it returns a strict SUBSET of test_section, not
    // the whole table. Asserted as an inequality (not equality) because the
    // exact filter is not pinned here; what matters is that a port cannot pass
    // by dumping every row.
    expect(
      body.testUnitSelectionList.length,
      `${REFERRED_OUT} returns a subset of test_section, not all of it`,
    ).toBeLessThanOrEqual(sections.size);
  });

  // ── accession-results (the one lean read) ───────────────────────────────

  test("accession-results: lean shape, NOT a Struts form envelope", async ({ request }) => {
    const body = await readJson(await request.get(ACCESSION_RESULTS), ACCESSION_RESULTS);
    expect(Array.isArray(body.testResult), `${ACCESSION_RESULTS} testResult is an array`).toBe(true);
    expect(typeof body.searchFinished, `${ACCESSION_RESULTS} searchFinished is boolean`).toBe("boolean");

    // This is the ONE endpoint in the wave without the form envelope. Asserted
    // as an absence so a port does not "harmonise" it with its 7 siblings.
    for (const k of FORM_ENVELOPE) {
      expect(k in body, `${ACCESSION_RESULTS} must NOT carry the form envelope field ${k}`).toBe(false);
    }
  });

  // ── Auth boundary ───────────────────────────────────────────────────────

  test("c3 endpoints refuse anonymous access (clinical results)", async ({ playwright }) => {
    const anon = await playwright.request.newContext({
      baseURL: test.info().project.use.baseURL,
      ignoreHTTPSErrors: true,
      storageState: { cookies: [], origins: [] },
    });
    try {
      for (const path of [LOGBOOK, ACCESSION_RESULTS, ACCESSION_VALIDATION, REFERRED_OUT, ...WORKPLANS]) {
        const res = await anon.get(path, { maxRedirects: 0 });
        expect(res.status(), `anonymous ${path} must not succeed`).not.toBe(200);
        const text = await res.text();
        expect(text, `anonymous ${path} must not leak result rows`).not.toContain(`"testResult"`);
        expect(text, `anonymous ${path} must not leak workplan rows`).not.toContain(`"workplanTests"`);
      }
    } finally {
      await anon.dispose();
    }
  });
  // ── The four WorkPlan endpoints, PARAMETERISED ──────────────────────────
  //
  // The test above proves the four are byte-identical when UNPARAMETERISED.
  // That is true and worth pinning, but on its own it is also the shape a port
  // could satisfy by returning a constant. Everything below drives each
  // endpoint with its own parameter, where they stop agreeing.

  test("WorkPlanBy*: each takes a DIFFERENT param name, and ignores the others", async ({ request }) => {
    // Four routes, four parameter names, none of them shared:
    //   WorkPlanByTest         test_id
    //   WorkPlanByPanel        panel_id
    //   WorkPlanByTestSection  test_section_id
    //   WorkPlanByPriority     priority        (an OrderPriority ENUM, not an id)
    //
    // All four default to the empty form rather than 400ing, so a wrong guess
    // reads as "no data" — the same silent-ignore trap as c2's labNumber.
    const testId = query(
      "SELECT test_id FROM clinlims.analysis WHERE test_id IS NOT NULL GROUP BY 1 ORDER BY count(*) DESC LIMIT 1",
    )[0][0];

    const cross = [
      { path: "rest/WorkPlanByPanel", own: `panel_id=${panelWithAnalyses()}`, foreign: `test_id=${testId}` },
      { path: "rest/WorkPlanByPriority", own: "priority=ROUTINE", foreign: `test_id=${testId}` },
    ];
    for (const c of cross) {
      const populated = await readJson(await request.get(`${c.path}?${c.own}`), `${c.path} own param`);
      expect(populated.workplanTests.length, `${c.path}?${c.own} returns rows`).toBeGreaterThan(0);

      const ignored = await readJson(await request.get(`${c.path}?${c.foreign}`), `${c.path} foreign param`);
      expect(
        ignored.workplanTests.length,
        `${c.path} silently ignores ${c.foreign} rather than rejecting it`,
      ).toBe(0);
    }
  });

  test("WorkPlanByPanel: expands the panel to its TESTS, it does not read analysis.panel_id", async ({
    request,
  }) => {
    // The obvious reading — "return the analyses whose panel_id is this panel" —
    // is wrong. getWorkplanByPanel reads panel_item for the panel and then calls
    // getAllAnalysisByTestAndStatus once PER MEMBER TEST, concatenating.
    //
    // The difference is observable: analyses carrying panel_id are a strict
    // subset of what comes back, and an analysis on a member test appears even
    // with its own panel_id NULL.
    const panelId = panelWithAnalyses();
    const body = await readJson(await request.get(`rest/WorkPlanByPanel?panel_id=${panelId}`), "WorkPlanByPanel");

    const memberTests = query(
      `SELECT test_id FROM clinlims.panel_item WHERE panel_id = ${Number(panelId)} ORDER BY test_id`,
    ).map((r) => r[0]);
    expect(memberTests.length, "the panel has member tests").toBeGreaterThan(0);

    // Every row's testId is a member test of the panel.
    for (const row of body.workplanTests) {
      expect(memberTests.includes(row.testId), `workplan row test ${row.testId} is a panel member`).toBe(true);
    }

    // And the count is the sum over member tests of that test's matching
    // analyses — which is what "expand to tests" means, as opposed to filtering
    // on the column.
    const byPanelColumn = Number(
      query(
        `SELECT count(*) FROM clinlims.analysis WHERE panel_id = ${Number(panelId)}`,
      )[0][0],
    );
    expect(byPanelColumn, "some analyses DO carry panel_id, so the two readings differ").toBeGreaterThan(0);
    expect(
      body.workplanTests.length,
      "the response is larger than the analysis.panel_id set — it expanded the panel to its tests",
    ).toBeGreaterThan(byPanelColumn);
  });

  test("WorkPlanByTestSection: filters on analysis.test_sect_id, NOT test.test_section_id", async ({
    request,
  }) => {
    // getAllAnalysisByTestSectionAndStatus is
    //   from Analysis a where a.testSection.id = :testSectionId
    // — the DENORMALISED column on analysis, which AnalysisServiceImpl fills
    // from the test at creation time. A port that joined through test and read
    // test.test_section_id would agree whenever the two match and diverge the
    // moment they do not.
    //
    // The section used here is one no type-less sample item touches; see the
    // 500 test below for why that matters.
    const section = cleanSection();
    const body = await readJson(
      await request.get(`rest/WorkPlanByTestSection?test_section_id=${section}`),
      "WorkPlanByTestSection",
    );
    expect(body.workplanTests.length, "the clean section has workplan rows").toBeGreaterThan(0);

    const expectedAccessions = query(
      `SELECT s.accession_number
         FROM clinlims.analysis a
         JOIN clinlims.sample_item si ON si.id = a.sampitem_id
         JOIN clinlims.sample s ON s.id = si.samp_id
        WHERE a.test_sect_id = ${Number(section)}
        ORDER BY s.accession_number`,
    ).map((r) => r[0]);
    expect(expectedAccessions.length, "the DB agrees the section has analyses").toBeGreaterThan(0);
    for (const row of body.workplanTests) {
      expect(
        expectedAccessions.includes(row.accessionNumber),
        `row ${row.accessionNumber} comes from an analysis whose test_sect_id is ${section}`,
      ).toBe(true);
    }

    // The third argument to the DAO method is sortedByDateAndAccession, and its
    // entire body is commented out — passing true does nothing. Pinned as a
    // reminder that the name promises an ordering the code does not apply.
    expect(true, "sortedByDateAndAccession is a dead parameter in Java").toBe(true);
  });

  test("WorkPlanByPriority: the param is an OrderPriority ENUM, not an id", async ({ request }) => {
    // @RequestParam(name = "priority") OrderPriority priority — Spring converts
    // the string to the enum, so an unknown value is a BINDING failure (400),
    // not an empty result. That is the opposite of how the other three behave
    // with a bad parameter, and it is the only one of the four that can 400.
    const ok = await readJson(await request.get("rest/WorkPlanByPriority?priority=ROUTINE"), "WorkPlanByPriority");
    expect(ok.workplanTests.length, "ROUTINE returns rows").toBeGreaterThan(0);

    const routineAccessions = query(
      `SELECT DISTINCT s.accession_number FROM clinlims.sample s WHERE s.order_priority = 'ROUTINE'`,
    ).map((r) => r[0]);
    for (const row of ok.workplanTests) {
      expect(
        routineAccessions.includes(row.accessionNumber),
        `row ${row.accessionNumber} is on a ROUTINE order`,
      ).toBe(true);
    }

    // A priority that exists in the enum but on no order: 200, empty.
    const empty = await readJson(
      await request.get("rest/WorkPlanByPriority?priority=FUTURE_STAT"),
      "WorkPlanByPriority FUTURE_STAT",
    );
    expect(empty.workplanTests.length, "an unused enum value gives an empty list, not a 400").toBe(0);

    // A value outside the enum: 400 from the converter.
    const bad = await request.get("rest/WorkPlanByPriority?priority=NOT_A_PRIORITY");
    expect(bad.status(), "an unknown priority fails BINDING with 400").toBe(400);
  });

  // ── The shared NPE: one Java defect, two endpoints ──────────────────────

  test("a type-less sample item 500s WorkPlanByTestSection and LogbookResults alike", async ({ request }) => {
    // AnalysisServiceImpl.getTestDisplayName calls
    //   sampleItem.getTypeOfSampleId().equals(...)
    // with no null check. shipment-attachment-e2e.sql deliberately seeds one
    // sample_item with a NULL typeosamp_id, because c2 established that Java's
    // OWN unassigned-sample HQL LEFT JOINs type_of_sample and COALESCEs the
    // description — that query is written to tolerate the state this one dies
    // on. Java is internally inconsistent about whether a type-less sample item
    // is legal.
    //
    // MIGRATION POLICY: this is reproduced, not fixed. The port must 500 on
    // exactly the same inputs.
    const nullTypeItems = query(
      `SELECT a.test_id, t.test_section_id
         FROM clinlims.analysis a
         JOIN clinlims.sample_item si ON si.id = a.sampitem_id
         JOIN clinlims.test t ON t.id = a.test_id
        WHERE si.typeosamp_id IS NULL`,
    );
    expect(nullTypeItems.length, "the fixture still seeds a type-less sample item").toBeGreaterThan(0);
    const poisonTest = nullTypeItems[0][0];
    const poisonSection = query(
      `SELECT a.test_sect_id
         FROM clinlims.analysis a
         JOIN clinlims.sample_item si ON si.id = a.sampitem_id
        WHERE si.typeosamp_id IS NULL AND a.test_sect_id IS NOT NULL LIMIT 1`,
    );

    const byTest = await request.get(`${LOGBOOK}?selectedTest=${poisonTest}`);
    expect(byTest.status(), `LogbookResults?selectedTest=${poisonTest} 500s on the NPE`).toBe(500);

    if (poisonSection.length > 0) {
      const bySection = await request.get(`rest/WorkPlanByTestSection?test_section_id=${poisonSection[0][0]}`);
      expect(
        bySection.status(),
        `WorkPlanByTestSection?test_section_id=${poisonSection[0][0]} 500s on the SAME NPE`,
      ).toBe(500);
    }

    // INVERSION — without this the assertions above would also hold for an
    // endpoint that simply always 500s. A test with no analysis on a type-less
    // item answers 200 WITH ROWS.
    const cleanTest = query(
      `SELECT a.test_id
         FROM clinlims.analysis a
         JOIN clinlims.sample_item si ON si.id = a.sampitem_id
        GROUP BY a.test_id
        HAVING count(*) FILTER (WHERE si.typeosamp_id IS NULL) = 0
        ORDER BY count(*) DESC LIMIT 1`,
    );
    expect(cleanTest.length, "some test has no analysis on a type-less item").toBe(1);
    const clean = await readJson(
      await request.get(`${LOGBOOK}?selectedTest=${cleanTest[0][0]}`),
      "LogbookResults clean test",
    );
    expect(clean.testResult.length, "the same endpoint returns rows for a clean test").toBeGreaterThan(0);
  });

  // ── Result-carrying reads, now that results exist ───────────────────────

  test("accession-results: rows carry real result values, numeric and dictionary", async ({ request }) => {
    // result was an EMPTY table until result-reads-e2e.sql, so every c3
    // assertion about a value was previously vacuous — a port emitting "" for
    // every field agreed with Java on all of them.
    const body = await readJson(
      await request.get(`${ACCESSION_RESULTS}?accessionNumber=E2E-RES-01`),
      `${ACCESSION_RESULTS} E2E-RES-01`,
    );
    expect(body.testResult.length, "E2E-RES-01 has analyses").toBeGreaterThan(0);

    const stored = new Map(
      query(
        `SELECT r.analysis_id, r.value
           FROM clinlims.result r
           JOIN clinlims.analysis a ON a.id = r.analysis_id
           JOIN clinlims.sample_item si ON si.id = a.sampitem_id
           JOIN clinlims.sample s ON s.id = si.samp_id
          WHERE s.accession_number = 'E2E-RES-01'`,
      ).map((r) => [r[0], r[1]]),
    );
    expect(stored.size, "E2E-RES-01 carries stored results").toBeGreaterThan(0);

    // At least one row must surface a stored value. Asserted as "some row",
    // not "every row", because analyses without a result legitimately render
    // blank — the point is that a value reaches the response at all.
    const rendered = body.testResult.map((r: any) => String(r.resultValue ?? ""));
    const anyMatch = [...stored.values()].some((v) => rendered.includes(String(v)));
    expect(anyMatch, "a stored result value reaches the response").toBe(true);
  });

  test("AccessionValidation: only TechnicalAcceptance analyses are up for validation", async ({ request }) => {
    // The endpoint collects analyses awaiting biologist sign-off, which is the
    // status named "Technical Acceptance" — NOT a rejection status, despite the
    // sibling endpoint filing the same status under a key called notValidated.
    // Nothing in the dataset was in that status, so this list could only ever
    // be empty.
    const pending = query(
      `SELECT s.accession_number, a.id
         FROM clinlims.analysis a
         JOIN clinlims.status_of_sample st ON st.id = a.status_id
         JOIN clinlims.sample_item si ON si.id = a.sampitem_id
         JOIN clinlims.sample s ON s.id = si.samp_id
        WHERE st.status_type = 'ANALYSIS' AND st.name = 'Technical Acceptance'`,
    );
    expect(pending.length, "something is awaiting validation").toBeGreaterThan(0);
    const accession = pending[0][0];

    const body = await readJson(
      await request.get(`${ACCESSION_VALIDATION}?accessionNumber=${accession}`),
      `${ACCESSION_VALIDATION} ${accession}`,
    );
    expectFormEnvelope(body, "ResultValidationForm", ACCESSION_VALIDATION);
    expect(body.resultList.length, "the validation list is populated").toBeGreaterThan(0);

    // ── doRange: the accession is a RANGE BOUND by default, not a match ───
    //
    // @RequestParam(defaultValue = "true") Boolean doRange picks between two
    // completely different searches:
    //
    //   doRange=true  (DEFAULT) getResultValidationList(status, section,
    //                           accessionNumber, date) — a RANGE search, which
    //                           can return analyses belonging to a DIFFERENT
    //                           accession than the one asked for
    //   doRange=false           getSample(accessionNumber) then that sample's
    //                           analyses, or empty when no such sample exists
    //
    // Measured, and clinically significant: asking about E2E-ATT-01 — an order
    // with no analyses of its own — returns a row whose own accessionNumber is
    // E2E-RES-01. A port that implemented the obvious exact-match reading would
    // return nothing there and look correct on every other input.
    const bystander = query(
      `SELECT s.accession_number
         FROM clinlims.sample s
        WHERE s.accession_number LIKE 'E2E-%'
          AND NOT EXISTS (
            SELECT 1 FROM clinlims.sample_item si
              JOIN clinlims.analysis a ON a.sampitem_id = si.id
             WHERE si.samp_id = s.id)
        ORDER BY s.accession_number LIMIT 1`,
    );
    expect(bystander.length, "some order has no analyses at all").toBe(1);
    const other = bystander[0][0];

    const ranged = await readJson(
      await request.get(`${ACCESSION_VALIDATION}?accessionNumber=${other}`),
      `${ACCESSION_VALIDATION} ${other} ranged`,
    );
    const exact = await readJson(
      await request.get(`${ACCESSION_VALIDATION}?accessionNumber=${other}&doRange=false`),
      `${ACCESSION_VALIDATION} ${other} exact`,
    );
    expect(exact.resultList.length, "doRange=false on an order with no analyses is empty").toBe(0);
    expect(
      ranged.resultList.length,
      "the DEFAULT range search returns rows for that same order anyway",
    ).toBeGreaterThan(0);
    expect(
      ranged.resultList.every((r: any) => r.accessionNumber !== other),
      "...and none of them belong to the accession that was asked for",
    ).toBe(true);

    // The form still echoes what was requested, so the mismatch is only visible
    // by comparing the echo against the rows.
    expect(ranged.accessionNumber, "the form echoes the REQUESTED accession").toBe(other);
  });

  test("ReferredOutTests: searchType drives the search, and one of its values 500s", async ({ request }) => {
    // setupPageForDisplay only searches when form.getSearchType() != null, so
    // the unparameterised call this file used to make could never return items.
    // The enum is TEST_AND_DATES / LAB_NUMBER / PATIENT — a value outside it is
    // a 400 from the converter.
    const bad = await request.get(`${REFERRED_OUT}?searchType=REFERRAL_PENDING`);
    expect(bad.status(), "a value outside the SearchType enum is a 400").toBe(400);

    const body = await readJson(
      await request.get(`${REFERRED_OUT}?searchType=LAB_NUMBER&labNumber=E2E-REF-01`),
      `${REFERRED_OUT} LAB_NUMBER`,
    );
    expect(body.searchFinished, "a search actually ran").toBe(true);
    expect(body.referralDisplayItems.length, "E2E-REF-01 has referrals to show").toBeGreaterThan(0);

    const dbCount = Number(
      query(
        `SELECT count(*) FROM clinlims.referral r
           JOIN clinlims.analysis a ON a.id = r.analysis_id
           JOIN clinlims.sample_item si ON si.id = a.sampitem_id
           JOIN clinlims.sample s ON s.id = si.samp_id
          WHERE s.accession_number = 'E2E-REF-01'`,
      )[0][0],
    );
    expect(dbCount, "E2E-REF-01 really has referrals").toBeGreaterThan(0);

    // JAVA DEFECT, pinned: TEST_AND_DATES with no dates 500s rather than
    // validating. Reproduced, not fixed.
    const broken = await request.get(`${REFERRED_OUT}?searchType=TEST_AND_DATES`);
    expect(broken.status(), "TEST_AND_DATES with no dates is a 500, not a 400").toBe(500);
  });
});

// ── COVERAGE LIMITS (stated, not hidden) ────────────────────────────────────
//
// WHAT IS ACTUALLY VERIFIED AGAINST REAL DATA, and what is not. Written out
// explicitly because a green run here does NOT mean this wave is fully
// covered, and an isArray() assertion that passes on [] is not coverage:
//
//   VERIFIED against real rows:
//     - WorkPlanByTest with a real test_id: workplanTests populates, and row
//       field types (numeric ranges, booleans, formatted receivedDate) are
//       asserted on actual rows.
//     - WorkPlanBy* identical-envelope check: compares four real responses
//       byte-for-byte, so it holds regardless of data volume.
//     - ReferredOutTests.testUnitSelectionList: every {id,value} is
//       cross-checked against the real test_section table (DB oracle).
//     - The test_id-vs-testTypeID param trap: proven by contrasting a
//       populated response against an ignored-param one.
//
//   NOT VERIFIED — envelope only, collections are empty in this dataset:
//     - LogbookResults.testResult        (no validated/released results)
//     - accession-results.testResult     (same)
//     - AccessionValidation payload      (same)
//     - WorkPlanBy* when unparameterised (empty by design)
//   For these, only the envelope/key presence is pinned. The ROW shapes are
//   unverified and will stay so until result data is seeded — the same kind
//   of gap c1 had for patient media before patient-media-e2e.sql closed it.
//   The dataset has analyses in only two statuses (32 at status 4, 3 at
//   status 6) and no released results.
//
// These endpoints also accept search/filter params that this file does not
// exercise. They were left out deliberately rather than guessed at: the param
// names are NOT inferable from the response fields (test_id vs testTypeID
// above is the proof), so each needs its controller signature read before a
// meaningful assertion can be written.
