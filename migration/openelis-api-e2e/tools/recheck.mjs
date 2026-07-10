// Re-verify the key baseline points against the live API.
import { request } from "@playwright/test";
const B = "https://localhost/api/OpenELIS-Global/";
const P = (ok, label, detail = "") =>
  console.log(`${ok ? "PASS" : "FAIL"}  ${label}${detail ? "  :: " + detail : ""}`);

const anon = await request.newContext({
  baseURL: B,
  ignoreHTTPSErrors: true,
  storageState: { cookies: [], origins: [] },
});
const authed = await request.newContext({
  baseURL: B,
  ignoreHTTPSErrors: true,
  storageState: "playwright/.auth/admin.json",
});

// 1) unauth protected endpoint -> login HTML (200), not JSON
{
  const r = await anon.get("rest/users");
  const t = await r.text();
  P(r.status() === 200 && t.startsWith("<!DOCTYPE html"), "unauth protected -> login HTML (200)", `status=${r.status()} len=${t.length}`);
}
// 2) authed session is authenticated + has csrf token + roles
let csrf = "";
{
  const s = await (await authed.get("session")).json();
  csrf = s.csrf || "";
  P(s.authenticated === true && !!s.csrf && s.roles?.includes("Global Administrator"), "authed session: authenticated + csrf + roles", `csrf=${csrf.slice(0, 8)}… roles=${s.roles?.length}`);
}
// 3) authed protected endpoint -> JSON data
{
  const r = await authed.get("rest/users");
  const t = await r.text();
  P(r.status() === 200 && t.startsWith("["), "authed protected -> JSON data", `status=${r.status()}`);
}
// 4) write verb WITHOUT csrf -> 403 (safe; no mutation)
{
  const r = await authed.fetch("rest/Organization", { method: "DELETE" });
  P(r.status() === 403, "write without CSRF -> 403 (CSRF enforced)", `status=${r.status()}`);
}
// 5) public whitelist reachable anon, not login page
{
  const r = await anon.get("rest/open-configuration-properties");
  const t = await r.text();
  P(r.status() === 200 && !t.startsWith("<!DOCTYPE html"), "public whitelist anon reachable", `status=${r.status()}`);
}
await anon.dispose();
await authed.dispose();
