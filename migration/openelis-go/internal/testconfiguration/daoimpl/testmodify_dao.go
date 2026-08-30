package daoimpl

import (
	"math"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"openelis-go/internal/common/audittrail"
	"openelis-go/internal/common/util"
)

// TestModifyDAO backs TestModifyEntry's POST — a DELETE-THEN-INSERT rewrite of
// everything hanging off one test.
//
// MEASURED against Java, one numeric modify and one dictionary modify:
//
//	sampletype_test   every row for the test deleted, then one inserted
//	panel_item        same
//	result_limits     same — and this is the ONLY audited table of the three
//	test              updated in place: loinc, uom, notify, in-lab, AMR, section,
//	                  active, orderable — and NOTHING else. description and
//	                  local_code keep the names the test was CREATED with, so a
//	                  renamed test still describes itself by its old name.
//	localization      the two existing rows are edited in place, for every
//	                  ACTIVE locale rather than just en and fr
//	test_result       INSERTED, never updated — see addResults for what that
//	                  does to a numeric test
//	sampletype_panel  one row, and only when the sample type had none
//
// `test.name` is not written by any of that and moves anyway: Hibernate maps
// the column to Test.getName(), a DERIVED getter that returns the localization's
// value. So editing the localization renames the column on the next flush of
// the test — which is what makes the rename screens work, and what makes
// description and local_code look stale beside it.
//
// The audit is three rows for the whole endpoint: a 'D' per deleted result
// limit, carrying the full row, and an 'I' per inserted one. The test updates,
// the localization edits, the join-row churn and the component sync are silent.
type TestModifyDAO struct {
	DB    *gorm.DB
	Audit *audittrail.Service
}

// TestModifyRow is the submission, shared across every set.
type TestModifyRow struct {
	TestID string
	TestAddRow
	// Locales is every ACTIVE supported locale. updateTestNames writes the
	// submitted English into all of them — it reads
	// nameLocalization.getLocalizedValue(locale) from a Localization built with
	// only en and fr set, and getLocalizedValue falls back to the English.
	Locales []string
}

// DeactivateDictionaryResults ports createDictionaryResultLimit's side effect.
//
// It turns every ACTIVE result of the test off, and it runs from
// createTestSets — BEFORE updateTestSets and therefore OUTSIDE its
// transaction. A modify that then fails leaves the test with no active results
// at all, so it is its own statement here too.
//
// Only the dictionary variants reach it. A NUMERIC modify never deactivates
// anything, which is why addResults leaves the old row active beside the new.
func (d *TestModifyDAO) DeactivateDictionaryResults(testID string) error {
	return d.DB.Exec(
		`UPDATE clinlims.test_result SET is_active = false WHERE test_id = ? AND is_active = true`,
		testID).Error
}

// Update runs the rewrite in one transaction — updateTestSets is @Transactional.
func (d *TestModifyDAO) Update(row TestModifyRow, sets []TestAddSet, sysUserID int64) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		ts := time.Now().UTC().Truncate(time.Millisecond)

		if err := tx.Exec(
			`DELETE FROM clinlims.sampletype_test WHERE test_id = ?`, row.TestID).Error; err != nil {
			return err
		}
		if err := tx.Exec(
			`DELETE FROM clinlims.panel_item WHERE test_id = ?`, row.TestID).Error; err != nil {
			return err
		}
		if err := d.deleteLimits(tx, row.TestID, sysUserID, ts); err != nil {
			return err
		}

		for _, set := range sets {
			if _, ok, err := scanOne(tx,
				`SELECT id::text FROM clinlims.type_of_sample WHERE id = ?`,
				set.SampleTypeID); err != nil {
				return err
			} else if !ok {
				// createTestSets skips a set whose sample type does not resolve,
				// and it runs BEFORE updateTestSets — so the three deletes above
				// have already happened and nothing puts them back. A modify
				// naming an unknown sample type strips the test of its joins and
				// its limits and answers 200.
				continue
			}
			if err := activateIfInactive(tx, "clinlims.test_section", row.TestSectionID, ts); err != nil {
				return err
			}
			// sortedTests here holds the modified test TOO — TestModifyEntry adds
			// it to the list where TestAdd leaves it out, so the "0" entry moves
			// the test being edited rather than being skipped.
			if err := d.reorderIncludingSelf(tx, set.OrderedTests, row.TestID, ts); err != nil {
				return err
			}
			if err := d.updateNames(tx, row, ts); err != nil {
				return err
			}
			if err := d.updateTest(tx, row, ts); err != nil {
				return err
			}
			if err := d.syncLegacyLoinc(tx, row.TestID, row.Loinc, ts); err != nil {
				return err
			}
			if err := tx.Exec(`
				INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id)
				VALUES (nextval('clinlims.sample_type_test_seq'), ?, ?)`,
				set.SampleTypeID, row.TestID).Error; err != nil {
				return err
			}
			if err := d.addPanels(tx, row, set.SampleTypeID, ts); err != nil {
				return err
			}
			if err := d.addResults(tx, row, ts); err != nil {
				return err
			}
			if err := d.addLimits(tx, row, sysUserID, ts); err != nil {
				return err
			}
			if err := d.syncComponent(tx, row, ts); err != nil {
				return err
			}
		}
		return nil
	})
}

