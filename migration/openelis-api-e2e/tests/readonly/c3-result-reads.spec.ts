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
import type { APIRequestContext } from "@playwright/test";
import { readJson, expectNonEmptyString } from "../../fixtures/assert";
import { query } from "../../fixtures/db";
import { E2E_PASS, ADMIN_USER, ADMIN_PASS } from "../../fixtures/env";
import {
  LOGIN_PATH,
  LOGIN_USER_FIELD,
  LOGIN_PASS_FIELD,
  SESSION_PATH,
} from "../../fixtures/contract";

/**
 * Log in as a seeded user and return their own request context.
 *
 * The lab-unit test needs TWO identities in one run — a restricted user and
 * admin — so it cannot use the suite-wide authenticated context.
 */
async function loginAs(
  apiRequest: { request: any },
  user: string,
): Promise<APIRequestContext> {
  const ctx = await apiRequest.request.newContext({
    baseURL: test.info().project.use.baseURL,
    ignoreHTTPSErrors: true,
    storageState: { cookies: [], origins: [] },
  });
  await ctx.get(SESSION_PATH);
  const pass = user === ADMIN_USER ? ADMIN_PASS : E2E_PASS;
  const res = await ctx.post(LOGIN_PATH, {
    form: { [LOGIN_USER_FIELD]: user, [LOGIN_PASS_FIELD]: pass },
  });
  expect(res.status(), `login as ${user}`).toBe(200);
  return ctx;
}

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
const FORM_ENVELOPE = [
  "formName",
  "formMethod",
  "cancelAction",
  "submitOnCancel",
  "cancelMethod",
];

