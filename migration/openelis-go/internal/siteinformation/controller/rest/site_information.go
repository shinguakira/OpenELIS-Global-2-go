// Package rest ports org.openelisglobal.siteinformation.controller.rest —
// SiteInformationRestController and SiteInformationMenuRestController.
// Folder layout mirrors the Java source during migration.
//
// ONE controller pair serves NINE configuration domains in Java, dispatched by
// request.getServletPath().contains(...). The port keeps that shape: five
// handlers, registered once per path, each reading its domain back out of the
// path exactly as Java does.
//
// Per constitution.md Layer IV: request/response mapping only. The dispatch
// table and every DTO decision live in the service (Layer III).
package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	authmw "openelis-go/internal/auth/middleware"
	"openelis-go/internal/common/web"
	"openelis-go/internal/siteinformation/daoimpl"
	"openelis-go/internal/siteinformation/form"
	"openelis-go/internal/siteinformation/service"
)

// SiteInformationRestController mirrors the Java controller pair.
type SiteInformationRestController struct {
	Service *service.SiteInformationService
}

// domains are the nine configuration modules, in path form.
var domains = []string{
	"NonConformityConfiguration",
	"WorkplanConfiguration",
	"PrintedReportsConfiguration",
	"SampleEntryConfig",
	"ResultConfiguration",
	"MenuStatementConfig",
	"PatientConfiguration",
	"ValidationConfiguration",
	"SiteInformation",
}

// deletableDomains is the @GetMapping list on showDeleteSiteInformation, and it
// has SEVEN entries where every other handler has nine.
//
// SampleEntryConfig and ValidationConfiguration have no Delete route at all —
// those two domains can be listed and edited but not deleted, and nothing in
// the Java source says why. Asking for one is a 404, measured.
var deletableDomains = []string{
	"MenuStatementConfig",
	"WorkplanConfiguration",
	"PatientConfiguration",
	"NonConformityConfiguration",
	"ResultConfiguration",
	"PrintedReportsConfiguration",
	"SiteInformation",
}

// deleteDomainOptions is SiteInformationMenuFormValidator's allow-list.
//
// Note "PaitientConfiguration": the delete validator accepts the FORM
// controller's misspelling, while PatientConfigurationMenu — where a client
// gets the value — answers "PatientConfiguration". Round-tripping the value the
// menu handed out is a 400, and the delete succeeds only for a caller that
// knows to send the typo. Measured both ways.
//
// "sampleEntryConfig" is in the list even though no DeleteSampleEntryConfig
// route exists, and "validationConfig" is absent, matching its missing route.
var deleteDomainOptions = map[string]bool{
	"non_conformityConfiguration": true,
	"WorkplanConfiguration":       true,
	"PrintedReportsConfiguration": true,
	"sampleEntryConfig":           true,
	"ResultConfiguration":         true,
	"MenuStatementConfig":         true,
	"PaitientConfiguration":       true,
	"SiteInformation":             true,
}

// Routes registers all ~52 paths.
//
// ADMIN-GATED: both Java controllers carry a class-level
// @PreAuthorize("hasRole('ADMIN')"). web.Register supplies authentication, the
// CSRF check on state-changing verbs, and the module-URL check; RequireAdmin
// adds the role.
func Routes(mux *http.ServeMux, ctrl *SiteInformationRestController) {
	for _, d := range domains {
		// showSiteInformation. NextPrevious shares the handler — the Java
		// mapping lists both and the TODO above it says the functionality was
		// never implemented.
		web.Register(mux, "GET", "rest/"+d, authmw.RequireAdmin(ctrl.show))
		web.Register(mux, "GET", "rest/NextPrevious"+d, authmw.RequireAdmin(ctrl.show))

		// showUpdateSiteInformation.
		web.Register(mux, "POST", "rest/"+d, authmw.RequireAdmin(ctrl.update))

		// cancelSiteInformation — a GET, even though every form in this module
		// reports cancelMethod "POST".
		web.Register(mux, "GET", "rest/Cancel"+d, authmw.RequireAdmin(ctrl.cancel))

		// showSiteInformationMenu.
		web.Register(mux, "GET", "rest/"+d+"Menu", authmw.RequireAdmin(ctrl.menu))
	}

	for _, d := range deletableDomains {
		// showDeleteSiteInformation is a GET that reads a REQUEST BODY. Unusual
		// enough to be worth stating plainly: the verb is GET and the selected
		// ids arrive as JSON.
		web.Register(mux, "GET", "rest/Delete"+d, authmw.RequireAdmin(ctrl.delete))
	}
}

