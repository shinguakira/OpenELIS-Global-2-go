// Package valueholder ports org.openelisglobal.patient.valueholder.
// Folder layout mirrors the Java source during migration.
//
// Patient/PatientIdDocument/PatientPhoto are Hibernate-XML-mapped in Java
// (hibernate/hbm/Patient.hbm.xml); column names and nullability here come from
// the real schema (information_schema), not the hbm metadata.
package valueholder

import "time"

// Patient mirrors patient.valueholder.Patient. Maps to clinlims.patient.
//
// IMPORTANT — birth_date is stored WRONG in this database and this port must
// NOT try to correct it. Java's Patient.setBirthDateForDisplay writes back
// into birthDate using a LENIENT, locale-dependent parse, so an entered
// "01/15/1990" was persisted as 1991-03-01. The corruption already happened at
// write time. The READ path is faithful — it returns the stored column — and
// that is exactly what this port reproduces. Deriving birthDate from
// entered_birth_date here would "fix" a Java bug, which is out of scope for a
// migration (see the c1 e2e spec, which pins the faithful read).
//
// EnteredBirthDate backs the JSON field `birthDateForDisplay`. It is its own
// persisted column holding the literal text the user typed, which may contain
// "XX" when the day/month is unknown — never regenerate it from birth_date.
type Patient struct {
	ID                  int64      `gorm:"column:id"`
	PersonID            int64      `gorm:"column:person_id"`
	Race                *string    `gorm:"column:race"`
	Gender              *string    `gorm:"column:gender"`
	BirthDate           *time.Time `gorm:"column:birth_date"`
	EpiFirstName        *string    `gorm:"column:epi_first_name"`
	EpiMiddleName       *string    `gorm:"column:epi_middle_name"`
	EpiLastName         *string    `gorm:"column:epi_last_name"`
	BirthTime           *time.Time `gorm:"column:birth_time"`
	DeathDate           *time.Time `gorm:"column:death_date"`
	NationalID          *string    `gorm:"column:national_id"`
	Ethnicity           *string    `gorm:"column:ethnicity"`
	SchoolAttend        *string    `gorm:"column:school_attend"`
	MedicareID          *string    `gorm:"column:medicare_id"`
	MedicaidID          *string    `gorm:"column:medicaid_id"`
	BirthPlace          *string    `gorm:"column:birth_place"`
	Lastupdated         *time.Time `gorm:"column:lastupdated"`
	ExternalID          *string    `gorm:"column:external_id"`
	ChartNumber         *string    `gorm:"column:chart_number"`
	EnteredBirthDate    *string    `gorm:"column:entered_birth_date"`
	FhirUUID            *string    `gorm:"column:fhir_uuid"`
	UpidCode            *string    `gorm:"column:upid_code"`
	MergedIntoPatientID *int64     `gorm:"column:merged_into_patient_id"`
	IsMerged            bool       `gorm:"column:is_merged"` // NOT NULL
	MergeDate           *time.Time `gorm:"column:merge_date"`
}

// TableName pins the GORM table name.
func (Patient) TableName() string { return "clinlims.patient" }

// PatientIdDocument mirrors patient.valueholder.PatientIdDocument.
// Maps to clinlims.patient_id_document. Note patient_id is a VARCHAR column,
// not a numeric FK — a non-numeric id therefore matches nothing rather than
// erroring, which is why the endpoint answers 200 [] for garbage input.
type PatientIdDocument struct {
	ID               int64      `gorm:"column:id"`
	PatientID        string     `gorm:"column:patient_id"`
	DocumentData     string     `gorm:"column:document_data"`
	ThumbnailData    string     `gorm:"column:thumbnail_data"`
	DocumentType     string     `gorm:"column:document_type"`
	DocumentCategory string     `gorm:"column:document_category"`
	Description      *string    `gorm:"column:description"`
	Deleted          bool       `gorm:"column:deleted"`
	LastUpdated      *time.Time `gorm:"column:last_updated"`
}

// TableName pins the GORM table name.
func (PatientIdDocument) TableName() string { return "clinlims.patient_id_document" }

// PatientPhoto mirrors patient.valueholder.PatientPhoto.
// Maps to clinlims.patient_photo (patient_id is UNIQUE — at most one row per
// patient). Same VARCHAR patient_id note as PatientIdDocument.
type PatientPhoto struct {
	ID            int64      `gorm:"column:id"`
	PatientID     string     `gorm:"column:patient_id"`
	PhotoData     string     `gorm:"column:photo_data"`
	ThumbnailData string     `gorm:"column:thumbnail_data"`
	PhotoType     string     `gorm:"column:photo_type"`
	LastUpdated   *time.Time `gorm:"column:last_updated"`
}

// TableName pins the GORM table name.
func (PatientPhoto) TableName() string { return "clinlims.patient_photo" }
