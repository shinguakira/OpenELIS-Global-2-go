# OpenELIS → Go — Endpoint-by-Endpoint Migration Order

Status: **draft / proposal**
Companion to [OpenELIS-Go-Migration-Plan.md](OpenELIS-Go-Migration-Plan.md) (the
*why*) and [endpoint-migration-taxonomy.md](endpoint-migration-taxonomy.md) (the
*what-kind* — endpoint types A–J, their porting recipe and dependency graph).
Endpoint source of truth: [openelis-api-e2e/fixtures/endpoints.generated.ts](openelis-api-e2e/fixtures/endpoints.generated.ts)
(425 endpoints extracted from Java `controller/rest` source).

## How to read this

Each endpoint is migrated **individually**. The unit of work is ONE row below:

1. Reimplement that one endpoint in Go (same DB, same JSON shape).
2. Run the parity test (`openelis-api-e2e`): hit Java and Go with the same
   request → assert identical response (and, for writes, identical DB rows).
3. Flip **one nginx line** to route that path → Go. Everything else stays Java.
4. Tick the box. Move to the next row.

**Ordering rule:** within a context, **all reads before any write**. Across
contexts, **reference data → identity → order spine → results → interop**. A
context's writes are not attempted until its reads pass parity, because the
write parity test *reads back* through the same endpoints.

Legend: **M** = HTTP verb(s). **Risk**: 🟢 trivial · 🟡 joins/logic · 🔴
clinical-safety or concurrency. Where a path has both GET and POST, the GET is
migrated in the read wave and the POST is listed again in the write wave.

---

## Wave 0 — Smoke: static / system reads (prove the pipe)

No DB joins, no clinical data. The point of this wave is to stand up the Go
service, the proxy route, and the parity harness against the **safest possible**
endpoints. `server-time` is the recommended very first endpoint.

| # | Endpoint | M | Tables / source | Risk |
|---|----------|---|-----------------|------|
| 0.1 | `rest/server-time` | GET | (none — clock) | 🟢 |
| 0.2 | `rest/open-configuration-properties` | GET | site_information (public subset) | 🟢 |
| 0.3 | `rest/configuration-properties` | GET | site_information | 🟢 |
| 0.4 | `rest/math-functions` | GET | (static) | 🟢 |
| 0.5 | `rest/analysis-status-types` | GET | status_of_sample / enum | 🟢 |
| 0.6 | `rest/sample-status-types` | GET | status_of_sample | 🟢 |
| 0.7 | `rest/sample-item-status-types` | GET | status_of_sample | 🟢 |
| 0.8 | `rest/supportedlocales/` | GET | supported_locale | 🟢 |
| 0.9 | `rest/supportedlocales/active` | GET | supported_locale | 🟢 |
| 0.10 | `rest/supportedlocales/fallback` | GET | supported_locale | 🟢 |
| 0.11 | `rest/supportedlocales/{id}` | GET | supported_locale | 🟢 |
| 0.12 | `rest/menu` | GET | menu | 🟡 (tree build) |
| 0.13 | `rest/menu/{elementId}` | GET | menu | 🟡 |
| 0.14 | `rest/admin/menu/{elementId}` | GET | menu | 🟡 |

**Exit:** Go service live behind proxy; 14 routes flipped; parity green; rollback
(flip line back) rehearsed at least once.

### Ops / infrastructure endpoints — outside the wave sequence

Not domain data, don't gate or get gated by anything downstream, so they
don't fit the dependency-driven wave numbering above. Tracked here instead
of inside a wave so they aren't invisible to planning (a real gap this repo
had until it was noticed — see `logging-adoption-plan.md`).

| Endpoint | M | Status | Note |
|---|---|---|---|
| `health` | GET | **done** (Go, `cmd/openelis/main.go`) | Placeholder only — `{"status":"UP"}`, not a port of Java's `/health/odoo` (Odoo-billing connectivity check). |
| `rest/logging`, `rest/logging/stream`, `rest/logging/test` | GET | **not started** — library decision made (zap, [logging-adoption-plan.md](logging-adoption-plan.md)), port itself not yet begun or scheduled | Admin-only. Runtime log-level switch + SSE live tail + test-line emitter. Not blocking anything; port only if/when worth the effort — see that doc's "not yet decided" note. |

---

## Wave 1 — Reference data: Dictionary + Test Catalog (reads)

The most-referenced data in the system. Everything downstream joins to these.

