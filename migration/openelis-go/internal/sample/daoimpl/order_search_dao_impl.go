package daoimpl

import "time"

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
		Joins("LEFT JOIN clinlims.localization_value AS lv ON lv.localization_id = d.name_localization_id AND lv.locale = ?", "en").
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
// assignment). Ordered by id, matching the DAO's natural order.
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
		// NO ORDER BY, deliberately. That criteria call goes to
		// getAllMatchingOrdered with an EMPTY order list, so Java's order is
		// DB-natural — and the samples[] array order is observable, so imposing
		// an ORDER BY here would diverge (measured: Java returns sample item
		// 10002 before 10001, which is neither id nor sort_order order).
		Where("si.samp_id = ? AND si.voided = false", sampleID).
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
		Joins("LEFT JOIN clinlims.localization_value AS tlv ON tlv.localization_id = tl.id AND tlv.locale = ?", "en").
		Joins("LEFT JOIN clinlims.panel AS p ON p.id = a.panel_id").
		Joins("LEFT JOIN clinlims.localization AS pl ON pl.id = p.name_localization_id").
		Joins("LEFT JOIN clinlims.localization_value AS plv ON plv.localization_id = pl.id AND plv.locale = ?", "en").
		Where("si.samp_id = ?", sampleID).
		Order("a.id").
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
