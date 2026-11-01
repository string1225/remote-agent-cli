const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const test = require('node:test');
const source = fs.readFileSync(process.env.UI_TEST_APP_SOURCE || 'web/app.js', 'utf8');

function harness(reply) {
  const elements = new Map(), cached = [];
  const state = {
    user:{id:'user'}, selected:{id:'client'}, serviceId:'project',
    session:{id:'session', taskId:'task', taskStatus:'running'}, sessions:[],
    submitting:new Set(), peer:{}, permissionMode:'full', listRevision:0,
  };
  const sandbox = {
    state, $:id => {
      if (!elements.has(id)) elements.set(id, {value:''});
      return elements.get(id);
    },
    WorkspaceCache:{write() {}}, rpc:reply, service:() => ({}),
    renderSessions() {}, cacheCurrentSession:() => cached.push({...state.session}),
    Map, Set,
  };
  vm.createContext(sandbox);
  vm.runInContext(source.slice(source.indexOf('function setRunning('), source.indexOf('function resetWorkspace(')), sandbox);
  vm.runInContext(source.slice(source.indexOf('async function refreshSessions('), source.indexOf('let sessionSignature;')), sandbox);
  return {...sandbox, cached};
}

for (const status of ['done', 'failed', 'cancelled']) {
  test(`list metadata restores sending after ${status} without a history response`, async () => {
    const h = harness(async () => ({sessions:[{id:'session', taskId:'task', taskStatus:status}], nextCursor:0}));
    h.setRunning();
    assert.equal(h.$('send-run').disabled, true);
    await h.refreshSessions(false, true);
    assert.equal(h.$('send-run').disabled, false);
    assert.equal(h.$('cancel-run').hidden, true);
    assert.equal(h.cached.at(-1).taskStatus, status);
  });
}

test('completion for another task or session cannot enable sending', async () => {
  const h = harness(async () => ({sessions:[
    {id:'session', taskId:'old-task', taskStatus:'done'},
    {id:'other-session', taskId:'task', taskStatus:'done'},
  ], nextCursor:0}));
  h.setRunning();
  await h.refreshSessions(false, true);
  assert.equal(h.$('send-run').disabled, true);
  assert.equal(h.state.session.taskStatus, 'running');
});

test('completion does not bypass submission, permissions, or connection checks', async () => {
  for (const guard of ['submitting', 'permission', 'connection']) {
    const h = harness(async () => ({sessions:[{id:'session', taskId:'task', taskStatus:'done'}], nextCursor:0}));
    if (guard === 'submitting') h.state.submitting.add('session');
    if (guard === 'permission') h.state.permissionMode = '';
    await h.refreshSessions(false, true);
    if (guard === 'connection') h.state.peer = null;
    h.setRunning();
    assert.equal(h.$('send-run').disabled, true, guard);
  }
});
