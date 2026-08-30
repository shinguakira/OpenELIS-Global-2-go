/**
 * e2 — TestRenameEntry, SelectListRenameEntry and the result select list.
 *
 * The last four endpoints of the testconfiguration package, and three of them
 * do something the name does not say:
 *
 *   - TestRenameEntry writes TWO localizations, because a test carries a name
 *     and a reporting name.
 *   - SelectListRenameEntry writes the localization AND two dictionary columns
 *     from the one English name.
 *   - POST /ResultSelectListAdd writes NOTHING. It reshapes the form and moves
 *     `page` to "2"; the write is POST /SaveResultSelectList.
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

function localizationValues(
  table: string,
  column: string,
  id: string,
): [string, string][] {
  return query(
    `SELECT lv.locale, COALESCE(lv.value, '')
       FROM ${table} e JOIN clinlims.localization_value lv
         ON lv.localization_id = e.${column}
      WHERE e.id = ${Number(id)} ORDER BY lv.locale`,
  ) as [string, string][];
}

test.describe("e2 — the rename and select-list screens", () => {
  let restore: (() => void)[] = [];

  test.beforeEach(() => {
    restore = [];
  });

  test.afterEach(() => {
    for (const r of restore) r();
    exec(
      `DELETE FROM clinlims.history WHERE timestamp > now() - interval '2 minutes'`,
    );
  });

  test("GET TestRenameEntry — the augmented test names", async ({ request }) => {
    const f = await readJson(
      await request.get("rest/TestRenameEntry"),
      "TestRenameEntry",
    );

    expect(f.formName, "formName").toBe("testRenameEntryForm");
    expect(Array.isArray(f.testList), "testList is a list").toBe(true);
    expect(f.testList.length, "and it is populated").toBeGreaterThan(0);
    // Four name fields, not two: a test has a name AND a reporting name.
    for (const key of [
      "nameEnglish",
      "nameFrench",
      "reportNameEnglish",
      "reportNameFrench",
      "testId",
    ]) {
      expect(f[key], `${key} default`).toBe("");
    }
  });

  test("POST TestRenameEntry — writes BOTH localizations", async ({
    request,
  }) => {
    const [[testId]] = query(
      `SELECT id::text FROM clinlims.test
        WHERE name_localization_id IS NOT NULL
          AND reporting_name_localization_id IS NOT NULL
        ORDER BY id LIMIT 1`,
    );
    const beforeName = localizationValues(
      "clinlims.test",
      "name_localization_id",
      testId,
    );
    const beforeReport = localizationValues(
      "clinlims.test",
      "reporting_name_localization_id",
      testId,
    );
    const [[nameColumn]] = query(
      `SELECT COALESCE(name, '') FROM clinlims.test WHERE id = ${Number(testId)}`,
    );

    restore.push(() => {
      for (const [col, rows] of [
        ["name_localization_id", beforeName],
        ["reporting_name_localization_id", beforeReport],
      ] as const) {
        for (const [locale, value] of rows) {
          exec(
            `UPDATE clinlims.localization_value lv
                SET value = '${value.replace(/'/g, "''")}'
               FROM clinlims.test e
              WHERE lv.localization_id = e.${col}
                AND e.id = ${Number(testId)} AND lv.locale = '${locale}'`,
          );
        }
      }
      // test.name moves too — see the assertion below.
      exec(
        `UPDATE clinlims.test SET name = '${nameColumn.replace(/'/g, "''")}'
          WHERE id = ${Number(testId)}`,
      );
    });

    const res = await post(request, "rest/TestRenameEntry", {
      testId,
      nameEnglish: "  e2tr-name-en  ",
      nameFrench: "  e2tr-name-fr  ",
      reportNameEnglish: "  e2tr-report-en  ",
      reportNameFrench: "  e2tr-report-fr  ",
    });
    expect(res.status(), "rename").toBe(200);

    // The POST sets cancelAction to a value the GET never carries.
    expect((await res.json()).cancelAction, "cancelAction on the POST").toBe(
      "CancelDictionary",
    );

    const nameAfter = Object.fromEntries(
      localizationValues("clinlims.test", "name_localization_id", testId),
    );
    expect(nameAfter.en, "the name localization, trimmed").toBe("e2tr-name-en");

    const reportAfter = Object.fromEntries(
      localizationValues(
        "clinlims.test",
        "reporting_name_localization_id",
        testId,
      ),
    );
    expect(reportAfter.en, "and the reporting one, which is a second row").toBe(
      "e2tr-report-en",
    );

    // AND the test.name column, which is the difference from every other
      // rename screen: those leave their entity alone, this one does not.
    expect(
      query(
        `SELECT COALESCE(name, '') FROM clinlims.test WHERE id = ${Number(testId)}`,
      )[0][0],
      "the test.s own name column IS written, unlike the other rename screens",
    ).toBe("e2tr-name-en");
    expect(nameColumn, "and it held something else before").not.toBe(
      "e2tr-name-en",
    );
  });

  test("GET SelectListRenameEntry — no formName, and whole Dictionary entities", async ({
    request,
  }) => {
    const f = await readJson(
      await request.get("rest/SelectListRenameEntry"),
      "SelectListRenameEntry",
    );

    // The bean sets no form name, so the key every other screen carries is
    // absent here.
    expect(f.formName, "formName is absent").toBe(undefined);
    expect(f.formMethod, "formMethod").toBe("POST");
    expect(
      Array.isArray(f.resultSelectOptionList),
      "resultSelectOptionList is a list",
    ).toBe(true);
    expect(f.resultSelectOptionList.length, "and it is populated").toBeGreaterThan(
      0,
    );

    const option = f.resultSelectOptionList[0];
    for (const key of ["id", "dictEntry", "isActive", "displayValue"]) {
      expect(option[key] !== undefined, `the entity carries ${key}`).toBe(true);
    }
  });

  test("POST SelectListRenameEntry — one name, three writes", async ({
    request,
  }) => {
    const f = await readJson(
      await request.get("rest/SelectListRenameEntry"),
      "SelectListRenameEntry",
    );
    const option = f.resultSelectOptionList[0];
    const [[dictEntry, localAbbrev]] = query(
      `SELECT COALESCE(dict_entry, ''), COALESCE(local_abbrev, '')
         FROM clinlims.dictionary WHERE id = ${Number(option.id)}`,
    );
    const before = localizationValues(
      "clinlims.dictionary",
      "name_localization_id",
      option.id,
    );

    restore.push(() => {
      exec(
        `UPDATE clinlims.dictionary
            SET dict_entry = '${dictEntry.replace(/'/g, "''")}',
                local_abbrev = '${localAbbrev.replace(/'/g, "''")}'
          WHERE id = ${Number(option.id)}`,
      );
      for (const [locale, value] of before) {
        exec(
          `UPDATE clinlims.localization_value lv
              SET value = '${value.replace(/'/g, "''")}'
             FROM clinlims.dictionary e
            WHERE lv.localization_id = e.name_localization_id
              AND e.id = ${Number(option.id)} AND lv.locale = '${locale}'`,
        );
      }
    });

    const res = await post(request, "rest/SelectListRenameEntry", {
      resultSelectOptionId: option.id,
      nameEnglish: "e2sl-renamed",
      nameFrench: "e2sl-renomme",
    });
    expect(res.status(), "rename").toBe(200);
    expect(
      Array.isArray((await res.json()).resultSelectOptionList),
      "the POST answers the rebuilt list",
    ).toBe(true);

    const [[nowEntry, nowAbbrev]] = query(
      `SELECT COALESCE(dict_entry, ''), COALESCE(local_abbrev, '')
         FROM clinlims.dictionary WHERE id = ${Number(option.id)}`,
    );
    // The ENGLISH name lands in three places.
    expect(nowEntry, "dict_entry takes the English name").toBe("e2sl-renamed");
    expect(nowAbbrev, "and so does local_abbrev").toBe("e2sl-renamed");

    const after = Object.fromEntries(
      localizationValues("clinlims.dictionary", "name_localization_id", option.id),
    );
    expect(after.en, "the localization takes it too").toBe("e2sl-renamed");
    if (before.some(([l]) => l === "fr")) {
      expect(after.fr, "and the French name reaches only there").toBe(
        "e2sl-renomme",
      );
    }
  });

  test("POST ResultSelectListAdd — a write endpoint that writes nothing", async ({
    request,
  }) => {
    const before = query(
      `SELECT count(*)::text FROM clinlims.dictionary`,
    )[0][0];

    const res = await post(request, "rest/ResultSelectListAdd", {
      nameEnglish: "e2rsl-name",
      nameFrench: "",
      normal: true,
      qualifiable: false,
    });
    expect(res.status(), "add").toBe(200);

    const body = await res.json();
    expect(body.page, "page moves to 2").toBe("2");
    // `if ("".equalsIgnoreCase(nameFrench)) setNameFrench(nameEnglish)` — the
    // blank half is filled from the other language.
    expect(body.nameFrench, "the blank French name takes the English one").toBe(
      "e2rsl-name",
    );
    expect(Array.isArray(body.tests), "and the form gains a test list").toBe(
      true,
    );

    expect(
      query(`SELECT count(*)::text FROM clinlims.dictionary`)[0][0],
      "no dictionary entry was created",
    ).toBe(before);
  });

  test("POST SaveResultSelectList — creates the dictionary entry and the results", async ({
    request,
  }) => {
    const [[testId]] = query(
      `SELECT tr.test_id::text FROM clinlims.test_result tr
        WHERE tr.tst_rslt_type = 'D' ORDER BY tr.test_id LIMIT 1`,
    );
    const beforeDict = query(
      `SELECT count(*)::text FROM clinlims.dictionary`,
    )[0][0];
    const beforeResults = query(
      `SELECT count(*)::text FROM clinlims.test_result WHERE test_id = ${Number(testId)}`,
    )[0][0];

    restore.push(() => {
      exec(
        `DELETE FROM clinlims.test_result WHERE value IN
           (SELECT id::text FROM clinlims.dictionary WHERE dict_entry = 'e2rsl-opt')`,
      );
      exec(
        `DELETE FROM clinlims.localization_value WHERE localization_id IN
           (SELECT name_localization_id FROM clinlims.dictionary WHERE dict_entry = 'e2rsl-opt')`,
      );
      const locs = query(
        `SELECT name_localization_id::text FROM clinlims.dictionary WHERE dict_entry = 'e2rsl-opt'`,
      ).map((r) => r[0]);
      exec(`DELETE FROM clinlims.dictionary WHERE dict_entry = 'e2rsl-opt'`);
      for (const id of locs) {
        exec(`DELETE FROM clinlims.localization WHERE id = ${Number(id)}`);
      }
    });

    const res = await post(request, "rest/SaveResultSelectList", {
      nameEnglish: "e2rsl-opt",
      nameFrench: "e2rsl-opt-fr",
      loincCode: "",
      testSelectListJson: JSON.stringify([
        {
          id: testId,
          items: [{ order: 3, normal: true, qualifiable: false }],
        },
      ]),
    });
    expect(res.status(), "save").toBe(200);

    const created = query(
      `SELECT id::text, COALESCE(local_abbrev, ''), COALESCE(is_active, ''),
              COALESCE(sort_order::text, '')
         FROM clinlims.dictionary WHERE dict_entry = 'e2rsl-opt'`,
    );
    expect(created.length, "one dictionary entry was created").toBe(1);
    const [dictId, abbrev, active, sortOrder] = created[0];
    expect(abbrev, "local_abbrev takes the English name too").toBe("e2rsl-opt");
    expect(active, "and it starts active").toBe("Y");
    expect(sortOrder, "with sort order 1").toBe("1");

    const results = query(
      `SELECT COALESCE(sort_order::text, ''), COALESCE(tst_rslt_type, ''),
              COALESCE(is_normal, false)::text, COALESCE(is_quantifiable, false)::text
         FROM clinlims.test_result
        WHERE test_id = ${Number(testId)} AND value = '${dictId}'`,
    );
    expect(results.length, "and one test_result for the named test").toBe(1);
    // setSortOrder(String.valueOf(order * 10)) — the same multiplier
    // TestActivation applies.
    expect(results[0][0], "the order submitted was multiplied by ten").toBe("30");
    expect(results[0][1], "typed as a dictionary result").toBe("D");

    expect(
      Number(query(`SELECT count(*)::text FROM clinlims.dictionary`)[0][0]) -
        Number(beforeDict),
      "exactly one dictionary row",
    ).toBe(1);
    expect(
      Number(
        query(
          `SELECT count(*)::text FROM clinlims.test_result WHERE test_id = ${Number(testId)}`,
        )[0][0],
      ) - Number(beforeResults),
      "exactly one test_result row",
    ).toBe(1);
  });
});
