// §8 — e1: admin config CRUD (STATE-CHANGING). Wave 6, the first WRITE wave.
//
// Lives in tests/mutating/ because every test here changes the database, and
// site_information is not ordinary data: it is the application's configuration.
// Java runs ConfigurationProperties.loadDBValuesIntoConfiguration() after every
// write, so editing a shipped row would change the behaviour of the whole
// application — including the other specs in the same run. Every test therefore
// creates and destroys its OWN row and never touches one that ships.
//
// ── WHY A READ DIFF IS NOT ENOUGH HERE ─────────────────────────────────────
// The parity question changes shape once writes are involved. A read spec
// compares two documents; a write spec has to ask whether both implementations
// left the same database behind. Those are different questions, and on this
// endpoint they have different answers:
//
//   - the INSERT response echoes back the valueType it was sent while the row
//     stores "text"
//   - the UPDATE response echoes back paramName and description while the row
//     keeps its old ones
//   - every write leaves an audit row in clinlims.history that no response
//     mentions at all
//
// A probe that trusted the responses would have found none of the three. The
// assertions below read the DATABASE for anything that matters.
import { test, expect } from "@playwright/test";
import type { APIRequestContext } from "@playwright/test";
import { readJson } from "../../fixtures/assert";
import { query, exec } from "../../fixtures/db";
import {
  SESSION_PATH,
  CSRF_SESSION_FIELD,
  CSRF_HEADER,
} from "../../fixtures/contract";

const SITE_INFORMATION = "rest/SiteInformation";
const SITE_INFORMATION_MENU = "rest/SiteInformationMenu";
const CANCEL = "rest/CancelSiteInformation";
const DELETE_SITE_INFORMATION = "rest/DeleteSiteInformation";
const DELETE_PATIENT_CONFIG = "rest/DeletePatientConfiguration";
const PATIENT_CONFIG_MENU = "rest/PatientConfigurationMenu";
const LAB_UNIT_CONFIG = "rest/labUnit/config";

/** A name this spec owns. Nothing ships with it and no Property maps to it. */
const PROBE = "e2eConfigProbe";

/** reference_tables row the audit trail keys site_information writes under. */
function siteInfoTableId(): string {
  const rows = query(
    `SELECT id FROM clinlims.reference_tables WHERE name = 'site_information'`,
  );
  expect(rows.length, "reference_tables has a site_information row").toBe(1);
  return rows[0][0];
}

/**
 * The masked CSRF token for this session.
 *
 * Read fresh for every write: Spring's XorCsrfTokenRequestAttributeHandler
 * returns a DIFFERENT masked string on each read, and the port reproduces that
 * (see fixtures/csrf.ts). Caching one would still work — they all un-mask to
 * the same token — but re-reading is what a real client does.
 */
async function csrfToken(request: APIRequestContext): Promise<string> {
  const body = await readJson(await request.get(SESSION_PATH), SESSION_PATH);
  const token = body[CSRF_SESSION_FIELD];
  expect(typeof token, `${SESSION_PATH} carries a ${CSRF_SESSION_FIELD}`).toBe(
    "string",
  );
  return token;
}

async function postConfig(
  request: APIRequestContext,
  path: string,
  data: Record<string, unknown>,
) {
  return request.post(path, {
    headers: {
      [CSRF_HEADER]: await csrfToken(request),
      "Content-Type": "application/json",
    },
    data,
  });
}

/** The audit rows a given site_information id accumulated, oldest first. */
function auditRows(
  id: string,
): { activity: string; user: string; changes: string }[] {
  return query(
    // The newlines are stripped IN SQL. `changes` is an XML fragment per
    // field, newline-separated, and psql -tA emits one ROW per line — so a
    // multi-line value is parsed as several rows and the activity column comes
    // back holding XML. That is not a hypothetical: it is what this helper did
    // on its first run.
    `SELECT h.activity, h.sys_user_id::text,
            replace(COALESCE(convert_from(h.changes, 'UTF8'), ''), chr(10), ' ')
       FROM clinlims.history h
      WHERE h.reference_table = ${siteInfoTableId()}
        AND h.reference_id = ${id}
      -- Ordered by TIMESTAMP, not id. Java's history ids are not chronological:
      -- Hibernate hands them out of a cached sequence block, so an insert's
      -- audit row can carry a HIGHER id than the update that followed it —
      -- observed as 866 (I) against 845 (U) and 846 (D) on one row. The
      -- sequence position is an allocation artefact; the timestamps are the
      -- record of what happened when.
      ORDER BY h.timestamp, h.id`,
  ).map(([activity, user, changes]) => ({ activity, user, changes }));
}

function probeIds(): string[] {
  return query(
    `SELECT id::text FROM clinlims.site_information WHERE name IN ('${PROBE}', 'RENAMED')`,
  ).map((r) => r[0]);
}

