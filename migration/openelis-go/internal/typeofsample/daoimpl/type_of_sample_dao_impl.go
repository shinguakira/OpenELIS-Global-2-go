// Package daoimpl ports org.openelisglobal.typeofsample.daoimpl.
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"database/sql"

	"openelis-go/internal/typeofsample/valueholder"
)

// TypeOfSampleDAOImpl ports TypeOfSampleDAOImpl — reads clinlims.type_of_sample.
type TypeOfSampleDAOImpl struct {
	DB *sql.DB
}

// GetAllSortOrdered mirrors TypeOfSampleDAOImpl.getAllTypeOfSamplesSortOrdered():
// all rows ordered by sort_order ascending.
// Name = description when non-blank, else local_abbrev (mirrors getLocalizedName()).
func (d *TypeOfSampleDAOImpl) GetAllSortOrdered() ([]valueholder.TypeOfSample, error) {
	rows, err := d.DB.Query(`
		SELECT id::text,
		       CASE WHEN description IS NOT NULL AND TRIM(description) != ''
		            THEN description
		            ELSE COALESCE(local_abbrev, '') END AS name,
		       COALESCE(sort_order, 0)
		FROM clinlims.type_of_sample
		ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []valueholder.TypeOfSample{}
	for rows.Next() {
		var t valueholder.TypeOfSample
		if err := rows.Scan(&t.ID, &t.Name, &t.SortOrder); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}
