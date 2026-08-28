package service

import (
	"strconv"
	"strings"
	"time"

	"openelis-go/internal/sample/daoimpl"
	"openelis-go/internal/sample/form"
)

// OrderDashboardService backs GET rest/order/dashboard.
type OrderDashboardService struct {
	DAO *daoimpl.SampleDAOImpl
}

// DashboardQuery carries the request parameters. Every field maps to a
// @RequestParam on OrderSearchRestController.getDashboard.
type DashboardQuery struct {
	Page     int
	PageSize int
	Search   string
	Status   string
	Priority string
	// IncludeExternal is accepted and NEVER READ, exactly as in Java. Kept as a
	// field so the inertness is visible in the code rather than being an
	// omission someone later "fixes".
	IncludeExternal bool
	StartDate       string
	EndDate         string
}

// GetDashboard mirrors getDashboard.
//
// THREE JAVA QUIRKS ARE PINNED HERE, not fixed — the c2 spec asserts each:
//  1. pageSize is echoed back but does not bound the result. It only shifts the
//     offset; the row count comes from page.defaultPageSize (see
//     SampleDAOImpl.DashboardPage).
//  2. externalCount is hardcoded 0 and never computed.
//  3. includeExternal is accepted and never read.
//
// totalCount is likewise ordersList.size(), i.e. the size of the CURRENT page
// after filtering — Java's own comment says "Simplified, should be total
// count".
func (s *OrderDashboardService) GetDashboard(q DashboardQuery) (*form.OrderDashboardDTO, error) {
	page, pageSize := q.Page, q.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}
	startingRecNo := ((page - 1) * pageSize) + 1

	rows, err := s.DAO.DashboardPage(startingRecNo)
	if err != nil {
		return nil, err
	}

	orders := make([]form.OrderDashboardItemDTO, 0, len(rows))
	for _, r := range rows {
		patientName := "---"
		if r.HasPatient {
			first, last := deref(r.PatientFirstName), deref(r.PatientLastName)
			// Java concatenates with a single space and then trims, so a
			// missing half leaves no leading/trailing space — but a patient
			// with neither name yields "" rather than "---", because the
			// "---" fallback only fires when there is no patient at all.
			patientName = strings.TrimSpace(first + " " + last)
		}

		priority := "routine"
		if r.OrderPriority != nil && *r.OrderPriority != "" {
			priority = strings.ToLower(*r.OrderPriority)
		}

		// Java's search filter is applied in Java, over the already-paged rows,
		// against the lab number OR the patient name, case-insensitively.
		if q.Search != "" {
			needle := strings.ToLower(q.Search)
			if !strings.Contains(strings.ToLower(r.AccessionNumber), needle) &&
				!strings.Contains(strings.ToLower(patientName), needle) {
				continue
			}
		}
		if q.Priority != "" && !strings.EqualFold(q.Priority, "all") &&
			!strings.EqualFold(priority, q.Priority) {
			continue
		}
		// Date filters compare against entered_date and are SKIPPED on an
		// unparseable value — Java catches IllegalArgumentException and falls
		// through rather than rejecting the request.
		if !withinDateFilter(r, q.StartDate, q.EndDate) {
			continue
		}

		storageSkipped := r.StorageSkipped != nil && *r.StorageSkipped

		// collectComplete: only sample items that HAVE analyses are considered,
		// and all of those must carry a collection date. With no such items the
		// flag stays false.
		collectComplete := r.ItemsWithTests > 0 && r.ItemsWithTestsDated == r.ItemsWithTests

		// labelComplete: storage-skipped short-circuits to true; otherwise
		// every sample item needs a storage assignment WITH a location. An
		// order with no sample items stays false, because Java guards the
		// allMatch behind a non-empty check (an unguarded allMatch over an
		// empty stream would have returned true).
		labelComplete := storageSkipped ||
			(r.ItemsTotal > 0 && r.ItemsWithStorageLoc == r.ItemsTotal)

		qaComplete := r.QAAllRequiredVerified

		orderStatus := "in_progress"
		switch {
		case qaComplete:
			orderStatus = "completed"
		case labelComplete:
			orderStatus = "pending_qa"
		}
		if q.Status != "" && !strings.EqualFold(q.Status, "all") && orderStatus != q.Status {
			continue
		}

		item := form.OrderDashboardItemDTO{
			ID:        strconv.FormatInt(r.ID, 10),
			LabNumber: r.AccessionNumber,
			// java.sql.Timestamp.toString(), or "" when null — a FORMATTED
			// STRING, not the epoch millis the b2/c1 entity endpoints emit
			// under the lowercase `lastupdated`. Same concept, different key
			// AND different type.
			LastUpdated: "",
			Priority:    priority,
			// Hardcoded on both: Java writes the literals.
			IsExternal:     false,
			ReturnedFromQA: false,
			PatientName:    patientName,
			FacilityName:   "---",
			StepProgress: form.OrderStepProgressDTO{
				Enter:   isEnterComplete(r),
				Collect: collectComplete,
				Label:   labelComplete,
				QA:      qaComplete,
			},
			Status:         orderStatus,
			StorageSkipped: storageSkipped,
		}
		if r.Lastupdated != nil {
			item.LastUpdated = formatSQLTimestamp(*r.Lastupdated)
		}
		orders = append(orders, item)
	}

	return &form.OrderDashboardDTO{
		Orders:     orders,
		TotalCount: len(orders),
		// Hardcoded 0, never computed — Java's own placeholder.
		ExternalCount: 0,
		Page:          page,
		// Echoed verbatim even though it did not bound anything.
		PageSize: pageSize,
	}, nil
}

// isEnterComplete ports OrderSearchRestController.isEnterComplete: a received
// date is required, then EITHER a linked patient (clinical) OR a recorded
// envWorkflowType observation (environmental).
func isEnterComplete(r daoimpl.OrderDashboardRow) bool {
	if r.ReceivedDate == nil {
		return false
	}
	if r.HasPatient {
		return true
	}
	return r.HasEnvWorkflowType
}

// withinDateFilter applies the startDate/endDate filters against ENTERED_DATE
// (not received or collection date), matching Java.
//
// An unparseable bound is IGNORED rather than rejected: Java wraps each
// java.sql.Date.valueOf in a try/catch that swallows IllegalArgumentException
// and falls through, so "startDate=garbage" silently disables that half of the
// filter instead of 400ing.
//
// A sample with a NULL entered_date is EXCLUDED whenever either bound is set,
// because Java tests `sampleDate == null || sampleDate.before(...)` — the null
// check is inside the skip condition, not a pass-through.
//
// The comparisons are strict (before/after), so both bounds are inclusive.
func withinDateFilter(r daoimpl.OrderDashboardRow, startDate, endDate string) bool {
	if start, ok := parseFilterDate(startDate); ok {
		if r.EnteredDate == nil || r.EnteredDate.Before(start) {
			return false
		}
	}
	if end, ok := parseFilterDate(endDate); ok {
		if r.EnteredDate == nil || r.EnteredDate.After(end) {
			return false
		}
	}
	return true
}

// parseFilterDate accepts exactly what java.sql.Date.valueOf does: "yyyy-mm-dd".
// Anything else (including an empty value) means "no bound".
func parseFilterDate(v string) (time.Time, bool) {
	if v == "" {
		return time.Time{}, false
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
