// Package valueholder ports org.openelisglobal.localization.valueholder.
// Folder layout mirrors the Java source during migration.
package valueholder

// SupportedLocale mirrors localization/valueholder/SupportedLocale.java — one row
// of the clinlims.supported_locale table. Id is int64 — the real Postgres type;
// string conversion for JSON happens in the controller DTO (matches Java's
// entity id vs Jackson-serialized BaseObject<String> split).
type SupportedLocale struct {
	Id          int64  `gorm:"column:id"`
	LocaleCode  string `gorm:"column:locale_code"`
	DisplayName string `gorm:"column:display_name"`
	Active      bool   `gorm:"column:is_active"`
	Fallback    bool   `gorm:"column:is_fallback"`
	SortOrder   int    `gorm:"column:sort_order"`
}

// TableName pins the GORM table name.
func (SupportedLocale) TableName() string { return "clinlims.supported_locale" }
