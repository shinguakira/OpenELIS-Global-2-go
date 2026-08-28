// Package rest ports AccessionValidationRestController (constitution.md Layer IV).
package rest

import (
	"net/http"

	"openelis-go/internal/common/web"
	"openelis-go/internal/resultvalidation/service"
)

// AccessionValidationRestController serves rest/AccessionValidation.
type AccessionValidationRestController struct {
	Service *service.ResultValidationService
}

// Routes registers the endpoint.
func Routes(mux *http.ServeMux, c *AccessionValidationRestController) {
	web.Register(mux, "GET", "rest/AccessionValidation", c.get)
}

func (c *AccessionValidationRestController) get(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// @RequestParam(defaultValue = "true") Boolean doRange — the DEFAULT is the
	// range search, so a caller who supplies only an accession number gets the
	// looser reading unless they opt out explicitly.
	doRange := true
	if v := q.Get("doRange"); v != "" {
		doRange = !(v == "false" || v == "0" || v == "off" || v == "no")
	}

	f, err := c.Service.Load(q.Get("accessionNumber"), q.Get("date"), q.Get("unitType"), doRange)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, web.ServletError(http.StatusInternalServerError))
		return
	}
	web.WriteJSON(w, http.StatusOK, f)
}
