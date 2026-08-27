# OpenELIS Global 2 → Go Migration Plan

Status: **draft / proposal** Source system: OpenELIS Global 2 (`develop`) — Java
21, Spring Framework 6.2 (traditional MVC, **not** Spring Boot), Jakarta EE 9
(`jakarta.*`), Hibernate/JPA, HAPI FHIR R4, Liquibase, **PostgreSQL**. Target: a
Go reimplementation of the OpenELIS backend. Companion docs in this folder:
`OpenMRS-Go-Migration-Plan` (see `openmrs-core-go/doc/GO_MIGRATION_PLAN.md`),
`OpenMRS-Analysis.md`, `e2e.md`.

> This is a _strategy_ document, not a line-by-line port script. OpenELIS is
> ~2,805 Java main files across ~120 self-contained domain packages, each using
> the same 5-layer pattern. A big-bang rewrite is not realistic. This plan
> defines the target architecture, the order of work, and how to prove
> behavioral parity at each step. It deliberately mirrors the OpenMRS Go plan so
> the two efforts in this workspace stay structurally consistent.

---

## 0. Decisions to make before writing any Go

| #   | Decision                                              | Recommendation & why                                                                                                                                                                                                                                                                                                                                                                 |
| --- | ----------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| D1  | **Keep the existing PostgreSQL schema, or redesign?** | **Keep it.** The schema is the real contract (Liquibase-owned, 993 changesets across 277 files, versioned `2.0.x`→`3.5.x`). Sharing it lets Go and Java run against the same DB during migration and makes parity testing possible. Redesign = a data-migration project _on top of_ a rewrite.                                                                                       |
| D2  | **Is Java plugin (`.jar`) compatibility a goal?**     | OpenELIS loads analyzer/interop plugins from the `openelisglobal-plugins` submodule (GenericASTM, GenericFile, GenericHL7) compiled against the Java API + `test-jar`. Go **cannot** load these. Decide early: "replace core, reimplement plugins natively in Go" vs "permanent hybrid where Java keeps serving plugin-driven interop." This is the single biggest scoping decision. |
| D3  | **What is the external contract?**                    | Two surfaces: (a) the **FHIR R4 API** (HAPI FHIR — the interoperability contract for lab orders/results) and (b) the **REST controllers** (`controller/rest`) that the React frontend consumes. The **JSP/legacy MVC UI is _not_ the contract** — the React frontend already talks REST. Port to preserve FHIR + REST behavior; drop the legacy MVC/form layer.                      |
| D4  | **Strangler-fig or clean cut?**                       | **Strangler-fig.** Run Go beside the Java WAR behind the existing nginx proxy; route endpoints to Go one bounded context at a time, both hitting the same Postgres. Ship and verify incrementally.                                                                                                                                                                                   |
| D5  | **FHIR: reimplement or facade?**                      | HAPI FHIR provides enormous machinery (R4 resource model, serialization, validation, the JPA server on port 8081). **Do not hand-reimplement FHIR from scratch.** Options: keep the HAPI FHIR server as a standalone service Go calls, or adopt a mature Go FHIR library and port only the OpenELIS-specific transformation logic (`fhir/transormation`, `dataexchange/fhir`).       |
| D6  | **Analyzer interop / scheduler**                      | ASTM (E1381/E1394), HL7v2, and file-import ingestion are hardware/instrument-facing and protocol-level. Plan to _re-implement the transports natively_ in Go (goroutines + net) or _keep them on Java_ during coexistence — but treat them as a distinct, high-risk subsystem, not "just another context."                                                                           |

**Assumed answers for the rest of this plan:** keep the schema (D1), replace
core over time via strangler-fig (D4), preserve FHIR R4 + REST contracts and
drop legacy MVC/JSP (D3), plugins reimplemented natively / out of scope for v1
(D2), FHIR via facade-or-library not hand-rolled (D5).

---

## 1. Architecture mapping (Java → Go)

OpenELIS uses a strict, repeated **5-layer pattern per domain**
(`Valueholder → DAO → Service → Controller → Form`). This regularity is a gift
for migration — one template, ~120 applications of it.