// deleteLimits removes every result limit of the test, auditing each with the
// row it removed.
func (d *TestModifyDAO) deleteLimits(tx *gorm.DB, testID string, sysUserID int64, ts time.Time) error {
	type limitRow struct {
		ID                 string  `gorm:"column:id"`
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
		AlwaysValidate     bool    `gorm:"column:always_validate"`
		ComponentID        string  `gorm:"column:component_id"`
		Lastupdated        *string `gorm:"column:lastupdated"`
	}
	rows := []limitRow{}
	if err := tx.Raw(`
		SELECT id::text AS id, test_result_type_id::text AS result_type_id,
		       COALESCE(gender, '') AS gender,
		       COALESCE(min_age, 0) AS min_age,
		       COALESCE(max_age, 'Infinity'::float8) AS max_age,
		       COALESCE(low_normal, '-Infinity'::float8) AS low_normal,
		       COALESCE(high_normal, 'Infinity'::float8) AS high_normal,
		       COALESCE(low_valid, '-Infinity'::float8) AS low_valid,
		       COALESCE(high_valid, 'Infinity'::float8) AS high_valid,
		       COALESCE(low_reporting_range, '-Infinity'::float8) AS low_reporting_range,
		       COALESCE(high_reporting_range, 'Infinity'::float8) AS high_reporting_range,
		       COALESCE(low_critical, 'Infinity'::float8) AS low_critical,
		       COALESCE(high_critical, 'Infinity'::float8) AS high_critical,
		       COALESCE(normal_dictionary_id::text, '') AS dictionary_normal_id,
		       COALESCE(always_validate, false) AS always_validate,
		       COALESCE(component_id, '') AS component_id,
		       to_char(lastupdated, 'YYYY-MM-DD HH24:MI:SS.MS') AS lastupdated
		  FROM clinlims.result_limits WHERE test_id = ? ORDER BY id`, testID).
		Scan(&rows).Error; err != nil {
		return err
	}

	for _, r := range rows {
		if err := tx.Exec(`DELETE FROM clinlims.result_limits WHERE id = ?`, r.ID).Error; err != nil {
			return err
		}
		// getChanges walks the entity's DECLARED FIELDS and emits only those
		// that differ from a blank ResultLimit — which is why `gender` is absent
		// when it is null and `dictionaryNormalId` is absent on a numeric limit,
		// and why the numeric bounds appear even when they hold the defaults
		// (they are compared against a blank object built by the SAME
		// constructor, so equal values would drop out — measured: they do not,
		// because the delete payload emits the persistent state wholesale).
		changes := audittrail.Field("testId", testID) +
			audittrail.Field("resultTypeId", r.ResultTypeID)
		if r.Gender != "" {
			changes += audittrail.Field("gender", r.Gender)
		}
		changes += audittrail.Field("minAge", javaDouble(r.MinAge)) +
			audittrail.Field("highCritical", javaDouble(r.HighCritical)) +
			audittrail.Field("lowCritical", javaDouble(r.LowCritical)) +
			audittrail.Field("maxAge", javaDouble(r.MaxAge)) +
			audittrail.Field("lowNormal", javaDouble(r.LowNormal)) +
			audittrail.Field("highNormal", javaDouble(r.HighNormal)) +
			audittrail.Field("lowValid", javaDouble(r.LowValid)) +
			audittrail.Field("highValid", javaDouble(r.HighValid)) +
			audittrail.Field("lowReportingRange", javaDouble(r.LowReportingRange)) +
			audittrail.Field("highReportingRange", javaDouble(r.HighReportingRange))
		if r.DictionaryNormalID != "" {
			changes += audittrail.Field("dictionaryNormalId", r.DictionaryNormalID)
		}
		changes += audittrail.Field("alwaysValidate", strconv.FormatBool(r.AlwaysValidate))
		if r.ComponentID != "" {
			changes += audittrail.Field("componentId", r.ComponentID)
		}
		if r.Lastupdated != nil {
			changes += audittrail.Field("lastupdated", *r.Lastupdated)
		}
		if err := d.Audit.Write(tx, "RESULT_LIMITS", r.ID, sysUserID,
			audittrail.ActivityDelete, &changes, ts); err != nil {
			return err
		}
	}
	return nil
}

