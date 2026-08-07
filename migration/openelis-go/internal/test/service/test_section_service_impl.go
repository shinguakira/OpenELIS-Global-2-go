// Package service ports org.openelisglobal.test.service — TestSectionServiceImpl.
// Folder layout mirrors the Java source during migration.
package service

import (
	"sort"
	"strings"

	"openelis-go/internal/test/daoimpl"
	"openelis-go/internal/test/valueholder"
)

// TestSectionService ports TestSectionServiceImpl — delegates to the DAO and
// applies the sort order that Java applies after fetching.
type TestSectionService struct {
	DAO *daoimpl.TestSectionDAOImpl
}

// GetAllTestSections mirrors TestSectionServiceImpl.getAllTestSections().
// Java sorts via options.sort((a,b) -> a.name.compareToIgnoreCase(b.name))
// inside TestCatalogEditorRestController.listLabUnits().
// The sort is applied here so callers get a consistent order.
func (s *TestSectionService) GetAllTestSections() ([]valueholder.TestSection, error) {
	list, err := s.DAO.GetAll()
	if err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool {
		return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)
	})
	return list, nil
}
