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
