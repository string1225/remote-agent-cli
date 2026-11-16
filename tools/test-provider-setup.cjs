const assert = require("node:assert/strict");
const fs = require("node:fs");
const vm = require("node:vm");

const source = fs.readFileSync("web/app.js", "utf8");
const elements = new Map();
const calls = [];
const element = (id) => {
  if (!elements.has(id)) elements.set(id, { showModal() { calls.push(["dialog", id]); }, focus() {} });
  return elements.get(id);
};
const sandbox = {
  $: element,
  location: { origin: "https://codex.sunny-string.cn" },
  basePath: "/agents",
  state: { selected: { name: "目标电脑" }, providers: {}, serviceId: "project" },
  service: () => ({ provider: "codex" }),
  rpc: async (...args) => { calls.push(["rpc", ...args]); return { id: "session" }; },
  refreshSessions: async () => calls.push(["refresh"]),
  loadSession: async (session) => calls.push(["load", session.id]),
  closeDrawers: () => {},
  notice: (message) => { throw Error(message); },
};
vm.createContext(sandbox);
vm.runInContext(source.slice(source.indexOf("function codexInstallCommand("), source.indexOf("function showService(")), sandbox);

async function test() {
  assert.equal(sandbox.codexInstallCommand("install"), "remote-agent codex install");
  assert.equal(sandbox.codexInstallCommand("update", "https://codex.sunny-string.cn/agents/"), "remote-agent codex update");
  assert.equal(sandbox.codexInstallCommand("install", "https://custom.test/agents/"), "remote-agent codex install --server https://custom.test/agents");

  await element("new-session").onclick();
  assert.deepEqual(calls, [["dialog", "provider-setup-dialog"]]);
  assert.equal(element("provider-setup-command").value, "remote-agent codex setup");
  assert.match(element("provider-setup-target").textContent, /目标电脑/);
  assert.equal(element("provider-setup-codex").hidden, false);
  assert.equal(element("provider-setup-qoder").hidden, true);

  calls.length = 0;
  sandbox.state.providers.codex = { available: true };
  await element("new-session").onclick();
  assert.equal(calls[0][0], "rpc");
  assert.equal(calls[0][1], "sessions.create");
  assert.equal(calls[0][2].serviceId, "project");
  assert.deepEqual(calls.slice(1), [["refresh"], ["load", "session"]]);

  sandbox.showProviderSetup("qoder");
  assert.equal(element("provider-setup-codex").hidden, true);
  assert.equal(element("provider-setup-qoder").hidden, false);
  assert.match(element("provider-setup-title").textContent, /Qoder CLI/);

  const html = fs.readFileSync("web/index.html", "utf8");
  for (const id of elements.keys()) assert(html.includes('id="' + id + '"'), "missing element " + id);
  assert(html.includes('data-copy="provider-setup-command"'));
  assert(html.includes("Windows 重启服务需要管理员权限"));
  console.log("PASS: optional default server; custom server preserved; missing CLI opens installation guide without creating a session; installed CLI creates a session; Qoder remains separate");
}

test().catch((error) => { console.error(error); process.exitCode = 1; });
