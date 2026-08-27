package service

import (
	"net/url"
	"strings"

	"openelis-go/internal/auth/daoimpl"
	"openelis-go/internal/auth/session"
)

// AuthzService ports ModuleAuthenticationInterceptor's permission decision.
//
// It is a SEPARATE mechanism from the programmatic role checks Java writes as
// `@PreAuthorize("hasRole('ADMIN')")` on a controller. Both can apply to one
// endpoint, they are evaluated in a fixed order, and they use DIFFERENT
// definitions of "admin". `rest/TestCatalog` is the endpoint that proves it —
// verified live:
//
//	user            module set has TestCatalog?  hasRole('ADMIN')?  Java
//	e2e_reception   no                           no                 401 Not Authorized
//	e2e_testmgmt    yes (Test Management role)   no                 500
//	admin           bypass (is_admin='Y')        yes                200
//
// A port that collapses the two into one check gets one of those three wrong.
type AuthzService struct {
	ModuleDAO *daoimpl.ModuleDAOImpl
}

// PermissionsAgentRole reflects `permissions.agent=Role` in
// SystemConfiguration.properties — verified as the effective value (there is no
// site_information override row in the dev DB).
//
// In Role mode the permitted-module set comes from
// system_user_role → system_role_module; `system_user_module` is never
// consulted (and is empty). The other mode (`isVerifyUserModule`) is not ported;
// if a deployment ever sets permissions.agent to something else, this port
// would silently apply the wrong rule — hence the constant, so the assumption
// is greppable.
const PermissionsAgentRole = "Role"

// HasPermission mirrors ModuleAuthenticationInterceptor.hasPermission for
// permissions.agent=Role:
//
//	hasPermissionForUrl(request, USE_PARAMETERS) || userModuleService.isUserAdmin(request)
//
// Note the admin bypass is `login_user.is_admin='Y'` ALONE — holding the Global
// Administrator ROLE does not bypass the module check.
//
// `contextPath` is the request path with the servlet context prefix already
// removed, e.g. "/rest/TestCatalog". TWO different forms of it are used below,
// and mixing them up silently breaks everything:
//   - the DB lookup uses the /rest-STRIPPED form ("/TestCatalog"), via
//     SystemModuleUrlDAOImpl.getByRequest → URLUtil.getReourcePathFromRequest;
//   - the auto-allow test uses the UN-stripped form, because
//     isRestFullPath() reads the interceptor's own `path` field, which
//     preHandle set to `requestURI - contextPath` with no stripping.
//
// Feeding the stripped path to IsRestFullPath would make auto-allow fail for
// every unmapped endpoint ("/organization-list" does not start with "/rest"),
// denying the entire ported surface.
func (s *AuthzService) HasPermission(p *session.Principal, contextPath string, query url.Values) (bool, error) {
	if p == nil {
		return false, nil
	}
	if p.IsAdmin {
		return true, nil
	}

	mappings, err := s.ModuleDAO.ModuleURLsForPath(ResourcePathFromRequest(contextPath))
	if err != nil {
		return false, err
	}

	// filterParamMatches: a mapping carrying a SystemModuleParam only applies
	// when the request's query parameter equals that value. Java compares
	// against request.getParameter(name), i.e. the raw single value.
	matched := mappings[:0:0]
	for _, m := range mappings {
		if m.ParamName != nil && m.ParamValue != nil {
			if query.Get(*m.ParamName) != *m.ParamValue {
				continue
			}
		}
		matched = append(matched, m)
	}

	if len(matched) == 0 {
		// THE AUTO-ALLOW RULE, reproduced deliberately. Java's in-source note:
		// "REST endpoints without SystemModuleUrl DB entries are auto-allowed
		// for any authenticated user. Admin-only controllers are protected via
		// @PreAuthorize." Unmapped NON-rest paths are denied instead.
		//
		// This is why most ported endpoints need no module data: they have no
		// mapping, so the module check is a no-op for them. rest/TestCatalog is
		// the exception — it IS mapped, which is exactly the case
		// auth-adoption-plan.md §9.2 warned would otherwise make Go more
		// permissive than Java with nothing to flag it.
		return IsRestFullPath(contextPath), nil
	}

	for _, m := range matched {
		if p.Modules[m.ModuleName] {
			return true, nil
		}
	}
	return false, nil
}

// ResourcePathFromRequest ports URLUtil.getReourcePathFromRequest: strip the
// context path (done by the caller), strip a `.do`/`.html` suffix, then strip a
// leading `/rest`.
//
// DIVERGENCE, deliberate and documented (auth-adoption-plan.md §6.5 / §9.3):
// Java strips with `path.split("/rest")[1]`, which
//   - throws ArrayIndexOutOfBoundsException for a path of exactly "/rest",
//   - yields "/a" for "/rest/a/rest/b" (it takes the FIRST segment, dropping
//     everything after the second "/rest"),
//   - and strips "/restore/x" to "ore/x" because the guard is startsWith("/rest").
//
// This port uses TrimPrefix, which matches Java for every path that can
// actually reach it (no ported or mapped url_path is exactly "/rest" or
// contains "/rest" twice; "/restore/..." behaves identically) and returns ""
// instead of panicking on "/rest". Reproducing an exception is not pinning
// behavior, it is copying a crash.
func ResourcePathFromRequest(requestPath string) string {
	path := requestPath
	if i := strings.Index(path, "?"); i >= 0 {
		path = path[:i]
	}
	if strings.Contains(path, ".do") || strings.Contains(path, ".html") {
		if dot := strings.LastIndex(path, "."); dot >= 0 {
			path = path[:dot]
		}
	}
	if strings.HasPrefix(path, "/rest") {
		path = strings.TrimPrefix(path, "/rest")
	}
	return path
}

// IsRestFullPath ports ModuleAuthenticationInterceptor.isRestFullPath, which
// decides both the auto-allow above and the SHAPE of a denial (JSON 401 vs an
// HTML redirect to /Home?access=denied).
//
// It takes the UN-stripped, context-relative path: Java reads the interceptor's
// `path` instance field, which preHandle set to `requestURI - contextPath`
// before any /rest stripping. Do not pass ResourcePathFromRequest's output.
//
// That the field is an instance field on a SINGLETON is a real thread-safety
// bug in Java (§6.4) — under concurrency one request can observe another's
// path and get an HTML redirect where it expected JSON. This port carries the
// path per request instead: a race is not a contract, so it is not pinned.
func IsRestFullPath(contextPath string) bool {
	return strings.HasPrefix(contextPath, "/rest") ||
		strings.HasPrefix(contextPath, "/Provider") ||
		strings.HasPrefix(contextPath, "/dbImage") ||
		strings.HasPrefix(contextPath, "/logging")
}
