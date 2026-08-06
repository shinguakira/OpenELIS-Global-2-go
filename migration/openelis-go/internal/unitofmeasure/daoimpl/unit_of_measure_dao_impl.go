// Package daoimpl ports org.openelisglobal.unitofmeasure.daoimpl.
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"database/sql"

	"openelis-go/internal/unitofmeasure/valueholder"
)

// UnitOfMeasureDAOImpl ports UnitOfMeasureDAOImpl — reads clinlims.unit_of_measure.
type UnitOfMeasureDAOImpl struct {
	DB *sql.DB
}

// GetAll mirrors BaseDAOImpl.getAll() — every UOM row, no ORDER BY.
func (d *UnitOfMeasureDAOImpl) GetAll() ([]valueholder.UnitOfMeasure, error) {
	rows, err := d.DB.Query(
		`SELECT id::text, COALESCE(name, '') FROM clinlims.unit_of_measure`)
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
