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
	// ActiveLocale is the locale getLocalizedName resolves against, taken from
	// site_information."default language locale" (language subtag only). Empty
	// falls back to "en" — see Locale().
	//
	// It is a field rather than the literal 'en' this DAO used to repeat in
	// every query: on a deployment whose default locale is not English, every
	// form list came back in English while Java returned the active locale.
	ActiveLocale string
}

// Locale returns the configured locale, or "en" when unset.
func (d *DisplayListDAOImpl) Locale() string {
	if d.ActiveLocale != "" {
		return d.ActiveLocale
	}
	return "en"
}

// activeFlagRow is idValueRow plus the is_active column, for the lists whose
// active filter Java applies in Java rather than in SQL.
//
// IsActive is a BOOL because type_of_sample.is_active is a real boolean column,
// unlike dictionary / test / test_section, which store the char 'Y'. Postgres
// accepts 'Y' as a boolean literal, so `is_active = 'Y'` works in SQL against
// both and hides the difference — but scanning the boolean into a Go string
// yields "true", never "Y", and the filter silently drops every row.
type activeFlagRow struct {
	ID       string `gorm:"column:id"`
	Value    string `gorm:"column:value"`
	IsActive bool   `gorm:"column:is_active"`
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
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = d.name_localization_id AND lv.locale = ?", d.Locale()).
		Where("c.name = ? AND d.is_active = ?", categoryName, "Y").
		// COLLATE "C" is byte order, which is what Java's String.compareTo does.
		// Postgres' default collation ignores case, so it puts "divorced"
		// before "DNA" where Java puts "DNA" first.
		Order(`COALESCE(NULLIF(lv.value, ''), d.dict_entry) COLLATE "C"`).
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
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = d.name_localization_id AND lv.locale = ?", d.Locale()).
		Where("c.name = ? AND d.is_active = ?", categoryName, "Y").
		// sort_order, NOT id. They usually agree — sort_order is typically
		// id*100 — but not always: in resultRejectionReasons, 1160 and 1161
		// carry 115233 and 115266, which places them BETWEEN 1152 and 1153.
		// Ordering by id puts them at the end and is wrong.
		Order("d.sort_order").
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
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = ts.name_localization_id AND lv.locale = ?", d.Locale()).
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
	// THE is_active FILTER IS DELIBERATELY NOT IN THE SQL.
	//
	// Java's HQL is `from TypeOfSample t where t.domain = :domainKey order by
	// t.sortOrder` — no active predicate — and createSampleTypeList drops the
	// inactive rows afterwards, in Java. That matters because sort_order has
	// heavy ties here (three rows at 0, seven at Integer.MAX_VALUE) and the tie
	// order is whatever the query plan produces. Adding the predicate to the
	// SQL changes the plan and therefore the tie order: measured, filtering in
	// SQL yields 30,32,31,… while Java yields 31,32,30,…
	//
	// So the filter runs after the scan, exactly where Java runs it.
	rows := []activeFlagRow{}
	err := d.DB.Table("clinlims.type_of_sample AS t").
		Select("t.id AS id, COALESCE(NULLIF(lv.value, ''), t.description) AS value, t.is_active AS is_active").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = t.name_localization_id AND lv.locale = ?", d.Locale()).
		Where("t.domain = 'H'").
		Order("t.sort_order").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]util.IdValuePair, 0, len(rows))
	for _, r := range rows {
		if r.IsActive {
			out = append(out, util.NewIdValuePair(r.ID, r.Value))
		}
	}
	return out, nil
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

