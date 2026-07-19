// Shared environment/config for the OpenELIS API parity suite.
// NOTE: trailing slash is required — Playwright resolves request paths with
// `new URL(path, baseURL)`, so paths must be relative (no leading slash) to keep
// the /api/OpenELIS-Global prefix.
export const BASE_URL =
  process.env.OE_BASE_URL || "https://localhost/api/OpenELIS-Global/";

// The Go port under test, for side-by-side parity: the SAME checks that run
// against Java (BASE_URL) also run against this base. Trailing slash required.
export const GO_BASE_URL = process.env.OE_GO_URL || "http://localhost:8090/";

export const ADMIN_USER = process.env.OE_ADMIN_USER || "admin";
export const ADMIN_PASS = process.env.OE_ADMIN_PASS || "adminADMIN!";

// Container name of the PostgreSQL DB (for the DB oracle).
export const DB_CONTAINER =
  process.env.OE_DB_CONTAINER || "openelisglobal-database";

// How to reach `docker` from the test runner.
//  - On Linux/CI where docker is on PATH:        OE_DOCKER="docker"
//  - On this Windows dev host (docker in WSL):   default → wsl.exe wrapper
export const DOCKER_MODE = process.env.OE_DOCKER || "wsl"; // "wsl" | "docker"
export const WSL_DISTRO = process.env.OE_WSL_DISTRO || "Ubuntu-24.04";

export const AUTH_STATE = "playwright/.auth/admin.json";
