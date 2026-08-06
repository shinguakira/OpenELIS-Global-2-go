// Package daoimpl ports org.openelisglobal.dictionarycategory.daoimpl.
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"database/sql"

	"openelis-go/internal/dictionarycategory/valueholder"
)

// DictionaryCategoryDAOImpl ports DictionaryCategoryDAOImpl — reads
// clinlims.dictionary_category.
type DictionaryCategoryDAOImpl struct {
	DB *sql.DB
}

// GetAll mirrors BaseDAOImpl.getAll() — every row, no ORDER BY (DB-natural order).
// Java: DictionaryCategoryServiceImpl inherits getAll() from BaseObjectService.
func (d *DictionaryCategoryDAOImpl) GetAll() ([]valueholder.DictionaryCategory, error) {
	rows, err := d.DB.Query(`
		SELECT id::text,
		       COALESCE(description, '')   AS description,
		       COALESCE(local_abbrev, '')  AS local_abbreviation,
		       COALESCE(name, '')          AS category_name,
		       EXTRACT(EPOCH FROM lastupdated)::bigint * 1000 AS lastupdated_ms
		FROM clinlims.dictionary_category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []valueholder.DictionaryCategory{}
	for rows.Next() {
		var c valueholder.DictionaryCategory
		var ms sql.NullInt64
		if err := rows.Scan(&c.ID, &c.Description, &c.LocalAbbreviation, &c.CategoryName, &ms); err != nil {
			return nil, err
		}
		if ms.Valid {
			v := ms.Int64
			c.Lastupdated = &v
		}
		list = append(list, c)
	}
	return list, rows.Err()
}
