// Package daoimpl ports org.openelisglobal.localization.daoimpl.
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"database/sql"

	"openelis-go/internal/localization/valueholder"
)

// SupportedLocaleDAO ports SupportedLocaleDAOImpl — reads clinlims.supported_locale.
// id is cast to text so it serializes as "1" (matching the Java String id), not a
// number. last_updated is intentionally not selected (the DTO omits it).
type SupportedLocaleDAO struct {
	DB *sql.DB
}

const cols = `id::text, locale_code, display_name, is_active, is_fallback, sort_order`

// GetAll mirrors BaseDAOImpl.getAll() — every row, no ORDER BY (DB-natural order).
func (d *SupportedLocaleDAO) GetAll() ([]valueholder.SupportedLocale, error) {
	return d.query(`SELECT ` + cols + ` FROM clinlims.supported_locale`)
}

// GetAllActive mirrors getAllMatchingOrdered("active", true, "sortOrder", false):
// WHERE is_active = true ORDER BY sort_order ASC.
func (d *SupportedLocaleDAO) GetAllActive() ([]valueholder.SupportedLocale, error) {
	return d.query(`SELECT ` + cols + ` FROM clinlims.supported_locale ` +
		`WHERE is_active = true ORDER BY sort_order ASC`)
}

// GetFallback mirrors getFallback() — the first row with is_fallback = true, or
// nil when none (the controller turns nil into 404).
func (d *SupportedLocaleDAO) GetFallback() (*valueholder.SupportedLocale, error) {
	list, err := d.query(`SELECT ` + cols + ` FROM clinlims.supported_locale WHERE is_fallback = true`)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	return &list[0], nil
}

func (d *SupportedLocaleDAO) query(q string) ([]valueholder.SupportedLocale, error) {
	rows, err := d.DB.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Non-nil slice so an empty result serializes as [] (not null).
	list := []valueholder.SupportedLocale{}
	for rows.Next() {
		var l valueholder.SupportedLocale
		if err := rows.Scan(&l.Id, &l.LocaleCode, &l.DisplayName, &l.Active, &l.Fallback, &l.SortOrder); err != nil {
			return nil, err
		}
		list = append(list, l)
	}
	return list, rows.Err()
}
