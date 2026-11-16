const assert = require("node:assert/strict");
const fs = require("node:fs");
const vm = require("node:vm");
const test = require("node:test");
const source = fs.readFileSync("web/app.js", "utf8");

const session = (id, project = "p1", status = "") => ({id, serviceId:project, title:id, taskId:"task-" + id, taskStatus:status});
const entry = (seq, text) => ({seq, turnId:"turn", role:"assistant", kind:"stdout", text});
const deferred = () => {
  let resolve;
  const promise = new Promise(done => { resolve = done; });
  return {promise, resolve};
};
function harness(reply) {
  const elements = new Map(), cache = new Map(), calls = [];
  const $ = id => {
    if (!elements.has(id)) elements.set(id, {value:"", hidden:false, scrollHeight:100, scrollTop:0});
    return elements.get(id);
  };
  const key = (type, scope) => JSON.stringify([type, scope]);
  const state = {
    user:{id:"user"}, selected:{id:"client", name:"Client"},
    services:[{id:"p1", name:"One", provider:"codex", workspace:"/one"}, {id:"p2", name:"Two", provider:"codex", workspace:"/two"}],
    serviceId:"p1", session:session("a", "p1", "running"), sessions:[], entries:[entry(1, "old a")],
    mode:"web", permissionMode:"full", backgroundRuns:true, providers:{codex:{available:true}},
    peer:{ws:{readyState:1}}, submitting:new Set(), readRevision:0, listRevision:0, listCursor:0,
  };
  const sandbox = {
    state, $, window:{chatAttachments:{snapshot:() => [], busy:() => false, restore() {}}}, document:{hidden:false},
    WorkspaceCache:{
      read:(type, scope) => cache.get(key(type, scope)),
      write:(type, scope, value) => cache.set(key(type, scope), JSON.parse(JSON.stringify(value))),
      remove:(type, scope) => cache.delete(key(type, scope)),
      history:(scope, current, entries) => cache.set(key("history", scope), JSON.parse(JSON.stringify({session:current, entries}))),
    },
    rpc:async (type, fields) => { calls.push([type, fields]); return reply(type, fields, state); },
    service:() => state.services.find(s => s.id === state.serviceId),
    closeSayso:null, renderSessions() {}, renderChat() {}, renderDevices() {}, folderName:value => value,
    notice:message => { throw Error(message); }, WebSocket:{OPEN:1}, crypto:{randomUUID:() => "accepted-turn"},
    Promise, Set, Map, console,
  };
  vm.createContext(sandbox);
  const evaluate = (start, end) => vm.runInContext(source.slice(source.indexOf(start), end ? source.indexOf(end) : undefined), sandbox);
  evaluate("function draftContext(", "function updateSessionContext(");
  evaluate("function setRunning(", "function resetWorkspace(");
  evaluate("function latestWebSession(", "let sessionSignature");
  evaluate("async function loadSession(", "function sessionTitle(");
  evaluate("function mergeEntries(", "function toolInvocation(");
  vm.runInContext("let workspacePoll;", sandbox);
  evaluate("async function pollSession(");
  evaluate('$("run-form").onsubmit =', '$("prompt").onkeydown =');
  return {...sandbox, calls};
}

test("running work allows project/session switching and restores its latest history and draft", async () => {
  const h = harness(async (type, fields) => type === "sessions.list"
    ? {sessions:[session(fields.serviceId === "p1" ? "a" : "b", fields.serviceId)], nextCursor:0}
    : {session:session(fields.sessionId, fields.sessionId === "a" ? "p1" : "p2", fields.sessionId === "a" ? "running" : ""), entries:[entry(1, "latest " + fields.sessionId)], nextCursor:0});
  h.$("prompt").value = "keep my a draft";
  h.state.submitting.add("a");
  await h.selectProject("p2");
  await h.loadSession(session("b", "p2"));
  assert.equal(h.state.session.id, "b");
  assert.equal(h.$("send-run").disabled, false);
  assert.equal(h.$("new-session").disabled, false);
  assert.equal(h.$("cancel-run").hidden, true);
  h.$("prompt").value = "keep my b draft";
  await h.selectProject("p1");
  await h.loadSession(session("a"));
  assert.equal(h.state.entries[0].text, "latest a");
  assert.equal(h.$("prompt").value, "keep my a draft");
  assert.equal(h.$("cancel-run").hidden, false);
  assert.equal(h.$("new-session").disabled, false);
  assert(!h.calls.some(([type]) => type.includes("cancel")));
});

