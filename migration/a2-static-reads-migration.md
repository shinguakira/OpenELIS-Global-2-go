# a2 — Static + first single-table DB reads (scoped migration plan)

Status: **Stages 1–3 complete ✅; Stage 4 (i18n) planned.**
Branch: `migration/a2-static-reads` (forked from `migration-base`).
Companion to [a1-server-time-migration.md](a1-server-time-migration.md),
[endpoint-migration-taxonomy.md](endpoint-migration-taxonomy.md),
[branch-naming.md](branch-naming.md).

This doc captures the source investigation for the a2 batch and freezes the
**re-scoped** plan agreed with the maintainer: a2 covers **7** endpoints —
the original 5 low-risk ones plus the 2 status-type endpoints that were pulled
in mid-branch (they share the same DB infrastructure pattern and belong here).

---

## 0. Why a2 was re-scoped (10 → 7)

The `branch-naming.md` a2 row listed 10 endpoints as "static / computed /
read-only config". Reading the Java source shows that framing is wrong for half
of them — only **2 of 10 are genuinely static**; the rest hit the DB, the
`ConfigurationProperties` subsystem, or build a cached recursive tree. Porting
all 10 now would drag Waves 1–5 infrastructure into the pilot.

**a2 is therefore scoped to 7 endpoints** with a coherent theme:
**"static handlers + first single-table Postgres reads + status-type
reference data (DB ids + i18n labels)."**

### In scope (7) — all ported ✅

| # | Endpoint | Data source | Bucket | Stage |
|---|----------|-------------|--------|-------|
| 1 | `GET /rest/math-functions` | hardcoded 14-item list, no DB | **static** | 1 |
| 2 | `GET /rest/sample-item-status-types` | hardcoded 3-item list, no DB | **static** | 1 |
| 3 | `GET /rest/supportedlocales` | `supported_locale` table, `getAll()` unordered | **1-table DB** | 2 |
| 4 | `GET /rest/supportedlocales/active` | same table, `WHERE is_active ORDER BY sort_order ASC` | **1-table DB** | 2 |
| 5 | `GET /rest/supportedlocales/fallback` | same table, single row or 404 | **1-table DB** | 2 |
| 6 | `GET /rest/analysis-status-types` | `status_of_sample` + `message_en.properties` | **DB + i18n** | 3→4 |
| 7 | `GET /rest/sample-status-types` | `status_of_sample` + `message_en.properties` | **DB + i18n** | 3→4 |

Endpoints 6–7 were initially listed as "deferred to Type B" in an earlier draft
but were pulled into a2 mid-branch: they share the same DB connection and follow
the same `(status_type, name) → id` pattern as the locale endpoints. Stage 3
ported them with hardcoded English labels; Stage 4 (this plan) replaces the
hardcoded values with the real `display_key → message_en.properties` lookup
that Java uses.

### Deferred (3) — moved to their proper branches

| Endpoint | Why not a2 | Goes to |
|----------|-----------|---------|
| `open-configuration-properties` | whole `ConfigurationProperties` subsystem + `site_information` + localization + OAuth/SAML | a dedicated config branch |
| `configuration-properties` | superset of open-config + name-regex from DB charset; auth-only | a dedicated config branch |
| `menu` | **Type C** — cached recursive tree + plugin injection + external JSON include/exclude filter | Type C (`migration/c-menu` later) |

---

## 1. Per-endpoint Java source (frozen baseline)

Shared: `IdValuePair` (`common/util/IdValuePair.java`) serializes as
`{"id": "...", "value": "..."}`.

### 1. `math-functions` — 🟢 trivial (static)
- Handler: `CalculatedValueRestController.getMathFunctions()`
  (`testcalculated/controller/rest/CalculatedValueRestController.java:109`)
  → `Operation.mathFunctions()` (`testcalculated/valueholder/Operation.java:135`).
