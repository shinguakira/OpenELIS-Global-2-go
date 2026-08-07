// Package valueholder ports org.openelisglobal.localization.valueholder.
// Folder layout mirrors the Java source during migration.
package valueholder

// LocalizationValue mirrors localization/valueholder/LocalizationValue.java —
// one row of clinlims.localization_value.
type LocalizationValue struct {
	ID     string
	Locale string
	Value  string
}

// Localization mirrors localization/valueholder/Localization.java.
// Values is keyed by locale string (e.g. "en", "fr").
// Lastupdated is milliseconds since epoch; nil when the DB row has no timestamp.
type Localization struct {
	ID          string
	Description string
	Values      map[string]LocalizationValue
	Lastupdated *int64
}
