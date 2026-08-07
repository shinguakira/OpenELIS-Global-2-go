// Package daoimpl ports org.openelisglobal.dictionarycategory.daoimpl.
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"gorm.io/gorm"

	"openelis-go/internal/dictionarycategory/valueholder"
)

// DictionaryCategoryDAOImpl ports DictionaryCategoryDAOImpl — reads
// clinlims.dictionary_category via GORM.
type DictionaryCategoryDAOImpl struct {
	DB *gorm.DB
}

// GetAll mirrors BaseDAOImpl.getAll() — every row, no ORDER BY (DB-natural order).
// Java: DictionaryCategoryServiceImpl inherits getAll() from BaseObjectService.
// id::text cast is explicit — OpenELIS PKs are BIGSERIAL but Java exposes them
// as String; GORM Raw keeps that cast rather than relying on pgx type coercion.
func (d *DictionaryCategoryDAOImpl) GetAll() ([]valueholder.DictionaryCategory, error) {
	var categories []valueholder.DictionaryCategory
	result := d.DB.Raw(`
		SELECT id::text                          AS id,
		       COALESCE(description, '')         AS description,
		       COALESCE(local_abbrev, '')        AS local_abbreviation,
		       COALESCE(name, '')               AS category_name,
		       lastupdated
		FROM clinlims.dictionary_category`).Scan(&categories)
	return categories, result.Error
}
