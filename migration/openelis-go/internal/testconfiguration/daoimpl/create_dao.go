package daoimpl

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"openelis-go/internal/common/audittrail"
)

// CreateDAO backs the *Create screens for Method, TestSection and SampleType.
//
// Each of them writes EIGHT rows in one transaction — the entity is the small
// part of it:
//
//	localization        1   description is the screen's own label
//	localization_value  2   'en' and 'fr', from the two submitted names
//	<entity>            1   name_localization_id points at the localization
//	system_module       3   Workplan, LogbookResults, ResultValidation
//	system_role_module  3   two of them to Results, one to Validation
//
// MEASURED, not read off the source: creating a method through Java moves
// exactly those six tables by those counts, and leaves ONE history row.
//
// That one row is for the ENTITY, with a NULL payload — the insert shape e1
// pinned. The localization, the modules and the role links are NOT audited,
// even though LOCALIZATION is flagged keep_history = 'Y', because the services
// behind them never set auditTrailLog. Two of the three tables in this wave now
// behave that way.
type CreateDAO struct {
	DB    *gorm.DB
	Audit *audittrail.Service
}

// writeTime matches Java's new Timestamp(System.currentTimeMillis()) —
// milliseconds, not the microseconds now() would store.
func writeTime() time.Time { return time.Now().UTC().Truncate(time.Millisecond) }

// The three modules every create makes, in the order Java makes them.
//
// The names are built by SystemModule's own convention: `Workplan:<name>` for
// the name and `Workplan=><name>` for the description. Both are measured off
// rows Java wrote.
var createdModules = []struct {
	Module string
	// Role is the system_role the module is linked to. Workplan and
	// LogbookResults go to Results; ResultValidation goes to Validation.
	Role string
}{
	{"Workplan", "Results"},
	{"LogbookResults", "Results"},
	{"ResultValidation", "Validation"},
}

// CreateSpec is one screen's differences from the others.
type CreateSpec struct {
	// Table is the entity table.
	Table string
	// LocalizationDescription is the literal createLocalization puts in
	// localization.description — "method name", "test unit name",
	// "type of sample name". It is a label, not a value, and it is per-screen.
	LocalizationDescription string
	// AuditTable is the reference_tables name the entity's history row keys on.
	AuditTable string
	// Columns are the entity columns beyond id/name/description/localization,
	// as literal SQL fragments keyed by column name.
	ExtraColumns map[string]any
	// NameColumn and DescriptionColumn differ: type_of_sample has no `name`.
	NameColumn        string
	DescriptionColumn string
}

// Create runs the whole eight-row write.
//
// One transaction, because insertMethod and its siblings are @Transactional and
// a half-made method — localization written, modules missing — would leave the
// screen offering a name whose permissions do not exist.
func (d *CreateDAO) Create(spec CreateSpec, english, french string, sysUserID int64) (string, error) {
	var entityID string
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		ts := writeTime()

		var locID string
		if err := tx.Raw(`
			INSERT INTO clinlims.localization (id, description, lastupdated)
			VALUES (nextval('clinlims.localization_seq'), ?, ?)
			RETURNING id::text`, spec.LocalizationDescription, ts).Scan(&locID).Error; err != nil {
			return err
		}

		// 'en' takes the English name and 'fr' the French one — setEnglish and
		// setFrench are setLocalizedValue("en"/"fr", …). No other locale is
		// written, however many the deployment has active.
		for locale, value := range map[string]string{"en": english, "fr": french} {
			if err := tx.Exec(`
				INSERT INTO clinlims.localization_value (id, localization_id, locale, value, last_updated)
				VALUES (nextval('clinlims.localization_value_seq'), ?, ?, ?, ?)`,
				locID, locale, value, ts).Error; err != nil {
				return err
			}
		}

		// type_of_sample has no `name` column — createTypeOfSample calls
		// setDescription only — so its two column names are the same one and it
		// must not be listed twice.
		cols := []string{spec.NameColumn, "name_localization_id", "lastupdated"}
		vals := []any{english, locID, ts}
		if spec.DescriptionColumn != spec.NameColumn {
			cols = append(cols, spec.DescriptionColumn)
			vals = append(vals, english)
		}
		for c, v := range spec.ExtraColumns {
			cols = append(cols, c)
			vals = append(vals, v)
		}
		placeholders := strings.TrimSuffix(strings.Repeat("?, ", len(cols)), ", ")
		sql := fmt.Sprintf(
			`INSERT INTO %s (id, %s) VALUES (nextval('%s_seq'), %s) RETURNING id::text`,
			spec.Table, strings.Join(cols, ", "), spec.Table, placeholders)
		if err := tx.Raw(sql, vals...).Scan(&entityID).Error; err != nil {
			return err
		}

		for _, m := range createdModules {
			var moduleID string
			if err := tx.Raw(`
				INSERT INTO clinlims.system_module (id, name, description)
				VALUES (nextval('clinlims.system_module_seq'), ?, ?)
				RETURNING id::text`,
				m.Module+":"+english, m.Module+"=>"+english).Scan(&moduleID).Error; err != nil {
				return err
			}
			// createRoleModule sets all four permissions to 'Y'.
			if err := tx.Exec(`
				INSERT INTO clinlims.system_role_module
				       (id, system_role_id, system_module_id, has_select, has_add, has_update, has_delete)
				VALUES (nextval('clinlims.system_role_module_seq'),
				        (SELECT id FROM clinlims.system_role WHERE trim(name) = ?), ?, 'Y', 'Y', 'Y', 'Y')`,
				m.Role, moduleID).Error; err != nil {
				return err
			}
		}

		// The entity's history row, and only the entity's. saveNewHistory sets
		// no payload, so changes is NULL rather than empty.
		return d.Audit.Write(tx, spec.AuditTable, entityID, sysUserID, audittrail.ActivityInsert, nil, ts)
	})
	return entityID, err
}
