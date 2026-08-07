// Package valueholder ports org.openelisglobal.test.valueholder.
// Folder layout mirrors the Java source during migration.
package valueholder

// TestSection mirrors test/valueholder/TestSection.java — one row of
// clinlims.test_section. Name holds the English localized name (resolved from
// localization_value JOIN); falls back to the raw name column.
type TestSection struct {
	ID   string `gorm:"column:id"`
	Name string `gorm:"column:name"`
}
