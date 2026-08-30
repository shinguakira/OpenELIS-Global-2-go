// Package daoimpl holds the DB access for the test-catalog editor's writes.
//
// The read side of this package (lab-units, sample-types, panels) composes
// foreign domain services and needs no DAO of its own. The writes below do:
// they move rows in tables no other ported service owns — test_sample_handling
// and its history, test_terminology_mapping, and the two join tables — and each
// one is a multi-statement upsert that has to be one transaction.
package daoimpl

import (
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
)

// EditorWriteDAO backs the editor's section saves.
//
// NOTHING here is audited. clinlims.history stays empty across every write in
// this file — measured, and worth stating because three of the tables involved
// (PANEL, PANEL_ITEM, SAMPLETYPE_TEST) carry keep_history='Y'. The services
// behind them never set auditTrailLog, so the flag has no effect, exactly as in
// the testconfiguration screens.
//
// Storage has an audit of its OWN, and it is a different mechanism: a JSON
// snapshot row in test_sample_handling_history, written only when the business
// state actually changed.
type EditorWriteDAO struct {
	DB *gorm.DB
}

// ---------------------------------------------------------------- storage

// StorageRow is one test_sample_handling row, in the field order snapshot()
// serialises.
type StorageRow struct {
	StorageCondition       *string `json:"storageCondition"`
	StorageConditionCustom *string `json:"storageConditionCustom"`
	StorageDuration        *int    `json:"storageDuration"`
	StorageDurationUnit    *string `json:"storageDurationUnit"`
	StabilityNotes         *string `json:"stabilityNotes"`
	ProtectFromLight       bool    `json:"protectFromLight"`
	DoNotFreeze            bool    `json:"doNotFreeze"`
	DoNotRefrigerate       bool    `json:"doNotRefrigerate"`
	DisposalMethod         *string `json:"disposalMethod"`
	DisposalTimeframe      *int    `json:"disposalTimeframe"`
	DisposalUnit           *string `json:"disposalUnit"`
	SpecialInstructions    *string `json:"specialInstructions"`
	OverrideRestricted     bool    `json:"overrideRestricted"`
}

// StoredStorage is what a read gives back: the row, plus the columns the
// snapshot does not carry.
type StoredStorage struct {
	ID      string
	Row     StorageRow
	Version int
	Found   bool
}

// GetStorage reads the single handling row for a test.
//
// getByTestId takes getAllMatching("testId", …).get(0) — so a test with two
// rows silently uses one of them, and which one is the plan's business. There
// is no unique constraint stopping that.
func (d *EditorWriteDAO) GetStorage(testID string) (StoredStorage, error) {
	var rows []struct {
		ID                     string  `gorm:"column:id"`
		StorageCondition       *string `gorm:"column:storage_condition"`
		StorageConditionCustom *string `gorm:"column:storage_condition_custom"`
		StorageDuration        *int    `gorm:"column:storage_duration"`
		StorageDurationUnit    *string `gorm:"column:storage_duration_unit"`
		StabilityNotes         *string `gorm:"column:stability_notes"`
		ProtectFromLight       bool    `gorm:"column:protect_from_light"`
		DoNotFreeze            bool    `gorm:"column:do_not_freeze"`
		DoNotRefrigerate       bool    `gorm:"column:do_not_refrigerate"`
		DisposalMethod         *string `gorm:"column:disposal_method"`
		DisposalTimeframe      *int    `gorm:"column:disposal_timeframe"`
		DisposalUnit           *string `gorm:"column:disposal_unit"`
		SpecialInstructions    *string `gorm:"column:special_instructions"`
		OverrideRestricted     bool    `gorm:"column:override_restricted"`
		Version                *int    `gorm:"column:version"`
	}
	err := d.DB.Raw(`
		SELECT id, storage_condition, storage_condition_custom, storage_duration,
		       storage_duration_unit, stability_notes,
		       COALESCE(protect_from_light, false) AS protect_from_light,
		       COALESCE(do_not_freeze, false) AS do_not_freeze,
		       COALESCE(do_not_refrigerate, false) AS do_not_refrigerate,
		       disposal_method, disposal_timeframe, disposal_unit, special_instructions,
		       COALESCE(override_restricted, false) AS override_restricted, version
		  FROM clinlims.test_sample_handling WHERE test_id = ? ORDER BY id`, testID).
		Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return StoredStorage{}, err
	}
	r := rows[0]
	out := StoredStorage{ID: r.ID, Found: true}
	if r.Version != nil {
		out.Version = *r.Version
	}
	out.Row = StorageRow{
		StorageCondition: r.StorageCondition, StorageConditionCustom: r.StorageConditionCustom,
		StorageDuration: r.StorageDuration, StorageDurationUnit: r.StorageDurationUnit,
		StabilityNotes: r.StabilityNotes, ProtectFromLight: r.ProtectFromLight,
		DoNotFreeze: r.DoNotFreeze, DoNotRefrigerate: r.DoNotRefrigerate,
		DisposalMethod: r.DisposalMethod, DisposalTimeframe: r.DisposalTimeframe,
		DisposalUnit: r.DisposalUnit, SpecialInstructions: r.SpecialInstructions,
		OverrideRestricted: r.OverrideRestricted,
	}
	return out, nil
}

