# Liquibase → Goose Migration Plan

Status: **deferred — execute only after Java is fully retired**  
During coexistence Java/Liquibase is the single schema owner. Go does NOT touch schema.

---

## 1. Inventory

| Item | Count |
|---|---|
| Liquibase XML changeset files | **277** |
| Latest changeset file | `uuid_columns.xml`, `validation_site_information.xml`, `update_roles.xml` |
| Schema | `clinlims` (PostgreSQL 14+) |
| Root changelog | `src/main/resources/liquibase/liquibase-db.xml` |

Changesets live under:
```
src/main/resources/liquibase/
    liquibase-db.xml          ← master include list
    common/
    3.5.x.x/
        001-initial.xml
        002-...xml
        ...
        039-test-method-links.xml   ← latest at time of writing
```

---

## 2. Extraction approach

### Step 1 — get the full ordered SQL out of Liquibase

> **Critical: run this against an EMPTY database, not production.**
> `liquibase:updateSQL` consults `DATABASECHANGELOG` and emits SQL only for
> changesets that are still **pending** for the target database. Production is
> fully migrated, so every changeset is already recorded there and the command
> would output essentially nothing. Only a clean database treats all 277
> changesets as pending and therefore emits the complete schema history.

```bash
# 1. create a throwaway empty database
createdb -h localhost -p 15432 -U postgres clinlims_extract

# 2. generate the full history against it
mvn liquibase:updateSQL \
  -Dliquibase.url="jdbc:postgresql://localhost:15432/clinlims_extract" \
  -Dliquibase.username=postgres \
  -Dliquibase.password=admin \
  -Dliquibase.outputFile=target/liquibase-full.sql
```

The output is a single SQL file containing every changeset in execution order,
each preceded by a `-- changeset author:id` comment. Because the target DB was
empty, this is the complete schema build from nothing.

Sanity check before proceeding — the file should contain 277 changeset markers:
```bash
grep -c '^-- changeset' target/liquibase-full.sql
```
If this returns 0 or a small number, you ran it against a database that already
has the schema. Drop and recreate the extract DB and retry.

### Step 2 — split into numbered goose files

Each Liquibase changeset becomes one goose file:

```
db/migrations/
    0001_initial_schema.sql
    0002_add_test_section.sql
    ...
    0277_uuid_columns.sql
```

Naming rule: zero-padded 4-digit sequence + snake_case of the changeset id or filename.

A script can split `target/liquibase-full.sql` on the `--changeset` comment
boundaries and write each block to the correct file. Example (Python):

```python
import re, pathlib

src = pathlib.Path("target/liquibase-full.sql").read_text()
# Keep the changeset id from each marker so filenames stay traceable.
chunks = re.split(r"^--\s*changeset\s+(\S+)", src, flags=re.M)
out = pathlib.Path("migration/openelis-go/db/migrations")
out.mkdir(parents=True, exist_ok=True)

# chunks = [preamble, id1, sql1, id2, sql2, ...]
pairs = list(zip(chunks[1::2], chunks[2::2]))
for i, (cs_id, sql) in enumerate(pairs, 1):
    slug = re.sub(r"[^a-z0-9]+", "_", cs_id.lower()).strip("_")
    (out / f"{i:04d}_{slug}.sql").write_text(
        f"-- +goose Up\n-- from liquibase changeset: {cs_id}\n{sql.strip()}\n\n"
        f"-- +goose Down\n-- TODO: write rollback\n"
    )
print(f"wrote {len(pairs)} migrations")
```

The printed count must be 277. Anything less means the extract was incomplete —
go back to Step 1.

---

## 3. Goose file format

```sql
-- +goose Up
CREATE TABLE clinlims.dictionary_category (
    id          BIGSERIAL PRIMARY KEY,
    description VARCHAR(255),
    local_abbrev VARCHAR(20),
    name        VARCHAR(255),
    lastupdated TIMESTAMP
);

-- +goose Down
DROP TABLE clinlims.dictionary_category;
```

### Special cases

| Situation | Directive |
|---|---|
| DDL that cannot run inside a transaction (`CREATE INDEX CONCURRENTLY`, `ALTER TYPE … ADD VALUE`) | Add `-- +goose NO TRANSACTION` before `-- +goose Up` |
| Rollback is destructive / impossible | Leave `-- +goose Down` body empty with a comment explaining why |
| Data migration (INSERT/UPDATE) | Write a matching DELETE/UPDATE in `-- +goose Down` |

---

## 4. Goose file layout in the repo

