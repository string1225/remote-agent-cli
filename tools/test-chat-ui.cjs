const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');

(async () => {
  const browser = await chromium.launch({headless:true, channel:process.env.PLAYWRIGHT_CHANNEL || 'msedge'});
  try {
    for (const width of [1440, 390]) {
      const page = await browser.newPage({viewport:{width, height:900}}), errors = [];
      page.setDefaultTimeout(5000);
      page.on('pageerror', error => errors.push(error.message));
      await page.addInitScript(() => { window.setInterval = () => 0; });
      await page.route('https://chat-ui.test/**', route => {
        const file = new URL(route.request().url()).pathname.slice(1) || 'index.html';
        if (file.startsWith('api/')) return route.fulfill({status:401, json:{error:'please sign in'}});
        const filename = file === 'app.js' && process.env.UI_TEST_APP_SOURCE
          ? process.env.UI_TEST_APP_SOURCE : path.join('web', file);
        const contentType = file.endsWith('.css') ? 'text/css' : file.endsWith('.js') ? 'text/javascript' : 'text/html';
        return route.fulfill({contentType, body:fs.readFileSync(filename)});
      });
      await page.goto('https://chat-ui.test/');
      await page.waitForFunction(() => document.getElementById('notice').textContent === '');
      await page.evaluate(() => {
        state.user = {id:'test-user'};
        state.selected = {id:'test-client', name:'测试客户端'};
        state.services = [{id:'project', name:'remote-agent-cli', provider:'codex', workspace:'C:\\Users\\junte\\Code'}];
        state.serviceId = 'project';
        state.session = {id:'test-session', serviceId:'project', title:'会话显示验证', taskId:'task', taskStatus:'running'};
        state.sessions = [state.session];
        state.peer = {ws:{readyState:WebSocket.OPEN}};
        state.permissionMode = 'full';
        state.entries = [
          {seq:1, turnId:'task', role:'user', kind:'message', text:'更换到 C:\\Users\\junte\\Code'},
          {seq:2, turnId:'task', role:'assistant', kind:'stdout', text:'我将把后续命令的工作目录切换到 `C:\\Users\\junte\\Code`，并确认目录可用。\n\n已切换到 `C:\\Users\\junte\\Code`，后续命令将在此目录执行。'},
          {seq:3, turnId:'task', role:'system', kind:'tool', text:'$ "C:\\Users\\junte\\.cache\\codex-runtimes\\codex-primary-runtime\\dependencies\\native\\powershell\\pwsh.exe" -Command Get-Location\nPRIVATE TOOL OUTPUT'},
          {seq:4, turnId:'task', role:'system', kind:'tool', text:'PRIVATE CONTINUATION'},
        ];
        $('auth').hidden = true;
        $('app').hidden = false;
        $('prompt').value = '保留的下一条草稿';
        renderChat();
      });
      const toggle = page.locator('.run-details-toggle'), details = page.locator('.run-details');
      assert.equal(await page.locator('#send-run').isDisabled(), true);
      assert.equal(await details.isVisible(), false);
      assert.equal(await page.locator('.message.system').count(), 0);
      assert(!(await page.locator('#messages').textContent()).includes('PRIVATE'));
      await toggle.click();
      await page.mouse.move(0, 0);
      assert.equal(await toggle.getAttribute('aria-expanded'), 'true');
      assert.equal(await details.isVisible(), true);
      const layout = await page.locator('.message.assistant').evaluate(article => {
        const header = article.querySelector('.message-label'), toggle = article.querySelector('button');
        const details = article.querySelector('.run-details'), reply = article.querySelector('.message-text');
        return {
          header:header.getBoundingClientRect().toJSON(), toggle:toggle.getBoundingClientRect().toJSON(),
          details:details.getBoundingClientRect().toJSON(), reply:reply.getBoundingClientRect().toJSON(),
          labelFont:parseFloat(getComputedStyle(header).fontSize), toggleFont:parseFloat(getComputedStyle(toggle).fontSize),
          toggleColor:getComputedStyle(toggle).color, detailsColor:getComputedStyle(details).color,
          border:getComputedStyle(details).borderBottomWidth,
        };
      });
      assert(layout.toggle.x > layout.header.x, 'toggle follows the Codex label');
      assert(layout.toggle.y < layout.header.bottom && layout.toggle.bottom > layout.header.y);
      assert(layout.details.y >= layout.header.bottom);
      assert(layout.reply.y >= layout.details.bottom);
      assert(layout.toggleFont < layout.labelFont);
      assert.equal(layout.toggleColor, 'rgb(139, 150, 168)');
      assert.equal(layout.detailsColor, layout.toggleColor);
      assert.equal(layout.border, '1px');
      await page.evaluate(async () => {
        rpc = async type => {
          if (type !== 'sessions.list') throw Error('History temporarily unavailable');
          return {sessions:[{...state.session, taskStatus:'done'}], nextCursor:0};
        };
        await refreshSessions(false, true);
        renderChat();
      });
      assert.equal(await page.locator('#send-run').isEnabled(), true, 'completion restores send without navigation');
      assert.equal(await page.locator('#cancel-run').isVisible(), false);
      assert.equal(await page.locator('#prompt').inputValue(), '保留的下一条草稿');
      assert.equal(await details.isVisible(), true, 'refresh preserves expanded details');
      if (process.env.UI_TEST_SCREENSHOT_DIR) {
        fs.mkdirSync(process.env.UI_TEST_SCREENSHOT_DIR, {recursive:true});
        await page.screenshot({path:path.join(process.env.UI_TEST_SCREENSHOT_DIR, `chat-${width}.png`)});
      }
      await toggle.press('Enter');
      assert.equal(await details.isVisible(), false);
      await toggle.press('Space');
      assert.equal(await details.isVisible(), true);
      assert.equal(await page.evaluate(() => document.documentElement.scrollWidth), width, 'no horizontal overflow');
      assert.deepEqual(errors, []);
      await page.close();
    }
    console.log('PASS: desktop/mobile layout, inline run details, keyboard toggle, expansion survives refresh, completion restores sending without navigation, tool output remains hidden');
  } finally {
    await browser.close();
  }
})().catch(error => { console.error(error); process.exitCode = 1; });
