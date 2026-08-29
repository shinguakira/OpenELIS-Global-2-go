package service

import (
	"math"
	"strings"

	"openelis-go/internal/common/util"

	commondaoimpl "openelis-go/internal/common/daoimpl"
	"openelis-go/internal/testconfiguration/daoimpl"
	"openelis-go/internal/testconfiguration/form"
)

// CreateService ports MethodCreateRestController, TestSectionCreateRestController
// and SampleTypeCreateRestController.
//
// Three screens, one write: eight rows across six tables, and one history row
// for the entity. See the DAO.
type CreateService struct {
	Lists *commondaoimpl.DisplayListDAOImpl
	DAO   *daoimpl.CreateDAO
}

// maxSortOrder is Integer.MAX_VALUE, which createTestSection and
// createTypeOfSample both use so a new row sorts last.
const maxSortOrder = math.MaxInt32

// MethodForm ports showMethodCreate + setupDisplayMethods.
func (s *CreateService) MethodForm() (*form.MethodCreateForm, error) {
	existing, err := s.Lists.Methods()
	if err != nil {
		return nil, err
	}
	inactive, err := s.Lists.InactiveMethods()
	if err != nil {
		return nil, err
	}
	en, fr, err := s.namePair("clinlims.method", "name", "", "e.id")
	if err != nil {
		return nil, err
	}
	f := form.NewMethodCreateForm()
	f.ExistingMethodList, f.InactiveMethodList = &existing, &inactive
	f.ExistingEnglishNames, f.ExistingFrenchNames = &en, &fr
	return &f, nil
}

// TestSectionForm ports showTestSectionCreate.
func (s *CreateService) TestSectionForm() (*form.TestSectionCreateForm, error) {
	existing, err := s.Lists.ActiveTestSections()
	if err != nil {
		return nil, err
	}
	inactive, err := s.Lists.InactiveTestSections()
	if err != nil {
		return nil, err
	}
	en, fr, err := s.namePair("clinlims.test_section", "name", "e.is_active = 'Y'", "e.sort_order")
	if err != nil {
		return nil, err
	}
	f := form.NewTestSectionCreateForm()
	f.ExistingTestUnitList, f.InactiveTestUnitList = &existing, &inactive
	f.ExistingEnglishNames, f.ExistingFrenchNames = &en, &fr
	return &f, nil
}

// SampleTypeForm ports showSampleTypeCreate.
func (s *CreateService) SampleTypeForm() (*form.SampleTypeCreateForm, error) {
	existing, err := s.Lists.ActiveHumanSampleTypes()
	if err != nil {
		return nil, err
	}
	inactive, err := s.Lists.InactiveHumanSampleTypes()
	if err != nil {
		return nil, err
	}
	en, fr, err := s.namePair("clinlims.type_of_sample", "description", "e.domain = 'H'", "e.sort_order")
	if err != nil {
		return nil, err
	}
	f := form.NewSampleTypeCreateForm()
	f.ExistingSampleTypeList, f.InactiveSampleTypeList = &existing, &inactive
	f.ExistingEnglishNames, f.ExistingFrenchNames = &en, &fr
	return &f, nil
}

// namePair builds the two "$name1$name2$" strings the create screens carry so
// the browser can refuse a duplicate before submitting.
//
// Both are seeded with the separator and append one after every name, so each
// carries a leading AND trailing "$" — an empty list yields "$", not "".
//
// Unlike UOM, whose French string is the hardcoded literal "French", these
// entities have real localization rows: the pair is the SAME query run twice,
// once per locale.
func (s *CreateService) namePair(table, fallbackColumn, where, order string) (string, string, error) {
	en, err := s.Lists.NamesIn(table, fallbackColumn, s.Lists.Locale(), where, order)
	if err != nil {
		return "", "", err
	}
	fr, err := s.Lists.NamesIn(table, fallbackColumn, "fr", where, order)
	if err != nil {
		return "", "", err
	}
	return joinNames(en), joinNames(fr), nil
}

