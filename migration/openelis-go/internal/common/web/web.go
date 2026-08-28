// Package web holds shared HTTP plumbing (mirrors the role of
// org.openelisglobal.common). This is migration-time scaffolding; the idiomatic
// Go reorganization happens at the end of the migration.
package web

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

// ContentTypeJSON is the Content-Type Java sends on every JSON response —
// verified live across ported reads, /session, login success, login failure and
// both authorization denial shapes. The charset parameter is what the servlet
// API's response.setCharacterEncoding("UTF-8") produces; there is no separate
// encoding header, and inventing one (`Character-Encoding`) declares nothing to
// any client or proxy.
const ContentTypeJSON = "application/json;charset=UTF-8"

// WriteJSON writes v as a JSON response — the analog of Spring returning
// ResponseEntity with produces=APPLICATION_JSON_VALUE.
//
// Two Go-vs-Jackson byte differences are corrected here, both invisible to any
// comparison that parses the JSON first:
//
//   - SetEscapeHTML(false) matches Jackson, which does NOT HTML-escape: Java
//     emits ">=", "<=", "&&" literally, whereas Go's encoder escapes them.
//   - Encode appends a trailing newline and Jackson does not, so every
//     response was one byte longer than the Java baseline with a different
//     Content-Length. Encoding into a buffer and trimming it is the only way
//     to keep SetEscapeHTML (json.Marshal has no equivalent switch).
//
// See util.JavaDouble for the third: Go renders a float64 5.0 as `5`.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		log.Printf("WriteJSON: encode failed: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", ContentTypeJSON)
	w.WriteHeader(status)
	_, _ = w.Write(bytes.TrimSuffix(buf.Bytes(), []byte{0x0a}))
}

// ── Default-deny wiring ─────────────────────────────────────────────────────
//
// Java's catch-all filter chain has NO securityMatcher and ends in
// `anyRequest().authenticated()` (SecurityConfig
// .defaultSecurityConfigurationFilterChain), so in Java a new endpoint is
// protected the moment it exists. Reproducing that property — not just the
// individual rules — is the single most important structural decision in the
// auth port: it makes "forgot to add auth" a non-event rather than a PHI leak.
//
// Register is the one choke point every ported route already goes through, so
// protection is applied THERE. Opening a route requires calling RegisterOpen by
// name, with a comment citing the Java rule that justifies it.
//
// The Protector is injected at startup rather than imported, because the
// middleware that implements it needs WriteJSON from this package — importing
// it back would be a cycle.

// Protector guards a route. ServeProtected either serves next (authorized) or
// writes the refusal itself.
type Protector interface {
	ServeProtected(w http.ResponseWriter, r *http.Request, next http.HandlerFunc)
}

// protector is read on every protected request, so routes may be registered
// before or after UseProtector without changing the outcome.
var protector atomic.Pointer[Protector]

// UseProtector installs the guard. Call it during startup, before serving.
func UseProtector(p Protector) { protector.Store(&p) }

// unconfiguredProtector is the FAIL-CLOSED default. If auth was never wired
// (say the DB never came up and startup took a degraded path), protected routes
// must refuse rather than serve data anonymously. There is no code path that
// should reach this in a healthy process, so it logs.
type unconfiguredProtector struct{}

func (unconfiguredProtector) ServeProtected(w http.ResponseWriter, r *http.Request, _ http.HandlerFunc) {
	log.Printf("SECURITY: no Protector installed; refusing %s %s", r.Method, r.URL.Path)
	http.Error(w, "authentication is not configured", http.StatusInternalServerError)
}

func currentProtector() Protector {
	if p := protector.Load(); p != nil {
		return *p
	}
	return unconfiguredProtector{}
}

// Register mounts h at both the full proxied path
// (/api/OpenELIS-Global/<restPath>) and the bare /<restPath>, so the nginx
// proxy can forward a single path here unchanged.
//
// The handler is PROTECTED: unauthenticated callers are refused, and
// state-changing verbs additionally require a valid CSRF token. Use
// RegisterOpen for the handful of paths Java itself leaves anonymous.
func Register(mux *http.ServeMux, method, restPath string, h http.HandlerFunc) {
	guarded := func(w http.ResponseWriter, r *http.Request) {
		currentProtector().ServeProtected(w, r, h)
	}
	mux.HandleFunc(method+" /api/OpenELIS-Global/"+restPath, guarded)
	mux.HandleFunc(method+" /"+restPath, guarded)
}

// RegisterOpen mounts h WITHOUT authentication.
//
// Every call site must carry a comment naming the Java rule that makes the path
// anonymous — an entry in SecurityConfig.OPEN_PAGES or LOGIN_PAGES. If you
// cannot cite one, the route belongs in Register.
func RegisterOpen(mux *http.ServeMux, method, restPath string, h http.HandlerFunc) {
	mux.HandleFunc(method+" /api/OpenELIS-Global/"+restPath, h)
	mux.HandleFunc(method+" /"+restPath, h)
}

// ServletErrorBody is the body Tomcat's default error page produces for an
// unhandled exception — the shape Java answers with when a controller throws,
// as opposed to the RFC 7807 ProblemDetail Spring produces for a BINDING
// failure. Two error paths, two envelopes, and c3 hits both.
//
//	{"timestamp":1787922187266,"status":500,"error":"Internal Server Error"}
type ServletErrorBody struct {
	Timestamp int64  `json:"timestamp"`
	Status    int    `json:"status"`
	Error     string `json:"error"`
}

// ServletError builds that body. The timestamp is epoch MILLIS, taken at
// response time exactly as Tomcat does.
func ServletError(status int) ServletErrorBody {
	return ServletErrorBody{
		Timestamp: nowMillis(),
		Status:    status,
		Error:     http.StatusText(status),
	}
}

// nowMillis is epoch milliseconds, the unit Tomcat stamps its error body with.
func nowMillis() int64 {
	return time.Now().UnixMilli()
}
