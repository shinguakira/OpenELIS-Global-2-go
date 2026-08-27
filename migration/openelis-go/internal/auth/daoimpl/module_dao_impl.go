package daoimpl

import (
	"gorm.io/gorm"

	"openelis-go/internal/auth/valueholder"
)

// ModuleDAOImpl ports the two reads ModuleAuthenticationInterceptor performs:
// SystemModuleUrlDAOImpl.getByUrlPath (per request) and
// RoleModuleServiceImpl.getAllPermittedPagesFromAgentId (once, at login).
type ModuleDAOImpl struct {
	DB *gorm.DB
}

// PermittedModulesForUser returns the module names a user may access, via
// system_user_role → system_role_module → system_module.
//
// Mirrors CustomFormAuthenticationSuccessHandler.getPermittedForms, which loops
// the user's roles and unions each role's modules. Java runs this ONCE at login
// and caches the result in the session (IActionConstants.PERMITTED_ACTIONS_MAP);
// this port does the same by putting it on the Principal — so a role granted
// mid-session does not take effect until the next login, exactly as in Java.
//
// No has_select/has_add/... filtering: RoleModuleDAOImpl's query is a bare
// `from RoleModule s where s.role.id = :param`, and getAllPermittedPagesFromAgentId
// takes every row's module name. Adding a permission-flag filter here would be
// stricter than Java, which is still a behavior change.
func (d *ModuleDAOImpl) PermittedModulesForUser(systemUserID int64) (map[string]bool, error) {
	var names []string
	err := d.DB.
		Table("clinlims.system_user_role AS sur").
		Joins("JOIN clinlims.system_role_module AS srm ON srm.system_role_id = sur.role_id").
		Joins("JOIN clinlims.system_module AS sm ON sm.id = srm.system_module_id").
		Where("sur.system_user_id = ?", systemUserID).
		Pluck("sm.name", &names).Error
	if err != nil {
		return nil, err
	}
	// A set, matching Java's HashSet — duplicates across roles collapse.
	out := make(map[string]bool, len(names))
	for _, n := range names {
		out[n] = true
	}
	return out, nil
}

// ModuleURLsForPath mirrors SystemModuleUrlDAOImpl.getByUrlPath: an EXACT match
// on url_path (no prefix, no LIKE), returning every mapping for that path along
// with its optional parameter constraint.
//
// Zero rows is the common case and is meaningful, not an error: it triggers the
// interceptor's auto-allow rule for /rest paths.
func (d *ModuleDAOImpl) ModuleURLsForPath(urlPath string) ([]valueholder.SystemModuleURL, error) {
	var rows []valueholder.SystemModuleURL
	err := d.DB.
		Table("clinlims.system_module_url AS u").
		Select("u.url_path AS url_path, m.name AS module_name, p.name AS param_name, p.value AS param_value").
		Joins("JOIN clinlims.system_module AS m ON m.id = u.system_module_id").
		Joins("LEFT JOIN clinlims.system_module_param AS p ON p.id = u.system_module_param_id").
		Where("u.url_path = ?", urlPath).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// PermissionsAgentOverride reads a `site_information` row named
// "permissions.agent", if one exists.
//
// This is the one configuration source Java and Go share. It is NOT the whole
// picture — see service.EffectivePermissionsAgent for why the property files on
// the Java container outrank it and why an env var exists as well.
//
// Returns ("", false, nil) when no row is present, which is the normal case:
// the WAR's SystemConfiguration.properties default then applies.
func (d *ModuleDAOImpl) PermissionsAgentOverride() (string, bool, error) {
	var values []string
	err := d.DB.
		Table("clinlims.site_information").
		Where("name = ?", "permissions.agent").
		Pluck("value", &values).Error
	if err != nil {
		return "", false, err
	}
	if len(values) == 0 {
		return "", false, nil
	}
	return values[0], true, nil
}
