package service

import (
	"fmt"
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

// PermissionsAgentRole is the only `permissions.agent` mode this port
// implements. It is the effective value on the dev stack and the WAR default in
// SystemConfiguration.properties.
//
// In Role mode the permitted-module set comes from
// system_user_role → system_role_module. The OTHER mode
// (UserModuleServiceImpl.isVerifyUserModule) reads `system_user_module` —
// modules assigned DIRECTLY to a user — and matches them against the action NAME
// by prefix instead of doing a system_module_url lookup. Different logic, not a
// variation, and not ported.
const PermissionsAgentRole = "Role"

// EffectivePermissionsAgent resolves which authorization model to run and
// REFUSES anything this port does not implement, so a misconfigured deployment
// fails at startup instead of silently applying the wrong rules — granting
// access Java denies, or denying access Java grants, with nothing to flag it.
//
// Resolution, highest precedence first:
//
//  1. `env` — the OE_PERMISSIONS_AGENT environment variable.
//  2. `dbOverride` — a `site_information` row named "permissions.agent".
//     DefaultConfigurationProperties.loadFromDatabase stores rows whose name is
//     not a Property enum under their raw name, so such a row really would be
//     picked up by Java's getPropertyValue("permissions.agent").
//  3. PermissionsAgentRole, the WAR default.
//
// WHY AN ENV VAR IS NEEDED AT ALL — the honest limit of this check. On the Java
// side the files under /var/lib/openelis-global/properties/ OUTRANK
// site_information (DefaultConfigurationProperties.init: the change-value file
// is copied in "prefer source", i.e. it overwrites). Those files live on the
// Java container's filesystem, which the Go service cannot read. So the DB row
// is the only shared source, and checking it alone would be a partial check
// dressed up as a complete one. The env var is the operator's way to tell the
// port what Java actually resolved.
func EffectivePermissionsAgent(env, dbOverride string) (string, error) {
	agent := strings.TrimSpace(env)
	if agent == "" {
		agent = strings.TrimSpace(dbOverride)
	}
	if agent == "" {
		return PermissionsAgentRole, nil
	}
	// Java compares with equalsIgnoreCase, so a deployment spelling it "ROLE"
	// is configured correctly and must start.
	if !strings.EqualFold(agent, PermissionsAgentRole) {
		return "", fmt.Errorf(
			"permissions.agent is %q, but this service only implements the %q model;"+
				" the other model (system_user_module direct assignment) is not ported."+
				" Refusing to start rather than applying the wrong authorization rules",
			agent, PermissionsAgentRole)
	}
	return PermissionsAgentRole, nil
}

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