// reorderIncludingSelf writes each id in the submitted order to its own index,
// with the "0" slot standing for the test being modified.
func (d *TestModifyDAO) reorderIncludingSelf(tx *gorm.DB, orderedTests []string, testID string, ts time.Time) error {
	for i, id := range orderedTests {
		target := id
		if id == "0" {
			target = testID
		}
		if err := tx.Exec(
			`UPDATE clinlims.test SET sort_order = ?, lastupdated = ? WHERE id = ?`,
			i, ts, target).Error; err != nil {
			return err
		}
	}
	return nil
}

// updateNames rewrites the two EXISTING localizations in place.
//
// The rows are not replaced and their ids do not change, so every test pointing
// at a shared localization is renamed with this one. The write covers every
// ACTIVE locale, not the en/fr pair the create path writes.
func (d *TestModifyDAO) updateNames(tx *gorm.DB, row TestModifyRow, ts time.Time) error {
	var ids struct {
		Name      *string `gorm:"column:name_localization_id"`
		Reporting *string `gorm:"column:reporting_name_localization_id"`
	}
	if err := tx.Raw(`
		SELECT name_localization_id::text, reporting_name_localization_id::text
		  FROM clinlims.test WHERE id = ?`, row.TestID).Scan(&ids).Error; err != nil {
		return err
	}

	write := func(locID *string, english, french string) error {
		if locID == nil || *locID == "" {
			return nil
		}
		for _, locale := range row.Locales {
			// getLocalizedValue(locale) on a Localization carrying only en and
			// fr answers the ENGLISH for any other locale, so a third supported
			// language is overwritten with the English text.
			value := strings.TrimSpace(english)
			if locale == "fr" {
				value = strings.TrimSpace(french)
			}
			if err := tx.Exec(`
				UPDATE clinlims.localization_value SET value = ?, last_updated = ?
				 WHERE localization_id = ? AND locale = ?`,
				value, ts, *locID, locale).Error; err != nil {
				return err
			}
		}
		return tx.Exec(`UPDATE clinlims.localization SET lastupdated = ? WHERE id = ?`,
			ts, *locID).Error
	}

	if err := write(ids.Name, row.NameEnglish, row.NameFrench); err != nil {
		return err
	}
	return write(ids.Reporting, row.ReportNameEnglish, row.ReportNameFrench)
}

// updateTest ports updateTestEntities.
//
// EIGHT columns, and `name` as the ninth because Hibernate derives it from the
// localization. description, local_code, guid, is_reportable and sort_order are
// all left where they were.
func (d *TestModifyDAO) updateTest(tx *gorm.DB, row TestModifyRow, ts time.Time) error {
	return tx.Exec(`
		UPDATE clinlims.test
		   SET loinc = ?, uom_id = ?, notify_results = ?, in_lab_only = ?,
		       antimicrobial_resistance = ?, test_section_id = ?, is_active = ?,
		       orderable = ?,
		       name = COALESCE(NULLIF((SELECT lv.value FROM clinlims.localization_value lv
		                                WHERE lv.localization_id = clinlims.test.name_localization_id
		                                  AND lv.locale = ?), ''), clinlims.test.description),
		       lastupdated = ?
		 WHERE id = ?`,
		row.Loinc, nullIfBlank(row.UomID), row.NotifyResults, row.InLabOnly,
		row.AntimicrobialResistance, nullIfBlank(row.TestSectionID), row.Active,
		row.Orderable, activeLocaleOf(row), ts, row.TestID).Error
}

