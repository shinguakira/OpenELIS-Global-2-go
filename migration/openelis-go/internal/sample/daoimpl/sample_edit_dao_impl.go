package daoimpl

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// SampleEditDAOImpl backs GET rest/SampleEdit (Wave 4.7).
type SampleEditDAOImpl struct {
	DB *gorm.DB
	// ActiveLocale is site_information."default language locale" (language
	// subtag). Empty falls back to "en". Not the literal 'en' this DAO used to
	// repeat: a non-English deployment got English type and test names.
	ActiveLocale string
}

// Locale returns the configured locale, or "en" when unset.
func (d *SampleEditDAOImpl) Locale() string {
	if d.ActiveLocale != "" {
		return d.ActiveLocale
	}
	return "en"
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
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = tos.name_localization_id AND lv.locale = ?", d.Locale()).
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
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = t.name_localization_id AND lv.locale = ?", d.Locale()).
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
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = t.name_localization_id AND lv.locale = ?", d.Locale()).
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
//
// subjectNumber is a real lookup, NOT the empty string this port first
// hardcoded: setPatientInfo calls getSubjectNumber, which reads the patient's
// PatientIdentity row of type SUBJECT and falls back to "" when there is none.
// The type is resolved BY NAME because TableIdService looks it up by name too —
// its id is 9 on this database and need not be elsewhere.
func (d *SampleEditDAOImpl) PatientForAccession(accessionNumber string) (*SampleEditPatientRow, error) {
	rows := []SampleEditPatientRow{}
	err := d.DB.Table("clinlims.patient AS p").
		Select(`p.id AS patient_id,
			COALESCE(pe.first_name, '') AS first_name,
			COALESCE(pe.last_name, '')  AS last_name,
			COALESCE(p.entered_birth_date, '') AS dob,
			COALESCE(p.gender, '') AS gender,
			COALESCE(p.national_id, '') AS national_id,
			COALESCE(subj.identity_data, '') AS subject_number`).
		Joins(`LEFT JOIN clinlims.patient_identity AS subj
			ON subj.patient_id = p.id
			AND subj.identity_type_id = (
			    SELECT id FROM clinlims.patient_identity_type
			     WHERE identity_type = 'SUBJECT'
			)`).
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

// MostRecentAccessionForPatient ports
// SampleEditRestController.getMostRecentAccessionNumberForPaitient.
//
// "Most recent" is the HIGHEST SAMPLE ID, not the latest date: Java walks every
// sample for the patient keeping `Integer.parseInt(sample.getId()) > maxId`.
// Ordering by entered_date or collection_date would pick a different row
// whenever ids and dates disagree.
//
// Returns "" when the patient has no samples, which leaves the caller on the
// blank-form branch exactly as a blank accessionNumber does.
func (d *SampleEditDAOImpl) MostRecentAccessionForPatient(patientID string) (string, error) {
	rows := []string{}
	err := d.DB.Table("clinlims.sample AS s").
		Select("s.accession_number").
		Joins("JOIN clinlims.sample_human AS sh ON sh.samp_id = s.id").
		Where("sh.patient_id = ? AND s.accession_number IS NOT NULL", patientID).
		Order("s.id::numeric DESC").
		Limit(1).
		Pluck("accession_number", &rows).Error
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", nil
	}
	return rows[0], nil
}

// ---------------------------------------------------------------------------
// The `if (sample != null)` half of SampleOrderService.getSampleOrderItem.
//
// getBaseSampleOrderItem stamps receivedDate/receivedTime/requestDate from the
// CLOCK and attaches the reference lists; then, when the accession resolves to
// a sample, that block OVERWRITES both dates and fills in ~15 further fields
// from the sample. An earlier version of this port stopped at the base item,
// which is correct only for a form load with no sample behind it.
// ---------------------------------------------------------------------------

// SampleEditObservation is one observation_history row: the value plus the
// value_type that decides how it is rendered.
type SampleEditObservation struct {
	Value     string `gorm:"column:value"`
	ValueType string `gorm:"column:value_type"`
	TypeName  string `gorm:"column:type_name"`
}

// ObservationsForSample returns the sample's observation history keyed by type
// name, carrying value_type so callers can pick between the raw value and the
// dictionary-resolved one.
func (d *SampleEditDAOImpl) ObservationsForSample(sampleID string) (map[string]SampleEditObservation, error) {
	rows := []SampleEditObservation{}
	err := d.DB.Table("clinlims.observation_history AS oh").
		Select("oh.value AS value, oh.value_type AS value_type, oht.type_name AS type_name").
		Joins("JOIN clinlims.observation_history_type AS oht ON oht.id = oh.observation_history_type_id").
		Where("oh.sample_id = ? AND oh.value IS NOT NULL", sampleID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]SampleEditObservation, len(rows))
	for _, r := range rows {
		out[r.TypeName] = r
	}
	return out, nil
}

// DictionaryLocalizedName resolves a dictionary id to its localized name, for
// the non-LITERAL branch of getValueForObservation.
//
// dictionaryService.getDataForId(...).getLocalizedName() falls back to
// dict_entry when there is no localization row, which is what the COALESCE
// reproduces.
func (d *SampleEditDAOImpl) DictionaryLocalizedName(dictID string) (string, error) {
	names := []string{}
	err := d.DB.Table("clinlims.dictionary AS dict").
		Select("COALESCE(lv.value, dict.dict_entry)").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = dict.name_localization_id AND lv.locale = ?", d.Locale()).
		Where("dict.id = ?", dictID).
		Limit(1).
		Scan(&names).Error
	if err != nil {
		return "", err
	}
	if len(names) == 0 {
		return "", nil
	}
	return names[0], nil
}