test("polling updates running output and terminal controls even without new entries", async () => {
  let complete = false;
  const h = harness(async (type) => type === "sessions.list"
    ? {sessions:[session("a", "p1", complete ? "done" : "running")], nextCursor:0}
    : {session:session("a", "p1", complete ? "done" : "running"), entries:complete ? [] : [entry(2, "ongoing output")], nextCursor:0});
  h.setRunning();
  assert.equal(h.$("send-run").disabled, true);
  await h.pollWorkspace();
  assert.equal(h.state.entries.at(-1).text, "ongoing output");
  complete = true;
  await h.pollWorkspace();
  assert.equal(h.$("send-run").disabled, false);
  assert.equal(h.$("cancel-run").hidden, true);
  assert.equal(h.state.sessions[0].taskStatus, "done");
  assert.equal(h.WorkspaceCache.read("history", h.draftContext()).session.taskStatus, "done");
});

test("a slow previous view cannot block a new poll or overwrite another project's session", async () => {
  const oldList = deferred(), oldHistory = deferred();
  let old = true;
  const h = harness(async (type, fields) => {
    if (old) return type === "sessions.list" ? oldList.promise : oldHistory.promise;
    return type === "sessions.list" ? {sessions:[session("b", "p2")], nextCursor:0}
      : {session:session(fields.sessionId, "p2"), entries:[entry(2, "only b")], nextCursor:0};
  });
  const poll = h.pollWorkspace(), samePoll = h.pollWorkspace();
  assert.equal(h.calls.length, 2, "one list/history request per view at a time");
  old = false;
  await h.selectProject("p2");
  await h.loadSession(session("b", "p2"));
  await h.pollWorkspace();
  oldList.resolve({sessions:[session("a")], nextCursor:0});
  oldHistory.resolve({session:session("a"), entries:[entry(9, "stale a must be ignored")], nextCursor:0});
  await Promise.all([poll, samePoll]);
  assert.equal(h.state.session.id, "b");
  assert.equal(h.state.sessions[0].id, "b");
  assert.deepEqual(Array.from(h.state.entries, row => row.text), ["only b"]);
});

test("initial paginated history is applied together and does not overlap a poll", async () => {
  const first = deferred();
  const h = harness(async (type, fields) => type === "sessions.list"
    ? {sessions:[session("a")], nextCursor:0}
    : fields.cursor === 0 ? first.promise : {session:session("a", "p1", "running"), entries:[entry(2, "second page")], nextCursor:0});
  const loading = h.loadSession(session("a"));
  await h.pollWorkspace();
  assert.equal(h.calls.filter(([type]) => type === "sessions.get").length, 1);
  first.resolve({session:session("a", "p1", "running"), entries:[entry(1, "first page")], nextCursor:1});
  await loading;
  assert.deepEqual(Array.from(h.state.entries, row => row.text), ["first page", "second page"]);
  assert.equal(h.state.loadingSession, null);
});

test("submission acknowledged after switching keeps the new session and its draft", async () => {
  const accepted = deferred();
  const h = harness(async (type, fields) => type === "tasks.run" ? accepted.promise
    : {session:session(fields.sessionId), entries:[], nextCursor:0});
  h.state.session = session("a");
  h.$("prompt").value = "submitted in a";
  h.saveCurrentDraft();
  const submitting = h.$("run-form").onsubmit({preventDefault() {}});
  await h.loadSession(session("b"));
  h.$("prompt").value = "new b draft";
  h.saveCurrentDraft();
  accepted.resolve({session:session("a", "p1", "running")});
  await submitting;
  assert.equal(h.state.session.id, "b");
  assert.equal(h.$("prompt").value, "new b draft");
  assert.equal(h.WorkspaceCache.read("draft", h.draftContext()).text, "new b draft");
  assert.equal(h.WorkspaceCache.read("pending", h.draftContext("a")), undefined);
  assert.equal(h.$("send-run").disabled, false);
});

test("submission acknowledgement preserves a newer draft in the same session", async () => {
  const accepted = deferred();
  const h = harness(async (type) => type === "tasks.run" ? accepted.promise
    : {session:session("a", "p1", "running"), entries:[], nextCursor:0});
  h.state.session = session("a");
  h.$("prompt").value = "submitted";
  h.saveCurrentDraft();
  const submitting = h.$("run-form").onsubmit({preventDefault() {}});
  h.$("prompt").value = "next prompt draft";
  h.saveCurrentDraft();
  accepted.resolve({session:session("a", "p1", "running")});
  await submitting;
  assert.equal(h.$("prompt").value, "next prompt draft");
  assert.equal(h.WorkspaceCache.read("draft", h.draftContext()).text, "next prompt draft");
});
