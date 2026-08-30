# Open items — work deliberately left undone

A register of everything the Go port does **not** yet do, kept in one place so
that "we know about it" is written down rather than remembered.

Nothing here is a bug report against Java. Java defects live in
[java-possible-bugs.md](java-possible-bugs.md) and
[java-defects-found.md](java-defects-found.md); this file is about the **port**.

## How to read this

Every item says how it was identified, because the two carry very different
weight:

| | |
|---|---|
| **measured** | the difference was observed against the live Java server |
| **source-read** | the gap was found by reading the Java source; the divergence is expected but has not been demonstrated |

and what its current effect is:

| | |
|---|---|
| 🔴 **divergent** | reachable input exists today on which Java and Go answer differently, or leave different database state |
| 🟡 **unreachable** | the branch exists in both, but no shipped data reaches it, so the two agree for now |
| ⚪ **harness** | the port is fine; the verification around it is not |

A 🟡 is not a free pass. Every wave so far has produced at least one 🟡 that
turned 🔴 the moment a fixture seeded the row that separates the branches — see
`OpenELIS-Go-Migration-Plan.md` §4. They are listed so the next person can seed
them deliberately instead of discovering them by accident.

---

## e2 — test-catalog writes and reads (Wave 6)

Full context in
[e2-testcatalog-writes-migration.md](e2-testcatalog-writes-migration.md).

Nothing on this branch is unported. The three items below are places where the
port and Java can answer DIFFERENTLY without either being wrong — all three are
Java reading from a cache or an unordered query, and all three are worked around
in the specs rather than in the port.

### 🟡 Agree today, and would not under other data

| # | What | Found by | Effect |
|---|---|---|---|
| e2-1 | `groupedDictionaryList` — the equal-SIZE ordering cannot be reproduced. `createGroupedDictionaryList` deduplicates the option groups through a `HashSet<String>` and then sorts by size with a STABLE sort, so groups of equal size come out in Java's HashSet bucket order — an accident of `String.hashCode` and the table size. The port emits first-appearance order instead. | **measured** — the two lists carry the same 19 groups with the same sizes and the same order WITHIN each group, in a different order between equal sizes. | 🟡 The set, the sizes and the intra-group order are identical, and those are what the screen reads. `e2-testadd-writes.spec.ts` and `e2-testmodify-writes.spec.ts` therefore assert the groups as a SET and the sizes as non-decreasing, not the sequence. Reproducing the bucket order is possible — Java's hash is deterministic — and buys nothing. |
| e2-2 | `DisplayListService`'s cached lists go stale against a direct SQL edit; the port reads live. Java loads `SAMPLE_TYPE_ACTIVE`, `TEST_SECTION_ACTIVE`, `PANELS`, `DICTIONARY_TEST_RESULTS` and the rest into a map refreshed only by a write THROUGH the application. | **measured** — a spec that activated a test section through the API and restored it with SQL left Java serving a section it no longer had, while Go answered from the table. | 🟡 In production nothing edits these tables out of band, so the two agree. Under the suite they do not, which is why `e2-testadd-writes` and `e2-editor-section-writes` end their `afterEach` with a throwaway create through the endpoint: one more application write rebuilds every list Java caches. A spec that restores fixture state with SQL and does NOT resync leaves the next suite reading Java's stale copy — that is how this was found, two files away from the spec that caused it. |
| e2-3 | Tie order under a `sort_order` with heavy duplicates is the query plan's, on both sides. `type_of_sample` has three rows at 0 and seven at `Integer.MAX_VALUE`; `test_section` has thirteen at `MAX_VALUE`. Neither Java's HQL nor the port adds a tiebreak. | **measured** — the order changed between runs after rows were updated, because an UPDATE moves a tuple to the end of the heap. | 🟡 Fixed once already, in the direction of matching Java's row SOURCE rather than imposing an order: `ActiveHumanSampleTypes` now reads its localized name through a scalar subquery instead of a LEFT JOIN, so the plan stays the plain scan Java's HQL produces. The same rule caught the terminology reads, which carry no `ORDER BY` because `getAllMatching` has none. Specs in this wave assert these lists as sets. |

### Not gaps — decisions recorded so they are not re-litigated

- **`POST /rest/test-catalog/panels` answers 500 in the port, on purpose.**
  `panel.name_localization_id` is NOT NULL and Java's `createPanel` never writes
  a localization, so the endpoint cannot succeed. Reproduced rather than
  repaired: a port that supplied the localization would create panel rows this
  Java deployment cannot. Defect 14 in
  [java-defects-found.md](java-defects-found.md).
- **`PUT /tests/{testId}/sample-results` answers 500 for a component re-sent
  without its id, on purpose.** Same reasoning; defect 15.
- **A NUMERIC `TestModifyEntry` save leaves the result it replaced ACTIVE.** Only
  the dictionary variants deactivate the old rows, so every numeric save adds
  another. The port does the same. Not raised as a defect because the editor
  that supersedes this screen does not use the path.
