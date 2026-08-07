// Package daoimpl ports org.openelisglobal.unitofmeasure.daoimpl.
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"gorm.io/gorm"

	"openelis-go/internal/unitofmeasure/valueholder"
)

// UnitOfMeasureDAOImpl ports UnitOfMeasureDAOImpl — reads clinlims.unit_of_measure
// via GORM.
type UnitOfMeasureDAOImpl struct {
	DB *gorm.DB
}

// GetAll mirrors BaseDAOImpl.getAll() — every UOM row, no ORDER BY.
func (d *UnitOfMeasureDAOImpl) GetAll() ([]valueholder.UnitOfMeasure, error) {
	var uoms []valueholder.UnitOfMeasure
	result := d.DB.Raw(
		`SELECT id::text AS id, COALESCE(name, '') AS name FROM clinlims.unit_of_measure`,
	).Scan(&uoms)
	return uoms, result.Error
}
