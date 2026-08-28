// Package service ports the org.openelisglobal.sample service layer for the c2
// reads. Per constitution.md Layer III, all data is compiled here so the
// controller only shapes a response.
package service

import (
	"errors"
	"strings"

	"openelis-go/internal/sample/daoimpl"
	"openelis-go/internal/sample/form"
)

// analysisStatusNotStarted is the status_of_sample.name that
// AnalysisStatus.NotStarted maps to (StatusService.addToAnalysisMap). Resolved
// by NAME rather than a hardcoded id so it survives a deployment whose
// status_of_sample table numbers differently — the same approach c1's merge
// service takes.
const (
	analysisStatusNotStarted = "Not Tested"
	statusTypeAnalysis       = "ANALYSIS"
)

// StatusResolver is the slice of the common StatusService this package needs.
type StatusResolver interface {
	IDByName(statusType, name string) string
}

// ErrUnassignedByAccessionBroken is returned unconditionally by
// GetUnassignedByAccession. See that method.
var ErrUnassignedByAccessionBroken = errors.New(
	"rest/sample/unassigned-by-accession is permanently broken in Java (invalid HQL); reproduced, not fixed")

// SampleService backs the c2 sample reads.
type SampleService struct {
	DAO *daoimpl.SampleDAOImpl
	// Status resolves the NotStarted analysis status. Required: without it the
	// all-by-accession list would include analyses in every status, which is
	// not what Java returns.
	Status StatusResolver
}

// GetAllByAccession backs GET rest/sample/all-by-accession/{accessionNumber},
// mirroring SampleRestController.getSampleByAccessionNumber + convertToForm.
//
// Java's shape, reproduced:
//   - no such sample                    -> (nil, nil), controller answers 404
//   - sample exists, no NotStarted rows -> ([], nil), controller answers 200 []
//
// That second case matters: this endpoint distinguishes "no such order" from
// "an order with nothing pending", while its siblings in the same wave answer
// 200 [] for both. There is no house rule — the c2 spec pins each separately.
//
// Only NotStarted analyses are listed. convertToForm builds its status set from
// statusService.getStatusID(AnalysisStatus.NotStarted) alone, so an order whose
// work has begun returns an empty list even though it has analyses.
func (s *SampleService) GetAllByAccession(accessionNumber string) ([]form.SampleSearchForm, error) {
	// getSampleByAccessionNumber returns null for a blank or whitespace-only
	// value before it queries anything, so a blank reaches the controller's
	// not-found branch rather than matching a row. Spring's path binding makes
	// this hard to hit through the URL, but the check is Java's and is cheap
	// to keep faithful.
	if strings.TrimSpace(accessionNumber) == "" {
		return nil, nil
	}
	sample, err := s.DAO.GetByAccessionNumber(normalizeAccession(accessionNumber))
	if err != nil || sample == nil {
		return nil, err
	}

	rows, err := s.DAO.AnalysesForSampleByStatus(sample.ID, s.notStartedStatusIDs())
	if err != nil {
		return nil, err
	}

	// Non-nil so an order with no pending analyses serializes as [] rather
	// than null — Java returns an empty ArrayList there.
	forms := make([]form.SampleSearchForm, 0, len(rows))
	for _, r := range rows {
		forms = append(forms, form.SampleSearchForm{
			ID:              sample.ID,
			AccessionNumber: sample.AccessionNumber,
			AnalysisID:      r.AnalysisID,
			SampleType:      r.SampleType,
			ReferralTest:    r.ReferralTest,
		})
	}
	return forms, nil
}

// GetUnassignedByAccession backs GET
// rest/sample/unassigned-by-accession/{accessionNumber} — which ALWAYS fails.
//
// SampleDAOImpl.getUnassignedSampleByAccessionNumber runs HQL referencing
// `r.canceled`, a property that does not exist on
// org.openelisglobal.referral.valueholder.Referral. Hibernate throws while
// PARSING the query, before touching data, so the failure is independent of
// input: a valid accession, a second valid one and a nonexistent one all
// produce it. The controller catches Exception and answers 500, which makes
// its own `return notFound()` branch unreachable dead code.
//
// PINNED, NOT FIXED. Migration policy is to reproduce Java's observable
// behavior and raise bugs separately; "porting" this endpoint therefore means
// reproducing a permanently-500 route. Repairing the HQL is a maintenance task
// for the Java side and is explicitly out of scope here.
//
// Reproduced as an error rather than a hardcoded 500 write so the controller
// keeps one error-to-status path, and so this shows up in logs the way a real
// failure would.
func (s *SampleService) GetUnassignedByAccession(string) ([]form.SampleSearchForm, error) {
	return nil, ErrUnassignedByAccessionBroken
}

// notStartedStatusIDs resolves the single status convertToForm filters on.
// Returns nil when no resolver is wired, which would list every analysis — the
// caller is wired at startup and treats a missing resolver as fatal.
func (s *SampleService) notStartedStatusIDs() []string {
	if s.Status == nil {
		return nil
	}
	return []string{s.Status.IDByName(statusTypeAnalysis, analysisStatusNotStarted)}
}

// normalizeAccession reproduces SampleServiceImpl's own truncation: everything
// from the first '.' is dropped before the lookup, so "E2E001.1" resolves to
// sample "E2E001". Applied to the ALL-by-accession path because
// getSampleByAccessionNumber does it too.
func normalizeAccession(accessionNumber string) string {
	if i := strings.Index(accessionNumber, "."); i >= 0 {
		return accessionNumber[:i]
	}
	return accessionNumber
}
