package services

import (
	"openelis-go/internal/common/util"
)

// Dictionary categories behind the patient-demographic lists. Each is the
// literal string DisplayListService passes to
// createFromDictionaryCategoryLocalizedSort, so all three are localized-sorted.
const (
	categoryMaritalStatus = "Marital Status Demographic Information"
	categoryNationality   = "Nationality Demographic Information"
	categoryEducation     = "Education Level Demographic Information"
	// patientProperties.addressDepartments carries the FULL entities of this
	// category, not {id,value} pairs.
	categoryAddressDepartments = "haitDepartments"
)

// MaritalList is ListType.PATIENT_MARITAL_STATUS.
func (s *DisplayListService) MaritalList() ([]util.IdValuePair, error) {
	return s.DAO.DictionaryByCategoryLocalizedSort(categoryMaritalStatus)
}

// NationalityList is ListType.PATIENT_NATIONALITY.
func (s *DisplayListService) NationalityList() ([]util.IdValuePair, error) {
	return s.DAO.DictionaryByCategoryLocalizedSort(categoryNationality)
}

// EducationList is ListType.PATIENT_EDUCATION.
func (s *DisplayListService) EducationList() ([]util.IdValuePair, error) {
	return s.DAO.DictionaryByCategoryLocalizedSort(categoryEducation)
}

// AddressDepartments is the full-entity list patientProperties emits.
func (s *DisplayListService) AddressDepartments() ([]util.IdValuePair, error) {
	return s.DAO.DictionaryByCategoryLocalizedSort(categoryAddressDepartments)
}

// Genders is ListType.GENDERS.
//
// The id is the gender TYPE ("M"/"F"), not the row id, and the label comes from
// the message bundle via name_key — so the wire shows {"id":"M","value":"1 = Male"}
// rather than the numeric id or the raw description.
func (s *DisplayListService) Genders() ([]util.IdValuePair, error) {
	rows, err := s.DAO.Genders()
	if err != nil {
		return nil, err
	}
	out := make([]util.IdValuePair, 0, len(rows))
	for _, r := range rows {
		label := r.Description
		if r.NameKey != nil {
			// CONTEXTUAL, not plain: gender.male is "Male" but this deployment
			// also ships gender.male.CI = "1 = Male", and that is what the form
			// shows.
			label = s.contextualMessage(*r.NameKey, r.Description)
		}
		out = append(out, util.NewIdValuePair(r.GenderType, label))
	}
	return out, nil
}

// PatientSearchCriteria is ListType.PATIENT_SEARCH_CRITERIA — a HARDCODED list,
// not a table.
//
// Note the ids are deliberately out of order against the numbering in the
// labels: entry "2" is displayed as "1. …" and entry "1" as "2. …". Java's own
// comment says to keep the id:value pairing if the order is ever changed, so
// the mismatch is intentional and must survive the port.
func (s *DisplayListService) PatientSearchCriteria() []util.IdValuePair {
	return []util.IdValuePair{
		util.NewIdValuePair("0", s.message("label.select.search.by", "")),
		util.NewIdValuePair("2", "1. "+s.message("label.select.last.name", "")),
		util.NewIdValuePair("1", "2. "+s.message("label.select.first.name", "")),
		util.NewIdValuePair("3", "3. "+s.message("label.select.last.first.name", "")),
		util.NewIdValuePair("4", "4. "+s.message("label.select.patient.ID", "")),
		// The last entry uses getContextualMessage while the four above use the
		// plain getMessage — Java mixes the two in one list.
		util.NewIdValuePair("5", "5. "+s.contextualMessage("quick.entry.accession.number", "")),
	}
}

// HealthRegions is ListType.PATIENT_HEALTH_REGIONS — organizations of the
// address-hierarchy level-1 type, or the legacy "Health Region" type. Empty on
// this dataset, and returned as [] rather than null.
func (s *DisplayListService) HealthRegions() ([]util.IdValuePair, error) {
	return s.DAO.OrganizationsByTypeName("Health Region")
}

// HealthDistricts is the level-2 equivalent. Also empty here.
func (s *DisplayListService) HealthDistricts() ([]util.IdValuePair, error) {
	return s.DAO.OrganizationsByTypeName("Health District")
}
