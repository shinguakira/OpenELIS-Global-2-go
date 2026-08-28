package service

import (
	"strings"

	"openelis-go/internal/result/daoimpl"
	"openelis-go/internal/result/form"
)

// Mapper for the Hibernate object graph rest/accession-results nests.

// localeDisplayNames renders the two locales the way
// Localization.getLocalesAndValuesOfLocalesWithValues does.
var localeDisplayNames = map[string]string{"en": "English", "fr": "French"}

// buildLocalization assembles the 13-key Localization object from its row and
// its per-locale values.
//
// Ten of those keys are derived from the same two values. localesWithValue,
// allActiveLocales, localesWithValueSortedForDisplay and localesSortedForDisplay
// are four separate getters that happen to return the same ["en","fr"] here —
// emitted separately because Jackson serialises each one.
func buildLocalization(
	id *string,
	locs map[string]daoimpl.LocalizationRow,
	vals map[string][]daoimpl.LocalizationValueRow,
) *form.LocalizationDTO {
	if id == nil || *id == "" {
		return nil
	}
	row, ok := locs[*id]
	if !ok {
		return nil
	}
	out := &form.LocalizationDTO{
		Lastupdated: row.Lastupdated,
		ID:          row.ID,
		Description: row.Description,
		Values:      map[string]form.LocalizedValueDTO{},
		ValuesAsMap: map[string]string{},
	}
	for _, v := range vals[*id] {
		out.Values[v.Locale] = form.LocalizedValueDTO{
			Lastupdated: v.Lastupdated,
			ID:          v.ID,
			Locale:      v.Locale,
			Value:       v.Value,
		}
		out.ValuesAsMap[v.Locale] = v.Value
		out.LocalesWithValue = append(out.LocalesWithValue, v.Locale)
		out.LocalesAndValuesOfLocalesWithValues = append(
			out.LocalesAndValuesOfLocalesWithValues,
			localeDisplayNames[v.Locale]+": "+v.Value)
		switch v.Locale {
		case "en":
			out.English = v.Value
		case "fr":
			out.French = v.Value
		}
	}
	// localizedValue follows the ACTIVE locale, which is English here.
	out.LocalizedValue = out.English
	out.AllActiveLocales = append([]string{}, out.LocalesWithValue...)
	out.LocalesWithValueSortedForDisplay = append([]string{}, out.LocalesWithValue...)
	out.LocalesSortedForDisplay = append([]string{}, out.LocalesWithValue...)
	return out
}

// synthesiseUOMLocalization ports UnitOfMeasure.getLocalization, which does NOT
// read the localization table at all — unit_of_measure has no
// name_localization_id column. It CONSTRUCTS a Localization from the UOM row
// with the English value set to the unit name and the French value set to the
// literal string "French", a placeholder left behind by whoever stubbed the
// method out. That literal reaches the client.
//
// The synthesised object also has no lastupdated, which is why this one
// localization carries twelve keys where every other carries thirteen.
func synthesiseUOMLocalization(id, description, name *string) *form.LocalizationDTO {
	if id == nil {
		return nil
	}
	english := deref(name)
	out := &form.LocalizationDTO{
		ID:          *id,
		Description: deref(description),
		Values: map[string]form.LocalizedValueDTO{
			"en": {Locale: "en", Value: english},
			"fr": {Locale: "fr", Value: "French"},
		},
		LocalizedValue:                   english,
		LocalesWithValue:                 []string{"en", "fr"},
		English:                          english,
		French:                           "French",
		ValuesAsMap:                      map[string]string{"en": english, "fr": "French"},
		AllActiveLocales:                 []string{"en", "fr"},
		LocalesWithValueSortedForDisplay: []string{"en", "fr"},
		LocalesSortedForDisplay:          []string{"en", "fr"},
		LocalesAndValuesOfLocalesWithValues: []string{
			"English: " + english, "French: French",
		},
	}
	return out
}