// UserSampleTypes ports UserServiceImpl.getUserSampleTypes(userId, roleName).
//
// NOT "the active sample types" — that is a different list
// (ActiveHumanSampleTypes) which SampleBatchEntrySetup uses. This one walks:
//
//	role name -> role id
//	 -> the user's lab units whose lab_roles contain that role id
//	 -> the ACTIVE test sections those lab units cover
//	     ("AllLabUnits" is a wildcard meaning every active section)
//	 -> the ACTIVE tests in those sections
//	 -> the distinct sample types those tests accept
//
// On the dev dataset that is 12 of the 14 active types; two carry no active
// test. A port that served ActiveHumanSampleTypes here would return 14 and be
// wrong on exactly the endpoint pair the c2 spec contrasts.
//
// The storage layout: user_lab_unit_roles -> lab_unit_roles ->
// lab_unit_role_map (the lab unit) -> lab_roles (its role ids).
//
// ORDER IS NOT REPRODUCED. Java collects the ids into a HashSet and iterates
// it, so the wire order is Java's string-hash bucket order — an implementation
// detail of the JDK, not a contract. The c2 spec compares this list as a SET
// for that reason; matching the bucket order would be reproducing an accident.
func (d *DisplayListDAOImpl) UserSampleTypes(systemUserID, roleName string) ([]util.IdValuePair, error) {
	labUnits := []string{}
	err := d.DB.Table("clinlims.lab_unit_roles AS lur").
		Select("DISTINCT lurm.lab_unit").
		Joins("JOIN clinlims.lab_unit_role_map AS lurm ON lurm.lab_unit_role_map_id = lur.lab_unit_role_map_id").
		Joins("JOIN clinlims.lab_roles AS lr ON lr.lab_unit_role_map_id = lurm.lab_unit_role_map_id").
		Joins("JOIN clinlims.system_role AS sr ON sr.id::text = lr.role").
		Where("lur.system_user_id::text = ? AND sr.name = ?", systemUserID, roleName).
		Pluck("lab_unit", &labUnits).Error
	if err != nil {
		return nil, err
	}

	sections := d.DB.Table("clinlims.test_section").Select("id").Where("is_active = 'Y'")
	allLabUnits := false
	for _, u := range labUnits {
		if u == "AllLabUnits" {
			allLabUnits = true
			break
		}
	}
	if !allLabUnits {
		// The non-wildcard branch filters the ACTIVE sections by id — Java
		// intersects its lab-unit list against TEST_SECTION_ACTIVE rather than
		// querying the sections directly, so an inactive section named by a lab
		// unit is still excluded.
		sections = sections.Where("id::text IN ?", labUnits)
	}

	// NO DISTINCT and NO ORDER BY, deliberately.
	//
	// Java walks the tests in the order its own unordered query returns them and,
	// per test, adds that test's sample types to a HashSet. The FIRST-SEEN order
	// of that walk decides the order of entries sharing a hash bucket, so it is
	// observable in the response even though nothing sorts it. Deduplicating in
	// SQL with DISTINCT would discard it.
	//
	// Verified: this walk yields 4,2,3,34,36,37,1,30,31,32,26,35, and running
	// that through JavaHashSetOrder reproduces Java's
	// 34,1,2,35,3,36,4,37,26,30,31,32 exactly — all four multi-entry buckets in
	// the right internal order.
	rows := []idValueRow{}
	err = d.DB.Table("clinlims.test AS t").
		Select("tos.id AS id, COALESCE(NULLIF(lv.value, ''), tos.description) AS value").
		Joins("JOIN clinlims.sampletype_test AS st ON st.test_id = t.id").
		Joins("JOIN clinlims.type_of_sample AS tos ON tos.id = st.sample_type_id").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = tos.name_localization_id AND lv.locale = ?", d.Locale()).
		Where("t.is_active = 'Y'").
		Where("t.test_section_id IN (?)", sections).
		// ctid is the PHYSICAL row order, which is what Java's unordered query
		// returns and therefore what decides first-seen order. Stated explicitly
		// so the result does not depend on the plan Postgres happens to pick:
		// measured, this reproduces Java's walk exactly.
		Order("t.ctid, st.ctid").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	// Dedup preserving first appearance — the HashSet.addAll semantics.
	seen := map[string]bool{}
	unique := make([]idValueRow, 0, len(rows))
	for _, r := range rows {
		if seen[r.ID] {
			continue
		}
		seen[r.ID] = true
		unique = append(unique, r)
	}
	return toPairs(unique), nil
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
	// Ordered by the TYPE CODE, not by id: Java returns E, H, P, R, U, which is
	// ids 2, 3, 5, 1, 4. Ordering by id looks natural and is wrong.
	err := d.DB.Table("clinlims.patient_type").
		Select("id, type, description, lastupdated").
		Order("type").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// GenderRow is one clinlims.gender row.
//
// GenderType ("M"/"F") is what the {id,value} list uses as its id — NOT the
// numeric primary key.
type GenderRow struct {
	GenderType  string  `gorm:"column:gender_type"`
	Description string  `gorm:"column:description"`
	NameKey     *string `gorm:"column:name_key"`
}

// Genders ports createGenderList's source.
func (d *DisplayListDAOImpl) Genders() ([]GenderRow, error) {
	rows := []GenderRow{}
	err := d.DB.Table("clinlims.gender").
		Select("gender_type, COALESCE(description, '') AS description, name_key").
		Order("id::numeric").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// OrganizationsByTypeName returns organizations carrying an organization type,
// sorted by name — the shape createPatientHealthRegions and its siblings use.
//
// Returns an empty slice rather than nil so the caller serializes [] and never
// null; both health lists are empty on the dev dataset and Java still emits [].
func (d *DisplayListDAOImpl) OrganizationsByTypeName(typeName string) ([]util.IdValuePair, error) {
	rows := []idValueRow{}
	err := d.DB.Table("clinlims.organization AS o").
		Select("o.id AS id, o.name AS value").
		Joins("JOIN clinlims.organization_organization_type AS oot ON oot.org_id = o.id").
		Joins("JOIN clinlims.organization_type AS t ON t.id = oot.org_type_id").
		Where("t.short_name = ? OR t.description = ?", typeName, typeName).
		Order("o.name").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return toPairs(rows), nil
}

// SiteInformation reads one clinlims.site_information value, or "" when the row
// is absent. Java reaches these through ConfigurationProperties, whose Property
// enum maps a code name to the row name — and the mapping is not always the
// obvious one: Property.AccessionFormat is stored under `acessionFormat`, with
// the typo, so looking up "accessionFormat" finds nothing.
func (d *DisplayListDAOImpl) SiteInformation(name string) (string, error) {
	var value string
	err := d.DB.Table("clinlims.site_information").
		Select("value").
		Where("name = ?", name).
		Limit(1).
		Scan(&value).Error
	if err != nil {
		return "", err
	}
	return value, nil
}
