// Package form ports org.openelisglobal.organization.form (constitution.md
// Layer V — "Forms/DTOs"): client<->server wire shapes, kept separate from
// the valueholder (Layer I) and the service that builds them (Layer III).
// Folder layout mirrors the Java source during migration.
package form

// OrgTypeDTO mirrors OrganizationType's JSON shape. Pointer fields with
// omitempty mirror Jackson's Include.NON_NULL (config/AppConfig.java) —
// dropped from the response when null, not emitted as JSON null.
type OrgTypeDTO struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Description    *string `json:"description,omitempty"`
	HierarchyLevel *int    `json:"hierarchyLevel,omitempty"`
	Lastupdated    *int64  `json:"lastupdated,omitempty"`
}

// OrganizationDTO mirrors the raw Organization entity's JSON shape (both
// organization-list and organization/{id} serialize the same entity type in
// Java, so they share this DTO here too). The self-referencing parent
// ("organization" field in Java, guarded there against infinite recursion by
// handleSelfReferencingParentOrg) is deliberately not embedded — see
// migration/b2-org-provider-migration.md. datim_org_code/datim_org_name are
// real DB columns but are never Hibernate-mapped in Java, so no endpoint ever
// serializes them; this DTO doesn't either.
type OrganizationDTO struct {
	ID                 string       `json:"id"`
	OrganizationName   string       `json:"organizationName"`
	City               *string      `json:"city,omitempty"`
	ZipCode            *string      `json:"zipCode,omitempty"`
	MlsSentinelLabFlag string       `json:"mlsSentinelLabFlag"`
	ShortName          *string      `json:"shortName,omitempty"`
	MultipleUnit       *string      `json:"multipleUnit,omitempty"`
	StreetAddress      *string      `json:"streetAddress,omitempty"`
	State              *string      `json:"state,omitempty"`
	InternetAddress    *string      `json:"internetAddress,omitempty"`
	CliaNum            *string      `json:"cliaNum,omitempty"`
	PwsID              *string      `json:"pwsId,omitempty"`
	Lastupdated        *int64       `json:"lastupdated,omitempty"`
	MlsLabFlag         *string      `json:"mlsLabFlag,omitempty"`
	IsActive           *string      `json:"isActive,omitempty"`
	LocalAbbrev        *string      `json:"organizationLocalAbbreviation,omitempty"`
	Code               *string      `json:"code,omitempty"`
	FhirUUID           *string      `json:"fhirUuid,omitempty"`
	OrganizationTypes  []OrgTypeDTO `json:"organizationTypes"`
	// TestSections is permanently [] — confirmed via exploration (full-repo
	// grep) that Organization.testSections is never populated anywhere in
	// the Java codebase. Hardcoded here rather than modeled as a real
	// relationship, matching that dead-field reality exactly.
	TestSections []struct{} `json:"testSections"`
}

// IdValuePairDTO mirrors common.util.IdValuePair — {"id":..., "value":...}.
type IdValuePairDTO struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

// SiteCodeDTO mirrors generateSiteCode()'s {"siteCode": "..."} response.
type SiteCodeDTO struct {
	SiteCode string `json:"siteCode"`
}