test.describe("e1 — SiteInformation config CRUD (writes)", () => {
  // Ids this spec created, so the cleanup can find their audit rows after the
  // row itself is gone.
  let touched: string[] = [];

  // The two tests below edit SHIPPED rows because no created row can reach
  // their branches. What they found is captured here and put back in afterEach,
  // which runs even when the test fails — a config row left flipped changes the
  // behaviour of every other spec in the run.
  let localizationRestore: { id: string; values: string[][] } | null = null;
  let settingRestore: { id: string; value: string; roleActive: string } | null =
    null;
  let configRestore: { id: string; value: string } | null = null;
  let localeRestore: {
    id: string;
    value: string;
    localizationId: string;
    fr: string;
  } | null = null;

  test.beforeEach(() => {
    touched = probeIds(); // anything a previous crashed run left behind
    localizationRestore = null;
    settingRestore = null;
    configRestore = null;
    localeRestore = null;
  });

  // Cleanup runs even when the test fails — a leftover row here is global
  // configuration, and a leftover audit row makes the next run's assertions
  // ambiguous.
  test.afterEach(async ({ request }) => {
    const ids = [...new Set([...touched, ...probeIds()])];
    for (const id of ids) {
      exec(
        `DELETE FROM clinlims.history WHERE reference_table = ${siteInfoTableId()} AND reference_id = ${id}`,
      );
    }
    exec(
      `DELETE FROM clinlims.site_information WHERE name IN ('${PROBE}', 'RENAMED')`,
    );

    if (localizationRestore) {
      for (const [locale, value] of localizationRestore.values) {
        exec(
          `UPDATE clinlims.localization_value SET value = '${value.replace(/'/g, "''")}'
            WHERE localization_id = ${Number(localizationRestore.id)} AND locale = '${locale}'`,
        );
      }
    }
    if (localeRestore) {
      // Restored THROUGH THE API, not with a direct UPDATE.
      //
      // The locale is a CACHED setting — which is exactly what the test two
      // above proves — so putting the row back by hand leaves both servers
      // still answering in French until something writes through them. An
      // earlier version of this cleanup did that and left the deployment
      // French for the rest of the run, which is how the flake was found.
      await postConfig(request, `${SITE_INFORMATION}?ID=${localeRestore.id}`, {
        paramName: "default language locale",
        value: localeRestore.value,
        siteInfoDomainName: "SiteInformation",
        valueType: "dictionary",
      });
      exec(
        `UPDATE clinlims.localization_value SET value = '${localeRestore.fr.replace(/'/g, "''")}'
          WHERE localization_id = ${Number(localeRestore.localizationId)} AND locale = 'fr'`,
      );
      exec(
        `DELETE FROM clinlims.history WHERE reference_table = ${siteInfoTableId()}
          AND reference_id = ${Number(localeRestore.id)}`,
      );
    }
    if (configRestore) {
      exec(
        `UPDATE clinlims.site_information SET value = '${configRestore.value}'
          WHERE id = ${Number(configRestore.id)}`,
      );
    }
    if (settingRestore) {
      exec(
        `UPDATE clinlims.site_information SET value = '${settingRestore.value}'
          WHERE id = ${Number(settingRestore.id)}`,
      );
      exec(
        `UPDATE clinlims.system_role SET active = ${settingRestore.roleActive}
          WHERE trim(name) = 'Results modifier'`,
      );
      // The restore is a direct write, so it leaves audit rows the application
      // did not make. Remove them too, or the next run's counts are off.
      exec(
        `DELETE FROM clinlims.history WHERE reference_table = ${siteInfoTableId()}
          AND reference_id = ${Number(settingRestore.id)}`,
      );
    }
  });

  test("the CRUD cycle writes the row Java writes AND the audit trail Java writes", async ({
    request,
  }) => {
    // ── INSERT ───────────────────────────────────────────────────────────
    const created = await postConfig(request, SITE_INFORMATION, {
      paramName: PROBE,
      value: "alpha",
      siteInfoDomainName: "SiteInformation",
      encrypted: false,
      // Deliberately NOT "text". The insert path hardcodes setValueType("text")
      // and the response echoes this back regardless, so response and row
      // disagree — the reason the row is asserted below and not the body.
      valueType: "boolean",
      description: "e2e config probe",
    });
    expect(created.status(), "insert").toBe(200);
    expect(
      (await created.json()).valueType,
      "the RESPONSE echoes what was sent",
    ).toBe("boolean");

    const rows = query(
      `SELECT id::text, value, value_type, domain_id::text, description, encrypted::text
         FROM clinlims.site_information WHERE name = '${PROBE}'`,
    );
    expect(rows.length, "exactly one row was created").toBe(1);
    const [id, value, valueType, domainId, description, encrypted] = rows[0];
    touched.push(id);

    expect(value, "the value is stored").toBe("alpha");
    expect(
      valueType,
      "...but value_type is FORCED to text, not what was sent",
    ).toBe("text");
    expect(domainId, "the new row lands in the site-identity domain").toBe("1");
    expect(description, "the description is stored on insert").toBe(
      "e2e config probe",
    );
    expect(encrypted, "encrypted is stored").toBe("false");

    // The audit trail. Nothing in any response mentions it, and a port that
    // wrote only the row passes every response assertion above.
    expect(
      auditRows(id).map((r) => r.activity),
      "insert leaves an I audit row",
    ).toEqual(["I"]);

    // ── UPDATE ───────────────────────────────────────────────────────────
    const updated = await postConfig(request, `${SITE_INFORMATION}?ID=${id}`, {
      paramName: "RENAMED",
      value: "beta",
      siteInfoDomainName: "SiteInformation",
      encrypted: false,
      valueType: "text",
      description: "changed",
    });
    expect(updated.status(), "update").toBe(200);
    const updatedBody = await updated.json();
    expect(updatedBody.paramName, "the RESPONSE echoes the new name").toBe(
      "RENAMED",
    );
    expect(updatedBody.description, "...and the new description").toBe(
      "changed",
    );

    const [[nameAfter, valueAfter, descriptionAfter]] = query(
      `SELECT name, value, description FROM clinlims.site_information WHERE id = ${id}`,
    );
    expect(valueAfter, "the value IS updated").toBe("beta");
    // The update path loads the row by id and calls setValue and nothing else.
    expect(
      nameAfter,
      "the name is NOT — the rename is echoed and dropped",
    ).toBe(PROBE);
    expect(descriptionAfter, "neither is the description").toBe(
      "e2e config probe",
    );

    const afterUpdate = auditRows(id);
    expect(
      afterUpdate.map((r) => r.activity),
      "update leaves a U audit row",
    ).toEqual(["I", "U"]);
    // The audit payload is the row's PREVIOUS state, as XML fragments — so the
    // U row records the value the update replaced, not the one it wrote.
    expect(afterUpdate[1].changes, "the U row records the OLD value").toContain(
      "<value>alpha</value>",
    );
    expect(afterUpdate[1].changes, "...not the new one").not.toContain("beta");

    // ── READ BACK ────────────────────────────────────────────────────────
    const readBack = await readJson(
      await request.get(`${SITE_INFORMATION}?ID=${id}`),
      `${SITE_INFORMATION}?ID=${id}`,
    );
    expect(readBack.paramName, "the read-back name is the STORED one").toBe(
      PROBE,
    );
    expect(readBack.value, "and the stored value").toBe("beta");

    // ── DELETE ───────────────────────────────────────────────────────────
    const deleted = await request.get(DELETE_SITE_INFORMATION, {
      headers: {
        [CSRF_HEADER]: await csrfToken(request),
        "Content-Type": "application/json",
      },
      // A GET that carries a request body. Unusual, and it is the contract.
      data: { selectedIDs: [id], siteInfoDomainName: "SiteInformation" },
    });
    expect(deleted.status(), "delete").toBe(200);
    expect(await deleted.json(), "the body is a bare JSON string").toBe(
      "Delete successful",
    );

    expect(
      query(`SELECT id FROM clinlims.site_information WHERE id = ${id}`).length,
      "the row is gone",
    ).toBe(0);

    const afterDelete = auditRows(id);
    expect(
      afterDelete.map((r) => r.activity),
      "delete leaves a D audit row",
    ).toEqual(["I", "U", "D"]);
    expect(
      afterDelete[2].changes,
      "the D row records the row that was removed",
    ).toContain(`<name>${PROBE}</name>`);
    // Every audit row is attributed to the acting user, not to a system id.
    expect(
      new Set(afterDelete.map((r) => r.user)).size,
      "all three audit rows share one acting user",
    ).toBe(1);
  });

  test("a write without the CSRF token is refused, with Java's hand-built body", async ({
    request,
  }) => {
    // The body is assembled by string concatenation in SecurityConfig's
    // accessDeniedHandler, so it carries a space after the brace and after each
    // colon and puts status BEFORE message. A marshalled map produces neither,
    // and the port did exactly that from p0 until this wave — undetected,
    // because the body is only reachable on a denial and no ported route had a
    // state-changing verb to be denied on.
    const res = await request.post(SITE_INFORMATION, {
      headers: { "Content-Type": "application/json" },
      data: {
        paramName: PROBE,
        value: "x",
        siteInfoDomainName: "SiteInformation",
      },
    });
    expect(res.status(), "no CSRF token").toBe(403);
    expect(await res.text(), "the exact hand-built body").toBe(
      '{ "status": 403, "message": "CSRF token missing or invalid" }',
    );

    expect(
      query(`SELECT id FROM clinlims.site_information WHERE name = '${PROBE}'`)
        .length,
      "and nothing was written",
    ).toBe(0);
  });

  test("the id parameter is ID, not id — and an unknown one is a 500", async ({
    request,
  }) => {
    // BaseController's constant is ID. `?id=` misses entirely, the handler
    // takes its is-new branch, and the caller gets a blank add-new form for a
    // row that exists — an edit screen showing empty fields.
    const known = query(
      `SELECT id::text, name FROM clinlims.site_information
        WHERE name = '24 hour clock'`,
    );
    expect(known.length, "the reference row is present").toBe(1);
    const [knownId, knownName] = known[0];

    const upper = await readJson(
      await request.get(`${SITE_INFORMATION}?ID=${knownId}`),
      "?ID=",
    );
    expect(upper.paramName, "?ID= loads the row").toBe(knownName);

    const lower = await readJson(
      await request.get(`${SITE_INFORMATION}?id=${knownId}`),
      "?id=",
    );
    expect(lower.paramName, "?id= is ignored and answers the blank form").toBe(
      "",
    );
    expect(lower.value, "...with no value either").toBe("");

    // INVERSION: the blank answer above is the is-new branch, not an empty row.
    // ID=0 is the explicit is-new request and produces the same document.
    const zero = await readJson(
      await request.get(`${SITE_INFORMATION}?ID=0`),
      "?ID=0",
    );
    expect(zero, "?id=<real> is indistinguishable from ?ID=0").toEqual(lower);

    // An id that matches no row: the service returns null and the next line
    // dereferences it, so this is Tomcat's 500 page rather than a 404.
    const missing = await request.get(`${SITE_INFORMATION}?ID=999999`);
    expect(missing.status(), "an unknown ID is a 500, not a 404").toBe(500);
  });

  test("POST answers a different form shape than GET", async ({ request }) => {
    // setupFormForRequest runs on the GET path and not on the POST path, so the
    // POST response is the submitted body over the bean defaults.
    const get = await readJson(
      await request.get(SITE_INFORMATION),
      SITE_INFORMATION,
    );
    expect(get.formName, "GET names the form for its domain").toBe(
      "SiteInformationForm",
    );
    expect(get.formAction, "and carries a formAction").toBe("SiteInformation");
    expect(get.cancelAction, "and a domain-specific cancel").toBe(
      "CancelSiteInformation",
    );

    const post = await postConfig(request, SITE_INFORMATION, {
      paramName: PROBE,
      value: "alpha",
      siteInfoDomainName: "SiteInformation",
    });
    expect(post.status()).toBe(200);
    const body = await post.json();
    touched.push(...probeIds());

    expect(body.formName, "POST answers the BEAN default, lowercased").toBe(
      "siteInformationForm",
    );
    expect("formAction" in body, "POST carries no formAction at all").toBe(
      false,
    );
    expect(body.cancelAction, "and the bean's cancel, not the domain's").toBe(
      "Home",
    );
    // cancelMethod says POST on both, and Cancel{domain} is a GET route — the
    // form tells the client to use a verb the route does not accept.
    expect(body.cancelMethod, "cancelMethod is POST").toBe("POST");
  });

  test("DeletePatientConfiguration rejects the domain name its own menu hands out", async ({
    request,
  }) => {
    // The menu controller spells the domain correctly; the form controller and
    // the delete validator's allow-list spell it "PaitientConfiguration". A
    // client that round-trips the value it was given is refused.
    const menu = await readJson(
      await request.get(PATIENT_CONFIG_MENU),
      PATIENT_CONFIG_MENU,
    );
    expect(
      menu.siteInfoDomainName,
      "the menu hands out the correct spelling",
    ).toBe("PatientConfiguration");

    const refused = await request.get(DELETE_PATIENT_CONFIG, {
      headers: {
        [CSRF_HEADER]: await csrfToken(request),
        "Content-Type": "application/json",
      },
      data: { selectedIDs: [], siteInfoDomainName: menu.siteInfoDomainName },
    });
    expect(refused.status(), "the value the menu gave is rejected").toBe(400);

    // A fourth error envelope: a bare ARRAY of Spring ObjectErrors — not the
    // RFC 7807 ProblemDetail a binding failure produces, not the per-field
    // `errors` map a @Valid form produces, not Tomcat's {timestamp,status}.
    const errors = await refused.json();
    expect(Array.isArray(errors), "the body is a bare array").toBe(true);
    expect(errors[0].field).toBe("siteInfoDomainName");
    expect(errors[0].rejectedValue, "and it names the value it refused").toBe(
      "PatientConfiguration",
    );
    expect(errors[0].code).toBe("error.field.option.invalid");

    // ...and the MISSPELLED one is accepted, which is what makes the above a
    // defect rather than ordinary validation.
    const accepted = await request.get(DELETE_PATIENT_CONFIG, {
      headers: {
        [CSRF_HEADER]: await csrfToken(request),
        "Content-Type": "application/json",
      },
      data: { selectedIDs: [], siteInfoDomainName: "PaitientConfiguration" },
    });
    expect(accepted.status(), "only the typo is accepted").toBe(200);
  });

  test("the menu paging block is constant, whatever the row count", async ({
    request,
  }) => {
    // createMenuList returns the FULL list and the client paginates, but the
    // form still ships a paging block — and it says the same thing on every
    // domain regardless of size.
    const site = await readJson(
      await request.get(SITE_INFORMATION_MENU),
      SITE_INFORMATION_MENU,
    );
    const patient = await readJson(
      await request.get(PATIENT_CONFIG_MENU),
      PATIENT_CONFIG_MENU,
    );

    expect(
      site.menuList.length,
      "the two domains hold different numbers of rows",
    ).not.toBe(patient.menuList.length);
    for (const [name, body] of [
      ["SiteInformation", site],
      ["PatientConfiguration", patient],
    ] as const) {
      expect(body.totalRecordCount, `${name} totalRecordCount`).toBe("");
      expect(body.fromRecordCount, `${name} fromRecordCount`).toBe("1");
      expect(body.toRecordCount, `${name} toRecordCount`).toBe("20");
    }
    // The list is the whole domain, so `to` bears no relation to it.
    expect(
      site.menuList.length,
      "and the list exceeds the reported page",
    ).toBeGreaterThan(20);

    // The menu is an ORACLE against the table, not just a shape check.
    const expected = query(
      `SELECT si.name FROM clinlims.site_information si
         JOIN clinlims.site_information_domain d ON d.id = si.domain_id
        WHERE d.name = 'siteIdentity'
        ORDER BY si.name`,
    ).map((r) => r[0]);
    expect(
      site.menuList.map((m: any) => m.name),
      "the menu is the domain's rows, ordered by the DATABASE collation",
    ).toEqual(expected);
  });

  test("invalid input is a 200 that writes NOTHING, not an error status", async ({
    request,
  }) => {
    // SiteInformationFormValidator and isValid both reject by calling
    // saveErrors and RETURNING THE FORM, so the response is a 200 carrying the
    // submitted values straight back. The only thing that distinguishes accept
    // from reject is whether the row appears — which is why every case below
    // asserts on the table rather than on the status.
    const cases: [string, Record<string, unknown>][] = [
      ["valueType outside the allow-list", { valueType: "BOGUS" }],
      [
        "siteInfoDomainName outside the allow-list",
        { siteInfoDomainName: "BOGUS" },
      ],
      ["tag outside the allow-list", { tag: "BOGUS" }],
      ["a blank paramName", { paramName: "" }],
    ];

    for (const [label, override] of cases) {
      const res = await postConfig(request, SITE_INFORMATION, {
        paramName: PROBE,
        value: "alpha",
        siteInfoDomainName: "SiteInformation",
        valueType: "text",
        ...override,
      });
      expect(res.status(), `${label}: still a 200`).toBe(200);
      expect(
        query(
          `SELECT id FROM clinlims.site_information WHERE name = '${PROBE}'`,
        ).length,
        `${label}: nothing written`,
      ).toBe(0);
    }

    // INVERSION: the same body WITHOUT an override does write, so the four
    // above are being refused rather than failing for some unrelated reason.
    const ok = await postConfig(request, SITE_INFORMATION, {
      paramName: PROBE,
      value: "alpha",
      siteInfoDomainName: "SiteInformation",
      valueType: "text",
    });
    expect(ok.status()).toBe(200);
    touched.push(...probeIds());
    expect(
      query(`SELECT id FROM clinlims.site_information WHERE name = '${PROBE}'`)
        .length,
      "the valid body writes",
    ).toBe(1);
  });

  test("the phone-format rules key on the row's NAME, not on any column", async ({
    request,
  }) => {
    // isValid decides by the name being written, so the same value is accepted
    // for one row and refused for another. "phone format" must match the format
    // regex; "phone format label" may be blank but not malformed.
    const badValue = "@@@bogus@@@";

    const refused = await postConfig(request, SITE_INFORMATION, {
      paramName: "phone format",
      value: badValue,
      siteInfoDomainName: "SiteInformation",
      valueType: "text",
    });
    expect(refused.status(), "still a 200").toBe(200);
    const [[stored]] = query(
      `SELECT value FROM clinlims.site_information WHERE name = 'phone format'`,
    );
    expect(stored, "the shipped row is untouched").toBe("xxxx-xxxx");

    // The SAME value under a name with no rule is accepted.
    const accepted = await postConfig(request, SITE_INFORMATION, {
      paramName: PROBE,
      value: badValue,
      siteInfoDomainName: "SiteInformation",
      valueType: "text",
    });
    expect(accepted.status()).toBe(200);
    touched.push(...probeIds());
    const rows = query(
      `SELECT value FROM clinlims.site_information WHERE name = '${PROBE}'`,
    );
    expect(rows.length, "a name with no format rule accepts it").toBe(1);
    expect(rows[0][0]).toBe(badValue);
  });

  test("an encrypted row stores CIPHERTEXT and reads back as plaintext", async ({
    request,
  }) => {
    // The service encrypts on write and decrypts on read, so the column never
    // holds what the caller sent and the form never shows what the column
    // holds. No shipped row is encrypted, so this whole path is unreachable
    // until a row like this one exists.
    const secret = "secret-value";
    const created = await postConfig(request, SITE_INFORMATION, {
      paramName: PROBE,
      value: secret,
      siteInfoDomainName: "SiteInformation",
      valueType: "text",
      encrypted: true,
    });
    expect(created.status(), "insert").toBe(200);

    const rows = query(
      `SELECT id::text, value, encrypted::text FROM clinlims.site_information WHERE name = '${PROBE}'`,
    );
    expect(rows.length, "the row exists").toBe(1);
    const [id, stored, encrypted] = rows[0];
    touched.push(id);

    expect(encrypted, "the flag is stored").toBe("true");
    expect(stored, "the column does NOT hold the plaintext").not.toBe(secret);
    // jasypt AES256: base64 of a 16-byte salt, a 16-byte IV and one AES block.
    expect(stored.length, "the column holds a 48-byte ciphertext").toBe(64);

    // ...and the read decrypts it back.
    const readBack = await readJson(
      await request.get(`${SITE_INFORMATION}?ID=${id}`),
      `${SITE_INFORMATION}?ID=${id}`,
    );
    expect(readBack.value, "the form shows the plaintext").toBe(secret);

    // The menu masks the DECRYPTED value, so the mask is as long as the secret
    // and not as long as the ciphertext.
    const menu = await readJson(
      await request.get(SITE_INFORMATION_MENU),
      SITE_INFORMATION_MENU,
    );
    const row = menu.menuList.find((m: any) => m.name === PROBE);
    expect(row, "the row is on the menu").toBeTruthy();
    expect(row.value, "masked to the PLAINTEXT length").toBe(
      "*".repeat(secret.length),
    );
    expect(row.value.length, "not to the ciphertext length").not.toBe(
      stored.length,
    );
  });

  test("labUnit/config drops labName when the value is blank", async ({
    request,
  }) => {
    const body = await readJson(
      await request.get(LAB_UNIT_CONFIG),
      LAB_UNIT_CONFIG,
    );

    // site_information has a SiteName row, and its value is empty — so this is
    // not "no row", it is a blank one, and the key is still absent.
    // The id is selected alongside the value on purpose: the DB helper splits
    // psql output by LINE, so a row whose only selected column is empty comes
    // back as an empty line and disappears from the result set entirely.
    const site = query(
      `SELECT id::text, COALESCE(value, '') FROM clinlims.site_information WHERE name = 'SiteName'`,
    );
    expect(site.length, "the SiteName row exists").toBe(1);
    expect(site[0][1] ?? "", "...and is blank").toBe("");
    expect("labName" in body, "so labName is absent, not empty").toBe(false);

    // orderEntryWorkflowType has no row at all, and the handler substitutes.
    expect(
      query(
        `SELECT id FROM clinlims.site_information WHERE name = 'orderEntryWorkflowType'`,
      ).length,
      "no workflow-type row is configured",
    ).toBe(0);
    expect(body.workflowType, "so the fallback is used").toBe("Both");

    // The other two come straight from their rows.
    const [[format]] = query(
      `SELECT value FROM clinlims.site_information WHERE name = 'acessionFormat'`,
    );
    expect(body.accessionFormat).toBe(format);
    const [[validate]] = query(
      `SELECT value FROM clinlims.site_information WHERE name = 'validateAccessionNumber'`,
    );
    expect(body.useAccessionNumberValidation).toBe(validate === "true");
  });

  test("a localization-tagged row writes to the LOCALIZATION table, not to site_information", async ({
    request,
  }) => {
    // The POST handler branches on the tag before it does anything else. For a
    // row tagged "localization" the site_information `value` column is a
    // localization ID, the content lives in localization_value, and
    // site_information is never touched — so a port that wrote the value column
    // would corrupt the pointer and lose the text.
    //
    // This is the one test that edits a SHIPPED row, because no created row can
    // carry the tag: the insert path never sets one. It restores what it found,
    // in an afterEach that runs even on failure.
    const [[siteId, localizationId, tag]] = query(
      `SELECT id::text, value, tag FROM clinlims.site_information WHERE name = 'bannerHeading'`,
    );
    expect(tag, "bannerHeading is the localization-tagged row").toBe(
      "localization",
    );

    const before = query(
      `SELECT locale, value FROM clinlims.localization_value
        WHERE localization_id = ${Number(localizationId)} ORDER BY locale`,
    );
    expect(before.length, "the localization has values").toBeGreaterThan(0);
    localizationRestore = { id: localizationId, values: before };

    const [[siteValueBefore, siteUpdatedBefore]] = query(
      `SELECT value, lastupdated::text FROM clinlims.site_information WHERE id = ${Number(siteId)}`,
    );

    const marker = "E2E LOCALIZATION PROBE";
    const res = await postConfig(request, `${SITE_INFORMATION}?ID=${siteId}`, {
      paramName: "bannerHeading",
      value: localizationId,
      siteInfoDomainName: "SiteInformation",
      valueType: "text",
      tag: "localization",
      localization: {
        id: localizationId,
        description: "",
        values: {
          en: { id: "0", locale: "en", value: marker },
          fr: { id: "0", locale: "fr", value: marker },
        },
      },
    });
    expect(res.status(), "the localization write").toBe(200);

    const after = query(
      `SELECT locale, value FROM clinlims.localization_value
        WHERE localization_id = ${Number(localizationId)} ORDER BY locale`,
    );
    expect(
      after.map((r) => r[1]),
      "every active locale now carries the submitted text",
    ).toEqual(after.map(() => marker));

    // ...and site_information is untouched, lastupdated included.
    const [[siteValueAfter, siteUpdatedAfter]] = query(
      `SELECT value, lastupdated::text FROM clinlims.site_information WHERE id = ${Number(siteId)}`,
    );
    expect(siteValueAfter, "the pointer column is unchanged").toBe(
      siteValueBefore,
    );
    expect(siteUpdatedAfter, "and the row was not written at all").toBe(
      siteUpdatedBefore,
    );
  });

  test("writing 'modify results role' toggles the ROLE, in another table", async ({
    request,
  }) => {
    // configurationSideEffects.siteInformationChanged: this setting IS the
    // "Results modifier" role's active flag. Storing the row and stopping would
    // leave the permission out of step with the screen that sets it.
    //
    // The Java code keys on Property.roleRequiredForModifyResults.getDBName(),
    // which is the string "modify results role" — not the constant's name. A
    // port that matched the constant would compile and never fire.
    const [[settingId, settingValue]] = query(
      `SELECT id::text, value FROM clinlims.site_information WHERE name = 'modify results role'`,
    );
    const [[roleActiveBefore]] = query(
      `SELECT active::text FROM clinlims.system_role WHERE trim(name) = 'Results modifier'`,
    );
    settingRestore = {
      id: settingId,
      value: settingValue,
      roleActive: roleActiveBefore,
    };

    // Flip it to whatever it is not.
    const next = settingValue === "true" ? "false" : "true";
    const res = await postConfig(
      request,
      `${SITE_INFORMATION}?ID=${settingId}`,
      {
        paramName: "modify results role",
        value: next,
        siteInfoDomainName: "SiteInformation",
        valueType: "boolean",
      },
    );
    expect(res.status(), "the setting write").toBe(200);

    const [[storedValue]] = query(
      `SELECT value FROM clinlims.site_information WHERE id = ${Number(settingId)}`,
    );
    expect(storedValue, "the setting is stored").toBe(next);

    const [[roleActiveAfter]] = query(
      `SELECT active::text FROM clinlims.system_role WHERE trim(name) = 'Results modifier'`,
    );
    expect(roleActiveAfter, "and the ROLE followed it").toBe(next);
    expect(roleActiveAfter, "which is a change, not a coincidence").not.toBe(
      roleActiveBefore,
    );
  });

  test("a write the database refuses is Tomcat's 500 page, not a form", async ({
    request,
  }) => {
    // site_information.name is varchar(32), so a longer one cannot be stored.
    // The controller LOOKS like it handles this — validateAndUpdateSiteInformation
    // catches LIMSRuntimeException, picks between errors.OptimisticLockException
    // and errors.UpdateException, and returns the form — but that path is not
    // the one taken: the failure surfaces at the transaction boundary and comes
    // back as the servlet error page. Measured, and the distinction matters,
    // because a 200-with-form and a 500 are different answers for a caller.
    const tooLong = "e2eNameThatIsDefinitelyLongerThanThirtyTwoChars";
    expect(tooLong.length, "longer than the column").toBeGreaterThan(32);

    const res = await postConfig(request, SITE_INFORMATION, {
      paramName: tooLong,
      value: "x",
      siteInfoDomainName: "SiteInformation",
      valueType: "text",
    });
    expect(res.status(), "the database refuses the row").toBe(500);

    const body = await res.json();
    expect(body.status, "Tomcat's error envelope, not plain text").toBe(500);
    expect(body.error).toBe("Internal Server Error");
    expect(typeof body.timestamp, "...with its epoch-millis timestamp").toBe(
      "number",
    );

    expect(
      query(
        `SELECT id FROM clinlims.site_information WHERE name LIKE 'e2eName%'`,
      ).length,
      "and nothing was written",
    ).toBe(0);
  });

  test("the configuration is a CACHE — a row changed outside the app is invisible", async ({
    request,
  }) => {
    // ConfigurationProperties is loaded into memory at startup and reloaded
    // only after a write through this application. A row changed by anything
    // else — a migration, a DBA, another service — is not picked up.
    //
    // This is the one behaviour where the obvious port is MORE correct than
    // Java and therefore wrong: reading the table per request answers the new
    // value, and Java answers the old one.
    const before = await readJson(
      await request.get(LAB_UNIT_CONFIG),
      LAB_UNIT_CONFIG,
    );
    const cached = before.accessionFormat;
    expect(cached, "the endpoint reports an accession format").toBeTruthy();

    const [[id, original]] = query(
      `SELECT id::text, value FROM clinlims.site_information WHERE name = 'acessionFormat'`,
    );
    expect(original, "and it matches the row, before anything changes").toBe(
      cached,
    );
    configRestore = { id, value: original };

    // Change it BEHIND the application's back.
    exec(
      `UPDATE clinlims.site_information SET value = 'E2E_CACHE_PROBE' WHERE id = ${Number(id)}`,
    );
    const [[nowStored]] = query(
      `SELECT value FROM clinlims.site_information WHERE id = ${Number(id)}`,
    );
    expect(nowStored, "the row really did change").toBe("E2E_CACHE_PROBE");

    const after = await readJson(
      await request.get(LAB_UNIT_CONFIG),
      LAB_UNIT_CONFIG,
    );
    expect(
      after.accessionFormat,
      "and the endpoint still answers the CACHED value",
    ).toBe(cached);
    expect(after.accessionFormat, "...not the stored one").not.toBe(
      "E2E_CACHE_PROBE",
    );
  });

  test("writing the default locale changes the language of the very next response", async ({
    request,
  }) => {
    // Two mechanisms, one visible effect: the write reloads
    // ConfigurationProperties, and localeResolver.setLocale puts the same
    // locale on the request. Either way the next response is rendered in the
    // new language.
    //
    // This needs a localization whose French text DIFFERS from its English
    // text, and the shipped data has none — bannerHeading reads "Test LIMS" in
    // both. So the test seeds the difference itself and puts it back, rather
    // than treating "no data" as a reason not to check.
    const [[localizationId]] = query(
      `SELECT value FROM clinlims.site_information WHERE name = 'bannerHeading'`,
    );
    const [[localeId, originalLocale]] = query(
      `SELECT id::text, value FROM clinlims.site_information WHERE name = 'default language locale'`,
    );
    const frBefore = query(
      `SELECT value FROM clinlims.localization_value
        WHERE localization_id = ${Number(localizationId)} AND locale = 'fr'`,
    );
    expect(frBefore.length, "the localization has a French row").toBe(1);

    localeRestore = {
      id: localeId,
      value: originalLocale,
      localizationId,
      fr: frBefore[0][0],
    };

    const french = "TEXTE FRANCAIS";
    exec(
      `UPDATE clinlims.localization_value SET value = '${french}'
        WHERE localization_id = ${Number(localizationId)} AND locale = 'fr'`,
    );

    const banner = async () => {
      const menu = await readJson(
        await request.get(SITE_INFORMATION_MENU),
        SITE_INFORMATION_MENU,
      );
      const row = menu.menuList.find((m: any) => m.name === "bannerHeading");
      expect(row, "bannerHeading is on the menu").toBeTruthy();
      return row.localization;
    };

    const before = await banner();
    expect(before.localizedValue, "English to start with").toBe(before.english);
    expect(
      before.localesAndValuesOfLocalesWithValues,
      "and the language NAMES are English too",
    ).toEqual(["English: " + before.english, "French: " + french]);

    // Flip the deployment's language.
    const flipped = await postConfig(
      request,
      `${SITE_INFORMATION}?ID=${localeId}`,
      {
        paramName: "default language locale",
        value: "fr-FR",
        siteInfoDomainName: "SiteInformation",
        valueType: "dictionary",
      },
    );
    expect(flipped.status(), "the locale write").toBe(200);

    const after = await banner();
    expect(
      after.localizedValue,
      "the localized value is now the French text",
    ).toBe(french);
    expect(
      after.localesAndValuesOfLocalesWithValues,
      "and the language names are rendered IN French",
    ).toEqual(["anglais: " + before.english, "français: " + french]);

    // INVERSION: the stored values did not change, only the language they are
    // read in — so this is a locale switch and not a write to the localization.
    expect(after.english, "the English text is untouched").toBe(before.english);
    expect(after.french, "and so is the French").toBe(french);

    // Put the language back HERE, and prove it took, rather than leaving it to
    // cleanup. GlobalLocaleResolver keeps `currentLocale` in a field on the
    // resolver — one field, for the whole process — so while this test runs the
    // ENTIRE deployment is in French, and a restore that silently failed would
    // hand every later test a French server. That is not hypothetical: it is
    // how this test's first version broke four unrelated specs.
    const restored = await postConfig(
      request,
      `${SITE_INFORMATION}?ID=${localeId}`,
      {
        paramName: "default language locale",
        value: originalLocale,
        siteInfoDomainName: "SiteInformation",
        valueType: "dictionary",
      },
    );
    expect(restored.status(), "the restore write").toBe(200);
    const back = await banner();
    expect(back.localizedValue, "and the language really is back").toBe(
      before.english,
    );
  });

  test("Cancel is a GET and answers a bare JSON string", async ({
    request,
  }) => {
    const res = await request.get(CANCEL);
    expect(res.status(), "cancel").toBe(200);
    expect(await res.json(), "the body is a string, not an object").toBe(
      "Cancellation successful",
    );
  });
});
