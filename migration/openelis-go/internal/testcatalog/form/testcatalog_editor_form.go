// Package form ports org.openelisglobal.testcatalog.form (constitution.md
// Layer V — "Forms/DTOs"). Folder layout mirrors the Java source during migration.
package form

// IdNameDTO mirrors the IdValuePair the Java controller returns for lab-units,
// sample-types, and panels: {id: string, name: string}.
type IdNameDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
