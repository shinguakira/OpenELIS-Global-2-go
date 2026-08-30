/**
 * e2 — the *Create screens for Method, TestSection and SampleType.
 *
 * The entity is the small part of what these write. Each create is EIGHT rows
 * across six tables, in one transaction:
 *
 *   localization        1   description is the screen's own label
 *   localization_value  2   'en' and 'fr', from the two submitted names
 *   <entity>            1   name_localization_id points at the localization
 *   system_module       3   Workplan, LogbookResults, ResultValidation
 *   system_role_module  3   two to the Results role, one to Validation
 *
 * and exactly ONE history row — for the entity, with a NULL payload. The
 * localization, the modules and the role links are not audited, though
 * LOCALIZATION is flagged keep_history = 'Y'. That is the third table in this
 * wave whose audit mechanism is simply switched off.
 *
 * A create therefore hands a new lab unit its permissions as well as its name,
 * which is why the transaction boundary matters: a half-made test section is
 * one the results screen cannot be granted on.
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

const EN = "e2crProbe";
const FR = "e2crProbeFr";

async function csrfToken(request: APIRequestContext): Promise<string> {
  const body = await readJson(await request.get(SESSION_PATH), SESSION_PATH);
  return body[CSRF_SESSION_FIELD];
}

async function post(
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

const SCREENS = [
  {
    path: "MethodCreate",
    formName: "methodCreateForm",
    listKey: "existingMethodList",
    inactiveKey: "inactiveMethodList",
    body: { methodEnglishName: EN, methodFrenchName: FR },
    table: "clinlims.method",
    nameColumn: "name",
    auditTable: "METHOD",
    localizationLabel: "method name",
    // is_active is the CHAR here, and a created method starts INACTIVE.
    activeCheck: `is_active = 'N'`,
  },
  {
    path: "TestSectionCreate",
    formName: "testSectionCreateForm",
    listKey: "existingTestUnitList",
    inactiveKey: "inactiveTestUnitList",
    body: { testUnitEnglishName: EN, testUnitFrenchName: FR },
    table: "clinlims.test_section",
    nameColumn: "name",
    auditTable: "TEST_SECTION",
    localizationLabel: "test unit name",
    activeCheck: `is_active = 'N' AND sort_order = 2147483647`,
  },
  {
    path: "SampleTypeCreate",
    formName: "sampleTypeCreateForm",
    listKey: "existingSampleTypeList",
    inactiveKey: "inactiveSampleTypeList",
    body: { sampleTypeEnglishName: EN, sampleTypeFrenchName: FR },
    table: "clinlims.type_of_sample",
    // No `name` column — createTypeOfSample calls setDescription only.
    nameColumn: "description",
    auditTable: "TYPE_OF_SAMPLE",
    localizationLabel: "type of sample name",
    // is_active here IS a real boolean, unlike the two above.
    activeCheck: `is_active = false AND domain = 'H' AND sort_order = 2147483647 AND local_abbrev = '${EN}'`,
  },
] as const;

/** Everything a probe create leaves behind, removed FK-first. */
function cleanUp() {
  exec(
    `DELETE FROM clinlims.system_role_module WHERE system_module_id IN
       (SELECT id FROM clinlims.system_module WHERE name LIKE '%${EN}%')`,
  );
  exec(`DELETE FROM clinlims.system_module WHERE name LIKE '%${EN}%'`);
  for (const s of SCREENS) {
    exec(
      `DELETE FROM clinlims.history
        WHERE reference_table = (SELECT id FROM clinlims.reference_tables WHERE name = '${s.auditTable}' LIMIT 1)
          AND reference_id IN (SELECT id FROM ${s.table} WHERE ${s.nameColumn} = '${EN}')`,
    );
    exec(
      `DELETE FROM clinlims.localization_value WHERE localization_id IN
         (SELECT name_localization_id FROM ${s.table} WHERE ${s.nameColumn} = '${EN}')`,
    );
  }
  const orphans: string[] = [];
  for (const s of SCREENS) {
    for (const [id] of query(
      `SELECT name_localization_id::text FROM ${s.table} WHERE ${s.nameColumn} = '${EN}' AND name_localization_id IS NOT NULL`,
    )) {
      orphans.push(id);
    }
  }
  for (const s of SCREENS) {
    exec(`DELETE FROM ${s.table} WHERE ${s.nameColumn} = '${EN}'`);
  }
  for (const id of orphans) {
    exec(`DELETE FROM clinlims.localization WHERE id = ${Number(id)}`);
  }
}

