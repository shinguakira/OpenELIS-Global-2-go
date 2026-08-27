// §4 — b2: organization/generate-site-code (STATE-CHANGING).
//
// Lives in tests/mutating/, not tests/readonly/, because this endpoint is not
// a read: OrganizationServiceImpl.generateSiteCode() executes
// `SELECT nextval('clinlims.site_code_seq')`, so every call permanently
// advances a shared sequence in the database. It was originally written into
// the read-only suite, where it ran under BOTH the api-readonly and go-parity
// projects — burning two site codes out of the shared sequence on every full
// test run, against a production-like database. Moved here so the read-only
// suite stays genuinely non-mutating.
//
// It is still wired into the go-parity project (see playwright.config.ts) so
// the Go port keeps this coverage — the assertion below is what pins the
// UTC-vs-host-local timezone fix.
import { test, expect } from "@playwright/test";
import { readJson, expectExactKeys } from "../../fixtures/assert";

const GENERATE_SITE_CODE = "rest/organization/generate-site-code";

test.describe("b2 — organization site-code generation (consumes a sequence)", () => {
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
});
