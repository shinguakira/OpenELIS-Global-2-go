// Package service — the read path behind
// GET rest/patient/merge/details/{patientId}.
package service

import (
	"errors"
	"strconv"
	"strings"

	"openelis-go/internal/patient/daoimpl"
	"openelis-go/internal/patient/form"
)

// internalIdentityTypes mirrors PatientMergeServiceImpl's own
// `List.of("GUID","AKA","MOTHER","MOTHERS_INITIAL")` — identity types Java
// hides from the human-facing identifiers list.
//
// The list is applied ONLY to identifiers[]. totalIdentifiers is set from the
// UNFILTERED collection, so the count and the array length legitimately
// disagree whenever a patient has one of these. That asymmetry is reproduced
// deliberately (the c1 e2e spec asserts both halves): "fixing" the count would
// diverge from Java.
var internalIdentityTypes = map[string]bool{
	"GUID":            true,
	"AKA":             true,
	"MOTHER":          true,
	"MOTHERS_INITIAL": true,
}

// identityDisplayNames mirrors getDisplayNameForIdentityType. Java matches on
// the UPPERCASED type name and falls through to the raw name when unmapped,
// or "Unknown" when null. The UI renders these labels, so returning the raw DB
// code (e.g. "NATIONAL") instead of "National ID" would look structurally
// valid while breaking the screen.
var identityDisplayNames = map[string]string{
	"SUBJECT":           "Subject Number",
	"NATIONAL":          "National ID",
	"ST":                "ST Number",
	"INSURANCE":         "Insurance ID",
	"OCCUPATION":        "Occupation",
	"ORG_SITE":          "Organization Site",
	"EDUCATION":         "Education",
	"MARITIAL":          "Marital Status",
	"NATIONALITY":       "Nationality",
	"OTHER NATIONALITY": "Other Nationality",
	"HEALTH DISTRICT":   "Health District",
	"HEALTH REGION":     "Health Region",
	"OB_NUMBER":         "OB Number",
	"PC_NUMBER":         "PC Number",
}

func displayNameForIdentityType(raw string) string {
	if raw == "" {
		return "Unknown"
	}
	if mapped, ok := identityDisplayNames[strings.ToUpper(raw)]; ok {
		return mapped
	}
	return raw
}

// Analysis status NAMES that countResultsForPatient excludes. Java asks
// IStatusService for AnalysisStatus.Canceled / SampleRejected / NotStarted;
// StatusService.addToAnalysisMap maps those enum constants onto these exact
// status_of_sample.name values, so resolving by name here is the same lookup
// Java performs — and, unlike hardcoding ids, it survives a deployment whose
// status_of_sample table uses different numbers.
const (
	analysisStatusNotStarted     = "Not Tested"
	analysisStatusCanceled       = "Test Canceled"
	analysisStatusSampleRejected = "Sample Rejected"
	statusTypeAnalysis           = "ANALYSIS"
)

// StatusResolver is the slice of the common StatusService this package needs.
// Declared as an interface so the service layer does not depend on the
// concrete status service (and so it can be nil — see below).
type StatusResolver interface {
	IDByName(statusType, name string) string
}

// PatientMergeService backs GET rest/patient/merge/details/{patientId}.
type PatientMergeService struct {
	DAO *daoimpl.PatientDAOImpl
	// Status resolves the analysis statuses excluded from totalResults.
	//
	// If nil, NOTHING is excluded and totalResults counts every analysis —
	// which does NOT match Java. An earlier revision shipped exactly that and
	// reported 28 where Java reports 0, because every analysis in the dev
	// dataset is "Not Tested" (an excluded status). Always wire this.
	Status StatusResolver
}

// excludedResultStatusIDs resolves the three excluded statuses to ids.
// Returns nil when no resolver is wired, which the caller treats as "exclude
// nothing" — see the warning on the Status field.
func (s *PatientMergeService) excludedResultStatusIDs() []string {
	if s.Status == nil {
		return nil
	}
	ids := make([]string, 0, 3)
	for _, name := range []string{analysisStatusCanceled, analysisStatusSampleRejected, analysisStatusNotStarted} {
		// Java's getStatusID returns "-1" for an unmapped status and that
		// value is still added to its exclusion set, so a "-1" here is
		// harmless (it matches no real status_id) and is kept rather than
		// filtered — same set semantics as Java.
		ids = append(ids, s.Status.IDByName(statusTypeAnalysis, name))
	}
	return ids
}

