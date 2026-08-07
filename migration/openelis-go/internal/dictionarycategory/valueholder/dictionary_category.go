// Package valueholder ports org.openelisglobal.dictionarycategory.valueholder.
// Folder layout mirrors the Java source during migration.
package valueholder

import "time"

// DictionaryCategory mirrors dictionarycategory/valueholder/DictionaryCategory.java.
// Maps to clinlims.dictionary_category.
// ID is int64 — the real Postgres BIGSERIAL type, no cast needed for GORM to
// scan it. Java's getId() returns String; that conversion happens at the DTO
// boundary in the controller (strconv.FormatInt), not here — same split as
// Hibernate entity (native type) vs Jackson DTO (serialized shape).
// Lastupdated is *time.Time; nil when the DB row has no timestamp.
type DictionaryCategory struct {
	ID                int64      `gorm:"column:id"`
	Description       string     `gorm:"column:description"`
	LocalAbbreviation string     `gorm:"column:local_abbreviation"`
	CategoryName      string     `gorm:"column:category_name"`
	Lastupdated       *time.Time `gorm:"column:lastupdated"`
}

// TableName pins the GORM table name — clinlims schema, singular (Postgres
// table is not pluralized), overriding GORM's default pluralization guess.
func (DictionaryCategory) TableName() string { return "clinlims.dictionary_category" }