// SampleEditRequesterPersonRow is the person behind a PROVIDER-type
// sample_requester row, plus the provider that person maps to.
type SampleEditRequesterPersonRow struct {
	PersonID   string  `gorm:"column:person_id"`
	FirstName  *string `gorm:"column:first_name"`
	LastName   *string `gorm:"column:last_name"`
	WorkPhone  *string `gorm:"column:work_phone"`
	Fax        *string `gorm:"column:fax"`
	Email      *string `gorm:"column:email"`
	ProviderID *string `gorm:"column:provider_id"`
}

// RequesterPerson ports SampleServiceImpl.getPersonRequester plus the
// getProviderByPerson lookup SampleOrderService runs on the result.
//
// Two traps, both reproduced:
//
//   - the requester type named "provider" is the PERSON type here.
//     getPersonRequester compares against PERSON_REQUESTER_TYPE_ID, which
//     initializeGlobalVariables resolves from requester_type WHERE
//     requester_type = 'provider'. So requester_type_id = 2 means "the
//     requester_id is a person id", not "a provider id".
//   - requester_id is therefore read as person.id directly. It is NOT
//     sample_human.provider_id, which is what order/search uses — the two
//     endpoints source the provider from different tables entirely.
//
// Types are matched by NAME because TableIdService looks them up by name; the
// numeric ids are deployment data.
func (d *SampleEditDAOImpl) RequesterPerson(sampleID string) (*SampleEditRequesterPersonRow, error) {
	rows := []SampleEditRequesterPersonRow{}
	err := d.DB.Table("clinlims.sample_requester AS sr").
		Select(`pe.id::text     AS person_id,
			pe.first_name  AS first_name,
			pe.last_name   AS last_name,
			pe.work_phone  AS work_phone,
			pe.fax         AS fax,
			pe.email       AS email,
			pr.id::text    AS provider_id`).
		Joins("JOIN clinlims.requester_type AS rt ON rt.id = sr.requester_type_id").
		Joins("JOIN clinlims.person AS pe ON pe.id = sr.requester_id").
		Joins("LEFT JOIN clinlims.provider AS pr ON pr.person_id = pe.id").
		Where("sr.sample_id = ? AND rt.requester_type = ?", sampleID, "provider").
		// getRequestersForSampleId is an unordered criteria query and the loop
		// returns the FIRST match, so physical order decides.
		Order("sr.ctid").
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

// SampleEditRequesterOrgRow is an organization requester.
type SampleEditRequesterOrgRow struct {
	ID   string  `gorm:"column:id"`
	Name *string `gorm:"column:name"`
	Code *string `gorm:"column:code"`
}

// RequesterOrganization ports SampleServiceImpl.getOrganizationRequester: walk
// the sample's requester rows, keep the ORGANIZATION-typed ones, and return the
// first whose organization carries the requested organization type.
//
// `code` is organization.code — NOT short_name. order/search's
// buildSampleOrderItems emits getShortName() under the same
// `referringSiteCode` key, so the two endpoints read different columns for what
// looks like one field. On this deployment the referring clinic has
// short_name '279' and a NULL code, which is why order/search emits the key
// and SampleEdit does not.
func (d *SampleEditDAOImpl) RequesterOrganization(sampleID, orgTypeShortName string) (*SampleEditRequesterOrgRow, error) {
	rows := []SampleEditRequesterOrgRow{}
	err := d.DB.Table("clinlims.sample_requester AS sr").
		Select("o.id::text AS id, o.name AS name, o.code AS code").
		Joins("JOIN clinlims.requester_type AS rt ON rt.id = sr.requester_type_id").
		Joins("JOIN clinlims.organization AS o ON o.id = sr.requester_id").
		Joins("JOIN clinlims.organization_organization_type AS oot ON oot.org_id = o.id").
		Joins("JOIN clinlims.organization_type AS ot ON ot.id = oot.org_type_id").
		Where("sr.sample_id = ? AND rt.requester_type = ? AND ot.short_name = ?",
			sampleID, "organization", orgTypeShortName).
		Order("sr.ctid").
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

// ProgramIDForSample is SampleEdit's program lookup, and it is NOT the one
// order/search uses.
//
// Both call getProgrammeSampleBySample, which routes by program NAME to a
// TABLE_PER_CLASS subclass table. But order/search wraps it in a fallback that
// resolves the id by name when the subclass table has no row, and
// SampleOrderService has no such fallback: `if (programSample != null)` and
// nothing else. So for a sample whose program name names a subclass,
// order/search emits a programId and SampleEdit emits none — from the same two
// rows.
func (d *SampleEditDAOImpl) ProgramIDForSample(sampleID, programName string) (string, error) {
	table := "clinlims.program_sample"
	switch n := strings.ToLower(programName); {
	case strings.Contains(n, "pathology"):
		table = "clinlims.pathology_sample"
	case strings.Contains(n, "immunohistochemistry"):
		table = "clinlims.immunohistochemistry_sample"
	case strings.Contains(n, "cytology"):
		table = "clinlims.cytology_sample"
	}
	ids := []string{}
	err := d.DB.Table(table).
		Select("program_id::text").
		Where("sample_id = ? AND program_id IS NOT NULL", sampleID).
		Limit(1).
		Scan(&ids).Error
	if err != nil {
		return "", err
	}
	if len(ids) == 0 {
		return "", nil
	}
	return ids[0], nil
}
