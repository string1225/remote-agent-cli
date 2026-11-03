const assert = require("node:assert/strict");
const fs = require("node:fs");
const vm = require("node:vm");
const test = require("node:test");
const source = fs.readFileSync(process.env.UI_TEST_APP_SOURCE || "web/app.js", "utf8");
const session = id => ({id, serviceId:"project", title:id});
const turns = count => Array.from({length:count}, (_, index) => [
  {seq:index * 4 + 1, turnId:String(index), role:"user", kind:"message", text:"prompt " + index},
  {seq:index * 4 + 2, turnId:String(index), role:"user", kind:"message", text:" continued"},
  {seq:index * 4 + 3, turnId:String(index), role:"system", kind:"tool", text:"$ pwd"},
  {seq:index * 4 + 4, turnId:String(index), role:"assistant", kind:"stdout", text:"reply " + index},
]).flat();
const deferred = () => {
  let resolve;
  const promise = new Promise(done => { resolve = done; });
  return {promise, resolve};
};
function harness(reply) {
  const elements = new Map(), cache = new Map(), calls = [], notices = [];
  const $ = id => {
    if (!elements.has(id)) elements.set(id, {value:"", scrollHeight:0, scrollTop:0, clientHeight:100});
    return elements.get(id);
  };
  const state = {
    user:{id:"account"}, selected:{id:"client"}, serviceId:"project", sessions:[], session:null,
    services:[{id:"project", provider:"codex"}], entries:[], submitting:new Set(), readRevision:0,
    peer:{ws:{readyState:1}}, mode:"web", permissionMode:"full",
  };
  const key = (kind, scope) => JSON.stringify([kind, scope]);
  const h = {
    state, $, window:{}, closeSayso:null, notice:message => notices.push(message),
    service:() => state.services[0], renderSessions() {},
    WorkspaceCache:{
      read:(kind, scope) => cache.get(key(kind, scope)),
      write:(kind, scope, value) => cache.set(key(kind, scope), value),
      remove:(kind, scope) => cache.delete(key(kind, scope)),
      history:(scope, current, entries) => cache.set(key("history", scope), {session:current, entries:[...entries]}),
    },
    rpc:async (type, fields) => { calls.push([type, fields]); return reply(type, fields); },
    renderChat() { $("chat-scroll").scrollHeight = state.entries.length * 100; },
  };
  vm.createContext(h);
  for (const [start, end] of [
    ["function draftContext(", "function updateSessionContext("],
    ["function setRunning(", "function resetWorkspace("],
    ["async function loadSession(", "function sessionTitle("],
    ["function mergeEntries(", "function toolInvocation("],
    ["async function pollSession(", "async function pollWorkspace("],
  ]) vm.runInContext(source.slice(source.indexOf(start), source.indexOf(end)), h);
  return {...h, calls, notices};
}

test("initial load requests two complete turns, then loads exactly two older turns per click", async () => {
  const rows = turns(7);
  const h = harness(async (_, fields) => {
    assert.equal(fields.turns, 2);
    const end = fields.before || 29, start = Math.max(1, end - 8);
    return {session:session("chat"), entries:rows.filter(row => row.seq >= start && row.seq < end), olderCursor:start > 1 ? start : 0, historyEnd:end, nextCursor:0};
  });
  await h.loadSession(session("chat"));
  assert.equal(h.calls.length, 1, "opening must not fetch ancient pages");
  assert.deepEqual(Array.from(h.state.entries, row => row.seq), [21,22,23,24,25,26,27,28]);
  assert.equal(h.state.olderCursor, 21);
  await h.loadMoreHistory();
  assert.equal(h.calls.at(-1)[1].before, 21);
  assert.equal(h.state.entries.length, 16);
  await h.loadMoreHistory();
  await h.loadMoreHistory();
  assert.equal(h.state.olderCursor, 0);
  assert.deepEqual(Array.from(h.state.entries, row => row.seq), rows.map(row => row.seq));
  await h.loadMoreHistory();
  assert.equal(h.calls.length, 4, "exhausted history performs no request");
});

