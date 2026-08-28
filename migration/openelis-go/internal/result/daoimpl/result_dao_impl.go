// Package daoimpl ports the queries behind rest/LogbookResults and
// rest/accession-results (constitution.md Layer II).
package daoimpl

import "gorm.io/gorm"

// ResultDAOImpl backs Wave 5.1 and 5.2.
type ResultDAOImpl struct {
	DB *gorm.DB

	// ActiveLocale is site_information "default language locale".
	ActiveLocale string
}

// Locale returns the configured locale, or "en" when unset.
func (d *ResultDAOImpl) Locale() string {
	if d.ActiveLocale != "" {
		return d.ActiveLocale
	}
	return "en"
}

// LogbookRow carries everything the logbook row builder reads.
type LogbookRow struct {
	AccessionNumber string  `gorm:"column:accession_number"`
	SequenceNumber  *string `gorm:"column:sequence_number"`
	SampleItemID    string  `gorm:"column:sample_item_id"`
	SampleType      *string `gorm:"column:sample_type"`

	TestDate     *string `gorm:"column:test_date"`
	ReceivedDate *string `gorm:"column:received_date"`

	TestID        string  `gorm:"column:test_id"`
	TestName      string  `gorm:"column:test_name"`
	TestSortOrder *string `gorm:"column:test_sort_order"`
	UnitOfMeasure *string `gorm:"column:unit_of_measure"`

	AnalysisID       string  `gorm:"column:analysis_id"`
	AnalysisStatusID string  `gorm:"column:analysis_status_id"`
	StatusIsRejected bool    `gorm:"column:status_is_rejected"`
	AnalysisType     *string `gorm:"column:analysis_type"`
	IsReportable     *string `gorm:"column:is_reportable"`

	ResultID    *string `gorm:"column:result_id"`
	ResultValue *string `gorm:"column:result_value"`
	ResultType  *string `gorm:"column:result_type"`
	SigDigits   *int    `gorm:"column:sig_digits"`
	Grouping    *int    `gorm:"column:grouping"`

	FallbackResultType *string `gorm:"column:fallback_result_type"`
	ReferredOut        bool    `gorm:"column:referred_out"`
	ReferralID         *string `gorm:"column:referral_id"`
	ReferralReasonID   *string `gorm:"column:referral_reason_id"`

	LimitID        *string  `gorm:"column:limit_id"`
	LowNormal      *float64 `gorm:"column:low_normal"`
	HighNormal     *float64 `gorm:"column:high_normal"`
	LowValid       *float64 `gorm:"column:low_valid"`
	HighValid      *float64 `gorm:"column:high_valid"`
	LowCritical    *float64 `gorm:"column:low_critical"`
	HighCritical   *float64 `gorm:"column:high_critical"`
	LimitSigDigits *int     `gorm:"column:limit_sig_digits"`

	PatientID        *string `gorm:"column:patient_id"`
	PatientFirstName *string `gorm:"column:patient_first_name"`
	PatientLastName  *string `gorm:"column:patient_last_name"`
	NationalID       *string `gorm:"column:national_id"`
	Gender           *string `gorm:"column:gender"`
	BirthDate        *string `gorm:"column:birth_date"`
	EnteredBirthDate *string `gorm:"column:entered_birth_date"`
}

