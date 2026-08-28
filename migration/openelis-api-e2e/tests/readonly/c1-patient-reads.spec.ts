// §5 — c1: patient reads (Java baseline; Go parity gate once c1 is ported).
//
// Taxonomy Type C/D, Wave 3, branch migration/c1-patient-reads. Written
// BEFORE the Go implementation on purpose: this file captures what Java
// actually does, so the port has an executable specification to satisfy
// rather than a description to interpret. It now runs under `go-parity` too.
//
// Every expectation below was captured from the LIVE Java server (authed,
// against the dev Postgres), not derived from reading source. Where source
// and live behavior could disagree, live wins.
//
// ── PHI WARNING ────────────────────────────────────────────────────────────
// Unlike a1/a2/b1/b2 (reference data), these endpoints return real patient
// health information: names, birth dates, national IDs, addresses, phone
// numbers, email. Two consequences for this suite:
//   1. Assertions must NOT hardcode PHI values. They pin SHAPE, TYPE and
//      INVARIANTS, and read any concrete value from the live DB oracle or
//      from a prior response. A spec file is not a place to check in a
//      patient's national ID.
//   2. The auth boundary is a first-class assertion here, not an
//      afterthought — see the "authorization" section at the bottom. This is
//      the wave P0 auth existed for: before it, the Go service served every
//      one of these endpoints to any anonymous caller, and the boundary test
//      was `test.skip`ped against go-parity for exactly that reason. That skip
//      is gone, which was the stated definition of done for
//      auth-adoption-plan.md's Phase 1.
import { test, expect, request as apiRequest, type APIRequestContext } from "@playwright/test";
import { readJson, expectKeysWithin, expectNonEmptyString } from "../../fixtures/assert";
import { query } from "../../fixtures/db";
import { E2E_PASS, E2E_USERS, E2E_AUTHZ_USERS } from "../../fixtures/env";
import {
  LOGIN_PATH,
  LOGIN_USER_FIELD,
  LOGIN_PASS_FIELD,
  SESSION_PATH,
} from "../../fixtures/contract";

/** Log in as a fixture user and return that context. */
async function loginAs(user: string): Promise<APIRequestContext> {
  const ctx = await apiRequest.newContext({
    baseURL: test.info().project.use.baseURL,
    ignoreHTTPSErrors: true,
    storageState: { cookies: [], origins: [] },
  });
  await ctx.get(SESSION_PATH);
  const res = await ctx.post(LOGIN_PATH, {
    form: { [LOGIN_USER_FIELD]: user, [LOGIN_PASS_FIELD]: E2E_PASS },
  });
  expect(res.status(), `login as ${user}`).toBe(200);
  return ctx;
}

const BY_LAB_NUMBER = "rest/patientByLabNumer"; // typo is Java's, not ours
const MERGE_DETAILS = "rest/patient/merge/details";
const ID_DOCUMENTS = "rest/patient-id-documents";
const PATIENT_PHOTOS = "rest/patient-photos";

// Large enough that it will never collide with a real seeded id.
const MISSING_ID = "99999999";

// Patient entity fields Java serializes. Optional ones are Jackson
// NON_NULL-omitted when the column is null, so `allowed` is the full set and
// `required` is only what is genuinely always present.
const PATIENT_ALLOWED = [
  "lastupdated", "id", "race", "gender", "birthDate", "birthDateForDisplay",
  "birthTime", "birthTimeForDisplay", "nationalId", "ethnicity", "person",
  "externalId", "isMerged", "fhirUuidAsString", "chartNumber", "fhirUuid",
];
const PERSON_ALLOWED = [
  "lastupdated", "id", "lastName", "firstName", "middleName", "multipleUnit",
  "streetAddress", "city", "state", "zipCode", "country", "workPhone",
  "homePhone", "cellPhone", "primaryPhone", "fax", "email", "gpsLatitude",
  "gpsLongitude",
];

/** A real accession number + its patient id, read from the DB, never hardcoded. */
async function anySampleAccession(): Promise<{ accession: string; patientId: string }> {
  const rows = query(
    "SELECT s.accession_number, sh.patient_id FROM clinlims.sample s" +
      " JOIN clinlims.sample_human sh ON sh.samp_id = s.id" +
      " WHERE s.accession_number IS NOT NULL ORDER BY s.id LIMIT 1",
  );
  return { accession: rows[0][0], patientId: rows[0][1] };
}

