// Package rest ports the c1 patient read endpoints. Folder layout mirrors the
// Java source during migration.
//
// Per constitution.md Layer IV: request/response mapping and service calls
// only — no DTO shaping here. See internal/patient/form (Layer V) and
// internal/patient/service (Layer III).
//
// ── ROUTE PROVENANCE ───────────────────────────────────────────────────────
// Every path below was confirmed against the LIVE Java server, not inferred.
// Two are easy to get wrong:
//   - rest/patientByLabNumer lives on SampleEditRestController (class-level
//     @RequestMapping("/rest/")), NOT on a patient controller, and its query
//     param is `accessionNumber` — despite the endpoint being named
//     "LabNumer". The obvious guess (labNumber) returns 400.
//   - rest/patient/merge/details/{id} is under PatientMergeRestController's
//     class-level @RequestMapping("/rest/patient/merge").
//
// ── SECURITY NOTE ──────────────────────────────────────────────────────────
// These endpoints return PHI: names, birth dates, national IDs, addresses,
// phone numbers, email. Two gates apply, both measured against live Java:
//
//		endpoint                          anon   authed, no Reception   Reception
//		patientByLabNumer                 302    200                    200
//		patient/merge/details/{id}        302    403 (bodiless)         200
//		patient-id-documents/{id}         302    200                    200
//		patient-photos/{id}/{isThumbnail} 302    200                    200
//
//	 1. Authentication, for all of them. Nothing extra is needed here: every
//	    route below goes through web.Register, which is DEFAULT-DENY since P0
//	    auth landed, so an unauthenticated caller gets the same 302 to
//	    /LoginPage Java sends, carrying no body.
//	 2. The "Reception" role, for merge/details only — declared at that route.
//
// Neither gate is the module system: none of these paths has a
// system_module_url row (checked directly, not assumed), so
// ModuleAuthenticationInterceptor auto-allows them, and none of the owning
// Java controllers carries an @PreAuthorize.
package rest

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	authmw "openelis-go/internal/auth/middleware"
	"openelis-go/internal/common/web"
	"openelis-go/internal/patient/service"
)

// PatientRestController groups the c1 read endpoints.
type PatientRestController struct {
	Service      *service.PatientService
	MergeService *service.PatientMergeService
}

// Routes registers the c1 endpoints.
func Routes(mux *http.ServeMux, ctrl *PatientRestController) {
	// GET rest/patientByLabNumer?accessionNumber=... — mirrors
	// SampleEditRestController.getPatientByLabNumber.
	//
	// Java's own order of checks, reproduced: blank/missing param -> 400
	// (empty body) BEFORE any lookup; then a miss -> 404 (empty body). Note
	// this is a real 404, unlike the b2 family where a missing id produces a
	// 500 via ObjectNotFoundException — the not-found contract is per-endpoint
	// in this codebase, not global.
	web.Register(mux, "GET", "rest/patientByLabNumer", func(w http.ResponseWriter, r *http.Request) {
		// Java trims via StringTrimmerEditor(false), so whitespace-only
		// becomes "" and is treated as blank rather than as a real value.
		accession := strings.TrimSpace(r.URL.Query().Get("accessionNumber"))
		if accession == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		dto, err := ctrl.Service.GetPatientByAccessionNumber(accession)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if dto == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		web.WriteJSON(w, http.StatusOK, dto)
	})

	// GET rest/patient/merge/details/{patientId} — mirrors
	// PatientMergeRestController.getMergeDetails.
	//
	// The three-way status split is Java's, reproduced deliberately:
	//   non-numeric id     -> 500 (NumberFormatException inside the Hibernate
	//                         usertype; a Java bug, pinned not fixed)
	//   numeric but absent -> 404
	//   found              -> 200
	// RECEPTION-GATED. Java checks it in the handler itself
	// (PatientMergeRestController.hasMergePermission ->
	// userRoleService.userInRole(userId, Constants.ROLE_RECEPTION)) and returns
	// ResponseEntity.status(FORBIDDEN).build() — a BODILESS 403.
	//
	// There is deliberately NO admin bypass. That check calls userInRole
	// directly; the `|| isUserAdmin(...)` fallback lives only in
	// ModuleAuthenticationInterceptor, which governs the separate module system.
	// Verified live: `is_admin='Y'` with no roles gets 403, and so does the
	// Global Administrator role — only Reception opens it. A port that "helpfully"
	// let admins through would diverge on both.
	web.Register(mux, "GET", "rest/patient/merge/details/{patientId}", authmw.RequireRole(
		"Reception",
		func(w http.ResponseWriter, r *http.Request) {
			dto, err := ctrl.MergeService.GetMergeDetails(r.PathValue("patientId"))
			if err != nil {
				var malformed service.ErrMalformedID
				if errors.As(err, &malformed) {
					http.Error(w, "internal error", http.StatusInternalServerError)
					return
				}
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if dto == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			web.WriteJSON(w, http.StatusOK, dto)
		}))

	// GET rest/patient-id-documents/{patientId} — mirrors
	// PatientManagementRestController.getIdDocuments.
	//
	// patientId is bound as a String against a VARCHAR column, so a
	// non-numeric value is NOT an error — it simply matches nothing and
	// returns 200 []. No 404 path exists on this endpoint at all.
	web.Register(mux, "GET", "rest/patient-id-documents/{patientId}", func(w http.ResponseWriter, r *http.Request) {
		dtos, err := ctrl.Service.GetIdDocuments(r.PathValue("patientId"))
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dtos)
	})

	// GET rest/patient-id-documents/{patientId}/{documentId}/full — mirrors
	// getIdDocumentFull.
	//
	// Asymmetric binding on ONE endpoint, reproduced: documentId is an Integer
	// in Java so a non-numeric value fails at binding -> 400, while patientId
	// (String/varchar, above) silently matches nothing. A port validating both
	// the same way would diverge on one of them.
	web.Register(mux, "GET", "rest/patient-id-documents/{patientId}/{documentId}/full", func(w http.ResponseWriter, r *http.Request) {
		documentID, err := strconv.ParseInt(r.PathValue("documentId"), 10, 32)
		if err != nil {
			http.Error(w, "invalid documentId", http.StatusBadRequest)
			return
		}
		dto, err := ctrl.Service.GetIdDocumentFull(r.PathValue("patientId"), documentID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dto)
	})

	// GET rest/patient-photos/{id}/{isThumbnail} — mirrors getPhoto.
	//
	// isThumbnail is a primitive boolean in Java, so Spring rejects anything
	// non-boolean at binding with a 400. Go's ParseBool is wired to answer 400
	// too rather than defaulting — a silent default would turn a client bug
	// into a wrong-image response.
	//
	// Spring's boolean binding also accepts on/off/yes/1/0; ParseBool accepts
	// 1/t/T/TRUE/true/True and their false counterparts. The overlap covers
	// every value the real frontend sends (`true`/`false`), and the divergence
	// on exotic spellings is noted rather than emulated.
	web.Register(mux, "GET", "rest/patient-photos/{id}/{isThumbnail}", func(w http.ResponseWriter, r *http.Request) {
		isThumbnail, err := strconv.ParseBool(r.PathValue("isThumbnail"))
		if err != nil {
			http.Error(w, "invalid isThumbnail", http.StatusBadRequest)
			return
		}
		dto, err := ctrl.Service.GetPhoto(r.PathValue("id"), isThumbnail)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		web.WriteJSON(w, http.StatusOK, dto)
	})
}
