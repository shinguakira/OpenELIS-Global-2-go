// Package rest ports org.openelisglobal.common.rest — the DisplayListController
// family of reference/list endpoints. Folder layout mirrors the Java source
// (common/rest) during migration.
package rest

import (
	"net/http"

	"openelis-go/internal/common/services"
	"openelis-go/internal/common/util"
	"openelis-go/internal/common/web"
)

// Routes registers the display-list REST endpoints. Mirrors the @GetMapping
// methods on DisplayListController (@RequestMapping("/rest/")).
func Routes(mux *http.ServeMux) {
	web.Register(mux, "GET", "rest/sample-item-status-types", SampleItemStatusTypes)
}

// SampleItemStatusTypes reproduces DisplayListController#getSampleItemStatusTypes
// (DisplayListController.java:505) — a hardcoded 3-item list:
//
//	GET /rest/sample-item-status-types
//	  -> [{"","All"},{"active","Active"},{"disposed","Disposed"}]
func SampleItemStatusTypes(w http.ResponseWriter, r *http.Request) {
	web.WriteJSON(w, http.StatusOK, []util.IdValuePair{
		util.NewIdValuePair("", "All"),
		util.NewIdValuePair("active", "Active"),
		util.NewIdValuePair("disposed", "Disposed"),
	})
}

// statusEntry maps a status (as Java's AnalysisStatus/OrderStatus enum does) to
// its status_of_sample row: internalName is the DB `name` used to resolve the id,
// value is the English localized label (Java gets this from localization; hardcoded
// here for the English baseline — full i18n is a later cross-cutting port).
type statusEntry struct {
	statusType   string
	internalName string
	value        string
}

// getAnalysisStatusTypes source (DisplayListController.java:459).
var analysisStatuses = []statusEntry{
	{"ANALYSIS", "Not Tested", "Not started"},
	{"ANALYSIS", "Test Canceled", "Canceled"},
	{"ANALYSIS", "Technical Acceptance", "Accepted by technician"},
	{"ANALYSIS", "Technical Rejected", "Not accepted by technician"},
	{"ANALYSIS", "Biologist Rejection", "Not accepted by biologist"},
}

// getSampleStatusTypes source (DisplayListController.java:483) — uses OrderStatus.
var sampleStatuses = []statusEntry{
	{"ORDER", "Test Entered", "No tests have been run for this sample"},
	{"ORDER", "Testing Started", "Some tests have been run on this sample"},
}

// statusList builds the list: a leading {"0",""} then each entry as
// {id = status id from the DB, value = localized label} — mirroring the Java.
func statusList(svc *services.StatusService, entries []statusEntry) []util.IdValuePair {
	list := []util.IdValuePair{util.NewIdValuePair("0", "")}
	for _, e := range entries {
		list = append(list, util.NewIdValuePair(svc.IDByName(e.statusType, e.internalName), e.value))
	}
	return list
}

// StatusRoutes registers the DB-backed status-type endpoints (need StatusService).
func StatusRoutes(mux *http.ServeMux, svc *services.StatusService) {
	web.Register(mux, "GET", "rest/analysis-status-types", func(w http.ResponseWriter, r *http.Request) {
		web.WriteJSON(w, http.StatusOK, statusList(svc, analysisStatuses))
	})
	web.Register(mux, "GET", "rest/sample-status-types", func(w http.ResponseWriter, r *http.Request) {
		web.WriteJSON(w, http.StatusOK, statusList(svc, sampleStatuses))
	})
}
