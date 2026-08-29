package service

import (
	"strconv"
	"strings"
	"sync"

	"openelis-go/internal/common/util"
	"openelis-go/internal/testconfiguration/form"
	"openelis-go/internal/unitofmeasure/daoimpl"
)

// nameSeparator is UomCreateRestController.NAME_SEPARATOR.
const nameSeparator = "$"

// frenchName is the literal UnitOfMeasure.getLocalization() puts in every
// French slot — `_localization.setFrench("French")`. Not a placeholder this
// port invented; it is what the endpoint answers.
const frenchName = "French"

// UomConfigService ports UomCreateRestController and
// UomRenameEntryRestController, plus the two DisplayListService lists they read.
//
// THE LISTS ARE CACHES, NOT VIEWS, and the two are not refreshed alike — which
// is the whole reason this type holds state instead of querying per request.
//
// DisplayListService loads every list once at startup into a map. UomCreate's
// handler then calls refreshList on UNIT_OF_MEASURE and on
// UNIT_OF_MEASURE_INACTIVE. Only the first does anything: refreshList's switch
// has a case for UNIT_OF_MEASURE and NONE for UNIT_OF_MEASURE_INACTIVE, so the
// second call falls through and that list keeps its startup snapshot for the
// life of the process.
//
// The two lists are built by different methods — createUnitOfMeasureList and
// createUOMList — which are the same six lines: getAll(), mapped to
// (id, localizedName), with no is_active filter in either. The filter is
// present in createUnitOfMeasureList only as a commented-out line. So
// `inactiveUomList` is not an inactive list; it is a stale copy of the other
// one, and after the first create the two diverge permanently.
//
// A port that read the table per request would answer two identical, current
// lists — more correct than Java and therefore wrong. e1-8 settled the same
// question for the configuration cache.
type UomConfigService struct {
	DAO *daoimpl.UnitOfMeasureDAOImpl

	mu sync.RWMutex
	// active is UNIT_OF_MEASURE: reloaded by every write through this service.
	active []util.IdValuePair
	// inactive is UNIT_OF_MEASURE_INACTIVE: loaded once and never reloaded,
	// because the refresh Java performs on it is a no-op.
	inactive []util.IdValuePair
	loaded   bool
}

// Load takes the startup snapshot both lists begin as.
func (s *UomConfigService) Load() error {
	rows, err := s.DAO.GetAllForNames()
	if err != nil {
		return err
	}
	pairs := toPairs(rows)
	s.mu.Lock()
	s.active = pairs
	s.inactive = pairs
	s.loaded = true
	s.mu.Unlock()
	return nil
}

// refreshActive is refreshList(UNIT_OF_MEASURE). There is deliberately no
// refreshInactive: the call Java makes for the other list does nothing.
func (s *UomConfigService) refreshActive() error {
	rows, err := s.DAO.GetAllForNames()
	if err != nil {
		return err
	}
	pairs := toPairs(rows)
	s.mu.Lock()
	s.active = pairs
	s.mu.Unlock()
	return nil
}

func (s *UomConfigService) lists() (active, inactive []util.IdValuePair) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.active, s.inactive
}

func toPairs(rows []daoimpl.UomRow) []util.IdValuePair {
	out := make([]util.IdValuePair, 0, len(rows))
	for _, r := range rows {
		// IdValuePair carries getLocalizedName(), which for a UOM resolves to
		// getDefaultLocalizedName() — the name column. There is no localization
		// row to read; see the form's doc comment.
		out = append(out, util.NewIdValuePair(strconv.FormatInt(r.ID, 10), r.Name))
	}
	return out
}

// CreateForm ports showUomCreate + setupDisplayItems.
func (s *UomConfigService) CreateForm() (*form.UomCreateForm, error) {
	f := form.NewUomCreateForm()
	if err := s.fillCreateDisplayItems(&f); err != nil {
		return nil, err
	}
	return &f, nil
}