Go paths **mirror the Java paths during migration** (see layout decision below):
`internal/<domain>/<layer>/`. The "Go path" column is that mirror; the reorg to
idiomatic Go happens at the end.

| OpenELIS Java layer                                                            | Java location (under `org.openelisglobal.*`)                                   | Go path (mirrored)                                                                                                                                                                                                         | Go equivalent                                                                                                                                                                    |
| ------------------------------------------------------------------------------ | ------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Valueholder** (Hibernate entity/POJO, ~120 `valueholder/` dirs)              | `<domain>/valueholder/`                                                        | `internal/<domain>/valueholder/`                                                                                                                                                                                           | Plain structs. Flatten the `BaseObject`/audit base classes into embedded structs.                                                                                                |
| **DAO / DAOImpl** (~118 interfaces / ~87 impls)                                | `<domain>/dao/`, `<domain>/daoimpl/`                                           | `internal/<domain>/dao/` (+ `daoimpl/`)                                                                                                                                                                                    | Repository interfaces with **explicit SQL** via `sqlc`/`pgx`. **No ORM** — write the queries. Biggest translation risk (see §4).                                                 |
| **Service** (~133 `service/` dirs, `@Transactional`)                           | `<domain>/service/` (+ `impl/`)                                                | `internal/<domain>/service/`                                                                                                                                                                                               | Go interfaces + implementations. Transaction boundary = the service method.                                                                                                      |
| **Controller** (~78, many with `controller/rest`)                              | `<domain>/controller/`, `<domain>/controller/rest/`                            | `internal/<domain>/controller/rest/`                                                                                                                                                                                       | **stdlib `net/http`** handlers (`ServeMux` method routing) mirroring the **REST** subcontrollers; a `Routes(mux)` per domain. **Skip the non-REST MVC controllers** (legacy UI). |
| **Form** (~51 form-backing beans)                                              | `<domain>/form/`                                                               | `internal/<domain>/form/`                                                                                                                                                                                                  | Request/response DTOs on the REST handlers. The Spring `Form` beans are an MVC artifact — do not port wholesale.                                                                 |
| **Spring Security + role/rolemodule**                                          | `login/`, `role/`, `rolemodule/`, `security/`                                  | Auth middleware + a `context.Context`-carried principal; port the module/permission model as an authorization layer.                                                                                                       |
| **Hibernate interceptors / audit** (`audittrail/`, `history/`, `interceptor/`) | `audittrail/`, `interceptor/`                                                  | Centralized audit + `sysUserId`/timestamp stamping in the tx/repository layer. Miss this and every write is subtly wrong.                                                                                                  |
| **Liquibase changelogs** (277 files / 993 changesets)                          | `src/main/resources/liquibase/`                                                | **Keep Liquibase as-is** during coexistence so both apps agree on schema; convert to `goose` only after Java is retired — extraction + tooling already done, see [liquibase-to-goose-plan.md](liquibase-to-goose-plan.md). |
| **HAPI FHIR R4**                                                               | `fhir/`, `dataexchange/fhir`, `referral/fhir`, `shipment/fhir`, `storage/fhir` | Facade to a FHIR service or a Go FHIR library (D5). Port only OpenELIS-specific transform logic.                                                                                                                           |
| **Analyzer import** (ASTM/HL7/file)                                            | `analyzer/`, `analyzerimport/`, `analyzerresults/` + plugins submodule         | Native Go transports (goroutines/net) **or** keep on Java (D2/D6). Distinct subsystem.                                                                                                                                     |
| **i18n / React Intl keys**                                                     | `frontend` `en.json`, Transifex                                                | Unchanged — the Go backend serves data, the React frontend keeps owning i18n. New keys still land in `en.json` only.                                                                                                       |

### Go project layout — mirror Java during migration, reorganize at the end

**Decision:** during the migration the Go folders **mirror the Java source
layout** (`org.openelisglobal.<domain>.<layer>` → `internal/<domain>/<layer>/`),
even though that carries Java's redundant per-layer nesting. Rationale: the team
is Java-native and must be able to find "where did `SampleRestController` go?"
by the same path. Accept the non-idiomatic structure now; **reorganize into
idiomatic Go once the port is complete** (that reorg is itself a final migration
step, verified by the parity suite).

