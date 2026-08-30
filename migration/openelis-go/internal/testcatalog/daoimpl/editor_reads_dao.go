package daoimpl

import (
	"gorm.io/gorm"
)

// EditorReadDAO backs the ten test-catalog reads that pair with no write — the
// list page, the editor envelope, the localization refs, the LOINC integrity
// check, the dictionary typeahead, siblings, the group summary, the analyzer
// table, and the two read-only controllers (reflex/calc and storage history).
//
// They are grouped here because they share one thing: the AUGMENTED test name,
// `Albumin(Urines)` — the localized name with the first sample type in
// brackets. Every one of these endpoints renders it, and three of them render
// something subtly different, so the rule is written once in augmentedName().
type EditorReadDAO struct {
	DB           *gorm.DB
	ActiveLocale string
	// AugmentNames is ConfigurationProperties.TEST_NAME_AUGMENTED — the
	// site_information row `augmentTestNameWithType`. Java reads it from a cache
	// loaded at startup, so it is passed in rather than queried per request.
	// When it is false NO name carries its sample type, here or anywhere else.
	AugmentNames bool
}

func (d *EditorReadDAO) locale() string {
	if d.ActiveLocale == "" {
		return "en"
	}
	return d.ActiveLocale
}

// variableSampleTypeAbbrev is the local_abbrev TestServiceImpl resolves
// VARIABLE_TYPE_OF_SAMPLE_ID from. A test whose sample type is that one is NOT
// augmented — "select the type at order time" is not a specimen.
const variableSampleTypeAbbrev = "Variable"

// augmentedName is buildAugmentedTestNameForLocale as SQL.
//
// The name is the test's localization for the active locale, falling back to
// `description` — Test.getName()'s own rule, not the `name` column. The suffix
// is the FIRST sampletype_test link's sample type, localized the same way, and
// it is left off when the test has no link, when that link is the Variable
// type, or when the deployment has augmentation switched off.
//
// Several shipped tests already carry their specimen in their own name, so the
// result reads doubled — "WBC(Whole Blood)(Whole Blood)". That is the wire
// value; it is not a bug to tidy.
func (d *EditorReadDAO) augmentedName(alias string) string {
	raw := "COALESCE(NULLIF(nlv.value, ''), " + alias + ".description)"
	if !d.AugmentNames {
		return raw
	}
	return raw + ` || COALESCE('(' || NULLIF(stlv2.value, '') || ')',
	                           '(' || tos.description || ')', '')`
}

// nameJoins is the join set augmentedName() depends on: the test's own
// localization, its first sample type, and that type's localization.
//
// The sample type is picked with a LATERAL over `ORDER BY st.id LIMIT 1`
// because Java takes `getTypeOfSampleTestsForTest(...).get(0)` — the first row
// of an unordered read, which on this data is insertion order.
const nameJoins = `
	LEFT JOIN clinlims.localization_value AS nlv
	       ON nlv.localization_id = t.name_localization_id AND nlv.locale = @locale
	LEFT JOIN LATERAL (
	     SELECT st.sample_type_id FROM clinlims.sampletype_test st
	      WHERE st.test_id = t.id ORDER BY st.id LIMIT 1
	) AS firstst ON true
	LEFT JOIN clinlims.type_of_sample AS tos
	       ON tos.id = firstst.sample_type_id
	      AND COALESCE(tos.local_abbrev, '') <> '` + variableSampleTypeAbbrev + `'
	LEFT JOIN clinlims.localization_value AS stlv2
	       ON stlv2.localization_id = tos.name_localization_id AND stlv2.locale = @locale`

// CatalogListRow is one test as the list page, siblings and the group summary
// all read it.
type CatalogListRow struct {
	TestID     string  `gorm:"column:test_id"`
	RawName    string  `gorm:"column:raw_name"`
	Name       string  `gorm:"column:name"`
	SampleType *string `gorm:"column:sample_type"`
	Code       *string `gorm:"column:code"`
	Domain     *string `gorm:"column:domain"`
	Loinc      *string `gorm:"column:loinc"`
	Active     bool    `gorm:"column:active"`
	AMR        bool    `gorm:"column:amr"`
}

// AllTestsForList reads every test with both names.
//
// NO ORDER BY: `testService.getAll()` is a criteria query with none, so the
// filtering and the sort happen in Java over whatever the heap returned. The
// sort is on the RAW name, and the augmented one is only substituted onto the
// page slice — sorting on the augmented name puts the rows in a different
// order.
func (d *EditorReadDAO) AllTestsForList() ([]CatalogListRow, error) {
	return d.listRows("")
}

