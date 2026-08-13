# b2 — Organization + Provider Migration

Status: **implemented, statically verified (go build/vet/gofmt clean), and
live-verified twice over** — first by direct curl side-by-side capture against
the real Java webapp and this Go port (same live Postgres instance), then by
turning that into committed Playwright spec files that assert the same
contract and actually executing them against both servers
(`test:readonly` / `test:go-parity`; see § 5). All 9 endpoints now have a
spec file wired into the `go-parity` project — **13/13 passing** (1 test,
`user-programs`, is correctly skipped under `go-parity`: it's deferred, not
implemented in Go — see § 2). Branch: `migration/b2-org-provider` (forked
from `migration-base`, per [branch-naming.md](branch-naming.md); the
Playwright spec files also live on this same branch — not a separate
`e2e/`-prefixed branch off `develop` as `branch-naming.md`'s stated policy
would otherwise call for, overridden by explicit user instruction this
session). Taxonomy Type B, Wave 2 per
[endpoint-migration-order.md](endpoint-migration-order.md).

**This live pass caught four things the original source-only implementation
had missed or mis-predicted** — see § 3 items 6, 7, 8, 9. Two are Go-side
implementation bugs, now fixed (route paths missing `rest/`, site-code date
using host-local time instead of UTC); two are corrections to Java behavior
that the original source-reading pass had characterized less precisely than
the live data shows (`organizationTypes` always-null regardless of real data,
and the `rest/practitioner` not-found path actually being a 3-way split, not
a clean 404/500 binary).

---

## 1. What shipped

9 endpoints across two domains, all reads. **Paths corrected this pass** — the
three `Provider`/`provider` routes were missing their `rest/` prefix until
live testing 404'd on them; see § 3.1.

| Endpoint                                   | Domain       | e2e spec file | Runs against Go? |
| ------------------------------------------- | ------------ | -------------- | ------------------ |
| `GET rest/organization/types`              | organization | `tests/readonly/b2-organization.spec.ts` | yes |
| `GET rest/organization-list`               | organization | same file | yes |
| `GET rest/organization/{id}`               | organization | same file | yes |
| `GET rest/organization/generate-site-code` | organization | same file | yes |
| `GET rest/departments-for-site`            | organization | same file | yes |
| `GET rest/Provider/raw/{id}`               | provider     | `tests/readonly/b2-provider.spec.ts` | yes |
| `GET rest/Provider/Person/{id}`            | provider     | same file | yes |
| `GET rest/practitioner`                    | provider     | same file | yes |
| `GET rest/provider/search`                 | provider     | same file | yes |

"Runs against Go" means matched by `playwright.config.ts`'s `go-parity`
project `testMatch` regex and actually passing there — confirmed by running
`npm run test:go-parity` for real against the live Go server, not assumed:
13/13 b2 tests pass (`user-programs` correctly `test.skip()`s itself under
`go-parity`, see § 2). The same 13 assertions (plus `user-programs`, 14/14)
also pass under `npm run test:readonly` against live Java. Full run logs are
in this session's transcript if anyone wants the raw `list` reporter output.

Code: `migration/openelis-go/internal/organization/` and
`migration/openelis-go/internal/provider/`, each following the established
valueholder → daoimpl → service → controller/rest layering (int64 IDs, GORM
query builder, `strconv.FormatInt` at the DTO boundary — same pattern as a2/b1,
see [a2-gorm-migration.md](a2-gorm-migration.md)).

## 2. What was deferred, and why

