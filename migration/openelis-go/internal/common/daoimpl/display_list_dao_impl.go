// Package daoimpl ports the reference-list reads behind
// org.openelisglobal.common.services.DisplayListService. Folder layout mirrors
// the Java source during migration.
package daoimpl

import (
	"time"

	"gorm.io/gorm"

	"openelis-go/internal/common/util"
)

// DisplayListDAOImpl backs the DisplayListService list types the Wave 4 form
// loads consume.
//
// LOCALIZED NAMES: Java's BaseObject.getLocalizedName() reads the entity's
// localization row for the active locale and falls back to the entity's own
// name column. The localized text lives in clinlims.localization_value keyed by
// (localization_id, locale) — clinlims.localization itself holds no text, only
// a description. Every list below therefore LEFT JOINs localization_value on
// locale 'en' and COALESCEs to the plain column, which is what the fallback
// does. Joining `localization` alone returns nothing useful.
type DisplayListDAOImpl struct {
	DB *gorm.DB
}

// idValueRow is the scan target for the {id,value} list queries.
type idValueRow struct {
	ID    string `gorm:"column:id"`
	Value string `gorm:"column:value"`
}

func toPairs(rows []idValueRow) []util.IdValuePair {
	out := make([]util.IdValuePair, 0, len(rows))
	for _, r := range rows {
		out = append(out, util.NewIdValuePair(r.ID, r.Value))
	}
	return out
}

// DictionaryByCategoryLocalizedSort ports
// createFromDictionaryCategoryLocalizedSort: every dictionary entry in a
// category, re-sorted by the LOCALIZED name.
//
// The sort is Java's BaseObject.sortByLocalizedName, applied in memory AFTER
// the query, so it orders by displayed text — not by id and not by sort_order.
// That is why initialSampleConditionList starts at "Broken Tubes" (id 844)
// rather than "Refrigerated" (id 840); ordering by id looks plausible and is
// wrong.
func (d *DisplayListDAOImpl) DictionaryByCategoryLocalizedSort(categoryName string) ([]util.IdValuePair, error) {
	rows := []idValueRow{}
	err := d.DB.Table("clinlims.dictionary AS d").
		Select("d.id AS id, COALESCE(NULLIF(lv.value, ''), d.dict_entry) AS value").
		Joins("JOIN clinlims.dictionary_category AS c ON c.id = d.dictionary_category_id").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = d.name_localization_id AND lv.locale = 'en'").
		Where("c.name = ?", categoryName).
		Order("value").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return toPairs(rows), nil
}

// DictionaryByCategory ports createDictionaryListForCategory — the same source
// WITHOUT the localized re-sort, so rows keep the DAO's own id order.
//
// REJECTION_REASONS uses this one and INITIAL_SAMPLE_CONDITION uses the sorted
// variant, which is why they are two methods: rejectReasonList comes back
// 1140, 1141, 1142 even though the second value begins with a space and would
// sort first by text. A port sharing one helper gets one of the two wrong.
func (d *DisplayListDAOImpl) DictionaryByCategory(categoryName string) ([]util.IdValuePair, error) {
	rows := []idValueRow{}
	err := d.DB.Table("clinlims.dictionary AS d").
		Select("d.id AS id, COALESCE(NULLIF(lv.value, ''), d.dict_entry) AS value").
		Joins("JOIN clinlims.dictionary_category AS c ON c.id = d.dictionary_category_id").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = d.name_localization_id AND lv.locale = 'en'").
		Where("c.name = ?", categoryName).
		Order("d.id::numeric").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return toPairs(rows), nil
}

// ActiveTestSections ports createTestSectionActiveList.
func (d *DisplayListDAOImpl) ActiveTestSections() ([]util.IdValuePair, error) {
	rows := []idValueRow{}
	err := d.DB.Table("clinlims.test_section AS ts").
		Select("ts.id AS id, COALESCE(NULLIF(lv.value, ''), ts.name) AS value").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = ts.name_localization_id AND lv.locale = 'en'").
		Where("ts.is_active = 'Y'").
		Order("ts.sort_order").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return toPairs(rows), nil
}

// ActiveHumanSampleTypes ports createSampleTypeList(false):
// getTypesForDomainBySortOrder(SampleDomain.HUMAN), filtered to active.
//
// The domain filter is not decoration — type_of_sample carries non-human
// domains too, and dropping it changes the list.
func (d *DisplayListDAOImpl) ActiveHumanSampleTypes() ([]util.IdValuePair, error) {
	rows := []idValueRow{}
	err := d.DB.Table("clinlims.type_of_sample AS t").
		Select("t.id AS id, COALESCE(NULLIF(lv.value, ''), t.description) AS value").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = t.name_localization_id AND lv.locale = 'en'").
		Where("t.is_active = 'Y' AND t.domain = 'H'").
		Order("t.sort_order").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return toPairs(rows), nil
}

// ReferralReasonRow carries the display_key so the service can resolve the
// label through the message bundle, as Java's getLocalizedName does.
//
// referral_reason has NO name_localization_id column — it localizes through
// display_key alone, unlike dictionary / test_section / type_of_sample. The
// stored `name` is the fallback, and it is TRIMMED: row 3 is stored as
// "Further testing required " with a trailing space and comes back without it.
type ReferralReasonRow struct {
	ID         string  `gorm:"column:id"`
	Name       string  `gorm:"column:name"`
	DisplayKey *string `gorm:"column:display_key"`
}

