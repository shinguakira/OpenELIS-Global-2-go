package daoimpl

import (
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// The Sample & Results section: result COMPONENTS, their interpretation rules,
// and the select-list options behind a dictionary component.
//
// All three follow the same shape — an upsert keyed on the id the caller echoed
// back, then a soft-delete of everything active it did not mention. Nothing is
// ever hard-deleted, which is why a re-added component has to REACTIVATE its
// row: the (test_id, code) unique index is still holding the slot.

// ComponentRow is one test_result_component.
type ComponentRow struct {
	ID                    string  `gorm:"column:id"`
	Code                  string  `gorm:"column:code"`
	Label                 string  `gorm:"column:label"`
	DisplayOrder          int     `gorm:"column:display_order"`
	ResultType            *string `gorm:"column:result_type"`
	UomID                 *string `gorm:"column:uom_id"`
	SignificantDigits     *int    `gorm:"column:significant_digits"`
	DefaultResult         *string `gorm:"column:default_result"`
	AllowMultipleReadings bool    `gorm:"column:allow_multiple_readings"`
}

// InterpretationRow is one test_result_interpretation.
type InterpretationRow struct {
	ID           string  `gorm:"column:id"`
	ComponentID  string  `gorm:"column:component_id"`
	ValueMatch   *string `gorm:"column:value_match"`
	Text         *string `gorm:"column:interpretation_text"`
	Severity     *string `gorm:"column:severity"`
	Color        *string `gorm:"column:color"`
	DisplayOrder int     `gorm:"column:display_order"`
}

// OptionRow is one select-list option — a test_result row scoped to a component.
type OptionRow struct {
	ID          string  `gorm:"column:id"`
	ComponentID string  `gorm:"column:component_id"`
	Value       *string `gorm:"column:value"`
	ValueName   *string `gorm:"column:value_name"`
	ResultType  *string `gorm:"column:result_type"`
	SortOrder   *string `gorm:"column:sort_order"`
	Normal      *bool   `gorm:"column:normal"`
}

// ActiveComponents ports getActiveComponentsByTestId — ordered by displayOrder
// then code, which is the only ordered read in this section.
func (d *EditorTestDAO) ActiveComponents(testID string) ([]ComponentRow, error) {
	rows := []ComponentRow{}
	err := d.DB.Raw(`
		SELECT id, code, label, COALESCE(display_order, 0) AS display_order, result_type,
		       uom_id::text AS uom_id, significant_digits::int AS significant_digits,
		       default_result, COALESCE(allow_multiple_readings, false) AS allow_multiple_readings
		  FROM clinlims.test_result_component
		 WHERE test_id = ? AND is_active = 'Y'
		 ORDER BY display_order, code`, testID).Scan(&rows).Error
	return rows, err
}

// ActiveInterpretations ports getActiveByComponentId — ordered by displayOrder
// alone, so two rules sharing an order come back in the plan's order.
func (d *EditorTestDAO) ActiveInterpretations(componentID string) ([]InterpretationRow, error) {
	rows := []InterpretationRow{}
	err := d.DB.Raw(`
		SELECT id, component_id, value_match, interpretation_text, severity, color,
		       COALESCE(display_order, 0) AS display_order
		  FROM clinlims.test_result_interpretation
		 WHERE component_id = ? AND is_active = 'Y'
		 ORDER BY display_order`, componentID).Scan(&rows).Error
	return rows, err
}

// ActiveOptions ports getActiveOptionsByComponentId.
//
// Two filters that are not in the SQL in Java and are here for the same reason
// they are there: a row whose result type is NOT a dictionary variant is
// REMOVED after the query, and the sort is on the numeric value of a column
// mapped as a String, nulls last. `sort_order` is numeric in the database, so
// the cast is the port's own — Java parses the string.
func (d *EditorTestDAO) ActiveOptions(componentID, locale string) ([]OptionRow, error) {
	rows := []OptionRow{}
	err := d.DB.Raw(`
		SELECT tr.id::text AS id, tr.component_id AS component_id, tr.value AS value,
		       COALESCE(NULLIF(lv.value, ''), dict.dict_entry) AS value_name,
		       tr.tst_rslt_type AS result_type, tr.sort_order::text AS sort_order,
		       tr.is_normal AS normal
		  FROM clinlims.test_result tr
		  LEFT JOIN clinlims.dictionary dict ON dict.id::text = tr.value
		  LEFT JOIN clinlims.localization_value lv
		         ON lv.localization_id = dict.name_localization_id AND lv.locale = ?
		 WHERE tr.component_id = ? AND tr.is_active = true
		   AND tr.tst_rslt_type IN ('D', 'M', 'C')
		 ORDER BY COALESCE(tr.sort_order, 2147483647), tr.id`, locale, componentID).Scan(&rows).Error
	return rows, err
}

// DesiredComponent is one component the caller asked for, with its children.
type DesiredComponent struct {
	ID                    string
	Code                  string
	Label                 string
	DisplayOrder          int
	ResultType            *string
	UomID                 *string
	SignificantDigits     *int
	DefaultResult         *string
	AllowMultipleReadings bool
	Interpretations       []DesiredInterpretation
	Options               []DesiredOption
}

// DesiredInterpretation is one interpretation rule the caller asked for.
type DesiredInterpretation struct {
	ID           string
	ValueMatch   *string
	Text         *string
	Severity     *string
	Color        *string
	DisplayOrder int
}

// DesiredOption is one select-list option the caller asked for.
type DesiredOption struct {
	ID         string
	Value      *string
	SortOrder  *int
	Normal     bool
	ResultType *string
}

// SaveSampleResults ports saveSampleResults — components, then the children of
// each, then the legacy mirror.
func (d *EditorTestDAO) SaveSampleResults(testID string, desired []DesiredComponent, sysUserID int64) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		ts := time.Now().UTC().Truncate(time.Millisecond)
		if err := saveComponents(tx, testID, desired, ts); err != nil {
			return err
		}

		// The children are keyed by the component's CODE, not its id, because a
		// newly inserted component had no id when the request was built.
		codeToID := map[string]string{}
		components, err := activeComponentsIn(tx, testID)
		if err != nil {
			return err
		}
		for _, c := range components {
			codeToID[c.Code] = c.ID
		}

		for _, want := range desired {
			componentID, ok := codeToID[want.Code]
			if !ok {
				continue
			}
			if err := saveInterpretations(tx, componentID, want.Interpretations, ts); err != nil {
				return err
			}
			if err := saveOptions(tx, testID, componentID, want.Options, ts); err != nil {
				return err
			}
		}
		return syncLegacyTestFields(tx, testID, ts)
	})
}

