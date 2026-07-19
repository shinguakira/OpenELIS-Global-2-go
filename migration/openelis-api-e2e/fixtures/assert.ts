// Shared strict-assertion helpers for the parity oracle.
//
// The suite's job is to prove endpoint BEHAVIOR, not reachability. Asserting a
// response's real shape/values already implies it wasn't the login HTML page, so
// these helpers replace the old "expect(200) + not-login" pattern. Orderings are
// only ever asserted where the Java source guarantees one (explicit Comparator /
// SQL ORDER BY) — never a coincidental DB-natural order.
import { expect, type APIResponse } from "@playwright/test";

/** Assert an OK JSON response and return the parsed body. */
export async function readJson(res: APIResponse, label: string): Promise<any> {
  expect(res.status(), `${label} status`).toBe(200);
  expect(
    (res.headers()["content-type"] ?? "").toLowerCase(),
    `${label} content-type`,
  ).toContain("application/json");
  return res.json();
}

/** Assert an array response: OK JSON, is an array, non-empty. Returns it. */
export async function readJsonArray(res: APIResponse, label: string): Promise<any[]> {
  const body = await readJson(res, label);
  expect(Array.isArray(body), `${label} is an array`).toBe(true);
  expect(body.length, `${label} non-empty`).toBeGreaterThan(0);
  return body;
}

/** Assert an object has EXACTLY these keys (order-independent). Catches both
 *  missing fields and extra fields a port might leak. */
export function expectExactKeys(obj: any, keys: string[], label: string): void {
  expect(Object.keys(obj).sort(), `${label} keys`).toEqual([...keys].sort());
}

/** Assert obj's keys are all within `allowed`, and every `required` key is
 *  present. Use instead of expectExactKeys when a port omits null-valued fields
 *  (Jackson JsonInclude.NON_NULL): the omittable fields go in `allowed` only. */
export function expectKeysWithin(
  obj: any,
  allowed: string[],
  required: string[],
  label: string,
): void {
  const keys = Object.keys(obj);
  for (const k of keys)
    expect(allowed.includes(k), `${label} unexpected key "${k}"`).toBe(true);
  for (const r of required)
    expect(keys.includes(r), `${label} missing required key "${r}"`).toBe(true);
}

/** Assert the values are all distinct. */
export function expectUnique(values: any[], label: string): void {
  expect(new Set(values).size, `${label} unique`).toBe(values.length);
}

/** Assert a sequence is non-decreasing. `caseInsensitive` lowercases strings
 *  before comparing (for Java compareToIgnoreCase / collation-robust checks). */
export function expectAscending(
  values: Array<string | number>,
  label: string,
  opts: { caseInsensitive?: boolean } = {},
): void {
  const norm = (v: string | number) =>
    opts.caseInsensitive && typeof v === "string" ? v.toLowerCase() : v;
  for (let i = 1; i < values.length; i++) {
    expect(
      norm(values[i - 1]) <= norm(values[i]),
      `${label} ascending at index ${i}: "${values[i - 1]}" → "${values[i]}"`,
    ).toBe(true);
  }
}

/** Assert a non-empty string field. */
export function expectNonEmptyString(v: any, label: string): void {
  expect(typeof v, `${label} is string`).toBe("string");
  expect(v.length, `${label} non-empty`).toBeGreaterThan(0);
}
