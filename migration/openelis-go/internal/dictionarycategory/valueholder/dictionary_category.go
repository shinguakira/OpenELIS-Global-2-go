// Package valueholder ports org.openelisglobal.dictionarycategory.valueholder.
// Folder layout mirrors the Java source during migration.
package valueholder

import "time"

// DictionaryCategory mirrors dictionarycategory/valueholder/DictionaryCategory.java.
// Maps to clinlims.dictionary_category.
// Lastupdated is *time.Time; nil when the DB row has no timestamp.
// The controller converts *time.Time → *int64 (epoch ms) for JSON output to match
// Jackson's default Date serialisation.
type DictionaryCategory struct {
	ID                string     `gorm:"column:id"`
	Description       string     `gorm:"column:description"`
	LocalAbbreviation string     `gorm:"column:local_abbreviation"`
	CategoryName      string     `gorm:"column:category_name"`
	Lastupdated       *time.Time `gorm:"column:lastupdated"`
}
