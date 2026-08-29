# e1 — Admin config CRUD (scoped migration plan)

Status: **ported, in the gate, 14 e2e tests**
Open items: [open-items.md](open-items.md) — three left, all of them either
unreachable without provoking a database error, or a cache concern rather than
a response one.
Branch: `migration/e1-config-crud` (from `migration-base`).
Companion docs:
[endpoint-migration-order.md](endpoint-migration-order.md) (Wave 6),
[endpoint-migration-taxonomy.md](endpoint-migration-taxonomy.md) (Type E),
[branch-naming.md](branch-naming.md).

e1 is the **write-path proving ground**. Everything before it was a read: the
port answered a GET and a differ compared two documents. Here the port has to
change the database, and the parity question becomes "did both implementations
leave the same row behind", which the read harness cannot ask.

---

## 0. Scope

One controller **pair** in Java serves **nine** configuration domains, chosen by
`request.getServletPath().contains(...)`:

| Domain path | `siteInfoDomainName` | `formName` |
|---|---|---|
| `SiteInformation` | `SiteInformation` | `SiteInformationForm` |
| `PatientConfiguration` | **`PaitientConfiguration`** | `PatientConfigurationForm` |
| `ResultConfiguration` | `ResultConfiguration` | **`resultConfigurationForm`** |
| `ValidationConfiguration` | `validationConfig` | `ValidationConfigurationForm` |
| `WorkplanConfiguration` | `WorkplanConfiguration` | `WorkplanConfigurationForm` |
| `SampleEntryConfig` | `sampleEntryConfig` | **`sampleEntryConfigForm`** |
| `PrintedReportsConfiguration` | `PrintedReportsConfiguration` | `PrintedReportsConfigurationForm` |
| `NonConformityConfiguration` | **`non_conformityConfiguration`** | `NonConformityConfigurationForm` |
| `MenuStatementConfig` | `MenuStatementConfig` | `MenuStatementConfigForm` |

Three of the nine domain names and two of the nine form names break the pattern.
They are the wire contract; the port reproduces them exactly. See §3.

Five handlers, ~52 routes:

| Handler | Routes | Method |
|---|---|---|
| show | `/{domain}` × 9 + `/NextPrevious{domain}` × 9 | GET |
| update | `/{domain}` × 9 | POST |
| cancel | `/Cancel{domain}` × 9 | GET |
| menu | `/{domain}Menu` × 9 | GET |
| delete | `/Delete{domain}` × **7** | GET (with a request body) |

**Delete has seven, not nine.** `DeleteSampleEntryConfig` and
`DeleteValidationConfiguration` do not exist — those two domains can be listed
and edited but not deleted, and nothing in the code says why.

Auth on both controllers: `@PreAuthorize("hasRole('ADMIN')")` at class level.

---

## 1. What the write path needs — and already has

The infrastructure this wave was supposed to build **landed with p0**:

| Requirement | Where it already is |
|---|---|
| Authentication, default-deny | `auth/middleware.Guard.ServeProtected` |
| CSRF, Spring Security 6 compatible | `auth/csrf` — `HttpSessionCsrfTokenRepository` + `XorCsrfTokenRequestAttributeHandler`, masked per read |
| CSRF enforcement on state-changing verbs | same `ServeProtected`, exempting GET/HEAD/OPTIONS/TRACE as `CsrfFilter` does |
| `hasRole('ADMIN')` | `auth/middleware.RequireAdmin` |
| Module-URL authorization | `auth/service.AuthzService.HasPermission` |

`web.Register` applies all of it to any route. So e1 is not "build the write
machinery" — it is porting five handlers onto machinery that is already proven
by the p0 specs.

**The token flow, measured:** `GET /session` returns `csrf`; a write sends it as
`X-CSRF-Token`. Without it Java answers
`{ "status": 403, "message": "CSRF token missing or invalid" }`.

---

## 2. Measured behaviour — the reference

All captured from the live Java server, admin session, one full CRUD cycle
(insert → update → read back → delete) against a probe row that was removed
afterwards. The database was left byte-identical to how it started.

### 2.1 GET `/{domain}` — blank form

