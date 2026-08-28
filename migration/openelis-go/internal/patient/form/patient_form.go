// Package form ports org.openelisglobal.patient.form + the merge DTOs
// (constitution.md Layer V — "Forms/DTOs"): the client<->server wire shapes,
// kept out of both the valueholder (Layer I) and the controller (Layer IV).
// Folder layout mirrors the Java source during migration.
package form

import "openelis-go/internal/common/util"

// PersonDTO is the nested `person` object inside PatientDTO. Pointer fields
// with omitempty mirror Jackson's Include.NON_NULL: a null column is DROPPED
// from the JSON entirely rather than emitted as null.
type PersonDTO struct {
	Lastupdated   *int64  `json:"lastupdated,omitempty"`
	ID            string  `json:"id"`
	LastName      *string `json:"lastName,omitempty"`
	FirstName     *string `json:"firstName,omitempty"`
	MiddleName    *string `json:"middleName,omitempty"`
	MultipleUnit  *string `json:"multipleUnit,omitempty"`
	StreetAddress *string `json:"streetAddress,omitempty"`
	City          *string `json:"city,omitempty"`
	State         *string `json:"state,omitempty"`
	ZipCode       *string `json:"zipCode,omitempty"`
	Country       *string `json:"country,omitempty"`
	WorkPhone     *string `json:"workPhone,omitempty"`
	HomePhone     *string `json:"homePhone,omitempty"`
	CellPhone     *string `json:"cellPhone,omitempty"`
	PrimaryPhone  *string `json:"primaryPhone,omitempty"`
	Fax           *string `json:"fax,omitempty"`
	Email         *string `json:"email,omitempty"`
	// util.JavaDouble, not float64: Go renders 5.0 as `5` and Jackson renders
	// the same java.lang.Double as `5.0`. The two parse equal, so only the raw
	// bytes and Content-Length differ.
	GpsLatitude  *util.JavaDouble `json:"gpsLatitude,omitempty"`
	GpsLongitude *util.JavaDouble `json:"gpsLongitude,omitempty"`
}

// PatientDTO mirrors the raw Patient entity's JSON shape, which Java
// serializes directly from Hibernate with no DTO of its own.
//
// Two fields deserve attention when comparing against Java:
//   - BirthDate is the STORED birth_date column, echoed faithfully. It is
//     known to hold a corrupted value for rows written through the Java UI
//     (a lenient locale-dependent parse ran at write time). Reproducing the
//     faithful read is correct; correcting the value is not this port's job.
//   - BirthTime is truncated to midnight because Java maps the column as
//     java.sql.Date, discarding the clock even though the column stores one.
//     See PatientService.toPatientDTO.
//
// FhirUUIDAsString is a computed Java getter with no backing column: it
// returns "" when the UUID is null, so it is ALWAYS present and never omitted.
type PatientDTO struct {
	Lastupdated         *int64     `json:"lastupdated,omitempty"`
	ID                  string     `json:"id"`
	Race                *string    `json:"race,omitempty"`
	Gender              *string    `json:"gender,omitempty"`
	BirthDate           *int64     `json:"birthDate,omitempty"`
	BirthDateForDisplay *string    `json:"birthDateForDisplay,omitempty"`
	BirthTime           *int64     `json:"birthTime,omitempty"`
	BirthTimeForDisplay *string    `json:"birthTimeForDisplay,omitempty"`
	BirthPlace          *string    `json:"birthPlace,omitempty"`
	DeathDate           *int64     `json:"deathDate,omitempty"`
	DeathDateForDisplay *string    `json:"deathDateForDisplay,omitempty"`
	EpiFirstName        *string    `json:"epiFirstName,omitempty"`
	EpiMiddleName       *string    `json:"epiMiddleName,omitempty"`
	EpiLastName         *string    `json:"epiLastName,omitempty"`
	NationalID          *string    `json:"nationalId,omitempty"`
	Ethnicity           *string    `json:"ethnicity,omitempty"`
	SchoolAttend        *string    `json:"schoolAttend,omitempty"`
	MedicareID          *string    `json:"medicareId,omitempty"`
	MedicaidID          *string    `json:"medicaidId,omitempty"`
	ChartNumber         *string    `json:"chartNumber,omitempty"`
	Person              *PersonDTO `json:"person,omitempty"`
	ExternalID          *string    `json:"externalId,omitempty"`
	UpidCode            *string    `json:"upidCode,omitempty"`
	MergedIntoPatientID *string    `json:"mergedIntoPatientId,omitempty"`
	IsMerged            bool       `json:"isMerged"`
	MergeDate           *int64     `json:"mergeDate,omitempty"`
	FhirUUID            *string    `json:"fhirUuid,omitempty"`
	FhirUUIDAsString    string     `json:"fhirUuidAsString"`
}

