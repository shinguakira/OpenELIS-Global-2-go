/**
 * e2 slice 1 — UOM create and rename.
 *
 * The smallest whole module in the testconfiguration package, and the shape all
 * 24 of those controllers share: a GET that builds a form out of
 * DisplayListService lists, and a POST that writes one row and answers the
 * bound form back.
 *
 * Three of the things pinned here are the opposite of what the Java source
 * reads like, and each one is a way a reasonable port goes wrong:
 *
 *   - the create writes NO audit row, though every signpost says it should
 *   - `existingFrenchNames` is the literal word "French", once per UOM
 *   - the rename writes `name` and leaves `description` on the old value
 *
 * Runs against Java (the measurement) and Go (the check).
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

const UOM_CREATE = "rest/UomCreate";
const UOM_RENAME = "rest/UomRenameEntry";

/** Names this spec owns. Nothing ships with them. */
const PROBE = "e2UomProbe";
const RENAMED = "e2UomRenamed";

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

function probeRows(): { id: string; name: string; description: string }[] {
  return query(
    `SELECT id::text, name, COALESCE(description, '')
       FROM clinlims.unit_of_measure
      WHERE name IN ('${PROBE}', '${RENAMED}') OR description IN ('${PROBE}', '${RENAMED}')
      ORDER BY id`,
  ).map(([id, name, description]) => ({ id, name, description }));
}

/** Every history row written for unit_of_measure in the window this test owns. */
function uomHistory(ids: string[]): string[][] {
  if (ids.length === 0) return [];
  return query(
    `SELECT h.activity, h.reference_id::text
       FROM clinlims.history h
       JOIN clinlims.reference_tables rt ON rt.id = h.reference_table
      WHERE upper(rt.name) = 'UNIT_OF_MEASURE'
        AND h.reference_id IN (${ids.map(Number).join(",")})`,
  );
}

