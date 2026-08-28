package daoimpl

// Loader for the Hibernate object graph rest/accession-results nests under each
// row's `result` key. One query per entity kind, keyed by the analyses on the
// order, because the shape is a tree over a handful of reference tables rather
// than a join that could be flattened.

// EntityGraphRow is every column the nested graph renders, for one analysis.
//
// The three "*_millis" values are epoch MILLISECONDS: Jackson writes a
// java.sql.Timestamp as a number, while the same instant also appears beside it
// pre-formatted as a string. Both go over the wire.
type EntityGraphRow struct {
	AnalysisID          string  `gorm:"column:analysis_id"`
	AnalysisLastupdated *int64  `gorm:"column:analysis_lastupdated"`
	AnalysisType        *string `gorm:"column:analysis_type"`
	AnalysisRevision    *string `gorm:"column:analysis_revision"`
	AnalysisEnteredDate *int64  `gorm:"column:analysis_entered_date"`
	AnalysisReportable  *string `gorm:"column:analysis_reportable"`
	AnalysisStatusID    *string `gorm:"column:analysis_status_id"`
	AnalysisReferredOut *bool   `gorm:"column:analysis_referred_out"`

	ResultID          *string  `gorm:"column:result_id"`
	ResultLastupdated *int64   `gorm:"column:result_lastupdated"`
	ResultSortOrder   *string  `gorm:"column:result_sort_order"`
	ResultReportable  *string  `gorm:"column:result_reportable"`
	ResultType        *string  `gorm:"column:result_type"`
	ResultValue       *string  `gorm:"column:result_value"`
	ResultMinNormal   *float64 `gorm:"column:result_min_normal"`
	ResultMaxNormal   *float64 `gorm:"column:result_max_normal"`
	ResultSigDigits   *int     `gorm:"column:result_sig_digits"`
	ResultGrouping    *int     `gorm:"column:result_grouping"`

	ItemID             string  `gorm:"column:item_id"`
	ItemLastupdated    *int64  `gorm:"column:item_lastupdated"`
	ItemSortOrder      *string `gorm:"column:item_sort_order"`
	ItemTypeOfSampleID *string `gorm:"column:item_type_of_sample_id"`
	ItemCollectionDate *int64  `gorm:"column:item_collection_date"`
	ItemStatusID       *string `gorm:"column:item_status_id"`
	ItemRejected       *bool   `gorm:"column:item_rejected"`
	ItemVoided         *bool   `gorm:"column:item_voided"`

	SampleID                string  `gorm:"column:sample_id"`
	SampleLastupdated       *int64  `gorm:"column:sample_lastupdated"`
	SampleAccessionNumber   string  `gorm:"column:sample_accession_number"`
	SampleEnteredDate       *int64  `gorm:"column:sample_entered_date"`
	SampleEnteredDisplay    *string `gorm:"column:sample_entered_display"`
	SampleReceivedMillis    *int64  `gorm:"column:sample_received_millis"`
	SampleReceivedDisplay   *string `gorm:"column:sample_received_display"`
	SampleReceivedTime      *string `gorm:"column:sample_received_time"`
	SampleCollectionMillis  *int64  `gorm:"column:sample_collection_millis"`
	SampleCollectionDisplay *string `gorm:"column:sample_collection_display"`
	SampleCollectionTime    *string `gorm:"column:sample_collection_time"`
	SampleIsConfirmation    *bool   `gorm:"column:sample_is_confirmation"`
	SamplePriority          *string `gorm:"column:sample_priority"`
	SampleStorageSkipped    *bool   `gorm:"column:sample_storage_skipped"`
	SampleStatusID          *string `gorm:"column:sample_status_id"`

	TosID          *string `gorm:"column:tos_id"`
	TosLastupdated *int64  `gorm:"column:tos_lastupdated"`
	TosDescription *string `gorm:"column:tos_description"`
	TosDomain      *string `gorm:"column:tos_domain"`
	TosLocalAbbrev *string `gorm:"column:tos_local_abbrev"`
	TosIsActive    *bool   `gorm:"column:tos_is_active"`
	TosSortOrder   *int    `gorm:"column:tos_sort_order"`
	TosLocalizID   *string `gorm:"column:tos_localiz_id"`

	SectionID          *string `gorm:"column:section_id"`
	SectionLastupdated *int64  `gorm:"column:section_lastupdated"`
	SectionIsActive    *string `gorm:"column:section_is_active"`
	SectionIsExternal  *string `gorm:"column:section_is_external"`
	SectionName        *string `gorm:"column:section_name"`
	SectionDescription *string `gorm:"column:section_description"`
	SectionSortOrder   *int    `gorm:"column:section_sort_order"`
	SectionLocalizID   *string `gorm:"column:section_localiz_id"`

	TestID                    string  `gorm:"column:test_entity_id"`
	TestLastupdated           *int64  `gorm:"column:test_lastupdated"`
	TestName                  *string `gorm:"column:test_entity_name"`
	TestSortOrder             *string `gorm:"column:test_entity_sort_order"`
	TestIsActive              *string `gorm:"column:test_is_active"`
	TestDescription           *string `gorm:"column:test_description"`
	TestNormalizedDescription *string `gorm:"column:test_normalized_description"`
	TestAugmentedName         *string `gorm:"column:test_augmented_name"`
	TestDomain                *string `gorm:"column:test_domain"`
	TestIsReportable          *string `gorm:"column:test_is_reportable"`
	TestLocalCode             *string `gorm:"column:test_local_code"`
	TestOrderable             *bool   `gorm:"column:test_orderable"`
	TestGUID                  *string `gorm:"column:test_guid"`
	TestInLabOnly             *bool   `gorm:"column:test_in_lab_only"`
	TestNotifyResults         *bool   `gorm:"column:test_notify_results"`
	TestAMR                   *bool   `gorm:"column:test_amr"`
	TestNameLocalizID         *string `gorm:"column:test_name_localiz_id"`
	TestRptLocalizID          *string `gorm:"column:test_rpt_localiz_id"`

	UomID          *string `gorm:"column:uom_id"`
	UomLastupdated *int64  `gorm:"column:uom_lastupdated"`
	UomName        *string `gorm:"column:uom_name"`
	UomDescription *string `gorm:"column:uom_description"`
	UomIsActive    *string `gorm:"column:uom_is_active"`

	PanelID          *string `gorm:"column:panel_id"`
	PanelLastupdated *int64  `gorm:"column:panel_lastupdated"`
	PanelIsActive    *string `gorm:"column:panel_is_active"`
	PanelName        *string `gorm:"column:panel_name"`
	PanelDescription *string `gorm:"column:panel_description"`
	PanelSortOrder   *int    `gorm:"column:panel_sort_order"`
	PanelLocalizID   *string `gorm:"column:panel_localiz_id"`
}

