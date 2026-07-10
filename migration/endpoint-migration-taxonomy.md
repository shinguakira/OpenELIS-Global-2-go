# OpenELIS → Go — Endpoint **Type** Taxonomy (migration lens)

Status: **draft / proposal**
Companion to [endpoint-migration-order.md](endpoint-migration-order.md) (the
*when*) and [OpenELIS-Go-Migration-Plan.md](OpenELIS-Go-Migration-Plan.md) (the
*why*). This doc is the *what-kind-of-thing* — it groups the 425 endpoints by
**migration character**, not by feature.

The order doc sequences endpoints. This doc explains *why a given endpoint is
easy or hard to port*, what it depends on, and what its parity test must assert.
When you pick up any endpoint, first find its **Type** here — that tells you the
porting recipe and the parity strategy before you read a line of Java.

---

## The type axes

Every endpoint is classified on three axes:

1. **I/O direction** — Read (safe, idempotent) vs Write (CSRF, DB mutation, tx).
2. **Data coupling** — how much it depends on other data/contexts being ported.
3. **Fidelity risk** — how silently Go can diverge from Java (the real cost).

The Types below are the useful *combinations* of those axes.

---

## Type A — Static / computed (no DB, or read-only config)

**Examples:** `rest/server-time`, `rest/math-functions`,
`rest/open-configuration-properties`, `rest/configuration-properties`.

| Property | Value |
|----------|-------|
| Direction | Read |
| Depends on | Nothing (or a single config table) |
| Fidelity risk | 🟢 trivial |
| Port recipe | Handler → return constant / one-row read. No service logic. |
| Parity test | Exact JSON equality. |
| Migration role | **Wave 0 smoke.** These exist to prove the pipe, not to move real load. |

---

## Type B — Reference lookups (single table, read-only)

**Examples:** `rest/dictionary-categories`, `rest/uom`,
`rest/organization/types`, `rest/sample-status-types`, `rest/supportedlocales/`,
`rest/test-catalog/lab-units`, `rest/TestNamesProvider`.

| Property | Value |
|----------|-------|
| Direction | Read |
| Depends on | One table; **everything downstream depends on THEM** |
| Fidelity risk | 🟢–🟡 (sort order, active-flag filtering, i18n label resolution) |
| Port recipe | `SELECT … WHERE active … ORDER BY …` → DTO. Watch the implicit `ORDER BY` and the "inactive rows hidden" filter — Java hides deactivated rows by default. |
| Parity test | Set-equality + order check. |
| Migration role | **Highest leverage.** Port these first (Wave 1–2); they unblock the widest fan-out. A bug here propagates into every join below. |

**Dependency note:** Type B has *no upstream* deps but *maximum downstream*
fan-out. This is why the order doc front-loads them.

---

## Type C — Join / projection reads (multi-table, derived shape)

**Examples:** `rest/test-catalog/tests/{testId}`, `rest/order/search`,
`rest/sample/all-by-accession/{accessionNumber}`, `rest/LogbookResults` (GET),
`rest/menu` (tree build), `rest/test-result-tree`.

| Property | Value |
|----------|-------|
| Direction | Read |
| Depends on | Multiple Type-B contexts being ported first |
| Fidelity risk | 🟡–🔴 — **this is where Hibernate lazy-loading bites** |
| Port recipe | Decide the object graph **up front**: one query with joins, or N queries assembled in the service. Java walks `sample.getSampleItems().getAnalyses()…` lazily; you must enumerate exactly what the response includes. |
| Parity test | Deep JSON diff on a seeded accession. Diff *ordering of nested arrays* too. |
| Migration role | The bulk of the read work (waves 3–5). Port only after the Type-B data they join to is green, or the diff is un-debuggable. |

**Sub-type C-tree:** `menu`, `*-tree`, `displayList/{listType}` build recursive
/ polymorphic structures. The `{listType}` dispatch is a big switch in Java —
port the switch faithfully; each list type is effectively its own mini-endpoint.

---

## Type D — Form-load reads (GET half of a GET/POST pair)

**Examples:** `rest/SamplePatientEntry` (GET), `rest/SampleEdit` (GET),
`rest/Dictionary` (GET), all the `*Configuration` (GET), `rest/AccessionValidation` (GET).

