package daoimpl

import (
	"math"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"openelis-go/internal/common/audittrail"
)

// TestAddDAO backs TestAdd — the widest write in the wave.
//
// One submission creates ONE TEST PER SAMPLE TYPE it names, and each of those
// drags a fan of rows behind it. MEASURED against Java, one numeric test and
// one dictionary test:
//
//	localization              2   name and reporting name, shared by every set
//	localization_value        4   'en' and 'fr' for each
//	test                      1   per sample type
//	test_terminology_mapping  1   per test, and only when the loinc is non-blank
//	sampletype_test           1   per test
//	panel_item                n   per test, one per panel named
//	test_result               n   one for numeric/text, one per dictionary
//	                              option for the dictionary variants, and NONE
//	                              for Titer — which is in no branch of
//	                              createTestResults
//	result_limits             n   per test
//	test_result_component     1   per test, the PRIMARY the new editor scopes by
//
// Three side effects turn rows back ON without being asked to: an inactive test
// section and an inactive panel are activated by being named, and the sample
// type follows the new test's own active flag — so submitting active="N"
// DEACTIVATES a live sample type.
//
// The audit is much narrower than the write. Measured: history holds one 'I'
// for the test, one 'I' per result limit, and one 'U' for the sample type ONLY
// when its flag actually changed. The localizations, the join rows, the panel
// items, the test results, the terminology mapping and the component are all
// silent — several of them from tables flagged keep_history='Y'.
type TestAddDAO struct {
	DB    *gorm.DB
	Audit *audittrail.Service
}

// TestAddSet is one sample type's worth of the write.
type TestAddSet struct {
	SampleTypeID string
	// OrderedTests is the sibling ordering the form submits. The new test takes
	// the index holding "0"; every other id is re-sorted to its own index.
	OrderedTests []string
}

// TestAddRow is the test being created, shared across every set.
type TestAddRow struct {
	NameEnglish             string
	NameFrench              string
	ReportNameEnglish       string
	ReportNameFrench        string
	TestSectionID           string
	UomID                   string
	Loinc                   string
	Active                  string
	Orderable               bool
	NotifyResults           bool
	InLabOnly               bool
	AntimicrobialResistance bool
	PanelIDs                []string
	// ResultTypeID is the type_of_test_result id the form submitted;
	// ResultTypeChar is its one-letter DB value. result_limits keys on the id,
	// test_result on the character.
	ResultTypeID          string
	ResultTypeChar        string
	SignificantDigits     string
	Dictionaries          []TestAddDictionary
	Limits                []TestAddLimit
	DictionaryReferenceID string
	// The six global numeric bounds. createResultLimits stamps the same valid
	// pair onto EVERY limit row, and writes the reporting range and the
	// criticals only when both criticals parsed.
	LowValid, HighValid         *float64
	LowReporting, HighReporting *float64
	LowCritical, HighCritical   *float64
}

// TestAddDictionary is one option of a dictionary-variant result.
type TestAddDictionary struct {
	DictionaryID   string
	IsQuantifiable bool
	IsDefault      bool
}

// TestAddLimit is one result_limits row, after extractLimits has split a
// gendered entry into its M and F halves.
type TestAddLimit struct {
	Gender     string
	MinAge     float64
	MaxAge     float64
	LowNormal  float64
	HighNormal float64
}

// The ResultLimit entity's own field defaults, which Hibernate writes for every
// column the form did not fill. They are NOT the column defaults — the two
// disagree on low_critical, where the column says -Infinity and the entity says
// +Infinity, and the entity wins because Hibernate names every column on the
// insert. Measured on a dictionary test, whose limit row Java writes without
// touching either critical.
var (
	limitDefaultMinAge       = 0.0
	limitDefaultLowNormal    = math.Inf(-1)
	limitDefaultLowValid     = math.Inf(-1)
	limitDefaultLowReporting = math.Inf(-1)
	limitDefaultLowCritical  = math.Inf(1)
)