### 1a. Dictionary
| # | Endpoint | M | Tables | Risk |
|---|----------|---|--------|------|
| 1.1 | `rest/dictionary-categories` | GET | dictionary_category | 🟢 |
| 1.2 | `rest/dictionary/category/{categoryName}` | GET | dictionary, dictionary_category | 🟡 |
| 1.3 | `rest/Dictionary` | GET | dictionary | 🟢 |
| 1.4 | `rest/DictionaryMenu` | GET | dictionary | 🟢 |
| 1.5 | `rest/displayList/{listType}` | GET | dictionary (+ many list sources) | 🟡 |
| 1.6 | `rest/paginatedDisplayList/{listType}` | GET | dictionary (+ paging) | 🟡 |

### 1b. Test catalog (read)
| # | Endpoint | M | Tables | Risk |
|---|----------|---|--------|------|
| 1.7 | `rest/test-catalog/tests` | GET | test, test_section | 🟡 |
| 1.8 | `rest/test-catalog/tests/{testId}` | GET | test (+ joins) | 🟡 |
| 1.9 | `rest/test-catalog/tests/{testId}/basic-info` | GET | test | 🟡 |
| 1.10 | `rest/test-catalog/tests/{testId}/ranges` | GET | result_limit, test_result | 🔴 (ref ranges) |
| 1.11 | `rest/test-catalog/tests/{testId}/sample-results` | GET | test_result | 🟡 |
| 1.12 | `rest/test-catalog/tests/{testId}/panels` | GET | panel, panel_item | 🟡 |
| 1.13 | `rest/test-catalog/tests/{testId}/analyzers` | GET | analyzer_test_map | 🟡 |
| 1.14 | `rest/test-catalog/tests/{testId}/siblings` | GET | test | 🟡 |
| 1.15 | `rest/test-catalog/tests/{testId}/localization` | GET | localization | 🟡 |
| 1.16 | `rest/test-catalog/tests/{testId}/terminology` | GET | test (LOINC) | 🟡 |
| 1.17 | `rest/test-catalog/tests/{testId}/loinc-integrity` | GET | test | 🟡 |
| 1.18 | `rest/test-catalog/tests/{testId}/storage` | GET | test storage | 🟡 |
| 1.19 | `rest/test-catalog/tests/{testId}/storage/history/` | GET | storage history | 🟡 |
| 1.20 | `rest/test-catalog/panels` | GET | panel | 🟢 |
| 1.21 | `rest/test-catalog/panels/{panelId}/test-order` | GET | panel_item | 🟡 |
| 1.22 | `rest/test-catalog/sample-types` | GET | type_of_sample | 🟢 |
| 1.23 | `rest/test-catalog/sample-types/{sampleTypeId}/test-order` | GET | type_of_sample_test | 🟡 |
| 1.24 | `rest/test-catalog/lab-units` | GET | test_section | 🟢 |
| 1.25 | `rest/test-catalog/dictionary` | GET | dictionary | 🟢 |
| 1.26 | `rest/test-catalog/group/summary` | GET | test | 🟡 |
| 1.27 | `rest/TestCatalog` | GET | test (legacy shape) | 🟡 |
| 1.28 | `rest/test-list` | GET | test | 🟢 |
| 1.29 | `rest/test-sample-types` | GET | type_of_sample_test | 🟡 |
| 1.30 | `rest/sample-type-tests` | GET | type_of_sample_test | 🟡 |
| 1.31 | `rest/tests-by-sample` | GET | type_of_sample_test | 🟡 |
| 1.32 | `rest/AllTestsForSampleTypeProvider` | GET | type_of_sample_test | 🟡 |
| 1.33 | `rest/TestNamesProvider` | GET | test | 🟢 |
| 1.34 | `rest/EntityNamesProvider` | GET | test/panel names | 🟢 |
| 1.35 | `rest/test-display-beans` | GET | test | 🟡 |
| 1.36 | `rest/test-display-beans-map` | GET | test | 🟡 |
| 1.37 | `rest/test-result-tree` | GET | test_result | 🟡 |
| 1.38 | `rest/result-tree` | GET | test_result | 🟡 |
| 1.39 | `rest/uom` | GET | unit_of_measure | 🟢 |
| 1.40 | `rest/methods-for-test/{testId}` | GET | test_method | 🟡 |
| 1.41 | `rest/test/{testId}/methods/` | GET | test_method | 🟡 |
| 1.42 | `rest/labUnit/config` | GET | test_section config | 🟡 |