// LocalizationRow is one localization plus its per-locale values.
type LocalizationRow struct {
	ID          string `gorm:"column:id"`
	Lastupdated *int64 `gorm:"column:lastupdated"`
	Description string `gorm:"column:description"`
}

// LocalizationValueRow is one localization_value row.
type LocalizationValueRow struct {
	LocalizationID string `gorm:"column:localization_id"`
	ID             string `gorm:"column:id"`
	Lastupdated    *int64 `gorm:"column:lastupdated"`
	Locale         string `gorm:"column:locale"`
	Value          string `gorm:"column:value"`
}

const graphSelect = `a.id::text AS analysis_id,
	trunc(EXTRACT(EPOCH FROM a.lastupdated) * 1000)::bigint AS analysis_lastupdated,
	a.analysis_type AS analysis_type,
	a.revision AS analysis_revision,
	trunc(EXTRACT(EPOCH FROM a.entry_date) * 1000)::bigint AS analysis_entered_date,
	a.is_reportable AS analysis_reportable,
	a.status_id::text AS analysis_status_id,
	a.referred_out AS analysis_referred_out,

	res.id::text AS result_id,
	trunc(EXTRACT(EPOCH FROM res.lastupdated) * 1000)::bigint AS result_lastupdated,
	res.sort_order::text AS result_sort_order,
	res.is_reportable AS result_reportable,
	res.result_type AS result_type,
	res.value AS result_value,
	res.min_normal AS result_min_normal,
	res.max_normal AS result_max_normal,
	res.significant_digits AS result_sig_digits,
	res.grouping AS result_grouping,

	si.id::text AS item_id,
	trunc(EXTRACT(EPOCH FROM si.lastupdated) * 1000)::bigint AS item_lastupdated,
	si.sort_order::text AS item_sort_order,
	si.typeosamp_id::text AS item_type_of_sample_id,
	trunc(EXTRACT(EPOCH FROM si.collection_date) * 1000)::bigint AS item_collection_date,
	si.status_id::text AS item_status_id,
	si.voided AS item_voided,
	si.rejected AS item_rejected,

	s.id::text AS sample_id,
	trunc(EXTRACT(EPOCH FROM s.lastupdated) * 1000)::bigint AS sample_lastupdated,
	s.accession_number AS sample_accession_number,
	trunc(EXTRACT(EPOCH FROM date_trunc('day', s.entered_date)) * 1000)::bigint AS sample_entered_date,
	to_char(s.entered_date, 'DD/MM/YYYY') AS sample_entered_display,
	trunc(EXTRACT(EPOCH FROM s.received_date) * 1000)::bigint AS sample_received_millis,
	to_char(s.received_date, 'DD/MM/YYYY') AS sample_received_display,
	to_char(s.received_date, 'HH24:MI') AS sample_received_time,
	trunc(EXTRACT(EPOCH FROM s.collection_date) * 1000)::bigint AS sample_collection_millis,
	to_char(s.collection_date, 'DD/MM/YYYY') AS sample_collection_display,
	to_char(s.collection_date, 'HH24:MI') AS sample_collection_time,
	s.is_confirmation AS sample_is_confirmation,
	s.order_priority AS sample_priority,
	s.storage_skipped AS sample_storage_skipped,
	s.status_id::text AS sample_status_id,

	tos.id::text AS tos_id,
	trunc(EXTRACT(EPOCH FROM tos.lastupdated) * 1000)::bigint AS tos_lastupdated,
	tos.description AS tos_description,
	tos.domain AS tos_domain,
	tos.local_abbrev AS tos_local_abbrev,
	tos.is_active AS tos_is_active,
	tos.sort_order AS tos_sort_order,
	tos.name_localization_id::text AS tos_localiz_id,

	ts.id::text AS section_id,
	trunc(EXTRACT(EPOCH FROM ts.lastupdated) * 1000)::bigint AS section_lastupdated,
	ts.is_active AS section_is_active,
	ts.is_external AS section_is_external,
	ts.name AS section_name,
	ts.description AS section_description,
	ts.sort_order AS section_sort_order,
	ts.name_localization_id::text AS section_localiz_id,

	t.id::text AS test_entity_id,
	trunc(EXTRACT(EPOCH FROM t.lastupdated) * 1000)::bigint AS test_lastupdated,
	t.name AS test_entity_name,
	t.sort_order AS test_entity_sort_order,
	t.is_active AS test_is_active,
	t.description AS test_description,
	t.normalized_description AS test_normalized_description,
	COALESCE(tnlv.value, t.name) || COALESCE(
		(SELECT '(' || COALESCE(atlv.value, atos.description) || ')'
		   FROM clinlims.sampletype_test AS atost
		   JOIN clinlims.type_of_sample AS atos ON atos.id = atost.sample_type_id
		   LEFT JOIN clinlims.localization AS atl ON atl.id = atos.name_localization_id
		   LEFT JOIN clinlims.localization_value AS atlv
		          ON atlv.localization_id = atl.id AND atlv.locale = 'en'
		  WHERE atost.test_id = t.id AND atos.local_abbrev IS DISTINCT FROM 'Variable'
		  ORDER BY atost.ctid LIMIT 1), '') AS test_augmented_name,
	t.domain AS test_domain,
	t.is_reportable AS test_is_reportable,
	t.local_code AS test_local_code,
	t.orderable AS test_orderable,
	t.guid::text AS test_guid,
	t.in_lab_only AS test_in_lab_only,
	t.notify_results AS test_notify_results,
	t.antimicrobial_resistance AS test_amr,
	t.name_localization_id::text AS test_name_localiz_id,
	t.reporting_name_localization_id::text AS test_rpt_localiz_id,

	uom.id::text AS uom_id,
	trunc(EXTRACT(EPOCH FROM uom.lastupdated) * 1000)::bigint AS uom_lastupdated,
	uom.name AS uom_name,
	uom.description AS uom_description,
	uom.is_active AS uom_is_active,

	p.id::text AS panel_id,
	trunc(EXTRACT(EPOCH FROM p.lastupdated) * 1000)::bigint AS panel_lastupdated,
	p.is_active AS panel_is_active,
	p.name AS panel_name,
	p.description AS panel_description,
	p.sort_order AS panel_sort_order,
	p.name_localization_id::text AS panel_localiz_id`

