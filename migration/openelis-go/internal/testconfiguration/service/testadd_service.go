package service

import (
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strconv"
	"strings"

	commondaoimpl "openelis-go/internal/common/daoimpl"
	"openelis-go/internal/common/util"
	"openelis-go/internal/testconfiguration/daoimpl"
	"openelis-go/internal/testconfiguration/form"
)

// TestAddService ports TestAddRestController and TestAddControllerUtills.
type TestAddService struct {
	Lists *commondaoimpl.DisplayListDAOImpl
	DAO   *daoimpl.TestAddDAO
	// Messages is message_en.properties — the result type labels and the age
	// range names are MessageUtil lookups, not columns.
	Messages map[string]string
}

// Form ports showTestAdd.
//
// Two of the eight lists are CONCATENATIONS of an active and an inactive list,
// in that order and with no re-sort, so the inactive rows land at the end.
func (s *TestAddService) Form() (*form.TestAddForm, error) {
	f := form.NewTestAddForm()

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

	resultTypes, err := s.localizedResultTypes()
	if err != nil {
		return nil, err
	}
	f.ResultTypeList = &resultTypes

	uoms, err := s.Lists.UnitsOfMeasure()
	if err != nil {
		return nil, err
	}
	f.UomList = &uoms

	activeSections, err := s.Lists.ActiveTestSections()
	if err != nil {
		return nil, err
	}
	inactiveSections, err := s.Lists.InactiveTestSections()
	if err != nil {
		return nil, err
	}
	labUnits := append(append([]util.IdValuePair{}, activeSections...), inactiveSections...)
	f.LabUnitList = &labUnits

	ageRanges, err := s.ageRanges()
	if err != nil {
		return nil, err
	}
	f.AgeRangeList = &ageRanges

	dictionary, err := s.Lists.DictionaryTestResults()
	if err != nil {
		return nil, err
	}
	f.DictionaryList = &dictionary

	grouped, err := s.groupedDictionary()
	if err != nil {
		return nil, err
	}
	f.GroupedDictionaryList = &grouped

	// form.setLoinc(new Test().getLoinc()) — null, and the form is NON_NULL, so
	// the key does not appear.
	return &f, nil
}

// localizedResultTypeLabels is createLocalizedResultTypeList's branch table,
// keyed by the type's DESCRIPTION. There is no branch for "Titer", so type 3 is
// silently dropped and the screen offers six of the seven result types.
var localizedResultTypeLabels = map[string]string{
	"Remark":                "result.type.freeText",
	"Dictionary":            "result.type.select",
	"Numeric":               "result.type.numeric",
	"Alpha,no range check":  "result.type.alpha",
	"Multiselect":           "result.type.multiselect",
	"Cascading Multiselect": "result.type.cascading",
}

func (s *TestAddService) localizedResultTypes() ([]util.IdValuePair, error) {
	rows, err := s.Lists.ResultTypes()
	if err != nil {
		return nil, err
	}
	out := []util.IdValuePair{}
	for _, r := range rows {
		key, ok := localizedResultTypeLabels[r.Description]
		if !ok {
			continue
		}
		out = append(out, util.NewIdValuePair(r.ID, s.message(key)))
	}
	return out, nil
}

// ageRangeNameKeys is getPredefinedAgeRanges' own branch table. It matches on
// the site information NAME, one branch each, and a resultAgeRange row whose
// name is none of the five reaches the list with a NULL displayed value.
var ageRangeNames = map[string]bool{
	"new born": true, "infant": true, "young child": true, "child": true, "adult": true,
}

func (s *TestAddService) ageRanges() ([]util.IdValuePair, error) {
	rows, err := s.Lists.PredefinedAgeRanges()
	if err != nil {
		return nil, err
	}
	out := make([]util.IdValuePair, 0, len(rows))
	for _, r := range rows {
		value := ""
		if ageRangeNames[r.Name] {
			// getLocalizedName(): the name_key through MessageUtil, falling back
			// to the entity's own name when the bundle has no such key.
			value = r.Name
			if r.NameKey != "" {
				if v, ok := s.Messages[r.NameKey]; ok && v != "" {
					value = v
				}
			}
		}
		out = append(out, util.NewIdValuePair(r.Value, value))
	}
	commondaoimpl.SortAgeRanges(out)
	return out, nil
}

