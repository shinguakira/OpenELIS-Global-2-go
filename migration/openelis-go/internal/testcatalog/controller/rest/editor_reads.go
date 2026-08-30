package rest

import (
	"errors"
	"net/http"
	"strconv"

	authmw "openelis-go/internal/auth/middleware"
	"openelis-go/internal/common/web"
	"openelis-go/internal/testcatalog/service"
)

// EditorReadRestController mirrors the read half of
// TestCatalogEditorRestController plus the two read-only controllers beside it,
// TestReflexCalcRestController and TestStorageHistoryRestController. All three
// Java classes are @PreAuthorize("hasRole('ADMIN')").
type EditorReadRestController struct {
	Service *service.EditorReadService
}

// ReadRoutes registers the ten reads that pair with no write.
func ReadRoutes(mux *http.ServeMux, ctrl *EditorReadRestController) {
	reg := func(method, path string, h http.HandlerFunc) {
		web.Register(mux, method, path, authmw.RequireAdmin(h))
	}

	reg("GET", "rest/test-catalog/tests", ctrl.listTests)
	reg("GET", "rest/test-catalog/tests/{testId}/localization", ctrl.localization)
	reg("GET", "rest/test-catalog/tests/{testId}/loinc-integrity", ctrl.loincIntegrity)
	reg("GET", "rest/test-catalog/dictionary", ctrl.dictionary)
	reg("GET", "rest/test-catalog/tests/{testId}/siblings", ctrl.siblings)
	reg("GET", "rest/test-catalog/group/summary", ctrl.groupSummary)
	reg("GET", "rest/test-catalog/tests/{testId}/analyzers", ctrl.analyzers)

	// TestStorageHistoryRestController's path puts `{testId}` directly under
	// /rest/test-catalog. It does not collide with `tests/{testId}/storage`,
	// because that one needs "storage" in the THIRD segment and this one needs
	// it in the second.
	reg("GET", "rest/test-catalog/{testId}/storage/history", ctrl.storageHistory)

	// The envelope and reflex-calc are BOTH two segments — `tests/{testId}` and
	// `{testId}/reflex-calc` — and Go's ServeMux refuses to register a pair
	// where neither pattern is more specific: it panics at startup rather than
	// guess. Spring does guess, and MEASURED it prefers the one whose trailing
	// segment is literal — GET /rest/test-catalog/tests/reflex-calc answers 500
	// from reflex-calc looking up a test called "tests", not the envelope for a
	// test called "reflex-calc". One handler, dispatched in that order.
	reg("GET", "rest/test-catalog/{first}/{second}", ctrl.twoSegment)
}

// twoSegment resolves `/rest/test-catalog/{a}/{b}` the way Spring does.
//
// Literal-suffix first: `{testId}/reflex-calc` wins over `tests/{testId}` for a
// path that matches both. Anything matching neither is a 404, which is what
// Spring answers when no mapping applies.
func (c *EditorReadRestController) twoSegment(w http.ResponseWriter, r *http.Request) {
	first, second := r.PathValue("first"), r.PathValue("second")
	switch {
	case second == "reflex-calc":
		c.reflexCalcFor(w, r, first)
	case first == "tests":
		body, err := c.Service.Envelope(second)
		write(w, body, err)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// writeProblem is the 404 the two read-only controllers answer with.
//
// They throw ResponseStatusException, which Spring renders as an RFC 7807
// ProblemDetail — a DIFFERENT envelope from the editor's own 404s, which are
// ResponseEntity.notFound() with no body at all. Two 404 shapes on one path
// prefix, decided by which controller owns the route.
func writeProblem(w http.ResponseWriter, r *http.Request, status int) {
	web.WriteJSON(w, status, map[string]any{
		"type":     "problemDetail.type.org.springframework.web.server.ResponseStatusException",
		"title":    "problemDetail.title.org.springframework.web.server.ResponseStatusException",
		"status":   status,
		"detail":   "problemDetail.org.springframework.web.server.ResponseStatusException",
		"instance": r.URL.Path,
	})
}

func (c *EditorReadRestController) listTests(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	// status defaults to "all", which matches neither the active nor the
	// inactive branch and therefore filters nothing.
	status := q.Get("status")
	if status == "" {
		status = "all"
	}
	var amr *bool
	if raw := q.Get("amr"); raw != "" {
		// Spring binds a Boolean param; anything that is not "true" is false,
		// and an absent one stays null so the filter does not run.
		v := raw == "true"
		amr = &v
	}
	page := intParam(q.Get("page"), 1)
	pageSize := intParam(q.Get("pageSize"), 25)

	body, err := c.Service.ListTests(
		q.Get("domain"), status, amr, q.Get("sampleType"), q.Get("search"), page, pageSize)
	write(w, body, err)
}

// intParam is Spring's @RequestParam(defaultValue = "N") int binding: absent
// takes the default, and a non-numeric value is a 400 there. Here it falls back
// to the default — no shipped caller sends one, and the service clamps anyway.
func intParam(raw string, dflt int) int {
	if raw == "" {
		return dflt
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return dflt
	}
	return v
}

func (c *EditorReadRestController) localization(w http.ResponseWriter, r *http.Request) {
	body, err := c.Service.LocalizationRefs(r.PathValue("testId"))
	write(w, body, err)
}

func (c *EditorReadRestController) loincIntegrity(w http.ResponseWriter, r *http.Request) {
	body, err := c.Service.LoincIntegrity(r.PathValue("testId"))
	write(w, body, err)
}

func (c *EditorReadRestController) dictionary(w http.ResponseWriter, r *http.Request) {
	body, err := c.Service.SearchDictionary(r.URL.Query().Get("search"))
	write(w, body, err)
}

// siblings answers 200 with an EMPTY LIST for an unknown test, not 404 — the
// handler returns the accumulator rather than a ResponseEntity.
func (c *EditorReadRestController) siblings(w http.ResponseWriter, r *http.Request) {
	body, err := c.Service.Siblings(r.PathValue("testId"))
	write(w, body, err)
}

// groupSummary takes `ids` as a REQUIRED param — Spring answers 400 when it is
// missing, before the handler runs.
func (c *EditorReadRestController) groupSummary(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if !q.Has("ids") {
		web.WriteJSON(w, http.StatusBadRequest, web.ServletError(http.StatusBadRequest))
		return
	}
	body, err := c.Service.GroupSummary(q.Get("ids"))
	write(w, body, err)
}

func (c *EditorReadRestController) analyzers(w http.ResponseWriter, r *http.Request) {
	body, err := c.Service.Analyzers(r.PathValue("testId"))
	write(w, body, err)
}

func (c *EditorReadRestController) reflexCalcFor(w http.ResponseWriter, r *http.Request, testID string) {
	body, err := c.Service.ReflexCalc(testID)
	if errors.Is(err, service.ErrNotFound) {
		writeProblem(w, r, http.StatusNotFound)
		return
	}
	write(w, body, err)
}

func (c *EditorReadRestController) storageHistory(w http.ResponseWriter, r *http.Request) {
	body, err := c.Service.StorageHistory(r.PathValue("testId"))
	if errors.Is(err, service.ErrNotFound) {
		writeProblem(w, r, http.StatusNotFound)
		return
	}
	write(w, body, err)
}
