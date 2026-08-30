package rest

import (
	"encoding/json"
	"net/http"

	authmw "openelis-go/internal/auth/middleware"
	"openelis-go/internal/common/web"
	"openelis-go/internal/testconfiguration/form"
	"openelis-go/internal/testconfiguration/service"
)

// CreateRestController mirrors MethodCreateRestController,
// TestSectionCreateRestController and SampleTypeCreateRestController.
//
// All three are @PreAuthorize("hasRole('ADMIN')").
type CreateRestController struct {
	Service *service.CreateService
}

// CreateRoutes registers the three GET/POST pairs.
func CreateRoutes(mux *http.ServeMux, ctrl *CreateRestController) {
	pairs := []struct {
		path string
		get  http.HandlerFunc
		post http.HandlerFunc
	}{
		{"MethodCreate", ctrl.methodForm, ctrl.methodCreate},
		{"TestSectionCreate", ctrl.testSectionForm, ctrl.testSectionCreate},
		{"SampleTypeCreate", ctrl.sampleTypeForm, ctrl.sampleTypeCreate},
		{"PanelCreate", ctrl.panelForm, ctrl.panelCreate},
	}
	for _, p := range pairs {
		web.Register(mux, "GET", "rest/"+p.path, authmw.RequireAdmin(p.get))
		web.Register(mux, "POST", "rest/"+p.path, authmw.RequireAdmin(p.post))
	}
}

func decodeCreate(w http.ResponseWriter, r *http.Request) (form.CreatePost, bool) {
	var post form.CreatePost
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		web.WriteJSON(w, http.StatusBadRequest, web.ServletError(http.StatusBadRequest))
		return post, false
	}
	return post, true
}

func (c *CreateRestController) methodForm(w http.ResponseWriter, r *http.Request) {
	f, err := c.Service.MethodForm()
	writeForm(w, f, err)
}

func (c *CreateRestController) methodCreate(w http.ResponseWriter, r *http.Request) {
	post, ok := decodeCreate(w, r)
	if !ok {
		return
	}
	f, err := c.Service.CreateMethod(post, actingUser(r))
	writeForm(w, f, err)
}

func (c *CreateRestController) testSectionForm(w http.ResponseWriter, r *http.Request) {
	f, err := c.Service.TestSectionForm()
	writeForm(w, f, err)
}

func (c *CreateRestController) testSectionCreate(w http.ResponseWriter, r *http.Request) {
	post, ok := decodeCreate(w, r)
	if !ok {
		return
	}
	f, err := c.Service.CreateTestSection(post, actingUser(r))
	writeForm(w, f, err)
}

func (c *CreateRestController) sampleTypeForm(w http.ResponseWriter, r *http.Request) {
	f, err := c.Service.SampleTypeForm()
	writeForm(w, f, err)
}

func (c *CreateRestController) sampleTypeCreate(w http.ResponseWriter, r *http.Request) {
	post, ok := decodeCreate(w, r)
	if !ok {
		return
	}
	f, err := c.Service.CreateSampleType(post, actingUser(r))
	writeForm(w, f, err)
}

// actingUser is getSysUserId(request) — the id the entity's audit row is
// attributed to. The ADMIN guard has already refused an unauthenticated caller.
func actingUser(r *http.Request) int64 {
	if p, ok := authmw.FromContext(r.Context()); ok {
		return p.SystemUserID
	}
	return 0
}

func (c *CreateRestController) panelForm(w http.ResponseWriter, r *http.Request) {
	f, err := c.Service.PanelForm()
	writeForm(w, f, err)
}

func (c *CreateRestController) panelCreate(w http.ResponseWriter, r *http.Request) {
	post, ok := decodeCreate(w, r)
	if !ok {
		return
	}
	f, err := c.Service.CreatePanel(post, actingUser(r))
	writeForm(w, f, err)
}
