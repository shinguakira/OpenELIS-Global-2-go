/**
 * e1 follow-ups — three parity gaps the first pass left behind.
 *
 * All three are invisible in the response body. Every one of them is a write
 * the port either makes differently from Java or does not make at all, so the
 * only oracle is the database: site_information for the value, clinlims.history
 * for the audit trail, system_role for the side effect.
 *
 * They run against BOTH targets, like every spec here. Against Java they are
 * the measurement; against Go they are the check.
 */
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
const DELETE_SITE_INFORMATION = "rest/DeleteSiteInformation";

/** A name this spec owns, distinct from the one e1-config-crud uses. */
const PROBE = "e2eGapProbe";

/** The shipped row whose write toggles a role in another table. */
const ROLE_SETTING = "modify results role";
const MODIFIER_ROLE = "Results modifier";

function siteInfoTableId(): string {
  const rows = query(
    `SELECT id FROM clinlims.reference_tables WHERE name = 'site_information'`,
  );
  expect(rows.length, "reference_tables has a site_information row").toBe(1);
  return rows[0][0];
}

/**
 * The reference_tables ids roles are audited under.
 *
 * PLURAL deliberately: this deployment ships THREE rows named SYSTEM_ROLE
 * (172, 174, 177), all with keep_history = 'Y'. Which one a write lands on is
 * decided by whatever the lookup returns, so the assertion accepts any of them
 * rather than pinning an id the data does not make unique.
 */
function roleTableIds(): string[] {
  const rows = query(
    `SELECT id FROM clinlims.reference_tables WHERE upper(name) = 'SYSTEM_ROLE'`,
  );
  expect(rows.length, "reference_tables has at least one SYSTEM_ROLE row")
    .toBeGreaterThan(0);
  return rows.map((r) => r[0]);
}

