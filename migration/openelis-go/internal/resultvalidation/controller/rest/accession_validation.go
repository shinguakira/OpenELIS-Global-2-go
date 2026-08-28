// Package rest ports AccessionValidationRestController (constitution.md Layer IV).
package rest

import (
	"net/http"
	"strconv"

	"openelis-go/internal/auth/middleware"

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

	// The AUTHENTICATED caller, not a fixed id: their lab-unit grants decide
	// which sections and which results this endpoint may return.
	f, err := c.Service.Load(sysUserID(r), q.Get("accessionNumber"), q.Get("date"), q.Get("unitType"), doRange)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, web.ServletError(http.StatusInternalServerError))
		return
	}
	web.WriteJSON(w, http.StatusOK, f)
}

// sysUserID is the authenticated principal, or the empty string when there is
// none. These routes are default-deny so that cannot happen; returning empty
// rather than a fallback id keeps a future open route from serving one
// caller's lab-unit scope to another.
func sysUserID(r *http.Request) string {
	if p, ok := middleware.FromContext(r.Context()); ok {
		return strconv.FormatInt(p.SystemUserID, 10)
	}
	return ""
}
