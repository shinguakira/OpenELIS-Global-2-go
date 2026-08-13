# OpenELIS Global 2 — API E2E Test Plan (Java baseline → migration parity oracle)

Goal: a **language-neutral, black-box** e2e suite (Playwright, **API mode — no
browser**) that strictly asserts the behavior of the current **Java** OpenELIS
backend over its **REST API** (+ FHIR R4), including **database state after
writes**. The same suite later runs unchanged against any re-implementation
(e.g. a Go port) — it is the parity oracle for the migration. This is the
OpenELIS analog of `e2e.md` (the OpenMRS plan).

- **Target under test:** OpenELIS Global 2 (`develop`) on Tomcat, PostgreSQL
  (`clinlims`, 375 tables), co-resident HAPI FHIR. Base URL:
  `https://localhost/api/OpenELIS-Global`.
- **REST prefix:** `…/rest/…` (every controller is class-mapped under `/rest`
  via a class-level `@RequestMapping("/rest")`, including `ProviderRestController`
  — there is no separate un-prefixed namespace); `~420` method endpoints across
  `112` controllers. **Confirmed root cause of a real bug**: §7 below used to list
  Provider paths without the `/rest` prefix (e.g. `/Provider/raw/{id}`), which
  is wrong — the real path is `/rest/Provider/raw/{id}`. The b2 Go port copied
  that wrong path verbatim and 404'd on every real call until live-tested
  against Java directly; see `b2-org-provider-migration.md` §3.1. Fixed here —
  §7 now lists the correct `/rest/...` paths. **Lesson for every future
  wave**: before registering ANY Go route, grep the target Java controller
  file for its class-level `@RequestMapping` — never trust a method-level
  `@GetMapping`'s path alone, and never trust this doc's own path listings
  without cross-checking the controller source, since this doc has already
  been wrong once.
- **FHIR prefix:** `…/fhir/*` (CapabilityStatement at `…/fhir/metadata`).
- **Auth:** **session cookie** via Spring Security form login —
  `POST /ValidateLogin?apiCall=true` (form `loginName`/`password`) → `{success:true}`;
  `GET /session` reports `{authenticated, userId, loginName, roles[]}`. Not HTTP
  Basic. (Verified live.)
- **Tool:** `@playwright/test` `APIRequestContext` (`request.get/post/...`) with a
  persistent cookie jar per worker; golden JSON snapshots for response shapes.
- **DB oracle:** `docker exec openelisglobal-database psql -U clinlims -d clinlims`
  to assert row state after writes (the write path is the strictest parity check).
- **Principle — what "parity-verified" actually means, non-negotiable:**
  - **No mocking, ever.** Every assertion in this suite runs against the real,
    live Java webapp (authenticated the same way a real client is) and the
    real, live Postgres it's connected to — never a stub, never a fixture
    substituted for the real HTTP round-trip. A Go endpoint that only ever
    got curled in isolation, or checked against assumed/remembered behavior,
    is **not verified** — it's a guess that happens to compile.
  - **Assert real field-by-field data, not just HTTP status.** `200 OK` proves
    reachability, nothing else — it would not have caught any of the b2
    findings (a wrong-path 404 is easy to notice, but a subtly wrong field, a
    wrong divergence, a silently dropped key would sail through a
    status-only check). Compare actual response bodies, field by field,
    against Java's actual response for the actual same input.
  - Still: assert on *contract + behavior*, not volatile fields
    (`lastupdated`, generated ids/uuids, accession numbers). Normalize/strip
    those before snapshotting — but don't drop them from the *shape* check,
    only from *exact-value* pinning (b2 found real bugs — a missing
    `lastupdated` field entirely — that a shape-blind snapshot would miss).
  - **Mine the target controller's own JUnit test first (§16), every time,
    not as an optional nice-to-have.** It's the fastest way to see Java's
    real URL, real request shape, and real edge cases before writing a
    single line of Go — the b2 routing bug (this doc's own §7 had the wrong
    path) would very likely have been caught immediately by reading
    `ProviderRestControllerTest.java` first, which was skipped.

---

## 0. Cross-cutting concerns (assert broadly, not per-endpoint)

> **Verified baseline contract (surprising, but real):** an unauthenticated call to
> a protected `/rest/**` endpoint returns **HTTP 200 with the login HTML page**
> (~11,880 bytes of `<!DOCTYPE html>…`, identical for every endpoint) — *not* a
> 401/403 JSON. Authenticated calls return JSON. So the auth discriminator the
> suite asserts is **"login-HTML body vs JSON data"**, not the status code. A
> migration target must replicate this (or the change must be explicit). Confirmed
> in `tests/readonly/00-crosscutting.spec.ts`.

