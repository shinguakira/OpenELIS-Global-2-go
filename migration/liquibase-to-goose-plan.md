# Liquibase → Goose Migration Plan

Status: **tooling done and verified; cutover still deferred until Java is fully
retired.** During coexistence, Java/Liquibase remains the single schema owner —
Go does NOT touch schema in production. What changed this round: the extraction
→ split → goose-file → embed pipeline described below is no longer a plan, it's
real, committed code (`migration/scripts/`, `migration/openelis-go/db/`,
`migration/openelis-go/cmd/migrate`), run end-to-end and verified against a
from-scratch database. See § 9 for exact results. Branch:
`migration/db-liquibase-to-goose` (forked from `migration-base`, per
[branch-naming.md](branch-naming.md)).

**Everything below was rewritten against what actually happened, not the
original plan** — the first draft of this doc got several load-bearing facts
wrong (see § 9.1). If you're re-running this extraction later (schema changed,
need to regenerate), trust this version and the two scripts, not your memory of
the old one.

---

## 1. Inventory (corrected — see § 9.1 for what the old numbers got wrong)

| Item                                                            | Value                                                                                                                      |
| --------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| Liquibase `<changeSet>` elements actually emitted by extraction | **993**                                                                                                                    |
| Root changelog                                                  | `src/main/resources/liquibase/base-changelog.xml`                                                                          |
| Changelog directories included                                  | `2.0.x.x` … `3.5.x.x`, `3.4.14.x`, `analyzer/`, `qc/` (17 directories, not just `3.5.x.x`)                                 |
| Schema                                                          | `clinlims` (PostgreSQL 14+)                                                                                                |
| Files using `<preConditions>`                                   | **218** (mostly `onFail="MARK_RAN"`, 215 files) — see § 7                                                                  |
| Files using `<loadData>`                                        | **0** — not used anywhere in this changelog                                                                                |
| Files using `<sqlFile>`                                         | 1                                                                                                                          |
| Files using `<modifySql>`                                       | 0                                                                                                                          |
| Pre-Liquibase baseline schema                                   | `db/dbInit/OpenELIS-Global.sql` — a `pg_dump` (PG 9.5.19), **not** part of the Liquibase changelog tree at all. See § 2.2. |

---

## 2. Extraction approach (what actually ran)

### 2.1 Getting the full ordered SQL out of Liquibase — no Maven, no Docker

The original plan called for `mvn liquibase:updateSQL` against a throwaway
Postgres. **Neither Maven nor Docker was available in the environment this ran
in** (`mvn`/`docker`: command not found). Two things made this a non- issue
instead of a blocker:

1. Liquibase ships a **standalone CLI** (no Maven needed) — downloaded directly
   from GitHub releases, matching the exact version pinned in `pom.xml`
   (`liquibase-core` **4.8.0**):

   ```bash
   curl -sL -o liquibase.zip \
     https://github.com/liquibase/liquibase/releases/download/v4.8.0/liquibase-4.8.0.zip
   unzip liquibase.zip -d liquibase-cli
   ```

2. Liquibase's `offline:postgresql` URL mode generates SQL **without any live
   database connection at all** — better than the original plan's
   throwaway-database approach, not just a workaround for one:
   ```bash
   cd liquibase-cli
   ./liquibase \
     --classpath="<repo>/src/main/resources" \
     --changelog-file="liquibase/base-changelog.xml" \
     --output-file="<out>/liquibase-full.sql" \
     updateSQL --url="offline:postgresql"
   ```
   `--classpath` must point at `src/main/resources` (not the `liquibase/`
   subdirectory) — `base-changelog.xml`'s
   `<include file="liquibase/2.0.x.x/ base.xml" />` paths are
   classpath-relative.

Output format: each changeset is preceded by a marker line of the exact form

```
-- Changeset <path>::<id>::<author>
```

(capital "Changeset" — not `-- changeset author:id` as the original plan
assumed). 993 of these markers appear in the output.

One unresolved changelog property: `validation_site_information.xml::1` contains
a literal `'${now}'` that offline mode never substitutes (Liquibase normally
resolves `${now}` via a live DB connection) — `split_liquibase_sql.py` rewrites
it to `now()`.

### 2.2 The baseline that Liquibase's own changelog assumes — the big gap in the original plan

The original plan asserted running `updateSQL` against an empty database "is the
complete schema build from nothing." **That's wrong.** Confirmed by running
migration 0001 against a truly empty database:

```
pq: relation "clinlims.login_user" does not exist
```