// SaveStorage upserts the handling row and records a snapshot when it changed.
//
// Two things are measured and neither is obvious. `version` is bumped on EVERY
// save, including one that changes nothing — it is a save counter, not a state
// counter. The history row is written only when the snapshot JSON differs, so a
// no-op re-save leaves version 2 with a single history row behind it.
//
// The save is a REPLACE, not a merge: every column is written from the request,
// so a field the caller omitted comes back null even if it had a value. That is
// what makes group/storage able to blank a duration it never mentions.
func (d *EditorWriteDAO) SaveStorage(testID string, desired StorageRow, sysUserID int64) (StoredStorage, error) {
	var out StoredStorage
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		existing, err := storageIn(tx, testID)
		if err != nil {
			return err
		}
		newJSON, err := json.Marshal(desired)
		if err != nil {
			return err
		}
		ts := time.Now().UTC().Truncate(time.Millisecond)
		version := existing.Version + 1

		var id string
		changeType := "INSERT"
		var previousJSON *string
		if existing.Found {
			changeType = "UPDATE"
			prior, err := json.Marshal(existing.Row)
			if err != nil {
				return err
			}
			s := string(prior)
			previousJSON = &s
			id = existing.ID
			if err := tx.Exec(`
				UPDATE clinlims.test_sample_handling
				   SET storage_condition = ?, storage_condition_custom = ?, storage_duration = ?,
				       storage_duration_unit = ?, stability_notes = ?, protect_from_light = ?,
				       do_not_freeze = ?, do_not_refrigerate = ?, disposal_method = ?,
				       disposal_timeframe = ?, disposal_unit = ?, special_instructions = ?,
				       override_restricted = ?, version = ?, is_active = 'Y', lastupdated = ?
				 WHERE id = ?`,
				desired.StorageCondition, desired.StorageConditionCustom, desired.StorageDuration,
				desired.StorageDurationUnit, desired.StabilityNotes, desired.ProtectFromLight,
				desired.DoNotFreeze, desired.DoNotRefrigerate, desired.DisposalMethod,
				desired.DisposalTimeframe, desired.DisposalUnit, desired.SpecialInstructions,
				desired.OverrideRestricted, version, ts, id).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Raw(`
				INSERT INTO clinlims.test_sample_handling
				       (id, test_id, storage_condition, storage_condition_custom, storage_duration,
				        storage_duration_unit, stability_notes, protect_from_light, do_not_freeze,
				        do_not_refrigerate, disposal_method, disposal_timeframe, disposal_unit,
				        special_instructions, override_restricted, version, is_active, lastupdated)
				VALUES (gen_random_uuid()::text, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'Y', ?)
				RETURNING id`,
				testID, desired.StorageCondition, desired.StorageConditionCustom, desired.StorageDuration,
				desired.StorageDurationUnit, desired.StabilityNotes, desired.ProtectFromLight,
				desired.DoNotFreeze, desired.DoNotRefrigerate, desired.DisposalMethod,
				desired.DisposalTimeframe, desired.DisposalUnit, desired.SpecialInstructions,
				desired.OverrideRestricted, version, ts).Scan(&id).Error; err != nil {
				return err
			}
		}

		// The snapshot comparison is on the JSON, not the columns — so it sees
		// exactly what the history row would store.
		changed := previousJSON == nil || *previousJSON != string(newJSON)
		if changed {
			if err := tx.Exec(`
				INSERT INTO clinlims.test_sample_handling_history
				       (id, test_sample_handling_id, changed_by, changed_at, change_type,
				        previous_values, new_values, lastupdated)
				VALUES (gen_random_uuid()::text, ?, ?, ?, ?, ?::jsonb, ?::jsonb, ?)`,
				id, sysUserID, ts, changeType, previousJSON, string(newJSON), ts).Error; err != nil {
				return err
			}
		}

		out = StoredStorage{ID: id, Row: desired, Version: version, Found: true}
		return nil
	})
	return out, err
}

