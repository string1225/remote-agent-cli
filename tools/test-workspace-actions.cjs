const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');

(async () => {
  const browser = await chromium.launch({headless:true, channel:process.env.PLAYWRIGHT_CHANNEL || 'msedge'});
  try {
    for (const width of [1440, 390]) {
      const page = await browser.newPage({viewport:{width, height:960}}), errors = [];
      page.setDefaultTimeout(8000);
      page.on('pageerror', error => errors.push(error.message));
      page.on('dialog', dialog => dialog.accept());
      await page.addInitScript(() => { window.setInterval = () => 0; });
      await page.route('https://actions.test/**', route => {
        const file = new URL(route.request().url()).pathname.slice(1) || 'index.html';
        if (file === 'api/me') return route.fulfill({json:{id:'tester', username:'测试账号'}});
        if (file === 'api/agents') return route.fulfill({json:[]});
        const filename = file === 'app.js' && process.env.UI_TEST_APP_SOURCE ? process.env.UI_TEST_APP_SOURCE : path.join('web', file);
        return route.fulfill({contentType:file.endsWith('.js') ? 'text/javascript' : file.endsWith('.css') ? 'text/css' : file.endsWith('.svg') ? 'image/svg+xml' : 'text/html', body:fs.readFileSync(filename)});
      });
      await page.goto('https://actions.test/');
      await page.locator('#app').waitFor({state:'visible'});
      await page.evaluate(() => {
        window.fixture = {
          project:{id:'project', name:'测试项目', provider:'codex', workspace:'C:\\Code'},
          session:{id:'session', serviceId:'project', title:'当前会话', taskId:'active', taskStatus:'running', updatedAt:Date.now()},
          archives:Array.from({length:55}, (_,i) => ({id:'archive-'+i, serviceId:'project', title:'旧会话 '+i, archivedAt:1, updatedAt:Date.now()})),
          queued:[], calls:[], lostEdit:true,
        };
        const capabilities = () => ({services:[fixture.project], providers:{codex:{available:true}}, allowedRoots:['C:\\Code'], permissionMode:'full', backgroundRuns:true, taskQueue:true, sessionManagement:true, projectSettings:true});
        rpc = async (type, fields = {}) => {
          fixture.calls.push([type, structuredClone(fields)]);
          if (type === 'services.list') return capabilities();
          if (type === 'services.update') { Object.assign(fixture.project, fields.service); return {...fixture.project}; }
          if (type === 'sessions.list') {
            const rows = fields.archived ? [...fixture.archives, ...(fixture.session.archivedAt && !fixture.session.deletedAt ? [fixture.session] : [])]
              : fixture.session.archivedAt || fixture.session.deletedAt ? [] : [fixture.session];
            const cursor = fields.cursor || 0;
            return {sessions:structuredClone(rows.slice(cursor, cursor+25)), nextCursor:cursor+25 < rows.length ? cursor+25 : 0};
          }
          if (type === 'sessions.get') return {session:{...fixture.session, queuedMessages:structuredClone(fixture.queued.filter(row => ['queued','editing'].includes(row.status)))}, entries:[], nextCursor:0};
          if (type === 'tasks.run') {
            fixture.queued.push({id:fields.id, prompt:fields.prompt, source:'web', status:'queued', createdAt:Date.now()});
            return {session:{...fixture.session}, taskId:fields.id};
          }
          if (type === 'tasks.promote') {
            const message = fixture.queued.find(row => row.id === fields.runId);
            message.status = 'running'; fixture.session.taskId = message.id;
            return {...message};
          }
          if (type === 'tasks.edit') {
            const message = fixture.queued.find(row => row.id === fields.runId);
            message.status = fields.body?.acknowledged ? 'edited' : 'editing';
            if (fixture.lostEdit) { fixture.lostEdit = false; throw Error('模拟编辑确认丢失'); }
            return {...message};
          }
          if (type === 'sessions.rename') { fixture.session.title = fields.title; fixture.session.customTitle = true; }
          else if (type === 'sessions.archive') fixture.session.archivedAt = fields.archived ? Date.now() : 0;
          else if (type === 'sessions.delete') fixture.session.deletedAt = Date.now();
          else throw Error('Unexpected fixture request '+type);
          return {...fixture.session};
        };
        state.selected = {id:'client', name:'测试电脑', os:'windows', online:true};
        state.agents = [state.selected]; state.peer = {ws:{readyState:WebSocket.OPEN}};
        refreshServices(capabilities()); state.serviceId = 'project'; state.session = {...fixture.session}; state.sessions = [state.session];
        state.entries = [{seq:1, turnId:'active', role:'user', kind:'message', text:'当前任务继续执行'}];
        connectionStatus('已连接', true); renderSessions(); renderDevices(); renderChat();
      });
      assert.equal(await page.locator('#send-run').isEnabled(), true, 'running session accepts queued messages');
      await page.locator('#prompt').fill('第一条排队消息'); await page.locator('#send-run').click();
      await page.locator('.queued-message').filter({hasText:'第一条排队消息'}).waitFor();
      assert.equal(await page.locator('#prompt').inputValue(), '');
      await page.locator('#prompt').fill('第二条排队消息'); await page.locator('#send-run').click();
      assert.equal(await page.locator('.queued-message').count(), 2);
      await page.locator('#prompt').fill('已经写好的下一条草稿');
      let editing = page.locator('.queued-message').filter({hasText:'第二条排队消息'});
      await editing.getByRole('button', {name:'编辑', exact:true}).click();
      await page.evaluate(() => pollSession());
      await editing.getByRole('button', {name:'编辑', exact:true}).click();
      await page.waitForFunction(() => document.getElementById('prompt').value.includes('第二条排队消息'));
      assert.equal(await page.locator('#prompt').inputValue(), '已经写好的下一条草稿\n\n第二条排队消息');
      assert.equal(await page.locator('.queued-message').count(), 1, 'edited message leaves the execution queue');
      if (process.env.UI_TEST_SCREENSHOT_DIR) {
        await page.evaluate(() => { document.getElementById('notice').hidden = true; });
        fs.mkdirSync(process.env.UI_TEST_SCREENSHOT_DIR, {recursive:true});
        await page.screenshot({path:path.join(process.env.UI_TEST_SCREENSHOT_DIR, `queue-${width}.png`)});
      }
      await page.locator('.queued-message').getByRole('button', {name:'插队', exact:true}).click();
      await page.waitForFunction(() => !document.querySelector('.queued-message'));
      const submissions = await page.evaluate(() => fixture.calls.filter(([type]) => type === 'tasks.run'));
      assert.equal(submissions.length, 2); assert(submissions.every(([,fields]) => !fields.interrupt));
      assert.equal(await page.locator('#prompt').inputValue(), '已经写好的下一条草稿\n\n第二条排队消息');

      if (width < 760) await page.locator('#open-conversations').click();
      let current = page.locator('#conversation-list>.conversation-card').filter({hasText:'当前会话'});
      await current.click({button:'right'});
      await page.locator('#session-menu-rename').click();
      await page.locator('#session-rename-form input').fill('重命名后的会话');
      await page.locator('#session-rename-form button[type=submit]').click();
      current = page.locator('#conversation-list>.conversation-card').filter({hasText:'重命名后的会话'});
      await current.waitFor(); await current.press('Shift+F10');
      await page.locator('#session-menu-archive').click();
      await page.locator('.archived-sessions>summary').click();
      await page.waitForFunction(() => document.querySelectorAll('.archived-sessions .conversation-card').length === 56);
      assert.equal(await page.locator('#conversation-list>.conversation-card').filter({hasText:'重命名后的会话'}).count(), 0);
      const archived = page.locator('.archived-sessions .conversation-card').filter({hasText:'重命名后的会话'});
      await archived.click({button:'right'}); await page.locator('#session-menu-archive').click();
      await current.waitFor();
      await page.evaluate(() => { fixture.session.taskStatus = 'done'; });
      await current.click({button:'right'}); await page.locator('#session-menu-delete').click();
      await page.waitForFunction(() => state.session === null);
      assert.equal(await current.count(), 0);
      if (width < 760) await page.keyboard.press('Escape');

      if (width < 760) await page.locator('#open-devices').click();
      await page.locator('.project-settings').click();
      assert.equal(await page.locator('#device-menu>button').first().textContent(), '设置');
      await page.locator('#device-menu-settings').click();
      await page.locator('#project-settings-form [name=name]').fill('新项目名称');
      await page.locator('#project-settings-form [name=workspace]').fill('C:\\Code\\new');
      await page.locator('#project-settings-form button[type=submit]').click();
      await page.locator('#project-settings-dialog').waitFor({state:'hidden'});
      assert.equal(await page.evaluate(() => fixture.project.id), 'project');
      assert.equal(await page.evaluate(() => fixture.project.name), '新项目名称');
      assert.equal(await page.evaluate(() => fixture.project.workspace), 'C:\\Code\\new');
      assert.equal(await page.evaluate(() => document.documentElement.scrollWidth), width);
      assert.deepEqual(errors, []);
      await page.close();
    }
    console.log('PASS: desktop/mobile project settings; live-session queueing and promotion; edit recovery preserves drafts; right-click/keyboard rename, archive, restore, delete; all archive pages load');
  } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode = 1; });
