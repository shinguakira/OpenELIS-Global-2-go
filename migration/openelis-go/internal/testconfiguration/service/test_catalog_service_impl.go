// Package service ports the data-assembly logic of
// org.openelisglobal.testconfiguration.controller.rest.TestCatalogRestController.createTestList().
// In Java that logic lives inline in the controller; here it is separated into
// a service to respect the layered architecture.
// Folder layout mirrors the Java source during migration.
package service

import (
	"math"
	"sort"

	"gorm.io/gorm"

	locvh "openelis-go/internal/localization/valueholder"
	"openelis-go/internal/testconfiguration/form"
	testvh "openelis-go/internal/test/valueholder"
)

// testRow is the raw DB projection from fetchTests — internal to the service.
// Fields are exported so GORM's Scan can populate them via reflection.
type testRow struct {
	TestID      string  `gorm:"column:test_id"`
	SectionName string  `gorm:"column:section_name"`
	SortOrder   int64   `gorm:"column:sort_order"`
	LocID       *string `gorm:"column:name_localization_id"`
	SampleType  string  `gorm:"column:sample_type"`
	Active      string  `gorm:"column:is_active"`
	Orderable   string  `gorm:"column:orderable"`
	Loinc       *string `gorm:"column:loinc"`
	UomName     string  `gorm:"column:uom_name"`
}

// locRow is the raw DB projection from fetchLocalizations — internal to the service.
type locRow struct {
	LocID         string  `gorm:"column:localization_id"`
	Description   *string `gorm:"column:description"`
	LastupdatedMs *int64  `gorm:"column:lastupdated_ms"`
	LvID          string  `gorm:"column:lv_id"`
	Locale        string  `gorm:"column:locale"`
	Value         string  `gorm:"column:value"`
}

// TestCatalogService assembles the TestCatalogForm from the DB via GORM.
// Mirrors the read path of TestCatalogRestController.showTestCatalog() +
// createTestList().
type TestCatalogService struct {
	DB *gorm.DB
}

// BuildForm mirrors TestCatalogRestController.showTestCatalog() +
// createTestList(): fetches all tests with their joined metadata, resolves
// localization values in bulk, sorts by (testUnit, sampleType, testSortOrder),
// and builds the form that the controller serialises to JSON.
func (s *TestCatalogService) BuildForm() (*form.TestCatalogForm, error) {
	tests, err := s.fetchTests()
	if err != nil {
		return nil, err
	}
	if len(tests) == 0 {
		return &form.TestCatalogForm{
			FormName:        "testCatalogForm",
			TestCatalogList: []testvh.TestCatalog{},
			TestSectionList: []string{},
		}, nil
	}

	// Collect unique localization IDs for the bulk fetch.
	locIDs := make([]string, 0, len(tests))
	seenLoc := map[string]bool{}
	for _, t := range tests {
		if t.LocID != nil && !seenLoc[*t.LocID] {
			locIDs = append(locIDs, *t.LocID)
			seenLoc[*t.LocID] = true
		}
	}
	locMap, err := s.fetchLocalizations(locIDs)
	if err != nil {
		return nil, err
	}

	// Sort by (sectionName, sampleType, testSortOrder) — mirrors Java Comparator.
	sort.SliceStable(tests, func(i, j int) bool {
		if tests[i].SectionName != tests[j].SectionName {
			return tests[i].SectionName < tests[j].SectionName
		}
		if tests[i].SampleType != tests[j].SampleType {
			return tests[i].SampleType < tests[j].SampleType
		}
		return tests[i].SortOrder < tests[j].SortOrder
	})

	catalogList := make([]testvh.TestCatalog, 0, len(tests))
	sectionList := []string{}
	seenSections := map[string]bool{}

	for _, t := range tests {
		loc := locvh.Localization{Values: map[string]locvh.LocalizationValue{}}
		if t.LocID != nil {
			if entry, ok := locMap[*t.LocID]; ok {
				loc = entry
			}
		}

		active := "Active"
		if t.Active != "Y" {
			active = "Not active"
		}
		orderable := "Orderable"
		if t.Orderable != "true" {
			orderable = "Not orderable"
		}

		item := testvh.TestCatalog{
			ID:               t.TestID,
			Localization:     loc,
			TestUnit:         t.SectionName,
			SampleType:       t.SampleType,
			Panel:            "None",
			ResultType:       "",
			Active:           active,
			Orderable:        orderable,
			Uom:              t.UomName,
			SignificantDigits: "n/a",
			TestSortOrder:    t.SortOrder,
		}
		if t.Loinc != nil {
			v := *t.Loinc
			item.Loinc = &v
		}
		catalogList = append(catalogList, item)

		if !seenSections[t.SectionName] {
			seenSections[t.SectionName] = true
			sectionList = append(sectionList, t.SectionName)
		}
	}

	return &form.TestCatalogForm{
		FormName:        "testCatalogForm",
		TestCatalogList: catalogList,
		TestSectionList: sectionList,
	}, nil
}

// fetchTests returns all tests with their section names and basic fields for sorting.
// Uses a LATERAL subquery to pick the first sample type per test (mirrors Java's
// typeOfSampleTests.get(0)).
// GORM Raw + Scan replaces manual rows.Next() / rows.Scan() cursor iteration.
func (s *TestCatalogService) fetchTests() ([]testRow, error) {
	var rows []testRow
	result := s.DB.Raw(`
		SELECT
		    t.id::text                                                              AS test_id,
		    COALESCE(ts_lv.value, ts.name, '')                                     AS section_name,
		    COALESCE(t.sort_order, ?)::bigint                                       AS sort_order,
		    t.name_localization_id::text                                            AS name_localization_id,
		    COALESCE(tos_lv.value, tos.description, tos.local_abbrev, 'n/a')       AS sample_type,
		    COALESCE(t.is_active, 'N')                                              AS is_active,
		    COALESCE(t.orderable::text, 'false')                                    AS orderable,
		    t.loinc                                                                 AS loinc,
		    COALESCE(uom.name, 'n/a')                                              AS uom_name
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
	return rows, result.Error
}

// fetchLocalizations builds a localizationID → Localization map for the given IDs.
// GORM Raw handles the IN ? clause with a slice natively — no manual placeholder
// building required.
func (s *TestCatalogService) fetchLocalizations(ids []string) (map[string]locvh.Localization, error) {
	if len(ids) == 0 {
		return map[string]locvh.Localization{}, nil
	}

	var rows []locRow
	result := s.DB.Raw(`
		SELECT lv.localization_id::text                              AS localization_id,
		       l.description                                         AS description,
		       EXTRACT(EPOCH FROM l.lastupdated)::bigint * 1000      AS lastupdated_ms,
		       lv.id::text                                           AS lv_id,
		       lv.locale                                             AS locale,
		       lv.value                                              AS value
		FROM clinlims.localization l
		JOIN clinlims.localization_value lv ON lv.localization_id = l.id
		WHERE l.id IN ?`, ids).Scan(&rows)
	if result.Error != nil {
		return nil, result.Error
	}

	// Aggregate the flat rows into a localizationID → Localization map.
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
