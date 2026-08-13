import { defineConfig } from "@playwright/test";
import { BASE_URL, GO_BASE_URL, AUTH_STATE } from "./fixtures/env";

// API-only parity suite for OpenELIS — no browser is launched; every test uses
// the `request` (APIRequestContext) fixture against the live REST/FHIR surface.
export default defineConfig({
  testDir: "./tests",
  timeout: 30_000,
  expect: { timeout: 10_000 },
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  reporter: [["list"]],
  use: {
    baseURL: BASE_URL,
    ignoreHTTPSErrors: true,
  },
  projects: [
    { name: "setup", testMatch: /auth\.setup\.ts/ },
    {
      name: "api-readonly",
      testMatch: /readonly[\\/].*\.spec\.ts/,
      dependencies: ["setup"],
      use: { storageState: AUTH_STATE },
    },
    {
      name: "api-mutating",
      testMatch: /mutating[\\/].*\.spec\.ts/,
      dependencies: ["setup"],
      use: { storageState: AUTH_STATE },
      workers: 1,
    },
    // Go port, side-by-side parity: the SAME assertions from the ported migration
    // units run against the Go service. No auth setup (Go needs none), anon cookie
    // jar. Only ported units are listed in testMatch — add each unit here as it
    // passes against Go. Run with:
    //   npx playwright test --project=go-parity   (requires the Go service up)
    {
      name: "go-parity",
      testMatch: /readonly[\\/](a1-server-time|a2-static-reads|b2-organization|b2-provider)\.spec\.ts/,
      use: {
        baseURL: GO_BASE_URL,
        ignoreHTTPSErrors: true,
        storageState: { cookies: [], origins: [] },
      },
    },
  ],
});
