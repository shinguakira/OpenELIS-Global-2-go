# OpenELIS Global 2 — User Manual

> A screen-by-screen guide to the OpenELIS Global 2 web application (React UI).
> Screenshots were captured from a running instance (version shown as **Test LIMS
> 3.2.1.x**) seeded with demo data. Your data and enabled modules may differ
> depending on configuration and user role.

---

## 1. Getting started

### Logging in

Open the application at **`https://localhost/`** (or your deployment URL). On a
development instance the default administrator credentials are
`admin` / `adminADMIN!`.

- **Legacy UI** is available at `https://localhost/api/OpenELIS-Global/`.
- **New React UI** (documented here) is the default at `https://localhost/`.

If the browser warns about the self-signed development certificate, choose
**Advanced → Proceed**.

### Navigation basics

Every screen shares a top bar (logo, global **search**, **notifications**,
**account**, **help**) and a **breadcrumb** under it. The **☰ menu** (top-left)
opens the main navigation; the **account** icon exposes language, password, and
logout. Most list screens share the same pattern: a **search/filter** panel on
top, a **results table** with `Items per page` paging, and action buttons
(Save / Print / Export).

### Home dashboard

The landing page is a tile dashboard of live workload counters — *In Progress
(awaiting result entry)*, *Ready for Validation*, *Orders Completed Today*,
*Average Turn-Around Time*, and more. Click a tile's expand icon to drill into
that queue.

![Home dashboard](img/user-manual/01-dashboard.png)

---

## 2. Order entry & patients

### Add Order (Test Request)

`Order → Add Order` — a guided wizard: **Patient Info → Program Selection → Add
Sample → Add Order**. Start by **searching for an existing patient** (by ID,
name, DOB, gender, or previous lab number) or registering a **New Patient**,
then add sample(s) and select the tests/panels to order.

![Add Order wizard](img/user-manual/02-add-order.png)

### Patient Management

Register and edit patient demographics (names, DOB, gender, identifiers,
contact, address). This is the master patient record used across all orders.

![Patient Management](img/user-manual/03-patient-management.png)

### Patient History

Look up a patient and review their full order/result history in one place.

![Patient History](img/user-manual/04-patient-history.png)

### Modify Order

Search an existing order (by accession/lab number) to amend samples, tests, or
patient details after initial entry.

![Modify Order](img/user-manual/05-modify-order.png)

### Batch Order Entry

Set up and enter **multiple orders in a batch** — useful for high-volume
reception where many samples share a program or test set.

![Batch Order Entry](img/user-manual/06-batch-order-entry.png)

---

## 3. Results & work plans

### Results by Test Unit (Logbook)

`Results` — select a **Test Unit** (lab section) to load its pending analyses
for **result entry**, then **Save**. This is the primary bench worklist.

![Logbook Results](img/user-manual/08-logbook-results.png)

### Analyzer Results

Review results imported automatically from connected **analyzers** (ASTM / HL7 /
file import), accept or reject them into the record.

![Analyzer Results](img/user-manual/09-analyzer-results.png)

### Results by Order (Accession)

Enter or review results for a **single order** by its accession/lab number.

![Accession Results](img/user-manual/10-accession-results.png)

### Results by Status

Filter the result queue by **status** (e.g. entered, referred, not started) to
work a specific stage of the pipeline.

![Status Results](img/user-manual/11-status-results.png)

### Results by Test / Range

Query results by **test and date range** for review or correction.

![Range Results](img/user-manual/12-range-results.png)

### Patient Results

View the consolidated **result report for a patient** across their orders.

![Patient Results](img/user-manual/07-patient-results.png)

---

## 4. Validation (result review)

### Result Validation

The review/approval step: a supervisor **validates** entered results (accept,
reject, or refer) before they are released/reported.

![Result Validation](img/user-manual/13-result-validation.png)

### Validation by Order (Accession)

Validate results for a specific order by accession number.

![Accession Validation](img/user-manual/14-accession-validation.png)

---

## 5. Referrals & electronic orders

### Referred-Out Tests

Track tests **referred to an external/reference laboratory** and record the
results returned.

![Referred Out Tests](img/user-manual/15-referred-out-tests.png)

### Electronic Orders

Inbox of **incoming electronic orders** (e.g. FHIR ServiceRequests from an EMR
such as OpenMRS). Accept an e-order to create the corresponding lab order.

![Electronic Orders](img/user-manual/16-electronic-orders.png)

---

## 6. Labels & barcodes

### Print Barcode

Generate and print **specimen/order barcode labels** for accessioned samples.

![Print Barcode](img/user-manual/17-print-barcode.png)

---

## 7. Sample storage & logistics

The **Storage** module models a cold-chain hierarchy: **Rooms → Devices
(freezers) → Shelves → Racks → Boxes**. Each list supports create/edit and
assigning sample items to positions.

### Storage — Rooms
![Storage Rooms](img/user-manual/18-storage-rooms.png)

### Storage — Devices (Freezers)
![Storage Devices](img/user-manual/19-storage-devices.png)

### Storage — Racks
![Storage Racks](img/user-manual/20-storage-racks.png)

### Storage — Boxes
![Storage Boxes](img/user-manual/21-storage-boxes.png)

### Freezer Monitoring

Record and monitor **temperature logs** for storage devices (cold-chain
compliance).

![Freezer Monitoring](img/user-manual/22-freezer-monitoring.png)

### Sample Shipment

Package samples into **boxes/shipments**, receive incoming shipments, and view
shipment reports.