- [ ] **Unauthenticated protected endpoint → login HTML (200), never JSON data**
      on `/rest/**` and `/Provider/**` (no session). Assert body starts with
      `<!DOCTYPE html` and leaks no records.
- [ ] **Public whitelist stays public** (no auth): `/rest/open-configuration-properties`,
      `/rest/site-branding/**`, `/rest/supportedlocales/active`, `/health/**`.
- [ ] **Authenticated session** — after `ValidateLogin`, `GET /session` →
      `authenticated:true`, correct `userId`/`loginName`, expected `roles[]`.
- [ ] **Bad credentials** — `ValidateLogin` with wrong password → `{success:false}` / not authenticated.
- [ ] **Logout / session invalidation** — after logout, `/session` → `authenticated:false`.
- [ ] **Role authorization (`@PreAuthorize`)** — a low-privilege user gets **403** on
      admin/EQA-gated endpoints (92 `@PreAuthorize` across 68 controllers). Assert
      both allow (correct role) and deny (wrong role) — the *inversion test*.
- [ ] **CSRF** — Spring `CsrfToken` is exposed via `/session`; assert whether state-
      changing `POST/PUT/DELETE` require it (send/omit and compare). Document the
      real enforcement (login succeeded without it in the probed build — verify per route).
- [ ] **Content-Type** `application/json` on REST reads; correct HTTP status codes.
- [ ] **Unknown id → 404** (not 500) on `…/{id}` reads.
- [ ] **Malformed/invalid body → 400** with an error payload (not 500).
- [ ] **Voided/inactive hiding** — inactive tests, voided samples, retired orgs not
      returned by default; explicit include flag returns them.
- [ ] **Idempotency** — every mutating test cleans up or runs against a freshly
      restored DB (see §17) so the suite is re-runnable.

---

## 1. Session, auth & bootstrap  `/session`, `/ValidateLogin`, `/rest/*configuration*`
- [ ] `GET /session` unauthenticated → `authenticated:false` + `sessionId`.
- [ ] `POST /ValidateLogin?apiCall=true` (admin) → `{success:true}`; session cookie set.
- [ ] `GET /session` authenticated → `userId=1`, `loginName=admin`, `roles[]` incl. `Global Administrator`.
- [ ] `GET /rest/open-configuration-properties` (public) → non-empty config bootstrap.
- [ ] `GET /rest/configuration-properties` (authed) → display-list/dropdown data.
- [ ] `GET /rest/menu`, `/rest/menu/{elementId}`, `/rest/admin/menu/{elementId}` → menu tree.
- [ ] `GET /rest/supportedlocales/active` (public) and `/rest/server-time`.

## 2. Test Catalog & configuration  (reference data — referenced by everything)
- [ ] `GET /rest/test-sample-types`, `/rest/TestCatalog` → catalog shape stable & non-empty.
- [ ] `GET /rest/uom`, `/rest/dictionary-categories`, `/rest/Dictionary` (admin) → lists.
- [ ] **Dictionary CRUD (admin):** `POST /rest/Dictionary` create → `GET` back → matches
      → `POST /rest/DeleteDictionary`; **DB:** row in `dictionary` created/removed.
- [ ] **Management menus:** `GET+POST` on `TestManagementConfigMenu`, `SampleTypeManagement`,
      `TestSectionManagement`, `PanelManagement`, `MethodManagement`, `UomManagement`.
- [ ] **Test creation:** `POST /rest/TestAdd` → new test; **DB:** `test` row + links
      (`test_result`, `test_section`, sample-type mappings). `POST /rest/TestModifyEntry`,
      `/TestRenameEntry`, `/TestActivation`, `/TestOrderability` → **DB** reflects flags.
- [ ] **Panels:** `PanelCreate` / `PanelOrder` / `PanelTestAssign` → **DB** `panel`,
      `panel_item`. **Sample types:** `SampleTypeCreate/Order/TestAssign`.
- [ ] `POST /rest/tests/{testId}/activate` (TestCatalogActivation) → **DB** `test.is_active`.
- [ ] Reflex/calc/alert rules: `TestReflexRule`, `CalculatedValue`, `TestAlertRule` read+write.
- [ ] **403** for non-admin on all `hasRole('ADMIN')` config endpoints.

## 3. Sample / Order entry  ← highest-value write path
- [ ] `GET /rest/SamplePatientEntry` → form/config shape.
- [ ] **Create order:** `POST /rest/SamplePatientEntry` (patient + sample + tests) →
      success; **DB:** new `sample`, `sample_human`, `sample_item`, `analysis` rows;
      accession number generated per site format.