**Exit:** all reference reads parity-green. This is the biggest single unblocking
wave — most other endpoints join to this data.

---

## Wave 2 — Organization + Provider (reads)

Referenced by orders and patients.

| # | Endpoint | M | Tables | Risk |
|---|----------|---|--------|------|
| 2.1 | `rest/Organization` | GET | organization | 🟢 |
| 2.2 | `rest/organization-list` | GET | organization | 🟢 |
| 2.3 | `rest/organization/{id}` | GET | organization | 🟢 |
| 2.4 | `rest/organization/search` | GET | organization | 🟡 |
| 2.5 | `rest/organization/types` | GET | organization_type | 🟢 |
| 2.6 | `rest/organization/generate-site-code` | GET | organization | 🟡 |
| 2.7 | `rest/OrganizationMenu` | GET | organization | 🟢 |
| 2.8 | `rest/OrganizationExport` | GET | organization | 🟡 |
| 2.9 | `rest/SearchOrganizationMenu` | GET | organization | 🟡 |
| 2.10 | `provider/search` | GET | provider, person | 🟡 |
| 2.11 | `rest/practitioner` | GET | provider, person | 🟡 |
| 2.12 | `rest/ProviderMenu` | GET | provider | 🟢 |
| 2.13 | `rest/SearchProviderMenu` | GET | provider | 🟡 |
| 2.14 | `Provider/Person/{id}` | GET | provider, person | 🟡 |
| 2.15 | `Provider/raw/{id}` | GET | provider | 🟢 |
| 2.16 | `rest/departments-for-site` | GET | organization/dept | 🟡 |

---

## Wave 3 — Identity: Patient (reads)

| # | Endpoint | M | Tables | Risk |
|---|----------|---|--------|------|
| 3.1 | `rest/patientByLabNumer` | GET | patient, sample | 🟡 |
| 3.2 | `rest/patient/merge/details/{patientId}` | GET | patient, person | 🟡 |
| 3.3 | `rest/patient-id-documents/{patientId}` | GET | patient_identity docs | 🟡 |
| 3.4 | `rest/patient-id-documents/{patientId}/{documentId}/full` | GET | docs (binary) | 🟡 |
| 3.5 | `rest/patient-photos/{id}/{isThumbnail}` | GET | patient photo (binary) | 🟡 |

---

## Wave 4 — Sample / order spine (reads)

The accession-centric core. Depends on waves 1–3.

| # | Endpoint | M | Tables | Risk |
|---|----------|---|--------|------|
| 4.1 | `rest/sample/all-by-accession/{accessionNumber}` | GET | sample, sample_item, analysis | 🔴 |
| 4.2 | `rest/sample/unassigned-by-accession/{accessionNumber}` | GET | sample, sample_item | 🟡 |
| 4.3 | `rest/order/search` | GET | sample | 🟡 |
| 4.4 | `rest/order/dashboard` | GET | sample, analysis (counts) | 🟡 |
| 4.5 | `rest/GenericSampleOrder` | GET | sample, sample_item | 🔴 |
| 4.6 | `rest/SamplePatientEntry` | GET | sample, patient (form load) | 🟡 |
| 4.7 | `rest/SampleEdit` | GET | sample, sample_item | 🟡 |
| 4.8 | `rest/SampleBatchEntrySetup` | GET | sample config | 🟡 |
| 4.9 | `rest/unassigned-sample/` | GET | sample_item | 🟡 |
| 4.10 | `rest/unassigned-sample/items` | GET | sample_item | 🟡 |
| 4.11 | `rest/unassigned-sample/items/search` | GET | sample_item | 🟡 |
| 4.12 | `rest/unassigned-sample/by-facility/{facilityId}` | GET | sample_item, organization | 🟡 |
| 4.13 | `rest/unassigned-sample/count-by-facility/{facilityId}` | GET | sample_item | 🟡 |
| 4.14 | `rest/order/{accessionNumber}/attachments` | GET | order attachments | 🟡 |
| 4.15 | `rest/order/attachments/{attachmentId}/download` | GET | attachment (binary) | 🟡 |
| 4.16 | `rest/order/attachments/{attachmentId}/view` | GET | attachment (binary) | 🟡 |
| 4.17 | `rest/getPendingAnalysisForTestProvider` | GET | analysis | 🟡 |

---

## Wave 5 — Result spine (reads) 🔴 clinical

Reference ranges, validation, logbook. Highest-fidelity wave — port with the
Java `resultvalidation`/`resultlimits` tests as the spec.

