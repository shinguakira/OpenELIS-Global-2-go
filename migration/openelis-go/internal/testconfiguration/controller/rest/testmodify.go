package rest

import (
	"encoding/json"
	"net/http"

	authmw "openelis-go/internal/auth/middleware"
	"openelis-go/internal/common/web"
	"openelis-go/internal/testconfiguration/form"
	"openelis-go/internal/testconfiguration/service"
)

// TestModifyRestController mirrors TestModifyEntryRestController, which is
// @PreAuthorize("hasRole('ADMIN')") like the rest of the package.
type TestModifyRestController struct {
	Service *service.TestModifyService
}

// TestModifyRoutes registers the GET/POST pair.
func TestModifyRoutes(mux *http.ServeMux, ctrl *TestModifyRestController) {
	web.Register(mux, "GET", "rest/TestModifyEntry", authmw.RequireAdmin(ctrl.show))
	web.Register(mux, "POST", "rest/TestModifyEntry", authmw.RequireAdmin(ctrl.update))
}

// show takes two OPTIONAL query parameters. Neither is required and neither is
// validated: an unknown id simply matches no test and the catalogue comes back
// empty, the same as the unfiltered load.
func (c *TestModifyRestController) show(w http.ResponseWriter, r *http.Request) {
	f, err := c.Service.Form(
		r.URL.Query().Get("sampleType"),
		r.URL.Query().Get("testSection"))
	writeForm(w, f, err)
}

func (c *TestModifyRestController) update(w http.ResponseWriter, r *http.Request) {
	var post form.TestModifyEntryPost
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		web.WriteJSON(w, http.StatusBadRequest, web.ServletError(http.StatusBadRequest))
		return
	}
	f, err := c.Service.Update(post, actingUser(r))
	writeForm(w, f, err)
}
