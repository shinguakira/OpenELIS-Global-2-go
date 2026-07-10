// ── Language-neutrality layer ────────────────────────────────────────────────
// This suite is a PARITY ORACLE: the same tests must run against the Java baseline
// today and a Go (or any) re-implementation later. Anything that *could* legitimately
// differ between implementations lives here as a config knob (env-overridable), so
// retargeting the suite at another language is CONFIG, not a test rewrite.
//
// What is NOT here (kept in the tests, because it IS the cross-language contract):
//   - endpoint paths + verbs (fixtures/endpoints.generated.ts — the frozen contract)
//   - response JSON shapes / DB row effects (the behavior the port must reproduce)
//
// Retarget at a Go port:  OE_BASE_URL=https://go-host/... npm test
// If the Go port changes auth to 401-JSON:  OE_UNAUTH_MODE=status

// Informational label for the target under test.
export const TARGET = process.env.OE_TARGET || "java";

// ── Auth / session mechanics (may be re-shaped by the port) ──────────────────
export const LOGIN_PATH =
  process.env.OE_LOGIN_PATH || "ValidateLogin?apiCall=true";
export const LOGIN_USER_FIELD = process.env.OE_LOGIN_USER_FIELD || "loginName";
export const LOGIN_PASS_FIELD = process.env.OE_LOGIN_PASS_FIELD || "password";
export const LOGIN_SUCCESS_FIELD = process.env.OE_LOGIN_SUCCESS_FIELD || "success";
export const SESSION_PATH = process.env.OE_SESSION_PATH || "session";
export const CSRF_SESSION_FIELD = process.env.OE_CSRF_FIELD || "csrf";
// header the port expects the CSRF token on for state-changing calls
export const CSRF_HEADER = process.env.OE_CSRF_HEADER || "X-CSRF-TOKEN";

// ── How the API answers an UNAUTHENTICATED protected request ─────────────────
// Java baseline: serves the login HTML page with HTTP 200.
// A Go port may instead return 401/403 JSON — set OE_UNAUTH_MODE=status.
export const UNAUTH_MODE = process.env.OE_UNAUTH_MODE || "login-html"; // "login-html" | "status"
const LOGIN_HTML_MARKER = "<!DOCTYPE html";

/** True when a response represents the "unauthenticated / blocked" outcome,
 *  under whichever convention the target implementation uses. */
export function isBlocked(status: number, body: string): boolean {
  if (UNAUTH_MODE === "status") return status === 401 || status === 403;
  return body.startsWith(LOGIN_HTML_MARKER);
}

/** True when a response is an authenticated app response (i.e. NOT blocked). */
export function isAuthedResponse(status: number, body: string): boolean {
  return !isBlocked(status, body);
}
