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

  test("ReferredOutTests: referredOutTestsForm envelope with its selection list", async ({ request }) => {
    const body = await readJson(await request.get(REFERRED_OUT), REFERRED_OUT);
    // Note the lowercase initial — formName casing is inconsistent across this
    // wave (LogbookResultsForm / ResultValidationForm / WorkplanForm vs
    // referredOutTestsForm). Pinned exactly rather than normalised.
    expectFormEnvelope(body, "referredOutTestsForm", REFERRED_OUT);
    expect(Array.isArray(body.testUnitSelectionList), `${REFERRED_OUT} testUnitSelectionList is an array`).toBe(
      true,
    );
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
});

// ── COVERAGE LIMITS (stated, not hidden) ────────────────────────────────────
//
// The dev dataset has analyses in only two statuses (32 rows at status 4, 3 at
// status 6) and no validated/released results, so the RESULT-BEARING paths of
// LogbookResults, AccessionValidation and accession-results return empty
// collections. Their envelopes are pinned; the populated row shapes are not
// reachable here and need seeded result data — the same kind of gap c1 had for
// patient media before patient-media-e2e.sql.
//
// These endpoints also accept search/filter params that this file does not
// exercise. They were left out deliberately rather than guessed at: the param
// names are NOT inferable from the response fields (test_id vs testTypeID
// above is the proof), so each needs its controller signature read before a
// meaningful assertion can be written.
