package daoimpl

import (
	"strings"
	"time"

	"gorm.io/gorm"

	"openelis-go/internal/common/audittrail"
)

// EditorTestDAO backs the editor's test-level writes: the create-in-place flow,
// the Basic Info save, the Sample & Results save with its copy-from shortcut,
// and activation.
//
// Three of those ARE audited, and it is the only part of the editor that is:
// creating a test leaves an 'I' with a NULL payload, and Basic Info and
// activation leave a 'U' carrying the values they replaced. The Sample & Results
// save is silent, as is every section save in editor_write_dao.go.
//
// Activation also has an acknowledgment table of its own —
// test_activation_acknowledgment — written ONLY when the operator overrode a
// coverage gap.
type EditorTestDAO struct {
	DB *gorm.DB
	// Audit writes the TEST and RESULT_LIMITS history rows. See the note above
	// for which of these writes reach it.
	Audit *audittrail.Service
}

// ------------------------------------------------------------ create tests

// CreateTestParams is TestCatalogCreationService.CreateTestParams.
type CreateTestParams struct {
	Name          string
	ReportingName string
	Code          string
	LabUnitID     string
	SampleTypeID  string
	Domain        string
	AMR           bool
	Orderable     bool
	Description   string
}

// CodeInUse ports codeInUse — an EQUALS-IGNORE-CASE scan over every test's
// local_code, in Java. There is no unique index behind it, so the check is
// advisory and two concurrent creates can both pass it.
func (d *EditorTestDAO) CodeInUse(code string) (bool, error) {
	if strings.TrimSpace(code) == "" {
		return false, nil
	}
	var rows []string
	err := d.DB.Raw(
		`SELECT id FROM clinlims.test WHERE upper(local_code) = upper(?) LIMIT 1`, code).
		Scan(&rows).Error
	return len(rows) > 0, err
}

