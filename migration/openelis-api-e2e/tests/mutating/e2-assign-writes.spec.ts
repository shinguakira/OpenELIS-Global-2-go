/**
 * e2 — the three *TestAssign screens.
 *
 * They attach a test to a sample type, a test to a test section, or a set of
 * tests to a panel. Two of them share a guard worth pinning: assigning to the
 * same id you are moving away from returns before any write, so the request is
 * a 200 that does nothing.
 *
 * All three carry a side effect the screen name does not suggest — assigning to
 * an INACTIVE sample type or test section turns it back on, and naming one to
 * deactivate turns that one off.
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

test.describe("e2 — the TestAssign screens", () => {
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

  test("GET SampleTypeTestAssign — two lists concatenated, and a toString()", async ({
    request,
  }) => {
    const f = await readJson(
      await request.get("rest/SampleTypeTestAssign"),
      "SampleTypeTestAssign",
    );

    expect(f.formName, "formName").toBe("sampleTypeTestAssignForm");
    expect(f.testId, "testId default").toBe("");
    expect(f.sampleTypeId, "sampleTypeId default").toBe("");
    expect(f.deactivateSampleTypeId, "deactivateSampleTypeId default").toBe("");

    // getListWithLeadingBlank puts an id "0" / value "" entry first.
    expect(f.sampleTypeList[0], "the leading blank").toEqual({
      id: "0",
      value: "",
    });

    // sampleTypeTestList is a MAP whose keys are the IdValuePair.toString() of
    // each sample type — Jackson renders a non-String map key that way — and
    // whose values are the ACTIVE tests under it.
    expect(Array.isArray(f.sampleTypeTestList), "not an array").toBe(false);
    expect(typeof f.sampleTypeTestList, "an object").toBe("object");
    expect(
      Object.keys(f.sampleTypeTestList)[0],
      "keyed by the pair.s toString, blank entry first",
    ).toBe("id=0, value=");
    const populated = Object.entries(f.sampleTypeTestList).find(
      ([, v]) => (v as unknown[]).length > 0,
    ) as [string, any[]];
    expect(populated, "at least one sample type carries tests").toBeTruthy();
    expect(
      Object.keys(populated[1][0]).sort(),
      "and each test is an id/value pair",
    ).toEqual(["id", "value"]);
  });

  test("GET TestSectionTestAssign — the same shape, sections instead", async ({
    request,
  }) => {
    const f = await readJson(
      await request.get("rest/TestSectionTestAssign"),
      "TestSectionTestAssign",
    );

    expect(f.formName, "formName").toBe("testSectionTestAssignForm");
    expect(f.testSectionList[0], "the leading blank").toEqual({
      id: "0",
      value: "",
    });
    expect(typeof f.sectionTestList, "sectionTestList is a map").toBe("object");
    expect(Object.keys(f.sectionTestList)[0], "blank entry first").toBe(
      "id=0, value=",
    );
    expect(f.testId, "testId default").toBe("");
  });

  test("GET PanelTestAssign — the panel list, and an empty selectedPanel", async ({
    request,
  }) => {
    const f = await readJson(
      await request.get("rest/PanelTestAssign"),
      "PanelTestAssign",
    );

    expect(f.formName, "formName").toBe("panelTestAssignForm");
    expect(Array.isArray(f.panelList), "panelList is a list").toBe(true);
    expect(f.panelId, "panelId default").toBe("");
    // The bean initialises selectedPanel either way, so the blank form carries
    // it with both halves empty rather than dropping the key.
    expect(f.selectedPanel, "selectedPanel on the blank form").toEqual({
      tests: [],
      availableTests: [],
    });
  });

  test("POST TestSectionTestAssign — moves the test and flips both sections", async ({
    request,
  }) => {
    const [[testId, sectionId]] = query(
      `SELECT id::text, COALESCE(test_section_id::text, '') FROM clinlims.test
        WHERE test_section_id IS NOT NULL ORDER BY id LIMIT 1`,
    );
    const [[targetSection, targetActive]] = query(
      `SELECT id::text, COALESCE(is_active, '') FROM clinlims.test_section
        WHERE id <> ${Number(sectionId)} ORDER BY id LIMIT 1`,
    );
    const [[oldActive]] = query(
      `SELECT COALESCE(is_active, '') FROM clinlims.test_section WHERE id = ${Number(sectionId)}`,
    );

    restore.push(() => {
      exec(
        `UPDATE clinlims.test SET test_section_id = ${Number(sectionId)} WHERE id = ${Number(testId)}`,
      );
      exec(
        `UPDATE clinlims.test_section SET is_active = '${targetActive}' WHERE id = ${Number(targetSection)}`,
      );
      exec(
        `UPDATE clinlims.test_section SET is_active = '${oldActive}' WHERE id = ${Number(sectionId)}`,
      );
    });

    const res = await post(request, "rest/TestSectionTestAssign", {
      testId,
      testSectionId: targetSection,
      deactivateTestSectionId: sectionId,
    });
    expect(res.status(), "assign").toBe(200);

    const [[nowSection]] = query(
      `SELECT COALESCE(test_section_id::text, '') FROM clinlims.test WHERE id = ${Number(testId)}`,
    );
    expect(nowSection, "the test moved").toBe(targetSection);

    const [[newActive]] = query(
      `SELECT COALESCE(is_active, '') FROM clinlims.test_section WHERE id = ${Number(targetSection)}`,
    );
    expect(newActive, "the section assigned TO is active").toBe("Y");

    const [[deactivated]] = query(
      `SELECT COALESCE(is_active, '') FROM clinlims.test_section WHERE id = ${Number(sectionId)}`,
    );
    expect(deactivated, "and the one named for deactivation is off").toBe("N");
  });

  test("POST TestSectionTestAssign — the same id twice writes nothing", async ({
    request,
  }) => {
    const [[testId, sectionId]] = query(
      `SELECT id::text, COALESCE(test_section_id::text, '') FROM clinlims.test
        WHERE test_section_id IS NOT NULL ORDER BY id LIMIT 1`,
    );
    const [[before]] = query(
      `SELECT COALESCE(is_active, '') FROM clinlims.test_section WHERE id = ${Number(sectionId)}`,
    );

    // `if (testSectionId.equals(deactivateTestSectionId)) return form;` — the
    // guard comes before every write.
    const res = await post(request, "rest/TestSectionTestAssign", {
      testId,
      testSectionId: sectionId,
      deactivateTestSectionId: sectionId,
    });
    expect(res.status(), "assign to the same section").toBe(200);

    const [[after]] = query(
      `SELECT COALESCE(is_active, '') FROM clinlims.test_section WHERE id = ${Number(sectionId)}`,
    );
    expect(after, "the section was not deactivated").toBe(before);
  });

  test("POST SampleTypeTestAssign — the join is REPLACED, not added to", async ({
    request,
  }) => {
    const [[testId, sampleTypeId]] = query(
      `SELECT t.id::text, st.sample_type_id::text
         FROM clinlims.test t JOIN clinlims.sampletype_test st ON st.test_id = t.id
        ORDER BY t.id LIMIT 1`,
    );
    const [[targetType, targetActive]] = query(
      `SELECT id::text, COALESCE(is_active, false)::text FROM clinlims.type_of_sample
        WHERE domain = 'H' AND id <> ${Number(sampleTypeId)} ORDER BY id LIMIT 1`,
    );
    const existing = query(
      `SELECT sample_type_id::text FROM clinlims.sampletype_test WHERE test_id = ${Number(testId)}`,
    ).map((r) => r[0]);

    restore.push(() => {
      exec(`DELETE FROM clinlims.sampletype_test WHERE test_id = ${Number(testId)}`);
      for (const st of existing) {
        exec(
          `INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id)
             VALUES (nextval('clinlims.sample_type_test_seq'), ${Number(st)}, ${Number(testId)})`,
        );
      }
      exec(
        `UPDATE clinlims.type_of_sample SET is_active = ${targetActive} WHERE id = ${Number(targetType)}`,
      );
    });

    const res = await post(request, "rest/SampleTypeTestAssign", {
      testId,
      sampleTypeId: targetType,
      deactivateSampleTypeId: "",
    });
    expect(res.status(), "assign").toBe(200);

    const after = query(
      `SELECT sample_type_id::text FROM clinlims.sampletype_test WHERE test_id = ${Number(testId)}`,
    ).map((r) => r[0]);
    expect(after, "exactly one link, to the submitted sample type").toEqual([
      targetType,
    ]);

    const [[nowActive]] = query(
      `SELECT COALESCE(is_active, false)::text FROM clinlims.type_of_sample WHERE id = ${Number(targetType)}`,
    );
    expect(nowActive, "the sample type assigned TO is active").toBe("true");
  });

  test("POST PanelTestAssign — the membership is replaced by currentTests", async ({
    request,
  }) => {
    const [[panelId]] = query(
      `SELECT panel_id::text FROM clinlims.panel_item GROUP BY panel_id ORDER BY panel_id LIMIT 1`,
    );
    const before = query(
      `SELECT test_id::text, COALESCE(sort_order, 0)::text FROM clinlims.panel_item
        WHERE panel_id = ${Number(panelId)} ORDER BY sort_order, id`,
    );
    restore.push(() => {
      exec(`DELETE FROM clinlims.panel_item WHERE panel_id = ${Number(panelId)}`);
      for (const [testId, sortOrder] of before) {
        exec(
          `INSERT INTO clinlims.panel_item (id, panel_id, test_id, sort_order, lastupdated)
             VALUES (nextval('clinlims.panel_item_seq'), ${Number(panelId)},
                     ${Number(testId)}, ${Number(sortOrder)}, now())`,
        );
      }
    });

    // Keep only the first member: updatePanelItems replaces the set.
    const keep = before[0][0];
    const res = await post(request, "rest/PanelTestAssign", {
      panelId,
      currentTests: [keep],
    });
    expect(res.status(), "assign").toBe(200);

    const after = query(
      `SELECT test_id::text FROM clinlims.panel_item WHERE panel_id = ${Number(panelId)}`,
    ).map((r) => r[0]);
    expect(after, "the panel holds only what was submitted").toEqual([keep]);
  });

  test("POST PanelTestAssign — a blank panel id writes nothing", async ({
    request,
  }) => {
    const [[panelId]] = query(
      `SELECT panel_id::text FROM clinlims.panel_item GROUP BY panel_id ORDER BY panel_id LIMIT 1`,
    );
    const before = query(
      `SELECT count(*)::text FROM clinlims.panel_item WHERE panel_id = ${Number(panelId)}`,
    )[0][0];

    // `if (!isBlankOrNull(panelId))` guards the whole block.
    const res = await post(request, "rest/PanelTestAssign", {
      panelId: "",
      currentTests: [],
    });
    expect(res.status(), "a blank panel id").toBe(200);
    expect(
      query(
        `SELECT count(*)::text FROM clinlims.panel_item WHERE panel_id = ${Number(panelId)}`,
      )[0][0],
      "and nothing was removed",
    ).toBe(before);
  });
});
