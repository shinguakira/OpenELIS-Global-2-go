package daoimpl

import (
	"time"

	"gorm.io/gorm"
)

// SampleEditDAOImpl backs GET rest/SampleEdit (Wave 4.7).
type SampleEditDAOImpl struct {
	DB *gorm.DB
}

// SampleEditItemRow is one sample_item that survived the status filter.
type SampleEditItemRow struct {
	ID             string     `gorm:"column:id"`
	SortOrder      string     `gorm:"column:sort_order"`
	TypeOfSampleID *string    `gorm:"column:typeosamp_id"`
	TypeOfSample   *string    `gorm:"column:type_of_sample_name"`
	CollectionDate *time.Time `gorm:"column:collection_date"`
}

// EnteredSampleItems ports SampleEditRestController.getSampleItems, which is
// getSampleItemsBySampleIdAndStatus(sampleId, ENTERED_STATUS_SAMPLE_LIST).
//
// ENTERED_STATUS_SAMPLE_LIST holds exactly one id: the status_of_sample row
// named `SampleEntered` (SampleStatus.Entered via IStatusService). Resolved by
// NAME here rather than by the id this database happens to use, because
// IStatusService resolves it by name too.
//
// This filter is the single most load-bearing thing on the endpoint: it decides
// existingTests, possibleTests AND the maxAccessionNumber suffix. No row in the
// stock dataset carries that status, which is why
// src/test/resources/fixtures/sample-edit-e2e.sql exists.
func (d *SampleEditDAOImpl) EnteredSampleItems(accessionNumber string) ([]SampleEditItemRow, error) {
	rows := []SampleEditItemRow{}
	err := d.DB.Table("clinlims.sample_item AS si").
		Select(`si.id, si.sort_order, si.typeosamp_id, si.collection_date,
			COALESCE(NULLIF(lv.value, ''), tos.description) AS type_of_sample_name`).
		Joins("JOIN clinlims.sample AS s ON s.id = si.samp_id").
		Joins("LEFT JOIN clinlims.type_of_sample AS tos ON tos.id = si.typeosamp_id").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = tos.name_localization_id AND lv.locale = 'en'").
		Joins(`JOIN clinlims.status_of_sample AS sos ON sos.id = si.status_id
			AND sos.status_type = 'SAMPLE' AND sos.name = 'SampleEntered'`).
		Where("s.accession_number = ?", accessionNumber).
		Order("si.sort_order::numeric").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// SampleEditAnalysisRow is one analysis behind an existingTests row.
type SampleEditAnalysisRow struct {
	AnalysisID    string `gorm:"column:analysis_id"`
	TestID        string `gorm:"column:test_id"`
	TestName      string `gorm:"column:test_name"`
	TestSortOrder string `gorm:"column:test_sort_order"`
	StatusName    string `gorm:"column:status_name"`
}

// AnalysesForItem ports getAnalysesBySampleItemsExcludingByStatusIds with
// excludedAnalysisStatusList = {AnalysisStatus.Canceled}.
//
// Only CANCELED is excluded — not "not started", not rejected. The test name is
// the localized one (TestServiceImpl.getUserLocalizedTestName), and sortOrder on
// the row is the TEST's sort order, not the sample item's.
func (d *SampleEditDAOImpl) AnalysesForItem(sampleItemID string) ([]SampleEditAnalysisRow, error) {
	rows := []SampleEditAnalysisRow{}
	err := d.DB.Table("clinlims.analysis AS a").
		Select(`a.id AS analysis_id, t.id AS test_id,
			COALESCE(NULLIF(lv.value, ''), t.name) AS test_name,
			COALESCE(t.sort_order::text, '') AS test_sort_order,
			sos.name AS status_name`).
		Joins("JOIN clinlims.test AS t ON t.id = a.test_id").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = t.name_localization_id AND lv.locale = 'en'").
		Joins("JOIN clinlims.status_of_sample AS sos ON sos.id = a.status_id").
		Where("a.sampitem_id = ?", sampleItemID).
		Where(`a.status_id <> (SELECT id FROM clinlims.status_of_sample
			WHERE status_type = 'ANALYSIS' AND name = 'Test Canceled')`).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// SampleEditPossibleTestRow is one candidate test for a sample item's type.
type SampleEditPossibleTestRow struct {
	TestID    string `gorm:"column:test_id"`
	TestName  string `gorm:"column:test_name"`
	SortOrder string `gorm:"column:sort_order"`
}

// PossibleTestsForSampleType ports addPossibleTestsToList: every
// sampletype_test row for the item's sample type whose test is ACTIVE and
// ORDERABLE.
//
// It does NOT exclude tests already ordered on the item — the list is "what this
// sample type supports", not "what is still addable", despite the name.
// Confirmed live: E2E-EDIT-01 has 2 items of the same type and returns 8 rows,
// 4 per item, including the ones already in existingTests.
func (d *SampleEditDAOImpl) PossibleTestsForSampleType(sampleTypeID string) ([]SampleEditPossibleTestRow, error) {
	rows := []SampleEditPossibleTestRow{}
	err := d.DB.Table("clinlims.sampletype_test AS st").
		Select(`t.id AS test_id,
			COALESCE(NULLIF(lv.value, ''), t.name) AS test_name,
			COALESCE(t.sort_order::text, '') AS sort_order`).
		Joins("JOIN clinlims.test AS t ON t.id = st.test_id").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = t.name_localization_id AND lv.locale = 'en'").
		Where("st.sample_type_id = ? AND t.is_active = 'Y' AND t.orderable = true", sampleTypeID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// SampleEditPatientRow is the patient block SampleEdit emits.
//
// Every field is a plain string because the form initialises them to "" — an
// accession with no patient still emits the keys, empty.
type SampleEditPatientRow struct {
	PatientID     string `gorm:"column:patient_id"`
	FirstName     string `gorm:"column:first_name"`
	LastName      string `gorm:"column:last_name"`
	DOB           string `gorm:"column:dob"`
	Gender        string `gorm:"column:gender"`
	NationalID    string `gorm:"column:national_id"`
	SubjectNumber string `gorm:"column:subject_number"`
}

// PatientForAccession returns the patient linked to the sample, or nil.
//
// `dob` is the STORED entered_birth_date, emitted RAW — getEnteredDOB does not
// reformat, unlike order/search which runs the same column through
// DateUtil.formatStringDate.
func (d *SampleEditDAOImpl) PatientForAccession(accessionNumber string) (*SampleEditPatientRow, error) {
	rows := []SampleEditPatientRow{}
	err := d.DB.Table("clinlims.patient AS p").
		Select(`p.id AS patient_id,
			COALESCE(pe.first_name, '') AS first_name,
			COALESCE(pe.last_name, '')  AS last_name,
			COALESCE(p.entered_birth_date, '') AS dob,
			COALESCE(p.gender, '') AS gender,
			COALESCE(p.national_id, '') AS national_id,
			'' AS subject_number`).
		Joins("JOIN clinlims.person AS pe ON pe.id = p.person_id").
		Joins("JOIN clinlims.sample_human AS sh ON sh.patient_id = p.id").
		Joins("JOIN clinlims.sample AS s ON s.id = sh.samp_id").
		Where("s.accession_number = ?", accessionNumber).
		Limit(1).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

// SampleEditSampleRow is the sample itself.
type SampleEditSampleRow struct {
	ID             string     `gorm:"column:id"`
	Priority       *string    `gorm:"column:order_priority"`
	IsConfirmation *bool      `gorm:"column:is_confirmation"`
	ReceivedDate   *time.Time `gorm:"column:received_date"`
}

// SampleByAccession returns the sample row, or nil when nothing matches — the
// signal for the noSampleFound branch.
func (d *SampleEditDAOImpl) SampleByAccession(accessionNumber string) (*SampleEditSampleRow, error) {
	rows := []SampleEditSampleRow{}
	err := d.DB.Table("clinlims.sample").
		Select("id, order_priority, is_confirmation, received_date").
		Where("accession_number = ?", accessionNumber).
		Limit(1).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}
