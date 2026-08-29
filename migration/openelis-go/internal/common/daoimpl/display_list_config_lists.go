package daoimpl

import "openelis-go/internal/common/util"

// The DisplayListService list types the e2 admin screens read.
//
// Every one of them renders getLocalizedName() / getLocalization()
// .getLocalizedValue(), which is localization_value for the active locale
// falling back to the entity's own column — the same LEFT JOIN the lists above
// use. What differs between them is the source query, the active filter and
// the sort, and each of those is Java's, not a tidier equivalent.

// Methods ports createMethodList: METHODS.
//
// methodService.getAll() — every method, active or not, and no ORDER BY. The
// list is NOT sorted anywhere on the way to the form, so the row order is the
// plan's; id order on the primary-key scan.
func (d *DisplayListDAOImpl) Methods() ([]util.IdValuePair, error) {
	return d.methodList("")
}

// InactiveMethods ports createInactiveMethod: METHODS_INACTIVE.
//
// getAllInActiveMethods(), so `is_active = 'N'` — method.is_active is the CHAR
// 'Y'/'N', not a boolean.
func (d *DisplayListDAOImpl) InactiveMethods() ([]util.IdValuePair, error) {
	return d.methodList("m.is_active = 'N'")
}

func (d *DisplayListDAOImpl) methodList(where string) ([]util.IdValuePair, error) {
	rows := []idValueRow{}
	q := d.DB.Table("clinlims.method AS m").
		Select(`m.id AS id, COALESCE(NULLIF(lv.value, ''), m.name) AS value`).
		Joins(`LEFT JOIN clinlims.localization_value AS lv
		         ON lv.localization_id = m.name_localization_id AND lv.locale = ?`, d.Locale()).
		Order("m.id")
	if where != "" {
		q = q.Where(where)
	}
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	return toPairs(rows), nil
}

// InactiveTestSections ports createInactiveTestSection: TEST_SECTION_INACTIVE.
//
// getAllInActiveTestSections(). ActiveTestSections above orders by sort_order
// because getAllActiveTestSections does; this one has no ordering of its own.
func (d *DisplayListDAOImpl) InactiveTestSections() ([]util.IdValuePair, error) {
	rows := []idValueRow{}
	err := d.DB.Table("clinlims.test_section AS ts").
		Select(`ts.id AS id, COALESCE(NULLIF(lv.value, ''), ts.name) AS value`).
		Joins(`LEFT JOIN clinlims.localization_value AS lv
		         ON lv.localization_id = ts.name_localization_id AND lv.locale = ?`, d.Locale()).
		Where("ts.is_active = 'N'").
		Order("ts.id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return toPairs(rows), nil
}

// Panels ports createPanelList(): PANELS.
//
// getAllPanels() then Collections.sort with PanelSortOrderComparator, which
// compares getSortOrderInt() — the sort_order column parsed as an INTEGER.
// The column is text here, so ordering by it in SQL would sort "10" before
// "9"; the cast is what Java's comparator does.
//
// Collections.sort is stable, so panels sharing a sort_order keep the order
// getAllPanels returned them in — which is the query plan's, id order in
// practice. Ordering by (sort_order::int, id) reproduces that.
func (d *DisplayListDAOImpl) Panels() ([]util.IdValuePair, error) {
	return d.panelList("")
}

// ActivePanels ports createPanelList(false): PANELS_ACTIVE.
func (d *DisplayListDAOImpl) ActivePanels() ([]util.IdValuePair, error) {
	return d.panelList("active")
}

// InactivePanels ports createPanelList(true): PANELS_INACTIVE.
//
// Like createSampleTypeList, the active filter runs in JAVA after the sort, not
// in the query — so it cannot change the plan or the tie order.
func (d *DisplayListDAOImpl) InactivePanels() ([]util.IdValuePair, error) {
	return d.panelList("inactive")
}

// panelRow is idValueRow plus the active flag.
//
// IsActive is a STRING, not a bool: panel.is_active is varchar(1) holding 'Y'
// or 'N', like dictionary / test / test_section and unlike type_of_sample,
// whose column is a real boolean. Scanning the char into a Go bool yields
// false for every row and empties the filtered lists — the same trap the
// active-flag note above records for the other direction.
type panelRow struct {
	ID       string `gorm:"column:id"`
	Value    string `gorm:"column:value"`
	IsActive string `gorm:"column:is_active"`
}

func (d *DisplayListDAOImpl) panelList(filter string) ([]util.IdValuePair, error) {
	rows := []panelRow{}
	err := d.DB.Table("clinlims.panel AS p").
		Select(`p.id AS id, COALESCE(NULLIF(lv.value, ''), p.name) AS value,
		        COALESCE(p.is_active, '') AS is_active`).
		Joins(`LEFT JOIN clinlims.localization_value AS lv
		         ON lv.localization_id = p.name_localization_id AND lv.locale = ?`, d.Locale()).
		// sort_order is NUMERIC here, so ordering by the column is what
		// getSortOrderInt() compares. It is not text — the cast an earlier
		// version applied fails outright.
		Order("p.sort_order, p.id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]util.IdValuePair, 0, len(rows))
	for _, r := range rows {
		active := r.IsActive == "Y"
		switch filter {
		case "active":
			if !active {
				continue
			}
		case "inactive":
			if active {
				continue
			}
		}
		out = append(out, util.NewIdValuePair(r.ID, r.Value))
	}
	return out, nil
}

// InactiveHumanSampleTypes ports createSampleTypeList(true): SAMPLE_TYPE_INACTIVE.
//
// The same query as ActiveHumanSampleTypes with the opposite half of the
// filter, and for the same reason the predicate stays out of the SQL — see the
// note there about sort_order ties changing with the plan.
func (d *DisplayListDAOImpl) InactiveHumanSampleTypes() ([]util.IdValuePair, error) {
	return d.humanSampleTypes(false)
}

func (d *DisplayListDAOImpl) humanSampleTypes(active bool) ([]util.IdValuePair, error) {
	rows := []activeFlagRow{}
	err := d.DB.Table("clinlims.type_of_sample AS t").
		Select(`t.id AS id, COALESCE(NULLIF(lv.value, ''), t.description) AS value, t.is_active AS is_active`).
		Joins(`LEFT JOIN clinlims.localization_value AS lv
		         ON lv.localization_id = t.name_localization_id AND lv.locale = ?`, d.Locale()).
		Where("t.domain = 'H'").
		Order("t.sort_order").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]util.IdValuePair, 0, len(rows))
	for _, r := range rows {
		if r.IsActive != active {
			continue
		}
		out = append(out, util.NewIdValuePair(r.ID, r.Value))
	}
	return out, nil
}

// AllSampleTypes ports createTypeOfSampleList: SAMPLE_TYPE — every human type,
// active or not, in sort order.
func (d *DisplayListDAOImpl) AllSampleTypes() ([]util.IdValuePair, error) {
	rows := []idValueRow{}
	err := d.DB.Table("clinlims.type_of_sample AS t").
		Select(`t.id AS id, COALESCE(NULLIF(lv.value, ''), t.description) AS value`).
		Joins(`LEFT JOIN clinlims.localization_value AS lv
		         ON lv.localization_id = t.name_localization_id AND lv.locale = ?`, d.Locale()).
		Where("t.domain = 'H'").
		Order("t.sort_order").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return toPairs(rows), nil
}