// ReferralReasons ports createReferralReasonList — the referral_reason table,
// not a dictionary category.
func (d *DisplayListDAOImpl) ReferralReasons() ([]ReferralReasonRow, error) {
	rows := []ReferralReasonRow{}
	err := d.DB.Table("clinlims.referral_reason").
		Select("id, name, display_key").
		Order("id::numeric").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// ReferralOrganizations ports createReferralOrganizationList: organizations
// carrying the `referralLab` organization type, sorted by NAME in Java rather
// than by id.
func (d *DisplayListDAOImpl) ReferralOrganizations() ([]util.IdValuePair, error) {
	rows := []idValueRow{}
	err := d.DB.Table("clinlims.organization AS o").
		Select("o.id AS id, o.name AS value").
		Joins("JOIN clinlims.organization_organization_type AS oot ON oot.org_id = o.id").
		Joins("JOIN clinlims.organization_type AS t ON t.id = oot.org_type_id").
		Where("t.short_name = ? OR t.description = ?", "referralLab", "referralLab").
		Order("o.name").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return toPairs(rows), nil
}

// ActivePractitionerPersons ports createActivePractitionerPersonsList.
//
// Two details that are easy to get wrong:
//   - the id is the PERSON id, not the provider id. provider and person are
//     separate tables and the form binds to the person.
//   - only ACTIVE providers are listed (provider.active, a real boolean column
//     — not the 'Y'/'N' char other tables use).
//
// Java sorts by last name with NULLs pushed to the END; `NULLS LAST` reproduces
// that, where a plain ORDER BY would put them first in Postgres' default for
// ascending order.
func (d *DisplayListDAOImpl) ActivePractitionerPersons() ([]util.IdValuePair, error) {
	rows := []idValueRow{}
	err := d.DB.Table("clinlims.provider AS pr").
		Select("pe.id AS id, COALESCE(pe.last_name, '') || ', ' || COALESCE(pe.first_name, '') AS value").
		Joins("JOIN clinlims.person AS pe ON pe.id = pr.person_id").
		Where("pr.active = true").
		Order("pe.last_name NULLS LAST").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return toPairs(rows), nil
}

// ReferringClinics ports createReferringClinicList: organizations of type
// `referring clinic`, sorted by name.
//
// The label is CONDITIONAL — "shortName - organizationName" when a short name
// exists, the bare name otherwise. Java's guard is
// GenericValidator.isBlankOrNull, so an EMPTY short name takes the bare-name
// branch just as a NULL one does; the referralLab org in the dev dataset stores
// "" and would get a stray " - " prefix from a null-only check.
func (d *DisplayListDAOImpl) ReferringClinics() ([]util.IdValuePair, error) {
	rows := []idValueRow{}
	err := d.DB.Table("clinlims.organization AS o").
		Select(`o.id AS id,
			CASE WHEN COALESCE(btrim(o.short_name), '') = '' THEN o.name
			     ELSE o.short_name || ' - ' || o.name END AS value`).
		Joins("JOIN clinlims.organization_organization_type AS oot ON oot.org_id = o.id").
		Joins("JOIN clinlims.organization_type AS t ON t.id = oot.org_type_id").
		Where("t.short_name = ?", "referring clinic").
		Order("o.name").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return toPairs(rows), nil
}

// ProjectRow is one clinlims.project row as the form loads emit it: the FULL
// entity, not an {id,value} pair.
//
// The column is `name`, exposed as `projectName`. ProgramCode is a pointer
// because the column is nullable and Include.NON_NULL drops the key on rows
// without one — a list where every row carries `programCode: null` is a
// different response.
type ProjectRow struct {
	ID          string     `gorm:"column:id"`
	ProjectName string     `gorm:"column:name"`
	Description *string    `gorm:"column:description"`
	IsActive    string     `gorm:"column:is_active"`
	ProgramCode *string    `gorm:"column:program_code"`
	Lastupdated *time.Time `gorm:"column:lastupdated"`
}

// Projects returns every project row, in id order.
func (d *DisplayListDAOImpl) Projects() ([]ProjectRow, error) {
	rows := []ProjectRow{}
	err := d.DB.Table("clinlims.project").
		Select("id, name, description, is_active, program_code, lastupdated").
		Order("id::numeric").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// PatientTypeRow is one clinlims.patient_type row.
//
// There is NO isActive column: the JSON's `isActive: "Y"` comes from
// EnumValueItemImpl, whose field is initialised to the literal "Y" and is never
// overwritten on this path. It is a constant in the response, not data — the
// service supplies it.
type PatientTypeRow struct {
	ID          string     `gorm:"column:id"`
	Type        string     `gorm:"column:type"`
	Description string     `gorm:"column:description"`
	Lastupdated *time.Time `gorm:"column:lastupdated"`
}

// PatientTypes returns every patient_type row.
func (d *DisplayListDAOImpl) PatientTypes() ([]PatientTypeRow, error) {
	rows := []PatientTypeRow{}
	err := d.DB.Table("clinlims.patient_type").
		Select("id, type, description, lastupdated").
		Order("id::numeric").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
