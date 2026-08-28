// §6 — c2 (continued): the Wave 4 form-load endpoints, 4.5–4.8.
//
// These four are in Wave 4's list exactly like the rest of c2
// (migration/endpoint-migration-order.md). An earlier pass excluded them as
// "their own unit of work", which is not a reason — it is scope narrowing. They
// live in their own file only because c2-sample-order-reads.spec.ts is already
// long, and they are wired into the same `go-parity` project.
//
//   4.5  rest/GenericSampleOrder      (GET, requires accessionNumber)
//   4.6  rest/SamplePatientEntry
//   4.7  rest/SampleEdit
//   4.8  rest/SampleBatchEntrySetup
//
// ── MIGRATION POLICY ───────────────────────────────────────────────────────
// This is a migration, not a bug-fix pass. Where Java is broken, these tests
// PIN the broken behavior so the port reproduces it knowingly; the defects are
// listed in migration/java-defects-found.md to be raised separately.
//
// ── WHY THE FIXTURES MATTER HERE ───────────────────────────────────────────
// SampleEdit filters its sample items by SampleStatus.Entered, and not one row
// in the stock dataset carries that status — so existingTests, possibleTests
// and the real maxAccessionNumber branch were all unreachable and the response
// was pinned to its fallbacks whatever the server did.
// src/test/resources/fixtures/sample-edit-e2e.sql seeds both an accession whose
// items are all Entered and one whose LAST item is not, which is what makes the
// filter observable rather than merely present.
import { test, expect } from "@playwright/test";
import { readJson, expectKeysWithin, expectNonEmptyString } from "../../fixtures/assert";
import { query } from "../../fixtures/db";

const GENERIC_SAMPLE_ORDER = "rest/GenericSampleOrder";
const SAMPLE_PATIENT_ENTRY = "rest/SamplePatientEntry";
const SAMPLE_EDIT = "rest/SampleEdit";
const BATCH_ENTRY_SETUP = "rest/SampleBatchEntrySetup";

/** dd/MM/yyyy, the format DateUtil.getCurrentDateAsText produces. */
function todayDDMMYYYY(): string {
  const d = new Date();
  const p = (n: number) => String(n).padStart(2, "0");
  return `${p(d.getDate())}/${p(d.getMonth() + 1)}/${d.getFullYear()}`;
}

/** Every {id,value} row, sorted, for order-independent comparison. */
function idValuePairs(rows: any[]): string[] {
  return rows.map((r) => `${r.id} | ${r.value}`).sort();
}

// ───────────────────────────────────────────────────────────────────────────
// 4.5 — rest/GenericSampleOrder
// ───────────────────────────────────────────────────────────────────────────
test.describe("c2 form loads — GenericSampleOrder", () => {
  test("GenericSampleOrder: 400 without the param, 404 for an unknown accession", async ({
    request,
  }) => {
    // accessionNumber is @RequestParam(required = true), so Spring rejects a
    // missing one at binding with its ProblemDetail envelope — five keys, not
    // the {error} shape the handler itself produces.
    const missing = await request.get(GENERIC_SAMPLE_ORDER);
    expect(missing.status(), `${GENERIC_SAMPLE_ORDER} without accessionNumber`).toBe(400);
    const problem = await missing.json();
    expectKeysWithin(
      problem,
      ["type", "title", "status", "detail", "instance"],
      ["type", "title", "status", "detail", "instance"],
      `${GENERIC_SAMPLE_ORDER} 400 ProblemDetail`,
    );
    expect(problem.status, "ProblemDetail.status mirrors the HTTP status").toBe(400);
    // Spring emits the UNRESOLVED message keys here — the ProblemDetail
    // message source is not wired — so these are literal
    // "problemDetail.*" strings rather than English sentences.
    expect(problem.type, "ProblemDetail.type is an unresolved message key").toContain(
      "MissingServletRequestParameterException",
    );
    expect(problem.instance, "ProblemDetail.instance is the request path").toContain(
      "/rest/GenericSampleOrder",
    );

    // An accession with no sample takes the handler's OWN not-found branch,
    // which is a different envelope: Map.of("error", "..."), one key, and the
    // message interpolates the value the caller sent.
    const unknown = await request.get(`${GENERIC_SAMPLE_ORDER}?accessionNumber=NOPE_XYZ`);
    expect(unknown.status(), `${GENERIC_SAMPLE_ORDER} unknown accession`).toBe(404);
    expect(await unknown.json(), `${GENERIC_SAMPLE_ORDER} 404 envelope`).toEqual({
      error: "No sample found with accession number: NOPE_XYZ",
    });
  });

  test("GenericSampleOrder: 500 for every accession that EXISTS — a Java defect", async ({
    request,
  }) => {
    // MIGRATION POLICY: pinned, not fixed. Same root cause as
    // unassigned-sample/items — a numeric id bound to a String-mapped
    // property:
    //
    //   Failed to retrieve notebook sample for accession: E2E001
    //   Parameter value [10002] did not match expected type [java.lang.String]
    //
    // The exception marks the transaction rollback-only, the commit at the
    // @Transactional boundary throws, and the handler's catch-all wraps the
    // resulting message into its {error} envelope with a 500.
    //
    // So this endpoint has an inverted success contract: a NONEXISTENT
    // accession is the only input that produces a clean answer (404), and
    // every real one fails. Verified across three different real accessions
    // rather than one, so it reads as structural rather than as one bad row.
    const accessions = query(
      "SELECT accession_number FROM clinlims.sample WHERE accession_number IS NOT NULL" +
        " ORDER BY id LIMIT 3",
    ).map((r) => r[0]);
    expect(accessions.length, "the dataset has samples to try").toBeGreaterThan(0);

    for (const accession of accessions) {
      const res = await request.get(
        `${GENERIC_SAMPLE_ORDER}?accessionNumber=${encodeURIComponent(accession)}`,
      );
      expect(res.status(), `${GENERIC_SAMPLE_ORDER}?accessionNumber=${accession}`).toBe(500);
      const body = await res.json();
      expect(Object.keys(body), `${GENERIC_SAMPLE_ORDER} 500 envelope`).toEqual(["error"]);
      // The prefix is the handler's; the tail is the propagated Spring
      // message. Pinning the prefix alone would let a port emit any failure
      // text, so the rollback wording is pinned too — it is what identifies
      // WHICH failure this is.
      expect(body.error, `${GENERIC_SAMPLE_ORDER} 500 message`).toBe(
        "Failed to retrieve generic sample order: Transaction silently rolled back" +
          " because it has been marked as rollback-only",
      );
    }
  });
});

