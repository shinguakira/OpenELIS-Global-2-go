// §4 — b2: provider reference reads (strict parity gate).
//
// Taxonomy Type B, Wave 2. Runs against the Go port via the go-parity project
// (see playwright.config.ts). All ids used below are discovered live from the
// running dataset (via provider/search and direct DB queries), never
// hardcoded — resilient to whichever dataset this suite runs against.
//
// Design (same as a2/b1/b2-organization): assert the real contract, not
// "HTTP 200". Three real, confirmed divergences between this port and Java
// are asserted explicitly on both sides (branched on testInfo.project.name),
// not silently pinned to one value — full writeup in
// migration/b2-org-provider-migration.md §3:
//   - not-found on a single-id lookup: 404 (Go) vs 500 (Java, confirmed bug)
//   - provider/search isActive: the real flag (Go) vs Java's confirmed
//     always-false bug ("Y".equals(Boolean) can never be true)
//   - rest/practitioner on a person with no linked provider: uniform 404
//     (Go) vs Java's real 200-with-empty-body
//
// Findings excluded here (real, out of scope): rest/ProviderMenu,
// rest/SearchProviderMenu — Struts-legacy AdminMenuForm envelope, no e2e
// contract, deferred (see migration/b2-org-provider-migration.md §2).
import { test, expect } from "@playwright/test";
import { readJson, expectKeysWithin, expectUnique, expectNonEmptyString } from "../../fixtures/assert";
import { query, count } from "../../fixtures/db";

const PROVIDER_RAW = "rest/Provider/raw";
const PROVIDER_PERSON = "rest/Provider/Person";
const PRACTITIONER = "rest/practitioner";
const PROVIDER_SEARCH = "rest/provider/search";

// Large enough that it will never collide with a real seeded/test id.
const MISSING_ID = "999999999";

const PERSON_ALLOWED = [
  "id", "lastName", "firstName", "middleName", "multipleUnit", "streetAddress",
  "city", "state", "zipCode", "country", "workPhone", "homePhone", "cellPhone",
  "primaryPhone", "fax", "email", "gpsLatitude", "gpsLongitude", "lastupdated",
];
const PROVIDER_ALLOWED = [
  "id", "externalId", "npi", "providerType", "person", "fhirUuid",
  "fhirUuidAsString", "active", "desynchronized", "lastupdated",
];

