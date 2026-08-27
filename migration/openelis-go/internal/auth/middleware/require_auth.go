// Package middleware ports the request-time half of Spring Security's default
// filter chain: the default-deny authorization rule, the principal handoff onto
// context.Context, and CSRF for state-changing verbs.
//
// Structural decision (auth-adoption-plan.md §3.2): protection is applied in
// common/web.Register, the single choke point every ported route already goes
// through, so a NEW ROUTE IS PROTECTED BY DEFAULT. Opening a route requires
// calling web.RegisterOpen explicitly, with a comment naming the Java
// justification. This is the opposite of opt-in protection and is what makes
// "forgot to add auth" a non-event rather than a PHI leak.
package middleware

import (
	"context"
	"net/http"

	"openelis-go/internal/auth/csrf"
	"openelis-go/internal/auth/session"
	"openelis-go/internal/common/web"
)

type ctxKey struct{}

var principalKey ctxKey

// FromContext returns the authenticated principal, if any. Handlers read the
// caller's identity from here; nothing else gets a session handle.
func FromContext(ctx context.Context) (*session.Principal, bool) {
	p, ok := ctx.Value(principalKey).(*session.Principal)
	return p, ok
}

// WithPrincipal attaches a principal (used by the auth middleware and by the
// login controller once a session exists).
func WithPrincipal(ctx context.Context, p *session.Principal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}

// Guard applies the request-time rules to a route. It implements
// web.Protector, which is how common/web.Register makes every ported route
// default-deny without importing this package (that would be a cycle).
//
// A nil Guard or nil Store refuses everything: fail closed, never fail open.
type Guard struct {
	Store session.Store
}

// ServeProtected is the default-deny rule — Java's
// `anyRequest().authenticated()` on the catch-all filter chain (SecurityConfig
// .defaultSecurityConfigurationFilterChain, which has NO securityMatcher) —
// followed by the CSRF check for state-changing verbs.
//
// The refusal is a 302 to <context>/LoginPage, exactly what Spring's
// LoginUrlAuthenticationEntryPoint emits — verified live for every protected
// path. It is deliberately NOT "corrected" into a 401 JSON: migration policy is
// to pin Java's observable behavior and raise bugs separately.
//
// A dead or unknown session id gets the SAME redirect, matching
// sessionManagement().invalidSessionUrl("/LoginPage") — which in Java intercepts
// even permitAll paths, so a client presenting a logged-out JSESSIONID is
// redirected rather than told `authenticated:false`.
func (g *Guard) ServeProtected(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	p, ok := g.principal(r)
	if !ok {
		session.Redirect(w, r, "/LoginPage")
		return
	}
	r = r.WithContext(WithPrincipal(r.Context(), p))

	// CSRF applies to state-changing verbs only, matching Spring's CsrfFilter
	// (which exempts GET/HEAD/OPTIONS/TRACE). The token arrives masked on
	// X-CSRF-TOKEN — header lookup is case-insensitive in net/http — or as the
	// `_csrf` form parameter; csrf.Valid un-masks before comparing.
	if !isSafeMethod(r.Method) && !csrf.Valid(p.CSRFToken, submittedToken(r)) {
		DenyAccess(w, r, "CSRF token missing or invalid")
		return
	}

	next(w, r)
}

// RequireRole wraps a handler with a programmatic role check, the analog of
// Java's in-controller guards (e.g.
// PatientMergeRestController.hasMergePermission, which requires "Reception"
// and returns a BODILESS 403 — ResponseEntity.status(FORBIDDEN).build()).
//
// Admin does NOT bypass these: the controller-level checks in Java call
// userRoleService.userInRole directly, with no isUserAdmin fallback. That
// fallback exists only in ModuleAuthenticationInterceptor.hasPermission, which
// governs the module system — a separate mechanism, and one no ported endpoint
// currently exercises (none of the ported paths has a system_module_url row, so
// its unmapped-/rest auto-allow covers them all). See auth-adoption-plan.md
// §2.7.
func RequireRole(role string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, ok := FromContext(r.Context())
		if !ok || !p.HasRole(role) {
			// Bodiless, matching ResponseEntity.status(403).build().
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

// principal resolves the session the request presents.
func (g *Guard) principal(r *http.Request) (*session.Principal, bool) {
	if g == nil || g.Store == nil {
		return nil, false
	}
	p, ok := g.Store.Get(session.IDFromRequest(r))
	if !ok || p == nil {
		// A session row with a nil principal is an anonymous bootstrap session
		// (minted by GET /session so the client has an id). It authenticates
		// nothing.
		return nil, false
	}
	return p, true
}

func isSafeMethod(m string) bool {
	switch m {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

// submittedToken reads the CSRF token from the header Spring's default
// CsrfTokenRequestAttributeHandler looks at, then the form parameter.
func submittedToken(r *http.Request) string {
	if v := r.Header.Get("X-CSRF-TOKEN"); v != "" {
		return v
	}
	if err := r.ParseForm(); err == nil {
		return r.PostFormValue("_csrf")
	}
	return ""
}

// DenyAccess reproduces SecurityConfig's accessDeniedHandler, including its
// path-dependent shape.
func DenyAccess(w http.ResponseWriter, r *http.Request, message string) {
	if isRestPath(r.URL.Path) {
		web.WriteJSON(w, http.StatusForbidden, map[string]any{
			"status":  403,
			"message": message,
		})
		return
	}
	session.Redirect(w, r, "/Home?access=denied")
}

// isRestPath mirrors the accessDeniedHandler's own test: the path (after the
// context path is stripped) starts with /rest or /Provider — plus the literal
// "/api/OpenELIS-Global/rest" the Java handler also checks, because it compares
// against the request URI when the app is proxied.
func isRestPath(path string) bool {
	rel := path
	if len(path) > len(session.ContextPath) && path[:len(session.ContextPath)] == session.ContextPath {
		rel = path[len(session.ContextPath):]
	}
	return hasPrefix(rel, "/rest") || hasPrefix(rel, "/Provider")
}

func hasPrefix(s, p string) bool { return len(s) >= len(p) && s[:len(p)] == p }
