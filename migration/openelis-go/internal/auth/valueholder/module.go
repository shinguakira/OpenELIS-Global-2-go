package valueholder

// RoleGlobalAdmin mirrors Constants.ROLE_GLOBAL_ADMIN.
//
// Two DIFFERENT admin notions exist in Java and they must not be collapsed —
// the e2e fixture seeds a user (`e2e_testmgmt`) that separates them:
//
//   - The MODULE check's bypass is `login_user.is_admin = 'Y'` ALONE
//     (UserModuleServiceImpl.isUserAdmin). Holding the Global Administrator
//     role does not bypass it.
//   - Spring's `hasRole('ADMIN')` is granted when `is_admin='Y'` OR the user
//     holds this role (CustomUserDetailsService.addAuthoritiesForRole adds
//     "ROLE_ADMIN" for it).
const RoleGlobalAdmin = "Global Administrator"

// SystemModuleURL mirrors systemmodule.valueholder.SystemModuleUrl joined to
// SystemModule and its optional SystemModuleParam. Maps the module-permission
// lookup that ModuleAuthenticationInterceptor performs per request.
//
// Note the asymmetry with Role: `system_module.name` is `character varying`,
// NOT the blank-padded `character(30)` that `system_role.name` is — so this one
// needs no trim. Verified against the live schema; do not assume either way.
type SystemModuleURL struct {
	URLPath    string
	ModuleName string
	// ParamName/ParamValue come from the optional system_module_param row. When
	// present, the mapping only applies if the request carries that exact query
	// parameter value (ModuleAuthenticationInterceptor.filterParamMatches).
	// Used by e.g. /ElisaAlgorithmWorkplan, which maps to a different module
	// per ?type= value.
	ParamName  *string
	ParamValue *string
}
