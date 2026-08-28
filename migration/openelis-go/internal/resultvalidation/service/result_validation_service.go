// Package service ports AccessionValidationRestController's assembly
// (constitution.md Layer III). Folder layout mirrors the Java source.
package service

import (
	"fmt"

	"openelis-go/internal/common/util"
	"strconv"
	"strings"

	referraldaoimpl "openelis-go/internal/referral/daoimpl"
	"openelis-go/internal/resultvalidation/daoimpl"
	"openelis-go/internal/resultvalidation/form"
)

// ResultValidationService builds the rest/AccessionValidation response.
type ResultValidationService struct {
	DAO      *daoimpl.ResultValidationDAOImpl
	Sections *referraldaoimpl.ReferralDAOImpl // for the RAW-named section list
}

// Load ports showAccessionValidationRange.
//
// The three search inputs are mutually exclusive and checked in this order —
// accessionNumber, then date, then unitType — so supplying two silently uses
// the first. doRange then picks between a RANGE search and an exact one; see
// ByAccessionRange for why that distinction is visible to a caller.
func (s *ResultValidationService) Load(accessionNumber, date, unitType string, doRange bool) (*form.ResultValidationForm, error) {
	f := form.NewResultValidationForm()

	sections, err := s.DAO.UserTestSections()
	if err != nil {
		return nil, err
	}
	f.TestSections = sections
	byName, err := s.Sections.TestSectionsByName()
	if err != nil {
		return nil, err
	}
	f.TestSectionsByName = byName

	switch {
	case accessionNumber != "":
		f.AccessionNumber = accessionNumber
	case date != "":
		f.TestDate = date
	case unitType != "":
		f.TestSectionID = &unitType
	}

	// The search runs only when at least one of the three was supplied.
	if accessionNumber == "" && date == "" && unitType == "" {
		return &f, nil
	}

	var rows []daoimpl.ValidationRow
	if doRange {
		switch {
		case unitType != "":
			rows, err = s.DAO.BySection(unitType)
		case accessionNumber != "":
			rows, err = s.DAO.ByAccessionRange(accessionNumber)
		default:
			// The date branch is not ported: no analysis in this dataset has a
			// started_date to search on, so there is nothing to verify an
			// implementation against and guessing at it would be untested code.
			rows = nil
		}
	} else if accessionNumber != "" {
		rows, err = s.DAO.BySample(accessionNumber)
	}
	if err != nil {
		return nil, err
	}

	f.ResultList = buildItems(rows)
	f.Paging.SearchTermToPage = searchTermToPage(f.ResultList)
	f.SearchFinished = true
	return &f, nil
}

// buildItems ports testResultListToAnalysisItemList plus setGroupingNumbers.
func buildItems(rows []daoimpl.ValidationRow) []form.AnalysisItemDTO {
	items := make([]form.AnalysisItemDTO, 0, len(rows))
	for _, r := range rows {
		items = append(items, buildItem(r))
	}

	// setGroupingNumbers starts its counter at ONE and increments on the first
	// row, so the first group is 2, not 1 — the workplan version of the same
	// routine comments that "the header is always going to be 0". Reproduced.
	grouping := 1
	current := ""
	first := true
	headIndex := -1
	for i := range items {
		if first || items[i].AccessionNumber != current {
			current = items[i].AccessionNumber
			headIndex = i
			grouping++
			first = false
		} else {
			// Both the head of the group and this row are marked, which is why
			// the flag is set on two rows rather than one.
			items[headIndex].MultipleResultForSample = true
			items[i].MultipleResultForSample = true
		}
		items[i].SampleGroupingNumber = grouping
	}
	return items
}

func buildItem(r daoimpl.ValidationRow) form.AnalysisItemDTO {
	item := form.NewAnalysisItem()
	item.AccessionNumber = r.AccessionNumber
	item.AnalysisID = r.AnalysisID
	item.TestID = r.TestID
	item.TestName = r.TestName
	if r.TestSortNumber != nil {
		item.TestSortNumber = *r.TestSortNumber
	}
	if r.ResultID != nil {
		item.ResultID = *r.ResultID
	}
	if r.ResultValue != nil {
		item.Result = *r.ResultValue
	}
	if r.ResultType != nil {
		item.ResultType = *r.ResultType
	}
	if r.SigDigits != nil {
		item.SignificantDigits = *r.SigDigits
	}

	// patientName is "LAST FIRST" separated by a SPACE. c2's SampleEdit renders
	// the same person as "Last, First" with a comma — same two columns, two
	// formats, and only a fixture with both endpoints exercised shows it.
	item.PatientName = strings.TrimSpace(deref(r.PatientLastName) + " " + deref(r.PatientFirstName))

	// patientInfo reads ENTERED_birth_date — the raw text column — where
	// LogbookResults formats the parsed birth_date. The two disagree on this
	// patient: entered "01/15/1990" against a stored 1991-03-01.
	item.PatientInfo = strings.Join([]string{
		deref(r.NationalID), deref(r.Gender), deref(r.EnteredBirthDate),
	}, ", ")

	// units is the UOM plus the RESULT's own normal range in parentheses,
	// formatted to the RESULT's significant digits: "UI/L ( 1.00-9.00 )".
	// normalRange below is a DIFFERENT range — the TEST's result_limits row,
	// formatted to the test_result significant digits: "7 - 40". Two ranges,
	// two sources, two formats, in one row.
	item.Units = augmentUOMWithRange(deref(r.UnitOfMeasure), r.MinNormal, r.MaxNormal, r.SigDigits)
	item.NormalRange = displayReferenceRange(r.LimitLowNormal, r.LimitHighNormal, r.LimitSigDigits, " - ")

	// low/high critical are ±Infinity in result_limits and are mapped to 0
	// rather than emitted as infinities, which JSON cannot represent anyway.
	item.LowerCritical = 0
	item.HigherCritical = 0

	item.Normal = true
	return item
}

func augmentUOMWithRange(uom string, min, max *float64, sig *int) string {
	rng := displayReferenceRange(min, max, sig, "-")
	if rng == "" {
		return uom
	}
	return uom + " ( " + rng + " )"
}

// displayReferenceRange renders "<low><sep><high>" with the given number of
// decimals, or "" when the range is absent or unbounded.
func displayReferenceRange(low, high *float64, sig *int, sep string) string {
	if low == nil || high == nil {
		return ""
	}
	digits := 0
	if sig != nil {
		digits = *sig
	}
	return fmtFixed(*low, digits) + sep + fmtFixed(*high, digits)
}

func fmtFixed(v float64, digits int) string {
	if digits <= 0 {
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	return fmt.Sprintf("%.*f", digits, v)
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// searchTermToPage lists each DISTINCT accession in the result against the page
// it lands on, exactly as the workplan paging does. Every row fits on page 1
// here; the key set is what a port can get wrong, and leaving it empty is the
// easy mistake because the block renders fine without it.
func searchTermToPage(items []form.AnalysisItemDTO) []util.IdValuePair {
	out := []util.IdValuePair{}
	seen := map[string]bool{}
	for _, it := range items {
		if seen[it.AccessionNumber] {
			continue
		}
		seen[it.AccessionNumber] = true
		out = append(out, util.IdValuePair{Id: it.AccessionNumber, Value: "1"})
	}
	return out
}