`2.0.x.x/convert_id_types.xml` _modifies_ `login_user` — it assumes the table
already exists. The Liquibase changelog tree here has **never** been a
from-scratch schema; it's the incremental history on top of a pre-Liquibase
baseline: `db/dbInit/OpenELIS-Global.sql`, a `pg_dump` (PostgreSQL 9.5.19) that
the Docker database image (`itechuw/openelis-global-2-database`) and the install
scripts already load before Liquibase ever runs. This isn't a new gap introduced
by this conversion — Liquibase itself has always depended on this same baseline
in production. Goose migrations built from this changelog inherit the identical
precondition: **apply only to a database that already has
`db/dbInit/OpenELIS-Global.sql` loaded.**

Loading that baseline for verification needed its own tooling
(`migration/scripts/load_baseline_dump.py`) because the dump uses `pg_dump`'s
`COPY ... FROM stdin; <tab-data> \.` bulk-load format for seed rows — a libpq
streaming-protocol construct, not something a plain string `Exec()` can run. The
script splits the dump into plain-SQL chunks and COPY chunks and drives each
through the right `psycopg2` API (`execute()` vs. `copy_expert()`). One
additional wrinkle it handles: some tables in this dump have their `COPY` block
deliberately commented out with `/* ... */` (data stripped for size) — a naive
"is this chunk non-empty" check has to strip block comments too, or it tries to
execute pure-comment chunks and `psycopg2` raises
`can't execute an empty query`.

### 2.3 Splitting into goose files

`migration/scripts/split_liquibase_sql.py` splits on the `-- Changeset` marker
and writes one file per changeset to
`migration/openelis-go/db/migrations/<seq>_<slug>.sql`, zero-padded 4-digit
sequence + `<source-file-stem>_<changeset-id>` slug. Both the source file stem
and the changeset id are needed in the slug — many changesets share a source
file (e.g. `new_tests.xml` alone contributes over 100).

Run it:

```bash
python3 migration/scripts/split_liquibase_sql.py <liquibase-full.sql> migration/openelis-go/db/migrations
```

It also writes `migration/openelis-go/db/MIGRATION_MANIFEST.tsv` — one row per
migration file recording its Liquibase source, whether it needed
`NO TRANSACTION`, whether a `Down` block was auto-generated, and whether an
idempotency guard was applied (see § 3, § 7).

---

## 3. Goose file format (as actually generated)

Every changeset body is wrapped in `-- +goose StatementBegin` /
`-- +goose StatementEnd`, **always**, not just when a `DO $$ ... $$` block is
detected. Reason, confirmed empirically: goose's own SQL-migration parser splits
a plain (unwrapped) file on `;` — which breaks any changeset containing a
dollar-quoted block (`DO $$ DECLARE ... END $$;`) into fragments, producing
`pq: unterminated dollar-quoted string`. Postgres itself has no problem
executing a multi-statement string with dollar-quoting in one `Exec()` call
(simple query protocol) — only goose's own internal splitting needed
suppressing, which `StatementBegin`/`StatementEnd` does by telling goose "send
this whole span as one opaque unit."

```sql
-- source: liquibase liquibase/2.1.x.x/external_connections.xml::4::csteele
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.basic_authentication_data (id INTEGER NOT NULL, external_connection_id INTEGER, username VARCHAR(255), password VARCHAR(255), last_updated date, CONSTRAINT basic_authentication_data_pkey PRIMARY KEY (id), CONSTRAINT fk_basic_authentication_data_external_connection FOREIGN KEY (external_connection_id) REFERENCES external_connection(id));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS clinlims.basic_authentication_data;
-- +goose StatementEnd
```

### 3.1 NO TRANSACTION

`split_liquibase_sql.py` scans for `CREATE INDEX CONCURRENTLY`,
`DROP INDEX CONCURRENTLY`, and `ALTER TYPE ... ADD VALUE` and prepends
`-- +goose NO TRANSACTION` when found. Confirmed count for this changelog:
**zero** — none of these patterns appear anywhere in it. The original plan
flagged this as a risk item; empirically it's a no-op for this codebase (but the
detection stays in the script since it costs nothing and a future changeset
could use one).

### 3.2 Down blocks — best-effort, not complete

