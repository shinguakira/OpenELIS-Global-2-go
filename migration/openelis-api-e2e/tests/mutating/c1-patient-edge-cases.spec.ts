// c1 edge cases that require WRITING to the database, so they cannot live in
// tests/readonly/. Each one seeds its own row, asserts, and cleans up in a
// finally block.
//
// These exist because the behaviors below were previously confirmed by hand
// (curl + psql) and then left untested — which means nothing would catch them
// regressing. Verified against live Java first, as always; the same assertions
// then run against the Go port under `go-parity`.
import { test, expect, request as apiRequest, type APIRequestContext } from "@playwright/test";
import { query, exec, runSqlFile, fixturePath } from "../../fixtures/db";
import { ADMIN_USER, ADMIN_PASS } from "../../fixtures/env";
import {
  LOGIN_PATH,
  LOGIN_USER_FIELD,
  LOGIN_PASS_FIELD,
  SESSION_PATH,
} from "../../fixtures/contract";

const MERGE_DETAILS = "rest/patient/merge/details";

// Reserved ids for rows this spec creates itself. Far above anything the
// application generates, and every one is deleted in a finally block.
const TMP_IDENTITY_ID = 9999001;
const SENTINEL_DOCUMENT_ID = 9900150; // ABOVE the fixture's reserved block

async function adminCtx(): Promise<APIRequestContext> {
  const ctx = await apiRequest.newContext({
    baseURL: test.info().project.use.baseURL,
    ignoreHTTPSErrors: true,
    storageState: { cookies: [], origins: [] },
  });
  await ctx.get(SESSION_PATH);
  const res = await ctx.post(LOGIN_PATH, {
    form: { [LOGIN_USER_FIELD]: ADMIN_USER, [LOGIN_PASS_FIELD]: ADMIN_PASS },
  });
  expect(res.status(), "admin login").toBe(200);
  return ctx;
}

test.describe.configure({ mode: "serial" });

