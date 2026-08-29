/**
 * e2 — the *RenameEntry screens for Method, Panel, SampleType and TestSection.
 *
 * Four Java controllers with one body between them, and the body does something
 * the screen name hides: it does NOT rename the entity. It loads the entity's
 * Localization and writes English and French onto that, leaving `method.name`,
 * `panel.name`, `type_of_sample.description` and `test_section.name` exactly as
 * they were. Every list renders the localization and falls back to the column
 * only when there is no localization row, so the rename is visible everywhere
 * and invisible in the table it appears to be about.
 *
 * UomRenameEntry is the one that really does write the column, because
 * unit_of_measure has no localization — see e2-uom-writes.spec.ts.
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

/** The four screens, and everything that differs between them. */
const SCREENS = [
  {
    path: "MethodRenameEntry",
    formName: "methodRenameEntryForm",
    listKey: "methodList",
    idKey: "methodId",
    table: "clinlims.method",
    nameColumn: "name",
    // The ONLY one of the four that answers 500 for an unknown id. Its lookup is
    // the generic methodService.get(id), and BaseObjectServiceImpl.get is
    // `getBaseObjectDAO().get(id).orElseThrow(ObjectNotFoundException)` — it
    // never returns null. The `if (method != null)` guard around the write is
    // therefore DEAD CODE, and the same guard in the other three works only
    // because they call an entity-specific finder that really does return null.
    unknownIdStatus: 500,
    postIncludesList: false,
  },
  {
    path: "PanelRenameEntry",
    formName: "panelRenameEntryForm",
    listKey: "panelList",
    idKey: "panelId",
    table: "clinlims.panel",
    nameColumn: "name",
    unknownIdStatus: 200,
    postIncludesList: false,
  },
  {
    path: "SampleTypeRenameEntry",
    formName: "sampleTypeRenameEntryForm",
    listKey: "sampleTypeList",
    idKey: "sampleTypeId",
    table: "clinlims.type_of_sample",
    nameColumn: "description",
    unknownIdStatus: 200,
    // The only one that re-populates its list on the SUCCESS path — it calls
    // setSampleTypeList again after the write, where the other three return the
    // bound form untouched.
    postIncludesList: true,
  },
  {
    path: "TestSectionRenameEntry",
    formName: "testSectionRenameEntryForm",
    listKey: "testSectionList",
    idKey: "testSectionId",
    table: "clinlims.test_section",
    nameColumn: "name",
    unknownIdStatus: 200,
    postIncludesList: false,
  },
] as const;

/** The localization values a row points at, per locale. */
function localizationOf(
  table: string,
  id: string,
): { locale: string; value: string }[] {
  return query(
    `SELECT lv.locale, COALESCE(lv.value, '')
       FROM ${table} e
       JOIN clinlims.localization_value lv ON lv.localization_id = e.name_localization_id
      WHERE e.id = ${Number(id)}
      ORDER BY lv.locale`,
  ).map(([locale, value]) => ({ locale, value }));
}

function nameColumnOf(table: string, column: string, id: string): string {
  const rows = query(
    `SELECT COALESCE(${column}, '') FROM ${table} WHERE id = ${Number(id)}`,
  );
  return rows.length ? rows[0][0] : "";
}

