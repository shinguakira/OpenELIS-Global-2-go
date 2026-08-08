# a2 GORM Migration

Status: **done** — merged on `migration/a2-static-reads`, commits `41c899f3f`,
`ebd8a6aaf`, `e2be9b367`. This doc now documents the actual shipped pattern, as
a reference template for the next domain migrated the same way — not a
forward-looking plan. Code samples below are the real committed code, not an
earlier draft; if you're copying a pattern for a new domain, copy from here.

Depends on: b1 GORM rewrite (merged to `migration-base` via PR #7, commit
c0f2aa6cf) — `gorm.io/gorm v1.31.2` in `go.mod` as a **direct** require (a2's
own daoimpl files import it directly now, same as b1's).

---

## What changed

Two a2 files used raw `database/sql`; both moved to GORM's query builder
(`.Where()`/`.Order()`/`.Select()`/`.Find()`/`.First()`), **not** `db.Raw()` —
neither query has a join or needs anything a builder can't express:

| File | Before | After |
|---|---|---|
| `internal/localization/daoimpl/supported_locale_dao_impl.go` | `DB *sql.DB`; manual `rows.Next()/rows.Scan()` | `DB *gorm.DB`; `.Where().Order().Find()` / `.First()` |
| `internal/common/services/status.go` | `NewStatusService(db *sql.DB, ...)` runs `db.Query()` itself | DAO extracted; `StatusService` takes the DAO, holds no DB handle |
| `cmd/openelis/main.go` | extracted `sqlDB` from `gormDB.DB()` for a2 | passes `gormDB` directly; `sqlDB` extraction removed entirely |

a1 (system/server-time) has no DB access — no change needed.

**Layer note:** `status.go` used to run `db.Query()` directly inside a
`services` package — the same DAO-bypass pattern that Codex/Copilot flagged in
b1's `TestCatalogService` (fixed in commit 9892eb2a2). Per the tightened rule in
[OpenELIS-Go-Migration-Plan.md](OpenELIS-Go-Migration-Plan.md) ("`daoimpl/` is
the only layer allowed to import a database or ORM package"), a DAO was
extracted for status too, instead of just swapping `*sql.DB` → `*gorm.DB` in
place and carrying the same violation forward.

**ORM-usage note:** the first pass at this (superseded, not what's below) used
`db.Raw(fullSQLString).Scan(&x)` for every query, including plain single-table
reads with no join — that's Hibernate-native-SQL-equivalent usage for queries
that don't need it. GORM's query builder (`Select`/`Where`/`Order`/`Joins`/`Find`)
is the equivalent of Hibernate's Criteria/HQL — use it for anything a join or a
filter can express; reserve `Raw()` for the same tier of complexity where
Hibernate itself would need `createNativeQuery()` (see
`testconfiguration/daoimpl/test_catalog_dao_impl.go`'s LATERAL+bulk-IN query —
the one `Raw()` call left in the whole module).

---

## File-by-file changes

### 1. `internal/localization/valueholder/supported_locale.go`

`Id` is `int64` — the real Postgres type, not a cast-to-string. JSON-string
conversion happens only in the controller DTO (§ below), matching Java's
entity-vs-Jackson split. `TableName()` pins the table since GORM's default
pluralization guess would be wrong:

```go
type SupportedLocale struct {
    Id          int64  `gorm:"column:id"`
    LocaleCode  string `gorm:"column:locale_code"`
    DisplayName string `gorm:"column:display_name"`
    Active      bool   `gorm:"column:is_active"`
    Fallback    bool   `gorm:"column:is_fallback"`
    SortOrder   int    `gorm:"column:sort_order"`
}

func (SupportedLocale) TableName() string { return "clinlims.supported_locale" }
```

### 2. `internal/localization/daoimpl/supported_locale_dao_impl.go`

`*gorm.DB` + the query builder — no `.Select()` needed at all here: every
column maps 1:1 to a struct field via `gorm` tags, and none are nullable (no
`COALESCE` needed, same as the original Java query had none):

```go
import (
    "errors"

    "gorm.io/gorm"
)

type SupportedLocaleDAO struct {
    DB *gorm.DB
}

func (d *SupportedLocaleDAO) GetAll() ([]valueholder.SupportedLocale, error) {
    var list []valueholder.SupportedLocale
    result := d.DB.Find(&list)
    if list == nil {
        list = []valueholder.SupportedLocale{}
    }
    return list, result.Error
}

func (d *SupportedLocaleDAO) GetAllActive() ([]valueholder.SupportedLocale, error) {
    var list []valueholder.SupportedLocale
    result := d.DB.Where("is_active = ?", true).Order("sort_order ASC").Find(&list)
    if list == nil {
        list = []valueholder.SupportedLocale{}
    }
    return list, result.Error
}

// GetFallback uses First() — GORM's canonical single-row read, mirrors JPA's
// getSingleResult()/NoResultException — instead of Find()+manual empty check.
func (d *SupportedLocaleDAO) GetFallback() (*valueholder.SupportedLocale, error) {
    var loc valueholder.SupportedLocale
    err := d.DB.Where("is_fallback = ?", true).First(&loc).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    return &loc, nil
}
```

### 3. New: `internal/common/valueholder/status_row.go`

DAO scan target — mirrors the `testconfiguration/valueholder/test_row.go`
pattern from the b1 fix. Fields exported for GORM reflection. `ID` is `int64`
(real Postgres type); `StatusService` converts to string only when it builds
its in-memory map (§5).

```go
// Package valueholder holds DB projections for internal/common domains.
package valueholder

// StatusRow is the raw projection of one clinlims.status_of_sample row.
type StatusRow struct {
    ID         int64  `gorm:"column:id"`
    StatusType string `gorm:"column:status_type"`
    Name       string `gorm:"column:name"`
    DisplayKey string `gorm:"column:display_key"`
}

func (StatusRow) TableName() string { return "clinlims.status_of_sample" }
```

### 4. New: `internal/common/daoimpl/status_dao_impl.go`

The extracted DAO — the only place in this domain that imports `gorm.io/gorm`.
`.Select()` is needed here (unlike SupportedLocale) because `display_key` is
nullable and needs `COALESCE`; there's still no join, so `.Find()`, not `Raw()`:

```go
// Package daoimpl ports the status_of_sample data access used by StatusService.
package daoimpl

import (
    "gorm.io/gorm"

    "openelis-go/internal/common/valueholder"
)

// StatusDAOImpl reads clinlims.status_of_sample.
type StatusDAOImpl struct {
    DB *gorm.DB
}

// GetAll returns every status_of_sample row.
func (d *StatusDAOImpl) GetAll() ([]valueholder.StatusRow, error) {
    var rows []valueholder.StatusRow
    result := d.DB.Select("id, status_type, name, COALESCE(display_key, '') AS display_key").Find(&rows)
    return rows, result.Error
}
```

### 5. `internal/common/services/status.go`

`StatusService` no longer holds a DB handle. `NewStatusService` takes the DAO,
calls it once, and builds the in-memory map — pure business logic, no SQL.
`statusEntry.id` is still `string` (that's the public contract `IDByName`
returns), so the int64 → string conversion happens right here, once, via
`strconv.FormatInt` — the one place this domain actually needs the string form:

```go
import (
    "strconv"

    "openelis-go/internal/common/daoimpl"
)

type StatusService struct {
    entryByKey map[string]statusEntry
}

// NewStatusService loads status_of_sample rows via the DAO into an in-memory
// map. msgs is the parsed message_en.properties (from i18n.Messages()).
func NewStatusService(dao *daoimpl.StatusDAOImpl, msgs map[string]string) (*StatusService, error) {
    rows, err := dao.GetAll()
    if err != nil {
        return nil, err
    }

    m := map[string]statusEntry{}
    for _, r := range rows {
        label := msgs[r.DisplayKey]
        if label == "" {
            label = r.Name
        }
        m[r.StatusType+"\x00"+r.Name] = statusEntry{id: strconv.FormatInt(r.ID, 10), label: label}
    }
    return &StatusService{entryByKey: m}, nil
}
```

`IDByName` / `LabelByName` are unchanged.

### 6. `cmd/openelis/main.go`

Remove the `sqlDB` extraction step. Wire a `StatusDAOImpl`, pass `gormDB`
directly to `SupportedLocaleDAO`:

```go
// Before:
sqlDB, err := gormDB.DB()
if err != nil {
    log.Fatalf("failed to extract *sql.DB from GORM: %v", err)
}
...
&localizationdao.SupportedLocaleDAO{DB: sqlDB}
...
commonservices.NewStatusService(sqlDB, msgs)

// After:
&localizationdao.SupportedLocaleDAO{DB: gormDB}
...
statusDAO := &commondaoimpl.StatusDAOImpl{DB: gormDB}
commonservices.NewStatusService(statusDAO, msgs)
```

Add the import: `commondaoimpl "openelis-go/internal/common/daoimpl"`.

---

## Verify

```bash
cd migration/openelis-go
go build ./...
go vet ./...
```

Both must produce no output. Also confirm no file under `internal/common/services/`
or `internal/localization/service/` imports `gorm.io/gorm` — only the two
`daoimpl/` packages should:

```bash
grep -rl 'gorm.io/gorm' internal/common/services internal/localization/service
```

Expect no output.
