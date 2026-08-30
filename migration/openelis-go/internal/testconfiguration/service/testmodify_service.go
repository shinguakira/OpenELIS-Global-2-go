package service

import (
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"strings"

	commondaoimpl "openelis-go/internal/common/daoimpl"
	"openelis-go/internal/common/util"
	locform "openelis-go/internal/localization/form"
	"openelis-go/internal/testconfiguration/daoimpl"
	"openelis-go/internal/testconfiguration/form"
)

// TestModifyService ports TestModifyEntryRestController.
type TestModifyService struct {
	Lists    *commondaoimpl.DisplayListDAOImpl
	Read     *daoimpl.TestModifyReadDAO
	DAO      *daoimpl.TestModifyDAO
	TestAdd  *TestAddService
	Messages map[string]string
}

// Form ports showTestModifyEntry.
//
// The eight lists are TestAdd's, with ONE difference: labUnitList is the active
// test sections alone. TestAdd concatenates the inactive ones onto it; this
// screen does not.
//
// testCatBeanList is built only when a filter is given. A blank GET answers
// with an empty list rather than the whole catalogue — an explicit guard on the
// initial page load, not a fallthrough.
func (s *TestModifyService) Form(sampleTypeID, testSectionID string) (*form.TestModifyEntryForm, error) {
	f := form.NewTestModifyEntryForm()

	activeTypes, err := s.Lists.ActiveHumanSampleTypes()
	if err != nil {
		return nil, err
	}
	inactiveTypes, err := s.Lists.InactiveHumanSampleTypes()
	if err != nil {
		return nil, err
	}
	sampleTypes := append(append([]util.IdValuePair{}, activeTypes...), inactiveTypes...)
	f.SampleTypeList = &sampleTypes

	panels, err := s.Lists.Panels()
	if err != nil {
		return nil, err
	}
	f.PanelList = &panels

	resultTypes, err := s.TestAdd.localizedResultTypes()
	if err != nil {
		return nil, err
	}
	f.ResultTypeList = &resultTypes

	uoms, err := s.Lists.UnitsOfMeasure()
	if err != nil {
		return nil, err
	}
	f.UomList = &uoms

	labUnits, err := s.Lists.ActiveTestSections()
	if err != nil {
		return nil, err
	}
	f.LabUnitList = &labUnits

	ageRanges, err := s.TestAdd.ageRanges()
	if err != nil {
		return nil, err
	}
	f.AgeRangeList = &ageRanges

	dictionary, err := s.Lists.DictionaryTestResults()
	if err != nil {
		return nil, err
	}
	f.DictionaryList = &dictionary

	grouped, err := s.TestAdd.groupedDictionary()
	if err != nil {
		return nil, err
	}
	f.GroupedDictionaryList = &grouped

	beans := []form.TestCatalogBean{}
	if strings.TrimSpace(sampleTypeID) != "" || strings.TrimSpace(testSectionID) != "" {
		beans, err = s.catalog(sampleTypeID, testSectionID)
		if err != nil {
			return nil, err
		}
	}
	f.TestCatBeanList = &beans
	return &f, nil
}

