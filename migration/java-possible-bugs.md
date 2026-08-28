# Java possible bugs — consolidated

Everything that looked like a **defect in the Java application** while porting,
across every wave so far (p0, b2, c1, c2, c3), gathered in one place so it can
be raised with the maintainers.

**Nothing here has been fixed.** The migration policy is to reproduce Java's
behaviour, not correct it, so the Go port answers identically on every item
below. See `OpenELIS-Go-Migration-Plan.md` §5.

## How to read this

| Column | Meaning |
|---|---|
| **Confidence** | 🔴 certain (root cause read in the source and reproduced) · 🟡 likely (behaviour measured, cause inferred) · ⚪ cosmetic |
| **Pinned** | ✅ an e2e test asserts the broken behaviour · — no test yet |

This file is a **triage list**. `java-defects-found.md` holds the full write-up
for the items that have one, including the exact reproduction and the test that
pins it; this file is the index, plus the items that never got their own entry.

**Not re-investigated.** These come from what surfaced during the porting work.
Where the cause is inferred rather than read, the Confidence column says so.

---

## A. Endpoints that fail on valid input

| # | Endpoint | Symptom | Confidence | Pinned | Wave |
|---|---|---|---|---|---|
| A1 | `GET rest/sample/unassigned-by-accession/{n}` | **Always 500**, for every input | 🔴 | ✅ | c2 |
| A2 | `GET rest/unassigned-sample/items` and `/items/search` | 500 as soon as **any row matches**; an empty result is the only clean answer | 🔴 | ✅ | c2 |
| A3 | `GET rest/GenericSampleOrder` | 500 for **every accession that exists**; a nonexistent one is the only input that answers cleanly (404) | 🔴 | ✅ | c2 |
| A4 | `GET rest/WorkPlanByTestSection`, `GET rest/LogbookResults?selectedTest=` | 500 when any matching analysis sits on a sample item with a NULL `typeosamp_id` | 🔴 | ✅ | c3 |
| A5 | `GET rest/ReferredOutTests?searchType=TEST_AND_DATES` | 500 with no dates supplied, where a value *outside* the enum is a clean 400 | 🔴 | ✅ | c3 |
| A6 | any endpoint reaching `QAService.nonconformingByDepricatedStatus` | NPE on a NULL `sample.status_id` | 🔴 | — | c3 |

**A2 / A3 share one root cause**: a numeric id bound to a String-mapped
property. `Parameter value [10002] did not match expected type
[java.lang.String]`. The exception marks the transaction rollback-only, the
commit at the `@Transactional` boundary throws, and the handler wraps it into a
500.

**A4** is the sharper one: Java's *own* unassigned-sample HQL LEFT JOINs
`type_of_sample` and COALESCEs the description — it is written to tolerate a
NULL type — while `AnalysisServiceImpl.getTestDisplayName` dereferences
`sampleItem.getTypeOfSampleId()` with no null check. **Two code paths in the
same application disagree about whether that state is legal.**

---

## B. Wrong data returned (no error raised)

