# openelis-api-e2e

Black-box, **browserless** API parity suite for the OpenELIS Global 2 backend.
It is a **parity oracle**: the *same* tests assert the Java baseline today and are
pointed, **unchanged**, at a Go (or any) re-implementation later. Coverage plan:
[`../openelis-api-e2e.md`](../openelis-api-e2e.md).

Every test uses Playwright's `request` (`APIRequestContext`) against the live
REST/FHIR surface and asserts **DB state** via a psql oracle. No browser is
launched.

---

## Why it survives the migration (design)

The suite is built so **retargeting at another language is configuration, not a
rewrite**. Three layers keep it language-neutral:

1. **Target = one env var.** Everything hits `OE_BASE_URL`. Point it at the Go
   port and run the identical suite:
   ```bash
   OE_BASE_URL=https://go-host/api/OpenELIS-Global/ npm test
   ```
2. **The contract is frozen, not re-derived.** `fixtures/endpoints.generated.ts`
   (425 endpoints + verb map) and the asserted **JSON shapes / DB effects** ARE
   the cross-language contract the port must satisfy. `tools/extract-endpoints.mjs`
   captures that catalog **from the Java source once** — it is a *baseline-capture*
   tool, **never run against the Go port** (Go has no Java annotations). The
   committed generated file is the source of truth.
3. **Implementation conventions are pluggable** — `fixtures/contract.ts` isolates
   anything a port could legitimately reshape (login path/fields, CSRF token
   location/header, and *how an unauthenticated call is answered*). The Java
   baseline serves a **login HTML page (HTTP 200)** for unauth calls; if the Go
   port returns **401 JSON** instead, flip one flag — no test edits:
   ```bash
   OE_UNAUTH_MODE=status npm test
   ```

**Frozen (the contract the port must reproduce)** vs **configurable (impl detail)**:

| Frozen — asserted as parity | Configurable — `fixtures/contract.ts` / env |
|---|---|
| endpoint paths + verbs (`endpoints.generated.ts`) | login path & form field names |
| response JSON shapes | CSRF token field (`/session.csrf`) + header name |
| DB row effects of writes | unauth convention (`login-html` \| `status`) |
| reference-data / schema invariants | base URL, DB container, docker vs wsl |

The **DB oracle is inherently language-neutral**: it asserts rows in the shared
PostgreSQL `clinlims` schema (migration decision D1 keeps the schema), which both
Java and Go write to — so it validates *either* backend without change.

---

## Prerequisites
- OpenELIS stack running, reachable at `https://localhost/api/OpenELIS-Global/`.
- Demo data loaded: `./src/test/resources/load-test-fixtures.sh --profile=core`
  (and/or `--profile=harness`). DB-oracle baselines assume this.
- Node 18+. `npm install` (no browser download — API mode only).

## Run
```bash
npm install
npm test                 # setup (login) → api-readonly → api-mutating
npm run test:readonly    # read-only + DB baseline + full-surface auth/reachability
npm run test:mutating    # write-path lifecycles (serial, reset DB)

# retarget at a re-implementation
OE_BASE_URL=https://go-host/api/OpenELIS-Global/ npm test
```
Full read-only surface (**451 tests**) runs in **~52 s** (3 workers, browserless).

## Config (env — see `fixtures/env.ts` + `fixtures/contract.ts`)
| Var | Default | Meaning |
|-----|---------|---------|
| `OE_BASE_URL` | `https://localhost/api/OpenELIS-Global/` | target base (**must** end `/`) |
| `OE_TARGET` | `java` | informational label (`java`/`go`) |
| `OE_UNAUTH_MODE` | `login-html` | `status` → assert 401/403 instead of login page |
| `OE_LOGIN_PATH` | `ValidateLogin?apiCall=true` | login endpoint |
| `OE_LOGIN_USER_FIELD` / `OE_LOGIN_PASS_FIELD` | `loginName` / `password` | login form fields |
| `OE_CSRF_FIELD` / `OE_CSRF_HEADER` | `csrf` / `X-CSRF-TOKEN` | CSRF token source & header |
| `OE_ADMIN_USER` / `OE_ADMIN_PASS` | `admin` / `adminADMIN!` | credentials |
| `OE_DB_CONTAINER` | `openelisglobal-database` | Postgres container |
| `OE_DOCKER` | `wsl` | `docker` on Linux/CI; `wsl` wraps via `wsl.exe` here |
| `OE_WSL_DISTRO` | `Ubuntu-24.04` | WSL distro when `OE_DOCKER=wsl` |

## Baseline contracts captured (must hold on the Go port)
- **Auth:** session cookie via `POST /ValidateLogin?apiCall=true` (form login);
  `GET /session` → `{authenticated, userId, loginName, roles[], csrf}`.
- **Unauthenticated protected endpoints return the login HTML page (HTTP 200)** —
  not 401/403. (Configurable via `OE_UNAUTH_MODE` if the port changes this.)
- **CSRF is enforced** on every state-changing verb: `POST/PUT/DELETE/OPTIONS`
  without the `/session.csrf` token → **403** (no `Allow` header exposed).
- **425 endpoints**: 244 read-only (GET), 128 form read+write (GET/POST),
  53 RESTful CRUD — see `ENDPOINT-CATALOG.md`.

## Layout
```
fixtures/
  env.ts                 # base URL, creds, DB access
  contract.ts            # pluggable impl conventions (auth/csrf/unauth mode)
  db.ts                  # psql DB oracle (query/count/assertCount)
  endpoints.generated.ts # FROZEN contract: 425 endpoints + verb map
tests/
  auth.setup.ts          # login → playwright/.auth/admin.json
  readonly/              # crosscutting, session, testcatalog, 90-auth-coverage, 91-db-inventory
  mutating/              # (next) serial, reset DB: order → result → validation → …
tools/
  extract-endpoints.mjs  # Java-baseline catalog capture (NOT run vs the port)
  recheck.mjs            # 5 quick contract re-verification probes
ENDPOINT-CATALOG.md      # endpoint → verbs → manipulation pattern
```

## Status
Read-only surface green (**451 tests**): auth boundary + reachability over every
GET endpoint, session/roles, catalog reads, and the DB oracle. Remaining: the
**mutating** write-path lifecycles (CSRF-authenticated) per `../openelis-api-e2e.md`
§3–§14, mining the Java `*RestControllerTest` classes.