// buildResultEntity assembles the whole nested graph for one analysis.
func buildResultEntity(
	r daoimpl.EntityGraphRow,
	locs map[string]daoimpl.LocalizationRow,
	vals map[string][]daoimpl.LocalizationValueRow,
) *form.ResultEntityDTO {
	if r.ResultID == nil {
		return nil
	}

	sample := &form.SampleEntityDTO{
		Lastupdated:              r.SampleLastupdated,
		IsActive:                 "Y",
		ID:                       r.SampleID,
		AccessionNumber:          r.SampleAccessionNumber,
		EnteredDate:              r.SampleEnteredDate,
		EnteredDateForDisplay:    deref(r.SampleEnteredDisplay),
		ReceivedTimestamp:        r.SampleReceivedMillis,
		ReceivedDateForDisplay:   deref(r.SampleReceivedDisplay),
		ReceivedTimeForDisplay:   deref(r.SampleReceivedTime),
		CollectionDate:           r.SampleCollectionMillis,
		CollectionDateForDisplay: deref(r.SampleCollectionDisplay),
		CollectionTimeForDisplay: deref(r.SampleCollectionTime),
		IsConfirmation:           derefB(r.SampleIsConfirmation),
		Priority:                 deref(r.SamplePriority),
		StorageSkipped:           derefB(r.SampleStorageSkipped),
		SampleProjects:           []any{},
		StatusID:                 deref(r.SampleStatusID),
		// objectId repeats the id, and boundTo/tableId are audit-trail
		// metadata the entity exposes as getters.
		ObjectID:                     r.SampleID,
		ReceivedDate:                 r.SampleReceivedMillis,
		Received24HourTimeForDisplay: deref(r.SampleReceivedTime),
		BoundTo:                      "SAMPLE",
		TableID:                      "1",
	}

	var tos *form.TypeOfSampleDTO
	if r.TosID != nil {
		active := derefB(r.TosIsActive)
		tos = &form.TypeOfSampleDTO{
			Lastupdated:       r.TosLastupdated,
			ID:                *r.TosID,
			Description:       deref(r.TosDescription),
			Domain:            deref(r.TosDomain),
			LocalAbbreviation: deref(r.TosLocalAbbrev),
			IsActive:          active,
			SortOrder:         derefI(r.TosSortOrder),
			Localization:      buildLocalization(r.TosLocalizID, locs, vals),
			Active:            active,
		}
	}

	item := &form.SampleItemEntityDTO{
		Lastupdated:    r.ItemLastupdated,
		ID:             r.ItemID,
		Sample:         sample,
		SortOrder:      deref(r.ItemSortOrder),
		TypeOfSample:   tos,
		TypeOfSampleID: deref(r.ItemTypeOfSampleID),
		CollectionDate: r.ItemCollectionDate,
		StatusID:       deref(r.ItemStatusID),
		Rejected:       derefB(r.ItemRejected),
		Voided:         derefB(r.ItemVoided),
		ChildAliquots:  nil,
		ObjectID:       r.ItemID,
		NestingLevel:   0,
		BoundTo:        "SAMPLE_ITEM",
		TableID:        "23",
	}

	section := buildSection(r, locs, vals)

	var uom *form.UnitOfMeasureDTO
	if r.UomID != nil {
		uom = &form.UnitOfMeasureDTO{
			Lastupdated:       r.UomLastupdated,
			Name:              deref(r.UomName),
			Key:               *r.UomID,
			IsActive:          deref(r.UomIsActive),
			ID:                *r.UomID,
			UnitOfMeasureName: deref(r.UomName),
			Description:       deref(r.UomDescription),
			Localization:      synthesiseUOMLocalization(r.UomID, r.UomDescription, r.UomName),
		}
	}

	localizedName := buildLocalization(r.TestNameLocalizID, locs, vals)
	reportingName := buildLocalization(r.TestRptLocalizID, locs, vals)

	// The test carries FIVE renderings of its own name. description is the raw
	// column; alternateTestDisplayValue is "description-name"; testDisplayValue
	// is "name-description"; augmentedTestName is the localized name plus the
	// sample type. Same two strings, four arrangements.
	testName := ""
	if localizedName != nil {
		testName = localizedName.LocalizedValue
	}
	desc := deref(r.TestDescription)
	test := &form.TestEntityDTO{
		Lastupdated:               r.TestLastupdated,
		Name:                      deref(r.TestName),
		SortOrder:                 deref(r.TestSortOrder),
		Key:                       r.TestID,
		IsActive:                  deref(r.TestIsActive),
		ID:                        r.TestID,
		TestSection:               section,
		Description:               desc,
		NormalizedDescription:     deref(r.TestDescription),
		Domain:                    deref(r.TestDomain),
		AlternateTestDisplayValue: desc + "-" + testName,
		IsReportable:              deref(r.TestIsReportable),
		UnitOfMeasure:             uom,
		LocalCode:                 deref(r.TestLocalCode),
		Orderable:                 derefB(r.TestOrderable),
		LocalizedTestName:         localizedName,
		LocalizedReportingName:    reportingName,
		GUID:                      deref(r.TestGUID),
		InLabOnly:                 derefB(r.TestInLabOnly),
		NotifyResults:             derefB(r.TestNotifyResults),
		AntimicrobialResistance:   derefB(r.TestAMR),
		Active:                    deref(r.TestIsActive) == "Y",
		TestDisplayValue:          testName + "-" + desc,
		// The localized name plus the sample type — the same string the ROW
		// emits as testName, reached through a different getter.
		AugmentedTestName: deref(r.TestAugmentedName),
	}
	if r.TestNormalizedDescription != nil {
		test.NormalizedDescription = *r.TestNormalizedDescription
	}

	var panel *form.PanelDTO
	if r.PanelID != nil {
		panel = &form.PanelDTO{
			Lastupdated:  r.PanelLastupdated,
			IsActive:     deref(r.PanelIsActive),
			ID:           *r.PanelID,
			PanelName:    deref(r.PanelName),
			Description:  deref(r.PanelDescription),
			SortOrderInt: derefI(r.PanelSortOrder),
			Localization: buildLocalization(r.PanelLocalizID, locs, vals),
		}
	}

	analysis := &form.AnalysisEntityDTO{
		Lastupdated:  r.AnalysisLastupdated,
		ID:           r.AnalysisID,
		SampleItem:   item,
		AnalysisType: deref(r.AnalysisType),
		TestSection:  section,
		Test:         test,
		Revision:     deref(r.AnalysisRevision),
		EnteredDate:  r.AnalysisEnteredDate,
		IsReportable: deref(r.AnalysisReportable),
		Panel:        panel,
		StatusID:     deref(r.AnalysisStatusID),
		ReferredOut:  derefB(r.AnalysisReferredOut),
		ObjectID:     r.AnalysisID,
		BoundTo:      "ANALYSIS",
		TableID:      "4",
	}

	return &form.ResultEntityDTO{
		Lastupdated:       r.ResultLastupdated,
		SortOrder:         deref(r.ResultSortOrder),
		IsActive:          "Y",
		ID:                *r.ResultID,
		Analysis:          analysis,
		IsReportable:      deref(r.ResultReportable),
		ResultType:        deref(r.ResultType),
		Value:             deref(r.ResultValue),
		MinNormal:         r.ResultMinNormal,
		MaxNormal:         r.ResultMaxNormal,
		SignificantDigits: derefI(r.ResultSigDigits),
		Grouping:          derefI(r.ResultGrouping),
	}
}