func joinNames(pairs []util.IdValuePair) string {
	var b strings.Builder
	b.WriteString(nameSeparator)
	for _, p := range pairs {
		b.WriteString(p.Value)
		b.WriteString(nameSeparator)
	}
	return b.String()
}

// CreateMethod ports postMethodCreate.
func (s *CreateService) CreateMethod(post form.CreatePost, sysUserID int64) (*form.MethodCreateForm, error) {
	f := form.NewMethodCreateForm()
	f.MethodEnglishName, f.MethodFrenchName, f.MethodCode = post.MethodEnglishName, post.MethodFrenchName, post.MethodCode

	spec := daoimpl.CreateSpec{
		Table:                   "clinlims.method",
		LocalizationDescription: "method name",
		AuditTable:              "METHOD",
		NameColumn:              "name",
		DescriptionColumn:       "description",
		// setIsActive("N") — method.is_active is the CHAR, not a boolean.
		ExtraColumns: map[string]any{"is_active": "N"},
	}
	// setCode(code.toUpperCase()) — and only when the field is non-blank, which
	// matters because method.code carries a UNIQUE constraint and every blank
	// would collide after the first.
	if code := derefOr(post.MethodCode); strings.TrimSpace(code) != "" {
		spec.ExtraColumns["code"] = strings.ToUpper(code)
	}
	_, err := s.DAO.Create(spec, derefOr(post.MethodEnglishName), derefOr(post.MethodFrenchName), sysUserID)
	if err != nil {
		// LIMSRuntimeException is caught and logged at DEBUG; the form comes
		// back as though the insert had worked.
		return &f, nil
	}
	return &f, nil
}

// CreateTestSection ports postTestSectionCreate.
func (s *CreateService) CreateTestSection(post form.CreatePost, sysUserID int64) (*form.TestSectionCreateForm, error) {
	f := form.NewTestSectionCreateForm()
	f.TestUnitEnglishName, f.TestUnitFrenchName = post.TestUnitEnglishName, post.TestUnitFrenchName

	_, err := s.DAO.Create(daoimpl.CreateSpec{
		Table:                   "clinlims.test_section",
		LocalizationDescription: "test unit name",
		AuditTable:              "TEST_SECTION",
		NameColumn:              "name",
		DescriptionColumn:       "description",
		ExtraColumns:            map[string]any{"is_active": "N", "sort_order": maxSortOrder},
	}, derefOr(post.TestUnitEnglishName), derefOr(post.TestUnitFrenchName), sysUserID)
	if err != nil {
		return &f, nil
	}
	return &f, nil
}

// CreateSampleType ports postSampleTypeCreate.
func (s *CreateService) CreateSampleType(post form.CreatePost, sysUserID int64) (*form.SampleTypeCreateForm, error) {
	f := form.NewSampleTypeCreateForm()
	f.SampleTypeEnglishName, f.SampleTypeFrenchName = post.SampleTypeEnglishName, post.SampleTypeFrenchName

	name := derefOr(post.SampleTypeEnglishName)
	// setLocalAbbreviation(name.length() > 10 ? name.substring(0, 10) : name) —
	// a byte-count truncation in Java, applied here on runes so a multi-byte
	// name is not cut mid-character. No shipped name reaches ten characters of
	// anything but ASCII, so the two agree on this data.
	abbrev := name
	if r := []rune(abbrev); len(r) > 10 {
		abbrev = string(r[:10])
	}

	_, err := s.DAO.Create(daoimpl.CreateSpec{
		Table:                   "clinlims.type_of_sample",
		LocalizationDescription: "type of sample name",
		AuditTable:              "TYPE_OF_SAMPLE",
		// type_of_sample has NO name column; createTypeOfSample calls
		// setDescription only, so the description carries the name.
		NameColumn:        "description",
		DescriptionColumn: "description",
		ExtraColumns: map[string]any{
			"domain": "H",
			// is_active here IS a real boolean, unlike method and test_section.
			"is_active":    false,
			"sort_order":   maxSortOrder,
			"local_abbrev": abbrev,
		},
	}, name, derefOr(post.SampleTypeFrenchName), sysUserID)
	if err != nil {
		return &f, nil
	}
	return &f, nil
}