// ───────────────────────────────────────────────────────────────────────────
// 4.6 — rest/SamplePatientEntry
// ───────────────────────────────────────────────────────────────────────────
test.describe("c2 form loads — SamplePatientEntry", () => {
  test("SamplePatientEntry: the form envelope and its literals", async ({ request }) => {
    const body = await readJson(await request.get(SAMPLE_PATIENT_ENTRY), SAMPLE_PATIENT_ENTRY);

    expectKeysWithin(
      body,
      ["cancelAction", "cancelMethod", "currentDate", "customNotificationLogic", "formMethod",
       "formName", "initialSampleConditionList", "orderEntryOnly", "patientProperties",
       "patientSearch", "patientUpdateStatus", "projects", "referralOrganizations",
       "referralReasons", "rejectReasonList", "sampleOrderItems", "sampleTypes", "sampleXML",
       "submitOnCancel", "testSectionList", "useReferral", "warning"],
      ["cancelAction", "cancelMethod", "currentDate", "formMethod", "formName",
       "patientProperties", "patientSearch", "patientUpdateStatus", "projects",
       "sampleOrderItems", "sampleTypes", "testSectionList"],
      `${SAMPLE_PATIENT_ENTRY} envelope`,
    );

    // Form metadata is hardcoded in the controller. formName differs per
    // endpoint and is the cheapest way to catch a port that serves one form
    // builder for all three loads.
    expect(body.formName, "formName").toBe("samplePatientEntryForm");
    expect(body.formMethod, "formMethod").toBe("POST");
    expect(body.cancelAction, "cancelAction").toBe("Home");
    expect(body.cancelMethod, "cancelMethod").toBe("POST");
    expect(body.patientUpdateStatus, "patientUpdateStatus is the literal ADD").toBe("ADD");
    // sampleXML is initialised to "" and never populated on the GET path.
    expect(body.sampleXML, "sampleXML is empty on load").toBe("");

    // DateUtil.getCurrentDateAsText() — dd/MM/yyyy, not ISO.
    expect(body.currentDate, "currentDate is today in dd/MM/yyyy").toBe(todayDDMMYYYY());

    // Primitive booleans on the form object, so all five always serialize.
    for (const k of ["customNotificationLogic", "orderEntryOnly", "submitOnCancel", "useReferral", "warning"]) {
      expect(typeof body[k], `${k} is a boolean`).toBe("boolean");
    }
    // warning is FALSE here and TRUE on SampleEdit — same key, opposite
    // literal, set by two different controllers.
    expect(body.warning, "SamplePatientEntry sets warning false").toBe(false);
  });

  test("SamplePatientEntry: reference lists come from the DB, not from a stub", async ({
    request,
  }) => {
    const body = await readJson(await request.get(SAMPLE_PATIENT_ENTRY), SAMPLE_PATIENT_ENTRY);

    // testSectionList is ListType.TEST_SECTION_ACTIVE — every ACTIVE test
    // section, and only those. Compared as a set against the DB so a port that
    // returns all sections (active or not) fails.
    const dbSections = query(
      "SELECT id, name FROM clinlims.test_section WHERE is_active = 'Y'",
    );
    expect(body.testSectionList.length, "testSectionList is the active test sections").toBe(
      dbSections.length,
    );
    expect(
      body.testSectionList.map((r: any) => r.id).sort(),
      "testSectionList ids match the active test sections",
    ).toEqual(dbSections.map((r) => r[0]).sort());

    // projects is EVERY project row, and unlike the {id,value} lists it is the
    // full entity — a port emitting {id,value} here would be a different
    // response even though the ids match.
    const dbProjects = query("SELECT id FROM clinlims.project");
    expect(body.projects.length, "projects is every project").toBe(dbProjects.length);
    expect(body.projects.map((p: any) => p.id).sort(), "project ids").toEqual(
      dbProjects.map((r) => r[0]).sort(),
    );
    for (const p of body.projects) {
      expectKeysWithin(
        p,
        // programCode is present on SOME rows only (nullable column), so it
        // belongs in the allowed set but not the required one.
        ["lastupdated", "id", "projectName", "description", "isActive", "programCode",
         "concatProjNameDesc", "organizations"],
        ["id", "projectName", "isActive", "concatProjNameDesc", "organizations"],
        "project row",
      );
      // concatProjNameDesc is DERIVED — name + "+" + description — not a
      // column. It is the field most likely to be dropped by a port that maps
      // the table straight through.
      const expected =
        p.description === undefined || p.description === null
          ? p.projectName
          : `${p.projectName}+${p.description}`;
      expect(p.concatProjNameDesc, `project ${p.id} concatProjNameDesc is derived`).toBe(expected);
      expect(Array.isArray(p.organizations), `project ${p.id} organizations is an array`).toBe(true);
    }

    // The dictionary-backed {id,value} lists. Non-empty is asserted explicitly:
    // every one of them would serialize as [] on a port that never wired the
    // list service, and [] passes any shape-only check.
    for (const k of ["initialSampleConditionList", "rejectReasonList", "referralReasons", "sampleTypes"]) {
      expect(Array.isArray(body[k]), `${k} is an array`).toBe(true);
      expect(body[k].length, `${k} is populated, not an empty stub`).toBeGreaterThan(0);
      for (const row of body[k]) {
        expectKeysWithin(row, ["id", "value"], ["id", "value"], `${k} row`);
        expectNonEmptyString(row.value, `${k} row value`);
      }
    }
  });

  test("SamplePatientEntry: sampleTypes is ROLE-FILTERED, unlike SampleBatchEntrySetup", async ({
    request,
  }) => {
    // The sharpest discriminator in this group. Both endpoints emit a
    // `sampleTypes` key of {id,value} rows, but they come from different
    // sources:
    //
    //   SamplePatientEntry     userService.getUserSampleTypes(user, ROLE_RECEPTION)
    //   SampleBatchEntrySetup  every ACTIVE type_of_sample
    //
    // On this dataset that is 12 vs 14. A port that builds one list and reuses
    // it for both passes every other assertion in this file.
    const entry = await readJson(await request.get(SAMPLE_PATIENT_ENTRY), SAMPLE_PATIENT_ENTRY);
    const batch = await readJson(await request.get(BATCH_ENTRY_SETUP), BATCH_ENTRY_SETUP);

    const allActive = query("SELECT id FROM clinlims.type_of_sample WHERE is_active = 'Y'").map(
      (r) => r[0],
    );
    expect(
      batch.sampleTypes.map((t: any) => t.id).sort(),
      "SampleBatchEntrySetup lists every active type_of_sample",
    ).toEqual([...allActive].sort());

    const entryIds = entry.sampleTypes.map((t: any) => t.id);
    // A strict, non-empty subset: proves the filter runs AND that it is not
    // filtering everything away.
    expect(entryIds.length, "SamplePatientEntry sampleTypes is non-empty").toBeGreaterThan(0);
    expect(
      entryIds.length,
      "SamplePatientEntry sampleTypes is a STRICT subset of the active types",
    ).toBeLessThan(allActive.length);
    for (const id of entryIds) {
      expect(allActive, `sample type ${id} is an active type_of_sample`).toContain(id);
    }
  });

  test("SamplePatientEntry: patientProperties and patientSearch sub-forms", async ({ request }) => {
    const body = await readJson(await request.get(SAMPLE_PATIENT_ENTRY), SAMPLE_PATIENT_ENTRY);

    // patientProperties on THIS path is the blank-form variant: the lists a
    // new-patient form needs, with no patient loaded. Contrast with
    // order/search's patientProperties, which is the POPULATED bean and shares
    // almost none of these keys — same key name, two different objects.
    expectKeysWithin(
      body.patientProperties,
      ["addressDepartments", "addressHierarchy", "birthDateForDisplay", "educationList",
       "genders", "healthDistricts", "healthRegions", "idDocuments", "maritialList",
       "nationalityList", "patientType", "patientTypes", "readOnly"],
      ["addressHierarchy", "birthDateForDisplay", "genders", "maritialList", "nationalityList",
       "patientType", "patientTypes", "readOnly"],
      `${SAMPLE_PATIENT_ENTRY} patientProperties`,
    );
    expect(body.patientProperties.birthDateForDisplay, "no patient loaded, so no birth date").toBe("");
    expect(body.patientProperties.patientType, "no patient loaded, so no patient type").toBe("");
    expect(body.patientProperties.readOnly, "readOnly is a primitive boolean").toBe(false);
    expect(body.patientProperties.addressHierarchy, "addressHierarchy is an empty map").toEqual({});
    expect(Array.isArray(body.patientProperties.idDocuments), "idDocuments is an array").toBe(true);

    // patientTypes is the full entity, not {id,value} — a different shape from
    // the sibling lists in the same object.
    expect(body.patientTypes, "patientTypes is not hoisted to the top level").toBeUndefined();
    for (const pt of body.patientProperties.patientTypes) {
      expectKeysWithin(
        pt,
        ["lastupdated", "isActive", "id", "type", "description"],
        ["id", "type", "description"],
        "patientType row",
      );
    }
    const dbPatientTypes = query("SELECT id FROM clinlims.patient_type");
    expect(
      body.patientProperties.patientTypes.map((p: any) => p.id).sort(),
      "patientTypes comes from patient_type",
    ).toEqual(dbPatientTypes.map((r) => r[0]).sort());

    // genders appears in BOTH patientProperties and patientSearch and must be
    // the same list — they are built from the same source.
    expectKeysWithin(
      body.patientSearch,
      ["defaultHeader", "genders", "loadFromServerWithPatient", "searchCriteria"],
      ["defaultHeader", "genders", "loadFromServerWithPatient", "searchCriteria"],
      `${SAMPLE_PATIENT_ENTRY} patientSearch`,
    );
    expect(
      idValuePairs(body.patientSearch.genders),
      "patientSearch.genders equals patientProperties.genders",
    ).toEqual(idValuePairs(body.patientProperties.genders));
    // SamplePatientEntry does NOT set loadFromServerWithPatient; SampleEdit
    // does. Pinned so a port does not share one PatientSearch builder.
    expect(body.patientSearch.loadFromServerWithPatient, "false on this form").toBe(false);
    expect(body.patientSearch.searchCriteria.length, "searchCriteria is populated").toBeGreaterThan(0);
  });

  test("SamplePatientEntry: sampleOrderItems is the FORM variant, not order/search's", async ({
    request,
  }) => {
    // Third distinct object under a key this migration has now seen three
    // times. order/search's sampleOrderItems carries labNo / collectionDate /
    // priority; this one carries the LISTS a blank order form needs, plus a
    // request date and a received date/time stamped at load.
    const body = await readJson(await request.get(SAMPLE_PATIENT_ENTRY), SAMPLE_PATIENT_ENTRY);
    const s = body.sampleOrderItems;

    expectKeysWithin(
      s,
      ["environmentalFields", "isEQASample", "modified", "paymentOptions", "priorityList",
       "programList", "providersList", "readOnly", "receivedDateForDisplay", "receivedTime",
       "referringSiteList", "requestDate", "testLocationCodeList"],
      ["environmentalFields", "isEQASample", "modified", "paymentOptions", "priorityList",
       "readOnly", "receivedDateForDisplay", "receivedTime", "requestDate"],
      `${SAMPLE_PATIENT_ENTRY} sampleOrderItems`,
    );
    // No labNo here — there is no sample yet. Asserting the ABSENCE stops a
    // port from reusing order/search's builder.
    for (const absent of ["labNo", "sampleId", "collectionDate", "priority"]) {
      expect(absent in s, `sampleOrderItems omits ${absent} on the blank form`).toBe(false);
    }

    expect(s.requestDate, "requestDate is today").toBe(todayDDMMYYYY());
    expect(s.receivedDateForDisplay, "receivedDateForDisplay is today").toBe(todayDDMMYYYY());
    expect(s.receivedTime, "receivedTime is HH:mm").toMatch(/^\d{2}:\d{2}$/);
    expect(s.environmentalFields, "environmentalFields is an empty map on load").toEqual({});
    for (const k of ["isEQASample", "modified", "readOnly"]) {
      expect(typeof s[k], `sampleOrderItems.${k} is a boolean`).toBe("boolean");
    }

    // paymentOptions is dictionary-backed; cross-checked against the DB the
    // same way the order/search spec does it.
    expect(s.paymentOptions.length, "paymentOptions is populated").toBeGreaterThan(0);
    for (const row of s.paymentOptions) {
      const hits = query(
        `SELECT dict_entry FROM clinlims.dictionary WHERE id = '${row.id}'`,
      );
      expect(hits.length, `paymentOption ${row.id} is a real dictionary row`).toBe(1);
      expect(row.value, `paymentOption ${row.id} value comes from dict_entry`).toBe(hits[0][0]);
    }

    // priorityList is an ENUM, not a table: its ids are names, not numbers.
    expect(s.priorityList.length, "priorityList is populated").toBeGreaterThan(0);
    for (const row of s.priorityList) {
      expect(row.id, `priority ${row.id} is a non-numeric enum name`).toMatch(/^[A-Z_]+$/);
    }
  });
});

