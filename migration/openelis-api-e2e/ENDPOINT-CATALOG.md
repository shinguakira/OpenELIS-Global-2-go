# OpenELIS REST endpoint catalog & manipulation patterns

Auto-generated from source by `tools/extract-endpoints.mjs`. **425** endpoints.

| pattern | count |
|---|---:|
| form read+write (GET/POST) | 128 |
| read-only (GET) | 244 |
| RESTful CRUD (GET/POST/PUT/DELETE) | 53 |

> Live note: the API does **not** expose verbs via `OPTIONS`/`Allow` — any non-GET
> without a CSRF token returns **403** (CSRF is enforced; token is in `GET /session`
> `.csrf`). So the verb map below is the authoritative source of truth; live
> per-verb probing requires CSRF + real payloads (the mutating suite, which resets
> the DB).

| endpoint | verbs | manipulation pattern |
|---|---|---|
| `Provider/FhirUuid` | POST | form read+write (GET/POST) |
| `Provider/Person/{id}` | GET | read-only (GET) |
| `Provider/raw/{id}` | GET | read-only (GET) |
| `provider/search` | GET | read-only (GET) |
| `rest/accession-results` | GET | read-only (GET) |
| `rest/AccessionValidation` | GET, POST | form read+write (GET/POST) |
| `rest/activate-reflexrule/{id}` | POST | form read+write (GET/POST) |
| `rest/activate-test-calculation/{id}` | POST | form read+write (GET/POST) |
| `rest/admin/menu/{elementId}` | GET | read-only (GET) |
| `rest/alert-notification-config/` | GET, POST | form read+write (GET/POST) |
| `rest/alerts/` | GET | read-only (GET) |
| `rest/alerts/{id}` | GET | read-only (GET) |
| `rest/alerts/{id}/acknowledge` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/alerts/{id}/resolve` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/alerts/count` | GET | read-only (GET) |
| `rest/alerts/dashboard` | GET | read-only (GET) |
| `rest/alerts/dashboard/summary` | GET | read-only (GET) |
| `rest/AllTestsForSampleTypeProvider` | GET | read-only (GET) |
| `rest/analysis-status-types` | GET | read-only (GET) |
| `rest/AnalyzerTestName` | GET, POST | form read+write (GET/POST) |
| `rest/AnalyzerTestNameMenu` | GET | read-only (GET) |
| `rest/api/tests/{id}/labelConfig` | GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/AuditTrailReport` | GET | read-only (GET) |
| `rest/AuditTrailReport/exportCsv` | GET | read-only (GET) |
| `rest/AuditTrailReport/exportPdf` | GET | read-only (GET) |
| `rest/BatchTestReassignment` | GET, POST | form read+write (GET/POST) |
| `rest/box-sample/` | POST | form read+write (GET/POST) |
| `rest/box-sample/{id}` | DELETE, GET | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/box-sample/{id}/reception-status` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/box-sample/{id}/remove` | POST | form read+write (GET/POST) |
| `rest/box-sample/by-box/{shippingBoxId}` | GET | read-only (GET) |
| `rest/box-sample/by-box/{shippingBoxId}/status/{status}` | GET | read-only (GET) |
| `rest/box-sample/by-sample/{sampleId}` | GET | read-only (GET) |
| `rest/box-sample/check-sample/{sampleId}` | GET | read-only (GET) |
| `rest/box-sample/count-by-box/{shippingBoxId}` | GET | read-only (GET) |
| `rest/box-sample/items` | POST | form read+write (GET/POST) |
| `rest/box-sample/items/{id}` | DELETE | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/box-sample/items/{id}/reception-status` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/box-sample/items/{id}/remove` | POST | form read+write (GET/POST) |
| `rest/box-sample/items/by-box/{shippingBoxId}` | GET | read-only (GET) |
| `rest/box-sample/items/check/{sampleItemId}` | GET | read-only (GET) |
| `rest/box-sample/items/count-by-box/{shippingBoxId}` | GET | read-only (GET) |
| `rest/calendar/holidays` | GET, POST | form read+write (GET/POST) |
| `rest/calendar/holidays/{id}` | DELETE, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/calendar/holidays/export` | GET | read-only (GET) |
| `rest/calendar/holidays/import` | POST | form read+write (GET/POST) |
| `rest/calendar/weekends` | GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/CancelAnalyzerTestName` | GET | read-only (GET) |
| `rest/CancelDictionary` | GET | read-only (GET) |
| `rest/CancelMenuStatementConfig` | GET | read-only (GET) |
| `rest/CancelNonConformityConfiguration` | GET | read-only (GET) |
| `rest/CancelOrganization` | GET | read-only (GET) |
| `rest/CancelPatientConfiguration` | GET | read-only (GET) |
| `rest/CancelPrintedReportsConfiguration` | GET | read-only (GET) |
| `rest/CancelResultConfiguration` | GET | read-only (GET) |
| `rest/CancelSampleEntryConfig` | GET | read-only (GET) |
| `rest/CancelSiteInformation` | GET | read-only (GET) |
| `rest/CancelValidationConfiguration` | GET | read-only (GET) |
| `rest/CancelWorkplanConfiguration` | GET | read-only (GET) |
| `rest/coldstorage/audit-trail/` | GET | read-only (GET) |
| `rest/coldstorage/reports/excursions` | GET | read-only (GET) |
| `rest/configuration-properties` | GET | read-only (GET) |
| `rest/DataExportStatus` | GET | read-only (GET) |
| `rest/DataExportStatus/{taskId}/attempts` | GET | read-only (GET) |
| `rest/DataExportStatus/{taskId}/trigger` | POST | form read+write (GET/POST) |
| `rest/deactivate-reflexrule/{id}` | POST | form read+write (GET/POST) |
| `rest/deactivate-test-calculation/{id}` | POST | form read+write (GET/POST) |
| `rest/DeactivateExternalConnection` | POST | form read+write (GET/POST) |
| `rest/DeleteAnalyzerTestName` | POST | form read+write (GET/POST) |
| `rest/DeleteDictionary` | POST | form read+write (GET/POST) |
| `rest/DeleteMenuStatementConfig` | GET | read-only (GET) |
| `rest/DeleteNonConformityConfiguration` | GET | read-only (GET) |
| `rest/DeleteOrganization` | POST | form read+write (GET/POST) |
| `rest/DeletePatientConfiguration` | GET | read-only (GET) |
| `rest/DeletePrintedReportsConfiguration` | GET | read-only (GET) |
| `rest/DeleteProvider` | POST | form read+write (GET/POST) |
| `rest/DeleteResultConfiguration` | GET | read-only (GET) |
| `rest/DeleteSiteInformation` | GET | read-only (GET) |
| `rest/DeleteUnifiedSystemUser` | POST | form read+write (GET/POST) |
| `rest/DeleteWorkplanConfiguration` | GET | read-only (GET) |
| `rest/departments-for-site` | GET | read-only (GET) |
| `rest/Dictionary` | GET, POST | form read+write (GET/POST) |
| `rest/dictionary-categories` | GET | read-only (GET) |
| `rest/dictionary/category/{categoryName}` | GET | read-only (GET) |
| `rest/DictionaryMenu` | GET | read-only (GET) |
| `rest/displayList/{listType}` | GET | read-only (GET) |
| `rest/ElectronicOrders` | GET | read-only (GET) |
| `rest/EntityNamesProvider` | GET | read-only (GET) |
| `rest/eqa/distributions` | GET, POST | form read+write (GET/POST) |
| `rest/eqa/distributions/{distributionId}/results` | GET, POST | form read+write (GET/POST) |
| `rest/eqa/distributions/{distributionId}/results/import` | POST | form read+write (GET/POST) |
| `rest/eqa/distributions/{distributionId}/statistics` | GET | read-only (GET) |
| `rest/eqa/distributions/{distributionId}/submit/{organizationId}` | POST | form read+write (GET/POST) |
| `rest/eqa/distributions/{distributionId}/submit/{organizationId}/approve-late` | POST | form read+write (GET/POST) |
| `rest/eqa/distributions/{id}` | GET | read-only (GET) |
| `rest/eqa/distributions/{id}/barcodes` | POST | form read+write (GET/POST) |
| `rest/eqa/distributions/{id}/status` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/eqa/eligible-organizations` | GET | read-only (GET) |
| `rest/eqa/my-programs/` | GET, POST | form read+write (GET/POST) |
| `rest/eqa/my-programs/{id}` | DELETE, GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/eqa/my-programs/providers` | GET | read-only (GET) |
| `rest/eqa/orders/` | GET | read-only (GET) |
| `rest/eqa/orders/summary` | GET | read-only (GET) |
| `rest/eqa/programs/` | GET, POST | form read+write (GET/POST) |
| `rest/eqa/programs/{id}` | GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/eqa/programs/{id}/tests` | GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/eqa/programs/{programId}/enrollments` | GET, POST | form read+write (GET/POST) |
| `rest/eqa/programs/{programId}/enrollments/{enrollmentId}` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/esig/admin/certifications` | GET | read-only (GET) |
| `rest/esig/admin/certifications/{username}` | DELETE | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/esig/certified/{username}` | GET | read-only (GET) |
| `rest/esig/certify` | POST | form read+write (GET/POST) |
| `rest/esig/enabled` | GET | read-only (GET) |
| `rest/esig/session-status/{username}` | GET | read-only (GET) |
| `rest/esig/sign` | POST | form read+write (GET/POST) |
| `rest/esig/signatures` | GET | read-only (GET) |
| `rest/ExternalConnection` | GET, POST | form read+write (GET/POST) |
| `rest/ExternalConnectionMenu` | GET | read-only (GET) |
| `rest/GenericSampleOrder` | GET, POST | form read+write (GET/POST) |
| `rest/GenericSampleOrder/{accessionNumber}` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/GenericSampleOrder/import` | POST | form read+write (GET/POST) |
| `rest/GenericSampleOrder/validate` | POST | form read+write (GET/POST) |
| `rest/getPendingAnalysisForTestProvider` | GET | read-only (GET) |
| `rest/inventory-storage-locations/` | GET, POST | form read+write (GET/POST) |
| `rest/inventory-storage-locations/{id}` | GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/inventory-storage-locations/{id}/deactivate` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/inventory-storage-locations/{id}/has-active-lots` | GET | read-only (GET) |
| `rest/inventory-storage-locations/{id}/path` | GET | read-only (GET) |
| `rest/inventory-storage-locations/{parentId}/children` | GET | read-only (GET) |
| `rest/inventory-storage-locations/code/{code}` | GET | read-only (GET) |
| `rest/inventory-storage-locations/top-level` | GET | read-only (GET) |
| `rest/inventory-storage-locations/type/{locationType}` | GET | read-only (GET) |
| `rest/inventory/items/` | GET, POST | form read+write (GET/POST) |
| `rest/inventory/items/{id}` | GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/inventory/items/{id}/activate` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/inventory/items/{id}/deactivate` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/inventory/items/{id}/stock` | GET | read-only (GET) |
| `rest/inventory/items/all` | GET | read-only (GET) |
| `rest/inventory/items/category/{category}` | GET | read-only (GET) |
| `rest/inventory/items/low-stock` | GET | read-only (GET) |
| `rest/inventory/items/search` | GET | read-only (GET) |
| `rest/inventory/items/type/{itemType}` | GET | read-only (GET) |
| `rest/inventory/items/types` | GET | read-only (GET) |
| `rest/inventory/lots/` | GET, POST | form read+write (GET/POST) |
| `rest/inventory/lots/{id}` | GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/inventory/lots/{id}/adjust` | POST | form read+write (GET/POST) |
| `rest/inventory/lots/{id}/dispose` | POST | form read+write (GET/POST) |
| `rest/inventory/lots/{id}/open` | POST | form read+write (GET/POST) |
| `rest/inventory/lots/{id}/qc-status` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/inventory/lots/{id}/status` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/inventory/lots/expired` | GET | read-only (GET) |
| `rest/inventory/lots/expiring` | GET | read-only (GET) |
| `rest/inventory/lots/item/{itemId}` | GET | read-only (GET) |
| `rest/inventory/lots/item/{itemId}/available` | GET | read-only (GET) |
| `rest/inventory/lots/item/{itemId}/total-quantity` | GET | read-only (GET) |
| `rest/inventory/lots/location/{locationId}` | GET | read-only (GET) |
| `rest/inventory/lots/lot-number/{lotNumber}` | GET | read-only (GET) |
| `rest/inventory/lots/process-expired` | POST | form read+write (GET/POST) |
| `rest/inventory/management/alerts` | GET | read-only (GET) |
| `rest/inventory/management/check-availability` | GET | read-only (GET) |
| `rest/inventory/management/consume` | POST | form read+write (GET/POST) |
| `rest/inventory/management/receive` | POST | form read+write (GET/POST) |
| `rest/inventory/transactions/{id}` | GET | read-only (GET) |
| `rest/inventory/transactions/date-range` | GET | read-only (GET) |
| `rest/inventory/transactions/lot/{lotId}` | GET | read-only (GET) |
| `rest/inventory/transactions/reference` | GET | read-only (GET) |
| `rest/inventory/transactions/type/{transactionType}` | GET | read-only (GET) |
| `rest/inventory/usage/{id}` | GET | read-only (GET) |
| `rest/inventory/usage/analysis/{analysisId}` | GET | read-only (GET) |
| `rest/inventory/usage/item/{itemId}` | GET | read-only (GET) |
| `rest/inventory/usage/lot/{lotId}` | GET | read-only (GET) |
| `rest/inventory/usage/test-result/{testResultId}` | GET | read-only (GET) |
| `rest/labUnit/config` | GET | read-only (GET) |
| `rest/ListPlugins` | GET | read-only (GET) |
| `rest/localizations/` | GET | read-only (GET) |
| `rest/localizations/{id}` | GET | read-only (GET) |
| `rest/localizations/{id}/translations` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/localizations/{id}/translations/{locale}` | POST | form read+write (GET/POST) |
| `rest/localizations/export/{locale}` | GET | read-only (GET) |
| `rest/localizations/import` | POST | form read+write (GET/POST) |
| `rest/localizations/missing/{locale}` | GET | read-only (GET) |
| `rest/localizations/stats` | GET | read-only (GET) |
| `rest/LogbookResults` | GET, POST | form read+write (GET/POST) |
| `rest/logoUpload/` | POST | form read+write (GET/POST) |
| `rest/math-functions` | GET | read-only (GET) |
| `rest/menu` | GET, POST | form read+write (GET/POST) |
| `rest/menu/{elementId}` | GET, POST | form read+write (GET/POST) |
| `rest/MenuStatementConfig` | GET, POST | form read+write (GET/POST) |
| `rest/MenuStatementConfigMenu` | GET | read-only (GET) |
| `rest/MethodCreate` | GET, POST | form read+write (GET/POST) |
| `rest/MethodRenameEntry` | GET, POST | form read+write (GET/POST) |
| `rest/methods-for-test/{testId}` | GET | read-only (GET) |
| `rest/nce/assign` | POST | form read+write (GET/POST) |
| `rest/nce/attachments/{attachmentId}/download` | GET | read-only (GET) |
| `rest/nce/categories` | GET | read-only (GET) |
| `rest/nce/dashboard` | GET | read-only (GET) |
| `rest/nce/generate-number` | GET | read-only (GET) |
| `rest/nce/history` | POST | form read+write (GET/POST) |
| `rest/nce/users` | GET | read-only (GET) |
| `rest/NCECorrectiveAction` | GET, POST | form read+write (GET/POST) |
| `rest/NextPreviousDictionary` | GET | read-only (GET) |
| `rest/NextPreviousMenuStatementConfig` | GET | read-only (GET) |
| `rest/NextPreviousNonConformityConfiguration` | GET | read-only (GET) |
| `rest/NextPreviousOrganization` | GET | read-only (GET) |
| `rest/NextPreviousPatientConfiguration` | GET | read-only (GET) |
| `rest/NextPreviousPrintedReportsConfiguration` | GET | read-only (GET) |
| `rest/NextPreviousResultConfiguration` | GET | read-only (GET) |
| `rest/NextPreviousSampleEntryConfig` | GET | read-only (GET) |
| `rest/NextPreviousSiteInformation` | GET | read-only (GET) |
| `rest/NextPreviousValidationConfiguration` | GET | read-only (GET) |
| `rest/NextPreviousWorkplanConfiguration` | GET | read-only (GET) |
| `rest/nonconformevents` | GET | read-only (GET) |
| `rest/nonconformingcorrectiveaction` | GET | read-only (GET) |
| `rest/NonConformityConfiguration` | GET, POST | form read+write (GET/POST) |
| `rest/NonConformityConfigurationMenu` | GET | read-only (GET) |
| `rest/notebook/auditTrail` | GET | read-only (GET) |
| `rest/notebook/create` | POST | form read+write (GET/POST) |
| `rest/notebook/dashboard/entries` | GET | read-only (GET) |
| `rest/notebook/dashboard/entries/{noteBookId}` | GET | read-only (GET) |
| `rest/notebook/dashboard/metrics` | GET | read-only (GET) |
| `rest/notebook/dashboard/notebooks` | GET | read-only (GET) |
| `rest/notebook/list` | GET | read-only (GET) |
| `rest/notebook/notebooksamples` | GET | read-only (GET) |
| `rest/notebook/questionnaires` | GET | read-only (GET) |
| `rest/notebook/samples` | GET | read-only (GET) |
| `rest/notebook/update/{noteBookId}` | POST | form read+write (GET/POST) |
| `rest/notebook/updatestatus/{noteBookId}` | POST | form read+write (GET/POST) |
| `rest/notebook/view/{noteBookId}` | GET | read-only (GET) |
| `rest/open-configuration-properties` | GET | read-only (GET) |
| `rest/order/{accessionNumber}/attachments` | GET, POST | form read+write (GET/POST) |
| `rest/order/attachments/{attachmentId}` | DELETE | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/order/attachments/{attachmentId}/download` | GET | read-only (GET) |
| `rest/order/attachments/{attachmentId}/view` | GET | read-only (GET) |
| `rest/order/dashboard` | GET | read-only (GET) |
| `rest/order/search` | GET | read-only (GET) |
| `rest/Organization` | GET, POST | form read+write (GET/POST) |
| `rest/organization-list` | GET | read-only (GET) |
| `rest/organization/{id}` | GET | read-only (GET) |
| `rest/organization/generate-site-code` | GET | read-only (GET) |
| `rest/organization/search` | GET | read-only (GET) |
| `rest/organization/types` | GET | read-only (GET) |
| `rest/OrganizationExport` | GET | read-only (GET) |
| `rest/OrganizationMenu` | GET | read-only (GET) |
| `rest/paginatedDisplayList/{listType}` | GET | read-only (GET) |
| `rest/PanelCreate` | GET, POST | form read+write (GET/POST) |
| `rest/PanelOrder` | GET, POST | form read+write (GET/POST) |
| `rest/PanelRenameEntry` | GET, POST | form read+write (GET/POST) |
| `rest/PanelTestAssign` | GET, POST | form read+write (GET/POST) |
| `rest/patient-id-documents/{documentId}` | DELETE, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/patient-id-documents/{patientId}` | GET | read-only (GET) |
| `rest/patient-id-documents/{patientId}/{documentId}/full` | GET | read-only (GET) |
| `rest/patient-photos/{id}/{isThumbnail}` | GET | read-only (GET) |
| `rest/patient/merge/details/{patientId}` | GET | read-only (GET) |
| `rest/patient/merge/execute` | POST | form read+write (GET/POST) |
| `rest/patient/merge/validate` | POST | form read+write (GET/POST) |
| `rest/patientByLabNumer` | GET | read-only (GET) |
| `rest/PatientConfiguration` | GET, POST | form read+write (GET/POST) |
| `rest/PatientConfigurationMenu` | GET | read-only (GET) |
| `rest/PatientManagement` | POST | form read+write (GET/POST) |
| `rest/practitioner` | GET | read-only (GET) |
| `rest/PrintedReportsConfiguration` | GET, POST | form read+write (GET/POST) |
| `rest/PrintedReportsConfigurationMenu` | GET | read-only (GET) |
| `rest/PrintWorkplanReport` | POST | form read+write (GET/POST) |
| `rest/projects` | GET | read-only (GET) |
| `rest/ProviderMenu` | GET | read-only (GET) |
| `rest/ReferredOutTests` | GET | read-only (GET) |
| `rest/reflexrule` | POST | form read+write (GET/POST) |
| `rest/reflexrule-options` | GET | read-only (GET) |
| `rest/reflexrules` | GET | read-only (GET) |
| `rest/reportnonconformingevent` | GET, POST | form read+write (GET/POST) |
| `rest/reportnonconformingevent/with-attachments` | POST | form read+write (GET/POST) |
| `rest/ReportPrint` | POST | form read+write (GET/POST) |
| `rest/reports/tat/detail` | GET | read-only (GET) |
| `rest/reports/tat/export` | GET | read-only (GET) |
| `rest/reports/tat/summary` | GET | read-only (GET) |
| `rest/reports/tat/trend` | GET | read-only (GET) |
| `rest/result-tree` | GET | read-only (GET) |
| `rest/ResultConfiguration` | GET, POST | form read+write (GET/POST) |
| `rest/ResultConfigurationMenu` | GET | read-only (GET) |
| `rest/ResultReportingConfiguration` | GET, POST | form read+write (GET/POST) |
| `rest/ResultSelectListAdd` | GET, POST | form read+write (GET/POST) |
| `rest/sample-item-status-types` | GET | read-only (GET) |
| `rest/sample-status-types` | GET | read-only (GET) |
| `rest/sample-type-requests/` | POST | form read+write (GET/POST) |
| `rest/sample-type-requests/{requestId}/cancel` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/sample-type-requests/{requestId}/fulfill` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/sample-type-requests/sample/{sampleId}` | GET | read-only (GET) |
| `rest/sample-type-requests/sample/{sampleId}/pending` | GET | read-only (GET) |
| `rest/sample-type-tests` | GET | read-only (GET) |
| `rest/sample/all-by-accession/{accessionNumber}` | GET | read-only (GET) |
| `rest/sample/unassigned-by-accession/{accessionNumber}` | GET | read-only (GET) |
| `rest/SampleBatchEntry` | POST | form read+write (GET/POST) |
| `rest/SampleBatchEntrySetup` | GET | read-only (GET) |
| `rest/SampleEdit` | GET, POST | form read+write (GET/POST) |
| `rest/SampleEntryConfig` | GET, POST | form read+write (GET/POST) |
| `rest/SampleEntryConfigMenu` | GET | read-only (GET) |
| `rest/SamplePatientEntry` | GET, POST | form read+write (GET/POST) |
| `rest/SamplePatientEntryBatch` | POST | form read+write (GET/POST) |
| `rest/SampleTypeCreate` | GET, POST | form read+write (GET/POST) |
| `rest/SampleTypeOrder` | GET, POST | form read+write (GET/POST) |
| `rest/SampleTypeRenameEntry` | GET, POST | form read+write (GET/POST) |
| `rest/SampleTypeTestAssign` | GET, POST | form read+write (GET/POST) |
| `rest/SaveResultSelectList` | POST | form read+write (GET/POST) |
| `rest/SearchExternalConnectionMenu` | GET | read-only (GET) |
| `rest/SearchOrganizationMenu` | GET | read-only (GET) |
| `rest/SearchProviderMenu` | GET | read-only (GET) |
| `rest/SearchUnifiedSystemUserMenu` | GET | read-only (GET) |
| `rest/SelectListRenameEntry` | GET, POST | form read+write (GET/POST) |
| `rest/server-time` | GET | read-only (GET) |
| `rest/shipping-box/` | GET, POST | form read+write (GET/POST) |
| `rest/shipping-box/{id}` | GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/shipping-box/{id}/archive` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/shipping-box/{id}/label/pdf` | GET | read-only (GET) |
| `rest/shipping-box/{id}/manifest-data` | GET | read-only (GET) |
| `rest/shipping-box/{id}/manifest/pdf` | GET | read-only (GET) |
| `rest/shipping-box/{id}/state` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/shipping-box/box-label-prefix` | GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/shipping-box/by-box-id/{boxId}` | GET | read-only (GET) |
| `rest/shipping-box/by-facility/{facilityId}` | GET | read-only (GET) |
| `rest/shipping-box/by-state/{state}` | GET | read-only (GET) |
| `rest/shipping-box/fhir-mapping-config` | GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/shipping-box/generate-box-number` | GET | read-only (GET) |
| `rest/shipping-box/import-from-fhir` | POST | form read+write (GET/POST) |
| `rest/shipping-box/site-organization-uuid` | GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/shipping-box/statistics` | GET | read-only (GET) |
| `rest/site-branding/` | GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/site-branding/logo/{type}` | DELETE, GET, POST | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/site-branding/reset` | POST | form read+write (GET/POST) |
| `rest/SiteInformation` | GET, POST | form read+write (GET/POST) |
| `rest/SiteInformationMenu` | GET | read-only (GET) |
| `rest/supportedlocales/` | GET, POST | form read+write (GET/POST) |
| `rest/supportedlocales/{id}` | DELETE, GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/supportedlocales/{id}/setFallback` | POST | form read+write (GET/POST) |
| `rest/supportedlocales/active` | GET | read-only (GET) |
| `rest/supportedlocales/fallback` | GET | read-only (GET) |
| `rest/systemAuditEvents` | GET | read-only (GET) |
| `rest/systemAuditEvents/entityTypes` | GET | read-only (GET) |
| `rest/systemAuditEvents/export` | GET | read-only (GET) |
| `rest/systemAuditEvents/exportPdf` | GET | read-only (GET) |
| `rest/systemroles` | GET | read-only (GET) |
| `rest/systemroles-testsections` | GET | read-only (GET) |
| `rest/test-calculation` | POST | form read+write (GET/POST) |
| `rest/test-calculations` | GET | read-only (GET) |
| `rest/test-catalog/{testId}/alerts/` | GET, POST | form read+write (GET/POST) |
| `rest/test-catalog/{testId}/alerts/{ruleId}` | DELETE, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/test-catalog/{testId}/alerts/roles` | GET | read-only (GET) |
| `rest/test-catalog/{testId}/reagents/` | GET, POST | form read+write (GET/POST) |
| `rest/test-catalog/{testId}/reagents/{reagentId}` | DELETE, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/test-catalog/{testId}/reflex-calc/` | GET | read-only (GET) |
| `rest/test-catalog/{testId}/storage/history/` | GET | read-only (GET) |
| `rest/test-catalog/dictionary` | GET | read-only (GET) |
| `rest/test-catalog/group/ranges` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/test-catalog/group/storage` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/test-catalog/group/summary` | GET | read-only (GET) |
| `rest/test-catalog/lab-units` | GET | read-only (GET) |
| `rest/test-catalog/panels` | GET, POST | form read+write (GET/POST) |
| `rest/test-catalog/panels/{panelId}/test-order` | GET | read-only (GET) |
| `rest/test-catalog/sample-types` | GET | read-only (GET) |
| `rest/test-catalog/sample-types/{sampleTypeId}/test-order` | GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/test-catalog/tests` | GET, POST | form read+write (GET/POST) |
| `rest/test-catalog/tests/{testId}` | GET | read-only (GET) |
| `rest/test-catalog/tests/{testId}/activate` | POST | form read+write (GET/POST) |
| `rest/test-catalog/tests/{testId}/analyzers` | GET | read-only (GET) |
| `rest/test-catalog/tests/{testId}/basic-info` | GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/test-catalog/tests/{testId}/localization` | GET | read-only (GET) |
| `rest/test-catalog/tests/{testId}/loinc-integrity` | GET | read-only (GET) |
| `rest/test-catalog/tests/{testId}/panels` | GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/test-catalog/tests/{testId}/ranges` | GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/test-catalog/tests/{testId}/sample-results` | GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/test-catalog/tests/{testId}/sample-results/copy-from/{sourceId}` | POST | form read+write (GET/POST) |
| `rest/test-catalog/tests/{testId}/siblings` | GET | read-only (GET) |
| `rest/test-catalog/tests/{testId}/storage` | GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/test-catalog/tests/{testId}/terminology` | GET, PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/test-display-beans` | GET | read-only (GET) |
| `rest/test-display-beans-map` | GET | read-only (GET) |
| `rest/test-list` | GET | read-only (GET) |
| `rest/test-result-tree` | GET | read-only (GET) |
| `rest/test-sample-types` | GET | read-only (GET) |
| `rest/test/{testId}/methods/` | GET, POST | form read+write (GET/POST) |
| `rest/test/{testId}/methods/{id}` | DELETE, PATCH | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/test/{testId}/methods/copyFrom/{sourceTestId}` | POST | form read+write (GET/POST) |
| `rest/test/{testId}/methods/inline-create` | POST | form read+write (GET/POST) |
| `rest/TestActivation` | GET, POST | form read+write (GET/POST) |
| `rest/TestAdd` | GET, POST | form read+write (GET/POST) |
| `rest/TestCatalog` | GET | read-only (GET) |
| `rest/TestModifyEntry` | GET, POST | form read+write (GET/POST) |
| `rest/TestNamesProvider` | GET | read-only (GET) |
| `rest/TestNotificationConfig` | GET, POST | form read+write (GET/POST) |
| `rest/TestNotificationConfig/raw/list` | GET | read-only (GET) |
| `rest/TestNotificationConfigMenu` | GET, POST | form read+write (GET/POST) |
| `rest/TestOrderability` | GET, POST | form read+write (GET/POST) |
| `rest/TestRenameEntry` | GET, POST | form read+write (GET/POST) |
| `rest/tests-by-sample` | GET | read-only (GET) |
| `rest/TestSectionCreate` | GET, POST | form read+write (GET/POST) |
| `rest/TestSectionOrder` | GET, POST | form read+write (GET/POST) |
| `rest/TestSectionRenameEntry` | GET, POST | form read+write (GET/POST) |
| `rest/TestSectionTestAssign` | GET, POST | form read+write (GET/POST) |
| `rest/trendsprojects` | GET | read-only (GET) |
| `rest/unassigned-sample/` | GET | read-only (GET) |
| `rest/unassigned-sample/{referralId}/assign-to-box` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/unassigned-sample/{referralId}/cancel` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/unassigned-sample/{referralId}/mark-lost` | PUT | RESTful CRUD (GET/POST/PUT/DELETE) |
| `rest/unassigned-sample/by-facility/{facilityId}` | GET | read-only (GET) |
| `rest/unassigned-sample/count-by-facility/{facilityId}` | GET | read-only (GET) |
| `rest/unassigned-sample/items` | GET | read-only (GET) |
| `rest/unassigned-sample/items/search` | GET | read-only (GET) |
| `rest/UnifiedSystemUser` | GET, POST | form read+write (GET/POST) |
| `rest/UnifiedSystemUserMenu` | GET | read-only (GET) |
| `rest/uom` | GET, POST | form read+write (GET/POST) |
| `rest/UomCreate` | GET, POST | form read+write (GET/POST) |
| `rest/UomRenameEntry` | GET, POST | form read+write (GET/POST) |
| `rest/user-programs` | GET | read-only (GET) |
| `rest/user-sample-types` | GET | read-only (GET) |
| `rest/user-test-sections/{roleName}` | GET | read-only (GET) |
| `rest/users` | GET | read-only (GET) |
| `rest/users/{roleName}` | GET | read-only (GET) |
| `rest/ValidationConfiguration` | GET, POST | form read+write (GET/POST) |
| `rest/ValidationConfigurationMenu` | GET | read-only (GET) |
| `rest/viewNonConformEvents` | GET, POST | form read+write (GET/POST) |
| `rest/WorkPlanByPanel` | GET | read-only (GET) |
| `rest/WorkPlanByPriority` | GET | read-only (GET) |
| `rest/WorkPlanByTest` | GET | read-only (GET) |
| `rest/WorkPlanByTestSection` | GET | read-only (GET) |
| `rest/WorkplanConfiguration` | GET, POST | form read+write (GET/POST) |
| `rest/WorkplanConfigurationMenu` | GET | read-only (GET) |