`infer_down()` generates a `Down` block only for changesets that are a
**single** statement matching one of five unambiguous shapes
(`CREATE TABLE`/`SEQUENCE`/`INDEX`/`VIEW`, single-column `ADD COLUMN`) — 130
of 993. Everything else (863) gets an honest `-- TODO` comment, not a guess.
Liquibase's own `future-rollback-sql` was tried first as a way to get real
rollback SQL — it produced an empty file (no `<rollback>` elements are defined
anywhere in this changelog, and offline mode can't auto-infer rollbacks that
need live DB state) — so there was no shortcut available here; the `-- TODO`
approach is the honest one.

---

## 4. Goose file layout in the repo (as shipped)

```
migration/openelis-go/
    db/
        embed.go                  -- package db; //go:embed migrations/*.sql
        MIGRATION_MANIFEST.tsv     -- one row per migration: source, guards, down
        migrations/
            0001_convert_id_types_1.sql
            0002_convert_id_types_2.sql
            ...
            0993_009_nullable_unit_of_measure_qc_009_nullable_unit_of_measure.sql
    cmd/
        migrate/main.go            -- standalone opt-in goose CLI (§ 5)
        _dbsetup/main.go           -- throwaway scratch-DB helper for verification
                                       (underscore prefix -> go build/vet ./... skip it)
```

`db/embed.go` is its own tiny package specifically so the `//go:embed`
directive's relative path works out to `migrations/*.sql` — `go:embed` paths
resolve relative to the `.go` file's own directory, so putting the directive
directly in `cmd/migrate/main.go` would have looked for
`cmd/migrate/db/migrations/*.sql`, which doesn't exist. `cmd/migrate` imports
`openelis-go/db` and uses its exported `db.Migrations` FS instead.

---

## 5. Go embed + CLI (as shipped, not the sketch)

`migration/openelis-go/cmd/migrate` is a **standalone, opt-in** CLI —
`cmd/openelis` (the main server) does not import it and nothing runs it
automatically. That's deliberate: per this doc's own status line, Go does not
touch schema during coexistence, so the tool that CAN touch schema stays outside
the server's own startup path.

```go
// cmd/migrate/main.go (trimmed)
conn, _ := sql.Open("postgres", *dsn)
conn.SetMaxOpenConns(1)                                 // see note below
conn.Exec("SET search_path TO clinlims, public")        // see note below
goose.SetBaseFS(db.Migrations)
goose.SetDialect("postgres")
goose.RunWithOptions(command, conn, "migrations", cmdArgs)  // "migrations", not "db/migrations" — see § 4
```

Usage:

```bash
go run ./cmd/migrate -dsn "postgres://user:pass@host:port/db?sslmode=disable" <goose-command> [args]
# e.g.: status, up, up-by-one, up-to <N>, down, validate
```

**Why `SET search_path` + `SetMaxOpenConns(1)`:** Liquibase's own extracted SQL
leaves plenty of table references unqualified (e.g. a FK to `data_export_task`,
not `clinlims.data_export_task`) — confirmed by running against a clean DB:
`pq: relation "data_export_task" does not exist` unless `search_path` resolves
`clinlims` first. In production this works _implicitly_, because Java connects
**as** the `clinlims` role and Postgres's default `search_path`
(`"$user", public`) resolves `"$user"` to a same-named schema — `clinlims`
happens to be both the role and the schema name (confirmed: `SHOW search_path`
as `postgres` on the real `clinlims` DB returns nothing role-specific; the role
literally named `clinlims` gets the schema access for free via the default
`"$user"` rule, no explicit config anywhere). Rather than depend on which role
`-dsn` happens to authenticate as, `cmd/migrate` sets it explicitly.
`SetMaxOpenConns(1)` is required for that `SET` to actually stick:
`database/sql` pools connections, and a `SET` issued on one pooled connection
silently doesn't apply to whichever _different_ physical connection goose's next
query happens to borrow. This is a one-shot CLI, not a server, so pinning to a
single connection for the whole run has no real cost.

`go.mod`: `github.com/pressly/goose/v3 v3.24.1`, direct require (confirmed via
`go mod tidy`, not hand-edited). `go.sum` gained checksum-only entries for
`modernc.org/sqlite`, `github.com/dustin/go-humanize`, and a few other
`modernc.org/*` packages — confirmed via `go mod why modernc.org/sqlite` these
come **only** through `github.com/pressly/goose/v3.test` (goose's own test suite
exercises a sqlite dialect internally); nothing in openelis-go imports or uses
sqlite. Same situation as the GORM `gorm.io/gorm.test` transitive entries from
the b1 migration — a checksum-ledger artifact of `go.sum` covering the full
module graph, not a real dependency.

---