// ───────────────────────────────────────────────────────────────────────────
// 4.7 — rest/SampleEdit
// ───────────────────────────────────────────────────────────────────────────
test.describe("c2 form loads — SampleEdit", () => {
  test("SampleEdit: no accession loads a blank form, an unknown one reports noSampleFound", async ({
    request,
  }) => {
    // Three states, all 200 — this endpoint never 404s. A port that answered
    // 404 for an unknown accession would be a different API.
    const blank = await readJson(await request.get(SAMPLE_EDIT), SAMPLE_EDIT);
    expect(blank.searchFinished, "no accession -> searchFinished false").toBe(false);
    expect(blank.noSampleFound, "no accession -> noSampleFound false, NOT true").toBe(false);
    // accessionNumber is never set on this branch, so NON_NULL drops it.
    expect("accessionNumber" in blank, "blank form omits accessionNumber").toBe(false);

    const unknown = await readJson(
      await request.get(`${SAMPLE_EDIT}?accessionNumber=NOPE_XYZ`),
      SAMPLE_EDIT,
    );
    expect(unknown.searchFinished, "unknown accession -> searchFinished TRUE").toBe(true);
    expect(unknown.noSampleFound, "unknown accession -> noSampleFound true").toBe(true);
    // The searched-for value is echoed back even though nothing matched.
    expect(unknown.accessionNumber, "unknown accession is echoed").toBe("NOPE_XYZ");

    // The patient block splits two ways on this branch, and the split is the
    // point: the scalar fields are initialised to "" on the form object, so
    // they are PRESENT AND EMPTY, while the two LISTS are left null and
    // dropped by Include.NON_NULL. A port that emitted nulls for the scalars,
    // or [] for the lists, would be wrong in opposite directions.
    for (const k of ["patientName", "dob", "gender", "nationalId", "patientId",
                     "subjectNumber", "maxAccessionNumber"]) {
      expect(k in unknown, `unknown accession keeps ${k}`).toBe(true);
      expect(unknown[k], `unknown accession leaves ${k} empty`).toBe("");
    }
    for (const k of ["existingTests", "possibleTests"]) {
      expect(k in unknown, `unknown accession drops ${k} entirely`).toBe(false);
    }
    // Primitive booleans, so they survive as false rather than disappearing.
    expect(unknown.ableToCancelResults, "ableToCancelResults is false").toBe(false);
    expect(unknown.isConfirmationSample, "isConfirmationSample is false").toBe(false);
  });

  test("SampleEdit: accession-format scalars and form literals", async ({ request }) => {
    const body = await readJson(
      await request.get(`${SAMPLE_EDIT}?accessionNumber=E2E-EDIT-01`),
      SAMPLE_EDIT,
    );

    expect(body.formName, "formName").toBe("SampleEditForm");
    expect(body.formAction, "formAction is set on this form only").toBe("SampleEdit");
    expect(body.formMethod, "formMethod").toBe("POST");
    expect(body.cancelAction, "cancelAction").toBe("Home");
    // warning is TRUE here and FALSE on SamplePatientEntry.
    expect(body.warning, "SampleEdit sets warning true").toBe(true);
    expect(body.currentDate, "currentDate is today in dd/MM/yyyy").toBe(todayDDMMYYYY());
    expect(body.sampleXML, "sampleXML is empty on load").toBe("");
    expect(body.newAccessionNumber, "newAccessionNumber is empty on load").toBe("");

    // Accession-format configuration, read from the configured generator.
    // NUMBERS, not strings — a port emitting "15" is a different response.
    expect(body.accessionFormat, "accessionFormat").toBe("SITEYEARNUM");
    expect(body.idSeparator, "idSeparator").toBe(";");
    expect(typeof body.maxAccessionLength, "maxAccessionLength is a number").toBe("number");
    expect(typeof body.editableAccession, "editableAccession is a number").toBe("number");
    expect(typeof body.nonEditableAccession, "nonEditableAccession is a number").toBe("number");
    // The split is a partition of the total, which is what makes these three
    // consistent rather than three independent constants.
    expect(
      body.editableAccession + body.nonEditableAccession,
      "editable + nonEditable is the accession length",
    ).toBe(body.maxAccessionLength);

    // isEditable comes from a SESSION attribute or ?type=readwrite, so a plain
    // GET is read-only. Pinned because it is the only stateful input on this
    // read path.
    expect(body.isEditable, "a plain GET is not editable").toBe(false);
  });

  test("SampleEdit: sample items are filtered by SampleEntered status", async ({ request }) => {
    // The behaviour sample-edit-e2e.sql exists for. getSampleItems calls
    // getSampleItemsBySampleIdAndStatus(id, {SampleStatus.Entered}), and no row
    // in the stock dataset has that status — so before the fixture, every
    // accession returned existingTests [], possibleTests [] and
    // maxAccessionNumber "<accession>-0", and a port that dropped the filter
    // produced exactly the same thing.
    const enteredStatus = query(
      "SELECT id FROM clinlims.status_of_sample" +
        " WHERE status_type = 'SAMPLE' AND name = 'SampleEntered'",
    );
    expect(enteredStatus.length, "the SampleEntered status exists").toBe(1);
    const enteredId = enteredStatus[0][0];

    for (const accession of ["E2E-EDIT-01", "E2E-EDIT-02"]) {
      const rows = query(
        "SELECT si.sort_order, si.status_id FROM clinlims.sample_item si" +
          " JOIN clinlims.sample s ON s.id = si.samp_id" +
          ` WHERE s.accession_number = '${accession}' ORDER BY si.sort_order::numeric`,
      );
      // Not a skip: the loader marks the fixture fatal.
      expect(rows.length, `${accession} fixture is loaded`).toBe(2);
      const entered = rows.filter((r) => r[1] === enteredId);

      const body = await readJson(
        await request.get(`${SAMPLE_EDIT}?accessionNumber=${accession}`),
        SAMPLE_EDIT,
      );
      expect(body.searchFinished, `${accession} searchFinished`).toBe(true);
      expect(body.noSampleFound, `${accession} noSampleFound`).toBe(false);

      // maxAccessionNumber appends the sort order of the LAST item in the
      // FILTERED list. E2E-EDIT-02's last item is excluded by status, so it
      // ends "-1" while its highest sort order is 2 — that gap is what a port
      // ignoring the filter gets wrong.
      const lastSortOrder = entered[entered.length - 1][0];
      expect(body.maxAccessionNumber, `${accession} maxAccessionNumber`).toBe(
        `${accession}-${lastSortOrder}`,
      );

      // One existingTests row per non-canceled analysis on a FILTERED item.
      const expectedTests = Number(
        query(
          "SELECT count(*) FROM clinlims.analysis a" +
            " JOIN clinlims.sample_item si ON si.id = a.sampitem_id" +
            " JOIN clinlims.sample s ON s.id = si.samp_id" +
            ` WHERE s.accession_number = '${accession}' AND si.status_id = ${enteredId}` +
            "   AND a.status_id <> (SELECT id FROM clinlims.status_of_sample" +
            "       WHERE status_type = 'ANALYSIS' AND name = 'Test Canceled')",
        )[0][0],
      );
      expect(body.existingTests.length, `${accession} existingTests`).toBe(expectedTests);
      expect(expectedTests, `${accession} really has tests, so this is not a vacuous check`)
        .toBeGreaterThan(0);

      for (const item of body.existingTests) {
        expectKeysWithin(
          item,
          ["accessionNumber", "add", "analysisId", "canCancel", "canRemoveSample", "canceled",
           "collectionDate", "collectionTime", "hasResults", "id", "removeSample",
           "sampleItemChanged", "sampleItemId", "sampleType", "sortOrder", "status", "testId",
           "testName"],
          ["accessionNumber", "add", "canCancel", "canRemoveSample", "canceled", "hasResults",
           "id", "removeSample", "sampleItemChanged", "sampleType", "testId", "testName"],
          "existingTests row",
        );
        // accessionNumber inside the row is SUFFIXED with the item's sort
        // order, unlike the top-level one.
        expect(item.accessionNumber, "row accessionNumber carries the item suffix").toMatch(
          new RegExp(`^${accession}-\\d+$`),
        );
        expect(item.id, "id duplicates testId").toBe(item.testId);
      }

      // possibleTests are the ADDABLE tests for the filtered items' sample
      // types. Only the FIRST row of each sample item carries accessionNumber
      // and sampleType — addPossibleTestsToList sets them once per item and
      // leaves them null on the rest, where Include.NON_NULL drops them. The
      // frontend reads them as group headers.
      //
      // So the count of rows carrying an accessionNumber equals the number of
      // FILTERED sample items, not the number of tests. A port that set the
      // header fields on every row would look more consistent and be wrong.
      expect(body.possibleTests.length, `${accession} possibleTests is populated`).toBeGreaterThan(0);
      const headers = body.possibleTests.filter((t: any) => "accessionNumber" in t);
      expect(headers.length, `${accession} one possibleTests header per filtered sample item`).toBe(
        entered.length,
      );
      expect(
        headers.length,
        `${accession} not every row is a header, so the pattern is observable`,
      ).toBeLessThan(body.possibleTests.length);

      for (const item of body.possibleTests) {
        expect(item.id, "possibleTests id duplicates testId").toBe(item.testId);
        expect(item.canceled, "an addable test is not canceled").toBe(false);
        expect(item.hasResults, "an addable test has no results").toBe(false);
        // sampleItemId is on EVERY row — it is the header fields specifically
        // that are once-per-item.
        expectNonEmptyString(item.sampleItemId, "possibleTests sampleItemId");
        if ("accessionNumber" in item) {
          expect(item.accessionNumber, "a header row's accession is suffixed").toMatch(
            new RegExp(`^${accession}-\\d+$`),
          );
          expectNonEmptyString(item.sampleType, "a header row carries sampleType");
        } else {
          expect("sampleType" in item, "a non-header row drops sampleType too").toBe(false);
        }
      }
    }
  });

  test("SampleEdit: patient block comes from the linked patient", async ({ request }) => {
    const body = await readJson(
      await request.get(`${SAMPLE_EDIT}?accessionNumber=E2E-EDIT-01`),
      SAMPLE_EDIT,
    );

    const [[patientId, firstName, lastName, dob, gender, nationalId]] = query(
      "SELECT p.id, pe.first_name, pe.last_name, COALESCE(p.entered_birth_date,'')," +
        " COALESCE(p.gender,''), COALESCE(p.national_id,'')" +
        " FROM clinlims.patient p" +
        " JOIN clinlims.person pe ON pe.id = p.person_id" +
        " JOIN clinlims.sample_human sh ON sh.patient_id = p.id" +
        " JOIN clinlims.sample s ON s.id = sh.samp_id" +
        " WHERE s.accession_number = 'E2E-EDIT-01'",
    );

    expect(body.patientId, "patientId").toBe(patientId);
    // getLastFirstName: "Last, First" — comma AND space, in that order.
    expect(body.patientName, "patientName is 'Last, First'").toBe(`${lastName}, ${firstName}`);
    expect(body.gender, "gender is the raw column").toBe(gender);
    expect(body.nationalId, "nationalId is the raw column").toBe(nationalId);
    // dob is the STORED entered_birth_date, emitted RAW — the same value
    // order/search reformats through DateUtil.formatStringDate. Two endpoints,
    // one column, two renderings.
    expect(body.dob, "dob is the stored entered_birth_date, unreformatted").toBe(dob);
  });
});

