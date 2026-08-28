// Package form ports org.openelisglobal.provider.form (constitution.md
// Layer V — "Forms/DTOs"): client<->server wire shapes, kept separate from
// the valueholder (Layer I) and the service that builds them (Layer III).
// Folder layout mirrors the Java source during migration.
package form

// PersonDTO mirrors Person's JSON shape (GET Provider/Person/{id}). Pointer
// fields with omitempty mirror Jackson's Include.NON_NULL.
type PersonDTO struct {
	ID            string   `json:"id"`
	LastName      *string  `json:"lastName,omitempty"`
	FirstName     *string  `json:"firstName,omitempty"`
	MiddleName    *string  `json:"middleName,omitempty"`
	MultipleUnit  *string  `json:"multipleUnit,omitempty"`
	StreetAddress *string  `json:"streetAddress,omitempty"`
	City          *string  `json:"city,omitempty"`
	State         *string  `json:"state,omitempty"`
	ZipCode       *string  `json:"zipCode,omitempty"`
	Country       *string  `json:"country,omitempty"`
	WorkPhone     *string  `json:"workPhone,omitempty"`
	HomePhone     *string  `json:"homePhone,omitempty"`
	CellPhone     *string  `json:"cellPhone,omitempty"`
	PrimaryPhone  *string  `json:"primaryPhone,omitempty"`
	Fax           *string  `json:"fax,omitempty"`
	Email         *string  `json:"email,omitempty"`
	GpsLatitude   *float64 `json:"gpsLatitude,omitempty"`
	GpsLongitude  *float64 `json:"gpsLongitude,omitempty"`
	Lastupdated   *int64   `json:"lastupdated,omitempty"`
}

// ProviderDTO mirrors Provider's JSON shape (GET Provider/raw/{id}, GET
// rest/practitioner) — the full entity with Person nested under "person",
// matching Provider.hbm.xml's eager (lazy="false") many-to-one.
// FhirUUIDAsString is a real second field on Provider.java
// (getFhirUuidAsString(), a Jackson-visible getter distinct from the UUID
// getFhirUuid()) — both are ported since Java emits both.
type ProviderDTO struct {
	ID               string    `json:"id"`
	ExternalID       *string   `json:"externalId,omitempty"`
	NPI              *string   `json:"npi,omitempty"`
	ProviderType     *string   `json:"providerType,omitempty"`
	Person           PersonDTO `json:"person"`
	FhirUUID         *string   `json:"fhirUuid,omitempty"`
	FhirUUIDAsString string    `json:"fhirUuidAsString"`
	Active           bool      `json:"active"`
	Desynchronized   bool      `json:"desynchronized"`
	Lastupdated      *int64    `json:"lastupdated,omitempty"`
}

// ProviderSearchResultDTO mirrors the hand-built Map<String,Object> row shape
// from ProviderRestController.searchProviders exactly (field-by-field, not a
// formal Java DTO class — see migration exploration notes).
type ProviderSearchResultDTO struct {
	ID         string  `json:"id"`
	PersonID   *string `json:"personId,omitempty"`
	FirstName  *string `json:"firstName,omitempty"`
	LastName   *string `json:"lastName,omitempty"`
	Name       *string `json:"name,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	Fax        *string `json:"fax,omitempty"`
	Email      *string `json:"email,omitempty"`
	ExternalID *string `json:"externalId,omitempty"`
	IsActive   bool    `json:"isActive"`
}

// SearchResultDTO mirrors the {providers, totalCount, page, pageSize}
// envelope GET provider/search returns.
type SearchResultDTO struct {
	Providers  []ProviderSearchResultDTO `json:"providers"`
	TotalCount int64                     `json:"totalCount"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"pageSize"`
}