test.describe("e2 — UOM create and rename", () => {
  let touched: string[] = [];

  test.beforeEach(() => {
    touched = probeRows().map((r) => r.id);
  });

  test.afterEach(() => {
    const ids = [...new Set([...touched, ...probeRows().map((r) => r.id)])];
    for (const id of ids) {
      exec(
        `DELETE FROM clinlims.history
           WHERE reference_id = ${Number(id)}
             AND reference_table IN (SELECT id FROM clinlims.reference_tables
                                      WHERE upper(name) = 'UNIT_OF_MEASURE')`,
      );
    }
    exec(
      `DELETE FROM clinlims.unit_of_measure
        WHERE name IN ('${PROBE}', '${RENAMED}') OR description IN ('${PROBE}', '${RENAMED}')`,
    );
  });

  test("GET UomCreate — the form, and the hardcoded French", async ({
    request,
  }) => {
    const form = await readJson(await request.get(UOM_CREATE), UOM_CREATE);

    expect(form.formName, "formName").toBe("uomCreateForm");
    expect(form.formMethod, "formMethod").toBe("POST");
    expect(form.cancelAction, "cancelAction").toBe("Home");
    expect(form.submitOnCancel, "submitOnCancel").toBe(false);
    expect(form.cancelMethod, "cancelMethod").toBe("POST");

    expect(Array.isArray(form.existingUomList), "existingUomList is a list")
      .toBe(true);
    expect(form.existingUomList.length, "and it is populated").toBeGreaterThan(0);
    expect(
      Object.keys(form.existingUomList[0]).sort(),
      "each entry is an id/value pair",
    ).toEqual(["id", "value"]);

    // The name strings carry a leading AND trailing separator — getExistingUomNames
    // seeds the builder with "$" and appends one after every name.
    expect(form.existingEnglishNames.startsWith("$"), "leading separator").toBe(
      true,
    );
    expect(form.existingEnglishNames.endsWith("$"), "trailing separator").toBe(
      true,
    );

    // And the French string is not French. UnitOfMeasure.getLocalization() is a
    // stub that builds a Localization in memory per call and ends with
    // setFrench("French") — a literal. unit_of_measure has no localization
    // column at all, so there is nothing else it could return.
    const french = form.existingFrenchNames.split("$").filter(Boolean);
    expect(french.length, "one entry per UOM").toBeGreaterThan(0);
    expect(
      new Set(french),
      'every entry is the literal word "French"',
    ).toEqual(new Set(["French"]));
  });

  test("POST UomCreate — writes the row, and NO audit row", async ({
    request,
  }) => {
    const res = await post(request, UOM_CREATE, { uomEnglishName: PROBE });
    expect(res.status(), "create").toBe(200);

    // The success branch returns the bound form WITHOUT calling
    // setupDisplayItems, so the lists the GET carries are absent here. Only the
    // validation-failure branch re-populates them — and that is a 200 as well.
    const body = await res.json();
    expect(Object.keys(body).sort(), "the POST body is not the GET body").toEqual(
      [
        "cancelAction",
        "cancelMethod",
        "formMethod",
        "formName",
        "submitOnCancel",
        "uomEnglishName",
      ],
    );
    expect(body.uomEnglishName, "the submitted name is echoed").toBe(PROBE);

    const rows = probeRows();
    expect(rows.length, "one row was created").toBe(1);
    touched.push(rows[0].id);

    // createUnitOfMeasure puts the SAME string in both columns.
    expect(rows[0].name, "name").toBe(PROBE);
    expect(rows[0].description, "description takes the same value").toBe(PROBE);

    const [[active]] = query(
      `SELECT is_active FROM clinlims.unit_of_measure WHERE id = ${rows[0].id}`,
    );
    expect(active, "is_active takes the column default").toBe("Y");

    // reference_tables has UNIT_OF_MEASURE with keep_history = 'Y', and
    // UnitOfMeasureServiceImpl extends AuditableBaseObjectServiceImpl — but it
    // never sets auditTrailLog = true, so the mechanism is off. A port that
    // audits this write is wrong.
    expect(
      uomHistory([rows[0].id]).length,
      "the create leaves no audit row, despite keep_history = Y",
    ).toBe(0);
  });

  test("the two UOM lists are the same query, and only ONE of them refreshes", async ({
    request,
  }) => {
    // `inactiveUomList` is not an inactive list. DisplayListService builds
    // UNIT_OF_MEASURE with createUnitOfMeasureList() and
    // UNIT_OF_MEASURE_INACTIVE with createUOMList(), and the two methods are
    // the same six lines — getAll(), mapped to (id, localizedName). Neither
    // filters on is_active; the filter is commented out in the first.
    // Asserted as a DELTA, not as an equality between the two lists. They are
    // built from the same query, but they are two independent snapshots taken
    // at different times, so in a suite that creates UOMs they drift apart and
    // stay apart — the run that found this had a UOM in `existingUomList` that
    // had already been deleted from the table, because nothing refreshed the
    // list after the delete. What is stable is the effect of one write.
    const before = await readJson(await request.get(UOM_CREATE), UOM_CREATE);

    const res = await post(request, UOM_CREATE, { uomEnglishName: PROBE });
    expect(res.status(), "create").toBe(200);
    const rows = probeRows();
    expect(rows.length, "one row was created").toBe(1);
    touched.push(rows[0].id);

    // The handler refreshes both list types. Only one of the calls does
    // anything: refreshList's switch has a case for UNIT_OF_MEASURE and none
    // for UNIT_OF_MEASURE_INACTIVE, so the second is a silent no-op and that
    // list keeps whatever it was loaded with at startup.
    const after = await readJson(await request.get(UOM_CREATE), UOM_CREATE);
    const ids = (l: { id: string }[]) => l.map((p) => p.id);

    expect(
      ids(after.existingUomList),
      "the refreshed list has the new UOM",
    ).toContain(rows[0].id);
    expect(
      ids(after.inactiveUomList),
      "the one whose refresh is a no-op does not",
    ).not.toContain(rows[0].id);

    // So a list that was a copy of the other is now permanently one row behind
    // — for the life of the process, not the request. Same shape as the
    // configuration cache e1-8 pinned: a port that reads the table per request
    // answers a list Java does not have.
    expect(
      after.inactiveUomList,
      "and it is exactly what it was before the write",
    ).toEqual(before.inactiveUomList);
  });

  test("GET UomRenameEntry — the form", async ({ request }) => {
    const form = await readJson(await request.get(UOM_RENAME), UOM_RENAME);

    expect(form.formName, "formName").toBe("uomRenameEntryForm");
    expect(Array.isArray(form.uomList), "uomList is a list").toBe(true);
    expect(form.uomList.length, "and it is populated").toBeGreaterThan(0);

    // The bean initialises these to "" rather than leaving them null.
    expect(form.nameEnglish, "nameEnglish default").toBe("");
    expect(form.uomId, "uomId default").toBe("");
  });

  test("POST UomRenameEntry — renames `name` and LEAVES `description`", async ({
    request,
  }) => {
    const created = await post(request, UOM_CREATE, { uomEnglishName: PROBE });
    expect(created.status(), "create").toBe(200);
    const before = probeRows();
    expect(before.length, "one row to rename").toBe(1);
    const id = before[0].id;
    touched.push(id);

    const res = await post(request, UOM_RENAME, {
      uomId: id,
      nameEnglish: `  ${RENAMED}  `, // padded on purpose — see below
    });
    expect(res.status(), "rename").toBe(200);

    const [[name, description]] = query(
      `SELECT name, COALESCE(description, '') FROM clinlims.unit_of_measure WHERE id = ${id}`,
    );

    // updateUomNames calls setUnitOfMeasureName(nameEnglish.trim()) and nothing
    // else. The padding is gone...
    expect(name, "the name is renamed, and trimmed").toBe(RENAMED);
    // ...and description keeps the value the CREATE put there. The two columns
    // agree when a UOM is created and disagree forever after it is renamed.
    expect(description, "description is not touched by a rename").toBe(PROBE);

    expect(
      uomHistory([id]).length,
      "the rename leaves no audit row either",
    ).toBe(0);
  });

  test("POST UomRenameEntry — an unknown id is a silent 200", async ({
    request,
  }) => {
    // getUnitOfMeasureById returns null and updateUomNames skips the whole
    // block. No row changes, nothing is reported, and the caller is told 200.
    const res = await post(request, UOM_RENAME, {
      uomId: "999999",
      nameEnglish: RENAMED,
    });
    expect(res.status(), "an id that does not exist").toBe(200);
    expect(probeRows().length, "and nothing was written").toBe(0);
  });
});