// ───────────────────────────────────────────────────────────────────────────
// 4.8 — rest/SampleBatchEntrySetup
// ───────────────────────────────────────────────────────────────────────────
test.describe("c2 form loads — SampleBatchEntrySetup", () => {
  test("SampleBatchEntrySetup: envelope, literals and the two project-data blocks", async ({
    request,
  }) => {
    const body = await readJson(await request.get(BATCH_ENTRY_SETUP), BATCH_ENTRY_SETUP);

    expectKeysWithin(
      body,
      ["cancelAction", "cancelMethod", "currentDate", "currentTime", "customNotificationLogic",
       "facilityIDCheck", "formMethod", "formName", "initialSampleConditionList", "localDBOnly",
       "orderEntryOnly", "patientInfoCheck", "patientUpdateStatus", "project", "projectDataEID",
       "projectDataVL", "projects", "sampleOrderItems", "sampleTypes", "sampleXML",
       "submitOnCancel", "testSectionList", "useReferral", "warning"],
      ["cancelAction", "cancelMethod", "currentDate", "currentTime", "formMethod", "formName",
       "patientUpdateStatus", "project", "projectDataEID", "projectDataVL", "projects",
       "sampleOrderItems", "sampleTypes", "testSectionList"],
      `${BATCH_ENTRY_SETUP} envelope`,
    );

    expect(body.formName, "formName").toBe("sampleBatchEntryForm");
    expect(body.formMethod, "formMethod").toBe("POST");
    expect(body.cancelAction, "cancelAction").toBe("Home");
    expect(body.patientUpdateStatus, "patientUpdateStatus").toBe("ADD");
    expect(body.project, "project is empty on load").toBe("");
    expect(body.sampleXML, "sampleXML is empty on load").toBe("");
    expect(body.currentDate, "currentDate is today in dd/MM/yyyy").toBe(todayDDMMYYYY());
    // currentTime is exclusive to this form — SamplePatientEntry has no such
    // key, though its sampleOrderItems carries a receivedTime.
    expect(body.currentTime, "currentTime is HH:mm").toMatch(/^\d{2}:\d{2}$/);

    for (const k of ["customNotificationLogic", "facilityIDCheck", "localDBOnly", "orderEntryOnly",
                     "patientInfoCheck", "submitOnCancel", "useReferral", "warning"]) {
      expect(typeof body[k], `${k} is a boolean`).toBe("boolean");
    }

    // projectDataEID and projectDataVL are two SEPARATE objects of the same
    // large shape — one per project flavour. A port that emitted one and
    // aliased the other would pass a key-count check but not this equality of
    // key SETS combined with their independence.
    const eidKeys = Object.keys(body.projectDataEID).sort();
    const vlKeys = Object.keys(body.projectDataVL).sort();
    expect(eidKeys.length, "projectDataEID is fully populated").toBeGreaterThan(50);
    expect(vlKeys, "projectDataVL has the same key set as projectDataEID").toEqual(eidKeys);
    // The list-valued members must be real lists, not nulls dropped by
    // NON_NULL — these are what the form's dropdowns bind to.
    for (const k of ["eidWhichPCRList", "eidSecondPCRReasonList", "hivStatusList", "isUnderInvestigationList"]) {
      expect(Array.isArray(body.projectDataEID[k]), `projectDataEID.${k} is an array`).toBe(true);
    }
  });

  test("SampleBatchEntrySetup: shares list builders with SamplePatientEntry where Java does", async ({
    request,
  }) => {
    // Three lists are built from the same source on both endpoints, and one is
    // not (sampleTypes — see the role-filter test). Asserting the shared ones
    // are IDENTICAL and that sampleTypes is not, together, pins which builder
    // each key uses. Either half alone would be satisfied by a port that
    // shared everything, or by one that shared nothing.
    const entry = await readJson(await request.get(SAMPLE_PATIENT_ENTRY), SAMPLE_PATIENT_ENTRY);
    const batch = await readJson(await request.get(BATCH_ENTRY_SETUP), BATCH_ENTRY_SETUP);

    for (const k of ["testSectionList", "initialSampleConditionList"]) {
      expect(idValuePairs(batch[k]), `${k} is identical on both form loads`).toEqual(
        idValuePairs(entry[k]),
      );
    }
    expect(
      batch.projects.map((p: any) => p.id).sort(),
      "projects is identical on both form loads",
    ).toEqual(entry.projects.map((p: any) => p.id).sort());

    expect(
      idValuePairs(batch.sampleTypes),
      "sampleTypes is NOT shared — batch is unfiltered, entry is role-filtered",
    ).not.toEqual(idValuePairs(entry.sampleTypes));
  });
});

