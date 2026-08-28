# Java defects found during the Go migration

Migration policy is **pin Java, do not fix it**: the Go port reproduces the
behaviour Java actually has, and defects are raised with the maintainers
separately rather than quietly corrected in the port. A port that "fixes" a bug
makes Go and Java disagree on a live endpoint, which is the one thing this
migration must not do.

That policy has a reproduce half and a raise half. The reproduce half has been
happening all along, in code comments and spec comments; this file is the raise
half, so the list is in one place when someone takes it to the Java side.

Every entry below was **measured on the running Java server**, not inferred from
reading the source, and each is pinned by a named test in
`migration/openelis-api-e2e/`.

---

> **Looking for the short list?** [java-possible-bugs.md](java-possible-bugs.md)
> is the consolidated triage index across every wave — severity, confidence and
> whether a test pins it — including the items that never got a full write-up
> here. This file is the detail.

## 1. `GET rest/sample/unassigned-by-accession/{accessionNumber}` — always 500

`SampleDAOImpl.getUnassignedSampleByAccessionNumber` runs HQL referencing
`r.canceled`, which is not a property of
`org.openelisglobal.referral.valueholder.Referral` (the column is `canceled`,
the field is not mapped under that name). Hibernate throws while **parsing**,
before touching data, so no input can succeed:

```
could not resolve property: canceled of:
org.openelisglobal.referral.valueholder.Referral
```

The controller catches `Exception` and answers 500, which makes its own
`return notFound()` branch unreachable dead code.

- **Observable:** 500, empty body, for a valid accession, a second valid one and
  a nonexistent one alike.
- **Pinned by:** `c2-sample-order-reads.spec.ts` — "unassigned-by-accession:
  PERMANENTLY BROKEN in Java — always 500".

---

## 2. `GET rest/unassigned-sample/items` and `/items/search` — 500 as soon as any row matches

`UnassignedSampleItemServiceImpl.buildSampleItemDTOs` calls
`referralDAO.getReferralsBySampleItemId(Integer)` once per result row, but
`SampleItem.id` is mapped as a `String`, so Hibernate refuses the binding:

```
IllegalArgumentException: Parameter value [10111] did not match
expected type [java.lang.String (n/a)]
```

The service catches it and returns an empty list, but the exception has already
marked the `@Transactional(readOnly = true)` transaction rollback-only, so the
commit at the AOP boundary throws `UnexpectedRollbackException` and the
**controller's** catch answers `ResponseEntity.status(500).build()`.

The mismatch is structural — an `Integer` argument against a `String`-mapped id
— so it is independent of the values involved and fires for **any** non-empty
result. A query that matches nothing never enters the loop and still answers
`200 []`.

- **Why it went unnoticed:** `clinlims.referral` is empty in the dev/demo
  dataset, so both endpoints answered `200 []` and looked healthy. Seeding the
  table is what surfaced it.
- **Likely fix (for the Java side):** take the id as `String` and drop the
  `Integer.parseInt` at the two call sites, matching the mapping.
- **Pinned by:** `c2-sample-order-reads.spec.ts` — "unassigned-sample/items: 500
  once ANY row matches".

---

## 3. `GET rest/order/attachments/{id}/download|view` — a missing id is 500, a deleted one is 404

`OrderAttachmentRestController.serveAttachment` guards with
`attachment == null || TRUE.equals(attachment.getIsDeleted())` and answers 404.
But `OrderAttachmentServiceImpl.get(id)` **throws** for an id with no row rather
than returning null, so the `== null` half of that guard is unreachable and a
missing id escapes as a 500.

- **Observable:** soft-deleted id → 404 (empty body); nonexistent id → 500.
- **Pinned by:** `c2-sample-order-reads.spec.ts` — "order/attachments/{id}/
  download vs /view".

---

## 4. `GET rest/order/dashboard` — paging that does not page

`SampleDAOImpl.getPageOfSamples`:

```java
int endingRecNo = startingRecNo + (page.defaultPageSize + 1);
query.setFirstResult(startingRecNo - 1);
query.setMaxResults(endingRecNo - 1);
```

so `maxResults = startingRecNo + defaultPageSize`. Two consequences:

- The **request's** `pageSize` never bounds the result. It only shifts the
  offset, so asking for 1 row still returns `defaultPageSize + 1` of them while
  the echoed `pageSize` says 1.
- The limit **grows with the offset**, so later pages return more rows than
  earlier ones.

Alongside it, in `OrderSearchRestController.getDashboard`:

- `externalCount` is hardcoded `0` and never computed.
- `includeExternal` is accepted as a `@RequestParam` and never read.
- `totalCount` is `ordersList.size()` — the size of the current page after
  filtering. Java's own comment says "Simplified, should be total count".

