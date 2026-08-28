// Package daoimpl ports the referral queries behind rest/ReferredOutTests
// (constitution.md Layer II). Folder layout mirrors the Java source.
package daoimpl

import (
	"openelis-go/internal/common/util"

	"gorm.io/gorm"
)

// ReferralDAOImpl backs rest/ReferredOutTests (Wave 5.4).
type ReferralDAOImpl struct {
	DB *gorm.DB

	// ActiveLocale is site_information "default language locale".
	ActiveLocale string
}

// Locale returns the configured locale, or "en" when unset.
func (d *ReferralDAOImpl) Locale() string {
	if d.ActiveLocale != "" {
		return d.ActiveLocale
	}
	return "en"
}

// ReferralDisplayRow carries everything convertToDisplayItem reads.
type ReferralDisplayRow struct {
	AccessionNumber  string  `gorm:"column:accession_number"`
	ReferredSendDate *string `gorm:"column:referred_send_date"`
	Status           *string `gorm:"column:status"`
	PatientLastName  *string `gorm:"column:patient_last_name"`
	PatientFirstName *string `gorm:"column:patient_first_name"`
	TestName         *string `gorm:"column:test_name"`
	OrganizationName *string `gorm:"column:organization_name"`
	AnalysisID       string  `gorm:"column:analysis_id"`
	ResultCount      int64   `gorm:"column:result_count"`
}

// ByAccessionNumber ports getReferralsByAccessionNumber, then the per-referral
// work convertToDisplayItem does, as one query.
//
// referringTestName is the BARE localized test name
// (test.getLocalizedTestName().getLocalizedValue()) — NOT the augmented form
// with the sample type that the same wave's WorkPlanByTestSection emits and
// that this endpoint's own testSelectionList uses. Three name builders in one
// wave; they are not interchangeable.
func (d *ReferralDAOImpl) ByAccessionNumber(accessionNumber string) ([]ReferralDisplayRow, error) {
	rows := []ReferralDisplayRow{}
	err := d.DB.Table("clinlims.referral AS r").
		Select(`s.accession_number AS accession_number,
			to_char(r.sent_date, 'DD/MM/YYYY') AS referred_send_date,
			r.status AS status,
			pe.last_name  AS patient_last_name,
			pe.first_name AS patient_first_name,
			COALESCE(lv.value, t.name) AS test_name,
			o.name AS organization_name,
			a.id::text AS analysis_id,
			(SELECT count(*) FROM clinlims.result res WHERE res.analysis_id = a.id) AS result_count`).
		Joins("JOIN clinlims.analysis AS a ON a.id = r.analysis_id").
		Joins("JOIN clinlims.sample_item AS si ON si.id = a.sampitem_id").
		Joins("JOIN clinlims.sample AS s ON s.id = si.samp_id").
		Joins("JOIN clinlims.test AS t ON t.id = a.test_id").
		Joins("LEFT JOIN clinlims.localization AS l ON l.id = t.name_localization_id").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = l.id AND lv.locale = ?", d.Locale()).
		Joins("LEFT JOIN clinlims.organization AS o ON o.id = r.organization_id").
		Joins("LEFT JOIN clinlims.sample_human AS sh ON sh.samp_id = s.id").
		Joins("LEFT JOIN clinlims.patient AS pa ON pa.id = sh.patient_id").
		Joins("LEFT JOIN clinlims.person AS pe ON pe.id = pa.person_id").
		Where("s.accession_number = ?", accessionNumber).
		Order("r.ctid").
		Scan(&rows).Error
	return rows, err
}

// AllActiveTests ports DisplayListService.createTestList — every ACTIVE test,
// valued by the AUGMENTED localized name (name plus "(sample type)"), sorted by
// that value with Java's String.compareTo.
//
// The sort is done in SQL with COLLATE "C", which is byte order and therefore
// the same comparison Java performs; the database's own collation ignores
// punctuation and would order this list differently.
func (d *ReferralDAOImpl) AllActiveTests() ([]util.IdValuePair, error) {
	rows := []util.IdValuePair{}
	augmented := `COALESCE(lv.value, t.name) || COALESCE(
		(SELECT '(' || COALESCE(tlv.value, tos.description) || ')'
		   FROM clinlims.sampletype_test AS tost
		   JOIN clinlims.type_of_sample AS tos ON tos.id = tost.sample_type_id
		   LEFT JOIN clinlims.localization AS tl ON tl.id = tos.name_localization_id
		   LEFT JOIN clinlims.localization_value AS tlv
		          ON tlv.localization_id = tl.id AND tlv.locale = '` + d.Locale() + `'
		  WHERE tost.test_id = t.id
		    AND tos.local_abbrev IS DISTINCT FROM 'Variable'
		  ORDER BY tost.ctid LIMIT 1), '')`
	err := d.DB.Table("clinlims.test AS t").
		Select("t.id::text AS id, "+augmented+" AS value").
		Joins("LEFT JOIN clinlims.localization AS l ON l.id = t.name_localization_id").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = l.id AND lv.locale = ?", d.Locale()).
		Where("t.is_active = ?", "Y").
		Order(`(` + augmented + `) COLLATE "C"`).
		Scan(&rows).Error
	return rows, err
}

// TestSectionsByName ports DisplayListService.createTestSectionByNameList.
//
// This is NOT the same list as ListType.TEST_SECTION, even though both are
// "the active test sections":
//
//	TEST_SECTION          COALESCE(localization value, name)  -- localized
//	TEST_SECTION_BY_NAME  section.getTestSectionName()        -- the RAW column
//
// The two disagree wherever a section's stored name is in a different language
// from its localization. On this deployment section 76 is stored as "Virologie"
// with an English localization of "Virology", so reusing the localized list
// here emits "Virology" where Java emits "Virologie" — one row out of ten, and
// invisible on the other nine.
func (d *ReferralDAOImpl) TestSectionsByName() ([]util.IdValuePair, error) {
	rows := []util.IdValuePair{}
	err := d.DB.Table("clinlims.test_section AS ts").
		Select("ts.id::text AS id, ts.name AS value").
		Where("ts.is_active = ?", "Y").
		Order("ts.sort_order").
		Scan(&rows).Error
	return rows, err
}