![Sample Shipment](img/user-manual/23-sample-shipment.png)

---

## 8. Reports

### Routine Reports

Generate standard operational reports (patient reports, workload, export, etc.)
by selecting a report and its parameters.

![Routine Reports](img/user-manual/24-routine-reports.png)

### Report (Study / Non-conformity)

Additional report category (study/aggregate and non-conforming-event reporting).

![Report](img/user-manual/25-report.png)

### Audit Trail Report

Query the **audit trail** — who changed what and when — for compliance review.

![Audit Trail Report](img/user-manual/34-audit-trail-report.png)

---

## 9. Specialized modules

These modules are enabled per deployment. Each is a case-based dashboard.

### Cytology
![Cytology Dashboard](img/user-manual/26-cytology-dashboard.png)

### Pathology
![Pathology Dashboard](img/user-manual/27-pathology-dashboard.png)

### Immunohistochemistry
![Immunohistochemistry Dashboard](img/user-manual/28-immunohistochemistry-dashboard.png)

### Lab Notebook

A structured **notebook** for recording bench procedures and linking them to
samples.

![Notebook Dashboard](img/user-manual/29-notebook-dashboard.png)

### EQA (External Quality Assessment)

Manage **proficiency-testing** programs: distributions, participants, orders,
and results.

![EQA Management](img/user-manual/30-eqa-management.png)

### Non-Conforming Events (NCE)

Log and track **non-conforming events / corrective actions** (quality
management).

![NCE Dashboard](img/user-manual/31-nce-dashboard.png)

### Alerts

Central **alerts** feed for the logged-in user (e.g. flagged results, tasks).

![Alerts](img/user-manual/32-alerts.png)

---

## 10. Administration

### Admin dashboard (Master Lists / Configuration)

The administration hub. The left menu and tiles cover the full configuration
surface, including:

- **User Management**, **Organization Management**, **Provider Management**
- **Test Management / Test Catalog**, **Sample Type**, **Panel**, **Method**,
  **Reflex Tests**, **Analyzer Test Name**, **Test Activation / Orderability**
- **Lab Number Management**, **Program Entry**, **Label Presets**,
  **Barcode Configuration**
- **Dictionary Menu**, **General Configurations**, **Application Properties**,
  **Menu Configuration**, **Localization**
- **External Connections**, **Result Reporting Configuration**,
  **FHIR Data Export Status**, **Test Notification Configuration**
- **List Plugins**, **Search Index Management**, **Logging Configuration**,
  **Calendar Management**

![Admin dashboard](img/user-manual/33-master-lists.png)

---

## 11. API reference

OpenELIS exposes two programmatic surfaces. **Note:** there is currently **no
OpenAPI/Swagger document** generated for the internal REST API — the contract is
defined by the Spring MVC controllers in source.

### 11.1 Internal REST API (used by the React UI)

- **Base:** `https://<host>/api/OpenELIS-Global/rest/…`
- **Auth:** session-based (Spring Security form login / same session as the UI).
- **Scale:** ~**126 REST controllers**, ~**751 endpoint mappings**.
- **Discovery:** endpoints are defined by `@RequestMapping` / `@GetMapping` /
  `@PostMapping` on classes under `src/main/java/org/openelisglobal/**/controller/rest/`.
  Examples: `TestCatalog`, `TestAdd`, `SampleTypeManagement`, `PanelManagement`,
  `Organization`, `Provider/Person/{id}`, `GenericSampleOrder/{accessionNumber}`,
  `DataExportStatus/{taskId}/trigger`, `TestNotificationConfig`.

Since there is no generated spec, treat the `controller/rest` packages as the
authoritative endpoint list. (Adding `springdoc-openapi` would auto-generate a
Swagger UI — see *Recommendations* below.)

### 11.2 FHIR R4 API (external interoperability contract)

- **Server:** HAPI FHIR `RestfulServer` (`org.openelisglobal.fhir.servlets.FhirRestfulServer`).
- **Co-resident FHIR store** runs as its own service (container `external-fhir-api`,
  mapped to `:8081` in the dev compose).
- **Capability statement (the machine-readable spec):** `GET …/fhir/metadata`
  returns the server's `CapabilityStatement` listing every supported resource
  and interaction.
- **Resources used** include `Patient`, `ServiceRequest` (orders),
  `DiagnosticReport` / `Observation` (results), `Task`, `Specimen`,
  `Practitioner`, `Organization`.
- See also `docs/directFHIRCommunication.md` and `docs/messages.md` for message
  examples.

### 11.3 Credentials & data

- Dev login: `admin` / `adminADMIN!`.
- Load demo/test data with
  `./src/test/resources/load-test-fixtures.sh --profile=core` (or `harness`).
  See [Test Data Strategy](.specify/guides/test-data-strategy.md).

---

## 12. Recommendations (documentation gaps found)

- **No OpenAPI/Swagger** for the internal REST API. Adding `springdoc-openapi`
  would expose `/v3/api-docs` + a Swagger UI, turning the 126 controllers into a
  browsable, testable spec.
- The `docs/` site is **deployment/technical**-focused; this file is the first
  **end-user, screen-by-screen** manual. Consider adding it to `mkdocs.yml` nav
  (e.g. under a new *User Guide* section).
- Screenshots here are from a demo instance — regenerate them per release with
  `frontend/playwright/manual-shots.mjs` (the script used to produce this set).

---

*Generated from a live instance. Screenshot source script:
`frontend/playwright/manual-shots.mjs`. Images: `docs/img/user-manual/`.*
