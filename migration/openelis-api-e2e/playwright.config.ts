import { defineConfig } from "@playwright/test";
import { BASE_URL, GO_BASE_URL, AUTH_STATE, GO_AUTH_STATE } from "./fixtures/env";

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
    // Login handshake against the Go port. Runs the SAME tests/auth.setup.ts as
    // `setup` — only the target and the output jar differ. Its existence is the
    // point: if the Go port ever needed a different login flow, this project
    // would not be able to reuse that file.
    {
      name: "setup-go",
      testMatch: /auth\.setup\.ts/,
      use: {
        baseURL: GO_BASE_URL,
        ignoreHTTPSErrors: true,
        storageState: { cookies: [], origins: [] },
      },
    },
    // Go port, side-by-side parity: the SAME assertions from the ported migration
    // units run against the Go service. Only ported units are listed in
    // testMatch — add each unit here as it passes against Go. Run with:
    //   npx playwright test --project=go-parity   (requires the Go service up)
    //
    // Since P0 auth landed, the Go service is default-deny like Java, so this
    // project authenticates first (setup-go) and carries the resulting cookie
    // jar. p0-auth.spec.ts drives its own anonymous contexts and ignores it.
    {
      name: "go-parity",
      // b1-testcatalog was ported and green long before it appeared here — the
      // ledger entry was simply never added, so nothing guarded b1 against
      // regression. If a spec passes against Go, it belongs in this list.
      //
      // Includes one mutating spec on purpose: b2-organization-sitecode
      // exercises generate-site-code, which consumes a DB sequence, so it
      // cannot live in tests/readonly/ — but the Go port still needs the
      // coverage (it pins the UTC-vs-host-local site-code date fix).
      testMatch:
        /(readonly[\\/](p0-auth|p0-authz|a1-server-time|a2-static-reads|b1-testcatalog|b2-organization|b2-provider|c1-patient-reads|c2-sample-order-reads|c2-sample-form-loads|c3-result-reads)|mutating[\\/](b2-organization-sitecode|c1-patient-edge-cases|e1-config-crud|e1-config-parity-gaps|e2-uom-writes|e2-rename-writes|e2-create-writes|e2-panel-create))\.spec\.ts/,
      dependencies: ["setup-go"],
      use: {
        baseURL: GO_BASE_URL,
        ignoreHTTPSErrors: true,
        storageState: GO_AUTH_STATE,
      },
    },
  ],
});
