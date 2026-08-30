package form

import (
	"openelis-go/internal/common/util"
	"openelis-go/internal/testconfiguration/daoimpl"
)

// TestRenameEntryForm is GET/POST /rest/TestRenameEntry.
//
// A test carries TWO localizations — its name and its reporting name — so this
// screen has four name fields where the other rename screens have two.
type TestRenameEntryForm struct {
	FormName       string `json:"formName"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	TestList *[]util.IdValuePair `json:"testList,omitempty"`

	NameEnglish       string `json:"nameEnglish"`
	NameFrench        string `json:"nameFrench"`
	ReportNameEnglish string `json:"reportNameEnglish"`
	ReportNameFrench  string `json:"reportNameFrench"`
	TestID            string `json:"testId"`
}

// SelectListRenameForm is GET/POST /rest/SelectListRenameEntry.
//
// It has NO formName — the bean does not set one, so the key is absent where
// every other screen in this package carries it.
type SelectListRenameForm struct {
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	ResultSelectOptionList *[]daoimpl.SelectOption `json:"resultSelectOptionList,omitempty"`

	NameEnglish          string `json:"nameEnglish"`
	NameFrench           string `json:"nameFrench"`
	ResultSelectOptionID string `json:"resultSelectOptionId"`
}

// ResultSelectListForm is POST /rest/ResultSelectListAdd and
// /rest/SaveResultSelectList.
//
// Like SelectListRenameForm it carries no formName. `normal` and `qualifiable`
// are primitives on the bean, so they are present as false rather than absent.
type ResultSelectListForm struct {
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	Normal      bool   `json:"normal"`
	Qualifiable bool   `json:"qualifiable"`
	Page        string `json:"page"`

	NameEnglish        *string                        `json:"nameEnglish,omitempty"`
	NameFrench         *string                        `json:"nameFrench,omitempty"`
	LoincCode          *string                        `json:"loincCode,omitempty"`
	TestSelectListJSON *string                        `json:"testSelectListJson,omitempty"`
	Tests              *[]util.IdValuePair            `json:"tests,omitempty"`
	TestDictionary     *map[string][]util.IdValuePair `json:"testDictionary,omitempty"`
}

// SelectListPost is the bound body for all four screens here.
type SelectListPost struct {
	TestID               *string `json:"testId"`
	NameEnglish          *string `json:"nameEnglish"`
	NameFrench           *string `json:"nameFrench"`
	ReportNameEnglish    *string `json:"reportNameEnglish"`
	ReportNameFrench     *string `json:"reportNameFrench"`
	ResultSelectOptionID *string `json:"resultSelectOptionId"`
	LoincCode            *string `json:"loincCode"`
	TestSelectListJSON   *string `json:"testSelectListJson"`
	Normal               bool    `json:"normal"`
	Qualifiable          bool    `json:"qualifiable"`
}

// NewTestRenameEntryForm reproduces `new TestRenameEntryForm()`.
func NewTestRenameEntryForm() TestRenameEntryForm {
	return TestRenameEntryForm{
		FormName: "testRenameEntryForm", FormMethod: "POST",
		CancelAction: "Home", SubmitOnCancel: false, CancelMethod: "POST",
	}
}

// NewSelectListRenameForm reproduces `new ResultSelectListRenameForm()`.
func NewSelectListRenameForm() SelectListRenameForm {
	return SelectListRenameForm{
		FormMethod: "POST", CancelAction: "Home",
		SubmitOnCancel: false, CancelMethod: "POST",
	}
}

// NewResultSelectListForm reproduces `new ResultSelectListForm()`, whose page
// initialiser is "1".
func NewResultSelectListForm() ResultSelectListForm {
	return ResultSelectListForm{
		FormMethod: "POST", CancelAction: "Home",
		SubmitOnCancel: false, CancelMethod: "POST", Page: "1",
	}
}