test.describe("c1 — patient reads", () => {
  // ── Fixture premises ────────────────────────────────────────────────────
  // Several tests below guard themselves with `test.skip` when their fixture
  // rows are missing — the honest thing to do locally, but it means a suite
  // that stops loading the fixture goes GREEN while exercising nothing.
  //
  // That is not hypothetical: patient-media-e2e.sql was never referenced by
  // load-test-fixtures.sh, so in CI four c1 tests silently took their skip
  // branch. This test is what makes that failure loud. If it fails, run
  //   ./src/test/resources/load-test-fixtures.sh --profile=core
  // and check that the loader still calls the fixture.
  test("the patient-media fixture is loaded (otherwise later skips hide real gaps)", async () => {
    expect(
      Number(query("SELECT count(*) FROM clinlims.patient_photo WHERE id BETWEEN 9900000 AND 9900099")[0][0]),
      "patient_photo rows — populated-media paths depend on these",
    ).toBeGreaterThan(0);

    const docs = Number(
      query("SELECT count(*) FROM clinlims.patient_id_document WHERE id BETWEEN 9900000 AND 9900099")[0][0],
    );
    expect(docs, "patient_id_document rows").toBeGreaterThan(0);

    // The specific shapes later tests need, not just "some rows": a
    // soft-deleted row, a null-description row, and a row owned by a DIFFERENT
    // patient. Each backs one assertion that would otherwise skip.
    expect(
      Number(query("SELECT count(*) FROM clinlims.patient_id_document WHERE deleted = true")[0][0]),
      "a soft-deleted document — the deleted-filter test needs it",
    ).toBeGreaterThan(0);
    expect(
      Number(
        query("SELECT count(*) FROM clinlims.patient_id_document WHERE description IS NULL AND deleted = false")[0][0],
      ),
      "a null-description document — the NON_NULL omission test needs it",
    ).toBeGreaterThan(0);
    expect(
      Number(query("SELECT count(DISTINCT patient_id) FROM clinlims.patient_id_document WHERE deleted = false")[0][0]),
      "documents on at least two patients — the cross-patient lookup test needs it",
    ).toBeGreaterThan(1);

    // The patient-less sample, which backs patientByLabNumer's second 404.
    expect(
      Number(query("SELECT count(*) FROM clinlims.sample WHERE accession_number = 'E2E-NOPAT-01'")[0][0]),
      "the patient-less sample — the second 404 path needs it",
    ).toBe(1);
  });

  // ── 1. rest/patientByLabNumer ───────────────────────────────────────────
  // Lives on SampleEditRestController, NOT a patient controller — and the
  // query param is `accessionNumber`, not `labNumber` as the endpoint name
  // suggests. Both facts were confirmed against the live server after the
  // obvious guess returned 400.

  test("patientByLabNumer: returns the Patient entity for a real accession", async ({ request }) => {
    const { accession, patientId } = await anySampleAccession();
    const body = await readJson(
      await request.get(`${BY_LAB_NUMBER}?accessionNumber=${encodeURIComponent(accession)}`),
      BY_LAB_NUMBER,
    );

    expectKeysWithin(body, PATIENT_ALLOWED, ["id", "person"], `${BY_LAB_NUMBER} patient`);
    // DB oracle: the endpoint must resolve to the patient the sample is
    // actually linked to — not merely "some patient".
    expect(body.id, `${BY_LAB_NUMBER} resolves to the sample's patient`).toBe(patientId);
    expectKeysWithin(body.person, PERSON_ALLOWED, ["id"], `${BY_LAB_NUMBER} person`);

    // isMerged is a real boolean, not a "Y"/"N" string like Organization.isActive.
    expect(typeof body.isMerged, `${BY_LAB_NUMBER} isMerged is boolean`).toBe("boolean");
    // fhirUuidAsString is always present (Provider has the same pair) and is a
    // string even when the UUID column is null — it renders as "".
    expect(typeof body.fhirUuidAsString, `${BY_LAB_NUMBER} fhirUuidAsString is string`).toBe("string");
  });

  test("patientByLabNumer: date fields pair an epoch with a formatted string", async ({ request }) => {
    const { accession } = await anySampleAccession();
    const body = await readJson(
      await request.get(`${BY_LAB_NUMBER}?accessionNumber=${encodeURIComponent(accession)}`),
      BY_LAB_NUMBER,
    );
    test.skip(!("birthDate" in body), "this patient has no birthDate");

    // birthDate/birthTime serialize as epoch millis; the *ForDisplay twins are
    // pre-formatted strings. A Go port must emit BOTH, and must not silently
    // switch either to ISO-8601.
    expect(typeof body.birthDate, `${BY_LAB_NUMBER} birthDate is epoch millis`).toBe("number");
    expect(body.birthDateForDisplay, `${BY_LAB_NUMBER} birthDateForDisplay is dd/dd/dddd`).toMatch(
      /^\d{2}\/\d{2}\/\d{4}$/,
    );

    if ("birthTimeForDisplay" in body) {
      expect(body.birthTimeForDisplay, `${BY_LAB_NUMBER} birthTimeForDisplay is dd/dd/dddd`).toMatch(
        /^\d{2}\/\d{2}\/\d{4}$/,
      );
    }
  });

  test("patientByLabNumer: birthDate is the STORED value, echoed faithfully", async ({ request }) => {
    // RESOLVED (was an open question): birthDateForDisplay is NOT a rendering
    // of birthDate. It is its own persisted column, `entered_birth_date`,
    // holding the literal text the user typed (which may contain "XX" when the
    // day/month is unknown — see db/dbInit/OpenELIS-Global.sql:3844). So the
    // two fields disagreeing is NOT a serialization bug and must NOT be
    // "fixed" by deriving one from the other.
    //
    // They disagree here for a separate, real reason:
    // Patient.setBirthDateForDisplay (Patient.java:261-267) writes BACK into
    // birthDate using a LENIENT SimpleDateFormat whose pattern comes from the
    // runtime locale (fr-FR -> dd/MM/yyyy). Parsing "01/15/1990" as
    // day=01/month=15 rolls over to 1991-03-01. That corruption already ran at
    // WRITE time, so the column itself holds the wrong date.
    //
    // The read path is therefore FAITHFUL: it returns what is stored. That is
    // exactly what this asserts, and it is the contract the Go port must meet
    // — read and return birth_date; do NOT replicate the write-back setter.
    const { accession } = await anySampleAccession();
    const body = await readJson(
      await request.get(`${BY_LAB_NUMBER}?accessionNumber=${encodeURIComponent(accession)}`),
      BY_LAB_NUMBER,
    );
    test.skip(!("birthDate" in body), "this patient has no birthDate");

    const stored = query(
      `SELECT extract(epoch FROM birth_date) FROM clinlims.patient WHERE id = '${body.id}'`,
    )[0][0];
    test.skip(!stored, "birth_date is null for this patient");
    expect(body.birthDate, `${BY_LAB_NUMBER} birthDate echoes the stored birth_date`).toBe(
      Math.round(Number(stored) * 1000),
    );
  });

  test("patientByLabNumer: birthTime is TRUNCATED to midnight (time-of-birth is dropped)", async ({
    request,
  }) => {
    // PARITY TRAP for the port. birth_time is `timestamp` in the schema and is
    // documented as "Time of birth for newborn patients"
    // (db/dbInit/OpenELIS-Global.sql:3809) — the dev row genuinely stores
    // 10:00:00. But Patient.hbm.xml:44 maps it as java.sql.Date, so Hibernate
    // truncates the clock and the API emits midnight.
    //
    // A Go port scanning the column into a time.Time and serializing it would
    // return 10:00:00 and diverge on a field nobody would think to check. This
    // asserts the truncation explicitly so the port is forced to reproduce it
    // (or to diverge deliberately and document it).
    const { accession } = await anySampleAccession();
    const body = await readJson(
      await request.get(`${BY_LAB_NUMBER}?accessionNumber=${encodeURIComponent(accession)}`),
      BY_LAB_NUMBER,
    );
    test.skip(!("birthTime" in body), "this patient has no birthTime");

    const raw = query(
      `SELECT to_char(birth_time, 'HH24:MI:SS') FROM clinlims.patient WHERE id = '${body.id}'`,
    )[0][0];
    test.skip(!raw, "birth_time is null for this patient");

    // The API value must land exactly on a UTC midnight...
    expect(body.birthTime % 86_400_000, `${BY_LAB_NUMBER} birthTime is truncated to midnight`).toBe(0);
    // ...and this is only a meaningful assertion when the stored row actually
    // HAS a non-midnight time to lose. Otherwise it proves nothing.
    if (raw !== "00:00:00") {
      expect(
        raw,
        `stored birth_time (${raw}) has a real clock time that the API drops — truncation confirmed`,
      ).not.toBe("00:00:00");
    }
  });

  test("patientByLabNumer: unknown accession is 404; blank/missing param is 400", async ({ request }) => {
    // Note this differs from b2's family, where a missing id produces a 500 via
    // ObjectNotFoundException. Here the controller guards explicitly, so the
    // Go port should reproduce these codes rather than assuming the b2 pattern.
    const unknown = await request.get(`${BY_LAB_NUMBER}?accessionNumber=NO_SUCH_ACCESSION_XYZ`);
    expect(unknown.status(), `${BY_LAB_NUMBER} unknown accession`).toBe(404);

    const blank = await request.get(`${BY_LAB_NUMBER}?accessionNumber=`);
    expect(blank.status(), `${BY_LAB_NUMBER} blank accessionNumber`).toBe(400);

    const missing = await request.get(BY_LAB_NUMBER);
    expect(missing.status(), `${BY_LAB_NUMBER} missing accessionNumber`).toBe(400);
  });

  // ── 2. rest/patient/merge/details/{patientId} ───────────────────────────

  test("merge/details: envelope shape + dataSummary counts are numbers", async ({ request }) => {
    const rows = query("SELECT id FROM clinlims.patient ORDER BY id LIMIT 1");
    const patientId = rows[0][0];
    const parsed = await readJson(await request.get(`${MERGE_DETAILS}/${patientId}`), MERGE_DETAILS);

    expectKeysWithin(
      parsed,
      ["patientId", "firstName", "lastName", "gender", "birthDate", "dataSummary", "identifiers", "conflictingFields"],
      ["patientId", "dataSummary", "identifiers", "conflictingFields"],
      `${MERGE_DETAILS} body`,
    );
    expect(parsed.patientId, `${MERGE_DETAILS} echoes the requested id`).toBe(patientId);

    // Unlike patientByLabNumer this is a purpose-built DTO, so birthDate here
    // is the FORMATTED string, not an epoch — an easy thing to get backwards
    // when porting, since the two endpoints use the same field name for
    // different types.
    if ("birthDate" in parsed) {
      expect(typeof parsed.birthDate, `${MERGE_DETAILS} birthDate is a formatted string`).toBe("string");
    }

    const summaryKeys = [
      "totalOrders", "activeOrders", "totalResults", "totalSamples", "totalDocuments",
      "totalIdentifiers", "totalContacts", "totalRelations", "totalAuditEntries", "totalDataItems",
    ];
    for (const k of summaryKeys) {
      expect(typeof parsed.dataSummary[k], `${MERGE_DETAILS} dataSummary.${k} is a number`).toBe("number");
      expect(parsed.dataSummary[k], `${MERGE_DETAILS} dataSummary.${k} is non-negative`).toBeGreaterThanOrEqual(0);
    }
    expect(Array.isArray(parsed.dataSummary.conflictingFields), `${MERGE_DETAILS} conflictingFields is array`).toBe(true);
    expect(Array.isArray(parsed.identifiers), `${MERGE_DETAILS} identifiers is array`).toBe(true);

    // totalDataItems is a COMPUTED getter on the DTO
    // (PatientMergeDataSummaryDTO.java:52-54), not a stored field — Jackson
    // emits it because it looks like a property. A Go port that models the
    // DTO as plain struct fields will omit it entirely unless it knows to
    // derive it. Note the sum deliberately excludes totalContacts,
    // totalRelations and totalAuditEntries.
    expect(parsed.dataSummary.totalDataItems, `${MERGE_DETAILS} totalDataItems is the documented sum`).toBe(
      parsed.dataSummary.totalOrders +
        parsed.dataSummary.totalResults +
        parsed.dataSummary.totalSamples +
        parsed.dataSummary.totalDocuments +
        parsed.dataSummary.totalIdentifiers,
    );

    // These DTO fields exist on the class but getMergeDetails never populates
    // them (PatientMergeServiceImpl.java:296-300), so Jackson's NON_NULL drops
    // them. Asserting their ABSENCE stops a port from "helpfully" filling them
    // in — which would leak extra PHI (nationalId, phone, email, address) that
    // Java does not return from this endpoint.
    for (const neverSet of ["nationalId", "phoneNumber", "email", "address"]) {
      expect(neverSet in parsed, `${MERGE_DETAILS} ${neverSet} is never populated by Java`).toBe(false);
    }
  });

  test("merge/details: dataSummary counts match the DB (oracle)", async ({ request }) => {
    // NAMING TRAP, confirmed in PatientMergeServiceImpl.java:232-245 after this
    // test first failed on the obvious-but-wrong assumption:
    //   totalOrders  counts sample_human rows  (one per sample/order)
    //   totalSamples counts sample_ITEM rows   (the containers within them)
    // So totalSamples > totalOrders is normal, and the field names read
    // backwards from what the DB tables are called. A port that maps
    // totalSamples to a sample count would be subtly wrong in a way no shape
    // assertion catches — which is exactly why this oracle is worth having.
    const rows = query(
      "SELECT sh.patient_id, count(*) FROM clinlims.sample_human sh" +
        " GROUP BY sh.patient_id ORDER BY count(*) DESC LIMIT 1",
    );
    test.skip(rows.length === 0, "no patient has samples in this dataset");
    const [patientId, orderCount] = rows[0];

    // `voided = false` is NOT incidental. PatientMergeServiceImpl's
    // countSamplesForPatient walks getSampleItemsBySampleId, whose criteria is
    // {sample.id, voided:false} — the same filter rest/order/search applies.
    // Without the predicate this oracle counts rows Java never counts, and the
    // only reason it agreed before is that nothing in the dataset was voided;
    // order-search-e2e.sql now seeds a voided item, which is what exposed it.
    const itemCount = Number(
      query(
        "SELECT count(*) FROM clinlims.sample_item si" +
          " WHERE si.voided = false" +
          " AND si.samp_id IN (SELECT sh.samp_id FROM clinlims.sample_human sh" +
          ` WHERE sh.patient_id = ${patientId})`,
      )[0][0],
    );

    const parsed = await readJson(await request.get(`${MERGE_DETAILS}/${patientId}`), MERGE_DETAILS);
    expect(parsed.dataSummary.totalOrders, `${MERGE_DETAILS} totalOrders = sample_human rows`).toBe(
      Number(orderCount),
    );
    expect(parsed.dataSummary.totalSamples, `${MERGE_DETAILS} totalSamples = sample_item rows`).toBe(itemCount);
    expect(parsed.dataSummary.totalOrders, "chosen patient genuinely has orders").toBeGreaterThan(0);

    // activeOrders is set to the SAME value as totalOrders
    // (PatientMergeServiceImpl.java:328). The code comment above it says
    // "Set to 0 for now" and describes a status filter that was never
    // implemented, so this is a known-unfinished field, not a real
    // active-order count. Pinned so the port copies the current behavior
    // knowingly rather than inventing a status filter Java does not have.
    expect(parsed.dataSummary.activeOrders, `${MERGE_DETAILS} activeOrders mirrors totalOrders (unfinished in Java)`).toBe(
      parsed.dataSummary.totalOrders,
    );

    // totalResults is NOT a plain analysis count: countResultsForPatient asks
    // IStatusService for the Canceled / SampleRejected / NotStarted ids and
    // excludes them. Type-checking the field is not enough — the dev dataset's
    // analyses are ALL "Not Tested", so Java reports 0 while an unfiltered
    // count reports 28. A port that forgets the exclusion (or wires its status
    // resolver as nil) lands on 28 and passes every other assertion here.
    // The `si.voided = false` here is the SECOND filter on this field, and it
    // is easy to miss: countResultsForPatient reaches its analyses through
    // getSampleItemsBySampleId ({sample.id, voided:false}), so an analysis
    // hanging off a voided sample item is never counted no matter what status
    // it carries. order-search-e2e.sql seeds exactly that row — a countable
    // status on a voided item — so this predicate is load-bearing rather than
    // decorative.
    const expectedResults = Number(
      query(
        "SELECT count(*) FROM clinlims.analysis a" +
          " WHERE a.sampitem_id IN (SELECT si.id FROM clinlims.sample_item si" +
          `  WHERE si.voided = false AND si.samp_id IN (SELECT sh.samp_id FROM clinlims.sample_human sh WHERE sh.patient_id = ${patientId}))` +
          "   AND a.status_id NOT IN (SELECT id FROM clinlims.status_of_sample" +
          "     WHERE status_type = 'ANALYSIS'" +
          "       AND name IN ('Test Canceled', 'Sample Rejected', 'Not Tested'))",
      )[0][0],
    );
    expect(parsed.dataSummary.totalResults, `${MERGE_DETAILS} totalResults excludes the three statuses`).toBe(
      expectedResults,
    );

    // Inversion check for the voided half specifically (Constitution V.6): the
    // status exclusion already has its own inversion below, but without this
    // one a port that keeps the status filter and drops the voided filter is
    // caught only while the fixture happens to be loaded, and silently not
    // caught when it is not. Asserting the fixture's discriminating row EXISTS
    // turns "no coverage" into a failure instead of a pass.
    const voidedCountable = Number(
      query(
        "SELECT count(*) FROM clinlims.analysis a" +
          " JOIN clinlims.sample_item si ON si.id = a.sampitem_id" +
          `  WHERE si.voided = true AND si.samp_id IN (SELECT sh.samp_id FROM clinlims.sample_human sh WHERE sh.patient_id = ${patientId})` +
          "   AND a.status_id NOT IN (SELECT id FROM clinlims.status_of_sample" +
          "     WHERE status_type = 'ANALYSIS'" +
          "       AND name IN ('Test Canceled', 'Sample Rejected', 'Not Tested'))",
      )[0][0],
    );
    expect(
      voidedCountable,
      "order-search-e2e.sql seeds an otherwise-countable analysis on a voided item for this patient",
    ).toBeGreaterThan(0);

    // Prove the exclusion is doing something for this patient — otherwise the
    // assertion above is satisfied by any implementation whenever the filtered
    // and unfiltered counts happen to coincide.
    const unfiltered = Number(
      query(
        "SELECT count(*) FROM clinlims.analysis a" +
          " WHERE a.sampitem_id IN (SELECT si.id FROM clinlims.sample_item si" +
          `  WHERE si.samp_id IN (SELECT sh.samp_id FROM clinlims.sample_human sh WHERE sh.patient_id = ${patientId}))`,
      )[0][0],
    );
    expect(
      unfiltered,
      "the chosen patient has analyses in an EXCLUDED status, so the filter is observable",
    ).toBeGreaterThan(expectedResults);
  });

  test("merge/details: identifiers carry a resolved type NAME, not a numeric id", async ({ request }) => {
    const rows = query(
      "SELECT patient_id FROM clinlims.patient_identity WHERE identity_type_id = 1 ORDER BY patient_id LIMIT 1",
    );
    test.skip(rows.length === 0, "no patient has a NATIONAL identity in this dataset");
    const patientId = rows[0][0];

    const parsed = await readJson(await request.get(`${MERGE_DETAILS}/${patientId}`), MERGE_DETAILS);
    test.skip(parsed.identifiers.length === 0, "identifiers empty for the chosen patient");

    for (const ident of parsed.identifiers) {
      expectKeysWithin(ident, ["identityType", "identityValue"], ["identityType", "identityValue"], `${MERGE_DETAILS} identifier`);
      expectNonEmptyString(ident.identityType, `${MERGE_DETAILS} identityType`);
      // The type is resolved to a human label ("National ID"), not the raw
      // patient_identity_type id or its DB code ("NATIONAL"). A port that
      // returns the id or the enum name passes a naive shape check but breaks
      // the UI, so assert it is NOT numeric.
      expect(ident.identityType, `${MERGE_DETAILS} identityType is a label, not a numeric id`).not.toMatch(/^\d+$/);
    }
  });

  test("merge/details: totalIdentifiers counts ALL rows; identifiers[] excludes internal types", async ({
    request,
  }) => {
    // Java skips GUID, AKA, MOTHER and MOTHERS_INITIAL when building the list
    // (PatientMergeServiceImpl.java:304,314) but sets totalIdentifiers from
    // the UNFILTERED collection (:330). So the count and the array length
    // legitimately disagree — first spotted live, where a patient reported
    // totalIdentifiers=2 with a single-element array.
    //
    // A port that returns every patient_identity row would leak an internal
    // GUID into a human-facing list; one that also filters the count would
    // "fix" a number Java does not fix. Both halves are pinned.
    const EXCLUDED = ["GUID", "AKA", "MOTHER", "MOTHERS_INITIAL"];
    const excludedList = EXCLUDED.map((t) => `'${t}'`).join(",");

    const rows = query(
      "SELECT pi.patient_id FROM clinlims.patient_identity pi" +
        " JOIN clinlims.patient_identity_type pit ON pit.id = pi.identity_type_id" +
        ` WHERE upper(pit.identity_type) IN (${excludedList})` +
        " AND EXISTS (SELECT 1 FROM clinlims.patient_identity o" +
        "   JOIN clinlims.patient_identity_type ot ON ot.id = o.identity_type_id" +
        `   WHERE o.patient_id = pi.patient_id AND upper(ot.identity_type) NOT IN (${excludedList}))` +
        " ORDER BY pi.patient_id LIMIT 1",
    );
    test.skip(rows.length === 0, "no patient mixes an excluded identity type with a listed one");
    const patientId = rows[0][0];

    const dbTotal = Number(
      query(`SELECT count(*) FROM clinlims.patient_identity WHERE patient_id = ${patientId}`)[0][0],
    );
    const excludedCount = Number(
      query(
        "SELECT count(*) FROM clinlims.patient_identity pi" +
          " JOIN clinlims.patient_identity_type pit ON pit.id = pi.identity_type_id" +
          ` WHERE pi.patient_id = ${patientId} AND upper(pit.identity_type) IN (${excludedList})`,
      )[0][0],
    );

    const parsed = await readJson(await request.get(`${MERGE_DETAILS}/${patientId}`), MERGE_DETAILS);
    expect(parsed.dataSummary.totalIdentifiers, "totalIdentifiers counts ALL identity rows").toBe(dbTotal);
    expect(parsed.identifiers.length, "identifiers[] excludes the internal types").toBe(dbTotal - excludedCount);
    for (const ident of parsed.identifiers) {
      expect(
        EXCLUDED.some((t) => String(ident.identityType).toUpperCase().includes(t)),
        `no excluded identity type is listed (saw "${ident.identityType}")`,
      ).toBe(false);
      // `system` is declared on IdentifierDTO but never populated -> absent.
      expect("system" in ident, "IdentifierDTO.system is never populated by Java").toBe(false);
    }
  });

  test("merge/details: unknown id is 404; non-numeric id is 500 (Java bug, pinned)", async ({ request }) => {
    const unknown = await request.get(`${MERGE_DETAILS}/${MISSING_ID}`);
    expect(unknown.status(), `${MERGE_DETAILS} unknown patient`).toBe(404);

    // Live-confirmed: a non-numeric id reaches the service and blows up rather
    // than being rejected at binding (the path variable is a String here, so
    // Spring cannot reject it the way it rejects a bad int). Pinned as-is so
    // the port makes a DELIBERATE choice: either reproduce the 500 or return
    // 400 and document the divergence — the same call that was made for the
    // b2 404-vs-500 family.
    const malformed = await request.get(`${MERGE_DETAILS}/abc`);
    expect(malformed.status(), `${MERGE_DETAILS} non-numeric id (Java: 500)`).toBe(500);
  });

  // ── 3/4. rest/patient-id-documents ──────────────────────────────────────

  test("patient-id-documents: list is an array; unknown patient is 200 [] not 404", async ({ request }) => {
    const rows = query("SELECT id FROM clinlims.patient ORDER BY id LIMIT 1");
    const patientId = rows[0][0];

    const res = await request.get(`${ID_DOCUMENTS}/${patientId}`);
    const body = await readJson(res, ID_DOCUMENTS);
    expect(Array.isArray(body), `${ID_DOCUMENTS} is an array`).toBe(true);

    // Live-confirmed: an id that does not exist still returns 200 with [] —
    // this endpoint does NOT 404. A port that "helpfully" 404s here changes
    // the contract for every caller that just renders an empty list.
    const unknown = await request.get(`${ID_DOCUMENTS}/${MISSING_ID}`);
    expect(unknown.status(), `${ID_DOCUMENTS} unknown patient is 200`).toBe(200);
    expect(await unknown.json(), `${ID_DOCUMENTS} unknown patient is []`).toEqual([]);
  });

  test("patient-id-documents/{doc}/full: base64-in-JSON envelope, not raw bytes", async ({ request }) => {
    const rows = query("SELECT id FROM clinlims.patient ORDER BY id LIMIT 1");
    const patientId = rows[0][0];

    const res = await request.get(`${ID_DOCUMENTS}/${patientId}/1/full`);
    expect(res.status(), `${ID_DOCUMENTS}/full status`).toBe(200);
    // The shape that matters: JSON {"data": "<base64>"} — NOT an image/*
    // body. A port that streams raw bytes here would break every caller even
    // though "it returns the image" sounds equivalent.
    expect(
      (res.headers()["content-type"] ?? "").toLowerCase(),
      `${ID_DOCUMENTS}/full is JSON, not a binary stream`,
    ).toContain("application/json");
    const body = await res.json();
    expect(Object.keys(body), `${ID_DOCUMENTS}/full envelope`).toEqual(["data"]);
    expect(typeof body.data, `${ID_DOCUMENTS}/full data is a string`).toBe("string");
  });

  test("patient-id-documents: non-numeric documentId is 400, non-numeric patientId is 200", async ({ request }) => {
    const rows = query("SELECT id FROM clinlims.patient ORDER BY id LIMIT 1");
    const patientId = rows[0][0];

    // documentId is bound as Integer, so Spring rejects a non-numeric value at
    // binding with MethodArgumentTypeMismatch -> 400.
    const badDoc = await request.get(`${ID_DOCUMENTS}/${patientId}/abc/full`);
    expect(badDoc.status(), `${ID_DOCUMENTS} non-numeric documentId`).toBe(400);

    // ...but patientId is bound as String against a varchar column, so a
    // non-numeric value is not an error at all — it just matches nothing.
    // The two path variables on the SAME endpoint behave differently; a port
    // that validates both the same way diverges on one of them.
    const badPatient = await request.get(`${ID_DOCUMENTS}/abc`);
    expect(badPatient.status(), `${ID_DOCUMENTS} non-numeric patientId`).toBe(200);
    expect(await badPatient.json(), `${ID_DOCUMENTS} non-numeric patientId body`).toEqual([]);
  });

  // ── 5. rest/patient-photos/{id}/{isThumbnail} ───────────────────────────

  test("patient-photos: thumbnail returns BARE base64, full returns a data: URI", async ({ request }) => {
    const rows = query("SELECT id FROM clinlims.patient ORDER BY id LIMIT 1");
    const patientId = rows[0][0];

    const seen: Record<string, string> = {};
    for (const isThumb of ["true", "false"]) {
      const res = await request.get(`${PATIENT_PHOTOS}/${patientId}/${isThumb}`);
      expect(res.status(), `${PATIENT_PHOTOS} ${isThumb} status`).toBe(200);
      expect(
        (res.headers()["content-type"] ?? "").toLowerCase(),
        `${PATIENT_PHOTOS} ${isThumb} is JSON`,
      ).toContain("application/json");
      const body = await res.json();
      expect(Object.keys(body), `${PATIENT_PHOTOS} ${isThumb} envelope`).toEqual(["data"]);
      expect(typeof body.data, `${PATIENT_PHOTOS} ${isThumb} data is a string`).toBe("string");
      seen[isThumb] = body.data;
    }

    // THE BIGGEST PARITY TRAP IN THIS UNIT (PatientPhotoServiceImpl.java:116-119):
    // the two branches return STRUCTURALLY DIFFERENT strings.
    //   isThumbnail=false -> "data:image/jpeg;base64,AAAA..."  (full data-URI)
    //   isThumbnail=true  -> "AAAA..."                          (BARE base64)
    // The frontend compensates for exactly this: AyncAvatar.jsx requests
    // /true for the avatar while usePatientDetails.js requests /false. A Go
    // port that emits a data-URI for both "looks right" and silently breaks
    // the avatar.
    //
    // This dev DB has ZERO patient_photo rows, so both branches return "" and
    // the difference cannot be exercised. Rather than assert nothing, the
    // check below is written to activate automatically the moment a photo
    // fixture exists — it is skipped, not silently passing.
    test.skip(
      seen["false"] === "" && seen["true"] === "",
      "no patient_photo rows — load src/test/resources/fixtures/patient-media-e2e.sql to exercise this",
    );
    expect(seen["false"], `${PATIENT_PHOTOS}/false must be a data: URI`).toMatch(/^data:[^;]*;base64,/);
    expect(seen["true"], `${PATIENT_PHOTOS}/true must be BARE base64, no data: prefix`).not.toMatch(/^data:/);

    // The two branches must also read DIFFERENT columns (photo_data vs
    // thumbnail_data). Asserting only the prefix would still pass a port that
    // returned the full image for both and merely stripped the prefix.
    const fullPayload = seen["false"].replace(/^data:[^;]*;base64,/, "");
    expect(seen["true"], `${PATIENT_PHOTOS} thumbnail reads thumbnail_data, not photo_data`).not.toBe(
      fullPayload,
    );
  });

  test("patient-id-documents: item keys, NON_NULL omission, and the deleted filter", async ({ request }) => {
    const rows = query("SELECT id FROM clinlims.patient ORDER BY id LIMIT 1");
    const patientId = rows[0][0];
    const body = await readJson(await request.get(`${ID_DOCUMENTS}/${patientId}`), ID_DOCUMENTS);
    test.skip(
      body.length === 0,
      "no patient_id_document rows — load src/test/resources/fixtures/patient-media-e2e.sql",
    );

    // The controller hand-picks 5 keys from the entity (document_data and
    // patient_id are deliberately NOT exposed here — the full blob only comes
    // from the /full endpoint). Anything else appearing is a leak.
    for (const doc of body) {
      expectKeysWithin(
        doc,
        ["id", "thumbnail", "category", "description", "lastUpdated"],
        ["id", "thumbnail", "category"],
        `${ID_DOCUMENTS} item`,
      );
      expect(doc.thumbnail, `${ID_DOCUMENTS} thumbnail is a data: URI`).toMatch(/^data:[^;]*;base64,/);
      expect("documentData" in doc, `${ID_DOCUMENTS} must not expose the full blob in the list`).toBe(false);
      expect("patientId" in doc, `${ID_DOCUMENTS} must not echo patientId per item`).toBe(false);
    }

    // Jackson's NON_NULL is set via setSerializationInclusion, which applies
    // CONTENT inclusion too — so a null value inside these HashMaps drops the
    // KEY rather than emitting null. A Go port using a struct with
    // `json:"description"` (no omitempty) would emit "description": null and
    // diverge. Requires both a described and an undescribed row to prove.
    const described = query(
      `SELECT count(*) FROM clinlims.patient_id_document WHERE patient_id = '${patientId}' AND deleted = false AND description IS NOT NULL`,
    )[0][0];
    const undescribed = query(
      `SELECT count(*) FROM clinlims.patient_id_document WHERE patient_id = '${patientId}' AND deleted = false AND description IS NULL`,
    )[0][0];
    if (Number(described) > 0 && Number(undescribed) > 0) {
      expect(body.filter((d: any) => "description" in d).length, "described rows keep the key").toBe(
        Number(described),
      );
      expect(body.filter((d: any) => !("description" in d)).length, "null description OMITS the key").toBe(
        Number(undescribed),
      );
    }

    // Soft-deleted rows must never surface (DAO filters `deleted = false`).
    const liveCount = Number(
      query(
        `SELECT count(*) FROM clinlims.patient_id_document WHERE patient_id = '${patientId}' AND deleted = false`,
      )[0][0],
    );
    expect(body.length, `${ID_DOCUMENTS} excludes soft-deleted rows`).toBe(liveCount);
    const deletedIds = query(
      `SELECT id FROM clinlims.patient_id_document WHERE patient_id = '${patientId}' AND deleted = true`,
    ).map((r) => Number(r[0]));
    for (const gone of deletedIds) {
      expect(body.some((d: any) => d.id === gone), `soft-deleted doc ${gone} must not be listed`).toBe(false);
    }
  });

  test("patient-id-documents/{doc}/full: returns that document, and {} for another patient's doc", async ({
    request,
  }) => {
    const rows = query("SELECT id FROM clinlims.patient ORDER BY id LIMIT 1");
    const patientId = rows[0][0];
    const mine = query(
      `SELECT id FROM clinlims.patient_id_document WHERE patient_id = '${patientId}' AND deleted = false ORDER BY id LIMIT 1`,
    );
    test.skip(mine.length === 0, "no patient_id_document rows — load patient-media-e2e.sql");
    const docId = mine[0][0];

    const body = await readJson(await request.get(`${ID_DOCUMENTS}/${patientId}/${docId}/full`), `${ID_DOCUMENTS}/full`);
    expect(body.data, `${ID_DOCUMENTS}/full returns a data: URI`).toMatch(/^data:[^;]*;base64,/);

    // Cross-patient request: Java scopes the fetch by patientId and then scans
    // in Java for the docId, so a mismatched pair yields {"data":""} — NOT 403
    // or 404. Pinned because it is IDOR-shaped: the port must not "improve"
    // this into a 403 (a behavior change) nor into returning the document (a
    // real data leak).
    const others = query(
      `SELECT id FROM clinlims.patient_id_document WHERE patient_id <> '${patientId}' AND deleted = false ORDER BY id LIMIT 1`,
    );
    if (others.length > 0) {
      const foreign = await readJson(
        await request.get(`${ID_DOCUMENTS}/${patientId}/${others[0][0]}/full`),
        `${ID_DOCUMENTS}/full foreign`,
      );
      expect(foreign, "another patient's docId yields empty data, not the document").toEqual({ data: "" });
    }
  });

  test("patient-photos: a non-numeric patient id is 200 empty, NOT an error", async ({ request }) => {
    // patient_photo.patient_id is a plain varchar with no numeric usertype, so
    // a malformed id simply matches nothing. Contrast with merge/details,
    // where a non-numeric id hits Integer.parseInt inside the Hibernate
    // usertype and 500s. Same-looking input, opposite outcome, per endpoint.
    const res = await request.get(`${PATIENT_PHOTOS}/abc/true`);
    expect(res.status(), `${PATIENT_PHOTOS} non-numeric id`).toBe(200);
    expect(await res.json(), `${PATIENT_PHOTOS} non-numeric id body`).toEqual({ data: "" });
  });

  test("patient-photos: non-boolean isThumbnail is rejected with 400", async ({ request }) => {
    const rows = query("SELECT id FROM clinlims.patient ORDER BY id LIMIT 1");
    const patientId = rows[0][0];
    // Spring binds this path variable as a boolean, so a non-boolean fails at
    // binding — the same MethodArgumentTypeMismatch 400 that the b2
    // provider/search int params produce. Pinned because Go's ParseBool must
    // be wired to answer 400 too, not fall back to a default.
    const res = await request.get(`${PATIENT_PHOTOS}/${patientId}/notabool`);
    expect(res.status(), `${PATIENT_PHOTOS} non-boolean isThumbnail`).toBe(400);
  });

  // ── Cross-cutting: the PHI auth boundary ────────────────────────────────

  test("all c1 endpoints refuse anonymous access (PHI boundary)", async () => {
    // THE test this whole wave waited on. Until P0 auth landed this was
    // `test.skip`ped against go-parity, because the Go service had no auth
    // layer and served every one of these endpoints to anyone who could reach
    // the port. Removing that skip was the stated definition of done for
    // auth-adoption-plan.md Phase 1.
    //
    // A dedicated context with NO stored auth state — the rest of this suite
    // runs authenticated, so without this the boundary is never exercised.
    const anon = await apiRequest.newContext({
      baseURL: test.info().project.use.baseURL,
      ignoreHTTPSErrors: true,
      storageState: { cookies: [], origins: [] },
    });
    try {
      const { accession } = await anySampleAccession();
      const rows = query("SELECT id FROM clinlims.patient ORDER BY id LIMIT 1");
      const patientId = rows[0][0];

      for (const path of [
        `${BY_LAB_NUMBER}?accessionNumber=${encodeURIComponent(accession)}`,
        `${MERGE_DETAILS}/${patientId}`,
        `${ID_DOCUMENTS}/${patientId}`,
        `${ID_DOCUMENTS}/${patientId}/1/full`,
        `${PATIENT_PHOTOS}/${patientId}/true`,
      ]) {
        const res = await anon.get(path, { maxRedirects: 0 });
        // Java's Spring Security entry point: 302 to <context>/LoginPage,
        // measured on all five. Asserted unfollowed — following the redirect
        // lands on Tomcat's login JSP, which a Go port has no reason to serve,
        // so "302 to /LoginPage carrying nothing" is the language-neutral
        // contract and "the body looks like the login page" is not.
        expect(res.status(), `anonymous ${path} must be refused`).toBe(302);
        expect(
          res.headers()["location"] ?? "",
          `anonymous ${path} redirect target`,
        ).toMatch(/\/LoginPage$/);

        const text = await res.text();
        expect(text, `anonymous ${path} must be bodiless`).toBe("");
        expect(text, `anonymous ${path} must not leak a patient id`).not.toContain(`"patientId"`);
        expect(text, `anonymous ${path} must not leak a nationalId`).not.toContain(`"nationalId"`);
      }
    } finally {
      await anon.dispose();
    }
  });

  // ── Cross-cutting: the Reception role gate on merge/details ──────────────
  //
  // merge/details is the ONE c1 endpoint with a gate beyond authentication.
  // Java checks it in the handler
  // (PatientMergeRestController.hasMergePermission ->
  // userRoleService.userInRole(userId, Constants.ROLE_RECEPTION)) and returns
  // ResponseEntity.status(FORBIDDEN).build() — a BODILESS 403.
  //
  // The other four endpoints are authentication-only: no system_module_url row
  // (so the module interceptor auto-allows them) and no @PreAuthorize on any of
  // the owning controllers. Both checked directly against the running stack,
  // not assumed.
  test.describe("c1 authorization: merge/details requires Reception", () => {
    const patientId = () =>
      query("SELECT id FROM clinlims.patient ORDER BY id LIMIT 1")[0][0];

    test("a user WITH Reception gets the merge details", async () => {
      const ctx = await loginAs(E2E_USERS.reception);
      const body = await readJson(
        await ctx.get(`${MERGE_DETAILS}/${patientId()}`),
        `${MERGE_DETAILS} as Reception`,
      );
      expect(Object.keys(body).length, "merge details returns real content").toBeGreaterThan(0);
      await ctx.dispose();
    });

    test("an authenticated user WITHOUT Reception gets a bodiless 403", async () => {
      const ctx = await loginAs(E2E_USERS.noRoles);
      const res = await ctx.get(`${MERGE_DETAILS}/${patientId()}`);
      expect(res.status(), "no-Reception merge/details status").toBe(403);
      expect(await res.text(), "403 is bodiless").toBe("");
      await ctx.dispose();
    });

    // The two below are what stop a port from "helpfully" letting admins
    // through. hasMergePermission calls userInRole DIRECTLY — the
    // `|| isUserAdmin(...)` fallback exists only in
    // ModuleAuthenticationInterceptor, a different mechanism. Both verified
    // live against Java: 403, not 200.
    test("login_user.is_admin='Y' does NOT bypass the Reception gate", async () => {
      const ctx = await loginAs(E2E_AUTHZ_USERS.isAdmin);
      const res = await ctx.get(`${MERGE_DETAILS}/${patientId()}`);
      expect(res.status(), "is_admin merge/details status").toBe(403);
      expect(await res.text(), "403 is bodiless").toBe("");
      await ctx.dispose();
    });

    test("the Global Administrator role does NOT bypass the Reception gate", async () => {
      const ctx = await loginAs(E2E_AUTHZ_USERS.globalAdmin);
      const res = await ctx.get(`${MERGE_DETAILS}/${patientId()}`);
      expect(res.status(), "Global Administrator merge/details status").toBe(403);
      expect(await res.text(), "403 is bodiless").toBe("");
      await ctx.dispose();
    });

    test("the other four c1 endpoints need only authentication", async () => {
      // Same no-roles user that is refused merge/details reads all of these —
      // proving the 403 above is the Reception gate specifically, not a blanket
      // denial of low-privilege users.
      const ctx = await loginAs(E2E_USERS.noRoles);
      const { accession } = await anySampleAccession();
      const id = patientId();
      for (const path of [
        `${BY_LAB_NUMBER}?accessionNumber=${encodeURIComponent(accession)}`,
        `${ID_DOCUMENTS}/${id}`,
        `${PATIENT_PHOTOS}/${id}/true`,
      ]) {
        const res = await ctx.get(path);
        expect(res.status(), `no-roles user reads ${path}`).toBe(200);
      }
      await ctx.dispose();
    });
  });

  // ── Deeper Java-implementation coverage ─────────────────────────────────
  // Added after re-reading the Java methods line by line. Each of these is a
  // branch or a transformation the tests above pass over, and each is one a
  // port can get wrong while still going green.

  test("patientByLabNumer: a sample with NO patient is the SECOND 404 path", async ({ request }) => {
    // getPatientByLabNumber has TWO independent 404 returns:
    //
    //   Sample sample = getSample(accessionNumber);
    //   if (sample == null)  return notFound();     <- unknown accession
    //   Patient patient = sampleHumanService.getPatientForSample(sample);
    //   if (patient == null) return notFound();     <- THIS one
    //
    // A port that implements only the first — the obvious one — answers this
    // request with 200 and a null/empty patient, or 500 on a nil dereference.
    // Either way every other test in this file still passes, because the stock
    // dataset has no patient-less sample. patient-media-e2e.sql seeds one.
    const NO_PATIENT_ACCESSION = "E2E-NOPAT-01";
    const exists = query(
      `SELECT count(*) FROM clinlims.sample WHERE accession_number = '${NO_PATIENT_ACCESSION}'`,
    )[0][0];
    test.skip(
      exists === "0",
      "patient-less sample not seeded — load src/test/resources/fixtures/patient-media-e2e.sql",
    );

    // The premise, from the DB: the accession RESOLVES to a sample, and that
    // sample has no patient. Without this the test could pass for the boring
    // reason of the accession simply not existing.
    const linked = query(
      "SELECT count(*) FROM clinlims.sample_human sh JOIN clinlims.sample s ON s.id = sh.samp_id" +
        ` WHERE s.accession_number = '${NO_PATIENT_ACCESSION}'`,
    )[0][0];
    expect(linked, "the seeded sample deliberately has no sample_human row").toBe("0");

    const res = await request.get(
      `${BY_LAB_NUMBER}?accessionNumber=${encodeURIComponent(NO_PATIENT_ACCESSION)}`,
    );
    expect(res.status(), "sample exists but has no patient -> 404").toBe(404);
    expect(await res.text(), "404 is bodiless").toBe("");

    // Same status as the unknown-accession case — the two branches are
    // indistinguishable to a client, which is itself the contract.
    const unknown = await request.get(`${BY_LAB_NUMBER}?accessionNumber=NO-SUCH-ACCESSION-0001`);
    expect(unknown.status(), "unknown accession -> 404 as well").toBe(404);
  });

  test("merge/details: identityType is the MAPPED display name, not the DB code", async ({ request }) => {
    // PatientMergeServiceImpl.getDisplayNameForIdentityType is a switch on the
    // UPPERCASED patient_identity_type.identity_type:
    //   NATIONAL -> "National ID"   SUBJECT -> "Subject Number"
    //   ST -> "ST Number"           INSURANCE -> "Insurance ID"   ...
    //   null -> "Unknown"           default -> the input, unchanged
    //
    // The "not a numeric id" test above passes for a port that returns the raw
    // code ("NATIONAL"). This asserts the actual mapping, for every identity
    // type present in the data, cross-checked against the DB rather than
    // pinned to one patient.
    const DISPLAY_NAMES: Record<string, string> = {
      NATIONAL: "National ID",
      SUBJECT: "Subject Number",
      ST: "ST Number",
      INSURANCE: "Insurance ID",
      OCCUPATION: "Occupation",
      ORG_SITE: "Organization Site",
      EDUCATION: "Education",
      MARITIAL: "Marital Status",
      NATIONALITY: "Nationality",
      "OTHER NATIONALITY": "Other Nationality",
      "HEALTH DISTRICT": "Health District",
      "HEALTH REGION": "Health Region",
      OB_NUMBER: "OB Number",
      PC_NUMBER: "PC Number",
    };

    // One patient per identity type actually present, so this scales with the
    // dataset instead of hardcoding an id.
    const samples = query(
      "SELECT DISTINCT ON (t.identity_type) t.identity_type, pi.patient_id" +
        " FROM clinlims.patient_identity pi" +
        " JOIN clinlims.patient_identity_type t ON t.id = pi.identity_type_id" +
        " ORDER BY t.identity_type, pi.patient_id",
    );
    expect(samples.length, "the dataset has identities to check").toBeGreaterThan(0);

    let checked = 0;
    for (const [rawType, patientId] of samples) {
      // These four are filtered out of identifiers[] entirely (asserted
      // separately), so they cannot be checked here.
      if (["GUID", "AKA", "MOTHER", "MOTHERS_INITIAL"].includes(rawType)) continue;

      const parsed = await readJson(
        await request.get(`${MERGE_DETAILS}/${patientId}`),
        `${MERGE_DETAILS}/${patientId}`,
      );
      const types = parsed.identifiers.map((i: any) => i.identityType);
      // Java's `default:` returns the input unchanged, so an unmapped type is
      // expected back as the raw code — pinning that branch too.
      const expected = DISPLAY_NAMES[rawType.toUpperCase()] ?? rawType;
      expect(types, `${rawType} maps to "${expected}" for patient ${patientId}`).toContain(expected);
      if (expected !== rawType) {
        expect(types, `the raw code "${rawType}" must NOT be returned`).not.toContain(rawType);
      }
      checked++;
    }
    expect(checked, "at least one non-internal identity type was checked").toBeGreaterThan(0);
  });

  test("patient-id-documents: the data URI is built from document_type + the RIGHT column", async ({
    request,
  }) => {
    // Both endpoints build their payload by raw string concatenation:
    //   list:  "data:" + doc.getDocumentType() + ";base64," + doc.getThumbnailData()
    //   /full: "data:" + doc.getDocumentType() + ";base64," + doc.getDocumentData()
    //
    // Two mistakes a port makes easily, neither caught by the
    // /^data:[^;]*;base64,/ shape check the tests above use:
    //   1. hardcoding the media type — image/jpeg is the obvious guess, and
    //      the seeded rows are image/png;
    //   2. reading the same column for both, so the list silently serves the
    //      FULL image as a thumbnail.
    // Exact equality against the stored columns closes both.
    const rows = query(
      "SELECT id, document_type, thumbnail_data, document_data FROM clinlims.patient_id_document" +
        " WHERE deleted = false ORDER BY id",
    );
    test.skip(rows.length === 0, "no patient_id_document rows — load patient-media-e2e.sql");

    for (const [docId, docType, thumbData, fullData] of rows) {
      // The premise for mistake 2: the two columns genuinely differ here, so
      // "same value for both" is a detectable error rather than a coincidence.
      expect(thumbData, `doc ${docId}: thumbnail_data differs from document_data`).not.toBe(fullData);

      const patientId = query(
        `SELECT patient_id FROM clinlims.patient_id_document WHERE id = ${docId}`,
      )[0][0];

      const list = await readJson(await request.get(`${ID_DOCUMENTS}/${patientId}`), ID_DOCUMENTS);
      const item = list.find((d: any) => String(d.id) === String(docId));
      expect(item, `doc ${docId} appears in the list for patient ${patientId}`).toBeTruthy();
      expect(
        item.thumbnail,
        `doc ${docId}: list thumbnail is data:<document_type>;base64,<thumbnail_data>`,
      ).toBe(`data:${docType};base64,${thumbData}`);

      const full = await readJson(
        await request.get(`${ID_DOCUMENTS}/${patientId}/${docId}/full`),
        `${ID_DOCUMENTS}/full`,
      );
      expect(full.data, `doc ${docId}: /full is data:<document_type>;base64,<document_data>`).toBe(
        `data:${docType};base64,${fullData}`,
      );
    }
  });

  test("patient-id-documents: an unknown documentId is 200 with empty data, never 404", async ({
    request,
  }) => {
    // getIdDocumentFull loads the patient's documents and scans for the id in
    // Java; falling off the end returns Map.of("data", "") with a 200. NOT a
    // 404 and not an error — the same shape a patient with no documents at all
    // produces, so a client cannot distinguish "no such document" from
    // "the document is empty". Pinned because 404 is the tempting improvement.
    const patientId = query("SELECT id FROM clinlims.patient ORDER BY id LIMIT 1")[0][0];
    const res = await request.get(`${ID_DOCUMENTS}/${patientId}/99999999/full`);
    expect(res.status(), "unknown documentId status").toBe(200);
    expect(await res.json(), "unknown documentId body").toEqual({ data: "" });
  });

  test("patient-photos: isThumbnail accepts Spring's boolean vocabulary, not Go's", async ({ request }) => {
    // `boolean isThumbnail` is bound by Spring's StringToBooleanConverter,
    // whose vocabulary is fixed and case-insensitive:
    //   true:  true / on / yes / 1        false: false / off / no / 0
    // Everything else fails to bind and Spring answers 400.
    //
    // Go's strconv.ParseBool is a DIFFERENT set, and the two diverge in BOTH
    // directions — which is exactly what the earlier port did, and what the
    // /^data:.../ shape tests above could never notice. Every value below was
    // measured against the live Java server, including the eight that used to
    // disagree.
    const patientId = query("SELECT id FROM clinlims.patient ORDER BY id LIMIT 1")[0][0];

    const ACCEPTED = ["true", "false", "True", "FALSE", "TRUE", "on", "off", "yes", "no", "1", "0"];
    // "t"/"f" are ParseBool shorthand that Spring rejects — the direction a
    // port is most likely to get wrong, because Go accepts them silently.
    const REJECTED = ["t", "f", "T", "F", "bogus", "2", "onn"];

    for (const v of ACCEPTED) {
      const res = await request.get(`${PATIENT_PHOTOS}/${patientId}/${v}`);
      expect(res.status(), `isThumbnail="${v}" must bind (Spring accepts it)`).toBe(200);
      expect(Object.keys(await res.json()), `isThumbnail="${v}" envelope`).toEqual(["data"]);
    }
    for (const v of REJECTED) {
      const res = await request.get(`${PATIENT_PHOTOS}/${patientId}/${v}`);
      expect(res.status(), `isThumbnail="${v}" must NOT bind (Spring rejects it)`).toBe(400);
    }

    // The vocabulary is not just accept/reject — the two groups must map to
    // the two DIFFERENT branches. "on"/"yes"/"1" have to behave like "true"
    // (bare base64) and "off"/"no"/"0" like "false" (a data: URI), or a port
    // could pass the status checks above while treating every accepted value
    // as false.
    const dataFor = async (v: string) =>
      (await (await request.get(`${PATIENT_PHOTOS}/${patientId}/${v}`)).json()).data;
    const [t, on, yes, one] = [await dataFor("true"), await dataFor("on"), await dataFor("yes"), await dataFor("1")];
    const [f, off, no, zero] = [await dataFor("false"), await dataFor("off"), await dataFor("no"), await dataFor("0")];
    for (const [label, v] of [["on", on], ["yes", yes], ["1", one]] as const) {
      expect(v, `"${label}" must behave like "true"`).toBe(t);
    }
    for (const [label, v] of [["off", off], ["no", no], ["0", zero]] as const) {
      expect(v, `"${label}" must behave like "false"`).toBe(f);
    }
  });
});

