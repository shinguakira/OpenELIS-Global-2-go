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
// its status_of_sample row by (status_type, DB name). The localized label is
// resolved at runtime via StatusService (display_key → message_en.properties),
// mirroring Java's BaseObject.getLocalizedName() path.
type statusEntry struct {
	statusType   string
	internalName string
}

// getAnalysisStatusTypes source (DisplayListController.java:459).
var analysisStatuses = []statusEntry{
	{"ANALYSIS", "Not Tested"},
	{"ANALYSIS", "Test Canceled"},
	{"ANALYSIS", "Technical Acceptance"},
	{"ANALYSIS", "Technical Rejected"},
	{"ANALYSIS", "Biologist Rejection"},
}

// getSampleStatusTypes source (DisplayListController.java:483) — uses OrderStatus.
var sampleStatuses = []statusEntry{
	{"ORDER", "Test Entered"},
	{"ORDER", "Testing Started"},
}

// statusList builds the list: a leading {"0",""} then each entry as
// {id = status id from the DB, value = label from display_key → message_en.properties}.
func statusList(svc *services.StatusService, entries []statusEntry) []util.IdValuePair {
	list := []util.IdValuePair{util.NewIdValuePair("0", "")}
	for _, e := range entries {
		list = append(list, util.NewIdValuePair(
			svc.IDByName(e.statusType, e.internalName),
			svc.LabelByName(e.statusType, e.internalName),
		))
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
