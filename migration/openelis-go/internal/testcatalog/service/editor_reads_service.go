package service

import (
	"sort"
	"strings"

	"openelis-go/internal/testcatalog/daoimpl"
	"openelis-go/internal/testcatalog/form"
)

// EditorReadService ports the ten test-catalog reads that pair with no write.
type EditorReadService struct {
	DAO *daoimpl.EditorReadDAO
}

// v1Sections is the SideNav order the envelope advertises. Compliance (v2) is
// hidden entirely in v1, so the set is the same for every domain — the field
// exists so the shell can branch once v2 lights up domain-conditional
// visibility.
var v1Sections = []string{
	"basic-info", "sample-results", "methods", "ranges",
	"storage", "panels", "terminology", "analyzers", "display-order",
}

// dictionaryLimit is the controller's own cap on the typeahead.
const dictionaryLimit = 50

// ListTests ports listTests.
//
// Five filters, all ANDed, all applied in memory over every test — and the
// SEARCH matches the raw localized name, not the augmented one the response
// shows. So searching "Albumin" finds "Albumin(Urines)" and searching "Urines"
// does not.
func (s *EditorReadService) ListTests(domain, status string, amr *bool, sampleType, search string,
	page, pageSize int) (*form.TestListPage, error) {

	rows, err := s.DAO.AllTestsForList()
	if err != nil {
		return nil, err
	}

	// Resolved once rather than per test — the only part of this Java method
	// that is not a linear scan.
	var sampleTypeIDs map[string]bool
	if strings.TrimSpace(sampleType) != "" {
		ids, err := s.DAO.TestsForSampleType(sampleType)
		if err != nil {
			return nil, err
		}
		sampleTypeIDs = map[string]bool{}
		for _, id := range ids {
			sampleTypeIDs[id] = true
		}
	}

	searchLower := strings.ToLower(search)
	filtered := []daoimpl.CatalogListRow{}
	for _, r := range rows {
		if strings.TrimSpace(domain) != "" && derefStr(r.Domain) != domain {
			continue
		}
		if status == "active" && !r.Active {
			continue
		}
		if status == "inactive" && r.Active {
			continue
		}
		if amr != nil && *amr != r.AMR {
			continue
		}
		if sampleTypeIDs != nil && !sampleTypeIDs[r.TestID] {
			continue
		}
		if strings.TrimSpace(searchLower) != "" &&
			!strings.Contains(strings.ToLower(r.RawName), searchLower) {
			continue
		}
		filtered = append(filtered, r)
	}

	// By the RAW name, case-insensitively. Collections.sort is stable, so equal
	// names keep the order the unordered scan returned them in.
	sort.SliceStable(filtered, func(i, j int) bool {
		return strings.ToLower(filtered[i].RawName) < strings.ToLower(filtered[j].RawName)
	})

	out := &form.TestListPage{
		Total: len(filtered),
		// Both are clamped UP to 1, so page=0&pageSize=0 answers page 1 of size
		// 1 rather than an error.
		PageSize: maxInt(1, pageSize),
		Page:     maxInt(1, page),
		Rows:     []form.TestListRow{},
	}
	from := minInt((out.Page-1)*out.PageSize, len(filtered))
	to := minInt(from+out.PageSize, len(filtered))
	for _, r := range filtered[from:to] {
		// The augmented name and the sample type are attached to the PAGE SLICE
		// only — ≤ pageSize lookups in Java, and the reason the sort above runs
		// on the raw name.
		out.Rows = append(out.Rows, form.TestListRow{
			TestID: r.TestID, Name: r.Name, SampleType: r.SampleType,
			Code: r.Code, Domain: r.Domain, Active: r.Active, AMR: r.AMR,
		})
	}
	return out, nil
}

// Envelope ports getEditorEnvelope.
func (s *EditorReadService) Envelope(testID string) (*form.EditorEnvelope, error) {
	row, err := s.DAO.OneTestForList(testID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrNotFound
	}
	return &form.EditorEnvelope{
		TestID: row.TestID, Name: row.Name, Code: row.Code, Domain: row.Domain,
		ApplicableSections: v1Sections,
	}, nil
}

// LocalizationRefs ports getLocalizationRefs.
//
// The two fields are emitted in a FIXED order — name then reportingName — and
// one whose localization id is null is dropped rather than sent as null.
func (s *EditorReadService) LocalizationRefs(testID string) (*form.LocalizationRefs, error) {
	row, err := s.DAO.LocalizationRefs(testID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrNotFound
	}
	refs := &form.LocalizationRefs{TestID: testID, Fields: []form.LocalizationFieldRef{}}
	for _, f := range []struct {
		field string
		id    *string
	}{{"name", row.NameID}, {"reportingName", row.ReportingID}} {
		if f.id != nil && *f.id != "" {
			refs.Fields = append(refs.Fields, form.LocalizationFieldRef{
				Field: f.field, LocalizationID: *f.id,
			})
		}
	}
	return refs, nil
}

