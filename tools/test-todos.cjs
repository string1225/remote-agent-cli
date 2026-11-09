const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {chromium} = require(process.env.PLAYWRIGHT_MODULE || 'playwright');

(async () => {
  const browser = await chromium.launch({headless:true, channel:process.env.PLAYWRIGHT_CHANNEL || 'chrome'});
  try {
    for (const width of [1440,390]) {
      const page = await browser.newPage({viewport:{width,height:900}, userAgent:width === 390 ? 'RemoteAgentAndroid/1.0' : undefined});
      const errors = [], calls = []; let version = 1, status = 'open', title = '<img src=x onerror=alert(1)>判断发布窗口', reply = null, empty = false;
      const todo = () => ({id:'todo',title,body:'**目标**：明确发布窗口。\n- **决策**：周五或周六发布。\n- **验收**：保留回滚并检查线上健康。\n[测试链接](javascript:alert(1))',status,version,clientId:'client',projectId:'project',sessionId:'session-with-a-long-persistent-identifier',reply});
      page.on('pageerror', e => errors.push(e.message));
      await page.addInitScript(() => { window.setInterval = () => 0; });
      await page.route('https://todo.test/**', async route => {
        const file = new URL(route.request().url()).pathname.slice(1) || 'index.html';
        if (file === 'api/me') return route.fulfill({json:{id:'alice',username:'Alice'}});
        if (file === 'api/agents') return route.fulfill({json:[{id:'client',name:'工作电脑',enrolled:true,online:true,os:'windows',services:[{id:'project',name:'项目',provider:'codex'}]}]});
        if (file === 'api/default-host') return route.fulfill({json:{agentId:'client',name:'工作电脑',online:true}});
        if (file === 'api/todos') {
          if (route.request().method() === 'GET') {
            const params = new URL(route.request().url()).searchParams;
            return route.fulfill({json:{todos:empty || params.get('status') && params.get('status') !== status || params.get('search') && !title.includes(params.get('search')) ? [] : [todo()]}});
          }
          const inData = route.request().postDataJSON(); calls.push(inData);
          assert.equal(inData.version, version, 'UI must use the latest todo version'); version++;
          if (inData.operation === 'update') title = inData.title;
          if (inData.operation === 'complete') status = 'completed';
          if (inData.operation === 'reopen') status = 'open';
          if (inData.operation === 'cancel') status = 'cancelled';
          if (inData.operation === 'reply') reply = {id:inData.requestId,text:inData.message,state:'delivered'};
          return route.fulfill({json:todo()});
        }
        if (file.startsWith('api/')) return route.fulfill({status:409,json:{error:'此测试仅模拟待办接口'}});
        const target = path.join('web', file);
        return route.fulfill({contentType:file.endsWith('.js') ? 'text/javascript' : file.endsWith('.css') ? 'text/css' : file.endsWith('.svg') ? 'image/svg+xml' : 'text/html',body:fs.readFileSync(target)});
      });
      await page.goto('https://todo.test/'); await page.locator('#app').waitFor({state:'visible'});
      if (width === 390) {
        assert.equal(await page.locator('#new-agent').isVisible(),false,'Android does not create clients');
        await page.locator('#open-devices').click();
      }
      await page.locator('#todos-entry').click();
      const dialog = page.locator('#todos-dialog'); await dialog.locator('.todo-card').waitFor();
      assert.equal(await dialog.locator('h3 img').count(),0,'todo titles render as text');
      assert.match(await dialog.locator('h3').textContent(),/onerror/);
      assert.equal(await dialog.locator('.todo-body strong').count(),3,'key points render as emphasized Markdown');
      assert.equal(await dialog.locator('.todo-body a[href^="javascript:"]').count(),0,'Markdown links must remain sanitized');
      const layout = await dialog.evaluate(el => {
        const box = selector => el.querySelector(selector).getBoundingClientRect().toJSON();
        return {search:box('input[type=search]'),filter:box('select'),refresh:box('.todo-toolbar button'),header:box('.todo-dialog-header'),host:box('.todo-host'),create:box('.todo-header-actions .primary')};
      });
      assert.equal(layout.filter.y,layout.refresh.y,'status and refresh align');
      assert.equal(layout.search.height,layout.filter.height,'controls share a height');
      assert.equal(layout.filter.height,layout.refresh.height);
      if (width === 1440) assert.equal(layout.search.y,layout.filter.y,'desktop toolbar stays on one row');
      assert(layout.host.bottom <= layout.header.bottom,'default host stays inside the heading');
      assert(layout.host.right <= layout.create.x,'default host sits beside the new-todo action');
      assert.equal(await dialog.locator('.todo-count').textContent(),'1 项');
      assert.equal(await dialog.locator('.todo-session-ref').getAttribute('title'),'Session session-with-a-long-persistent-identifier');
      await dialog.getByRole('button',{name:'回复',exact:true}).click(); await dialog.getByRole('textbox',{name:'回复内容'}).fill('周五发布，保留回滚');
      await dialog.getByRole('button',{name:'发送到原会话'}).click(); await dialog.getByText('上次回复 · 已送入原会话：周五发布，保留回滚').waitFor();
      await dialog.getByRole('button',{name:'编辑',exact:true}).click();
      const editor = page.locator('.todo-editor'); await editor.getByRole('textbox',{name:'待办标题'}).fill('已确认发布窗口'); await editor.getByRole('button',{name:'保存',exact:true}).click(); await editor.waitFor({state:'detached'});
      const filter = dialog.getByRole('combobox',{name:'待办状态'});
      await dialog.getByRole('button',{name:'完成',exact:true}).click(); await dialog.getByRole('heading',{name:'暂无待处理任务'}).waitFor();
      await filter.selectOption('completed'); await dialog.getByRole('button',{name:'重新打开'}).click();
      await filter.selectOption('open'); await dialog.getByRole('button',{name:'取消',exact:true}).click();
      await filter.selectOption('cancelled'); await dialog.getByRole('button',{name:'重新打开'}).waitFor();
      assert.equal(calls.filter(c=>c.operation==='reply').length,1); assert.equal(status,'cancelled');
      assert.equal(await dialog.evaluate(el=>el.scrollWidth<=el.clientWidth),true,'todo dialog must fit phone width');
      if (process.env.UI_TEST_SCREENSHOT_DIR) { fs.mkdirSync(process.env.UI_TEST_SCREENSHOT_DIR,{recursive:true}); await dialog.screenshot({path:path.join(process.env.UI_TEST_SCREENSHOT_DIR,`todos-${width}.png`)}); }
      await dialog.getByRole('searchbox',{name:'搜索待办'}).fill('没有这个任务');
      await dialog.getByRole('heading',{name:'没有匹配的待办'}).waitFor();
      empty = true;
      await dialog.getByRole('searchbox',{name:'搜索待办'}).fill(''); await filter.selectOption('open');
      await dialog.getByRole('heading',{name:'暂无待处理任务'}).waitFor();
      assert.equal(await dialog.locator('.todo-count').textContent(),'0 项');
      if (process.env.UI_TEST_SCREENSHOT_DIR) await dialog.screenshot({path:path.join(process.env.UI_TEST_SCREENSHOT_DIR,`todos-empty-${width}.png`)});
      await dialog.getByRole('button',{name:'＋ 新建待办',exact:true}).click();
      await page.locator('.todo-editor').getByRole('textbox',{name:'待办标题'}).waitFor();
      await page.locator('.todo-editor').getByRole('button',{name:'取消',exact:true}).click();
      assert.equal(await dialog.evaluate(el=>el.scrollWidth<=el.clientWidth),true,'empty state and toolbar must fit phone width');
      assert.deepEqual(errors,[]); await page.close();
    }
    console.log('PASS: desktop/Android todo list, safe rendering, reply, edit, complete, cancel, current versions and phone layout');
  } finally { await browser.close(); }
})().catch(e=>{console.error(e);process.exitCode=1});
