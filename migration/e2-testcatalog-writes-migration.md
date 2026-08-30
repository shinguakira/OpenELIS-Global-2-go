# e2 — Test-catalog writes (scoped migration plan)

Status: **complete — all 37 writes and all 45 reads ported; three gates green**
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

Of this controller's **19 reads**, three are ported (b1 did `lab-units`,
`sample-types`, `panels`). The rest are unported, and several are the only way
to SEE what the write beside them did — `GET /tests/{testId}/ranges` is how you
check `PUT /tests/{testId}/ranges`.

Two more controllers in the same package are read-only and unported:

| Method | Path | Controller |
|---|---|---|
| GET | `/rest/test-catalog/{testId}/reflex-calc` | `TestReflexCalcRestController` |
| GET | `/rest/test-catalog/{testId}/storage/history` | `TestStorageHistoryRestController` |

**`testcatalog` totals: 13 writes, 21 reads.**

### 0.2 `testconfiguration` — the legacy screens, now REST (`/rest/{Name}`)

24 controllers, each a **GET form / POST save** pair. 23 carry a write; **24
write mappings** in total (`ResultSelectListAdd` has two).
`TestCatalogRestController` is read-only and was ported in b1, so **23 of the 24
reads** are unported.

**`testconfiguration` totals: 24 writes, 24 reads.**

| Group | Controllers |
|---|---|
| Test | `TestAdd` · `TestModifyEntry` · `TestActivation` · `TestOrderability` · `TestRenameEntry` |
| Panel | `PanelCreate` · `PanelOrder` · `PanelRenameEntry` · `PanelTestAssign` |
| Sample type | `SampleTypeCreate` · `SampleTypeOrder` · `SampleTypeRenameEntry` · `SampleTypeTestAssign` |
| Test section | `TestSectionCreate` · `TestSectionOrder` · `TestSectionRenameEntry` · `TestSectionTestAssign` |
| Method | `MethodCreate` · `MethodRenameEntry` |
| UOM | `UomCreate` · `UomRenameEntry` |
| Select lists | `ResultSelectListAdd` (2 writes) · `SelectListRenameEntry` |

**Total: 37 write endpoints and 45 reads, of which 41 reads are unported**
(b1 had done four: `lab-units`, `sample-types`, `panels`, `TestCatalog`).

Counted from the source, not estimated — an earlier draft of this section said
38 writes, 25 `testconfiguration` write mappings, 18 `testcatalog` reads and
"~30 companion reads", and every one of those numbers was wrong. The counts
above come from `grep -c '@PostMapping\|@PutMapping\|@DeleteMapping'` and
`grep -c '@GetMapping'` over both packages.

That is bigger than every previous wave, and it is all one branch. §5 is the
ORDER the work is done in, not a smaller scope: **the branch is finished when it
covers every endpoint above — the 37 writes AND the 41 unported reads.**

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
documented rule already describes something other than the practice.

**Settled by the maintainer: forking from e1 is accepted for this branch, this
time.** It is not a new rule. A later branch that wants the same has to ask
again, or `branch-naming.md` has to be amended.

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

So each group below ports **the write and the read that shows it first**, even
where the read was not on the wave list. The alternative — asserting with
hand-written SQL against five join tables — is how a spec ends up testing its
own query instead of the port.

**That is a sequencing rule, not a scope rule.** It says which read to reach for
while a write is being ported; it does not say the other reads are out. §0's
completion criterion stands: every endpoint in §0, writes and reads alike. The
reads that pair with no write are group 7 in §5 — last, because nothing depends
on them, not because they are optional.

An earlier version of this paragraph said only "the write and the read that
shows it", with no such qualifier, and it was read as the scope. The 37 writes
were delivered, ten `testcatalog` reads were left unported, and §7's status line
was changed to "complete" — all of it defensible against this paragraph and none
of it against §0. See AGENTS.md rule 4, which exists because of exactly this.

---

## 5. Order of work

All of it ships on this branch. The order below is what each group teaches the
next, and blast radius — clinical writes last. It is not a scope boundary.

