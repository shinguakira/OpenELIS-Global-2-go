// Setup project: establish an authenticated session and persist cookies.
// Uses the contract layer so a port that reshapes login is retargeted by config.
import { test as setup, expect } from "@playwright/test";
import fs from "node:fs";
import { ADMIN_USER, ADMIN_PASS, AUTH_STATE } from "../fixtures/env";
import {
  LOGIN_PATH,
  LOGIN_USER_FIELD,
  LOGIN_PASS_FIELD,
  LOGIN_SUCCESS_FIELD,
  SESSION_PATH,
} from "../fixtures/contract";

setup("authenticate admin", async ({ request }) => {
  // 1) touch session to receive a session id / cookie
  await request.get(SESSION_PATH);
  // 2) form login
  const res = await request.post(LOGIN_PATH, {
    form: { [LOGIN_USER_FIELD]: ADMIN_USER, [LOGIN_PASS_FIELD]: ADMIN_PASS },
  });
  expect(res.ok(), "login should return 2xx").toBeTruthy();
  expect((await res.json())[LOGIN_SUCCESS_FIELD], "login success flag").toBe(
    true,
  );
  // 3) confirm the session is authenticated
  const session = await (await request.get(SESSION_PATH)).json();
  expect(session.authenticated).toBe(true);
  expect(session.loginName).toBe(ADMIN_USER);
  // 4) persist cookie jar for the other projects
  fs.mkdirSync("playwright/.auth", { recursive: true });
  await request.storageState({ path: AUTH_STATE });
});
