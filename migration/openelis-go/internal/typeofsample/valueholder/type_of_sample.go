// Package valueholder ports org.openelisglobal.typeofsample.valueholder.
// Folder layout mirrors the Java source during migration.
package valueholder

// TypeOfSample mirrors typeofsample/valueholder/TypeOfSample.java — one row of
// clinlims.type_of_sample. Name is the display name returned by the endpoint:
// description when non-empty, otherwise localAbbreviation.
// ID is int64 — the real Postgres type; string conversion for JSON happens in
// the controller DTO.
type TypeOfSample struct {
	ID        int64  `gorm:"column:id"`
	Name      string `gorm:"column:name"`
	SortOrder int    `gorm:"column:sort_order"`
}

// TableName pins the GORM table name.
func (TypeOfSample) TableName() string { return "clinlims.type_of_sample" }
