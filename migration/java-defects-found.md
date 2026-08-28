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
- **`stepProgress.enter` means two different things.** `order/search` hardcodes
  it to `true`; `order/dashboard` computes it from the received date plus
  patient/workflow type.
- **`birthDateForDisplay` is raw in one endpoint and reformatted in another.**
  c1's `patientByLabNumer` emits the stored `entered_birth_date` unchanged;
  `order/search` runs it through `DateUtil.formatStringDate` first.