test.describe("b2 — provider reference reads", () => {
  test("provider/search: envelope shape + DB-count oracle + isActive divergence", async ({ request }, testInfo) => {
    const body = await readJson(await request.get(`${PROVIDER_SEARCH}?search=`), PROVIDER_SEARCH);
    expectKeysWithin(
      body,
      ["providers", "totalCount", "page", "pageSize"],
      ["providers", "totalCount", "page", "pageSize"],
      `${PROVIDER_SEARCH} envelope`,
    );
    expect(Array.isArray(body.providers), `${PROVIDER_SEARCH} providers is array`).toBe(true);
    // DB oracle: an unfiltered search's totalCount must equal the real row
    // count, live, regardless of how big the dataset is (live-confirmed:
    // Java's empty-string search matches every row too, not a no-op filter).
    expect(body.totalCount, `${PROVIDER_SEARCH} totalCount matches DB`).toBe(count("provider"));
    expect(body.providers.length, `${PROVIDER_SEARCH} providers.length <= totalCount`).toBeLessThanOrEqual(
      body.totalCount,
    );

    // Independent oracle for the isActive divergence below — real DB values,
    // not derived from the API response under test.
    const activeRows = query("SELECT id, active FROM clinlims.provider");
    const realActive = new Map(activeRows.map(([id, active]) => [id, active === "t"]));

    for (const row of body.providers) {
      expectKeysWithin(
        row,
        ["id", "personId", "firstName", "lastName", "name", "phone", "fax", "email", "externalId", "isActive"],
        ["id", "isActive"],
        `${PROVIDER_SEARCH} row`,
      );
      expectNonEmptyString(row.id, `${PROVIDER_SEARCH} row id`);
      expect(typeof row.isActive, `${PROVIDER_SEARCH} row ${row.id} isActive is boolean`).toBe("boolean");

      // migration/b2-org-provider-migration.md §3.2 #3, live-confirmed against
      // all 3 real dev providers (one inactive, two active): Java's
      // `"Y".equals(provider.getActive())` compares a Boolean to the string
      // "Y" and can never be true, so isActive is always false regardless of
      // the real column. This port returns the real flag instead.
      if (testInfo.project.name === "go-parity") {
        expect(
          row.isActive,
          `${PROVIDER_SEARCH} row ${row.id} isActive matches DB active column (Go: real flag)`,
        ).toBe(realActive.get(row.id));
      } else {
        expect(
          row.isActive,
          `${PROVIDER_SEARCH} row ${row.id} isActive (Java: confirmed always-false bug)`,
        ).toBe(false);
      }
    }
    expectUnique(body.providers.map((r: any) => r.id), `${PROVIDER_SEARCH} id`);
  });

  test("provider/search: pageSize caps the page without affecting totalCount", async ({ request }) => {
    const body = await readJson(
      await request.get(`${PROVIDER_SEARCH}?search=&page=1&pageSize=1`),
      PROVIDER_SEARCH,
    );
    expect(body.pageSize, `${PROVIDER_SEARCH} pageSize echoes request`).toBe(1);
    expect(body.providers.length, `${PROVIDER_SEARCH} providers capped to pageSize`).toBeLessThanOrEqual(1);
    expect(body.totalCount, `${PROVIDER_SEARCH} totalCount unaffected by pageSize`).toBe(count("provider"));
  });

  test("provider/search?search=<lastName>: substring match includes the target", async ({ request }) => {
    const all = await readJson(await request.get(`${PROVIDER_SEARCH}?search=`), PROVIDER_SEARCH);
    const target = all.providers.find((p: any) => p.lastName);
    test.skip(!target, "no provider with a lastName in this dataset");

    const body = await readJson(
      await request.get(`${PROVIDER_SEARCH}?search=${encodeURIComponent(target.lastName)}`),
      `${PROVIDER_SEARCH}?search=<lastName>`,
    );
    expect(
      body.providers.some((p: any) => p.id === target.id),
      `${PROVIDER_SEARCH}?search=${target.lastName} includes provider ${target.id}`,
    ).toBe(true);
  });

  test("provider/search?phone=<no-match>: empty result, not an error", async ({ request }) => {
    const body = await readJson(
      await request.get(`${PROVIDER_SEARCH}?phone=00000NOMATCH99999`),
      `${PROVIDER_SEARCH}?phone=<no-match>`,
    );
    expect(body.providers, `${PROVIDER_SEARCH}?phone=no-match providers`).toEqual([]);
    expect(body.totalCount, `${PROVIDER_SEARCH}?phone=no-match totalCount`).toBe(0);
  });

  test("Provider/raw/{id}: full entity + nested person; not-found diverges by design", async ({
    request,
  }, testInfo) => {
    const all = await readJson(await request.get(`${PROVIDER_SEARCH}?search=`), PROVIDER_SEARCH);
    test.skip(all.providers.length === 0, "no providers in this dataset");
    const sample = all.providers[0];

    const body = await readJson(await request.get(`${PROVIDER_RAW}/${sample.id}`), `${PROVIDER_RAW}/{id}`);
    expectKeysWithin(
      body,
      PROVIDER_ALLOWED,
      ["id", "person", "fhirUuidAsString", "active", "desynchronized"],
      `${PROVIDER_RAW}/{id}`,
    );
    expect(body.id, `${PROVIDER_RAW}/{id} id`).toBe(sample.id);
    expect(body.person.id, `${PROVIDER_RAW}/{id} person.id`).toBe(sample.personId);
    expectKeysWithin(body.person, PERSON_ALLOWED, ["id"], `${PROVIDER_RAW}/{id} person`);

    // Not-found: real, confirmed, deliberate divergence — migration doc §3.2 #1.
    const missing = await request.get(`${PROVIDER_RAW}/${MISSING_ID}`);
    if (testInfo.project.name === "go-parity") {
      expect(missing.status(), `${PROVIDER_RAW}/{id} missing (Go: real 404)`).toBe(404);
    } else {
      expect(missing.status(), `${PROVIDER_RAW}/{id} missing (Java: confirmed 500 bug)`).toBe(500);
    }
  });

  test("Provider/Person/{id}: person shape; not-found diverges by design", async ({ request }, testInfo) => {
    const all = await readJson(await request.get(`${PROVIDER_SEARCH}?search=`), PROVIDER_SEARCH);
    test.skip(all.providers.length === 0, "no providers in this dataset");
    const personId = all.providers[0].personId;

    const body = await readJson(await request.get(`${PROVIDER_PERSON}/${personId}`), `${PROVIDER_PERSON}/{id}`);
    expectKeysWithin(body, PERSON_ALLOWED, ["id"], `${PROVIDER_PERSON}/{id}`);
    expect(body.id, `${PROVIDER_PERSON}/{id} id`).toBe(personId);

    // Provider/Person/{id} is a plain person lookup — unlike rest/practitioner
    // below, it does not care whether the person has a linked provider at
    // all (live-confirmed against a person with zero provider rows: 200 with
    // full person data, no ambiguity).
    const missing = await request.get(`${PROVIDER_PERSON}/${MISSING_ID}`);
    if (testInfo.project.name === "go-parity") {
      expect(missing.status(), `${PROVIDER_PERSON}/{id} missing (Go: real 404)`).toBe(404);
    } else {
      expect(missing.status(), `${PROVIDER_PERSON}/{id} missing (Java: confirmed 500 bug)`).toBe(500);
    }
  });

  test("rest/practitioner: linked person returns Provider; missing person diverges by design", async ({
    request,
  }, testInfo) => {
    const all = await readJson(await request.get(`${PROVIDER_SEARCH}?search=`), PROVIDER_SEARCH);
    test.skip(all.providers.length === 0, "no providers in this dataset");
    const personId = all.providers[0].personId;

    // providerId is misleadingly named — it's really a Person id (kept as-is
    // to match every real Java caller; see migration doc §4).
    const body = await readJson(await request.get(`${PRACTITIONER}?providerId=${personId}`), PRACTITIONER);
    expect(body.person.id, `${PRACTITIONER} person.id`).toBe(personId);
    expectKeysWithin(
      body,
      PROVIDER_ALLOWED,
      ["id", "person", "fhirUuidAsString", "active", "desynchronized"],
      `${PRACTITIONER} body`,
    );

    const missing = await request.get(`${PRACTITIONER}?providerId=${MISSING_ID}`);
    if (testInfo.project.name === "go-parity") {
      expect(missing.status(), `${PRACTITIONER} missing person (Go: real 404)`).toBe(404);
    } else {
      expect(missing.status(), `${PRACTITIONER} missing person (Java: confirmed 500 bug)`).toBe(500);
    }
  });

  test("rest/practitioner: person with no linked provider — 3-way divergence by design", async ({
    request,
  }, testInfo) => {
    const unlinked = query(
      "SELECT id FROM clinlims.person WHERE id NOT IN (SELECT person_id FROM clinlims.provider) LIMIT 1",
    );
    test.skip(unlinked.length === 0, "every person in this dataset has a linked provider");
    const personId = unlinked[0][0];

    const res = await request.get(`${PRACTITIONER}?providerId=${personId}`);
    // migration/b2-org-provider-migration.md §3.2 #9: Java's
    // ProviderDAOImpl.getProviderByPerson runs a clean HQL query, finds no
    // rows, and returns null (no exception) — the controller's `return
    // provider;` then returns null, and Spring's @ResponseBody serializes
    // that as HTTP 200 with a genuinely empty body (Content-Length: 0,
    // confirmed via response headers, not just an empty-looking curl
    // output). This port keeps a uniform 404 across this whole endpoint
    // family instead of adding a third, easy-to-mishandle response shape
    // (e.g. `response.json()` on an empty body throws for a real caller) —
    // deliberate, open to being overridden.
    if (testInfo.project.name === "go-parity") {
      expect(res.status(), `${PRACTITIONER} unlinked person (Go: uniform 404)`).toBe(404);
    } else {
      expect(res.status(), `${PRACTITIONER} unlinked person (Java: real 200)`).toBe(200);
      const text = await res.text();
      expect(text, `${PRACTITIONER} unlinked person (Java: genuinely empty body)`).toBe("");
    }
  });
});