// groupedDictionary ports createGroupedDictionaryList.
//
// The dictionary ids of each test's dictionary-variant results are joined with
// commas into ONE string, the strings go into a HashSet — which is what
// deduplicates two tests offering the same options — and each surviving string
// is split back apart and resolved to pairs. The groups are then sorted by
// SIZE, with Collections.sort, which is stable.
//
// Java's iteration order over that HashSet is its bucket order, and nothing in
// the port can or should reproduce it: it is an accident of String.hashCode and
// the table size. Equal-sized groups therefore come out in a different order
// here. The set of groups, their sizes and the order WITHIN each group are all
// identical, and those are what the screen reads.
func (s *TestAddService) groupedDictionary() ([][]util.IdValuePair, error) {
	rows, err := s.Lists.DictionaryResultGroups()
	if err != nil {
		return nil, err
	}

	type group struct {
		ids   []string
		pairs []util.IdValuePair
	}
	ordered := []group{}
	currentTest := ""
	for i, r := range rows {
		if i == 0 || r.TestID != currentTest {
			currentTest = r.TestID
			ordered = append(ordered, group{})
		}
		g := &ordered[len(ordered)-1]
		g.ids = append(g.ids, r.Value)
		if r.Found {
			// getDictionaryById returning null drops the ENTRY, not the group.
			g.pairs = append(g.pairs, util.NewIdValuePair(r.Value, r.Name))
		}
	}

	seen := map[string]bool{}
	out := [][]util.IdValuePair{}
	for _, g := range ordered {
		key := strings.Join(g.ids, ",")
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, g.pairs)
	}
	sort.SliceStable(out, func(i, j int) bool { return len(out[i]) < len(out[j]) })
	return out, nil
}

func (s *TestAddService) message(key string) string {
	if v, ok := s.Messages[key]; ok && v != "" {
		return v
	}
	return key
}

// testAddParams is TestAddControllerUtills.TestAddParams — the parsed jsonWad.
//
// Every field is read with a cast to String, so a JSON number where a string is
// expected is a ClassCastException and a 500. The numeric block reads its
// values with toString() instead, which accepts either.
type testAddParams struct {
	TestNameEnglish       string `json:"testNameEnglish"`
	TestNameFrench        string `json:"testNameFrench"`
	TestReportNameEnglish string `json:"testReportNameEnglish"`
	TestReportNameFrench  string `json:"testReportNameFrench"`
	TestSection           string `json:"testSection"`
	DictionaryReference   string `json:"dictionaryReference"`
	Panels                []struct {
		ID string `json:"id"`
	} `json:"panels"`
	Uom         string `json:"uom"`
	Loinc       string `json:"loinc"`
	ResultType  string `json:"resultType"`
	SampleTypes []struct {
		TypeID string `json:"typeId"`
		Tests  []struct {
			ID json.RawMessage `json:"id"`
		} `json:"tests"`
	} `json:"sampleTypes"`
	Active                  string `json:"active"`
	Orderable               string `json:"orderable"`
	NotifyResults           string `json:"notifyResults"`
	InLabOnly               string `json:"inLabOnly"`
	AntimicrobialResistance string `json:"antimicrobialResistance"`

	LowValid           json.RawMessage `json:"lowValid"`
	HighValid          json.RawMessage `json:"highValid"`
	LowReportingRange  json.RawMessage `json:"lowReportingRange"`
	HighReportingRange json.RawMessage `json:"highReportingRange"`
	LowCritical        json.RawMessage `json:"lowCritical"`
	HighCritical       json.RawMessage `json:"highCritical"`
	SignificantDigits  json.RawMessage `json:"significantDigits"`
	ResultLimits       []struct {
		Gender           *bool           `json:"gender"`
		HighAgeRange     json.RawMessage `json:"highAgeRange"`
		LowNormal        json.RawMessage `json:"lowNormal"`
		HighNormal       json.RawMessage `json:"highNormal"`
		LowNormalFemale  json.RawMessage `json:"lowNormalFemale"`
		HighNormalFemale json.RawMessage `json:"highNormalFemale"`
	} `json:"resultLimits"`

	Dictionary []struct {
		ID        string `json:"id"`
		Qualified string `json:"qualified"`
	} `json:"dictionary"`
	DefaultTestResult string `json:"defaultTestResult"`
}