func storageIn(tx *gorm.DB, testID string) (StoredStorage, error) {
	d := &EditorWriteDAO{DB: tx}
	return d.GetStorage(testID)
}

// ------------------------------------------------------------ terminology

// TerminologyRow is one test_terminology_mapping.
type TerminologyRow struct {
	ID           string  `gorm:"column:id"`
	Source       string  `gorm:"column:source"`
	Code         string  `gorm:"column:code"`
	Relationship *string `gorm:"column:relationship"`
	IsActive     string  `gorm:"column:is_active"`
}

// ActiveTerminology reads the active mappings.
//
// NO ORDER BY, deliberately. getAllMatching is a criteria query with no
// ordering, so Java gets the heap order — which on rows this suite has just
// inserted is INSERTION order. Adding `ORDER BY id` looks tidier and is wrong:
// these ids are UUIDs, so it would sort them at random relative to the sequence
// the caller sent them in, and both the response and pickLegacyLoinc's
// "first LOINC mapping" read that order.
func (d *EditorWriteDAO) ActiveTerminology(testID string) ([]TerminologyRow, error) {
	rows := []TerminologyRow{}
	err := d.DB.Raw(`
		SELECT id, source, code, relationship, is_active
		  FROM clinlims.test_terminology_mapping
		 WHERE test_id = ? AND is_active = 'Y'`, testID).Scan(&rows).Error
	return rows, err
}

// DesiredMapping is one mapping the caller asked for.
type DesiredMapping struct {
	Source       string
	Code         string
	Relationship *string
}

// SaveTerminology ports saveMappingsForTest.
//
// The natural key is (source, code) and the DB enforces it per test, so a
// mapping that was soft-deleted and is asked for again is REACTIVATED rather
// than inserted — an insert would collide. Anything active and no longer asked
// for goes to is_active = 'N'; nothing is ever deleted.
//
// Then the legacy `test.loinc` column is re-derived from whatever survived:
// the first active LOINC mapping with relationship SAME_AS, else the first
// active LOINC mapping, else NULL. Dropping every LOINC mapping therefore
// CLEARS the column the rest of the application reads.
func (d *EditorWriteDAO) SaveTerminology(testID string, desired []DesiredMapping, sysUserID int64) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		ts := time.Now().UTC().Truncate(time.Millisecond)

		existing := []TerminologyRow{}
		if err := tx.Raw(`
			SELECT id, source, code, relationship, is_active
			  FROM clinlims.test_terminology_mapping WHERE test_id = ?`, testID).
			Scan(&existing).Error; err != nil {
			return err
		}
		byKey := map[string]TerminologyRow{}
		for _, m := range existing {
			byKey[m.Source+" "+m.Code] = m
		}

		desiredKeys := map[string]bool{}
		for _, want := range desired {
			key := want.Source + " " + want.Code
			desiredKeys[key] = true
			if target, ok := byKey[key]; ok {
				if err := tx.Exec(`
					UPDATE clinlims.test_terminology_mapping
					   SET relationship = ?, is_active = 'Y', lastupdated = ? WHERE id = ?`,
					want.Relationship, ts, target.ID).Error; err != nil {
					return err
				}
				continue
			}
			if err := tx.Exec(`
				INSERT INTO clinlims.test_terminology_mapping
				       (id, test_id, source, code, relationship, is_active, lastupdated)
				VALUES (gen_random_uuid()::text, ?, ?, ?, ?, 'Y', ?)`,
				testID, want.Source, want.Code, want.Relationship, ts).Error; err != nil {
				return err
			}
		}

		for _, m := range existing {
			if m.IsActive != "Y" || desiredKeys[m.Source+" "+m.Code] {
				continue
			}
			if err := tx.Exec(`
				UPDATE clinlims.test_terminology_mapping
				   SET is_active = 'N', lastupdated = ? WHERE id = ?`, ts, m.ID).Error; err != nil {
				return err
			}
		}

		return applyLoincToLegacyTest(tx, testID, ts)
	})
}

