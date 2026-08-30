package daoimpl

import (
	"gorm.io/gorm"
)

// TestModifyReadDAO backs TestModifyEntry's GET — the test catalogue the screen
// lists, and everything each row hangs off.
//
// The catalogue is only built when a FILTER is given. A blank GET returns an
// empty testCatBeanList rather than every test, which is a deliberate guard on
// the initial page load and not an accident of the query.
type TestModifyReadDAO struct {
	DB           *gorm.DB
	ActiveLocale string
}

func (d *TestModifyReadDAO) locale() string {
	if d.ActiveLocale == "" {
		return "en"
	}
	return d.ActiveLocale
}

// CatalogTestRow is one `test`, with the columns the bean reads directly.
type CatalogTestRow struct {
	ID               string  `gorm:"column:id"`
	Loinc            *string `gorm:"column:loinc"`
	IsActive         bool    `gorm:"column:is_active"`
	Orderable        bool    `gorm:"column:orderable"`
	NotifyResults    bool    `gorm:"column:notify_results"`
	InLabOnly        bool    `gorm:"column:in_lab_only"`
	AMR              bool    `gorm:"column:amr"`
	SortOrder        *int    `gorm:"column:sort_order"`
	UomName          *string `gorm:"column:uom_name"`
	TestSectionName  *string `gorm:"column:test_section_name"`
	NameLocID        *string `gorm:"column:name_loc_id"`
	ReportingLocID   *string `gorm:"column:reporting_loc_id"`
	NameLocDesc      *string `gorm:"column:name_loc_desc"`
	NameLocUpdated   *int64  `gorm:"column:name_loc_updated"`
	ReportLocDesc    *string `gorm:"column:report_loc_desc"`
	ReportLocUpdated *int64  `gorm:"column:report_loc_updated"`
}

// CatalogTests ports getAllTests(false) plus the two filters.
//
// The base query is `from Test Order by description` — every test, ordered by
// description — and the filters run over that list in Java. sampleType wins:
// `if (sampleType) … else if (testSection)`, so submitting both applies only
// the first.
//
// test.is_active is a VARCHAR 'Y'/'N' here, not the boolean type_of_sample
// carries; `isActive()` compares to "Y".
func (d *TestModifyReadDAO) CatalogTests(sampleTypeID, testSectionID string) ([]CatalogTestRow, error) {
	rows := []CatalogTestRow{}
	q := d.DB.Table("clinlims.test AS t").
		Select(`t.id::text AS id,
		        t.loinc AS loinc,
		        (t.is_active = 'Y') AS is_active,
		        COALESCE(t.orderable, false) AS orderable,
		        COALESCE(t.notify_results, false) AS notify_results,
		        COALESCE(t.in_lab_only, false) AS in_lab_only,
		        COALESCE(t.antimicrobial_resistance, false) AS amr,
		        t.sort_order::int AS sort_order,
		        uom.name AS uom_name,
		        COALESCE(NULLIF(tslv.value, ''), ts.name) AS test_section_name,
		        t.name_localization_id::text AS name_loc_id,
		        t.reporting_name_localization_id::text AS reporting_loc_id,
		        nl.description AS name_loc_desc,
		        trunc(EXTRACT(EPOCH FROM nl.lastupdated) * 1000)::bigint AS name_loc_updated,
		        rl.description AS report_loc_desc,
		        trunc(EXTRACT(EPOCH FROM rl.lastupdated) * 1000)::bigint AS report_loc_updated`).
		Joins("LEFT JOIN clinlims.unit_of_measure AS uom ON uom.id = t.uom_id").
		Joins("LEFT JOIN clinlims.test_section AS ts ON ts.id = t.test_section_id").
		Joins(`LEFT JOIN clinlims.localization_value AS tslv
		         ON tslv.localization_id = ts.name_localization_id AND tslv.locale = ?`, d.locale()).
		Joins("LEFT JOIN clinlims.localization AS nl ON nl.id = t.name_localization_id").
		Joins("LEFT JOIN clinlims.localization AS rl ON rl.id = t.reporting_name_localization_id").
		Order("t.description")

	switch {
	case sampleTypeID != "":
		q = q.Where(`EXISTS (SELECT 1 FROM clinlims.sampletype_test st
		                      WHERE st.test_id = t.id AND st.sample_type_id = ?)`, sampleTypeID)
	case testSectionID != "":
		q = q.Where("t.test_section_id = ?", testSectionID)
	}
	err := q.Scan(&rows).Error
	return rows, err
}

// CatalogStringRow is one (test id, value) pair for the per-test sub-lists.
type CatalogStringRow struct {
	TestID string `gorm:"column:test_id"`
	Value  string `gorm:"column:value"`
}

// PanelNamesByTest ports createPanelList: the panels a test belongs to, by
// their LOCALIZATION value — getLocalizedValueById, so the entity's own name is
// not a fallback here the way it is elsewhere.
func (d *TestModifyReadDAO) PanelNamesByTest() ([]CatalogStringRow, error) {
	rows := []CatalogStringRow{}
	err := d.DB.Table("clinlims.panel_item AS pi").
		Select(`pi.test_id::text AS test_id,
		        COALESCE(lv.value, '') AS value`).
		Joins("JOIN clinlims.panel AS p ON p.id = pi.panel_id").
		Joins(`LEFT JOIN clinlims.localization_value AS lv
		         ON lv.localization_id = p.name_localization_id AND lv.locale = ?`, d.locale()).
		Order("pi.test_id, pi.id").
		Scan(&rows).Error
	return rows, err
}

