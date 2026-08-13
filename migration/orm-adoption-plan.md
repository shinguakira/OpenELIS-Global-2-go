# ORM Adoption Plan — GORM + Goose

Status: **GORM rewrite in progress (b1 complete), Goose migration deferred**
Branch: `migration/b1-dictionary-testcatalog` (GORM rewrite applied here first)

---

## Decision

| Concern        | Choice                    | Rationale                                                                    |
| -------------- | ------------------------- | ---------------------------------------------------------------------------- |
| ORM            | **GORM**                  | De facto Go standard (~37k stars, 2013, widest community)                    |
| Migration tool | **Goose**                 | SQL-based like Liquibase, plain `.sql` files, easy conversion                |
| Timing — ORM   | **Now** (b1 rewrite)      | Sets correct pattern before more domains pile up                             |
| Timing — Goose | **After Java is retired** | Both apps share same DB; Liquibase stays the single owner during coexistence |

---

## 1. GORM adoption (current work)

### What changes

| Before                                        | After                                                       |
| --------------------------------------------- | ----------------------------------------------------------- |
| `*sql.DB` in every DAO struct                 | `*gorm.DB` in every b1 DAO struct                           |
| Manual `rows.Next()` / `rows.Scan()` loop     | `db.Raw(...).Scan(&result)`                                 |
| `github.com/lib/pq` only                      | `gorm.io/gorm` + `gorm.io/driver/postgres` added            |
| `internal/common/db.Open()` returns `*sql.DB` | `internal/common/db.OpenGORM()` returns `*gorm.DB`          |
| a2 domains use `*sql.DB` directly             | a2 domains get `*sql.DB` extracted from GORM: `gormDB.DB()` |

### Query strategy

GORM offers two modes. b1 uses `db.Raw().Scan()` for all DAO methods:

```
db.Raw("SELECT id::text AS id, ...").Scan(&results)
```

**Why not `db.Find()` for simple tables?** All OpenELIS primary keys are
`BIGSERIAL` (int64 in PostgreSQL). Java returns them as strings (`getId()` →
`String`). Go valueholders keep `ID string` to match. `db.Find()` generates
`SELECT *` with no `::text` cast — pgx strict typing rejects scanning int64 →
string. `db.Raw()` keeps explicit `::text` casts, maintaining parity.

Future domains that don't need string IDs (or that use UUIDs) can use
`db.Find()` directly. b1 uses `db.Raw().Scan()` everywhere for consistency.

### Lastupdated type change

`DictionaryCategory.Lastupdated` changes from `*int64` (epoch ms) to
`*time.Time`.

- GORM scans `TIMESTAMP` → `*time.Time` naturally
- Controller converts `*time.Time` → `*int64` using `.UnixMilli()` before JSON
  serialization
- JSON output is identical: `"lastupdated": 1712345678000`

### What GORM adds for future write operations

```go
db.Create(&category)           // INSERT
db.Save(&category)             // UPDATE
db.Delete(&category)           // DELETE (soft delete if DeletedAt field present)
db.Transaction(func(tx *gorm.DB) error { ... }) // transaction boundary
```

These calls live in **`daoimpl/` only** — the DAO exposes them as methods
(`Insert`, `Update`, `Delete`). The service layer orchestrates those DAO methods
and owns the business decisions; it never holds a `*gorm.DB`. This is the same
rule the read path follows and it mirrors Java, where `@Transactional` sits on
the service but Hibernate session access stays in the DAO.

---

## 2. Goose migration plan (deferred — after Java is retired)

### When

Only after Java is fully retired and Go owns the schema. During coexistence,
Liquibase (Java) is the single schema owner. Go does not touch schema.

### How — Liquibase XML → Goose SQL

Liquibase stores changesets as XML wrapping plain SQL. Goose uses plain `.sql`
files with `-- +goose Up` / `-- +goose Down` markers.

**Liquibase changeset:**

```xml
<!-- src/main/resources/liquibase/3.5.x.x/001-initial.xml -->
<changeSet id="001" author="openelis">
    <createTable tableName="dictionary_category" schemaName="clinlims">
        <column name="id" type="BIGSERIAL"><constraints primaryKey="true"/></column>
        <column name="description" type="VARCHAR(255)"/>
        <column name="local_abbrev" type="VARCHAR(20)"/>
        <column name="name" type="VARCHAR(255)"/>
        <column name="lastupdated" type="TIMESTAMP"/>
    </createTable>
</changeSet>
```

**Equivalent Goose file:**

```sql
-- db/migrations/001_initial.sql
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

### Conversion approach

Liquibase has 993 changesets across 277 XML files. The SQL inside each changeset
can be extracted directly — it is standard PostgreSQL DDL.

**The extraction, split, and goose-file generation described below are no longer
just a plan — they're done and verified (967/993 apply cleanly against a
from-scratch database).** Full detail, the actual tooling (all Go —
`migration/openelis-go/cmd/splitliquibase`, `cmd/loadbaseline`, `cmd/migrate`),
and the itemized list of what's still open live in
[liquibase-to-goose-plan.md](liquibase-to-goose-plan.md) — that is the
authoritative doc for this conversion now; treat the summary below as a pointer,
not a second source of truth to keep in sync by hand.

Steps (as actually run — see liquibase-to-goose-plan.md §2 for the real
commands):

1. Extract via the standalone Liquibase CLI's `offline:postgresql` mode — no
   Maven, no live database needed at all (better than the originally planned
   `mvn liquibase:updateSQL` against a throwaway DB).
2. Split by changeset into numbered goose files (`cmd/splitliquibase`).
3. Idempotency guards + best-effort `-- +goose Down` generated automatically
   where safe; honest `-- TODO` where not (liquibase-to-goose-plan.md §3, §7).
4. Run goose against a clean, baselined DB and verify — 967/993 apply cleanly;
   26 known open items are itemized in liquibase-to-goose-plan.md §9.2.
5. Baseline production so goose does not re-run the history (see the full
   procedure in [liquibase-to-goose-plan.md](liquibase-to-goose-plan.md) §6).
6. Remove Liquibase and Java; goose takes ownership. **Still gated on Java
   retirement — unchanged from the original plan.**

### Goose file layout

```
migration/openelis-go/db/migrations/
    0001_convert_id_types_1.sql
    0002_convert_id_types_2.sql
    ...
    0993_009_nullable_unit_of_measure_qc_009_nullable_unit_of_measure.sql
```

### Goose version tracking

Goose creates a `goose_db_version` table (like Liquibase's `databasechangelog`).
Production already carries the full Liquibase-built schema, so goose must be
told those migrations are done **without executing them** — goose has no
built-in "baseline" command for this, so it is done by seeding the version table
directly. Full procedure in
[liquibase-to-goose-plan.md](liquibase-to-goose-plan.md) §6.

---

## 3. Summary of tools

| Tool                      | Purpose                                     | Status                            |
| ------------------------- | ------------------------------------------- | --------------------------------- |
| `gorm.io/gorm`            | ORM — queries, writes, transactions         | **adopted in b1**                 |
| `gorm.io/driver/postgres` | GORM PostgreSQL driver (pgx under the hood) | **adopted in b1**                 |
| Liquibase                 | Schema migrations                           | **keep as-is during coexistence** |
| Goose                     | Replaces Liquibase after Java retired       | **future**                        |
