package service

import (
	"errors"
	"math"
	"sort"
	"strings"

	"openelis-go/internal/common/util"
	"openelis-go/internal/testcatalog/daoimpl"
	"openelis-go/internal/testcatalog/form"
)

// EditorTestService ports the test-level half of
// TestCatalogEditorRestController plus TestCatalogActivationRestController.
type EditorTestService struct {
	DAO *daoimpl.EditorTestDAO
	// ActiveLocale is the deployment's default language, used to label a
	// dictionary-backed option.
	ActiveLocale string
}

// ErrCodeInUse is the 409 the create flow answers when local_code is taken. It
// is its own status so the UI can flag the field rather than the form.
var ErrCodeInUse = errors.New("testcatalog: code in use")

// domains is the controller's own allow-list. The column takes any string; this
// is the only thing stopping one.
var domains = map[string]bool{"CLINICAL": true, "ENVIRONMENTAL": true, "VECTOR": true}

func (s *EditorTestService) locale() string {
	if s.ActiveLocale == "" {
		return "en"
	}
	return s.ActiveLocale
}

// ------------------------------------------------------------ create tests

// CreateTest ports createTest.
//
// Five fields are required and `labUnitId` is not among them, so a test can be
// created with no lab unit at all — and then it is invisible on Add Order,
// which filters by section.
func (s *EditorTestService) CreateTest(body form.CreateTestRequest, sysUserID int64) (*form.CreatedTest, error) {
	if strings.TrimSpace(body.Name) == "" || strings.TrimSpace(body.ReportingName) == "" ||
		strings.TrimSpace(body.Code) == "" || strings.TrimSpace(body.Domain) == "" ||
		!domains[body.Domain] || strings.TrimSpace(body.SampleTypeID) == "" {
		return nil, ErrUnprocessable
	}
	inUse, err := s.DAO.CodeInUse(body.Code)
	if err != nil {
		return nil, err
	}
	if inUse {
		return nil, ErrCodeInUse
	}

	description := ""
	if body.Description != nil {
		description = *body.Description
	}
	testID, err := s.DAO.CreateInactiveTest(daoimpl.CreateTestParams{
		Name:          body.Name,
		ReportingName: body.ReportingName,
		Code:          body.Code,
		LabUnitID:     body.LabUnitID,
		SampleTypeID:  body.SampleTypeID,
		Domain:        body.Domain,
		AMR:           body.AMR != nil && *body.AMR,
		Orderable:     body.Orderable != nil && *body.Orderable,
		Description:   description,
	}, sysUserID)
	if err != nil {
		return nil, err
	}
	return &form.CreatedTest{TestID: testID}, nil
}

// ------------------------------------------------------------- basic info

// GetBasicInfo ports getBasicInfo.
func (s *EditorTestService) GetBasicInfo(testID string) (*form.BasicInfo, error) {
	row, err := s.DAO.BasicInfo(testID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrNotFound
	}
	return basicInfoDTO(row), nil
}

// SaveBasicInfo ports saveBasicInfo.
//
// `name` is immutable here and the check is not a plain equality: a null name
// passes, and a submitted one is compared against the STORED name with a blank
// stored value treated as "". So sending the name back unchanged is fine and
// sending a new one is 422.
func (s *EditorTestService) SaveBasicInfo(testID string, body form.BasicInfo, sysUserID int64) (*form.BasicInfo, error) {
	row, err := s.DAO.BasicInfo(testID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrNotFound
	}
	if body.Domain != nil && !domains[*body.Domain] {
		return nil, ErrUnprocessable
	}
	if body.Name != nil {
		stored := ""
		if row.Name != nil {
			stored = *row.Name
		}
		if *body.Name != stored {
			return nil, ErrUnprocessable
		}
	}

	update := daoimpl.BasicInfoUpdate{
		Code:        body.Code,
		Description: body.Description,
		Domain:      body.Domain,
		AMR:         body.AntimicrobialResistance,
		Orderable:   body.Orderable,
		// active=true is deliberately dropped: activation is gated on range
		// coverage and has to go through POST .../activate, so Basic Info can
		// only ever persist a DEACTIVATION.
		Deactivate: body.Active != nil && !*body.Active,
	}
	if body.LabUnitID != nil {
		update.LabUnitID = *body.LabUnitID
	}
	if body.SampleTypeID != nil {
		update.SampleTypeID = *body.SampleTypeID
	}

	if err := s.DAO.SaveBasicInfo(testID, update, sysUserID); err != nil {
		return nil, err
	}
	saved, err := s.DAO.BasicInfo(testID)
	if err != nil {
		return nil, err
	}
	return basicInfoDTO(saved), nil
}

func basicInfoDTO(row *daoimpl.BasicInfoRow) *form.BasicInfo {
	amr, active, orderable := row.AMR, row.Active, row.Orderable
	return &form.BasicInfo{
		TestID: row.ID, Name: row.Name, Code: row.Code, Description: row.Description,
		Domain: row.Domain, LabUnitID: row.LabUnitID, SampleTypeID: row.SampleTypeID,
		AntimicrobialResistance: &amr, Active: &active, Orderable: &orderable,
	}
}

