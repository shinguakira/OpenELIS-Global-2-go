# e2 — Test-catalog writes (scoped migration plan)

Status: **opening — scope measured, nothing ported yet**
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

That is bigger than every previous wave. It does not land in one branch-worth
of work, and §5 slices it.

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

Slice 1's are **measured** against the live server. The rest are read off the
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

So each slice below ports **the write and the read that shows it**, even where
the read was not on the wave list. The alternative — asserting with hand-written
SQL against five join tables — is how a spec ends up testing its own query
instead of the port.

---

## 5. Order of work

Ordered by how much each slice teaches about the next, and by blast radius.
Clinical writes last.

| # | Slice | Endpoints | Why here |
|---|---|---|---|
| 1 | **UOM** — `UomCreate`, `UomRenameEntry` | 2 W + 2 R | Smallest whole module. Establishes the GET-form/POST-save shape all 24 `testconfiguration` controllers share, the localization question, and the audit payload for a non-`site_information` entity. |
| 2 | **Method**, **TestSection**, **SampleType**, **Panel** — create / rename / order / assign | 13 W + 13 R | Structurally identical to slice 1. Cheap once the shape is proven; the differences are the join tables in `*TestAssign` and `*Order`. |
| 3 | **Select lists** — `ResultSelectListAdd`, `SelectListRenameEntry` | 3 W + 2 R | Dictionary-backed; feeds result entry. |
| 4 | **Test lifecycle** — `TestAdd`, `TestModifyEntry`, `TestActivation`, `TestOrderability`, `TestRenameEntry`, `POST tests/{id}/activate` | 6 W + 5 R | Touches `test`, which c2 and c3 read. |
| 5 | **Editor, non-clinical** — `POST /tests`, `POST /panels`, `basic-info`, `terminology`, `storage`, `group/storage`, `sample-types/{id}/test-order`, `tests/{id}/panels`, `sample-results`, `copy-from` | 10 W + ~12 R | The modern editor surface. Bulk writes appear here. |
| 6 | **Ranges** 🔴 — `PUT /tests/{testId}/ranges`, `PUT /group/ranges` | 2 W + 2 R | Reference ranges decide whether a result reads as normal. Last, and only with the rest green. |

Each slice: measure against live Java → spec that passes on Java and fails on
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