// numericResultTypeID and the dictionary-variant ids are
// TypeOfTestResultServiceImpl.ResultType's own — the enum carries the database
// ids as constants, so the branch does not read the table.
const (
	numericResultTypeID              = "4"
	dictionaryResultTypeID           = "2"
	multiselectResultTypeID          = "6"
	cascadingMultiselectResultTypeID = "7"
)

func isNumericByID(id string) bool { return id == numericResultTypeID }

func isDictionaryVariantByID(id string) bool {
	return id == dictionaryResultTypeID || id == multiselectResultTypeID ||
		id == cascadingMultiselectResultTypeID
}

// resultTypeChars maps the submitted type id to the character test_result
// stores. Java resolves it through typeOfTestResultService.getResultTypeById.
var resultTypeChars = map[string]string{
	"1": "R", "2": "D", "3": "T", "4": "N", "5": "A", "6": "M", "7": "C",
}

// Add ports postTestAdd.
//
// The BindingResult the validator fills is NEVER CHECKED: the controller runs
// the parse, the create and the cache refreshes whatever it says. Invalid input
// is written, and the only reject that can stop anything is a parse failure —
// which leaves obj null and NPEs out of extractTestAddParms into a 500.
func (s *TestAddService) Add(post form.TestAddPost, sysUserID int64) (*form.TestAddForm, error) {
	f := form.NewTestAddForm()
	if post.JSONWad != nil {
		f.JSONWad = *post.JSONWad
	}
	f.Loinc = post.Loinc

	var params testAddParams
	if err := json.Unmarshal([]byte(f.JSONWad), &params); err != nil {
		// parser.parse threw: obj stays null and the first obj.get() NPEs.
		return nil, errBadJSONWad
	}

	row, err := buildTestAddRow(params)
	if err != nil {
		return nil, err
	}

	sets := make([]daoimpl.TestAddSet, 0, len(params.SampleTypes))
	for _, st := range params.SampleTypes {
		set := daoimpl.TestAddSet{SampleTypeID: st.TypeID}
		for _, t := range st.Tests {
			// String.valueOf(get("id")) — a number here is stringified rather
			// than rejected, unlike every other id in this payload.
			set.OrderedTests = append(set.OrderedTests, rawToString(t.ID))
		}
		sets = append(sets, set)
	}

	if _, err := s.DAO.Add(row, sets, sysUserID); err != nil {
		return nil, err
	}
	return &f, nil
}

// extractLimits ports extractLimits: one row per entry, and a SECOND row for a
// gendered entry carrying the female normals.
//
// lowAge chains — each entry starts where the previous one ended, beginning at
// "0" — so the ages come from the list's ORDER, not from any submitted low
// value. `(Boolean) get("gender")` NPEs when the key is missing, and
// setMinAge/setLowNormal and their siblings take primitive doubles, so a blank
// or unparseable bound is an NPE too. Both are 500s, raised before any write.
func extractLimits(params testAddParams) ([]daoimpl.TestAddLimit, error) {
	out := []daoimpl.TestAddLimit{}
	lowAge := "0"
	for _, l := range params.ResultLimits {
		if l.Gender == nil {
			return nil, errBadJSONWad
		}
		highAge := rawToString(l.HighAgeRange)

		gender := ""
		if *l.Gender {
			gender = "M"
		}
		limit, err := buildLimit(gender, lowAge, highAge,
			rawToString(l.LowNormal), rawToString(l.HighNormal))
		if err != nil {
			return nil, err
		}
		out = append(out, limit)

		if *l.Gender {
			female, err := buildLimit("F", lowAge, highAge,
				rawToString(l.LowNormalFemale), rawToString(l.HighNormalFemale))
			if err != nil {
				return nil, err
			}
			out = append(out, female)
		}
		lowAge = highAge
	}
	return out, nil
}

