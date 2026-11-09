const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {chromium} = require(process.env.PLAYWRIGHT_MODULE || 'playwright');

(async () => {
  const browser = await chromium.launch({headless:true,channel:process.env.PLAYWRIGHT_CHANNEL || 'msedge'});
  try {
    for (const width of [1440,390]) {
      const context = await browser.newContext({viewport:{width,height:960},reducedMotion:'reduce'});
      const page = await context.newPage(), errors = [];
      page.on('pageerror', error => errors.push(error.message));
      let agents = [
        {id:'one',name:'工作电脑',os:'windows',online:false,services:[{id:'active',name:'使用中的项目',provider:'codex',workspace:'C:/work'},{id:'hidden',name:'旧项目',provider:'codex',archivedAt:100}]},
        {id:'two',name:'离线主机',os:'linux',online:false,services:[{id:'other',name:'<img src=x onerror=alert(1)>',provider:'qoder',archivedAt:200}]},
      ];
      let failRestore = false, holdRead = false, delayedRead, delayedWrite, holdWrite = false;
      await context.addInitScript(() => { window.setInterval = () => 0; });
      await context.route('https://preferences.test/**', async route => {
        const file = new URL(route.request().url()).pathname.slice(1) || 'index.html';
        if (file === 'api/me') return route.fulfill({json:{id:'owner',username:'测试用户',admin:false}});
        if (file === 'api/agents') {
          const snapshot = structuredClone(agents);
          if (holdRead) { holdRead = false; await new Promise(resolve => { delayedRead = resolve; }); }
          return route.fulfill({json:snapshot});
        }
        const match = file.match(/^api\/agents\/(.+)\/projects\/(.+)\/archive$/);
        if (match) {
          const body = route.request().postDataJSON();
          if (failRestore) return route.fulfill({status:500,json:{error:'模拟保存失败'}});
          if (holdWrite) { holdWrite = false; await new Promise(resolve => { delayedWrite = resolve; }); }
          const project = agents.find(a => a.id === match[1]).services.find(p => p.id === match[2]);
          project.archivedAt = body.archived ? Date.now() : 0;
          return route.fulfill({json:project});
        }
        if (file === 'api/codex') return route.fulfill({json:{enabled:false}});
        if (file === 'api/default-host') return route.fulfill({json:{agentId:'one',online:false}});
        if (file === 'api/tokens') return route.fulfill({json:[]});
        if (file.startsWith('api/')) return route.fulfill({status:409,json:{error:'模拟环境'}});
        return route.fulfill({contentType:file.endsWith('.css')?'text/css':file.endsWith('.js')?'text/javascript':'text/html',body:fs.readFileSync(path.join('web',file))});
      });
      const openMenu = async () => {
        if (width === 390 && !(await page.locator('#account-toggle').isVisible())) await page.locator('#open-devices').click();
        await page.locator('#account-toggle').click();
      };
      await page.goto('https://preferences.test/');
      await page.waitForFunction(()=>state.agents.length === 2);
      assert.equal(await page.locator('html').getAttribute('data-theme'),'dark','dark is the first-visit default');
      await page.evaluate(async()=>{await connectDevice(state.agents[0]);state.session={id:'history',serviceId:'active',title:'原历史'};state.sessions=[state.session];$('prompt').value='未发送的草稿';renderDevices();});
      assert.equal(await page.locator('.project-button').count(),1,'archived projects are hidden even when the client returns its unfiltered catalog');
      await openMenu(); await page.locator('#appearance-entry').click();
      await page.getByLabel('明亮主题',{exact:true}).check();
      assert.equal(await page.evaluate(()=>localStorage.getItem('remote-agent.theme')),'light');
      await page.reload(); await page.locator('#app').waitFor({state:'visible'});
      assert.equal(await page.locator('html').getAttribute('data-theme'),'light','refresh preserves the saved choice');
      await openMenu(); await page.locator('#appearance-entry').click();
      await page.getByLabel('暗黑主题',{exact:true}).check();
      if (process.env.UI_TEST_SCREENSHOT_DIR) {
        fs.mkdirSync(process.env.UI_TEST_SCREENSHOT_DIR,{recursive:true});
        await page.screenshot({path:path.join(process.env.UI_TEST_SCREENSHOT_DIR,`appearance-${width}.png`)});
      }
      await page.getByRole('button',{name:'关闭界面设置',exact:true}).click();
      await page.evaluate(async()=>{await connectDevice(state.agents[0]);state.session={id:'history',serviceId:'active',title:'原历史'};state.sessions=[state.session];$('prompt').value='未发送的草稿';renderDevices();});
      if (width === 390 && !(await page.locator('.project-settings').isVisible())) await page.locator('#open-devices').click();
      // A catalog read started before the write cannot undo the archive.
      holdRead = true;
      await page.evaluate(()=>{window.pendingCatalogRead = refresh();});
      await page.waitForFunction(()=>true);
      await page.locator('.project-settings').click(); await page.locator('#device-menu-archive').click();
      await page.waitForFunction(()=>!!state.agents[0].services[0].archivedAt);
      assert(delayedRead); delayedRead(); await page.evaluate(()=>window.pendingCatalogRead);
      assert.equal(await page.locator('.project-button').count(),0);
      assert.equal(await page.locator('#prompt').inputValue(),'未发送的草稿','archiving leaves the open conversation and draft intact');
      await openMenu(); await page.locator('#archived-projects-entry').click();
      await page.waitForFunction(()=>document.querySelectorAll('.archived-project-row').length === 3);
      assert.equal(await page.locator('.archived-project-row img').count(),0,'project names remain inert text');
      assert.match(await page.locator('.project-archive-list').textContent(),/离线主机.*Qoder.*离线/);
      const restore = page.getByRole('button',{name:'恢复 <img src=x onerror=alert(1)>（离线主机）',exact:true});
      failRestore = true; await restore.click();
      await page.getByRole('status').filter({hasText:'模拟保存失败'}).waitFor();
      assert.equal(await page.locator('.archived-project-row').count(),3,'failed writes retain the archive');
      failRestore = false; await restore.click();
      await page.waitForFunction(()=>document.querySelectorAll('.archived-project-row').length === 2);
      assert.equal(await page.evaluate(()=>state.selected.id),'one','restoring another device never switches the current workspace');
      if (process.env.UI_TEST_SCREENSHOT_DIR) await page.screenshot({path:path.join(process.env.UI_TEST_SCREENSHOT_DIR,`archives-${width}.png`)});
      await page.getByRole('button',{name:'关闭已归档项目',exact:true}).click();
      // Reconnecting with an archived saved selection chooses an active project.
      await page.evaluate(async()=>{disconnect();await connectDevice(state.agents[0]);});
      assert.equal(await page.evaluate(()=>state.serviceId),null);
      await page.evaluate(async()=>{await connectDevice(state.agents[0],{projectId:'active',sessionId:'history'});});
      assert.equal(await page.evaluate(()=>state.serviceId),'active','existing history deep links remain accessible');
      // Theme storage events synchronize tabs, without depending on a reload.
      const second = await context.newPage(); await second.goto('https://preferences.test/');
      await second.evaluate(()=>{localStorage.setItem('remote-agent.theme','light');});
      await page.waitForFunction(()=>document.documentElement.dataset.theme === 'light');
      await second.evaluate(()=>{localStorage.removeItem('remote-agent.theme');});
      await page.waitForFunction(()=>document.documentElement.dataset.theme === 'dark'); await second.close();
      await page.evaluate(async()=>{
        const module = await import('./sayso/sayso.js');
        const now = new Date().toISOString();
        state.mode='sayso';closeDrawers();renderChat();
        $('sayso-root').hidden=false;
        window.disposeSaysoPreview=module.mount($('sayso-root'), async()=>({projects:[{id:'preview',name:'讨论项目',rootPath:'/preview',color:'#527d48',createdAt:now}],sessions:[{id:'discussion',projectId:'preview',title:'需求讨论',status:'paused',startedAt:now,recordings:[],transcript:[{id:'line',text:'保留历史讨论与录音，继续整理下一步行动。',atMs:0,createdAt:now}]}],actions:[],runs:[],settings:{analysisProvider:'codex',analysisModel:'default',analysisIntervalMinutes:2,executor:'codex',executionModel:'default'}}),()=>{},'测试',{projectId:'preview',sessionId:'discussion'});
      });
      await page.locator('#sayso-root .app-shell').waitFor({state:'attached'});
      await page.locator('#sayso-root .utterance>p').waitFor({state:'visible'});
      assert.equal(await page.locator('#sayso-root').getAttribute('data-theme'),'dark');
      assert.equal(await page.locator('#sayso-root .workspace').evaluate(el=>getComputedStyle(el).backgroundColor),'rgb(25, 31, 41)');
      for (const selector of ['.session-title-line h1','.panel-header h2','.utterance>p']) {
        assert.equal(await page.locator(`#sayso-root ${selector}`).first().evaluate(el=>getComputedStyle(el).color),'rgb(223, 229, 239)','dark SaySo headings and transcript remain legible');
      }
      if (process.env.UI_TEST_SCREENSHOT_DIR) await page.screenshot({path:path.join(process.env.UI_TEST_SCREENSHOT_DIR,`sayso-dark-${width}.png`)});
      await page.evaluate(()=>{window.openAppearanceSettings();document.querySelector('input[value="light"]').click();});
      assert.equal(await page.locator('#sayso-root').getAttribute('data-theme'),'light','mounted SaySo follows theme changes');
      if (process.env.UI_TEST_SCREENSHOT_DIR) await page.screenshot({path:path.join(process.env.UI_TEST_SCREENSHOT_DIR,`appearance-light-${width}.png`)});
      await page.evaluate(()=>{$('appearance-dialog').close();window.disposeSaysoPreview();$('sayso-root').hidden=true;});
      // Late responses from a previous account cannot change the new account catalog.
      holdWrite = true;
      await page.evaluate(()=>{window.pendingArchive=window.archiveProject(state.agents[0],state.agents[0].services[0]);});
      while (!delayedWrite) await new Promise(resolve=>setTimeout(resolve,10));
      await page.evaluate(()=>{state.user={id:'different'};state.agents=[];}); delayedWrite();
      await page.evaluate(()=>window.pendingArchive);
      assert.equal(await page.evaluate(()=>state.agents.length),0);
      assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth <= innerWidth),true,'no horizontal page overflow');
      assert.deepEqual(errors,[]); await context.close();
    }
    const page = await browser.newPage();
    await page.addInitScript(()=>{Storage.prototype.getItem=()=>{throw Error('blocked')};Storage.prototype.setItem=()=>{throw Error('blocked')};window.setInterval=()=>0;});
    await page.route('https://blocked.test/**',route=>{
      const file=new URL(route.request().url()).pathname.slice(1)||'index.html';
      if(file.startsWith('api/'))return route.fulfill({status:401,json:{error:'sign in'}});
      return route.fulfill({contentType:file.endsWith('.css')?'text/css':file.endsWith('.js')?'text/javascript':'text/html',body:fs.readFileSync(path.join('web',file))});
    });
    await page.goto('https://blocked.test/');
    await page.evaluate(()=>window.openAppearanceSettings()); await page.getByLabel('明亮主题',{exact:true}).check();
    assert.equal(await page.locator('html').getAttribute('data-theme'),'light');
    assert.match(await page.locator('.theme-status').textContent(),/存储不可用/);
    await page.close();
    console.log('PASS: desktop/mobile archive, offline restore, failure/stale/account isolation, history/drafts, default/persisted/cross-tab/blocked-storage themes and SaySo');
  } finally {await browser.close();}
})().catch(error=>{console.error(error);process.exitCode=1;});
