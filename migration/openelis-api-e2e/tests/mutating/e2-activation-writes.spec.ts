/**
 * e2 — TestActivation and TestOrderability.
 *
 * The two screens that turn tests on and off. They share the *Order screens'
 * double-encoded jsonChangeList, and every other thing about their request
 * shape had to be measured, because none of it is guessable:
 *
 *   - ALL FOUR keys must be present. A missing one is `parser.parse(null)`,
 *     which is a NullPointerException rather than the ParseException the catch
 *     is written for, so the request is a 500.
 *   - Each activateTest entry needs an `activated` field the handler never
 *     reads. The validator requires it; omitting it refuses the whole request
 *     at 200, with nothing written.
 *   - TestOrderability's ids must be JSON STRINGS. Its copy of
 *     getIdsForActions casts the id value to String, so a numeric id is a
 *     ClassCastException and a 500. TestActivation's copy takes a number.
 *   - The stored sort order is the submitted one TIMES TEN.
 *   - TestOrderability answers {"status":"success"}, not the form.
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

/** A test that is active and belongs to a sample type, so both screens list it. */
function anyActiveTest() {
  const [[id, isActive, sortOrder, orderable]] = query(
    `SELECT t.id::text, COALESCE(t.is_active, ''), COALESCE(t.sort_order::text, ''),
            COALESCE(t.orderable, false)::text
       FROM clinlims.test t
       JOIN clinlims.sampletype_test st ON st.test_id = t.id
      WHERE t.is_active = 'Y'
      ORDER BY t.id
      LIMIT 1`,
  );
  return { id, isActive, sortOrder, orderable };
}

/** The four keys TestActivation insists on, with only the named one populated. */
function activationBody(overrides: Record<string, string>) {
  return {
    jsonChangeList: JSON.stringify({
      activateSample: "[]",
      deactivateSample: "[]",
      activateTest: "[]",
      deactivateTest: "[]",
      ...overrides,
    }),
  };
}

