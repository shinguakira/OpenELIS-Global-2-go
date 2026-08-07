// Package valueholder ports org.openelisglobal.unitofmeasure.valueholder.
// Folder layout mirrors the Java source during migration.
package valueholder

// UnitOfMeasure mirrors unitofmeasure/valueholder/UnitOfMeasure.java.
// Maps to clinlims.unit_of_measure columns.
type UnitOfMeasure struct {
	ID   string `gorm:"column:id"`
	Name string `gorm:"column:name"`
}