- [ ] `GET /rest/all-by-accession/{accessionNumber}` / `/rest/unassigned-by-accession/{n}`
      → returns the created order's items.
- [ ] `GET /rest/SampleEdit` + `POST /rest/SampleEdit` (modify order) → **DB** updated.
- [ ] `GET /rest/patientByLabNumer?...` lookup by accession.
- [ ] **Generic sample order:** `POST /rest/GenericSampleOrder`, `PUT /{accessionNumber}`,
      `POST /rest/GenericSampleOrder/validate`, `POST /rest/GenericSampleOrder/import`
      (multipart) → **DB** rows; validation returns 400 on bad payload.
- [ ] **Batch entry:** `POST /rest/SampleBatchEntry`, `/rest/SamplePatientEntryBatch` →
      **DB** multiple orders in one call.
- [ ] `GET /rest/dashboard`, `/rest/search` (OrderSearch) → counts reflect created orders.
- [ ] Order attachments: `OrderAttachmentRestController` CRUD (5 endpoints) → **DB** attachment rows.

## 4. Patient / Person
- [ ] **Create/update patient:** `POST /rest/PatientManagement` (person+identifiers) →
      **DB:** `person`, `patient`, `patient_identity` rows; identifier uniqueness enforced.
- [ ] `GET /rest/patient-id-documents/{patientId}` list; `PUT`/`DELETE` a document → **DB**.
- [ ] `GET /rest/patient-photos/{id}/{isThumbnail}` (binary) → correct content-type.
- [ ] **Patient merge:** `PatientMergeRestController` (3 endpoints) → **DB:** source
      `patient.is_merged=true`, `merged_into_patient_id` set; orders re-pointed.
- [ ] Invalid/duplicate identifier → 400.

## 5. Results / result entry
- [ ] `GET /rest/LogbookResults?testUnit=...` → pending analyses for a test unit.
- [ ] **Enter results:** `POST /rest/LogbookResults` → **DB:** `result` rows created,
      `analysis.status` advanced; value serialized per `type_of_test_result` (numeric/dictionary/text).
- [ ] `GET /rest/accession-results` (AccessionResults) → results for one order.
- [ ] Workplan reads: `WorkplanByTest` / `ByTestSection` / `ByPanel` / `ByPriority`,
      `PrintWorkplanReport` → correct worklists.
- [ ] Analyzer test-name mapping: `GET+POST /rest/AnalyzerTestName`, `AnalyzerTestNameMenu`,
      `DeleteAnalyzerTestName` → **DB** mapping rows.

## 6. Result validation
- [ ] `GET /rest/AccessionValidation` → results awaiting validation.
- [ ] **Validate:** `POST /rest/AccessionValidation` (accept/reject/refer) → **DB:**
      `analysis.status`/`result` transitions; released results become reportable.
- [ ] Reject path → **DB** status = rejected; refer path creates referral rows.

## 7. Organization / Provider
- [x] `GET /rest/organization-list`, `/rest/organization/{id}`,
      `/rest/organization/types`, `/rest/organization/generate-site-code`,
      `/rest/departments-for-site` — b2, live-verified both sides, see
      `b2-org-provider-migration.md`.
- [ ] `/rest/organization/search` (paginated Type-C search, own group).
- [ ] **Org CRUD:** `POST /rest/Organization` create → `GET` back → `GET /rest/CancelOrganization`;
      **DB:** `organization` row; retire hides by default.
- [x] `GET /rest/Provider/raw/{id}`, `/rest/Provider/Person/{id}`,
      `/rest/provider/search`, `/rest/practitioner` — b2, live-verified both
      sides. **Paths corrected**: previously listed here without the `/rest`
      prefix (wrong — see the REST-prefix note above); the b2 Go
      implementation copied that wrong path and 404'd on every real call
      until caught by live testing against Java.
- [ ] `POST /rest/Provider/FhirUuid` → **DB** `provider`.

## 8. Reports & audit
- [ ] `POST /rest/ReportPrint` (routine report) → returns report artifact (PDF/data), 200.
- [ ] TAT report (`/rest/tat/...`, 4 endpoints) → turnaround metrics.
- [ ] **Audit (admin):** `AuditTrailReport` (3) + `SystemAuditEvent` (4) →
      history rows; **403** for non-admin.

## 9. Storage / ColdStorage / Shipment / Inventory  (rich RESTful CRUD)
- [ ] **Storage hierarchy** (rooms/devices/shelves/racks/boxes) read + create + assign;
      **DB:** `storage_room/…/storage_box`, `sample_storage_assignment`.
