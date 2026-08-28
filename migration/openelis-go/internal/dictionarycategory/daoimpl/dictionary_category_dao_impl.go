// Package daoimpl ports org.openelisglobal.dictionarycategory.daoimpl.
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"gorm.io/gorm"

	"openelis-go/internal/dictionarycategory/valueholder"
)

// DictionaryCategoryDAOImpl ports DictionaryCategoryDAOImpl — reads
// clinlims.dictionary_category via GORM's query builder (Select + Find), the
// same auto-generated-SQL path Hibernate uses for simple entity reads. The
// Select() fragment is needed only for the COALESCE(...,”) null-guards — the
// FROM clause comes from DictionaryCategory.TableName(), not a manual string.
type DictionaryCategoryDAOImpl struct {
	DB *gorm.DB
}

// GetAll mirrors BaseDAOImpl.getAll() — every row, no ORDER BY (DB-natural order).
// Java: DictionaryCategoryServiceImpl inherits getAll() from BaseObjectService.
func (d *DictionaryCategoryDAOImpl) GetAll() ([]valueholder.DictionaryCategory, error) {
	var categories []valueholder.DictionaryCategory
	result := d.DB.
		Select(`id,
		        COALESCE(description, '')   AS description,
		        COALESCE(local_abbrev, '')  AS local_abbreviation,
		        COALESCE(name, '')          AS category_name,
		        lastupdated`).
		Find(&categories)
	return categories, result.Error
}
