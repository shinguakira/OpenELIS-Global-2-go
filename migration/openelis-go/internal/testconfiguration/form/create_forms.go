package form

import "openelis-go/internal/common/util"

// The *Create forms.
//
// Same rule as the rename ones: the lists are POINTERS because the POST success
// path does not call setupDisplayItems, so they are absent from a create's
// response and present on the validation-failure branch — which is also a 200
// for two of these three and a 400 for MethodCreate.

// MethodCreateForm is GET/POST /rest/MethodCreate.
type MethodCreateForm struct {
	FormName       string `json:"formName"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	ExistingMethodList   *[]util.IdValuePair `json:"existingMethodList,omitempty"`
	InactiveMethodList   *[]util.IdValuePair `json:"inactiveMethodList,omitempty"`
	ExistingEnglishNames *string             `json:"existingEnglishNames,omitempty"`
	ExistingFrenchNames  *string             `json:"existingFrenchNames,omitempty"`

	MethodEnglishName *string `json:"methodEnglishName,omitempty"`
	MethodFrenchName  *string `json:"methodFrenchName,omitempty"`
	MethodCode        *string `json:"methodCode,omitempty"`
}

// TestSectionCreateForm is GET/POST /rest/TestSectionCreate.
type TestSectionCreateForm struct {
	FormName       string `json:"formName"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	ExistingTestUnitList *[]util.IdValuePair `json:"existingTestUnitList,omitempty"`
	InactiveTestUnitList *[]util.IdValuePair `json:"inactiveTestUnitList,omitempty"`
	ExistingEnglishNames *string             `json:"existingEnglishNames,omitempty"`
	ExistingFrenchNames  *string             `json:"existingFrenchNames,omitempty"`

	TestUnitEnglishName *string `json:"testUnitEnglishName,omitempty"`
	TestUnitFrenchName  *string `json:"testUnitFrenchName,omitempty"`
}

// SampleTypeCreateForm is GET/POST /rest/SampleTypeCreate.
type SampleTypeCreateForm struct {
	FormName       string `json:"formName"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	ExistingSampleTypeList *[]util.IdValuePair `json:"existingSampleTypeList,omitempty"`
	InactiveSampleTypeList *[]util.IdValuePair `json:"inactiveSampleTypeList,omitempty"`
	ExistingEnglishNames   *string             `json:"existingEnglishNames,omitempty"`
	ExistingFrenchNames    *string             `json:"existingFrenchNames,omitempty"`

	SampleTypeEnglishName *string `json:"sampleTypeEnglishName,omitempty"`
	SampleTypeFrenchName  *string `json:"sampleTypeFrenchName,omitempty"`
}

// CreatePost is the bound body for all three. ALLOWED_FIELDS differs only in
// the names, so one struct decodes any of them and the service picks the pair
// its screen owns.
type CreatePost struct {
	MethodEnglishName     *string `json:"methodEnglishName"`
	MethodFrenchName      *string `json:"methodFrenchName"`
	MethodCode            *string `json:"methodCode"`
	TestUnitEnglishName   *string `json:"testUnitEnglishName"`
	TestUnitFrenchName    *string `json:"testUnitFrenchName"`
	SampleTypeEnglishName *string `json:"sampleTypeEnglishName"`
	SampleTypeFrenchName  *string `json:"sampleTypeFrenchName"`
}

// NewMethodCreateForm reproduces `new MethodCreateForm()`.
func NewMethodCreateForm() MethodCreateForm {
	return MethodCreateForm{
		FormName: "methodCreateForm", FormMethod: "POST",
		CancelAction: "Home", SubmitOnCancel: false, CancelMethod: "POST",
	}
}

// NewTestSectionCreateForm reproduces `new TestSectionCreateForm()`.
func NewTestSectionCreateForm() TestSectionCreateForm {
	return TestSectionCreateForm{
		FormName: "testSectionCreateForm", FormMethod: "POST",
		CancelAction: "Home", SubmitOnCancel: false, CancelMethod: "POST",
	}
}

// NewSampleTypeCreateForm reproduces `new SampleTypeCreateForm()`.
func NewSampleTypeCreateForm() SampleTypeCreateForm {
	return SampleTypeCreateForm{
		FormName: "sampleTypeCreateForm", FormMethod: "POST",
		CancelAction: "Home", SubmitOnCancel: false, CancelMethod: "POST",
	}
}
