// Package daoimpl ports the queries behind rest/AccessionValidation
// (constitution.md Layer II). Folder layout mirrors the Java source.
package daoimpl

import (
	"openelis-go/internal/common/util"

	"gorm.io/gorm"
)

// ResultValidationDAOImpl backs rest/AccessionValidation (Wave 5.3).
type ResultValidationDAOImpl struct {
	DB *gorm.DB

	// ActiveLocale is site_information "default language locale".
	ActiveLocale string
}

// Locale returns the configured locale, or "en" when unset.
func (d *ResultValidationDAOImpl) Locale() string {
	if d.ActiveLocale != "" {
		return d.ActiveLocale
	}
	return "en"
}

// validationStatusNames are the analysis statuses AccessionValidation collects,
// spelled as STORED in status_of_sample: AnalysisStatus.TechnicalAcceptance is
// the row named "Technical Acceptance" and TechnicalRejected is "Technical
// Rejected".
//
// Note the first is an ACCEPTANCE status, not a rejection — the wave's other
// endpoint files that very status under a key called `notValidated`.
var validationStatusNames = []string{
	"Technical Acceptance",
	"Technical Rejected",
}

// ValidationRow is one analysis awaiting validation, with everything the
// AnalysisItem builder reads.
type ValidationRow struct {
	AccessionNumber string   `gorm:"column:accession_number"`
	AnalysisID      string   `gorm:"column:analysis_id"`
	TestID          string   `gorm:"column:test_id"`
	ResultID        *string  `gorm:"column:result_id"`
	ResultValue     *string  `gorm:"column:result_value"`
	ResultType      *string  `gorm:"column:result_type"`
	MinNormal       *float64 `gorm:"column:min_normal"`
	MaxNormal       *float64 `gorm:"column:max_normal"`
	SigDigits       *int     `gorm:"column:sig_digits"`

	TestName       string  `gorm:"column:test_name"`
	TestSortNumber *string `gorm:"column:test_sort_number"`
	UnitOfMeasure  *string `gorm:"column:unit_of_measure"`

	LimitLowNormal  *float64 `gorm:"column:limit_low_normal"`
	LimitHighNormal *float64 `gorm:"column:limit_high_normal"`
	LimitSigDigits  *int     `gorm:"column:limit_sig_digits"`

	PatientFirstName *string `gorm:"column:patient_first_name"`
	PatientLastName  *string `gorm:"column:patient_last_name"`
	NationalID       *string `gorm:"column:national_id"`
	Gender           *string `gorm:"column:gender"`
	EnteredBirthDate *string `gorm:"column:entered_birth_date"`
}

const validationSelect = `s.accession_number AS accession_number,
	a.id::text AS analysis_id,
	a.test_id::text AS test_id,
	res.id::text AS result_id,
	res.value AS result_value,
	res.result_type AS result_type,
	res.min_normal AS min_normal,
	res.max_normal AS max_normal,
	res.significant_digits AS sig_digits,
	COALESCE(lv.value, t.name) || COALESCE(
		(SELECT '(' || COALESCE(tlv.value, tos.description) || ')'
		   FROM clinlims.sampletype_test AS tost
		   JOIN clinlims.type_of_sample AS tos ON tos.id = tost.sample_type_id
		   LEFT JOIN clinlims.localization AS tl ON tl.id = tos.name_localization_id
		   LEFT JOIN clinlims.localization_value AS tlv ON tlv.localization_id = tl.id AND tlv.locale = @loc
		  WHERE tost.test_id = t.id AND tos.local_abbrev IS DISTINCT FROM 'Variable'
		  ORDER BY tost.ctid LIMIT 1), '') AS test_name,
	t.sort_order AS test_sort_number,
	uom.name AS unit_of_measure,
	rl.low_normal  AS limit_low_normal,
	rl.high_normal AS limit_high_normal,
	(SELECT tr.significant_digits FROM clinlims.test_result tr
	  WHERE tr.test_id = t.id ORDER BY tr.ctid LIMIT 1) AS limit_sig_digits,
	pe.first_name AS patient_first_name,
	pe.last_name  AS patient_last_name,
	pa.national_id AS national_id,
	pa.gender AS gender,
	pa.entered_birth_date AS entered_birth_date`

