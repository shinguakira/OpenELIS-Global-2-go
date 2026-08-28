package form

// PendingAnalysisDTO mirrors the org.json object
// PendingAnalysisForTestProviderRestController.createJsonGroupedAnalysis
// builds: four groups, keyed by the tag passed to createPendingList.
//
// No omitempty anywhere. Java assembles this with org.json rather than Jackson,
// so the response never passes through Include.NON_NULL — an empty group is
// emitted as [] and the key is always present. Adding omitempty would drop
// keys Java always sends.
type PendingAnalysisDTO struct {
	NotStarted          []PendingAnalysisItem `json:"notStarted"`
	TechnicianRejection []PendingAnalysisItem `json:"technicianRejection"`
	BiologistRejection  []PendingAnalysisItem `json:"biologistRejection"`
	NotValidated        []PendingAnalysisItem `json:"notValidated"`
}

// PendingAnalysisItem is one pending analysis: the accession it belongs to and
// the analysis id AS A STRING (see the service for why the type differs from
// all-by-accession's numeric analysisId).
type PendingAnalysisItem struct {
	LabNo string `json:"labNo"`
	ID    string `json:"id"`
}
