// Package daoimpl ports org.openelisglobal.unitofmeasure.daoimpl.
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"gorm.io/gorm"

	"openelis-go/internal/unitofmeasure/valueholder"
)

// UomTypeMapDAOImpl ports UomTypeMapDAOImpl — reads clinlims.uom_type_map
// via GORM.
type UomTypeMapDAOImpl struct {
	DB *gorm.DB
}

// GetUnitOfMeasuresByType mirrors UomTypeMapDAOImpl.getUnitOfMeasuresByType():
// JPQL: SELECT m.unitOfMeasure FROM UomTypeMap m WHERE m.uomType = :uomType
// GORM Raw handles the positional parameter natively.
func (d *UomTypeMapDAOImpl) GetUnitOfMeasuresByType(uomType string) ([]valueholder.UnitOfMeasure, error) {
	var uoms []valueholder.UnitOfMeasure
	result := d.DB.Raw(`
		SELECT u.id::text AS id, COALESCE(u.name, '') AS name
		FROM clinlims.unit_of_measure u
		JOIN clinlims.uom_type_map m ON m.uom_id = u.id
		WHERE m.uom_type = ?`, uomType).Scan(&uoms)
	return uoms, result.Error
}
