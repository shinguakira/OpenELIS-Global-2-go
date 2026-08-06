// Package daoimpl ports org.openelisglobal.test.daoimpl — TestSectionDAOImpl.
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"database/sql"

	"openelis-go/internal/test/valueholder"
)

// TestSectionDAOImpl ports TestSectionDAOImpl — reads clinlims.test_section.
type TestSectionDAOImpl struct {
	DB *sql.DB
}

// GetAll mirrors TestSectionDAOImpl.getAllTestSections() — every section with
// English name resolved from localization_value; falls back to the raw name column.
func (d *TestSectionDAOImpl) GetAll() ([]valueholder.TestSection, error) {
	rows, err := d.DB.Query(`
		SELECT ts.id::text,
		       COALESCE(lv.value, ts.name, '') AS name
		FROM clinlims.test_section ts
		LEFT JOIN clinlims.localization_value lv
		    ON lv.localization_id = ts.name_localization_id
		    AND lv.locale = 'en'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []valueholder.TestSection{}
	for rows.Next() {
		var ts valueholder.TestSection
		if err := rows.Scan(&ts.ID, &ts.Name); err != nil {
			return nil, err
		}
		list = append(list, ts)
	}
	return list, rows.Err()
}
