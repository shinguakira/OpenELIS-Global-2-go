// Package valueholder ports org.openelisglobal.typeofsample.valueholder.
// Folder layout mirrors the Java source during migration.
package valueholder

// TypeOfSample mirrors typeofsample/valueholder/TypeOfSample.java — one row of
// clinlims.type_of_sample. Name is the display name returned by the endpoint:
// description when non-empty, otherwise localAbbreviation.
type TypeOfSample struct {
	ID        string `gorm:"column:id"`
	Name      string `gorm:"column:name"`
	SortOrder int    `gorm:"column:sort_order"`
}