// catalog ports createTestCatBeanList.
func (s *TestModifyService) catalog(sampleTypeID, testSectionID string) ([]form.TestCatalogBean, error) {
	tests, err := s.Read.CatalogTests(sampleTypeID, testSectionID)
	if err != nil {
		return nil, err
	}
	if len(tests) == 0 {
		return []form.TestCatalogBean{}, nil
	}

	panels, err := s.Read.PanelNamesByTest()
	if err != nil {
		return nil, err
	}
	panelsByTest := map[string][]string{}
	for _, p := range panels {
		panelsByTest[p.TestID] = append(panelsByTest[p.TestID], p.Value)
	}

	sampleTypes, err := s.Read.SampleTypeNameByTest()
	if err != nil {
		return nil, err
	}
	sampleTypeByTest := map[string]string{}
	for _, st := range sampleTypes {
		if _, seen := sampleTypeByTest[st.TestID]; !seen {
			sampleTypeByTest[st.TestID] = st.Value
		}
	}

	results, err := s.Read.ActiveResultsByTest()
	if err != nil {
		return nil, err
	}
	resultsByTest := map[string][]commondaoimplResult{}
	for _, r := range results {
		resultsByTest[r.TestID] = append(resultsByTest[r.TestID], commondaoimplResult(r))
	}

	limits, err := s.Read.LimitsByTest()
	if err != nil {
		return nil, err
	}
	limitsByTest := map[string][]daoimpl.CatalogLimitRow{}
	for _, l := range limits {
		limitsByTest[l.TestID] = append(limitsByTest[l.TestID], l)
	}

	locales, err := s.Lists.ActiveLocales()
	if err != nil {
		return nil, err
	}
	values, err := s.Lists.LocalizationValues()
	if err != nil {
		return nil, err
	}

	out := make([]form.TestCatalogBean, 0, len(tests))
	for _, t := range tests {
		bean := form.TestCatalogBean{
			ID:                      t.ID,
			TestUnit:                derefOr(t.TestSectionName),
			SampleType:              "n/a",
			Panel:                   "None",
			Uom:                     "n/a",
			SignificantDigits:       "n/a",
			Loinc:                   t.Loinc,
			NotifyResults:           t.NotifyResults,
			InLabOnly:               t.InLabOnly,
			AntimicrobialResistance: t.AMR,
			// The bean's own default, used when test.sort_order is NULL.
			TestSortOrder: math.MaxInt32,
		}
		if t.SortOrder != nil {
			bean.TestSortOrder = *t.SortOrder
		}
		if name, ok := sampleTypeByTest[t.ID]; ok {
			bean.SampleType = name
		}
		if names := panelsByTest[t.ID]; len(names) > 0 {
			bean.Panel = strings.Join(names, ", ")
		}
		if t.UomName != nil && *t.UomName != "" {
			bean.Uom = *t.UomName
		}
		if t.NameLocID != nil && *t.NameLocID != "" {
			bean.Localization = locform.BuildLocalization(*t.NameLocID, derefOr(t.NameLocDesc),
				t.NameLocUpdated, values[*t.NameLocID], locales, s.Lists.Locale())
		}
		if t.ReportingLocID != nil && *t.ReportingLocID != "" {
			bean.ReportLocalization = locform.BuildLocalization(*t.ReportingLocID, derefOr(t.ReportLocDesc),
				t.ReportLocUpdated, values[*t.ReportingLocID], locales, s.Lists.Locale())
		}

		// getResultType falls back to ALPHA when a test has no active result —
		// deliberately, because ALPHA accepts any input where NUMERIC would
		// reject it.
		testResults := resultsByTest[t.ID]
		bean.ResultType = "A"
		if len(testResults) > 0 {
			bean.ResultType = testResults[0].ResultType
		}

		if t.IsActive {
			bean.Active = "Active"
		} else {
			bean.Active = "Not active"
		}
		if t.Orderable {
			bean.Orderable = "Orderable"
		} else {
			bean.Orderable = "Not orderable"
		}

		if bean.ResultType == "N" && len(testResults) > 0 {
			digits := derefOr(testResults[0].SignificantDigits)
			if digits != "" {
				bean.SignificantDigits = digits
			}
			bean.HasLimitValues = true
			limitBeans := s.limitBeans(limitsByTest[t.ID], bean.SignificantDigits)
			bean.ResultLimits = &limitBeans
		}

		bean.HasDictionaryValues = isDictionaryVariantChar(bean.ResultType)
		if bean.HasDictionaryValues {
			names, ids := []string{}, []string{}
			for _, r := range testResults {
				name, id, ok := dictionaryEntry(r)
				if !ok {
					continue
				}
				names = append(names, name)
				ids = append(ids, id)
			}
			bean.DictionaryValues = &names
			bean.DictionaryIDs = &ids
			reference := s.referenceValue(limitsByTest[t.ID])
			bean.ReferenceValue = &reference
			// getDictionaryIdByDictEntry matches the reference TEXT against the
			// option texts; "n/a" short-circuits to null, and so does a
			// reference naming an option this test does not offer.
			if reference != "n/a" {
				for i, name := range names {
					if name == reference {
						id := ids[i]
						bean.ReferenceID = &id
						break
					}
				}
			}
		}
		out = append(out, bean)
	}

	// testUnit, then sampleType, then panel, then sort order — all three strings
	// with String.compareTo, which is byte order.
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.TestUnit != b.TestUnit {
			return a.TestUnit < b.TestUnit
		}
		if a.SampleType != b.SampleType {
			return a.SampleType < b.SampleType
		}
		if a.Panel != b.Panel {
			return a.Panel < b.Panel
		}
		return a.TestSortOrder < b.TestSortOrder
	})
	return out, nil
}

