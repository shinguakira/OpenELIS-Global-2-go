# a2 GORM Migration

Branch: new `migration/a2-gorm` (fork from `migration-base`)  
Depends on: b1 GORM rewrite (commit c0f2aa6cf) — `gorm.io/gorm v1.31.2` already in `go.mod`

---

## What needs changing

Two a2 files still use raw `database/sql`:

| File | Current | Change to |
|---|---|---|
| `internal/localization/daoimpl/supported_locale_dao_impl.go` | `DB *sql.DB`; manual `rows.Next()/rows.Scan()` | `DB *gorm.DB`; `db.Raw().Scan()` |
| `internal/common/services/status.go` | `NewStatusService(db *sql.DB, ...)` | `NewStatusService(db *gorm.DB, ...)` |
| `cmd/openelis/main.go` | extracts `sqlDB` from `gormDB.DB()` for a2 | pass `gormDB` directly; remove `sqlDB` extraction |

a1 (system/server-time) has no DB access — no change needed.

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
    return list, result.Error
}

func (d *SupportedLocaleDAO) GetAllActive() ([]valueholder.SupportedLocale, error) {
    var list []valueholder.SupportedLocale
    result := d.DB.Raw(
        `SELECT ` + cols + ` FROM clinlims.supported_locale WHERE is_active = true ORDER BY sort_order ASC`,
    ).Scan(&list)
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

Note: the `query()` helper is deleted; the three methods inline the scan directly.

### 3. `internal/common/services/status.go`

Replace `*sql.DB` + manual cursor with `*gorm.DB` + `Raw().Scan()`.

Define an internal scan target with exported fields:

```go
import "gorm.io/gorm"

type statusRow struct {
    ID         string `gorm:"column:id"`
    StatusType string `gorm:"column:status_type"`
    Name       string `gorm:"column:name"`
    DisplayKey string `gorm:"column:display_key"`
}

func NewStatusService(db *gorm.DB, msgs map[string]string) (*StatusService, error) {
    var rows []statusRow
    result := db.Raw(`
        SELECT id::text AS id, status_type, name, COALESCE(display_key, '') AS display_key
        FROM clinlims.status_of_sample`).Scan(&rows)
    if result.Error != nil {
        return nil, result.Error
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

### 4. `cmd/openelis/main.go`

Remove the `sqlDB` extraction step; pass `gormDB` directly to both a2 constructors:

```go
// Before:
sqlDB, err := gormDB.DB()
...
&localizationdao.SupportedLocaleDAO{DB: sqlDB}
...
commonservices.NewStatusService(sqlDB, msgs)

// After:
&localizationdao.SupportedLocaleDAO{DB: gormDB}
...
commonservices.NewStatusService(gormDB, msgs)
```

---

## Verify

```bash
cd migration/openelis-go
go build ./...
go vet ./...
```

Both must produce no output.

---

## Empty-slice invariant

`GetAll()` and `GetAllActive()` must still return `[]SupportedLocale{}` (not nil)
when the table is empty, so the controller serialises `[]` not `null`.
GORM's `Scan()` into a pre-declared `var list []T` leaves it as nil on zero rows.
Fix with:

```go
if list == nil {
    list = []valueholder.SupportedLocale{}
}
return list, result.Error
```
