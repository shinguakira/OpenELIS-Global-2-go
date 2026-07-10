// One-off: capture screenshots of primary screens for the user manual.
// Run from frontend/: node playwright/manual-shots.mjs
import { chromium } from "@playwright/test";
import fs from "fs";
import path from "path";

const BASE = process.env.BASE_URL || "https://localhost";
const OUT = path.resolve("../docs/img/user-manual");
fs.mkdirSync(OUT, { recursive: true });

// [slug, route, waitMs] — primary, param-less screens grouped by workflow
const SCREENS = [
  ["01-dashboard", "/"],
  ["02-add-order", "/SamplePatientEntry"],
  ["03-patient-management", "/PatientManagement"],
  ["04-patient-history", "/PatientHistory"],
  ["05-modify-order", "/ModifyOrder"],
  ["06-batch-order-entry", "/SampleBatchEntrySetup"],
  ["07-patient-results", "/PatientResults"],
  ["08-logbook-results", "/LogbookResults"],
  ["09-analyzer-results", "/AnalyzerResults"],
  ["10-accession-results", "/AccessionResults"],
  ["11-status-results", "/StatusResults"],
  ["12-range-results", "/RangeResults"],
  ["13-result-validation", "/ResultValidation"],
  ["14-accession-validation", "/AccessionValidation"],
  ["15-referred-out-tests", "/ReferredOutTests"],
  ["16-electronic-orders", "/ElectronicOrders"],
  ["17-print-barcode", "/PrintBarcode"],
  ["18-storage-rooms", "/Storage/rooms"],
  ["19-storage-devices", "/Storage/devices"],
  ["20-storage-racks", "/Storage/racks"],
  ["21-storage-boxes", "/Storage/boxes"],
  ["22-freezer-monitoring", "/FreezerMonitoring"],
  ["23-sample-shipment", "/SampleShipment"],
  ["24-routine-reports", "/RoutineReports"],
  ["25-report", "/Report"],
  ["26-cytology-dashboard", "/CytologyDashboard"],
  ["27-pathology-dashboard", "/PathologyDashboard"],
  ["28-immunohistochemistry-dashboard", "/ImmunohistochemistryDashboard"],
  ["29-notebook-dashboard", "/NoteBookDashboard"],
  ["30-eqa-management", "/EQAManagement"],
  ["31-nce-dashboard", "/NceDashboard"],
  ["32-alerts", "/Alerts"],
  ["33-master-lists", "/MasterListsPage"],
  ["34-audit-trail-report", "/AuditTrailReport"],
];

const authFile = "playwright/.auth/user.json";
const hasAuth = fs.existsSync(authFile);

const browser = await chromium.launch({
  args: ["--ignore-certificate-errors"],
});
const ctx = await browser.newContext({
  ignoreHTTPSErrors: true,
  viewport: { width: 1440, height: 900 },
  ...(hasAuth ? { storageState: authFile } : {}),
});
const page = await ctx.newPage();

if (!hasAuth) {
  await page.goto(BASE + "/login", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2500);
  await page.fill(
    'input[placeholder="Username"]',
    process.env.TEST_USER || "admin",
  );
  await page.fill(
    'input[placeholder="Password"]',
    process.env.TEST_PASS || "adminADMIN!",
  );
  await page.click('button:has-text("Login")');
  await page.waitForTimeout(6000);
}

const results = [];
for (const [slug, route] of SCREENS) {
  try {
    await page.goto(BASE + route, {
      waitUntil: "domcontentloaded",
      timeout: 45000,
    });
    await page.waitForTimeout(3500); // let SPA render
    const title = await page.title();
    await page.screenshot({
      path: path.join(OUT, slug + ".png"),
      fullPage: false,
    });
    results.push(`OK   ${slug}  <- ${route}  (${title})`);
  } catch (e) {
    results.push(`FAIL ${slug}  <- ${route}  :: ${e.message.split("\n")[0]}`);
  }
}
console.log(results.join("\n"));
await browser.close();
