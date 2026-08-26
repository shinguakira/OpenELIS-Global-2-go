// §6 — c2: sample / order reads (Java baseline; Go parity gate once ported).
//
// Taxonomy Type C/D, Wave 4, branch migration/c2-sample-order-reads. Same
// contract as the c1 spec: written BEFORE the Go port so the port has an
// executable specification, and every expectation below was captured from the
// LIVE Java server rather than read off the source.
//
// NOT yet in playwright.config.ts's `go-parity` testMatch — no Go
// implementation exists to run it against. Add it there as c2 lands.
//
// ── SCOPE NOTE ─────────────────────────────────────────────────────────────
// Wave 4 lists 17 endpoints. This file covers the ones that are reachable and
// meaningfully assertable against the current dataset. Three groups are
// deliberately excluded, each for a stated reason (see the bottom of the
// file): the form-load endpoints, the binary attachment endpoints, and the
// mutating shipment endpoints that share the unassigned-sample controller.
//
// ── MIGRATION POLICY ───────────────────────────────────────────────────────
// This is a migration, not a bug-fix pass. Where Java is broken, these tests
// PIN the broken behavior so the port reproduces it knowingly. Fixing any of
// it is out of scope and must be raised separately with the maintainers.
import { test, expect } from "@playwright/test";
import { readJson, expectKeysWithin, expectNonEmptyString } from "../../fixtures/assert";
import { query } from "../../fixtures/db";

const ALL_BY_ACCESSION = "rest/sample/all-by-accession";
const UNASSIGNED_BY_ACCESSION = "rest/sample/unassigned-by-accession";
const ORDER_SEARCH = "rest/order/search";
const ORDER_DASHBOARD = "rest/order/dashboard";
const UNASSIGNED = "rest/unassigned-sample";
const PENDING_ANALYSIS = "rest/getPendingAnalysisForTestProvider";

/** A real accession number, read from the DB rather than hardcoded. */
function anyAccession(): string {
  const rows = query(
    "SELECT accession_number FROM clinlims.sample WHERE accession_number IS NOT NULL ORDER BY id LIMIT 1",
  );
  return rows[0][0];
}