test.describe("e2 — TestActivation and TestOrderability", () => {
  let restore: (() => void)[] = [];

  test.beforeEach(() => {
    restore = [];
    exec(
      `DELETE FROM clinlims.history WHERE timestamp > now() - interval '30 seconds'`,
    );
  });

  test.afterEach(() => {
    for (const r of restore) r();
    exec(
      `DELETE FROM clinlims.history WHERE timestamp > now() - interval '2 minutes'`,
    );
  });

  function restoreTest(t: ReturnType<typeof anyActiveTest>) {
    restore.push(() =>
      exec(
        `UPDATE clinlims.test
            SET is_active = '${t.isActive}',
                sort_order = ${t.sortOrder === "" ? "NULL" : t.sortOrder},
                orderable = ${t.orderable}
          WHERE id = ${Number(t.id)}`,
      ),
    );
  }

  test("GET TestActivation — a bean per sample type, split active/inactive", async ({
    request,
  }) => {
    const f = await readJson(
      await request.get("rest/TestActivation"),
      "TestActivation",
    );

    expect(f.formName, "formName").toBe("testActivationForm");
    expect(f.jsonChangeList, "jsonChangeList default").toBe("");
    expect(Array.isArray(f.activeTestList), "activeTestList is a list").toBe(true);
    expect(f.activeTestList.length, "and it is populated").toBeGreaterThan(0);

    const bean = f.activeTestList[0];
    expect(Object.keys(bean).sort(), "the bean's shape").toEqual([
      "activeTests",
      "inactiveTests",
      "sampleType",
    ]);
    expect(
      Object.keys(bean.sampleType).sort(),
      "sampleType is an id/value pair",
    ).toEqual(["id", "value"]);

    // "if not active we use alphabetical ordering, the default is display order"
    const inactiveNames = f.inactiveTestList.map((b: any) => b.sampleType.value);
    expect(
      inactiveNames,
      "the inactive sample types are in alphabetical order",
    ).toEqual([...inactiveNames].sort());
  });

  test("POST TestActivation — the stored sort order is the submitted one TIMES TEN", async ({
    request,
  }) => {
    const target = anyActiveTest();
    restoreTest(target);

    const res = await post(
      request,
      "rest/TestActivation",
      activationBody({
        activateTest: JSON.stringify([
          { id: Number(target.id), sortOrder: 7, activated: "true" },
        ]),
      }),
    );
    expect(res.status(), "activate").toBe(200);

    const [[isActive, sortOrder]] = query(
      `SELECT COALESCE(is_active, ''), COALESCE(sort_order::text, '')
         FROM clinlims.test WHERE id = ${Number(target.id)}`,
    );
    expect(isActive, "the test is active").toBe("Y");
    // setSortOrder(String.valueOf(set.sortOrder * 10)) — 7 goes in, 70 lands.
    expect(sortOrder, "and the submitted order was multiplied by ten").toBe("70");

    // The POST answers the REBUILT lists, not the bound form.
    const body = await res.json();
    expect(
      Array.isArray(body.activeTestList),
      "the response carries the rebuilt lists",
    ).toBe(true);
  });

  test("POST TestActivation — omitting `activated` refuses the write, at 200", async ({
    request,
  }) => {
    const target = anyActiveTest();
    restoreTest(target);

    // The handler never reads `activated`. The VALIDATOR requires it, so the
    // whole request is refused — and a refusal here is a 200 carrying the form.
    const res = await post(
      request,
      "rest/TestActivation",
      activationBody({
        activateTest: JSON.stringify([{ id: Number(target.id), sortOrder: 7 }]),
      }),
    );
    expect(res.status(), "refused by the validator").toBe(200);

    const [[sortOrder]] = query(
      `SELECT COALESCE(sort_order::text, '') FROM clinlims.test WHERE id = ${Number(target.id)}`,
    );
    expect(sortOrder, "and nothing was written").toBe(target.sortOrder);
  });

  test("POST TestActivation — deactivating leaves the sort order alone", async ({
    request,
  }) => {
    const target = anyActiveTest();
    restoreTest(target);

    const res = await post(
      request,
      "rest/TestActivation",
      activationBody({
        deactivateTest: JSON.stringify([{ id: Number(target.id) }]),
      }),
    );
    expect(res.status(), "deactivate").toBe(200);

    const [[isActive, sortOrder]] = query(
      `SELECT COALESCE(is_active, ''), COALESCE(sort_order::text, '')
         FROM clinlims.test WHERE id = ${Number(target.id)}`,
    );
    expect(isActive, "the test is inactive").toBe("N");
    // getDeactivatedTests calls setIsActive only.
    expect(sortOrder, "and the sort order is unchanged").toBe(target.sortOrder);
  });

  test("POST TestActivation — a missing key is a 500", async ({ request }) => {
    const target = anyActiveTest();
    const res = await post(request, "rest/TestActivation", {
      jsonChangeList: JSON.stringify({
        activateTest: JSON.stringify([
          { id: Number(target.id), sortOrder: 7, activated: "true" },
        ]),
      }),
    });
    expect(res.status(), "the other three keys are absent").toBe(500);
  });

  test("GET TestOrderability — the same bean, split on orderable", async ({
    request,
  }) => {
    const f = await readJson(
      await request.get("rest/TestOrderability"),
      "TestOrderability",
    );

    expect(f.formName, "formName").toBe("testOrderabilityForm");
    expect(f.jsonChangeList, "jsonChangeList default").toBe("");
    expect(Array.isArray(f.orderableTestList), "orderableTestList is a list").toBe(
      true,
    );
    expect(f.orderableTestList.length, "and it is populated").toBeGreaterThan(0);
    expect(
      Object.keys(f.orderableTestList[0]).sort(),
      "it reuses TestActivationBean, so the keys carry the other screen's names",
    ).toEqual(["activeTests", "inactiveTests", "sampleType"]);
  });

  test("POST TestOrderability — string ids, and the answer is not the form", async ({
    request,
  }) => {
    const target = anyActiveTest();
    restoreTest(target);
    const want = target.orderable !== "true";

    const res = await post(request, "rest/TestOrderability", {
      jsonChangeList: JSON.stringify({
        activateTest: want ? JSON.stringify([{ id: target.id }]) : "[]",
        deactivateTest: want ? "[]" : JSON.stringify([{ id: target.id }]),
      }),
    });
    expect(res.status(), "toggle").toBe(200);
    // Not the form: a bare status map.
    expect(await res.json(), "the response body").toEqual({ status: "success" });

    const [[orderable, isActive, sortOrder]] = query(
      `SELECT COALESCE(orderable, false)::text, COALESCE(is_active, ''),
              COALESCE(sort_order::text, '')
         FROM clinlims.test WHERE id = ${Number(target.id)}`,
    );
    expect(orderable, "orderable followed the change list").toBe(String(want));
    expect(isActive, "is_active is untouched").toBe(target.isActive);
    expect(sortOrder, "and so is the sort order").toBe(target.sortOrder);
  });

  test("POST TestOrderability — a NUMERIC id is a 500", async ({ request }) => {
    const target = anyActiveTest();
    // `(String) ((JSONObject) actionArray.get(i)).get("id")` — this screen's
    // copy of getIdsForActions casts the id value to String, so a number is a
    // ClassCastException. TestActivation's copy does not cast, and takes one.
    const res = await post(request, "rest/TestOrderability", {
      jsonChangeList: JSON.stringify({
        activateTest: JSON.stringify([{ id: Number(target.id) }]),
        deactivateTest: "[]",
      }),
    });
    expect(res.status(), "a numeric id where a string is cast").toBe(500);
  });
});