// ───────────────────────────────────────────────────────────────────────────
// Auth boundary
// ───────────────────────────────────────────────────────────────────────────
test.describe("c2 form loads — auth", () => {
  test("form-load endpoints refuse anonymous access", async ({ playwright }) => {
    // Same default-deny boundary the rest of c2 asserts. These four assemble
    // patient names, national ids and provider lists, so an anonymous 200 here
    // would be a PHI leak, not just a missing redirect.
    // storageState MUST be passed explicitly. Without it the context inherits
    // the project's authenticated cookie jar, every request comes back 200, and
    // this test reports a passing auth boundary while asserting nothing — which
    // is exactly what it did on its first run.
    const anon = await playwright.request.newContext({
      baseURL: test.info().project.use.baseURL,
      ignoreHTTPSErrors: true,
      storageState: { cookies: [], origins: [] },
    });
    try {
      for (const path of [
        `${GENERIC_SAMPLE_ORDER}?accessionNumber=E2E001`,
        SAMPLE_PATIENT_ENTRY,
        `${SAMPLE_EDIT}?accessionNumber=E2E-EDIT-01`,
        BATCH_ENTRY_SETUP,
      ]) {
        const res = await anon.get(path, { maxRedirects: 0 });
        expect(res.status(), `anonymous ${path} must not succeed`).not.toBe(200);
        const text = await res.text();
        expect(text, `anonymous ${path} must not leak a patient name`).not.toContain(
          `"patientName"`,
        );
        expect(text, `anonymous ${path} must not leak a national id`).not.toContain(`"nationalId"`);
      }
    } finally {
      await anon.dispose();
    }
  });
});