// activeLocaleOf is the locale getName() resolves through — the deployment's
// default, which is the first active locale.
func activeLocaleOf(row TestModifyRow) string {
	if len(row.Locales) > 0 {
		return row.Locales[0]
	}
	return "en"
}

// syncLegacyLoinc is the update-path version: it deactivates any active LOINC
// mapping whose code no longer matches, reuses one that does, and otherwise
// inserts.
func (d *TestModifyDAO) syncLegacyLoinc(tx *gorm.DB, testID, loinc string, ts time.Time) error {
	code := strings.TrimSpace(loinc)
	if err := tx.Exec(`
		UPDATE clinlims.test_terminology_mapping
		   SET is_active = 'N', lastupdated = ?
		 WHERE test_id = ? AND source = 'LOINC' AND is_active = 'Y'
		   AND code IS DISTINCT FROM ?`, ts, testID, nullIfBlank(code)).Error; err != nil {
		return err
	}
	if code == "" {
		return nil
	}
	var existing []string
	if err := tx.Raw(`
		SELECT id FROM clinlims.test_terminology_mapping
		 WHERE test_id = ? AND source = 'LOINC' AND code = ? ORDER BY id LIMIT 1`,
		testID, code).Scan(&existing).Error; err != nil {
		return err
	}
	if len(existing) > 0 {
		return tx.Exec(`
			UPDATE clinlims.test_terminology_mapping
			   SET relationship = COALESCE(relationship, 'SAME_AS'), is_active = 'Y', lastupdated = ?
			 WHERE id = ?`, ts, existing[0]).Error
	}
	return tx.Exec(`
		INSERT INTO clinlims.test_terminology_mapping
		       (id, test_id, source, code, relationship, is_active, lastupdated)
		VALUES (gen_random_uuid()::text, ?, 'LOINC', ?, 'SAME_AS', 'Y', ?)`,
		testID, code, ts).Error
}

// addPanels re-attaches the test to the panels the form names.
//
// The sampletype_panel insert is the one write here with a condition: a sample
// type that already has ANY panel gets nothing, so the row appears only the
// first time a panel is attached to a test of that sample type — and it names
// whichever panel happened to be first.
func (d *TestModifyDAO) addPanels(tx *gorm.DB, row TestModifyRow, sampleTypeID string, ts time.Time) error {
	for _, panelID := range row.PanelIDs {
		if err := tx.Exec(`
			INSERT INTO clinlims.panel_item (id, panel_id, test_id, lastupdated)
			VALUES (nextval('clinlims.panel_item_seq'), ?, ?, ?)`,
			panelID, row.TestID, ts).Error; err != nil {
			return err
		}
		var existing []string
		if err := tx.Raw(
			`SELECT id FROM clinlims.sampletype_panel WHERE sample_type_id = ? LIMIT 1`,
			sampleTypeID).Scan(&existing).Error; err != nil {
			return err
		}
		if len(existing) == 0 {
			if err := tx.Exec(`
				INSERT INTO clinlims.sampletype_panel (id, sample_type_id, panel_id)
				VALUES (nextval('clinlims.sample_type_panel_seq'), ?, ?)`,
				sampleTypeID, panelID).Error; err != nil {
				return err
			}
		}
		if err := activateIfInactive(tx, "clinlims.panel", panelID, ts); err != nil {
			return err
		}
	}
	return nil
}