test.describe("e2 — the RenameEntry screens write localization, not the row", () => {
  // Every test restores what it touched: these are shared reference rows that
  // b1, c2 and c3 all read, so a rename left behind changes other specs.
  let restore: { table: string; id: string; values: { locale: string; value: string }[] }[] = [];

  test.beforeEach(() => {
    restore = [];
  });

  test.afterEach(() => {
    for (const r of restore) {
      for (const v of r.values) {
        exec(
          `UPDATE clinlims.localization_value lv
              SET value = '${v.value.replace(/'/g, "''")}'
             FROM ${r.table} e
            WHERE lv.localization_id = e.name_localization_id
              AND e.id = ${Number(r.id)} AND lv.locale = '${v.locale}'`,
        );
      }
    }
  });

  for (const s of SCREENS) {
    test(`GET ${s.path} — the form`, async ({ request }) => {
      const form = await readJson(await request.get(`rest/${s.path}`), s.path);

      expect(form.formName, "formName").toBe(s.formName);
      expect(form.formMethod, "formMethod").toBe("POST");
      expect(form.cancelAction, "cancelAction").toBe("Home");
      expect(form.submitOnCancel, "submitOnCancel").toBe(false);
      expect(form.cancelMethod, "cancelMethod").toBe("POST");

      const list = form[s.listKey];
      expect(Array.isArray(list), `${s.listKey} is a list`).toBe(true);
      expect(list.length, "and it is populated").toBeGreaterThan(0);
      expect(
        Object.keys(list[0]).sort(),
        "each entry is an id/value pair",
      ).toEqual(["id", "value"]);

      // The bean initialises these to "" rather than leaving them null, so they
      // are present on the blank form as empty strings.
      expect(form.nameEnglish, "nameEnglish default").toBe("");
      expect(form.nameFrench, "nameFrench default").toBe("");
      expect(form[s.idKey], `${s.idKey} default`).toBe("");
    });

    test(`POST ${s.path} — writes localization and leaves ${s.table.split(".")[1]}.${s.nameColumn}`, async ({
      request,
    }) => {
      const form = await readJson(await request.get(`rest/${s.path}`), s.path);
      const target = form[s.listKey][0];

      const before = localizationOf(s.table, target.id);
      expect(
        before.length,
        "the row the screen offers has a localization to write",
      ).toBeGreaterThan(0);
      restore.push({ table: s.table, id: target.id, values: before });

      const columnBefore = nameColumnOf(s.table, s.nameColumn, target.id);
      const en = `e2ren-${s.path}-en`;
      const fr = `e2ren-${s.path}-fr`;

      const res = await post(request, `rest/${s.path}`, {
        [s.idKey]: target.id,
        // Padded on purpose: the controller trims before the setter, and the
        // database is not what does it.
        nameEnglish: `  ${en}  `,
        nameFrench: `  ${fr}  `,
      });
      expect(res.status(), "rename").toBe(200);

      // Whether the POST carries the list back is per-controller, not a rule:
      // three of the four return the bound form untouched, and
      // SampleTypeRenameEntry calls setSampleTypeList again after the write.
      const body = await res.json();
      if (s.postIncludesList) {
        expect(
          Array.isArray(body[s.listKey]),
          "this one re-populates its list on the success path",
        ).toBe(true);
      } else {
        expect(
          body[s.listKey],
          "the list is absent from the POST response",
        ).toBe(undefined);
      }
      expect(body[s.idKey], "the submitted id is echoed").toBe(target.id);

      const after = localizationOf(s.table, target.id);
      const value = (loc: string) =>
        after.find((v) => v.locale === loc)?.value;

      expect(value("en"), "the English localization is written, trimmed").toBe(en);
      if (before.some((v) => v.locale === "fr")) {
        expect(value("fr"), "and the French one").toBe(fr);
      }

      // The point of the whole spec: the entity's own column is untouched.
      expect(
        nameColumnOf(s.table, s.nameColumn, target.id),
        `${s.nameColumn} is not what this screen writes`,
      ).toBe(columnBefore);
    });

    test(`POST ${s.path} — an unknown id is a silent 200`, async ({
      request,
    }) => {
      // Guarded by `if (entity != null)`, so the block is skipped whole and the
      // caller is told nothing.
      const res = await post(request, `rest/${s.path}`, {
        [s.idKey]: "999999",
        nameEnglish: "e2ren-nobody",
        nameFrench: "e2ren-personne",
      });
      expect(res.status(), "an id that does not exist").toBe(s.unknownIdStatus);
      expect(
        query(
          `SELECT count(*)::text FROM clinlims.localization_value WHERE value LIKE 'e2ren-nobody%'`,
        )[0][0],
        "and nothing was written either way",
      ).toBe("0");
    });
  }

  test("the rename leaves no audit row", async ({ request }) => {
    // reference_tables carries LOCALIZATION with keep_history = 'Y', and
    // LocalizationServiceImpl extends AuditableBaseObjectServiceImpl — and, as
    // with UNIT_OF_MEASURE, never sets auditTrailLog = true. The mechanism is
    // off. A port that writes the history row would be more correct than Java.
    const s = SCREENS[0];
    const form = await readJson(await request.get(`rest/${s.path}`), s.path);
    const target = form[s.listKey][0];
    const before = localizationOf(s.table, target.id);
    restore.push({ table: s.table, id: target.id, values: before });

    const [[locId]] = query(
      `SELECT name_localization_id::text FROM ${s.table} WHERE id = ${Number(target.id)}`,
    );

    const res = await post(request, `rest/${s.path}`, {
      [s.idKey]: target.id,
      nameEnglish: "e2ren-audit-en",
      nameFrench: "e2ren-audit-fr",
    });
    expect(res.status(), "rename").toBe(200);

    expect(
      query(
        `SELECT count(*)::text
           FROM clinlims.history h
           JOIN clinlims.reference_tables rt ON rt.id = h.reference_table
          WHERE upper(rt.name) = 'LOCALIZATION' AND h.reference_id = ${Number(locId)}`,
      )[0][0],
      "no history row, despite keep_history = Y on LOCALIZATION",
    ).toBe("0");
  });
});
