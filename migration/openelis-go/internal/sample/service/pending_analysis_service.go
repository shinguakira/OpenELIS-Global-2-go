package service

import (
	"strconv"

	"openelis-go/internal/sample/daoimpl"
	"openelis-go/internal/sample/form"
)

// Analysis status NAMES for the four groups
// PendingAnalysisForTestProviderRestController builds, resolved through the
// same StatusService lookup Java uses in its static initialiser.
//
// The enum-to-name mapping is StatusService.addToAnalysisMap's, read from the
// source rather than inferred — several do not follow from the enum name:
// TechnicalRejected is "Technical Rejected" but BiologistRejected is
// "Biologist Rejection", and NotStarted is "Not Tested".
const (
	analysisStatusTechnicalRejected   = "Technical Rejected"
	analysisStatusBiologistRejected   = "Biologist Rejection"
	analysisStatusTechnicalAcceptance = "Technical Acceptance"
)

// PendingAnalysisService backs GET rest/getPendingAnalysisForTestProvider.
type PendingAnalysisService struct {
	DAO    *daoimpl.SampleDAOImpl
	Status StatusResolver
}

// GetPendingForTest mirrors createJsonGroupedAnalysis: four independent
// queries against the same test, one per status group, each under its own key.
//
// Java builds the response with org.json, not Jackson, so every group key is
// ALWAYS present even when its array is empty — Include.NON_NULL never sees
// this object. The DTO's non-omitempty slices reproduce that.
func (s *PendingAnalysisService) GetPendingForTest(testID string) (*form.PendingAnalysisDTO, error) {
	dto := &form.PendingAnalysisDTO{}
	targets := []struct {
		statusName string
		into       *[]form.PendingAnalysisItem
	}{
		{analysisStatusNotStarted, &dto.NotStarted},
		{analysisStatusTechnicalRejected, &dto.TechnicianRejection},
		{analysisStatusBiologistRejected, &dto.BiologistRejection},
		{analysisStatusTechnicalAcceptance, &dto.NotValidated},
	}

	for _, t := range targets {
		rows, err := s.DAO.AnalysesByTestAndStatus(testID, s.statusIDs(t.statusName))
		if err != nil {
			return nil, err
		}
		items := make([]form.PendingAnalysisItem, 0, len(rows))
		for _, r := range rows {
			items = append(items, form.PendingAnalysisItem{
				LabNo: r.LabNo,
				// STRING, not a number: the controller does
				// analysisObject.put("id", analysis.getId()) and getId()
				// returns a String. Contrast all-by-accession in this same
				// wave, which Integer.parseInt's the very same id — two
				// endpoints over one entity, two JSON types.
				ID: strconv.FormatInt(r.AnalysisID, 10),
			})
		}
		*t.into = items
	}
	return dto, nil
}

func (s *PendingAnalysisService) statusIDs(name string) []string {
	if s.Status == nil {
		return nil
	}
	return []string{s.Status.IDByName(statusTypeAnalysis, name)}
}