// LoincIntegrity ports getLoincIntegrity.
//
// Two ways a result silently mis-routes, both surfaced as warnings: a test that
// SHOULD receive results — active AND orderable — with no LOINC at all, and two
// active tests sharing one, where the resolver takes whichever comes first.
func (s *EditorReadService) LoincIntegrity(testID string) (*form.LoincIntegrity, error) {
	row, err := s.DAO.OneTestForList(testID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrNotFound
	}
	orderable, err := s.DAO.IsOrderable(testID)
	if err != nil {
		return nil, err
	}

	loinc := ""
	if row.Loinc != nil {
		loinc = *row.Loinc
	}
	out := &form.LoincIntegrity{
		Loinc:      row.Loinc,
		Active:     row.Active,
		NoLoinc:    row.Active && orderable && strings.TrimSpace(loinc) == "",
		Duplicates: []form.TestRef{},
	}
	if strings.TrimSpace(loinc) == "" {
		return out, nil
	}
	others, err := s.DAO.ActiveTestsByLoinc(loinc)
	if err != nil {
		return nil, err
	}
	for _, other := range others {
		if other.TestID == testID {
			continue
		}
		out.Duplicates = append(out.Duplicates, form.TestRef{
			TestID: other.TestID, Name: other.Name,
		})
	}
	return out, nil
}

// SearchDictionary ports searchDictionaryOptions.
//
// A blank search returns NOTHING — deliberately, so the control does not dump
// the whole dictionary on focus.
func (s *EditorReadService) SearchDictionary(search string) ([]form.DictionaryOption, error) {
	out := []form.DictionaryOption{}
	if strings.TrimSpace(search) == "" {
		return out, nil
	}
	rows, err := s.DAO.SearchDictionary(strings.TrimSpace(search), dictionaryLimit)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		out = append(out, form.DictionaryOption{ID: r.ID, Name: r.Name})
	}
	return out, nil
}

// Siblings ports siblings.
//
// The stem is the augmented name with its trailing "(SampleType)" cut off, so
// the grouping is by analyte across specimens. An unknown test is an EMPTY LIST,
// not a 404 — the only endpoint in this group that answers that way.
//
// The rows reuse TestListRow and fill only testId, name and sampleType, so
// `active` comes back FALSE for every sibling however active it is. Measured.
func (s *EditorReadService) Siblings(testID string) ([]form.TestListRow, error) {
	out := []form.TestListRow{}
	row, err := s.DAO.OneTestForList(testID)
	if err != nil || row == nil {
		return out, err
	}
	stem := nameStem(row.Name)
	if stem == "" {
		return out, nil
	}
	actives, err := s.DAO.ActiveTestsForList()
	if err != nil {
		return nil, err
	}
	for _, other := range actives {
		if !strings.EqualFold(stem, nameStem(other.Name)) {
			continue
		}
		out = append(out, form.TestListRow{
			TestID: other.TestID, Name: other.Name, SampleType: other.SampleType,
		})
	}
	return out, nil
}

// nameStem is the augmented name without its "(SampleType)" tail.
//
// It cuts at the LAST '(' and only when that is not the first character, so a
// name that opens with a bracket keeps all of it.
func nameStem(name string) string {
	if paren := strings.LastIndex(name, "("); paren > 0 {
		return strings.TrimSpace(name[:paren])
	}
	return strings.TrimSpace(name)
}

// GroupSummary ports groupSummary.
//
// A "group" is whatever ids the admin selected — there is no stored family
// entity. Blank entries and ids naming nothing are skipped in silence, so the
// response can be shorter than the request with no indication of which id was
// dropped.
func (s *EditorReadService) GroupSummary(ids string) ([]form.GroupTestSummary, error) {
	out := []form.GroupTestSummary{}
	for _, raw := range strings.Split(ids, ",") {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		row, err := s.DAO.OneTestForList(id)
		if err != nil {
			return nil, err
		}
		if row == nil {
			continue
		}
		out = append(out, form.GroupTestSummary{
			TestID: row.TestID, Name: row.Name, Code: row.Code,
			SampleType: row.SampleType, Loinc: row.Loinc, Active: row.Active,
		})
	}
	return out, nil
}

// Analyzers ports getAnalyzers.
func (s *EditorReadService) Analyzers(testID string) (*form.AnalyzersResponse, error) {
	row, err := s.DAO.OneTestForList(testID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrNotFound
	}
	rows, err := s.DAO.AnalyzersForTest(testID)
	if err != nil {
		return nil, err
	}
	resp := &form.AnalyzersResponse{TestID: testID, Analyzers: []form.AnalyzerRow{}}
	for _, r := range rows {
		resp.Analyzers = append(resp.Analyzers, form.AnalyzerRow{
			AnalyzerID: r.AnalyzerID, AnalyzerName: r.AnalyzerName,
			AnalyzerTestName: r.AnalyzerTestName,
		})
	}
	// By analyzer name, then by the analyzer's own name for the test — a stable
	// order so the read-only table does not reshuffle between loads.
	sort.SliceStable(resp.Analyzers, func(i, j int) bool {
		a, b := resp.Analyzers[i], resp.Analyzers[j]
		an, bn := strings.ToLower(derefStr(a.AnalyzerName)), strings.ToLower(derefStr(b.AnalyzerName))
		if an != bn {
			return an < bn
		}
		return strings.ToLower(derefStr(a.AnalyzerTestName)) <
			strings.ToLower(derefStr(b.AnalyzerTestName))
	})
	return resp, nil
}

