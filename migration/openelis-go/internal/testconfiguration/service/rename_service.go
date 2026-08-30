package service

import (
	"errors"
	"strings"

	commondaoimpl "openelis-go/internal/common/daoimpl"
	"openelis-go/internal/testconfiguration/daoimpl"
	"openelis-go/internal/testconfiguration/form"
)

// RenameService ports the four *RenameEntry controllers whose lists this port
// already knows how to build: Method, Panel, SampleType and TestSection.
//
// The lists here are read LIVE, not cached. That is not an oversight and not a
// contradiction of the UOM service beside it: those controllers call
// DisplayListService.getList, which reads the map that is refreshed on every
// write through the application — and every write that changes these lists goes
// through the application. UomCreate's `inactiveUomList` is the odd one out
// because its refresh is a no-op, which is why that one is pinned as a cache.
//
// If a screen is later found whose list Java serves stale, it moves here as a
// cache and the spec says so. Nothing is cached on suspicion.
type RenameService struct {
	Lists *commondaoimpl.DisplayListDAOImpl
	DAO   *daoimpl.RenameDAO
}

// MethodForm ports showMethodRenameEntry — METHODS.
func (s *RenameService) MethodForm() (*form.MethodRenameEntryForm, error) {
	list, err := s.Lists.Methods()
	if err != nil {
		return nil, err
	}
	f := form.NewMethodRenameEntryForm()
	f.MethodList = &list
	return &f, nil
}

// PanelForm ports showPanelRenameEntry — PANELS.
func (s *RenameService) PanelForm() (*form.PanelRenameEntryForm, error) {
	list, err := s.Lists.Panels()
	if err != nil {
		return nil, err
	}
	f := form.NewPanelRenameEntryForm()
	f.PanelList = &list
	return &f, nil
}

// SampleTypeForm ports showSampleTypeRenameEntry — SAMPLE_TYPE_ACTIVE.
func (s *RenameService) SampleTypeForm() (*form.SampleTypeRenameEntryForm, error) {
	list, err := s.Lists.ActiveHumanSampleTypes()
	if err != nil {
		return nil, err
	}
	f := form.NewSampleTypeRenameEntryForm()
	f.SampleTypeList = &list
	return &f, nil
}

// TestSectionForm ports showTestSectionRenameEntry — TEST_SECTION_ACTIVE.
func (s *RenameService) TestSectionForm() (*form.TestSectionRenameEntryForm, error) {
	list, err := s.Lists.ActiveTestSections()
	if err != nil {
		return nil, err
	}
	f := form.NewTestSectionRenameEntryForm()
	f.TestSectionList = &list
	return &f, nil
}

// ErrEntityNotFound is the ObjectNotFoundException Java throws for an id that
// names nothing — see rename below. The controller turns it into the 500 Java
// answers.
var ErrEntityNotFound = errors.New("testconfiguration: entity not found")

// rename is the body all four POST handlers share: load the entity's
// localization, write English and French onto it, trimmed.
//
// WHAT AN UNKNOWN ID DOES IS PER-SCREEN, and the four disagree.
//
// All four guard the write with `if (entity != null)`, which reads like "an id
// that names nothing is a silent 200". For three of them it is: Panel,
// SampleType and TestSection call an entity-specific finder —
// getPanelById, getTypeOfSampleById, getTestSectionById — that really does
// return null, the block is skipped, and the caller is told 200 with nothing
// written.
//
// MethodRenameEntry calls the GENERIC methodService.get(id), and
// BaseObjectServiceImpl.get is
// `getBaseObjectDAO().get(id).orElseThrow(() -> new ObjectNotFoundException(...))`.
// It never returns null, so the guard beneath it is dead code and the request
// ends as a 500. Measured on all four; the difference is one method call.
func (s *RenameService) rename(kind, id, english, french string, throwOnMiss bool) error {
	if strings.TrimSpace(id) == "" {
		if throwOnMiss {
			return ErrEntityNotFound
		}
		return nil
	}
	locID, err := s.DAO.LocalizationIDFor(kind, id)
	if err != nil {
		return err
	}
	if locID == "" {
		if throwOnMiss {
			return ErrEntityNotFound
		}
		// Either the entity does not exist, or it exists with no localization.
		// Java cannot tell those apart either: getLocalization() on a row with a
		// null FK yields null and the setter call would throw, which the
		// surrounding try/catch swallows at DEBUG.
		return nil
	}
	return s.DAO.Rename(nil, locID, strings.TrimSpace(english), strings.TrimSpace(french))
}

// RenameMethod ports updateMethodRenameEntry — the one that throws on a miss.
func (s *RenameService) RenameMethod(post form.RenamePost) (*form.MethodRenameEntryForm, error) {
	f := form.NewMethodRenameEntryForm()
	f.MethodID, f.NameEnglish, f.NameFrench = bind(post.MethodID, post.NameEnglish, post.NameFrench)
	return &f, s.rename("method", f.MethodID, f.NameEnglish, f.NameFrench, true)
}

// RenamePanel ports updatePanelRenameEntry.
func (s *RenameService) RenamePanel(post form.RenamePost) (*form.PanelRenameEntryForm, error) {
	f := form.NewPanelRenameEntryForm()
	f.PanelID, f.NameEnglish, f.NameFrench = bind(post.PanelID, post.NameEnglish, post.NameFrench)
	return &f, s.rename("panel", f.PanelID, f.NameEnglish, f.NameFrench, false)
}

// RenameSampleType ports updateSampleTypeRenameEntry.
//
// The ONLY one of the four that re-populates its list on the SUCCESS path: it
// calls setSampleTypeList again after the write, so this POST answers the list
// where the other three answer the bound form alone. Measured.
func (s *RenameService) RenameSampleType(post form.RenamePost) (*form.SampleTypeRenameEntryForm, error) {
	f := form.NewSampleTypeRenameEntryForm()
	f.SampleTypeID, f.NameEnglish, f.NameFrench = bind(post.SampleTypeID, post.NameEnglish, post.NameFrench)
	if err := s.rename("sampleType", f.SampleTypeID, f.NameEnglish, f.NameFrench, false); err != nil {
		return nil, err
	}
	list, err := s.Lists.ActiveHumanSampleTypes()
	if err != nil {
		return nil, err
	}
	f.SampleTypeList = &list
	return &f, nil
}

// RenameTestSection ports updateTestSectionRenameEntry.
func (s *RenameService) RenameTestSection(post form.RenamePost) (*form.TestSectionRenameEntryForm, error) {
	f := form.NewTestSectionRenameEntryForm()
	f.TestSectionID, f.NameEnglish, f.NameFrench = bind(post.TestSectionID, post.NameEnglish, post.NameFrench)
	return &f, s.rename("testSection", f.TestSectionID, f.NameEnglish, f.NameFrench, false)
}

// bind applies the submitted values over the bean defaults, which are "" and
// not null — so an omitted key answers "" rather than dropping out.
func bind(id, english, french *string) (string, string, string) {
	return derefOr(id), derefOr(english), derefOr(french)
}

func derefOr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
