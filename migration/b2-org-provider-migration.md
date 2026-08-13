# b2 — Organization + Provider Migration

Status: **implemented, statically verified (go build/vet/gofmt clean); NOT yet
runtime-verified against a live DB or the e2e parity suite** — the dev Postgres
instance was unreachable for the whole implementation session (see §
Verification). Branch: `migration/b2-org-provider` (forked from
`migration-base`, per [branch-naming.md](branch-naming.md)). Taxonomy Type B,
Wave 2 per [endpoint-migration-order.md](endpoint-migration-order.md).

---

## 1. What shipped

9 endpoints across two domains, all reads:

| Endpoint                                   | Domain       | e2e contract?                                  |
| ------------------------------------------ | ------------ | ---------------------------------------------- |
| `GET rest/organization/types`              | organization | yes — `tests/readonly/b2-organization.spec.ts` |
| `GET rest/organization-list`               | organization | yes — same file                                |
| `GET rest/organization/{id}`               | organization | no                                             |
| `GET rest/organization/generate-site-code` | organization | no                                             |
| `GET rest/departments-for-site`            | organization | no                                             |
| `GET Provider/raw/{id}`                    | provider     | no                                             |
| `GET Provider/Person/{id}`                 | provider     | no                                             |
| `GET rest/practitioner`                    | provider     | no                                             |
| `GET provider/search`                      | provider     | no                                             |

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

## 3. Real Java bugs found — decisions made, not silent fixes or silent copies

Confirmed via full source trace (controller → service → DAO → hbm.xml → schema),
not guessed. None of these are pinned by any existing e2e assertion, so nothing
breaks CI either way — the choice below is a documented judgment call per
endpoint, open to being overridden.

1. **`organization/{id}`, `Provider/raw/{id}`, `Provider/Person/{id}`: 500, not
   404, for any not-found id.** Root cause: `BaseObjectServiceImpl.get(id)`
   throws `ObjectNotFoundException` instead of returning null when a row isn't
   found; none of these three controllers catch it, so it propagates to
   `ControllerSetup`'s generic `RuntimeException` handler → HTTP 500. The
   `if (x != null) {...} else { return 404 }` branch in each controller is dead
   code — it can never take the `else`. **This port returns real 404s instead.**
   This was clearly never the intended behavior (why write unreachable
   404-returning code otherwise), and reproducing an exception-handling gap into
   new Go code would be actively worse REST design with no test locking it in.

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
   column. **This port returns the real active flag.** Reproducing an
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
   gap.

5. **`OrganizationMenu`/`SearchOrganizationMenu` (deferred, not shipped):
   `totalRecordCount` always shows the grand total, even on a filtered search.**
   See § 2 table — noted here since it's the same class of bug as #3/#4, for
   whoever picks these endpoints up later.

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
  `organization-list` (avoiding N+1), rather than trying to replicate
  Hibernate's lazy `Set` loading — the exploration found that lazy-load timing
  is itself environment-dependent in the Java app, so eager-and-always-populate
  is the safer, unambiguous choice here.
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

## 5. Verification

**Static**: `go build ./...`, `go vet ./...`, `gofmt -l` all clean across
`internal/organization/`, `internal/provider/`, and the `cmd/openelis/main.go`
wiring.

**Runtime**: **not done.** The dev Postgres instance (`localhost:15432`) was
unreachable for this entire implementation session, and `docker` isn't on this
environment's PATH to bring it back up. Nothing in this doc's endpoint list has
been exercised against a real database, and the two endpoints with an existing
frozen contract (`organization/types`, `organization-list` in
`openelis-api-e2e/tests/readonly/b2-organization.spec.ts`) have **not** been run
through `playwright test --project=go-parity`. This is a real, outstanding gap,
not an oversight being glossed over — next step for whoever has DB access: bring
the stack up, run the existing b2 e2e spec against this branch, and work through
whatever it finds.

## 6. Not yet decided

- Whether `user-programs`' session/RBAC dependency should be unblocked by
  building minimal session infrastructure now, or left deferred until a later
  wave.
- Whether the 404-vs-500 and `isActive`/`pageSize` divergences in § 3 should be
  the final call, or whether faithful-reproduction is preferred for any of them
  once real parity testing is possible.
