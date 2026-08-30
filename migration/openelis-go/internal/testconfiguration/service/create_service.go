package service

import (
	"math"
	"strings"
	"time"

	"gorm.io/gorm"

	"openelis-go/internal/common/util"
	locform "openelis-go/internal/localization/form"

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

// PanelForm ports showPanelCreate + setupDisplayItems.
//
// The two panel lists are NOT id/value pairs like every other list on these
// screens: each entry is a sample type name with the panels tied to it, and the
// panels are whole entities with their Localization nested. See the DTO.
func (s *CreateService) PanelForm() (*form.PanelCreateForm, error) {
	sampleTypes, err := s.Lists.ActiveHumanSampleTypes()
	if err != nil {
		return nil, err
	}
	joined, err := s.Lists.PanelsJoinedToSampleTypes()
	if err != nil {
		return nil, err
	}
	locales, err := s.Lists.ActiveLocales()
	if err != nil {
		return nil, err
	}
	values, err := s.Lists.LocalizationValues()
	if err != nil {
		return nil, err
	}

	active := s.panelsBySampleType(joined, values, locales, true)
	inactive := s.panelsBySampleType(joined, values, locales, false)

	existing := make([]form.SampleTypePanelDTO, 0, len(sampleTypes))
	inactiveList := make([]form.SampleTypePanelDTO, 0, len(sampleTypes))
	for _, st := range sampleTypes {
		existing = append(existing, form.SampleTypePanelDTO{
			TypeOfSampleName: st.Value, Panels: active[st.Value],
		})
		inactiveList = append(inactiveList, form.SampleTypePanelDTO{
			TypeOfSampleName: st.Value, Panels: inactive[st.Value],
		})
	}

	en, fr, err := s.namePair("clinlims.panel", "name", "", "e.id")
	if err != nil {
		return nil, err
	}

	f := form.NewPanelCreateForm()
	f.ExistingPanelList, f.InactivePanelList = &existing, &inactiveList
	f.ExistingSampleTypeList = &sampleTypes
	f.ExistingEnglishNames, f.ExistingFrenchNames = &en, &fr
	return &f, nil
}

// panelsBySampleType is createTypeOfSamplePanelMap.
//
// The map entry is created as soon as a join row for that sample type is seen,
// BEFORE the active filter — so a sample type whose panels all fail the filter
// keeps an empty list rather than dropping out. Absent and empty are different
// documents here; see the DTO.
func (s *CreateService) panelsBySampleType(
	joined []commondaoimpl.PanelRow,
	values map[string][]locform.LocalizationValue,
	locales []string,
	wantActive bool,
) map[string]*[]form.PanelDTO {
	out := map[string]*[]form.PanelDTO{}
	for _, r := range joined {
		if _, ok := out[r.SampleTypeName]; !ok {
			empty := []form.PanelDTO{}
			out[r.SampleTypeName] = &empty
		}
		if (r.IsActive == "Y") != wantActive {
			continue
		}
		list := append(*out[r.SampleTypeName], form.PanelDTO{
			Lastupdated:  r.Lastupdated,
			IsActive:     r.IsActive,
			ID:           r.ID,
			PanelName:    r.Name,
			Description:  r.Description,
			Loinc:        r.Loinc,
			SortOrderInt: r.SortOrder,
			Localization: locform.BuildLocalization(
				r.LocalizationID, r.LocDescription, r.LocUpdated,
				values[r.LocalizationID], locales, s.Lists.Locale()),
		})
		out[r.SampleTypeName] = &list
	}
	return out
}

// CreatePanel ports postPanelCreate.
//
// Nine rows, not eight: a panel is created already tied to the sample type the
// form chose, through a sampletype_panel row the other creates have no
// equivalent of. Its system_module DESCRIPTIONS also differ —
// `Workplan=>panel=><name>` where the others build `Workplan=><name>` — because
// this controller has its own copy of createSystemModule.
func (s *CreateService) CreatePanel(post form.CreatePost, sysUserID int64) (*form.PanelCreateForm, error) {
	f := form.NewPanelCreateForm()
	f.PanelEnglishName, f.PanelFrenchName, f.SampleTypeID = post.PanelEnglishName, post.PanelFrenchName, post.SampleTypeID

	sampleTypeID := derefOr(post.SampleTypeID)
	_, err := s.DAO.Create(daoimpl.CreateSpec{
		Table:                   "clinlims.panel",
		LocalizationDescription: "panel name",
		AuditTable:              "PANEL",
		NameColumn:              "name",
		DescriptionColumn:       "description",
		ModuleInfix:             "panel=>",
		ExtraColumns: map[string]any{
			"is_active":  "N",
			"sort_order": maxSortOrder,
			// setLoinc(form.getPanelLoinc()), and the submitted value DOES reach
			// it. panelLoinc is absent from ALLOWED_FIELDS, which reads like it
			// cannot be bound — but initBinder.setAllowedFields governs FORM
			// binding and these endpoints take @RequestBody, which is Jackson.
			// The allow-list is dead configuration on every REST controller in
			// this package. Measured: the column comes back holding what was sent.
			"loinc": loincOrNil(post.PanelLoinc),
		},
		AfterEntity: func(tx *gorm.DB, panelID string, ts time.Time) error {
			if sampleTypeID == "" {
				return nil
			}
			// The table is sampletype_panel and the sequence is
			// sample_type_panel_seq — the two are spelled differently, and the
			// name that matches the table does not exist.
			return tx.Exec(`
				INSERT INTO clinlims.sampletype_panel (id, sample_type_id, panel_id)
				VALUES (nextval('clinlims.sample_type_panel_seq'), ?, ?)`,
				sampleTypeID, panelID).Error
		},
	}, derefOr(post.PanelEnglishName), derefOr(post.PanelFrenchName), sysUserID)
	if err != nil {
		return &f, nil
	}
	return &f, nil
}

// loincOrNil keeps an absent panelLoinc out of the column as NULL rather than
// as the empty string, which is what setLoinc(null) leaves behind.
func loincOrNil(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}
