package service

import (
	"math"

	"openelis-go/internal/common/util"

	"openelis-go/internal/testcatalog/daoimpl"
	"openelis-go/internal/testcatalog/form"
)

// The Reference Ranges section. 🔴 Clinical: these rows decide whether a result
// reads as normal, so the validation below is the only thing between a typo and
// a silently wrong flag on a patient's report.

// rangeGenders is the controller's allow-list. Note what is NOT in it: the
// BLANK gender, which is accepted and means "both" — the coverage validator
// counts such a range toward each sex.
var rangeGenders = map[string]bool{"M": true, "F": true}

// GetRanges ports getRanges.
func (s *EditorTestService) GetRanges(testID string) (*form.RangesResponse, error) {
	row, err := s.DAO.BasicInfo(testID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrNotFound
	}
	return s.rangesResponse(testID)
}

// SaveRanges ports saveRanges.
//
// Validation runs over EVERY range before a single write, so one bad row makes
// the whole save a 422 that changed nothing. The rules are narrow — a known
// gender or none, a non-negative minimum, and a maximum strictly above it — and
// they do not check the VALUE bounds at all, so a range whose low normal is
// above its high normal is accepted.
func (s *EditorTestService) SaveRanges(testID string, body form.RangesResponse, sysUserID int64) (*form.RangesResponse, error) {
	row, err := s.DAO.BasicInfo(testID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrNotFound
	}
	if err := validateRanges(body.Ranges); err != nil {
		return nil, err
	}
	if err := s.DAO.SaveRanges(testID, toDesiredRanges(body.Ranges), sysUserID); err != nil {
		return nil, err
	}
	return s.rangesResponse(testID)
}

// SaveGroupRanges ports saveGroupRanges.
//
// The same ranges are written to every test named — and the incoming IDS ARE
// DROPPED, because an id in a shared set belongs to no single test. So a group
// save always INSERTS on each test and deletes whatever numeric ranges those
// tests had. Running it twice therefore replaces the rows it created the first
// time rather than updating them.
func (s *EditorTestService) SaveGroupRanges(body form.GroupRangesUpdate, sysUserID int64) error {
	if len(body.TestIDs) == 0 {
		return ErrUnprocessable
	}
	if err := validateRanges(body.Ranges); err != nil {
		return err
	}
	desired := toDesiredRanges(body.Ranges)
	for i := range desired {
		desired[i].ID = ""
	}
	for _, testID := range body.TestIDs {
		row, err := s.DAO.BasicInfo(testID)
		if err != nil {
			return err
		}
		if row == nil {
			continue
		}
		if err := s.DAO.SaveRanges(testID, desired, sysUserID); err != nil {
			return err
		}
	}
	return nil
}

func validateRanges(ranges []form.RangeDTO) error {
	for _, r := range ranges {
		if r.Gender != nil && *r.Gender != "" && !rangeGenders[*r.Gender] {
			return ErrUnprocessable
		}
		min := 0.0
		if r.MinAge != nil {
			min = float64(*r.MinAge)
		}
		max := math.Inf(1)
		if r.MaxAge != nil {
			max = float64(*r.MaxAge)
		}
		if min < 0 || max <= min {
			return ErrUnprocessable
		}
	}
	return nil
}

// toDesiredRanges ports toResultLimits: an absent bound takes the ENTITY
// default, which is where the low-critical asymmetry comes from — it defaults
// to POSITIVE infinity, not negative.
func toDesiredRanges(ranges []form.RangeDTO) []daoimpl.DesiredRange {
	out := make([]daoimpl.DesiredRange, 0, len(ranges))
	for _, r := range ranges {
		out = append(out, daoimpl.DesiredRange{
			ID:           r.ID,
			ComponentID:  blankToNil(r.ComponentID),
			Gender:       blankToNil(r.Gender),
			MinAge:       unbox(r.MinAge, 0),
			MaxAge:       unbox(r.MaxAge, math.Inf(1)),
			LowNormal:    unbox(r.LowNormal, math.Inf(-1)),
			HighNormal:   unbox(r.HighNormal, math.Inf(1)),
			LowCritical:  unbox(r.LowCritical, math.Inf(1)),
			HighCritical: unbox(r.HighCritical, math.Inf(1)),
			LowValid:     unbox(r.LowValid, math.Inf(-1)),
			HighValid:    unbox(r.HighValid, math.Inf(1)),
		})
	}
	return out
}

func (s *EditorTestService) rangesResponse(testID string) (*form.RangesResponse, error) {
	rows, err := s.DAO.Ranges(testID)
	if err != nil {
		return nil, err
	}
	resp := &form.RangesResponse{TestID: testID, Ranges: []form.RangeDTO{}}
	limits := make([]daoimpl.LimitAge, 0, len(rows))
	for _, r := range rows {
		resp.Ranges = append(resp.Ranges, form.RangeDTO{
			ID: r.ID, ComponentID: r.ComponentID, Gender: r.Gender,
			MinAge: finiteOrNil(r.MinAge), MaxAge: finiteOrNil(r.MaxAge),
			LowNormal: finiteOrNil(r.LowNormal), HighNormal: finiteOrNil(r.HighNormal),
			LowCritical: finiteOrNil(r.LowCritical), HighCritical: finiteOrNil(r.HighCritical),
			LowValid: finiteOrNil(r.LowValid), HighValid: finiteOrNil(r.HighValid),
			LowReporting: finiteOrNil(r.LowReporting), HighReporting: finiteOrNil(r.HighReporting),
		})
		gender := ""
		if r.Gender != nil {
			gender = *r.Gender
		}
		limits = append(limits, daoimpl.LimitAge{Gender: gender, MinAge: r.MinAge, MaxAge: r.MaxAge})
	}
	// The coverage report counts EVERY limit, including the dictionary ones the
	// editor cannot manage — so a dictionary test's reference row participates
	// in the gate for a section it does not belong to.
	resp.Coverage = validateCoverage(limits)
	return resp, nil
}

// finiteOrNil is the controller's finiteOrNull: ±Infinity and NaN become null
// so the bound serialises as JSON, which means "unbounded" and "unset" are the
// same thing on the wire.
func finiteOrNil(v float64) *util.JavaDouble {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return nil
	}
	out := util.JavaDouble(v)
	return &out
}

func unbox(v *util.JavaDouble, dflt float64) float64 {
	if v == nil {
		return dflt
	}
	return float64(*v)
}

func blankToNil(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}