- [ ] **Shipment** (`/rest/shipment/*`, 21 endpoints): `GET /`, `/{id}`, `/by-state/{state}`,
      `POST /` create box, `PUT /{id}/state`, `PUT /{id}/archive`, `GET /{id}/manifest/pdf`,
      `POST /import-from-fhir`, `GET /statistics`; **DB** box + box_sample rows.
- [ ] `BoxSample` (17) + `UnassignedSample` (8) endpoints — assign/unassign samples.
- [ ] **Freezer monitoring:** `/rest/coldstorage/reports`, `/rest/coldstorage/audit-trail`.
- [ ] **Inventory** (Lot 17 / Item 13 / Location 11 / Transaction 5 / Usage 5 / Mgmt 4):
      full CRUD lifecycle with **DB** assertions on each.

## 10. Referral / Electronic Orders
- [ ] `GET /rest/ReferredOutTests` → referred tests list.
- [ ] `GET /rest/ElectronicOrders` (incoming e-orders inbox) → pending FHIR ServiceRequests;
      accept flow creates a lab order (**DB** sample from e-order).

## 11. EQA (role-gated)
- [ ] `GET /rest/eqa/programs/{id}`, `/{id}/tests` (needs `RECEPTION`/`RESULTS`).
- [ ] `PUT /rest/eqa/programs/{id}` mutations need `Global Administrator`; distributions
      need `EQA Coordinator` → **403** without role, 200 with.
- [ ] Enrollment / Orders / Result / Submission / MyPrograms / Alert dashboards read.

## 12. QC / QAEvent / NCE
- [ ] `GET /rest/nonconformevents`; `GET+POST /rest/reportnonconformingevent`;
      `POST /rest/reportnonconformingevent/with-attachments` (multipart) → **DB** NCE rows.
- [ ] NCE enhancement dashboard (7) + correction action (3) + view (2) read/write.

## 13. Admin / configuration
- [ ] **Users:** `GET /rest/users`, `/rest/users/{roleName}`, `GET+POST /rest/UnifiedSystemUser`
      → **DB** `system_user`, `system_user_role`; login as created user → role checks.
- [ ] **External connections (admin):** `GET+POST /rest/ExternalConnection`,
      `DeactivateExternalConnection` → **DB**.
- [ ] **FHIR data export (admin):** `GET /rest/DataExportStatus`, `/{taskId}/attempts`,
      `POST /{taskId}/trigger` → task queue state.
- [ ] **Site info / branding / logo / localization / calendar / alerts / notification config /
      label presets / result-reporting config / notebook** — read + write with **DB** checks;
      role gates asserted.

## 14. FHIR R4 API  `…/fhir/*`
- [ ] `GET /fhir/metadata` → `CapabilityStatement` lists supported resources/interactions.
- [ ] `Patient`, `ServiceRequest` (orders), `DiagnosticReport`/`Observation` (results),
      `Task`, `Specimen`, `Practitioner`, `Organization`: read (`GET /fhir/{Resource}/{id}`,
      search) return valid R4 resources with `resourceType`, `id`, `meta`.
- [ ] **Interop parity:** create an order via REST → the corresponding FHIR
      `ServiceRequest`/`Specimen` appears; enter+validate results → `DiagnosticReport`.
- [ ] Unknown resource id → FHIR `OperationOutcome` 404 (not 500).

---

## 15. DB restoration / fixtures strategy  ("each db restoration")

The suite must control DB state so writes are deterministic and re-runnable.

- **Seed profiles:** `./src/test/resources/load-test-fixtures.sh --profile=core|harness`
  (foundational providers/orgs, storage hierarchy, demo patients/accessions).
- **Reset before a mutating group:** `--reset` wipes and reloads a clean baseline.
- **Per-test isolation options (pick per group):**
  - **Restore-around:** snapshot → run mutation test → restore. Fastest correctness:
    `pg_dump -n clinlims` a baseline once, `psql` restore between destructive groups.
  - **Create-then-cleanup:** each test creates with a unique marker
    (e.g. name `E2E-<uuid>`) and deletes/voids in teardown — keeps the suite additive.
  - **Read-only groups** (catalog, dashboards, metadata) need no restoration.
- **DB assertions helper:** a thin wrapper running
  `docker exec openelisglobal-database psql -U clinlims -d clinlims -tA -c "<sql>"`
  and returning parsed rows, e.g. `assertRowCount('sample', where, n)`,
  `assertRow('result', {analysis_id}, {value})`.