| # | Endpoint | M | Tables | Risk |
|---|----------|---|--------|------|
| 5.1 | `rest/LogbookResults` | GET | analysis, result, test | 🔴 |
| 5.2 | `rest/accession-results` | GET | result, analysis | 🔴 |
| 5.3 | `rest/AccessionValidation` | GET | result, analysis | 🔴 |
| 5.4 | `rest/ReferredOutTests` | GET | referral, analysis | 🟡 |
| 5.5 | `rest/WorkPlanByTest` | GET | analysis | 🟡 |
| 5.6 | `rest/WorkPlanByPanel` | GET | analysis, panel | 🟡 |
| 5.7 | `rest/WorkPlanByTestSection` | GET | analysis, test_section | 🟡 |
| 5.8 | `rest/WorkPlanByPriority` | GET | analysis | 🟡 |

---

## Wave 6 — Admin / config CRUD (read + write, per module)

These are self-contained admin screens sharing a **Get / NextPrevious / Cancel /
save-POST / Delete** pattern. Migrate each module as a **complete set** (all its
rows together) because they're isolated from the clinical spine and are the
lowest-risk place to prove the **write** path + CSRF + DB-write parity.

Recommended first write-module: **SiteInformation** (config only, no clinical
impact). Then Dictionary, Organization, Provider, then the test-catalog writes.

Per module the set is: `X` (GET+POST), `XMenu` (GET), `NextPreviousX` (GET),
`CancelX` (GET), `DeleteX` (GET/POST). Modules:

- `SiteInformation` · `PatientConfiguration` · `ResultConfiguration` ·
  `ValidationConfiguration` · `WorkplanConfiguration` · `SampleEntryConfig` ·
  `PrintedReportsConfiguration` · `NonConformityConfiguration` ·
  `MenuStatementConfig`
- `Dictionary` (POST + `DeleteDictionary`) · `Organization` (POST +
  `DeleteOrganization`) · Provider (`DeleteProvider`) · `UnifiedSystemUser`
  (+ `DeleteUnifiedSystemUser`) · `ExternalConnection` (+ `Deactivate…`)
- `AnalyzerTestName` (+ `CancelAnalyzerTestName`, `DeleteAnalyzerTestName`)

### Test-catalog writes (after 6-config proven)
`rest/test-catalog/tests` POST · `rest/test-catalog/tests/{testId}/activate` ·
`rest/test-catalog/panels` POST · `rest/test-catalog/group/ranges` PUT 🔴 ·
`rest/test-catalog/group/storage` PUT ·
`rest/test-catalog/sample-types/{sampleTypeId}/test-order` PUT ·
`TestAdd` · `TestModifyEntry` · `TestActivation` · `TestOrderability` ·
`TestRenameEntry` · `PanelCreate` · `PanelOrder` · `PanelRenameEntry` ·
`PanelTestAssign` · `SampleTypeCreate` · `SampleTypeOrder` ·
`SampleTypeRenameEntry` · `SampleTypeTestAssign` · `MethodCreate` ·
`MethodRenameEntry` · `UomCreate` · `UomRenameEntry` · `TestSectionCreate` ·
`TestSectionOrder` · `TestSectionRenameEntry` · `TestSectionTestAssign` ·
`ResultSelectListAdd` · `SaveResultSelectList` · `SelectListRenameEntry`

---

## Wave 7 — Order + Result **writes** 🔴 (the clinical lifecycle)

Only after waves 1–6 reads/writes are green. This is the heart of the LIS and
the hardest parity: create order → enter result → validate. Each needs
per-test DB reset + row-level assertions.

| Order | Endpoint | M | Note |
|-------|----------|---|------|
| 7.1 | `rest/SamplePatientEntry` | POST | create order (accession numbering 🔴) |
| 7.2 | `rest/SamplePatientEntryBatch` | POST | batch create |
| 7.3 | `rest/SampleBatchEntry` | POST | batch |
| 7.4 | `rest/GenericSampleOrder` | POST | create |
| 7.5 | `rest/GenericSampleOrder/{accessionNumber}` | PUT | update |
| 7.6 | `rest/GenericSampleOrder/validate` | POST | validate |
| 7.7 | `rest/GenericSampleOrder/import` | POST | import |
| 7.8 | `rest/SampleEdit` | POST | amend order |
| 7.9 | `rest/PatientManagement` | POST | patient create/update |
| 7.10 | `rest/patient/merge/validate` + `/execute` | POST | merge 🔴 |
| 7.11 | `rest/LogbookResults` | POST | **enter results** 🔴 |
| 7.12 | `rest/AccessionValidation` | POST | **validate results** 🔴 |
| 7.13 | `rest/reflexrule` (+ activate/deactivate) | POST | reflex firing 🔴 |
| 7.14 | `rest/test-calculation` (+ activate/deactivate) | POST | calc rules 🔴 |
| 7.15 | `rest/BatchTestReassignment` | POST | reassign |

