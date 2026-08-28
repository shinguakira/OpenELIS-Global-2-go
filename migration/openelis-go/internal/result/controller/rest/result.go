// Package rest ports LogbookResultsRestController and
// AccessionResultsRestController (constitution.md Layer IV).
package rest

import (
	"net/http"

	"openelis-go/internal/common/web"
	"openelis-go/internal/result/service"
)

// ResultRestController serves rest/LogbookResults and rest/accession-results.
type ResultRestController struct {
	Service *service.ResultService
}

// Routes registers both.
func Routes(mux *http.ServeMux, c *ResultRestController) {
	web.Register(mux, "GET", "rest/LogbookResults", c.logbook)
	web.Register(mux, "GET", "rest/accession-results", c.accessionResults)
}

func (c *ResultRestController) logbook(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f, err := c.Service.Logbook(q.Get("labNumber"), q.Get("selectedTest"))
	if err != nil {
		// Includes the reproduced getTestDisplayName NPE, which Java surfaces
		// as Tomcat's default error body.
		web.WriteJSON(w, http.StatusInternalServerError, web.ServletError(http.StatusInternalServerError))
		return
	}
	web.WriteJSON(w, http.StatusOK, f)
}

func (c *ResultRestController) accessionResults(w http.ResponseWriter, r *http.Request) {
	f, err := c.Service.AccessionResults(r.URL.Query().Get("accessionNumber"))
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, web.ServletError(http.StatusInternalServerError))
		return
	}
	web.WriteJSON(w, http.StatusOK, f)
}
