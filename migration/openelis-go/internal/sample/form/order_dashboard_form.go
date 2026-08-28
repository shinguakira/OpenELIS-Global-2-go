package form

// OrderDashboardDTO mirrors the response map OrderSearchRestController
// .getDashboard builds. Every key is put unconditionally, so none omits.
type OrderDashboardDTO struct {
	Orders     []OrderDashboardItemDTO `json:"orders"`
	TotalCount int                     `json:"totalCount"`
	// Hardcoded 0 in Java and never computed — the source comment calls it a
	// "Placeholder for external orders count". Pinned, not fixed.
	ExternalCount int `json:"externalCount"`
	Page          int `json:"page"`
	// Echoed back verbatim. It does NOT bound the result — see the DAO.
	PageSize int `json:"pageSize"`
}

// OrderDashboardItemDTO is one row. Java builds it as a HashMap and puts every
// key unconditionally (patientName falls back to "---" rather than being
// omitted), so nothing here takes omitempty.
type OrderDashboardItemDTO struct {
	ID        string `json:"id"`
	LabNumber string `json:"labNumber"`
	// java.sql.Timestamp.toString() or "" — a formatted string, NOT the epoch
	// millis that the b2/c1 entity endpoints emit under `lastupdated`.
	LastUpdated string `json:"lastUpdated"`
	Priority    string `json:"priority"`
	// Real booleans, not the "Y"/"N" strings several tables in this schema use.
	// Both are hardcoded literals in Java.
	IsExternal     bool                 `json:"isExternal"`
	ReturnedFromQA bool                 `json:"returnedFromQA"`
	PatientName    string               `json:"patientName"`
	FacilityName   string               `json:"facilityName"`
	StepProgress   OrderStepProgressDTO `json:"stepProgress"`
	Status         string               `json:"status"`
	StorageSkipped bool                 `json:"storageSkipped"`
}

// OrderStepProgressDTO is the four-flag workflow map.
type OrderStepProgressDTO struct {
	Enter   bool `json:"enter"`
	Collect bool `json:"collect"`
	Label   bool `json:"label"`
	QA      bool `json:"qa"`
}
