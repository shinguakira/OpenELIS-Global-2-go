// Package daoimpl — the count/identity lookups behind
// GET rest/patient/merge/details/{patientId}.
package daoimpl

// PatientIdentityRow is a scan target for a patient_identity row LEFT-joined to
// its type name.
//
// IdentityTypeName is a POINTER because the join is outer: nil means the row's
// identity_type_id is NULL and no type could be resolved. Java reaches that
// state too — it loads the row and then throws inside
// patientIdentityTypeService.get(null) — so nil must become a 500, not a
// dropped row. See GetPatientIdentities.
type PatientIdentityRow struct {
	IdentityTypeName *string `gorm:"column:identity_type"`
	IdentityData     string  `gorm:"column:identity_data"`
}

// GetPatientIdentities mirrors PatientMergeServiceImpl.getPatientIdentities:
// SELECT * FROM patient_identity WHERE patient_id = :patientId
// joined here to patient_identity_type so the display name is available in one
// query. Returns ALL rows unfiltered — the GUID/AKA/MOTHER/MOTHERS_INITIAL
// exclusion is a DISPLAY concern and is applied in the service, because Java
// counts the unfiltered list for totalIdentifiers while listing the filtered
// one. Filtering here would make that documented mismatch impossible to
// reproduce.
//
// LEFT JOIN, not an inner one. patient_identity.identity_type_id is NULLABLE,
// and Java's query is `SELECT * FROM patient_identity WHERE patient_id = ?` —
// no join at all — so a row with a null type is still LOADED and still counted
// in totalIdentifiers. An inner join silently drops it, which understates
// totalIdentifiers and turns Java's error into a 200.
//
// Measured live by seeding one such row: Java answers 500 (its
// patientIdentityTypeService.get(null) throws inside the identifier loop),
// while the inner-join version of this query answered 200 with the row simply
// missing. IdentityType is therefore a *string: nil means "the row exists but
// its type is unresolvable", which the service turns back into Java's 500.
//
// A dangling (non-null but absent) type id cannot occur — patient_identity has
// a FK to patient_identity_type — so nil is unambiguously the null case.
func (d *PatientDAOImpl) GetPatientIdentities(patientID string) ([]PatientIdentityRow, error) {
	rows := []PatientIdentityRow{}
	err := d.DB.Table("clinlims.patient_identity AS pi").
		Select("pit.identity_type, pi.identity_data").
		Joins("LEFT JOIN clinlims.patient_identity_type pit ON pit.id = pi.identity_type_id").
		Where("pi.patient_id = ?", patientID).
		Order("pi.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// CountOrdersForPatient mirrors PatientMergeServiceImpl.countOrdersForPatient:
// SELECT COUNT(*) FROM sample_human WHERE patient_id = :patientId
//
// Despite the name "orders", this counts SAMPLE_HUMAN rows. Java's field naming
// reads backwards from the tables (see CountSampleItemsForPatient below);
// preserved deliberately.
func (d *PatientDAOImpl) CountOrdersForPatient(patientID string) (int64, error) {
	var n int64
	err := d.DB.Table("clinlims.sample_human").Where("patient_id = ?", patientID).Count(&n).Error
	return n, err
}

// CountSampleItemsForPatient mirrors countSamplesForPatient, which walks every
// Sample for the patient and sums their SampleItems. It populates the field
// Java calls `totalSamples` — so `totalSamples` is a count of sample_ITEM rows,
// not of samples. Confirmed against live data (21 items vs 19 samples) and
// pinned by the c1 e2e spec's DB oracle.
//
// `voided = false` is Java's, not an addition: the walk goes through
// SampleItemServiceImpl.getSampleItemsBySampleId, whose criteria is
// {sample.id, voided:false}. It was missing here until a voided row was seeded
// (order-search-e2e.sql) and the c1 oracle tightened to match — before that,
// nothing in the dataset was voided, so the omission was invisible.
func (d *PatientDAOImpl) CountSampleItemsForPatient(patientID string) (int64, error) {
	var n int64
	err := d.DB.Table("clinlims.sample_item si").
		Where("si.voided = false").
		Where("si.samp_id IN (?)",
			d.DB.Table("clinlims.sample_human").Select("samp_id").Where("patient_id = ?", patientID)).
		Count(&n).Error
	return n, err
}

// CountResultsForPatient mirrors countResultsForPatient: analyses for the
// patient's sample items, EXCLUDING the Canceled / SampleRejected / NotStarted
// statuses. Java resolves those three status ids through IStatusService at
// runtime; this port takes them as a parameter so the caller owns that mapping
// and the DAO stays free of business rules.
func (d *PatientDAOImpl) CountResultsForPatient(patientID string, excludedStatusIDs []string) (int64, error) {
	// Same `voided = false` as CountSampleItemsForPatient, for the same reason:
	// countResultsForPatient reaches its analyses through
	// getSampleItemsBySampleId, so an analysis hanging off a voided sample item
	// is never counted.
	q := d.DB.Table("clinlims.analysis a").
		Where("a.sampitem_id IN (?)",
			d.DB.Table("clinlims.sample_item si").Select("si.id").
				Where("si.voided = false").
				Where("si.samp_id IN (?)",
					d.DB.Table("clinlims.sample_human").Select("samp_id").Where("patient_id = ?", patientID)))
	if len(excludedStatusIDs) > 0 {
		q = q.Where("a.status_id NOT IN (?)", excludedStatusIDs)
	}
	var n int64
	err := q.Count(&n).Error
	return n, err
}
