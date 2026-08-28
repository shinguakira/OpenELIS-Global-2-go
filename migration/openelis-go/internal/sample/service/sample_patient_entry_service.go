package service

import (
	"time"

	commonservices "openelis-go/internal/common/services"
	"openelis-go/internal/common/util"
	"openelis-go/internal/sample/form"
	batchservice "openelis-go/internal/samplebatchentry/service"
)

// addressDepartmentsCategory is the dictionary category behind
// patientProperties.addressDepartments — emitted as FULL entities.
const addressDepartmentsCategory = "haitDepartments"

// SamplePatientEntryService backs GET rest/SamplePatientEntry (Wave 4.6).
type SamplePatientEntryService struct {
	Lists *commonservices.DisplayListService
	// SysUserID decides the role-filtered sampleTypes list.
	SysUserID string
}

// GetForm builds the blank order-entry form. The endpoint takes no parameters
// and reads no sample.
func (s *SamplePatientEntryService) GetForm() (*form.SamplePatientEntryDTO, error) {
	now := time.Now().In(displayZone())
	today := now.Format("02/01/2006")

	conditions, err := s.Lists.InitialSampleConditionList()
	if err != nil {
		return nil, err
	}
	rejects, err := s.Lists.RejectReasonList()
	if err != nil {
		return nil, err
	}
	referralReasons, err := s.Lists.ReferralReasons()
	if err != nil {
		return nil, err
	}
	referralOrgs, err := s.Lists.ReferralOrganizations()
	if err != nil {
		return nil, err
	}
	// ROLE-FILTERED, unlike SampleBatchEntrySetup's every-active-type list.
	sampleTypes, err := s.Lists.UserSampleTypes(s.SysUserID)
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
	orderItems, err := batchservice.BuildBlankOrderItems(s.Lists, today, now.Format("15:04"))
	if err != nil {
		return nil, err
	}
	patientProps, genders, err := s.buildPatientProperties()
	if err != nil {
		return nil, err
	}

	return &form.SamplePatientEntryDTO{
		FormName:   "samplePatientEntryForm",
		FormMethod: "POST",

		CancelAction: "Home",
		CancelMethod: "POST",

		CurrentDate: today,
		// FALSE here; SampleEdit hardcodes it true.
		Warning: false,

		PatientUpdateStatus: "ADD",
		SampleXML:           "",

		InitialSampleConditionList: conditions,
		RejectReasonList:           rejects,
		ReferralReasons:            referralReasons,
		ReferralOrganizations:      referralOrgs,
		SampleTypes:                sampleTypes,
		TestSectionList:            sections,
		Projects:                   projects,

		PatientProperties: patientProps,
		PatientSearch: &form.PatientSearchDTO{
			DefaultHeader: true,
			// The SAME list patientProperties carries — one source, two places.
			Genders: genders,
			// Not set on this form. The MVC SampleEdit twin sets it true, but
			// the REST SampleEdit omits the whole object, so this is the only
			// endpoint in the wave that emits one.
			LoadFromServerWithPatient: false,
			SearchCriteria:            s.Lists.PatientSearchCriteria(),
		},

		SampleOrderItems: orderItems,
	}, nil
}

// buildPatientProperties assembles the blank-form patient block, returning the
// gender list too so patientSearch can reuse the identical slice — Java builds
// it once and hands the same list to both.
func (s *SamplePatientEntryService) buildPatientProperties() (*form.PatientEntryPropertiesDTO, []util.IdValuePair, error) {
	genders, err := s.Lists.Genders()
	if err != nil {
		return nil, nil, err
	}
	education, err := s.Lists.EducationList()
	if err != nil {
		return nil, nil, err
	}
	marital, err := s.Lists.MaritalList()
	if err != nil {
		return nil, nil, err
	}
	nationality, err := s.Lists.NationalityList()
	if err != nil {
		return nil, nil, err
	}
	// FULL dictionary entities, not {id,value} — the heavy shape, sitting
	// beside four lists that use the light one.
	departments, err := s.Lists.DictionaryEntities(addressDepartmentsCategory)
	if err != nil {
		return nil, nil, err
	}
	patientTypes, err := s.Lists.PatientTypeEntities()
	if err != nil {
		return nil, nil, err
	}
	regions, err := s.Lists.HealthRegions()
	if err != nil {
		return nil, nil, err
	}
	districts, err := s.Lists.HealthDistricts()
	if err != nil {
		return nil, nil, err
	}

	return &form.PatientEntryPropertiesDTO{
		AddressDepartments: departments,
		AddressHierarchy:   map[string]any{},
		// Initialised on the bean, so present and empty rather than dropped.
		BirthDateForDisplay: "",
		PatientType:         "",
		EducationList:       education,
		Genders:             genders,
		MaritialList:        marital,
		NationalityList:     nationality,
		HealthDistricts:     districts,
		HealthRegions:       regions,
		IDDocuments:         []any{},
		PatientTypes:        patientTypes,
		ReadOnly:            false,
	}, genders, nil
}
