// Package form ports org.openelisglobal.dictionarycategory.form (constitution.md
// Layer V — "Forms/DTOs": client<->server wire shapes, kept separate from both
// the valueholder (Layer I) and the service that builds them (Layer III).
// Folder layout mirrors the Java source during migration.
package form

// DictionaryCategoryDTO is the JSON shape for each DictionaryCategory row.
// lastupdated is epoch-milliseconds (int64) to match Jackson's default Date
// serialisation; omitted when nil (mirrors @JsonInclude(NON_NULL)).
type DictionaryCategoryDTO struct {
	ID                string `json:"id"`
	Description       string `json:"description"`
	LocalAbbreviation string `json:"localAbbreviation"`
	CategoryName      string `json:"categoryName"`
	Lastupdated       *int64 `json:"lastupdated,omitempty"`
}
