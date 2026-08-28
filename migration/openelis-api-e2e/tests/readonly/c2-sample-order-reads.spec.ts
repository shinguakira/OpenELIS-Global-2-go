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
// Wave 4 lists 17 endpoints. This file covers the sample/order reads; the four
// Type-D form loads moved to c2-sample-form-loads.spec.ts once they were
// ported. The ONLY group still outside c2 is the mutating shipment endpoints
// that share the unassigned-sample controller — reason at the bottom of the
// file. Earlier drafts also excluded the binary attachment endpoints and the
// form loads; both are covered now, and the "table is empty" exclusions were
// fixture bugs that have since been seeded.
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

  test("order/search: the response BYTES, not just the parsed values", async ({ request }) => {
    // Everything else in this file asserts on `await res.json()`, which is blind
    // by construction to any difference that survives a round trip. Two such
    // differences were real and shipped:
    //
    //   - Go's json.Encoder.Encode appends a trailing newline; Jackson does
    //     not. Every response was one byte longer with a different
    //     Content-Length.
    //   - Go renders a float64 5.0 as `5`; Jackson renders the same
    //     java.lang.Double as `5.0`. Both parse to 5, so a field-by-field
    //     comparison of the DECODED objects reports parity.
    //
    // Neither is visible to a JSON assertion, so this one reads the raw text.
    const res = await request.get(`${ORDER_SEARCH}?labNumber=${encodeURIComponent(anyAccession())}`);
    expect(res.status(), `${ORDER_SEARCH} raw fetch`).toBe(200);
    const text = await res.text();

    expect(text.endsWith("\n"), "the body must not end with a newline — Jackson emits none").toBe(false);
    expect(text.endsWith("}"), "it ends at the closing brace").toBe(true);
    // Content-Length has to agree with the bytes, which is what the stray
    // newline actually broke.
    const declared = res.headers()["content-length"];
    if (declared !== undefined) {
      expect(Number(declared), "Content-Length matches the body length").toBe(Buffer.byteLength(text, "utf8"));
    }

    // A Double is rendered with its fraction digit. sample_item.quantity is the
    // only such column reachable here, so the assertion is anchored to a sample
    // item that HAS one — with a NULL quantity the outer key is the string ""
    // and there is no number to format.
    const withQuantity = query(
      `SELECT s.accession_number, si.quantity
         FROM clinlims.sample_item si
         JOIN clinlims.sample s ON s.id = si.samp_id
        WHERE si.quantity IS NOT NULL AND si.voided = false
        ORDER BY si.id LIMIT 1`,
    );
    expect(withQuantity.length, "some sample item carries a quantity").toBe(1);
    const [accession, quantity] = withQuantity[0];
    const raw = await (await request.get(`${ORDER_SEARCH}?labNumber=${encodeURIComponent(accession)}`)).text();

    // Postgres renders numeric 5.0 as "5.0" and Java's Double.toString agrees,
    // so the column value is its own oracle here.
    expect(
      raw.includes(`"quantity":${quantity}`),
      `quantity is serialized as ${quantity}, the way Double.toString writes it`,
    ).toBe(true);
    expect(
      new RegExp(`"quantity":${String(quantity).replace(/\.0$/, "")}([,}])`).test(raw),
      "...and NOT with the fraction digit stripped, which is Go's default",
    ).toBe(false);
  });

  test("order/search: a missing labNumber is rejected with 400", async ({ request }) => {
    const res = await request.get(ORDER_SEARCH);
    expect(res.status(), `${ORDER_SEARCH} missing labNumber`).toBe(400);
  });
  test("order/search: the WHOLE sampleOrderItems map, on an order that trips every branch", async ({
    request,
  }) => {
    // `buildSampleOrderItems` puts ~25 keys, almost all conditionally. On the
    // stock dataset every condition is false — no sample carries a provider, a
    // requester organization, a program or any observation history — so the
    // live response was a six-key object and a port that built only those six
    // keys passed. The test was green and the parity was fictional.
    //
    // order-search-full-e2e.sql seeds E2E-FULL-01 to trip all of them at once.
    // Every value below is oracled against the row it came from, so a port that
    // reads the wrong column or the wrong observation type fails on the VALUE,
    // not merely on the key being present.
    const body = await readJson(
      await request.get(`${ORDER_SEARCH}?labNumber=E2E-FULL-01`),
      `${ORDER_SEARCH} E2E-FULL-01`,
    );
    const soi = body.sampleOrderItems;

    // ── provider: sample_human.provider_id -> provider -> person ──────────
    const prov = query(
      `SELECT pe.id, pe.first_name, pe.last_name, pe.work_phone, pe.email, pe.fax
         FROM clinlims.sample s
         JOIN clinlims.sample_human sh ON sh.samp_id = s.id
         JOIN clinlims.provider pr ON pr.id = sh.provider_id
         JOIN clinlims.person pe ON pe.id = pr.person_id
        WHERE s.accession_number = 'E2E-FULL-01'`,
    );
    expect(prov.length, "E2E-FULL-01 has a provider (fixture loaded?)").toBe(1);
    const [pid, pfirst, plast, pphone, pemail, pfax] = prov[0];
    expect(soi.providerPersonId, "providerPersonId is the PERSON id, not the provider id").toBe(String(pid));
    expect(soi.providerFirstName, "providerFirstName == person.first_name").toBe(pfirst);
    expect(soi.providerLastName, "providerLastName == person.last_name").toBe(plast);
    // The remaining three are put RAW inside the same guard, so a person with
    // no phone omits just that key rather than emitting "". Whichever way this
    // deployment's person row falls, the response has to agree with the column.
    for (const [key, col] of [
      ["providerWorkPhone", pphone],
      ["providerEmail", pemail],
      ["providerFax", pfax],
    ] as const) {
      if (col === null || col === "") {
        expect(key in soi, `${key} is absent when the column is empty`).toBe(false);
      } else {
        expect(soi[key], `${key} == its person column`).toBe(col);
      }
    }

    // ── referring site / department: sample_requester, split by ORG TYPE ───
    // requester_type_id only separates organization from provider; which
    // organization is the "site" and which the "department" is decided by the
    // organization's TYPE name. A port that keys off requester_type_id gets the
    // two swapped, and this oracle catches it.
    const orgOf = (typeName: string) =>
      query(
        `SELECT o.id, o.name, o.short_name
           FROM clinlims.sample s
           JOIN clinlims.sample_requester sr ON sr.sample_id = s.id
           JOIN clinlims.organization o ON o.id = sr.requester_id
           JOIN clinlims.organization_organization_type oot ON oot.org_id = o.id
           JOIN clinlims.organization_type t ON t.id = oot.org_type_id
          WHERE s.accession_number = 'E2E-FULL-01' AND t.short_name = '${typeName}'
          LIMIT 1`,
      );
    const site = orgOf("referring clinic");
    expect(site.length, "E2E-FULL-01 has a referring-clinic requester").toBe(1);
    expect(soi.referringSiteId, "referringSiteId == the referring-clinic org").toBe(String(site[0][0]));
    expect(soi.referringSiteName, "referringSiteName == organization.name").toBe(site[0][1]);
    expect(soi.referringSiteCode, "referringSiteCode == organization.short_name").toBe(site[0][2]);

    const dept = orgOf("dept");
    if (dept.length > 0) {
      expect(soi.referringSiteDepartmentId, "departmentId == the dept-typed org").toBe(String(dept[0][0]));
      expect(soi.referringSiteDepartmentName, "departmentName == organization.name").toBe(dept[0][1]);
    } else {
      // Java PROMOTES a lone department into the site slot, so the absence of a
      // dept-typed org must leave both department keys off rather than emitting
      // empties.
      expect("referringSiteDepartmentId" in soi, "no dept org -> no departmentId key").toBe(false);
      expect("referringSiteDepartmentName" in soi, "no dept org -> no departmentName key").toBe(false);
    }

    // ── program: NAME from observation history, ID from program_sample ─────
    // Two different tables feed one logical field. A port that read the name
    // off program.name would still look right until the two disagree.
    const progId = query(
      `SELECT ps.program_id FROM clinlims.sample s
         JOIN clinlims.program_sample ps ON ps.sample_id = s.id
        WHERE s.accession_number = 'E2E-FULL-01'`,
    );
    expect(progId.length, "E2E-FULL-01 has a program_sample row").toBe(1);
    expect(soi.programId, "programId comes from program_sample").toBe(String(progId[0][0]));

    // ── observation history: one row per key, matched by TYPE NAME ─────────
    const obs = new Map(
      query(
        `SELECT t.type_name, oh.value
           FROM clinlims.sample s
           JOIN clinlims.observation_history oh ON oh.sample_id = s.id
           JOIN clinlims.observation_history_type t ON t.id = oh.observation_history_type_id
          WHERE s.accession_number = 'E2E-FULL-01'`,
      ).map((r) => [r[0], r[1]]),
    );
    expect(obs.size, "E2E-FULL-01 carries observation history").toBeGreaterThan(0);
    for (const [key, typeName] of [
      ["program", "program"],
      ["paymentOptionSelection", "paymentStatus"],
      ["billingReferenceNumber", "billingRefNumber"],
      ["testLocationCode", "testLocationCode"],
      ["otherLocationCode", "testLocationCodeOther"],
      ["requestDate", "requestDate"],
      ["nextVisitDate", "nextVisitDate"],
      ["provisionalClinicalDiagnosis", "provisionalClinicalDiagnosis"],
    ] as const) {
      if (obs.has(typeName)) {
        expect(soi[key], `${key} == observation ${typeName}`).toBe(obs.get(typeName));
      } else {
        // observation_history_type is deployment data: this database has no
        // testLocationCode row at all, so the key can never appear. Pinning the
        // absence stops a port from inventing a default.
        expect(key in soi, `${key} is absent when ${typeName} is not a known observation type`).toBe(false);
      }
    }

    // ── priority: the RAW enum name, upper case ────────────────────────────
    // order/dashboard lower-cases the same column. Two endpoints, two casings,
    // and only a seeded non-null priority shows it.
    const pri = query(`SELECT order_priority FROM clinlims.sample WHERE accession_number = 'E2E-FULL-01'`)[0][0];
    expect(soi.priority, "priority is the raw enum name, not lower-cased").toBe(pri);

    // ── INVERSION ─────────────────────────────────────────────────────────
    // Everything above would also pass against a port that emitted these keys
    // unconditionally with hardcoded values. E2E001 has none of the underlying
    // rows, so every conditional key must be ABSENT there — that is what makes
    // the assertions above evidence of a conditional build rather than of a
    // constant.
    const lean = await readJson(await request.get(`${ORDER_SEARCH}?labNumber=E2E001`), `${ORDER_SEARCH} E2E001`);
    for (const key of [
      "providerPersonId",
      "providerFirstName",
      "providerLastName",
      "referringSiteId",
      "referringSiteName",
      "referringSiteCode",
      "programId",
      "program",
      "paymentOptionSelection",
      "billingReferenceNumber",
      "requestDate",
      "nextVisitDate",
      "provisionalClinicalDiagnosis",
    ]) {
      expect(key in lean.sampleOrderItems, `${key} is absent on an order with no such row`).toBe(false);
      expect(key in soi, `${key} IS present on the fully-populated order`).toBe(true);
    }

    // `priority` is deliberately NOT in that list. sample.order_priority is
    // nullable but carries DEFAULT 'ROUTINE', so an ordinary insert always
    // stores a value and the key is always emitted — an earlier draft of this
    // test asserted it absent on E2E001 and failed for exactly that reason.
    // E2E-FULL-03 is seeded with an EXPLICIT NULL, which is the only way to
    // reach Java's `if (sample.getPriority() != null)` guard.
    expect(lean.sampleOrderItems.priority, "a defaulted order still emits priority").toBe("ROUTINE");
    const nulled = await readJson(
      await request.get(`${ORDER_SEARCH}?labNumber=E2E-FULL-03`),
      `${ORDER_SEARCH} E2E-FULL-03`,
    );
    expect(
      query(`SELECT order_priority IS NULL FROM clinlims.sample WHERE accession_number = 'E2E-FULL-03'`)[0][0],
      "E2E-FULL-03 really has a NULL order_priority",
    ).toBe("t");
    expect("priority" in nulled.sampleOrderItems, "an explicitly NULL priority drops the key").toBe(false);
  });

  test("order/search: program resolution takes three different paths", async ({ request }) => {
    // The program keys are the one place where reading the obvious column is
    // wrong. ProgramSampleDAOImpl.getProgrammeSampleBySample picks an ENTITY
    // CLASS from the program NAME, and ProgramSample is
    // @Inheritance(TABLE_PER_CLASS) — each subclass has its OWN table. So:
    //
    //   (a) a plain name  -> queries program_sample, finds the row,
    //                        programId = program_sample.program_id
    //   (b) a name containing pathology/cytology/immunohistochemistry
    //                     -> queries the SUBCLASS table, finds nothing, and
    //                        falls back to matching the name against the
    //                        program list. program_sample is ignored entirely.
    //   (c) no program observation at all
    //                     -> a different query supplies BOTH keys, and
    //                        `program` is the program's own NAME.
    //
    // E2E-FULL-02 exists to separate (b) from (a): its observation names one
    // program while its program_sample row points at another, so the two
    // branches give different ids and a port cannot satisfy both by accident.
    const psProgram = (accession: string) =>
      query(
        `SELECT ps.program_id, p.name
           FROM clinlims.sample s
           JOIN clinlims.program_sample ps ON ps.sample_id = s.id
           JOIN clinlims.program p ON p.id = ps.program_id
          WHERE s.accession_number = '${accession}'`,
      );
    const obsProgram = (accession: string) =>
      query(
        `SELECT oh.value
           FROM clinlims.sample s
           JOIN clinlims.observation_history oh ON oh.sample_id = s.id
           JOIN clinlims.observation_history_type t ON t.id = oh.observation_history_type_id
          WHERE s.accession_number = '${accession}' AND t.type_name = 'program'`,
      );

    // ── (a) ───────────────────────────────────────────────────────────────
    const a = await readJson(
      await request.get(`${ORDER_SEARCH}?labNumber=E2E-FULL-01`),
      `${ORDER_SEARCH} E2E-FULL-01`,
    );
    const aObs = obsProgram("E2E-FULL-01");
    const aPs = psProgram("E2E-FULL-01");
    expect(aObs.length, "E2E-FULL-01 has a program observation").toBe(1);
    expect(aPs.length, "E2E-FULL-01 has a program_sample row").toBe(1);
    expect(a.sampleOrderItems.program, "(a) program is the OBSERVATION value").toBe(aObs[0][0]);
    expect(a.sampleOrderItems.programId, "(a) programId comes from program_sample").toBe(aPs[0][0]);

    // ── (b) ───────────────────────────────────────────────────────────────
    const b = await readJson(
      await request.get(`${ORDER_SEARCH}?labNumber=E2E-FULL-02`),
      `${ORDER_SEARCH} E2E-FULL-02`,
    );
    const bObs = obsProgram("E2E-FULL-02")[0][0];
    const bPs = psProgram("E2E-FULL-02");
    expect(
      /pathology|cytology|immunohistochemistry/i.test(bObs),
      "E2E-FULL-02 names a subclass-routed program",
    ).toBe(true);
    const namedId = query(`SELECT id FROM clinlims.program WHERE name = '${bObs}'`)[0][0];
    expect(
      bPs[0][0],
      "the fixture must point program_sample somewhere ELSE, or (b) is indistinguishable from (a)",
    ).not.toBe(namedId);
    expect(b.sampleOrderItems.program, "(b) program is still the observation value").toBe(bObs);
    expect(
      b.sampleOrderItems.programId,
      "(b) programId is the NAMED program, NOT the one program_sample points at",
    ).toBe(namedId);

    // ── (c) ───────────────────────────────────────────────────────────────
    const c = await readJson(
      await request.get(`${ORDER_SEARCH}?labNumber=E2E-FULL-03`),
      `${ORDER_SEARCH} E2E-FULL-03`,
    );
    expect(obsProgram("E2E-FULL-03").length, "E2E-FULL-03 has NO program observation").toBe(0);
    const cPs = psProgram("E2E-FULL-03");
    expect(cPs.length, "E2E-FULL-03 has a program_sample row").toBe(1);
    expect(
      c.sampleOrderItems.program,
      "(c) program falls back to the program table NAME, with no observation to read",
    ).toBe(cPs[0][1]);
    expect(c.sampleOrderItems.programId, "(c) programId comes from program_sample").toBe(cPs[0][0]);
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

    // ...and in the DB's PHYSICAL order, which is neither id nor sortOrder.
    //
    // The assertion above sorts both sides, so it says nothing about order —
    // and order is observable here, since samples[] is an array the frontend
    // renders in sequence. SampleItemDAOImpl has an HQL
    // getSampleItemsBySampleId ending `order by sampleItem.sortOrder`, but the
    // SERVICE method of the same name does NOT call it: it builds a criteria
    // map and calls getAllMatching, which has no ordering at all. The
    // controller calls the service, so Postgres decides, and what it returns
    // is scan order.
    //
    // This is a real discriminator on the stock dataset, not a theoretical one:
    // E2E001's items come back 10002 first even though 10002 has the HIGHER id
    // and the LATER sortOrder. A port that added `ORDER BY id` or
    // `ORDER BY sort_order` — the two obvious guesses — reverses them.
    const physicalOrder = query(
      `SELECT id FROM clinlims.sample_item WHERE samp_id = ${sampleId} AND voided = false ORDER BY ctid`,
    ).map((r) => r[0]);
    expect(
      body.samples.map((s: any) => s.id),
      `${ORDER_SEARCH} samples keep the DB's physical order, not id or sortOrder order`,
    ).toEqual(physicalOrder);

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

      // quantity is NOT a mirror — it is the sharpest null-policy split in the
      // envelope, and it took a fixture with a NULL quantity to expose it.
      // sample_item.quantity is a Double put at two sites:
      //   outer   put("quantity", q != null ? q : "")   -> number, or the STRING ""
      //   nested  put("quantity", q)                    -> number, or DROPPED by NON_NULL
      // So one column yields three shapes. An earlier version of this test
      // asserted the two were equal; it passed only because every sample_item
      // in the stock dataset happens to carry a quantity.
      const q = query(`SELECT quantity FROM clinlims.sample_item WHERE id = ${Number(item.id)}`)[0][0];
      if (q === null) {
        expect(item.quantity, "outer quantity coalesces NULL to the empty string").toBe("");
        expect("quantity" in item.sampleXML, "sampleXML drops a null quantity entirely").toBe(false);
      } else {
        expect(item.quantity, "outer quantity is the numeric column value").toBe(Number(q));
        expect(item.sampleXML.quantity, "sampleXML repeats a non-null quantity").toBe(item.quantity);
      }

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

  test("order/dashboard: stepProgress ignores VOIDED sample items", async ({ request }) => {
    // Every count behind stepProgress stands in for a pass over
    // sampleItemService.getSampleItemsBySampleId, whose criteria map is
    // {sample.id, voided:false} — so a voided item contributes to neither
    // collect nor label.
    //
    // E2E-VOIDED-01 is the only order in the dataset where that matters: its
    // ONE item carrying analyses is the voided one, and that item has a
    // collection date. Counting it makes collect look complete. Java reports
    // false; a port without the filter reports true, and no other assertion on
    // this endpoint looks at stepProgress per row.
    const items = query(
      `SELECT si.voided,
              (SELECT count(*) FROM clinlims.analysis a WHERE a.sampitem_id = si.id),
              (si.collection_date IS NOT NULL)
         FROM clinlims.sample_item si
         JOIN clinlims.sample s ON s.id = si.samp_id
        WHERE s.accession_number = 'E2E-VOIDED-01'
        ORDER BY si.id`,
    );
    const voidedWithTests = items.filter((r) => r[0] === "t" && Number(r[1]) > 0 && r[2] === "t");
    const liveWithTests = items.filter((r) => r[0] === "f" && Number(r[1]) > 0);
    expect(voidedWithTests.length, "E2E-VOIDED-01 has a voided, dated item WITH analyses").toBeGreaterThan(0);
    expect(liveWithTests.length, "...and no LIVE item with analyses, so collect must be false").toBe(0);

    const body = await readJson(await request.get(ORDER_DASHBOARD), ORDER_DASHBOARD);
    const row = body.orders.find((o: any) => o.labNumber === "E2E-VOIDED-01");
    expect(row, "E2E-VOIDED-01 is on the dashboard page").toBeTruthy();
    expect(
      row.stepProgress.collect,
      "collect is false: the only item with analyses is voided, so no item counts",
    ).toBe(false);
    // label is computed over the same voided-filtered list, so it must not be
    // rescued by the voided item's storage assignment either.
    expect(typeof row.stepProgress.label, "label is still a boolean").toBe("boolean");
  });

  test("unassigned-sample: rows keep the DB's physical order", async ({ request }) => {
    // getUnassignedReferrals is `FROM Referral r WHERE ...` with NO ordering, so
    // the array order is whatever Postgres scans — and the shipment dashboard
    // renders the array in sequence, which makes it observable.
    //
    // Every other assertion on this endpoint matches rows by id, so a port that
    // returned the same rows in a different order passed. The Go query adds five
    // JOINs that Java resolves lazily per row, and the planner reordered them:
    // Java led with E2E-REF-01, the port with E2E-REF-03.
    const rows = await readJson(await request.get(UNASSIGNED), UNASSIGNED);
    expect(rows.length, "the referral fixture seeds unassigned rows").toBeGreaterThan(1);

    const physicalOrder = query(
      `SELECT r.id FROM clinlims.referral r
        WHERE r.assigned_to_box_id IS NULL
          AND (r.lost_status IS NULL OR r.lost_status = false)
          AND r.status IS NOT NULL AND r.status <> 'CANCELED'
        ORDER BY r.ctid`,
    ).map((r) => r[0]);
    expect(
      rows.map((r: any) => String(r.id)),
      `${UNASSIGNED} rows are in scan order, not id or date order`,
    ).toEqual(physicalOrder);

    // by-facility runs the same predicate with an organization filter, so it
    // inherits the same ordering. Checked on the facility with more than one
    // row, since a single-row list cannot show an ordering bug.
    const byFacility = query(
      `SELECT r.organization_id, count(*) FROM clinlims.referral r
        WHERE r.organization_id IS NOT NULL
          AND r.assigned_to_box_id IS NULL
          AND (r.lost_status IS NULL OR r.lost_status = false)
          AND r.status IS NOT NULL AND r.status <> 'CANCELED'
        GROUP BY 1 HAVING count(*) > 1 ORDER BY 2 DESC LIMIT 1`,
    );
    expect(byFacility.length, "one facility has more than one unassigned referral").toBe(1);
    const facilityId = byFacility[0][0];
    const facilityRows = await readJson(
      await request.get(`${UNASSIGNED}/by-facility/${facilityId}`),
      `${UNASSIGNED}/by-facility`,
    );
    expect(
      facilityRows.map((r: any) => String(r.id)),
      `${UNASSIGNED}/by-facility keeps scan order too`,
    ).toEqual(
      query(
        `SELECT r.id FROM clinlims.referral r
          WHERE r.organization_id = ${Number(facilityId)}
            AND r.assigned_to_box_id IS NULL
            AND (r.lost_status IS NULL OR r.lost_status = false)
            AND r.status IS NOT NULL AND r.status <> 'CANCELED'
          ORDER BY r.ctid`,
      ).map((r) => r[0]),
    );
  });

test("order/dashboard: binding failures are 400, not a silent default", async ({ request }) => {
    // page and pageSize bind as int, includeExternal as boolean, so Spring
    // rejects an unconvertible value BEFORE the handler runs. The port
    // originally used a "parse, else use the default" helper and answered 200
    // here — a divergence its own comment acknowledged and declined to fix.
    for (const q of ["page=abc", "pageSize=abc", "includeExternal=abc"]) {
      const res = await request.get(`${ORDER_DASHBOARD}?${q}`);
      expect(res.status(), `${ORDER_DASHBOARD}?${q} is a bind failure`).toBe(400);
      const body = await res.json();
      expectKeysWithin(
        body,
        ["type", "title", "status", "detail", "instance"],
        ["type", "title", "status", "detail", "instance"],
        `${ORDER_DASHBOARD} 400 ProblemDetail`,
      );
      expect(body.status, "ProblemDetail.status").toBe(400);
      // type/title name MethodArgumentTypeMismatchException while detail names
      // org.springframework.beans.TypeMismatchException — two different classes
      // in one body. Pinned because it looks like a transcription slip and is
      // not.
      expect(body.type, "type names the method-argument exception").toBe(
        "problemDetail.type.org.springframework.web.method.annotation" +
          ".MethodArgumentTypeMismatchException",
      );
      expect(body.detail, "detail names the beans exception instead").toBe(
        "problemDetail.org.springframework.beans.TypeMismatchException",
      );
      expect(body.instance, "instance is the request path").toContain("/rest/order/dashboard");
    }

    // An EMPTY value is not a bind failure — the default applies.
    for (const q of ["page=", "pageSize="]) {
      const res = await request.get(`${ORDER_DASHBOARD}?${q}`);
      expect(res.status(), `${ORDER_DASHBOARD}?${q} falls back to the default`).toBe(200);
    }

    // String-typed params bind to anything, so garbage in them is a 200. The
    // contrast is the point: only the typed params can fail.
    for (const q of ["startDate=abc", "priority=abc", "status=abc", "search=abc"]) {
      const res = await request.get(`${ORDER_DASHBOARD}?${q}`);
      expect(res.status(), `${ORDER_DASHBOARD}?${q} is a String param, so 200`).toBe(200);
    }

    // includeExternal follows Spring's StringToBooleanConverter vocabulary,
    // which is NOT Go's: "t" and "f" are rejected even though strconv.ParseBool
    // accepts them.
    for (const v of ["on", "off", "yes", "no", "1", "0", "TRUE"]) {
      const res = await request.get(`${ORDER_DASHBOARD}?includeExternal=${v}`);
      expect(res.status(), `includeExternal=${v} is in Spring's vocabulary`).toBe(200);
    }
    for (const v of ["t", "f"]) {
      const res = await request.get(`${ORDER_DASHBOARD}?includeExternal=${v}`);
      expect(res.status(), `includeExternal=${v} is Go shorthand, not Spring's`).toBe(400);
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

  test("unassigned-sample: dashboard rows, one per branch of compileSampleData", async ({ request }) => {
    // Was `expect(Array.isArray(body)).toBe(true)` and nothing else, because
    // clinlims.referral is empty in the stock dataset — a check that passes on
    // [] forever, so the row shape was never compared against Java at all.
    // shipment-attachment-e2e.sql now seeds ten referrals: five visible here,
    // five excluded, each exercising exactly one branch.
    const rows = await readJson(await request.get(UNASSIGNED), UNASSIGNED);
    expect(Array.isArray(rows), `${UNASSIGNED} is an array`).toBe(true);

    const byId = new Map<string, any>(rows.map((r: any) => [r.id, r]));
    const fixture = query(
      "SELECT r.id, COALESCE(r.organization_id::text,''), COALESCE(r.organization_name,'')," +
        " COALESCE(r.priority,''), COALESCE(r.referral_reason_id::text,''), r.lost_status," +
        " COALESCE(r.status,''), COALESCE(r.assigned_to_box_id::text,''), si.voided, s.accession_number" +
        " FROM clinlims.referral r" +
        " JOIN clinlims.analysis a ON a.id = r.analysis_id" +
        " JOIN clinlims.sample_item si ON si.id = a.sampitem_id" +
        " JOIN clinlims.sample s ON s.id = si.samp_id" +
        " WHERE s.accession_number LIKE 'E2E-REF-%' ORDER BY r.id::numeric",
    );
    // Not a skip: the loader marks the fixture fatal, so an empty result means
    // it did not run and this test would otherwise silently assert nothing.
    expect(fixture.length, "shipment-attachment-e2e.sql is loaded").toBe(10);

    // ── the five exclusion rules, one row each ────────────────────────────
    // A port implementing four of the five still returns a plausible list, so
    // each is named and checked separately rather than as one count.
    for (const [id, , , , , lost, status, boxId, voided] of fixture) {
      const present = byId.has(id);
      if (lost === "t") {
        expect(present, `lost referral ${id} is excluded`).toBe(false);
      } else if (status === "CANCELED") {
        expect(present, `canceled referral ${id} is excluded`).toBe(false);
      } else if (boxId !== "") {
        expect(present, `box-assigned referral ${id} is excluded`).toBe(false);
      } else if (status === "") {
        // The HQL is `r.status != 'CANCELED'`, and NULL != 'CANCELED' is
        // UNKNOWN, not TRUE — so a NULL status is excluded by three-valued
        // logic rather than by any explicit rule. A port written with Go's
        // `status != "CANCELED"` INCLUDES this row.
        expect(present, `NULL-status referral ${id} is excluded by 3-valued logic`).toBe(false);
      } else {
        // Voided sample items are NOT filtered here. getUnassignedReferrals
        // has no `si.voided` predicate, unlike the /items query — so this row
        // is present here and absent there. A port that shares one filter
        // between the two endpoints gets exactly one of them wrong.
        expect(present, `referral ${id} is present (voided=${voided} does not exclude here)`).toBe(true);
      }
    }

    // ── row shape, per branch ─────────────────────────────────────────────
    for (const [id, orgId, orgName, priority, reasonId] of fixture) {
      const row = byId.get(id);
      if (!row) continue;

      // The row is a HashMap, so Jackson's NON_NULL drops null values: which
      // keys are PRESENT is itself the assertion.
      expectKeysWithin(
        row,
        ["id", "referralDate", "priority", "daysUnassigned", "accessionNumber", "sampleId",
         "referralTestName", "testId", "destinationFacilityName", "destinationFacilityId",
         "referralReasonId"],
        ["id", "priority", "daysUnassigned", "accessionNumber", "sampleId",
         "referralTestName", "testId", "destinationFacilityName"],
        `${UNASSIGNED} row ${id}`,
      );

      if (orgId !== "") {
        expect(row.destinationFacilityId, `row ${id} destinationFacilityId`).toBe(orgId);
      } else {
        // The ELSE branch emits the free-text name WITHOUT an id. Pinning the
        // ABSENCE matters: a port that emits `destinationFacilityId: null`
        // or "" adds a key Java never sends.
        expect("destinationFacilityId" in row, `row ${id} omits destinationFacilityId entirely`).toBe(false);
        expect(row.destinationFacilityName, `row ${id} falls back to organization_name`).toBe(orgName);
      }

      // priority is `getPriority() != null ? getPriority() : "Normal"` — the
      // literal string, not a lookup and not the DB value uppercased.
      expect(row.priority, `row ${id} priority`).toBe(priority === "" ? "Normal" : priority);

      if (reasonId === "") {
        expect("referralReasonId" in row, `row ${id} omits a null referralReasonId`).toBe(false);
      } else {
        expect(row.referralReasonId, `row ${id} referralReasonId`).toBe(reasonId);
      }
    }

    // ── daysUnassigned is COMPUTED, not stored ────────────────────────────
    // TimeUnit.MILLISECONDS.toDays(now - requestDate), i.e. truncated toward
    // zero, and 0L when there is no request date. Derived from the row's own
    // referralDate so this does not rot as the fixture ages.
    for (const row of rows) {
      if (typeof row.referralDate === "number") {
        const expected = Math.floor((Date.now() - row.referralDate) / 86_400_000);
        // Tolerance 1: the server and the test read their own clocks, so a
        // run that straddles a day boundary can legitimately differ by one.
        expect(
          Math.abs(row.daysUnassigned - expected),
          `daysUnassigned for ${row.id} tracks referralDate`,
        ).toBeLessThanOrEqual(1);
      } else {
        // No request date: BOTH the referralDate key disappears (NON_NULL) and
        // daysUnassigned falls back to the literal 0.
        expect(row.daysUnassigned, `daysUnassigned is 0 without a referralDate`).toBe(0);
      }
    }
  });

  test("unassigned-sample/items: 500 once ANY row matches — a Java defect the empty table hid", async ({
    request,
  }) => {
    // MIGRATION POLICY: pinned, not fixed.
    //
    // UnassignedSampleItemServiceImpl.buildSampleItemDTOs calls
    // referralDAO.getReferralsBySampleItemId(Integer), but SampleItem.id is
    // mapped as a String, so Hibernate rejects the binding:
    //   Parameter value [10111] did not match expected type [java.lang.String]
    // The service catches it and returns an empty list, but the exception
    // already marked the read-only transaction rollback-only, so the commit at
    // the @Transactional boundary throws and the CONTROLLER's catch turns it
    // into `ResponseEntity.status(500).build()` — a 500 with an EMPTY body.
    //
    // This is unconditional given at least one matching row: the mismatch is
    // between an Integer argument and a String-mapped id, not between values.
    // With clinlims.referral empty the loop never ran, so both endpoints
    // returned 200 [] and looked healthy. That is precisely why seeding the
    // table is part of the work rather than a follow-up.
    const candidates = Number(
      query(
        "SELECT count(*) FROM clinlims.referral r" +
          " JOIN clinlims.analysis a ON a.id = r.analysis_id" +
          " JOIN clinlims.sample_item si ON si.id = a.sampitem_id" +
          " WHERE r.assigned_to_box_id IS NULL" +
          "   AND (r.lost_status IS NULL OR r.lost_status = false)" +
          "   AND r.status IS NOT NULL AND r.status <> 'CANCELED'" +
          "   AND (si.rejected IS NULL OR si.rejected = false)" +
          "   AND (si.voided IS NULL OR si.voided = false)",
      )[0][0],
    );
    expect(candidates, "the fixture leaves at least one row for /items to choke on").toBeGreaterThan(0);

    const items = await request.get(`${UNASSIGNED}/items`);
    expect(items.status(), `${UNASSIGNED}/items is 500 whenever a row matches`).toBe(500);
    expect(await items.text(), `${UNASSIGNED}/items 500 carries no body`).toBe("");

    const searched = await request.get(`${UNASSIGNED}/items/search?accessionNumber=E2E-REF`);
    expect(searched.status(), `${UNASSIGNED}/items/search is 500 the same way`).toBe(500);
    expect(await searched.text(), `${UNASSIGNED}/items/search 500 carries no body`).toBe("");

    // A search term matching nothing takes the other path: the loop never
    // runs, no binding happens, and the endpoint answers 200 [].
    const empty = await request.get(`${UNASSIGNED}/items/search?accessionNumber=NO-SUCH-ACCESSION-XYZ`);
    expect(empty.status(), "a search matching nothing is 200").toBe(200);
    expect(await empty.json(), "a search matching nothing returns []").toEqual([]);

    // accessionNumber is a required @RequestParam here (no defaultValue), so
    // omitting it is a 400 — unlike /items, which takes none.
    const missing = await request.get(`${UNASSIGNED}/items/search`);
    expect(missing.status(), `${UNASSIGNED}/items/search without accessionNumber`).toBe(400);
  });

  test("unassigned-sample/count-by-facility: agrees with by-facility, and both really filter", async ({
    request,
  }) => {
    // Two organizations are seeded with different numbers of referrals, so
    // this is a real filter rather than "everything, twice".
    const orgs = query(
      "SELECT r.organization_id, count(*) FROM clinlims.referral r" +
        " WHERE r.organization_id IS NOT NULL" +
        "   AND r.assigned_to_box_id IS NULL" +
        "   AND (r.lost_status IS NULL OR r.lost_status = false)" +
        "   AND r.status IS NOT NULL AND r.status <> 'CANCELED'" +
        " GROUP BY r.organization_id ORDER BY count(*) DESC",
    );
    expect(orgs.length, "the fixture seeds referrals across at least two facilities").toBeGreaterThan(1);

    let previous = -1;
    for (const [facilityId, dbCount] of orgs) {
      const counted = await readJson(
        await request.get(`${UNASSIGNED}/count-by-facility/${facilityId}`),
        `${UNASSIGNED}/count-by-facility`,
      );
      expect(Object.keys(counted), `${UNASSIGNED}/count-by-facility envelope`).toEqual(["count"]);

      const listed = await readJson(
        await request.get(`${UNASSIGNED}/by-facility/${facilityId}`),
        `${UNASSIGNED}/by-facility`,
      );
      expect(Array.isArray(listed), `${UNASSIGNED}/by-facility is an array`).toBe(true);

      // countUnassignedSamplesByFacility re-applies a lost/canceled filter
      // that getUnassignedSamplesByDestinationFacility does not — but the SQL
      // has already excluded both, so the extra pass removes nothing and the
      // two agree. Measured, not assumed: an earlier draft asserted only
      // `count <= length`, which is satisfied by returning 0 forever.
      expect(counted.count, `count-by-facility(${facilityId}) equals by-facility length`).toBe(listed.length);
      expect(counted.count, `count-by-facility(${facilityId}) matches the DB`).toBe(Number(dbCount));

      // Every listed row really belongs to the requested facility.
      for (const row of listed) {
        expect(row.destinationFacilityId, `by-facility(${facilityId}) row ${row.id}`).toBe(facilityId);
      }

      // The two facilities must not return the same number, or "filtering"
      // and "not filtering" would be indistinguishable.
      if (previous >= 0) {
        expect(counted.count, "the two facilities differ, so the filter is observable").not.toBe(previous);
      }
      previous = counted.count;
    }
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

  test("getPendingAnalysisForTestProvider: the four groups, oracled per status", async ({ request }) => {
    // The only assertion this endpoint had was the 400 for a missing param, so
    // its entire payload — four status-grouped lists — was unverified. Under
    // the field-by-field diff it also turned out to be the one c2 response
    // whose array order is not reproducible; see the ordering note below.
    //
    // Each group is one status, resolved by NAME because the numeric ids are
    // deployment data:
    // Each group is one status. The enum constant is NOT the stored name —
    // StatusService.addToAnalysisMap maps them by matching status_of_sample.name
    // literally, so AnalysisStatus.NotStarted is the row named "Not Tested" and
    // BiologistRejected is "Biologist Rejection" (not "Biologist Rejected").
    // Resolved by that stored NAME here, because the numeric ids are deployment
    // data.
    //   notStarted          NotStarted          -> "Not Tested"
    //   technicianRejection TechnicalRejected   -> "Technical Rejected"
    //   biologistRejection  BiologistRejected   -> "Biologist Rejection"
    //   notValidated        TechnicalAcceptance -> "Technical Acceptance"
    //                       ^ an ACCEPTANCE status under a "notValidated" key
    const groups = [
      ["notStarted", "Not Tested"],
      ["technicianRejection", "Technical Rejected"],
      ["biologistRejection", "Biologist Rejection"],
      ["notValidated", "Technical Acceptance"],
    ] as const;

    // A test id that actually has pending analyses, so the assertions below are
    // not all vacuously true on empty arrays.
    const testIds = query(
      `SELECT a.test_id, count(*) FROM clinlims.analysis a
        WHERE a.test_id IS NOT NULL GROUP BY 1 ORDER BY 2 DESC LIMIT 1`,
    );
    expect(testIds.length, "some test has analyses").toBe(1);
    const testId = testIds[0][0];

    const body = await readJson(await request.get(`${PENDING_ANALYSIS}?testId=${testId}`), PENDING_ANALYSIS);
    expect(Object.keys(body).sort(), `${PENDING_ANALYSIS} emits exactly the four groups`).toEqual(
      [...groups].map(([k]) => k).sort(),
    );

    let populated = 0;
    for (const [key, statusName] of groups) {
      const expected = query(
        `SELECT a.id, s.accession_number
           FROM clinlims.analysis a
           JOIN clinlims.sample_item si ON si.id = a.sampitem_id
           JOIN clinlims.sample s ON s.id = si.samp_id
           JOIN clinlims.status_of_sample sos ON sos.id = a.status_id
          WHERE a.test_id = ${Number(testId)}
            AND sos.status_type = 'ANALYSIS' AND sos.name = '${statusName}'
          ORDER BY s.accession_number`,
      );
      const rows = body[key];
      expect(Array.isArray(rows), `${key} is an array`).toBe(true);

      // MEMBERSHIP is the contract; order within it is not (see below).
      expect(
        rows.map((r: any) => r.id).sort(),
        `${key} contains exactly the ${statusName} analyses for this test`,
      ).toEqual(expected.map((r) => r[0]).sort());

      // Each entry carries the accession of the sample its analysis hangs off,
      // which is the join a port could get wrong while still returning the
      // right ids.
      const labNoById = new Map(expected.map((r) => [r[0], r[1]]));
      for (const r of rows) {
        expect(Object.keys(r).sort(), `${key} row shape`).toEqual(["id", "labNo"]);
        expect(r.labNo, `${key} row ${r.id} carries its own sample's accession`).toBe(labNoById.get(r.id));
      }

      // ORDERING — and specifically what is NOT asserted.
      //
      // The HQL ends `order by a.sampleItem.sample.accessionNumber` and nothing
      // else, so rows sharing an accession number are TIED and Postgres decides
      // their relative order from the query plan. That is not stable even
      // within one server: measured on this dataset, the two calls after a
      // fresh connection return a tie group in one order and later calls return
      // it in another, on Java and on the port alike.
      //
      // So the only ordering Java actually promises is the one it asked for,
      // and that is all this asserts. Do NOT "fix" a tie-order mismatch by
      // adding a secondary sort key to the port: Java has none, no available
      // column reproduces the observed tie order anyway, and inventing one
      // makes the port deterministic in a way the original is not.
      // The comparison uses POSTGRES as the oracle rather than sorting in JS.
      // `ORDER BY accession_number` sorts under the database collation, which on
      // this deployment ignores punctuation at the primary level — so E2E001
      // precedes E2E-EDIT-01, while a JS `<` (code-point order) puts the
      // hyphen first and gets the opposite answer. A port that sorted in Go
      // rather than in SQL would inherit the JS answer.
      //
      // Comparing the labNo SEQUENCE also sidesteps the tie problem: rows tied
      // on accession number share a labNo, so the sequence is the same whichever
      // order the plan puts them in.
      expect(
        rows.map((r: any) => r.labNo),
        `${key} is ordered by accession_number under the DB collation`,
      ).toEqual(expected.map((r) => r[1]));

      if (rows.length > 0) populated++;
    }
    expect(populated, "at least one group is non-empty, so the loop above proved something").toBeGreaterThan(0);
  });

  // ── rest/order/{accessionNumber}/attachments ────────────────────────────

  test("order/{accession}/attachments: populated rows, soft-delete filter and ordering", async ({
    request,
  }) => {
    // clinlims.order_attachment is empty in the stock dataset, so the 200 path
    // could only ever be observed as []. shipment-attachment-e2e.sql seeds
    // three rows on E2E-ATT-01: one typed, one with a NULL file_type, and one
    // soft-deleted.
    const seeded = query(
      "SELECT oa.id, oa.original_file_name, COALESCE(oa.file_type,''), oa.file_size_bytes," +
        " oa.is_deleted, to_char(oa.uploaded_at, 'YYYY-MM-DD HH24:MI:SS')" +
        " FROM clinlims.order_attachment oa" +
        " JOIN clinlims.sample s ON s.id = oa.sample_id" +
        " WHERE s.accession_number = 'E2E-ATT-01' ORDER BY oa.uploaded_at DESC",
    );
    // Not a skip: the loader marks the fixture fatal.
    expect(seeded.length, "shipment-attachment-e2e.sql seeds three attachments").toBe(3);
    const live = seeded.filter((r) => r[4] === "f");
    expect(live.length, "two of the three are active").toBe(2);

    const body = await readJson(
      await request.get(`rest/order/E2E-ATT-01/attachments`),
      "order attachments",
    );
    expect(Array.isArray(body), "order attachments is an array").toBe(true);

    // findActiveBySampleId is `where isDeleted = false order by uploadedAt desc`,
    // so BOTH the soft-delete filter and the ordering are asserted here rather
    // than the length alone.
    expect(
      body.map((a: any) => String(a.id)),
      "active attachments, newest upload first",
    ).toEqual(live.map((r) => r[0]));

    for (const [i, row] of body.entries()) {
      const [dbId, dbName, dbType, dbSize, , dbUploaded] = live[i];

      // toDto is a Map.of with exactly five entries — no `sampleId`, no
      // `isDeleted`, no `uploadedBy`. Those columns exist and must not leak.
      expectKeysWithin(
        row,
        ["id", "fileName", "fileType", "fileSizeBytes", "uploadedAt"],
        ["id", "fileName", "fileType", "fileSizeBytes", "uploadedAt"],
        "order attachment row",
      );
      // id and fileSizeBytes are NUMBERS here, unlike the String ids most of
      // this codebase emits — the column is a real integer and Map.of keeps it.
      expect(typeof row.id, "attachment id is a number").toBe("number");
      expect(String(row.id), "attachment id").toBe(dbId);
      expect(row.fileName, "fileName is original_file_name").toBe(dbName);
      expect(typeof row.fileSizeBytes, "fileSizeBytes is a number").toBe("number");
      expect(String(row.fileSizeBytes), "fileSizeBytes").toBe(dbSize);

      // A NULL file_type becomes "" HERE, while the download path turns the
      // same NULL into application/octet-stream. One column, two null
      // policies — a port that normalises them to one value breaks one caller.
      expect(row.fileType, "fileType is '' when the column is null").toBe(dbType);

      // uploadedAt is java.sql.Timestamp.toString(), which appends a
      // fractional-second part: "2025-05-04 10:00:00.0", not ISO-8601 and not
      // epoch millis.
      expect(row.uploadedAt, "uploadedAt is Timestamp.toString()").toBe(`${dbUploaded}.0`);
    }

    // A known order with zero attachments -> 200 []; an unknown order -> 404
    // with a JSON error envelope {"error":"Order not found"}. Two distinct
    // shapes for two distinct conditions, and the 404 body is JSON rather than
    // the empty body most other 404s in this codebase return — so a port must
    // emit the envelope, not just the status.
    const emptyOne = query(
      "SELECT s.accession_number FROM clinlims.sample s" +
        " WHERE NOT EXISTS (SELECT 1 FROM clinlims.order_attachment oa WHERE oa.sample_id = s.id)" +
        " ORDER BY s.id LIMIT 1",
    );
    if (emptyOne.length > 0) {
      const res = await request.get(`rest/order/${encodeURIComponent(emptyOne[0][0])}/attachments`);
      expect(res.status(), "a known order with no attachments is 200").toBe(200);
      expect(await res.json(), "a known order with no attachments returns []").toEqual([]);
    }

    const unknown = await request.get("rest/order/NO_SUCH_ACCESSION_XYZ/attachments");
    expect(unknown.status(), "order attachments unknown accession is 404").toBe(404);
    expect(await unknown.json(), "order attachments 404 carries a JSON error envelope").toEqual({
      error: "Order not found",
    });
  });

  test("order/attachments/{id}/download vs /view: same bytes, different disposition", async ({
    request,
  }) => {
    // Entirely unverifiable before the fixture existed: with no rows, only the
    // 404 branch was reachable, and 404 says nothing about the Content-Type /
    // Content-Disposition behaviour that is the whole point of these two.
    const rows = query(
      "SELECT oa.id, oa.original_file_name, COALESCE(oa.file_type,''), oa.file_size_bytes, oa.is_deleted" +
        " FROM clinlims.order_attachment oa" +
        " JOIN clinlims.sample s ON s.id = oa.sample_id" +
        " WHERE s.accession_number = 'E2E-ATT-01' ORDER BY oa.id::numeric",
    );
    expect(rows.length, "shipment-attachment-e2e.sql is loaded").toBe(3);

    for (const [id, fileName, fileType, size, isDeleted] of rows) {
      if (isDeleted === "t") {
        // serveAttachment refuses a soft-deleted row before it ever looks at
        // the bytes, on BOTH endpoints, with an empty body.
        for (const mode of ["download", "view"]) {
          const res = await request.get(`rest/order/attachments/${id}/${mode}`);
          expect(res.status(), `soft-deleted attachment ${id} ${mode} is 404`).toBe(404);
          expect(await res.text(), `soft-deleted ${mode} 404 carries no body`).toBe("");
        }
        continue;
      }

      // A NULL file_type falls back to application/octet-stream here — the
      // same column the list endpoint renders as "".
      const expectedType = fileType === "" ? "application/octet-stream" : fileType;

      for (const [mode, disposition] of [["download", "attachment"], ["view", "inline"]]) {
        const res = await request.get(`rest/order/attachments/${id}/${mode}`);
        expect(res.status(), `attachment ${id} ${mode} is 200`).toBe(200);

        // Spring appends `;charset=UTF-8` even to application/pdf and
        // application/octet-stream, because the media type is built through
        // MediaType.parseMediaType and written by a ResourceHttpMessageConverter
        // with the default charset attached. Pinned as-is: a port emitting a
        // bare `application/pdf` is a different response.
        expect(res.headers()["content-type"], `attachment ${id} ${mode} Content-Type`).toBe(
          `${expectedType};charset=UTF-8`,
        );
        expect(
          res.headers()["content-disposition"],
          `attachment ${id} ${mode} Content-Disposition`,
        ).toBe(`${disposition}; filename="${fileName}"`);
        expect(res.headers()["content-length"], `attachment ${id} ${mode} Content-Length`).toBe(size);

        // The bytes themselves, not just the headers.
        const buf = await res.body();
        expect(buf.length, `attachment ${id} ${mode} body length`).toBe(Number(size));
        const [[dbHex]] = query(
          `SELECT encode(file_content, 'hex') FROM clinlims.order_attachment WHERE id = ${id}`,
        );
        expect(buf.toString("hex"), `attachment ${id} ${mode} body bytes`).toBe(dbHex);
      }
    }

    // MIGRATION POLICY: pinned, not fixed. A MISSING id is a 500, not the 404
    // a soft-deleted one produces — OrderAttachmentServiceImpl.get throws
    // rather than returning null, so the `attachment == null` guard in
    // serveAttachment is unreachable for that case. Two "not there" conditions,
    // two different statuses.
    const missing = await request.get("rest/order/attachments/999999/download");
    expect(missing.status(), "a missing attachment id is 500, not 404").toBe(500);
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
//     - order/search: the whole envelope — nine top-level keys,
//       patientProperties' always-present identity fields, samples[] with the
//       sampleXML twin and the all-or-nothing storage block, and every
//       paymentOption {id,value} cross-checked against clinlims.dictionary.
//     - order/search: voided sample items are excluded (order-search-e2e.sql
//       seeds the dataset's only voided row — without it, deleting the
//       predicate from the port left the suite green).
//     - order/search: the FULL sampleOrderItems map on E2E-FULL-01 — provider
//       via person, referring site and department split by ORGANIZATION TYPE,
//       programId from program_sample, and every observation-history value by
//       type name, each oracled against its own row, with E2E001 as the
//       inversion proving all of them are conditional.
//     - order/search: quantity's three shapes (number / "" / absent) from the
//       one column, which only a NULL-quantity sample item can show.
//     - order/search: all THREE program-resolution paths, on E2E-FULL-01/02/03
//       — program_sample, the TABLE_PER_CLASS subclass miss that falls back to
//       a name lookup, and the no-observation branch that takes both keys from
//       program_sample. -02 points its program_sample at a different program
//       from the one it names, so the branches cannot agree by accident.
//     - order/search: priority survives a DEFAULT and disappears only on an
//       EXPLICIT NULL (E2E-FULL-03).
//     - order/search: the raw response BYTES — no trailing newline, and a
//       Double rendered as 5.0 rather than 5. Neither is visible to any
//       assertion on the DECODED body, and both were wrong.
//     - order/search: samples[] keeps the DB's physical order (E2E001 returns
//       item 10002 first, which is neither id nor sortOrder order).
//     - order/dashboard: stepProgress ignores VOIDED items — E2E-VOIDED-01's
//       only item with analyses is the voided one, so collect must be false.
//     - unassigned-sample + by-facility: rows keep scan order, checked against
//       ctid, on a facility with more than one row.
//     - getPendingAnalysisForTestProvider: all four status groups oracled by
//       stored status NAME, row shape, the labNo join, and ordering compared
//       against Postgres itself (the DB collation ignores punctuation, so
//       E2E001 precedes E2E-EDIT-01). The within-tie order is deliberately
//       NOT asserted — it is unstable on Java too.
//     - order/dashboard: every labNumber must be a real sample
//       accession_number, and the page is a strict subset of the table.
//     - order/dashboard quirks: pageSize-ignored / externalCount-0 /
//       includeExternal-inert, each proven by contrasting real responses.
//     - unassigned-by-accession: always-500 proven across three inputs.
//     - unassigned-sample: dashboard row shape per branch of compileSampleData,
//       all five exclusion rules one row each, and daysUnassigned derived from
//       the row's own referralDate.
//     - unassigned-sample/items + /items/search: the Java 500 (see below), the
//       200 [] a non-matching search still returns, and the 400 for a missing
//       param.
//     - count-by-facility EQUALS by-facility length, across two facilities with
//       different counts. Two earlier drafts were both wrong here: one asserted
//       equality when the tables were empty, the next weakened it to
//       `count <= length`, which returning 0 forever satisfies.
//     - order/{accession}/attachments: row shape, the soft-delete filter and
//       the uploadedAt-DESC ordering.
//     - order/attachments/{id}/download|view: Content-Type (including Spring's
//       appended charset), Content-Disposition, Content-Length and the BYTES,
//       plus 404 for soft-deleted and 500 for missing.
//
//   All of the above became verifiable only after the fixtures existed.
//   clinlims.referral and clinlims.order_attachment are empty in the stock
//   dataset; src/test/resources/fixtures/shipment-attachment-e2e.sql seeds
//   both, and order-search-e2e.sql seeds the voided sample item. Between them
//   they turned four "green" assertions into failures that were real port
//   defects.
//
// ── DELIBERATELY NOT COVERED (and why) ──────────────────────────────────────
//
// Exactly one exclusion remains, and it is not "the table is empty" — that is a
// fixture bug, not a scope boundary, and the ones this wave had are now seeded.
//
// MUTATING endpoints on the unassigned-sample controller —
// POST/PUT assignSampleToBox, markSampleAsLost, cancelReferral
// (UnassignedSampleRestController.java:126, :155, :184). They change referral
// state and must never run from the read-only suite. They belong to the
// shipment feature module (an h-* branch), not to c2.
//
// The Type-D form loads (rest/GenericSampleOrder, rest/SamplePatientEntry,
// rest/SampleEdit, rest/SampleBatchEntrySetup) used to be listed here as a
// second exclusion. They are now ported and pinned in
// c2-sample-form-loads.spec.ts, so c2 has exactly one exclusion, not three.
//
// ── JAVA DEFECTS THIS FILE PINS ─────────────────────────────────────────────
//
// Reproduced, never fixed, and listed in migration/java-defects-found.md so
// they can be raised with the maintainers rather than silently corrected here:
// unassigned-by-accession's invalid HQL, unassigned-sample/items' Integer-vs-
// String binding, order/dashboard's paging and hardcoded counters, and the
// download/view split where a missing id is a 500 but a deleted one is a 404.