### Layer rules — what goes in each Go file (MANDATORY)

Before writing any Go code for a domain, check
`src/main/java/org/openelisglobal/<domain>/` and create a corresponding Go file
for each Java layer present.

| Java layer                             | Go file                                       | Contains                                                                       | Must NOT contain                                              |
| -------------------------------------- | --------------------------------------------- | ------------------------------------------------------------------------------ | ------------------------------------------------------------- |
| `valueholder/X.java`                   | `internal/<domain>/valueholder/x.go`          | Plain structs; no JSON tags                                                    | SQL, HTTP, business logic                                     |
| `daoimpl/XDAOImpl.java`                | `internal/<domain>/daoimpl/x_dao_impl.go`     | **All** database/ORM access — SQL, row scanning, projection→entity aggregation | Business logic, HTTP                                          |
| `service/XServiceImpl.java`            | `internal/<domain>/service/x_service_impl.go` | Calls DAO(s); business logic                                                   | SQL, HTTP, any DB/ORM import (`database/sql`, `gorm.io/gorm`) |
| `controller/rest/XRestController.java` | `internal/<domain>/controller/rest/x.go`      | Parse request → call service → convert to DTO → write JSON                     | **SQL, DB/ORM imports, business logic**                       |

**`daoimpl/` is the only layer allowed to import a database or ORM package.** A
`*gorm.DB` (or `*sql.DB`) field belongs in a DAO struct and nowhere else.
Services hold a `*daoimpl.XDAOImpl`, never a DB handle — that is what makes them
testable without a database.

**The controller/rest layer is a thin HTTP adapter — nothing more.** If you find
yourself writing a `DB.Raw()`, a `DB.Query()`, or a `for rows.Next()` loop in a
controller or service file, stop and move it to `daoimpl/`.

JSON tags live on DTO structs **in the controller package**, not on valueholder
structs. Valueholders are plain Go structs with no serialization concerns.

**Reference implementation:** `internal/localization/` (a2) — one file per Java
layer, correct separation throughout.

```
cmd/openelis/main.go                         # entrypoint; wires each domain's Routes()
internal/common/web/                          # shared HTTP plumbing (~ org.openelisglobal.common)
internal/system/controller/rest/system.go     # <- system/controller/rest/SystemRestController.java (a1)
internal/<domain>/controller/rest/            # <- <domain>/controller/rest/*RestController.java
internal/<domain>/service/                     # <- <domain>/service/
internal/<domain>/dao/  (or daoimpl/)          # <- <domain>/dao/ + daoimpl/ (pgx/sqlc, see §4)
internal/<domain>/valueholder/                 # <- <domain>/valueholder/ (entity structs)
internal/db/                                   # pgx connection, tx manager, sqlc
internal/auth/                                 # login, role/rolemodule, privilege checks
test/parity/                                   # golden tests comparing Go vs Java (REST + FHIR)
```

> The **end-of-migration reorg** collapses Java's layer-dirs
> (`controller/rest/`, `daoimpl/`) into idiomatic flat domain packages
> (`internal/system/`, a few files each). Deferred deliberately so Java devs
> navigate familiar paths throughout the port.

---

## 2. Recommended migration order (dependency-driven)

Port bounded contexts in dependency order — reference data and identity first,
then the lab result spine, then interop. The OpenELIS domain graph roughly
layers:

```
Dictionary / TestCatalog (test, panel, typeoftestresult) ─ referenced by almost everything
Organization / Provider / Address ─ referenced by orders & patients
Person ──▶ Patient ──▶ Sample (accession) ──▶ Analysis ──▶ Result
Login / Role / RoleModule (auth, referenced everywhere)
Analyzer import (ASTM/HL7/file) ──▶ AnalyzerResults ──▶ Result
FHIR / DataExchange / Referral (external contract, depends on Sample+Result)
```

**Suggested phase order:**

1. **Foundations** — `domain` structs, db/tx layer, audit/timestamp plumbing,
   site & system configuration (`config`, `configuration`, `siteinformation`),
   auth skeleton (`login`, `role`, `rolemodule`). _(No feature yet — the
   frame.)_
