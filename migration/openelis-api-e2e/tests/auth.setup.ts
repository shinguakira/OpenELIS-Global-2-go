// Setup project: establish an authenticated session and persist cookies.
// Uses the contract layer so a port that reshapes login is retargeted by config.
//
// The SAME handshake runs against Java (`setup`) and the Go port (`setup-go`) —
// there is deliberately no Go-specific login branch here. If the Go port ever
// needs one, the port is wrong, not this file. The only thing that varies per
// project is the target (taken from the project's own baseURL) and which cookie
// jar the result is written to.
import { test as setup, expect } from "@playwright/test";
import fs from "node:fs";
import path from "node:path";
import { ADMIN_USER, ADMIN_PASS, AUTH_STATE, GO_AUTH_STATE } from "../fixtures/env";
import {
  LOGIN_PATH,
  LOGIN_USER_FIELD,
  LOGIN_PASS_FIELD,
  LOGIN_SUCCESS_FIELD,
  SESSION_PATH,
} from "../fixtures/contract";

setup("authenticate admin", async ({ request }) => {
  const statePath =
    setup.info().project.name === "setup-go" ? GO_AUTH_STATE : AUTH_STATE;

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
  fs.mkdirSync(path.dirname(statePath), { recursive: true });
  await request.storageState({ path: statePath });
});
