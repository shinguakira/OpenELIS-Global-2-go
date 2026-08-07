// Package service ports the data-assembly logic of
// org.openelisglobal.testconfiguration.controller.rest.TestCatalogRestController.createTestList().
// In Java that logic lives inline in the controller; here it is separated into
// a service to respect the layered architecture.
// Folder layout mirrors the Java source during migration.
package service

import (
	"sort"

	locvh "openelis-go/internal/localization/valueholder"
	"openelis-go/internal/testconfiguration/daoimpl"
	"openelis-go/internal/testconfiguration/form"
	testvh "openelis-go/internal/test/valueholder"
)

// TestCatalogService assembles the TestCatalogForm.
// Mirrors the read path of TestCatalogRestController.showTestCatalog() +
// createTestList(). All DB access goes through the DAO — this layer holds no SQL.
type TestCatalogService struct {
	DAO *daoimpl.TestCatalogDAOImpl
}

// BuildForm mirrors TestCatalogRestController.showTestCatalog() +
// createTestList(): fetches all tests with their joined metadata, resolves
// localization values in bulk, sorts by (testUnit, sampleType, testSortOrder),
// and builds the form that the controller serialises to JSON.
func (s *TestCatalogService) BuildForm() (*form.TestCatalogForm, error) {
	tests, err := s.DAO.GetAllTestRows()
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

	// Collect unique localization IDs for the bulk fetch — the Go replacement
	// for Java's per-test localizationService call (N+1 → 1 query).
	locIDs := make([]string, 0, len(tests))
	seenLoc := map[string]bool{}
	for _, t := range tests {
		if t.LocID != nil && !seenLoc[*t.LocID] {
			locIDs = append(locIDs, *t.LocID)
			seenLoc[*t.LocID] = true
		}
	}
	locMap, err := s.DAO.GetLocalizationsByIDs(locIDs)
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
			ID:                t.TestID,
			Localization:      loc,
			TestUnit:          t.SectionName,
			SampleType:        t.SampleType,
			Panel:             "None",
			ResultType:        "",
			Active:            active,
			Orderable:         orderable,
			Uom:               t.UomName,
			SignificantDigits: "n/a",
			TestSortOrder:     t.SortOrder,
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
