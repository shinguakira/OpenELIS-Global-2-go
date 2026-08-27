// Package session holds the server-side session store and the JSESSIONID
// cookie mechanics. It stands in for the servlet container's HttpSession, which
// is what Java uses (there is no session replication configured on the Java
// side either — see migration/auth-adoption-plan.md §3.3).
//
// KNOWN LIMITATION, stated up front: an in-memory store means sessions die on
// restart and are not shared across replicas — and, during strangler
// coexistence, a user authenticated against Java is NOT authenticated against
// Go. That is parity with Java's own limitation, not a regression, but it is
// why the Go service stays loopback-bound until a whole path group is cut over.
// The Store interface exists so a Redis/DB-backed implementation can replace
// this one without touching any caller.
package session

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	"openelis-go/internal/auth/valueholder"
)

// Principal is the authenticated identity carried on context.Context, per
// tech-stack-diff.md's prescription ("custom middleware + context.Context
// principal"). It merges Java's two overlapping notions — the Spring
// SecurityContext and the `userSessionData` session attribute — into one value.
//
// Keeping them merged closes trap §6.3: in Java, `authenticated` is
// `userSessionData != null`, NOT "has a SecurityContext", so a session can hold
// a valid SecurityContext and still report authenticated:false. With a single
// principal in a single session, that divergence cannot arise.
type Principal struct {
	SystemUserID int64
	LoginName    string
	FirstName    string
	LastName     string
	// IsAdmin is `login_user.is_admin = 'Y'` and NOTHING else. It is the
	// module check's bypass (UserModuleServiceImpl.isUserAdmin) — which the
	// Global Administrator role does NOT confer. For Spring's
	// hasRole('ADMIN'), use HasAdminAuthority instead.
	IsAdmin bool
	// Roles holds TRIMMED role names (see auth/valueholder.Role).
	Roles []string
	// Modules is the permitted-module set, unioned across the user's roles at
	// LOGIN — mirroring Java, which computes it once in the success handler and
	// caches it in the session as PERMITTED_ACTIONS_MAP. A role granted
	// mid-session therefore takes effect only at the next login, in both.
	Modules map[string]bool
	// CSRFToken is the RAW per-session token. It is never sent as-is — every
	// value handed to a client goes through csrf.Mask.
	CSRFToken string
}

// HasRole reports whether the principal holds a role, comparing the way Java
// does: trimmed and case-insensitively (UserRoleServiceImpl.userInRole).
func (p *Principal) HasRole(name string) bool {
	want := strings.TrimSpace(name)
	for _, r := range p.Roles {
		if strings.EqualFold(strings.TrimSpace(r), want) {
			return true
		}
	}
	return false
}

// HasAdminAuthority reproduces Spring's `hasRole('ADMIN')`, the expression
// every `@PreAuthorize("hasRole('ADMIN')")` in the Java controllers uses.
//
// CustomUserDetailsService.getGrantedAuthorities adds the ROLE_ADMIN authority
// in two independent cases: the user holds the Global Administrator role
// (addAuthoritiesForRole special-cases it), or `login_user.is_admin='Y'` (which
// synthesises that same role). Hence the OR — and hence why this is NOT the
// same predicate as IsAdmin. The fixture user `e2e_testmgmt` exists precisely
// to keep the two apart in the e2e oracle.
func (p *Principal) HasAdminAuthority() bool {
	return p.IsAdmin || p.HasRole(valueholder.RoleGlobalAdmin)
}

// Store is the session backend. Implementations must be safe for concurrent
// use.
type Store interface {
	// New creates a session and returns its id.
	New(p *Principal, ttl time.Duration) (string, error)
	// Get returns the principal for id, refreshing its sliding expiry. The
	// bool is false for unknown or expired ids.
	Get(id string) (*Principal, bool)
	// Delete invalidates a session (logout).
	Delete(id string)
}

type entry struct {
	principal *Principal
	ttl       time.Duration
	expires   time.Time
}

// MemoryStore is the in-memory Store, standing in for the servlet container's
// session map.
type MemoryStore struct {
	mu       sync.Mutex
	sessions map[string]*entry
}

// NewMemoryStore returns an empty store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{sessions: make(map[string]*entry)}
}

// NewID returns an opaque session id: 128 bits from crypto/rand, uppercase hex.
//
// Deliberately NOT a UUID and deliberately not parseable — Tomcat's ids are
// opaque 32-char uppercase hex and nothing in the application reads structure
// out of them, so matching the shape keeps any client-side assumption (a
// length check, a regex in a log scraper) working across the cutover.
func NewID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(b)), nil
}

// New creates a session. The caller supplies the TTL because it is PER USER in
// Java: login_user.user_time_out minutes (default 20 when absent), applied as
// HttpSession.setMaxInactiveInterval.
func (s *MemoryStore) New(p *Principal, ttl time.Duration) (string, error) {
	id, err := NewID()
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = &entry{principal: p, ttl: ttl, expires: time.Now().Add(ttl)}
	return id, nil
}

// Get returns the principal and slides the expiry forward, matching
// maxInactiveInterval semantics (idle timeout, not absolute lifetime).
func (s *MemoryStore) Get(id string) (*Principal, bool) {
	if id == "" {
		return nil, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.sessions[id]
	if !ok {
		return nil, false
	}
	if time.Now().After(e.expires) {
		delete(s.sessions, id)
		return nil, false
	}
	e.expires = time.Now().Add(e.ttl)
	return e.principal, true
}

// Delete invalidates a session.
func (s *MemoryStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
}