// addResults inserts the submitted results.
//
// It only ever INSERTS. The dictionary variants deactivate the old rows first,
// outside this transaction — the numeric and text ones do not, so every numeric
// modify leaves ANOTHER active result row on the test, and getResultType then
// reads the newest of a growing pile. Measured, and reproduced rather than
// fixed.
func (d *TestModifyDAO) addResults(tx *gorm.DB, row TestModifyRow, ts time.Time) error {
	switch {
	case isDictionaryVariant(row.ResultTypeChar):
		order := 10
		for _, dict := range row.Dictionaries {
			var resultID string
			if err := tx.Raw(`
				INSERT INTO clinlims.test_result
				       (id, test_id, tst_rslt_type, sort_order, is_active, value,
				        is_quantifiable, is_normal, significant_digits, lastupdated)
				VALUES (nextval('clinlims.test_result_seq'), ?, ?, ?, true, ?, ?, false, NULL, ?)
				RETURNING id::text`,
				row.TestID, row.ResultTypeChar, order, dict.DictionaryID,
				dict.IsQuantifiable, ts).Scan(&resultID).Error; err != nil {
				return err
			}
			if dict.IsDefault {
				if err := tx.Exec(
					`UPDATE clinlims.test SET default_test_result_id = ?, lastupdated = ? WHERE id = ?`,
					resultID, ts, row.TestID).Error; err != nil {
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
			row.TestID, row.ResultTypeChar, nullIfBlank(row.SignificantDigits), ts).Error
	}
	return nil
}

// addLimits inserts the submitted limits, each with its 'I' history row. The
// shapes are the create path's.
func (d *TestModifyDAO) addLimits(tx *gorm.DB, row TestModifyRow, sysUserID int64, ts time.Time) error {
	dao := &TestAddDAO{DB: d.DB, Audit: d.Audit}
	return dao.insertLimits(tx, row.TestID, row.TestAddRow, ts, sysUserID)
}

// syncComponent is the update half of syncPrimaryComponentFromLegacy: the
// component already exists, so it takes the newest active result's type and
// digits and the test's current unit — and keeps the LABEL it was created with,
// which is the test's name at CREATE time, not now.
func (d *TestModifyDAO) syncComponent(tx *gorm.DB, row TestModifyRow, ts time.Time) error {
	var newest struct {
		ResultType        *string `gorm:"column:tst_rslt_type"`
		SignificantDigits *int    `gorm:"column:significant_digits"`
	}
	if err := tx.Raw(`
		SELECT tst_rslt_type, significant_digits
		  FROM clinlims.test_result
		 WHERE test_id = ? AND is_active = true
		 ORDER BY id DESC LIMIT 1`, row.TestID).Scan(&newest).Error; err != nil {
		return err
	}

	var componentIDs []string
	if err := tx.Raw(`
		SELECT id FROM clinlims.test_result_component
		 WHERE test_id = ? AND is_active = 'Y'
		 ORDER BY (code = 'PRIMARY') DESC, id LIMIT 1`, row.TestID).Scan(&componentIDs).Error; err != nil {
		return err
	}

	var componentID string
	if len(componentIDs) == 0 {
		label := row.NameEnglish
		if strings.TrimSpace(label) == "" {
			label = "PRIMARY"
		}
		if err := tx.Raw(`
			INSERT INTO clinlims.test_result_component
			       (id, test_id, code, label, display_order, result_type, uom_id,
			        significant_digits, allow_multiple_readings, is_active, lastupdated)
			VALUES (gen_random_uuid()::text, ?, 'PRIMARY', ?, 0, ?, ?, ?, false, 'Y', ?)
			RETURNING id`,
			row.TestID, label, newest.ResultType, nullIfBlank(row.UomID),
			newest.SignificantDigits, ts).Scan(&componentID).Error; err != nil {
			return err
		}
	} else {
		componentID = componentIDs[0]
		if err := tx.Exec(`
			UPDATE clinlims.test_result_component
			   SET uom_id = ?, result_type = ?, significant_digits = ?, lastupdated = ?
			 WHERE id = ?`,
			nullIfBlank(row.UomID), newest.ResultType, newest.SignificantDigits,
			ts, componentID).Error; err != nil {
			return err
		}
	}

	if err := tx.Exec(
		`UPDATE clinlims.test_result SET component_id = ? WHERE test_id = ? AND component_id IS NULL`,
		componentID, row.TestID).Error; err != nil {
		return err
	}
	return tx.Exec(
		`UPDATE clinlims.result_limits SET component_id = ? WHERE test_id = ? AND component_id IS NULL`,
		componentID, row.TestID).Error
}

// javaDouble renders a double the way the audit payload carries it —
// Double.toString, so "0.0" and "Infinity" rather than "0" and "+Inf".
//
// util.JavaDoubleString is the JSON form and QUOTES the non-numeric values,
// because Jackson has to; the XML payload does not, so the quotes come off.
func javaDouble(v float64) string {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return strings.Trim(util.JavaDoubleString(v), `"`)
	}
	return util.JavaDoubleString(v)
}