| # | Group | Endpoints | Why here |
|---|---|---|---|
| 1 ✅ | **UOM** — `UomCreate`, `UomRenameEntry` | 2 W + 2 R | Smallest whole module. Establishes the GET-form/POST-save shape all 24 `testconfiguration` controllers share, the localization question, and the audit payload for a non-`site_information` entity. |
| 2 ✅ | **Method**, **TestSection**, **SampleType**, **Panel** — create / rename / order / assign | 14 W + 14 R | Structurally identical to the UOM pair. Cheap once the shape is proven; the differences are the join tables in `*TestAssign` and `*Order`. |
| 3 ✅ | **Select lists** — `ResultSelectListAdd`, `SelectListRenameEntry` | 3 W + 2 R | Dictionary-backed; feeds result entry. |
| 4 ✅ | **Test lifecycle** — `TestAdd`, `TestModifyEntry`, `TestActivation`, `TestOrderability`, `TestRenameEntry`, `POST tests/{id}/activate` | 6 W + 5 R | Touches `test`, which c2 and c3 read. |
| 5 ✅ | **Editor, non-clinical** — `POST /tests`, `POST /panels`, `basic-info`, `terminology`, `storage`, `group/storage`, `sample-types/{id}/test-order`, `tests/{id}/panels`, `sample-results`, `copy-from` | 10 W + 7 R | The modern editor surface. Bulk writes appear here. |
| 6 ✅ | **Ranges** 🔴 — `PUT /tests/{testId}/ranges`, `PUT /group/ranges` | 2 W + 1 R | Reference ranges decide whether a result reads as normal. Last, and only with the rest green. |
| 7 ✅ | **The reads that pair with no write** — `GET /tests`, `/tests/{testId}`, `/tests/{testId}/localization`, `/tests/{testId}/loinc-integrity`, `/dictionary`, `/tests/{testId}/siblings`, `/group/summary`, `/tests/{testId}/analyzers`, `/{testId}/reflex-calc`, `/{testId}/storage/history` | 0 W + 10 R | Last because nothing depends on them, NOT because they are optional — §0 counts them. All ten are `testcatalog`; every `testconfiguration` read came with its write. |

Writes: 2 + 14 + 3 + 6 + 10 + 2 = **37**. Reads: 2 + 14 + 2 + 5 + 7 + 1 + 10 =
**41**, the count §0 gives as unported. Both sums are here so the next edit to
this table has to keep them adding up.

Each group: measure against live Java → spec that passes on Java and fails on
the port → port → gate (`api-readonly`, `api-mutating`, `go-parity`) → record
what was found in [open-items.md](open-items.md).

---

## 6. Where the specs live — settled

[branch-naming.md](branch-naming.md) says, in bold: *"Before adding or updating
any e2e test, ASK the user whether it should be added to e2e. Do not add e2e
specs unprompted."* e1's parity specs nonetheless live on the migration branch,
in `openelis-api-e2e`, which is how they run in the `go-parity` gate.

**Settled by the maintainer: everything for this branch stays ON this branch —
the port and the parity specs alike.** There is no separate `e2e/` branch for
e2, and nothing here is to be deferred elsewhere. The `go-parity` gate runs
against the port in the same commit, which is the point.

---

## 7. Outcome — every endpoint in §0

All 37 writes and all 45 reads are ported: the 31 reads that came with a write,
the 4 b1 had already done, and the 10 of §5 group 7 that pair with none.

Every commit below passed all three gates before it was made — `api-readonly`,
`api-mutating`, `go-parity`, zero failures — and each carries its own parity
spec, written against the LIVE Java server and run against both stacks.

### 7.1 The commits

| Commit | Endpoints | W |
|---|---|---|
| `227a0229b` | `UomCreate`, `UomRenameEntry` | 2 |
| `70925b499` | `MethodRenameEntry`, `TestSectionRenameEntry`, `SampleTypeRenameEntry`, `PanelRenameEntry` | 4 |
| `1916cb19e` | `MethodCreate`, `TestSectionCreate`, `SampleTypeCreate` | 3 |
| `7fefcc43e` | `PanelCreate` | 1 |
| `b9473c1fd` | `PanelOrder`, `TestSectionOrder`, `SampleTypeOrder` | 3 |
| `1a5f38f45` | `TestActivation`, `TestOrderability` | 2 |
| `d89de5b82` | `SampleTypeTestAssign`, `TestSectionTestAssign`, `PanelTestAssign` | 3 |
| `064f5df2b` | `TestRenameEntry`, `SelectListRenameEntry`, `ResultSelectListAdd`, `SaveResultSelectList` | 4 |
| `9f1d93335` | `TestAdd` | 1 |
| `1b0b159de` | `TestModifyEntry` | 1 |
| `60b4f22e4` | editor sections: `storage`, `group/storage`, `terminology`, `sample-types/{id}/test-order`, `tests/{id}/panels`, `POST /panels` | 6 |
| `0e620f436` | editor tests: `POST /tests`, `basic-info`, `sample-results`, `copy-from`, `ranges` 🔴, `group/ranges` 🔴, `activate` | 7 |
| (this one) | group 7 — the ten reads that pair with no write | 0 |