```json
{"formName":"SiteInformationForm","formAction":"SiteInformation","formMethod":"POST",
 "cancelAction":"CancelSiteInformation","submitOnCancel":false,"cancelMethod":"POST",
 "paramName":"","description":"","value":"","encrypted":false,"valueType":"text",
 "siteInfoDomainName":"SiteInformation","editable":true,"tag":"","descriptionKey":""}
```

### 2.2 GET `/{domain}?ID=83` — loaded row

`paramName`, `description`, `value`, `valueType` come from the row.
`description` is `getInstruction()`: `instructionKey` → `descriptionKey` →
the raw `description` column, first non-blank wins.

**`tag` disappears.** The blank form carries `"tag":""`; a loaded row whose
`tag` column is NULL drops the key entirely under `Include.NON_NULL`. Same
endpoint, same field, present or absent depending on which branch ran.

### 2.3 POST `/{domain}` — update

- The id comes from the **query string** (`?ID=148`), never the body.
- The body carries `paramName` / `value` / `description` / `valueType` / …

### 2.4 GET `/{domain}Menu`

Root: `formName` (`siteInformationMenuForm`), `formMethod`, `cancelAction:"Home"`,
`submitOnCancel`, `cancelMethod`, `totalRecordCount`, `fromRecordCount`,
`toRecordCount`, `selectedIDs`, `siteInfoDomainName`, `menuList`.

Each `menuList` row: `lastupdated` (epoch ms), `id`, `name`, `description`,
`value`, `encrypted`, `valueType`, `instructionKey`, `domain` (`{id,name,description}`),
`group`, `descriptionKey`.

### 2.5 GET `/Cancel{domain}`

`200`, body is the bare JSON string `"Cancellation successful"`.

### 2.6 GET `/Delete{domain}` — a GET with a body

Requires `{"selectedIDs":["148"],"siteInfoDomainName":"SiteInformation"}`.
`200` with the bare string `"Delete successful"`.

---

## 3. Java defects and traps measured here

Reproduced, not fixed — see the migration policy in
[OpenELIS-Go-Migration-Plan.md](OpenELIS-Go-Migration-Plan.md) §5. Candidates
for [java-possible-bugs.md](java-possible-bugs.md) once pinned by a spec.

| # | What | Evidence |
|---|---|---|
| E1-a | **The id parameter is `ID`, not `id`.** `?id=83` is silently ignored and the endpoint returns the blank "add new" form — an edit screen showing empty fields for a row that exists. | `?id=83` and `?ID=83` measured side by side |
| E1-b | **`?ID=<nonexistent>` → 500.** `siteInformationService.get()` returns null and the next line dereferences it. | `?ID=999999` |
| E1-c | **INSERT silently forces `value_type='text'`** — `setValueType("text")` is hardcoded on the new-row branch — while the response echoes back the `valueType` that was submitted. Sent `boolean`, stored `text`, told `boolean`. | probe row 148 |
| E1-d | **UPDATE writes ONLY `value`.** `paramName` and `description` are read off the form, echoed back in the response, and never persisted. Renaming a row through this endpoint reports success and changes nothing. | POST `paramName:"RENAMED"`, DB kept `e2eWriteProbe` |
| E1-e | **POST returns a different form shape than GET.** `setupFormForRequest` is not called on the POST path, so the response carries the bean defaults: `formName` `siteInformationForm` (lowercase) against GET's `SiteInformationForm`, `cancelAction` `Home` against `CancelSiteInformation`, and **no `formAction` at all**. | both measured |
| E1-f | **`PaitientConfiguration`** — misspelled domain name, on the wire. | GET `/PatientConfiguration` |
| E1-g | **`cancelMethod` says `POST`; `Cancel{domain}` is a GET route.** The form tells the client to submit the cancel with a verb the route does not accept. | mapping vs response |
| E1-h | **Paging is decorative.** `toRecordCount:"20"` while `menuList` carries 30 rows, and `totalRecordCount` is `""`. Same shape as the c2 dashboard paging defect. | `SiteInformationMenu` |
| E1-i | **A fourth error envelope.** `Delete{domain}` validation failure returns a bare JSON **array** of Spring `ObjectError`s (`codes`/`arguments`/`defaultMessage`/`objectName`/`field`/`bindingFailure`/`code`) — not the ProblemDetail, not the per-field `errors` map, not Tomcat's `{timestamp,status,error}`. | delete without `siteInfoDomainName` |
| E1-j | **`isEditable` is a suffix test where equality was meant**: `"Accession number prefix".endsWith(siteInformation.getName())`, so any row whose name is a suffix of that string would be locked too. | source |
| E1-k | **`Delete` exists for 7 of 9 domains.** | mapping list |
| E1-l | **`DeletePatientConfiguration` rejects the domain name its own menu hands out.** `PatientConfigurationMenu` answers `siteInfoDomainName: "PatientConfiguration"`; the delete validator's allow-list contains only the FORM controller's misspelling `PaitientConfiguration`. A client that round-trips the value it was given gets a 400, and the delete succeeds only for a caller that knows to send the typo. | both spellings measured: 400 vs 200 |
| E1-m | **An encrypted row makes every config menu 500.** `hideEncryptedFields` masks the DECRYPTED value, so the decrypt runs on read; a row whose value column is not valid jasypt ciphertext throws, and the resulting `BaseErrors` object is itself unserialisable (`No serializer found for DefaultMessageCodesResolver`). Two failures compound into a 500 with no usable message. | seeded a plaintext encrypted row; SiteInformationMenu went 200 → 500 |
| E1-n | **The CSRF denial body is hand-built, not marshalled.** `{ "status": 403, "message": "..." }` — spaces after the brace and colons, status before message. Not a Java defect; noted because the PORT had it wrong since p0 and nothing caught it: the body is only reachable on a denial, and until e1 no ported route had a state-changing verb to be denied on. | e1 write probe |

