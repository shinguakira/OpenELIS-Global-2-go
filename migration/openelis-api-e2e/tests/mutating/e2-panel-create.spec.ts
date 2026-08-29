/**
 * e2 — PanelCreate.
 *
 * The fourth *Create screen, and the one that breaks the pattern the other
 * three share. It writes NINE rows, not eight: a panel arrives already tied to
 * a sample type, through a sampletype_panel row. Its system_module descriptions
 * are spelled differently too — `Workplan=>panel=><name>` where the others
 * build `Workplan=><name>` — because this controller carries its own copy of
 * createSystemModule.
 *
 * Its GET is the odd one as well: existingPanelList is not a list of id/value
 * pairs but a list of sample-type names with whole Panel ENTITIES nested under
 * them, Localization and all.
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

const EN = "e2pnProbe";
const FR = "e2pnProbeFr";

async function post(
  request: APIRequestContext,
  path: string,
  data: Record<string, unknown>,
) {
  const body = await readJson(await request.get(SESSION_PATH), SESSION_PATH);
  return request.post(path, {
    headers: {
      [CSRF_HEADER]: body[CSRF_SESSION_FIELD],
      "Content-Type": "application/json",
    },
    data,
  });
}

function cleanUp() {
  exec(
    `DELETE FROM clinlims.system_role_module WHERE system_module_id IN
       (SELECT id FROM clinlims.system_module WHERE name LIKE '%${EN}%')`,
  );
  exec(`DELETE FROM clinlims.system_module WHERE name LIKE '%${EN}%'`);
  exec(
    `DELETE FROM clinlims.sampletype_panel WHERE panel_id IN
       (SELECT id FROM clinlims.panel WHERE name = '${EN}')`,
  );
  exec(
    `DELETE FROM clinlims.history WHERE reference_table =
       (SELECT id FROM clinlims.reference_tables WHERE name = 'PANEL' LIMIT 1)
       AND reference_id IN (SELECT id FROM clinlims.panel WHERE name = '${EN}')`,
  );
  const locs = query(
    `SELECT name_localization_id::text FROM clinlims.panel WHERE name = '${EN}'`,
  ).map((r) => r[0]);
  exec(
    `DELETE FROM clinlims.localization_value WHERE localization_id IN
       (SELECT name_localization_id FROM clinlims.panel WHERE name = '${EN}')`,
  );
  exec(`DELETE FROM clinlims.panel WHERE name = '${EN}'`);
  for (const id of locs) {
    exec(`DELETE FROM clinlims.localization WHERE id = ${Number(id)}`);
  }
}

test.describe("e2 — PanelCreate writes nine rows", () => {
  test.beforeEach(() => cleanUp());
  test.afterEach(() => cleanUp());

  test("GET PanelCreate — sample types with whole Panel entities nested", async ({
    request,
  }) => {
    const f = await readJson(await request.get("rest/PanelCreate"), "PanelCreate");

    expect(f.formName, "formName").toBe("panelCreateForm");
    expect(Array.isArray(f.existingSampleTypeList), "existingSampleTypeList").toBe(
      true,
    );
    expect(f.existingPanelList.length, "one entry per ACTIVE sample type").toBe(
      f.existingSampleTypeList.length,
    );
    expect(f.inactivePanelList.length, "and the same for the inactive list").toBe(
      f.existingSampleTypeList.length,
    );

    // Absent and empty are DIFFERENT here. createTypeOfSamplePanelMap creates
    // the map entry as soon as a sampletype_panel row for that sample type is
    // seen, before the active filter runs — so a sample type with join rows
    // whose panels all fail the filter keeps `panels: []`, and one with no join
    // rows at all has no `panels` key at all.
    const withKey = f.existingPanelList.filter((e: any) => e.panels !== undefined);
    expect(
      withKey.length,
      "only the sample types with join rows carry the key",
    ).toBeGreaterThan(0);
    expect(withKey.length, "and fewer than every sample type").toBeLessThan(
      f.existingPanelList.length,
    );
    expect(
      f.inactivePanelList.filter((e: any) => e.panels !== undefined).length,
      "the same sample types carry it in the inactive list",
    ).toBe(withKey.length);

    // The nested panels are entities, not projections.
    const panel = withKey
      .map((e: any) => e.panels)
      .flat()
      .find(Boolean);
    expect(panel, "at least one panel is nested").toBeTruthy();
    for (const key of [
      "id",
      "panelName",
      "description",
      "isActive",
      "sortOrderInt",
      "localization",
    ]) {
      expect(panel[key] !== undefined, `the panel entity carries ${key}`).toBe(
        true,
      );
    }
    expect(
      panel.localization.values,
      "and its Localization is the full object, not an id",
    ).toBeTruthy();

    for (const key of ["existingEnglishNames", "existingFrenchNames"]) {
      expect(f[key].startsWith("$"), `${key} leading separator`).toBe(true);
      expect(f[key].endsWith("$"), `${key} trailing separator`).toBe(true);
    }
  });

  test("POST PanelCreate — nine rows, and a different module description", async ({
    request,
  }) => {
    const f = await readJson(await request.get("rest/PanelCreate"), "PanelCreate");
    const sampleTypeId = f.existingSampleTypeList[0].id;

    const count = (sql: string) => Number(query(sql)[0][0]);
    const before = {
      panel: count(`SELECT count(*)::text FROM clinlims.panel`),
      join: count(`SELECT count(*)::text FROM clinlims.sampletype_panel`),
      module: count(`SELECT count(*)::text FROM clinlims.system_module`),
    };

    const res = await post(request, "rest/PanelCreate", {
      panelEnglishName: EN,
      panelFrenchName: FR,
      sampleTypeId,
      // Sent on purpose. panelLoinc is NOT in ALLOWED_FIELDS, and it is stored
      // anyway: initBinder.setAllowedFields governs FORM binding, and these
      // endpoints take a JSON @RequestBody, which Jackson maps. The allow-list
      // is dead configuration on every REST controller in this package.
      panelLoinc: "12345-6",
    });
    expect(res.status(), "create").toBe(200);

    expect(
      count(`SELECT count(*)::text FROM clinlims.panel`) - before.panel,
      "one panel",
    ).toBe(1);
    expect(
      count(`SELECT count(*)::text FROM clinlims.sampletype_panel`) - before.join,
      "the ninth row — the sample type link the other creates do not write",
    ).toBe(1);
    expect(
      count(`SELECT count(*)::text FROM clinlims.system_module`) - before.module,
      "three modules",
    ).toBe(3);

    const [[panelId, isActive, sortOrder, loinc]] = query(
      `SELECT id::text, is_active, sort_order::text, COALESCE(loinc, '<null>')
         FROM clinlims.panel WHERE name = '${EN}'`,
    );
    expect(isActive, "a created panel starts inactive").toBe("N");
    expect(sortOrder, "and sorts last").toBe("2147483647");
    expect(
      loinc,
      "panelLoinc IS stored — setAllowedFields does not reach a JSON body",
    ).toBe("12345-6");

    expect(
      query(
        `SELECT sample_type_id::text FROM clinlims.sampletype_panel WHERE panel_id = ${Number(panelId)}`,
      ),
      "the link points at the submitted sample type",
    ).toEqual([[sampleTypeId]]);

    // The description this controller builds is NOT the one the other three do.
    expect(
      query(
        `SELECT name, COALESCE(description, '') FROM clinlims.system_module
          WHERE name LIKE '%${EN}%' ORDER BY id`,
      ),
      "Workplan=>panel=><name>, not Workplan=><name>",
    ).toEqual([
      [`Workplan:${EN}`, `Workplan=>panel=>${EN}`],
      [`LogbookResults:${EN}`, `LogbookResults=>panel=>${EN}`],
      [`ResultValidation:${EN}`, `ResultValidation=>panel=>${EN}`],
    ]);

    const audit = query(
      `SELECT h.activity, COALESCE(convert_from(h.changes, 'UTF8'), '<null>')
         FROM clinlims.history h
         JOIN clinlims.reference_tables rt ON rt.id = h.reference_table
        WHERE upper(rt.name) = 'PANEL' AND h.reference_id = ${Number(panelId)}`,
    );
    expect(audit.length, "the panel is audited").toBe(1);
    expect(audit[0], "as an insert with no payload").toEqual(["I", "<null>"]);
  });
});