- **Pinned by:** `c2-sample-order-reads.spec.ts` — "order/dashboard: pageSize is
  echoed but IGNORED; externalCount is hardcoded 0".

---

## 5. Method-security denials surface as 500, not 403

A `@PreAuthorize` denial produces an `AccessDeniedException` that never reaches
the configured `accessDeniedHandler`, so the client sees
`{"timestamp":…,"status":500,"error":"Internal Server Error"}` where a 403 is
intended.

- **Pinned by:** `p0-authz.spec.ts`.

---

## 6. `programId` can contradict the sample's own `program_sample` row

`buildSampleOrderItems` resolves the program id through
`ProgramSampleDAOImpl.getProgrammeSampleBySample`, which picks the JPA entity
**class** from the program **name**: a name containing `pathology`,
`cytology` or `immunohistochemistry` selects `PathologySample`,
`CytologySample` or `ImmunohistochemistrySample`. Those are
`@Inheritance(TABLE_PER_CLASS)` and live in their own tables, so the sample's
actual `program_sample` row is invisible to the query, which returns null. The
controller then falls back to scanning `programService.getAll()` for a program
whose name equals the observation value.

The result is that for such a sample `programId` is the id of the **named**
program, and the `program_sample` row the sample really points at is ignored.
Seeded and pinned as `E2E-FULL-02`: `program_sample.program_id = 2`
(`Routine Testing`), observation `program = "Cytology"`, response
`programId = "5"`.

- **Pinned by:** `c2-sample-order-reads.spec.ts`, "program resolution takes
  three different paths".

---

## 7. A type-less sample item NPEs two result endpoints (c3)

`AnalysisServiceImpl.getTestDisplayName` calls
`sampleItem.getTypeOfSampleId().equals(...)` with no null check. Java's OWN
unassigned-sample HQL LEFT JOINs `type_of_sample` and COALESCEs the
description — it is written to tolerate a NULL `typeosamp_id` — so the two
code paths disagree about whether that state is legal.

Reachable from at least two endpoints, which is how it was found:
`rest/LogbookResults?selectedTest=N` and
`rest/WorkPlanByTestSection?test_section_id=N` both 500 as soon as one
matching analysis sits on such an item. A test or section with none answers
200 with rows.

- **Pinned by:** `c3-result-reads.spec.ts`, "a type-less sample item 500s
  WorkPlanByTestSection and LogbookResults alike", with the 200-plus-rows
  inversion beside it.

---

## 8. `AccessionValidation` returns a DIFFERENT accession's results by default

`showAccessionValidationRange` takes
`@RequestParam(defaultValue = "true") Boolean doRange`, and the two branches
are not variations on one search:

| `doRange` | behaviour |
|---|---|
| `true` (default) | `getResultValidationList(status, section, accessionNumber, date)` — a RANGE search |
| `false` | `getSample(accessionNumber)`, then that sample's analyses; empty when no such sample |

Measured: asking for `E2E-ATT-01` — an order with no analyses of its own —
returns one row whose own `accessionNumber` is `E2E-RES-01`, while the form
still echoes `E2E-ATT-01`. A clinical screen showing another order's result
under the requested accession is worth raising on its own; for the port it
means the obvious exact-match reading is wrong and looks right on every other
input.

- **Pinned by:** `c3-result-reads.spec.ts`, "AccessionValidation: only
  TechnicalAcceptance analyses are up for validation".

---

## 9. `ReferredOutTests?searchType=TEST_AND_DATES` 500s instead of validating

`SearchType` is `TEST_AND_DATES` / `LAB_NUMBER` / `PATIENT`. A value outside
the enum is a clean 400 from Spring's converter, but the in-enum
`TEST_AND_DATES` with no dates supplied throws instead of reporting a
validation error.

- **Pinned by:** `c3-result-reads.spec.ts`, "ReferredOutTests: searchType
  drives the search, and one of its values 500s".

---

## 10. `QAService.nonconformingByDepricatedStatus` NPEs on a NULL sample status

`Sample.getStatusId()` is dereferenced with no null check, so any endpoint
reaching that code 500s for a sample whose `status_id` is NULL — found via
`rest/WorkPlanByTest`, which returned 500 for every `test_id` that had
analyses. Every stock sample carries a status, so this only fires on data
written outside the app.

---

## 11. `UnitOfMeasure.getLocalization()` ships a placeholder as the French name

`unit_of_measure` has no `name_localization_id` column, so the getter builds a
`Localization` on the fly — and sets the French value to the literal string
`"French"`:

