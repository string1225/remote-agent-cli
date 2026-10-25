const assert = require('node:assert/strict');
const path = require('node:path');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');

(async () => {
  const browser = await chromium.launch({ headless: true, channel: process.env.PLAYWRIGHT_CHANNEL || 'msedge' });
  let page;
  try {
    page = await browser.newPage({ viewport: { width: 1100, height: 800 } });
    page.setDefaultTimeout(6000);
    const errors = [];
    page.on('pageerror', error => errors.push(error.message));
    await page.route('https://maintenance.test/**', route => route.fulfill({ contentType: 'text/html', body: '<!doctype html><html><head></head><body><div id="device-menu"><button id="device-menu-delete">删除</button></div><p id="enroll-command-note"></p></body></html>' }));
    await page.goto('https://maintenance.test/');
    await page.evaluate(() => {
      window.$ = id => document.getElementById(id);
      window.node = (tag, text, className) => { const element = document.createElement(tag); if (text !== undefined) element.textContent = text; if (className) element.className = className; return element; };
      window.state = {};
      window.basePath = '';
      window.showInstallCommands = () => {};
      window.fixture = { requests: [], loseAck: true, completed: false, operation: null };
      window.WebSocket = class {
        static OPEN = 1;
        constructor() { this.readyState = 1; setTimeout(() => this.onmessage?.({ data: JSON.stringify({ type: 'ready', transport: 'relay-v1' }) }), 5); }
        send(raw) {
          const message = JSON.parse(raw), request = message.request;
          if (!request) return;
          let data;
          if (request.type === 'services.list') data = { remoteMaintenance: true, remoteClientUpdate: true, clientVersion: 'fixture-old', providers: {} };
          if (request.type === 'maintenance.start') {
            fixture.requests.push(request.body.requestId);
            fixture.operation ||= { id: 'operation', requestId: request.body.requestId, kind: 'codex', status: 'running', stage: 'download', message: '正在下载官方 Codex CLI', downloaded: 1024, total: 4096 };
            if (fixture.loseAck) { fixture.loseAck = false; this.close(); return; }
            data = fixture.operation;
          }
          if (request.type === 'maintenance.status') {
            if (fixture.completed && fixture.operation) fixture.operation = { ...fixture.operation, status: 'succeeded', stage: 'complete', message: 'Codex CLI 更新完成', afterVersion: '1.2.3' };
            data = fixture.operation || {};
          }
          setTimeout(() => {
            if (request.type === 'services.list') {
              const encoded = btoa(String.fromCharCode(...new TextEncoder().encode(JSON.stringify(data))));
              this.onmessage?.({ data: JSON.stringify({ type: 'event', event: { id: request.id, type: 'chunk', data: encoded } }) });
              data = { chunked: true };
            }
            this.onmessage?.({ data: JSON.stringify({ type: 'event', event: { id: request.id, type: 'result', data } }) });
          }, 5);
        }
        close() { this.readyState = 3; this.onclose?.(); }
      };
    });
    await page.addStyleTag({ path: path.resolve('web/style.css') });
    await page.addStyleTag({ path: path.resolve('web/maintenance.css') });
    await page.addScriptTag({ path: path.resolve('web/device-maintenance.js') });
    const device = { id: 'target-device', name: '测试电脑', os: 'windows', online: true };
    await page.evaluate(d => prepareMaintenanceMenu({ ...d, online: false }), device);
    assert.equal(await page.locator('#device-menu-maintenance').isDisabled(), true);
    await page.evaluate(d => prepareMaintenanceMenu(d, { id: 'project' }), device);
    assert.equal(await page.locator('#device-menu-maintenance').isHidden(), true);
    for (const os of ['darwin', 'linux']) {
      await page.evaluate(d => { prepareMaintenanceMenu(d); document.getElementById("device-menu").hidden=false; }, { ...device, os, id: os + '-client' });
      await page.locator('#device-menu-maintenance').click();
      const update = page.getByRole('menuitem', { name: '更新客户端', exact: true });
      assert.equal(await update.isEnabled(), true, os + ' update is disabled');
      await update.click();
      await page.waitForFunction(() => !$('maintenance-start').disabled);
      assert.match(await page.locator('#maintenance-target').innerText(), new RegExp(os));
      await page.locator('#maintenance-close').click();
    }
    await page.evaluate(d => { prepareMaintenanceMenu(d); document.getElementById("device-menu").hidden=false; }, device);
    await page.locator('#device-menu-maintenance').click();
    assert.equal(await page.getByRole('menu', { name: '客户端更新与工具' }).isVisible(), true);
    await page.getByRole('menuitem', { name: '安装／更新 Codex CLI' }).click();
    assert.match(await page.locator('#maintenance-target').innerText(), /target-device/);
    await page.locator('#maintenance-start').click();
    await page.waitForFunction(() => $('maintenance-message').textContent.includes('重试会核对同一请求'));
    await page.locator('#maintenance-start').click();
    await page.waitForFunction(() => $('maintenance-progress').value === 0.25);
    assert.equal(await page.locator('#maintenance-start').isDisabled(), true);
    assert.equal(await page.evaluate(() => fixture.requests.length === 2 && fixture.requests[0] === fixture.requests[1]), true, 'lost acknowledgement duplicated installation request');
    await page.locator('#maintenance-close').click();
    await page.evaluate(d => openDeviceMaintenance(d, 'codex'), device);
    await page.waitForFunction(() => $('maintenance-progress').value === 0.25 && !$('maintenance-progress').hidden);
    await page.evaluate(() => { fixture.completed = true; });
    await page.waitForFunction(() => $('maintenance-version').textContent.includes('1.2.3'));
    assert.equal(await page.locator('#maintenance-start').isEnabled(), true);
    await page.locator('#maintenance-close').click();
    await page.evaluate(() => showInstallCommands({ repair: true, agent: { id: 'target-device', name: '目标电脑', hostname: 'old-host' }, windows: 'remote-agent connect --client-id target-device' }));
    assert.match(await page.locator('#repair-target-client').innerText(), /target-device.*old-host/);
    assert.match(await page.locator('#enroll-command-note').innerText(), /BIND/);
    await page.evaluate(() => showInstallCommands({ agent: { id: 'new-client' } }));
    assert.equal(await page.locator('#repair-target-client').count(), 0);
    assert.deepEqual(errors, []);
    console.log('PASS: submenu, offline/project state, target ID, lost-ack idempotency, download progress, reopen during installation and completion');
  } catch (error) {
    if (page) console.error(await page.locator('#device-maintenance-dialog').innerText());
    throw error;
  } finally {
    await browser.close();
  }
})().catch(error => { console.error(error); process.exitCode = 1; });
