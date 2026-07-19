// §4 — b2: organization + program reference reads (strict parity gate).
//
// Taxonomy Type B, Wave 2. The provider-side reads are all param-driven
// (search / menu), so the param-less b2 reads are organization types/list +
// user programs. Same design as b1: assert the real contract, not "HTTP 200".
//
// Findings excluded here (real, out of scope): GET rest/Organization returns 500
// on Java itself; rest/organization/search is a paginated Type-C search (its own
// group); rest/EntityNamesProvider / getPendingAnalysisForTestProvider need query
// params (400 without).
import { test, expect } from "@playwright/test";
import {
  readJsonArray,
  expectExactKeys,
  expectKeysWithin,
  expectUnique,
  expectNonEmptyString,
} from "../../fixtures/assert";

const ORG_TYPES = "rest/organization/types";
const ORG_LIST = "rest/organization-list";
const USER_PROGRAMS = "rest/user-programs";

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

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

  test("user-programs: IdValuePair rows, unique ids", async ({ request }) => {
    const body = await readJsonArray(await request.get(USER_PROGRAMS), USER_PROGRAMS);
    for (const row of body) {
      expectExactKeys(row, ["id", "value"], `${USER_PROGRAMS} row`);
      expectNonEmptyString(row.id, `${USER_PROGRAMS} id`);
      expectNonEmptyString(row.value, `${USER_PROGRAMS} value`);
    }
    expectUnique(body.map((r: any) => r.id), `${USER_PROGRAMS} id`);
  });
});
