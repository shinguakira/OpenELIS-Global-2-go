package daoimpl

import "time"

// DefaultPageSize is `page.defaultPageSize` from SystemConfiguration.properties.
//
// It is what actually bounds the dashboard's result set — the request's own
// pageSize does NOT (see DashboardPage). Mirrored as a constant because the Go
// service cannot read the Java container's properties file; if a deployment
// changes it, this must change with it.
const DefaultPageSize = 20

// OrderDashboardRow is one dashboard row, with every derived flag computed in
// SQL rather than by re-querying per sample.
//
// Java does this with a per-sample cascade of service calls (sample items,
// their analyses, their storage assignments, the QA checklist, an observation
// lookup). Folding it into one query changes the number of round trips, not
// the answer — each aggregate below reproduces one of those checks exactly, and
// the c2 spec's DB oracle compares the resulting rows against the sample table.
type OrderDashboardRow struct {
	ID              int64      `gorm:"column:id"`
	AccessionNumber string     `gorm:"column:accession_number"`
	Lastupdated     *time.Time `gorm:"column:lastupdated"`
	OrderPriority   *string    `gorm:"column:order_priority"`
	StorageSkipped  *bool      `gorm:"column:storage_skipped"`
	ReceivedDate    *time.Time `gorm:"column:received_date"`
	EnteredDate     *time.Time `gorm:"column:entered_date"`

	PatientFirstName *string `gorm:"column:patient_first_name"`
	PatientLastName  *string `gorm:"column:patient_last_name"`
	HasPatient       bool    `gorm:"column:has_patient"`

	ItemsTotal            int64 `gorm:"column:items_total"`
	ItemsWithTests        int64 `gorm:"column:items_with_tests"`
	ItemsWithTestsDated   int64 `gorm:"column:items_with_tests_dated"`
	ItemsWithStorageLoc   int64 `gorm:"column:items_with_storage_loc"`
	QAAllRequiredVerified bool  `gorm:"column:qa_all_required_verified"`
	HasEnvWorkflowType    bool  `gorm:"column:has_env_workflow_type"`
}

// DashboardPage mirrors SampleDAOImpl.getPageOfSamples, INCLUDING its paging
// bug — which the c2 spec pins rather than fixes.
//
//	int endingRecNo = startingRecNo + (page.defaultPageSize + 1);
//	query.setFirstResult(startingRecNo - 1);
//	query.setMaxResults(endingRecNo - 1);
//
// so maxResults = startingRecNo + defaultPageSize. Two consequences:
//   - the REQUEST's pageSize never bounds the result. It only shifts the
//     offset, so asking for 1 row still returns defaultPageSize + 1 of them
//     while the echoed pageSize says 1.
//   - the limit GROWS with the offset, so later pages return more rows than
//     earlier ones.
//
// Both are reproduced deliberately. `order by s.id` is Java's.
func (d *SampleDAOImpl) DashboardPage(startingRecNo int) ([]OrderDashboardRow, error) {
	offset := startingRecNo - 1
	if offset < 0 {
		offset = 0
	}
	limit := startingRecNo + DefaultPageSize

	rows := []OrderDashboardRow{}
	err := d.DB.Table("clinlims.sample AS s").
		Select(`s.id, s.accession_number, s.lastupdated, s.order_priority,
			s.storage_skipped, s.received_date, s.entered_date,
			pe.first_name AS patient_first_name, pe.last_name AS patient_last_name,
			(sh.patient_id IS NOT NULL) AS has_patient,
			COALESCE(agg.items_total, 0)          AS items_total,
			COALESCE(agg.items_with_tests, 0)     AS items_with_tests,
			COALESCE(agg.items_with_tests_dated, 0) AS items_with_tests_dated,
			COALESCE(agg.items_with_storage_loc, 0) AS items_with_storage_loc,
			COALESCE(qa.all_required_verified, false) AS qa_all_required_verified,
			(oh.id IS NOT NULL) AS has_env_workflow_type`).
		Joins("LEFT JOIN clinlims.sample_human AS sh ON sh.samp_id = s.id").
		Joins("LEFT JOIN clinlims.patient AS pa ON pa.id = sh.patient_id").
		Joins("LEFT JOIN clinlims.person AS pe ON pe.id = pa.person_id").
		Joins(`LEFT JOIN (
			SELECT si.samp_id,
			       count(*) AS items_total,
			       count(*) FILTER (WHERE ana.n > 0) AS items_with_tests,
			       count(*) FILTER (WHERE ana.n > 0 AND si.collection_date IS NOT NULL) AS items_with_tests_dated,
			       count(*) FILTER (WHERE ssa.location_id IS NOT NULL) AS items_with_storage_loc
			  FROM clinlims.sample_item si
			  LEFT JOIN LATERAL (
			       SELECT count(*) AS n FROM clinlims.analysis a WHERE a.sampitem_id = si.id
			  ) ana ON true
			  LEFT JOIN clinlims.sample_storage_assignment ssa ON ssa.sample_item_id = si.id
			 GROUP BY si.samp_id
			) AS agg ON agg.samp_id = s.id`).
		Joins("LEFT JOIN clinlims.sample_qa_checklist AS qa ON qa.sample_id = s.id").
		Joins(`LEFT JOIN clinlims.observation_history AS oh
			ON oh.sample_id = s.id
			AND oh.observation_history_type_id = (
			    SELECT id FROM clinlims.observation_history_type WHERE type_name = 'envWorkflowType'
			)`).
		Order("s.id").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
