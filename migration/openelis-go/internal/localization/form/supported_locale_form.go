// Package form ports org.openelisglobal.localization.form (constitution.md
// Layer V — "Forms/DTOs"). Folder layout mirrors the Java source during migration.
package form

// SupportedLocaleDTO mirrors SupportedLocaleRestController.SupportedLocaleDTO.
// Field order matches the Java DTO's JSON output.
type SupportedLocaleDTO struct {
	Id          string `json:"id"`
	LocaleCode  string `json:"localeCode"`
	DisplayName string `json:"displayName"`
	Active      bool   `json:"active"`
	Fallback    bool   `json:"fallback"`
	SortOrder   int    `json:"sortOrder"`
}
