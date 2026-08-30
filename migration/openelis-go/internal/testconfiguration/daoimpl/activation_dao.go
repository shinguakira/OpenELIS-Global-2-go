package daoimpl

import (
	"sort"
	"strconv"
	"time"

	"gorm.io/gorm"

	"openelis-go/internal/common/audittrail"
)

// ActivationDAO backs TestActivation and TestOrderability — the two screens
// that turn tests (and, for TestActivation, sample types) on and off.
type ActivationDAO struct {
	DB           *gorm.DB
	Audit        *audittrail.Service
	ActiveLocale string
}

func (d *ActivationDAO) locale() string {
	if d.ActiveLocale != "" {
		return d.ActiveLocale
	}
	return "en"
}

// TestRow is one test as the activation screens see it.
type TestRow struct {
	SampleTypeID string `gorm:"column:sample_type_id"`
	ID           string `gorm:"column:id"`
	Name         string `gorm:"column:name"`
	// AugmentedName is getLocalizedTestNameWithType: the localized test name
	// with the sample type appended in brackets. Several shipped tests already
	// carry the type in their own name, so the result reads doubled —
	// "WBC(Whole Blood)(Whole Blood)" — and that is the wire value.
	AugmentedName string `gorm:"column:augmented_name"`
	IsActive      string `gorm:"column:is_active"`
	Orderable     bool   `gorm:"column:orderable"`
	SortOrder     string `gorm:"column:sort_order"`
}

// TestsBySampleType ports typeOfSampleService.getAllTestsBySampleTypeId for
// every sample type at once, with the LOCALIZED test name.
//
// The name is getUserLocalizedTestName, which resolves the test's
// name_localization_id for the active locale and falls back to the column —
// the same rule every other list here follows.
func (d *ActivationDAO) TestsBySampleType() ([]TestRow, error) {
	rows := []TestRow{}
	err := d.DB.Table("clinlims.sampletype_test AS tost").
		Select(`tost.sample_type_id::text AS sample_type_id,
		        t.id::text AS id,
		        COALESCE(NULLIF(lv.value, ''), t.name) AS name,
		        COALESCE(NULLIF(lv.value, ''), t.name) ||
		          '(' || COALESCE(NULLIF(stlv.value, ''), st.description) || ')' AS augmented_name,
		        COALESCE(t.is_active, '') AS is_active,
		        COALESCE(t.orderable, false) AS orderable,
		        COALESCE(t.sort_order::text, '') AS sort_order`).
		Joins("JOIN clinlims.test AS t ON t.id = tost.test_id").
		Joins("JOIN clinlims.type_of_sample AS st ON st.id = tost.sample_type_id").
		Joins(`LEFT JOIN clinlims.localization_value AS stlv
		         ON stlv.localization_id = st.name_localization_id AND stlv.locale = ?`, d.locale()).
		Joins(`LEFT JOIN clinlims.localization_value AS lv
		         ON lv.localization_id = t.name_localization_id AND lv.locale = ?`, d.locale()).
		Order("tost.id").
		Scan(&rows).Error
	return rows, err
}

// AllTests reads every test, for TestOrderability's list.
func (d *ActivationDAO) AllTests() ([]TestRow, error) {
	rows := []TestRow{}
	err := d.DB.Table("clinlims.test AS t").
		Select(`'' AS sample_type_id, '' AS augmented_name, t.id::text AS id,
		        COALESCE(NULLIF(lv.value, ''), t.name) AS name,
		        COALESCE(t.is_active, '') AS is_active,
		        COALESCE(t.orderable, false) AS orderable,
		        COALESCE(t.sort_order::text, '') AS sort_order`).
		Joins(`LEFT JOIN clinlims.localization_value AS lv
		         ON lv.localization_id = t.name_localization_id AND lv.locale = ?`, d.locale()).
		Order("t.id").
		Scan(&rows).Error
	return rows, err
}

// TestChange is one test's new state.
type TestChange struct {
	ID string
	// Active is written to test.is_active as the CHAR 'Y'/'N'.
	Active bool
	// SortOrder is written only when Set — a deactivation leaves it alone.
	SortOrder    int
	SortOrderSet bool
}

// SampleTypeChange is one sample type's new state. type_of_sample.is_active is
// a real boolean, unlike test.is_active.
type SampleTypeChange struct {
	ID           string
	Active       bool
	SortOrder    int
	SortOrderSet bool
}