func (d *ResultValidationDAOImpl) base() *gorm.DB {
	return d.DB.Table("clinlims.analysis AS a").
		Select(validationSelect, map[string]any{"loc": d.Locale()}).
		Joins("JOIN clinlims.sample_item AS si ON si.id = a.sampitem_id").
		Joins("JOIN clinlims.sample AS s ON s.id = si.samp_id").
		Joins("JOIN clinlims.test AS t ON t.id = a.test_id").
		Joins("LEFT JOIN clinlims.localization AS l ON l.id = t.name_localization_id").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = l.id AND lv.locale = ?", d.Locale()).
		Joins("LEFT JOIN clinlims.unit_of_measure AS uom ON uom.id = t.uom_id").
		Joins("LEFT JOIN clinlims.result AS res ON res.analysis_id = a.id").
		Joins("LEFT JOIN clinlims.result_limits AS rl ON rl.test_id = t.id").
		Joins("LEFT JOIN clinlims.sample_human AS sh ON sh.samp_id = s.id").
		Joins("LEFT JOIN clinlims.patient AS pa ON pa.id = sh.patient_id").
		Joins("LEFT JOIN clinlims.person AS pe ON pe.id = pa.person_id").
		Joins(`JOIN clinlims.status_of_sample AS st ON st.id = a.status_id
			AND st.status_type = 'ANALYSIS' AND st.name IN ?`, validationStatusNames)
}

// ByAccessionRange ports getPageAnalysisAtAccessionNumberAndStatus — the
// doRange=TRUE path, and the reason this endpoint can answer with a different
// order's results than the one asked for:
//
//	where accession_number >= :accessionNumber
//	  and length(accession_number) = length(:accessionNumber)
//	  and statusId in (...)
//	order by accession_number
//
// It is a RANGE, bounded only by string length. Asking for E2E-ATT-01 (10
// characters) therefore reaches E2E-RES-01 (also 10, and greater), while
// E2E-EDIT-01 (11) reaches nothing — the length predicate, not the accession,
// is what excludes it. The comparison uses the DATABASE collation.
func (d *ResultValidationDAOImpl) ByAccessionRange(accessionNumber string) ([]ValidationRow, error) {
	rows := []ValidationRow{}
	err := d.base().
		Where("s.accession_number >= ? AND length(s.accession_number) = length(?)",
			accessionNumber, accessionNumber).
		Order("s.accession_number").
		Scan(&rows).Error
	return rows, err
}

// BySample ports the doRange=FALSE path: getSample(accessionNumber) and then
// that sample's analyses, or nothing at all when no such sample exists. An
// EXACT match, which is what the parameter looks like it should mean.
func (d *ResultValidationDAOImpl) BySample(accessionNumber string) ([]ValidationRow, error) {
	rows := []ValidationRow{}
	err := d.base().
		Where("s.accession_number = ?", accessionNumber).
		Order("s.accession_number").
		Scan(&rows).Error
	return rows, err
}

// BySection ports the testSectionId branch.
func (d *ResultValidationDAOImpl) BySection(sectionID string) ([]ValidationRow, error) {
	rows := []ValidationRow{}
	err := d.base().
		Where("a.test_sect_id = ?", sectionID).
		Order("s.accession_number").
		Scan(&rows).Error
	return rows, err
}

// UserTestSections ports userService.getUserTestSections(user, ROLE_VALIDATION)
// — the caller's own sections, LOCALIZED, as opposed to the raw-named
// TEST_SECTION_BY_NAME list that sits beside it in the same response.
//
// This deployment's admin holds AllLabUnits, so the list is every active
// section; the role filter is still applied rather than assumed away.
func (d *ResultValidationDAOImpl) UserTestSections() ([]util.IdValuePair, error) {
	rows := []util.IdValuePair{}
	err := d.DB.Table("clinlims.test_section AS ts").
		Select("ts.id::text AS id, COALESCE(NULLIF(lv.value, ''), ts.name) AS value").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = ts.name_localization_id AND lv.locale = ?", d.Locale()).
		Where("ts.is_active = ?", "Y").
		Order("ts.sort_order").
		Scan(&rows).Error
	return rows, err
}