- Mapping: `@RequestMapping("/rest/")` + `@GetMapping("math-functions")` → `/rest/math-functions`.
- Returns `List<IdValuePair>`, **14 fixed** entries (English literals baked in):
  `+`=Plus, `-`=Minus, `/`=Divided By, `*`=Multiplied By, `(`=Open Bracket,
  `)`=Close Bracket, `==`=Equals, `!=`=Does Not Equal, `>=`=Is Greater Than Or
  Equal, `<=`=Is Less Than Or Equal, `IS_IN_NORMAL_RANGE`=Is With In Normal
  Range, `IS_OUTSIDE_NORMAL_RANGE`=Is Out Side Normal Range, `&&`=And, `||`=Or.
- Port: hardcode the 14 pairs. No DB, no i18n.

### 2. `sample-item-status-types` — 🟢 trivial (static)
- Handler: `DisplayListController.getSampleItemStatusTypes()`
  (`common/rest/DisplayListController.java:503`).
- Mapping: `/rest/sample-item-status-types`.
- Returns `List<IdValuePair>`, **3 fixed** entries: `{"","All"}`,
  `{"active","Active"}`, `{"disposed","Disposed"}`. English literals, no DB.
- Port: hardcode the 3 pairs. (Sits next to the two DB-backed status endpoints
  but is genuinely static — do not confuse with `analysis-/sample-status-types`.)

### 3. `supportedlocales` — 🟢 easy (1-table DB), **no trailing slash**
- Handler: `SupportedLocaleRestController.getAllLocales()`
  (`localization/controller/rest/SupportedLocaleRestController.java:54`).
- Mapping: class `@RequestMapping("/rest/supportedlocales")` + `@GetMapping` (empty)
  → **`/rest/supportedlocales`, no trailing slash**. Spring Framework 6 disabled
  trailing-slash matching by default, so **`/rest/supportedlocales/` returns 404**
  (this is the current e2e failure — a wrong test path, not a Java bug).
- Returns `List<SupportedLocaleDTO>` = `{id, localeCode, displayName, active,
  fallback, sortOrder}` (DTO inline in the controller).
- Source: `supportedLocaleService.getAll()` → base DAO `getAll()` on table
  `supported_locale`. **All** rows (active + inactive), **no explicit order**
  (Hibernate default — NOT sorted by `sortOrder`).

### 4. `supportedlocales/active` — 🟢 easy (1-table DB), **public**
- Handler: `SupportedLocaleRestController.getActiveLocales()` (`:66`).
- Mapping: `/rest/supportedlocales/active`. **Public** — whitelisted in
  `SecurityConfig.java:107` `OPEN_PAGES` (served without auth).
- Returns `List<SupportedLocaleDTO>`.
- Source: `getAllActive()` → `getAllMatchingOrdered("active", true, "sortOrder",
  false)` = **`WHERE is_active = true ORDER BY sort_order ASC`**. The one
  ordering rule the Go port must reproduce.

### 5. `supportedlocales/fallback` — 🟢 trivial (1-table DB), single object or 404
- Handler: `SupportedLocaleRestController.getFallbackLocale()` (`:94`).
- Mapping: `/rest/supportedlocales/fallback`. **Precedence nuance:** coexists
  with `@GetMapping("/{id}")` (`:79`); Spring matches literal `/fallback` over
  the `{id}` pattern. A Go router must give literal segments priority so
  `fallback` isn't captured as an id.
- Returns a **single** `SupportedLocaleDTO` (not a list), or **404** if no row
  has `is_fallback = true`.
- Source: `getFallback()` → `getAllMatching("fallback", true)` = `WHERE
  is_fallback = true`, first / empty.

---

## 2. `supported_locale` — the one new DB dependency (the point of a2)

Endpoints 3–5 all read table `supported_locale` via
`SupportedLocaleService` → `SupportedLocaleDAO`. Columns used (from the DTO):
`id`, `locale_code`, `display_name`, `is_active`, `is_fallback`, `sort_order`.

Go work a2 introduces (mirroring the Java folder layout, per the
Java-mirror decision — see [[go-folder-mirrors-java]]):
- a Postgres connection (`database/sql` + `pq`, or `pgx`) wired from the same env
  the Java app uses (`DATABASE_HOST/PORT/NAME/USER/PASSWORD`), pointed at the
  **same** DB the parity harness runs against;
- `internal/localization/valueholder/` — the `SupportedLocaleDTO` struct
  (`id, localeCode, displayName, active, fallback, sortOrder`);
