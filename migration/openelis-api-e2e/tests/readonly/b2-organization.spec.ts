// §4 — b2: organization + program reference reads (strict parity gate).
//
// Taxonomy Type B, Wave 2. Also runs against the Go port via the go-parity
// project (see playwright.config.ts) — except user-programs, which is
// session/RBAC-dependent and deliberately not implemented in Go yet (see
// migration/b2-org-provider-migration.md §2), so it's skipped there.
//
// Design (same as a2/b1): assert the real contract, not "HTTP 200". Where
// live testing found Go deliberately diverging from a confirmed Java bug or
// serialization quirk, both sides' real behavior is asserted explicitly
// (branched on testInfo.project.name), not silently pinned to one value —
// see migration/b2-org-provider-migration.md §3 for the full writeup of each:
//   - organization/{id} not-found: 404 (Go) vs 500 (Java, confirmed bug)
//   - organization-list/{id} organizationTypes: real array (Go) vs Java's
//     unconditional null (Hibernate lazy="true" + no forced load)
//
// Findings excluded here (real, out of scope): GET rest/Organization returns 500
// on Java itself; rest/organization/search is a paginated Type-C search (its own
// group); rest/EntityNamesProvider / getPendingAnalysisForTestProvider need query
// params (400 without).
import { test, expect } from "@playwright/test";
import {
  readJson,
  readJsonArray,
  expectExactKeys,
  expectKeysWithin,
  expectUnique,
  expectNonEmptyString,
} from "../../fixtures/assert";

const ORG_TYPES = "rest/organization/types";
const ORG_LIST = "rest/organization-list";
const ORG_BY_ID = "rest/organization";
const GENERATE_SITE_CODE = "rest/organization/generate-site-code";
const DEPARTMENTS_FOR_SITE = "rest/departments-for-site";
const USER_PROGRAMS = "rest/user-programs";

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
// Large enough that it will never collide with a real seeded/test id.
const MISSING_ID = "999999999";

