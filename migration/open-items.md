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

## e1 — admin config CRUD (Wave 6)

Full context in [e1-config-crud-migration.md](e1-config-crud-migration.md).

### 🔴 Divergent today

| # | What | Found by | What it takes |
|---|---|---|---|
| e1-5 | **Write-failure paths are not ported.** Java distinguishes `StaleObjectStateException` (→ `errors.OptimisticLockException`) from any other failure (→ `errors.UpdateException`) and returns the form carrying the error. The port answers a bare 500. | source-read | Needs the `saveErrors` envelope measured first — it is the same shape question as e1-2. |

### Side effects not ported

| # | What | Found by | Effect |
|---|---|---|---|
| e1-8 | `ConfigurationProperties.loadDBValuesIntoConfiguration()` and `DisplayListService.getInstance().refreshLists()` run after **every** write. The Go service builds its display lists once at startup and never refreshes them, so a write that changes a list-backing row leaves Go serving stale lists where Java serves fresh ones. 🔴 in principle; not yet demonstrated. | source-read | Needs a cache-invalidation hook on the write path. |
| e1-10 | `localeResolver.setLocale(...)` runs when the row written is `default language locale`, changing the request's locale in place. | source-read | Only observable across a locale change. |

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

Two things surfaced only because those checks existed:

- **An update that changes nothing writes nothing** — not the row, not
  `lastupdated`, not an audit row. The port had been stamping
  `lastupdated = now()` unconditionally, so every save button press moved it.
- **Java's `history.id` is not chronological.** Hibernate hands ids out of a
  cached sequence block, so an insert's audit row can carry a HIGHER id than
  the update that followed it — 866 (I) against 845 (U) and 846 (D) on one row.
  Anything reading the audit trail has to order by `timestamp`.

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

---

## Suggested order

1. **e1-5** — the write-failure envelopes. Reachable only on a database error,
   so it needs a way to provoke one before it can be pinned.
2. **e1-8** — the display-list cache the port never invalidates.
3. **e1-10** — `localeResolver.setLocale` on the default-language row.