| # | What | Why it matters | Confidence | Pinned | Wave |
|---|---|---|---|---|---|
| B1 | `AccessionValidation` returns **another accession's results** | `doRange` defaults to `true`, which is a RANGE search: `accession_number >= ? AND length(accession_number) = length(?)`. Asking for `E2E-ATT-01` returns a row belonging to `E2E-RES-01` **while the form echoes the requested number**. A clinical screen showing another order's result under the number you asked for. | 🔴 | ✅ | c3 |
| B2 | `programId` can contradict the sample's own `program_sample` row | `getProgrammeSampleBySample` picks a JPA entity **class** from the program NAME, and those classes are `TABLE_PER_CLASS`. A name containing `cytology`/`pathology`/`immunohistochemistry` queries a different TABLE, finds nothing, and falls back to a name lookup — ignoring the row the sample actually points at. | 🔴 | ✅ | c2 |
| B3 | `patient/merge/details` counted **voided** sample items | `totalSamples` and `totalResults` included voided rows. | 🟡 | ✅ | c1 |
| B4 | `order/dashboard`'s `stepProgress` counted **voided** sample items | `collect` flipped to true on an order whose only item carrying analyses is the voided one. | 🔴 | ✅ | c2 |
| B5 | `order/dashboard` paging does not page | `maxResults = startingRecNo + defaultPageSize`, so the requested `pageSize` never bounds the result and the limit **grows with the offset** — later pages return more rows than earlier ones. The echoed `pageSize` still reports what was asked for. | 🔴 | ✅ | c2 |
| B6 | `order/dashboard` hardcoded counters | `externalCount` is always 0 and `includeExternal` is inert. | 🔴 | ✅ | c2 |
| B7 | `UnitOfMeasure.getLocalization()` ships a placeholder | The table has no localization column, so the getter builds one in code and sets the French value to the **literal string `"French"`**. It reaches the client in four places of every `accession-results` response. A French deployment shows "French" as the unit of measure. | 🔴 | ✅ | c3 |
| B9 | `lowerCritical` is emitted as the JSON **string** `"Infinity"` | `ResultLimit` initialises `lowCritical` to **POSITIVE**_INFINITY where every other low bound is NEGATIVE. When a test has age bands but no default band, `defaultResultLimit` returns `new ResultLimit()` rather than null, the fold guard tests the wrong sentinel and misses, and Jackson writes the infinity as a string. A field declared `double` changes JSON type per row, and `higherCritical` on the SAME row folds to 0. | 🔴 | ✅ | c3 |
| B8 | `AccessionValidation` shows NO reference range for an age-banded test | It resolves the limit through the `patient == null` branch, which only matches a blank-gender `0..Infinity` row. `LogbookResults` resolves the same analysis to the age-appropriate band and renders it — so the screen where a biologist accepts or rejects a result is the one without a reference range. | 🔴 | ✅ | c3 |

---

## C. Status-code and error-shape problems

| # | What | Confidence | Pinned | Wave |
|---|---|---|---|---|
| C1 | Method-security denials surface as **500, not 403** — the `AccessDeniedException` never reaches the configured `accessDeniedHandler` | 🔴 | ✅ | p0 |
| C2 | `merge/details/{non-numeric}` → **500** where a 400 belongs (the path variable is a String, so Spring cannot reject it at binding) | 🔴 | ✅ | c1 |
| C3 | `order/attachments/{id}/download\|view` — a **missing** id is 500 while a **soft-deleted** one is 404 | 🔴 | ✅ | c2 |
| C4 | Provider search with a huge `page` overflows `(page-1)*pageSize` into a **negative SQL OFFSET**, which Postgres rejects → 500. Currently masked because the 32-bit binding rejects the input first. | 🟡 | ✅ | b2 |
| C5 | Binding failures return **unresolved message keys** — `problemDetail.org.springframework.beans.TypeMismatchException` rather than a sentence, because no `MessageSource` is wired for `ProblemDetail` | 🔴 | ✅ | c2 |
| C6 | Three different error envelopes for the same class of failure: RFC 7807 `ProblemDetail` (`@RequestParam` enum), a per-field `errors` map (`@Valid` form), and Tomcat's `{timestamp,status,error}` (unhandled exception) | ⚪ | ✅ | c2/c3 |

On **C5**, note the ProblemDetail is also internally inconsistent: `type` and
`title` name `MethodArgumentTypeMismatchException` while `detail` names
`org.springframework.beans.TypeMismatchException`.

---

## D. Non-deterministic output

| # | What | Confidence | Pinned | Wave |
|---|---|---|---|---|
| D1 | `getPendingAnalysisForTestProvider`'s array order is **not reproducible, on Java itself** — the HQL sorts on `accession_number` alone and tied rows come back in one order over a fresh connection and another later | 🔴 | ✅ | c2 |
| D2 | `samples[]` is unordered despite a DAO method that orders it — `SampleItemDAOImpl.getSampleItemsBySampleId` ends `order by sortOrder`, but the **service** of the same name (the one the controller calls) ignores it and runs an unordered `getAllMatching` | 🔴 | ✅ | c2 |
| D3 | `getUnassignedReferrals` is likewise unordered, and the shipment dashboard renders the array in sequence | 🔴 | ✅ | c2 |

D2 and D3 are the same shape: the array order is observable in the UI and
nothing guarantees it.

---

## E. Dead code and inert parameters

