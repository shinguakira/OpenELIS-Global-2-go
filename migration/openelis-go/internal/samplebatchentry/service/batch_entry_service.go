// Package service ports the sample batch-entry form load (Wave 4.8).
// Folder layout mirrors the Java source during migration.
package service

import (
	"time"

	commonservices "openelis-go/internal/common/services"
	"openelis-go/internal/samplebatchentry/form"
)

// hivResultCategory is the dictionary category behind ProjectData.hivStatusList.
const hivResultCategory = "HIVResult"

// BatchEntrySetupService backs GET rest/SampleBatchEntrySetup.
type BatchEntrySetupService struct {
	Lists *commonservices.DisplayListService
	// Zone is the display zone dates and times render in — the same one the
	// other c2 endpoints use, resolved once at wiring rather than per call.
	Zone *time.Location
}

// GetSetup builds the form.
//
// Everything here is reference data plus bean defaults: the endpoint takes no
// parameters and reads no sample.
func (s *BatchEntrySetupService) GetSetup() (*form.BatchEntrySetupDTO, error) {
	now := time.Now().In(s.Zone)
	today := now.Format("02/01/2006")

	conditions, err := s.Lists.InitialSampleConditionList()
	if err != nil {
		return nil, err
	}
	// Every ACTIVE human type — NOT the role-filtered list SamplePatientEntry
	// uses under the same key.
	sampleTypes, err := s.Lists.ActiveSampleTypes()
	if err != nil {
		return nil, err
	}
	sections, err := s.Lists.TestSectionList()
	if err != nil {
		return nil, err
	}
	projects, err := s.Lists.ProjectEntities()
	if err != nil {
		return nil, err
	}
	orderItems, err := BuildBlankOrderItems(s.Lists, today, now.Format("15:04"))
	if err != nil {
		return nil, err
	}
	projectData, err := s.buildProjectData()
	if err != nil {
		return nil, err
	}
	// Two SEPARATE instances with identical contents — Java constructs one per
	// project flavour and populates neither. Built twice rather than aliased so
	// a future divergence cannot silently affect both.
	projectDataVL, err := s.buildProjectData()
	if err != nil {
		return nil, err
	}

	return &form.BatchEntrySetupDTO{
		FormName:   "sampleBatchEntryForm",
		FormMethod: "POST",

		CancelAction: "Home",
		CancelMethod: "POST",

		CurrentDate: today,
		// Exclusive to this form; SamplePatientEntry has no top-level time.
		CurrentTime: now.Format("15:04"),

		PatientUpdateStatus: "ADD",
		Project:             "",
		SampleXML:           "",

		InitialSampleConditionList: conditions,
		SampleTypes:                sampleTypes,
		TestSectionList:            sections,
		Projects:                   projects,
		SampleOrderItems:           orderItems,
		ProjectDataEID:             projectData,
		ProjectDataVL:              projectDataVL,
	}, nil
}

// buildProjectData returns a fresh ProjectData bean: two booleans true, the
// rest false, three empty lists, and the HIVResult dictionary entities.
func (s *BatchEntrySetupService) buildProjectData() (*form.ProjectDataDTO, error) {
	hiv, err := s.Lists.DictionaryEntities(hivResultCategory)
	if err != nil {
		return nil, err
	}
	return &form.ProjectDataDTO{
		// The only two fields the bean initialises to true.
		AbbottOrRocheAnalysis: true,
		PreservCytTaken:       true,

		EidWhichPCRList:          []any{},
		EidSecondPCRReasonList:   []any{},
		IsUnderInvestigationList: []any{},
		HivStatusList:            hiv,
	}, nil
}

// BuildBlankOrderItems assembles the blank-form sampleOrderItems shared by
// SampleBatchEntrySetup and SamplePatientEntry.
//
// getBaseSampleOrderItem stamps requestDate, receivedDateForDisplay and
// receivedTime from the CLOCK before any sample is consulted — they are load
// timestamps, not data.
func BuildBlankOrderItems(lists *commonservices.DisplayListService, today, nowTime string) (*form.SampleOrderFormDTO, error) {
	payment, err := lists.PaymentOptions()
	if err != nil {
		return nil, err
	}
	program, err := lists.ProgramList()
	if err != nil {
		return nil, err
	}
	providers, err := lists.ProvidersList()
	if err != nil {
		return nil, err
	}
	referring, err := lists.ReferringSiteList()
	if err != nil {
		return nil, err
	}
	locations, err := lists.TestLocationCodeList()
	if err != nil {
		return nil, err
	}
	return &form.SampleOrderFormDTO{
		RequestDate:            today,
		ReceivedDateForDisplay: today,
		ReceivedTime:           nowTime,
		PaymentOptions:         payment,
		PriorityList:           lists.PriorityList(),
		ProgramList:            program,
		ProvidersList:          providers,
		ReferringSiteList:      referring,
		TestLocationCodeList:   locations,
		EnvironmentalFields:    map[string]any{},
	}, nil
}
