// Package rest ports ReferredOutTestsRestController (constitution.md Layer IV).
package rest

import (
	"log"
	"net/http"
	"time"

	"openelis-go/internal/common/web"
	"openelis-go/internal/referral/form"
	"openelis-go/internal/referral/service"
)

// ReferredOutTestsRestController serves rest/ReferredOutTests.
type ReferredOutTestsRestController struct {
	Service *service.ReferredOutTestsService
}

// Routes registers the endpoint.
func Routes(mux *http.ServeMux, c *ReferredOutTestsRestController) {
	web.Register(mux, "GET", "rest/ReferredOutTests", c.get)
}

func (c *ReferredOutTestsRestController) get(w http.ResponseWriter, r *http.Request) {
	searchType := r.URL.Query().Get("searchType")

	// The form is @Valid-bound, so a searchType outside the enum fails binding
	// BEFORE the controller runs — and Spring answers with a per-field `errors`
	// map, not the RFC 7807 ProblemDetail the WorkPlan endpoints produce for
	// the same class of failure. Two binding failures, two envelopes, both in
	// this one wave.
	if searchType != "" && !form.IsSearchType(searchType) {
		web.WriteJSON(w, http.StatusBadRequest, form.BindErrorBody{
			Timestamp: time.Now().UnixMilli(),
			Status:    http.StatusBadRequest,
			Errors: map[string]string{
				"searchType": "Failed to convert property value of type 'java.lang.String'" +
					" to required type 'org.openelisglobal.referral.form.ReferredOutTestsForm$SearchType'" +
					" for property 'searchType'; Failed to convert from type [java.lang.String]" +
					" to type [org.openelisglobal.referral.form.ReferredOutTestsForm$SearchType]" +
					" for value [" + searchType + "]",
			},
		})
		return
	}

	f, err := c.Service.Load(searchType, r.URL.Query().Get("labNumber"))
	if err != nil {
		// Java answers 500 here too, but it LOGS the cause. Swallowing it made a
		// broken query indistinguishable from the reproduced TEST_AND_DATES
		// defect.
		log.Printf("ReferredOutTests: %v", err)
		web.WriteJSON(w, http.StatusInternalServerError, web.ServletError(http.StatusInternalServerError))
		return
	}
	web.WriteJSON(w, http.StatusOK, f)
}
