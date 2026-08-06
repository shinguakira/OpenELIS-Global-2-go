// Package service ports the data-assembly logic of
// org.openelisglobal.testconfiguration.controller.rest.TestCatalogRestController.createTestList().
// In Java that logic lives inline in the controller; here it is separated into
// a service to respect the layered architecture.
// Folder layout mirrors the Java source during migration.
package service

import (
	"database/sql"
	"math"
	"sort"
	"strconv"

	locvh "openelis-go/internal/localization/valueholder"
	"openelis-go/internal/testconfiguration/form"
	testvh "openelis-go/internal/test/valueholder"
)

// testRow is the raw DB projection from fetchTests — internal to the service.
type testRow struct {
	testID      string
	sectionName string
	sortOrder   int64
	locID       sql.NullString
	sampleType  string
	active      string
	orderable   string
	loinc       sql.NullString
	uomName     string
}

// TestCatalogService assembles the TestCatalogForm from the DB.
// Mirrors the read path of TestCatalogRestController.showTestCatalog() +
// createTestList().
type TestCatalogService struct {
	DB *sql.DB
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
		if t.locID.Valid && !seenLoc[t.locID.String] {
			locIDs = append(locIDs, t.locID.String)
			seenLoc[t.locID.String] = true
		}
	}
	locMap, err := s.fetchLocalizations(locIDs)
	if err != nil {
		return nil, err
	}

	// Sort by (sectionName, sampleType, testSortOrder) — mirrors Java Comparator.
	sort.SliceStable(tests, func(i, j int) bool {
		if tests[i].sectionName != tests[j].sectionName {
			return tests[i].sectionName < tests[j].sectionName
		}
		if tests[i].sampleType != tests[j].sampleType {
			return tests[i].sampleType < tests[j].sampleType
		}
		return tests[i].sortOrder < tests[j].sortOrder
	})

	catalogList := make([]testvh.TestCatalog, 0, len(tests))
	sectionList := []string{}
	seenSections := map[string]bool{}

	for _, t := range tests {
		loc := locvh.Localization{Values: map[string]locvh.LocalizationValue{}}
		if t.locID.Valid {
			if entry, ok := locMap[t.locID.String]; ok {
				loc = entry
			}
		}

		active := "Active"
		if t.active != "Y" {
			active = "Not active"
		}
		orderable := "Orderable"
		if t.orderable != "true" {
			orderable = "Not orderable"
		}

		item := testvh.TestCatalog{
			ID:                t.testID,
			Localization:      loc,
			TestUnit:          t.sectionName,
			SampleType:        t.sampleType,
			Panel:             "None",
			ResultType:        "",
			Active:            active,
			Orderable:         orderable,
			Uom:               t.uomName,
			SignificantDigits:  "n/a",
			TestSortOrder:     t.sortOrder,
		}
		if t.loinc.Valid {
			v := t.loinc.String
			item.Loinc = &v
		}
		catalogList = append(catalogList, item)

		if !seenSections[t.sectionName] {
			seenSections[t.sectionName] = true
			sectionList = append(sectionList, t.sectionName)
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
func (s *TestCatalogService) fetchTests() ([]testRow, error) {
	rows, err := s.DB.Query(`
		SELECT
		    t.id::text,
		    COALESCE(ts_lv.value, ts.name, '') AS section_name,
		    COALESCE(t.sort_order, $1)::bigint  AS sort_order,
		    t.name_localization_id::text,
		    COALESCE(tos_lv.value, tos.description, tos.local_abbrev, 'n/a') AS sample_type,
		    COALESCE(t.is_active, 'N'),
		    COALESCE(t.orderable::text, 'false'),
		    t.loinc,
		    COALESCE(uom.name, 'n/a') AS uom_name
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
		math.MaxInt32)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []testRow
	for rows.Next() {
		var r testRow
		if err := rows.Scan(
			&r.testID,
			&r.sectionName,
			&r.sortOrder,
			&r.locID,
			&r.sampleType,
			&r.active,
			&r.orderable,
			&r.loinc,
			&r.uomName,
		); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

// fetchLocalizations builds a localizationID → Localization map for the given IDs.
func (s *TestCatalogService) fetchLocalizations(ids []string) (map[string]locvh.Localization, error) {
	if len(ids) == 0 {
		return map[string]locvh.Localization{}, nil
	}

	placeholders := make([]any, len(ids))
	paramSQL := "$1"
	for i, id := range ids {
		placeholders[i] = id
		if i > 0 {
			paramSQL += ", $" + strconv.Itoa(i+1)
		}
	}

	q := `SELECT lv.localization_id::text,
		         l.description,
		         EXTRACT(EPOCH FROM l.lastupdated)::bigint * 1000,
		         lv.id::text,
		         lv.locale,
		         lv.value
		  FROM clinlims.localization l
		  JOIN clinlims.localization_value lv ON lv.localization_id = l.id
		  WHERE l.id IN (` + paramSQL + `)`

	rows, err := s.DB.Query(q, placeholders...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]locvh.Localization{}
	for rows.Next() {
		var locID, lvID, locale, value string
		var description sql.NullString
		var lastupdatedMs sql.NullInt64
		if err := rows.Scan(&locID, &description, &lastupdatedMs, &lvID, &locale, &value); err != nil {
			return nil, err
		}
		entry, ok := result[locID]
		if !ok {
			entry = locvh.Localization{
				ID:     locID,
				Values: map[string]locvh.LocalizationValue{},
			}
			if description.Valid {
				entry.Description = description.String
			}
			if lastupdatedMs.Valid {
				ms := lastupdatedMs.Int64
				entry.Lastupdated = &ms
			}
		}
		entry.Values[locale] = locvh.LocalizationValue{
			ID:     lvID,
			Locale: locale,
			Value:  value,
		}
		result[locID] = entry
	}
	return result, rows.Err()
}
