// Spring Security 6 CSRF token masking — the single detail most likely to be
// silently wrong in a port.
//
// Spring's default `XorCsrfTokenRequestAttributeHandler` never hands a client
// the raw session token. Every read produces
//     masked = base64url( random(n) || (token XOR random) )   where n = len(token)
// so two `GET /session` calls in the SAME session return two DIFFERENT `csrf`
// strings that both un-mask to the same underlying token. A port that stores a
// token and compares the submitted value with `==` rejects every request; a
// port that returns a fresh RANDOM string each time looks identical from the
// outside until you actually un-mask. Hence the helper: the spec asserts the
// values differ AND that they decode to one stable token.
//
// Verified live against Java (see p0-auth.spec.ts "csrf token is XOR-masked").

/** Decode base64url (Spring emits URL-safe, unpadded). */
function b64urlDecode(s: string): Buffer {
  const pad = s.length % 4 === 0 ? "" : "=".repeat(4 - (s.length % 4));
  return Buffer.from(s.replace(/-/g, "+").replace(/_/g, "/") + pad, "base64");
}

/**
 * Reverse Spring's XOR mask. Returns null when the value cannot be a masked
 * token (odd length, or not decodable) — callers assert on that rather than
 * throwing, so a port returning a plain token is reported as a diff, not a
 * crash.
 */
export function unmaskCsrf(masked: string): string | null {
  let raw: Buffer;
  try {
    raw = b64urlDecode(masked);
  } catch {
    return null;
  }
  if (raw.length === 0 || raw.length % 2 !== 0) return null;
  const n = raw.length / 2;
  const out = Buffer.alloc(n);
  for (let i = 0; i < n; i++) out[i] = raw[i] ^ raw[n + i];
  return out.toString("utf8");
}
