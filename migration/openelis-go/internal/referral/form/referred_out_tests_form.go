// Package form ports org.openelisglobal.referral.form.ReferredOutTestsForm
// (constitution.md Layer V). Folder layout mirrors the Java source.
package form

import (
	"openelis-go/internal/common/util"
	sampleform "openelis-go/internal/sample/form"
)

// ReferredOutTestsForm is what rest/ReferredOutTests returns.
//
// formName is "referredOutTestsForm" — LOWERCASE initial, unlike the
// WorkplanForm / LogbookResultsForm / ResultValidationForm siblings. Pinned
// exactly rather than normalised.
//
// searchType, labNumber and referralDisplayItems are POINTERS because the
// endpoint emits them only once a search has run: setupPageForDisplay fills
// referralDisplayItems solely when form.getSearchType() != null, and the other
// two are simply the bound request parameters echoed back. An unparameterised
// call carries neither.
type ReferredOutTestsForm struct {
	FormName       string `json:"formName"`
	FormMethod     string `json:"formMethod"`
	CancelAction   string `json:"cancelAction"`
	SubmitOnCancel bool   `json:"submitOnCancel"`
	CancelMethod   string `json:"cancelMethod"`

	SearchType            *string                   `json:"searchType,omitempty"`
	ReferralDisplayItems  *[]ReferralDisplayItemDTO `json:"referralDisplayItems,omitempty"`
	TestUnitSelectionList []util.IdValuePair        `json:"testUnitSelectionList"`
	TestSelectionList     []util.IdValuePair        `json:"testSelectionList"`
	LabNumber             *string                   `json:"labNumber,omitempty"`

	PatientSearch  sampleform.PatientSearchDTO `json:"patientSearch"`
	SearchFinished bool                        `json:"searchFinished"`
}

// ReferralDisplayItemDTO is one referral row.
//
// referralResultsDisplay, resultDate and notes are absent on a referral with no
// results and no notes, so they are pointers: convertToDisplayItem sets the
// first two only inside `if (!resultList.isEmpty())`, and notes comes back null
// when the analysis has none.
//
// referralStatus and referralStatusDisplay are the SAME value twice — one is
// the enum, the other its toString. Emitted separately because the screen reads
// both.
type ReferralDisplayItemDTO struct {
	AccessionNumber        string  `json:"accessionNumber"`
	ReferredSendDate       string  `json:"referredSendDate"`
	ReferralStatus         string  `json:"referralStatus"`
	ReferralStatusDisplay  string  `json:"referralStatusDisplay"`
	PatientLastName        string  `json:"patientLastName"`
	PatientFirstName       string  `json:"patientFirstName"`
	ReferringTestName      string  `json:"referringTestName"`
	ReferralResultsDisplay *string `json:"referralResultsDisplay,omitempty"`
	ResultDate             *string `json:"resultDate,omitempty"`
	ReferenceLabDisplay    *string `json:"referenceLabDisplay,omitempty"`
	Notes                  *string `json:"notes,omitempty"`
	AnalysisID             string  `json:"analysisId"`
}

// NewReferredOutTestsForm returns the envelope literals.
func NewReferredOutTestsForm() ReferredOutTestsForm {
	return ReferredOutTestsForm{
		FormName:     "referredOutTestsForm",
		FormMethod:   "POST",
		CancelAction: "Home",
		CancelMethod: "POST",
	}
}

// SearchTypes is the ReferredOutTestsForm.SearchType enum. A value outside it
// is a BINDING failure — but not the RFC 7807 ProblemDetail the WorkPlan
// endpoints produce: this form is @Valid-bound, so Spring answers with a
// per-field `errors` map instead. Two binding failures, two envelopes.
func SearchTypes() []string { return []string{"TEST_AND_DATES", "LAB_NUMBER", "PATIENT"} }

// IsSearchType reports whether v is a member of the enum.
func IsSearchType(v string) bool {
	for _, s := range SearchTypes() {
		if s == v {
			return true
		}
	}
	return false
}

// BindErrorBody is the shape Spring returns for a failed @Valid form binding:
//
//	{"timestamp":...,"status":400,"errors":{"<field>":"<message>"}}
type BindErrorBody struct {
	Timestamp int64             `json:"timestamp"`
	Status    int               `json:"status"`
	Errors    map[string]string `json:"errors"`
}