// logbookSelect formats the two dates IN SQL, and they are NOT the same format:
// testDate carries a time and receivedDate does not, even though the workplan
// rows render the received date WITH one. Same column, three renderings across
// the wave.
const logbookSelect = `s.accession_number AS accession_number,
	si.sort_order::text AS sequence_number,
	si.id::text AS sample_item_id,
	COALESCE(tlv.value, tos.description) AS sample_type,
	-- testDate is the CLOCK, not a stored column: it moves between two calls
	-- seconds apart. The analysis's entry_date looks like the obvious source
	-- and is not it.
	to_char(now(), 'DD/MM/YYYY HH24:MI') AS test_date,
	to_char(s.received_date, 'DD/MM/YYYY') AS received_date,
	a.test_id::text AS test_id,
	COALESCE(lv.value, t.name) || COALESCE(
		(SELECT '(' || COALESCE(stlv.value, stos.description) || ')'
		   FROM clinlims.sampletype_test AS tost
		   JOIN clinlims.type_of_sample AS stos ON stos.id = tost.sample_type_id
		   LEFT JOIN clinlims.localization AS stl ON stl.id = stos.name_localization_id
		   LEFT JOIN clinlims.localization_value AS stlv ON stlv.localization_id = stl.id AND stlv.locale = @loc
		  WHERE tost.test_id = t.id AND stos.local_abbrev IS DISTINCT FROM 'Variable'
		  ORDER BY tost.ctid LIMIT 1), '') AS test_name,
	t.sort_order AS test_sort_order,
	uom.name AS unit_of_measure,
	a.id::text AS analysis_id,
	a.status_id::text AS analysis_status_id,
	(SELECT st.name = 'Technical Rejected' FROM clinlims.status_of_sample st
	  WHERE st.id = a.status_id AND st.status_type = 'ANALYSIS') AS status_is_rejected,
	a.analysis_type AS analysis_type,
	-- reportable is the TEST's column, not the analysis's. Every analysis here
	-- is is_reportable='Y' while every test is 'N', so reading the wrong one is
	-- wrong on every row.
	t.is_reportable AS is_reportable,
	-- resultType falls back to the TEST_RESULT's type when the analysis has no
	-- result yet — a referred-out row still reports 'N'.
	(SELECT tr.tst_rslt_type FROM clinlims.test_result tr
	  WHERE tr.test_id = t.id ORDER BY tr.id LIMIT 1) AS fallback_result_type,
	(r.id IS NOT NULL) AS referred_out,
	r.id::text AS referral_id,
	r.referral_reason_id::text AS referral_reason_id,
	res.id::text AS result_id,
	res.value AS result_value,
	res.result_type AS result_type,
	res.significant_digits AS sig_digits,
	res.grouping AS grouping,
	rl.id::text AS limit_id,
	rl.low_normal AS low_normal,
	rl.high_normal AS high_normal,
	rl.low_valid AS low_valid,
	rl.high_valid AS high_valid,
	rl.low_critical AS low_critical,
	rl.high_critical AS high_critical,
	(SELECT tr.significant_digits FROM clinlims.test_result tr
	  WHERE tr.test_id = t.id ORDER BY tr.id LIMIT 1) AS limit_sig_digits,
	pa.id::text AS patient_id,
	pe.first_name AS patient_first_name,
	pe.last_name AS patient_last_name,
	pa.national_id AS national_id,
	pa.gender AS gender,
	to_char(pa.birth_date, 'DD/MM/YYYY') AS birth_date,
	pa.entered_birth_date AS entered_birth_date`

func (d *ResultDAOImpl) base() *gorm.DB {
	return d.DB.Table("clinlims.analysis AS a").
		Select(logbookSelect, map[string]any{"loc": d.Locale()}).
		Joins("JOIN clinlims.sample_item AS si ON si.id = a.sampitem_id").
		Joins("JOIN clinlims.sample AS s ON s.id = si.samp_id").
		Joins("LEFT JOIN clinlims.type_of_sample AS tos ON tos.id = si.typeosamp_id").
		Joins("LEFT JOIN clinlims.localization AS tl ON tl.id = tos.name_localization_id").
		Joins("LEFT JOIN clinlims.localization_value AS tlv ON tlv.localization_id = tl.id AND tlv.locale = ?", d.Locale()).
		Joins("JOIN clinlims.test AS t ON t.id = a.test_id").
		Joins("LEFT JOIN clinlims.localization AS l ON l.id = t.name_localization_id").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = l.id AND lv.locale = ?", d.Locale()).
		Joins("LEFT JOIN clinlims.unit_of_measure AS uom ON uom.id = t.uom_id").
		Joins("LEFT JOIN clinlims.result AS res ON res.analysis_id = a.id").
		Joins("LEFT JOIN clinlims.referral AS r ON r.analysis_id = a.id").
		// The patient joins come BEFORE the result_limits one because the band
		// lookup below reads the patient's gender and birth date.
		Joins("LEFT JOIN clinlims.sample_human AS sh ON sh.samp_id = s.id").
		Joins("LEFT JOIN clinlims.patient AS pa ON pa.id = sh.patient_id").
		Joins("LEFT JOIN clinlims.person AS pe ON pe.id = pa.person_id").
		// result_limits is per test AND per AGE BAND (min_age/max_age are in
		// DAYS) and optionally per gender, so a test can carry six rows. Java
		// resolves ONE through getResultLimitForAnalysis; joining on test_id
		// alone multiplies every analysis by its band count — measured: one
		// order went from 4 rows to 9.
		Joins(`LEFT JOIN clinlims.result_limits AS rl ON rl.id = (
			SELECT rl2.id FROM clinlims.result_limits rl2
			 WHERE rl2.test_id = t.id
			   AND (rl2.gender IS NULL OR rl2.gender = '' OR rl2.gender = pa.gender)
			   AND (pa.birth_date IS NULL OR (
			         EXTRACT(EPOCH FROM age(now(), pa.birth_date)) / 86400 >= rl2.min_age
			     AND EXTRACT(EPOCH FROM age(now(), pa.birth_date)) / 86400 <  rl2.max_age))
			 ORDER BY rl2.id LIMIT 1)`)
}