func buildLimit(gender, lowAge, highAge, lowNormal, highNormal string) (daoimpl.TestAddLimit, error) {
	minAge, maxAge := doubleWithInfinity(lowAge), doubleWithInfinity(highAge)
	low, high := doubleWithInfinity(lowNormal), doubleWithInfinity(highNormal)
	if minAge == nil || maxAge == nil || low == nil || high == nil {
		return daoimpl.TestAddLimit{}, errBadJSONWad
	}
	return daoimpl.TestAddLimit{
		Gender: gender, MinAge: *minAge, MaxAge: *maxAge,
		LowNormal: *low, HighNormal: *high,
	}, nil
}

// doubleWithInfinity ports StringUtil.doubleWithInfinity: blank and
// unparseable both come back null, and the two infinity literals are spelled
// out rather than parsed.
func doubleWithInfinity(s string) *float64 {
	switch strings.TrimSpace(s) {
	case "":
		return nil
	case "Infinity":
		v := inf(1)
		return &v
	case "-Infinity":
		v := inf(-1)
		return &v
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &v
}

// rawToString is `obj.get(key).toString()`: a JSON string loses its quotes, a
// number keeps its literal text, and a missing key NPEs — which the caller
// turns into the 500 Java raises.
func rawToString(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return string(raw)
}

// ErrTestAddPayload is the NullPointerException / ClassCastException Java
// raises out of extractTestAddParms and createTestSets: a jsonWad that will not
// parse, a resultLimits entry with no `gender` key, or a numeric bound that
// StringUtil.doubleWithInfinity turns into null and a primitive setter then
// unboxes. All three are raised BEFORE addTests, so nothing is written.
var errBadJSONWad = errors.New("testconfiguration: jsonWad cannot be built into a test")

func inf(sign int) float64 { return math.Inf(sign) }

// buildTestAddRow turns the parsed jsonWad into the row both TestAdd and
// TestModifyEntry write. createTestSets is duplicated between the two
// controllers character for character in this part, so the port is not.
func buildTestAddRow(params testAddParams) (daoimpl.TestAddRow, error) {
	row := daoimpl.TestAddRow{
		NameEnglish:             params.TestNameEnglish,
		NameFrench:              params.TestNameFrench,
		ReportNameEnglish:       params.TestReportNameEnglish,
		ReportNameFrench:        params.TestReportNameFrench,
		TestSectionID:           params.TestSection,
		UomID:                   params.Uom,
		Loinc:                   params.Loinc,
		Active:                  params.Active,
		Orderable:               params.Orderable == "Y",
		NotifyResults:           params.NotifyResults == "Y",
		InLabOnly:               params.InLabOnly == "Y",
		AntimicrobialResistance: params.AntimicrobialResistance == "Y",
		ResultTypeID:            params.ResultType,
		ResultTypeChar:          resultTypeChars[params.ResultType],
		DictionaryReferenceID:   params.DictionaryReference,
	}
	for _, p := range params.Panels {
		row.PanelIDs = append(row.PanelIDs, p.ID)
	}

	switch {
	case isNumericByID(params.ResultType):
		row.SignificantDigits = rawToString(params.SignificantDigits)
		row.LowValid = doubleWithInfinity(rawToString(params.LowValid))
		row.HighValid = doubleWithInfinity(rawToString(params.HighValid))
		row.LowReporting = doubleWithInfinity(rawToString(params.LowReportingRange))
		row.HighReporting = doubleWithInfinity(rawToString(params.HighReportingRange))
		row.LowCritical = doubleWithInfinity(rawToString(params.LowCritical))
		row.HighCritical = doubleWithInfinity(rawToString(params.HighCritical))
		limits, err := extractLimits(params)
		if err != nil {
			return row, err
		}
		row.Limits = limits
	case isDictionaryVariantByID(params.ResultType):
		for _, d := range params.Dictionary {
			row.Dictionaries = append(row.Dictionaries, daoimpl.TestAddDictionary{
				DictionaryID:   d.ID,
				IsQuantifiable: d.Qualified == "Y",
				// isDefault compares the option id to defaultTestResult, so a
				// blank defaultTestResult matches a blank option id.
				IsDefault: d.ID == params.DefaultTestResult,
			})
		}
	}
	return row, nil
}
