// Package daoimpl holds the DB access for the test-catalog read path.
// In Java this data is loaded implicitly by Hibernate off the Test entity from
// inside TestCatalogRestController.createTestList(); Go makes it explicit here so
// SQL stays confined to the DAO layer.
// Folder layout mirrors the Java source during migration.
package daoimpl

import (
	"math"

	"gorm.io/gorm"

	locvh "openelis-go/internal/localization/valueholder"
	"openelis-go/internal/testconfiguration/valueholder"
)

// locRow is the flat projection of the localization join — internal to the DAO.
// GetLocalizationsByIDs aggregates these rows into Localization entities.
type locRow struct {
	LocID         string  `gorm:"column:localization_id"`
	Description   *string `gorm:"column:description"`
	LastupdatedMs *int64  `gorm:"column:lastupdated_ms"`
	LvID          string  `gorm:"column:lv_id"`
	Locale        string  `gorm:"column:locale"`
	Value         string  `gorm:"column:value"`
}

// TestCatalogDAOImpl reads the test-catalog projections from clinlims.
type TestCatalogDAOImpl struct {
	DB *gorm.DB
}

// GetAllTestRows returns every test with its section name, sample type, UOM and
// sorting fields resolved in a single query.
// The LATERAL subquery picks the first sample type per test — mirrors Java's
// typeOfSampleTests.get(0).
// COALESCE(t.sort_order, MaxInt32) mirrors Java's NumberUtils.isNumber() guard:
// tests with no numeric sort order sort last.
func (d *TestCatalogDAOImpl) GetAllTestRows() ([]valueholder.TestRow, error) {
	var rows []valueholder.TestRow
	result := d.DB.Raw(`
		SELECT
		    t.id::text                                                        AS test_id,
		    COALESCE(ts_lv.value, ts.name, '')                                AS section_name,
		    COALESCE(t.sort_order, ?)::bigint                                 AS sort_order,
		    t.name_localization_id::text                                      AS name_localization_id,
		    COALESCE(tos_lv.value, tos.description, tos.local_abbrev, 'n/a')  AS sample_type,
		    COALESCE(t.is_active, 'N')                                        AS is_active,
		    COALESCE(t.orderable::text, 'false')                              AS orderable,
		    t.loinc                                                           AS loinc,
		    COALESCE(uom.name, 'n/a')                                         AS uom_name
		FROM clinlims.test t
		LEFT JOIN clinlims.test_section ts ON ts.id = t.test_section_id
		LEFT JOIN clinlims.localization_value ts_lv ON (
		    ts_lv.localization_id = ts.name_localization_id AND ts_lv.locale = 'en'
		)
		LEFT JOIN LATERAL (
		    SELECT tost.sample_type_id
		    FROM clinlims.sampletype_test tost
		    WHERE tost.test_id = t.id
		    LIMIT 1
		) first_tost ON true
		LEFT JOIN clinlims.type_of_sample tos ON tos.id = first_tost.sample_type_id
		LEFT JOIN clinlims.localization_value tos_lv ON (
		    tos_lv.localization_id = tos.name_localization_id AND tos_lv.locale = 'en'
		)
		LEFT JOIN clinlims.unit_of_measure uom ON uom.id = t.uom_id`,
		math.MaxInt32).Scan(&rows)
	if rows == nil {
		rows = []valueholder.TestRow{}
	}
	return rows, result.Error
}

// GetLocalizationsByIDs returns a localizationID → Localization map for the given
// IDs, aggregating the flat (localization × locale) join rows into entities.
// This is the bulk replacement for Java's per-test localizationService lookups.
func (d *TestCatalogDAOImpl) GetLocalizationsByIDs(ids []string) (map[string]locvh.Localization, error) {
	if len(ids) == 0 {
		return map[string]locvh.Localization{}, nil
	}

	var rows []locRow
	result := d.DB.Raw(`
		SELECT lv.localization_id::text                          AS localization_id,
		       l.description                                     AS description,
		       EXTRACT(EPOCH FROM l.lastupdated)::bigint * 1000  AS lastupdated_ms,
		       lv.id::text                                       AS lv_id,
		       lv.locale                                         AS locale,
		       lv.value                                          AS value
		FROM clinlims.localization l
		JOIN clinlims.localization_value lv ON lv.localization_id = l.id
		WHERE l.id IN ?`, ids).Scan(&rows)
	if result.Error != nil {
		return nil, result.Error
	}

	locMap := map[string]locvh.Localization{}
	for _, row := range rows {
		entry, ok := locMap[row.LocID]
		if !ok {
			entry = locvh.Localization{
				ID:     row.LocID,
				Values: map[string]locvh.LocalizationValue{},
			}
			if row.Description != nil {
				entry.Description = *row.Description
			}
			if row.LastupdatedMs != nil {
				ms := *row.LastupdatedMs
				entry.Lastupdated = &ms
			}
		}
		entry.Values[row.Locale] = locvh.LocalizationValue{
			ID:     row.LvID,
			Locale: row.Locale,
			Value:  row.Value,
		}
		locMap[row.LocID] = entry
	}
	return locMap, nil
}
