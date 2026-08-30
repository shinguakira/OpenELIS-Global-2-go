package rest

import (
	"encoding/json"
	"net/http"

	authmw "openelis-go/internal/auth/middleware"
	"openelis-go/internal/common/web"
	"openelis-go/internal/testconfiguration/form"
	"openelis-go/internal/testconfiguration/service"
)

// TestAddRestController mirrors TestAddRestController, which is
// @PreAuthorize("hasRole('ADMIN')") like the rest of this package.
type TestAddRestController struct {
	Service *service.TestAddService
}

// TestAddRoutes registers the GET/POST pair.
func TestAddRoutes(mux *http.ServeMux, ctrl *TestAddRestController) {
	web.Register(mux, "GET", "rest/TestAdd", authmw.RequireAdmin(ctrl.showTestAdd))
	web.Register(mux, "POST", "rest/TestAdd", authmw.RequireAdmin(ctrl.postTestAdd))
}

func (c *TestAddRestController) showTestAdd(w http.ResponseWriter, r *http.Request) {
	f, err := c.Service.Form()
	writeForm(w, f, err)
}

func (c *TestAddRestController) postTestAdd(w http.ResponseWriter, r *http.Request) {
	var post form.TestAddPost
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		web.WriteJSON(w, http.StatusBadRequest, web.ServletError(http.StatusBadRequest))
		return
	}
	f, err := c.Service.Add(post, actingUser(r))
	writeForm(w, f, err)
}
