# a2 GORM Migration

Branch: `migration/a2-static-reads`
Depends on: b1 GORM rewrite (merged to `migration-base` via PR #7,
commit c0f2aa6cf) — `gorm.io/gorm v1.31.2` already in `go.mod` (currently
marked `// indirect` since nothing on this branch imports it directly yet;
that clears once the DAO edits below land).

---

## What needs changing

Two a2 files still use raw `database/sql`:

| File | Current | Change to |
|---|---|---|
| `internal/localization/daoimpl/supported_locale_dao_impl.go` | `DB *sql.DB`; manual `rows.Next()/rows.Scan()` | `DB *gorm.DB`; `db.Raw().Scan()` |
| `internal/common/services/status.go` | `NewStatusService(db *sql.DB, ...)` runs `db.Query()` itself | Extract a DAO; `StatusService` takes the DAO, holds no DB handle |
| `cmd/openelis/main.go` | extracts `sqlDB` from `gormDB.DB()` for a2 | pass `gormDB` directly; remove `sqlDB` extraction |

a1 (system/server-time) has no DB access — no change needed.

**Layer note:** `status.go` currently runs `db.Query()` directly inside a
`services` package — the same DAO-bypass pattern that Codex/Copilot flagged in
b1's `TestCatalogService` (fixed in commit 9892eb2a2). Per the tightened rule in
[OpenELIS-Go-Migration-Plan.md](OpenELIS-Go-Migration-Plan.md) ("`daoimpl/` is
the only layer allowed to import a database or ORM package"), this migration
extracts a DAO for status too, rather than just swapping `*sql.DB` → `*gorm.DB`
in place and carrying the same violation forward.

---

## File-by-file changes

### 1. `internal/localization/valueholder/supported_locale.go`

Add `gorm:"column:..."` tags so `Raw().Scan()` matches the SQL aliases:

```go
type SupportedLocale struct {
    Id          string `gorm:"column:id"`
    LocaleCode  string `gorm:"column:locale_code"`
    DisplayName string `gorm:"column:display_name"`
    Active      bool   `gorm:"column:is_active"`
    Fallback    bool   `gorm:"column:is_fallback"`
    SortOrder   int    `gorm:"column:sort_order"`
}
```

### 2. `internal/localization/daoimpl/supported_locale_dao_impl.go`

Replace `*sql.DB` + helper `query()` method with `*gorm.DB` + `Raw().Scan()`:

```go
import "gorm.io/gorm"

type SupportedLocaleDAO struct {
    DB *gorm.DB
}

const cols = `id::text AS id, locale_code, display_name, is_active, is_fallback, sort_order`

func (d *SupportedLocaleDAO) GetAll() ([]valueholder.SupportedLocale, error) {
    var list []valueholder.SupportedLocale
    result := d.DB.Raw(`SELECT ` + cols + ` FROM clinlims.supported_locale`).Scan(&list)
    if list == nil {
        list = []valueholder.SupportedLocale{}
    }
    return list, result.Error
}

func (d *SupportedLocaleDAO) GetAllActive() ([]valueholder.SupportedLocale, error) {
    var list []valueholder.SupportedLocale
    result := d.DB.Raw(
        `SELECT ` + cols + ` FROM clinlims.supported_locale WHERE is_active = true ORDER BY sort_order ASC`,
    ).Scan(&list)
    if list == nil {
        list = []valueholder.SupportedLocale{}
    }
    return list, result.Error
}

func (d *SupportedLocaleDAO) GetFallback() (*valueholder.SupportedLocale, error) {
    var list []valueholder.SupportedLocale
    result := d.DB.Raw(
        `SELECT ` + cols + ` FROM clinlims.supported_locale WHERE is_fallback = true`,
    ).Scan(&list)
    if result.Error != nil {
        return nil, result.Error
    }
    if len(list) == 0 {
        return nil, nil
    }
    return &list[0], nil
}
```

Note: the `query()` helper is deleted; the three methods inline the scan directly
(see empty-slice invariant below — inlined directly this time instead of as a
follow-up).

### 3. New: `internal/common/valueholder/status_row.go`

DAO scan target — mirrors the `testconfiguration/valueholder/test_row.go`
pattern from the b1 fix. Fields exported for GORM reflection.

```go
// Package valueholder holds DB projections for internal/common domains.
package valueholder

// StatusRow is the raw projection of one clinlims.status_of_sample row.
type StatusRow struct {
    ID         string `gorm:"column:id"`
    StatusType string `gorm:"column:status_type"`
    Name       string `gorm:"column:name"`
    DisplayKey string `gorm:"column:display_key"`
}
```

### 4. New: `internal/common/daoimpl/status_dao_impl.go`

The extracted DAO — the only place in this domain that imports `gorm.io/gorm`:

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
    result := d.DB.Raw(`
        SELECT id::text AS id, status_type, name, COALESCE(display_key, '') AS display_key
        FROM clinlims.status_of_sample`).Scan(&rows)
    return rows, result.Error
}
```

### 5. `internal/common/services/status.go`

`StatusService` no longer holds a DB handle. `NewStatusService` takes the DAO,
calls it once, and builds the in-memory map — pure business logic, no SQL:

```go
import "openelis-go/internal/common/daoimpl"

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
        m[r.StatusType+"\x00"+r.Name] = statusEntry{id: r.ID, label: label}
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
