# e2 — Test-catalog writes (scoped migration plan)

Status: **in progress — the branch owns all 38 writes; 23 are ported and in the gate, TestModifyEntry is written but not yet committed (see §7)**
Branch: `migration/e2-testcatalog-writes` (from `migration/e1-config-crud`, not
from `migration-base` — see §1).
Companion docs:
[endpoint-migration-order.md](endpoint-migration-order.md) (Wave 6),
[endpoint-migration-taxonomy.md](endpoint-migration-taxonomy.md) (Type E),
[branch-naming.md](branch-naming.md),
[open-items.md](open-items.md).

e1 proved the write path on configuration rows: one controller pair, nine
domains, one table, one audit shape. e2 is the same *kind* of work at roughly
five times the surface, against the catalog the clinical spine reads from.

The difference that matters is not the count. e1's writes touched
`site_information`, which nothing else joins to. e2's writes change **test**,
**panel**, **sample_type**, **test_section**, **unit_of_measure**, their join
tables and their `localization` rows — the reference data every ported read
wave already answers from. A wrong write here is visible in b1, c2 and c3
before anyone looks at the catalog screen.

---

## 0. Scope

Two Java packages, both already `@RestController`, both under `/rest`.

### 0.1 `testcatalog` — the editor (`/rest/test-catalog/...`)

`TestCatalogEditorRestController`, 1550 lines, 31 mappings. **12 are writes:**

| Method | Path | Note |
|---|---|---|
| POST | `/tests` | create |
| PUT | `/tests/{testId}/basic-info` | |
| PUT | `/tests/{testId}/sample-results` | |
| POST | `/tests/{testId}/sample-results/copy-from/{sourceId}` | copies another test's shape |
| PUT | `/tests/{testId}/ranges` | 🔴 clinical — reference ranges |
| PUT | `/group/ranges` | 🔴 clinical — bulk, across tests |
| PUT | `/tests/{testId}/storage` | |
| PUT | `/group/storage` | bulk |
| PUT | `/sample-types/{sampleTypeId}/test-order` | ordering |
| PUT | `/tests/{testId}/terminology` | LOINC etc. |
| PUT | `/tests/{testId}/panels` | |
| POST | `/panels` | create |

`TestCatalogActivationRestController` adds one more:

| Method | Path |
|---|---|
| POST | `/rest/test-catalog/tests/{testId}/activate` |

Of this controller's 18 **reads**, only three are ported (b1 did `lab-units`,
`sample-types`, `panels`). The rest are unported, and several are the natural
oracle for the write beside them — `GET /tests/{testId}/ranges` is how you see
what `PUT /tests/{testId}/ranges` did.

### 0.2 `testconfiguration` — the legacy screens, now REST (`/rest/{Name}`)

24 controllers, each a **GET form / POST save** pair. 23 carry a write; 25
write mappings in total (`ResultSelectListAdd` has two).
`TestCatalogRestController` is read-only and was ported in b1.

| Group | Controllers |
|---|---|
| Test | `TestAdd` · `TestModifyEntry` · `TestActivation` · `TestOrderability` · `TestRenameEntry` |
| Panel | `PanelCreate` · `PanelOrder` · `PanelRenameEntry` · `PanelTestAssign` |
| Sample type | `SampleTypeCreate` · `SampleTypeOrder` · `SampleTypeRenameEntry` · `SampleTypeTestAssign` |
| Test section | `TestSectionCreate` · `TestSectionOrder` · `TestSectionRenameEntry` · `TestSectionTestAssign` |
| Method | `MethodCreate` · `MethodRenameEntry` |
| UOM | `UomCreate` · `UomRenameEntry` |
| Select lists | `ResultSelectListAdd` (2 writes) · `SelectListRenameEntry` |

**Total: 38 write endpoints, plus ~30 companion reads that are not ported.**

That is bigger than every previous wave, and it is all one branch. §5 is the
ORDER the work is done in, not a smaller scope: the branch is finished when it
covers every endpoint above.

---

## 1. Fork point, and why it is not `migration-base`

