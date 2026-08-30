package service

import (
	"encoding/json"
	"errors"
	"regexp"

	commondaoimpl "openelis-go/internal/common/daoimpl"
	"openelis-go/internal/testconfiguration/daoimpl"
	"openelis-go/internal/testconfiguration/form"
)

// OrderService ports PanelOrderRestController, SampleTypeOrderRestController
// and TestSectionOrderRestController — three screens that reorder a reference
// list and write nothing else.
type OrderService struct {
	Lists *commondaoimpl.DisplayListDAOImpl
	DAO   *daoimpl.OrderDAO
}

// activateSet is Java's ActivateSet: one id and the sort order it moves to.
type activateSet struct {
	ID        json.Number `json:"id"`
	SortOrder json.Number `json:"sortOrder"`
	// Activated is REQUIRED by TestActivationFormValidator —
	// validateField(..., true, 5, "^$|^true$|^false$") — and is not read by the
	// handler at all. Omitting it refuses the whole request, at 200 with no
	// write. The *Order screens do not have it.
	Activated *string `json:"activated"`
}

// ErrChangeListShape is the ClassCastException Java throws when the value under
// the screen's key is a real ARRAY rather than a string holding one.
//
// `String action = (String) root.get(key)` sits OUTSIDE the try block that
// catches ParseException, so nothing catches it and the request ends as a 500.
// Measured on all three screens: the shape a reasonable client would send —
// a nested array — is the one that fails.
var ErrChangeListShape = errors.New("testconfiguration: change list is not double-encoded")

// parseChangeList unpacks jsonChangeList, which is DOUBLE-ENCODED.
//
// The submitted field is a JSON string. Parsing it yields an object whose value
// under the screen's key is ITSELF a JSON string, parsed again to reach the
// array of {id, sortOrder}.
//
// An outer body that will not parse is NOT an error: JSONUtils.getAsObject
// yields nothing, the validator's isEmpty check passes it through, and
// getActivateSetForActions returns an empty list — a 200 that wrote nothing.
// An inner value of the wrong TYPE is a different outcome; see above.
func parseChangeList(changeList, key string) ([]activateSet, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal([]byte(changeList), &root); err != nil {
		return nil, nil
	}
	raw, ok := root[key]
	if !ok {
		// A missing key is `parser.parse(null)` — NullPointerException, not the
		// ParseException the catch is written for. 500.
		return nil, ErrChangeListShape
	}
	var inner string
	if err := json.Unmarshal(raw, &inner); err != nil {
		return nil, ErrChangeListShape
	}
	var sets []activateSet
	if err := json.Unmarshal([]byte(inner), &sets); err != nil {
		return nil, nil
	}
	return sets, nil
}

// sortOrderFormat is the validator's rule for the sort order:
// validateFieldAndCharset(..., 3, "0-9") — at most THREE digits, and digits
// only. A four-digit order is refused, and a refusal is a 200 that writes
// nothing rather than an error status.
var sortOrderFormat = regexp.MustCompile(`^[0-9]{1,3}$`)

// idFormat is ValidationHelper.validateIdField.
var idFormat = regexp.MustCompile(`^[0-9]+$`)

// orderScreen is one screen's table, key, audit table and payload field.
type orderScreen struct {
	key        string
	table      string
	auditTable string
	// payloadField is the BEAN PROPERTY the audit payload names, not the
	// column. Panel and TestSection call setSortOrderInt and emit
	// `<sortOrderInt>`; TypeOfSample calls setSortOrder and emits
	// `<sortOrder>`. One column, two names in the history — the same
	// declared-field rule e1 pinned for the delete payload.
	payloadField string
}

var orderScreens = map[string]orderScreen{
	"PanelOrder":       {"panels", "clinlims.panel", "PANEL", "sortOrderInt"},
	"SampleTypeOrder":  {"sampleTypes", "clinlims.type_of_sample", "TYPE_OF_SAMPLE", "sortOrder"},
	"TestSectionOrder": {"testSections", "clinlims.test_section", "TEST_SECTION", "sortOrderInt"},
}

// PanelOrderForm ports showPanelOrder.
func (s *OrderService) PanelOrderForm() (*form.OrderForm, error) {
	panels, err := s.Lists.Panels()
	if err != nil {
		return nil, err
	}
	f := form.NewOrderForm("panelOrderForm")
	f.PanelList = &panels
	return &f, nil
}

// SampleTypeOrderForm ports showSampleTypeOrder — SAMPLE_TYPE, which is every
// human type rather than only the active ones.
func (s *OrderService) SampleTypeOrderForm() (*form.OrderForm, error) {
	types, err := s.Lists.AllSampleTypes()
	if err != nil {
		return nil, err
	}
	f := form.NewOrderForm("sampleTypeOrderForm")
	f.SampleTypeList = &types
	return &f, nil
}

// TestSectionOrderForm ports showTestSectionOrder.
func (s *OrderService) TestSectionOrderForm() (*form.OrderForm, error) {
	sections, err := s.Lists.ActiveTestSections()
	if err != nil {
		return nil, err
	}
	f := form.NewOrderForm("testSectionOrderForm")
	f.TestSectionList = &sections
	return &f, nil
}

// Reorder applies one screen's change list.
//
// Java loads each entity, sets its sort order and hands the whole list to
// updateAll, so the writes share one transaction and each moved row leaves an
// audit entry carrying the order it replaced.
func (s *OrderService) Reorder(screen string, post form.OrderPost, sysUserID int64) (*form.OrderForm, error) {
	sc := orderScreens[screen]
	f := form.NewOrderForm(form.OrderFormNames[screen])
	if post.JSONChangeList != nil {
		f.JSONChangeList = *post.JSONChangeList
	}

	sets, err := parseChangeList(f.JSONChangeList, sc.key)
	if err != nil {
		return nil, err
	}
	if len(sets) == 0 {
		return &f, nil
	}

	// The validator runs over the WHOLE list before anything is written, and a
	// single bad entry refuses the request — so a change list with one
	// four-digit order moves none of its rows.
	ids := make([]string, 0, len(sets))
	orders := make([]int, 0, len(sets))
	for _, set := range sets {
		id, order := set.ID.String(), set.SortOrder.String()
		if !idFormat.MatchString(id) || !sortOrderFormat.MatchString(order) {
			return &f, nil
		}
		n, convErr := set.SortOrder.Int64()
		if convErr != nil {
			return &f, nil
		}
		ids = append(ids, id)
		orders = append(orders, int(n))
	}

	if err := s.DAO.Reorder(sc.table, sc.auditTable, sc.payloadField, ids, orders, sysUserID); err != nil {
		// HibernateException is caught and logged at DEBUG; the form comes back
		// as though the write had worked.
		return &f, nil
	}
	return &f, nil
}
