// Package valueholder ports org.openelisglobal.test.valueholder.
// Folder layout mirrors the Java source during migration.
package valueholder

// TestSection mirrors test/valueholder/TestSection.java — one row of
// clinlims.test_section. Name holds the English localized name (resolved from
// localization_value JOIN); falls back to the raw name column.
// ID is int64 — the real Postgres type; string conversion for JSON happens in
// the controller DTO.
type TestSection struct {
	ID   int64  `gorm:"column:id"`
	Name string `gorm:"column:name"`
}

// TableName pins the GORM table name.
func (TestSection) TableName() string { return "clinlims.test_section" }
