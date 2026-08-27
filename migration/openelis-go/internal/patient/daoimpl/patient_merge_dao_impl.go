// Package daoimpl — the count/identity lookups behind
// GET rest/patient/merge/details/{patientId}.
package daoimpl

// PatientIdentityRow is a scan target for a patient_identity row joined to its
// type name. Java resolves the type via patientIdentityTypeService.get(); this
// port joins instead, which avoids that call's ObjectNotFoundException-on-miss
// behavior (a real Java 500 path) while producing the same data for every row
// whose type actually exists.
type PatientIdentityRow struct {
	IdentityTypeName string `gorm:"column:identity_type"`
	IdentityData     string `gorm:"column:identity_data"`
}

// GetPatientIdentities mirrors PatientMergeServiceImpl.getPatientIdentities:
// SELECT * FROM patient_identity WHERE patient_id = :patientId
// joined here to patient_identity_type so the display name is available in one
// query. Returns ALL rows unfiltered — the GUID/AKA/MOTHER/MOTHERS_INITIAL
// exclusion is a DISPLAY concern and is applied in the service, because Java
// counts the unfiltered list for totalIdentifiers while listing the filtered
// one. Filtering here would make that documented mismatch impossible to
// reproduce.
func (d *PatientDAOImpl) GetPatientIdentities(patientID string) ([]PatientIdentityRow, error) {
	rows := []PatientIdentityRow{}
	err := d.DB.Table("clinlims.patient_identity AS pi").
		Select("pit.identity_type, pi.identity_data").
		Joins("JOIN clinlims.patient_identity_type pit ON pit.id = pi.identity_type_id").
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
func (d *PatientDAOImpl) CountSampleItemsForPatient(patientID string) (int64, error) {
	var n int64
	err := d.DB.Table("clinlims.sample_item si").
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
	q := d.DB.Table("clinlims.analysis a").
		Where("a.sampitem_id IN (?)",
			d.DB.Table("clinlims.sample_item si").Select("si.id").
				Where("si.samp_id IN (?)",
					d.DB.Table("clinlims.sample_human").Select("samp_id").Where("patient_id = ?", patientID)))
	if len(excludedStatusIDs) > 0 {
		q = q.Where("a.status_id NOT IN (?)", excludedStatusIDs)
	}
	var n int64
	err := q.Count(&n).Error
	return n, err
}