// Add runs the whole write in one transaction — addTests is @Transactional, and
// a test whose sampletype_test row is missing is invisible to every screen that
// lists tests by sample type.
func (d *TestAddDAO) Add(row TestAddRow, sets []TestAddSet, sysUserID int64) ([]string, error) {
	created := []string{}
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		ts := time.Now().UTC().Truncate(time.Millisecond)

		// The two localizations are written ONCE, before the loop, and every set
		// points at the same pair. Their descriptions are the literals
		// LocalizationType carries, not the names being stored.
		nameLoc, err := insertTestLocalization(tx, "test name", row.NameEnglish, row.NameFrench, ts)
		if err != nil {
			return err
		}
		reportLoc, err := insertTestLocalization(tx, "test report name",
			row.ReportNameEnglish, row.ReportNameFrench, ts)
		if err != nil {
			return err
		}

		for _, set := range sets {
			sampleTypeDesc, ok, err := scanOne(tx,
				`SELECT COALESCE(description, '') FROM clinlims.type_of_sample WHERE id = ?`,
				set.SampleTypeID)
			if err != nil {
				return err
			}
			if !ok {
				// getTypeOfSampleById returned null and `continue` skips the
				// whole set — no test, no join row, no results.
				continue
			}

			testID, err := d.insertTest(tx, row, set, sampleTypeDesc, nameLoc, reportLoc, ts, sysUserID)
			if err != nil {
				return err
			}
			created = append(created, testID)

			if err := d.syncLegacyLoinc(tx, testID, row.Loinc, ts); err != nil {
				return err
			}
			if err := activateIfInactive(tx, "clinlims.test_section", row.TestSectionID, ts); err != nil {
				return err
			}
			if err := reorderSiblings(tx, set.OrderedTests, ts); err != nil {
				return err
			}
			if err := d.updateSampleTypeActive(tx, set.SampleTypeID, row.Active == "Y", ts, sysUserID); err != nil {
				return err
			}
			if err := tx.Exec(`
				INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id)
				VALUES (nextval('clinlims.sample_type_test_seq'), ?, ?)`,
				set.SampleTypeID, testID).Error; err != nil {
				return err
			}
			for _, panelID := range row.PanelIDs {
				if err := tx.Exec(`
					INSERT INTO clinlims.panel_item (id, panel_id, test_id, lastupdated)
					VALUES (nextval('clinlims.panel_item_seq'), ?, ?, ?)`,
					panelID, testID, ts).Error; err != nil {
					return err
				}
				if err := activateIfInactive(tx, "clinlims.panel", panelID, ts); err != nil {
					return err
				}
			}
			if err := d.insertResults(tx, testID, row, ts); err != nil {
				return err
			}
			if err := d.insertLimits(tx, testID, row, ts, sysUserID); err != nil {
				return err
			}
			if err := d.syncPrimaryComponent(tx, testID, row, ts); err != nil {
				return err
			}
		}
		return nil
	})
	return created, err
}

// insertTest writes the test row and its history.
//
// normalized_description is NOT set here: a BEFORE INSERT trigger on the table
// derives it from the description, so both stacks get the same value from the
// same place.
func (d *TestAddDAO) insertTest(tx *gorm.DB, row TestAddRow, set TestAddSet,
	sampleTypeDesc, nameLoc, reportLoc string, ts time.Time, sysUserID int64) (string, error) {

	// The description is the English name with the sample type in brackets —
	// which is where the doubled test names the rest of this wave reads come
	// from. `name` and `local_code` are both the bare English name.
	description := row.NameEnglish + "(" + sampleTypeDesc + ")"

	var testID string
	if err := tx.Raw(`
		INSERT INTO clinlims.test
		       (id, name, description, local_code, loinc, is_active, orderable,
		        notify_results, in_lab_only, antimicrobial_resistance, is_reportable,
		        test_section_id, uom_id, sort_order, guid, domain,
		        name_localization_id, reporting_name_localization_id, lastupdated)
		VALUES (nextval('clinlims.test_seq'), ?, ?, ?, ?, ?, ?, ?, ?, ?, 'N',
		        ?, ?, ?, gen_random_uuid()::text, 'CLINICAL', ?, ?, ?)
		RETURNING id::text`,
		row.NameEnglish, description, row.NameEnglish, row.Loinc,
		row.Active, row.Orderable, row.NotifyResults, row.InLabOnly,
		row.AntimicrobialResistance,
		nullIfBlank(row.TestSectionID), nullIfBlank(row.UomID),
		newTestSortOrder(set.OrderedTests), nameLoc, reportLoc, ts).
		Scan(&testID).Error; err != nil {
		return "", err
	}
	// Insert audit rows carry a NULL payload.
	if err := d.Audit.Write(tx, "TEST", testID, sysUserID, audittrail.ActivityInsert, nil, ts); err != nil {
		return "", err
	}
	return testID, nil
}