// ReflexCalc ports ReflexCalcViewServiceImpl.getForTest.
func (s *EditorReadService) ReflexCalc(testID string) (*form.ReflexCalcView, error) {
	row, err := s.DAO.OneTestForList(testID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrNotFound
	}

	view := &form.ReflexCalcView{
		ReflexRules:  []form.ReflexRow{},
		CalculatedBy: []form.CalcRow{},
		FeedsInto:    []form.CalcRow{},
	}

	rules, err := s.DAO.ReflexRulesForTest(testID)
	if err != nil {
		return nil, err
	}
	for _, r := range rules {
		reflexTests := r.AddedTestName
		trigger := describeTrigger(r)
		// The rule name falls back to the added test's name, so a rule with no
		// internal note is labelled by what it adds.
		ruleName := r.InternalNote
		if ruleName == nil || strings.TrimSpace(*ruleName) == "" {
			ruleName = reflexTests
		}
		view.ReflexRules = append(view.ReflexRules, form.ReflexRow{
			ID: r.ID, RuleName: ruleName, TriggerCondition: trigger, ReflexTests: reflexTests,
		})
	}

	calcs, err := s.DAO.ActiveCalculations()
	if err != nil {
		return nil, err
	}
	operations, err := s.DAO.Operations()
	if err != nil {
		return nil, err
	}
	byCalc := map[string][]daoimpl.OperationRow{}
	for _, op := range operations {
		byCalc[op.CalculationID] = append(byCalc[op.CalculationID], op)
	}

	for _, c := range calcs {
		row := form.CalcRow{ID: c.ID, Name: c.Name, OutputTest: c.OutputTestName}
		formula := buildFormula(c, byCalc[c.ID])
		row.Formula = formula
		switch {
		case c.TestID != nil && *c.TestID == testID:
			// This test is the calculation's OUTPUT.
			view.CalculatedBy = append(view.CalculatedBy, row)
		case referencesTest(byCalc[c.ID], testID):
			// This test is one of its INPUTS.
			view.FeedsInto = append(view.FeedsInto, row)
		}
	}
	return view, nil
}

// describeTrigger ports describeTrigger: the dictionary value the rule fires
// on, or the free-text one, prefixed by the relation when there is one.
func describeTrigger(r daoimpl.ReflexRuleRow) string {
	value := ""
	if r.TestResultValue != nil && *r.TestResultValue != "" {
		value = *r.TestResultValue
	} else if r.NonDictionaryValue != nil {
		value = *r.NonDictionaryValue
	}
	relation := ""
	if r.Relation != nil && *r.Relation != "" {
		relation = *r.Relation + " "
	}
	if strings.TrimSpace(value) == "" {
		if strings.TrimSpace(relation) == "" {
			return "Any result"
		}
		return strings.TrimSpace(relation)
	}
	return strings.TrimSpace(relation + value)
}

// buildFormula joins a calculation's terms, falling back to its stored `result`
// when it has none.
func buildFormula(c daoimpl.CalculationRow, ops []daoimpl.OperationRow) *string {
	if len(ops) == 0 {
		return c.Result
	}
	parts := make([]string, 0, len(ops))
	for _, op := range ops {
		token := ""
		if op.Value != nil && strings.TrimSpace(*op.Value) != "" {
			token = *op.Value
		} else if op.Type != nil {
			token = *op.Type
		}
		parts = append(parts, token)
	}
	joined := strings.Join(parts, " ")
	return &joined
}

// referencesTest is `operations.anyMatch(op -> testId.equals(op.getValue()))` —
// a calculation feeds INTO this test when one of its terms is the test's id.
func referencesTest(ops []daoimpl.OperationRow, testID string) bool {
	for _, op := range ops {
		if op.Value != nil && *op.Value == testID {
			return true
		}
	}
	return false
}

// StorageHistory ports getHistory.
//
// A missing TEST is a 404; a test with no storage config is an empty list.
func (s *EditorReadService) StorageHistory(testID string) ([]form.StorageHistoryEntry, error) {
	row, err := s.DAO.OneTestForList(testID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrNotFound
	}
	rows, err := s.DAO.StorageHistory(testID)
	if err != nil {
		return nil, err
	}
	out := []form.StorageHistoryEntry{}
	for _, r := range rows {
		out = append(out, form.StorageHistoryEntry{
			Lastupdated: r.Lastupdated, ID: r.ID,
			TestSampleHandlingID: r.TestSampleHandlingID,
			ChangedBy:            r.ChangedBy, ChangedAt: r.ChangedAt,
			ChangeType: r.ChangeType, PreviousValues: r.PreviousValues,
			NewValues: r.NewValues,
		})
	}
	return out, nil
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