func buildSection(
	r daoimpl.EntityGraphRow,
	locs map[string]daoimpl.LocalizationRow,
	vals map[string][]daoimpl.LocalizationValueRow,
) *form.TestSectionDTO {
	if r.SectionID == nil {
		return nil
	}
	return &form.TestSectionDTO{
		Lastupdated:     r.SectionLastupdated,
		IsActive:        deref(r.SectionIsActive),
		ID:              *r.SectionID,
		IsExternal:      deref(r.SectionIsExternal),
		TestSectionName: deref(r.SectionName),
		Description:     deref(r.SectionDescription),
		SortOrderInt:    derefI(r.SectionSortOrder),
		Localization:    buildLocalization(r.SectionLocalizID, locs, vals),
	}
}

func derefB(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

var _ = strings.TrimSpace

// attachEntityGraph loads the nested Result entity for every row of an
// accession-results response and hangs it off the row's `result` key.
//
// Kept separate from the row build because only this endpoint carries it:
// LogbookResults nests the five-key ResultRefDTO for the same field.
func (s *ResultService) attachEntityGraph(accessionNumber string, rows []form.TestResultRowDTO) error {
	graph, err := s.DAO.EntityGraphForAccession(accessionNumber)
	if err != nil {
		return err
	}
	locs, vals, err := s.DAO.Localizations()
	if err != nil {
		return err
	}
	byAnalysis := map[string]*form.ResultEntityDTO{}
	for _, g := range graph {
		if e := buildResultEntity(g, locs, vals); e != nil {
			byAnalysis[g.AnalysisID] = e
		}
	}
	for i := range rows {
		if e, ok := byAnalysis[rows[i].AnalysisID]; ok {
			rows[i].Result = e
		}
	}
	return nil
}
