// Package rest ports org.openelisglobal.testconfiguration.controller.rest.
// Folder layout mirrors the Java source during migration.
//
// Per constitution.md Layer IV: request/response mapping only.
package rest

import (
	"encoding/json"
	"net/http"

	authmw "openelis-go/internal/auth/middleware"
	"openelis-go/internal/common/web"
	"openelis-go/internal/testconfiguration/form"
	"openelis-go/internal/testconfiguration/service"
)

// UomConfigRestController mirrors UomCreateRestController and
// UomRenameEntryRestController — two Java classes, one type here because they
// share a service and a table.
//
// Both are @PreAuthorize("hasRole('ADMIN')") in Java, so both go behind
// RequireAdmin.
type UomConfigRestController struct {
	Service *service.UomConfigService
}

// Routes registers the four endpoints of e2 slice 1.
func UomRoutes(mux *http.ServeMux, ctrl *UomConfigRestController) {
	web.Register(mux, "GET", "rest/UomCreate", authmw.RequireAdmin(ctrl.createForm))
	web.Register(mux, "POST", "rest/UomCreate", authmw.RequireAdmin(ctrl.create))
	web.Register(mux, "GET", "rest/UomRenameEntry", authmw.RequireAdmin(ctrl.renameForm))
	web.Register(mux, "POST", "rest/UomRenameEntry", authmw.RequireAdmin(ctrl.rename))
}

func (c *UomConfigRestController) createForm(w http.ResponseWriter, r *http.Request) {
	f, err := c.Service.CreateForm()
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, web.ServletError(http.StatusInternalServerError))
		return
	}
	web.WriteJSON(w, http.StatusOK, f)
}

func (c *UomConfigRestController) create(w http.ResponseWriter, r *http.Request) {
	var post form.UomCreatePost
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		web.WriteJSON(w, http.StatusBadRequest, web.ServletError(http.StatusBadRequest))
		return
	}
	f, err := c.Service.Create(post)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, web.ServletError(http.StatusInternalServerError))
		return
	}
	web.WriteJSON(w, http.StatusOK, f)
}

func (c *UomConfigRestController) renameForm(w http.ResponseWriter, r *http.Request) {
	f, err := c.Service.RenameForm()
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, web.ServletError(http.StatusInternalServerError))
		return
	}
	web.WriteJSON(w, http.StatusOK, f)
}

func (c *UomConfigRestController) rename(w http.ResponseWriter, r *http.Request) {
	var post form.UomRenameEntryPost
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		web.WriteJSON(w, http.StatusBadRequest, web.ServletError(http.StatusBadRequest))
		return
	}
	f, err := c.Service.Rename(post)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, web.ServletError(http.StatusInternalServerError))
		return
	}
	web.WriteJSON(w, http.StatusOK, f)
}