// EntityGraphForAccession loads the graph for every analysis on one order.
func (d *ResultDAOImpl) EntityGraphForAccession(accessionNumber string) ([]EntityGraphRow, error) {
	rows := []EntityGraphRow{}
	err := d.DB.Table("clinlims.analysis AS a").
		Select(graphSelect).
		Joins("JOIN clinlims.sample_item AS si ON si.id = a.sampitem_id").
		Joins("JOIN clinlims.sample AS s ON s.id = si.samp_id").
		Joins("LEFT JOIN clinlims.type_of_sample AS tos ON tos.id = si.typeosamp_id").
		Joins("JOIN clinlims.test AS t ON t.id = a.test_id").
		Joins("LEFT JOIN clinlims.test_section AS ts ON ts.id = t.test_section_id").
		Joins("LEFT JOIN clinlims.unit_of_measure AS uom ON uom.id = t.uom_id").
		Joins("LEFT JOIN clinlims.localization AS tnl ON tnl.id = t.name_localization_id").
		Joins("LEFT JOIN clinlims.localization_value AS tnlv ON tnlv.localization_id = tnl.id AND tnlv.locale = ?", d.Locale()).
		Joins("LEFT JOIN clinlims.panel AS p ON p.id = a.panel_id").
		Joins("LEFT JOIN clinlims.result AS res ON res.analysis_id = a.id").
		Where("s.accession_number = ?", accessionNumber).
		Scan(&rows).Error
	return rows, err
}