- **Golden baseline counts** (post `core`+`harness`, for sanity): `test=189`,
  `dictionary=701`, `type_of_sample=15`, `test_result=337`, `patient≈3+`,
  `sample≈23`, `analysis≈14`, `clinlims` tables `=375`.

---

## 16. Mining the JUnit tests  ("partly copy of JUnit kind of")

OpenELIS already ships **~488 test Java files / 90,889 LOC**, many
`*RestControllerTest.java` extending `BaseWebContextSensitiveTest` with DBUnit
XML fixtures under `src/test/resources/testdata/`. **These are the behavioral
spec — port their assertions to API calls:**

- Per-controller tests → the request/response shape + status codes for that endpoint
  (e.g. `AnalyzerPluginConfigRestControllerTest`, `OrganizationRestControllerTest`,
  `AddressHierarchyRestControllerTest`, storage/report `Base*RestControllerTest`).
- **Security tests** (`AnalyzerSecurityTest`, `AnalyzerPluginConfigRestControllerSecurityTest`)
  → the exact 401/403 authorization matrix.
- **DBUnit datasets** (`testdata/*.xml`) → known-state fixtures to translate into
  seed SQL or load directly, giving deterministic inputs identical to the JUnit run.
- Service/DAO integration tests → invariants to assert via the API + DB oracle
  (cascade, audit fields, status transitions).

Workflow: for each domain, open its `*RestControllerTest`, lift the arrange
(fixture) + assert (expected JSON/status/DB), and re-express as a Playwright
`request.*` call + snapshot + `psql` check. Same inputs, same expectations,
language-neutral.

---

## 17. Harness design (Playwright, API mode)

```
openelis-e2e/
  playwright.config.ts     # projects: setup(login) → readonly → mutating(serial)
  fixtures/
    auth.ts                # ValidateLogin, returns storageState (cookie jar)
    db.ts                  # psql helper: query(), assertRow(), reset(), seed(profile)
    normalize.ts           # strip lastupdated/ids/accession before snapshot
  tests/
    00-crosscutting.spec.ts  # §0 auth/authz/csrf/404/400 matrix
    01-session.spec.ts
    02-testcatalog.spec.ts
    03-sample-order.spec.ts  # serial; reset before
    04-patient.spec.ts
    05-results.spec.ts
    06-validation.spec.ts
    07-organization-provider.spec.ts
    08-reports-audit.spec.ts
    09-storage-shipment-inventory.spec.ts
    10-referral-eorders.spec.ts
    11-eqa.spec.ts
    12-qc-nce.spec.ts
    13-admin.spec.ts
    14-fhir.spec.ts
    __snapshots__/           # golden JSON (contract shapes)
```

- **No browser:** use `request` (`APIRequestContext`) only; `ignoreHTTPSErrors:true`
  for the dev cert; `baseURL = https://localhost/api/OpenELIS-Global`.
- **Auth fixture:** `POST /ValidateLogin?apiCall=true`, persist cookies via
  `storageState`; a second low-privilege user context for authz/inversion tests.
- **Mutating specs run serially** (`--workers=1`, `test.describe.serial`) against a
  reset DB; read-only specs run parallel.
- **CI:** run against the same `docker compose` stack; gate on parity snapshots +
  DB assertions. The identical suite is later pointed at the Go port's base URL.

---

## 18. Execution order & coverage tracking

1. **Cross-cutting (§0) + session (§1) + read-only reference (§2, §7 reads, §14 metadata)**
   — no writes, immediate assurance the harness + auth are correct.
2. **Read-only dashboards/reports** (§5 reads, §8, §9 reads, §11 reads).
3. **Mutating CRUD lifecycles** (§3 order → §4 patient → §5 result entry → §6 validation →
   §7 org/provider → §9 storage/shipment/inventory → §12 NCE → §13 admin) — each on a
   reset/seeded DB, each with DB-state assertions. These are the strictest parity checks.
4. **FHIR interop parity** (§14) — REST-created order surfaces as FHIR resources.

**Coverage ledger:** track `covered / total` per domain against the ~420 endpoints.
Target order = highest-value clinical write paths first (order → result → validation),
then reference/config, then admin. Endpoints requiring the analyzer mock/bridge
(harness lane) are out of scope for this API oracle unless that infra is started.

---

*Companion: `openelis-api-e2e/` (suite), `baseline-performance.md`,
`OpenELIS-Go-Migration-Plan.md`. Auth flow, REST prefix, FHIR mount, and DB
oracle all verified live on commit `0c6a4a62d`.*