- **`Test.getName()` is derived from the localization, so `test.name` moves on a
  flush that never names the column** — while `description` and `local_code` are
  never rewritten by the modify path. Both stacks behave identically; it is
  written down because it looks like a port bug from either side.

---

## e1 — admin config CRUD (Wave 6)

Full context in [e1-config-crud-migration.md](e1-config-crud-migration.md).

### 🔴 Divergent today

| # | What | Found by | What it takes |
|---|---|---|---|

### Side effects not ported

| # | What | Found by | Effect |
|---|---|---|---|

## Closed

| # | What | How |
|---|---|---|
| e1-1 | Writes left no audit row. Java writes a `clinlims.history` row for every insert, update and delete; the port now writes the same three, with the same payloads. | Ported `internal/common/audittrail` — shared, because every future write wave needs it. The payload is the row's state BEFORE the write: NULL on insert, `<value>old</value>` on update, the nine-field row dump on delete. Verified byte-identical against Java apart from the wall-clock `<lastupdated>`. |
| e1-11 | e1 had no e2e spec and was not in the gate. | `tests/mutating/e1-config-crud.spec.ts`, eight tests, added to the `go-parity` `testMatch`. Written audit-assertion-first: it passed against Java and FAILED against Go on "insert leaves an I audit row" before the port was fixed. |
| e1-12 | The write probe compared only the target table. | It now dumps `history` at each step. |
| e1-2 | The POST validators were not ported. | `validForm` carries all three allow-lists. A rejection is a **200 with the form echoed back and no write** — measured on every list; the only observable difference between accept and refuse is whether the row appears, which is why the spec asserts on the table. |
| e1-3 | The localization POST branch was not ported. | `updateLocalization`. A row tagged `localization` writes to **localization_value** and leaves `site_information` untouched, `lastupdated` included. Pinned against the shipped `bannerHeading` row, which the spec restores. |
| e1-4 | `isValid` was not ported. | The required name and the four phone rules, which key on the row's NAME rather than on any column — so the same value is refused for `phone format` and accepted for another row. Both directions asserted. |
| e1-6 | jasypt `AES256TextEncryptor` was not ported. | Parameters read out of jasypt with a JDK instead of guessed: PBEWithHMACSHA512AndAES_256, PBKDF2-HMAC-SHA512, 1000 iterations, and the layout is **salt‖iv‖ciphertext** — the opposite of what the jasypt source reads like, settled by decrypting its own output. The password is `kspass`, from `volume/properties/common.properties`, not the `dev` default. The port encrypts on write, decrypts on read, and masks the **decrypted** value on the menu. |
| e1-7 | `isEditable`'s sample-count half was not ported. | Computed rather than assumed. |
| e1-9 | `configurationSideEffects.siteInformationChanged` was not ported. | The `modify results role` branch toggles the **"Results modifier" role** in another table, and is pinned. The names are the PROPERTY DB names — `modify results role`, `siteNumber` — not the enum constants; matching the constants would compile and never fire, which is what the port did until the spec caught it. The `siteNumber` branch has no row in this deployment and stays unverified. |
| e1-13 | `GET rest/labUnit/config` was not ported. | Ported, and it drops `labName` when the value is BLANK rather than only when the row is missing — site_information row 33 exists with an empty value and Java still omits the key. |
| e1-5 | Write-failure paths were not ported. | **measured** by writing a name longer than `site_information.name`'s varchar(32): Java answers Tomcat's `{"timestamp","status":500,"error"}` page, NOT the form the controller looks like it returns — the failure surfaces at the transaction boundary and the saveErrors/UpdateException branch is never reached. The port answered a plain-text 500; it now answers the same envelope. |
| e1-8 | The configuration cache was not ported — and the register had it BACKWARDS. | Java's ConfigurationProperties is loaded at startup and refreshed only by a write through the application, so a row changed by anything else is invisible. The port read the table per request, which made it MORE correct than Java and therefore wrong: measured by editing `acessionFormat` directly, where Java kept answering the old value and the port answered the new one. The port now caches and reloads on write. |
| e1-10 | `localeResolver.setLocale` was not ported, and "no data to test it with" was the reason given. That was wrong: the data was two SQL statements away. | The spec seeds the French text itself, flips the locale through the API, and asserts the next response comes back in French — `localizedValue` and the display-language names both. The port's locale now comes from the configuration cache, so a write to the row changes it. The test restores the language INLINE and proves it took: `GlobalLocaleResolver` holds `currentLocale` in one field for the whole process, so while it runs the entire deployment is French, and a restore that silently failed handed every later test a French server — which is how the flake was found. |

