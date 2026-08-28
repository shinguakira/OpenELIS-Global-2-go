// Package daoimpl ports org.openelisglobal.patient.daoimpl (+ the person and
// sample-human lookups the ported endpoints need). Folder layout mirrors the
// Java source during migration.
//
// This is the only layer allowed to import GORM.
package daoimpl

import (
	"errors"

	"gorm.io/gorm"

	"openelis-go/internal/patient/valueholder"
	personvh "openelis-go/internal/person/valueholder"
)

// PatientDAOImpl ports PatientDAOImpl + PersonDAOImpl + the SampleHumanDAOImpl
// lookup that patientByLabNumer needs.
type PatientDAOImpl struct {
	DB *gorm.DB
}

// GetPatientByAccessionNumber mirrors the two-step chain behind
// GET rest/patientByLabNumer:
//
//	SampleDAOImpl.getSampleByAccessionNumber  (HQL: from Sample where accession_number = :n)
//	SampleHumanDAOImpl.getPatientForSample    (HQL: join Patient/SampleHuman on sampleId)
//
// Collapsed into a single join here — Java issues two queries only because it
// hands a Sample entity between two services; the resulting rows are identical
// and one round trip is strictly better. Returns (nil, nil) when the accession
// does not exist OR exists with no linked patient: Java answers 404 for both,
// so the two cases do not need to be distinguished.
func (d *PatientDAOImpl) GetPatientByAccessionNumber(accessionNumber string) (*valueholder.Patient, error) {
	var p valueholder.Patient
	err := d.DB.Table("clinlims.patient AS pat").
		Select("pat.*").
		Joins("JOIN clinlims.sample_human sh ON sh.patient_id = pat.id").
		Joins("JOIN clinlims.sample s ON s.id = sh.samp_id").
		Where("s.accession_number = ?", accessionNumber).
		Order("pat.id ASC").
		Limit(1).
		Scan(&p).Error
	if err != nil {
		return nil, err
	}
	if p.ID == 0 {
		return nil, nil
	}
	return &p, nil
}

// GetPersonByID mirrors personService.get(id). Returns (nil, nil) on a miss.
func (d *PatientDAOImpl) GetPersonByID(id int64) (*personvh.Person, error) {
	var person personvh.Person
	result := d.DB.First(&person, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &person, nil
}

// GetPatientByID mirrors patientDAO.getData(id). Returns (nil, nil) on a miss.
//
// NOTE: id is taken as a string, not an int64, on purpose. Java's Patient id
// uses LIMSStringNumberUserType, whose nullSafeSet does Integer.parseInt — so a
// NON-NUMERIC id throws NumberFormatException and surfaces as HTTP 500, while a
// numeric-but-absent id is a clean 404. Callers reproduce that split; see
// PatientService.GetMergeDetails.
func (d *PatientDAOImpl) GetPatientByID(id string) (*valueholder.Patient, error) {
	var p valueholder.Patient
	result := d.DB.First(&p, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &p, nil
}

// GetIdDocuments mirrors PatientIdDocumentDAOImpl.getByPatientId:
// FROM PatientIdDocument d WHERE d.patientId = :patientId AND d.deleted = false
// No ORDER BY in Java, so none here — order is DB-natural either way.
func (d *PatientDAOImpl) GetIdDocuments(patientID string) ([]valueholder.PatientIdDocument, error) {
	docs := []valueholder.PatientIdDocument{}
	err := d.DB.Where("patient_id = ? AND deleted = false", patientID).Find(&docs).Error
	if err != nil {
		return nil, err
	}
	return docs, nil
}

// GetPhoto mirrors PatientPhotoDAOImpl's uniqueResult() lookup by patient id.
// Returns (nil, nil) when the patient has no photo row.
func (d *PatientDAOImpl) GetPhoto(patientID string) (*valueholder.PatientPhoto, error) {
	photos := []valueholder.PatientPhoto{}
	if err := d.DB.Where("patient_id = ?", patientID).Limit(1).Find(&photos).Error; err != nil {
		return nil, err
	}
	if len(photos) == 0 {
		return nil, nil
	}
	return &photos[0], nil
}
