// Package service ports LogbookResultsRestController and
// AccessionResultsRestController (constitution.md Layer III).
package service

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"openelis-go/internal/common/util"
	"openelis-go/internal/result/daoimpl"
	"openelis-go/internal/result/form"
)

// ResultService builds both result-carrying reads.
type ResultService struct {
	DAO *daoimpl.ResultDAOImpl
}

// ErrTypelessSampleItem is the NPE Java throws from getTestDisplayName.
var ErrTypelessSampleItem = errors.New("analysis on a sample item with no type of sample")

// Logbook ports rest/LogbookResults.
//
// Both parameters are optional and, when neither is supplied, no search runs —
// which is why the unparameterised call returns an empty testResult and a spec
// that only made that call could assert nothing about the rows.
func (s *ResultService) Logbook(labNumber, selectedTest string) (*form.LogbookResultsForm, error) {
	f := form.NewLogbookResultsForm()
	f.CurrentDate = currentDate()

	var rows []daoimpl.LogbookRow
	var err error
	switch {
	case labNumber != "":
		acc := labNumber
		f.AccessionNumber = &acc
		rows, err = s.DAO.ByAccessionNumber(labNumber)
	case selectedTest != "":
		// JAVA DEFECT, reproduced: checked before the rows are built so the
		// port fails on exactly the inputs Java fails on.
		poisoned, perr := s.DAO.TestHasTypelessItem(selectedTest)
		if perr != nil {
			return nil, perr
		}
		if poisoned {
			return nil, ErrTypelessSampleItem
		}
		rows, err = s.DAO.ByTest(selectedTest)
	default:
		return &f, nil
	}
	if err != nil {
		return nil, err
	}

	f.TestResult = buildRows(rows)
	f.Paging.SearchTermToPage = searchTermToPage(f.TestResult)
	// searchFinished is set on the labNumber path ONLY. The selectedTest path
	// returns rows with the flag still false — the screen shows results while
	// reporting that no search completed.
	f.SearchFinished = labNumber != ""
	return &f, nil
}

// AccessionResults ports rest/accession-results — the same rows in a lean
// envelope with no form fields at all.
func (s *ResultService) AccessionResults(accessionNumber string) (*form.AccessionResultsResponse, error) {
	out := form.AccessionResultsResponse{TestResult: []form.TestResultRowDTO{}}
	if accessionNumber == "" {
		return &out, nil
	}
	acc := accessionNumber
	out.AccessionNumber = &acc
	rows, err := s.DAO.ByAccessionNumberDesc(accessionNumber)
	if err != nil {
		return nil, err
	}
	out.TestResult = buildRows(rows)
	// The ROWS carry no patient at all here — patientName and patientInfo come
	// out as a single SPACE (last + " " + first over two empty strings),
	// nationalId as "" and patientId is absent entirely. The patient is instead
	// flattened onto the response ROOT below. LogbookResults does the opposite,
	// repeating the patient on every row.
	for i := range out.TestResult {
		out.TestResult[i].PatientName = " "
		out.TestResult[i].PatientInfo = " "
		out.TestResult[i].NationalID = ""
		out.TestResult[i].PatientID = nil
	}
	if err := s.attachEntityGraph(accessionNumber, out.TestResult); err != nil {
		return nil, err
	}
	if len(rows) > 0 {
		r := rows[0]
		out.FirstName = strPtr(deref(r.PatientFirstName))
		out.LastName = strPtr(deref(r.PatientLastName))
		out.DOB = strPtr(deref(r.EnteredBirthDate))
		out.Gender = strPtr(deref(r.Gender))
		// st and subjectNumber are emitted EMPTY, not omitted: the response
		// bean initialises them to "" and Include.NON_NULL keeps an empty
		// string.
		out.ST = strPtr("")
		out.SubjectNumber = strPtr("")
		out.NationalID = strPtr(deref(r.NationalID))
	}
	out.SearchFinished = true
	return &out, nil
}

func buildRows(rows []daoimpl.LogbookRow) []form.TestResultRowDTO {
	items := make([]form.TestResultRowDTO, 0, len(rows))
	for _, r := range rows {
		items = append(items, buildRow(r))
	}

	// Same grouping routine as the validation list: the counter starts at ONE
	// and increments on the first row, so the first group is 2. showSampleDetails
	// marks the head of each group.
	// The logbook groups by SAMPLE ITEM, not by accession: an order with two
	// items produces two groups under one accession, and showSampleDetails
	// marks the head of each. The validation list next door groups by accession
	// with the same counter, so the two disagree on any multi-item order.
	grouping := 1
	current := ""
	first := true
	for i := range items {
		if first || items[i].SampleItemID != current {
			current = items[i].SampleItemID
			grouping++
			first = false
			items[i].ShowSampleDetails = true
		}
		items[i].SampleGroupingNumber = grouping
	}
	return items
}

