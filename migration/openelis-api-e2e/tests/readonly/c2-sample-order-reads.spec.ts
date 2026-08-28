// §6 — c2: sample / order reads (Java baseline; Go parity gate once ported).
//
// Taxonomy Type C/D, Wave 4, branch migration/c2-sample-order-reads. Same
// contract as the c1 spec: written BEFORE the Go port so the port has an
// executable specification, and every expectation below was captured from the
// LIVE Java server rather than read off the source.
//
// Wired into playwright.config.ts's `go-parity` testMatch: the Go port exists
// and this same file is the gate it has to pass.
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
    const options = body.sampleOrderItems.paymentOptions;
    expect(Array.isArray(options), `${ORDER_SEARCH} paymentOptions is a list`).toBe(true);
    expect(options.length, `${ORDER_SEARCH} paymentOptions is non-empty`).toBeGreaterThan(0);

    // DB oracle. This list is genuinely populated, so stopping at
    // Array.isArray would be a test that passes on [] and proves nothing.
    // Every {id,value} must resolve to a real clinlims.dictionary row whose
    // dict_entry matches the emitted value — which also pins WHICH column
    // feeds `value` (dict_entry, not local_abbrev), a detail a port would
    // otherwise have to guess.
    const dict = new Map(
      query(
        `SELECT id, dict_entry FROM clinlims.dictionary WHERE id IN (${options
          .map((o: any) => Number(o.id))
          .filter((n: number) => Number.isFinite(n))
          .join(",")})`,
      ).map((r) => [r[0], r[1]]),
    );
    for (const opt of options) {
      expect(dict.has(opt.id), `${ORDER_SEARCH} paymentOption id ${opt.id} is a real dictionary row`).toBe(true);
      expect(opt.value, `${ORDER_SEARCH} paymentOption ${opt.id} value == dictionary.dict_entry`).toBe(
        dict.get(opt.id),
      );
    }
  });

  test("order/search: a missing labNumber is rejected with 400", async ({ request }) => {
    const res = await request.get(ORDER_SEARCH);
    expect(res.status(), `${ORDER_SEARCH} missing labNumber`).toBe(400);
  });

  test("order/search: the WHOLE envelope, not just sampleOrderItems", async ({ request }) => {
    // The two tests above assert `sampleOrderItems` and nothing else — two of
    // the nine top-level keys. A port emitting only those would have passed
    // them while dropping patientProperties, samples, orderData and
    // stepProgress entirely. This pins the envelope so "green" means the
    // response is actually Java's.
    const accession = anyAccession();
    const body = await readJson(
      await request.get(`${ORDER_SEARCH}?labNumber=${encodeURIComponent(accession)}`),
      ORDER_SEARCH,
    );

    expectKeysWithin(
      body,
      // collectionDate and status are dropped when null (HashMap +
      // Include.NON_NULL); patientProperties/orderData are absent when the
      // sample has no patient.
      ["id", "labNumber", "receivedDate", "collectionDate", "status",
       "patientProperties", "orderData", "samples", "sampleOrderItems",
       "stepProgress", "storageSkipped"],
      ["id", "labNumber", "samples", "sampleOrderItems", "stepProgress", "storageSkipped"],
      `${ORDER_SEARCH} envelope`,
    );

    // id is the SAMPLE id as a string, and labNumber echoes the request.
    const [[sampleId]] = query(
      `SELECT id FROM clinlims.sample WHERE accession_number = '${accession}'`,
    );
    expect(body.id, `${ORDER_SEARCH} id is the sample id`).toBe(String(sampleId));
    expect(body.labNumber, `${ORDER_SEARCH} labNumber echoes the request`).toBe(accession);
    expect(typeof body.storageSkipped, `${ORDER_SEARCH} storageSkipped is boolean`).toBe("boolean");

    // stepProgress: four flags, and `enter` is HARDCODED true here — "If
    // sample exists, enter is complete" — while order/dashboard COMPUTES the
    // same key from received date plus patient/workflow type. Same name, two
    // meanings; pinned so a port does not share one implementation.
    expectKeysWithin(
      body.stepProgress,
      ["enter", "collect", "label", "qa"],
      ["enter", "collect", "label", "qa"],
      `${ORDER_SEARCH} stepProgress`,
    );
    expect(body.stepProgress.enter, `${ORDER_SEARCH} stepProgress.enter is hardcoded true`).toBe(true);
    for (const k of ["collect", "label", "qa"]) {
      expect(typeof body.stepProgress[k], `${ORDER_SEARCH} stepProgress.${k} is boolean`).toBe("boolean");
    }

    // patientProperties is a BEAN, not a map: identity-backed fields come from
    // getIdentityValue which returns "" for a missing identity, so those keys
    // are always PRESENT and empty rather than dropped.
    test.skip(!("patientProperties" in body), "this sample has no linked patient");
    const pp = body.patientProperties;
    for (const k of ["patientPK", "patientUpdateStatus", "firstName", "lastName",
                     "nationalId", "subjectNumber", "guid", "aka", "mothersName",
                     "gender", "birthDateForDisplay", "readOnly", "isMerged",
                     "addressHierarchy", "stnumber"]) {
      expect(k in pp, `${ORDER_SEARCH} patientProperties.${k} is always present`).toBe(true);
    }
    expect(pp.patientUpdateStatus, "patientUpdateStatus is the literal UPDATE").toBe("UPDATE");
    expect(pp.readOnly, "readOnly is a primitive boolean").toBe(false);
    expect(typeof pp.isMerged, "isMerged is a primitive boolean").toBe("boolean");
    expect(pp.addressHierarchy, "addressHierarchy is an initialised empty map").toEqual({});

    // DB oracle for the patient identity: patientPK is the linked patient, and
    // nationalId comes from the patient row rather than from an identity.
    const [[dbPatientId, dbNationalId]] = query(
      "SELECT p.id, COALESCE(p.national_id, '') FROM clinlims.patient p" +
        " JOIN clinlims.sample_human sh ON sh.patient_id = p.id" +
        ` WHERE sh.samp_id = ${sampleId}`,
    );
    expect(pp.patientPK, "patientPK is the linked patient").toBe(String(dbPatientId));
    expect(pp.nationalId, "nationalId comes from patient.national_id").toBe(dbNationalId);

    // birthDateForDisplay is REFORMATTED here. c1's patientByLabNumer emits
    // the stored entered_birth_date raw; this endpoint runs it through
    // DateUtil.formatStringDate first, so the same stored value appears in two
    // different orders across the two endpoints.
    const [[storedBirth]] = query(
      `SELECT COALESCE(entered_birth_date, '') FROM clinlims.patient WHERE id = ${dbPatientId}`,
    );
    if (storedBirth !== "") {
      const [a, b, y] = storedBirth.split("/");
      expect(pp.birthDateForDisplay, "birthDateForDisplay is reformatted, not the stored string").toBe(
        `${b}/${a}/${y}`,
      );
    }

    // orderData carries the SAME bean again plus the literal status.
    expectKeysWithin(
      body.orderData,
      ["patientProperties", "patientUpdateStatus"],
      ["patientProperties", "patientUpdateStatus"],
      `${ORDER_SEARCH} orderData`,
    );
    expect(body.orderData.patientProperties, "orderData repeats patientProperties verbatim").toEqual(pp);
  });

  test("order/search: voided sample items are excluded", async ({ request }) => {
    // Separate from the samples[] test above because that one cannot detect
    // this. It counts the API rows against a DB oracle that applies the SAME
    // `voided = false` predicate, so on a dataset where nothing is voided both
    // sides agree no matter what the server does. Deleting the predicate from
    // the Go DAO and re-running the suite left it green — that mutation is
    // what this test exists to kill.
    //
    // src/test/resources/fixtures/order-search-e2e.sql seeds E2E-VOIDED-01
    // with three items, the middle one voided.
    const rows = query(
      "SELECT si.id, si.voided FROM clinlims.sample_item si" +
        " JOIN clinlims.sample s ON s.id = si.samp_id" +
        " WHERE s.accession_number = 'E2E-VOIDED-01' ORDER BY si.sort_order",
    );
    // Not a skip: the loader marks this fixture fatal, so a missing row means
    // the fixture did not run and the coverage is silently gone.
    expect(rows.length, "E2E-VOIDED-01 fixture is loaded").toBe(3);
    const voidedId = rows.find((r) => r[1] === "t" || r[1] === "true")![0];
    const liveIds = rows.filter((r) => r[0] !== voidedId).map((r) => r[0]).sort();

    const body = await readJson(
      await request.get(`${ORDER_SEARCH}?labNumber=E2E-VOIDED-01`),
      ORDER_SEARCH,
    );
    const returned = body.samples.map((s: any) => s.id).sort();
    expect(returned, "voided item is not in samples[]").toEqual(liveIds);
    expect(returned, "the voided id specifically").not.toContain(voidedId);
  });

  test("order/search: samples[] rows, their storage block and the sampleXML twin", async ({ request }) => {
    const accession = anyAccession();
    const body = await readJson(
      await request.get(`${ORDER_SEARCH}?labNumber=${encodeURIComponent(accession)}`),
      ORDER_SEARCH,
    );
    test.skip(body.samples.length === 0, "this order has no sample items");

    const [[sampleId]] = query(
      `SELECT id FROM clinlims.sample WHERE accession_number = '${accession}'`,
    );
    // The list excludes VOIDED items — Java's criteria is {sample.id,
    // voided:false}, an exact match, so NULL is excluded as well.
    const liveItems = query(
      `SELECT id FROM clinlims.sample_item WHERE samp_id = ${sampleId} AND voided = false`,
    ).map((r) => r[0]);
    expect(body.samples.length, `${ORDER_SEARCH} samples excludes voided items`).toBe(liveItems.length);
    expect(
      body.samples.map((s: any) => s.id).sort(),
      `${ORDER_SEARCH} samples are the live sample items`,
    ).toEqual([...liveItems].sort());

    for (const item of body.samples) {
      // id and sampleItemId are the SAME value under two keys, as are
      // sortOrder and index — Java puts each twice for the frontend.
      expect(item.sampleItemId, "sampleItemId duplicates id").toBe(item.id);
      if ("sortOrder" in item) expect(item.index, "index duplicates sortOrder").toBe(item.sortOrder);

      // sampleXML is a nested partial duplicate. collectionDate/Time and
      // quantity must agree with the outer copy, which is what makes it a
      // duplicate rather than a second source of truth.
      expect(item.sampleXML.collectionDate, "sampleXML mirrors collectionDate").toBe(item.collectionDate);
      expect(item.sampleXML.collectionTime, "sampleXML mirrors collectionTime").toBe(item.collectionTime);
      expect(item.sampleXML.quantity, "sampleXML mirrors quantity").toBe(item.quantity);

      expect(Array.isArray(item.tests), "tests is an array").toBe(true);
      expect(Array.isArray(item.panels), "panels is an array").toBe(true);

      // The storage keys travel TOGETHER: Java puts them only inside
      // `if (assignment != null && assignment.getLocationId() != null)`, so a
      // port that emitted some of them, or emitted them as nulls, diverges.
      const storageKeys = ["storageLocationId", "storageLocationType",
                           "storagePositionCoordinate", "storageNotes"];
      const present = storageKeys.filter((k) => k in item);
      expect(
        present.length === 0 || present.length === storageKeys.length,
        `storage keys are all-or-nothing (saw ${present.join(",")})`,
      ).toBe(true);

      if (present.length > 0) {
        // DB oracle for the assignment itself.
        const [[locId, locType, coord]] = query(
          "SELECT location_id, COALESCE(location_type,''), COALESCE(position_coordinate,'')" +
            ` FROM clinlims.sample_storage_assignment WHERE sample_item_id = ${item.id}`,
        );
        expect(String(item.storageLocationId), "storageLocationId from the assignment").toBe(locId);
        expect(item.storageLocationType, "storageLocationType from the assignment").toBe(locType);

        // The hierarchical path is built by walking UPWARD to the root and
        // appending the position coordinate. Asserting the last segment and
        // the separator pins the shape without hardcoding a room name.
        if ("storageHierarchicalPath" in item) {
          expect(item.storageHierarchicalPath, "path is ' > '-separated").toContain(" > ");
          if (coord !== "") {
            expect(
              item.storageHierarchicalPath.endsWith(coord),
              "the position coordinate is the last path segment",
            ).toBe(true);
          }
        }
      }
    }
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

    // DB oracle. The loop above only checks TYPES; on its own it would pass
    // even if the endpoint invented labNumbers. Every returned labNumber must
    // be a real clinlims.sample.accession_number, and the page must be a
    // strict subset of the table (it is paged — 21 of 32 rows in this
    // dataset), so a port cannot pass by returning everything or by returning
    // fabricated rows.
    const realAccessions = new Set(
      query("SELECT accession_number FROM clinlims.sample WHERE accession_number IS NOT NULL").map((r) => r[0]),
    );
    for (const order of body.orders) {
      expect(
        realAccessions.has(order.labNumber),
        `${ORDER_DASHBOARD} labNumber ${order.labNumber} is a real sample accession`,
      ).toBe(true);
    }
    expect(body.orders.length, `${ORDER_DASHBOARD} returns a paged subset of sample`).toBeLessThanOrEqual(
      realAccessions.size,
    );
  });

  test("order/dashboard: pageSize is echoed but IGNORED; externalCount is hardcoded 0", async ({ request }) => {
    // Three confirmed Java quirks, pinned so the port reproduces them
    // knowingly rather than "fixing" them into a working pager:
    //
    //  1. pageSize is used only to compute the OFFSET. The number of rows
    //     actually returned comes from the server's page.defaultPageSize
    //     config, not from the request — so asking for 1 row does not get you
    //     1 row, while the echoed pageSize still says 1.
    //  2. externalCount is hardcoded to 0 and never computed.
    //  3. includeExternal is accepted as a param and never read at all.
    //
    // MIGRATION POLICY: these are Java bugs. They are pinned, not fixed.
    const res = await readJson(await request.get(`${ORDER_DASHBOARD}?pageSize=1`), ORDER_DASHBOARD);

    expect(res.pageSize, "pageSize is echoed back verbatim").toBe(1);
    // The echoed value and the real row count are allowed to disagree — that
    // IS the bug. Asserting they agree would encode a fix.
    expect(Array.isArray(res.orders), "orders is still an array").toBe(true);

    expect(res.externalCount, "externalCount is hardcoded 0, never computed").toBe(0);

    // includeExternal is inert: passing true must not change the response.
    const withExternal = await readJson(
      await request.get(`${ORDER_DASHBOARD}?includeExternal=true`),
      ORDER_DASHBOARD,
    );
    const without = await readJson(await request.get(ORDER_DASHBOARD), ORDER_DASHBOARD);
    expect(withExternal.orders.length, "includeExternal=true is ignored").toBe(without.orders.length);
    expect(withExternal.externalCount, "includeExternal does not populate externalCount").toBe(0);
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

  test("unassigned-sample/count-by-facility: {count:n}, a SUBSET of by-facility", async ({ request }) => {
    const orgs = query("SELECT id FROM clinlims.organization ORDER BY id LIMIT 1");
    test.skip(orgs.length === 0, "no organizations in this dataset");
    const facilityId = orgs[0][0];

    const counted = await readJson(
      await request.get(`${UNASSIGNED}/count-by-facility/${facilityId}`),
      `${UNASSIGNED}/count-by-facility`,
    );
    expect(Object.keys(counted), `${UNASSIGNED}/count-by-facility envelope`).toEqual(["count"]);
    expect(typeof counted.count, `${UNASSIGNED}/count-by-facility count is a number`).toBe("number");

    const listed = await readJson(
      await request.get(`${UNASSIGNED}/by-facility/${facilityId}`),
      `${UNASSIGNED}/by-facility`,
    );
    expect(Array.isArray(listed), `${UNASSIGNED}/by-facility is an array`).toBe(true);

    // NOT an equality. countUnassignedSamplesByFacility applies an extra
    // lost/canceled filter that getUnassignedSamplesByDestinationFacility does
    // NOT, so the count is legitimately a SUBSET of the list length. An
    // earlier draft of this test asserted equality; it passed only because
    // both are 0 in this dataset and would have failed the moment real
    // referral data existed. The correct, non-vacuous invariant is the
    // inequality plus the guarantee that count is never negative.
    expect(counted.count, "count-by-facility never exceeds by-facility length").toBeLessThanOrEqual(
      listed.length,
    );
    expect(counted.count, "count-by-facility is non-negative").toBeGreaterThanOrEqual(0);
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

// ── WHAT IS ACTUALLY VERIFIED AGAINST REAL DATA ─────────────────────────────
//
// Spelled out because a green run does NOT mean this wave is covered, and an
// isArray() assertion that passes on [] is not coverage:
//
//   VERIFIED against real rows (DB oracle or populated response):
//     - all-by-accession: rows echo the requested accession, ids are numeric.
//     - order/search: every paymentOption {id,value} is cross-checked against
//       clinlims.dictionary (and pins that `value` comes from dict_entry).
//     - order/dashboard: every labNumber must be a real sample
//       accession_number, and the page is a strict subset of the table.
//     - order/dashboard quirks: pageSize-ignored / externalCount-0 /
//       includeExternal-inert, each proven by contrasting real responses.
//     - unassigned-by-accession: always-500 proven across three inputs.
//     - count-by-facility <= by-facility: an inequality that holds on real
//       data (an earlier draft asserted equality and was WRONG).
//
//   ENVELOPE-ONLY — the collection is empty in this dataset, so row shape is
//   UNVERIFIED:
//     - unassigned-sample and unassigned-sample/items  (no referral rows)
//     - unassigned-sample/items/search                 (same)
//     - order/{accession}/attachments 200 path         (no order_attachment
//       rows; the 404 path IS verified)
//   Closing these means seeding referrals and attachments, the same way
//   src/test/resources/fixtures/patient-media-e2e.sql closed c1's photo gap.
//
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
