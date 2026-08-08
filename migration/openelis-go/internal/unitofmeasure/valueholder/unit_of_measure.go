// Package valueholder ports org.openelisglobal.unitofmeasure.valueholder.
// Folder layout mirrors the Java source during migration.
package valueholder

// UnitOfMeasure mirrors unitofmeasure/valueholder/UnitOfMeasure.java.
// Maps to clinlims.unit_of_measure columns.
// ID is int64 — the real Postgres type; string conversion for JSON happens in
// the controller DTO, matching Java's entity (native id) vs Jackson (String) split.
type UnitOfMeasure struct {
	ID   int64  `gorm:"column:id"`
	Name string `gorm:"column:name"`
}

// TableName pins the GORM table name.
func (UnitOfMeasure) TableName() string { return "clinlims.unit_of_measure" }
