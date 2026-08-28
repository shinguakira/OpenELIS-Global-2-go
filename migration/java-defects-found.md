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

## Not defects — deliberate asymmetries worth knowing

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