// CreateInactiveTest ports createInactiveTest.
//
// A new test starts INACTIVE and stays that way until POST .../activate lets it
// through the coverage gate — which is the whole point of the two-step flow.
// The reporting name falls back to the name; the description falls back to the
// name as well, because it doubles as the legacy base name.
//
// `test.name` is never set: Hibernate maps that column to a derived getter over
// the localization, so it lands as the localized name.
func (d *EditorTestDAO) CreateInactiveTest(p CreateTestParams, sysUserID int64) (string, error) {
	var testID string
	err := d.DB.Transaction(func(tx *gorm.DB) error {
		ts := time.Now().UTC().Truncate(time.Millisecond)

		// createNewLocalization sets BOTH locales to the same string here — the
		// create flow has one name field, not the two the legacy screen has.
		nameLoc, err := insertEditorLocalization(tx, "test name", p.Name, p.Name, ts)
		if err != nil {
			return err
		}
		reportingName := p.ReportingName
		if strings.TrimSpace(reportingName) == "" {
			reportingName = p.Name
		}
		reportLoc, err := insertEditorLocalization(tx, "test report name", reportingName, reportingName, ts)
		if err != nil {
			return err
		}

		description := p.Description
		if strings.TrimSpace(description) == "" {
			description = p.Name
		}

		var sectionID any
		if strings.TrimSpace(p.LabUnitID) != "" {
			var found []string
			if err := tx.Raw(`SELECT id FROM clinlims.test_section WHERE id = ?`, p.LabUnitID).
				Scan(&found).Error; err != nil {
				return err
			}
			if len(found) > 0 {
				// Assigning a test to an inactive lab unit ACTIVATES it, so a
				// create can turn a disabled section back on.
				if err := tx.Exec(`
					UPDATE clinlims.test_section SET is_active = 'Y', lastupdated = ?
					 WHERE id = ? AND is_active = 'N'`, ts, p.LabUnitID).Error; err != nil {
					return err
				}
				sectionID = p.LabUnitID
			}
		}

		if err := tx.Raw(`
			INSERT INTO clinlims.test
			       (id, name, description, local_code, domain, test_section_id, is_active,
			        orderable, antimicrobial_resistance, is_reportable, guid, sort_order,
			        name_localization_id, reporting_name_localization_id, lastupdated)
			VALUES (nextval('clinlims.test_seq'), ?, ?, ?, ?, ?, 'N', ?, ?, 'N',
			        gen_random_uuid()::text, 0, ?, ?, ?)
			RETURNING id::text`,
			p.Name, description, p.Code, p.Domain, sectionID,
			p.Orderable, p.AMR, nameLoc, reportLoc, ts).Scan(&testID).Error; err != nil {
			return err
		}

		// Creating a test IS audited — one 'I' with a NULL payload, the same
		// shape TestAdd leaves.
		if err := d.Audit.Write(tx, "TEST", testID, sysUserID, audittrail.ActivityInsert, nil, ts); err != nil {
			return err
		}

		if strings.TrimSpace(p.SampleTypeID) != "" {
			if err := tx.Exec(`
				INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id)
				VALUES (nextval('clinlims.sample_type_test_seq'), ?, ?)`,
				p.SampleTypeID, testID).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return testID, err
}

func insertEditorLocalization(tx *gorm.DB, description, english, french string, ts time.Time) (string, error) {
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

// ------------------------------------------------------------- basic info

// BasicInfoRow is the test as the Basic Info section reads it.
type BasicInfoRow struct {
	ID           string  `gorm:"column:id"`
	Name         *string `gorm:"column:name"`
	Code         *string `gorm:"column:code"`
	Description  *string `gorm:"column:description"`
	Domain       *string `gorm:"column:domain"`
	LabUnitID    *string `gorm:"column:lab_unit_id"`
	SampleTypeID *string `gorm:"column:sample_type_id"`
	AMR          bool    `gorm:"column:amr"`
	Active       bool    `gorm:"column:active"`
	Orderable    bool    `gorm:"column:orderable"`
}

// BasicInfo reads one test's identity block.
//
// sampleTypeId is the FIRST sampletype_test link, which is what
// testService.getTypeOfSample returns — the editor models one sample type per
// test even though the join table does not.
func (d *EditorTestDAO) BasicInfo(testID string) (*BasicInfoRow, error) {
	rows := []BasicInfoRow{}
	err := d.DB.Raw(`
		SELECT t.id::text AS id, t.name AS name, t.local_code AS code, t.description AS description,
		       t.domain AS domain, t.test_section_id::text AS lab_unit_id,
		       (SELECT st.sample_type_id::text FROM clinlims.sampletype_test st
		         WHERE st.test_id = t.id ORDER BY st.id LIMIT 1) AS sample_type_id,
		       COALESCE(t.antimicrobial_resistance, false) AS amr,
		       (t.is_active = 'Y') AS active,
		       COALESCE(t.orderable, false) AS orderable
		  FROM clinlims.test t WHERE t.id = ?`, testID).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return &rows[0], nil
}

// BasicInfoUpdate is the set of fields a Basic Info save may move. A nil means
// the caller did not send the key, and the column is left alone — the handler
// applies "only what the caller actually sent" so a partial PUT cannot silently
// deactivate a test.
type BasicInfoUpdate struct {
	Code         *string
	Description  *string
	Domain       *string
	AMR          *bool
	Orderable    *bool
	LabUnitID    string
	SampleTypeID string
	// Deactivate is set only when the body carried active=false. active=true is
	// IGNORED here: turning a test on has to go through the coverage gate at
	// POST .../activate, so Basic Info can only ever turn one off.
	Deactivate bool
}

// SaveBasicInfo ports saveBasicInfo's write half.
func (d *EditorTestDAO) SaveBasicInfo(testID string, u BasicInfoUpdate, sysUserID int64) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		ts := time.Now().UTC().Truncate(time.Millisecond)

		before, err := testAuditState(tx, testID)
		if err != nil {
			return err
		}

		sets := []string{}
		args := []any{}
		if u.Code != nil && strings.TrimSpace(*u.Code) != "" {
			sets = append(sets, "local_code = ?")
			args = append(args, *u.Code)
		}
		if u.Description != nil {
			// A non-null description is applied even when BLANK — the guard is on
			// null, not on emptiness, so "" clears the column.
			sets = append(sets, "description = ?")
			args = append(args, *u.Description)
		}
		if u.Domain != nil {
			sets = append(sets, "domain = ?")
			args = append(args, *u.Domain)
		}
		if u.AMR != nil {
			sets = append(sets, "antimicrobial_resistance = ?")
			args = append(args, *u.AMR)
		}
		if u.Orderable != nil {
			sets = append(sets, "orderable = ?")
			args = append(args, *u.Orderable)
		}
		if strings.TrimSpace(u.LabUnitID) != "" {
			var found []string
			if err := tx.Raw(`SELECT id FROM clinlims.test_section WHERE id = ?`, u.LabUnitID).
				Scan(&found).Error; err != nil {
				return err
			}
			if len(found) > 0 {
				if err := tx.Exec(`
					UPDATE clinlims.test_section SET is_active = 'Y', lastupdated = ?
					 WHERE id = ? AND is_active = 'N'`, ts, u.LabUnitID).Error; err != nil {
					return err
				}
				sets = append(sets, "test_section_id = ?")
				args = append(args, u.LabUnitID)
			}
		}
		if u.Deactivate {
			sets = append(sets, "is_active = 'N'")
		}

		// testService.update runs whether or not a field moved, and the flush
		// rewrites `name` from the localization either way.
		sets = append(sets,
			`name = COALESCE(NULLIF((SELECT lv.value FROM clinlims.localization_value lv
			                          WHERE lv.localization_id = clinlims.test.name_localization_id
			                            AND lv.locale = 'en'), ''), clinlims.test.description)`,
			"lastupdated = ?")
		args = append(args, ts, testID)
		if err := tx.Exec(
			`UPDATE clinlims.test SET `+strings.Join(sets, ", ")+` WHERE id = ?`, args...).Error; err != nil {
			return err
		}

		// The audit row records what the save REPLACED, and only for the fields
		// that actually moved — so a PUT that re-sends the stored values leaves
		// no trace. The sample-type relink below is not part of it.
		after, err := testAuditState(tx, testID)
		if err != nil {
			return err
		}
		if err := writeTestUpdateAudit(tx, d.Audit, testID, before, after, sysUserID, ts); err != nil {
			return err
		}

		if strings.TrimSpace(u.SampleTypeID) == "" {
			return nil
		}
		// The link is reconciled to ONE sample type — replace-all, and skipped
		// entirely when the test already has exactly that one link.
		current := []string{}
		if err := tx.Raw(
			`SELECT sample_type_id::text FROM clinlims.sampletype_test WHERE test_id = ? ORDER BY id`,
			testID).Scan(&current).Error; err != nil {
			return err
		}
		if len(current) == 1 && current[0] == u.SampleTypeID {
			return nil
		}
		if err := tx.Exec(`DELETE FROM clinlims.sampletype_test WHERE test_id = ?`, testID).Error; err != nil {
			return err
		}
		return tx.Exec(`
			INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id)
			VALUES (nextval('clinlims.sample_type_test_seq'), ?, ?)`,
			u.SampleTypeID, testID).Error
	})
}

// TestSectionExists guards the lab-unit assignment.
func (d *EditorTestDAO) TestSectionExists(id string) (bool, error) {
	var rows []string
	err := d.DB.Raw(`SELECT id FROM clinlims.test_section WHERE id = ?`, id).Scan(&rows).Error
	return len(rows) > 0, err
}

// ---------------------------------------------------------------- activation

// LimitAge is one reference range reduced to what the coverage gate reads.
type LimitAge struct {
	Gender string  `gorm:"column:gender"`
	MinAge float64 `gorm:"column:min_age"`
	MaxAge float64 `gorm:"column:max_age"`
}

// LimitAges reads a test's ranges for the coverage validation.
func (d *EditorTestDAO) LimitAges(testID string) ([]LimitAge, error) {
	rows := []LimitAge{}
	err := d.DB.Raw(`
		SELECT COALESCE(gender, '') AS gender,
		       COALESCE(min_age, 0) AS min_age,
		       COALESCE(max_age, 'Infinity'::float8) AS max_age
		  FROM clinlims.result_limits WHERE test_id = ? ORDER BY id`, testID).Scan(&rows).Error
	return rows, err
}

// Activate turns a test on, recording an acknowledgment when one was needed.
//
// Activation also forces `orderable` TRUE regardless of what the test carried:
// the order picker filters on it, so activation alone was not enough to make a
// test appear (OGC-1116).
func (d *EditorTestDAO) Activate(testID string, gapsAcknowledged *string, sysUserID int64) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		ts := time.Now().UTC().Truncate(time.Millisecond)
		before, err := testAuditState(tx, testID)
		if err != nil {
			return err
		}
		if gapsAcknowledged != nil {
			if err := tx.Exec(`
				INSERT INTO clinlims.test_activation_acknowledgment
				       (id, test_id, user_id, acknowledged_at, gaps_acknowledged, lastupdated)
				VALUES (gen_random_uuid()::text, ?, ?, ?, ?::jsonb, ?)`,
				testID, sysUserID, ts, *gapsAcknowledged, ts).Error; err != nil {
				return err
			}
		}
		if err := tx.Exec(`
			UPDATE clinlims.test SET is_active = 'Y', orderable = true, lastupdated = ?
			 WHERE id = ?`, ts, testID).Error; err != nil {
			return err
		}
		after, err := testAuditState(tx, testID)
		if err != nil {
			return err
		}
		// Activating an already-active, already-orderable test moves nothing and
		// is therefore not audited.
		return writeTestUpdateAudit(tx, d.Audit, testID, before, after, sysUserID, ts)
	})
}