## 6. Cutover procedure — still deferred, unchanged from the original plan

These steps still run **after Java is retired** (WAR shut down, no traffic).
Nothing in this round changes this section's content or its premise — it's
reproduced here so the whole procedure lives in one document.

1. **Freeze Liquibase** — remove the Liquibase Maven plugin from `pom.xml` or
   comment out its `<phase>` binding so it cannot accidentally run.

2. **Schema parity check** — dump both the Liquibase-managed schema and the
   goose-managed schema (baseline + all migrations applied to a clean DB) and
   diff them:

   ```bash
   pg_dump --schema-only clinlims > liquibase.sql
   pg_dump --schema-only clinlims_goose > goose.sql
   diff liquibase.sql goose.sql
   ```

   Diff must be empty (modulo `goose_db_version` / `databasechangelog`). **Given
   § 9's 26 known-open items, this parity check is expected to fail until those
   are resolved — see § 9.2.**

3. **Baseline production** — production already carries the full schema
   (Liquibase-managed), so goose must record every migration as applied
   **without executing any of it**.

   Goose has no built-in baseline command (`up`, `down`, `status`, `version`,
   `up-to <VERSION>` are the real subcommands — there is no `up-to-date`). The
   supported way is to let goose create its version table, then seed it
   directly.

   ```bash
   # a. let goose create goose_db_version without running anything.
   goose -dir db/migrations postgres "$DSN" up-to 0

   # b. seed one row per migration file, marking each applied.
   #    version_id is the numeric prefix of each filename.
   psql "$DSN" <<'SQL'
   INSERT INTO goose_db_version (version_id, is_applied, tstamp)
   SELECT v, true, now()
   FROM generate_series(1, 993) AS v
   ON CONFLICT DO NOTHING;
   SQL
   ```

   993, not 277 — see § 9.1. **Verify this on a restored production snapshot
   before touching production.**

4. **Verify** — `goose -dir db/migrations postgres "$DSN" status` must show
   every migration as `Applied`, nothing `Pending`. Confirm the app starts and
   `goose up` is a no-op.

5. **Future schema changes** — add a new `0994_*.sql` goose file, commit,
   deploy. From this point goose behaves normally; the baseline seed is a
   one-time operation.

---

## 7. Risk items — corrected against what actually happened

| Risk                     | Original plan said    | What's actually true                                                                       |
| ------------------------ | --------------------- | ------------------------------------------------------------------------------------------ |
| `<loadData>` CSV inserts | "~10–15 changesets"   | **0** files use `<loadData>` anywhere in this changelog                                    |
| `<preConditions>`        | same "~10–15" bucket  | **218 files** (215 with `onFail="MARK_RAN"`) — the dominant risk category, not a minor one |
| `<sqlFile>` includes     | not mentioned as rare | 1 file                                                                                     |
| `modifySql`              | listed as a risk      | 0 files — not used                                                                         |
| NO TRANSACTION DDL       | listed as a risk      | 0 occurrences (§ 3.1)                                                                      |