func saveComponents(tx *gorm.DB, testID string, desired []DesiredComponent, ts time.Time) error {
	existing, err := activeComponentsIn(tx, testID)
	if err != nil {
		return err
	}
	existingByID := map[string]bool{}
	for _, e := range existing {
		existingByID[e.ID] = true
	}

	kept := map[string]bool{}
	for _, want := range desired {
		if want.ID != "" && existingByID[want.ID] {
			if err := tx.Exec(`
				UPDATE clinlims.test_result_component
				   SET code = ?, label = ?, display_order = ?, result_type = ?, uom_id = ?,
				       significant_digits = ?, default_result = ?, allow_multiple_readings = ?,
				       lastupdated = ?
				 WHERE id = ?`,
				want.Code, want.Label, want.DisplayOrder, want.ResultType, want.UomID,
				want.SignificantDigits, want.DefaultResult, want.AllowMultipleReadings,
				ts, want.ID).Error; err != nil {
				return err
			}
			kept[want.ID] = true
			continue
		}

		// A soft-deleted component still holds the (test_id, code) slot, so a
		// re-added code REACTIVATES that row. Inserting instead would violate
		// the unique index.
		var dead []string
		if err := tx.Raw(`
			SELECT id FROM clinlims.test_result_component
			 WHERE test_id = ? AND code = ? AND is_active <> 'Y' LIMIT 1`,
			testID, want.Code).Scan(&dead).Error; err != nil {
			return err
		}
		if len(dead) > 0 {
			if err := tx.Exec(`
				UPDATE clinlims.test_result_component
				   SET label = ?, display_order = ?, result_type = ?, uom_id = ?,
				       significant_digits = ?, default_result = ?, allow_multiple_readings = ?,
				       is_active = 'Y', lastupdated = ?
				 WHERE id = ?`,
				want.Label, want.DisplayOrder, want.ResultType, want.UomID,
				want.SignificantDigits, want.DefaultResult, want.AllowMultipleReadings,
				ts, dead[0]).Error; err != nil {
				return err
			}
			kept[dead[0]] = true
			continue
		}

		if err := tx.Exec(`
			INSERT INTO clinlims.test_result_component
			       (id, test_id, code, label, display_order, result_type, uom_id,
			        significant_digits, default_result, allow_multiple_readings, is_active, lastupdated)
			VALUES (gen_random_uuid()::text, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'Y', ?)`,
			testID, want.Code, want.Label, want.DisplayOrder, want.ResultType, want.UomID,
			want.SignificantDigits, want.DefaultResult, want.AllowMultipleReadings, ts).Error; err != nil {
			return err
		}
	}

	for _, e := range existing {
		if kept[e.ID] {
			continue
		}
		if err := tx.Exec(`
			UPDATE clinlims.test_result_component SET is_active = 'N', lastupdated = ?
			 WHERE id = ?`, ts, e.ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func saveInterpretations(tx *gorm.DB, componentID string, desired []DesiredInterpretation, ts time.Time) error {
	existing := []InterpretationRow{}
	if err := tx.Raw(`
		SELECT id, component_id, value_match, interpretation_text, severity, color,
		       COALESCE(display_order, 0) AS display_order
		  FROM clinlims.test_result_interpretation
		 WHERE component_id = ? AND is_active = 'Y' ORDER BY display_order`, componentID).
		Scan(&existing).Error; err != nil {
		return err
	}
	existingByID := map[string]bool{}
	for _, e := range existing {
		existingByID[e.ID] = true
	}

	kept := map[string]bool{}
	for _, want := range desired {
		if want.ID != "" && existingByID[want.ID] {
			if err := tx.Exec(`
				UPDATE clinlims.test_result_interpretation
				   SET value_match = ?, interpretation_text = ?, severity = ?, color = ?,
				       display_order = ?, lastupdated = ?
				 WHERE id = ?`,
				want.ValueMatch, want.Text, want.Severity, want.Color,
				want.DisplayOrder, ts, want.ID).Error; err != nil {
				return err
			}
			kept[want.ID] = true
			continue
		}
		// Unlike components, an interpretation has no unique slot to reclaim —
		// a fresh row every time.
		if err := tx.Exec(`
			INSERT INTO clinlims.test_result_interpretation
			       (id, component_id, value_match, interpretation_text, severity, color,
			        display_order, is_active, lastupdated)
			VALUES (gen_random_uuid()::text, ?, ?, ?, ?, ?, ?, 'Y', ?)`,
			componentID, want.ValueMatch, want.Text, want.Severity, want.Color,
			want.DisplayOrder, ts).Error; err != nil {
			return err
		}
	}

	for _, e := range existing {
		if kept[e.ID] {
			continue
		}
		if err := tx.Exec(`
			UPDATE clinlims.test_result_interpretation SET is_active = 'N', lastupdated = ?
			 WHERE id = ?`, ts, e.ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func saveOptions(tx *gorm.DB, testID, componentID string, desired []DesiredOption, ts time.Time) error {
	// The "existing" set is what getActiveOptionsByComponentId returns — and
	// that read DROPS every non-dictionary row, so an option whose result type
	// is numeric is invisible to the reconcile and is never deactivated.
	existing := []string{}
	if err := tx.Raw(`
		SELECT id::text FROM clinlims.test_result
		 WHERE component_id = ? AND is_active = true
		   AND tst_rslt_type IN ('D', 'M', 'C')
		 ORDER BY COALESCE(sort_order, 2147483647), id`, componentID).Scan(&existing).Error; err != nil {
		return err
	}
	existingByID := map[string]bool{}
	for _, id := range existing {
		existingByID[id] = true
	}

	kept := map[string]bool{}
	for _, want := range desired {
		if want.ID != "" && existingByID[want.ID] {
			if err := tx.Exec(`
				UPDATE clinlims.test_result
				   SET value = ?, sort_order = ?, is_normal = ?, tst_rslt_type = ?, lastupdated = ?
				 WHERE id = ?`,
				want.Value, want.SortOrder, want.Normal, want.ResultType, ts, want.ID).Error; err != nil {
				return err
			}
			kept[want.ID] = true
			continue
		}
		if err := tx.Exec(`
			INSERT INTO clinlims.test_result
			       (id, test_id, component_id, value, sort_order, is_normal, tst_rslt_type,
			        is_active, is_quantifiable, significant_digits, lastupdated)
			VALUES (nextval('clinlims.test_result_seq'), ?, ?, ?, ?, ?, ?, true, false, NULL, ?)`,
			testID, componentID, want.Value, want.SortOrder, want.Normal,
			want.ResultType, ts).Error; err != nil {
			return err
		}
	}

	for _, id := range existing {
		if kept[id] {
			continue
		}
		if err := tx.Exec(
			`UPDATE clinlims.test_result SET is_active = false, lastupdated = ? WHERE id = ?`,
			ts, id).Error; err != nil {
			return err
		}
	}
	return nil
}

// syncLegacyTestFields mirrors the PRIMARY component back onto the columns the
// OLD Test Modify screen still reads — test.uom_id and
// test_result.significant_digits.
//
// The M1 backfill seeded the component FROM those columns; this is the inverse,
// and it is what keeps the two editors agreeing during the transition. When the
// test has no legacy test_result row at all — a test born in the new editor — a
// single row is SEEDED for the non-dictionary types, because getResultType()
// would otherwise fall back to ALPHA and the legacy screen would show the wrong
// type.
func syncLegacyTestFields(tx *gorm.DB, testID string, ts time.Time) error {
	components, err := activeComponentsIn(tx, testID)
	if err != nil {
		return err
	}
	if len(components) == 0 {
		return nil
	}
	primary := components[0]
	for _, c := range components {
		if c.Code == "PRIMARY" {
			primary = c
			break
		}
	}

	if err := tx.Exec(
		`UPDATE clinlims.test SET uom_id = ?, lastupdated = ? WHERE id = ?`,
		primary.UomID, ts, testID).Error; err != nil {
		return err
	}

	var digits any
	if primary.SignificantDigits != nil {
		digits = strconv.Itoa(*primary.SignificantDigits)
	}

	legacy := []struct {
		ID    string  `gorm:"column:id"`
		Value *string `gorm:"column:value"`
	}{}
	if err := tx.Raw(`
		SELECT id::text AS id, value FROM clinlims.test_result
		 WHERE test_id = ? AND is_active = true
		 ORDER BY COALESCE(result_group, 0) DESC, id DESC`, testID).Scan(&legacy).Error; err != nil {
		return err
	}

	dictionary := primary.ResultType != nil && strings.Contains("DMC", *primary.ResultType)

	if len(legacy) == 0 {
		if primary.ResultType == nil || dictionary {
			// A dictionary type gets its rows from the options instead.
			return nil
		}
		return tx.Exec(`
			INSERT INTO clinlims.test_result
			       (id, test_id, tst_rslt_type, sort_order, is_active, significant_digits,
			        is_quantifiable, is_normal, lastupdated)
			VALUES (nextval('clinlims.test_result_seq'), ?, ?, 1, true, ?, false, false, ?)`,
			testID, *primary.ResultType, digits, ts).Error
	}

	for _, row := range legacy {
		hasValue := row.Value != nil && strings.TrimSpace(*row.Value) != ""
		if dictionary && !hasValue {
			// A dictionary component cannot carry a value-less legacy row, so
			// the one the numeric flow left behind is switched off.
			if err := tx.Exec(
				`UPDATE clinlims.test_result SET is_active = false, lastupdated = ? WHERE id = ?`,
				ts, row.ID).Error; err != nil {
				return err
			}
			continue
		}
		sets := "significant_digits = ?"
		args := []any{digits}
		if primary.ResultType != nil {
			sets += ", tst_rslt_type = ?"
			args = append(args, *primary.ResultType)
		}
		args = append(args, ts, row.ID)
		if err := tx.Exec(
			`UPDATE clinlims.test_result SET `+sets+`, lastupdated = ? WHERE id = ?`, args...).Error; err != nil {
			return err
		}
	}
	return nil
}

// CopyComponentsFromTest ports copyComponentsFromTest.
//
// A component whose CODE the target already has is skipped whole — including
// its interpretations and options — so a partial copy is possible and a repeat
// copy is a no-op. Note what it does NOT do: no syncLegacyTestFields, so the
// legacy columns are left as they were.
func (d *EditorTestDAO) CopyComponentsFromTest(sourceTestID, targetTestID string, locale string, sysUserID int64) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		ts := time.Now().UTC().Truncate(time.Millisecond)
		sources, err := activeComponentsIn(tx, sourceTestID)
		if err != nil {
			return err
		}
		for _, src := range sources {
			var taken []string
			if err := tx.Raw(
				`SELECT id FROM clinlims.test_result_component WHERE test_id = ? AND code = ? LIMIT 1`,
				targetTestID, src.Code).Scan(&taken).Error; err != nil {
				return err
			}
			if len(taken) > 0 {
				continue
			}

			var copyID string
			if err := tx.Raw(`
				INSERT INTO clinlims.test_result_component
				       (id, test_id, code, label, display_order, result_type, uom_id,
				        significant_digits, default_result, allow_multiple_readings, is_active, lastupdated)
				VALUES (gen_random_uuid()::text, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'Y', ?)
				RETURNING id`,
				targetTestID, src.Code, src.Label, src.DisplayOrder, src.ResultType, src.UomID,
				src.SignificantDigits, src.DefaultResult, src.AllowMultipleReadings, ts).
				Scan(&copyID).Error; err != nil {
				return err
			}

			interps, err := activeInterpretationsIn(tx, src.ID)
			if err != nil {
				return err
			}
			desiredInterps := make([]DesiredInterpretation, 0, len(interps))
			for _, i := range interps {
				// The copies carry NO id, so every one is inserted fresh.
				desiredInterps = append(desiredInterps, DesiredInterpretation{
					ValueMatch: i.ValueMatch, Text: i.Text, Severity: i.Severity,
					Color: i.Color, DisplayOrder: i.DisplayOrder,
				})
			}
			if err := saveInterpretations(tx, copyID, desiredInterps, ts); err != nil {
				return err
			}

			options, err := activeOptionsIn(tx, src.ID, locale)
			if err != nil {
				return err
			}
			desiredOptions := make([]DesiredOption, 0, len(options))
			for _, o := range options {
				normal := o.Normal != nil && *o.Normal
				desiredOptions = append(desiredOptions, DesiredOption{
					Value: o.Value, SortOrder: parseSortOrder(o.SortOrder),
					Normal: normal, ResultType: o.ResultType,
				})
			}
			if err := saveOptions(tx, targetTestID, copyID, desiredOptions, ts); err != nil {
				return err
			}
		}
		return nil
	})
}

func activeComponentsIn(tx *gorm.DB, testID string) ([]ComponentRow, error) {
	d := &EditorTestDAO{DB: tx}
	return d.ActiveComponents(testID)
}

func activeInterpretationsIn(tx *gorm.DB, componentID string) ([]InterpretationRow, error) {
	d := &EditorTestDAO{DB: tx}
	return d.ActiveInterpretations(componentID)
}

func activeOptionsIn(tx *gorm.DB, componentID, locale string) ([]OptionRow, error) {
	d := &EditorTestDAO{DB: tx}
	return d.ActiveOptions(componentID, locale)
}

func parseSortOrder(s *string) *int {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	n, err := strconv.Atoi(strings.TrimSpace(*s))
	if err != nil {
		return nil
	}
	return &n
}
