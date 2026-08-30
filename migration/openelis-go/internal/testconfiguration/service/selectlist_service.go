package service

import (
	"encoding/json"
	"sort"
	"strings"

	commondaoimpl "openelis-go/internal/common/daoimpl"
	"openelis-go/internal/common/util"
	locform "openelis-go/internal/localization/form"
	"openelis-go/internal/testconfiguration/daoimpl"
	"openelis-go/internal/testconfiguration/form"
)

// SelectListService ports TestRenameEntryRestController,
// SelectListRenameEntryRestController and ResultSelectListAddRestController.
type SelectListService struct {
	Lists      *commondaoimpl.DisplayListDAOImpl
	Activation *daoimpl.ActivationDAO
	DAO        *daoimpl.SelectListDAO
}

// TestRenameForm ports showTestRenameEntry — ALL_TESTS, whose values are the
// augmented test names.
func (s *SelectListService) TestRenameForm() (*form.TestRenameEntryForm, error) {
	tests, err := s.Activation.TestsBySampleType()
	if err != nil {
		return nil, err
	}
	list := make([]util.IdValuePair, 0, len(tests))
	for _, t := range tests {
		if t.IsActive != "Y" {
			continue
		}
		list = append(list, util.NewIdValuePair(t.ID, t.AugmentedName))
	}
	// createTestList sorts by the displayed VALUE with String.compareTo, which
	// is byte order rather than a collation.
	sort.SliceStable(list, func(i, j int) bool { return list[i].Value < list[j].Value })

	f := form.NewTestRenameEntryForm()
	f.TestList = &list
	return &f, nil
}

// RenameTest ports updateTestRenameEntry.
//
// It sets cancelAction to "CancelDictionary" on the way out — a value no other
// screen in this package uses, and one the GET does not carry.
func (s *SelectListService) RenameTest(post form.SelectListPost) (*form.TestRenameEntryForm, error) {
	f := form.NewTestRenameEntryForm()
	f.CancelAction = "CancelDictionary"
	f.TestID = derefOr(post.TestID)
	f.NameEnglish, f.NameFrench = derefOr(post.NameEnglish), derefOr(post.NameFrench)
	f.ReportNameEnglish = derefOr(post.ReportNameEnglish)
	f.ReportNameFrench = derefOr(post.ReportNameFrench)

	if strings.TrimSpace(f.TestID) == "" {
		return &f, nil
	}
	err := s.DAO.RenameTest(f.TestID,
		strings.TrimSpace(f.NameEnglish), strings.TrimSpace(f.NameFrench),
		strings.TrimSpace(f.ReportNameEnglish), strings.TrimSpace(f.ReportNameFrench))
	if err != nil {
		// LIMSRuntimeException is caught and logged.
		return &f, nil
	}
	return &f, nil
}

// SelectListRenameForm ports showUomRenameEntry — the misnamed handler behind
// /SelectListRenameEntry.
func (s *SelectListService) SelectListRenameForm() (*form.SelectListRenameForm, error) {
	options, err := s.selectOptions()
	if err != nil {
		return nil, err
	}
	f := form.NewSelectListRenameForm()
	f.ResultSelectOptionList = &options
	return &f, nil
}

// selectOptions assembles the Dictionary entities the screen lists.
func (s *SelectListService) selectOptions() ([]daoimpl.SelectOption, error) {
	rows, err := s.DAO.SelectListOptions()
	if err != nil {
		return nil, err
	}
	locales, err := s.Lists.ActiveLocales()
	if err != nil {
		return nil, err
	}
	values, err := s.Lists.LocalizationValues()
	if err != nil {
		return nil, err
	}
	out := make([]daoimpl.SelectOption, 0, len(rows))
	for _, r := range rows {
		opt := daoimpl.SelectOption{
			Lastupdated:       r.Lastupdated,
			NameKey:           r.NameKey,
			ID:                r.ID,
			IsActive:          r.IsActive,
			DictEntry:         r.DictEntry,
			LocalAbbreviation: r.LocalAbbreviation,
			SortOrder:         r.SortOrder,
			// displayValue is getDisplayValue(), which falls back to the
			// dictionary entry when there is no localized name.
			DisplayValue: r.DictEntry,
		}
		if r.CategoryID != nil {
			opt.DictionaryCategory = &daoimpl.DictionaryCategoryDTO{
				Lastupdated:  r.CategoryUpdated,
				ID:           *r.CategoryID,
				CategoryName: derefOr(r.CategoryName),
				Description:  r.CategoryDesc,
				LocalAbbrev:  r.CategoryAbbrev,
			}
		}
		if r.LocalizationID != nil && *r.LocalizationID != "" {
			opt.LocalizedDictionaryName = locform.BuildLocalization(
				*r.LocalizationID, derefOr(r.LocDescription), r.LocUpdated,
				values[*r.LocalizationID], locales, s.Lists.Locale())
			if opt.LocalizedDictionaryName != nil &&
				opt.LocalizedDictionaryName.LocalizedValue != "" {
				opt.DisplayValue = opt.LocalizedDictionaryName.LocalizedValue
			}
		}
		out = append(out, opt)
	}
	return out, nil
}

