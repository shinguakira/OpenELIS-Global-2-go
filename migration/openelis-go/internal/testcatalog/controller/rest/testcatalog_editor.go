// Package rest ports org.openelisglobal.testcatalog.controller.rest.TestCatalogEditorRestController
// — three read-only reference endpoints: /lab-units, /sample-types, /panels.
// Folder layout mirrors the Java source during migration.
package rest

import (
	"net/http"
	"strconv"

	"openelis-go/internal/common/web"
	panelsvc "openelis-go/internal/panel/service"
	panelvh "openelis-go/internal/panel/valueholder"
	testsvc "openelis-go/internal/test/service"
	testvh "openelis-go/internal/test/valueholder"
	tossvc "openelis-go/internal/typeofsample/service"
	tosvh "openelis-go/internal/typeofsample/valueholder"
)

// idNameDTO mirrors the IdValuePair the Java controller returns for lab-units,
// sample-types, and panels: {id: string, name: string}.
type idNameDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func sectionToDTO(ts testvh.TestSection) idNameDTO {
	return idNameDTO{ID: strconv.FormatInt(ts.ID, 10), Name: ts.Name}
}

func sampleTypeToDTO(t tosvh.TypeOfSample) idNameDTO {
	return idNameDTO{ID: strconv.FormatInt(t.ID, 10), Name: t.Name}
}

func panelToDTO(p panelvh.Panel) idNameDTO {
	return idNameDTO{ID: strconv.FormatInt(p.ID, 10), Name: p.PanelName}
}

// TestCatalogEditorRestController mirrors TestCatalogEditorRestController.
type TestCatalogEditorRestController struct {
	TestSectionService  *testsvc.TestSectionService
	TypeOfSampleService *tossvc.TypeOfSampleService
	PanelService        *panelsvc.PanelService
}

// Routes registers /rest/test-catalog/lab-units, /sample-types, /panels.
func Routes(mux *http.ServeMux, ctrl *TestCatalogEditorRestController) {
	web.Register(mux, "GET", "rest/test-catalog/lab-units", func(w http.ResponseWriter, r *http.Request) {
		sections, err := ctrl.TestSectionService.GetAllTestSections()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		dtos := make([]idNameDTO, len(sections))
		for i, s := range sections {
			dtos[i] = sectionToDTO(s)
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})

	web.Register(mux, "GET", "rest/test-catalog/sample-types", func(w http.ResponseWriter, r *http.Request) {
		samples, err := ctrl.TypeOfSampleService.GetAllTypeOfSamplesSortOrdered()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		dtos := make([]idNameDTO, len(samples))
		for i, s := range samples {
			dtos[i] = sampleTypeToDTO(s)
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})

	web.Register(mux, "GET", "rest/test-catalog/panels", func(w http.ResponseWriter, r *http.Request) {
		panels, err := ctrl.PanelService.GetAllActivePanels()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		dtos := make([]idNameDTO, len(panels))
		for i, p := range panels {
			dtos[i] = panelToDTO(p)
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})
}
