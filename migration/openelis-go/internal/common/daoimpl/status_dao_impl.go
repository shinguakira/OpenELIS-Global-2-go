// Package daoimpl ports the status_of_sample data access used by StatusService.
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"gorm.io/gorm"

	"openelis-go/internal/common/valueholder"
)

// StatusDAOImpl reads clinlims.status_of_sample via GORM's query builder.
type StatusDAOImpl struct {
	DB *gorm.DB
}

// GetAll returns every status_of_sample row.
func (d *StatusDAOImpl) GetAll() ([]valueholder.StatusRow, error) {
	var rows []valueholder.StatusRow
	result := d.DB.Select("id, status_type, name, COALESCE(display_key, '') AS display_key").Find(&rows)
	return rows, result.Error
}