[branch-naming.md](branch-naming.md) says migration branches fork from
`migration-base`. `migration-base` is at the c2 merge (PR #13) and does **not**
contain c3, p0-auth or e1 — and e2 needs two of those:

- **p0-auth**, because every controller here is `@PreAuthorize("hasRole('ADMIN')")`
  and the port's default-deny + role gate comes from that branch.
- **e1**, because `internal/common/audittrail` is where the `clinlims.history`
  writer lives, and every write below is audited.

e1 itself was cut this way — its ancestry runs c2 → c3 → p0-auth → e1 — so the
documented rule already describes something other than the practice. Recorded
here rather than quietly repeated; the rule should either be amended or the
branches should be rebased, and that is a call for the maintainer.

---

## 2. What e1 leaves behind that e2 can use

| Piece | Where | Reused as-is |
|---|---|---|
| `clinlims.history` writer, `keep_history` gate, XML payload format | `internal/common/audittrail` | yes |
| The rule that an unchanged write writes NOTHING — no row, no `lastupdated`, no audit | e1 `UpdateValue` | the pattern, per table |
| Primary write + side effects in ONE transaction | e1 `Update` / `DAO.Tx` | yes |
| CSRF on writes, and the shape of a refusal | p0-auth + e1 spec | yes |
| Payload field order = entity DECLARED field order | e1 `deleteChanges` | the rule, per entity |

The last one is the trap that will repeat. `getChanges` is reflective and
generic: for every entity written below, the audit payload's field list and
order come from that class's declaration order, and a field is emitted only
when it differs from a blank object. e1 got this wrong once by hardcoding the
fields one row happened to have. There are ~8 entities in e2.

---

## 3. Traps

UOM's are **measured** against the live server. The rest are read off the
source and still have to be.

### 3.1 UOM — measured

Probed `GET`/`POST rest/UomCreate` against Java. Four answers, and two of them
are the opposite of what the source suggests.

**The localization NPE does not exist.** `setupDisplayItems` calls
`uom.getLocalization().getLocalizedValue(locale)` over every UOM, and the
insert path never writes a localization — which reads like a create that breaks
its own form. It does not: `unit_of_measure` has **no localization column at
all**, and `UnitOfMeasure.getLocalization()` is a stub that builds a fresh
`Localization` in memory from the entity's own fields every call. It cannot
return null. Creating a UOM and reloading the form answers **200**, measured.

**`existingFrenchNames` is the literal string `French`, once per UOM.** The
same stub ends with `_localization.setFrench("French")` — hardcoded. The form
really does answer `"$French$French$French$…"`, and it is a wire value the port
has to reproduce rather than a bug to fix in passing.

**The insert writes NO audit row.** `reference_tables` has `UNIT_OF_MEASURE`
with `keep_history = 'Y'`, and `UnitOfMeasureServiceImpl` extends
`AuditableBaseObjectServiceImpl` — but never sets `auditTrailLog = true`, so
the mechanism is off. Measured: creating a UOM leaves `clinlims.history`
untouched. This is the same shape as e1's role side effect, where every
signpost pointed at an audit row that is not written. **A port that audits this
write is wrong.**

**The POST response is not the GET response.** On success the handler returns
the bound form without calling `setupDisplayItems`, so the body is
`{formName, formMethod, cancelAction, submitOnCancel, cancelMethod,
uomEnglishName}` — the display lists are absent. They appear only on the
validation-failure branch, which returns 200 as well. Same trap as e1's echo
form.

The row itself: `name` and `description` both take the submitted name,
`is_active` takes the column default `'Y'`, `lastupdated` is set.
`createUnitOfMeasure(identifyingName, userId)` accepts `userId` and drops it —
with no audit row, it has nowhere to show, but the same signature recurs in the
sibling controllers where it may.

**A rename moves `name` and leaves `description`.** `updateUomNames` calls
`setUnitOfMeasureName(nameEnglish.trim())` and nothing else, so the two columns
agree the moment a UOM is created and disagree from the first rename onward.
The trim is Java's, not the database's. An id that does not exist is a silent
**200**: the block is guarded by `if (unitOfMeasure != null)` and skipped
whole.

**`inactiveUomList` is not an inactive list, and it does not refresh.** This is
the one worth reading twice. `DisplayListService` builds `UNIT_OF_MEASURE` with
`createUnitOfMeasureList()` and `UNIT_OF_MEASURE_INACTIVE` with
`createUOMList()` — two names for the same six lines, `getAll()` mapped to
`(id, localizedName)`, neither filtering on `is_active`; the filter survives in
the first only as a commented-out line. The create handler then refreshes both,
and only one refresh exists: `refreshList`'s switch has a case for
`UNIT_OF_MEASURE` and none for `UNIT_OF_MEASURE_INACTIVE`, so the second call
falls through in silence and that list keeps its startup snapshot for the life
of the process.

Measured: after a create, `existingUomList` carries the new UOM and
`inactiveUomList` does not, and the latter is byte-identical to what it was
before the write. The run that pinned this went further — `existingUomList`
held a UOM that had already been deleted from the table, because nothing
refreshes on a delete either. **The port caches both lists at startup and
reloads only the first**, which is the e1-8 lesson applied to a second cache.

### 3.2 Still to measure

**The insert failure is swallowed.** `catch (LIMSRuntimeException e)
{ LogEvent.logDebug(e); }` — a failed insert returns 200 with the form. e1-5
measured the opposite for site_information, where the failure surfaced at the
transaction boundary as Tomcat's 500 page. Both shapes exist in this codebase;
which one applies is per-endpoint. `unit_of_measure.name` is `varchar(26)`,
which is the same lever e1-5 used.

**`$`-separated name strings.** `getExistingUomNames` builds
`"$name1$name2$"` — leading and trailing separator included. Confirmed on the
wire; the sibling controllers build their own and have not been checked.

**`DisplayListService.refreshList`.** The same cache-not-view problem e1-8
settled for `ConfigurationProperties`: these lists are held in memory and
refreshed only by a write through the application. The port must cache and
refresh on the same events, not read the table per request — being more correct
than Java is the failure mode this project treats as a defect.

**Bulk writes across tests.** `PUT /group/ranges` and `PUT /group/storage`
change many rows in one request. The transaction boundary and the number of
audit rows are both observable, and both need measuring before either is
ported.

---

## 4. Why the reads come with the writes

e1 could assert on the database alone because `site_information` is flat. The
catalog is not: a test's ranges, storage, panels and terminology live in
separate tables reached through joins, and the endpoint that renders them is
the only practical oracle for the endpoint that changes them.

So each group below ports **the write and the read that shows it**, even where
the read was not on the wave list. The alternative — asserting with hand-written
SQL against five join tables — is how a spec ends up testing its own query
instead of the port.

---

## 5. Order of work

All of it ships on this branch. The order below is what each group teaches the
next, and blast radius — clinical writes last. It is not a scope boundary.

| # | Group | Endpoints | Why here |
|---|---|---|---|
| 1 ✅ | **UOM** — `UomCreate`, `UomRenameEntry` | 2 W + 2 R | Smallest whole module. Establishes the GET-form/POST-save shape all 24 `testconfiguration` controllers share, the localization question, and the audit payload for a non-`site_information` entity. |
| 2 ✅ | **Method**, **TestSection**, **SampleType**, **Panel** — create / rename / order / assign | 13 W + 13 R | Structurally identical to the UOM pair. Cheap once the shape is proven; the differences are the join tables in `*TestAssign` and `*Order`. |
| 3 ✅ | **Select lists** — `ResultSelectListAdd`, `SelectListRenameEntry` | 3 W + 2 R | Dictionary-backed; feeds result entry. |
| 4 ◑ | **Test lifecycle** — `TestAdd`, `TestModifyEntry`, `TestActivation`, `TestOrderability`, `TestRenameEntry`, `POST tests/{id}/activate` | 6 W + 5 R | Touches `test`, which c2 and c3 read. All but `POST tests/{id}/activate` are done — see §7. |
| 5 | **Editor, non-clinical** — `POST /tests`, `POST /panels`, `basic-info`, `terminology`, `storage`, `group/storage`, `sample-types/{id}/test-order`, `tests/{id}/panels`, `sample-results`, `copy-from` | 10 W + ~12 R | The modern editor surface. Bulk writes appear here. |
| 6 | **Ranges** 🔴 — `PUT /tests/{testId}/ranges`, `PUT /group/ranges` | 2 W + 2 R | Reference ranges decide whether a result reads as normal. Last, and only with the rest green. |

Each group: measure against live Java → spec that passes on Java and fails on
the port → port → gate (`api-readonly`, `api-mutating`, `go-parity`) → record
what was found in [open-items.md](open-items.md).

---

## 6. Open question for the maintainer

[branch-naming.md](branch-naming.md) says, in bold: *"Before adding or updating
any e2e test, ASK the user whether it should be added to e2e. Do not add e2e
specs unprompted."* e1's parity specs nonetheless live on the migration branch,
in `openelis-api-e2e`, which is how they run in the `go-parity` gate.

e2 needs the same thing 38 times. Before the first spec is written, confirm:
specs land on `migration/e2-testcatalog-writes` alongside the port, as e1 did —
or on a separate `e2e/` branch off `develop`, in which case the gate cannot run
against the port until both are merged.

---

## 7. Progress — 2026-08-30 JST

Paused mid-branch at the maintainer's request. This section is the state of the
work, not a scope change: §5 still applies and the branch still owns all 38
writes.

### 7.1 Landed — 23 of 38 writes, 9 commits

Every commit below passed all three gates before it was made: `api-readonly`,
`api-mutating`, `go-parity`, zero failures.

| Commit | Endpoints |
|---|---|
| `227a0229b` | `UomCreate`, `UomRenameEntry` |
| `70925b499` | `MethodRenameEntry`, `TestSectionRenameEntry`, `SampleTypeRenameEntry`, `PanelRenameEntry` |
| `1916cb19e` | `MethodCreate`, `TestSectionCreate`, `SampleTypeCreate` |
| `7fefcc43e` | `PanelCreate` |
| `b9473c1fd` | `PanelOrder`, `TestSectionOrder`, `SampleTypeOrder` |
| `1a5f38f45` | `TestActivation`, `TestOrderability` |
| `d89de5b82` | `SampleTypeTestAssign`, `TestSectionTestAssign`, `PanelTestAssign` |
| `064f5df2b` | `TestRenameEntry`, `SelectListRenameEntry`, `ResultSelectListAdd`, `SaveResultSelectList` |
| `9f1d93335` | `TestAdd` |

`9f1d93335` also carries three fixes the port needed and one harness fix:

- the Go connection now names `clinlims` on its `search_path`, because the
  `test` table's BEFORE INSERT trigger calls `UNACCENT()` unqualified and
  `unaccent` lives in that schema. Java never has to say this: it connects as
  the `clinlims` user, so `"$user"` resolves it.
- `ActiveHumanSampleTypes` reads its localized name through a scalar subquery
  rather than a LEFT JOIN, so the row source stays the plain scan Java's HQL
  produces and the `sort_order` tie order matches.
- `go-parity` runs on ONE worker, like `api-mutating`. `fullyParallel: false`
  only serialises tests within a file, so separate mutating spec files had been
  racing against a single shared database.

### 7.2 In the working tree, NOT committed — `TestModifyEntry`

Implementation and spec are complete and green:

```
migration/openelis-go/internal/testconfiguration/daoimpl/testmodify_dao.go
migration/openelis-go/internal/testconfiguration/daoimpl/testmodify_read_dao.go
migration/openelis-go/internal/testconfiguration/form/testmodify_forms.go
migration/openelis-go/internal/testconfiguration/service/testmodify_service.go
migration/openelis-go/internal/testconfiguration/controller/rest/testmodify.go
migration/openelis-api-e2e/tests/mutating/e2-testmodify-writes.spec.ts
migration/openelis-api-e2e/playwright.config.ts      (spec added to go-parity)
migration/openelis-go/cmd/openelis/main.go           (wiring)
migration/openelis-go/internal/testconfiguration/service/testadd_service.go
                                                     (buildTestAddRow extracted
                                                      so both endpoints share it)
```

Measured state at the pause:

- `e2-testmodify-writes.spec.ts` — 7/7 against Java, 7/7 against Go.
- `api-readonly` — 637 passed, 0 failed.
- `api-mutating` — 101 passed, 0 failed.
- `go-parity` — **NOT re-run in full since this change entered the tree.**

That last line is why it is uncommitted. Rule 1 of AGENTS.md is that a failing
gate is never committed, and an unrun gate is not a passing one. To land it,
run the go-parity project with the Go service up, then commit if it is clean.

### 7.3 What TestAdd and TestModifyEntry measured

Recorded here because none of it is visible from the source alone.

- The sample type follows the new test's OWN active flag, so creating an
  inactive test DEACTIVATES a live sample type. An inactive test section or
  panel is turned back ON just by being named.
- `addTests` only mutates the in-memory test when a dictionary option is the
  default — but it is a managed entity, so the flush writes
  `default_test_result_id` anyway.
- Titer is in no branch of `createTestResults`: a Titer test is created with no
  results at all.
- The ResultLimit ENTITY defaults, not the column defaults, are what land. The
  two disagree on `low_critical` and the entity's `+Infinity` wins.
- `test.name` is not among the columns `TestModifyEntry` updates and moves
  anyway: Hibernate maps the column to `Test.getName()`, a DERIVED getter that
  returns the localization's value. `description` and `local_code` are never
  rewritten, so a renamed test keeps describing itself by its old name.
- A NUMERIC modify does not deactivate the results it replaces — only the
  dictionary variants do — so every numeric save leaves another active
  `test_result` row behind. Reproduced, not fixed.
- The audit is far narrower than either write. TestAdd: one `I` for the test,
  one per result limit, and a `U` for the sample type only when its flag really
  moved. TestModifyEntry: one `D` per deleted result limit and one `I` per
  inserted one, and nothing else — the test updates, the localization edits, the
  join-row churn and the component sync are all silent, several of them from
  tables flagged `keep_history='Y'`.
- `normalized_description` comes from a BEFORE INSERT trigger, so both stacks
  get it from the same place.

### 7.4 Remaining — 14 of 38 writes

**`testcatalog` / `TestCatalogEditorRestController`** (12 writes, ~15 companion
reads; b1 ported only `lab-units`, `sample-types`, `panels`):

| Endpoint | Group |
|---|---|
| `POST /rest/test-catalog/tests` | 5 |
| `PUT /rest/test-catalog/tests/{testId}/basic-info` | 5 |
| `PUT /rest/test-catalog/tests/{testId}/sample-results` | 5 |
| `POST /rest/test-catalog/tests/{testId}/sample-results/copy-from/{sourceId}` | 5 |
| `PUT /rest/test-catalog/tests/{testId}/storage` | 5 |
| `PUT /rest/test-catalog/group/storage` | 5 |
| `PUT /rest/test-catalog/sample-types/{sampleTypeId}/test-order` | 5 |
| `PUT /rest/test-catalog/tests/{testId}/terminology` | 5 |
| `PUT /rest/test-catalog/tests/{testId}/panels` | 5 |
| `POST /rest/test-catalog/panels` | 5 |
| `PUT /rest/test-catalog/tests/{testId}/ranges` 🔴 | 6 |
| `PUT /rest/test-catalog/group/ranges` 🔴 | 6 |

**`testcatalog` / `TestCatalogActivationRestController`** (1 write):

| Endpoint | Group |
|---|---|
| `POST /rest/test-catalog/tests/{testId}/activate` | 4 |

**Plus** the reads each of those pairs with — `GET /tests`, `GET /tests/{id}`,
`/localization`, `/loinc-integrity`, `/basic-info`, `/sample-results`,
`/dictionary`, `/ranges`, `/siblings`, `/group/summary`, `/storage`,
`/analyzers`, `/sample-types/{id}/test-order`, `/terminology` — for the reason
§4 gives: a write cannot be measured through a read that does not exist yet.

### 7.5 Environment notes for whoever picks this up

WSL localhost port forwarding was down for part of this session. Both notes
below are environment variables, not code changes.

Playwright against Java, when `https://localhost/` refuses the connection:

    OE_BASE_URL="https://<wsl-ip>/api/OpenELIS-Global/" npx playwright test --project=api-mutating

The Go service against the dockerised database:

    TZ=UTC OE_DB_HOST=<wsl-ip> OE_DB_PORT=15432 go run ./cmd/openelis

`TZ=UTC` is not optional. `rest/server-time` reads `$TZ` first and falls through
to the zone ABBREVIATION when it is unset, which answers `JST` on this host and
fails the IANA check in `a1-server-time`. The Java container runs `TZ=UTC`.

Get the WSL IP with `wsl -d Ubuntu-24.04 -- hostname -I` (first address).
