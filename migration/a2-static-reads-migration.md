# a2 — Static + first single-table DB reads (scoped migration plan)

Status: **investigation saved / NOT YET IMPLEMENTED** (no Go code written yet).
Branch (to be created by maintainer from `migration-base`):
`migration/a2-static-reads`.
Companion to [a1-server-time-migration.md](a1-server-time-migration.md),
[endpoint-migration-taxonomy.md](endpoint-migration-taxonomy.md),
[branch-naming.md](branch-naming.md).

This doc captures the source investigation for the a2 batch and freezes the
**re-scoped** plan agreed with the maintainer: a2 covers **5** endpoints, not
the 10 originally listed under "Type A static reads".

---

## 0. Why a2 was re-scoped (10 → 5)

The `branch-naming.md` a2 row listed 10 endpoints as "static / computed /
read-only config". Reading the Java source shows that framing is wrong for half
of them — only **2 of 10 are genuinely static**; the rest hit the DB, the
`ConfigurationProperties` subsystem, or build a cached recursive tree. Porting
all 10 now would drag Waves 1–5 infrastructure into the pilot.

**a2 is therefore scoped to the 5 low-risk endpoints** and given a coherent
theme: **"static handlers + the first single-table Postgres read."**

### In scope (5)

| # | Endpoint | Data source | Bucket |
|---|----------|-------------|--------|
| 1 | `GET /rest/math-functions` | hardcoded 14-item list, no DB | **static** |
| 2 | `GET /rest/sample-item-status-types` | hardcoded 3-item list, no DB | **static** |
| 3 | `GET /rest/supportedlocales` | `supported_locale` table, `getAll()` unordered | **1-table DB** |
| 4 | `GET /rest/supportedlocales/active` | same table, `WHERE is_active ORDER BY sort_order ASC` | **1-table DB** |
| 5 | `GET /rest/supportedlocales/fallback` | same table, single row or 404 | **1-table DB** |

a2's new capability = **Go → Postgres**, proven on the safest possible table
(`supported_locale`: one flat table, tiny DTO, one `ORDER BY`, one `WHERE`,
one single-row lookup). Endpoints 1–2 need no infra beyond the a1 static pattern.

### Deferred (5) — moved to their proper branches

| Endpoint | Why not a2 | Goes to |
|----------|-----------|---------|
| `analysis-status-types` | reads `status_of_sample` via `StatusService`; **DB-generated IDs** + localized names | Type B (reference reads) |
| `sample-status-types` | same `StatusService` / `status_of_sample` mechanism | Type B (reference reads) |
| `open-configuration-properties` | whole `ConfigurationProperties` subsystem + `site_information` + localization + OAuth/SAML | a dedicated config branch |
| `configuration-properties` | superset of open-config + name-regex from DB charset; auth-only | a dedicated config branch |
| `menu` | **Type C** — cached recursive tree + plugin injection + external JSON include/exclude filter | Type C (`migration/c-menu` later) |

`branch-naming.md` a2 row should be updated to reflect this 5-endpoint scope.

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

## 3. e2e parity gate — STRICT rewrite landed ✅ (Java baseline captured)

The shallow "200 + not-login" a2 loop was **replaced** with strict per-endpoint
tests in `tests/readonly/a2-static-reads.spec.ts`. Verified:
**8/8 pass against Java**, and a one-value mutation was confirmed to turn it red
(the assertions have teeth). Live-captured Java baseline (the frozen contract the
Go port must reproduce):

| Endpoint | Captured baseline (Java) |
|----------|--------------------------|
| `math-functions` | array of **14** `{id,value}`; `+`→Plus … `\|\|`→Or |
| `sample-item-status-types` | exactly `[{"","All"},{"active","Active"},{"disposed","Disposed"}]` |
| `supportedlocales` | 2 rows: `en`(active,**fallback**,sort 1), `fr`(active,not-fallback,sort 2); **`id` is a STRING** (`"1"`), `active`/`fallback` booleans, `sortOrder` int |
| `supportedlocales/active` | same 2 rows (both active), ascending `sortOrder` |
| `supportedlocales/fallback` | **single object** (not array) = the `en` row; 200 (seed has a fallback) |

