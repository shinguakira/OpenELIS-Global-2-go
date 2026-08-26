// §5 — c1: patient reads (Java baseline; Go parity gate once c1 is ported).
//
// Taxonomy Type C/D, Wave 3, branch migration/c1-patient-reads. Written
// BEFORE the Go implementation on purpose: this file captures what Java
// actually does, so the port has an executable specification to satisfy
// rather than a description to interpret. It is deliberately NOT yet in
// playwright.config.ts's `go-parity` testMatch — there is no Go
// implementation to run it against. Add it there as c1 lands.
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
//      afterthought — see the "unauthenticated" test. Java answers 302
//      (redirect to login) for these; a Go port that serves them anonymously
//      would be a PHI leak, and the current Go service has no auth layer at
//      all (that is why it is bound to loopback — see
//      migration/openelis-go/docker-compose.go.yml).
import { test, expect } from "@playwright/test";
import { readJson, expectKeysWithin, expectNonEmptyString } from "../../fixtures/assert";
import { query } from "../../fixtures/db";

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

    // PARITY TRAP, live-observed and deliberately pinned loosely: on the same
    // record Java rendered birthDateForDisplay as "01/15/1990" (MM/dd/yyyy)
    // but birthTimeForDisplay as "15/01/1990" (dd/MM/yyyy) — two different
    // field orders from one source date. Whether that is intentional or a
    // Java bug is unresolved, so this asserts only the shape. Resolve it
    // before porting, then tighten this to the exact expected order; do not
    // let the Go port guess.
    if ("birthTimeForDisplay" in body) {
      expect(body.birthTimeForDisplay, `${BY_LAB_NUMBER} birthTimeForDisplay is dd/dd/dddd`).toMatch(
        /^\d{2}\/\d{2}\/\d{4}$/,
      );
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

    const itemCount = Number(
      query(
        "SELECT count(*) FROM clinlims.sample_item si" +
          " WHERE si.samp_id IN (SELECT sh.samp_id FROM clinlims.sample_human sh" +
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

  test("merge/details: GUID identities are counted but NOT listed (live-confirmed)", async ({ request }) => {
    // Live finding worth pinning: patient 1203 reported totalIdentifiers=2
    // while `identifiers` held exactly 1 entry. The missing one is
    // patient_identity_type 14 = GUID — an internal identifier that is counted
    // in the summary but excluded from the human-facing list.
    //
    // This is asserted as a real behavior because a Go port that naively
    // returns every patient_identity row would emit an extra element and leak
    // an internal GUID into a UI list. If this turns out to be a Java bug
    // rather than deliberate filtering, this test is the thing that forces
    // the decision to be made explicitly instead of by accident.
    const rows = query(
      "SELECT pi.patient_id FROM clinlims.patient_identity pi WHERE pi.identity_type_id = 14" +
        " AND EXISTS (SELECT 1 FROM clinlims.patient_identity o WHERE o.patient_id = pi.patient_id" +
        " AND o.identity_type_id <> 14) ORDER BY pi.patient_id LIMIT 1",
    );
    test.skip(rows.length === 0, "no patient has a GUID identity alongside another type");
    const patientId = rows[0][0];

    const dbTotal = Number(
      query(`SELECT count(*) FROM clinlims.patient_identity WHERE patient_id = ${patientId}`)[0][0],
    );
    const guidCount = Number(
      query(
        `SELECT count(*) FROM clinlims.patient_identity WHERE patient_id = ${patientId} AND identity_type_id = 14`,
      )[0][0],
    );

    const parsed = await readJson(await request.get(`${MERGE_DETAILS}/${patientId}`), MERGE_DETAILS);
    expect(parsed.dataSummary.totalIdentifiers, "totalIdentifiers counts ALL identity rows").toBe(dbTotal);
    expect(parsed.identifiers.length, "identifiers list EXCLUDES GUID rows").toBe(dbTotal - guidCount);
    for (const ident of parsed.identifiers) {
      expect(ident.identityType, "no GUID identity is listed").not.toMatch(/guid/i);
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

  // ── 5. rest/patient-photos/{id}/{isThumbnail} ───────────────────────────

  test("patient-photos: same base64 envelope for both thumbnail and full", async ({ request }) => {
    const rows = query("SELECT id FROM clinlims.patient ORDER BY id LIMIT 1");
    const patientId = rows[0][0];

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
    }
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

  test("all c1 endpoints refuse anonymous access (PHI boundary)", async ({ playwright }) => {
    // A dedicated context with NO stored auth state — the rest of this suite
    // runs authenticated, so without this the auth boundary is never actually
    // exercised for these endpoints.
    const anon = await playwright.request.newContext({
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
        // Java answers 302 → login. The invariant that matters is "not a
        // successful PHI-bearing 200", so accept any non-2xx and assert no
        // patient data came back.
        expect(res.status(), `anonymous ${path} must not succeed`).not.toBe(200);
        const text = await res.text();
        expect(text, `anonymous ${path} must not leak a patient id`).not.toContain(`"patientId"`);
        expect(text, `anonymous ${path} must not leak a nationalId`).not.toContain(`"nationalId"`);
      }
    } finally {
      await anon.dispose();
    }
  });
});
