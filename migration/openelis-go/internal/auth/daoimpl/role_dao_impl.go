package daoimpl

import (
	"strings"

	"gorm.io/gorm"
)

// RoleDAOImpl ports the userrole/role read paths the auth flow uses:
// UserRoleServiceImpl.getRoleIdsForUser + RoleServiceImpl.getRoleById, which
// Java calls in a loop (LoginPageController.setLabunitRolesForExistingUser).
// One join replaces the N+1 — same rows, same names.
type RoleDAOImpl struct {
	DB *gorm.DB
}

// RolesForUser returns the TRIMMED role names granted to a system user via
// system_user_role → system_role.
//
// The trim is the whole point: system_role.name is character(30) and comes back
// blank-padded (see valueholder.Role). Java trims at every comparison site
// (`roleService.getRoleById(roleId).getName().trim()`); this port trims once,
// here, so no caller can forget.
//
// Ordering: Java collects into a HashSet, so it has NO defined order and the
// /session DTO's `roles` is an unordered set. Sorting here would invent a
// guarantee Java does not make — but an unordered result is also untestable, so
// the ORDER BY is deliberate and documented as a divergence in shape only, not
// in content. The e2e oracle compares role SETS, never positions.
func (d *RoleDAOImpl) RolesForUser(systemUserID int64) ([]string, error) {
	var names []string
	err := d.DB.
		Table("clinlims.system_user_role AS sur").
		Select("sr.name").
		Joins("JOIN clinlims.system_role AS sr ON sr.id = sur.role_id").
		Where("sur.system_user_id = ?", systemUserID).
		Order("sr.id").
		Pluck("sr.name", &names).Error
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(names))
	for _, n := range names {
		out = append(out, strings.TrimSpace(n))
	}
	return out, nil
}
