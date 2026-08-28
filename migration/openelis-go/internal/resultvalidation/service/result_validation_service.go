// Package service ports AccessionValidationRestController's assembly
// (constitution.md Layer III). Folder layout mirrors the Java source.
package service

import (
	"fmt"
	"math"

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
func (s *ResultValidationService) Load(systemUserID, accessionNumber, date, unitType string, doRange bool) (*form.ResultValidationForm, error) {
	f := form.NewResultValidationForm()

	// The caller's lab units bound BOTH lists below. Java resolves them with
	// getUserTestSections(user, ROLE_VALIDATION) and then filters the results
	// with filterAnalysisResultsByLabUnitRoles — two separate calls, and a port
	// that skips them serves every section and every result to every caller.
	roleID, err := s.DAO.RoleIDByName(roleValidation)
	if err != nil {
		return nil, err
	}
	units, err := s.DAO.UserLabUnits(systemUserID, roleID)
	if err != nil {
		return nil, err
	}

	sections, err := s.DAO.UserTestSections(units)
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
			rows, err = s.DAO.ByStartedDate(date)
		}
	} else if accessionNumber != "" {
		rows, err = s.DAO.BySample(accessionNumber)
	}
	if err != nil {
		return nil, err
	}

	// filterAnalysisResultsByLabUnitRoles. The section list above is what the
	// screen offers; this is what it may actually show.
	rows = filterByLabUnits(rows, units)

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
		// getFormattedResult runs the value through getResultValue, which pads a
		// numeric result to the RESULT row's significant digits — 6.5 on a
		// 2-digit result goes out as "6.50".
		item.Result = fmtFixedValue(*r.ResultValue, derefI(r.SigDigits))
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

	// The criticals, and the reason a numeric field can answer with the JSON
	// STRING "Infinity".
	//
	// setResultLimitDependencies folds the sentinels to zero:
	//
	//     lowCritical  == NEGATIVE_INFINITY ? 0 : lowCritical
	//     highCritical == POSITIVE_INFINITY ? 0 : highCritical
	//
	// Every stored row here holds exactly those sentinels, so a resolved limit
	// yields 0 and 0 — which is why a port that hardcoded both to 0 agreed with
	// Java on every row until a test with age bands was seeded.
	//
	// The third case is the defect. getResultLimitForTestAndPatient does NOT
	// return null when the test HAS limits but none is the default band:
	// defaultResultLimit falls through to `return new ResultLimit()`, and that
	// bean initialises
	//
	//     private double lowCritical = Double.POSITIVE_INFINITY;
	//
	// where every other low bound on it is NEGATIVE_INFINITY. The guard tests
	// for NEGATIVE_INFINITY, misses, and +Infinity reaches Jackson — which has
	// no JSON number for it and writes the string "Infinity". highCritical
	// carries the same initialiser and IS caught, so it comes back 0: one row,
	// two bounds, and they disagree about what an unset critical looks like.
	//
	// A test with NO result_limits rows returns null instead, the dependencies
	// are never set, and both stay at the bean default 0.
	switch {
	case r.HasDefaultLimit:
		item.LowerCritical = criticalOrZero(r.LimitLowCritical, math.Inf(-1))
		item.HigherCritical = criticalOrZero(r.LimitHighCritical, math.Inf(1))
	case r.HasAnyLimit:
		item.LowerCritical = util.JavaDouble(math.Inf(1))
		item.HigherCritical = 0
	default:
		item.LowerCritical = 0
		item.HigherCritical = 0
	}

	// nonconforming is set for a technically REJECTED analysis:
	//   testResultItem.isNonconforming() || matches(statusId, TechnicalRejected)
	item.Nonconforming = r.StatusIsRejected

	// normal compares the value against the RESOLVED limit. With no limit — the
	// age-banded case where this endpoint resolves none — there is nothing to be
	// abnormal against, so it stays true.
	item.Normal = isNormalResult(r)
	return item
}

// criticalOrZero folds one sentinel to zero and passes everything else through,
// including the OTHER infinity — the asymmetry is the point.
func criticalOrZero(v *float64, sentinel float64) util.JavaDouble {
	if v == nil || *v == sentinel {
		return 0
	}
	return util.JavaDouble(*v)
}

// isNormalResult ports ResultsValidationUtility.isNormalResult for the numeric
// case: inside the limit's normal range is normal, outside is not. A missing
// limit or a non-numeric value is normal.
func isNormalResult(r daoimpl.ValidationRow) bool {
	if r.LimitLowNormal == nil || r.LimitHighNormal == nil || r.ResultValue == nil {
		return true
	}
	v, err := strconv.ParseFloat(*r.ResultValue, 64)
	if err != nil {
		return true
	}
	return v >= *r.LimitLowNormal && v <= *r.LimitHighNormal
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

// fmtFixedValue pads a numeric result string to the given decimals, passing
// through anything non-numeric or unpadded.
func fmtFixedValue(v string, digits int) string {
	if digits <= 0 || v == "" {
		return v
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return v
	}
	return fmt.Sprintf("%.*f", digits, f)
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

// roleValidation is Constants.ROLE_VALIDATION — the role whose lab-unit grants
// decide what this endpoint may return.
const roleValidation = "Validation"

// filterByLabUnits ports filterAnalysisResultsByLabUnitRoles: keep only the
// analyses whose test section is one the caller holds the role on. A caller
// with the AllLabUnits sentinel keeps everything.
func filterByLabUnits(rows []daoimpl.ValidationRow, units []string) []daoimpl.ValidationRow {
	if daoimpl.HasAllLabUnits(units) {
		return rows
	}
	allowed := make(map[string]bool, len(units))
	for _, u := range units {
		allowed[u] = true
	}
	out := make([]daoimpl.ValidationRow, 0, len(rows))
	for _, r := range rows {
		if r.TestSectionID != nil && allowed[*r.TestSectionID] {
			out = append(out, r)
		}
	}
	return out
}

// derefI is zero for a nil count.
func derefI(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}