| e1-14 | The write path read `encrypted` off the SUBMISSION on both branches. | Java reads it off the submission only for a NEW row; an existing one is loaded by id and only `setValue` is called, so `encryptSiteInformation` tests the flag the COLUMN holds. The port broke the row both ways — an update with `encrypted` omitted stored plaintext under a row still flagged encrypted (every later read then 500s), and `encrypted: true` on an unencrypted row stored ciphertext nothing would ever decrypt. Both directions measured against Java and pinned. |
| e1-15 | The delete payload's field list was hardcoded to the nine columns a created row has. | `getChanges` walks the entity's DECLARED fields and keeps the ones that differ from a blank object, so a populated `tag`, `instruction_key`, `dictionary_category_id` or `description_key` IS in the payload. The shipped `bannerHeading` row (82) carries three of them, sits in the SiteInformation domain and deletes like any other row. The field ORDER is declaration order, which is why `instructionKey` lands between `valueType` and `domain` — measured, then covered by a unit test that needs no server. |
| e1-16 | The primary write and the side effects were in SEPARATE transactions. | `persistData` is `@Transactional` and covers both, so a failing side effect takes the configuration write and its audit row down with it. The port committed the write, then ran the side effects, then answered 500 over a half-applied change. One transaction now; `loadDBValuesIntoConfiguration` stays outside it, where Java has it. The `siteNumber` branch also carried a hardcoded `sysUserId = 1`, which was right for admin and wrong for everyone else. |
| e1-17 | The deployment compose did not pass `OE_ENCRYPTION_PASSWORD`. | Running the documented side-by-side command gave Go Spring's `dev` default while the Java container beside it used `kspass` from `volume/properties/common.properties` — on the SAME database. Neither could read the other's encrypted rows, and nothing failed at startup to say so. `docker-compose.go.yml` now carries the key, guarded by a test that compares it against the properties file. |

Two things surfaced only because those checks existed:

- **An update that changes nothing writes nothing** — not the row, not
  `lastupdated`, not an audit row. The port had been stamping
  `lastupdated = now()` unconditionally, so every save button press moved it.
- **Java's `history.id` is not chronological.** Hibernate hands ids out of a
  cached sequence block, so an insert's audit row can carry a HIGHER id than
  the update that followed it — 866 (I) against 845 (U) and 846 (D) on one row.
  Anything reading the audit trail has to order by `timestamp`.

One reported gap turned out **not** to be one, and it is recorded because the
evidence points the wrong way:

- **The role side effect writes NO audit row, and that is Java's behaviour.**
  Everything on paper says otherwise — `siteInformationChanged` goes through
  `roleService.update`, `RoleServiceImpl` sets `auditTrailLog = true`, and
  `reference_tables` ships THREE rows named `SYSTEM_ROLE` (172, 174, 177), all
  `keep_history = 'Y'`. Toggling `modify results role` through Java flips
  `system_role.active` and leaves exactly ONE history row, for
  `site_information`, and nothing under any of those ids. The port's bare
  `UPDATE` is therefore the faithful behaviour; adding the audit row would make
  it more correct than the thing it ports. `e1-config-parity-gaps.spec.ts` pins
  the absence so the argument does not have to be had twice.

---

## Earlier waves

| # | Wave | What | Effect |
|---|---|---|---|
| b2-1 | b2 | `rest/user-programs` is not implemented in Go. `b2-organization.spec.ts` skips it explicitly under `go-parity`, so the gap is at least visible in the run output. | 🟡 |
| w0-1 | 0 | `rest/logging`, `rest/logging/stream`, `rest/logging/test` are not started. The library decision is made (`zap`, see [logging-adoption-plan.md](logging-adoption-plan.md)); the port is not scheduled and blocks nothing. | 🟡 |
| w0-2 | 0 | `health` answers a placeholder `{"status":"UP"}` rather than porting Java's `/health/odoo` connectivity check. | 🟡 |

---

## A note on running the gate

The Go service must be started with `OE_ENCRYPTION_PASSWORD` set to the same
key Java uses, or the encrypted-row test fails against Go for a configuration
reason rather than a porting one. This deployment's key is `kspass`.

The Java side must be reachable. When WSL localhost forwarding is down, pass the
WSL IP instead: `OE_BASE_URL="https://<wsl-ip>/api/OpenELIS-Global/"` for the
Playwright projects, and `OE_DB_HOST=<wsl-ip> OE_DB_PORT=15432` for the Go
service. The Go service also needs `TZ=UTC` — `rest/server-time` reads the `TZ`
environment variable first and falls through to the zone ABBREVIATION when it is
unset, which answers `JST` on a Japanese host and fails the IANA check in
`a1-server-time`. The Java container runs `TZ=UTC`.

---

## Suggested order

Nothing is open on e1 or e2 — e2-1 through e2-3 are agreements that hold on the
shipped data and are pinned as such, not work left undone.

The one item worth carrying forward is not an e1 gap but a cross-cutting one:
every other ported endpoint takes its locale as a value captured at STARTUP
(`ActiveLocale` is passed to each DAO at construction), while Java resolves it
per request. A locale change therefore switches Java's other endpoints and not
the port's. It is invisible today because no localization row outside the one
this spec seeds has French text that differs from its English text — the same
condition that hid e1-10, so it should be seeded rather than waited for.