// newTestSortOrder is the index the form marked with "0".
//
// Java only calls setSortOrder inside that branch, so a list with no "0" leaves
// the field null and the column NULL — the column's own default of 2147483647
// never applies, because Hibernate names the column either way.
func newTestSortOrder(orderedTests []string) any {
	for i, id := range orderedTests {
		if id == "0" {
			return i
		}
	}
	return nil
}

// reorderSiblings re-sorts every OTHER test in the list to its own index.
func reorderSiblings(tx *gorm.DB, orderedTests []string, ts time.Time) error {
	for i, id := range orderedTests {
		if id == "0" {
			continue
		}
		// testService.get() would throw on an unknown id, but these come from
		// the list the screen itself rendered.
		if err := tx.Exec(
			`UPDATE clinlims.test SET sort_order = ?, lastupdated = ? WHERE id = ?`,
			i, ts, id).Error; err != nil {
			return err
		}
	}
	return nil
}

// activateIfInactive turns a test section or a panel back on.
//
// Guarded on `"N".equals(getIsActive())` in Java, which is the WHERE clause
// here — a row already active is not touched, so no lastupdated moves.
func activateIfInactive(tx *gorm.DB, table, id string, ts time.Time) error {
	if strings.TrimSpace(id) == "" {
		return nil
	}
	return tx.Exec(
		`UPDATE `+table+` SET is_active = 'Y', lastupdated = ? WHERE id = ? AND is_active = 'N'`,
		ts, id).Error
}

// updateSampleTypeActive ports `typeOfSample.setActive("Y".equals(active))`.
//
// MEASURED: when the flag already holds the submitted value, Hibernate finds
// the entity clean and issues neither the UPDATE nor the history row —
// type_of_sample 1 kept its 2012 lastupdated across a create that named it.
// When it does change, the 'U' payload carries the value being REPLACED.
func (d *TestAddDAO) updateSampleTypeActive(tx *gorm.DB, sampleTypeID string, active bool,
	ts time.Time, sysUserID int64) error {

	old, ok, err := scanOne(tx,
		`SELECT COALESCE(is_active, false)::text FROM clinlims.type_of_sample WHERE id = ?`, sampleTypeID)
	if err != nil || !ok {
		return err
	}
	if old == strconv.FormatBool(active) {
		return nil
	}
	if err := tx.Exec(
		`UPDATE clinlims.type_of_sample SET is_active = ?, lastupdated = ? WHERE id = ?`,
		active, ts, sampleTypeID).Error; err != nil {
		return err
	}
	changes := audittrail.Field("isActive", old)
	return d.Audit.Write(tx, "TYPE_OF_SAMPLE", sampleTypeID, sysUserID,
		audittrail.ActivityUpdate, &changes, ts)
}

// syncLegacyLoinc ports TestTerminologyMappingServiceImpl.syncLegacyLoinc.
//
// On a brand new test there is nothing to deactivate and nothing to reuse, so
// the whole method reduces to one insert — and to nothing at all when the loinc
// is blank. test_terminology_mapping has no reference_tables row, so it is not
// audited.
func (d *TestAddDAO) syncLegacyLoinc(tx *gorm.DB, testID, loinc string, ts time.Time) error {
	code := strings.TrimSpace(loinc)
	if code == "" {
		return nil
	}
	return tx.Exec(`
		INSERT INTO clinlims.test_terminology_mapping
		       (id, test_id, source, code, relationship, is_active, lastupdated)
		VALUES (gen_random_uuid()::text, ?, 'LOINC', ?, 'SAME_AS', 'Y', ?)`,
		testID, code, ts).Error
}

