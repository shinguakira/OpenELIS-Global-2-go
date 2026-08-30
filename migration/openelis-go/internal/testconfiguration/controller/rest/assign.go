package rest

import (
	"encoding/json"
	"net/http"

	authmw "openelis-go/internal/auth/middleware"
	"openelis-go/internal/common/web"
	"openelis-go/internal/testconfiguration/form"
	"openelis-go/internal/testconfiguration/service"
)

// AssignRestController mirrors SampleTypeTestAssignRestController,
// TestSectionTestAssignRestController and PanelTestAssignRestController.
type AssignRestController struct {
	Service *service.AssignService
}

// AssignRoutes registers the three GET/POST pairs.
func AssignRoutes(mux *http.ServeMux, ctrl *AssignRestController) {
	web.Register(mux, "GET", "rest/SampleTypeTestAssign", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			f, err := ctrl.Service.SampleTypeForm()
			writeForm(w, f, err)
		}))
	web.Register(mux, "POST", "rest/SampleTypeTestAssign", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			post, ok := decodeAssign(w, r)
			if !ok {
				return
			}
			f, err := ctrl.Service.AssignSampleType(post, actingUser(r))
			writeForm(w, f, err)
		}))

	web.Register(mux, "GET", "rest/TestSectionTestAssign", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			f, err := ctrl.Service.TestSectionForm()
			writeForm(w, f, err)
		}))
	web.Register(mux, "POST", "rest/TestSectionTestAssign", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			post, ok := decodeAssign(w, r)
			if !ok {
				return
			}
			f, err := ctrl.Service.AssignTestSection(post, actingUser(r))
			writeForm(w, f, err)
		}))

	web.Register(mux, "GET", "rest/PanelTestAssign", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			// setupDisplayItems fills selectedPanel only when the form already
			// carries a panel id, which the GET has no way to supply — so the
			// blank form answers the panel list alone.
			f, err := ctrl.Service.PanelForm(r.URL.Query().Get("panelId"))
			writeForm(w, f, err)
		}))
	web.Register(mux, "POST", "rest/PanelTestAssign", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			post, ok := decodeAssign(w, r)
			if !ok {
				return
			}
			f, err := ctrl.Service.AssignPanelTests(post, actingUser(r))
			writeForm(w, f, err)
		}))
}

func decodeAssign(w http.ResponseWriter, r *http.Request) (form.AssignPost, bool) {
	var post form.AssignPost
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		web.WriteJSON(w, http.StatusBadRequest, web.ServletError(http.StatusBadRequest))
		return post, false
	}
	return post, true
}
