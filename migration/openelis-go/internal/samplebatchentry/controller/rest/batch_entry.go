// Package rest ports
// org.openelisglobal.samplebatchentry.controller.rest.SampleBatchEntrySetupRestController.
// Folder layout mirrors the Java source during migration.
//
// AUTH: the route goes through web.Register, which is default-deny since P0.
package rest

import (
	"log"
	"net/http"

	"openelis-go/internal/common/web"
	"openelis-go/internal/samplebatchentry/service"
)

// BatchEntrySetupRestController groups the Wave 4.8 read.
type BatchEntrySetupRestController struct {
	Service *service.BatchEntrySetupService
}

// Routes registers GET rest/SampleBatchEntrySetup. It takes no parameters.
func Routes(mux *http.ServeMux, ctrl *BatchEntrySetupRestController) {
	web.Register(mux, "GET", "rest/SampleBatchEntrySetup", func(w http.ResponseWriter, r *http.Request) {
		dto, err := ctrl.Service.GetSetup()
		if err != nil {
			log.Printf("c2: SampleBatchEntrySetup failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dto)
	})
}
