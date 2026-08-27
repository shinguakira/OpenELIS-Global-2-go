// Package csrf reproduces Spring Security 6's CSRF token handling.
//
// Java runs the DEFAULT setup: HttpSessionCsrfTokenRepository (the token lives
// in the session) behind XorCsrfTokenRequestAttributeHandler (every value handed
// to a client is masked). Parameter name `_csrf`, header `X-CSRF-TOKEN`.
//
// The masking is the trap (migration/auth-adoption-plan.md §2.6, verified live:
// two `GET /session` reads in ONE session return two different `csrf` strings).
// A port that stores a token and compares the submitted value with `==` rejects
// every request. A port that returns fresh random bytes each time looks
// identical from the outside but cannot validate anything.
//
//	masked = base64url( random(n) || (token XOR random) )   where n = len(token)
//
// Un-masking splits the halves and XORs them back. Spring emits base64url
// WITHOUT padding.
package csrf

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
)

// TokenBytes is the length of a raw CSRF token. Spring's default
// (DefaultCsrfToken via UUID.randomUUID().toString()) is a 36-character UUID
// string; the exact length is not part of the observable contract — what is
// observable is that masked values are 2x the raw length once decoded, and that
// two reads un-mask to the same value. 32 bytes of base64url text keeps the
// same shape with a stronger token.
const TokenBytes = 24

// NewToken mints a raw per-session token.
func NewToken() (string, error) {
	b := make([]byte, TokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// Mask returns a fresh masked encoding of token. Call it on EVERY read: handing
// the same masked value out twice would fail the parity oracle and, more to the
// point, defeat the BREACH mitigation the masking exists for.
func Mask(token string) (string, error) {
	raw := []byte(token)
	pad := make([]byte, len(raw))
	if _, err := rand.Read(pad); err != nil {
		return "", err
	}
	out := make([]byte, 0, len(raw)*2)
	out = append(out, pad...)
	for i := range raw {
		out = append(out, raw[i]^pad[i])
	}
	return base64.RawURLEncoding.EncodeToString(out), nil
}

// Unmask reverses Mask. Returns ("", false) when the value cannot be a masked
// token, so a caller reports "invalid token" rather than panicking on client
// input.
func Unmask(masked string) (string, bool) {
	raw, err := base64.RawURLEncoding.DecodeString(masked)
	if err != nil {
		// Tolerate padded input: some clients round-trip the value through a
		// library that re-adds '='. Spring's Base64.getUrlDecoder() accepts it.
		raw, err = base64.URLEncoding.DecodeString(masked)
		if err != nil {
			return "", false
		}
	}
	if len(raw) == 0 || len(raw)%2 != 0 {
		return "", false
	}
	n := len(raw) / 2
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = raw[i] ^ raw[n+i]
	}
	return string(out), true
}

// Valid reports whether a client-submitted (masked) value carries the expected
// raw token. Constant-time on the token comparison, matching Spring's
// CsrfFilter, which uses MessageDigest.isEqual rather than String.equals.
func Valid(expected, submitted string) bool {
	got, ok := Unmask(submitted)
	if !ok {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(got)) == 1
}
