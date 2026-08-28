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
| e1-1 | **Writes leave no audit row.** Java writes a `clinlims.history` row for every insert, update and delete — activity `I`/`U`/`D` plus `sys_user_id` — and the port writes none. The table already holds 18 such rows for `site_information`, and the three left by this wave's own Java probe are ids 847/848/849. | **measured** | Port the audit write into the DAO's three mutating methods, and teach the write probe to compare `history` as well as the target table (see e1-12). |
| e1-2 | **The POST validators are not ported.** `SiteInformationFormValidator` checks `valueType`, `siteInfoDomainName` and `tag` against allow-lists; the port accepts anything. Note the domain list here is a THIRD list — it includes `externalConnections` and `validationConfig`, which the delete validator's list does not. | source-read | Three allow-lists plus the `errors` envelope Java returns on rejection. Measure the rejection shape first — it is a `@Valid` form, so likely the per-field `errors` map rather than the ObjectError array the delete uses. |
| e1-3 | **The localization POST branch is not ported.** `if ("localization".equals(form.getTag()))` sends the write to `validateAndUpdateLocalization`, which updates the **localization** table and never touches `site_information`. The port writes `site_information.value` regardless. Reachable today: `bannerHeading` (id 82) is tagged `localization`. | source-read | Port `validateAndUpdateLocalization`, including `languageChanged` and the per-locale update. Measure against id 82 with a restorable value. |
| e1-4 | **`isValid` is not ported.** A blank `paramName` is rejected by Java (`error.SiteInformation.name.required`), as are malformed values for `phone format`, `phone format label`, `phone international format label` and `phone international validation`. The port stores them. | source-read | Port the four checks and `PhoneNumberService.validatePhoneFormat`. |
| e1-5 | **Write-failure paths are not ported.** Java distinguishes `StaleObjectStateException` (→ `errors.OptimisticLockException`) from any other failure (→ `errors.UpdateException`) and returns the form carrying the error. The port answers a bare 500. | source-read | Needs the `saveErrors` envelope measured first — it is the same shape question as e1-2. |

### 🟡 Unreachable on this data

| # | What | Found by | What it takes |
|---|---|---|---|
| e1-6 | **jasypt `AES256TextEncryptor` is not ported**, so `hideEncryptedFields` masks the stored ciphertext where Java masks the decrypted plaintext — sixty-four asterisks against twelve. No shipped row carries `encrypted = true`. | **measured** — Java's own POST produced the ciphertext, and PBKDF2-HMAC over SHA-512/256/1, both salt/IV orders, 1..5000 iterations does not reproduce it against the known plaintext | Read jasypt's actual key-derivation parameters instead of guessing. The deployed container ships a JRE with no `javap` and no compiler, so this needs a JDK. **Do not seed an encrypted row until it works**: a plaintext value in that column makes Java 500 on every config menu, because the decrypt throws and the resulting error object is itself unserialisable. |
| e1-7 | **`isEditable`'s sample-count half is not ported.** Java returns `sampleService.getCount() == 0` for the accession-prefix row; the port returns true. Reachable only on a database with zero samples, and this one has samples. | source-read | One count query. Cheap, but there is no way to observe it here without emptying the sample table. |

### Side effects not ported

| # | What | Found by | Effect |
|---|---|---|---|
| e1-8 | `ConfigurationProperties.loadDBValuesIntoConfiguration()` and `DisplayListService.getInstance().refreshLists()` run after **every** write. The Go service builds its display lists once at startup and never refreshes them, so a write that changes a list-backing row leaves Go serving stale lists where Java serves fresh ones. 🔴 in principle; not yet demonstrated. | source-read | Needs a cache-invalidation hook on the write path. |
| e1-9 | `configurationSideEffects.siteInformationChanged(siteInformation)` runs on every persist. Its contents have not been read yet. | source-read | Read `common/util/ConfigurationSideEffects.java` and decide per side effect. |
| e1-10 | `localeResolver.setLocale(...)` runs when the row written is `default language locale`, changing the request's locale in place. | source-read | Only observable across a locale change. |

### Not ported, same controller

| # | What |
|---|---|
| e1-13 | `GET rest/labUnit/config` lives in `SiteInformationRestController` and is Wave 1 item 1.42. It was left alone because it belongs to a different wave, not because it is hard. |

---

## Harness and ledger

| # | What | Effect |
|---|---|---|
| e1-11 | **e1 has no e2e spec.** Every check in this wave was an ad-hoc script under the session scratchpad — a field diff, a byte diff and a write probe — none of which is part of any gate. Until a spec exists and its filename is added to `go-parity`'s `testMatch` in `playwright.config.ts`, nothing guards e1 against regression. This is exactly the state b1 was found in. | ⚪ |
| e1-12 | **The write probe compares only the target table.** It reads `site_information` before and after each step and never looks at `history`, which is how e1-1 went unnoticed through a run that reported "書き込み一致". A write-parity check has to compare every table the write touches, not the obvious one. | ⚪ |

---

## Earlier waves

| # | Wave | What | Effect |
|---|---|---|---|
| b2-1 | b2 | `rest/user-programs` is not implemented in Go. `b2-organization.spec.ts` skips it explicitly under `go-parity`, so the gap is at least visible in the run output. | 🟡 |
| w0-1 | 0 | `rest/logging`, `rest/logging/stream`, `rest/logging/test` are not started. The library decision is made (`zap`, see [logging-adoption-plan.md](logging-adoption-plan.md)); the port is not scheduled and blocks nothing. | 🟡 |
| w0-2 | 0 | `health` answers a placeholder `{"status":"UP"}` rather than porting Java's `/health/odoo` connectivity check. | 🟡 |

---

## Suggested order

1. **e1-11** — write the e1 spec and get the wave into the ledger. Everything
   else on this list is unguarded until that exists.
2. **e1-12 then e1-1** — teach the probe to compare `history`, watch it go red,
   then port the audit write. In that order, so the test fails first.
3. **e1-2, e1-4, e1-5** — the validation and error envelopes, which share one
   measurement.
4. **e1-3** — the localization write branch.
5. **e1-6** — jasypt, once a JDK is available.