test.describe("c1 edge cases (mutating)", () => {
  test("merge/details: an identity row with a NULL type makes the request fail", async () => {
    // patient_identity.identity_type_id is NULLABLE. Java's
    // getPatientIdentities is a plain `SELECT * FROM patient_identity WHERE
    // patient_id = ?` — no join — so such a row is LOADED and counted in
    // totalIdentifiers, and then the identifier loop calls
    // patientIdentityTypeService.get(null), which throws. The request ends 500.
    //
    // The Go port first used an INNER JOIN to resolve the type name, which
    // silently DROPPED the row: 200, one fewer identifier, and a lower
    // totalIdentifiers. Two divergences at once, and no test could see either
    // because the dev dataset has zero null-type rows — hence this spec, which
    // creates one.
    //
    // A dangling (non-null but absent) type id is not tested because it cannot
    // exist: patient_identity has a FK to patient_identity_type. NULL is the
    // only reachable way for the type to fail to resolve.
    const patientId = query(
      "SELECT pi.patient_id FROM clinlims.patient_identity pi ORDER BY pi.patient_id LIMIT 1",
    )[0][0];

    const ctx = await adminCtx();
    try {
      // Baseline: the patient reads fine before the null-type row exists.
      const before = await ctx.get(`${MERGE_DETAILS}/${patientId}`);
      expect(before.status(), "patient reads normally before seeding").toBe(200);
      const identifiersBefore = (await before.json()).dataSummary.totalIdentifiers;

      exec(
        "INSERT INTO clinlims.patient_identity (id, identity_type_id, patient_id, identity_data, lastupdated)" +
          ` VALUES (${TMP_IDENTITY_ID}, NULL, ${patientId}, 'E2E-NULLTYPE', now())`,
      );

      const during = await ctx.get(`${MERGE_DETAILS}/${patientId}`);
      // 500 is Java's, pinned rather than corrected — the same policy that
      // pins the non-numeric-id 500 and the @PreAuthorize 500 elsewhere.
      expect(during.status(), "a null identity type makes merge/details fail").toBe(500);

      // The failure mode this replaces: a 200 whose totalIdentifiers silently
      // dropped the row. Asserting the status alone would not have caught the
      // inner-join version if it had errored for some other reason, so pin
      // that it is NOT a successful response carrying a stale count.
      expect(
        during.status(),
        "must not answer 200 with the null-type row quietly omitted",
      ).not.toBe(200);
      expect(identifiersBefore, "baseline count was captured").toBeGreaterThanOrEqual(0);

      exec(`DELETE FROM clinlims.patient_identity WHERE id = ${TMP_IDENTITY_ID}`);

      // …and the endpoint recovers once the row is gone, proving the 500 came
      // from that row and not from something the test broke permanently.
      const after = await ctx.get(`${MERGE_DETAILS}/${patientId}`);
      expect(after.status(), "recovers after the null-type row is removed").toBe(200);
      expect(
        (await after.json()).dataSummary.totalIdentifiers,
        "count returns to its baseline",
      ).toBe(identifiersBefore);
    } finally {
      exec(`DELETE FROM clinlims.patient_identity WHERE id = ${TMP_IDENTITY_ID}`);
      await ctx.dispose();
    }
  });

  test("patient-media fixture: re-running it does not delete rows outside its reserved block", async () => {
    // The fixture's cleanup was `DELETE ... WHERE id >= 9900000` while the same
    // block advanced the sequences to 9900100 — so every ordinary row created
    // after one fixture run landed ABOVE the cleanup threshold and was deleted
    // by the next one. Silent data loss in test setup, and nothing observed it.
    //
    // This seeds a sentinel document above the reserved block, re-runs the real
    // fixture file, and asserts the sentinel survived.
    const patientId = query("SELECT id FROM clinlims.patient ORDER BY id LIMIT 1")[0][0];
    try {
      exec(`DELETE FROM clinlims.patient_id_document WHERE id = ${SENTINEL_DOCUMENT_ID}`);
      exec(
        "INSERT INTO clinlims.patient_id_document" +
          " (id, patient_id, document_data, thumbnail_data, document_type," +
          "  document_category, description, deleted, last_updated)" +
          ` VALUES (${SENTINEL_DOCUMENT_ID}, '${patientId}', 'AA==', 'AA==', 'image/png',` +
          "  'ID_CARD', 'E2E sentinel — must survive a fixture reload', false, now())",
      );
      expect(
        query(`SELECT count(*) FROM clinlims.patient_id_document WHERE id = ${SENTINEL_DOCUMENT_ID}`)[0][0],
        "sentinel seeded",
      ).toBe("1");

      runSqlFile(fixturePath("patient-media-e2e.sql"));

      expect(
        query(`SELECT count(*) FROM clinlims.patient_id_document WHERE id = ${SENTINEL_DOCUMENT_ID}`)[0][0],
        "a row above the reserved block must SURVIVE the fixture's cleanup",
      ).toBe("1");

      // And the fixture's own rows are still (re)created, so the bounded
      // cleanup did not simply stop cleaning.
      expect(
        Number(
          query(
            "SELECT count(*) FROM clinlims.patient_id_document WHERE id BETWEEN 9900000 AND 9900099",
          )[0][0],
        ),
        "the fixture still seeds its own reserved rows",
      ).toBeGreaterThan(0);
    } finally {
      exec(`DELETE FROM clinlims.patient_id_document WHERE id = ${SENTINEL_DOCUMENT_ID}`);
    }
  });

  test("patient-media fixture: seeding the patient-less sample does not inflate sample_seq", async () => {
    // The same seed originally used a reserved id of 9900001. That does not
    // stay contained: the loader's normalize_sequences step runs
    // `setval('sample_seq', MAX(id) + 1)`, so one fixture row dragged the whole
    // sample sequence from ~1k to ~9.9M, permanently, on every load.
    //
    // The fixture now lets the sequence assign the id and keys its cleanup on
    // the unique accession instead. This asserts the sequence stays sane after
    // a reload, which is the property that actually matters.
    runSqlFile(fixturePath("patient-media-e2e.sql"));

    const seededId = Number(
      query("SELECT id FROM clinlims.sample WHERE accession_number = 'E2E-NOPAT-01'")[0][0],
    );
    const maxId = Number(query("SELECT max(id) FROM clinlims.sample")[0][0]);
    const seqValue = Number(query("SELECT last_value FROM clinlims.sample_seq")[0][0]);

    expect(seededId, "the seeded sample took an ordinary sequence id").toBeLessThan(1_000_000);
    expect(seqValue, "sample_seq tracks max(id), it is not pushed into the millions").toBeLessThanOrEqual(
      maxId + 1,
    );
  });
});