async function csrfToken(request: APIRequestContext): Promise<string> {
  const body = await readJson(await request.get(SESSION_PATH), SESSION_PATH);
  return body[CSRF_SESSION_FIELD];
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

/** The audit payloads written for one site_information id, oldest first. */
function auditRows(
  id: string,
): { activity: string; user: string; changes: string }[] {
  return query(
    `SELECT h.activity, h.sys_user_id::text,
            replace(COALESCE(convert_from(h.changes, 'UTF8'), ''), chr(10), ' ')
       FROM clinlims.history h
      WHERE h.reference_table = ${siteInfoTableId()}
        AND h.reference_id = ${id}
      ORDER BY h.timestamp, h.id`,
  ).map(([activity, user, changes]) => ({ activity, user, changes }));
}

function probeIds(): string[] {
  return query(
    `SELECT id::text FROM clinlims.site_information WHERE name = '${PROBE}'`,
  ).map((r) => r[0]);
}

test.describe("e1 follow-ups — writes that only the database shows", () => {
  let touched: string[] = [];
  let roleRestore: { id: string; value: string; roleActive: string } | null =
    null;

  test.beforeEach(() => {
    touched = probeIds();
    roleRestore = null;
  });

  test.afterEach(() => {
    for (const id of [...new Set([...touched, ...probeIds()])]) {
      exec(
        `DELETE FROM clinlims.history WHERE reference_table = ${siteInfoTableId()} AND reference_id = ${id}`,
      );
    }
    exec(`DELETE FROM clinlims.site_information WHERE name = '${PROBE}'`);

    if (roleRestore) {
      exec(
        `UPDATE clinlims.site_information SET value = '${roleRestore.value}'
          WHERE id = ${Number(roleRestore.id)}`,
      );
      exec(
        `UPDATE clinlims.system_role SET active = ${roleRestore.roleActive}
          WHERE trim(name) = '${MODIFIER_ROLE}'`,
      );
      exec(
        `DELETE FROM clinlims.history WHERE reference_table = ${siteInfoTableId()}
          AND reference_id = ${Number(roleRestore.id)}`,
      );
      for (const table of roleTableIds()) {
        exec(
          `DELETE FROM clinlims.history WHERE reference_table = ${table}
            AND reference_id IN (SELECT id FROM clinlims.system_role WHERE trim(name) = '${MODIFIER_ROLE}')`,
        );
      }
    }
  });

  // ──────────────────────────────────────────────────────────────────────────
  // P1-a
  // ──────────────────────────────────────────────────────────────────────────
  test("an existing row encrypts by its STORED flag, not the submitted one", async ({
    request,
  }) => {
    // Java builds the entity two different ways. A new row takes `encrypted`
    // off the form; an existing one is LOADED by id and only setValue is
    // called, so encryptSiteInformation reads the flag the column already
    // holds. The submitted flag is not consulted on an update at all.
    const created = await postConfig(request, SITE_INFORMATION, {
      paramName: PROBE,
      value: "first-secret",
      encrypted: true,
      siteInfoDomainName: "SiteInformation",
      valueType: "text",
    });
    expect(created.status(), "create").toBe(200);

    const ids = probeIds();
    expect(ids.length, "one probe row was created").toBe(1);
    const id = ids[0];
    touched.push(id);

    const [[storedFlag, firstColumn]] = query(
      `SELECT encrypted::text, COALESCE(value, '') FROM clinlims.site_information WHERE id = ${id}`,
    );
    expect(storedFlag, "the row is flagged encrypted").toBe("true");
    expect(firstColumn, "and the column holds ciphertext, not the plaintext")
      .not.toBe("first-secret");

    // The update omits `encrypted` entirely. The bean default is false, so a
    // port that reads the submission would store PLAINTEXT here — and every
    // later read would try to decrypt it, because the column's own flag is
    // still true.
    const updated = await postConfig(request, `${SITE_INFORMATION}?ID=${id}`, {
      paramName: PROBE,
      value: "second-secret",
      siteInfoDomainName: "SiteInformation",
      valueType: "text",
    });
    expect(updated.status(), "update").toBe(200);

    const [[flagAfter, columnAfter]] = query(
      `SELECT encrypted::text, COALESCE(value, '') FROM clinlims.site_information WHERE id = ${id}`,
    );
    expect(flagAfter, "the update does not touch the encrypted column").toBe(
      "true",
    );
    expect(
      columnAfter,
      "and the new value is stored as ciphertext, because the ROW says encrypted",
    ).not.toBe("second-secret");
    expect(columnAfter, "the value did change").not.toBe(firstColumn);

    // The round trip proves the ciphertext is readable: the form decrypts it.
    const form = await readJson(
      await request.get(`${SITE_INFORMATION}?ID=${id}`),
      SITE_INFORMATION,
    );
    expect(form.value, "the form shows the plaintext back").toBe(
      "second-secret",
    );
  });

  test("submitting encrypted=true does not encrypt an unencrypted row", async ({
    request,
  }) => {
    // The other direction of the same rule. Storing ciphertext while the
    // column's flag stays false makes the value unreadable: nothing will ever
    // decrypt it again, because no reader looks at a row that is not flagged.
    const created = await postConfig(request, SITE_INFORMATION, {
      paramName: PROBE,
      value: "plain-one",
      encrypted: false,
      siteInfoDomainName: "SiteInformation",
      valueType: "text",
    });
    expect(created.status(), "create").toBe(200);

    const ids = probeIds();
    expect(ids.length, "one probe row was created").toBe(1);
    const id = ids[0];
    touched.push(id);

    const updated = await postConfig(request, `${SITE_INFORMATION}?ID=${id}`, {
      paramName: PROBE,
      value: "plain-two",
      encrypted: true, // ignored on an existing row
      siteInfoDomainName: "SiteInformation",
      valueType: "text",
    });
    expect(updated.status(), "update").toBe(200);

    const [[flagAfter, columnAfter]] = query(
      `SELECT encrypted::text, COALESCE(value, '') FROM clinlims.site_information WHERE id = ${id}`,
    );
    expect(flagAfter, "the flag is still false").toBe("false");
    expect(columnAfter, "and the column holds the plaintext it was sent").toBe(
      "plain-two",
    );
  });

  // ──────────────────────────────────────────────────────────────────────────
  // P1-c
  // ──────────────────────────────────────────────────────────────────────────
  test("toggling the modify-results role writes NO audit row for the role", async ({
    request,
  }) => {
    // This one reads backwards, so it is worth stating why.
    //
    // Everything on paper says the role change is audited. It goes through
    // roleService.update; RoleServiceImpl sets auditTrailLog = true; and
    // reference_tables ships THREE rows named SYSTEM_ROLE, all keep_history
    // = 'Y'. A review that stopped there would report the port's bare UPDATE
    // as a lost audit event.
    //
    // MEASURED, it is not. Toggling the setting through Java flips
    // system_role.active and leaves exactly ONE history row — the
    // site_information update — and nothing under any SYSTEM_ROLE table id.
    // The audit lookup resolves the Role entity to a table name that does not
    // match those rows, so the write proceeds unrecorded.
    //
    // So the port's direct UPDATE is the faithful behaviour, and this test
    // exists to keep it that way: adding an audit row here would be a port
    // that is MORE correct than Java, which is the one thing this project
    // treats as a defect.
    const [[settingId, settingValue]] = query(
      `SELECT id::text, COALESCE(value, '') FROM clinlims.site_information WHERE name = '${ROLE_SETTING}'`,
    );
    const [[roleId, roleActive]] = query(
      `SELECT id::text, active::text FROM clinlims.system_role WHERE trim(name) = '${MODIFIER_ROLE}'`,
    );
    roleRestore = { id: settingId, value: settingValue, roleActive };

    const next = roleActive === "true" ? "false" : "true";
    const posted = await postConfig(
      request,
      `${SITE_INFORMATION}?ID=${settingId}`,
      {
        paramName: ROLE_SETTING,
        value: next,
        siteInfoDomainName: "SiteInformation",
        valueType: "text",
      },
    );
    expect(posted.status(), "the setting was written").toBe(200);

    const [[activeAfter]] = query(
      `SELECT active::text FROM clinlims.system_role WHERE id = ${roleId}`,
    );
    expect(activeAfter, "the role followed the setting").toBe(next);

    const roleHistory = query(
      `SELECT h.activity, h.sys_user_id::text,
              replace(COALESCE(convert_from(h.changes, 'UTF8'), ''), chr(10), ' ')
         FROM clinlims.history h
        WHERE h.reference_table IN (${roleTableIds().join(",")})
          AND h.reference_id = ${roleId}
        ORDER BY h.timestamp, h.id`,
    );
    expect(
      roleHistory.length,
      "an authorization-bearing row changed and Java recorded nothing",
    ).toBe(0);

    // The configuration write itself IS audited, so the absence above is the
    // role's, not the request's — the mechanism ran, it just did not reach
    // system_role.
    const settingHistory = auditRows(settingId);
    expect(
      settingHistory.length,
      "the site_information update is in the trail",
    ).toBe(1);
    expect(settingHistory[0].activity, "recorded as an update").toBe("U");
    expect(
      settingHistory[0].changes,
      "carrying the value it replaced",
    ).toBe(`<value>${settingValue}</value>`);
  });

  // ──────────────────────────────────────────────────────────────────────────
  // P2-b
  // ──────────────────────────────────────────────────────────────────────────
  test("a delete records the optional columns the row actually carries", async ({
    request,
  }) => {
    // getChanges walks every field of the entity and emits the ones that differ
    // from a blank object, so a populated tag / instruction_key /
    // description_key IS part of the payload. The insert path cannot set those
    // columns, which is why this row is seeded in SQL — but the DELETE endpoint
    // takes arbitrary selected ids, and the shipped bannerHeading row (82)
    // carries all three.
    exec(
      `INSERT INTO clinlims.site_information
              (id, name, description, value, encrypted, domain_id, value_type,
               "group", tag, instruction_key, description_key, lastupdated)
       VALUES (nextval('clinlims.site_information_seq'), '${PROBE}', 'gap probe',
               'probe-value', false, 1, 'text', 0, 'localization',
               'instructions.${PROBE}', 'siteInfo.${PROBE}', now())`,
    );
    const ids = probeIds();
    expect(ids.length, "the seeded row is there").toBe(1);
    const id = ids[0];
    touched.push(id);

    const deleted = await request.get(DELETE_SITE_INFORMATION, {
      headers: {
        [CSRF_HEADER]: await csrfToken(request),
        "Content-Type": "application/json",
      },
      data: { selectedIDs: [id], siteInfoDomainName: "SiteInformation" },
    });
    expect(deleted.status(), "delete").toBe(200);
    expect(
      query(`SELECT id FROM clinlims.site_information WHERE id = ${id}`).length,
      "the row is gone",
    ).toBe(0);

    const rows = auditRows(id);
    expect(rows.length, "one audit row for the delete").toBe(1);
    const { activity, changes } = rows[0];
    expect(activity, "recorded as a delete").toBe("D");

    // The three the fixed field list dropped.
    expect(changes, "the tag it carried").toContain(
      "<tag>localization</tag>",
    );
    expect(changes, "the instruction key it carried").toContain(
      `<instructionKey>instructions.${PROBE}</instructionKey>`,
    );
    expect(changes, "the description key it carried").toContain(
      `<descriptionKey>siteInfo.${PROBE}</descriptionKey>`,
    );

    // Order is the wire contract: getChanges walks the entity's DECLARED
    // fields, so instructionKey sits between valueType and domain, tag after
    // group, and descriptionKey last before the superclass lastupdated.
    const order = [
      "name",
      "description",
      "value",
      "encrypted",
      "valueType",
      "instructionKey",
      "domain",
      "group",
      "tag",
      "schedule",
      "descriptionKey",
      "lastupdated",
    ];
    const seen = [...changes.matchAll(/<([a-zA-Z]+)>/g)].map((m) => m[1]);
    expect(seen, "the payload's field order").toEqual(order);
  });
});