test.describe("b2 — organization + program reference reads", () => {
  test("organization/types: {id,name,description} rows, unique ids", async ({ request }) => {
    const body = await readJsonArray(await request.get(ORG_TYPES), ORG_TYPES);
    for (const row of body) {
      // lastupdated is NON_NULL-omitted on some rows → optional.
      expectKeysWithin(
        row,
        ["lastupdated", "id", "name", "description"],
        ["id", "name", "description"],
        `${ORG_TYPES} row`,
      );
      expectNonEmptyString(row.id, `${ORG_TYPES} id`);
      expectNonEmptyString(row.name, `${ORG_TYPES} name`);
      expect(typeof row.description, `${ORG_TYPES} description`).toBe("string");
      if ("lastupdated" in row)
        expect(typeof row.lastupdated, `${ORG_TYPES} lastupdated`).toBe("number");
    }
    expectUnique(body.map((r: any) => r.id), `${ORG_TYPES} id`);
    // order: getAll(), no ORDER BY → DB-natural, NOT asserted.
  });

  test("organization-list: organization rows with valid fhirUuid + Y/N active flag", async ({ request }) => {
    const body = await readJsonArray(await request.get(ORG_LIST), ORG_LIST);
    const allowed = [
      "lastupdated", "isActive", "id", "mlsLabFlag", "mlsSentinelLabFlag",
      "organizationName", "shortName", "organizationTypes", "fhirUuid", "testSections",
    ];
    for (const row of body) {
      // mlsLabFlag is NON_NULL-omitted on some rows; organizationTypes is present
      // but may be null. Require the always-present core fields.
      expectKeysWithin(
        row,
        allowed,
        ["id", "isActive", "organizationName", "shortName", "fhirUuid", "testSections", "organizationTypes"],
        `${ORG_LIST} row`,
      );
      expectNonEmptyString(row.id, `${ORG_LIST} id`);
      expect(["Y", "N"], `${ORG_LIST} isActive is a Y/N flag`).toContain(row.isActive);
      expectNonEmptyString(row.organizationName, `${ORG_LIST} organizationName`);
      expect(row.fhirUuid, `${ORG_LIST} fhirUuid is a UUID`).toMatch(UUID_RE);
      expect(Array.isArray(row.testSections), `${ORG_LIST} testSections is array`).toBe(true);
      expect(
        row.organizationTypes === null || Array.isArray(row.organizationTypes),
        `${ORG_LIST} organizationTypes is null or array`,
      ).toBe(true);
    }
    expectUnique(body.map((r: any) => r.id), `${ORG_LIST} id`);
    // order: not guaranteed by Java (not name-sorted) → NOT asserted.
  });

  test("organization/{id}: matches its organization-list row; not-found diverges by design", async ({
    request,
  }, testInfo) => {
    // Pick a real id from the live list rather than hardcoding one — resilient
    // to whichever dataset this runs against.
    const list = await readJsonArray(await request.get(ORG_LIST), ORG_LIST);
    const sample = list[0];
    const body = await readJson(await request.get(`${ORG_BY_ID}/${sample.id}`), `${ORG_BY_ID}/{id}`);

    // organization/{id} and organization-list serialize the same entity in
    // Java, so a given id's fields must agree between the two endpoints.
    expect(body.id, `${ORG_BY_ID}/{id} id`).toBe(sample.id);
    expect(body.organizationName, `${ORG_BY_ID}/{id} organizationName`).toBe(sample.organizationName);
    expect(body.fhirUuid, `${ORG_BY_ID}/{id} fhirUuid`).toBe(sample.fhirUuid);
    expect(body.shortName ?? null, `${ORG_BY_ID}/{id} shortName`).toBe(sample.shortName ?? null);
    expect(
      body.organizationTypes === null || Array.isArray(body.organizationTypes),
      `${ORG_BY_ID}/{id} organizationTypes is null or array`,
    ).toBe(true);

    // Not-found: real, confirmed, deliberate divergence — see
    // migration/b2-org-provider-migration.md §3.2 #1. Java's
    // BaseObjectServiceImpl.get(id) throws ObjectNotFoundException on a miss,
    // uncaught by the controller → 500 (live-confirmed, not theoretical).
    // This port returns a real 404 instead.
    const missing = await request.get(`${ORG_BY_ID}/${MISSING_ID}`);
    if (testInfo.project.name === "go-parity") {
      expect(missing.status(), `${ORG_BY_ID}/{id} missing (Go: real 404)`).toBe(404);
    } else {
      expect(missing.status(), `${ORG_BY_ID}/{id} missing (Java: confirmed 500 bug)`).toBe(500);
    }
  });

  test("organization/generate-site-code: S<UTC yyMMdd>-<5-digit seq>", async ({ request }) => {
    const body = await readJson(await request.get(GENERATE_SITE_CODE), GENERATE_SITE_CODE);
    expectExactKeys(body, ["siteCode"], `${GENERATE_SITE_CODE} body`);

    // Format per OrganizationServiceImpl.generateSiteCode(): "S" + yyMMdd +
    // "-" + a 5-digit zero-padded clinlims.site_code_seq value.
    const m = body.siteCode.match(/^S(\d{6})-(\d{5})$/);
    expect(m, `${GENERATE_SITE_CODE} siteCode format "${body.siteCode}"`).not.toBeNull();

    // The date component must be TODAY IN UTC, not host-local time. This is
    // the exact assertion that would have caught a real bug found this
    // session: Java's container is pinned to TZ=UTC (docker-compose.yml),
    // but this port originally called plain time.Now(), which used whatever
    // timezone the Go binary's host happened to be in — producing a
    // different site-code date near UTC day boundaries. Fixed to
    // time.Now().UTC(); see migration/b2-org-provider-migration.md §3.1 #7.
    // (Negligible, accepted flake risk: a request issued in the same
    // millisecond as a UTC midnight rollover could see the date computed
    // here disagree with the server's — not worth guarding against.)
    const now = new Date();
    const expectedDate =
      String(now.getUTCFullYear() % 100).padStart(2, "0") +
      String(now.getUTCMonth() + 1).padStart(2, "0") +
      String(now.getUTCDate()).padStart(2, "0");
    expect(m![1], `${GENERATE_SITE_CODE} date component is today in UTC`).toBe(expectedDate);
  });

  test("departments-for-site: IdValuePair rows for a real org id; edge cases", async ({ request }) => {
    const list = await readJsonArray(await request.get(ORG_LIST), ORG_LIST);
    const parentId = list[0].id;

    // NOTE: not readJsonArray — in this dev dataset no organization has any
    // children, so an empty array is the real, expected response, not a
    // signal something is broken. Shape is checked either way.
    const body = await readJson(
      await request.get(`${DEPARTMENTS_FOR_SITE}?refferingSiteId=${parentId}`),
      DEPARTMENTS_FOR_SITE,
    );
    expect(Array.isArray(body), `${DEPARTMENTS_FOR_SITE} is array`).toBe(true);
    for (const row of body) {
      expectExactKeys(row, ["id", "value"], `${DEPARTMENTS_FOR_SITE} row`);
      expectNonEmptyString(row.id, `${DEPARTMENTS_FOR_SITE} id`);
      expectNonEmptyString(row.value, `${DEPARTMENTS_FOR_SITE} value`);
    }

    // A nonexistent parent id isn't validated on either side — just a WHERE
    // org_id = X with no matches, so still 200 + [], not a 404.
    const unknownParentBody = await readJson(
      await request.get(`${DEPARTMENTS_FOR_SITE}?refferingSiteId=${MISSING_ID}`),
      `${DEPARTMENTS_FOR_SITE} unknown parent`,
    );
    expect(Array.isArray(unknownParentBody), `${DEPARTMENTS_FOR_SITE} unknown parent is array`).toBe(true);
    expect(unknownParentBody.length, `${DEPARTMENTS_FOR_SITE} unknown parent is empty`).toBe(0);

    // Missing required query param: status matches on both sides (400); body
    // format differs (Java: Spring ProblemDetail JSON; Go: plain text), so
    // only status is asserted.
    const missingParam = await request.get(DEPARTMENTS_FOR_SITE);
    expect(missingParam.status(), `${DEPARTMENTS_FOR_SITE} missing refferingSiteId`).toBe(400);
  });

  test("user-programs: IdValuePair rows, unique ids", async ({ request }, testInfo) => {
    // Not a pure reference read — Java filters by the calling user's own
    // lab-unit/test-section permissions (session/RBAC-scoped). The Go port
    // has no session/RBAC layer yet and deliberately doesn't implement this
    // endpoint at all (see migration/b2-org-provider-migration.md §2) rather
    // than silently return an unfiltered list. Skip under go-parity instead
    // of asserting a 404 there — that 404 isn't a "port behavior" worth
    // pinning, it's just "not built yet".
    test.skip(testInfo.project.name === "go-parity", "user-programs is deferred, not implemented in Go");

    const body = await readJsonArray(await request.get(USER_PROGRAMS), USER_PROGRAMS);
    for (const row of body) {
      expectExactKeys(row, ["id", "value"], `${USER_PROGRAMS} row`);
      expectNonEmptyString(row.id, `${USER_PROGRAMS} id`);
      expectNonEmptyString(row.value, `${USER_PROGRAMS} value`);
    }
    expectUnique(body.map((r: any) => r.id), `${USER_PROGRAMS} id`);
  });
});