func buildRow(r daoimpl.LogbookRow) form.TestResultRowDTO {
	row := form.NewTestResultRow()
	row.AccessionNumber = r.AccessionNumber
	row.SampleItemID = r.SampleItemID
	row.AnalysisID = r.AnalysisID
	row.AnalysisStatusID = r.AnalysisStatusID
	row.TestID = r.TestID
	row.TestName = r.TestName

	row.SequenceNumber = deref(r.SequenceNumber)
	row.SampleType = deref(r.SampleType)
	row.TestDate = deref(r.TestDate)
	row.ReceivedDate = deref(r.ReceivedDate)
	row.TestSortOrder = deref(r.TestSortOrder)
	row.UnitsOfMeasure = deref(r.UnitOfMeasure)
	if r.AnalysisType != nil && *r.AnalysisType != "" {
		row.AnalysisMethod = *r.AnalysisType
	}
	if r.IsReportable != nil && *r.IsReportable != "" {
		row.Reportable = *r.IsReportable
	}

	// The normal range is emitted BOTH as two numbers and as a formatted
	// string, and the string is rounded to the test_result significant digits
	// while the numbers are raw.
	row.LowerNormalRange = derefF(r.LowNormal)
	row.UpperNormalRange = derefF(r.HighNormal)
	// The "abnormal" pair is result_limits' VALID range, not a separate
	// abnormal one — low_valid/high_valid under a different name.
	row.LowerAbnormalRange = derefF(r.LowValid)
	row.UpperAbnormalRange = derefF(r.HighValid)
	row.NormalRange = displayRange(r.LowNormal, r.HighNormal, r.LimitSigDigits)
	row.ResultLimitID = deref(r.LimitID)

	// significantDigits comes from TEST_RESULT, not from the result row —
	// AccessionValidation reads the result's own value for the same-named
	// field, so the two endpoints disagree on this patient's rows (0 here, 2
	// there).
	row.SignificantDigits = derefI(r.LimitSigDigits)

	// referredOut and its two ids come from the referral table; a row with no
	// referral omits the ids rather than emitting empty strings.
	row.ReferredOut = r.ReferredOut
	row.ShadowReferredOut = r.ReferredOut
	if r.ReferralID != nil {
		row.ReferralID = r.ReferralID
	}
	if r.ReferralReasonID != nil {
		row.ReferralReasonID = r.ReferralReasonID
	}
	if r.FallbackResultType != nil {
		row.ResultType = *r.FallbackResultType
	}

	if r.PatientID != nil {
		row.PatientID = r.PatientID
	}
	row.NationalID = deref(r.NationalID)
	// "Last, First" with a COMMA here, where AccessionValidation renders the
	// same two columns as "Last First" with a space.
	row.PatientName = strings.TrimSpace(deref(r.PatientLastName)) + ", " + strings.TrimSpace(deref(r.PatientFirstName))
	// patientInfo formats the parsed BIRTH_DATE, where AccessionValidation
	// emits the raw entered_birth_date text column. The two disagree whenever
	// what was typed is not what was stored.
	row.PatientInfo = strings.Join([]string{deref(r.NationalID), deref(r.Gender), deref(r.BirthDate)}, ", ")

	if r.ResultID != nil {
		row.ResultID = *r.ResultID
		row.ResultType = deref(r.ResultType)
		row.ResultValue = deref(r.ResultValue)
		// shadowResultValue is the same value under a second key — the form
		// keeps a copy to detect edits.
		row.ShadowResultValue = row.ResultValue
		row.ResultValueLog = logOf(row.ResultValue)
		row.Result = &form.ResultRefDTO{
			IsActive:          "Y",
			ID:                *r.ResultID,
			SignificantDigits: derefI(r.LimitSigDigits),
			Grouping:          derefI(r.Grouping),
		}
	}
	return row
}

// logOf renders the base-10 logarithm of a numeric result to two decimals,
// which is what the logbook shows in resultValueLog. A non-numeric or absent
// value leaves the bean's "--" initialiser in place.
func logOf(v string) string {
	f, err := strconv.ParseFloat(v, 64)
	if err != nil || f <= 0 {
		if err == nil && f == 0 {
			return "0"
		}
		return "--"
	}
	l := math.Log10(f)
	if l == math.Trunc(l) {
		return strconv.FormatFloat(l, 'f', -1, 64)
	}
	return fmt.Sprintf("%.2f", l)
}

func displayRange(low, high *float64, sig *int) string {
	if low == nil || high == nil {
		return ""
	}
	d := derefI(sig)
	return fmtFixed(*low, d) + " - " + fmtFixed(*high, d)
}

func fmtFixed(v float64, digits int) string {
	if digits <= 0 {
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	return fmt.Sprintf("%.*f", digits, v)
}

func searchTermToPage(items []form.TestResultRowDTO) []util.IdValuePair {
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

// currentDate is DateUtil.getCurrentDateAsText — dd/MM/yyyy.
func currentDate() string { return time.Now().Format("02/01/2006") }

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefF(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

func derefI(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func strPtr(s string) *string { return &s }
