// Package service ports org.openelisglobal.testcatalog.service — specifically
// the aggregation TestCatalogEditorRestController.java does inline in Java
// (it calls TestSectionService/TypeOfSampleService/PanelService directly from
// the controller). Per constitution.md Layer III/IV, that orchestration and
// the resulting DTO shaping belong in a service, not the controller — this
// package is that service, composed over the three underlying domain services
// rather than a DAO (its data isn't its own; it's a cross-domain view).
// Folder layout mirrors the Java source during migration.
package service

import (
	"strconv"

	"openelis-go/internal/testcatalog/form"

	panelservice "openelis-go/internal/panel/service"
	panelvh "openelis-go/internal/panel/valueholder"
	testservice "openelis-go/internal/test/service"
	testvh "openelis-go/internal/test/valueholder"
	tosservice "openelis-go/internal/typeofsample/service"
	tosvh "openelis-go/internal/typeofsample/valueholder"
)

// TestCatalogEditorService mirrors the three read methods
// TestCatalogEditorRestController delegates to.
type TestCatalogEditorService struct {
	TestSectionService  *testservice.TestSectionService
	TypeOfSampleService *tosservice.TypeOfSampleService
	PanelService        *panelservice.PanelService
}

// GetLabUnits mirrors GET /rest/test-catalog/lab-units.
func (s *TestCatalogEditorService) GetLabUnits() ([]form.IdNameDTO, error) {
	sections, err := s.TestSectionService.GetAllTestSections()
	if err != nil {
		return nil, err
	}
	dtos := make([]form.IdNameDTO, len(sections))
	for i, ts := range sections {
		dtos[i] = sectionToDTO(ts)
	}
	return dtos, nil
}

// GetSampleTypes mirrors GET /rest/test-catalog/sample-types.
func (s *TestCatalogEditorService) GetSampleTypes() ([]form.IdNameDTO, error) {
	samples, err := s.TypeOfSampleService.GetAllTypeOfSamplesSortOrdered()
	if err != nil {
		return nil, err
	}
	dtos := make([]form.IdNameDTO, len(samples))
	for i, t := range samples {
		dtos[i] = sampleTypeToDTO(t)
	}
	return dtos, nil
}

// GetPanels mirrors GET /rest/test-catalog/panels.
func (s *TestCatalogEditorService) GetPanels() ([]form.IdNameDTO, error) {
	panels, err := s.PanelService.GetAllActivePanels()
	if err != nil {
		return nil, err
	}
	dtos := make([]form.IdNameDTO, len(panels))
	for i, p := range panels {
		dtos[i] = panelToDTO(p)
	}
	return dtos, nil
}

func sectionToDTO(ts testvh.TestSection) form.IdNameDTO {
	return form.IdNameDTO{ID: strconv.FormatInt(ts.ID, 10), Name: ts.Name}
}

func sampleTypeToDTO(t tosvh.TypeOfSample) form.IdNameDTO {
	return form.IdNameDTO{ID: strconv.FormatInt(t.ID, 10), Name: t.Name}
}

func panelToDTO(p panelvh.Panel) form.IdNameDTO {
	return form.IdNameDTO{ID: strconv.FormatInt(p.ID, 10), Name: p.PanelName}
}