test.describe("c2 — sample + order reads", () => {
  // ── rest/sample/all-by-accession/{accessionNumber} ──────────────────────

  test("all-by-accession: rows carry the requested accession and a numeric analysisId", async ({ request }) => {
    const accession = anyAccession();
    const body = await readJson(
      await request.get(`${ALL_BY_ACCESSION}/${encodeURIComponent(accession)}`),
      ALL_BY_ACCESSION,
    );
    expect(Array.isArray(body), `${ALL_BY_ACCESSION} is an array`).toBe(true);
    test.skip(body.length === 0, "this accession has no analyses");

    for (const row of body) {
      expectKeysWithin(
        row,
        ["id", "accessionNumber", "sampleType", "referralTest", "analysisId"],
        ["id", "accessionNumber"],
        `${ALL_BY_ACCESSION} row`,
      );
      // Note the type split: `id` and `analysisId` are NUMBERS here, unlike the
      // b1/b2 reference endpoints where every id is stringified. A port that
      // applies the usual strconv.FormatInt at the DTO boundary would emit
      // strings and diverge.
      expect(typeof row.id, `${ALL_BY_ACCESSION} id is a number, not a string`).toBe("number");
      if ("analysisId" in row) {
        expect(typeof row.analysisId, `${ALL_BY_ACCESSION} analysisId is a number`).toBe("number");
      }
      expect(row.accessionNumber, `${ALL_BY_ACCESSION} echoes the requested accession`).toBe(accession);
    }
  });

  test("all-by-accession: an unknown accession is a real 404", async ({ request }) => {
    // Not 200-with-[] — this endpoint distinguishes "no such order" from "an
    // order with no analyses". Worth pinning explicitly because the sibling
    // list endpoints in this same wave (unassigned-sample/items,
    // unassigned-sample/items/search) DO return 200 [] for a miss, and c1's
    // patient-id-documents does too. There is no single house rule; each
    // endpoint has to be checked.
    const res = await request.get(`${ALL_BY_ACCESSION}/NO_SUCH_ACCESSION_XYZ`);
    expect(res.status(), `${ALL_BY_ACCESSION} unknown accession`).toBe(404);
  });

  // ── rest/sample/unassigned-by-accession/{accessionNumber} ───────────────

  test("unassigned-by-accession: PERMANENTLY BROKEN in Java — always 500", async ({ request }) => {
    // This endpoint cannot succeed for ANY input. Its HQL
    // (SampleDAOImpl.getUnassignedSampleByAccessionNumber) references
    // `r.canceled`, a property that does not exist on
    // org.openelisglobal.referral.valueholder.Referral, so Hibernate throws
    // QueryException while PARSING the query — before it ever looks at data.
    // Confirmed from the server log, and confirmed behaviorally below against
    // a valid accession, a second valid accession, and a nonexistent one.
    //
    // The controller catches Exception and returns 500
    // (SampleRestController.java:82-84), so the `return notFound()` branch at
    // :77 is unreachable dead code.
    //
    // MIGRATION POLICY: this is pinned, NOT fixed. Porting it means either
    // reproducing a permanently-500 endpoint or consciously declining to port
    // it. Fixing the Java HQL is a separate maintenance task and explicitly
    // out of scope for the migration.
    const accession = anyAccession();
    for (const input of [accession, "NO_SUCH_ACCESSION_XYZ"]) {
      const res = await request.get(`${UNASSIGNED_BY_ACCESSION}/${encodeURIComponent(input)}`);
      expect(res.status(), `${UNASSIGNED_BY_ACCESSION}/${input} is 500 regardless of input`).toBe(500);
    }
  });

  // ── rest/order/search ───────────────────────────────────────────────────

  test("order/search: param is labNumber; returns the order-entry form envelope", async ({ request }) => {
    // The param is `labNumber`, NOT `accessionNumber` — an easy wrong guess,
    // and one that fails closed with a 400 rather than silently returning
    // everything.
    const accession = anyAccession();
    const body = await readJson(
      await request.get(`${ORDER_SEARCH}?labNumber=${encodeURIComponent(accession)}`),
      ORDER_SEARCH,
    );
    expect(body.sampleOrderItems, `${ORDER_SEARCH} has sampleOrderItems`).toBeTruthy();
    expect(body.sampleOrderItems.labNo, `${ORDER_SEARCH} labNo echoes the request`).toBe(accession);

    // This is a Type-D form load, not a lean read: it carries reference lists
    // (payment options, etc.) alongside the order data. A port must build the
    // whole envelope, not just the sample row.
    expect(Array.isArray(body.sampleOrderItems.paymentOptions), `${ORDER_SEARCH} paymentOptions is a list`).toBe(
      true,
    );
  });

  test("order/search: a missing labNumber is rejected with 400", async ({ request }) => {
    const res = await request.get(ORDER_SEARCH);
    expect(res.status(), `${ORDER_SEARCH} missing labNumber`).toBe(400);
  });

  // ── rest/order/dashboard ────────────────────────────────────────────────

  test("order/dashboard: paged envelope with typed order rows", async ({ request }) => {
    const body = await readJson(await request.get(ORDER_DASHBOARD), ORDER_DASHBOARD);
    expect(typeof body.pageSize, `${ORDER_DASHBOARD} pageSize is a number`).toBe("number");
    expect(Array.isArray(body.orders), `${ORDER_DASHBOARD} orders is an array`).toBe(true);
    test.skip(body.orders.length === 0, "no orders in this dataset");

    for (const order of body.orders.slice(0, 20)) {
      expectNonEmptyString(order.labNumber, `${ORDER_DASHBOARD} labNumber`);
      // Booleans must stay booleans — not "Y"/"N" strings, which is the
      // convention several other tables in this schema use.
      expect(typeof order.returnedFromQA, `${ORDER_DASHBOARD} returnedFromQA is boolean`).toBe("boolean");
      expect(typeof order.isExternal, `${ORDER_DASHBOARD} isExternal is boolean`).toBe("boolean");
      // lastUpdated here is a FORMATTED STRING ("2025-01-01 12:00:00.0"), not
      // the epoch-millis number that the b2/c1 entity endpoints emit for
      // lastupdated. Same concept, different field name AND different type —
      // pinned so a port does not standardise it.
      if ("lastUpdated" in order) {
        expect(typeof order.lastUpdated, `${ORDER_DASHBOARD} lastUpdated is a string, not epoch`).toBe("string");
      }
    }
  });

  // ── rest/unassigned-sample (the trailing-slash trap) ────────────────────

  test("unassigned-sample: bare path works, trailing slash 404s (Spring 6)", async ({ request }) => {
    // The controller uses a bare @GetMapping with no value, so the path is
    // exactly /rest/unassigned-sample. Spring 6 removed automatic
    // trailing-slash matching, so the /-suffixed form falls through to the
    // JSP error page and returns 404 text/html.
    //
    // Worth pinning because the migration plan's own endpoint list writes this
    // one WITH a trailing slash — following that list literally produces a
    // route nobody can call. (The a2 spec carries the same warning for
    // supportedlocales.)
    const ok = await request.get(UNASSIGNED);
    expect(ok.status(), `${UNASSIGNED} (no trailing slash) is 200`).toBe(200);
    expect(Array.isArray(await ok.json()), `${UNASSIGNED} returns an array`).toBe(true);

    const slash = await request.get(`${UNASSIGNED}/`);
    expect(slash.status(), `${UNASSIGNED}/ (trailing slash) is 404`).toBe(404);
  });

  test("unassigned-sample/items: array; search variant requires accessionNumber", async ({ request }) => {
    const items = await readJson(await request.get(`${UNASSIGNED}/items`), `${UNASSIGNED}/items`);
    expect(Array.isArray(items), `${UNASSIGNED}/items is an array`).toBe(true);

    const accession = anyAccession();
    const searched = await readJson(
      await request.get(`${UNASSIGNED}/items/search?accessionNumber=${encodeURIComponent(accession)}`),
      `${UNASSIGNED}/items/search`,
    );
    expect(Array.isArray(searched), `${UNASSIGNED}/items/search is an array`).toBe(true);

    // accessionNumber is a required @RequestParam here (no defaultValue), so
    // omitting it is a 400 — unlike /items, which takes none.
    const missing = await request.get(`${UNASSIGNED}/items/search`);
    expect(missing.status(), `${UNASSIGNED}/items/search without accessionNumber`).toBe(400);
  });

  test("unassigned-sample/count-by-facility: {count:n} matching the list length", async ({ request }) => {
    const orgs = query("SELECT id FROM clinlims.organization ORDER BY id LIMIT 1");
    test.skip(orgs.length === 0, "no organizations in this dataset");
    const facilityId = orgs[0][0];

    const counted = await readJson(
      await request.get(`${UNASSIGNED}/count-by-facility/${facilityId}`),
      `${UNASSIGNED}/count-by-facility`,
    );
    expect(Object.keys(counted), `${UNASSIGNED}/count-by-facility envelope`).toEqual(["count"]);
    expect(typeof counted.count, `${UNASSIGNED}/count-by-facility count is a number`).toBe("number");

    // Cross-check the count against the list endpoint for the same facility —
    // the two must agree, which no single-endpoint shape assertion can catch.
    const listed = await readJson(
      await request.get(`${UNASSIGNED}/by-facility/${facilityId}`),
      `${UNASSIGNED}/by-facility`,
    );
    expect(Array.isArray(listed), `${UNASSIGNED}/by-facility is an array`).toBe(true);
    expect(counted.count, "count-by-facility agrees with by-facility length").toBe(listed.length);
  });

  test("unassigned-sample/by-facility: a non-numeric facilityId is 400", async ({ request }) => {
    // facilityId binds as Integer, so Spring rejects a non-numeric value at
    // binding — contrast with c1's patient endpoints, where a String-bound
    // path variable against a varchar column simply matches nothing.
    const res = await request.get(`${UNASSIGNED}/by-facility/abc`);
    expect(res.status(), `${UNASSIGNED}/by-facility non-numeric id`).toBe(400);
  });

  // ── rest/getPendingAnalysisForTestProvider ──────────────────────────────

  test("getPendingAnalysisForTestProvider: required param enforced", async ({ request }) => {
    const res = await request.get(PENDING_ANALYSIS);
    expect(res.status(), `${PENDING_ANALYSIS} without its required param`).toBe(400);
  });

  // ── rest/order/{accessionNumber}/attachments ────────────────────────────

  test("order/{accession}/attachments: 200 [] when known, 404 {error} when not", async ({ request }) => {
    const accession = anyAccession();
    const body = await readJson(
      await request.get(`rest/order/${encodeURIComponent(accession)}/attachments`),
      "order attachments",
    );
    expect(Array.isArray(body), "order attachments is an array").toBe(true);

    // A known order with zero attachments -> 200 []; an unknown order -> 404
    // with a JSON error envelope {"error":"Order not found"}. Two distinct
    // shapes for two distinct conditions, and the 404 body is JSON rather than
    // the empty body most other 404s in this codebase return — so a port must
    // emit the envelope, not just the status.
    const unknown = await request.get("rest/order/NO_SUCH_ACCESSION_XYZ/attachments");
    expect(unknown.status(), "order attachments unknown accession is 404").toBe(404);
    expect(await unknown.json(), "order attachments 404 carries a JSON error envelope").toEqual({
      error: "Order not found",
    });
  });

  // ── Auth boundary ───────────────────────────────────────────────────────

  test("c2 endpoints refuse anonymous access", async ({ playwright }) => {
    // Sample/order data is clinical and patient-linked (the dashboard rows
    // carry patientName), so the auth boundary matters here as much as in c1.
    const anon = await playwright.request.newContext({
      baseURL: test.info().project.use.baseURL,
      ignoreHTTPSErrors: true,
      storageState: { cookies: [], origins: [] },
    });
    try {
      const accession = anyAccession();
      for (const path of [
        ORDER_DASHBOARD,
        `${ALL_BY_ACCESSION}/${encodeURIComponent(accession)}`,
        `${UNASSIGNED}/items`,
        `${ORDER_SEARCH}?labNumber=${encodeURIComponent(accession)}`,
      ]) {
        const res = await anon.get(path, { maxRedirects: 0 });
        expect(res.status(), `anonymous ${path} must not succeed`).not.toBe(200);
        const text = await res.text();
        expect(text, `anonymous ${path} must not leak a patient name`).not.toContain(`"patientName"`);
      }
    } finally {
      await anon.dispose();
    }
  });
});

// ── DELIBERATELY NOT COVERED (and why) ──────────────────────────────────────
//
// 1. MUTATING endpoints on the unassigned-sample controller —
//    POST/PUT assignSampleToBox, markSampleAsLost, cancelReferral
//    (UnassignedSampleRestController.java:126, :155, :184). They change
//    referral state and must never run from the read-only suite. They belong
//    to the shipment feature module (an h-* branch), not to c2.
//
// 2. Binary attachment endpoints —
//    rest/order/attachments/{id}/download and .../view. The dataset has no
//    attachments, so only the empty path could be exercised, and that path
//    tells us nothing about the Content-Type/disposition behavior that
//    actually matters. Covering these properly needs a seeded attachment
//    fixture, the same way c1 needed patient-media-e2e.sql for photos.
//
// 3. Type-D form loads — rest/GenericSampleOrder, rest/SamplePatientEntry,
//    rest/SampleEdit, rest/SampleBatchEntrySetup. These return large
//    form-backing envelopes assembled from many reference lists; pinning them
//    meaningfully is its own unit of work and they are closer to Type D than
//    to a sample/order read. order/search is included above as the one
//    representative of that shape.
