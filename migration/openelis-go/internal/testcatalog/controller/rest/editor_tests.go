package rest

import (
	"errors"
	"net/http"

	authmw "openelis-go/internal/auth/middleware"
	"openelis-go/internal/common/web"
	"openelis-go/internal/testcatalog/form"
	"openelis-go/internal/testcatalog/service"
)

// EditorTestRestController mirrors the test-level half of
// TestCatalogEditorRestController, plus TestCatalogActivationRestController —
// two Java classes, both @PreAuthorize("hasRole('ADMIN')"), both under
// /rest/test-catalog.
type EditorTestRestController struct {
	Service *service.EditorTestService
}

// TestRoutes registers the create flow, Basic Info, Sample & Results and
// activation.
func TestRoutes(mux *http.ServeMux, ctrl *EditorTestRestController) {
	reg := func(method, path string, h http.HandlerFunc) {
		web.Register(mux, method, path, authmw.RequireAdmin(h))
	}

	reg("POST", "rest/test-catalog/tests", ctrl.createTest)

	reg("GET", "rest/test-catalog/tests/{testId}/basic-info", ctrl.getBasicInfo)
	reg("PUT", "rest/test-catalog/tests/{testId}/basic-info", ctrl.saveBasicInfo)

	reg("GET", "rest/test-catalog/tests/{testId}/sample-results", ctrl.getSampleResults)
	reg("PUT", "rest/test-catalog/tests/{testId}/sample-results", ctrl.saveSampleResults)
	reg("POST", "rest/test-catalog/tests/{testId}/sample-results/copy-from/{sourceId}", ctrl.copySampleResults)

	reg("POST", "rest/test-catalog/tests/{testId}/activate", ctrl.activate)
}

// createTest answers 201 with the new id, 409 when the code is taken, and 422
// for a body missing any of the five required fields.
func (c *EditorTestRestController) createTest(w http.ResponseWriter, r *http.Request) {
	var body form.CreateTestRequest
	if !decode(w, r, &body) {
		return
	}
	created, err := c.Service.CreateTest(body, actingUser(r))
	switch {
	case err == nil:
		web.WriteJSON(w, http.StatusCreated, created)
	case errors.Is(err, service.ErrUnprocessable):
		w.WriteHeader(http.StatusUnprocessableEntity)
	case errors.Is(err, service.ErrCodeInUse):
		w.WriteHeader(http.StatusConflict)
	default:
		web.WriteJSON(w, http.StatusInternalServerError, web.ServletError(http.StatusInternalServerError))
	}
}

func (c *EditorTestRestController) getBasicInfo(w http.ResponseWriter, r *http.Request) {
	dto, err := c.Service.GetBasicInfo(r.PathValue("testId"))
	write(w, dto, err)
}

func (c *EditorTestRestController) saveBasicInfo(w http.ResponseWriter, r *http.Request) {
	var body form.BasicInfo
	if !decode(w, r, &body) {
		return
	}
	dto, err := c.Service.SaveBasicInfo(r.PathValue("testId"), body, actingUser(r))
	write(w, dto, err)
}

func (c *EditorTestRestController) getSampleResults(w http.ResponseWriter, r *http.Request) {
	dto, err := c.Service.GetSampleResults(r.PathValue("testId"))
	write(w, dto, err)
}

func (c *EditorTestRestController) saveSampleResults(w http.ResponseWriter, r *http.Request) {
	var body form.SampleResults
	if !decode(w, r, &body) {
		return
	}
	dto, err := c.Service.SaveSampleResults(r.PathValue("testId"), body, actingUser(r))
	write(w, dto, err)
}

func (c *EditorTestRestController) copySampleResults(w http.ResponseWriter, r *http.Request) {
	dto, err := c.Service.CopySampleResults(
		r.PathValue("testId"), r.PathValue("sourceId"), actingUser(r))
	write(w, dto, err)
}

// activate answers 409 WITH the coverage report when a gap is unacknowledged,
// and 200 with the same report either way otherwise.
func (c *EditorTestRestController) activate(w http.ResponseWriter, r *http.Request) {
	// The body is @RequestBody(required = false), so an absent one is not an
	// error — it simply carries no acknowledgment.
	var body *form.ActivateRequest
	var parsed form.ActivateRequest
	if r.ContentLength != 0 {
		if !decode(w, r, &parsed) {
			return
		}
		body = &parsed
	}

	report, err := c.Service.Activate(r.PathValue("testId"), body, actingUser(r))
	var gap *service.ErrCoverageGap
	switch {
	case err == nil:
		web.WriteJSON(w, http.StatusOK, report)
	case errors.As(err, &gap):
		web.WriteJSON(w, http.StatusConflict, gap.Report)
	case errors.Is(err, service.ErrNotFound):
		w.WriteHeader(http.StatusNotFound)
	default:
		web.WriteJSON(w, http.StatusInternalServerError, web.ServletError(http.StatusInternalServerError))
	}
}

// RangeRoutes registers the Reference Ranges pair and the group save.
//
// Separate from TestRoutes only to keep the clinical surface visible at the
// call site: these are the writes that decide whether a result reads as normal.
func RangeRoutes(mux *http.ServeMux, ctrl *EditorTestRestController) {
	reg := func(method, path string, h http.HandlerFunc) {
		web.Register(mux, method, path, authmw.RequireAdmin(h))
	}
	reg("GET", "rest/test-catalog/tests/{testId}/ranges", ctrl.getRanges)
	reg("PUT", "rest/test-catalog/tests/{testId}/ranges", ctrl.saveRanges)
	reg("PUT", "rest/test-catalog/group/ranges", ctrl.saveGroupRanges)
}

func (c *EditorTestRestController) getRanges(w http.ResponseWriter, r *http.Request) {
	dto, err := c.Service.GetRanges(r.PathValue("testId"))
	write(w, dto, err)
}

func (c *EditorTestRestController) saveRanges(w http.ResponseWriter, r *http.Request) {
	var body form.RangesResponse
	if !decode(w, r, &body) {
		return
	}
	dto, err := c.Service.SaveRanges(r.PathValue("testId"), body, actingUser(r))
	write(w, dto, err)
}

// saveGroupRanges answers with NO BODY on success — ResponseEntity.ok().build().
func (c *EditorTestRestController) saveGroupRanges(w http.ResponseWriter, r *http.Request) {
	var body form.GroupRangesUpdate
	if !decode(w, r, &body) {
		return
	}
	err := c.Service.SaveGroupRanges(body, actingUser(r))
	switch {
	case err == nil:
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, service.ErrUnprocessable):
		w.WriteHeader(http.StatusUnprocessableEntity)
	default:
		web.WriteJSON(w, http.StatusInternalServerError, web.ServletError(http.StatusInternalServerError))
	}
}