---

## Wave 8 — Feature modules (self-contained; read then write, any order)

Each is an isolated module — schedule by team priority. Within each, reads
first. Rough size in parentheses.

- **Inventory** (~45 endpoints): `rest/inventory/**`,
  `rest/inventory-storage-locations/**` — large but fully self-contained CRUD.
- **Shipping / cold-storage** (~30): `rest/shipping-box/**`,
  `rest/box-sample/**`, `rest/coldstorage/**`.
- **EQA** (~20): `rest/eqa/**`.
- **NCE / non-conformity** (~12): `rest/nce/**`, `rest/nonconform*`,
  `rest/reportnonconformingevent`.
- **Notebook** (~13): `rest/notebook/**`.
- **Alerts** (~8): `rest/alerts/**`, `rest/alert-notification-config/`.
- **e-Signature** (~9): `rest/esig/**`.
- **Localization admin** (~10): `rest/localizations/**`.
- **Calendar** (~6): `rest/calendar/**`.
- **Sample-type requests** (~5): `rest/sample-type-requests/**`.
- **Users / roles** (~10): `rest/users`, `rest/UnifiedSystemUser*`,
  `rest/systemroles*`, `rest/user-*`, `rest/SearchUnifiedSystemUserMenu`.
- **Reports / audit** (~12): `rest/reports/tat/**`, `rest/AuditTrailReport*`,
  `rest/systemAuditEvents*`, `rest/ReportPrint`, `rest/PrintWorkplanReport`.
- **Site branding / logo** (~5): `rest/site-branding/**`, `rest/logoUpload/`.
- **Plugins / data-export** (~5): `rest/ListPlugins`, `rest/DataExportStatus*`.

---

## Wave 9 — Electronic orders + FHIR interop (interop tail)

Depends on order + result spine (waves 4–7). Per migration plan D5, prefer a
**facade to the existing HAPI FHIR server** over a Go reimplementation.

- `rest/ElectronicOrders` (GET) — e-order inbox.
- `Provider/FhirUuid` (POST).
- `rest/shipping-box/import-from-fhir` (POST), `rest/shipping-box/fhir-mapping-config`.
- The FHIR R4 surface itself (`/fhir/**`, HAPI, :8081) — **kept on Java / facade**,
  not ported endpoint-by-endpoint (see plan §D5).

---

## Wave 10 — Analyzer interop (may stay on Java permanently)

Per plan D2/D6: ASTM/HL7/file ingestion and the Java plugin ecosystem
(`GenericASTM`, `GenericFile`, `GenericHL7`) **cannot be loaded by Go**. Endpoints
touching analyzer import (`rest/AnalyzerTestName*`, analyzer result acceptance)
are the likely **permanent hybrid tail** — leave routed to Java unless/until the
transports are natively reimplemented. Not scheduled for endpoint-level cutover
in v1.

---

## Summary sequencing

| Wave | Theme | Endpoints (approx) | Gate to enter |
|------|-------|--------------------|---------------|
| 0 | Static/system reads | 14 | Go service + proxy + harness up |
| 1 | Dictionary + test catalog reads | 42 | W0 green |
| 2 | Org + provider reads | 16 | W1 green |
| 3 | Patient reads | 5 | W2 green |
| 4 | Sample/order reads | 17 | W1–3 green |
| 5 | Result reads 🔴 | 8 | W4 green |
| 6 | Admin/config CRUD (first writes) | ~90 | W1 green (parallel-safe) |
| 7 | Order+result writes 🔴 | ~18 | W4–6 green |
| 8 | Feature modules | ~190 | own context reads green |
| 9 | e-orders + FHIR facade | ~6 | W7 green |
| 10 | Analyzer interop | (hybrid) | — likely stays Java |

**Critical path to "core lab on Go":** W0 → W1 → W2 → W3 → W4 → W5 → W6(config
+ test-catalog writes) → W7. That path, not the 425-endpoint total, is the real
milestone. Feature modules (W8) fan out in parallel once their context's
reference reads (W1–2) are green.