func (c *SiteInformationRestController) show(w http.ResponseWriter, r *http.Request) {
	// UPPERCASE "ID". BaseController's constant is ID, so `?id=` is not read at
	// all — it misses, the handler takes its is-new branch, and the caller gets
	// a blank add-new form for a row that exists. Reproduced, not corrected.
	id := r.URL.Query().Get("ID")

	f, found, err := c.Service.Show(r.URL.Path, id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !found {
		// An id that matches no row: Java's siteInformationService.get returns
		// null and the very next line dereferences it, so the request ends as
		// Tomcat's 500 page rather than a 404.
		web.WriteJSON(w, http.StatusInternalServerError, web.ServletError(http.StatusInternalServerError))
		return
	}
	web.WriteJSON(w, http.StatusOK, f)
}

func (c *SiteInformationRestController) update(w http.ResponseWriter, r *http.Request) {
	// consumes = APPLICATION_JSON_VALUE, so a request with no JSON content type
	// is rejected by the media-type check BEFORE the body is read — a 415,
	// where the delete handler answers 400 for the same empty request.
	if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		web.WriteJSON(w, http.StatusUnsupportedMediaType, form.UnsupportedMediaProblem(r.URL.Path))
		return
	}

	var post form.SiteInformationPost
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		web.WriteJSON(w, http.StatusBadRequest, form.NotReadableProblem(r.URL.Path))
		return
	}

	f, err := c.Service.Update(post, r.URL.Query().Get("ID"), actingUser(r))
	if errors.Is(err, daoimpl.ErrNoSuchRow) {
		// Updating an id that does not exist: Java loads null and dereferences
		// it on the next line, so this is Tomcat's 500 page — the same answer
		// the GET gives for an unknown ID, and for the same reason.
		web.WriteJSON(w, http.StatusInternalServerError, web.ServletError(http.StatusInternalServerError))
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	web.WriteJSON(w, http.StatusOK, f)
}

// actingUser is getSysUserId(request) — the id every audit row is attributed
// to. The guard has already refused an unauthenticated caller, so the principal
// is always present on these routes.
func actingUser(r *http.Request) int64 {
	if p, ok := authmw.FromContext(r.Context()); ok {
		return p.SystemUserID
	}
	return 0
}

func (c *SiteInformationRestController) cancel(w http.ResponseWriter, r *http.Request) {
	// The body is the bare JSON string, not an object.
	web.WriteJSON(w, http.StatusOK, "Cancellation successful")
}

func (c *SiteInformationRestController) delete(w http.ResponseWriter, r *http.Request) {
	var req form.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		web.WriteJSON(w, http.StatusBadRequest, form.NotReadableProblem(r.URL.Path))
		return
	}

	domain := ""
	if req.SiteInfoDomainName != nil {
		domain = *req.SiteInfoDomainName
	}
	if !deleteDomainOptions[domain] {
		// A fourth error envelope: a bare ARRAY of Spring ObjectErrors, not the
		// ProblemDetail a binding failure produces and not the per-field
		// `errors` map a @Valid form produces.
		web.WriteJSON(w, http.StatusBadRequest, []form.ObjectError{{
			Codes: []string{
				"error.field.option.invalid.siteInformationMenuForm.siteInfoDomainName",
				"error.field.option.invalid.siteInfoDomainName",
				"error.field.option.invalid.java.lang.String",
				"error.field.option.invalid",
			},
			Arguments:      []string{"siteInfoDomainName"},
			DefaultMessage: "Field siteInfoDomainName is not a valid option",
			ObjectName:     "siteInformationMenuForm",
			Field:          "siteInfoDomainName",
			RejectedValue:  domain,
			BindingFailure: false,
			Code:           "error.field.option.invalid",
		}})
		return
	}

	if err := c.Service.Delete(req.SelectedIDs, actingUser(r)); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	web.WriteJSON(w, http.StatusOK, "Delete successful")
}

func (c *SiteInformationRestController) menu(w http.ResponseWriter, r *http.Request) {
	f, err := c.Service.Menu(r.URL.Path)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	web.WriteJSON(w, http.StatusOK, f)
}
