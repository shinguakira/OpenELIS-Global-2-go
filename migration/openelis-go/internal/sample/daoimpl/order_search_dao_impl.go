package daoimpl

import (
	"strings"
	"time"
)

// IDValuePair mirrors org.openelisglobal.common.util.IdValuePair, the shape
// every DisplayListService list is built from.
type IDValuePair struct {
	ID    string `gorm:"column:id" json:"id"`
	Value string `gorm:"column:value" json:"value"`
}

// PaymentOptions mirrors
// DisplayListService.getList(SAMPLE_PATIENT_PAYMENT_OPTIONS), which is
// createFromDictionaryCategoryLocalizedSort("patientPayment").
//
// Two filters that are easy to miss, both verified against the live response:
//   - is_active = 'Y' ONLY. The category holds seven rows; Java returns four,
//     because getDictionaryEntrysByCategoryAbbreviation hardcodes
//     `d.isActive = 'Y'`.
//   - sorted by the LOCALIZED name, not by id and not by dict_entry. The live
//     order is 1120, 1122, 1121, 1123 — alphabetical by value, which is not
//     id order.
//
// The value itself is the localized name with dict_entry as the fallback, the
// same resolution the b1 test-section read already uses.
func (d *SampleDAOImpl) PaymentOptions() ([]IDValuePair, error) {
	rows := []IDValuePair{}
	err := d.DB.Table("clinlims.dictionary AS d").
		Select("d.id::text AS id, COALESCE(lv.value, d.dict_entry) AS value").
		Joins("JOIN clinlims.dictionary_category AS c ON c.id = d.dictionary_category_id").
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = d.name_localization_id AND lv.locale = ?", d.Locale()).
		Where("c.name = ? AND d.is_active = ?", "patientPayment", "Y").
		Order("value").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// OrderSearchSampleRow is the sample-level data order/search needs.
type OrderSearchSampleRow struct {
	ID              int64      `gorm:"column:id"`
	AccessionNumber string     `gorm:"column:accession_number"`
	ReceivedDate    *time.Time `gorm:"column:received_date"`
	CollectionDate  *time.Time `gorm:"column:collection_date"`
	Status          *string    `gorm:"column:status"`
	OrderPriority   *string    `gorm:"column:order_priority"`
	StorageSkipped  *bool      `gorm:"column:storage_skipped"`
}

// OrderSearchSample loads the sample row behind GET rest/order/search.
func (d *SampleDAOImpl) OrderSearchSample(accessionNumber string) (*OrderSearchSampleRow, error) {
	rows := []OrderSearchSampleRow{}
	err := d.DB.Table("clinlims.sample").
		Select("id, accession_number, received_date, collection_date, status, order_priority, storage_skipped").
		Where("accession_number = ?", accessionNumber).
		Limit(1).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

// OrderSearchPatientRow carries the patient/person columns PatientInfoBean
// reads directly. The identity-backed fields come from PatientIdentities.
type OrderSearchPatientRow struct {
	PatientID          int64      `gorm:"column:patient_id"`
	NationalID         *string    `gorm:"column:national_id"`
	Gender             *string    `gorm:"column:gender"`
	EnteredBirthDate   *string    `gorm:"column:entered_birth_date"`
	IsMerged           *bool      `gorm:"column:is_merged"`
	PatientLastupdated *time.Time `gorm:"column:patient_lastupdated"`

	PersonID          int64      `gorm:"column:person_id"`
	FirstName         *string    `gorm:"column:first_name"`
	LastName          *string    `gorm:"column:last_name"`
	StreetAddress     *string    `gorm:"column:street_address"`
	City              *string    `gorm:"column:city"`
	PrimaryPhone      *string    `gorm:"column:primary_phone"`
	Email             *string    `gorm:"column:email"`
	PersonLastupdated *time.Time `gorm:"column:person_lastupdated"`
}

// PatientForSample loads the patient linked to a sample through sample_human.
// Returns (nil, nil) when the sample has no patient — Java then omits both
// patientProperties and orderData entirely.
func (d *SampleDAOImpl) PatientForSample(sampleID int64) (*OrderSearchPatientRow, error) {
	rows := []OrderSearchPatientRow{}
	err := d.DB.Table("clinlims.sample_human AS sh").
		Select(`pa.id AS patient_id, pa.national_id, pa.gender, pa.entered_birth_date,
			pa.is_merged, pa.lastupdated AS patient_lastupdated,
			pe.id AS person_id, pe.first_name, pe.last_name, pe.street_address, pe.city,
			pe.primary_phone, pe.email, pe.lastupdated AS person_lastupdated`).
		Joins("JOIN clinlims.patient AS pa ON pa.id = sh.patient_id").
		Joins("JOIN clinlims.person AS pe ON pe.id = pa.person_id").
		Where("sh.samp_id = ?", sampleID).
		Limit(1).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

// PatientIdentitiesByType returns the patient's identity values keyed by the
// TYPE NAME ("ST", "SUBJECT", "MOTHER", "AKA", ...), which is how
// PatientIdentityTypeMap.getIdentityValue looks them up.
func (d *SampleDAOImpl) PatientIdentitiesByType(patientID int64) (map[string]string, error) {
	type row struct {
		TypeName string `gorm:"column:type_name"`
		Value    string `gorm:"column:value"`
	}
	rows := []row{}
	err := d.DB.Table("clinlims.patient_identity AS pi").
		Select("t.identity_type AS type_name, COALESCE(pi.identity_data, '') AS value").
		Joins("JOIN clinlims.patient_identity_type AS t ON t.id = pi.identity_type_id").
		Where("pi.patient_id = ?", patientID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, r := range rows {
		out[r.TypeName] = r.Value
	}
	return out, nil
}

// AddressPartsByName returns the person's address values keyed by
// address_part.part_name ("village", "commune", "department", ...).
//
// Java resolves the three part IDs it cares about once
// (initAddressPartIds) and then calls getByPersonIdAndPartId per part; keying
// by name here is the same lookup with the indirection removed. A missing part
// yields "" in Java, so absence and empty are equivalent downstream.
func (d *SampleDAOImpl) AddressPartsByName(personID int64) (map[string]string, error) {
	type row struct {
		PartName string `gorm:"column:part_name"`
		Value    string `gorm:"column:value"`
	}
	rows := []row{}
	err := d.DB.Table("clinlims.person_address AS pa").
		Select("ap.part_name AS part_name, COALESCE(pa.value, '') AS value").
		Joins("JOIN clinlims.address_part AS ap ON ap.id = pa.address_part_id").
		Where("pa.person_id = ?", personID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, r := range rows {
		out[r.PartName] = r.Value
	}
	return out, nil
}

// OrderSearchSampleItemRow is one sample item plus everything the samples[]
// block needs, joined in one pass.
type OrderSearchSampleItemRow struct {
	ID             int64      `gorm:"column:id"`
	SortOrder      *string    `gorm:"column:sort_order"`
	TypeOfSampleID *int64     `gorm:"column:typeosamp_id"`
	SampleTypeName *string    `gorm:"column:sample_type_name"`
	CollectionDate *time.Time `gorm:"column:collection_date"`
	ReceivedDate   *time.Time `gorm:"column:received_date"`
	Quantity       *float64   `gorm:"column:quantity"`
	UomID          *int64     `gorm:"column:uom_id"`
	Collector      *string    `gorm:"column:collector"`

	CollectionConditions *string `gorm:"column:collection_conditions"`
	CollectionMethod     *string `gorm:"column:collection_method"`
	SampleTemperature    *string `gorm:"column:sample_temperature"`
	SpecimenOrigin       *string `gorm:"column:specimen_origin"`

	StorageLocationID         *int64  `gorm:"column:storage_location_id"`
	StorageLocationType       *string `gorm:"column:storage_location_type"`
	StoragePositionCoordinate *string `gorm:"column:storage_position_coordinate"`
	StorageNotes              *string `gorm:"column:storage_notes"`
}

// SampleItemsForSample mirrors sampleItemService.getSampleItemsBySampleId plus
// the per-item lookups searchOrder performs (type of sample, storage
// assignment).
func (d *SampleDAOImpl) SampleItemsForSample(sampleID int64) ([]OrderSearchSampleItemRow, error) {
	rows := []OrderSearchSampleItemRow{}
	err := d.DB.Table("clinlims.sample_item AS si").
		Select(`si.id, si.sort_order, si.typeosamp_id, tos.description AS sample_type_name,
			si.collection_date, si.received_date, si.quantity, si.uom_id, si.collector,
			si.collection_conditions, si.collection_method, si.sample_temperature, si.specimen_origin,
			ssa.location_id AS storage_location_id, ssa.location_type AS storage_location_type,
			ssa.position_coordinate AS storage_position_coordinate, ssa.notes AS storage_notes`).
		Joins("LEFT JOIN clinlims.type_of_sample AS tos ON tos.id = si.typeosamp_id").
		Joins("LEFT JOIN clinlims.sample_storage_assignment AS ssa ON ssa.sample_item_id = si.id").
		// voided = false is Java's: SampleItemServiceImpl.getSampleItemsBySampleId
		// builds criteria {sample.id, voided:false} — an EXACT match, so a NULL
		// voided is excluded too.
		//
		// ORDER BY ctid — the physical order, which is what Java's unordered
		// query actually returns.
		//
		// SampleItemDAOImpl has an HQL getSampleItemsBySampleId that ends
		// `order by sampleItem.sortOrder`, but the SERVICE method of the same
		// name does NOT call it: it builds a criteria map and calls
		// getAllMatching, which has no ordering at all. The controller calls the
		// service, so sortOrder never enters into it and Postgres decides —
		// measured as seqscan order (E2E001 returns item 10002 first, ctid (0,2),
		// ahead of 10001 at ctid (0,36); by id or sortOrder it would be second).
		//
		// This query carries two LEFT JOINs that Java resolves lazily per row, so
		// without an explicit order the planner is free to emit a different
		// sequence — and did: E2E-EDIT-01 came back reversed. samples[] order is
		// observable in the response, so that is a real divergence, not a detail.
		Where("si.samp_id = ? AND si.voided = false", sampleID).
		Order("si.ctid").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// SampleItemTestRow is one entry of a sample item's tests[] / panels[].
type SampleItemTestRow struct {
	SampleItemID    int64   `gorm:"column:sample_item_id"`
	TestID          *int64  `gorm:"column:test_id"`
	TestName        *string `gorm:"column:test_name"`
	TestDescription *string `gorm:"column:test_description"`
	PanelID         *int64  `gorm:"column:panel_id"`
	PanelName       *string `gorm:"column:panel_name"`
}

// AnalysesForSampleItems mirrors analysisService.getAnalysesBySampleItem run
// per item, with the test and panel joined in.
//
// Both names are the LOCALIZED ones (test.localizedName / panel.localizedName),
// resolved through localization_value with the raw name as the fallback — the
// same pattern the b1 test-section read uses. `description` is the raw column.
func (d *SampleDAOImpl) AnalysesForSampleItems(sampleID int64) ([]SampleItemTestRow, error) {
	rows := []SampleItemTestRow{}
	err := d.DB.Table("clinlims.analysis AS a").
		Select(`a.sampitem_id AS sample_item_id,
			t.id AS test_id, COALESCE(tlv.value, t.name) AS test_name, t.description AS test_description,
			p.id AS panel_id, COALESCE(plv.value, p.name) AS panel_name`).
		Joins("JOIN clinlims.sample_item AS si ON si.id = a.sampitem_id").
		Joins("LEFT JOIN clinlims.test AS t ON t.id = a.test_id").
		Joins("LEFT JOIN clinlims.localization AS tl ON tl.id = t.name_localization_id").
		Joins("LEFT JOIN clinlims.localization_value AS tlv ON tlv.localization_id = tl.id AND tlv.locale = ?", d.Locale()).
		Joins("LEFT JOIN clinlims.panel AS p ON p.id = a.panel_id").
		Joins("LEFT JOIN clinlims.localization AS pl ON pl.id = p.name_localization_id").
		Joins("LEFT JOIN clinlims.localization_value AS plv ON plv.localization_id = pl.id AND plv.locale = ?", d.Locale()).
		Where("si.samp_id = ?", sampleID).
		// AnalysisDAOImpl.getAnalysesBySampleItem is
		// `from Analysis a where a.sampleItem.id = :id` with NO ordering, run once
		// per item, so Java's per-item order is physical. `a.id` was an ordering
		// this port invented; ctid reproduces the scan order instead, the same way
		// the sample-item query does.
		//
		// No sample item in this database has analyses whose ctid order differs
		// from their id order, so no test can currently tell the two apart — the
		// change is for faithfulness, not to fix an observed diff.
		Order("a.ctid").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// StoragePathRow carries the names needed to render a hierarchical path for
// one assignment, walked to whatever depth the location type implies.
type StoragePathRow struct {
	RoomName   *string `gorm:"column:room_name"`
	DeviceName *string `gorm:"column:device_name"`
	ShelfLabel *string `gorm:"column:shelf_label"`
	RackLabel  *string `gorm:"column:rack_label"`
	BoxLabel   *string `gorm:"column:box_label"`
}

// StoragePathFor resolves the ancestry of one storage location.
//
// Mirrors SampleStorageServiceImpl.buildHierarchicalPathForAssignment, which
// switches on location_type and walks UPWARD:
//
//	room   -> room
//	device -> room > device
//	shelf  -> room > device > shelf
//	rack   -> room > device > shelf > rack
//	box    -> room > device > shelf > rack > box
//
// Each branch requires its FULL ancestry to be present: Java guards with
// `if (room != null && device != null && shelf != null)` and leaves the path
// null when any link is missing, so a rack whose shelf lost its device yields
// no path rather than a partial one. Reproduced by returning a row whose
// pointers are nil and letting the caller apply the same rule.
//
// The position coordinate is appended by the caller, not here — Java appends it
// after the switch, and only when non-blank.
func (d *SampleDAOImpl) StoragePathFor(locationType string, locationID int64) (*StoragePathRow, error) {
	rows := []StoragePathRow{}
	var err error
	base := d.DB.Table("clinlims.storage_room AS ro")

	switch locationType {
	case "room":
		err = base.Select("ro.name AS room_name").Where("ro.id = ?", locationID).Limit(1).Scan(&rows).Error
	case "device":
		err = d.DB.Table("clinlims.storage_device AS de").
			Select("ro.name AS room_name, de.name AS device_name").
			Joins("LEFT JOIN clinlims.storage_room AS ro ON ro.id = de.parent_room_id").
			Where("de.id = ?", locationID).Limit(1).Scan(&rows).Error
	case "shelf":
		err = d.DB.Table("clinlims.storage_shelf AS sh").
			Select("ro.name AS room_name, de.name AS device_name, sh.label AS shelf_label").
			Joins("LEFT JOIN clinlims.storage_device AS de ON de.id = sh.parent_device_id").
			Joins("LEFT JOIN clinlims.storage_room AS ro ON ro.id = de.parent_room_id").
			Where("sh.id = ?", locationID).Limit(1).Scan(&rows).Error
	case "rack":
		err = d.DB.Table("clinlims.storage_rack AS ra").
			Select("ro.name AS room_name, de.name AS device_name, sh.label AS shelf_label, ra.label AS rack_label").
			Joins("LEFT JOIN clinlims.storage_shelf AS sh ON sh.id = ra.parent_shelf_id").
			Joins("LEFT JOIN clinlims.storage_device AS de ON de.id = sh.parent_device_id").
			Joins("LEFT JOIN clinlims.storage_room AS ro ON ro.id = de.parent_room_id").
			Where("ra.id = ?", locationID).Limit(1).Scan(&rows).Error
	case "box":
		err = d.DB.Table("clinlims.storage_box AS bo").
			Select(`ro.name AS room_name, de.name AS device_name, sh.label AS shelf_label,
				ra.label AS rack_label, bo.label AS box_label`).
			Joins("LEFT JOIN clinlims.storage_rack AS ra ON ra.id = bo.parent_rack_id").
			Joins("LEFT JOIN clinlims.storage_shelf AS sh ON sh.id = ra.parent_shelf_id").
			Joins("LEFT JOIN clinlims.storage_device AS de ON de.id = sh.parent_device_id").
			Joins("LEFT JOIN clinlims.storage_room AS ro ON ro.id = de.parent_room_id").
			Where("bo.id = ?", locationID).Limit(1).Scan(&rows).Error
	default:
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

// QAAllRequiredVerified mirrors
// SampleQaChecklistServiceImpl.areAllItemsVerified: false when no checklist row
// exists, otherwise the stored all_required_verified flag.
func (d *SampleDAOImpl) QAAllRequiredVerified(sampleID int64) (bool, error) {
	vals := []bool{}
	err := d.DB.Table("clinlims.sample_qa_checklist").
		Where("sample_id = ?", sampleID).
		Limit(1).
		Pluck("COALESCE(all_required_verified, false)", &vals).Error
	if err != nil {
		return false, err
	}
	if len(vals) == 0 {
		return false, nil
	}
	return vals[0], nil
}

// SampleOrderExtrasRow carries the conditional halves of buildSampleOrderItems
// that come from tables rather than from the sample row.
//
// Every field is a pointer: Java puts each value only when its source row
// exists, and Include.NON_NULL then drops the key. A zero value would emit the
// key with "" and diverge.
type SampleOrderExtrasRow struct {
	ProviderPersonID  *string `gorm:"column:provider_person_id"`
	ProviderFirstName *string `gorm:"column:provider_first_name"`
	ProviderLastName  *string `gorm:"column:provider_last_name"`
	ProviderWorkPhone *string `gorm:"column:provider_work_phone"`
	ProviderEmail     *string `gorm:"column:provider_email"`
	ProviderFax       *string `gorm:"column:provider_fax"`

	ReferringSiteID   *string `gorm:"column:site_id"`
	ReferringSiteName *string `gorm:"column:site_name"`
	ReferringSiteCode *string `gorm:"column:site_code"`

	DepartmentID   *string `gorm:"column:dept_id"`
	DepartmentName *string `gorm:"column:dept_name"`
}

// SampleOrderExtras loads the provider and the referring site/department for a
// sample in one query.
//
// The program is NOT here: its resolution has three branches that depend on the
// observation value, so it lives in ProgramIDForSampleByName /
// ProgramSampleByAccession instead.
//
// The site/department split is by ORGANISATION TYPE — RequesterService calls
// getOrganizationRequester twice, with the type named "referring clinic" and
// the one named "dept". sample_requester.requester_type_id only separates
// organization (1) from provider (2) and does NOT decide which is which.
//
// Types are matched by NAME because TableIdService looks them up by name; their
// ids (5 and 11 here) are deployment data.
func (d *SampleDAOImpl) SampleOrderExtras(sampleID int64) (*SampleOrderExtrasRow, error) {
	rows := []SampleOrderExtrasRow{}
	err := d.DB.Table("clinlims.sample AS s").
		Select(`pe.id::text            AS provider_person_id,
			pe.first_name          AS provider_first_name,
			pe.last_name           AS provider_last_name,
			pe.work_phone          AS provider_work_phone,
			pe.email               AS provider_email,
			pe.fax                 AS provider_fax,
			site.id::text          AS site_id,
			site.name              AS site_name,
			site.short_name        AS site_code,
			dept.id::text          AS dept_id,
			dept.name              AS dept_name`).
		Joins("LEFT JOIN clinlims.sample_human AS sh ON sh.samp_id = s.id").
		Joins("LEFT JOIN clinlims.provider AS pr ON pr.id = sh.provider_id").
		Joins("LEFT JOIN clinlims.person AS pe ON pe.id = pr.person_id").
		Joins(`LEFT JOIN clinlims.organization AS site ON site.id = (
			SELECT sr.requester_id FROM clinlims.sample_requester sr
			  JOIN clinlims.organization_organization_type oot ON oot.org_id = sr.requester_id
			  JOIN clinlims.organization_type ot ON ot.id = oot.org_type_id
			 WHERE sr.sample_id = s.id AND ot.short_name = 'referring clinic'
			 LIMIT 1)`).
		Joins(`LEFT JOIN clinlims.organization AS dept ON dept.id = (
			SELECT sr.requester_id FROM clinlims.sample_requester sr
			  JOIN clinlims.organization_organization_type oot ON oot.org_id = sr.requester_id
			  JOIN clinlims.organization_type ot ON ot.id = oot.org_type_id
			 WHERE sr.sample_id = s.id AND ot.short_name = 'dept'
			 LIMIT 1)`).
		Where("s.id = ?", sampleID).
		Limit(1).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

// ObservationValues returns the raw observation-history values for a sample,
// keyed by observation type name — the map behind
// observationHistoryService.getRawValueForSample.
//
// A type NAME absent from observation_history_type simply yields no entry, so a
// deployment missing one (this one has no testLocationCode row) omits the key
// exactly as Java does.
func (d *SampleDAOImpl) ObservationValues(sampleID int64) (map[string]string, error) {
	type row struct {
		TypeName string `gorm:"column:type_name"`
		Value    string `gorm:"column:value"`
	}
	rows := []row{}
	err := d.DB.Table("clinlims.observation_history AS oh").
		Select("oht.type_name AS type_name, oh.value AS value").
		Joins("JOIN clinlims.observation_history_type AS oht ON oht.id = oh.observation_history_type_id").
		Where("oh.sample_id = ? AND oh.value IS NOT NULL", sampleID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(rows))
	for _, r := range rows {
		out[r.TypeName] = r.Value
	}
	return out, nil
}

// programSubclassTable reproduces the entity-class selection in
// ProgramSampleDAOImpl.getProgrammeSampleBySample.
//
// Java picks a CLASS from the program NAME, and ProgramSample is
// @Inheritance(TABLE_PER_CLASS) — so each subclass lives in its own table and a
// row in program_sample is INVISIBLE to a query for a subclass. The name
// therefore decides which table is searched, and a mismatch silently yields no
// row rather than an error.
//
// The three tests run in Java's order: "pathology" first, then
// "immunohistochemistry", then "cytology". Only a name containing two of them
// could tell the order apart, but it is reproduced rather than tidied.
func programSubclassTable(programName string) string {
	n := strings.ToLower(programName)
	switch {
	case strings.Contains(n, "pathology"):
		return "clinlims.pathology_sample"
	case strings.Contains(n, "immunohistochemistry"):
		return "clinlims.immunohistochemistry_sample"
	case strings.Contains(n, "cytology"):
		return "clinlims.cytology_sample"
	}
	return "clinlims.program_sample"
}

// ProgramIDForSampleByName resolves programId the way Java does when the sample
// HAS a program observation: look in the table the NAME selects, and fall back
// to matching the name against the program list when that table has no row.
//
// Returns "" when neither path resolves, which is Java putting no programId key.
func (d *SampleDAOImpl) ProgramIDForSampleByName(sampleID int64, programName string) (string, error) {
	// No ORDER BY: Hibernate issues `... where sample_id = ? limit 1` with no
	// ordering either, so both take whatever the table hands back first.
	ids := []string{}
	err := d.DB.Table(programSubclassTable(programName)).
		Select("program_id::text").
		Where("sample_id = ? AND program_id IS NOT NULL", sampleID).
		Limit(1).
		Scan(&ids).Error
	if err != nil {
		return "", err
	}
	if len(ids) > 0 && ids[0] != "" {
		return ids[0], nil
	}

	// Fallback: `for (Program p : programService.getAll())` breaking on the
	// first name match. Program names are unique here, so the scan order of
	// getAll() cannot change the answer; ctid keeps it deterministic anyway.
	err = d.DB.Table("clinlims.program").
		Select("id::text").
		Where("name = ?", programName).
		Order("ctid").
		Limit(1).
		Scan(&ids).Error
	if err != nil {
		return "", err
	}
	if len(ids) > 0 {
		return ids[0], nil
	}
	return "", nil
}

// ProgramSampleByAccession is the ELSE branch, taken when the sample has no
// program observation at all. Java calls
// getProgramSamplesByAccessionNumberOrProgramName(accessionNumber) and reads
// element 0, taking BOTH keys from it — so `program` here is the program's own
// NAME, not an observation value.
//
// The HQL is `from ProgramSample ps join ps.program p join ps.sample s where
// lower(p.programName) like :f or lower(s.accessionNumber) like :f order by
// ps.id`, and under TABLE_PER_CLASS Hibernate expands `from ProgramSample` to a
// UNION over the base table and all three subclass tables — hence the union
// below. The filter is `%<accession, lower-cased>%`, matched against the
// program name as well, so an accession that happens to be a substring of a
// program name can match a different sample's row entirely.
func (d *SampleDAOImpl) ProgramSampleByAccession(accessionNumber string) (programID string, programName string, err error) {
	type row struct {
		ProgramID   string `gorm:"column:program_id"`
		ProgramName string `gorm:"column:program_name"`
	}
	rows := []row{}
	filter := "%" + strings.ToLower(accessionNumber) + "%"
	sql := `
		SELECT ps.id AS ps_id, p.id::text AS program_id, p.name AS program_name
		  FROM (
		        SELECT id, program_id, sample_id FROM clinlims.program_sample
		         UNION ALL
		        SELECT id, program_id, sample_id FROM clinlims.pathology_sample
		         UNION ALL
		        SELECT id, program_id, sample_id FROM clinlims.immunohistochemistry_sample
		         UNION ALL
		        SELECT id, program_id, sample_id FROM clinlims.cytology_sample
		       ) AS ps
		  JOIN clinlims.program p ON p.id = ps.program_id
		  JOIN clinlims.sample s ON s.id = ps.sample_id
		 WHERE lower(p.name) LIKE ? OR lower(s.accession_number) LIKE ?
		 ORDER BY ps.id
		 LIMIT 1`
	if err := d.DB.Raw(sql, filter, filter).Scan(&rows).Error; err != nil {
		return "", "", err
	}
	if len(rows) == 0 {
		return "", "", nil
	}
	return rows[0].ProgramID, rows[0].ProgramName, nil
}
