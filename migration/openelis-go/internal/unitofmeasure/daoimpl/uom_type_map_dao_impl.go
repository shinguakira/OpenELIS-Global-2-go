// Package daoimpl ports org.openelisglobal.unitofmeasure.daoimpl.
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"database/sql"

	"openelis-go/internal/unitofmeasure/valueholder"
)

// UomTypeMapDAOImpl ports UomTypeMapDAOImpl — reads clinlims.uom_type_map.
type UomTypeMapDAOImpl struct {
	DB *sql.DB
}

// GetUnitOfMeasuresByType mirrors UomTypeMapDAOImpl.getUnitOfMeasuresByType():
// JPQL: SELECT m.unitOfMeasure FROM UomTypeMap m WHERE m.uomType = :uomType
func (d *UomTypeMapDAOImpl) GetUnitOfMeasuresByType(uomType string) ([]valueholder.UnitOfMeasure, error) {
	rows, err := d.DB.Query(`
		SELECT u.id::text, COALESCE(u.name, '')
		FROM clinlims.unit_of_measure u
		JOIN clinlims.uom_type_map m ON m.uom_id = u.id
		WHERE m.uom_type = $1`, uomType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []valueholder.UnitOfMeasure{}
	for rows.Next() {
		var u valueholder.UnitOfMeasure
		if err := rows.Scan(&u.ID, &u.Name); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, rows.Err()
}