- `internal/localization/dao/` — `getAll` / `getAllActive` / `getFallback`
  reproducing the exact SQL (incl. `ORDER BY sort_order ASC` on active, and the
  no-order on getAll);
- `internal/localization/service/` — thin pass-through (mirrors Java service);
- `internal/localization/controller/rest/` — the 3 routes, `/fallback` before
  `/{id}` in match priority, single-object vs list shapes, 404-on-empty.

**Verify before coding:** the exact column names / `is_active` vs `active`
mapping against the live schema (Hibernate `@Column`), so the Go SQL matches.

---

## 3. e2e parity gate — STRICT ✅ (Java baseline captured, 9/9 pass against Go)

The shallow "200 + not-login" a2 loop was **replaced** with strict per-endpoint
tests in `tests/readonly/a2-static-reads.spec.ts`.
`go-parity` runs `a1-server-time.spec.ts` + `a2-static-reads.spec.ts`; all
**9 tests pass against both Java and Go** (a1: 2 tests, a2: 7 tests).

Live-captured Java baseline (the frozen contract the Go port must reproduce):

| Endpoint | Captured baseline (Java) |
|----------|--------------------------|
| `math-functions` | array of **14** `{id,value}`; `+`→Plus … `\|\|`→Or |
| `sample-item-status-types` | exactly `[{"","All"},{"active","Active"},{"disposed","Disposed"}]` |
| `supportedlocales` | 2 rows: `en`(active,**fallback**,sort 1), `fr`(active,not-fallback,sort 2); **`id` is a STRING** (`"1"`), `active`/`fallback` booleans, `sortOrder` int |
| `supportedlocales/active` | same 2 rows (both active), ascending `sortOrder` |
| `supportedlocales/fallback` | **single object** (not array) = the `en` row; 200 (seed has a fallback) |
| `analysis-status-types` | `[{id:"0",value:""},{id:"4",value:"Not started"},…]` — 6 entries |
| `sample-status-types` | `[{id:"0",value:""},{id:"1",value:"No tests have been run…"},…]` — 3 entries |

What each test now guarantees (pass ⇒ behavior, not just reachability):
- **exact deep-equality** (`toEqual`) for all 7 endpoints — any drift in value,
  type, count, or key set fails;
- **strict types** on locale rows (`id` string, `active`/`fallback` boolean,
  `sortOrder` int, exactly the 6 keys) — catches Go serialization drift;
- **cross-view invariants**: `/active` = the active subset of the full list
  (same ids, each row serialized identically), non-decreasing `sortOrder`;
  `/fallback` = the full list's single fallback row as an object (or 404 if none);
- Status-type values are pinned to the live-captured English labels (which come
  from `message_en.properties` via `display_key` — see Stage 4 below).


---

## 4. Status-type Java source (endpoints 6–7)

Both endpoints are in `DisplayListController.java` (`:459`, `:483`) and call
`IStatusService.getStatusID(enum)` + `IStatusService.getStatusName(enum)`.

`getStatusName(AnalysisStatus)` traces through:
```
StatusService.analysisStatusToObjectMap.get(enum)          → StatusOfSample row
StatusOfSample.getLocalizedName()                          → BaseObject.getLocalizedName()
  nameKey = status_of_sample.display_key  (HBM: StatusOfSample.hbm.xml:36)
  MessageUtil.getContextualMessage(nameKey)                → message_en.properties lookup
  fallback: StatusOfSample.getDefaultLocalizedName()       → getStatusOfSampleName() = DB `name` col
```

Key finding: **`display_key` is a real column in `status_of_sample`**, mapped via
the legacy Hibernate HBM file (`src/main/resources/hibernate/hbm/StatusOfSample.hbm.xml:36`)
as `<property name="nameKey" column="display_key">`. The `@Transient` annotation
on `BaseObject.nameKey` is overridden by the HBM mapping for this entity.

`StatusService` builds maps at `@PostConstruct` by matching DB `name` column
to enum members:
- `"Not Tested"` → `AnalysisStatus.NotStarted` → `display_key` = `status.test.notStarted`
  → `message_en.properties` → **"Not started"**
- `"Test Entered"` → `OrderStatus.Entered` → `display_key` = `status.sample.notStarted`
  → `message_en.properties` → **"No tests have been run for this sample"**

