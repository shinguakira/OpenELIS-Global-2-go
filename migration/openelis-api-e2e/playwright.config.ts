import { defineConfig } from "@playwright/test";
import { BASE_URL, AUTH_STATE } from "./fixtures/env";

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
  ],
});