24 writes in `testconfiguration`, 13 in `testcatalog`. The reads that show a
write came with it, for the reason §4 gives; the ten that show none came last,
as §5 group 7.

### 7.2 What the measurements changed

None of this is visible from the Java source alone; each cost a probe against
the running server, and several contradicted a careful reading.

**Writes that are not what they look like**

- The sample type follows a NEW TEST'S own active flag, so creating an inactive
  test through `TestAdd` DEACTIVATES a live sample type. An inactive test
  section or panel is turned back ON just by being named — by `TestAdd`,
  `TestModifyEntry`, `POST /tests` and `basic-info` alike.
- `test.name` is not among the columns `TestModifyEntry` or `basic-info` write,
  and it moves anyway: Hibernate maps the column to `Test.getName()`, a DERIVED
  getter over the localization. `description` and `local_code` are never
  rewritten by the modify path, so a renamed test keeps describing itself by its
  old name.
- A NUMERIC `TestModifyEntry` save does not deactivate the results it replaces —
  only the dictionary variants do — so every numeric save leaves another active
  `test_result` row behind and `getResultType` reads the newest of a growing
  pile. Reproduced, not fixed.
- A storage save is a REPLACE. `group/storage` therefore clears every field its
  document does not name, and its `version` counter is bumped on every save
  whether the state changed or not — only the JSON snapshot row is conditional.
- `group/ranges` DISCARDS the ids it is sent, because a shared band belongs to
  no single test. It always inserts on each test and deletes what was there, so
  running it twice replaces rather than updates.
- `active: true` on `basic-info` is IGNORED. Activation is gated on range
  coverage and has to go through `POST .../activate`, so `basic-info` can only
  ever turn a test OFF.
- Only an OPEN-ENDED top band (max = +Infinity) reaches the top of the
  reportable lifetime, so bands 0–15 and 15–30 leave 30+ uncovered and the test
  cannot be activated without an acknowledgment. A test with NO ranges is EMPTY,
  not GAP, and activates freely.

**The audit is narrower than the write, everywhere**

- `TestAdd`: one `'I'` for the test, one per result limit, and a `'U'` for the
  sample type only when its flag really moved. The localizations, join rows,
  panel items, test results, terminology mapping and component are all silent —
  several from tables flagged `keep_history='Y'`.
- `TestModifyEntry`: a `'D'` per deleted result limit carrying the whole row in
  declared-field order, an `'I'` per inserted one, and nothing else.
- The editor's section saves write NO history at all.
- The editor's test-level writes are the exception: a create leaves an `'I'`
  with a NULL payload, and `basic-info` and `activate` leave a `'U'` carrying
  the values they replaced — in which `testSection` renders the section's
  DESCRIPTION, not its id. A save that changes nothing leaves nothing.

**Traps a port walks into**

- The ResultLimit ENTITY defaults, not the column defaults, are what land. The
  two disagree on `low_critical`, and the entity's `+Infinity` wins.
- `normalized_description` comes from a BEFORE INSERT trigger, so both stacks
  get it from the same place — but the trigger's plpgsql calls `UNACCENT()`
  unqualified, and `unaccent` lives in `clinlims`. The Go connection now names
  that schema on its `search_path`; without it no test can be inserted at all.
  Java never has to say this: it connects as the `clinlims` user, so `"$user"`
  resolves the same schema.
- Imposing an ORDER BY where Java has none is a bug, not a tidy-up. Two of them
  were found: `ActiveHumanSampleTypes` had to read its localized name through a
  scalar subquery rather than a LEFT JOIN so the row source stays the plain scan
  Java's HQL produces, and the terminology reads carry no ORDER BY because
  `getAllMatching` has none — the ids are UUIDs, so sorting them would randomise
  the order the caller sent.
- `go-parity` now runs on one worker, like `api-mutating`. `fullyParallel:
  false` only serialises tests within a file, so separate mutating spec files
  had been racing against a single shared database.

### 7.3 What the ten reads measured

They share the AUGMENTED name — the localized test name with the first sample
type in brackets — and most of what is worth recording is where they differ from
each other about it.