2. **Reference data** — `dictionary`/`dictionarycategory`, **test catalog**
   (`test`, `panel`, `typeoftestresult`, `testresult`, `unitofmeasure`),
   `organization`, `provider`, `address`/`citystatezip`.
3. **Identity** — `person` → `patient`, plus `samplehuman`.
4. **Sample/order core** — `sample`, `genericsample`, accession numbering, order
   entry.
5. **Result spine** — `analysis`, `result`, `resultvalidation`, `resultlimits`
   (the clinical-safety core: validation rules, reference ranges).
6. **Analyzer interop** — `analyzer`, `analyzerimport`, `analyzerresults`
   (ASTM/HL7/file). High-risk; may stay on Java per D2/D6.
7. **Interoperability / reporting** — `fhir`, `dataexchange`, `referral`,
   `report`/`reports`, `dataexport`.
8. **Cross-cutting features** — `barcode`, `coldstorage`/`shipment`/`storage`,
   `qc`/`qaevent`, `eqa`, `notification`.

---

## 3. The first vertical slice (prove the pattern end-to-end)

Before scaling out, build **one** context all the way through to lock the
template. Recommended: **Test-catalog read path** (self-contained, referenced
everywhere, low clinical risk) — or **Sample read path** if you want the
identity spine first.

1. `domain.Test`, `domain.TypeOfTestResult`, `domain.Panel`, `domain.TestResult`
   structs.
2. `test.Repository` with a real query: `GET test by id` (joins panel, result
   type, units, reference ranges).
3. `test.Service` with the read method + authorization check (role/module
   privilege).
4. `rest` handler mirroring the existing `controller/rest` response shape the
   React frontend consumes.
5. **Parity test**: same seeded Postgres, hit both Java and Go, diff the JSON.
   This becomes the acceptance bar for every subsequent context.

Then add create/update (write path exercises validators, transactions, audit
fields, accession/uniqueness) before moving on.

---

## 4. Hardest parts — where behavior silently diverges

- **Hibernate lazy loading & cascade.** Java freely walks
  `sample.getSampleItems().getAnalyses()...`; Hibernate fetches on access. In Go
  you decide up front what each query loads. Map the actual object graph each
  service method touches.
- **Save/update cascade semantics.** Saving a `Sample` cascades to sample
  items/analyses/results per the mappings. Replicate cascade order explicitly,
  in one transaction.
- **Audit fields & interceptors.** `sysUserId`, `lastupdated`, timestamps, and
  audit-trail rows are set by interceptors (`audittrail/`, `interceptor/`), not
  by callers. Centralize in the tx/repository layer.
- **Result validation & reference ranges** (`resultvalidation`, `resultlimits`)
  = **clinical safety rules**. Port faithfully with their tests as the spec; do
  not "clean up."
- **Spring Security + role/rolemodule privileges.** Enumerate every access check
  and reproduce as explicit middleware — a missing check is a security hole in a
  regulated LIS.
- **FHIR R4 fidelity.** HAPI does resource modelling, serialization, and
  validation. Hand-rolling this diverges silently — use a facade or a real Go
  FHIR library (D5).
- **Analyzer protocols.** ASTM E1381/E1394 framing and HL7v2 parsing are
  exacting and instrument-specific; the plugin architecture makes them
  pluggable. Treat as its own subsystem (D6).
- **Transaction boundaries.** `@Transactional` is invisible in the source. A
  service method calling several DAOs is _one_ transaction — get boundaries
  wrong and you get partial writes to patient/result records.
- **Accession numbering & sequences.** Order/accession generation has
  concurrency and format rules driven by site configuration — build the config
  plumbing early (Foundations).