// insertResults ports createTestResults, and sets the test's default result.
//
// A text-only ("A"/"R") or numeric ("N") type gets ONE row at sort order 1; a
// dictionary variant ("D"/"M"/"C") gets one per option at 10, 20, 30…; Titer
// ("T") matches neither branch and gets NONE, so a Titer test is created with
// no results at all.
//
// `testResult.getDefault()` sends the chosen option back onto the test. Java
// only mutates the in-memory entity there, but it is a managed one, so the
// flush writes default_test_result_id — MEASURED: the probe test came back
// pointing at its second option. No history row is written for that update.
func (d *TestAddDAO) insertResults(tx *gorm.DB, testID string, row TestAddRow, ts time.Time) error {
	switch {
	case isDictionaryVariant(row.ResultTypeChar):
		order := 10
		for _, dict := range row.Dictionaries {
			var resultID string
			// significant_digits is named even though the dictionary branch never
			// sets it: the COLUMN defaults to 0, Hibernate writes the unset field
			// as NULL, and leaving the column out would silently store the 0.
			if err := tx.Raw(`
				INSERT INTO clinlims.test_result
				       (id, test_id, tst_rslt_type, sort_order, is_active, value,
				        is_quantifiable, is_normal, significant_digits, lastupdated)
				VALUES (nextval('clinlims.test_result_seq'), ?, ?, ?, true, ?, ?, false, NULL, ?)
				RETURNING id::text`,
				testID, row.ResultTypeChar, order, dict.DictionaryID,
				dict.IsQuantifiable, ts).Scan(&resultID).Error; err != nil {
				return err
			}
			if dict.IsDefault {
				if err := tx.Exec(
					`UPDATE clinlims.test SET default_test_result_id = ? WHERE id = ?`,
					resultID, testID).Error; err != nil {
					return err
				}
			}
			order += 10
		}
	case isTextOnlyVariant(row.ResultTypeChar) || row.ResultTypeChar == "N":
		return tx.Exec(`
			INSERT INTO clinlims.test_result
			       (id, test_id, tst_rslt_type, sort_order, is_active, significant_digits,
			        is_quantifiable, is_normal, lastupdated)
			VALUES (nextval('clinlims.test_result_seq'), ?, ?, 1, true, ?, false, false, ?)`,
			testID, row.ResultTypeChar, nullIfBlank(row.SignificantDigits), ts).Error
	}
	return nil
}

// insertLimits ports createResultLimits and createDictionaryResultLimit.
//
// RESULT_LIMITS is one of the two tables here that IS audited, one 'I' per row.
func (d *TestAddDAO) insertLimits(tx *gorm.DB, testID string, row TestAddRow,
	ts time.Time, sysUserID int64) error {

	insert := func(gender any, minAge, maxAge, lowNormal, highNormal,
		lowValid, highValid, lowReporting, highReporting, lowCritical, highCritical float64,
		dictionaryNormalID any) error {

		var limitID string
		if err := tx.Raw(`
			INSERT INTO clinlims.result_limits
			       (id, test_id, test_result_type_id, gender, min_age, max_age,
			        low_normal, high_normal, low_valid, high_valid,
			        low_reporting_range, high_reporting_range, low_critical, high_critical,
			        normal_dictionary_id, always_validate, lastupdated)
			VALUES (nextval('clinlims.result_limits_seq'), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, false, ?)
			RETURNING id::text`,
			testID, row.ResultTypeID, gender, minAge, maxAge,
			lowNormal, highNormal, lowValid, highValid,
			lowReporting, highReporting, lowCritical, highCritical,
			dictionaryNormalID, ts).Scan(&limitID).Error; err != nil {
			return err
		}
		return d.Audit.Write(tx, "RESULT_LIMITS", limitID, sysUserID,
			audittrail.ActivityInsert, nil, ts)
	}

	if isDictionaryVariant(row.ResultTypeChar) {
		// One row, and only when a reference was chosen. Every numeric column
		// keeps the ENTITY default, including the two criticals at +Infinity.
		if strings.TrimSpace(row.DictionaryReferenceID) == "" {
			return nil
		}
		return insert(nil, limitDefaultMinAge, math.Inf(1),
			limitDefaultLowNormal, math.Inf(1),
			limitDefaultLowValid, math.Inf(1),
			limitDefaultLowReporting, math.Inf(1),
			limitDefaultLowCritical, math.Inf(1),
			row.DictionaryReferenceID)
	}

	// Numeric. The reporting range and the criticals are written only when BOTH
	// criticals parsed; otherwise all four keep the entity defaults.
	writeCriticals := row.LowCritical != nil && row.HighCritical != nil
	for _, l := range row.Limits {
		lowReporting, highReporting := limitDefaultLowReporting, math.Inf(1)
		lowCritical, highCritical := limitDefaultLowCritical, math.Inf(1)
		if writeCriticals {
			lowReporting = derefFloat(row.LowReporting, lowReporting)
			highReporting = derefFloat(row.HighReporting, highReporting)
			lowCritical, highCritical = *row.LowCritical, *row.HighCritical
		}
		if err := insert(nullIfBlank(l.Gender), l.MinAge, l.MaxAge,
			l.LowNormal, l.HighNormal,
			derefFloat(row.LowValid, limitDefaultLowValid),
			derefFloat(row.HighValid, math.Inf(1)),
			lowReporting, highReporting, lowCritical, highCritical, nil); err != nil {
			return err
		}
	}
	return nil
}

