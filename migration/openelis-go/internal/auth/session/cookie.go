package session

import (
	"net/http"
	"strings"
)

// CookieName must stay JSESSIONID: the React frontend and the e2e suite send
// back whatever the server set, and the nginx proxy forwards it unchanged, so
// renaming it would silently break every existing client.
const CookieName = "JSESSIONID"

// ContextPath is the proxied servlet context the Java WAR is deployed under.
// The Go service answers on BOTH this prefix and the bare path (see
// common/web.Register), so anything path-scoped — the cookie, a redirect —
// has to know which one the current request came in on.
const ContextPath = "/api/OpenELIS-Global"

// CookiePath returns the Path attribute for the session cookie, matching
// Tomcat: the cookie is scoped to the context the request was served from.
// Verified live on Java: `Set-Cookie: JSESSIONID=…; Path=/api/OpenELIS-Global`.
func CookiePath(r *http.Request) string {
	if InContext(r) {
		return ContextPath
	}
	return "/"
}

// InContext reports whether the request arrived on the proxied context path.
func InContext(r *http.Request) bool {
	return r.URL.Path == ContextPath || strings.HasPrefix(r.URL.Path, ContextPath+"/")
}

// Redirect sends the same 302 Java sends, resolved against whichever prefix the
// request came in on. `target` is context-relative and starts with "/", e.g.
// "/LoginPage".
//
// Deliberately NOT http.Redirect: for a GET whose client accepts HTML, that
// helper appends a short `<a href="…">Found</a>.` body. Java sends
// Content-Length: 0 on every one of these — measured on the anonymous-access
// redirect, the csrf-less logout denial and the successful logout alike — and
// on a PHI endpoint "the refusal carries no body" is worth being exact about.
// Writing the header directly keeps the response byte-identical.
func Redirect(w http.ResponseWriter, r *http.Request, target string) {
	prefix := ""
	if InContext(r) {
		prefix = ContextPath
	}
	w.Header().Set("Location", prefix+target)
	w.WriteHeader(http.StatusFound)
}

// SetCookie issues the session cookie.
//
// HttpOnly and Secure are UNCONDITIONAL, matching web.xml's
// <cookie-config><http-only>true</http-only><secure>true</secure> — which Java
// applies regardless of transport. Deliberately not gated on a dev flag: a flag
// that turns Secure off is a flag that can ship. Browsers and Playwright both
// treat http://localhost as a trustworthy origin, so the loopback parity runs
// still receive it.
//
// SameSite=Lax is an addition, not a divergence in observable behavior: Tomcat
// sets no SameSite attribute, so browsers apply their own Lax default anyway.
// Stating it explicitly removes the dependency on that default.
func SetCookie(w http.ResponseWriter, r *http.Request, id string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    id,
		Path:     CookiePath(r),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// IDFromRequest returns the session id the client presented, or "".
func IDFromRequest(r *http.Request) string {
	c, err := r.Cookie(CookieName)
	if err != nil {
		return ""
	}
	return c.Value
}
