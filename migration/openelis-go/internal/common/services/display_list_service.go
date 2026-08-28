package services

import (
	"strings"

	"openelis-go/internal/common/daoimpl"
	"openelis-go/internal/common/util"
)

// DisplayListService ports org.openelisglobal.common.services.DisplayListService
// for the list types the Wave 4 form loads consume.
//
// Java caches these in a static map and exposes getList / getFreshList; the
// distinction is a caching one and has no observable effect on a response, so
// this port queries每 time and does not model the cache.
type DisplayListService struct {
	DAO *daoimpl.DisplayListDAOImpl
	// Messages is the message bundle (message_en.properties), used wherever
	// Java resolves a label through MessageUtil rather than a column.
	Messages map[string]string
}

// message resolves a bundle key, falling back to the supplied default the way
// MessageUtil does when a key is absent.
func (s *DisplayListService) message(key, fallback string) string {
	if v, ok := s.Messages[key]; ok && v != "" {
		return v
	}
	return fallback
}

// InitialSampleConditionList is ListType.INITIAL_SAMPLE_CONDITION —
// dictionary category "specimen reception condition", re-sorted by localized
// name.
func (s *DisplayListService) InitialSampleConditionList() ([]util.IdValuePair, error) {
	return s.DAO.DictionaryByCategoryLocalizedSort("specimen reception condition")
}

// RejectReasonList is ListType.REJECTION_REASONS — dictionary category
// "resultRejectionReasons", NOT re-sorted (see the DAO).
func (s *DisplayListService) RejectReasonList() ([]util.IdValuePair, error) {
	return s.DAO.DictionaryByCategory("resultRejectionReasons")
}

// PaymentOptions is ListType.SAMPLE_PATIENT_PAYMENT_OPTIONS — dictionary
// category "patientPayment", localized-sorted.
func (s *DisplayListService) PaymentOptions() ([]util.IdValuePair, error) {
	return s.DAO.DictionaryByCategoryLocalizedSort("patientPayment")
}

// TestLocationCodeList is ListType.TEST_LOCATION_CODE — category
// "testLocationCode", unsorted.
func (s *DisplayListService) TestLocationCodeList() ([]util.IdValuePair, error) {
	return s.DAO.DictionaryByCategory("testLocationCode")
}

// ProgramList is ListType.DICTIONARY_PROGRAM — category "programs", unsorted.
func (s *DisplayListService) ProgramList() ([]util.IdValuePair, error) {
	return s.DAO.DictionaryByCategory("programs")
}

// TestSectionList is ListType.TEST_SECTION_ACTIVE.
func (s *DisplayListService) TestSectionList() ([]util.IdValuePair, error) {
	return s.DAO.ActiveTestSections()
}

// ActiveSampleTypes is ListType.SAMPLE_TYPE_ACTIVE — every active human
// type_of_sample. SampleBatchEntrySetup uses this; the other two form loads use
// the role-filtered UserSampleTypes instead.
func (s *DisplayListService) ActiveSampleTypes() ([]util.IdValuePair, error) {
	return s.DAO.ActiveHumanSampleTypes()
}

// ReferralOrganizations is ListType.REFERRAL_ORGANIZATIONS.
func (s *DisplayListService) ReferralOrganizations() ([]util.IdValuePair, error) {
	return s.DAO.ReferralOrganizations()
}

// ReferralReasons is ListType.REFERRAL_REASONS.
//
// referral_reason localizes through display_key alone — it has no
// name_localization_id — so the label comes from the message bundle with the
// stored name as fallback. The name is TRIMMED either way: row 3 is stored as
// "Further testing required " and Java returns it without the trailing space.
func (s *DisplayListService) ReferralReasons() ([]util.IdValuePair, error) {
	rows, err := s.DAO.ReferralReasons()
	if err != nil {
		return nil, err
	}
	out := make([]util.IdValuePair, 0, len(rows))
	for _, r := range rows {
		label := strings.TrimSpace(r.Name)
		if r.DisplayKey != nil {
			label = strings.TrimSpace(s.message(*r.DisplayKey, label))
		}
		out = append(out, util.NewIdValuePair(r.ID, label))
	}
	return out, nil
}

// PriorityList is ListType.ORDER_PRIORITY — a hardcoded enum walk, in this
// order, with labels from the message bundle. NOT a table: the ids are enum
// NAMES, which is why they are upper-case words rather than numbers.
func (s *DisplayListService) PriorityList() []util.IdValuePair {
	return []util.IdValuePair{
		util.NewIdValuePair("ROUTINE", s.message("label.priority.routine", "Routine")),
		util.NewIdValuePair("ASAP", s.message("label.priority.asap", "ASAP")),
		util.NewIdValuePair("STAT", s.message("label.priority.stat", "STAT")),
		util.NewIdValuePair("TIMED", s.message("label.priority.timed", "Timed")),
		util.NewIdValuePair("FUTURE_STAT", s.message("label.priority.futureStat", "Future STAT")),
	}
}

// ProvidersList is ListType.PRACTITIONER_PERSONS.
//
// The id is the PERSON id, not the provider id — providers and persons are
// separate tables and the form binds to the person. Sorted by last name, with
// NULL last names pushed to the END rather than sorted as empty strings.
func (s *DisplayListService) ProvidersList() ([]util.IdValuePair, error) {
	return s.DAO.ActivePractitionerPersons()
}

// ReferringSiteList is ListType.SAMPLE_PATIENT_REFERRING_CLINIC.
//
// The value is "shortName - organizationName" when a short name exists and the
// bare organization name otherwise — a conditional label, not a column.
func (s *DisplayListService) ReferringSiteList() ([]util.IdValuePair, error) {
	return s.DAO.ReferringClinics()
}

// Projects returns the full project entities the form loads emit.
func (s *DisplayListService) Projects() ([]daoimpl.ProjectRow, error) {
	return s.DAO.Projects()
}

// PatientTypes returns the full patient_type entities.
func (s *DisplayListService) PatientTypes() ([]daoimpl.PatientTypeRow, error) {
	return s.DAO.PatientTypes()
}