// Localizations loads every localization row and its per-locale values, keyed
// by id, so the graph mapper can attach them without another round trip per
// entity.
func (d *ResultDAOImpl) Localizations() (map[string]LocalizationRow, map[string][]LocalizationValueRow, error) {
	locs := []LocalizationRow{}
	err := d.DB.Table("clinlims.localization").
		Select("id::text AS id, trunc(EXTRACT(EPOCH FROM lastupdated) * 1000)::bigint AS lastupdated, description").
		Scan(&locs).Error
	if err != nil {
		return nil, nil, err
	}
	vals := []LocalizationValueRow{}
	err = d.DB.Table("clinlims.localization_value").
		Select(`localization_id::text AS localization_id, id::text AS id,
			trunc(EXTRACT(EPOCH FROM last_updated) * 1000)::bigint AS lastupdated, locale, value`).
		Order("localization_id, locale").
		Scan(&vals).Error
	if err != nil {
		return nil, nil, err
	}
	byID := make(map[string]LocalizationRow, len(locs))
	for _, l := range locs {
		byID[l.ID] = l
	}
	byLoc := map[string][]LocalizationValueRow{}
	for _, v := range vals {
		byLoc[v.LocalizationID] = append(byLoc[v.LocalizationID], v)
	}
	return byID, byLoc, nil
}
