// Package daoimpl ports org.openelisglobal.testconfiguration.daoimpl and the
// write paths the testconfiguration REST controllers reach through their
// services. Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// RenameDAO backs the *RenameEntry screens.
//
// Every one of them except UomRenameEntry does the SAME thing: load the entity,
// take its Localization, set English and French on it, and update THAT — the
// entity's own name column is never touched. So a renamed method still has its
// original `method.name` in the table and shows the new name everywhere,
// because every list renders the localization and falls back to the column only
// when the localization is missing.
//
// UomRenameEntry is the exception and does the opposite: unit_of_measure has no
// localization at all, so it writes the name column directly. It lives in the
// unitofmeasure package for that reason.
//
// NO AUDIT ROW. reference_tables carries LOCALIZATION with keep_history = 'Y',
// and LocalizationServiceImpl extends AuditableBaseObjectServiceImpl — and, as
// with UNIT_OF_MEASURE, never sets auditTrailLog = true. The mechanism is off,
// so a rename leaves clinlims.history untouched.
type RenameDAO struct {
	DB *gorm.DB
}

// ErrNoSuchEntity is returned when the id names nothing.
//
// The controllers do NOT surface it: each guards with `if (entity != null)` and
// skips the block, so an unknown id is a silent 200 that writes nothing.
var ErrNoSuchEntity = errors.New("testconfiguration: no such row")

// renameTarget is one screen's table and the column holding its name
// localization.
type renameTarget struct {
	table string
	// column is the FK to localization. `test` has TWO — the second is the
	// reporting name, which only TestRenameEntry writes.
	column string
}

var renameTargets = map[string]renameTarget{
	"method":       {"clinlims.method", "name_localization_id"},
	"panel":        {"clinlims.panel", "name_localization_id"},
	"sampleType":   {"clinlims.type_of_sample", "name_localization_id"},
	"testSection":  {"clinlims.test_section", "name_localization_id"},
	"test":         {"clinlims.test", "name_localization_id"},
	"testReport":   {"clinlims.test", "reporting_name_localization_id"},
	"selectOption": {"clinlims.dictionary", "name_localization_id"},
}

// LocalizationIDFor returns the localization a screen's entity points at, or
// "" when the entity does not exist.
func (d *RenameDAO) LocalizationIDFor(kind, id string) (string, error) {
	t, ok := renameTargets[kind]
	if !ok {
		return "", fmt.Errorf("rename: unknown target %q", kind)
	}
	var out []string
	err := d.DB.Raw(
		fmt.Sprintf(`SELECT COALESCE(%s::text, '') FROM %s WHERE id = ?`, t.column, t.table), id).
		Scan(&out).Error
	if err != nil {
		return "", err
	}
	if len(out) == 0 {
		return "", nil
	}
	return out[0], nil
}

// Rename writes the English and French values of one localization.
//
// Java sets both through Localization.setEnglish / setFrench, which are
// setLocalizedValue("en", …) and setLocalizedValue("fr", …) — so the write is
// per-LOCALE rows in localization_value, not columns on `localization`. Both
// values are trimmed by the controller before they arrive.
//
// A locale row that does not exist is not created: setLocalizedValue updates
// the value held for that locale, and a deployment without an `fr` row for this
// localization simply has nothing to update. UPDATE matching zero rows is the
// faithful outcome, not an error.
func (d *RenameDAO) Rename(tx *gorm.DB, localizationID, english, french string) error {
	db := d.DB
	if tx != nil {
		db = tx
	}
	for locale, value := range map[string]string{"en": english, "fr": french} {
		err := db.Exec(`
			UPDATE clinlims.localization_value
			   SET value = ?, last_updated = now()
			 WHERE localization_id = ? AND locale = ?`,
			value, localizationID, locale).Error
		if err != nil {
			return err
		}
	}
	return d.DB.Exec(
		`UPDATE clinlims.localization SET lastupdated = now() WHERE id = ?`, localizationID).Error
}