// -------------------------------------------------------- sample & results

// GetSampleResults ports getSampleResults.
func (s *EditorTestService) GetSampleResults(testID string) (*form.SampleResults, error) {
	ok, err := s.DAO.BasicInfo(testID)
	if err != nil {
		return nil, err
	}
	if ok == nil {
		return nil, ErrNotFound
	}
	return s.sampleResults(testID)
}

// SaveSampleResults ports saveSampleResults.
//
// Every component needs a code and a label, and the codes must be unique within
// the request — the DB enforces (test_id, code) as well, but rejecting here
// keeps a raw 500 off the wire.
func (s *EditorTestService) SaveSampleResults(testID string, body form.SampleResults, sysUserID int64) (*form.SampleResults, error) {
	row, err := s.DAO.BasicInfo(testID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrNotFound
	}

	codes := map[string]bool{}
	desired := make([]daoimpl.DesiredComponent, 0, len(body.Components))
	for _, c := range body.Components {
		if strings.TrimSpace(c.Code) == "" || strings.TrimSpace(c.Label) == "" {
			return nil, ErrUnprocessable
		}
		if codes[c.Code] {
			return nil, ErrUnprocessable
		}
		codes[c.Code] = true

		want := daoimpl.DesiredComponent{
			ID: c.ID, Code: c.Code, Label: c.Label,
			DisplayOrder:          intOrZero(c.DisplayOrder),
			ResultType:            c.ResultType,
			UomID:                 c.UomID,
			SignificantDigits:     c.SignificantDigits,
			DefaultResult:         c.DefaultResult,
			AllowMultipleReadings: c.AllowMultipleReadings != nil && *c.AllowMultipleReadings,
		}
		for _, i := range c.Interpretations {
			want.Interpretations = append(want.Interpretations, daoimpl.DesiredInterpretation{
				ID: i.ID, ValueMatch: i.ValueMatch, Text: i.Text, Severity: i.Severity,
				Color: i.Color, DisplayOrder: intOrZero(i.DisplayOrder),
			})
		}
		for _, o := range c.Options {
			// An option with no result type of its own inherits the COMPONENT's,
			// which is how a select list gets its D/M/C marker.
			resultType := o.ResultType
			if resultType == nil {
				resultType = c.ResultType
			}
			want.Options = append(want.Options, daoimpl.DesiredOption{
				ID: o.ID, Value: o.Value, SortOrder: o.SortOrder,
				Normal: o.Normal != nil && *o.Normal, ResultType: resultType,
			})
		}
		desired = append(desired, want)
	}

	if err := s.DAO.SaveSampleResults(testID, desired, sysUserID); err != nil {
		return nil, err
	}
	return s.sampleResults(testID)
}

// CopySampleResults ports copySampleResults.
//
// Only the TARGET is checked. A source id naming nothing simply has no active
// components, so the request is a 200 that copied nothing.
func (s *EditorTestService) CopySampleResults(testID, sourceID string, sysUserID int64) (*form.SampleResults, error) {
	row, err := s.DAO.BasicInfo(testID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrNotFound
	}
	if err := s.DAO.CopyComponentsFromTest(sourceID, testID, s.locale(), sysUserID); err != nil {
		return nil, err
	}
	return s.sampleResults(testID)
}

func (s *EditorTestService) sampleResults(testID string) (*form.SampleResults, error) {
	components, err := s.DAO.ActiveComponents(testID)
	if err != nil {
		return nil, err
	}
	out := &form.SampleResults{TestID: testID, Components: []form.ResultComponentDTO{}}
	for _, c := range components {
		displayOrder, allowMultiple := c.DisplayOrder, c.AllowMultipleReadings
		dto := form.ResultComponentDTO{
			ID: c.ID, Code: c.Code, Label: c.Label, DisplayOrder: &displayOrder,
			ResultType: c.ResultType, UomID: c.UomID, SignificantDigits: c.SignificantDigits,
			DefaultResult: c.DefaultResult, AllowMultipleReadings: &allowMultiple,
			Interpretations: []form.InterpretationDTO{}, Options: []form.OptionDTO{},
		}
		interps, err := s.DAO.ActiveInterpretations(c.ID)
		if err != nil {
			return nil, err
		}
		for _, i := range interps {
			order := i.DisplayOrder
			dto.Interpretations = append(dto.Interpretations, form.InterpretationDTO{
				ID: i.ID, ValueMatch: i.ValueMatch, Text: i.Text, Severity: i.Severity,
				Color: i.Color, DisplayOrder: &order,
			})
		}
		options, err := s.DAO.ActiveOptions(c.ID, s.locale())
		if err != nil {
			return nil, err
		}
		for _, o := range options {
			dto.Options = append(dto.Options, form.OptionDTO{
				ID: o.ID, Value: o.Value, ValueName: o.ValueName, ResultType: o.ResultType,
				SortOrder: parseIntOrNil(o.SortOrder), Normal: o.Normal,
			})
		}
		out.Components = append(out.Components, dto)
	}
	return out, nil
}

