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
// RoleReception is Constants.ROLE_RECEPTION — the role whose lab units decide
// which sample types the order-entry forms offer.
const RoleReception = "Reception"

type DisplayListService struct {
	DAO *daoimpl.DisplayListDAOImpl
	// Messages is the message bundle (message_en.properties), used wherever
	// Java resolves a label through MessageUtil rather than a column.
	Messages map[string]string
	// StringContext is site_information.stringContext (here: "CI").
	// MessageUtil.getContextualMessage appends "." + this to the key and uses
	// the result WHEN THAT KEY EXISTS, falling back to the bare key otherwise.
	// It is the difference between gender.male ("Male") and gender.male.CI
	// ("1 = Male") — the deployment ships both and the form shows the second.
	StringContext string
}

// message resolves a bundle key.
//
// A MISSING key resolves to the KEY ITSELF, which is what Spring's
// MessageSource does and what the live response shows: the bundle has no
// label.select.last.first.name, and Java renders that entry as
// "3. label.select.last.first.name". Substituting invented English there would
// be a different response.
//
// The fallback argument is only for callers that have a genuine non-key
// default; pass "" to get the key-as-value behaviour.
func (s *DisplayListService) message(key, fallback string) string {
	if v, ok := s.Messages[key]; ok && v != "" {
		return v
	}
	if fallback != "" {
		return fallback
	}
	return key
}

// contextualMessage ports MessageUtil.getContextualMessage: try the
// suffixed key first, fall back to the bare key, then to the supplied default.
func (s *DisplayListService) contextualMessage(key, fallback string) string {
	if s.StringContext != "" {
		if v, ok := s.Messages[key+"."+s.StringContext]; ok && v != "" {
			return v
		}
	}
	return s.message(key, fallback)
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

// UserSampleTypes is getUserSampleTypes(userId, ROLE_RECEPTION) — the
// ROLE-FILTERED sample types SamplePatientEntry and SampleEdit use, as opposed
// to ActiveSampleTypes which SampleBatchEntrySetup uses. Same JSON key, two
// different lists; see the DAO.
func (s *DisplayListService) UserSampleTypes(systemUserID string) ([]util.IdValuePair, error) {
	pairs, err := s.DAO.UserSampleTypes(systemUserID, RoleReception)
	if err != nil {
		return nil, err
	}
	// Java collects the ids into a HashSet and iterates it, so the wire order is
	// HashMap bucket order rather than anything the query controls. See
	// JavaHashSetOrder.
	return JavaHashSetOrder(pairs), nil
}

// Projects returns the full project entities the form loads emit.
func (s *DisplayListService) Projects() ([]daoimpl.ProjectRow, error) {
	return s.DAO.Projects()
}

// PatientTypes returns the full patient_type entities.
func (s *DisplayListService) PatientTypes() ([]daoimpl.PatientTypeRow, error) {
	return s.DAO.PatientTypes()
}