| Endpoint(s)                                                | Why deferred                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| ---------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GET rest/Organization`                                    | Type-D form-load, not Type-B — embeds `departmentList`, `orgTypes`, `selectedTypes`, address-part scalars (`commune`/`village`/`department`). Much larger scope than a reference read.                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `GET/rest/OrganizationMenu`, `rest/SearchOrganizationMenu` | Struts-legacy `AdminMenuForm` envelope (13 top-level JSON keys); a real, confirmed Java bug (`totalRecordCount` always shows the grand total, even on a filtered search — the search-specific count is computed into a request attribute, then unconditionally overwritten); no e2e contract exists yet to pin the intended shape.                                                                                                                                                                                                                                                                                                                         |
| `GET rest/OrganizationExport`                              | Needs `FhirTransformService` (outputs a FHIR `Bundle`, not an `Organization` DTO) — the migration plan's own D5 principle is facade-not-reimplement for FHIR; porting the transform layer is its own unit of work.                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `GET rest/ProviderMenu`, `rest/SearchProviderMenu`         | Same Struts-menu envelope pattern/complexity as OrganizationMenu; no e2e contract; the Java controller double-fetches the same query per request (confirmed real waste, not a semantic requirement).                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `GET rest/user-programs`                                   | **Not a pure reference read** — `UserServiceImpl.getUserPrograms()` filters the program list by the _calling user's own_ lab-unit/test-section permissions for the "Reception" role (session-scoped, RBAC-dependent). The Go service has no session/auth layer at all yet (a1/a2/b1/b2 are all unauthenticated reads) — implementing this without that infrastructure would mean either building session+RBAC now (out of scope for a Type-B pass) or silently returning the full unfiltered list (a real access-control regression, not an acceptable simplification per Constitution Principle VIII). Deferred until session/RBAC infrastructure exists. |
| `GET rest/organization/search`                             | Out of scope from the start — flagged by the existing e2e spec's own comment as a paginated Type-C search, its own group.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |

## 3. Findings from live Java-vs-Go comparison

### 3.1 Bugs in this Go port itself — found live, now fixed

Static review (build/vet/gofmt, source-trace-by-eye) missed these; only
curling both servers side by side against the same live data surfaced them.

6. **Three provider routes were missing the `rest/` prefix and 404'd on every
   request.** `ProviderRestController` carries a class-level
   `@RequestMapping("/rest")` in Java — every method-level `@GetMapping` path
   is relative to that. This port had registered `Provider/raw/{id}`,
   `Provider/Person/{id}`, and `provider/search` (no `rest/`) instead of
   `rest/Provider/raw/{id}`, `rest/Provider/Person/{id}`,
   `rest/provider/search`. Every call to these three routes 404'd against a
   real deployment; `rest/organization/*` and `rest/practitioner` were
   unaffected (their registrations already included the correct prefix).
   Caught the moment live Java-vs-Go curls were compared side by side — Java
   returned real data, Go returned `NoHandlerFoundException`-shaped 404s for
   the *same* nominally-equivalent path. Fixed in
   `internal/provider/controller/rest/provider.go`; re-verified live after the
   fix (§ 5).

7. **`generate-site-code` used host-local time instead of UTC.** Java's
   `LocalDate.now()` resolves via `ZoneId.systemDefault()`, and
   `docker-compose.yml` pins the webapp container to `TZ=${TZ:-UTC}`
   (confirmed live: `docker exec openelisglobal-webapp date` reports UTC) — so
   Java's real, deployed behavior is UTC-based site codes. This port's Go
   service called plain `time.Now()`, which uses the process's own host
   timezone. Running the Go binary directly on this (JST) dev machine while
   Java ran in its UTC container produced two different site-code dates for
   requests issued in the same instant whenever the true UTC time was still on
   the previous day (JST is UTC+9, so any request in the first 9 hours of a
   JST day lands on "yesterday" in UTC) — live-observed as `S260813-...` from
   Java vs `S260814-...` from Go at the same moment. Fixed by switching to
   `time.Now().UTC()` in `internal/organization/service/organization_service.go`;
   re-verified live — both sides now agree on the date component (the numeric
   sequence itself legitimately differs per call, since it's a shared,
   incrementing DB sequence and every call — Java's or Go's — consumes the
   next value).

### 3.2 Real Java behaviors/bugs — decisions made, not silent fixes or silent copies

Originally found via full source trace (controller → service → DAO → hbm.xml
→ schema); items 1, 8, and 9 were additionally confirmed or corrected against
real live responses this pass (noted inline). None of these are pinned by any
existing e2e assertion yet, so nothing breaks CI either way — the choice below
is a documented judgment call per endpoint, open to being overridden.

1. **`organization/{id}`, `Provider/raw/{id}`, `Provider/Person/{id}`,
   `rest/practitioner`: 500, not 404, for any not-found id.** Root cause:
   `BaseObjectServiceImpl.get(id)` throws `ObjectNotFoundException` instead of
   returning null when a row isn't found; none of these controllers catch it,
   so it propagates to `ControllerSetup`'s generic `RuntimeException` handler
   → HTTP 500. The `if (x != null) {...} else { return 404 }` branch in each
   controller is dead code — it can never take the `else`. **This port
   returns real 404s instead.** This was clearly never the intended behavior
   (why write unreachable 404-returning code otherwise), and reproducing an
   exception-handling gap into new Go code would be actively worse REST design
   with no test locking it in. `rest/practitioner` added to this list this
   pass — live-tested with a nonexistent person id and got the same 500,
   confirming it hits the identical bug class via `personService.getPersonById`
   even though it wasn't in the original source-trace's list.

2. **`GET rest/Organization` (deferred, not shipped): 500 when called with no
   query params at all** — a related but distinct bug from #1: a missing `ID`
   param leaves `id = ""` (not `null`), which fails the `isNew` check, falls
   into `organizationService.get("")`, and hits the same
   `ObjectNotFoundException` path. Confirmed independently by the existing e2e
   spec's own exclusion comment. Not applicable to this port since the endpoint
   itself is deferred — noted here for whoever picks up `rest/Organization`
   later.

3. **`provider/search`: `isActive` is always `false`.** Java's
   `providerData.put("isActive", "Y".equals(provider.getActive()))` compares a
   `Boolean` to the string `"Y"` — always false, regardless of the real `active`
   column. **Live-confirmed this pass, not just source-theorized**: searched
   for all 3 real providers (one `active=false`, two `active=true` in the DB)
   and Java's `isActive` was `false` for all three, unconditionally. **This
   port returns the real active flag** — confirmed correct against the same 3
   rows (`false`/`true`/`true`, matching the DB exactly). Reproducing an
   always-false field would be actively misleading to any new consumer of this
   endpoint.

4. **`provider/search`: client-supplied `pageSize` is silently capped.** Java's
   DAO methods internally fetch `page.defaultPageSize + 1` rows (a server
   _config_ value), not `pageSize + 1` — the controller's own trim-to-`pageSize`
   logic checks against the wrong bound, so a request for `pageSize=50` returns
   at most `page.defaultPageSize` rows (commonly ~20) with no error or
   indication. **This port has no `page.defaultPageSize` config concept at all
   and uses the caller's `pageSize` directly as the real SQL `LIMIT`** —
   simpler, and arguably more correct REST behavior, not just a reproduction
   gap. **Not live-exercised this pass** — the dev DB only has 3 providers
   total, too few to trigger Java's cap (commonly ~20) either way; this item
   is still source-confirmed only, not live-confirmed. Said plainly rather
   than implied, per the same standard as everything else in this section.

5. **`OrganizationMenu`/`SearchOrganizationMenu` (deferred, not shipped):
   `totalRecordCount` always shows the grand total, even on a filtered search.**
   See § 2 table — noted here since it's the same class of bug as #3/#4, for
   whoever picks these endpoints up later.

8. **`organization-list` / `organization/{id}`: `organizationTypes` is
   *always* `null` in Java's real response, even for organizations that do
   have real rows in the `organization_organization_type` join table.** Root
   cause, confirmed via `Organization.hbm.xml`: the `organizationTypes`
   `<set>` mapping is `lazy="true"`. The service methods run inside
   `@Transactional(readOnly = true)`, which commits and closes the Hibernate
   session before the controller hands the entity to Jackson; Jackson (via the
   Hibernate5 module, without forced lazy-loading) serializes an uninitialized
   lazy collection proxy as `null` rather than triggering a load or throwing.
   This is **not** environment-dependent/lazy-timing-flaky as an earlier draft
   of this doc speculated — it was live-tested against 3 real organizations
   (one with no assigned type, two with real join-table rows for types 5 and
   6) and Java returned `null` unconditionally, every time, regardless of
   whether real data existed. **This port deliberately keeps its current,
   different behavior**: `GetTypesForOrgIDs` eager-batch-loads the real join
   data in one query per `organization-list`/`organization/{id}` call and
   returns a real array (populated or `[]`), never `null`. Reasoning: Java's
   `null` here isn't reachable application logic a client could rely on, it's
   an artifact of open-session-in-view being off for this particular field;
   returning the real, correct data is strictly more useful and is exactly the
   kind of "avoid N+1, but don't reproduce an accidental serialization gap"
   choice this project has made elsewhere (see #1, #3). Flagged here as a
   **confirmed, deliberate divergence** an e2e spec must assert asymmetrically
   (Java: `organizationTypes === null`; Go: a real array) rather than pin with
   a single shared `toEqual`.

9. **`rest/practitioner?providerId=<personId>`: three different outcomes for
   three different kinds of "not found", not a clean 404/200 split.** Traced
   via `DisplayListController.getProviderInformation` +
   `ProviderDAOImpl.getProviderByPerson`:
   - Person id doesn't exist at all → `personService.getPersonById` hits the
     same `ObjectNotFoundException` → 500 (bug #1's class, live-confirmed).
   - Person exists but has **no** linked `Provider` row →
     `getProviderByPerson` runs `from Provider p where p.person.id = :personId`,
     gets an empty list, and returns `null` cleanly (no exception) — the
     controller's `return provider;` then returns `null`, and Spring's
     `@ResponseBody` serializes that as **HTTP 200 with a genuinely empty body**
     (`Content-Length: 0`, live-confirmed with response headers, not just an
     empty-looking curl output). This is a real, reachable, working code path
     in Java, not a bug in the same sense as #1 — it just happens to produce
     an ambiguous 200.
   - Person exists and has a linked Provider → 200 with the real Provider+Person
     JSON (already covered by #1's "otherwise it works" case).

   An earlier draft of this doc assumed the middle case also hit the 500 bug;
   live testing showed it's actually a distinct, clean 200-empty-body case.
   **This port keeps a uniform 404 for both "person missing" and "person
   exists but unlinked"** rather than adding a third response shape
   (200-empty) to match Java exactly — deliberate, for REST consistency across
   this whole endpoint family (every other not-found case in this file is a
   404), and because a bare 200-empty-body response is easy for a real caller
   to mishandle (e.g. `response.json()` on an empty body throws). Flagged as
   an open judgment call, not a silent mismatch — see § 6.

## 4. Schema / ORM notes for whoever extends this

- **Organization, OrganizationType, Provider, Person are all
  Hibernate-XML-mapped in Java** (`hibernate/hbm/*.hbm.xml`), not JPA-annotated.
  Column names/nullability for this port came from the real DB schema
  (`db/dbInit/OpenELIS-Global.sql` + the relevant Liquibase changesets), not the
  hbm.xml metadata — confirmed stale in multiple places:
  - `Organization.organizationName`: hbm says `length=40`, real column is
    `varchar(80)`.
  - `Provider.externalId`: hbm says `length=10`, real column is `varchar(50)`.
  - `Provider.personId`: hbm's `<many-to-one>` doesn't declare
    `not-null="true"`, but the DB itself has `NOT NULL` + `MATCH FULL` FK
    (`prov_person_fk`). Trust the DDL over the hbm.xml for constraints, always.
- **`OrganizationType.name` maps to DB column `short_name`, not `name`** — easy
  to get backwards.
- **`organization_type` is a genuine many-to-many lookup**, linked via the join
  table `organization_organization_type` (composite PK `(org_id, org_type_id)`),
  not a single embedded FK on `organization`. One organization can have multiple
  types. `GetTypesForOrgIDs` batch-loads this in one JOIN query for
  `organization-list` (avoiding N+1) and returns real data instead of
  replicating Hibernate's `lazy="true"` Set, which Java's own stack serializes
  as an unconditional `null` — root-caused and live-confirmed, see § 3.2 item 8
  for the full mechanism and why this port doesn't copy it.
- **`Organization.testSections` is permanently `[]`** — confirmed via a
  full-repo grep that nothing in the Java codebase ever calls
  `setTestSections()` on an `Organization`. Hardcoded to `[]` in the DTO rather
  than modeled as a real relationship.
- **The self-referencing `Organization.organization` (parent org) field is NOT
  ported** — Java guards infinite JSON recursion with
  `@JsonIgnoreProperties({"organization"})` plus a runtime
  `handleSelfReferencingParentOrg` fixup. Neither `organization-list` nor
  `organization/{id}`'s DTO embeds the parent at all in this port. Not covered
  by the existing e2e contract (its allowed-keys list doesn't include
  `"organization"` either). Flagged as a known gap, not silently matched by
  coincidence.
- **`datim_org_code`/`datim_org_name`** are real columns on
  `clinlims.organization` but are never Hibernate-mapped in Java (grepped — no
  hbm `<property>` for either) — no Java endpoint ever serializes them, so this
  port's valueholder doesn't include them either. Matching API behavior, not
  schema completeness.
- **`rest/practitioner`'s `providerId` query param is actually a Person id**,
  not a Provider id — confirmed via `DisplayListService`'s provider-picker list,
  which is keyed by Person id throughout the app. Kept as-is (misleading name
  and all) since every real caller already uses it that way.
- **All nullable DB columns are ported as Go pointer fields**
  (`*string`/`*int`/`*time.Time`), scanned by GORM directly with no `COALESCE` —
  `nil` + `json:",omitempty"` mirrors Jackson's `Include.NON_NULL` (null fields
  dropped from the response entirely, not emitted as JSON `null`) exactly,
  including for fields where NULL and `""` are both real, distinct possible
  values in Postgres.
- **`Provider`/`Person` DTOs were missing `lastupdated` entirely** — the
  valueholders already scanned the column (`Provider.Lastupdated`,
  `Person.Lastupdated`), but `providerDTO`/`personDTO` in
  `internal/provider/controller/rest/provider.go` never surfaced it, so it was
  silently dropped at the DTO boundary. Caught live: Java's real
  `Provider/raw/{id}`, `Provider/Person/{id}`, and `rest/practitioner`
  responses all include `lastupdated` (epoch millis) at both the provider and
  the nested person level. Fixed — both DTOs now include it
  (`json:"lastupdated,omitempty"`), same `UnixMilli()` encoding as
  `organizationDTO.Lastupdated`. Re-verified live: exact epoch-millis match on
  both levels for all 3 real providers.

## 5. Verification

**Static**: `go build ./...`, `go vet ./...`, `gofmt -l` all clean across
`internal/organization/`, `internal/provider/`, and the `cmd/openelis/main.go`
wiring.

**Runtime — live side-by-side capture (done this pass).** The dev stack
(`docker-compose`, running inside WSL2 Ubuntu-24.04 on this dev machine) came
up successfully after two unrelated infra fixes: a stale
`databasechangeloglock` row blocking Java's Liquibase startup (cleared with a
direct, standard `UPDATE ... SET locked=false`), and repeated WSL2
idle-shutdown cycling between disconnected `wsl.exe` calls (fixed by holding a
persistent background `wsl.exe -d Ubuntu-24.04 -- sleep 3600` process alive;
Windows' `localhost`→WSL2 port-forwarding relay also intermittently stopped
proxying specific ports after VM restarts during this session — worked around
by hitting the WSL VM's real IP directly instead of `localhost` when that
happened, not by changing anything in the app stack itself).

Methodology, precisely:

1. **Authenticated against the real Java webapp** exactly the way the
   existing Playwright suite does (`tests/auth.setup.ts` /
   `fixtures/contract.ts`): touch `session`, `POST ValidateLogin?apiCall=true`
   with `loginName=admin`/`password=adminADMIN!` form fields, confirm
   `{"success":true}`, reuse the resulting session cookie for every subsequent
   call — via `curl` with a cookie jar, not a browser.
2. **Queried the live DB directly** (`docker exec ... psql`) to find real,
   varied row ids to test against — not synthetic ones: 3 organizations (one
   with no assigned type, two with real `organization_organization_type`
   rows), 3 providers (one `active=false`, two `active=true`, one with a
   deliberately-odd `shortName` value of literal `""`), 213 people (one
   linked to no provider at all, used as an edge case).
3. **Curled every one of the 9 endpoints against both servers** — same live
   Postgres instance, same ids, same query params — for the happy path *and*
   edge cases per endpoint: a nonexistent id, a missing required query param,
   an org with no assigned type, a person with no linked provider, an empty
   search string, a search with no matches, explicit `page`/`pageSize`
   values. Full transcript of the final, all-green run is in this session's
   scratchpad (`compare_b2_results3.txt`) if anyone wants the raw output.
4. **Compared every JSON field**, not just HTTP status — this is what
   surfaced § 3.1's two Go bugs and § 3.2 items 8/9's precise Java behavior;
   status-code-only comparison would have missed all four.
5. Two full passes: the first surfaced the four findings in § 3; all four were
   fixed/root-caused, the Go server was rebuilt, and a second full pass
   confirmed every one of the 9 endpoints now matches on every field except
   the deliberately-documented divergences (§ 3.2 items 1, 3, 8, 9).

**Runtime — committed Playwright specs, actually executed (done this pass).**
The curl-level findings above were turned into real spec files and run for
real, not just written and assumed correct:

- `tests/readonly/b2-organization.spec.ts` extended with 3 new tests
  (`organization/{id}`, `generate-site-code`, `departments-for-site`); its 2
  existing tests (`organization/types`, `organization-list`) untouched in
  substance. `user-programs` now `test.skip()`s itself under `go-parity` with
  an inline reason, instead of either failing there or being silently
  excluded.
- `tests/readonly/b2-provider.spec.ts` written from scratch: 7 tests across
  all 4 provider endpoints, ids discovered live via `provider/search` and
  direct DB queries rather than hardcoded, so the suite isn't tied to this
  session's specific dataset.
- `playwright.config.ts`'s `go-parity` project `testMatch` extended to
  include both files; `package.json` gained a `test:go-parity` script
  matching the existing `test:readonly`/`test:mutating` convention.
- Every divergence in § 3.2 (items 1, 3, 9) that's actually exercised by live
  data is asserted **on both sides explicitly** — branched on
  `testInfo.project.name === "go-parity"` — not silently pinned to whichever
  server's value happened to be captured first. Item 8
  (`organizationTypes`) didn't need a branch: the existing assertion
  (`null` or array) already tolerated both.
- Executed for real: `npm run test:readonly` (Java) — 479/479 passing,
  including all pre-existing specs (no regressions from the `playwright.config.ts`/
  `package.json` changes). `npm run test:go-parity` (Go) — 21/22 passing, 1
  skipped (`user-programs`, by design); the one failure is `a1-server-time`'s
  IANA-timezone check, pre-existing and unrelated to b2 (Go's `rest/server-time`
  emits a host-timezone abbreviation instead of an IANA id when run from a
  non-UTC host — same bug *class* as § 3.1 #7 but different code, flagged
  separately, not fixed here to keep this change scoped to b2). All 13 b2
  tests across both files pass under `go-parity`.

**What this *is* now**: a committed, repeatable, two-project Playwright
contract — not just a one-time manual capture — proving this port behaves
identically to Java for all 9 endpoints modulo the divergences documented in
§ 3, each backed by an assertion that actually runs on both sides.

## 6. Open questions / not yet decided

- Whether `user-programs`' session/RBAC dependency should be unblocked by
  building minimal session infrastructure now, or left deferred until a later
  wave.
- Whether § 3.2's divergences are the final call, or whether
  faithful-reproduction is preferred for any of them now that real parity
  testing is possible:
  - #1 (404 vs 500 on not-found, 4 endpoints)
  - #3 (`isActive`, now live-confirmed real-vs-always-false)
  - #4 (`pageSize` cap — still source-only, not live-confirmed; the dev
    dataset is too small to exercise it)
  - #8 (`organizationTypes`: real array vs Java's unconditional `null`)
  - #9 (`rest/practitioner` on an unlinked person: 404 vs Java's real
    200-empty-body)
