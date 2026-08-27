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

/** Encode bytes as unpadded base64url, the way Spring emits masked tokens. */
function b64urlEncode(buf: Buffer): string {
  return buf
    .toString("base64")
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/, "");
}

/**
 * Mask a raw token the way Spring's XorCsrfTokenRequestAttributeHandler does:
 *     masked = base64url( random(n) || (token XOR random) ),  n = len(token)
 *
 * Needed to forge a token DETERMINISTICALLY. The obvious forgery — corrupt one
 * character of a real masked value — is not deterministic against Java: the
 * server un-masks, gets a byte that is now arbitrary, and runs it through
 * `Utf8.decode`. Whether that byte happens to be valid UTF-8 decides whether
 * Java answers with its clean access-denied redirect (302) or throws and
 * answers 500. Measured live: roughly 60% 500, 40% 302 over 12 runs.
 *
 * Building a VALID mask of a DIFFERENT-but-still-ASCII token avoids the coin
 * toss entirely, and tests the stronger property anyway — that the server
 * really un-masks and compares, rather than merely choking on malformed input.
 */
export function maskCsrf(token: string): string {
  const raw = Buffer.from(token, "utf8");
  const pad = Buffer.alloc(raw.length);
  for (let i = 0; i < pad.length; i++) pad[i] = (i * 7 + 13) % 256; // deterministic
  const xored = Buffer.alloc(raw.length);
  for (let i = 0; i < raw.length; i++) xored[i] = raw[i] ^ pad[i];
  return b64urlEncode(Buffer.concat([pad, xored]));
}

/**
 * Return a token of the SAME length that is still printable ASCII but is not
 * `token` — a well-formed credential belonging to nobody.
 */
export function differentAsciiToken(token: string): string {
  const chars = [...token];
  // Rotate the first alphanumeric character within a safe ASCII range. Both
  // targets' tokens are ASCII (Java: a UUID, Go: base64url), so the result
  // stays valid UTF-8 and the server reaches its compare step.
  for (let i = 0; i < chars.length; i++) {
    if (/[0-9a-y]/.test(chars[i])) {
      chars[i] = String.fromCharCode(chars[i].charCodeAt(0) + 1);
      return chars.join("");
    }
  }
  throw new Error(`cannot derive a different token from ${JSON.stringify(token)}`);
}