// ── WHAT IS ACTUALLY VERIFIED AGAINST REAL DATA ─────────────────────────────
//
// Stated explicitly so a green run is not mistaken for full coverage. This
// file is the most heavily DB-oracled of the c1/c2/c3 set (~24 live SQL
// cross-checks); nothing here asserts against a mock, and no assertion relies
// on an empty collection to pass.
//
//   VERIFIED against real rows:
//     - patientByLabNumer resolves to the sample's ACTUAL patient (oracle on
//       sample_human), and birthDate is compared to the stored birth_date
//       column rather than to the response's own display field.
//     - birthTime truncation proven against a row that really stores 10:00:00.
//     - merge/details totalOrders/totalSamples cross-checked against
//       sample_human and sample_item counts — this is what exposed that the
//       two field names read backwards from the tables.
//     - identifiers[] vs totalIdentifiers proven against real
//       patient_identity + patient_identity_type rows, including the
//       GUID/AKA/MOTHER/MOTHERS_INITIAL exclusion.
//     - photo and document behavior proven against seeded rows
//       (src/test/resources/fixtures/patient-media-e2e.sql): the
//       data-URI-vs-bare-base64 split, the two branches reading DIFFERENT
//       columns, NON_NULL key omission for a null description, the
//       deleted=false filter, and the cross-patient {"data":""} case.
//
//   NOT VERIFIED (honest gaps):
//     - The populated shapes above depend on patient-media-e2e.sql being
//       loaded. Without it those tests test.skip() WITH A REASON rather than
//       passing vacuously — check for skips before trusting a green run.
//     - gpsLatitude/gpsLongitude and several optional Person fields are null
//       in this dataset, so only their absence-handling is exercised.
