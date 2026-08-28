// DB oracle — run SQL against the OpenELIS PostgreSQL and assert row state.
// Portable: uses `docker exec` directly on CI/Linux, or wraps via wsl.exe on
// this Windows dev host (where the Docker daemon lives inside WSL).
import { execFileSync } from "node:child_process";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
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

/** Absolute path of a fixture SQL file, from the repo root. */
export function fixturePath(name: string): string {
  return `src/test/resources/fixtures/${name}`;
}

/**
 * Execute a fixture SQL file against the same database the API is using.
 *
 * Streams the file through psql's stdin rather than `-f <path>`: the file lives
 * on the test runner's filesystem, which on this dev host is NOT the filesystem
 * the database container sees. Runs with ON_ERROR_STOP so a broken fixture
 * fails the test instead of silently half-applying.
 */
export function runSqlFile(relativePath: string): string {
  const sql = readFileSync(resolve(process.cwd(), "..", "..", relativePath), "utf8");
  const psql = [
    "docker", "exec", "-i", DB_CONTAINER,
    "psql", "-U", "clinlims", "-d", "clinlims", "-v", "ON_ERROR_STOP=1", "-f", "-",
  ];
  try {
    if (DOCKER_MODE === "docker") {
      return execFileSync(psql[0], psql.slice(1), { encoding: "utf8", input: sql });
    }
    return execFileSync("wsl.exe", ["-d", WSL_DISTRO, "--", ...psql], {
      encoding: "utf8",
      input: sql,
    });
  } catch (e: any) {
    throw new Error(
      `fixture failed: ${relativePath}\n${e.stderr?.toString?.() || e.message}`,
    );
  }
}

/** Run arbitrary SQL for setup/teardown; throws on error. */
export function exec(sql: string): void {
  query(sql.trimEnd().endsWith(";") ? sql : sql + ";");
}