**The real risk, confirmed by actually running the result against a clean
database, is `onFail="MARK_RAN"` preConditions.** Liquibase evaluates these
against live DB state at apply time and silently skips the changeset if the
precondition fails (e.g. "only insert this reference row if it doesn't already
exist"). Offline-mode extraction has no DB to check preconditions against, so it
can't reproduce that skip — it just always emits the statement. Running the raw
extraction against a clean DB hit real
`duplicate key value violates unique constraint` errors on ordinary `INSERT`s
almost immediately.

**Fix applied, not just documented:** `split_liquibase_sql.py` rewrites the
safely-rewritable statement shapes to be idempotent by construction —
`CREATE TABLE/SEQUENCE/INDEX` → add `IF NOT EXISTS`,
`ALTER TABLE ... ADD [COLUMN]` → add `IF NOT EXISTS`, plain `INSERT` → append
`ON CONFLICT DO NOTHING`, and the DROP-side mirror
(`DROP TABLE/SEQUENCE/INDEX/VIEW`, `ALTER TABLE ... DROP COLUMN` → add
`IF EXISTS`, needed because a from- scratch sequence can also hit a changeset
dropping something an earlier changeset never created). This is applied
**per-statement** within a changeset (a dollar-quote/string-literal-aware
splitter, `split_top_level_statements`), not just to whole-body single-statement
changesets — most changesets in this changelog are multi-statement.

Two real regexp bugs surfaced and got fixed while building this (both confirmed
via the actual Postgres error, not guessed):

- A naive `\s+(?!IF NOT EXISTS\b)` lookahead lets the regex engine **backtrack**
  `\s+` down to consuming only one space when the source SQL has irregular
  double-spacing (`CREATE SEQUENCE  IF NOT EXISTS ...`, which occurs in this
  changelog's own generated SQL), producing double- guarded output
  (`CREATE SEQUENCE IF NOT EXISTS  IF NOT EXISTS ...`). Fixed by never combining
  a variable-length `\s+` with an immediately adjacent lookahead — match the
  keyword prefix and check "already guarded" as a fully separate regex against
  whatever follows.
- `ALTER TABLE ... ADD` is ambiguous (`ADD COLUMN` vs. `ADD CONSTRAINT` /
  `PRIMARY KEY` / `UNIQUE` / `FOREIGN KEY` / `CHECK` / `EXCLUDE`) and only
  `ADD [COLUMN]` accepts `IF NOT EXISTS` in Postgres — a first pass blindly
  inserted it after every bare `ADD`, producing
  `ALTER TABLE ... ADD IF NOT EXISTS CONSTRAINT ...` (syntax error).

What's genuinely **not** mechanically fixable, confirmed by the full
verification run — see § 9.2 for the itemized list (26 changesets): plain
`ON CONFLICT DO NOTHING` guards a statement's own conflicting write, but the
remaining failures happen one level removed — a subquery like
`(SELECT id FROM system_role WHERE name = 'Reception')` used inside an
`INSERT`'s `VALUES` returns 2+ rows (because an earlier un-guarded changeset
already duplicated that reference row) or 0 rows, and Postgres raises the error
while evaluating the subquery expression, **before** conflict resolution is ever
reached. `ON CONFLICT` cannot help; fixing this means either tracing back which
earlier changeset(s) produced the duplicate, or making the subquery itself
defensive (`LIMIT 1`) — a judgment call about which of several historical
duplicate rows is "correct" that needs a human, not a regex.

---

## 8. Timeline estimate — superseded by § 9

The original estimate (~2 weeks, itemized by phase) assumed manual extraction
and ~15 hand-conversions. That's no longer the right way to think about
remaining effort — § 9 gives the actual, current, itemized remainder (26
changesets, one well-understood root cause) instead of a phase estimate.

---

## 9. What actually happened this round (verification results)

### 9.1 Corrections to the original draft of this doc

The original draft of this plan was written without running any of it. Once
actually executed, several of its load-bearing facts turned out wrong:

| Claim in the original draft                                                             | Reality                                                                                                                                                                            |
| --------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 277 changesets total                                                                    | **993** — the draft undercounted by ~3.6x (likely counted only `3.5.x.x/*.xml`, missing `2.0.x.x` through `3.4.14.x`, `analyzer/`, `qc/`)                                          |
| Root changelog is `liquibase/liquibase-db.xml`                                          | Doesn't exist. Real file: `liquibase/base-changelog.xml`                                                                                                                           |
| `updateSQL` against an empty DB gives "the complete schema build from nothing"          | False — the changelog assumes the `db/dbInit/OpenELIS-Global.sql` baseline already exists (§ 2.2); it is not a from-scratch build                                                  |
| `<loadData>`/`<preConditions>` together are "~10–15 changesets" needing hand conversion | `<loadData>`: 0 files. `<preConditions>`: **218 files** — this was the dominant risk, not a minor one                                                                              |
| Extraction requires Maven + a live throwaway Postgres                                   | Neither was available in this environment, and neither turned out to be necessary — the standalone Liquibase CLI's `offline:postgresql` mode needs no DB connection at all (§ 2.1) |

### 9.2 End-to-end verification: 967 / 993 apply cleanly from scratch

Verification method: build the standalone `migrate` CLI, create a scratch
Postgres database, load `db/dbInit/OpenELIS-Global.sql` (§ 2.2) into it, then
run every migration in sequence via `up-by-one` in a loop (this diagnostic
driver force-marks a failing version as applied-without-running so the loop can
continue and surface every distinct failure in one pass, instead of stopping at
the first one — this is a scan for reporting purposes, not a real apply; **the
26 items below did not actually run**).

**Result: 967 of 993 (97.4%) applied with a real, successful `Exec()` against a
genuinely empty-then-baselined database.**

The remaining 26 are a single root cause — the subquery-cardinality problem from
§ 7's last paragraph — surfacing at these versions:

| Version(s)             | File(s)                                                                    | Table                                                                                                                                       |
| ---------------------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| 0216, 0261, 0264, 0544 | `add_menu_study_electronic_order*`, `add_tb_menu*`, `generic_sample_menu*` | `system_module_url`                                                                                                                         |
| 0234–0249 (12 files)   | `fix_database_bugs_retroc_*`                                               | `system_role_module`                                                                                                                        |
| 0254, 0255             | `add_admin_default_roles_1/2`                                              | `user_lab_unit_roles`, `system_user_role`                                                                                                   |
| 0666, 0884             | `eqa_007_add_eqa_menu_items_*`, `shipment_007_add_menu_and_role_*`         | `system_role_module`                                                                                                                        |
| 0740–0742              | `007_insert_seed_data_*` (query timeout/rate-limit/max-fields config)      | `system_configuration` (table itself not found at this point — likely a further cascade from an earlier un-fixed item; not yet root-caused) |

All 26 are recorded in `migration/openelis-go/db/MIGRATION_MANIFEST.tsv`
(`idempotency_guarded=False`) for anyone picking this up. Fixing them for real
requires per-changeset tracing of the specific historical duplication (which
earlier changeset(s) added extra `system_role`/`system_module`/ `system_user`
rows with the same `name`) — exactly the kind of manual reconciliation § 7
already predicted was necessary and mechanical guards can't safely do (a regex
can't judge which of several duplicate historical rows is the "correct" FK
target).

