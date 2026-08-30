package daoimpl

import (
	"strconv"
	"time"

	"gorm.io/gorm"

	"openelis-go/internal/common/audittrail"
)

// AssignDAO backs the three *TestAssign screens: they attach a test to a sample
// type, to a test section, or a set of tests to a panel.
//
// All three write a JOIN and, on the side, flip an is_active flag on the thing
// being assigned to — assigning a test to an inactive test section ACTIVATES
// that section, which is a permission-shaped side effect the screen name does
// not suggest.
type AssignDAO struct {
	DB    *gorm.DB
	Audit *audittrail.Service
}

func now() time.Time { return time.Now().UTC().Truncate(time.Millisecond) }

// AssignSampleType ports sampleTypeTestAssignService.update.
//
// The join is REPLACED, not added to: every sampletype_test row for the test is
// deleted and one new row written, so a test belongs to exactly one sample type
// after this call however many it had before.
func (d *AssignDAO) AssignSampleType(testID, sampleTypeID, deactivateID string, sysUserID int64) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		ts := now()

		if err := tx.Exec(
			`DELETE FROM clinlims.sampletype_test WHERE test_id = ?`, testID).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id)
			VALUES (nextval('clinlims.sample_type_test_seq'), ?, ?)`,
			sampleTypeID, testID).Error; err != nil {
			return err
		}

		// Assigning to an inactive sample type turns it back on.
		if err := d.setSampleTypeActive(tx, sampleTypeID, true, sysUserID, ts); err != nil {
			return err
		}
		if deactivateID != "" {
			if err := d.setSampleTypeActive(tx, deactivateID, false, sysUserID, ts); err != nil {
				return err
			}
		}
		return nil
	})
}

func (d *AssignDAO) setSampleTypeActive(tx *gorm.DB, id string, active bool, sysUserID int64, ts time.Time) error {
	var old []bool
	if err := tx.Raw(
		`SELECT COALESCE(is_active, false) FROM clinlims.type_of_sample WHERE id = ?`, id).
		Scan(&old).Error; err != nil {
		return err
	}
	if len(old) == 0 || old[0] == active {
		// Already in the requested state: Hibernate finds the entity clean and
		// writes neither the UPDATE nor the history row.
		return nil
	}
	if err := tx.Exec(
		`UPDATE clinlims.type_of_sample SET is_active = ?, lastupdated = ? WHERE id = ?`,
		active, ts, id).Error; err != nil {
		return err
	}
	changes := audittrail.Field("active", strconv.FormatBool(old[0]))
	return d.Audit.Write(tx, "TYPE_OF_SAMPLE", id, sysUserID, audittrail.ActivityUpdate, &changes, ts)
}

// AssignTestSection ports testSectionTestAssignService.updateTestAndTestSections.
//
// Unlike the sample-type screen there is no join table: test.test_section_id is
// the link, so a test has exactly one section by construction.
func (d *AssignDAO) AssignTestSection(testID, sectionID, deactivateID string, sysUserID int64) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		ts := now()

		var old []string
		if err := tx.Raw(
			`SELECT COALESCE(test_section_id::text, '') FROM clinlims.test WHERE id = ?`, testID).
			Scan(&old).Error; err != nil {
			return err
		}
		if len(old) > 0 && old[0] != sectionID {
			if err := tx.Exec(
				`UPDATE clinlims.test SET test_section_id = ?, lastupdated = ? WHERE id = ?`,
				sectionID, ts, testID).Error; err != nil {
				return err
			}
			changes := audittrail.Field("testSection", old[0])
			if err := d.Audit.Write(tx, "TEST", testID, sysUserID,
				audittrail.ActivityUpdate, &changes, ts); err != nil {
				return err
			}
		}

		if err := d.setTestSectionActive(tx, sectionID, "Y", sysUserID, ts); err != nil {
			return err
		}
		if deactivateID != "" {
			if err := d.setTestSectionActive(tx, deactivateID, "N", sysUserID, ts); err != nil {
				return err
			}
		}
		return nil
	})
}

func (d *AssignDAO) setTestSectionActive(tx *gorm.DB, id, want string, sysUserID int64, ts time.Time) error {
	var old []string
	if err := tx.Raw(
		`SELECT COALESCE(is_active, '') FROM clinlims.test_section WHERE id = ?`, id).
		Scan(&old).Error; err != nil {
		return err
	}
	if len(old) == 0 || old[0] == want {
		return nil
	}
	if err := tx.Exec(
		`UPDATE clinlims.test_section SET is_active = ?, lastupdated = ? WHERE id = ?`,
		want, ts, id).Error; err != nil {
		return err
	}
	changes := audittrail.Field("isActive", old[0])
	return d.Audit.Write(tx, "TEST_SECTION", id, sysUserID, audittrail.ActivityUpdate, &changes, ts)
}

// PanelTest is one row of a panel's membership.
type PanelTest struct {
	TestID    string `gorm:"column:test_id"`
	TestName  string `gorm:"column:test_name"`
	SortOrder int    `gorm:"column:sort_order"`
}

// PanelItems reads a panel's current membership, in display order.
func (d *AssignDAO) PanelItems(panelID string) ([]PanelTest, error) {
	rows := []PanelTest{}
	err := d.DB.Table("clinlims.panel_item").
		Select(`test_id::text AS test_id, COALESCE(test_name, '') AS test_name,
		        COALESCE(sort_order, 0)::int AS sort_order`).
		Where("panel_id = ?", panelID).
		Order("sort_order, id").
		Scan(&rows).Error
	return rows, err
}

// AssignPanelTests ports panelItemService.updatePanelItems: the panel's
// membership is REPLACED by the submitted list.
func (d *AssignDAO) AssignPanelTests(panelID string, testIDs []string, sysUserID int64) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		ts := now()
		if err := tx.Exec(
			`DELETE FROM clinlims.panel_item WHERE panel_id = ?`, panelID).Error; err != nil {
			return err
		}
		for i, testID := range testIDs {
			if err := tx.Exec(`
				INSERT INTO clinlims.panel_item (id, panel_id, test_id, sort_order, lastupdated)
				VALUES (nextval('clinlims.panel_item_seq'), ?, ?, ?, ?)`,
				panelID, testID, i, ts).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
