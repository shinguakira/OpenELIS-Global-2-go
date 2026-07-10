// DB oracle — run SQL against the OpenELIS PostgreSQL and assert row state.
// Portable: uses `docker exec` directly on CI/Linux, or wraps via wsl.exe on
// this Windows dev host (where the Docker daemon lives inside WSL).
import { execFileSync } from "node:child_process";
import { expect } from "@playwright/test";
import { DB_CONTAINER, DOCKER_MODE, WSL_DISTRO } from "./env";

function run(sql: string): string {
  const psql = [
    "docker",
    "exec",
    "-i",
    DB_CONTAINER,
    "psql",
    "-U",
    "clinlims",
    "-d",
    "clinlims",
    "-tA", // unaligned + tuples-only; default field separator is already "|"
    "-c",
    sql,
  ];
  try {
    if (DOCKER_MODE === "docker") {
      return execFileSync(psql[0], psql.slice(1), { encoding: "utf8" });
    }
    // wsl.exe -d <distro> -- docker exec ... psql ... -c "<sql>"
    return execFileSync("wsl.exe", ["-d", WSL_DISTRO, "--", ...psql], {
      encoding: "utf8",
    });
  } catch (e: any) {
    throw new Error(
      `DB query failed: ${sql}\n${e.stderr?.toString?.() || e.message}`,
    );
  }
}

/** Run SQL, return rows as arrays of column strings. */
export function query(sql: string): string[][] {
  const out = run(sql).trim();
  if (out === "") return [];
  return out.split("\n").map((line) => line.split("|"));
}

/** SELECT count(*) FROM <schema-qualified table> [WHERE ...] */
export function count(table: string, where?: string): number {
  const sql = `SELECT count(*) FROM clinlims.${table}${where ? " WHERE " + where : ""};`;
  return parseInt(query(sql)[0]?.[0] ?? "0", 10);
}

export function assertCount(table: string, where: string, expected: number) {
  expect(count(table, where), `${table} WHERE ${where}`).toBe(expected);
}

/** Baseline sanity numbers used by read-only smoke checks. */
export const BASELINE = {
  schemaTables: 375,
  test: 189,
  dictionary: 701,
  typeOfSample: 15,
};