| Property | Value |
|----------|-------|
| Direction | Read (but paired with a Type-F write on the same path) |
| Depends on | Its own context's Type-B/C reads |
| Fidelity risk | 🟡 — returns *both* data and form-scaffolding (dropdown option lists, defaults) |
| Port recipe | Same as Type C, but the DTO also carries the "what can this form contain" lists. Reuse Type-B endpoints internally rather than re-querying. |
| Parity test | JSON diff; verify the option-lists match. |
| Migration role | Migrate the **GET in the read wave, the POST later** in the write wave. They share a URL but are two separate migration units. |

---

## Type E — Admin CRUD sets (isolated config modules)

**Examples (each a *set*):** `SiteInformation`, `PatientConfiguration`,
`ValidationConfiguration`, `WorkplanConfiguration`, `MenuStatementConfig`,
`NonConformityConfiguration`, plus the `X` / `XMenu` / `NextPreviousX` /
`CancelX` / `DeleteX` satellites around each.

| Property | Value |
|----------|-------|
| Direction | Read + Write |
| Depends on | Almost nothing else — **self-contained** |
| Fidelity risk | 🟢–🟡 (config, no clinical impact) |
| Port recipe | Migrate the whole 5-endpoint set together. The `NextPrevious`/`Cancel` are navigation helpers, not real writes. |
| Parity test | **First place to prove the WRITE path**: POST → assert DB row → GET reads it back → identical. CSRF token flow exercised here. |
| Migration role | **The write-path proving ground** (Wave 6). Low blast radius, so break the CSRF/tx/audit machinery here, not on patient records. Start with `SiteInformation`. |

**Why grouped:** these share one code template in Java and one migration recipe.
Do them as batches, not individually.

---

## Type F — Clinical lifecycle writes 🔴 (the crown jewels)

**Examples:** `rest/SamplePatientEntry` (POST), `rest/LogbookResults` (POST),
`rest/AccessionValidation` (POST), `rest/GenericSampleOrder` (POST/PUT),
`rest/reflexrule`, `rest/test-calculation`, `rest/patient/merge/execute`.

| Property | Value |
|----------|-------|
| Direction | Write |
| Depends on | The **entire** read spine (waves 1–5) + Type-E write machinery proven |
| Fidelity risk | 🔴 **highest** — accession numbering, cascade saves, audit rows, reference-range validation, reflex firing, transaction boundaries |
| Port recipe | Replicate the `@Transactional` boundary as ONE Go tx; replicate cascade order explicitly (sample→item→analysis→result); stamp audit fields in the repo layer; port the validation/reference-range rules **with the Java unit tests as the spec**. |
| Parity test | Golden DB-state diff: seed → POST to Java on DB-A, POST to Go on DB-B (identical seed) → dump and diff *all* touched tables incl. `audit_trail`. Not just the HTTP response. |
| Migration role | **Last** (Wave 7). Never migrate one of these until its context's reads and the Type-E write machinery are green. |

**These are why the whole migration is strangler-fig and not big-bang.** A silent
divergence here corrupts patient/result data.

---

## Type G — Binary / file endpoints

**Examples:** `rest/patient-photos/{id}/{isThumbnail}`,
`rest/order/attachments/{attachmentId}/download`, `rest/*/label/pdf`,
`rest/*/manifest/pdf`, `rest/AuditTrailReport/exportPdf`, `*/exportCsv`,
`rest/logoUpload/`.

| Property | Value |
|----------|-------|
| Direction | Read (download) or Write (upload) |
| Depends on | Its context |
| Fidelity risk | 🟡 — content-type, byte-for-byte (images) vs semantic (PDF/CSV) equality |
| Port recipe | Stream bytes; match `Content-Type`/`Content-Disposition`. PDF generation (JasperReports in Java) is **hard to byte-match** — assert on generated *data*, not the rendered PDF bytes. |
| Parity test | Images/CSV: byte or row equality. PDF: assert the source data, mark rendered-bytes as "documented divergence." |
| Migration role | Migrate late, per-context. PDF renderers are a candidate to **keep on Java** even after the data endpoint moves. |

---

## Type H — Large self-contained feature modules