// ActiveTestsForList is `from Test WHERE is_Active = 'Y' Order by description`,
// which is what siblings walks.
func (d *EditorReadDAO) ActiveTestsForList() ([]CatalogListRow, error) {
	return d.listRows("WHERE t.is_active = 'Y' ORDER BY t.description")
}

func (d *EditorReadDAO) listRows(tail string) ([]CatalogListRow, error) {
	rows := []CatalogListRow{}
	sql := `
		SELECT t.id::text AS test_id,
		       COALESCE(NULLIF(nlv.value, ''), t.description) AS raw_name,
		       ` + d.augmentedName("t") + ` AS name,
		       COALESCE(NULLIF(stlv2.value, ''), tos.description) AS sample_type,
		       t.local_code AS code, t.domain AS domain, t.loinc AS loinc,
		       (t.is_active = 'Y') AS active,
		       COALESCE(t.antimicrobial_resistance, false) AS amr
		  FROM clinlims.test AS t` + nameJoins + " " + tail
	err := d.DB.Raw(sql, map[string]any{"locale": d.locale()}).Scan(&rows).Error
	return rows, err
}

// TestsForSampleType is the one-query id set the list page's sampleType filter
// uses, instead of asking each test for its sample types while filtering.
func (d *EditorReadDAO) TestsForSampleType(sampleTypeID string) ([]string, error) {
	ids := []string{}
	err := d.DB.Raw(
		`SELECT test_id::text FROM clinlims.sampletype_test WHERE sample_type_id = ? ORDER BY id`,
		sampleTypeID).Scan(&ids).Error
	return ids, err
}

// OneTestForList reads a single test in the same shape, for the endpoints that
// take an id.
func (d *EditorReadDAO) OneTestForList(testID string) (*CatalogListRow, error) {
	// Not listRows: that one takes only the locale parameter, and the id has to
	// ride in the same named map rather than be string-built into the predicate.
	rows := []CatalogListRow{}
	err := d.DB.Raw(`
		SELECT t.id::text AS test_id,
		       COALESCE(NULLIF(nlv.value, ''), t.description) AS raw_name,
		       `+d.augmentedName("t")+` AS name,
		       COALESCE(NULLIF(stlv2.value, ''), tos.description) AS sample_type,
		       t.local_code AS code, t.domain AS domain, t.loinc AS loinc,
		       (t.is_active = 'Y') AS active,
		       COALESCE(t.antimicrobial_resistance, false) AS amr
		  FROM clinlims.test AS t`+nameJoins+` WHERE t.id = @id`,
		map[string]any{"locale": d.locale(), "id": testID}).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return &rows[0], nil
}

// ActiveTestsByLoinc is getActiveTestsByLoinc — the same LOINC, active only.
// The resolver that routes analyzer results takes get(0) of this list, which is
// why two rows here is a warning worth surfacing.
func (d *EditorReadDAO) ActiveTestsByLoinc(loinc string) ([]CatalogListRow, error) {
	rows := []CatalogListRow{}
	err := d.DB.Raw(`
		SELECT t.id::text AS test_id,
		       COALESCE(NULLIF(nlv.value, ''), t.description) AS raw_name,
		       `+d.augmentedName("t")+` AS name,
		       COALESCE(NULLIF(stlv2.value, ''), tos.description) AS sample_type,
		       t.local_code AS code, t.domain AS domain, t.loinc AS loinc,
		       (t.is_active = 'Y') AS active,
		       COALESCE(t.antimicrobial_resistance, false) AS amr
		  FROM clinlims.test AS t`+nameJoins+`
		 WHERE t.loinc = @loinc AND t.is_active = 'Y'`,
		map[string]any{"locale": d.locale(), "loinc": loinc}).Scan(&rows).Error
	return rows, err
}

// LocalizationRefRow is the two localization ids a test points at.
type LocalizationRefRow struct {
	NameID      *string `gorm:"column:name_id"`
	ReportingID *string `gorm:"column:reporting_id"`
}

// LocalizationRefs bridges testId → the backing localization ids. The editor
// then reads and writes the per-locale values through /rest/localizations/{id};
// there is no per-test translation store.
func (d *EditorReadDAO) LocalizationRefs(testID string) (*LocalizationRefRow, error) {
	rows := []LocalizationRefRow{}
	err := d.DB.Raw(`
		SELECT name_localization_id::text AS name_id,
		       reporting_name_localization_id::text AS reporting_id
		  FROM clinlims.test WHERE id = ?`, testID).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return &rows[0], nil
}