- **`GET /tests` sorts on a name it does not show.** The sort is by the RAW
  localized name; the augmented one is substituted onto the PAGE SLICE
  afterwards, along with the sample type. So `search` matches the analyte and
  not the specimen: `search=Urines` finds nothing while every row's `name` ends
  in `(Urines)`. A port that sorts on the augmented name pages the catalogue in
  a different order.
- **`page` and `pageSize` are clamped UP to 1**, so `?page=0&pageSize=0` is a
  200 carrying one row, not a 400. A page past the end echoes the page it was
  asked for and returns nothing.
- **`siblings` says every sibling is inactive.** It reuses the list page's row
  type and fills only `testId`, `name` and `sampleType`; `active`, `amr` and
  `coverageIncomplete` are primitive booleans that serialise as false because
  nothing set them. `coverageIncomplete` is hardcoded false on the list page
  too — the decoration was left for a later milestone and never wired.
- **`siblings` is also the only endpoint here that answers 200 for an unknown
  test.** It returns the accumulator rather than a `ResponseEntity`, so a
  missing test is an empty list. Every other one is a 404.
- **There are two 404 SHAPES under one path prefix.** The editor's own are
  `ResponseEntity.notFound()` with no body; the two read-only controllers throw
  `ResponseStatusException`, which Spring renders as an RFC 7807 ProblemDetail
  carrying unresolved message keys (`problemDetail.type.org.springframework…`).
- **`/dictionary` matches the abbreviation and the entry as ONE string.** The
  query is `upper(local_abbrev || ': ' || dict_entry) LIKE upper('X%')`, so
  searching `pos` returns "College or University" through its abbreviation. The
  label is the RAW `dict_entry`, not the localized name — the only dictionary
  list in this package that is not localized. A blank search returns nothing, so
  the control does not dump the whole dictionary on focus.
- **`reflex-calc` renders the added test's name UNaugmented**, while everything
  else on the same screen is augmented. Two name rules, one view.
- **`storage/history` serialises the ENTITY**, so its two `jsonb` columns arrive
  as STRINGS of JSON rather than as objects — the entity types them as `String`
  and Jackson has nothing to parse them into. A test with no storage config is
  an empty list; only a missing TEST is a 404.

One routing difference had to be handled rather than reproduced directly.
`tests/{testId}` and `{testId}/reflex-calc` are both two segments, and Go's
`ServeMux` refuses to register a pair where neither pattern is more specific —
it panics at startup rather than guess. Spring does guess, and MEASURED it
prefers the literal SUFFIX: `GET /rest/test-catalog/tests/reflex-calc` answers
500 from reflex-calc looking up a test named "tests", not the envelope for a
test named "reflex-calc". The port dispatches both from one handler in that
order, which reproduces the answer for every path including that one.

### 7.4 Two Java defects, reproduced rather than repaired

Both are recorded in [java-defects-found.md](java-defects-found.md) and pinned
by the specs.

- **`POST /rest/test-catalog/panels` cannot succeed.**
  `panel.name_localization_id` is NOT NULL and `createPanel` never writes a
  localization, so every non-blank name is a 500 and nothing survives it. A
  blank name is a clean 422. The legacy `PanelCreate` screen next door gets this
  right, writing the localization first. Defect 14.
- **`PUT /tests/{testId}/sample-results` is a 500 when a component is re-sent
  without its id.** The match is on id alone, so the component is inserted
  afresh and collides with the `(test_id, code)` unique index. The UI always
  echoes the id; a hand-written client that does not gets a 500 rather than a
  422.

Fixing either upstream is a small change in shape but a behaviour change to a
shipped endpoint, and belongs in its own PR against `develop` — not in a
migration branch whose contract is to reproduce what runs today.

### 7.5 Environment notes

WSL localhost port forwarding was down for part of this work. Both notes below
are environment variables, not code changes.

Playwright against Java, when `https://localhost/` refuses the connection:

    OE_BASE_URL="https://<wsl-ip>/api/OpenELIS-Global/" npx playwright test --project=api-mutating

The Go service against the dockerised database:

    TZ=UTC OE_DB_HOST=<wsl-ip> OE_DB_PORT=15432 go run ./cmd/openelis

`TZ=UTC` is not optional. `rest/server-time` reads `$TZ` first and falls through
to the zone ABBREVIATION when it is unset, which answers `JST` on this host and
fails the IANA check in `a1-server-time`. The Java container runs `TZ=UTC`.

Get the WSL IP with `wsl -d Ubuntu-24.04 -- hostname -I` (first address).
