// Package form holds the c2 sample DTOs — constitution.md Layer V.
package form

// SampleSearchForm mirrors org.openelisglobal.sample.form.SampleSearchForm,
// the row shape GET rest/sample/all-by-accession/{n} returns a list of.
//
// TYPE TRAP, pinned by the c2 e2e spec: `id` and `analysisId` are NUMBERS here.
// Java builds them with Integer.parseInt(sample.getId()) /
// Integer.parseInt(analysis.getId()) even though both are strings on the
// entities, so Jackson emits bare integers. Every b1/b2 reference endpoint in
// this port stringifies its ids instead — applying that habit here would emit
// "1000" where Java emits 1000 and diverge on every row.
//
// sampleType and referralTest are omitted when null: Java only calls the
// setter inside a null guard (`if (typeOfSample != null)` /
// `if (analysis.getTest() != null)`), and the global Include.NON_NULL then
// drops the unset field rather than emitting null.
type SampleSearchForm struct {
	ID              int64   `json:"id"`
	AccessionNumber string  `json:"accessionNumber"`
	SampleType      *string `json:"sampleType,omitempty"`
	ReferralTest    *string `json:"referralTest,omitempty"`
	AnalysisID      int64   `json:"analysisId"`
}
