// Package valueholder ports org.openelisglobal.dictionarycategory.valueholder.
// Folder layout mirrors the Java source during migration.
package valueholder

// DictionaryCategory mirrors dictionarycategory/valueholder/DictionaryCategory.java.
// Maps to clinlims.dictionary_category columns.
// Lastupdated is milliseconds since epoch; nil when the DB row has no timestamp
// (matches Jackson @JsonInclude(NON_NULL) behaviour).
type DictionaryCategory struct {
	ID                string
	Description       string
	LocalAbbreviation string
	CategoryName      string
	Lastupdated       *int64
}
