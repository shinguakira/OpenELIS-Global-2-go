// Package rest ports org.openelisglobal.system.controller.rest — the system
// domain's REST controllers. Folder layout mirrors the Java source
// (system/controller/rest) during migration; idiomatic Go reorg comes at the end.
package rest

import (
	"net/http"
	"os"
	"time"

	"openelis-go/internal/common/web"
)

// Routes registers the system REST endpoints. Mirrors the @GetMapping methods on
// SystemRestController (@RequestMapping("/rest")).
func Routes(mux *http.ServeMux) {
	web.Register(mux, "GET", "rest/server-time", ServerTime)
}

// ServerTime reproduces SystemRestController#getServerTime:
//
//	GET /rest/server-time -> {"date","time","timezone"}
//
// date     = ISO local date (yyyy-MM-dd)
// time     = HH:mm (24-hour)
// timezone = IANA system zone id (e.g. "Etc/UTC")
func ServerTime(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	web.WriteJSON(w, http.StatusOK, map[string]string{
		"date":     now.Format("2006-01-02"),
		"time":     now.Format("15:04"),
		"timezone": systemZoneID(),
	})
}

// systemZoneID mirrors ZoneId.systemDefault().getId() — an IANA zone id
// (e.g. "Etc/UTC"), not Go's zone abbreviation ("UTC"/"JST"). Prefers the TZ
// env var (set in the container), then the resolved local IANA name, and only
// falls back to the abbreviation on a dev box with no zone info.
func systemZoneID() string {
	if tz := os.Getenv("TZ"); tz != "" {
		return tz
	}
	if name := time.Local.String(); name != "" && name != "Local" {
		return name
	}
	name, _ := time.Now().Zone()
	return name
}
