// Package service ports org.openelisglobal.patient.service (the read paths
// behind the c1 endpoints). Folder layout mirrors the Java source during
// migration.
//
// Per constitution.md Layer III, this layer compiles the COMPLETE response DTO;
// the controller (Layer IV) only maps request/response. No GORM import here.
package service

import (
	"strconv"
	"time"

	"openelis-go/internal/patient/daoimpl"
	"openelis-go/internal/patient/form"
	"openelis-go/internal/patient/valueholder"
	personvh "openelis-go/internal/person/valueholder"
)

// PatientService holds no DB handle — all access goes through the DAO.
type PatientService struct {
	DAO *daoimpl.PatientDAOImpl
}

// ErrMalformedID signals a patient id that Java would have handed to
// Integer.parseInt inside LIMSStringNumberUserType, producing a
// NumberFormatException and therefore HTTP 500 (NOT 404). The controller maps
// this to 500 so the port reproduces Java's split exactly:
//
//	numeric-but-absent id -> 404
//	non-numeric id        -> 500
//
// Reproducing a 500 is deliberate. It is a Java bug, and a migration pins bugs
// rather than fixing them.
type ErrMalformedID struct{ ID string }

func (e ErrMalformedID) Error() string { return "malformed patient id: " + e.ID }

// GetPatientByAccessionNumber backs GET rest/patientByLabNumer.
// Returns (nil, nil) when no patient is linked to that accession — Java
// answers 404 both when the sample is missing and when it has no patient.
func (s *PatientService) GetPatientByAccessionNumber(accessionNumber string) (*form.PatientDTO, error) {
	p, err := s.DAO.GetPatientByAccessionNumber(accessionNumber)
	if err != nil || p == nil {
		return nil, err
	}
	person, err := s.DAO.GetPersonByID(p.PersonID)
	if err != nil {
		return nil, err
	}
	dto := toPatientDTO(*p, person)
	return &dto, nil
}

// GetIdDocuments backs GET rest/patient-id-documents/{patientId}.
// Always returns a non-nil slice so an empty result serializes as [] rather
// than null — Java answers 200 [] for an unknown patient, never 404.
func (s *PatientService) GetIdDocuments(patientID string) ([]form.IdDocumentDTO, error) {
	docs, err := s.DAO.GetIdDocuments(patientID)
	if err != nil {
		return nil, err
	}
	dtos := make([]form.IdDocumentDTO, 0, len(docs))
	for _, d := range docs {
		dto := form.IdDocumentDTO{
			ID: d.ID,
			// Java concatenates unconditionally, so this is always a data: URI
			// string and the key is always present.
			Thumbnail:   "data:" + d.DocumentType + ";base64," + d.ThumbnailData,
			Category:    d.DocumentCategory,
			Description: d.Description,
		}
		if d.LastUpdated != nil {
			ms := d.LastUpdated.UnixMilli()
			dto.LastUpdated = &ms
		}
		dtos = append(dtos, dto)
	}
	return dtos, nil
}

// GetIdDocumentFull backs GET rest/patient-id-documents/{patientId}/{documentId}/full.
//
// Java fetches every non-deleted document for the patient and then scans in
// Java for the matching id — it never queries by document id. The observable
// consequence, which this reproduces: asking for a documentId that belongs to a
// DIFFERENT patient yields an empty data string rather than 403/404 or that
// document. Always 200; the empty string is how "not found" is expressed.
func (s *PatientService) GetIdDocumentFull(patientID string, documentID int64) (form.DataEnvelopeDTO, error) {
	docs, err := s.DAO.GetIdDocuments(patientID)
	if err != nil {
		return form.DataEnvelopeDTO{}, err
	}
	for _, d := range docs {
		if d.ID == documentID {
			return form.DataEnvelopeDTO{Data: "data:" + d.DocumentType + ";base64," + d.DocumentData}, nil
		}
	}
	return form.DataEnvelopeDTO{Data: ""}, nil
}

// GetPhoto backs GET rest/patient-photos/{id}/{isThumbnail}.
//
// THE KEY ASYMMETRY, and the single easiest thing to get wrong in this unit
// (PatientPhotoServiceImpl.java:116-119):
//
//	isThumbnail=false -> "data:<photoType>;base64,<photoData>"  (full data URI)
//	isThumbnail=true  -> "<thumbnailData>"                       (BARE base64)
//
// The thumbnail branch has NO data: prefix and reads a DIFFERENT column. The
// frontend depends on it — the avatar requests /true, the full view /false — so
// emitting a data: URI for both silently breaks the avatar. A missing photo
// yields an empty data string; Java cannot distinguish "no record" from "no
// thumbnail" here and neither does this.
func (s *PatientService) GetPhoto(patientID string, isThumbnail bool) (form.DataEnvelopeDTO, error) {
	photo, err := s.DAO.GetPhoto(patientID)
	if err != nil {
		return form.DataEnvelopeDTO{}, err
	}
	if photo == nil {
		return form.DataEnvelopeDTO{Data: ""}, nil
	}
	if isThumbnail {
		return form.DataEnvelopeDTO{Data: photo.ThumbnailData}, nil
	}
	return form.DataEnvelopeDTO{Data: "data:" + photo.PhotoType + ";base64," + photo.PhotoData}, nil
}