// commondaoimplResult is a local alias so the catalog can carry result rows
// without importing the DAO type name twice.
type commondaoimplResult = daoimpl.CatalogResultRow

// dictionaryEntry ports getDictionaryValue / getDictionaryId, which are the
// same method twice over one returning the name and one the id.
//
// A non-dictionary row, a blank value, or a value naming no dictionary row is
// dropped from BOTH lists — CollectionUtils.addIgnoreNull — so the two stay
// index-aligned, which is what getDictionaryIdByDictEntry relies on.
func dictionaryEntry(r daoimpl.CatalogResultRow) (name, id string, ok bool) {
	if !isDictionaryVariantChar(r.ResultType) || strings.TrimSpace(r.Value) == "" {
		return "", "", false
	}
	if r.DictionaryName == nil {
		return "", "", false
	}
	name, id = *r.DictionaryName, r.Value
	// A quantifiable option is suffixed in BOTH lists, so the id stops being an
	// id — "542 Qualifiable" is what the screen receives.
	if r.IsQuantifiable {
		name += " Qualifiable"
		id += " Qualifiable"
	}
	return name, id, true
}

// referenceValue ports createReferenceValueForDictionaryType: the FIRST limit
// carrying a dictionary normal, rendered as that dictionary's name.
//
// A test converted from a numeric type keeps its stale numeric limit, so
// picking limit 0 blindly would render a range where a name belongs — the
// scan for a non-blank dictionaryNormalId is the fix, and "n/a" is what a test
// with no such limit gets.
func (s *TestModifyService) referenceValue(limits []daoimpl.CatalogLimitRow) string {
	for _, l := range limits {
		if strings.TrimSpace(l.DictionaryNormalID) == "" {
			continue
		}
		if l.DictionaryName != nil {
			return *l.DictionaryName
		}
		return ""
	}
	return "n/a"
}

// limitBeans ports getResultLimits: the limits sorted by minAge and rendered.
func (s *TestModifyService) limitBeans(limits []daoimpl.CatalogLimitRow, significantDigits string) []form.ResultLimitBean {
	sorted := append([]daoimpl.CatalogLimitRow{}, limits...)
	// `(int) (o1.getMinAge() - o2.getMinAge())` — a double subtraction CAST to
	// int, so two ages less than one apart compare EQUAL and the sort keeps
	// their original order.
	sort.SliceStable(sorted, func(i, j int) bool {
		return int(sorted[i].MinAge-sorted[j].MinAge) < 0
	})

	out := make([]form.ResultLimitBean, 0, len(sorted))
	for _, l := range sorted {
		bean := form.ResultLimitBean{
			Gender:         "n/a",
			AgeRange:       s.displayAgeRange(l),
			NormalRange:    s.displayRange(l, l.LowNormal, l.HighNormal, significantDigits),
			ValidRange:     s.displayRange(l, l.LowValid, l.HighValid, significantDigits),
			ReportingRange: s.displayRange(l, l.LowReportingRange, l.HighReportingRange, significantDigits),
			CriticalRange:  s.displayRange(l, l.LowCritical, l.HighCritical, significantDigits),
		}
		if strings.TrimSpace(l.Gender) != "" {
			bean.Gender = l.Gender
		}
		out = append(out, bean)
	}
	return out
}

// displayRange ports getDisplayNormalRange and the four getDisplay*Range
// wrappers, which all reduce to it for the NUMERIC result type and to "" for
// every other one.
func (s *TestModifyService) displayRange(l daoimpl.CatalogLimitRow, low, high float64, significantDigits string) string {
	if l.ResultTypeID != numericResultTypeID {
		return ""
	}
	if math.IsInf(low, -1) && math.IsInf(high, 1) {
		return s.message("result.anyValue")
	}
	if low == high {
		return ""
	}
	if math.IsInf(high, 1) {
		return "> " + doubleWithSignificantDigits(low, significantDigits)
	}
	if math.IsInf(low, -1) {
		return "< " + doubleWithSignificantDigits(high, significantDigits)
	}
	return doubleWithSignificantDigits(low, significantDigits) + "-" +
		doubleWithSignificantDigits(high, significantDigits)
}

