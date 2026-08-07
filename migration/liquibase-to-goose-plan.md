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

```bash
mvn liquibase:updateSQL \
  -Dliquibase.url="jdbc:postgresql://localhost:15432/clinlims" \
  -Dliquibase.username=postgres \
  -Dliquibase.password=admin \
  -Dliquibase.outputFile=target/liquibase-full.sql
```

This produces a single SQL file with every applied changeset in execution order,
wrapped in `--changeset author:id` comments. The changesets that Liquibase has
already applied are marked; unrun ones come after.

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
parts = re.split(r"--\s*changeset\s+\S+", src)
out = pathlib.Path("migration/openelis-go/db/migrations")
out.mkdir(parents=True, exist_ok=True)
for i, sql in enumerate(parts[1:], 1):        # skip preamble
    (out / f"{i:04d}_changeset.sql").write_text(
        f"-- +goose Up\n{sql.strip()}\n\n-- +goose Down\n-- TODO: write rollback\n"
    )
```

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
   Diff must be empty (modulo `goose_db_version` table).

3. **Mark migrations as applied** — on the production DB that already has the
   Liquibase schema, tell goose that all migrations are done without re-running:
   ```bash
   goose -dir db/migrations postgres "DSN" up-to-date
   ```
   This inserts rows into `goose_db_version` for every file without executing the SQL.

4. **Verify** — run `goose status` and confirm all migrations show as `Applied`.

5. **Future schema changes** — add a new `0278_*.sql` goose file, commit, deploy.
   The app applies it at startup (or you run the CLI manually before restart).

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