// applyLoincToLegacyTest re-derives test.loinc, and writes only when it moved.
func applyLoincToLegacyTest(tx *gorm.DB, testID string, ts time.Time) error {
	active := []TerminologyRow{}
	if err := tx.Raw(`
		SELECT id, source, code, relationship, is_active
		  FROM clinlims.test_terminology_mapping
		 WHERE test_id = ? AND is_active = 'Y'`, testID).Scan(&active).Error; err != nil {
		return err
	}
	var loinc *string
	for _, m := range active {
		if m.Source != "LOINC" {
			continue
		}
		if m.Relationship != nil && *m.Relationship == "SAME_AS" {
			code := m.Code
			loinc = &code
			break
		}
		if loinc == nil {
			code := m.Code
			loinc = &code
		}
	}
	// `Objects.equals(test.getLoinc(), loinc)` short-circuits an unchanged
	// value, so no UPDATE and no lastupdated bump.
	return tx.Exec(`
		UPDATE clinlims.test SET loinc = ?, lastupdated = ?
		 WHERE id = ? AND loinc IS DISTINCT FROM ?`, loinc, ts, testID, loinc).Error
}

// ------------------------------------------------------- sample-type order

// TestOrderRow is one test under a sample type, for the display-order screen.
type TestOrderRow struct {
	TestID       string  `gorm:"column:test_id"`
	TestName     *string `gorm:"column:test_name"`
	DisplayOrder *int    `gorm:"column:display_order"`
}

// TestOrder reads the junction rows for a sample type. Java sorts them in
// memory, so the query only has to return them.
func (d *EditorWriteDAO) TestOrder(sampleTypeID string) ([]TestOrderRow, error) {
	rows := []TestOrderRow{}
	err := d.DB.Raw(`
		SELECT st.test_id::text AS test_id, t.name AS test_name, st.display_order
		  FROM clinlims.sampletype_test st
		  LEFT JOIN clinlims.test t ON t.id = st.test_id
		 WHERE st.sample_type_id = ? ORDER BY st.id`, sampleTypeID).Scan(&rows).Error
	return rows, err
}

// SaveTestOrder ports updateDisplayOrder.
//
// It walks the junction rows the sample type already has and writes an order to
// the ones the request names. A test id that is not under this sample type is
// not inserted — it is simply not found and nothing happens.
func (d *EditorWriteDAO) SaveTestOrder(sampleTypeID string, orderByTestID map[string]int) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		ts := time.Now().UTC().Truncate(time.Millisecond)
		rows := []struct {
			ID     string `gorm:"column:id"`
			TestID string `gorm:"column:test_id"`
		}{}
		if err := tx.Raw(
			`SELECT id::text AS id, test_id::text AS test_id
			   FROM clinlims.sampletype_test WHERE sample_type_id = ? ORDER BY id`,
			sampleTypeID).Scan(&rows).Error; err != nil {
			return err
		}
		for _, row := range rows {
			order, ok := orderByTestID[row.TestID]
			if !ok {
				continue
			}
			if err := tx.Exec(
				`UPDATE clinlims.sampletype_test SET display_order = ? WHERE id = ?`,
				order, row.ID).Error; err != nil {
				return err
			}
		}
		_ = ts
		return nil
	})
}

// ------------------------------------------------------------------ panels

// PanelMembershipRow is one panel a test belongs to.
type PanelMembershipRow struct {
	PanelID   string  `gorm:"column:panel_id"`
	PanelName *string `gorm:"column:panel_name"`
	SortOrder *string `gorm:"column:sort_order"`
}

// TestPanels reads the panel_item rows for a test, in id order — Java sorts by
// panel name afterwards.
func (d *EditorWriteDAO) TestPanels(testID string) ([]PanelMembershipRow, error) {
	rows := []PanelMembershipRow{}
	err := d.DB.Raw(`
		SELECT pi.panel_id::text AS panel_id, p.name AS panel_name, pi.sort_order::text AS sort_order
		  FROM clinlims.panel_item pi
		  LEFT JOIN clinlims.panel p ON p.id = pi.panel_id
		 WHERE pi.test_id = ? ORDER BY pi.id`, testID).Scan(&rows).Error
	return rows, err
}

// PanelTests reads a panel's members, for the read-only position preview.
func (d *EditorWriteDAO) PanelTests(panelID string) ([]TestOrderRow, error) {
	rows := []TestOrderRow{}
	err := d.DB.Raw(`
		SELECT pi.test_id::text AS test_id, t.name AS test_name,
		       pi.sort_order::int AS display_order
		  FROM clinlims.panel_item pi
		  LEFT JOIN clinlims.test t ON t.id = pi.test_id
		 WHERE pi.panel_id = ? ORDER BY pi.id`, panelID).Scan(&rows).Error
	return rows, err
}

