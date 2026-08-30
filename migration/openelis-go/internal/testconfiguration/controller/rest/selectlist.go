package rest

import (
	"encoding/json"
	"net/http"

	authmw "openelis-go/internal/auth/middleware"
	"openelis-go/internal/common/web"
	"openelis-go/internal/testconfiguration/form"
	"openelis-go/internal/testconfiguration/service"
)

// SelectListRestController mirrors TestRenameEntryRestController,
// SelectListRenameEntryRestController and ResultSelectListAddRestController.
type SelectListRestController struct {
	Service *service.SelectListService
}

// SelectListRoutes registers the four endpoints. ResultSelectListAdd's POST is
// one of them and writes nothing — the write is /SaveResultSelectList.
func SelectListRoutes(mux *http.ServeMux, ctrl *SelectListRestController) {
	web.Register(mux, "GET", "rest/TestRenameEntry", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			f, err := ctrl.Service.TestRenameForm()
			writeForm(w, f, err)
		}))
	web.Register(mux, "POST", "rest/TestRenameEntry", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			post, ok := decodeSelectList(w, r)
			if !ok {
				return
			}
			f, err := ctrl.Service.RenameTest(post)
			writeForm(w, f, err)
		}))

	web.Register(mux, "GET", "rest/SelectListRenameEntry", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			f, err := ctrl.Service.SelectListRenameForm()
			writeForm(w, f, err)
		}))
	web.Register(mux, "POST", "rest/SelectListRenameEntry", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			post, ok := decodeSelectList(w, r)
			if !ok {
				return
			}
			f, err := ctrl.Service.RenameSelectOption(post, actingUser(r))
			writeForm(w, f, err)
		}))

	web.Register(mux, "GET", "rest/ResultSelectListAdd", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			f, err := ctrl.Service.ResultSelectListAdd(form.SelectListPost{})
			// The GET answers page "1"; only the POST moves it to "2".
			if err == nil {
				f.Page = "1"
				f.Tests, f.TestDictionary = nil, nil
				f.NameEnglish, f.NameFrench = nil, nil
			}
			writeForm(w, f, err)
		}))
	web.Register(mux, "POST", "rest/ResultSelectListAdd", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			post, ok := decodeSelectList(w, r)
			if !ok {
				return
			}
			f, err := ctrl.Service.ResultSelectListAdd(post)
			writeForm(w, f, err)
		}))

	web.Register(mux, "POST", "rest/SaveResultSelectList", authmw.RequireAdmin(
		func(w http.ResponseWriter, r *http.Request) {
			post, ok := decodeSelectList(w, r)
			if !ok {
				return
			}
			f, err := ctrl.Service.SaveResultSelectList(post, actingUser(r))
			writeForm(w, f, err)
		}))
}

func decodeSelectList(w http.ResponseWriter, r *http.Request) (form.SelectListPost, bool) {
	var post form.SelectListPost
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		web.WriteJSON(w, http.StatusBadRequest, web.ServletError(http.StatusBadRequest))
		return post, false
	}
	return post, true
}