// SampleTypeNameByTest ports testService.getTypeOfSample: the FIRST
// sampletype_test row's sample type, localized. A test with none renders "n/a".
func (d *TestModifyReadDAO) SampleTypeNameByTest() ([]CatalogStringRow, error) {
	rows := []CatalogStringRow{}
	err := d.DB.Table("clinlims.sampletype_test AS st").
		Select(`st.test_id::text AS test_id,
		        COALESCE(NULLIF(lv.value, ''), tos.description, '') AS value`).
		Joins("JOIN clinlims.type_of_sample AS tos ON tos.id = st.sample_type_id").
		Joins(`LEFT JOIN clinlims.localization_value AS lv
		         ON lv.localization_id = tos.name_localization_id AND lv.locale = ?`, d.locale()).
		Order("st.test_id, st.id").
		Scan(&rows).Error
	return rows, err
}

// CatalogResultRow is one ACTIVE test_result, with its dictionary entry
// resolved.
type CatalogResultRow struct {
	TestID            string  `gorm:"column:test_id"`
	ResultType        string  `gorm:"column:result_type"`
	Value             string  `gorm:"column:value"`
	IsQuantifiable    bool    `gorm:"column:is_quantifiable"`
	SignificantDigits *string `gorm:"column:significant_digits"`
	DictionaryName    *string `gorm:"column:dictionary_name"`
}

// ActiveResultsByTest ports getAllActiveTestResultsPerTest.
//
// The order is `resultGroup, id` DESCENDING — the DAO's `descending` flag
// applies to every order property — which is why a dictionary test lists its
// options newest first and why getResultType reads the LAST row inserted.
func (d *TestModifyReadDAO) ActiveResultsByTest() ([]CatalogResultRow, error) {
	rows := []CatalogResultRow{}
	err := d.DB.Table("clinlims.test_result AS tr").
		Select(`tr.test_id::text AS test_id,
		        COALESCE(tr.tst_rslt_type, '') AS result_type,
		        COALESCE(tr.value, '') AS value,
		        COALESCE(tr.is_quantifiable, false) AS is_quantifiable,
		        tr.significant_digits::text AS significant_digits,
		        COALESCE(NULLIF(dlv.value, ''), dict.dict_entry) AS dictionary_name`).
		Joins("LEFT JOIN clinlims.dictionary AS dict ON dict.id::text = tr.value").
		Joins(`LEFT JOIN clinlims.localization_value AS dlv
		         ON dlv.localization_id = dict.name_localization_id AND dlv.locale = ?`, d.locale()).
		Where("tr.is_active = true").
		Order("tr.test_id, tr.result_group DESC NULLS FIRST, tr.id DESC").
		Scan(&rows).Error
	return rows, err
}

// CatalogLimitRow is one result_limits row, with its dictionary normal
// resolved.
type CatalogLimitRow struct {
	TestID             string  `gorm:"column:test_id"`
	ResultTypeID       string  `gorm:"column:result_type_id"`
	Gender             string  `gorm:"column:gender"`
	MinAge             float64 `gorm:"column:min_age"`
	MaxAge             float64 `gorm:"column:max_age"`
	LowNormal          float64 `gorm:"column:low_normal"`
	HighNormal         float64 `gorm:"column:high_normal"`
	LowValid           float64 `gorm:"column:low_valid"`
	HighValid          float64 `gorm:"column:high_valid"`
	LowReportingRange  float64 `gorm:"column:low_reporting_range"`
	HighReportingRange float64 `gorm:"column:high_reporting_range"`
	LowCritical        float64 `gorm:"column:low_critical"`
	HighCritical       float64 `gorm:"column:high_critical"`
	DictionaryNormalID string  `gorm:"column:dictionary_normal_id"`
	DictionaryName     *string `gorm:"column:dictionary_name"`
}

// LimitsByTest ports resultLimitService.getResultLimits(test).
func (d *TestModifyReadDAO) LimitsByTest() ([]CatalogLimitRow, error) {
	rows := []CatalogLimitRow{}
	err := d.DB.Table("clinlims.result_limits AS rl").
		Select(`rl.test_id::text AS test_id,
		        rl.test_result_type_id::text AS result_type_id,
		        COALESCE(rl.gender, '') AS gender,
		        COALESCE(rl.min_age, 0) AS min_age,
		        COALESCE(rl.max_age, 'Infinity'::float8) AS max_age,
		        COALESCE(rl.low_normal, '-Infinity'::float8) AS low_normal,
		        COALESCE(rl.high_normal, 'Infinity'::float8) AS high_normal,
		        COALESCE(rl.low_valid, '-Infinity'::float8) AS low_valid,
		        COALESCE(rl.high_valid, 'Infinity'::float8) AS high_valid,
		        COALESCE(rl.low_reporting_range, '-Infinity'::float8) AS low_reporting_range,
		        COALESCE(rl.high_reporting_range, 'Infinity'::float8) AS high_reporting_range,
		        COALESCE(rl.low_critical, 'Infinity'::float8) AS low_critical,
		        COALESCE(rl.high_critical, 'Infinity'::float8) AS high_critical,
		        COALESCE(rl.normal_dictionary_id::text, '') AS dictionary_normal_id,
		        COALESCE(NULLIF(dlv.value, ''), dict.dict_entry) AS dictionary_name`).
		Joins("LEFT JOIN clinlims.dictionary AS dict ON dict.id = rl.normal_dictionary_id").
		Joins(`LEFT JOIN clinlims.localization_value AS dlv
		         ON dlv.localization_id = dict.name_localization_id AND dlv.locale = ?`, d.locale()).
		Order("rl.test_id, rl.id").
		Scan(&rows).Error
	return rows, err
}
