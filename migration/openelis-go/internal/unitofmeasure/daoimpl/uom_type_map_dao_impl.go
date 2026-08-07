// Package daoimpl ports org.openelisglobal.unitofmeasure.daoimpl.
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"gorm.io/gorm"

	"openelis-go/internal/unitofmeasure/valueholder"
)

// UomTypeMapDAOImpl ports UomTypeMapDAOImpl — reads clinlims.uom_type_map
// via GORM's query builder. A single INNER JOIN with one predicate is well
// within what .Joins() expresses cleanly — Raw() is reserved for queries a
// builder genuinely can't express (LATERAL subqueries, bulk IN aggregation),
// same boundary Hibernate draws between JPQL/Criteria and native SQL.
type UomTypeMapDAOImpl struct {
	DB *gorm.DB
}

// GetUnitOfMeasuresByType mirrors UomTypeMapDAOImpl.getUnitOfMeasuresByType():
// JPQL: SELECT m.unitOfMeasure FROM UomTypeMap m WHERE m.uomType = :uomType
func (d *UomTypeMapDAOImpl) GetUnitOfMeasuresByType(uomType string) ([]valueholder.UnitOfMeasure, error) {
	var uoms []valueholder.UnitOfMeasure
	result := d.DB.
		Table("clinlims.unit_of_measure AS u").
		Select("u.id AS id, COALESCE(u.name, '') AS name").
		Joins("JOIN clinlims.uom_type_map m ON m.uom_id = u.id").
		Where("m.uom_type = ?", uomType).
		Find(&uoms)
	return uoms, result.Error
}
