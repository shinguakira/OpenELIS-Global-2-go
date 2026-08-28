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

	// ValidateRejected is site_information validateTechnicalRejection, read once
	// at startup the way ConfigurationProperties does. When false, technically
	// REJECTED analyses must NOT be offered for validation.
	ValidateRejected bool
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
// validationStatusName is added unconditionally.
//
// getValidationStatus adds TechnicalAcceptance always and TechnicalRejected
// ONLY when site_information validateTechnicalRejection is "true". Including
// the second unconditionally puts technically REJECTED analyses on the
// validation screen at sites that switched that off, offering them for an
// acceptance workflow they were meant to be excluded from.
//
// Nothing in the dev dataset was in the rejected status, so neither branch of
// the condition was observable and the shortcut matched byte-for-byte.
const (
	statusTechnicalAcceptance = "Technical Acceptance"
	statusTechnicalRejected   = "Technical Rejected"
)

// ValidationStatusNames returns the statuses this endpoint collects, honouring
// the rejected-tests setting.
func (d *ResultValidationDAOImpl) ValidationStatusNames() []string {
	names := []string{statusTechnicalAcceptance}
	if d.ValidateRejected {
		names = append(names, statusTechnicalRejected)
	}
	return names
}

// ValidationRow is one analysis awaiting validation, with everything the
// AnalysisItem builder reads.
type ValidationRow struct {
	AccessionNumber string `gorm:"column:accession_number"`
	AnalysisID      string `gorm:"column:analysis_id"`
	TestID          string `gorm:"column:test_id"`
	// The analysis's own section, used to filter by the caller's lab units.
	TestSectionID *string  `gorm:"column:test_section_id"`
	ResultID      *string  `gorm:"column:result_id"`
	ResultValue   *string  `gorm:"column:result_value"`
	ResultType    *string  `gorm:"column:result_type"`
	MinNormal     *float64 `gorm:"column:min_normal"`
	MaxNormal     *float64 `gorm:"column:max_normal"`
	SigDigits     *int     `gorm:"column:sig_digits"`

	TestName       string  `gorm:"column:test_name"`
	TestSortNumber *string `gorm:"column:test_sort_number"`
	// True when the analysis is technically REJECTED — the flag both
	// nonconforming and the screen's highlighting key off.
	StatusIsRejected bool    `gorm:"column:status_is_rejected"`
	UnitOfMeasure    *string `gorm:"column:unit_of_measure"`

	LimitLowNormal    *float64 `gorm:"column:limit_low_normal"`
	LimitHighNormal   *float64 `gorm:"column:limit_high_normal"`
	LimitLowCritical  *float64 `gorm:"column:limit_low_critical"`
	LimitHighCritical *float64 `gorm:"column:limit_high_critical"`

	// HasDefaultLimit is whether the default-band row joined; HasAnyLimit is
	// whether the test has ANY result_limits row. The pair distinguishes Java's
	// three outcomes — a resolved limit, a SYNTHESIZED one, and none at all —
	// which a single nullable column cannot.
	HasDefaultLimit bool `gorm:"column:has_default_limit"`
	HasAnyLimit     bool `gorm:"column:has_any_limit"`
	LimitSigDigits  *int `gorm:"column:limit_sig_digits"`

	PatientFirstName *string `gorm:"column:patient_first_name"`
	PatientLastName  *string `gorm:"column:patient_last_name"`
	NationalID       *string `gorm:"column:national_id"`
	Gender           *string `gorm:"column:gender"`
	EnteredBirthDate *string `gorm:"column:entered_birth_date"`
}

