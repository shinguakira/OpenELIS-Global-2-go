/**
 * e2 — the three *Order screens: PanelOrder, SampleTypeOrder, TestSectionOrder.
 *
 * They move sort_order and write nothing else. What makes them worth a spec is
 * the request shape: jsonChangeList is DOUBLE-ENCODED. The submitted field is a
 * JSON string; parsing it yields an object whose value under the screen's key
 * is ITSELF a JSON string, parsed again to reach the array of {id, sortOrder}.
 *
 * A body that will not parse at either level is not an error —
 * getActivateSetForActions catches ParseException, logs at DEBUG and returns an
 * empty list, so the request is a 200 that wrote nothing.
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

const SCREENS = [
  {
    path: "PanelOrder",
    formName: "panelOrderForm",
    listKey: "panelList",
    changeKey: "panels",
    table: "clinlims.panel",
    auditTable: "PANEL",
    payloadField: "sortOrderInt",
  },
  {
    path: "SampleTypeOrder",
    formName: "sampleTypeOrderForm",
    listKey: "sampleTypeList",
    changeKey: "sampleTypes",
    table: "clinlims.type_of_sample",
    auditTable: "TYPE_OF_SAMPLE",
    payloadField: "sortOrder",
  },
  {
    path: "TestSectionOrder",
    formName: "testSectionOrderForm",
    listKey: "testSectionList",
    changeKey: "testSections",
    table: "clinlims.test_section",
    auditTable: "TEST_SECTION",
    payloadField: "sortOrderInt",
  },
] as const;

function sortOrderOf(table: string, id: string): string {
  return query(
    `SELECT COALESCE(sort_order::text, '<null>') FROM ${table} WHERE id = ${Number(id)}`,
  )[0][0];
}

test.describe("e2 — the Order screens move sort_order", () => {
  let restore: { table: string; id: string; sortOrder: string }[] = [];

  test.beforeEach(() => {
    restore = [];
  });

  test.afterEach(() => {
    // These are shared reference rows every read wave sorts by, so a sort order
    // left moved changes other specs.
    for (const r of restore) {
      exec(
        `UPDATE ${r.table} SET sort_order = ${r.sortOrder === "<null>" ? "NULL" : r.sortOrder}
          WHERE id = ${Number(r.id)}`,
      );
    }
  });

  for (const s of SCREENS) {
    test(`GET ${s.path} — the form`, async ({ request }) => {
      const f = await readJson(await request.get(`rest/${s.path}`), s.path);

      expect(f.formName, "formName").toBe(s.formName);
      expect(f.formMethod, "formMethod").toBe("POST");
      expect(f.cancelAction, "cancelAction").toBe("Home");
      expect(f.submitOnCancel, "submitOnCancel").toBe(false);
      expect(f.cancelMethod, "cancelMethod").toBe("POST");

      expect(Array.isArray(f[s.listKey]), `${s.listKey} is a list`).toBe(true);
      expect(f[s.listKey].length, "and it is populated").toBeGreaterThan(0);
      expect(
        Object.keys(f[s.listKey][0]).sort(),
        "each entry is an id/value pair",
      ).toEqual(["id", "value"]);

      // Initialised to "" on the bean, so present rather than absent.
      expect(f.jsonChangeList, "jsonChangeList default").toBe("");
    });

    test(`POST ${s.path} — the change list is double-encoded`, async ({
      request,
    }) => {
      const f = await readJson(await request.get(`rest/${s.path}`), s.path);
      const target = f[s.listKey][0];
      const before = sortOrderOf(s.table, target.id);
      restore.push({ table: s.table, id: target.id, sortOrder: before });

      // The inner array is a STRING inside the outer object. Sending it as a
      // real array is the shape a reasonable client would use, and Java reads
      // nothing from it — see the refusal test below.
      const inner = JSON.stringify([{ id: Number(target.id), sortOrder: 424 }]);
      const res = await post(request, `rest/${s.path}`, {
        jsonChangeList: JSON.stringify({ [s.changeKey]: inner }),
      });
      expect(res.status(), "reorder").toBe(200);

      expect(sortOrderOf(s.table, target.id), "the sort order moved").toBe("424");
    });

    test(`POST ${s.path} — an inner ARRAY is a 500`, async ({
      request,
    }) => {
      const f = await readJson(await request.get(`rest/${s.path}`), s.path);
      const target = f[s.listKey][0];
      const before = sortOrderOf(s.table, target.id);
      restore.push({ table: s.table, id: target.id, sortOrder: before });

      // `String action = (String) root.get(key)` sits OUTSIDE the try block
      // that catches ParseException, so a real array throws
      // ClassCastException and nothing catches it. The shape a reasonable
      // client would send is the one that fails.
      const res = await post(request, `rest/${s.path}`, {
        jsonChangeList: JSON.stringify({
          [s.changeKey]: [{ id: Number(target.id), sortOrder: 425 }],
        }),
      });
      expect(res.status(), "a nested array, not a nested string").toBe(500);
      expect(sortOrderOf(s.table, target.id), "and nothing was written").toBe(
        before,
      );
    });

    test(`POST ${s.path} — a four-digit sort order is refused, at 200`, async ({
      request,
    }) => {
      const f = await readJson(await request.get(`rest/${s.path}`), s.path);
      const target = f[s.listKey][0];
      const before = sortOrderOf(s.table, target.id);
      restore.push({ table: s.table, id: target.id, sortOrder: before });

      // validateFieldAndCharset(..., 3, "0-9") — at most three digits. A
      // rejection is a 200 carrying the form back, not an error status.
      const res = await post(request, `rest/${s.path}`, {
        jsonChangeList: JSON.stringify({
          [s.changeKey]: JSON.stringify([
            { id: Number(target.id), sortOrder: 4242 },
          ]),
        }),
      });
      expect(res.status(), "refused by the validator").toBe(200);
      expect(sortOrderOf(s.table, target.id), "and nothing was written").toBe(
        before,
      );
    });

    test(`POST ${s.path} — the reorder IS audited`, async ({ request }) => {
      const f = await readJson(await request.get(`rest/${s.path}`), s.path);
      const target = f[s.listKey][0];
      const before = sortOrderOf(s.table, target.id);
      restore.push({ table: s.table, id: target.id, sortOrder: before });

      exec(
        `DELETE FROM clinlims.history WHERE timestamp > now() - interval '30 seconds'`,
      );
      const inner = JSON.stringify([{ id: Number(target.id), sortOrder: 426 }]);
      await post(request, `rest/${s.path}`, {
        jsonChangeList: JSON.stringify({ [s.changeKey]: inner }),
      });

      // Unlike the renames and the UOM writes, these ARE audited: one row per
      // row moved, carrying the order it replaced. The payload names the BEAN
      // PROPERTY, and the three screens do not agree on it — Panel and
      // TestSection emit sortOrderInt, TypeOfSample emits sortOrder.
      const audit = query(
        `SELECT h.activity, COALESCE(convert_from(h.changes, 'UTF8'), '')
           FROM clinlims.history h
           JOIN clinlims.reference_tables rt ON rt.id = h.reference_table
          WHERE upper(rt.name) = '${s.auditTable}'
            AND h.reference_id = ${Number(target.id)}
            AND h.timestamp > now() - interval '30 seconds'`,
      );
      expect(audit.length, "one history row for the row moved").toBe(1);
      expect(audit[0][0], "recorded as an update").toBe("U");
      expect(audit[0][1], "carrying the order it replaced").toBe(
        `<${s.payloadField}>${before}</${s.payloadField}>`,
      );

      exec(
        `DELETE FROM clinlims.history WHERE timestamp > now() - interval '30 seconds'`,
      );
    });
  }
});