test("long replies finish their transport pages inside the same snapshot", async () => {
  const h = harness(async (_, fields) => fields.cursor === 0
    ? {session:session("chat"), entries:turns(1), olderCursor:1, historyEnd:9, nextCursor:4}
    : {session:session("chat"), entries:turns(2).slice(4), olderCursor:0, historyEnd:9, nextCursor:0});
  await h.loadSession(session("chat"));
  assert.equal(h.calls[1][1].cursor, 4);
  assert.equal(h.calls[1][1].before, 9);
  assert.equal(h.state.entries.length, 8);
  assert.equal(h.state.olderCursor, 1, "continuation metadata cannot erase earlier history");
});

test("load-more preserves the visible message and live polling keeps newer replies", async () => {
  const older = deferred();
  const h = harness(async (_, fields) => fields.turns
    ? fields.before ? older.promise : {session:session("chat"), entries:turns(4).slice(8), olderCursor:9, historyEnd:17, nextCursor:0}
    : {session:session("chat"), entries:[{seq:17, turnId:"live", role:"assistant", text:"new live reply"}], nextCursor:0});
  await h.loadSession(session("chat"));
  const scroll = h.$("chat-scroll");
  scroll.scrollTop = 50;
  const anchorSeq = h.state.entries[0].seq;
  const anchor = {getBoundingClientRect:() => ({top:h.state.entries.findIndex(row => row.seq === anchorSeq) * 100 - scroll.scrollTop})};
  h.$("messages").firstElementChild = anchor;
  const loading = h.loadMoreHistory();
  await h.loadMoreHistory();
  assert.equal(h.calls.length, 2, "duplicate clicks cannot create concurrent requests");
  await h.pollSession();
  assert.equal(h.calls.at(-1)[1].cursor, 16);
  const beforeTop = anchor.getBoundingClientRect().top;
  older.resolve({session:session("chat"), entries:turns(2), olderCursor:0, historyEnd:9, nextCursor:0});
  await loading;
  assert.equal(anchor.getBoundingClientRect().top, beforeTop);
  assert.equal(h.state.entries.at(-1).text, "new live reply");
  assert.equal(h.state.entries.length, 17);
});

test("switching sessions discards slow history without requesting its next page", async () => {
  const slow = deferred();
  const h = harness(async (_, fields) => fields.sessionId === "old" ? slow.promise
    : {session:session("new"), entries:turns(1), olderCursor:0, nextCursor:0});
  const old = h.loadSession(session("old"));
  await h.loadSession(session("new"));
  slow.resolve({session:session("old"), entries:turns(2), olderCursor:1, nextCursor:8});
  await old;
  assert.equal(h.state.session.id, "new");
  assert.equal(h.state.entries.length, 4);
  assert.equal(h.calls.length, 2);
});

test("failed load-more retains its cursor for retry and never changes the draft", async () => {
  let fail = true;
  const h = harness(async (_, fields) => {
    if (fields.before && fail) throw Error("retry me");
    return {session:session("chat"), entries:fields.before ? turns(2) : turns(4).slice(8), olderCursor:fields.before ? 0 : 9, nextCursor:0};
  });
  await h.loadSession(session("chat"));
  h.$("prompt").value = "keep my draft";
  await h.loadMoreHistory();
  assert.equal(h.state.olderCursor, 9);
  assert.equal(h.state.loadingOlder, null);
  assert.deepEqual(h.notices, ["retry me"]);
  fail = false;
  await h.loadMoreHistory();
  assert.equal(h.state.entries.length, 16);
  assert.equal(h.$("prompt").value, "keep my draft");
});

test("legacy clients and cached transcripts display two turns with older turns available", async () => {
  const h = harness(async () => ({session:session("chat"), entries:turns(6), nextCursor:0}));
  h.WorkspaceCache.history(["account", "client", "project", "chat"], session("chat"), turns(6));
  await h.loadSession(session("chat"), {localOnly:true});
  assert.equal(h.state.entries.length, 8);
  assert.equal(h.calls.length, 0);
  const peer = h.state.peer;
  h.state.peer = null;
  await h.loadMoreHistory();
  assert.equal(h.state.entries.length, 16, "older cached turns remain readable offline");
  assert.equal(h.calls.length, 0);
  h.state.peer = peer;
  await h.loadSession(session("chat"));
  assert.equal(h.state.entries.length, 8);
  await h.loadMoreHistory();
  assert.equal(h.state.entries.length, 16);
  assert.equal(h.calls.length, 1, "legacy transcript remains in memory for load-more");
});