```java
_localization.setEnglish(this.getDefaultLocalizedName());
_localization.setFrench("French");
```

That literal reaches the client inside every `rest/accession-results`
response, in `values.fr.value`, `french`, `valuesAsMap.fr` and
`localesAndValuesOfLocalesWithValues`. A French deployment shows "French" as
the unit of measure.

- **Pinned by:** the c3 parity gate, byte-for-byte.

---

## 12. `AccessionValidation` shows no reference range where the logbook shows one

`ResultsValidationUtility.createTestResultItem` resolves the reference range
with `getResultLimitForTestAndPatient(test, currentPatient)`, and on that path
the patient carries neither a birth date nor a gender — so the service takes
its `patient == null` branch and returns `defaultResultLimit`, which only ever
matches a row whose gender is BLANK and whose age limits are the defaults
`0..Infinity`.

A test whose `result_limits` rows are all age-banded therefore has NO default
row, gets no limit, and its `normalRange` comes out empty — while
`rest/LogbookResults` resolves the SAME analysis through
`getResultLimitForAnalysis`, picks the age-appropriate band, and renders it.

Measured on one analysis, both endpoints, same request:

| endpoint | resultLimitId | normalRange |
|---|---|---|
| `LogbookResults` | 11 | `4.00 - 10.00` |
| `AccessionValidation` | — | `""` |

So the validation screen — the one where a biologist decides whether a result
is acceptable — is the screen that shows no reference range for any
age-banded test.

- **Pinned by:** `c3-result-reads.spec.ts`, "ONE result_limits band per
  analysis, chosen by the patient's age".

---

## 13. `lowerCritical` answers with the STRING `"Infinity"` from a field declared `double`

`ResultLimit` initialises its bounds in the valueholder, and one of them is
initialised to the wrong infinity:

```java
private double highCritical = Double.POSITIVE_INFINITY;
private double lowCritical  = Double.POSITIVE_INFINITY;   // <- every other LOW bound is NEGATIVE
private double lowNormal    = Double.NEGATIVE_INFINITY;
private double lowValid     = Double.NEGATIVE_INFINITY;
private double lowReportingRange = Double.NEGATIVE_INFINITY;
```

That initialiser is normally invisible, because a limit loaded from the table
overwrites every field. It becomes reachable through
`ResultLimitServiceImpl.defaultResultLimit`, which does **not** return null
when a test has limits but none of them is the default band:

```java
for (ResultLimit limit : resultLimits) {
    if (isBlankOrNull(limit.getGender()) && limit.ageLimitsAreDefault()) return limit;
}
return new ResultLimit();          // <- a SYNTHESIZED limit, all fields at their initialisers
```

`setResultLimitDependencies` then folds each bound against **its own**
sentinel:

```java
testItem.setLowerCritical( resultLimit.getLowCritical()  == Double.NEGATIVE_INFINITY ? 0 : ... );
testItem.setHigherCritical(resultLimit.getHighCritical() == Double.POSITIVE_INFINITY ? 0 : ... );
```

`highCritical` matches its sentinel and folds to 0. `lowCritical` holds
`POSITIVE_INFINITY`, does not match `NEGATIVE_INFINITY`, and survives to
Jackson — which has no JSON number for an infinity and, with
`QUOTE_NON_NUMERIC_NUMBERS` on by default, writes it as a **string**.

Measured on one `AccessionValidation` response, three rows of one order:

| analysis | default band resolved? | `lowerCritical` | `higherCritical` |
|---|---|---|---|
| 173 | yes | `0.0` | `0.0` |
| 176 | yes | `0.0` | `0.0` |
| 175 | **no** (age-banded only) | `"Infinity"` | `0.0` |