const validationSelect = `s.accession_number AS accession_number,
	a.id::text AS analysis_id,
	a.test_id::text AS test_id,
	a.test_sect_id::text AS test_section_id,
	(SELECT st2.name = 'Technical Rejected' FROM clinlims.status_of_sample st2
	  WHERE st2.id = a.status_id AND st2.status_type = 'ANALYSIS') AS status_is_rejected,
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
	rl.low_normal    AS limit_low_normal,
	rl.high_normal   AS limit_high_normal,
	rl.low_critical  AS limit_low_critical,
	rl.high_critical AS limit_high_critical,
	(rl.id IS NOT NULL) AS has_default_limit,
	EXISTS (SELECT 1 FROM clinlims.result_limits rl3 WHERE rl3.test_id = t.id) AS has_any_limit,
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
		// ONE limit, and NOT the patient's age band — the DEFAULT limit.
		//
		// createTestResultItem calls getResultLimitForTestAndPatient(test,
		// currentPatient) with a patient that carries neither a birth date nor a
		// gender on this path, so the service takes its `patient == null` branch
		// and returns defaultResultLimit — the row whose gender is BLANK and whose
		// age limits are the defaults 0..Infinity. A test with no such row gets NO
		// limit at all, and its normalRange comes out empty.
		//
		// Measured, and it is an asymmetry rather than a guess: LogbookResults
		// resolves the SAME analysis to result_limits row 11 and renders
		// "4.00 - 10.00", while AccessionValidation renders "" — one analysis, two
		// screens, two answers. Using the age-band rule here (which is correct for
		// the logbook) invents a reference range the validation screen does not
		// show.
		Joins(`LEFT JOIN clinlims.result_limits AS rl ON rl.id = (
			SELECT rl2.id FROM clinlims.result_limits rl2
			 WHERE rl2.test_id = t.id
			   AND (rl2.gender IS NULL OR rl2.gender = '')
			   AND rl2.min_age = 0
			   AND rl2.max_age = 'Infinity'::double precision
			 ORDER BY rl2.id LIMIT 1)`).
		Joins("LEFT JOIN clinlims.sample_human AS sh ON sh.samp_id = s.id").
		Joins("LEFT JOIN clinlims.patient AS pa ON pa.id = sh.patient_id").
		Joins("LEFT JOIN clinlims.person AS pe ON pe.id = pa.person_id").
		Joins(`JOIN clinlims.status_of_sample AS st ON st.id = a.status_id
			AND st.status_type = 'ANALYSIS' AND st.name IN ?`, d.ValidationStatusNames())
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
		Order("s.accession_number, t.sort_order::int, a.id").
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
		// sortByAccessionNumberAndOrder — accession, then the TEST sort order.
		Order("s.accession_number, t.sort_order::int, a.id").
		Scan(&rows).Error
	return rows, err
}

// ByStartedDate ports the `date` branch: getAnalysisStartedOn(date) filtered
// to the validation statuses.
//
// No analysis in the dev dataset carried a started_date, so this branch could
// only ever answer with an empty list — which is indistinguishable from not
// having implemented it at all. That is why it was left out and why leaving it
// out went unnoticed.
//
// The parameter arrives as dd/MM/yyyy, the format the form echoes back.
func (d *ResultValidationDAOImpl) ByStartedDate(date string) ([]ValidationRow, error) {
	rows := []ValidationRow{}
	err := d.base().
		Where("to_char(a.started_date, 'DD/MM/YYYY') = ?", date).
		Order("s.accession_number, t.sort_order::int, a.id").
		Scan(&rows).Error
	return rows, err
}

// BySection ports the testSectionId branch.
func (d *ResultValidationDAOImpl) BySection(sectionID string) ([]ValidationRow, error) {
	rows := []ValidationRow{}
	err := d.base().
		Where("a.test_sect_id = ?", sectionID).
		Order("s.accession_number, t.sort_order::int, a.id").
		Scan(&rows).Error
	return rows, err
}

// UserLabUnits ports the lab-unit half of getUserTestSections: the lab units
// on which this user holds the given role.
//
//	user_lab_unit_roles  (system_user_id)              the parent row
//	lab_unit_role_map    (id, lab_unit)                one row per lab unit
//	lab_roles            (lab_unit_role_map_id, role)  the roles ON that unit
//	lab_unit_roles       (system_user_id, map_id)      the join
//
// lab_unit holds a test_section id as text, except for the literal
// 'AllLabUnits' sentinel which means "every section".
func (d *ResultValidationDAOImpl) UserLabUnits(systemUserID, roleID string) ([]string, error) {
	units := []string{}
	err := d.DB.Table("clinlims.lab_unit_roles AS lur").
		Select("m.lab_unit").
		Joins("JOIN clinlims.lab_unit_role_map AS m ON m.lab_unit_role_map_id = lur.lab_unit_role_map_id").
		Joins("JOIN clinlims.lab_roles AS lr ON lr.lab_unit_role_map_id = m.lab_unit_role_map_id").
		Where("lur.system_user_id = ? AND lr.role = ?", systemUserID, roleID).
		Scan(&units).Error
	return units, err
}

// RoleIDByName resolves a system_role id by name. The ids are deployment data.
func (d *ResultValidationDAOImpl) RoleIDByName(name string) (string, error) {
	ids := []string{}
	err := d.DB.Table("clinlims.system_role").
		Select("id::text").
		Where("trim(name) = ?", name).
		Limit(1).
		Scan(&ids).Error
	if err != nil || len(ids) == 0 {
		return "", err
	}
	return ids[0], nil
}

// UserTestSections ports userService.getUserTestSections(user, ROLE_VALIDATION):
// the ACTIVE sections, LOCALIZED, narrowed to the caller's own lab units —
// unless the caller holds AllLabUnits, which returns every one.
//
// Returning them all unconditionally is not a cosmetic shortcut: paired with
// the unfiltered result list below it lets a user authorised for one lab unit
// read validation-pending results from every other one. The only user with lab
// units in the dev data holds AllLabUnits, which is why it went unnoticed.
func (d *ResultValidationDAOImpl) UserTestSections(units []string) ([]util.IdValuePair, error) {
	rows := []util.IdValuePair{}
	q := d.DB.Table("clinlims.test_section AS ts").
		Select("ts.id::text AS id, COALESCE(NULLIF(lv.value, ''), ts.name) AS value").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = ts.name_localization_id AND lv.locale = ?", d.Locale()).
		Where("ts.is_active = ?", "Y")
	if !HasAllLabUnits(units) {
		q = q.Where("ts.id::text IN ?", units)
	}
	err := q.Order("ts.sort_order").Scan(&rows).Error
	return rows, err
}

// AllLabUnits is the sentinel value UnifiedSystemUserController writes for a
// user who is not restricted to particular units.
const AllLabUnits = "AllLabUnits"

// HasAllLabUnits reports whether the sentinel is present.
func HasAllLabUnits(units []string) bool {
	for _, u := range units {
		if u == AllLabUnits {
			return true
		}
	}
	return false
}