// ---------------------------------------------------------------- activation

// ErrCoverageGap is the 409 the activation gate answers with. The report rides
// along as the body, which is why it is carried on the error.
type ErrCoverageGap struct {
	Report *form.CoverageReport
}

func (e *ErrCoverageGap) Error() string { return "testcatalog: reference range coverage gap" }

// Activate ports activateTest.
//
// The gate is on GAPS only — an OVERLAP is reported and does not block. An
// acknowledgment is any non-blank string, and it is stored verbatim as jsonb,
// so a non-JSON acknowledgment is a 500 from the database rather than a 422.
func (s *EditorTestService) Activate(testID string, body *form.ActivateRequest, sysUserID int64) (*form.CoverageReport, error) {
	row, err := s.DAO.BasicInfo(testID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrNotFound
	}

	limits, err := s.DAO.LimitAges(testID)
	if err != nil {
		return nil, err
	}
	report := validateCoverage(limits)

	acknowledged := body != nil && body.GapsAcknowledged != nil &&
		strings.TrimSpace(*body.GapsAcknowledged) != ""
	if hasGaps(report) && !acknowledged {
		return nil, &ErrCoverageGap{Report: report}
	}

	var ack *string
	if hasGaps(report) {
		ack = body.GapsAcknowledged
	}
	if err := s.DAO.Activate(testID, ack, sysUserID); err != nil {
		return nil, err
	}
	return report, nil
}

func hasGaps(r *form.CoverageReport) bool {
	return (r.Male != nil && len(r.Male.Gaps) > 0) || (r.Female != nil && len(r.Female.Gaps) > 0)
}

// coverageEpsilon is the tolerance the validator compares ages with — well
// under an hour expressed in years.
const coverageEpsilon = 1e-9

// validateCoverage ports RangeCoverageValidationService.validate.
func validateCoverage(limits []daoimpl.LimitAge) *form.CoverageReport {
	return &form.CoverageReport{
		Male:   coverageForSex(limits, "M"),
		Female: coverageForSex(limits, "F"),
	}
}

// coverageForSex walks the sorted age windows tracking the frontier covered
// from age 0.
//
// The tail rule is the one that surprises: only an OPEN-ENDED top band
// (max = +Infinity) covers to the top of the reportable lifetime, so bands
// 0–15 and 15–30 leave 30+ uncovered and the test cannot be activated without
// an acknowledgment. A test with no ranges at all is EMPTY, not GAP — and
// EMPTY does not block.
func coverageForSex(limits []daoimpl.LimitAge, sex string) *form.SexCoverage {
	coverage := &form.SexCoverage{
		Sex: sex, Gaps: []form.AgeInterval{}, Overlaps: []form.AgeInterval{},
	}

	applicable := []daoimpl.LimitAge{}
	for _, l := range limits {
		if appliesToSex(l.Gender, sex) {
			applicable = append(applicable, l)
		}
	}
	if len(applicable) == 0 {
		coverage.Status = "EMPTY"
		return coverage
	}

	// Collections.sort is stable, so equal minAges keep the DAO's order.
	sort.SliceStable(applicable, func(i, j int) bool {
		return applicable[i].MinAge < applicable[j].MinAge
	})

	coveredTo := 0.0
	for _, l := range applicable {
		if l.MinAge > coveredTo+coverageEpsilon {
			coverage.Gaps = append(coverage.Gaps, form.AgeInterval{FromAge: util.JavaDouble(coveredTo), ToAge: util.JavaDouble(l.MinAge)})
		} else if l.MinAge < coveredTo-coverageEpsilon {
			coverage.Overlaps = append(coverage.Overlaps,
				form.AgeInterval{FromAge: util.JavaDouble(l.MinAge), ToAge: util.JavaDouble(math.Min(coveredTo, l.MaxAge))})
		}
		coveredTo = math.Max(coveredTo, l.MaxAge)
	}

	if !math.IsInf(coveredTo, 0) && !math.IsNaN(coveredTo) {
		coverage.Gaps = append(coverage.Gaps,
			form.AgeInterval{FromAge: util.JavaDouble(coveredTo), ToAge: util.JavaDouble(math.Inf(1))})
	}

	switch {
	case len(coverage.Gaps) > 0:
		// Gaps drive the safety gate, so they outrank overlaps for the status.
		coverage.Status = "GAP"
	case len(coverage.Overlaps) > 0:
		coverage.Status = "OVERLAP"
	default:
		coverage.Status = "COMPLETE"
	}
	return coverage
}

// appliesToSex: a blank or "A"/"ALL"/"B"/"BOTH" gender counts toward BOTH
// sexes, so one gender-less range can complete the coverage for each.
func appliesToSex(gender, sex string) bool {
	g := strings.ToUpper(strings.TrimSpace(gender))
	if g == "" || g == "A" || g == "ALL" || g == "B" || g == "BOTH" {
		return true
	}
	return g == sex
}

func intOrZero(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}
