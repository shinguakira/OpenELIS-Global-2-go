// §1 Session & auth.
import { test, expect } from "@playwright/test";

test.describe("session", () => {
  test("authenticated session reports admin identity and roles", async ({
    request,
  }) => {
    const s = await (await request.get("session")).json();
    expect(s.authenticated).toBe(true);
    expect(s.loginName).toBe("admin");
    expect(s.userId).toBeTruthy();
    expect(Array.isArray(s.roles)).toBe(true);
    expect(s.roles).toContain("Global Administrator");
  });

  test("server-time endpoint responds", async ({ request }) => {
    const res = await request.get("rest/server-time");
    expect(res.status()).toBe(200);
  });
});