```
migration/openelis-go/
    db/
        migrations/
            0001_initial_schema.sql
            0002_add_test_section.sql
            ...
            0277_uuid_columns.sql
```

---

## 5. Go embed pattern

Embed the migration files into the binary so they ship without a separate
directory. Run them at startup or via a `--migrate` CLI flag.

```go
package main

import (
    "embed"
    "database/sql"

    "github.com/pressly/goose/v3"
)

//go:embed db/migrations/*.sql
var migrations embed.FS

func runMigrations(db *sql.DB) error {
    goose.SetBaseFS(migrations)
    return goose.Up(db, "db/migrations")
}
```

`goose.Up` applies all pending migrations in sequence and records them in
the `goose_db_version` table it creates automatically.

Install goose:
```bash
go get github.com/pressly/goose/v3
```

---

## 6. Cutover procedure

These steps run after Java is retired (WAR shut down, no traffic):

1. **Freeze Liquibase** — remove Liquibase Maven plugin from `pom.xml` or comment
   out the `<phase>` binding so it cannot accidentally run.

2. **Schema parity check** — dump both the Liquibase-managed schema and the
   goose-managed schema (applied against a clean DB) and diff them:
   ```bash
   pg_dump --schema-only clinlims > liquibase.sql
   pg_dump --schema-only clinlims_goose > goose.sql
   diff liquibase.sql goose.sql
   ```
   Diff must be empty (modulo the `goose_db_version` / `databasechangelog` tables).

3. **Baseline production** — production already carries the full schema, so goose
   must record every migration as applied **without executing any of it**.

   Goose has no built-in baseline command (`up`, `down`, `status`, `version`,
   `up-to <VERSION>` are the real subcommands — there is no `up-to-date`). The
   supported way is to let goose create its version table, then seed it directly.

   ```bash
   # a. let goose create goose_db_version without running anything.
   #    `up-to 0` applies nothing but initialises the table.
   goose -dir db/migrations postgres "$DSN" up-to 0

   # b. seed one row per migration file, marking each applied.
   #    version_id is the numeric prefix of each filename.
   psql "$DSN" <<'SQL'
   INSERT INTO goose_db_version (version_id, is_applied, tstamp)
   SELECT v, true, now()
   FROM generate_series(1, 277) AS v
   ON CONFLICT DO NOTHING;
   SQL
   ```

   Adjust the upper bound (`277`) to the actual migration count. If goose's
   version table already contains row `0` from step (a), the `ON CONFLICT` guard
   keeps the insert idempotent.

   **Verify this on a restored production snapshot before touching production.**

4. **Verify** — `goose -dir db/migrations postgres "$DSN" status` must show every
   migration as `Applied` and list nothing as `Pending`. Then confirm the app
   starts and `goose up` is a no-op.

5. **Future schema changes** — add a new `0278_*.sql` goose file, commit, deploy.
   The app applies it at startup (or you run the CLI manually before restart).
   From this point goose behaves normally; the baseline is a one-time operation.

---

## 7. Risk items

These Liquibase features have no direct SQL equivalent and need hand-conversion:

| Risk | Detail | Action |
|---|---|---|
| `<loadData>` CSV inserts | Liquibase can bulk-insert from a CSV file | Replace with `INSERT INTO … VALUES (…)` statements in the goose Up block |
| `<preconditions>` | Liquibase skips a changeset based on DB state (e.g. `<tableExists>`) | Convert to a `DO $$ BEGIN IF NOT EXISTS … END $$` PL/pgSQL guard in the goose file |
| `<sqlFile>` includes | External `.sql` files referenced from XML | Inline the SQL content directly into the goose file |
| `onFail="MARK_RAN"` preconditions | Liquibase silently marks a changeset applied when its precondition fails | goose has no equivalent; evaluate per case — usually the guard `IF NOT EXISTS` suffices |
| XML `modifySql` | Liquibase can patch generated SQL per DB dialect | Unlikely in a Postgres-only project; check and inline any actual modifications |

Estimated hand-conversion effort: ~10–15 changesets out of 277 use `<loadData>`
or `<preconditions>`. The rest are standard DDL `<sql>` blocks.

---

## 8. Timeline estimate

| Phase | Effort |
|---|---|
| Script extraction + split (steps 1–2) | 1–2 days |
| Hand-convert risk-item changesets (~15) | 2–3 days |
| Write `-- +goose Down` rollbacks | 3–5 days (skip for DDL-only; focus on data migrations) |
| Schema parity check + fix diffs | 1–2 days |
| Cutover on staging, then production | 1 day |
| **Total** | **~2 weeks** |