Relevant keys in `message_en.properties` (lines 6903–6916):
```
status.sample.notStarted      = No tests have been run for this sample
status.sample.started         = Some tests have been run on this sample
status.test.biologist.reject  = Not accepted by biologist
status.test.canceled          = Canceled
status.test.notStarted        = Not started
status.test.tech.accepted     = Accepted by technician
status.test.tech.rejected     = Not accepted by technician
```

---

## 5. Stage 4 plan — i18n: replace hardcoded labels with real lookup

**Why:** Stage 3 ported endpoints 6–7 with hardcoded English values (a known
shortcut). The correct source is `display_key` → `message_en.properties`. Stage 4
implements the real lookup, matching Java's behavior exactly and establishing the
shared i18n infrastructure all future migration units will reuse.

### Step 1 — Properties loader (`internal/common/i18n/properties.go`)

- Parse `src/main/resources/languages/message_en.properties` (key `=` value lines;
  `#` comment lines ignored; leading/trailing whitespace trimmed).
- Return `map[string]string` (`"status.test.notStarted"` → `"Not started"`).
- Use `//go:embed` to bundle the file into the binary at compile time — same
  immutability guarantee as the Java WAR. The embed path is relative to the Go
  module root, pointing at the Java source tree.
- Load once at startup; pass into any service that needs label resolution.

### Step 2 — Extend `StatusService` (`internal/common/services/status.go`)

- Add `display_key` to the query:
  `SELECT id::text, status_type, name, display_key FROM clinlims.status_of_sample`
- Accept a `map[string]string` (the properties map) at construction.
- Store `(status_type + "\x00" + name) → {id, label}` where
  `label = props[displayKey]` if `displayKey != ""`, else fallback to `name`
  (mirrors Java's `getDefaultLocalizedName()` path).
- Expose `EntryByName(statusType, name string) (id, label string)` (or two
  separate accessors — `IDByName` stays, `LabelByName` is added).

### Step 3 — Remove hardcoded values from `display_list.go`

- Drop the `value string` field from `statusEntry`; the label now comes from
  `StatusService`.
- `statusList` calls `svc.EntryByName(e.statusType, e.internalName)` for both id
  and label in one shot.

### Step 4 — Wire in `main.go`

- Load `i18n.LoadProperties(...)` once at startup → pass `props` to
  `NewStatusService(db, props)`.

### Fallback / error contract

- Missing `display_key` (empty or null): fall back to the `name` column —
  same as Java's `getDefaultLocalizedName()`.
- Missing properties key (key present but not in the file): fall back to the
  key string itself — same as Java's
  `localizedName.equals(nameKey.trim())` fallback path.
- Properties file missing at embed time: compile error — intentional, same as
  the WAR refusing to build without its resources.

### Cross-cutting note for future migration units

The `display_key` → `message_en.properties` pattern appears in many entities
beyond `status_of_sample` (`gender`, `type_of_sample`, `test_section`, etc. all
extend `BaseObject` and likely have `display_key` columns via their own HBM
files). Every future migration unit that returns a localized name **must**:
1. Fetch `display_key` alongside the row.
2. Look it up in the shared `i18n.LoadProperties` map (loaded once at startup).
3. Fall back to the `name`/`description` column if `display_key` is null/empty.

The `internal/common/i18n` package built in Stage 4 is the shared infrastructure
for all subsequent migration units.

---

## 6. Commit groups (historical + pending)

| # | Commit | Status |
|---|--------|--------|
| 1 | Plan doc + `branch-naming.md` a2 scope update | ✅ done |
| 2 | Static handlers: `math-functions`, `sample-item-status-types` | ✅ done |
| 3 | Postgres wiring + `supported_locale` DAO/service/controller + e2e strict rewrite + `go-parity` extension | ✅ done |
| 4 | Type-B status reads: `status_of_sample` `StatusService` + `analysis-/sample-status-types` routes (Stage 3 — hardcoded labels) | ✅ done |
| 5 | **Stage 4 — i18n**: `internal/common/i18n` properties loader, extend `StatusService` to use `display_key`, remove hardcoded labels from `display_list.go` | **planned** |