- **Controller URL mapping — a class-level `@RequestMapping` prefix is easy to
  miss and easy to get wrong from memory.** Spring controllers commonly
  declare `@RequestMapping("/rest")` (or similar) once at the class level;
  every `@GetMapping`/`@PostMapping` in that file is relative to it — the
  method-level annotation alone does not tell you the real path. **Real,
  confirmed incident**: the b2 wave registered 3 Go routes
  (`Provider/raw/{id}`, `Provider/Person/{id}`, `provider/search`) without
  the `/rest` prefix `ProviderRestController` actually requires — copied from
  a wrong path listing in `openelis-api-e2e.md` itself (now fixed) — and
  every one of them 404'd on every real call until caught by live-testing
  against Java directly (source review and unit-level Go testing both missed
  it; only a live side-by-side request against the real Java server
  surfaced it). See `b2-org-provider-migration.md` §3.1. **Before registering
  any Go route for any future wave**: grep the target Java controller file
  for its class-level `@RequestMapping` yourself — do not trust a prior
  wave's doc, a remembered path, or a method-level annotation in isolation.
  Also affects: server timezone assumptions (see `time.Now()` vs
  `time.Now().UTC()` in the same doc, §3.1 #7) — anything derived from "how
  Java behaves" needs to be read from Java's actual source or observed live,
  never assumed.

---

## 5. Parity & testing strategy — reuse this workspace's e2e harness

`migration/openelis-api-e2e/` is the **language-neutral, black-box Playwright
(API-mode) parity oracle** for this migration (plan: `openelis-api-e2e.md`;
the OpenMRS analog was the original template, `e2e.md`, now superseded for
OpenELIS by this dedicated suite — it is **not** empty, it's the active,
growing gate: 479 tests as of the b2 wave). **What "parity-verified" requires
— non-negotiable, see `openelis-api-e2e.md`'s Principle section for the full
version with the incident that motivated it:**

- **No mocking, ever.** Every assertion runs against the real, live Java
  webapp and the real, live Postgres it's connected to. A Go endpoint checked
  only in isolation, or against assumed/remembered Java behavior, is a guess,
  not a verification — say so plainly if that's all that's been done for a
  given endpoint, don't call it "verified."
- **Assert real field-by-field response data, not just HTTP status.** `200`
  proves reachability, nothing about correctness.
- **Mine the target controller's own JUnit test first**
  (`*RestControllerTest.java`, `openelis-api-e2e.md` §16) before writing any
  Go — it's the fastest path to Java's real URL, request shape, and edge
  cases, and skipping it has already let a real bug (b2's wrong route
  prefix, `b2-org-provider-migration.md` §3.1) go undetected past
  implementation and into a "looks done" state.
- **Golden-master / parity harness.** Same live request against Java and Go;
  assert identical JSON (REST + FHIR) or identical DB state (writes). Primary
  gate — see `go-parity` project in `playwright.config.ts`; a spec isn't
  "parity-verified" until it's matched by that project's `testMatch` *and*
  has actually been run and passed there, not merely written.
- Target both the **FHIR R4 API** and the React-facing **REST controllers**
  (`/api/OpenELIS-Global`).
- **NO GO UNIT TESTS. Every test in this migration is an e2e test.** This
  supersedes an earlier version of this bullet, which said to "port high-value
  unit tests … to Go table tests." That was wrong and it has already misled
  work: `migration/openelis-go` contains **zero** `*_test.go` files, and that is
  the intended state, not a gap to fill.

  **Why:** a Go unit test asserts what the person writing it *believes* Java
  does. It has no oracle. The entire value of this migration's test suite is
  that **Java and Go are checked against each other**, on the same live request,
  in the same run — which only an e2e test in
  `migration/openelis-api-e2e/` can do. A green Go unit test proves the port
  matches its author's assumption; it proves nothing about parity, and it makes
  a wrong assumption look verified. That is worse than no test.

  Consequence to accept, not work around: **a behavior with no Java-observable
  counterpart gets no test.** Internal-only concerns (memory reclamation,
  startup configuration refusal) are fixed and *documented as untested*, never
  given a Go unit test to make the change feel covered.
- **Keep the existing Playwright suite** (42 specs in `frontend/playwright`)
  running against the strangler proxy as a full-stack regression net during
  coexistence.
- Wire CI (`e2e-playwright.yml`) to run the parity projects per context as it
  lands — `api-readonly` (Java) and `go-parity` (Go) over the same specs.

---

## 6. Coexistence / rollout (strangler-fig)

1. Deploy Go beside the Java WAR; both connect to the **same PostgreSQL**
   (`openelisglobal-database`, `15432:5432`).
2. Use the existing **nginx proxy** (`openelisglobal-proxy`, 80/443) as the
   router. Default: everything → Java (`oe.openelis.org:8080/8443`).
3. As each context passes parity, flip its REST/FHIR routes → Go (per-endpoint /
   feature-flag).
4. Liquibase stays the single owner of schema during this window (both apps read
   the same tables).
5. Monitor error rates/latency for divergence.
6. Retire Java routes only after Go serves them and is stable. **Analyzer
   plugins and FHIR** are the likely long-lived hybrid tail (D2/D5/D6).

---

## 7. Rough sequencing summary

| Phase             | Contexts                                                             | Exit criterion                                             |
| ----------------- | -------------------------------------------------------------------- | ---------------------------------------------------------- |
| P0 Foundations    | domain, db, tx, audit, site/system config, auth skeleton             | Test-catalog (or Sample) read slice (§3) passes parity     |
| P1 Reference data | dictionary, test catalog, organization, provider, address            | Parity on reads + writes                                   |
| P2 Identity       | person, patient, samplehuman                                         | Parity incl. patient identifier rules                      |
| P3 Sample/order   | sample, genericsample, accession numbering                           | Parity incl. accession format/uniqueness                   |
| P4 Result spine   | analysis, result, resultvalidation, resultlimits                     | Validation + reference-range rules match Java specs        |
| P5 Interop-in     | analyzer, analyzerimport, analyzerresults                            | ASTM/HL7/file ingestion parity (or documented Java-hybrid) |
| P6 Interop-out    | fhir, dataexchange, referral, report, dataexport                     | FHIR R4 + report parity                                    |
| P7 Cross-cutting  | barcode, storage/coldstorage/shipment, qc/qaevent, eqa, notification | Parity                                                     |
| P8 Cutover        | routing, monitoring, retire Java routes                              | All ported REST/FHIR routes on Go, stable                  |

---

## 8. Open questions to resolve with stakeholders

- **D1–D6 above** (schema, plugins, contract, rollout, FHIR strategy,
  analyzer/scheduler).
- **Plugins (D2) is decisive**: OpenELIS's analyzer/interop ecosystem lives in
  `openelisglobal-plugins` (Java). If plugins must keep working, the Java core
  cannot be fully retired — the realistic target becomes "Go serves core lab
  REST/FHIR; Java keeps plugin-driven analyzer interop" (permanent hybrid).
  Resolve before committing a timeline.
- **FHIR (D5)**: keep HAPI FHIR JPA server as-is (it already runs standalone
  on 8081) and have Go call it, vs. a Go-native FHIR layer? The former
  dramatically de-risks v1.
- **Legacy MVC/JSP UI**: confirm it is fully superseded by the React frontend
  (it appears to be) so the Form/non-REST controller layers can be dropped
  rather than ported.
- **Postgres stays** — no reason to change engine; keeps coexistence simple
  (unlike OpenMRS which is MySQL).
- **Team size / timeline** — scopes how many contexts run in parallel once the
  P0 template is locked.

---

## Appendix — measured baseline (source, `develop`)

| Metric                                   |                                                        Value |
| ---------------------------------------- | -----------------------------------------------------------: |
| Backend Java main files                  |                                                       ~2,805 |
| Domain packages (`org.openelisglobal.*`) |                                                         ~120 |
| Valueholder / DAO / DAOImpl dirs         |                                            ~120 / ~118 / ~87 |
| Service / Controller / Form dirs         |                                             ~133 / ~78 / ~51 |
| Liquibase changelog files / changesets   |                                   277 files / 993 changesets |
| `createTable` operations                 |                                        ~318 across ~70 files |
| DB engine                                |                                                   PostgreSQL |
| Frontend                                 |           React 17 + Carbon (Vite/Vitest), talks REST + FHIR |
| E2E                                      |       Playwright (42 specs) + Cypress (34 specs, deprecated) |
| Git submodules                           |               `dataexport`, `plugins` (+ FHIR server, tools) |
| External contracts                       | FHIR R4 (HAPI, :8081) + REST controllers (`controller/rest`) |