What each test now guarantees (pass ⇒ behavior, not just reachability):
- **exact deep-equality** for the two compiled-in lists (any drift/reorder fails);
- **strict types** on locale rows (`id` string, `active`/`fallback` boolean,
  `sortOrder` int, exactly the 6 keys) — catches Go serialization drift;
- **cross-view invariants**: `/active` = the active subset of the full list
  (same ids, each row serialized identically), non-decreasing `sortOrder`;
  `/fallback` = the full list's single fallback row as an object (or 404 if none).

Remaining at migrate time: add `a2-static-reads.spec.ts` to the `go-parity`
testMatch so these 5 also run against the Go port. Path already fixed to
`supportedlocales` (no trailing slash).

---

## 3-old. What the e2e checked BEFORE the rewrite  ⚠️ (near nothing)

Files: `tests/readonly/03-type-a.spec.ts` (a2 loop) + `playwright.config.ts`
(`go-parity` project) + `fixtures/contract.ts`.

**Per-endpoint, the entire assertion today is:**
```
expect(res.status()).toBe(200);
expect(isAuthedResponse(status, body)).toBe(true);   // body does NOT start with "<!DOCTYPE html"
```
That's it. **No JSON parse, no keys, no values, no count, no ordering, no
content-type.** This is exactly the "checking 200 is meaningless" problem we
already fixed for a1 — the a1 pilot has a deep shape + IANA-timezone assertion,
but the a2 siblings are shallow 200-only.

Concrete gaps for the 5:

| Endpoint | In `03-type-a` TYPE_A list? | What's asserted | Runs vs Go? |
|----------|-----------------------------|-----------------|-------------|
| `math-functions` | yes | 200 + not-login only | **no** |
| `sample-item-status-types` | yes | 200 + not-login only | **no** |
| `supportedlocales` (real, no slash) | **NO** — list has `supportedlocales/` (slash) which **404s** on Java | the real path is untested; the listed one is a failing test | **no** |
| `supportedlocales/active` | yes | 200 + not-login only | **no** |
| `supportedlocales/fallback` | yes | 200 + not-login only | **no** |

Plus:
- **`go-parity` project greps `/server-time/`** — so **none of these 5 run
  against Go** at all. Zero Go parity coverage today.
- **`90-endpoint-auth-coverage.spec.ts`** hits the param-less ones too, but only
  for the same generic check (anon→blocked, authed→app-response) — reachability,
  not shape.

**Bottom line:** for the 5 a2 endpoints the suite currently proves only "Java
returns 200 and not the login page" (and even that is broken for
`supportedlocales/` due to the trailing slash). There is **no behavioral parity
assertion** and **no Go-side check** yet.

---

## 4. What a real a2 parity gate needs (to add when we migrate — NOT yet)

Behavioral assertions (deep, like a1's), each an exact/near-exact contract:

- `math-functions`: JSON array of exactly **14** `{id,value}`; assert the full id
  set (`+ - / * ( ) == != >= <= IS_IN_NORMAL_RANGE IS_OUTSIDE_NORMAL_RANGE && ||`)
  and their values. Fixed contract → exact equality.
- `sample-item-status-types`: exactly `[{"","All"},{"active","Active"},
  {"disposed","Disposed"}]`. Exact equality.
- `supportedlocales`: array of DTOs with keys `{id, localeCode, displayName,
  active, fallback, sortOrder}`; assert the key set + expected locale codes.
- `supportedlocales/active`: subset where `active===true`, **assert
  `sortOrder` is ascending** (the one ordering behavior).
- `supportedlocales/fallback`: **single object** (not array) with
  `fallback===true`, or 404 — assert the shape, not just 200.

Harness changes:
- Fix the path: `supportedlocales/` → `supportedlocales` (no slash).
- Extend `go-parity` grep from `/server-time/` to also match the 5 once ported
  (or split a dedicated grep list), so the SAME assertions run Java + Go.
- (e2e edits are the **e2e track** — fork from `develop`, prefix `e2e`, and
  ask first per `branch-naming.md`.)

---

## 5. Planned commit groups (when implementing — NOT yet)

1. this plan doc + `branch-naming.md` a2 scope update.
2. static handlers: `math-functions`, `sample-item-status-types`.
3. Postgres wiring + `supported_locale` DAO/service/controller (the 3 locale reads).
4. deep parity assertions + `go-parity` extension + trailing-slash fix (e2e track).

No code is written until the maintainer creates `migration/a2-static-reads`.