So a field the DTO declares as `double` changes JSON **type** depending on
which rows exist in `result_limits`, and the two bounds of the same row
disagree about what "unset" looks like. Any consumer that parses
`lowerCritical` as a number fails on exactly the age-banded tests — the same
rows that already carry [defect 12](#12-accessionvalidation-shows-no-reference-range-where-the-logbook-shows-one)'s
empty `normalRange`.

A test with **no** `result_limits` rows at all is a third case:
`getResultLimitForTestAndPatient` returns null, the dependencies are never
set, and both bounds stay at the bean default `0.0`.

- **Pinned by:** `c3-result-reads.spec.ts`, "a test with bands but NO default
  band answers lowerCritical \"Infinity\"".


## Not defects — deliberate asymmetries worth knowing

- **`result` is TWO different objects under one key.** `LogbookResults` nests a
  five-key reference (`isActive`, `id`, `significantDigits`, `grouping`,
  `fhirUuidAsString`); `accession-results` nests the FULLY serialised Hibernate
  entity — analysis, sample item, sample, type of sample, test section, test,
  unit of measure, panel and all their `Localization` objects, roughly 300
  fields. Nothing in the code selects between them: it is which associations
  happen to be initialised when Jackson reaches the object.
- **`LogbookResults` leaves `searchFinished` FALSE on the `selectedTest` path**
  while still returning rows. Only the `labNumber` path sets it.
- **`testDate` is the clock, not a stored column.** It changes between two
  calls seconds apart. The analysis's `entry_date` looks like the obvious
  source and is not it.
- **`accession-results` flattens the patient onto the ROOT and blanks it in the
  rows** — `patientName` and `patientInfo` come out as a single SPACE,
  `nationalId` as `""`, and `patientId` is absent. `LogbookResults` does the
  opposite, repeating the patient on every row and omitting it from the root.
- **Two patient-name formats and two birth-date COLUMNS in one wave.**
  `LogbookResults` renders "Last, First" and formats the parsed `birth_date`;
  `AccessionValidation` renders "Last First" and emits the raw
  `entered_birth_date` text. On the seeded patient those disagree —
  `01/03/1991` against `01/15/1990`.
- **`significantDigits` has two sources.** `LogbookResults` reads
  `test_result.significant_digits`; `AccessionValidation` reads the result
  row's own. Same field name, different numbers for the same result.
- **`result_limits` is per test AND per AGE BAND** (`min_age`/`max_age` in
  DAYS) and optionally per gender. Joining on `test_id` alone multiplies every
  analysis by its band count — one order went from 4 rows to 9.
- **`normalRange` and `units` carry DIFFERENT ranges.** `normalRange` is the
  test's `result_limits` row formatted to the `test_result` significant digits
  (`7 - 40`); the range inside `units` is the RESULT row's own
  `min_normal`/`max_normal` formatted to the RESULT's significant digits
  (`UI/L ( 1.00-9.00 )`).
- **`upperAbnormalRange`/`lowerAbnormalRange` are `result_limits`' VALID
  range** (`high_valid`/`low_valid`) under a different name.
- **`reportable` is the TEST's column, not the analysis's.** Every seeded
  analysis is `Y` while every test is `N`.
- **The logbook groups by SAMPLE ITEM; the validation list groups by
  ACCESSION** — with the same counter, which starts at ONE and increments on
  the first row, so the first group is 2.

- **`WorkPlanByPanel` expands the panel to its TESTS; it does not read
  `analysis.panel_id`.** `getWorkplanByPanel` reads `panel_item` for the panel
  and calls `getAllAnalysisByTestAndStatus` once per member test,
  concatenating. An analysis on a member test appears with its own `panel_id`
  NULL, and the response is strictly larger than the set of analyses carrying
  that `panel_id`.
- **`getAllAnalysisByTestSectionAndStatus`'s third parameter is dead.**
  `sortedByDateAndAccession` guards a block whose entire body is commented
  out, so passing `true` applies no ordering. The callers pass `true`.
- **`WorkPlanByTestSection` filters `analysis.test_sect_id`, not
  `test.test_section_id`.** The column on `analysis` is a denormalised copy
  that `AnalysisServiceImpl` fills at creation time. Joining through `test`
  agrees whenever the two match and diverges the moment they do not.

These look like bugs, are not, and are pinned so a port does not "tidy" them:

- **`voided` filtering is inconsistent by design between two shipment queries.**
  `getUnassignedReferrals` (the dashboard) has no `si.voided` predicate, while
  `getUnassignedReferralsGroupedBySampleItem` (`/items`) does — so a referral on
  a voided sample item appears on the dashboard and not in `/items`.
- **A NULL `referral.status` is excluded everywhere.** The HQL is
  `r.status != 'CANCELED'`, and `NULL != 'CANCELED'` is UNKNOWN, not TRUE. A
  port written with Go's `status != "CANCELED"` would include those rows.
- **`file_type` has two null policies.** The attachment list renders a NULL as
  `""`; `serveAttachment` turns the same NULL into `application/octet-stream`.
- **`sampleOrderItems` is three different objects, and the two that carry a
  sample disagree on where every shared field comes from.** `order/search`
  builds a `HashMap` in the controller; `SampleEdit` builds a
  `SampleOrderItem` bean through `SampleOrderService`. Same key names,
  different sources:

  | key | `order/search` | `SampleEdit` |
  |---|---|---|
  | `provider*` | `sample_human.provider_id` → provider → person | `sample_requester` PERSON row → person, then `getProviderByPerson` |
  | `referringSiteCode` | `organization.short_name` | `organization.code` |
  | `referringSiteDepartmentName` | emitted | never emitted |
  | `programId` | subclass lookup **with** a name fallback | subclass lookup, **no** fallback |
  | `receivedTime` | `sample.received_date` | clock, then overwritten by `sample.received_date` |

  On this deployment the referring clinic has `short_name = '279'` and a NULL
  `code`, so `order/search` emits `referringSiteCode` and `SampleEdit` does
  not — from the same organization. Sharing one builder across the two gets
  at least one of them wrong.

- **`getPendingAnalysisForTestProvider`'s array order is not reproducible, on
  either implementation.** The HQL ends `order by
  a.sampleItem.sample.accessionNumber` and nothing else, so analyses sharing an
  accession number are tied and Postgres decides from the plan. Measured on
  this dataset: the first calls over a fresh connection return a tie group in
  one order and later calls return it in another — Java included. Membership
  and the accession ordering are the contract; the tie order is not, and a port
  must NOT add a secondary sort key to stabilise it (no available column
  reproduces the observed order anyway).

- **`ORDER BY accession_number` sorts under the DATABASE collation, which
  ignores punctuation.** `E2E001` precedes `E2E-EDIT-01` there, while a
  code-point sort puts the hyphen first. Any ordering a port performs in Go
  rather than in SQL gets this backwards.

- **The AnalysisStatus enum constants are not the stored status names.**
  `StatusService.addToAnalysisMap` matches `status_of_sample.name` literally,
  so `NotStarted` is the row named `Not Tested` and `BiologistRejected` is
  `Biologist Rejection`. The endpoint also files `TechnicalAcceptance` — an
  acceptance status — under a key called `notValidated`.

- **Two counters behind `order/dashboard`'s `stepProgress` are voided-filtered
  and it is not obvious.** `collect` and `label` are computed over
  `sampleItemService.getSampleItemsBySampleId`, whose criteria map is
  `{sample.id, voided:false}`. A port that aggregates `sample_item` directly
  counts voided rows, and `E2E-VOIDED-01` — whose only item carrying analyses
  is the voided one — flips `collect` from false to true.

- **`requester_type = 'provider'` means the requester is a PERSON.**
  `SampleServiceImpl.initializeGlobalVariables` sets
  `PERSON_REQUESTER_TYPE_ID` from the `requester_type` row named `provider`,
  and `getPersonRequester` then reads that row's `requester_id` as a
  `person.id`. A port that treats it as a `provider.id` joins the wrong table
  and usually finds nothing.

- **Two readers over `observation_history`, chosen per field.**
  `getRawValueForSample` returns `value` as stored. `getValueForSample`
  returns it only when `value_type` is `L`; otherwise it treats the value as a
  DICTIONARY ID and renders that row's localized name. `SampleEdit` uses the
  raw reader for `paymentOptionSelection`, `testLocationCode` and `program`
  and the resolving one for the other five. Using one reader throughout is
  right for whichever half it happens to match, and echoes a numeric id to the
  user for the rest.

- **`samples[]` is unordered, despite a DAO method that orders it.**
  `SampleItemDAOImpl.getSampleItemsBySampleId` ends `order by
  sampleItem.sortOrder`, but `SampleItemServiceImpl.getSampleItemsBySampleId`
  — the method the controller actually calls — ignores it and builds a
  criteria map for `getAllMatching`, which has no ordering. The array order is
  therefore whatever Postgres scans, and on the stock dataset E2E001 comes
  back with sample item 10002 before 10001: higher id, later `sortOrder`.
  Reading the DAO instead of the service leads a port straight to
  `ORDER BY sort_order`, which reverses it.
- **`stepProgress.enter` means two different things.** `order/search` hardcodes
  it to `true`; `order/dashboard` computes it from the received date plus
  patient/workflow type.
- **`birthDateForDisplay` is raw in one endpoint and reformatted in another.**
  c1's `patientByLabNumer` emits the stored `entered_birth_date` unchanged;
  `order/search` runs it through `DateUtil.formatStringDate` first.
- **`quantity` is a number, the string `""`, or absent — same column, one
  response.** `sample_item.quantity` is a `Double` put into a
  `Map<String, Object>` twice by `order/search`: the outer item does
  `q != null ? q : ""`, so a NULL becomes a JSON *string*; the nested
  `sampleXML` puts it raw, so a NULL is dropped by `Include.NON_NULL`. A typed
  client reading `quantity` as a number breaks on any sample item with no
  quantity. Invisible on the stock dataset — every row there has one — and only
  surfaced once `order-search-full-e2e.sql` seeded a NULL.