// DictionaryOptionRow is one typeahead hit.
type DictionaryOptionRow struct {
	ID   string `gorm:"column:id"`
	Name string `gorm:"column:name"`
}

// SearchDictionary ports getDictionaryEntrysByCategoryAbbreviation(filter, null).
//
// The method name says category; the query matches the ENTRY. With a local
// abbreviation present it matches `ABBREV: entry` as one string — so searching
// "pos" finds "College or University" through its abbreviation, not its text.
// Prefix match, case-insensitive, active only, ordered by dict_entry, and the
// controller caps the result at 50.
//
// The label is the RAW dict_entry, not the localized name — the one dictionary
// list in this package that is not localized.
func (d *EditorReadDAO) SearchDictionary(prefix string, limit int) ([]DictionaryOptionRow, error) {
	rows := []DictionaryOptionRow{}
	err := d.DB.Raw(`
		SELECT d.id::text AS id, d.dict_entry AS name
		  FROM clinlims.dictionary AS d
		 WHERE d.is_active = 'Y'
		   AND ( (d.local_abbrev IS NOT NULL
		          AND upper(d.local_abbrev || ': ' || d.dict_entry) LIKE upper(@pattern))
		      OR (d.local_abbrev IS NULL AND upper(d.dict_entry) LIKE upper(@pattern)) )
		 ORDER BY d.dict_entry ASC
		 LIMIT @lim`,
		map[string]any{"pattern": prefix + "%", "lim": limit}).Scan(&rows).Error
	return rows, err
}

// AnalyzerRow is one analyzer that can run a test.
type AnalyzerRow struct {
	AnalyzerID       string  `gorm:"column:analyzer_id"`
	AnalyzerName     *string `gorm:"column:analyzer_name"`
	AnalyzerTestName *string `gorm:"column:analyzer_test_name"`
}

// AnalyzersForTest reads analyzer_test_map. Read-only here: the source of truth
// is the analyzer record, edited on the Analyzer configuration surface.
func (d *EditorReadDAO) AnalyzersForTest(testID string) ([]AnalyzerRow, error) {
	rows := []AnalyzerRow{}
	err := d.DB.Raw(`
		SELECT m.analyzer_id::text AS analyzer_id, a.name AS analyzer_name,
		       m.analyzer_test_name AS analyzer_test_name
		  FROM clinlims.analyzer_test_map AS m
		  LEFT JOIN clinlims.analyzer AS a ON a.id = m.analyzer_id
		 WHERE m.test_id = ?`, testID).Scan(&rows).Error
	return rows, err
}

// ReflexRuleRow is one reflex rule touching a test.
type ReflexRuleRow struct {
	ID                 string  `gorm:"column:id"`
	AddedTestName      *string `gorm:"column:added_test_name"`
	TestResultValue    *string `gorm:"column:test_result_value"`
	NonDictionaryValue *string `gorm:"column:non_dictionary_value"`
	Relation           *string `gorm:"column:relation"`
	InternalNote       *string `gorm:"column:internal_note"`
}

// ReflexRulesForTest reads the reflex rules whose TRIGGER is this test.
//
// The added test's name is getLocalizedName() — NOT the augmented one, unlike
// every other name in this file. Two name rules, one screen.
func (d *EditorReadDAO) ReflexRulesForTest(testID string) ([]ReflexRuleRow, error) {
	rows := []ReflexRuleRow{}
	err := d.DB.Raw(`
		SELECT r.id::text AS id,
		       COALESCE(NULLIF(alv.value, ''), added.description) AS added_test_name,
		       tr.value AS test_result_value,
		       r.non_dictionary_value AS non_dictionary_value,
		       r.relation AS relation,
		       r.internal_note AS internal_note
		  FROM clinlims.test_reflex AS r
		  LEFT JOIN clinlims.test AS added ON added.id = r.add_test_id
		  LEFT JOIN clinlims.localization_value AS alv
		         ON alv.localization_id = added.name_localization_id AND alv.locale = @locale
		  LEFT JOIN clinlims.test_result AS tr ON tr.id = r.tst_rslt_id
		 WHERE r.test_id = @testId
		 ORDER BY r.id`,
		map[string]any{"locale": d.locale(), "testId": testID}).Scan(&rows).Error
	return rows, err
}

