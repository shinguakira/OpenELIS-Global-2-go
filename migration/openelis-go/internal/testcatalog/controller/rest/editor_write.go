package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	authmw "openelis-go/internal/auth/middleware"
	"openelis-go/internal/common/web"
	"openelis-go/internal/testcatalog/form"
	"openelis-go/internal/testcatalog/service"
)

// EditorWriteRestController mirrors the section-save half of
// TestCatalogEditorRestController. Same class, so the same class-level
// @PreAuthorize("hasRole('ADMIN')") covers every route below.
type EditorWriteRestController struct {
	Service *service.EditorWriteService
}

// WriteRoutes registers the storage, terminology, display-order and panel
// endpoints — the four sections whose saves do not touch result definitions.
func WriteRoutes(mux *http.ServeMux, ctrl *EditorWriteRestController) {
	reg := func(method, path string, h http.HandlerFunc) {
		web.Register(mux, method, path, authmw.RequireAdmin(h))
	}

	reg("GET", "rest/test-catalog/tests/{testId}/storage", ctrl.getStorage)
	reg("PUT", "rest/test-catalog/tests/{testId}/storage", ctrl.saveStorage)
	reg("PUT", "rest/test-catalog/group/storage", ctrl.saveGroupStorage)

	reg("GET", "rest/test-catalog/tests/{testId}/terminology", ctrl.getTerminology)
	reg("PUT", "rest/test-catalog/tests/{testId}/terminology", ctrl.saveTerminology)

	reg("GET", "rest/test-catalog/sample-types/{sampleTypeId}/test-order", ctrl.getTestOrder)
	reg("PUT", "rest/test-catalog/sample-types/{sampleTypeId}/test-order", ctrl.saveTestOrder)

	reg("GET", "rest/test-catalog/tests/{testId}/panels", ctrl.getTestPanels)
	reg("PUT", "rest/test-catalog/tests/{testId}/panels", ctrl.saveTestPanels)
	reg("GET", "rest/test-catalog/panels/{panelId}/test-order", ctrl.getPanelTestOrder)
	reg("POST", "rest/test-catalog/panels", ctrl.createPanel)
}

// write maps the service's outcome onto the status codes the Java handlers
// answer with: 404 for a missing test or sample type, 422 for a rejected body,
// 500 for anything else — which for POST /panels is every valid request.
func write(w http.ResponseWriter, body any, err error) {
	switch {
	case err == nil:
		web.WriteJSON(w, http.StatusOK, body)
	case errors.Is(err, service.ErrNotFound):
		w.WriteHeader(http.StatusNotFound)
	case errors.Is(err, service.ErrUnprocessable):
		w.WriteHeader(http.StatusUnprocessableEntity)
	default:
		web.WriteJSON(w, http.StatusInternalServerError, web.ServletError(http.StatusInternalServerError))
	}
}

// decode reads a JSON body. Spring answers a malformed one with its own 400
// before the handler runs.
func decode(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		web.WriteJSON(w, http.StatusBadRequest, web.ServletError(http.StatusBadRequest))
		return false
	}
	return true
}

func actingUser(r *http.Request) int64 {
	if p, ok := authmw.FromContext(r.Context()); ok {
		return p.SystemUserID
	}
	return 0
}

func (c *EditorWriteRestController) getStorage(w http.ResponseWriter, r *http.Request) {
	dto, err := c.Service.GetStorage(r.PathValue("testId"))
	write(w, dto, err)
}

func (c *EditorWriteRestController) saveStorage(w http.ResponseWriter, r *http.Request) {
	var body form.StorageDTO
	if !decode(w, r, &body) {
		return
	}
	dto, err := c.Service.SaveStorage(r.PathValue("testId"), body, actingUser(r))
	write(w, dto, err)
}

// saveGroupStorage answers with NO BODY on success — ResponseEntity.ok().build().
func (c *EditorWriteRestController) saveGroupStorage(w http.ResponseWriter, r *http.Request) {
	var body form.GroupStorageUpdate
	if !decode(w, r, &body) {
		return
	}
	err := c.Service.SaveGroupStorage(body, actingUser(r))
	switch {
	case err == nil:
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, service.ErrUnprocessable):
		w.WriteHeader(http.StatusUnprocessableEntity)
	default:
		web.WriteJSON(w, http.StatusInternalServerError, web.ServletError(http.StatusInternalServerError))
	}
}

func (c *EditorWriteRestController) getTerminology(w http.ResponseWriter, r *http.Request) {
	dto, err := c.Service.GetTerminology(r.PathValue("testId"))
	write(w, dto, err)
}

func (c *EditorWriteRestController) saveTerminology(w http.ResponseWriter, r *http.Request) {
	var body form.TerminologyResponse
	if !decode(w, r, &body) {
		return
	}
	dto, err := c.Service.SaveTerminology(r.PathValue("testId"), body, actingUser(r))
	write(w, dto, err)
}

func (c *EditorWriteRestController) getTestOrder(w http.ResponseWriter, r *http.Request) {
	dto, err := c.Service.GetTestOrder(r.PathValue("sampleTypeId"))
	write(w, dto, err)
}

func (c *EditorWriteRestController) saveTestOrder(w http.ResponseWriter, r *http.Request) {
	var body form.DisplayOrderUpdate
	if !decode(w, r, &body) {
		return
	}
	dto, err := c.Service.SaveTestOrder(r.PathValue("sampleTypeId"), body)
	write(w, dto, err)
}

func (c *EditorWriteRestController) getTestPanels(w http.ResponseWriter, r *http.Request) {
	dto, err := c.Service.GetTestPanels(r.PathValue("testId"))
	write(w, dto, err)
}

func (c *EditorWriteRestController) saveTestPanels(w http.ResponseWriter, r *http.Request) {
	var body form.PanelMembershipUpdate
	if !decode(w, r, &body) {
		return
	}
	dto, err := c.Service.SaveTestPanels(r.PathValue("testId"), body)
	write(w, dto, err)
}

func (c *EditorWriteRestController) getPanelTestOrder(w http.ResponseWriter, r *http.Request) {
	dto, err := c.Service.GetPanelTestOrder(r.PathValue("panelId"))
	write(w, dto, err)
}

// createPanel answers 201 on success — which it never reaches, because the
// insert cannot satisfy panel.name_localization_id. See the DAO.
func (c *EditorWriteRestController) createPanel(w http.ResponseWriter, r *http.Request) {
	var body form.CreatePanelRequest
	if !decode(w, r, &body) {
		return
	}
	dto, err := c.Service.CreatePanel(body)
	switch {
	case err == nil:
		web.WriteJSON(w, http.StatusCreated, dto)
	case errors.Is(err, service.ErrUnprocessable):
		w.WriteHeader(http.StatusUnprocessableEntity)
	default:
		web.WriteJSON(w, http.StatusInternalServerError, web.ServletError(http.StatusInternalServerError))
	}
}