// SaveTestPanels ports setMembershipsForTest.
//
// An upsert over the panels named, then a DELETE of every membership not named
// — panel_item has no soft-delete flag, so a dropped membership is gone.
func (d *EditorWriteDAO) SaveTestPanels(testID string, positionByPanelID map[string]*int) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		ts := time.Now().UTC().Truncate(time.Millisecond)
		existing := []struct {
			ID      string `gorm:"column:id"`
			PanelID string `gorm:"column:panel_id"`
		}{}
		if err := tx.Raw(
			`SELECT id::text AS id, panel_id::text AS panel_id
			   FROM clinlims.panel_item WHERE test_id = ? ORDER BY id`, testID).
			Scan(&existing).Error; err != nil {
			return err
		}
		byPanel := map[string]string{}
		for _, e := range existing {
			byPanel[e.PanelID] = e.ID
		}

		for panelID, position := range positionByPanelID {
			if itemID, ok := byPanel[panelID]; ok {
				if err := tx.Exec(
					`UPDATE clinlims.panel_item SET sort_order = ?, lastupdated = ? WHERE id = ?`,
					position, ts, itemID).Error; err != nil {
					return err
				}
				continue
			}
			// getPanelById returning null skips the membership silently. The
			// controller rejects an unknown panel with 422 before this runs, so
			// the branch is unreachable from the endpoint — kept because the
			// service is shared.
			var found []string
			if err := tx.Raw(`SELECT id FROM clinlims.panel WHERE id = ?`, panelID).
				Scan(&found).Error; err != nil {
				return err
			}
			if len(found) == 0 {
				continue
			}
			if err := tx.Exec(`
				INSERT INTO clinlims.panel_item (id, panel_id, test_id, sort_order, lastupdated)
				VALUES (nextval('clinlims.panel_item_seq'), ?, ?, ?, ?)`,
				panelID, testID, position, ts).Error; err != nil {
				return err
			}
		}

		for _, e := range existing {
			if _, keep := positionByPanelID[e.PanelID]; keep {
				continue
			}
			if err := tx.Exec(`DELETE FROM clinlims.panel_item WHERE id = ?`, e.ID).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// CreatePanel reproduces createPanel — INCLUDING the fact that it cannot work.
//
// `panel.name_localization_id` is NOT NULL and this path never creates a
// localization, so the insert violates the constraint and Java answers 500 for
// every non-blank name. Measured live, and recorded in java-defects-found.md.
// The port issues the same insert and fails the same way rather than inventing
// a localization Java does not write; a port that "fixed" it would create rows
// the Java deployment cannot.
func (d *EditorWriteDAO) CreatePanel(name string) (string, error) {
	var id string
	err := d.DB.Raw(`
		INSERT INTO clinlims.panel (id, name, description, is_active, sort_order,
		                            name_localization_id, lastupdated)
		VALUES (nextval('clinlims.panel_seq'), ?, ?, 'Y', 2147483647, NULL, ?)
		RETURNING id::text`,
		strings.TrimSpace(name), strings.TrimSpace(name),
		time.Now().UTC().Truncate(time.Millisecond)).Scan(&id).Error
	return id, err
}

// TestExists is the `testService.getTestById(testId) == null` guard every
// handler in this file opens with.
func (d *EditorWriteDAO) TestExists(testID string) (bool, error) {
	return d.exists(`SELECT id FROM clinlims.test WHERE id = ?`, testID)
}

// SampleTypeExists guards the display-order pair.
func (d *EditorWriteDAO) SampleTypeExists(sampleTypeID string) (bool, error) {
	return d.exists(`SELECT id FROM clinlims.type_of_sample WHERE id = ?`, sampleTypeID)
}

// PanelExists guards the membership save.
func (d *EditorWriteDAO) PanelExists(panelID string) (bool, error) {
	return d.exists(`SELECT id FROM clinlims.panel WHERE id = ?`, panelID)
}

func (d *EditorWriteDAO) exists(sql, id string) (bool, error) {
	// A non-numeric id would be a cast error on a numeric key column, which is
	// the 500 Java raises from the same lookup.
	var rows []string
	if err := d.DB.Raw(sql, id).Scan(&rows).Error; err != nil {
		return false, err
	}
	return len(rows) > 0, nil
}