// CalculationRow is one calculation, with its output test resolved.
type CalculationRow struct {
	ID             string  `gorm:"column:id"`
	Name           *string `gorm:"column:name"`
	Result         *string `gorm:"column:result"`
	TestID         *string `gorm:"column:test_id"`
	OutputTestName *string `gorm:"column:output_test_name"`
	Active         *bool   `gorm:"column:active"`
}

// ActiveCalculations reads every calculation that is not explicitly inactive —
// `Boolean.FALSE.equals(getActive())` skips only an explicit false, so a NULL
// active column stays in the list.
func (d *EditorReadDAO) ActiveCalculations() ([]CalculationRow, error) {
	rows := []CalculationRow{}
	err := d.DB.Raw(`
		SELECT c.id::text AS id, c.name AS name, c.result AS result,
		       c.test_id::text AS test_id,
		       COALESCE(NULLIF(olv.value, ''), out.description) AS output_test_name,
		       c.active AS active
		  FROM clinlims.calculation AS c
		  LEFT JOIN clinlims.test AS out ON out.id = c.test_id
		  LEFT JOIN clinlims.localization_value AS olv
		         ON olv.localization_id = out.name_localization_id AND olv.locale = @locale
		 WHERE c.active IS DISTINCT FROM false
		 ORDER BY c.id`,
		map[string]any{"locale": d.locale()}).Scan(&rows).Error
	return rows, err
}

// OperationRow is one term of a calculation's formula.
type OperationRow struct {
	CalculationID string  `gorm:"column:calculation_id"`
	Type          *string `gorm:"column:type"`
	Value         *string `gorm:"column:value"`
}

// Operations reads every calculation's terms, in the order the formula joins
// them — Operation implements Comparable on its order column.
func (d *EditorReadDAO) Operations() ([]OperationRow, error) {
	rows := []OperationRow{}
	err := d.DB.Raw(`
		SELECT calculation_id::text AS calculation_id, type, value
		  FROM clinlims.calculation_operation
		 ORDER BY calculation_id, operation_order, id`).Scan(&rows).Error
	return rows, err
}

// StorageHistoryRow is one test_sample_handling_history row. The controller
// serialises the ENTITY, so these are its bean properties rather than a DTO's.
type StorageHistoryRow struct {
	ID                   string  `gorm:"column:id"`
	TestSampleHandlingID string  `gorm:"column:test_sample_handling_id"`
	ChangedBy            *string `gorm:"column:changed_by"`
	ChangedAt            *int64  `gorm:"column:changed_at"`
	ChangeType           *string `gorm:"column:change_type"`
	PreviousValues       *string `gorm:"column:previous_values"`
	NewValues            *string `gorm:"column:new_values"`
	Lastupdated          *int64  `gorm:"column:lastupdated"`
}

// StorageHistory reads a test's storage change trail, newest first.
//
// A test with no handling row answers an EMPTY LIST, not a 404 — only a missing
// TEST is a 404. The two jsonb columns come back as STRINGS on the wire, because
// the entity types them as String and Jackson has nothing to parse them into.
func (d *EditorReadDAO) StorageHistory(testID string) ([]StorageHistoryRow, error) {
	rows := []StorageHistoryRow{}
	err := d.DB.Raw(`
		SELECT h.id AS id, h.test_sample_handling_id AS test_sample_handling_id,
		       h.changed_by::text AS changed_by,
		       trunc(EXTRACT(EPOCH FROM h.changed_at) * 1000)::bigint AS changed_at,
		       h.change_type AS change_type,
		       h.previous_values::text AS previous_values,
		       h.new_values::text AS new_values,
		       trunc(EXTRACT(EPOCH FROM h.lastupdated) * 1000)::bigint AS lastupdated
		  FROM clinlims.test_sample_handling_history AS h
		 WHERE h.test_sample_handling_id = (
		       SELECT s.id FROM clinlims.test_sample_handling s
		        WHERE s.test_id = ? ORDER BY s.id LIMIT 1)
		 ORDER BY h.changed_at DESC, h.id DESC`, testID).Scan(&rows).Error
	return rows, err
}

// IsOrderable answers the second half of the noLoinc warning: a test that
// cannot receive results is not warned about for lacking a LOINC.
func (d *EditorReadDAO) IsOrderable(testID string) (bool, error) {
	rows := []bool{}
	err := d.DB.Raw(
		`SELECT COALESCE(orderable, false) FROM clinlims.test WHERE id = ?`, testID).
		Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return false, err
	}
	return rows[0], nil
}