// GetMergeDetails backs GET rest/patient/merge/details/{patientId}.
//
// Error contract, reproducing Java exactly:
//
//   - non-numeric id      -> ErrMalformedID (controller answers 500, matching
//     Java's NumberFormatException path)
//
//   - numeric but absent  -> (nil, nil) (controller answers 404)
//
//   - identity row with a NULL type -> ErrUnresolvableIdentityType (500, also
//     matching Java — see that error's doc)
//
// The "Reception" role gate Java applies to this endpoint IS reproduced, but a
// layer up: authmw.RequireRole wraps the route in the controller, which is
// where Go can express what Java writes as an in-handler check.
func (s *PatientMergeService) GetMergeDetails(patientID string) (*form.MergeDetailsDTO, error) {
	if _, err := strconv.ParseInt(patientID, 10, 64); err != nil {
		return nil, ErrMalformedID{ID: patientID}
	}

	p, err := s.DAO.GetPatientByID(patientID)
	if err != nil || p == nil {
		return nil, err
	}
	person, err := s.DAO.GetPersonByID(p.PersonID)
	if err != nil {
		return nil, err
	}

	identities, err := s.DAO.GetPatientIdentities(patientID)
	if err != nil {
		return nil, err
	}
	identifiers := make([]form.IdentifierDTO, 0, len(identities))
	for _, row := range identities {
		if row.IdentityTypeName == nil {
			// identity_type_id is NULL. Java loads the row anyway and then
			// calls patientIdentityTypeService.get(null) inside this same
			// loop, which throws — the request ends as a 500. Measured live by
			// seeding one such row.
			//
			// Returning an error here reproduces that. The alternatives are
			// both wrong: skipping the row understates totalIdentifiers and
			// answers 200 where Java errors, and inventing a placeholder type
			// name invents data Java never emits.
			return nil, ErrUnresolvableIdentityType
		}
		if internalIdentityTypes[strings.ToUpper(*row.IdentityTypeName)] {
			continue // hidden from the list, but still counted below
		}
		identifiers = append(identifiers, form.IdentifierDTO{
			IdentityType:  displayNameForIdentityType(*row.IdentityTypeName),
			IdentityValue: row.IdentityData,
		})
	}

	orders, err := s.DAO.CountOrdersForPatient(patientID)
	if err != nil {
		return nil, err
	}
	sampleItems, err := s.DAO.CountSampleItemsForPatient(patientID)
	if err != nil {
		return nil, err
	}
	results, err := s.DAO.CountResultsForPatient(patientID, s.excludedResultStatusIDs())
	if err != nil {
		return nil, err
	}

	summary := form.MergeDataSummaryDTO{
		TotalOrders: int(orders),
		// activeOrders is set to the SAME value as totalOrders in Java. The
		// comment above it describes a status filter that was never
		// implemented, so this is a known-unfinished field, not a real
		// active-order count. Copied knowingly.
		ActiveOrders: int(orders),
		TotalResults: int(results),
		// Java's `totalSamples` counts sample_ITEM rows; `totalOrders` counts
		// sample_human rows. The names read backwards from the tables — see
		// the DAO. Preserved.
		TotalSamples: int(sampleItems),
		// Never populated by getMergeDetails; they are Java primitives so they
		// still serialize, as 0.
		TotalDocuments:    0,
		TotalIdentifiers:  len(identities), // UNFILTERED — see internalIdentityTypes
		TotalContacts:     0,
		TotalRelations:    0,
		TotalAuditEntries: 0,
		// Non-nil so they serialize as [] rather than null; Java initialises
		// both to empty lists and never populates them on this path.
		ConflictingFields:        []string{},
		ConflictingIdentityTypes: []string{},
	}
	// Computed in Java, not stored: deliberately excludes contacts, relations
	// and auditEntries.
	summary.TotalDataItems = summary.TotalOrders + summary.TotalResults +
		summary.TotalSamples + summary.TotalDocuments + summary.TotalIdentifiers

	dto := form.MergeDetailsDTO{
		PatientID: strconv.FormatInt(p.ID, 10),
		Gender:    p.Gender,
		// birthDate here is the FORMATTED string (entered_birth_date), NOT the
		// epoch that patientByLabNumer emits under the same field name.
		BirthDate:         p.EnteredBirthDate,
		DataSummary:       summary,
		Identifiers:       identifiers,
		ConflictingFields: []string{},
	}
	if person != nil {
		dto.FirstName = person.FirstName
		dto.LastName = person.LastName
	}
	// nationalId / phoneNumber / email / address exist on the Java DTO but are
	// never populated here, so NON_NULL drops them. Left unset on purpose —
	// populating them would leak PHI this endpoint does not return.
	return &dto, nil
}

// ErrUnresolvableIdentityType signals a patient_identity row whose
// identity_type_id is NULL, so no type name can be resolved.
//
// Java does not guard this: getPatientIdentities loads every row with a plain
// `SELECT *`, and the identifier loop then calls
// patientIdentityTypeService.get(identity.getIdentityTypeId()) with null, which
// throws and surfaces as HTTP 500. Verified by seeding one such row against the
// live server.
//
// The controller maps this to 500 for that reason — pinning Java's behavior,
// not choosing it. The column is nullable and carries a FK, so NULL is the only
// way the type can fail to resolve; a dangling id cannot exist.
var ErrUnresolvableIdentityType = errors.New("patient_identity.identity_type_id is null; no identity type to resolve")
