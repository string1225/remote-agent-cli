const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {chromium} = require(process.env.PLAYWRIGHT_MODULE || 'playwright');

(async () => {
  const browser = await chromium.launch({headless:true, channel:process.env.PLAYWRIGHT_CHANNEL || 'msedge'});
  try {
    for (const width of [1440, 390]) {
      const page = await browser.newPage({viewport:{width,height:960}}), errors = [], requests = [];
      let token = '', nextMetadata;
      page.on('pageerror', error => errors.push(error.message));
      page.on('dialog', dialog => dialog.accept());
      await page.addInitScript(() => { window.setInterval = () => 0; });
      await page.route('https://project-api.test/**', async route => {
        const request = route.request(), file = new URL(request.url()).pathname.replace(/^\/agents\//,'') || 'index.html';
        if (file === 'api/me') return route.fulfill({json:{id:'owner',username:'Owner'}});
        if (file === 'api/agents') return route.fulfill({json:[]});
        if (file === 'api/default-host') return route.fulfill({json:{agentId:'',online:false}});
        if (file === 'api/coordination' || file === 'api/assistants') return route.fulfill({status:409,json:{error:'No default host in test fixture'}});
        if (file.includes('/session-token')) {
          requests.push({method:request.method(),file,body:request.postDataJSON()});
          if (request.method() === 'GET') {
            if (nextMetadata) { const wait = nextMetadata; nextMetadata = undefined; await wait; }
            return route.fulfill({json:{configured:!!token}});
          }
          if (request.method() === 'PUT') { token = request.postDataJSON().token || 'ra_project_generated-test-token-0123456789'; return route.fulfill({json:{configured:true,token}}); }
          if (request.method() === 'DELETE') { token = ''; return route.fulfill({json:{configured:false}}); }
          return route.fulfill({json:{token}});
        }
        const source = path.join(process.env.UI_TEST_SOURCE || 'web', file);
        return route.fulfill({contentType:file.endsWith('.js')?'text/javascript':file.endsWith('.css')?'text/css':file.endsWith('.svg')?'image/svg+xml':'text/html',body:fs.readFileSync(source)});
      });
      await page.goto('https://project-api.test/agents/');
      await page.locator('#app').waitFor({state:'visible'});
      async function open(project='project') {
        await page.evaluate(async id => {
          const project={id,name:'API 项目',workspace:'C:\\Code',provider:'codex'};
          state.peer = {}; state.projectSettings = true; state.roots = ['C:\\Code']; state.services = [project];
          connectDevice = async () => {};
          await window.openProjectSettings({id:'client'},project);
        },project);
        await page.locator('.project-session-api').waitFor();
        await page.locator('.project-session-api').evaluate(element => { element.open = true; });
      }
      await open();
      const status = page.locator('.project-session-api [data-status]');
      await page.waitForFunction(() => document.querySelector('[data-status]').textContent.includes('尚未配置'));
      assert.equal(await page.locator('.project-session-api [data-copy]').isDisabled(),true);
      const endpoint = 'https://project-api.test/agents/api/agents/client/projects/project/sessions';
      assert.equal(await page.locator('[data-endpoint]').inputValue(),endpoint);
      assert((await page.locator('[data-example]').inputValue()).includes('"message"'));
      assert((await page.locator('[data-example]').inputValue()).includes('"execute":true'));
      assert(!(await page.locator('[data-example]').inputValue()).includes('\n+'));
      await page.getByRole('button',{name:'生成 Token',exact:true}).click();
      await page.waitForFunction(() => document.querySelector('[data-token]').value.length > 0);
      assert.equal(await page.locator('[data-token]').getAttribute('type'),'password');
      await page.getByRole('button',{name:'查看 Token',exact:true}).click();
      assert.equal(await page.locator('[data-token]').getAttribute('type'),'text');
      await page.getByRole('button',{name:'隐藏 Token',exact:true}).click();
      const custom='custom-project-token-0123456789-abcdef';
      await page.locator('[data-custom] input').fill(custom);
      await page.getByRole('button',{name:'保存自定义 Token',exact:true}).click();
      await page.waitForFunction(expected => document.querySelector('[data-token]').value === expected,custom);
      assert.equal(await page.locator('[data-custom] input').inputValue(),'');
      assert(!(await page.locator('[data-example]').inputValue()).includes(custom));
      if (process.env.UI_TEST_SCREENSHOT_DIR) {
        fs.mkdirSync(process.env.UI_TEST_SCREENSHOT_DIR,{recursive:true});
        await page.screenshot({path:path.join(process.env.UI_TEST_SCREENSHOT_DIR,`project-session-api-${width}.png`)});
      }
      await page.locator('#project-settings-dialog .close-dialog').click();
      await page.waitForFunction(() => document.querySelector('[data-token]').value === '');
      assert.equal(await page.locator('[data-token]').inputValue(),'');
      await open();
      await page.waitForFunction(() => document.querySelector('[data-status]').textContent.includes('已配置'));
      await page.getByRole('button',{name:'查看 Token',exact:true}).click();
      await page.waitForFunction(expected => document.querySelector('[data-token]').value === expected,custom);
      await page.getByRole('button',{name:'撤销 Token',exact:true}).click();
      await page.waitForFunction(() => document.querySelector('[data-status]').textContent.includes('已撤销'));
      assert.equal(await page.locator('[data-token]').inputValue(),'');
      assert.equal(await page.locator('.project-session-api [data-copy]').isDisabled(),true);
      await page.locator('#project-settings-dialog .close-dialog').click();
      let release;
      nextMetadata = new Promise(resolve => { release=resolve; });
      await open('slow-project');
      await page.waitForFunction(() => document.querySelector('[data-status]').textContent.includes('正在读取'));
      await page.locator('#project-settings-dialog .close-dialog').click();
      release();
      await page.waitForTimeout(80);
      assert.equal(await page.locator('[data-token]').inputValue(),'');
      assert.equal(await status.textContent(),'');
      assert.equal(await page.evaluate(() => document.documentElement.scrollWidth),width);
      assert.deepEqual(errors,[]);
      assert(requests.some(row => row.method === 'PUT' && row.body.token === custom));
      await page.close();
    }
    console.log('PASS: desktop/mobile project session API settings; generate/custom token/reveal/revoke; subpath endpoint and message/execute example; secrets cleared on close and delayed response ignored');
  } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode=1; });