// IdDocumentDTO is one element of GET rest/patient-id-documents/{patientId}.
// Java builds this as a HashMap with exactly five puts; the full blob
// (document_data) is deliberately NOT among them — only /full returns it.
//
// Description and LastUpdated are pointers because Jackson's NON_NULL applies
// CONTENT inclusion to Maps: a null value drops the KEY entirely rather than
// emitting null. omitempty reproduces that.
type IdDocumentDTO struct {
	ID          int64   `json:"id"`
	Thumbnail   string  `json:"thumbnail"`
	Category    string  `json:"category"`
	Description *string `json:"description,omitempty"`
	LastUpdated *int64  `json:"lastUpdated,omitempty"`
}

// DataEnvelopeDTO is the {"data": "..."} shape returned by both binary
// endpoints (patient-photos and patient-id-documents/.../full). Java returns
// JSON here, never raw bytes and never an image/* Content-Type.
type DataEnvelopeDTO struct {
	Data string `json:"data"`
}

// IdentifierDTO is one element of PatientMergeDetailsDTO.identifiers.
// Java's class also declares a `system` field that getMergeDetails never
// populates, so NON_NULL drops it — it is intentionally absent here rather
// than emitted as an empty string.
type IdentifierDTO struct {
	IdentityType  string `json:"identityType"`
	IdentityValue string `json:"identityValue"`
}

// MergeDataSummaryDTO mirrors PatientMergeDataSummaryDTO.
//
// TotalDataItems is a COMPUTED getter in Java (not a stored field): the sum of
// orders + results + samples + documents + identifiers, deliberately excluding
// contacts/relations/auditEntries. Modelling it as a plain field and forgetting
// to populate it would silently drop it from the response.
//
// TotalDocuments/TotalContacts/TotalRelations/TotalAuditEntries are never set
// by getMergeDetails and stay 0 — they are primitives in Java, so they DO
// serialize (as 0) rather than being omitted.
type MergeDataSummaryDTO struct {
	TotalOrders              int      `json:"totalOrders"`
	ActiveOrders             int      `json:"activeOrders"`
	TotalResults             int      `json:"totalResults"`
	TotalSamples             int      `json:"totalSamples"`
	TotalDocuments           int      `json:"totalDocuments"`
	TotalIdentifiers         int      `json:"totalIdentifiers"`
	TotalContacts            int      `json:"totalContacts"`
	TotalRelations           int      `json:"totalRelations"`
	TotalAuditEntries        int      `json:"totalAuditEntries"`
	ConflictingFields        []string `json:"conflictingFields"`
	ConflictingIdentityTypes []string `json:"conflictingIdentityTypes"`
	TotalDataItems           int      `json:"totalDataItems"`
}

// MergeDetailsDTO mirrors PatientMergeDetailsDTO.
//
// NationalID/PhoneNumber/Email/Address exist on the Java class but are NEVER
// populated by getMergeDetails, so NON_NULL drops them. They are omitted here
// too — populating them would leak PHI this endpoint does not return, which
// the c1 e2e spec asserts as an ABSENCE.
type MergeDetailsDTO struct {
	PatientID         string              `json:"patientId"`
	FirstName         *string             `json:"firstName,omitempty"`
	LastName          *string             `json:"lastName,omitempty"`
	Gender            *string             `json:"gender,omitempty"`
	BirthDate         *string             `json:"birthDate,omitempty"`
	DataSummary       MergeDataSummaryDTO `json:"dataSummary"`
	Identifiers       []IdentifierDTO     `json:"identifiers"`
	ConflictingFields []string            `json:"conflictingFields"`
}
