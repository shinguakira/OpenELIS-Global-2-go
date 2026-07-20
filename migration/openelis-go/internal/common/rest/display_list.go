// Package rest ports org.openelisglobal.common.rest — the DisplayListController
// family of reference/list endpoints. Folder layout mirrors the Java source
// (common/rest) during migration.
package rest

import (
	"net/http"

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
