// Package valueholder ports org.openelisglobal.unitofmeasure.valueholder.
// Folder layout mirrors the Java source during migration.
package valueholder

// UomTypeMap mirrors unitofmeasure/valueholder/UomTypeMap.java.
// Maps to clinlims.uom_type_map — links a UOM to a usage-type string
// (e.g. "SAMPLE_COLLECTION").
type UomTypeMap struct {
	ID            int
	UnitOfMeasure UnitOfMeasure
	UomType       string
}