### 9.3 Bugs found and fixed along the way (all confirmed via real error output, not guessed)

1. `db.Raw()`-only Maven/Docker extraction path — unavailable in this
   environment; replaced with the standalone-CLI offline-mode path (§ 2.1).
2. Baseline assumption wrong (§ 2.2) — found by literally running migration 0001
   against an empty DB and reading the error.
3. `pg_dump` `COPY ... FROM stdin` blocks needed dedicated handling — a plain
   string `Exec()` cannot run them; `load_baseline_dump.py` was written to split
   and drive them through `psycopg2.copy_expert()`.
4. Comment-only chunks (including commented-out
   `/*COPY ... FROM stdin; ... \. */` blocks for tables whose seed data was
   stripped from the baseline dump) made `psycopg2` raise
   `can't execute an empty query` — fixed by stripping both `--` and `/* */`
   comments before deciding a chunk has nothing to execute.
5. goose's own semicolon-splitting breaks `DO $$ ... $$` blocks — fixed with
   universal `StatementBegin`/`StatementEnd` wrapping (§ 3).
6. `search_path` — Liquibase's extracted SQL has unqualified table refs that
   only resolve because production connects as the `clinlims` role (§ 5) — fixed
   by setting it explicitly in `cmd/migrate`, plus the `database/sql`
   connection-pooling gotcha that comes with that fix.
7. `onFail="MARK_RAN"` preconditions can't survive offline extraction (§ 7) —
   fixed for the mechanically-safe subset with per-statement idempotency guards;
   the regex backtracking bug and the `ADD COLUMN` vs. `ADD CONSTRAINT`
   ambiguity bug (§ 7) were both found by reading the actual Postgres syntax
   error, not anticipated in advance.
8. DROP-side mirror of the same idempotency-guard idea, needed once the
   full-scan diagnostic surfaced `DROP TABLE`/`ALTER TABLE ... DROP COLUMN` on
   objects a from-scratch sequence never created.

---

## 10. Reproducing this

```bash
# 1. extract (no Maven/Docker needed)
cd liquibase-cli   # standalone CLI, downloaded per § 2.1
./liquibase --classpath="<repo>/src/main/resources" \
  --changelog-file="liquibase/base-changelog.xml" \
  --output-file="<out>/liquibase-full.sql" \
  updateSQL --url="offline:postgresql"

# 2. split into goose files
python3 migration/scripts/split_liquibase_sql.py <out>/liquibase-full.sql migration/openelis-go/db/migrations

# 3. build the CLI
cd migration/openelis-go && go build ./...

# 4. verify against a scratch DB (needs a reachable Postgres server + superuser creds)
go run ./cmd/_dbsetup    # edit the DSN inside if your dev DB differs from localhost:15432 postgres/admin
python3 ../scripts/load_baseline_dump.py ../../db/dbInit/OpenELIS-Global.sql "postgres://postgres:admin@localhost:15432/clinlims_goose_verify"
go run ./cmd/migrate -dsn "postgres://postgres:admin@localhost:15432/clinlims_goose_verify?sslmode=disable" up
```