| # | What | Confidence | Wave |
|---|---|---|---|
| E1 | `getAllAnalysisByTestSectionAndStatus`'s third parameter, `sortedByDateAndAccession`, guards a block whose **entire body is commented out**. Every caller passes `true`. | 🔴 | c3 |
| E2 | `LogbookResults` leaves `searchFinished` **false** on the `selectedTest` path while returning rows — only the `labNumber` path sets it. The screen shows results while reporting that no search completed. | 🔴 | c3 |
| E3 | `stepProgress.enter` means two different things: `order/search` hardcodes it to `true`, `order/dashboard` computes it | 🔴 | c2 |

---

## F. Misspelled public identifiers

These are **in the wire contract** — renaming them is a breaking change, which
is presumably why they survived. Listed so nobody "fixes" them in the port.

| Identifier | Should be | Where |
|---|---|---|
| `rest/patientByLabNumer` | `...LabNumber` | public URL |
| `refferingSiteId` | `referring...` | public query param |
| `acessionFormat` | `accession...` | `site_information` key |
| `recievedDate` | `received...` | `LogbookResults` param, `ReportExternalImport` field |
| `oderpriority` | `orderPriority` | HQL parameter name |
| `nonconformingByDepricatedStatus` | `Deprecated` | method name |

---

## G. Same field, two meanings — traps rather than bugs

Not defects, but every one of them cost a diff cycle, and a port that
"harmonises" any of them is wrong. Full detail in `java-defects-found.md`.

- **`sampleOrderItems` is three different objects** across `order/search`,
  `SamplePatientEntry` and `SampleEdit`, and the two that carry a sample source
  the provider from **different tables** (`sample_human.provider_id` vs a
  PERSON-typed `sample_requester` row) and `referringSiteCode` from **different
  columns** (`short_name` vs `code`).
- **`result` is two shapes under one key** — five keys in `LogbookResults`, the
  entire serialised Hibernate graph (~300 fields) in `accession-results`.
  Nothing selects between them; it is which associations Jackson finds
  initialised.
- **`quantity` is a number, the string `""`, or absent** — one column, two put
  sites, three outputs.
- **Two patient-name formats and two birth-date COLUMNS** in one wave:
  `"Last, First"` off `birth_date` vs `"Last First"` off the raw
  `entered_birth_date`. On the seeded patient those disagree.
- **`significantDigits` has two sources**, `test_result` vs the result row.
- **`requester_type = 'provider'` means the requester is a PERSON.**
- **AnalysisStatus enum constants are not the stored names** — `NotStarted` is
  the row named `Not Tested`, `BiologistRejected` is `Biologist Rejection`.
- **`ORDER BY accession_number` sorts under the DB collation** (punctuation
  ignored, `E2E001` first) while Java's in-memory sort is `String.compareTo`
  (byte order, `E2E001` last). Both orderings are load-bearing in the same
  response.
- **A NULL `referral.status` is excluded everywhere** — `r.status != 'CANCELED'`
  is UNKNOWN for NULL, not TRUE.
- **`notes` is not the note text.** `getNotesAsString(analysis, true, true,
  "<br/>", false)` prefixes each note with its TYPE label and its timestamp —
  `Internal 28/08/2026 20:14 : <text>` — and joins multiple notes with
  `<br/>`. An unknown note type contributes an empty label and the line still
  opens with its space. A port that returns the column is wrong, and a test
  that asserts `toContain(text)` cannot tell the two apart.
- **`file_type` has two null policies** — `""` in the list,
  `application/octet-stream` on download.

---

## Suggested triage order

1. **B1** — a clinical screen showing another order's result under the requested
   accession number is the only item here with a patient-safety edge.
2. **A1–A5** — endpoints that fail on valid input. A2/A3 share one fix.
3. **B4, B5, B3** — silently wrong counts and paging.
4. **B7** — a French deployment displaying "French" as a unit of measure.
5. Everything else.

---

## Where the detail lives

- `java-defects-found.md` — full write-up per defect: reproduction, root cause,
  the test that pins it, and the deliberate asymmetries a port must not tidy.
- `openelis-api-e2e/tests/readonly/*.spec.ts` — the assertions themselves. Every
  pinned item has a test whose name says it is pinning Java behaviour.