// ByAccessionNumberDesc is accession-results' ordering: sample items DESCENDING.
// The logbook walks the same order the other way, so the two endpoints present
// one order's items in opposite sequences.
func (d *ResultDAOImpl) ByAccessionNumberDesc(accessionNumber string) ([]LogbookRow, error) {
	rows := []LogbookRow{}
	err := d.base().
		Where("s.accession_number = ?", accessionNumber).
		Order("si.sort_order DESC, a.id").
		Scan(&rows).Error
	return rows, err
}

// ByAccessionNumber returns every analysis on one order, in sample-item then
// test-sort order — the grouping the logbook renders.
func (d *ResultDAOImpl) ByAccessionNumber(accessionNumber string) ([]LogbookRow, error) {
	rows := []LogbookRow{}
	err := d.base().
		Where("s.accession_number = ?", accessionNumber).
		Order("si.sort_order, t.sort_order::int, a.id").
		Scan(&rows).Error
	return rows, err
}

// CurrentDate is DateUtil.getCurrentDateAsText — dd/MM/yyyy — read from the
// DATABASE clock rather than the Go process's.
//
// time.Now() takes the process's local zone, and Go on Windows does not honour
// the TZ environment variable, so the port rendered tomorrow's date for the
// nine hours after local midnight while Java, in UTC, was still on today.
// Reading the same clock Java formats from removes the question.
func (d *ResultDAOImpl) CurrentDate() (string, error) {
	values := []string{}
	err := d.DB.Raw("SELECT to_char(now(), 'DD/MM/YYYY')").Scan(&values).Error
	if err != nil || len(values) == 0 {
		return "", err
	}
	return values[0], nil
}

// ByTest returns every analysis of one test.
//
// A single analysis sitting on a sample item with a NULL typeosamp_id makes
// Java throw from getTestDisplayName, so callers must check
// TestHasTypelessItem first — see that method.
func (d *ResultDAOImpl) ByTest(testID string) ([]LogbookRow, error) {
	rows := []LogbookRow{}
	err := d.base().
		Where("a.test_id = ?", testID).
		// DESCENDING on the SEQUENCE ACCESSION — accession plus the sample item's
		// sequence, not the accession alone. reverseSortByAccessionAndSequence
		// compares getSequenceAccessionNumber(), and the difference is visible
		// whenever one order has several items:
		//
		//   E2E-EDIT-01 items 1 and 2 come back 2-then-1, reversed
		//   E2E-RES-01's two analyses share item 1, tie, and keep scan order
		//
		// An earlier version of this ordering used `a.id DESC`, derived from a
		// two-row sample where the two happened to agree. Adding a third analysis
		// separated them.
		//
		// COLLATE "C" because Java compares with String.compareTo — byte order,
		// not the database collation. ctid last, for the stable sort's tie.
		Order(`(s.accession_number || '-' || si.sort_order) COLLATE "C" DESC, a.ctid`).
		Scan(&rows).Error
	return rows, err
}

// TestHasTypelessItem reports whether any analysis of the test sits on a sample
// item with no type of sample.
//
// JAVA DEFECT, reproduced rather than fixed: AnalysisServiceImpl
// .getTestDisplayName dereferences sampleItem.getTypeOfSampleId() unguarded, so
// one such row makes rest/LogbookResults?selectedTest=N answer 500 — while
// Java's own unassigned-sample HQL is written to tolerate exactly that state.
func (d *ResultDAOImpl) TestHasTypelessItem(testID string) (bool, error) {
	var n int64
	err := d.DB.Table("clinlims.analysis AS a").
		Joins("JOIN clinlims.sample_item AS si ON si.id = a.sampitem_id").
		Where("a.test_id = ? AND si.typeosamp_id IS NULL", testID).
		Count(&n).Error
	return n > 0, err
}