// displayAgeRange ports getDisplayAgeRange.
//
// The months arithmetic divides by `365 / 12` — INTEGER division in Java, so
// the divisor is 30, not 30.4167. A port using the real ratio drifts by a day
// per month.
func (s *TestModifyService) displayAgeRange(l daoimpl.CatalogLimitRow) string {
	if l.MinAge == 0 && math.IsInf(l.MaxAge, 1) {
		return s.message("age.anyAge")
	}
	day := s.message("abbreviation.day.single")
	month := s.message("abbreviation.month.single")
	year := s.message("abbreviation.year.single")

	minDays, minMonths, minYears := splitAge(l.MinAge)
	if math.IsInf(l.MaxAge, 1) {
		return ">" + formatAge(minYears) + year + formatAge(minMonths) + month + formatAge(minDays) + day
	}
	maxDays, maxMonths, maxYears := splitAge(l.MaxAge)
	return formatAge(minDays) + day + "/" + formatAge(minMonths) + month + "/" + formatAge(minYears) + year +
		"-" +
		formatAge(maxDays) + day + "/" + formatAge(maxMonths) + month + "/" + formatAge(maxYears) + year
}

func splitAge(age float64) (days, months, years float64) {
	remaining := age
	years = math.Floor(remaining / 365)
	remaining -= years * 365
	// 365 / 12 is integer division in Java: 30.
	months = math.Floor(remaining / 30)
	remaining -= math.Floor(months * 30)
	return remaining, months, years
}

// formatAge is DecimalFormat with maximumFractionDigits 0 — HALF_EVEN rounding,
// which is what math.RoundToEven does.
func formatAge(v float64) string {
	return strconv.FormatFloat(math.RoundToEven(v), 'f', -1, 64)
}

// doubleWithSignificantDigits ports StringUtil.doubleWithSignificantDigits: a
// blank or "-1" digit count falls back to String.valueOf(double), which keeps
// the trailing ".0".
func doubleWithSignificantDigits(v float64, significantDigits string) string {
	if strings.TrimSpace(significantDigits) == "" || significantDigits == "-1" {
		return strings.Trim(util.JavaDoubleString(v), `"`)
	}
	digits, err := strconv.Atoi(significantDigits)
	if err != nil {
		return strings.Trim(util.JavaDoubleString(v), `"`)
	}
	return strconv.FormatFloat(v, 'f', digits, 64)
}

func isDictionaryVariantChar(t string) bool {
	return t != "" && strings.Contains("DMC", t)
}

func (s *TestModifyService) message(key string) string {
	if v, ok := s.Messages[key]; ok && v != "" {
		return v
	}
	return key
}

// Update ports postTestModifyEntry.
//
// The BindingResult is populated and then IGNORED for the write, exactly as in
// TestAdd — setupDisplayItems runs on errors but the create still goes ahead.
func (s *TestModifyService) Update(post form.TestModifyEntryPost, sysUserID int64) (*form.TestModifyEntryForm, error) {
	f := form.NewTestModifyEntryForm()
	if post.JSONWad != nil {
		f.JSONWad = *post.JSONWad
	}
	f.Loinc = post.Loinc

	var params testAddParams
	if err := json.Unmarshal([]byte(f.JSONWad), &params); err != nil {
		return nil, errBadJSONWad
	}
	var wad struct {
		TestID string `json:"testId"`
	}
	_ = json.Unmarshal([]byte(f.JSONWad), &wad)
	if strings.TrimSpace(wad.TestID) == "" {
		// updateTestSets reads the id straight into its queries; a null one is
		// an ObjectNotFoundException out of testService.get.
		return nil, errBadJSONWad
	}

	base, err := buildTestAddRow(params)
	if err != nil {
		return nil, err
	}
	locales, err := s.Lists.ActiveLocales()
	if err != nil {
		return nil, err
	}
	row := daoimpl.TestModifyRow{TestID: wad.TestID, TestAddRow: base, Locales: locales}

	// The dictionary branch deactivates the existing results from
	// createTestSets — before the transaction, and only for D/M/C.
	if isDictionaryVariantByID(params.ResultType) {
		if err := s.DAO.DeactivateDictionaryResults(row.TestID); err != nil {
			return nil, err
		}
	}

	sets := make([]daoimpl.TestAddSet, 0, len(params.SampleTypes))
	for _, st := range params.SampleTypes {
		set := daoimpl.TestAddSet{SampleTypeID: st.TypeID}
		for _, t := range st.Tests {
			set.OrderedTests = append(set.OrderedTests, rawToString(t.ID))
		}
		sets = append(sets, set)
	}

	if err := s.DAO.Update(row, sets, sysUserID); err != nil {
		return nil, err
	}
	return &f, nil
}
