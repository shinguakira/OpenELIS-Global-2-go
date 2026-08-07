// Package daoimpl ports org.openelisglobal.test.daoimpl — TestSectionDAOImpl.
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"gorm.io/gorm"

	"openelis-go/internal/test/valueholder"
)

// TestSectionDAOImpl ports TestSectionDAOImpl — reads clinlims.test_section
// via GORM.
type TestSectionDAOImpl struct {
	DB *gorm.DB
}

// GetAll mirrors TestSectionDAOImpl.getAllTestSections() — every section with
// English name resolved from localization_value; falls back to the raw name column.
func (d *TestSectionDAOImpl) GetAll() ([]valueholder.TestSection, error) {
	var sections []valueholder.TestSection
	result := d.DB.Raw(`
		SELECT ts.id::text AS id,
		       COALESCE(lv.value, ts.name, '') AS name
		FROM clinlims.test_section ts
		LEFT JOIN clinlims.localization_value lv
		    ON lv.localization_id = ts.name_localization_id
		    AND lv.locale = 'en'`).Scan(&sections)
	return sections, result.Error
}