// fillCreateDisplayItems ports setupDisplayItems — the four fields the POST
// success branch does NOT set.
func (s *UomConfigService) fillCreateDisplayItems(f *form.UomCreateForm) error {
	if !s.loaded {
		if err := s.Load(); err != nil {
			return err
		}
	}
	active, inactive := s.lists()
	f.ExistingUomList = &active
	f.InactiveUomList = &inactive

	// getExistingUomNames reads getAll() LIVE, not the cached list — so the
	// name strings can disagree with existingUomList, which is a snapshot.
	rows, err := s.DAO.GetAllForNames()
	if err != nil {
		return err
	}
	english := namesString(rows, false)
	french := namesString(rows, true)
	f.ExistingEnglishNames = &english
	f.ExistingFrenchNames = &french
	return nil
}

// namesString ports getExistingUomNames: the builder is SEEDED with the
// separator and appends one after every name, so a leading and a trailing "$"
// are both part of the value. An empty table yields "$", not "".
func namesString(rows []daoimpl.UomRow, french bool) string {
	var b strings.Builder
	b.WriteString(nameSeparator)
	for _, r := range rows {
		if french {
			b.WriteString(frenchName)
		} else {
			b.WriteString(r.Name)
		}
		b.WriteString(nameSeparator)
	}
	return b.String()
}

// Create ports postUomCreate.
//
// The response is the BOUND FORM, not the display form: the success branch
// returns without calling setupDisplayItems, so the four list/name fields stay
// absent. Only the validation-failure branch re-populates them — and answers
// 200 as well, so the two outcomes are told apart by which keys are present
// rather than by status.
func (s *UomConfigService) Create(post form.UomCreatePost) (*form.UomCreateForm, error) {
	f := form.NewUomCreateForm()
	f.UomEnglishName = post.UomEnglishName

	name := ""
	if post.UomEnglishName != nil {
		name = *post.UomEnglishName
	}

	if _, err := s.DAO.Insert(name); err != nil {
		// LIMSRuntimeException is caught and logged at DEBUG, and the form is
		// returned as if the insert had worked. Reproduced: a refused insert is
		// a 200 carrying the submitted name back.
		return &f, nil
	}

	// refreshList(UNIT_OF_MEASURE) — and refreshList(UNIT_OF_MEASURE_INACTIVE),
	// which Java calls and which does nothing.
	if err := s.refreshActive(); err != nil {
		return nil, err
	}
	return &f, nil
}

// RenameForm ports showUomRenameEntry — one list, and it is the refreshed one.
func (s *UomConfigService) RenameForm() (*form.UomRenameEntryForm, error) {
	if !s.loaded {
		if err := s.Load(); err != nil {
			return nil, err
		}
	}
	f := form.NewUomRenameEntryForm()
	active, _ := s.lists()
	f.UomList = &active
	return &f, nil
}

// Rename ports updateUomRenameEntry + updateUomNames.
//
// An id that does not exist is a SILENT 200: updateUomNames guards on
// `unitOfMeasure != null` and skips the block, so nothing is written and
// nothing is reported. The response is the bound form either way, without the
// list — setUomList runs only on the validation-failure branch.
func (s *UomConfigService) Rename(post form.UomRenameEntryPost) (*form.UomRenameEntryForm, error) {
	f := form.NewUomRenameEntryForm()
	if post.UomID != nil {
		f.UomID = *post.UomID
	}
	if post.NameEnglish != nil {
		f.NameEnglish = *post.NameEnglish
	}

	if f.UomID == "" {
		return &f, nil
	}
	existing, err := s.DAO.GetByID(f.UomID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return &f, nil
	}

	if err := s.DAO.UpdateName(f.UomID, strings.TrimSpace(f.NameEnglish)); err != nil {
		// HibernateException is caught and logged at DEBUG here too.
		return &f, nil
	}

	// getFreshList(UNIT_OF_MEASURE) — the rename path refreshes only that one.
	if err := s.refreshActive(); err != nil {
		return nil, err
	}
	return &f, nil
}
