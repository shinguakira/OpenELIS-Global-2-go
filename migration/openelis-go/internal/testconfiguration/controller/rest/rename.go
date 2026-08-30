package rest

import (
	"encoding/json"
	"net/http"

	authmw "openelis-go/internal/auth/middleware"
	"openelis-go/internal/common/web"
	"openelis-go/internal/testconfiguration/form"
	"openelis-go/internal/testconfiguration/service"
)

// RenameRestController mirrors MethodRenameEntryRestController,
// PanelRenameEntryRestController, SampleTypeRenameEntryRestController and
// TestSectionRenameEntryRestController — four Java classes with one body
// between them.
//
// All four are @PreAuthorize("hasRole('ADMIN')").
type RenameRestController struct {
	Service *service.RenameService
}

// RenameRoutes registers the four GET/POST pairs.
func RenameRoutes(mux *http.ServeMux, ctrl *RenameRestController) {
	pairs := []struct {
		path string
		get  http.HandlerFunc
		post http.HandlerFunc
	}{
		{"MethodRenameEntry", ctrl.methodForm, ctrl.methodRename},
		{"PanelRenameEntry", ctrl.panelForm, ctrl.panelRename},
		{"SampleTypeRenameEntry", ctrl.sampleTypeForm, ctrl.sampleTypeRename},
		{"TestSectionRenameEntry", ctrl.testSectionForm, ctrl.testSectionRename},
	}
	for _, p := range pairs {
		web.Register(mux, "GET", "rest/"+p.path, authmw.RequireAdmin(p.get))
		web.Register(mux, "POST", "rest/"+p.path, authmw.RequireAdmin(p.post))
	}
}

// writeForm is the shared tail: a service error is Tomcat's 500 page, anything
// else is the form at 200.
func writeForm(w http.ResponseWriter, f any, err error) {
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, web.ServletError(http.StatusInternalServerError))
		return
	}
	web.WriteJSON(w, http.StatusOK, f)
}

// decodeRename reads the bound body. A body that will not parse is a 400
// before any handler logic runs.
func decodeRename(w http.ResponseWriter, r *http.Request) (form.RenamePost, bool) {
	var post form.RenamePost
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		web.WriteJSON(w, http.StatusBadRequest, web.ServletError(http.StatusBadRequest))
		return post, false
	}
	return post, true
}

func (c *RenameRestController) methodForm(w http.ResponseWriter, r *http.Request) {
	f, err := c.Service.MethodForm()
	writeForm(w, f, err)
}

func (c *RenameRestController) methodRename(w http.ResponseWriter, r *http.Request) {
	post, ok := decodeRename(w, r)
	if !ok {
		return
	}
	f, err := c.Service.RenameMethod(post)
	writeForm(w, f, err)
}

func (c *RenameRestController) panelForm(w http.ResponseWriter, r *http.Request) {
	f, err := c.Service.PanelForm()
	writeForm(w, f, err)
}

func (c *RenameRestController) panelRename(w http.ResponseWriter, r *http.Request) {
	post, ok := decodeRename(w, r)
	if !ok {
		return
	}
	f, err := c.Service.RenamePanel(post)
	writeForm(w, f, err)
}

func (c *RenameRestController) sampleTypeForm(w http.ResponseWriter, r *http.Request) {
	f, err := c.Service.SampleTypeForm()
	writeForm(w, f, err)
}

func (c *RenameRestController) sampleTypeRename(w http.ResponseWriter, r *http.Request) {
	post, ok := decodeRename(w, r)
	if !ok {
		return
	}
	f, err := c.Service.RenameSampleType(post)
	writeForm(w, f, err)
}

func (c *RenameRestController) testSectionForm(w http.ResponseWriter, r *http.Request) {
	f, err := c.Service.TestSectionForm()
	writeForm(w, f, err)
}

func (c *RenameRestController) testSectionRename(w http.ResponseWriter, r *http.Request) {
	post, ok := decodeRename(w, r)
	if !ok {
		return
	}
	f, err := c.Service.RenameTestSection(post)
	writeForm(w, f, err)
}
