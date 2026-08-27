// Package form ports org.openelisglobal.unitofmeasure.form (constitution.md
// Layer V — "Forms/DTOs"). Folder layout mirrors the Java source during migration.
package form

// UnitOfMeasureDTO mirrors the {id, value} shape the Java controller returns.
// Java uses IdValuePair serialized as {"id": "...", "value": "..."}.
type UnitOfMeasureDTO struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}
