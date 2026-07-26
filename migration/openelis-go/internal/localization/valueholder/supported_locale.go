// Package valueholder ports org.openelisglobal.localization.valueholder.
// Folder layout mirrors the Java source during migration.
package valueholder

// SupportedLocale mirrors localization/valueholder/SupportedLocale.java — one row
// of the clinlims.supported_locale table. Id is a string to match the Java entity
// (BaseObject<String>), even though the DB column is numeric.
type SupportedLocale struct {
	Id          string
	LocaleCode  string
	DisplayName string
	Active      bool
	Fallback    bool
	SortOrder   int
}