// RenameSelectOption ports updateUomRenameEntry — the second handler with that
// name, on a different controller, renaming a dictionary entry.
//
// The response carries the REBUILT list either way: both the success and the
// failure branch call setResultSelectOptionList before returning.
func (s *SelectListService) RenameSelectOption(post form.SelectListPost, sysUserID int64) (*form.SelectListRenameForm, error) {
	f := form.NewSelectListRenameForm()
	f.ResultSelectOptionID = derefOr(post.ResultSelectOptionID)
	f.NameEnglish, f.NameFrench = derefOr(post.NameEnglish), derefOr(post.NameFrench)

	if strings.TrimSpace(f.ResultSelectOptionID) != "" {
		// A RuntimeException is caught and turned into `renamed = false`, which
		// only changes whether an error is flashed — the body is the same.
		_ = s.DAO.RenameSelectOption(f.ResultSelectOptionID, f.NameEnglish, f.NameFrench, sysUserID)
	}

	options, err := s.selectOptions()
	if err != nil {
		return nil, err
	}
	f.ResultSelectOptionList = &options
	return &f, nil
}

// ResultSelectListAdd ports showResultSelectListAddToTest — a POST that WRITES
// NOTHING.
//
// It sets page to "2", copies whichever name was left blank from the other
// language, and fills two lists. The screen's real write is
// /SaveResultSelectList.
func (s *SelectListService) ResultSelectListAdd(post form.SelectListPost) (*form.ResultSelectListForm, error) {
	f := form.NewResultSelectListForm()
	f.Page = "2"
	f.Normal, f.Qualifiable = post.Normal, post.Qualifiable
	en, fr := derefOr(post.NameEnglish), derefOr(post.NameFrench)
	// `if ("".equalsIgnoreCase(nameEnglish)) setNameEnglish(nameFrench)` — the
	// blank one takes the other's value, and a name that is null rather than
	// empty is left alone.
	if post.NameEnglish != nil && en == "" {
		en = fr
	} else if post.NameFrench != nil && fr == "" {
		fr = en
	}
	f.NameEnglish, f.NameFrench = &en, &fr
	f.LoincCode = post.LoincCode
	f.TestSelectListJSON = post.TestSelectListJSON

	tests, err := s.dictionaryResultTests()
	if err != nil {
		return nil, err
	}
	f.Tests = &tests
	dict, err := s.testSelectDictionary()
	if err != nil {
		return nil, err
	}
	f.TestDictionary = &dict
	return &f, nil
}

// dictionaryResultTests ports testService.getAllTestsByDictionaryResult.
func (s *SelectListService) dictionaryResultTests() ([]util.IdValuePair, error) {
	rows, err := s.Activation.TestsBySampleType()
	if err != nil {
		return nil, err
	}
	out := []util.IdValuePair{}
	seen := map[string]bool{}
	for _, t := range rows {
		if seen[t.ID] {
			continue
		}
		seen[t.ID] = true
		out = append(out, util.NewIdValuePair(t.ID, t.AugmentedName))
	}
	return out, nil
}

// testSelectDictionary ports resultSelectListService.getTestSelectDictionary:
// the dictionary options each test already offers.
func (s *SelectListService) testSelectDictionary() (map[string][]util.IdValuePair, error) {
	rows, err := s.DAO.TestSelectDictionary()
	if err != nil {
		return nil, err
	}
	out := map[string][]util.IdValuePair{}
	for _, r := range rows {
		out[r.TestID] = append(out[r.TestID], util.NewIdValuePair(r.DictionaryID, r.Value))
	}
	return out, nil
}

// SaveResultSelectList ports SaveResultSelectList — the real write.
//
// testSelectListJson is permissive in a way nothing else in this package is: it
// accepts EITHER a bare array or an object with a `tests` key holding a JSON
// string. Both are parsed; anything else is an error the caller never sees,
// because the failure branch returns the same body as the success one.
func (s *SelectListService) SaveResultSelectList(post form.SelectListPost, sysUserID int64) (*form.ResultSelectListForm, error) {
	f := form.NewResultSelectListForm()
	f.Normal, f.Qualifiable = post.Normal, post.Qualifiable
	f.NameEnglish, f.NameFrench = post.NameEnglish, post.NameFrench
	f.LoincCode, f.TestSelectListJSON = post.LoincCode, post.TestSelectListJSON

	byTest, ok := parseTestSelectList(derefOr(post.TestSelectListJSON))
	if !ok {
		return &f, nil
	}
	_, err := s.DAO.AddSelectList(
		derefOr(post.NameEnglish), derefOr(post.NameFrench), derefOr(post.LoincCode),
		byTest, sysUserID)
	if err != nil {
		return &f, nil
	}
	return &f, nil
}

// parseTestSelectList reads either accepted shape of testSelectListJson.
func parseTestSelectList(raw string) (map[string][]daoimpl.TestResultItem, bool) {
	if strings.TrimSpace(raw) == "" {
		return nil, false
	}
	type item struct {
		ID          *string     `json:"id"`
		Order       json.Number `json:"order"`
		Normal      bool        `json:"normal"`
		Qualifiable bool        `json:"qualifiable"`
	}
	type testEntry struct {
		ID    string `json:"id"`
		Items []item `json:"items"`
	}

	var entries []testEntry
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		// The object form: {"tests": "<json array as a string>"}
		var wrapper struct {
			Tests string `json:"tests"`
		}
		if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
			return nil, false
		}
		if err := json.Unmarshal([]byte(wrapper.Tests), &entries); err != nil {
			return nil, false
		}
	}

	out := map[string][]daoimpl.TestResultItem{}
	for _, e := range entries {
		for _, it := range e.Items {
			order, convErr := it.Order.Int64()
			if convErr != nil {
				continue
			}
			out[e.ID] = append(out[e.ID], daoimpl.TestResultItem{
				ID: it.ID, Order: int(order),
				Normal: it.Normal, Qualifiable: it.Qualifiable,
			})
		}
	}
	return out, true
}