// --- DTO shaping (Layer III owns this, not the controller) ---

func toPersonDTO(p personvh.Person) form.PersonDTO {
	dto := form.PersonDTO{
		ID:            strconv.FormatInt(p.ID, 10),
		LastName:      p.LastName,
		FirstName:     p.FirstName,
		MiddleName:    p.MiddleName,
		MultipleUnit:  p.MultipleUnit,
		StreetAddress: p.StreetAddress,
		City:          p.City,
		State:         p.State,
		ZipCode:       p.ZipCode,
		Country:       p.Country,
		WorkPhone:     p.WorkPhone,
		HomePhone:     p.HomePhone,
		CellPhone:     p.CellPhone,
		PrimaryPhone:  p.PrimaryPhone,
		Fax:           p.Fax,
		Email:         p.Email,
		GpsLatitude:   p.GpsLatitude,
		GpsLongitude:  p.GpsLongitude,
	}
	if p.Lastupdated != nil {
		ms := p.Lastupdated.UnixMilli()
		dto.Lastupdated = &ms
	}
	return dto
}

// truncateToMidnightMillis reproduces Hibernate mapping birth_time/death_date
// as java.sql.Date: the clock component is DISCARDED even though the column
// stores one (the dev row holds 10:00:00 and Java emits midnight).
//
// Truncation is done in UTC to match the Java container, which docker-compose
// pins to TZ=UTC. Using the host's local zone here would shift the value by the
// UTC offset — the same class of bug already found and fixed in b2's
// generate-site-code.
func truncateToMidnightMillis(t time.Time) int64 {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC).UnixMilli()
}

func toPatientDTO(p valueholder.Patient, person *personvh.Person) form.PatientDTO {
	dto := form.PatientDTO{
		ID:                  strconv.FormatInt(p.ID, 10),
		Race:                p.Race,
		Gender:              p.Gender,
		BirthDateForDisplay: p.EnteredBirthDate,
		BirthPlace:          p.BirthPlace,
		EpiFirstName:        p.EpiFirstName,
		EpiMiddleName:       p.EpiMiddleName,
		EpiLastName:         p.EpiLastName,
		NationalID:          p.NationalID,
		Ethnicity:           p.Ethnicity,
		SchoolAttend:        p.SchoolAttend,
		MedicareID:          p.MedicareID,
		MedicaidID:          p.MedicaidID,
		ChartNumber:         p.ChartNumber,
		ExternalID:          p.ExternalID,
		UpidCode:            p.UpidCode,
		IsMerged:            p.IsMerged,
		FhirUUID:            p.FhirUUID,
	}
	if person != nil {
		pd := toPersonDTO(*person)
		dto.Person = &pd
	}
	// birth_date is echoed as stored — see the valueholder's note on why this
	// port does not "correct" the known write-time corruption.
	if p.BirthDate != nil {
		ms := p.BirthDate.UTC().UnixMilli()
		dto.BirthDate = &ms
	}
	if p.BirthTime != nil {
		ms := truncateToMidnightMillis(*p.BirthTime)
		dto.BirthTime = &ms
		// birthTimeForDisplay is NOT persisted — Java derives it at load time
		// with the DATE formatter (dd/MM/yyyy), which is why it can disagree
		// with birthDateForDisplay's stored text. Reproduced, not reconciled.
		s := p.BirthTime.UTC().Format("02/01/2006")
		dto.BirthTimeForDisplay = &s
	}
	if p.DeathDate != nil {
		ms := truncateToMidnightMillis(*p.DeathDate)
		dto.DeathDate = &ms
		s := p.DeathDate.UTC().Format("02/01/2006")
		dto.DeathDateForDisplay = &s
	}
	if p.MergedIntoPatientID != nil {
		s := strconv.FormatInt(*p.MergedIntoPatientID, 10)
		dto.MergedIntoPatientID = &s
	}
	if p.MergeDate != nil {
		ms := p.MergeDate.UTC().UnixMilli()
		dto.MergeDate = &ms
	}
	if p.Lastupdated != nil {
		ms := p.Lastupdated.UnixMilli()
		dto.Lastupdated = &ms
	}
	// Computed getter with no backing column: "" when the UUID is null, so the
	// key is ALWAYS present.
	if p.FhirUUID != nil {
		dto.FhirUUIDAsString = *p.FhirUUID
	}
	return dto
}
