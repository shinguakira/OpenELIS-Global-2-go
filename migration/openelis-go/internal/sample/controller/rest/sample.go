// Package rest ports org.openelisglobal.sample.controller.rest.SampleRestController
// (class-level @RequestMapping("/rest/sample")) for the c2 read endpoints.
// Folder layout mirrors the Java source during migration.
//
// Per constitution.md Layer IV: request/response mapping and service calls
// only — the DTO lives in internal/sample/form (Layer V) and the decisions in
// internal/sample/service (Layer III).
//
// AUTH: both routes go through web.Register, which is default-deny since P0,
// so an unauthenticated caller gets Java's 302 to /LoginPage. Neither path has
// a system_module_url row and SampleRestController carries no @PreAuthorize,
// so authentication is the only gate — checked, not assumed.
package rest

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"openelis-go/internal/auth/middleware"
	authsession "openelis-go/internal/auth/session"
	"openelis-go/internal/common/web"
	"openelis-go/internal/sample/daoimpl"
	"openelis-go/internal/sample/form"
	"openelis-go/internal/sample/service"
)

// SampleRestController groups the c2 sample reads.
type SampleRestController struct {
	Service *service.SampleService
}

// Routes registers the c2 sample endpoints.
func Routes(mux *http.ServeMux, ctrl *SampleRestController) {
	// GET rest/sample/all-by-accession/{accessionNumber}
	//
	// Java's status split, reproduced:
	//   no such sample                    -> 404 (bodiless)
	//   sample with no NotStarted rows    -> 200 []
	//   otherwise                         -> 200 [rows]
	//
	// The controller wraps everything in `catch (Exception)` -> 500, so a
	// query failure is a 500 rather than a propagated error page. Same here.
	web.Register(mux, "GET", "rest/sample/all-by-accession/{accessionNumber}", func(w http.ResponseWriter, r *http.Request) {
		forms, err := ctrl.Service.GetAllByAccession(r.PathValue("accessionNumber"))
		if err != nil {
			log.Printf("c2: all-by-accession failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if forms == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		web.WriteJSON(w, http.StatusOK, forms)
	})

	// GET rest/sample/unassigned-by-accession/{accessionNumber}
	//
	// ALWAYS 500 — see SampleService.GetUnassignedByAccession. Java's HQL is
	// invalid and throws at parse time, so no input can succeed and the
	// controller's own not-found branch is unreachable. Registered rather than
	// omitted so a client gets the same status Java gives it; pinned, not
	// fixed.
	web.Register(mux, "GET", "rest/sample/unassigned-by-accession/{accessionNumber}", func(w http.ResponseWriter, r *http.Request) {
		if _, err := ctrl.Service.GetUnassignedByAccession(r.PathValue("accessionNumber")); err != nil {
			// Bodiless, matching ResponseEntity.status(500).build().
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Unreachable, exactly as in Java.
		w.WriteHeader(http.StatusInternalServerError)
	})
}

// PendingAnalysisRestController mirrors
// PendingAnalysisForTestProviderRestController (class-level
// @RequestMapping("/rest")).
type PendingAnalysisRestController struct {
	Service *service.PendingAnalysisService
}

// PendingAnalysisRoutes registers GET rest/getPendingAnalysisForTestProvider.
func PendingAnalysisRoutes(mux *http.ServeMux, ctrl *PendingAnalysisRestController) {
	web.Register(mux, "GET", "rest/getPendingAnalysisForTestProvider", func(w http.ResponseWriter, r *http.Request) {
		testID := r.URL.Query().Get("testId")

		// testId is @RequestParam WITHOUT required=false, so Spring rejects a
		// MISSING param at binding with its own 400 before the handler runs.
		// A PRESENT-but-blank one reaches the handler and hits
		// GenericValidator.isBlankOrNull, which returns 400 with this exact
		// plain-text body. Two different 400s; only the second carries a body.
		if testID == "" {
			if !r.URL.Query().Has("testId") {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			http.Error(w, "Internal error, please contact Admin and file bug report", http.StatusBadRequest)
			return
		}

		dto, err := ctrl.Service.GetPendingForTest(testID)
		if err != nil {
			log.Printf("c2: getPendingAnalysisForTestProvider failed: %v", err)
			http.Error(w, "Internal error, please contact Admin and file bug report", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dto)
	})
}

// OrderAttachmentRestController mirrors OrderAttachmentRestController
// (class-level @RequestMapping("/rest/order")) for its READ endpoints: the
// attachment list plus the download and view streams. The POST upload and the
// soft delete are writes and stay out of the c2 read scope.
type OrderAttachmentRestController struct {
	Service *service.SampleService
}

// OrderAttachmentRoutes registers GET rest/order/{accessionNumber}/attachments.
func OrderAttachmentRoutes(mux *http.ServeMux, ctrl *OrderAttachmentRestController) {
	web.Register(mux, "GET", "rest/order/{accessionNumber}/attachments", func(w http.ResponseWriter, r *http.Request) {
		dtos, err := ctrl.Service.GetOrderAttachments(r.PathValue("accessionNumber"))
		if err != nil {
			log.Printf("c2: order attachments failed: %v", err)
			web.WriteJSON(w, http.StatusInternalServerError, form.ErrorDTO{Error: "Failed to save attachment"})
			return
		}
		if dtos == nil {
			// 404 WITH A BODY: Map.of("error", "Order not found"). Not the
			// bodiless 404 all-by-accession returns for the same missing
			// accession — two endpoints in this wave, two not-found shapes.
			web.WriteJSON(w, http.StatusNotFound, form.ErrorDTO{Error: "Order not found"})
			return
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})

	// GET rest/order/attachments/{attachmentId}/download  (Content-Disposition: attachment)
	// GET rest/order/attachments/{attachmentId}/view      (Content-Disposition: inline)
	//
	// One handler, two dispositions — exactly how Java does it, both delegating
	// to serveAttachment with a different literal.
	for _, m := range []struct{ route, disposition string }{
		{"rest/order/attachments/{attachmentId}/download", "attachment"},
		{"rest/order/attachments/{attachmentId}/view", "inline"},
	} {
		disposition := m.disposition
		web.Register(mux, "GET", m.route, func(w http.ResponseWriter, r *http.Request) {
			id, err := strconv.ParseInt(r.PathValue("attachmentId"), 10, 64)
			if err != nil {
				// attachmentId binds as Integer, so Spring rejects a
				// non-numeric value at binding time, before the handler runs.
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			content, err := ctrl.Service.GetAttachmentContent(id)
			if err != nil {
				// A missing id is a 500 in Java, not a 404 — the service's
				// get() throws rather than returning null. Pinned, not fixed.
				if !errors.Is(err, daoimpl.ErrAttachmentNotFound) {
					log.Printf("c2: order attachment %s failed: %v", disposition, err)
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if content == nil {
				// Soft-deleted, or no bytes: ResponseEntity.notFound().build()
				// — 404 with no body.
				w.WriteHeader(http.StatusNotFound)
				return
			}
			// Spring writes the parsed MediaType back WITH the default charset
			// attached, so even application/pdf and application/octet-stream
			// arrive as "…;charset=UTF-8". Measured live; a bare content type
			// is a different response.
			w.Header().Set("Content-Type", content.ContentType+";charset=UTF-8")
			w.Header().Set("Content-Disposition", disposition+`; filename="`+content.FileName+`"`)
			w.Header().Set("Content-Length", strconv.Itoa(len(content.Bytes)))
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write(content.Bytes); err != nil {
				log.Printf("c2: order attachment %s write failed: %v", disposition, err)
			}
		})
	}
}

// UnassignedSampleRestController mirrors UnassignedSampleRestController
// (class-level @RequestMapping("/rest/unassigned-sample")) for its GET
// endpoints. The three PUT routes are writes and out of the c2 read scope.
type UnassignedSampleRestController struct {
	Service *service.UnassignedSampleService
}

// UnassignedSampleRoutes registers the five unassigned-sample GET endpoints.
func UnassignedSampleRoutes(mux *http.ServeMux, ctrl *UnassignedSampleRestController) {
	// GET rest/unassigned-sample — a BARE @GetMapping, so the path is exactly
	// this with NO trailing slash. Spring 6 dropped automatic trailing-slash
	// matching, so "/rest/unassigned-sample/" is a 404 there and must be here
	// too. Go's ServeMux would happily treat the two as one pattern, so the
	// distinction has to be deliberate: registering only the bare form leaves
	// the slashed form unmatched, which is what produces the 404.
	web.Register(mux, "GET", "rest/unassigned-sample", func(w http.ResponseWriter, r *http.Request) {
		dtos, err := ctrl.Service.GetUnassignedForDashboard()
		if err != nil {
			log.Printf("c2: unassigned-sample failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})

	// GET rest/unassigned-sample/items
	//
	// 500 with an EMPTY body whenever the query matches anything — Java's
	// behavior, reproduced rather than fixed. w.WriteHeader alone (no
	// http.Error, no JSON envelope) is what makes Content-Length 0, matching
	// ResponseEntity.status(500).build(). See
	// service.ErrUnassignedItemsUnserializable for the Hibernate cause.
	web.Register(mux, "GET", "rest/unassigned-sample/items", func(w http.ResponseWriter, r *http.Request) {
		dtos, err := ctrl.Service.GetUnassignedItems("")
		if err != nil {
			if !errors.Is(err, service.ErrUnassignedItemsUnserializable) {
				log.Printf("c2: unassigned-sample/items failed: %v", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})

	// GET rest/unassigned-sample/items/search?accessionNumber=
	//
	// accessionNumber is a required @RequestParam with no default, so Spring
	// answers 400 when it is absent — unlike /items, which takes no params.
	web.Register(mux, "GET", "rest/unassigned-sample/items/search", func(w http.ResponseWriter, r *http.Request) {
		if !r.URL.Query().Has("accessionNumber") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		dtos, err := ctrl.Service.GetUnassignedItems(r.URL.Query().Get("accessionNumber"))
		if err != nil {
			// Same bodiless 500 as /items, for the same reason: a search that
			// matches rows hits the identical DTO-building path.
			if !errors.Is(err, service.ErrUnassignedItemsUnserializable) {
				log.Printf("c2: unassigned-sample/items/search failed: %v", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})

	// GET rest/unassigned-sample/by-facility/{facilityId}
	//
	// facilityId binds as Integer, so a non-numeric value fails Spring's
	// binding with a 400 before the handler runs — in contrast with c1's
	// patient endpoints, where a String-bound path variable against a varchar
	// column simply matches nothing and returns 200.
	web.Register(mux, "GET", "rest/unassigned-sample/by-facility/{facilityId}", func(w http.ResponseWriter, r *http.Request) {
		facilityID, err := strconv.ParseInt(r.PathValue("facilityId"), 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		dtos, err := ctrl.Service.GetUnassignedByFacility(facilityID)
		if err != nil {
			log.Printf("c2: unassigned-sample/by-facility failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})

	// GET rest/unassigned-sample/count-by-facility/{facilityId}
	//
	// Returns {"count": n} — a one-key HashMap, never a bare number. The count
	// is a SUBSET of by-facility's length, not its equal: see
	// UnassignedSampleService.CountUnassignedByFacility.
	web.Register(mux, "GET", "rest/unassigned-sample/count-by-facility/{facilityId}", func(w http.ResponseWriter, r *http.Request) {
		facilityID, err := strconv.ParseInt(r.PathValue("facilityId"), 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		count, err := ctrl.Service.CountUnassignedByFacility(facilityID)
		if err != nil {
			log.Printf("c2: unassigned-sample/count-by-facility failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, form.CountDTO{Count: count})
	})
}

// OrderDashboardRestController mirrors OrderSearchRestController's /dashboard
// endpoint (class-level @RequestMapping("/rest/order")).
type OrderDashboardRestController struct {
	Service *service.OrderDashboardService
}

// OrderDashboardRoutes registers GET rest/order/dashboard.
func OrderDashboardRoutes(mux *http.ServeMux, ctrl *OrderDashboardRestController) {
	web.Register(mux, "GET", "rest/order/dashboard", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		// BINDING FAILURES ARE 400, not a silent default.
		//
		// page and pageSize bind as int and includeExternal as boolean, so a
		// value Spring cannot convert is rejected before the handler runs, with
		// a ProblemDetail body. An earlier version of this port used
		// atoiDefault here and answered 200 with the default — a divergence
		// that its own comment acknowledged and declined to emulate. Measured:
		//   ?page=abc  ?pageSize=abc  ?includeExternal=abc   -> 400
		//   ?page=      ?pageSize=                           -> 200 (empty is
		//                                                      not a failure;
		//                                                      the default wins)
		//   ?startDate=abc ?priority=abc ?status=abc ?search=abc -> 200 (String
		//                                                      params bind to
		//                                                      anything)
		page, ok := bindInt(q, "page", 1)
		if !ok {
			web.WriteJSON(w, http.StatusBadRequest, form.TypeMismatchProblem(r))
			return
		}
		pageSize, ok := bindInt(q, "pageSize", 100)
		if !ok {
			web.WriteJSON(w, http.StatusBadRequest, form.TypeMismatchProblem(r))
			return
		}
		includeExternal := false
		if raw := q.Get("includeExternal"); raw != "" {
			// Spring's StringToBooleanConverter vocabulary — "t"/"f" are NOT
			// in it and fail to bind, unlike Go's strconv.ParseBool.
			v, bok := parseSpringBoolean(raw)
			if !bok {
				web.WriteJSON(w, http.StatusBadRequest, form.TypeMismatchProblem(r))
				return
			}
			includeExternal = v
		}

		dto, err := ctrl.Service.GetDashboard(service.DashboardQuery{
			Page:            page,
			PageSize:        pageSize,
			Search:          q.Get("search"),
			Status:          q.Get("status"),
			Priority:        q.Get("priority"),
			IncludeExternal: includeExternal,
			StartDate:       q.Get("startDate"),
			EndDate:         q.Get("endDate"),
		})
		if err != nil {
			log.Printf("c2: order/dashboard failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dto)
	})
}

// bindInt ports Spring's int @RequestParam binding: an absent or EMPTY value
// takes the default, anything non-numeric is a bind failure.
func bindInt(q url.Values, name string, def int) (int, bool) {
	raw := q.Get(name)
	if raw == "" {
		return def, true
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return n, true
}

// parseSpringBoolean ports Spring's StringToBooleanConverter. Duplicated from
// the c1 patient controller deliberately: the two are separate ports of the
// same Spring behaviour and neither should start depending on the other.
//
//	true:  "true", "on", "yes", "1"
//	false: "false", "off", "no", "0"
//
// Anything else — "t", "f", "abc" — fails to bind. Measured on the live server
// across the full matrix.
func parseSpringBoolean(raw string) (value bool, ok bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true", "on", "yes", "1":
		return true, true
	case "false", "off", "no", "0":
		return false, true
	default:
		return false, false
	}
}

// OrderSearchRestController mirrors OrderSearchRestController's /search
// endpoint (class-level @RequestMapping("/rest/order")).
type OrderSearchRestController struct {
	Service *service.OrderSearchService
}

// OrderSearchRoutes registers GET rest/order/search.
func OrderSearchRoutes(mux *http.ServeMux, ctrl *OrderSearchRestController) {
	web.Register(mux, "GET", "rest/order/search", func(w http.ResponseWriter, r *http.Request) {
		labNumber := r.URL.Query().Get("labNumber")
		// labNumber is @RequestParam(required = false), so a missing value
		// reaches the handler and is rejected there with a BODILESS 400 —
		// ResponseEntity.badRequest().build(). Blank and whitespace-only take
		// the same branch (`labNumber.trim().isEmpty()`).
		if strings.TrimSpace(labNumber) == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		dto, err := ctrl.Service.GetOrderByLabNumber(labNumber)
		if err != nil {
			log.Printf("c2: order/search failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if dto == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		web.WriteJSON(w, http.StatusOK, dto)
	})
}

// SampleEditRestController mirrors
// org.openelisglobal.sample.controller.rest.SampleEditRestController
// (@RequestMapping("/rest/") + @GetMapping("SampleEdit")).
//
// NOT the MVC SampleEditController, which maps /SampleEdit and sets two extra
// keys the REST class leaves commented out.
type SampleEditRestController struct {
	Service *service.SampleEditService
	// Sessions backs the SampleEditWritable attribute. Java keeps it in the
	// HttpSession, so a form opened with ?type=readwrite stays editable on
	// later requests that omit the parameter.
	Sessions *authsession.MemoryStore
}

// SampleEditRoutes registers GET rest/SampleEdit.
//
// Both params are @RequestParam(required = false), so every combination is a
// 200 — this endpoint never 404s, even for an accession that matches nothing.
func SampleEditRoutes(mux *http.ServeMux, ctrl *SampleEditRestController) {
	web.Register(mux, "GET", "rest/SampleEdit", func(w http.ResponseWriter, r *http.Request) {
		dto, err := ctrl.Service.GetSampleEdit(service.SampleEditRequest{
			AccessionNumber:    r.URL.Query().Get("accessionNumber"),
			PatientID:          r.URL.Query().Get("patientId"),
			SysUserID:          sysUserID(r),
			Editable:           ctrl.resolveEditable(r),
			AllowedToCancelAll: allowedToCancelAll(r),
		})
		if err != nil {
			log.Printf("c2: SampleEdit failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dto)
	})
}

// SamplePatientEntryRestController mirrors
// org.openelisglobal.sample.controller.rest.SamplePatientEntryRestController
// (@RequestMapping("/rest/") + @GetMapping("SamplePatientEntry")).
type SamplePatientEntryRestController struct {
	Service *service.SamplePatientEntryService
}

// SamplePatientEntryRoutes registers GET rest/SamplePatientEntry. No params.
func SamplePatientEntryRoutes(mux *http.ServeMux, ctrl *SamplePatientEntryRestController) {
	web.Register(mux, "GET", "rest/SamplePatientEntry", func(w http.ResponseWriter, r *http.Request) {
		dto, err := ctrl.Service.GetForm(sysUserID(r))
		if err != nil {
			log.Printf("c2: SamplePatientEntry failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dto)
	})
}

// resolveEditable ports isEditable:
//
//	"readwrite".equals(session.getAttribute(SAMPLE_EDIT_WRITABLE))
//	  || "readwrite".equals(request.getParameter("type"))
//
// and the WRITE half of the same controller, which stores the attribute when
// ?type=readwrite arrives WITHOUT an accession number.
//
// Measured on the live server: opening ?type=readwrite with no accession makes
// a LATER ?accessionNumber=... request (no type parameter) still report
// isEditable true. An earlier version of this port read only the query
// parameter, so that second request came back read-only.
func (ctrl *SampleEditRestController) resolveEditable(r *http.Request) bool {
	q := r.URL.Query()
	typeParam := q.Get("type") == "readwrite"

	if ctrl.Sessions == nil {
		return typeParam
	}
	id := authsession.IDFromRequest(r)

	// Java only WRITES the attribute on the branch with no accession number —
	// the else-branch of the accession check. Writing it unconditionally would
	// make a one-off ?accessionNumber=X&type=readwrite stick for the session,
	// which Java does not do.
	if typeParam && strings.TrimSpace(q.Get("accessionNumber")) == "" {
		ctrl.Sessions.SetAttribute(id, authsession.SampleEditWritable, "readwrite")
	}

	return typeParam || ctrl.Sessions.Attribute(id, authsession.SampleEditWritable) == "readwrite"
}

// sysUserID is Java's BaseController.getSysUserId(request): the id of the
// authenticated user, which the role-filtered sampleTypes list depends on.
//
// Returns "" when there is no principal. That cannot happen on these routes —
// web.Register is default-deny — but returning "" rather than a fallback id
// keeps a future open route from silently serving another user's list.
func sysUserID(r *http.Request) string {
	if p, ok := middleware.FromContext(r.Context()); ok {
		return strconv.FormatInt(p.SystemUserID, 10)
	}
	return ""
}

// allowedToCancelAll ports SampleEditRestController's
//
//	userModuleService.isUserAdmin(request)
//	  || userRoleService.userInRole(getSysUserId(request), ABLE_TO_CANCEL_ROLE_NAMES)
//
// isUserAdmin is login_user.is_admin = 'Y' alone, which is exactly what
// Principal.IsAdmin holds — NOT the Global Administrator role.
func allowedToCancelAll(r *http.Request) bool {
	p, ok := middleware.FromContext(r.Context())
	if !ok {
		return false
	}
	if p.IsAdmin {
		return true
	}
	for _, role := range service.AbleToCancelRoles {
		if p.HasRole(role) {
			return true
		}
	}
	return false
}
