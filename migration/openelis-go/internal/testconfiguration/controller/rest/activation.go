package rest

import (
	"encoding/json"
	"net/http"

	authmw "openelis-go/internal/auth/middleware"
	"openelis-go/internal/common/web"
	"openelis-go/internal/testconfiguration/form"
	"openelis-go/internal/testconfiguration/service"
)

// ActivationRestController mirrors TestActivationRestController and
// TestOrderabilityRestController.
type ActivationRestController struct {
	Service *service.ActivationService
}

// ActivationRoutes registers the two GET/POST pairs.
func ActivationRoutes(mux *http.ServeMux, ctrl *ActivationRestController) {
	web.Register(mux, "GET", "rest/TestActivation", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			f, err := ctrl.Service.TestActivationForm()
			writeForm(w, f, err)
		}))
	web.Register(mux, "POST", "rest/TestActivation", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			post, ok := decodeActivation(w, r)
			if !ok {
				return
			}
			f, err := ctrl.Service.ApplyTestActivation(post, actingUser(r))
			writeForm(w, f, err)
		}))

	web.Register(mux, "GET", "rest/TestOrderability", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			f, err := ctrl.Service.TestOrderabilityForm()
			writeForm(w, f, err)
		}))
	web.Register(mux, "POST", "rest/TestOrderability", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			post, ok := decodeActivation(w, r)
			if !ok {
				return
			}
			_, err := ctrl.Service.ApplyTestOrderability(post, actingUser(r))
			if err != nil {
				web.WriteJSON(w, http.StatusInternalServerError, web.ServletError(http.StatusInternalServerError))
				return
			}
			// NOT the form: this handler ends
			// `return ResponseEntity.ok(Collections.singletonMap("status", "success"))`,
			// having built a list it then throws away.
			web.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
		}))
}

func decodeActivation(w http.ResponseWriter, r *http.Request) (form.ActivationPost, bool) {
	var post form.ActivationPost
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		web.WriteJSON(w, http.StatusBadRequest, web.ServletError(http.StatusBadRequest))
		return post, false
	}
	return post, true
}