// Apply writes both sets in one transaction — TestActivationService.updateAll
// walks four lists inside one @Transactional method.
//
// Each changed row leaves an audit entry carrying the value it replaced, which
// is what testService.update and typeOfSampleService.update do with
// auditTrailLog set. A row already in the requested state is skipped: Hibernate
// finds the entity clean and writes neither the UPDATE nor the history row.
func (d *ActivationDAO) Apply(tests []TestChange, sampleTypes []SampleTypeChange, sysUserID int64) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		ts := time.Now().UTC().Truncate(time.Millisecond)

		for _, c := range tests {
			var old []struct {
				IsActive  string `gorm:"column:is_active"`
				SortOrder string `gorm:"column:sort_order"`
			}
			if err := tx.Raw(`SELECT COALESCE(is_active, '') AS is_active,
			                         COALESCE(sort_order::text, '') AS sort_order
			                    FROM clinlims.test WHERE id = ?`, c.ID).Scan(&old).Error; err != nil {
				return err
			}
			if len(old) == 0 {
				continue
			}
			want := "N"
			if c.Active {
				want = "Y"
			}
			changes := ""
			if old[0].IsActive != want {
				changes += audittrail.Field("isActive", old[0].IsActive)
			}
			newOrder := old[0].SortOrder
			if c.SortOrderSet {
				newOrder = strconv.Itoa(c.SortOrder)
				if old[0].SortOrder != newOrder {
					changes += audittrail.Field("sortOrder", old[0].SortOrder)
				}
			}
			if changes == "" {
				continue
			}
			if err := tx.Exec(
				`UPDATE clinlims.test SET is_active = ?, sort_order = ?, lastupdated = ? WHERE id = ?`,
				want, newOrder, ts, c.ID).Error; err != nil {
				return err
			}
			if err := d.Audit.Write(tx, "TEST", c.ID, sysUserID,
				audittrail.ActivityUpdate, &changes, ts); err != nil {
				return err
			}
		}

		for _, c := range sampleTypes {
			var old []struct {
				IsActive  bool   `gorm:"column:is_active"`
				SortOrder string `gorm:"column:sort_order"`
			}
			if err := tx.Raw(`SELECT COALESCE(is_active, false) AS is_active,
			                         COALESCE(sort_order::text, '') AS sort_order
			                    FROM clinlims.type_of_sample WHERE id = ?`, c.ID).Scan(&old).Error; err != nil {
				return err
			}
			if len(old) == 0 {
				continue
			}
			changes := ""
			if old[0].IsActive != c.Active {
				changes += audittrail.Field("active", strconv.FormatBool(old[0].IsActive))
			}
			newOrder := old[0].SortOrder
			if c.SortOrderSet {
				newOrder = strconv.Itoa(c.SortOrder)
				if old[0].SortOrder != newOrder {
					changes += audittrail.Field("sortOrder", old[0].SortOrder)
				}
			}
			if changes == "" {
				continue
			}
			if err := tx.Exec(
				`UPDATE clinlims.type_of_sample SET is_active = ?, sort_order = ?, lastupdated = ? WHERE id = ?`,
				c.Active, newOrder, ts, c.ID).Error; err != nil {
				return err
			}
			if err := d.Audit.Write(tx, "TYPE_OF_SAMPLE", c.ID, sysUserID,
				audittrail.ActivityUpdate, &changes, ts); err != nil {
				return err
			}
		}
		return nil
	})
}

// SetOrderable ports TestOrderability's write: test.orderable and nothing else.
func (d *ActivationDAO) SetOrderable(ids []string, orderable bool, sysUserID int64) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		ts := time.Now().UTC().Truncate(time.Millisecond)
		for _, id := range ids {
			var old []bool
			if err := tx.Raw(
				`SELECT COALESCE(orderable, false) FROM clinlims.test WHERE id = ?`, id).
				Scan(&old).Error; err != nil {
				return err
			}
			if len(old) == 0 || old[0] == orderable {
				continue
			}
			if err := tx.Exec(
				`UPDATE clinlims.test SET orderable = ?, lastupdated = ? WHERE id = ?`,
				orderable, ts, id).Error; err != nil {
				return err
			}
			changes := audittrail.Field("orderable", strconv.FormatBool(old[0]))
			if err := d.Audit.Write(tx, "TEST", id, sysUserID,
				audittrail.ActivityUpdate, &changes, ts); err != nil {
				return err
			}
		}
		return nil
	})
}

// SortTestsForDisplay is the comparator createTestList applies before splitting
// the list: numeric sort_order ascending, and a test WITHOUT a numeric order
// sorts after one that has it. Collections.sort is STABLE, so ties keep the
// order the loader returned — which is why sort.SliceStable is the right call
// and sort.Slice is not.
func SortTestsForDisplay(tests []TestRow) []TestRow {
	out := append([]TestRow(nil), tests...)
	sort.SliceStable(out, func(i, j int) bool {
		a, aerr := strconv.Atoi(out[i].SortOrder)
		b, berr := strconv.Atoi(out[j].SortOrder)
		switch {
		case aerr == nil && berr == nil:
			return a < b
		case aerr == nil:
			return true
		default:
			return false
		}
	})
	return out
}

// TestsByTestSection is the section equivalent of TestsBySampleType: the tests
// under each test_section, keyed by section id.
//
// test.test_section_id is the link — there is no join table — but the augmented
// name still needs the test's SAMPLE type, so the join to sampletype_test is
// still here. A test with no sample type drops out, exactly as it does from
// getAllTestsBySampleTypeId's side.
func (d *ActivationDAO) TestsByTestSection() (map[string][]TestRow, error) {
	rows := []TestRow{}
	err := d.DB.Table("clinlims.test AS t").
		Select(`COALESCE(t.test_section_id::text, '') AS sample_type_id,
		        t.id::text AS id,
		        COALESCE(NULLIF(lv.value, ''), t.name) AS name,
		        COALESCE(NULLIF(lv.value, ''), t.name) ||
		          '(' || COALESCE(NULLIF(stlv.value, ''), st.description) || ')' AS augmented_name,
		        COALESCE(t.is_active, '') AS is_active,
		        COALESCE(t.orderable, false) AS orderable,
		        COALESCE(t.sort_order::text, '') AS sort_order`).
		Joins("JOIN clinlims.sampletype_test AS tost ON tost.test_id = t.id").
		Joins("JOIN clinlims.type_of_sample AS st ON st.id = tost.sample_type_id").
		Joins(`LEFT JOIN clinlims.localization_value AS lv
		         ON lv.localization_id = t.name_localization_id AND lv.locale = ?`, d.locale()).
		Joins(`LEFT JOIN clinlims.localization_value AS stlv
		         ON stlv.localization_id = st.name_localization_id AND stlv.locale = ?`, d.locale()).
		Order("t.id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := map[string][]TestRow{}
	for _, r := range rows {
		out[r.SampleTypeID] = append(out[r.SampleTypeID], r)
	}
	return out, nil
}