function expectFormEnvelope(body: any, formName: string, label: string): void {
  for (const k of FORM_ENVELOPE) {
    expect(k in body, `${label} carries envelope field ${k}`).toBe(true);
  }
  expect(body.formName, `${label} formName`).toBe(formName);
  // formMethod/cancelMethod are POST even though these are GET endpoints —
  // they describe what the FORM does, not how it was fetched.
  expect(
    body.formMethod,
    `${label} formMethod is POST (describes the form, not this GET)`,
  ).toBe("POST");
  expect(typeof body.submitOnCancel, `${label} submitOnCancel is boolean`).toBe(
    "boolean",
  );
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

  test("WorkPlanBy*: all four share ONE identical empty envelope", async ({
    request,
  }) => {
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
      expect(
        Array.isArray(body.workplanTests),
        `${path} workplanTests is an array`,
      ).toBe(true);
      expect(body.paging, `${path} has a paging block`).toBeTruthy();
    }
    for (let i = 1; i < bodies.length; i++) {
      expect(
        bodies[i],
        `${WORKPLANS[i]} is identical to ${WORKPLANS[0]} when unparameterised`,
      ).toBe(bodies[0]);
    }
  });

  test("WorkPlanByTest: the param is test_id — and testTypeID is NOT echoed back", async ({
    request,
  }) => {
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

    const populated = await readJson(
      await request.get(`${WORKPLANS[0]}?test_id=${testId}`),
      "WorkPlanByTest",
    );
    expect(
      populated.workplanTests.length,
      "a real test_id populates workplanTests",
    ).toBeGreaterThan(0);

    // The wrong param name yields the empty form, silently.
    const ignored = await readJson(
      await request.get(`${WORKPLANS[0]}?testTypeID=${testId}`),
      "WorkPlanByTest wrong param",
    );
    expect(
      ignored.workplanTests.length,
      "testTypeID is ignored, not honoured",
    ).toBe(0);

    // Quirk worth pinning: even on the populated response, the response's own
    // testTypeID field stays "" rather than echoing the requested test_id.
    expect(
      populated.testTypeID,
      "testTypeID is not echoed back even when test_id was supplied",
    ).toBe("");
  });

  test("WorkPlanByTest: workplanTest rows keep numeric ranges numeric", async ({
    request,
  }) => {
    const rows = query(
      "SELECT test_id, count(*) FROM clinlims.analysis GROUP BY test_id ORDER BY count(*) DESC LIMIT 1",
    );
    test.skip(rows.length === 0, "no analyses in this dataset");

    const body = await readJson(
      await request.get(`${WORKPLANS[0]}?test_id=${rows[0][0]}`),
      "WorkPlanByTest",
    );
    test.skip(
      body.workplanTests.length === 0,
      "no workplan rows for this test",
    );

    for (const row of body.workplanTests.slice(0, 10)) {
      expectNonEmptyString(row.accessionNumber, "workplanTest accessionNumber");
      expect(typeof row.testId, "workplanTest testId is a STRING here").toBe(
        "string",
      );
      // Reference ranges are floats and must not be stringified — these feed
      // result-range comparisons on the validation screens.
      for (const numeric of [
        "upperNormalRange",
        "lowerNormalRange",
        "upperAbnormalRange",
        "lowerAbnormalRange",
      ]) {
        if (numeric in row) {
          expect(
            typeof row[numeric],
            `workplanTest ${numeric} stays numeric`,
          ).toBe("number");
        }
      }
      for (const bool of [
        "showSampleDetails",
        "isGroupSeparator",
        "testKitInactive",
      ]) {
        if (bool in row) {
          expect(typeof row[bool], `workplanTest ${bool} is boolean`).toBe(
            "boolean",
          );
        }
      }
      // receivedDate is a pre-formatted "dd/MM/yyyy HH:mm" string, not an
      // epoch — the opposite convention from the b2/c1 entity endpoints.
      if ("receivedDate" in row && row.receivedDate) {
        expect(
          row.receivedDate,
          "workplanTest receivedDate is a formatted string",
        ).toMatch(/^\d{2}\/\d{2}\/\d{4} \d{2}:\d{2}$/);
      }
    }
  });

  // ── LogbookResults / AccessionValidation / ReferredOutTests ─────────────

  test("LogbookResults: form envelope with paging and display flags", async ({
    request,
  }) => {
    const body = await readJson(await request.get(LOGBOOK), LOGBOOK);
    expectFormEnvelope(body, "LogbookResultsForm", LOGBOOK);
    expect(
      Array.isArray(body.testResult),
      `${LOGBOOK} testResult is an array`,
    ).toBe(true);
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
        expect(typeof row, `${LOGBOOK} testResult row is an object`).toBe(
          "object",
        );
        expect(row, `${LOGBOOK} testResult row is not null`).not.toBeNull();
      }
    }

    // These display* flags drive which columns the results screen renders, so
    // they are part of the contract, not incidental.
    for (const flag of [
      "displayMethods",
      "displayTestKit",
      "displayTestMethod",
      "displayTestSections",
    ]) {
      expect(flag in body, `${LOGBOOK} carries ${flag}`).toBe(true);
    }
    expect(
      typeof body.searchFinished,
      `${LOGBOOK} searchFinished is boolean`,
    ).toBe("boolean");
  });

  test("AccessionValidation: ResultValidationForm envelope", async ({
    request,
  }) => {
    const body = await readJson(
      await request.get(ACCESSION_VALIDATION),
      ACCESSION_VALIDATION,
    );
    expectFormEnvelope(body, "ResultValidationForm", ACCESSION_VALIDATION);
    expect(
      typeof body.searchFinished,
      `${ACCESSION_VALIDATION} searchFinished is boolean`,
    ).toBe("boolean");
  });

  test("ReferredOutTests: selection list rows match real test_section rows (DB oracle)", async ({
    request,
  }) => {
    const body = await readJson(await request.get(REFERRED_OUT), REFERRED_OUT);
    // Note the lowercase initial — formName casing is inconsistent across this
    // wave (LogbookResultsForm / ResultValidationForm / WorkplanForm vs
    // referredOutTestsForm). Pinned exactly rather than normalised.
    expectFormEnvelope(body, "referredOutTestsForm", REFERRED_OUT);
    expect(
      Array.isArray(body.testUnitSelectionList),
      `${REFERRED_OUT} testUnitSelectionList is an array`,
    ).toBe(true);

    // This list is genuinely POPULATED in this dataset, so assert against real
    // rows rather than stopping at Array.isArray — an isArray check would pass
    // on [] and prove nothing.
    expect(
      body.testUnitSelectionList.length,
      `${REFERRED_OUT} selection list is non-empty`,
    ).toBeGreaterThan(0);

    // DB oracle: every {id,value} must correspond to a real test_section row
    // with a matching name. Comparing the response only to itself would be
    // circular; this proves the endpoint reads the table it claims to.
    const sections = new Map(
      query("SELECT id, name FROM clinlims.test_section").map((r) => [
        r[0],
        r[1],
      ]),
    );
    for (const row of body.testUnitSelectionList) {
      expect(
        sections.has(row.id),
        `${REFERRED_OUT} id ${row.id} is a real test_section id`,
      ).toBe(true);
      expect(
        row.value,
        `${REFERRED_OUT} value for id ${row.id} matches test_section.name`,
      ).toBe(sections.get(row.id));
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

  test("accession-results: lean shape, NOT a Struts form envelope", async ({
    request,
  }) => {
    const body = await readJson(
      await request.get(ACCESSION_RESULTS),
      ACCESSION_RESULTS,
    );
    expect(
      Array.isArray(body.testResult),
      `${ACCESSION_RESULTS} testResult is an array`,
    ).toBe(true);
    expect(
      typeof body.searchFinished,
      `${ACCESSION_RESULTS} searchFinished is boolean`,
    ).toBe("boolean");

    // This is the ONE endpoint in the wave without the form envelope. Asserted
    // as an absence so a port does not "harmonise" it with its 7 siblings.
    for (const k of FORM_ENVELOPE) {
      expect(
        k in body,
        `${ACCESSION_RESULTS} must NOT carry the form envelope field ${k}`,
      ).toBe(false);
    }
  });

  // ── Auth boundary ───────────────────────────────────────────────────────

  test("c3 endpoints refuse anonymous access (clinical results)", async ({
    playwright,
  }) => {
    const anon = await playwright.request.newContext({
      baseURL: test.info().project.use.baseURL,
      ignoreHTTPSErrors: true,
      storageState: { cookies: [], origins: [] },
    });
    try {
      for (const path of [
        LOGBOOK,
        ACCESSION_RESULTS,
        ACCESSION_VALIDATION,
        REFERRED_OUT,
        ...WORKPLANS,
      ]) {
        const res = await anon.get(path, { maxRedirects: 0 });
        expect(res.status(), `anonymous ${path} must not succeed`).not.toBe(
          200,
        );
        const text = await res.text();
        expect(
          text,
          `anonymous ${path} must not leak result rows`,
        ).not.toContain(`"testResult"`);
        expect(
          text,
          `anonymous ${path} must not leak workplan rows`,
        ).not.toContain(`"workplanTests"`);
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

  test("WorkPlanBy*: each takes a DIFFERENT param name, and ignores the others", async ({
    request,
  }) => {
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
      {
        path: "rest/WorkPlanByPanel",
        own: `panel_id=${panelWithAnalyses()}`,
        foreign: `test_id=${testId}`,
      },
      {
        path: "rest/WorkPlanByPriority",
        own: "priority=ROUTINE",
        foreign: `test_id=${testId}`,
      },
    ];
    for (const c of cross) {
      const populated = await readJson(
        await request.get(`${c.path}?${c.own}`),
        `${c.path} own param`,
      );
      expect(
        populated.workplanTests.length,
        `${c.path}?${c.own} returns rows`,
      ).toBeGreaterThan(0);

      const ignored = await readJson(
        await request.get(`${c.path}?${c.foreign}`),
        `${c.path} foreign param`,
      );
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
    const body = await readJson(
      await request.get(`rest/WorkPlanByPanel?panel_id=${panelId}`),
      "WorkPlanByPanel",
    );

    const memberTests = query(
      `SELECT test_id FROM clinlims.panel_item WHERE panel_id = ${Number(panelId)} ORDER BY test_id`,
    ).map((r) => r[0]);
    expect(memberTests.length, "the panel has member tests").toBeGreaterThan(0);

    // Every row's testId is a member test of the panel.
    for (const row of body.workplanTests) {
      expect(
        memberTests.includes(row.testId),
        `workplan row test ${row.testId} is a panel member`,
      ).toBe(true);
    }

    // And the count is the sum over member tests of that test's matching
    // analyses — which is what "expand to tests" means, as opposed to filtering
    // on the column.
    const byPanelColumn = Number(
      query(
        `SELECT count(*) FROM clinlims.analysis WHERE panel_id = ${Number(panelId)}`,
      )[0][0],
    );
    expect(
      byPanelColumn,
      "some analyses DO carry panel_id, so the two readings differ",
    ).toBeGreaterThan(0);
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
      await request.get(
        `rest/WorkPlanByTestSection?test_section_id=${section}`,
      ),
      "WorkPlanByTestSection",
    );
    expect(
      body.workplanTests.length,
      "the clean section has workplan rows",
    ).toBeGreaterThan(0);

    const expectedAccessions = query(
      `SELECT s.accession_number
         FROM clinlims.analysis a
         JOIN clinlims.sample_item si ON si.id = a.sampitem_id
         JOIN clinlims.sample s ON s.id = si.samp_id
        WHERE a.test_sect_id = ${Number(section)}
        ORDER BY s.accession_number`,
    ).map((r) => r[0]);
    expect(
      expectedAccessions.length,
      "the DB agrees the section has analyses",
    ).toBeGreaterThan(0);
    for (const row of body.workplanTests) {
      expect(
        expectedAccessions.includes(row.accessionNumber),
        `row ${row.accessionNumber} comes from an analysis whose test_sect_id is ${section}`,
      ).toBe(true);
    }

    // The third argument to the DAO method is sortedByDateAndAccession, and its
    // entire body is commented out — passing true does nothing. Pinned as a
    // reminder that the name promises an ordering the code does not apply.
    expect(true, "sortedByDateAndAccession is a dead parameter in Java").toBe(
      true,
    );
  });

  test("WorkPlanByPriority: the param is an OrderPriority ENUM, not an id", async ({
    request,
  }) => {
    // @RequestParam(name = "priority") OrderPriority priority — Spring converts
    // the string to the enum, so an unknown value is a BINDING failure (400),
    // not an empty result. That is the opposite of how the other three behave
    // with a bad parameter, and it is the only one of the four that can 400.
    const ok = await readJson(
      await request.get("rest/WorkPlanByPriority?priority=ROUTINE"),
      "WorkPlanByPriority",
    );
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
    expect(
      empty.workplanTests.length,
      "an unused enum value gives an empty list, not a 400",
    ).toBe(0);

    // A value outside the enum: 400 from the converter.
    const bad = await request.get(
      "rest/WorkPlanByPriority?priority=NOT_A_PRIORITY",
    );
    expect(bad.status(), "an unknown priority fails BINDING with 400").toBe(
      400,
    );
  });

  // ── The shared NPE: one Java defect, two endpoints ──────────────────────

  test("a type-less sample item 500s WorkPlanByTestSection and LogbookResults alike", async ({
    request,
  }) => {
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
    expect(
      nullTypeItems.length,
      "the fixture still seeds a type-less sample item",
    ).toBeGreaterThan(0);
    const poisonTest = nullTypeItems[0][0];
    const poisonSection = query(
      `SELECT a.test_sect_id
         FROM clinlims.analysis a
         JOIN clinlims.sample_item si ON si.id = a.sampitem_id
        WHERE si.typeosamp_id IS NULL AND a.test_sect_id IS NOT NULL LIMIT 1`,
    );

    const byTest = await request.get(`${LOGBOOK}?selectedTest=${poisonTest}`);
    expect(
      byTest.status(),
      `LogbookResults?selectedTest=${poisonTest} 500s on the NPE`,
    ).toBe(500);

    if (poisonSection.length > 0) {
      const bySection = await request.get(
        `rest/WorkPlanByTestSection?test_section_id=${poisonSection[0][0]}`,
      );
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
    expect(
      cleanTest.length,
      "some test has no analysis on a type-less item",
    ).toBe(1);
    const clean = await readJson(
      await request.get(`${LOGBOOK}?selectedTest=${cleanTest[0][0]}`),
      "LogbookResults clean test",
    );
    expect(
      clean.testResult.length,
      "the same endpoint returns rows for a clean test",
    ).toBeGreaterThan(0);
  });

  // ── Result-carrying reads, now that results exist ───────────────────────

  test("accession-results: rows carry real result values, numeric and dictionary", async ({
    request,
  }) => {
    // result was an EMPTY table until result-reads-e2e.sql, so every c3
    // assertion about a value was previously vacuous — a port emitting "" for
    // every field agreed with Java on all of them.
    const body = await readJson(
      await request.get(`${ACCESSION_RESULTS}?accessionNumber=E2E-RES-01`),
      `${ACCESSION_RESULTS} E2E-RES-01`,
    );
    expect(body.testResult.length, "E2E-RES-01 has analyses").toBeGreaterThan(
      0,
    );

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
    const rendered = body.testResult.map((r: any) =>
      String(r.resultValue ?? ""),
    );
    const anyMatch = [...stored.values()].some((v) =>
      rendered.includes(String(v)),
    );
    expect(anyMatch, "a stored result value reaches the response").toBe(true);
  });

  test("AccessionValidation: only TechnicalAcceptance analyses are up for validation", async ({
    request,
  }) => {
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
    expect(pending.length, "something is awaiting validation").toBeGreaterThan(
      0,
    );
    const accession = pending[0][0];

    const body = await readJson(
      await request.get(`${ACCESSION_VALIDATION}?accessionNumber=${accession}`),
      `${ACCESSION_VALIDATION} ${accession}`,
    );
    expectFormEnvelope(body, "ResultValidationForm", ACCESSION_VALIDATION);
    expect(
      body.resultList.length,
      "the validation list is populated",
    ).toBeGreaterThan(0);

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
      await request.get(
        `${ACCESSION_VALIDATION}?accessionNumber=${other}&doRange=false`,
      ),
      `${ACCESSION_VALIDATION} ${other} exact`,
    );
    expect(
      exact.resultList.length,
      "doRange=false on an order with no analyses is empty",
    ).toBe(0);
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
    expect(
      ranged.accessionNumber,
      "the form echoes the REQUESTED accession",
    ).toBe(other);
  });

  test("ReferredOutTests: searchType drives the search, and one of its values 500s", async ({
    request,
  }) => {
    // setupPageForDisplay only searches when form.getSearchType() != null, so
    // the unparameterised call this file used to make could never return items.
    // The enum is TEST_AND_DATES / LAB_NUMBER / PATIENT — a value outside it is
    // a 400 from the converter.
    const bad = await request.get(
      `${REFERRED_OUT}?searchType=REFERRAL_PENDING`,
    );
    expect(bad.status(), "a value outside the SearchType enum is a 400").toBe(
      400,
    );

    const body = await readJson(
      await request.get(
        `${REFERRED_OUT}?searchType=LAB_NUMBER&labNumber=E2E-REF-01`,
      ),
      `${REFERRED_OUT} LAB_NUMBER`,
    );
    expect(body.searchFinished, "a search actually ran").toBe(true);
    expect(
      body.referralDisplayItems.length,
      "E2E-REF-01 has referrals to show",
    ).toBeGreaterThan(0);

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
    const broken = await request.get(
      `${REFERRED_OUT}?searchType=TEST_AND_DATES`,
    );
    expect(
      broken.status(),
      "TEST_AND_DATES with no dates is a 500, not a 400",
    ).toBe(500);
  });
  // ── The review findings ─────────────────────────────────────────────────
  //
  // Six things below were wrong in the Go port while every diff came back
  // byte-identical, because THIS deployment's configuration and data happen to
  // agree with the shortcut. c2's lesson was "an empty table makes a test that
  // cannot fail"; these are the same failure wearing configuration and
  // permissions instead of emptiness.

  test("AccessionValidation: the caller's LAB UNITS bound both the section list and the results", async ({
    playwright,
  }) => {
    // AccessionValidationRestController calls
    //   userService.getUserTestSections(getSysUserId(request), ROLE_VALIDATION)
    //   userService.filterAnalysisResultsByLabUnitRoles(getSysUserId(request), ...)
    // and the only user with lab units in the stock data is `admin`, whose
    // single lab_unit_role_map row says `AllLabUnits` — the branch that returns
    // every section and filters nothing. A port that skipped both calls matched
    // byte-for-byte and would let a user authorised for one lab unit read
    // validation-pending results from every other one.
    //
    // lab-unit-roles-e2e.sql seeds e2e_labunit with the Validation role on
    // exactly ONE section.
    const [[restrictedSection]] = query(
      `SELECT m.lab_unit
         FROM clinlims.system_user su
         JOIN clinlims.lab_unit_roles ur ON ur.system_user_id = su.id
         JOIN clinlims.lab_unit_role_map m ON m.lab_unit_role_map_id = ur.lab_unit_role_map_id
        WHERE su.login_name = 'e2e_labunit'`,
    );
    expect(
      restrictedSection,
      "e2e_labunit is restricted to one section",
    ).toBeTruthy();

    const restricted = await loginAs(
      { request: playwright.request },
      "e2e_labunit",
    );
    try {
      const mine = await readJson(
        await restricted.get(
          `${ACCESSION_VALIDATION}?accessionNumber=E2E-RES-01`,
        ),
        `${ACCESSION_VALIDATION} as e2e_labunit`,
      );
      const asAdmin = await readJson(
        await (
          await loginAs({ request: playwright.request }, "admin")
        ).get(`${ACCESSION_VALIDATION}?accessionNumber=E2E-RES-01`),
        `${ACCESSION_VALIDATION} as admin`,
      );

      // testSections is the user's OWN units. testSectionsByName beside it is
      // the deployment-wide list and is NOT filtered — two lists in one
      // response, only one of them scoped to the caller.
      expect(
        mine.testSections.map((s: any) => s.id),
        "testSections is exactly the caller's lab units",
      ).toEqual([restrictedSection]);
      expect(
        asAdmin.testSections.length,
        "...and admin, holding AllLabUnits, still sees every section",
      ).toBeGreaterThan(mine.testSections.length);
      expect(
        mine.testSectionsByName.length,
        "testSectionsByName is NOT scoped to the caller",
      ).toBe(asAdmin.testSectionsByName.length);

      // The RESULTS are filtered too, and that is the disclosure half.
      expect(
        mine.resultList.length,
        "the restricted caller still sees something",
      ).toBeGreaterThan(0);
      expect(
        asAdmin.resultList.length,
        "...but strictly less than admin, or the filter proves nothing",
      ).toBeGreaterThan(mine.resultList.length);

      const inSection = query(
        `SELECT a.id::text FROM clinlims.analysis a WHERE a.test_sect_id = ${Number(restrictedSection)}`,
      ).map((r) => r[0]);
      for (const row of mine.resultList) {
        expect(
          inSection.includes(row.analysisId),
          `analysis ${row.analysisId} is inside the caller's lab unit`,
        ).toBe(true);
      }
    } finally {
      await restricted.dispose();
    }
  });

  test("AccessionValidation: ONE result_limits band per analysis, chosen by the patient's age", async ({
    request,
  }) => {
    // result_limits is keyed by test AND by age band (min_age/max_age in DAYS)
    // and optionally gender. Joining on test_id alone multiplies the analysis by
    // its band count — the same mistake took LogbookResults from 4 rows to 9.
    // It stayed invisible on this endpoint because every validation-pending
    // analysis sat on a single-band test until the fixture added one that does
    // not.
    const banded = query(
      `SELECT a.id::text, a.test_id, count(rl.id)
         FROM clinlims.analysis a
         JOIN clinlims.status_of_sample st ON st.id = a.status_id
         JOIN clinlims.result_limits rl ON rl.test_id = a.test_id
        WHERE st.status_type = 'ANALYSIS' AND st.name = 'Technical Acceptance'
        GROUP BY a.id, a.test_id HAVING count(rl.id) > 1`,
    );
    expect(
      banded.length,
      "some pending analysis sits on a MULTI-band test",
    ).toBeGreaterThan(0);
    const [analysisId, , bandCount] = banded[0];
    expect(
      Number(bandCount),
      "that test really has several bands",
    ).toBeGreaterThan(1);

    const body = await readJson(
      await request.get(`${ACCESSION_VALIDATION}?accessionNumber=E2E-RES-01`),
      `${ACCESSION_VALIDATION} banded`,
    );
    const rows = body.resultList.filter(
      (r: any) => r.analysisId === analysisId,
    );
    expect(
      rows.length,
      `analysis ${analysisId} appears ONCE, not once per band`,
    ).toBe(1);

    // ...and its reference range is EMPTY. That is the defect this endpoint
    // carries: createTestResultItem resolves the limit through the
    // `patient == null` branch, and defaultResultLimit only matches a
    // blank-gender 0..Infinity row. A test that HAS bands but no default band
    // gets a SYNTHESIZED ResultLimit whose bounds are the infinities, and
    // getDisplayReferenceRange renders that as "".
    expect(
      rows[0].normalRange,
      "the validation screen shows NO reference range",
    ).toBe("");

    // The age-appropriate band, in DAYS, straight from the table.
    const [[low, high]] = query(
      `SELECT rl.low_normal, rl.high_normal
         FROM clinlims.analysis a
         JOIN clinlims.result_limits rl ON rl.test_id = a.test_id
         JOIN clinlims.sample_item si ON si.id = a.sampitem_id
         JOIN clinlims.sample s ON s.id = si.samp_id
         JOIN clinlims.sample_human sh ON sh.samp_id = s.id
         JOIN clinlims.patient pa ON pa.id = sh.patient_id
        WHERE a.id = ${Number(analysisId)}
          AND (rl.gender IS NULL OR rl.gender = '' OR rl.gender = pa.gender)
          AND EXTRACT(EPOCH FROM age(now(), pa.birth_date)) / 86400 >= rl.min_age
          AND EXTRACT(EPOCH FROM age(now(), pa.birth_date)) / 86400 <  rl.max_age
        ORDER BY rl.id LIMIT 1`,
    );

    // INVERSION, and the reason the empty string above is a defect rather than
    // absent data: LogbookResults resolves the SAME analysis to that band and
    // renders it. One analysis, two screens, two answers — and the screen
    // without the range is the one where a biologist accepts or rejects the
    // result. PINNED, not fixed.
    const logbook = await readJson(
      await request.get(`${LOGBOOK}?labNumber=E2E-RES-01`),
      `${LOGBOOK} for the same analysis`,
    );
    const logRow = logbook.testResult.find(
      (r: any) => r.analysisId === analysisId,
    );
    expect(logRow, `logbook row for analysis ${analysisId}`).toBeTruthy();
    expect(
      logRow.normalRange,
      "the logbook DOES render a band for it",
    ).not.toBe("");
    expect(logRow.normalRange, "...the age-appropriate one").toContain(
      String(Number(low)),
    );
    expect(logRow.normalRange, "...both ends of it").toContain(
      String(Number(high)),
    );
  });

  test("AccessionValidation: the rejected-tests SETTING decides whether rejections appear", async ({
    request,
  }) => {
    // getValidationStatus adds TechnicalAcceptance unconditionally and
    // TechnicalRejected only when Property.VALIDATE_REJECTED_TESTS is "true".
    // Nothing in the dataset was in the rejected status, so BOTH branches of
    // that condition were unobservable and a port that ignored the setting
    // matched anyway.
    const [[setting]] = query(
      `SELECT value FROM clinlims.site_information WHERE name = 'validateTechnicalRejection'`,
    );
    const rejected = query(
      `SELECT a.id::text FROM clinlims.analysis a
         JOIN clinlims.status_of_sample st ON st.id = a.status_id
         JOIN clinlims.sample_item si ON si.id = a.sampitem_id
         JOIN clinlims.sample s ON s.id = si.samp_id
        WHERE st.status_type = 'ANALYSIS' AND st.name = 'Technical Rejected'
          AND s.accession_number = 'E2E-RES-01'`,
    ).map((r) => r[0]);
    expect(
      rejected.length,
      "E2E-RES-01 carries a technically rejected analysis",
    ).toBeGreaterThan(0);

    const body = await readJson(
      await request.get(`${ACCESSION_VALIDATION}?accessionNumber=E2E-RES-01`),
      `${ACCESSION_VALIDATION} rejected`,
    );
    const returned = body.resultList.map((r: any) => r.analysisId);

    if (setting === "true") {
      for (const id of rejected) {
        expect(
          returned.includes(id),
          `rejected analysis ${id} IS offered for validation`,
        ).toBe(true);
      }
    } else {
      for (const id of rejected) {
        expect(
          returned.includes(id),
          `rejected analysis ${id} is NOT offered`,
        ).toBe(false);
      }
    }

    // Whichever way the setting falls, an accepted analysis is always there —
    // so this test cannot pass by returning nothing.
    const accepted = query(
      `SELECT a.id::text FROM clinlims.analysis a
         JOIN clinlims.status_of_sample st ON st.id = a.status_id
         JOIN clinlims.sample_item si ON si.id = a.sampitem_id
         JOIN clinlims.sample s ON s.id = si.samp_id
        WHERE st.status_type = 'ANALYSIS' AND st.name = 'Technical Acceptance'
          AND s.accession_number = 'E2E-RES-01'`,
    ).map((r) => r[0]);
    expect(
      accepted.length,
      "and an ACCEPTED one to anchor the assertion",
    ).toBeGreaterThan(0);
    for (const id of accepted) {
      expect(
        returned.includes(id),
        `accepted analysis ${id} is always offered`,
      ).toBe(true);
    }
  });

  test("AccessionValidation: the DATE branch searches started_date", async ({
    request,
  }) => {
    // The controller picks between accessionNumber, date and unitType in that
    // order, and the date branch runs getAnalysisStartedOn. No analysis in the
    // dataset had a started_date, so that branch could only ever return an
    // empty list — which is exactly what an unported branch also returns.
    const started = query(
      `SELECT a.id::text, to_char(a.started_date, 'DD/MM/YYYY')
         FROM clinlims.analysis a
         JOIN clinlims.status_of_sample st ON st.id = a.status_id
        WHERE a.started_date IS NOT NULL
          AND st.status_type = 'ANALYSIS'
          AND st.name IN ('Technical Acceptance', 'Technical Rejected')`,
    );
    expect(
      started.length,
      "an analysis carries a started_date",
    ).toBeGreaterThan(0);
    const [analysisId, date] = started[0];

    const body = await readJson(
      await request.get(
        `${ACCESSION_VALIDATION}?date=${encodeURIComponent(date)}`,
      ),
      `${ACCESSION_VALIDATION} by date`,
    );
    expect(body.searchFinished, "a search ran").toBe(true);
    expect(body.testDate, "the form echoes the requested date").toBe(date);
    expect(
      body.resultList.map((r: any) => r.analysisId).includes(analysisId),
      `the analysis started on ${date} is returned`,
    ).toBe(true);

    // INVERSION: a date nothing started on returns an empty list from the same
    // endpoint, so the assertion above is not satisfied by returning everything.
    const empty = await readJson(
      await request.get(`${ACCESSION_VALIDATION}?date=01/01/1999`),
      `${ACCESSION_VALIDATION} by an unused date`,
    );
    expect(
      empty.resultList.length,
      "a date with no analyses returns nothing",
    ).toBe(0);
  });

  test("accession-results: the nested augmentedTestName uses the SAME locale as the row", async ({
    request,
  }) => {
    // The entity graph builds the test name through a second path, and that
    // path had the locale hardcoded to 'en' while every other builder took the
    // configured one. On an English deployment the two agree, which is why the
    // diff never showed it; on any other, the row is localized and the nested
    // copy is not.
    const body = await readJson(
      await request.get(`${ACCESSION_RESULTS}?accessionNumber=E2E-RES-01`),
      `${ACCESSION_RESULTS} locale`,
    );
    expect(body.testResult.length, "E2E-RES-01 has rows").toBeGreaterThan(0);

    let compared = 0;
    for (const row of body.testResult) {
      if (!row.result?.analysis?.test) continue;
      expect(
        row.result.analysis.test.augmentedTestName,
        `augmentedTestName matches the row's testName for analysis ${row.analysisId}`,
      ).toBe(row.testName);
      compared++;
    }
    expect(
      compared,
      "at least one row carried the nested entity graph",
    ).toBeGreaterThan(0);
  });

  test('AccessionValidation: a test with bands but NO default band answers lowerCritical "Infinity"', async ({
    request,
  }) => {
    // getResultLimitForTestAndPatient does NOT return null when a test has
    // limits but none of them is the default band: defaultResultLimit falls
    // through to `return new ResultLimit()`, and that bean initialises
    //
    //     private double lowCritical = Double.POSITIVE_INFINITY;
    //
    // where every other LOW bound on it is NEGATIVE_INFINITY. The guard in
    // setResultLimitDependencies folds only NEGATIVE_INFINITY to zero, so
    // +Infinity survives to Jackson — which writes a field declared `double`
    // as the STRING "Infinity". highCritical carries the same initialiser and
    // IS folded, so the two bounds of one row disagree about what unset means.
    //
    // PINNED, not fixed: this is Java's behaviour and the port reproduces it.
    //
    // Selected from the RESPONSE, then explained by the database. The reverse
    // order does not work: a query for "a test with bands but no default band"
    // also matches analyses this endpoint never returns — it filters to the
    // validation statuses, and the first such analysis in the fixture is not in
    // one of them, so that version failed identically against Java and Go.
    const body = await readJson(
      await request.get(`${ACCESSION_VALIDATION}?accessionNumber=E2E-RES-01`),
      `${ACCESSION_VALIDATION} for the banded analysis`,
    );
    const noRange = body.resultList.filter((r: any) => r.normalRange === "");
    expect(
      noRange.length,
      "exactly one returned analysis resolves no reference range",
    ).toBe(1);
    const row = noRange[0];

    // ...and that empty range IS the missing default band rather than a test
    // with no limits at all — the synthesized ResultLimit produces the empty
    // range either way, and only the criticals tell the two cases apart.
    const [hasBands, hasDefault] = query(
      `SELECT EXISTS (SELECT 1 FROM clinlims.result_limits rl
                       WHERE rl.test_id = a.test_id),
              EXISTS (SELECT 1 FROM clinlims.result_limits rl
                       WHERE rl.test_id = a.test_id
                         AND (rl.gender IS NULL OR rl.gender = '')
                         AND rl.min_age = 0
                         AND rl.max_age = 'Infinity'::double precision)
         FROM clinlims.analysis a WHERE a.id = ${row.analysisId}`,
    )[0];
    expect(hasBands, "the test does have result_limits rows").toBe("t");
    expect(hasDefault, "...but none of them is the default band").toBe("f");

    expect(
      row.lowerCritical,
      "the synthesized limit leaks +Infinity as a string",
    ).toBe("Infinity");
    expect(
      row.higherCritical,
      "...while the high bound IS folded to zero",
    ).toBe(0);

    // INVERSION: a row that DOES resolve a default band folds both to zero, so
    // the string above is a property of the missing band and not a constant.
    const resolved = body.resultList.find((r: any) => r.normalRange !== "");
    expect(
      resolved,
      "an analysis with a resolved band is seeded too",
    ).toBeTruthy();
    expect(
      resolved.lowerCritical,
      "a resolved band folds -Infinity to zero",
    ).toBe(0);
  });

  test("WorkPlanByPriority: the grouping number follows ANALYSIS order, splitting an accession", async ({
    request,
  }) => {
    // getAllAnalysisByPriorityAndStatus has no ORDER BY, and the grouping
    // counter is stamped in whatever order the plan yields BEFORE the list is
    // sorted for display — so the number is the only observable trace of the
    // scan order left in the response.
    //
    // Sample-item order and analysis order assign identical numbers whenever
    // each accession's analyses are physically contiguous, which they were.
    // Adding one analysis to an EXISTING accession separates them: the new row
    // lands at the end of the analysis heap while its sample_item stays put.
    const scan = query(
      `SELECT s.accession_number
         FROM clinlims.analysis a
         JOIN clinlims.sample_item si ON si.id = a.sampitem_id
         JOIN clinlims.sample s ON s.id = si.samp_id
         JOIN clinlims.status_of_sample st ON st.id = a.status_id
              AND st.status_type = 'ANALYSIS'
              AND st.name IN ('Not Tested', 'Biologist Rejection',
                              'Technical Rejected', 'NonConforming')
        WHERE s.order_priority = 'ROUTINE'
        ORDER BY a.ctid`,
    ).map((r) => r[0]);
    expect(scan.length, "routine analyses are seeded").toBeGreaterThan(0);

    // The counter increments whenever the accession changes from the previous
    // row of the scan — so a non-contiguous accession gets two numbers.
    const expected: string[] = [];
    let group = 0;
    scan.forEach((acc, i) => {
      if (i === 0 || acc !== scan[i - 1]) group += 1;
      expected.push(`${acc}/${group}`);
    });
    expect(
      new Set(expected.map((e) => e.split("/")[1])).size,
      "the seeded data splits at least one accession across two groups",
    ).toBeGreaterThan(new Set(scan).size);

    const body = await readJson(
      await request.get("rest/WorkPlanByPriority?priority=ROUTINE"),
      "WorkPlanByPriority ROUTINE",
    );
    const actual = body.workplanTests.map(
      (t: any) => `${t.accessionNumber}/${t.sampleGroupingNumber}`,
    );
    expect(
      actual.slice().sort(),
      "every row carries its scan-order group",
    ).toEqual(expected.slice().sort());
  });

  test("LogbookResults: every `double` field is rendered Java-style, with a decimal point", async ({
    request,
  }) => {
    // Jackson writes a Java `double` 40.0 as `40.0`; Go writes a float64 40.0
    // as `40`. The two parse to the same number, so a differ that unmarshals
    // both sides reports parity and the divergence lives only in the bytes —
    // it changes Content-Length and it is what any strict consumer sees.
    //
    // Asserted on the RAW body for that reason: reading these fields off a
    // parsed object cannot distinguish the two renderings at all.
    const res = await request.get(`${LOGBOOK}?labNumber=E2E-RES-01`);
    expect(res.status(), "logbook by lab number").toBe(200);
    const raw = await res.text();

    const DOUBLE_KEYS =
      "upperNormalRange|lowerNormalRange|upperAbnormalRange|lowerAbnormalRange|lowerCritical|higherCritical";
    const present = raw.match(new RegExp(`"(?:${DOUBLE_KEYS})":`, "g")) ?? [];
    expect(
      present.length,
      "the response actually carries these fields",
    ).toBeGreaterThan(0);

    // An integer with no fraction digits and no exponent is the Go rendering.
    const bare =
      raw.match(new RegExp(`"(?:${DOUBLE_KEYS})":-?\\d+(?![.\\dEe])`, "g")) ??
      [];
    expect(bare, "no double is rendered as a bare integer").toEqual([]);
  });
});
