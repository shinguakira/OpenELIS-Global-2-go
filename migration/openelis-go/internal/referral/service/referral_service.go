// Package service ports ReferredOutTestsRestController.setupPageForDisplay
// (constitution.md Layer III). Folder layout mirrors the Java source.
package service

import (
	"errors"

	commonservices "openelis-go/internal/common/services"
	"openelis-go/internal/referral/daoimpl"
	"openelis-go/internal/referral/form"
	sampleform "openelis-go/internal/sample/form"
)

// ReferredOutTestsService builds the rest/ReferredOutTests response.
type ReferredOutTestsService struct {
	DAO   *daoimpl.ReferralDAOImpl
	Lists *commonservices.DisplayListService
}

// Load ports setupPageForDisplay.
//
//	if (form.getSearchType() != null) {
//	    form.setReferralDisplayItems(referralService.getReferralItems(form));
//	    form.setSearchFinished(true);
//	}
//	form.setTestSelectionList(... ALL_TESTS);
//	form.setTestUnitSelectionList(... TEST_SECTION_BY_NAME);
//
// So the two reference lists are attached unconditionally and the search runs
// only when a searchType was supplied — which is why the unparameterised call
// can never return items, and why a spec that only ever made that call could
// assert nothing about them.
func (s *ReferredOutTestsService) Load(searchType, labNumber string) (*form.ReferredOutTestsForm, error) {
	f := form.NewReferredOutTestsForm()

	if searchType != "" {
		items, err := s.searchItems(searchType, labNumber)
		if err != nil {
			return nil, err
		}
		st := searchType
		f.SearchType = &st
		f.ReferralDisplayItems = &items
		f.SearchFinished = true
	}
	// labNumber is echoed whenever it was bound, independently of whether a
	// search ran — it is a form field, not a search result.
	if labNumber != "" {
		ln := labNumber
		f.LabNumber = &ln
	}

	// TEST_SECTION_BY_NAME, not TEST_SECTION — the raw name column rather than
	// the localized value. See TestSectionsByName for why the two differ.
	sections, err := s.DAO.TestSectionsByName()
	if err != nil {
		return nil, err
	}
	f.TestUnitSelectionList = sections

	tests, err := s.DAO.AllActiveTests()
	if err != nil {
		return nil, err
	}
	f.TestSelectionList = tests

	genders, err := s.Lists.Genders()
	if err != nil {
		return nil, err
	}
	f.PatientSearch = sampleform.PatientSearchDTO{
		DefaultHeader:  true,
		Genders:        genders,
		SearchCriteria: s.Lists.PatientSearchCriteria(),
	}
	return &f, nil
}

// searchItems ports ReferralServiceImpl.getReferralItems' switch.
//
// TEST_AND_DATES is a JAVA DEFECT reproduced deliberately: with no dates bound
// it throws rather than reporting a validation error, so the endpoint answers
// 500. PATIENT with no patient selected returns an empty list, not an error —
// two in-enum values, two different failure modes.
func (s *ReferredOutTestsService) searchItems(searchType, labNumber string) ([]form.ReferralDisplayItemDTO, error) {
	switch searchType {
	case "LAB_NUMBER":
		rows, err := s.DAO.ByAccessionNumber(labNumber)
		if err != nil {
			return nil, err
		}
		return convert(rows), nil
	case "TEST_AND_DATES":
		return nil, ErrTestAndDatesUnbound
	default: // PATIENT, with no patient id bound
		return []form.ReferralDisplayItemDTO{}, nil
	}
}

func convert(rows []daoimpl.ReferralDisplayRow) []form.ReferralDisplayItemDTO {
	out := make([]form.ReferralDisplayItemDTO, 0, len(rows))
	for _, r := range rows {
		item := form.ReferralDisplayItemDTO{
			AccessionNumber: r.AccessionNumber,
			AnalysisID:      r.AnalysisID,
		}
		// convertTimestampToStringDate on a null sent_date yields "", not a
		// dropped key — the setter always runs.
		if r.ReferredSendDate != nil {
			item.ReferredSendDate = *r.ReferredSendDate
		}
		if r.Status != nil {
			// The enum and its toString, emitted under two keys.
			item.ReferralStatus = *r.Status
			item.ReferralStatusDisplay = *r.Status
		}
		if r.PatientLastName != nil {
			item.PatientLastName = *r.PatientLastName
		}
		if r.PatientFirstName != nil {
			item.PatientFirstName = *r.PatientFirstName
		}
		if r.TestName != nil {
			item.ReferringTestName = *r.TestName
		}
		// referenceLabDisplay is set only inside `if (organization != null)`.
		if r.OrganizationName != nil {
			name := *r.OrganizationName
			item.ReferenceLabDisplay = &name
		}

		// The two result-dependent fields, set TOGETHER inside
		// `if (!resultList.isEmpty())`. A numeric result is rendered with its
		// UNIT appended — getAppropriateResultValue returns "13.75 UI/L", not the
		// bare number — so the unit is part of the value rather than a field of
		// its own.
		if r.ResultValue != nil && *r.ResultValue != "" {
			display := *r.ResultValue
			if r.UnitOfMeasure != nil && *r.UnitOfMeasure != "" {
				display += " " + *r.UnitOfMeasure
			}
			item.ReferralResultsDisplay = &display
			if r.CompletedDate != nil {
				d := *r.CompletedDate
				item.ResultDate = &d
			}
		}

		// notes is INDEPENDENT of the result: getNotesAsString runs either way
		// and returns null when the analysis carries none.
		if r.Notes != nil && *r.Notes != "" {
			n := *r.Notes
			item.Notes = &n
		}
		out = append(out, item)
	}
	return out
}

// ErrTestAndDatesUnbound is the exception TEST_AND_DATES throws when no dates
// were supplied. A JAVA DEFECT, reproduced: the value is inside the SearchType
// enum so it binds cleanly, and only then does the search blow up — a 500 where
// a validation error belongs.
var ErrTestAndDatesUnbound = errors.New("referred-out search by test and dates with no dates bound")