**Examples:** `rest/inventory/**` (~45), `rest/shipping-box/**` +
`rest/box-sample/**` (~30), `rest/eqa/**` (~20), `rest/notebook/**` (~13),
`rest/nce/**` (~12), `rest/esig/**` (~9), `rest/localizations/**` (~10),
`rest/calendar/**` (~6).

| Property | Value |
|----------|-------|
| Direction | Read + Write (full CRUD sub-apps) |
| Depends on | Type-B reference data only; **not** the clinical spine |
| Fidelity risk | 🟡 (own tables, own logic, little cross-talk) |
| Port recipe | Treat each module as a **mini-migration**: its own reads-then-writes internal order, its own parity subset. |
| Parity test | Module-scoped seed + CRUD round-trip. |
| Migration role | **Parallelizable** (Wave 8). Once Type-B is green, different people/agents can own different modules independently. This is where the migration *scales out*. |

---

## Type I — Interop / external contract (FHIR, e-orders)

**Examples:** `rest/ElectronicOrders`, `Provider/FhirUuid`,
`rest/shipping-box/import-from-fhir`, and the whole `/fhir/**` HAPI surface.

| Property | Value |
|----------|-------|
| Direction | Read + Write |
| Depends on | Order + result spine (waves 4–7) |
| Fidelity risk | 🔴 FHIR R4 resource fidelity |
| Port recipe | **Facade, don't reimplement** (plan D5): keep HAPI FHIR (:8081) as a service Go calls; port only the OpenELIS-specific transform. |
| Parity test | FHIR resource equality against the capability statement. |
| Migration role | Interop tail (Wave 9). Largely stays a hybrid. |

---

## Type J — Analyzer / plugin-bound (likely permanent Java) 🔴

**Examples:** `rest/AnalyzerTestName*`, analyzer-result acceptance paths, and
anything reaching the `GenericASTM`/`GenericFile`/`GenericHL7` Java plugins.

| Property | Value |
|----------|-------|
| Direction | Read + Write + long-lived transports |
| Depends on | Java plugin `.jar`s that **Go cannot load** |
| Fidelity risk | 🔴 protocol-exact (ASTM E1381/E1394, HL7v2) |
| Port recipe | Native Go transports **or** keep on Java (plan D2/D6). |
| Parity test | n/a for v1 — documented Java-hybrid. |
| Migration role | **Do not schedule for v1 cutover.** The permanent hybrid tail. |

---

## Dependency graph between types (what must exist before what)

```
Type A (static)            ── no deps ── migrate anytime (used as smoke)
        │
Type B (reference reads)   ── no upstream deps, MAX downstream fan-out ── FIRST
        │
        ├─▶ Type C (join reads)        needs B
        ├─▶ Type D (form-load reads)   needs B (+ own C)
        │
Type E (admin CRUD)        ── near-independent ── first WRITE proving ground
        │        (proves CSRF / tx / audit machinery)
        ▼
Type F (clinical writes) 🔴 needs C + D reads green AND E write-machinery green
        │
        ├─▶ Type I (FHIR/interop)   needs F
        └─▶ Type G (reports/PDF)    needs F data

Type H (feature modules)   ── needs only B ── PARALLEL fan-out, independent of F
Type J (analyzer/plugin) 🔴 ── out of band ── stays Java
```

**One-line reading:** port **B** first (unblocks everything), prove writes on
**E** (safe), then do **F** (dangerous, gated on B/C/D/E), and fan **H** out in
parallel the whole time. **I/J** are the hybrid tail.

---

## Type → wave cross-reference

| Type | Character | Waves (order doc) | Count (approx) |
|------|-----------|-------------------|----------------|
| A | Static/computed | 0 | ~6 |
| B | Reference reads | 1, 2 | ~55 |
| C | Join reads | 3, 4, 5 | ~40 |
| D | Form-load reads | 4, 6 (GET halves) | ~20 |
| E | Admin CRUD sets | 6 | ~90 |
| F | Clinical writes 🔴 | 7 | ~18 |
| G | Binary/file | 4, 8 (per context) | ~20 |
| H | Feature modules | 8 | ~190 |
| I | FHIR/interop | 9 | ~6 |
| J | Analyzer/plugin 🔴 | 10 (hybrid) | ~10 |