// syncPrimaryComponent ports
// TestResultComponentServiceImpl.syncPrimaryComponentFromLegacy.
//
// On a new test there is never an existing component, so it always inserts —
// then repoints the results and the limits it just wrote, which legacy created
// with a NULL component_id, onto it. Neither the component nor the repointing
// is audited.
//
// resultType and significantDigits come from the NEWEST result by id, which for
// a dictionary variant is the LAST option submitted.
func (d *TestAddDAO) syncPrimaryComponent(tx *gorm.DB, testID string, row TestAddRow, ts time.Time) error {
	var newest struct {
		ResultType        *string `gorm:"column:tst_rslt_type"`
		SignificantDigits *int    `gorm:"column:significant_digits"`
	}
	if err := tx.Raw(`
		SELECT tst_rslt_type, significant_digits
		  FROM clinlims.test_result
		 WHERE test_id = ? AND is_active = true
		 ORDER BY id DESC LIMIT 1`, testID).Scan(&newest).Error; err != nil {
		return err
	}

	// primaryLabel falls back to the literal "PRIMARY" for a blank test name.
	label := row.NameEnglish
	if strings.TrimSpace(label) == "" {
		label = "PRIMARY"
	}

	var componentID string
	if err := tx.Raw(`
		INSERT INTO clinlims.test_result_component
		       (id, test_id, code, label, display_order, result_type, uom_id,
		        significant_digits, allow_multiple_readings, is_active, lastupdated)
		VALUES (gen_random_uuid()::text, ?, 'PRIMARY', ?, 0, ?, ?, ?, false, 'Y', ?)
		RETURNING id`,
		testID, label, newest.ResultType, nullIfBlank(row.UomID),
		newest.SignificantDigits, ts).Scan(&componentID).Error; err != nil {
		return err
	}

	if err := tx.Exec(
		`UPDATE clinlims.test_result SET component_id = ? WHERE test_id = ? AND component_id IS NULL`,
		componentID, testID).Error; err != nil {
		return err
	}
	return tx.Exec(
		`UPDATE clinlims.result_limits SET component_id = ? WHERE test_id = ? AND component_id IS NULL`,
		componentID, testID).Error
}

// insertTestLocalization writes one localization plus its 'en' and 'fr' values.
//
// Only those two locales are written, however many the deployment has active —
// createNewLocalization calls setEnglish and setFrench and nothing else.
func insertTestLocalization(tx *gorm.DB, description, english, french string, ts time.Time) (string, error) {
	var id string
	if err := tx.Raw(`
		INSERT INTO clinlims.localization (id, description, lastupdated)
		VALUES (nextval('clinlims.localization_seq'), ?, ?)
		RETURNING id::text`, description, ts).Scan(&id).Error; err != nil {
		return "", err
	}
	for _, lv := range []struct{ locale, value string }{{"en", english}, {"fr", french}} {
		if err := tx.Exec(`
			INSERT INTO clinlims.localization_value (id, localization_id, locale, value, last_updated)
			VALUES (nextval('clinlims.localization_value_seq'), ?, ?, ?, ?)`,
			id, lv.locale, lv.value, ts).Error; err != nil {
			return "", err
		}
	}
	return id, nil
}

// isDictionaryVariant is ResultType.isDictionaryVariant: "DMC".contains(type).
// A blank type is not one, and — a quirk of contains — a multi-letter string is
// unless it is not a substring of "DMC".
func isDictionaryVariant(t string) bool {
	return t != "" && strings.Contains("DMC", t)
}

// isTextOnlyVariant is ResultType.isTextOnlyVariant: "AR".contains(type).
func isTextOnlyVariant(t string) bool {
	return t != "" && strings.Contains("AR", t)
}

func scanOne(tx *gorm.DB, sql string, args ...any) (string, bool, error) {
	rows := []string{}
	if err := tx.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return "", false, err
	}
	if len(rows) == 0 {
		return "", false, nil
	}
	return rows[0], true, nil
}

func nullIfBlank(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func derefFloat(p *float64, fallback float64) float64 {
	if p == nil {
		return fallback
	}
	return *p
}