test.describe("e2 — a Create writes eight rows and audits one", () => {
  test.beforeEach(() => cleanUp());
  test.afterEach(() => cleanUp());

  for (const s of SCREENS) {
    test(`GET ${s.path} — the form`, async ({ request }) => {
      const form = await readJson(await request.get(`rest/${s.path}`), s.path);

      expect(form.formName, "formName").toBe(s.formName);
      expect(form.formMethod, "formMethod").toBe("POST");
      expect(form.cancelAction, "cancelAction").toBe("Home");
      expect(form.submitOnCancel, "submitOnCancel").toBe(false);
      expect(form.cancelMethod, "cancelMethod").toBe("POST");

      for (const key of [s.listKey, s.inactiveKey]) {
        expect(Array.isArray(form[key]), `${key} is a list`).toBe(true);
      }
      expect(form[s.listKey].length, "the active list is populated")
        .toBeGreaterThan(0);

      // Both name strings are seeded with the separator and append one after
      // every name, so each carries a leading AND a trailing "$".
      for (const key of ["existingEnglishNames", "existingFrenchNames"]) {
        expect(typeof form[key], `${key} is a string`).toBe("string");
        expect(form[key].startsWith("$"), `${key} leading separator`).toBe(true);
        expect(form[key].endsWith("$"), `${key} trailing separator`).toBe(true);
      }
      // Unlike UOM, whose French string is the literal word "French", these
      // entities have real localization rows.
      expect(
        form.existingFrenchNames.split("$").filter(Boolean).every((v: string) => v === "French"),
        "the French names are not a hardcoded literal here",
      ).toBe(false);
    });

    test(`POST ${s.path} — eight rows, one audit row`, async ({ request }) => {
      const count = (sql: string) => Number(query(sql)[0][0]);
      const before = {
        localization: count(`SELECT count(*)::text FROM clinlims.localization`),
        localizationValue: count(`SELECT count(*)::text FROM clinlims.localization_value`),
        entity: count(`SELECT count(*)::text FROM ${s.table}`),
        module: count(`SELECT count(*)::text FROM clinlims.system_module`),
        roleModule: count(`SELECT count(*)::text FROM clinlims.system_role_module`),
      };

      const res = await post(request, `rest/${s.path}`, s.body);
      expect(res.status(), "create").toBe(200);

      // The success branch returns the bound form WITHOUT setupDisplayItems, so
      // the lists and the name strings are absent.
      const body = await res.json();
      expect(body[s.listKey], "the lists are absent from the POST response").toBe(
        undefined,
      );
      expect(body.existingEnglishNames, "and so are the name strings").toBe(
        undefined,
      );

      expect(
        count(`SELECT count(*)::text FROM clinlims.localization`) - before.localization,
        "one localization",
      ).toBe(1);
      expect(
        count(`SELECT count(*)::text FROM clinlims.localization_value`) - before.localizationValue,
        "two localization values — 'en' and 'fr', and no other active locale",
      ).toBe(2);
      expect(
        count(`SELECT count(*)::text FROM ${s.table}`) - before.entity,
        "one entity",
      ).toBe(1);
      expect(
        count(`SELECT count(*)::text FROM clinlims.system_module`) - before.module,
        "three system modules",
      ).toBe(3);
      expect(
        count(`SELECT count(*)::text FROM clinlims.system_role_module`) - before.roleModule,
        "three role links",
      ).toBe(3);

      // The entity, and the shape createX gives it.
      const rows = query(
        `SELECT id::text FROM ${s.table} WHERE ${s.nameColumn} = '${EN}' AND ${s.activeCheck}`,
      );
      expect(rows.length, "the entity carries the flags its create sets").toBe(1);
      const entityId = rows[0][0];

      // The localization it points at, and the label the screen gave it.
      const [[label]] = query(
        `SELECT COALESCE(l.description, '') FROM clinlims.localization l
           JOIN ${s.table} e ON e.name_localization_id = l.id WHERE e.id = ${Number(entityId)}`,
      );
      expect(label, "the localization's description is the screen's label").toBe(
        s.localizationLabel,
      );

      const values = query(
        `SELECT lv.locale, lv.value FROM clinlims.localization_value lv
           JOIN ${s.table} e ON e.name_localization_id = lv.localization_id
          WHERE e.id = ${Number(entityId)} ORDER BY lv.locale`,
      );
      expect(values, "'en' takes the English name and 'fr' the French one").toEqual(
        [["en", EN], ["fr", FR]],
      );

      // The three modules, named by SystemModule's own convention.
      const modules = query(
        `SELECT name, COALESCE(description, '') FROM clinlims.system_module
          WHERE name LIKE '%${EN}%' ORDER BY id`,
      );
      expect(modules, "Workplan, LogbookResults, ResultValidation").toEqual([
        [`Workplan:${EN}`, `Workplan=>${EN}`],
        [`LogbookResults:${EN}`, `LogbookResults=>${EN}`],
        [`ResultValidation:${EN}`, `ResultValidation=>${EN}`],
      ]);

      // Two to Results, one to Validation, all four permissions granted.
      const links = query(
        `SELECT trim(r.name), srm.has_select, srm.has_add, srm.has_update, srm.has_delete
           FROM clinlims.system_role_module srm
           JOIN clinlims.system_role r ON r.id = srm.system_role_id
           JOIN clinlims.system_module m ON m.id = srm.system_module_id
          WHERE m.name LIKE '%${EN}%' ORDER BY m.id`,
      );
      expect(links, "the role links the create grants").toEqual([
        ["Results", "Y", "Y", "Y", "Y"],
        ["Results", "Y", "Y", "Y", "Y"],
        ["Validation", "Y", "Y", "Y", "Y"],
      ]);

      // ONE history row, for the entity, with a NULL payload.
      const audit = query(
        `SELECT h.activity, h.sys_user_id::text, COALESCE(convert_from(h.changes, 'UTF8'), '<null>')
           FROM clinlims.history h
           JOIN clinlims.reference_tables rt ON rt.id = h.reference_table
          WHERE upper(rt.name) = '${s.auditTable}' AND h.reference_id = ${Number(entityId)}`,
      );
      expect(audit.length, "the entity is audited").toBe(1);
      expect(audit[0][0], "recorded as an insert").toBe("I");
      expect(audit[0][2], "with no payload — saveNewHistory sets none").toBe(
        "<null>",
      );

      // And nothing else is.
      const [[locId]] = query(
        `SELECT name_localization_id::text FROM ${s.table} WHERE id = ${Number(entityId)}`,
      );
      expect(
        query(
          `SELECT count(*)::text FROM clinlims.history h
             JOIN clinlims.reference_tables rt ON rt.id = h.reference_table
            WHERE upper(rt.name) = 'LOCALIZATION' AND h.reference_id = ${Number(locId)}`,
        )[0][0],
        "the localization is NOT audited, despite keep_history = Y",
      ).toBe("0");
    });
  }
});