---

## 4. Parity strategy — what changes now that writes are involved

A read spec compares two documents. A write spec has to answer a different
question: **did both implementations leave the same row behind?**

The recipe, run independently under `api-mutating` (Java) and `go-parity` (Go):

1. **Create** a row the fixture owns, through the API.
2. **Assert the DATABASE** — not the response. The response is what the
   implementation claims; the row is what it did. E1-c and E1-d are both cases
   where those two disagree, and a spec that trusted the response would have
   found neither.
3. **Read it back** through GET and compare to the row.
4. **Delete** it through the API and assert it is gone.

Every step is its own assertion, so a port that writes nothing still passes
step 1's response check and fails step 2.

**Isolation.** `site_information` is global configuration — `ConfigurationProperties.loadDBValuesIntoConfiguration()`
runs on every write, so a test that edits a real row changes the behaviour of
the whole application, including the other specs in the same run. The spec
therefore creates and destroys **its own** row and never touches a shipped one.
It must clean up even when it fails.

---

## 4.1 The encrypted-row branch — closed

Java masks the **decrypted** value, so reproducing `hideEncryptedFields` meant
porting jasypt. The parameters were read out of the library with a JDK rather
than guessed at:

| | |
|---|---|
| algorithm | PBEWithHMACSHA512AndAES_256 |
| key | PBKDF2-HMAC-SHA512, 1000 iterations, 256-bit |
| cipher | AES-256-CBC, PKCS#5 |
| wire | Base64( **salt ‖ iv ‖ ciphertext** ), 16 bytes each |
| password | `kspass`, from `volume/properties/common.properties` |

Two things cost time and are worth writing down. The layout reads like
iv-then-salt in the jasypt source — `StandardPBEByteEncryptor` prepends the
salt and then prepends the IV to that result — and it is not; decrypting
jasypt's own output settles it. And the password is **not** the `dev` fallback
in the `@Value` expression: this deployment overrides it, and a value
encrypted under one password is unreadable under the other, which is what made
the first search fail against what turned out to be the right algorithm.

**Still no fixture seeds an encrypted row.** The e2e spec creates one through
the API and deletes it, so the ciphertext is always the application's own. A
row whose column holds plaintext makes Java 500 on every config menu (E1-m).

**Running the gate:** the Go service needs `OE_ENCRYPTION_PASSWORD` set to the
same key, or the encrypted-row test fails for a configuration reason rather
than a porting one.

---

## 5. Order of work

1. Port `show` + `menu` (reads) for all nine domains, with the dispatch table.
2. Port `cancel` — trivial, and it proves the route/verb mapping.
3. Port `update` (POST) with E1-c/E1-d/E1-e reproduced.
4. Port `delete`, with its 7-domain list and the array error envelope.
5. Write the mutating parity spec per §4; promote to `go-parity` only once
   green.
