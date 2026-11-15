const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {chromium} = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const web = path.resolve(process.env.UI_WEB_ROOT || 'web');
const screenshots = process.env.UI_TEST_SCREENSHOT_DIR;
const baseline = process.env.UI_TEST_BASELINE === '1';

async function fixture(page) {
  await page.addInitScript(() => { window.setInterval = () => 0; localStorage.setItem('remote-agent.theme','light'); });
  const projects = [{id:'project',name:'Remote Agent',provider:'codex'}, {id:'notes',name:'产品笔记',provider:'codex'}, {id:'tools',name:'桌面工具',provider:'qoder'}];
  const clients = [
    {id:'client',name:'主力工作站',os:'windows',online:true,enrolled:true,services:projects},
    {id:'laptop',name:'随身笔记本',os:'darwin',online:false,enrolled:true,services:[{id:'lab',name:'实验项目',provider:'codex'}]},
    {id:'linux',name:'构建服务器',os:'linux',online:false,enrolled:true,services:[]},
  ];
  const now = Date.now();
  const session = {id:'current',serviceId:'project',title:'工作台布局与运行状态',taskId:'current-task',taskStatus:'running',taskAcceptedAt:now-15000,taskStartedAt:now-14000,createdAt:now-500000,updatedAt:now};
  const entries = [
    {seq:1,turnId:'previous-task',role:'user',kind:'message',text:'检查一下客户端连接和更新渠道。',createdAt:now-90000},
    {seq:2,turnId:'previous-task',role:'assistant',kind:'stdout',text:'客户端连接正常，更新渠道已核对。原有会话与工作区都已保留。',createdAt:now-88000},
    {seq:3,turnId:'previous-task',role:'system',kind:'done',text:'',createdAt:now-82000},
    {seq:4,turnId:'current-task',role:'user',kind:'message',text:'让工作台紧凑一些，把运行状态放在最新回答的底部。',createdAt:now-15000},
    {seq:5,turnId:'current-task',role:'system',kind:'tool',text:'$ node --check web/workbench.js\n',activity:{id:'check',name:'检查页面与交互',kind:'command_execution',status:'running',startedAt:now-13000},createdAt:now-13000},
    {seq:6,turnId:'current-task',role:'assistant',kind:'stdout',text:'已把客户端管家移到客户端行右侧，并压缩导航和会话列表的间距。\n\n正在检查运行状态、完成时间，以及浅色和深色主题下的显示。',createdAt:now-10000},
  ];
  await page.route('https://compact.test/**', route => {
    const file = new URL(route.request().url()).pathname.slice(1) || 'index.html';
    const data = file === 'api/me' ? {id:'demo-owner',username:'演示工作区',admin:false}
      : file === 'api/agents' ? clients : file === 'api/todos' ? {todos:[]}
      : file === 'api/default-host' ? {agentId:'client',name:'主力工作站',online:true}
      : file === 'api/assistants' ? {assistants:[]} : null;
    if (data) return route.fulfill({json:data});
    if (file.startsWith('api/')) return route.fulfill({json:{}});
    const target = path.join(web,file);
    if (!fs.existsSync(target)) return route.fulfill({status:404,body:''});
    return route.fulfill({contentType:file.endsWith('.css')?'text/css':file.endsWith('.js')?'text/javascript':file.endsWith('.svg')?'image/svg+xml':'text/html',body:fs.readFileSync(target)});
  });
  await page.routeWebSocket('**/api/agents/*/connect', socket => {
    socket.onMessage(wire => {
      const message = JSON.parse(wire); if (message.type === 'ping') { socket.send(JSON.stringify({type:'pong'})); return; }
      const request = message.request; if (!request) return;
      let data = {};
      if (request.type === 'services.list') data = {services:projects,providers:{codex:{available:true},qoder:{available:true}},permissionMode:'full',taskQueue:true,backgroundRuns:true,sessionManagement:true,runActivities:true,workspaceReview:true,clientVersion:'fixture'};
      if (request.type === 'sessions.list') data = {sessions:[session,...['项目目录检查','发布验收记录','继续昨日工作','新界面讨论'].map((title,index)=>({id:'other-'+index,serviceId:'project',title,taskStatus:'done',createdAt:now-(index+1)*900000,updatedAt:now-(index+1)*900000}))],nextCursor:0};
      if (request.type === 'sessions.get') data = {session,entries,nextCursor:0};
      socket.send(JSON.stringify({type:'event',event:{id:request.id,type:'result',data}}));
    });
    socket.send(JSON.stringify({type:'ready',transport:'relay-v1'}));
  });
  await page.goto('https://compact.test/#client=client&project=project&session=current');
  await page.waitForFunction(()=>state.session?.id==='current' && state.entries.some(entry=>entry.seq===6));
  await page.evaluate(()=>{renderChat(); renderDevices();});
  await page.waitForFunction(()=>innerWidth<=760 || document.getElementById('app').style.getPropertyValue('--devices-width').endsWith('px'));
}

(async () => {
  const browser = await chromium.launch({headless:true,channel:process.env.PLAYWRIGHT_CHANNEL || 'msedge'});
  try {
    if (screenshots) fs.mkdirSync(screenshots,{recursive:true});
    for (const width of [1440,1280,1279,1024,390]) {
      const page = await browser.newPage({viewport:{width,height:960},timezoneId:'Asia/Shanghai'}), errors=[];
      page.setDefaultTimeout(6000); page.on('pageerror',error=>errors.push(error.message));
      try { await fixture(page); } catch(error) {
        console.error('Fixture diagnostics',errors,await page.evaluate(()=>({notice:document.getElementById('notice').textContent,user:state.user?.id,client:state.selected?.id,project:state.serviceId,session:state.session?.id,entries:state.entries.length,hash:location.hash})));
        throw error;
      }
      if (baseline) {
        if (screenshots) await page.screenshot({path:path.join(screenshots,`before-${width}.png`),fullPage:true});
        await page.close(); continue;
      }
      if (width <= 760) await page.locator('#open-devices').click();
      const clientRow=page.locator('.device-node').first().locator('.device-row');
      assert.equal(await clientRow.locator('.device-add').isVisible(),width>=1280,'inline add moves into menu below 1280px');
      for (const button of [clientRow.locator('.client-steward'),page.locator('.project-steward').first()]) {
        assert.equal(await button.textContent(),'');
        assert.equal(await button.locator('svg[aria-hidden=true]').count(),1);
        assert.match(await button.getAttribute('aria-label'),/打开.*管家/);
      }
      await clientRow.locator('.device-menu-button').click();
      assert.equal(await page.locator('#device-menu-add-project').isVisible(),width<1280);
      if (width<1280) await page.locator('#device-menu-add-project').click();
      else { await page.keyboard.press('Escape');await clientRow.locator('.device-add').click(); }
      await page.locator('#service-dialog').waitFor({state:'visible'});
      assert.equal(await page.evaluate(()=>state.selected.id),'client');
      await page.locator('#service-dialog .close-dialog').click();
      await page.locator('.device-node').nth(1).locator('.device-menu-button').click();
      if (width<1280) assert(await page.locator('#device-menu-add-project').isDisabled());
      assert.equal(await page.evaluate(()=>document.activeElement.id),'device-menu-settings','offline menu focuses an enabled action');
      await page.keyboard.press('Escape');
      if (width<=760 && await page.locator('#open-devices').getAttribute('aria-expanded')!=='true') await page.locator('#open-devices').click();
      await page.locator('.project-settings').first().click();
      assert(await page.locator('#device-menu-add-project').isHidden(),'project menu cannot expose client-level add');
      await page.keyboard.press('Escape');
      if (width<=760 && await page.locator('#open-devices').getAttribute('aria-expanded')!=='true') await page.locator('#open-devices').click();
      await page.evaluate(()=>{document.documentElement.dataset.theme='dark';document.documentElement.style.colorScheme='dark';});
      for (const [os,file] of [['macOS','macos-dark.svg'],['Linux','linux-dark.svg']]) {
        const logo=page.locator(`.device-icon[alt="${os}"]`);
        assert((await logo.evaluate(element=>getComputedStyle(element).content)).includes(file));
        assert.equal(await page.evaluate(async file=>(await fetch('icons/'+file)).status,file),200);
      }
      if (screenshots) await page.screenshot({path:path.join(screenshots,`sidebar-dark-${width}.png`),fullPage:true,animations:'disabled'});
      await page.evaluate(()=>{document.documentElement.dataset.theme='light';document.documentElement.style.colorScheme='light';closeDrawers();});
      const badge = page.locator('.workbench-status');
      assert.equal(await page.locator('.workbench-bar .workbench-status').count(),0);
      assert.match(await badge.textContent(),/执行中.*秒/);
      assert.equal(await badge.evaluate(element=>element.closest('.message').dataset.key),'current-task/assistant/stdout');
      assert.equal(await badge.evaluate(element=>element.parentElement===element.closest('.message').lastElementChild),true);
      assert.equal(await page.locator('.run-state-label').evaluate(element=>getComputedStyle(element).animationName),'run-color-marquee');
      if (width > 760) {
        const bounds = await page.locator('#devices-pane').boundingBox();
        assert(Math.abs(bounds.width-Math.max(180,(width-12)*.1575))<2,'compact first column retains readable minimum width');
        const second=await page.locator('#conversations-pane').boundingBox();
        assert(Math.abs(second.width-Math.max(180,(width-12)*.1725))<2,'compact second column retains readable minimum width');
      }
      if (screenshots) await page.screenshot({path:path.join(screenshots,`running-${width}.png`),fullPage:true});
      await page.evaluate(()=>{
        state.session.taskStatus='done';state.session.taskCompletedAt=Date.now();
        state.entries.push({seq:7,turnId:'current-task',role:'system',kind:'done',text:'',createdAt:Date.now()});renderChat();
      });
      assert.match(await badge.textContent(),/已完成/);
      assert.equal(await badge.evaluate(element=>getComputedStyle(element).color),'rgb(6, 118, 71)');
      const footer=page.locator('.message').filter({has:badge}).locator('.message-run-footer');
      assert.match(await footer.locator('.run-timing').textContent(),/完成于.*总耗时/);
      assert.equal(await page.locator('.message-label .run-timing').count(),0);
      if (width > 760) {
        const b=await badge.boundingBox(),time=await footer.locator('.run-timing').boundingBox();
        assert(time.x>=b.x+b.width,'completion timestamp follows status');
      }
      if (screenshots) await page.screenshot({path:path.join(screenshots,`done-${width}.png`),fullPage:true});
      await page.evaluate(()=>{document.documentElement.dataset.theme='dark';document.documentElement.style.colorScheme='dark';});
      assert.equal(await badge.evaluate(element=>getComputedStyle(element).color),'rgb(142, 226, 183)');
      if (screenshots) await page.screenshot({path:path.join(screenshots,`dark-${width}.png`),fullPage:true,animations:'disabled'});
      await page.evaluate(()=>{
        state.session.taskStatus='failed';state.entries=state.entries.filter(entry=>entry.seq!==7);
        state.entries.push({seq:8,turnId:'current-task',role:'system',kind:'error',text:'模拟验证失败，请检查日志。',createdAt:Date.now()});renderChat();
      });
      assert.match(await badge.textContent(),/执行失败/);
      assert.equal(await badge.evaluate(element=>getComputedStyle(element).color),'rgb(255, 165, 174)');
      await page.evaluate(()=>{document.documentElement.dataset.theme='light';document.documentElement.style.colorScheme='light';});
      assert.equal(await badge.evaluate(element=>getComputedStyle(element).color),'rgb(180, 35, 24)');
      if (screenshots) await page.screenshot({path:path.join(screenshots,`failed-${width}.png`),fullPage:true});
      await page.evaluate(()=>{state.session.taskStatus='cancelled';renderChat();});
      assert.match(await badge.textContent(),/已停止/,'explicit stop metadata takes precedence over an earlier error entry');
      await page.evaluate(()=>{
        state.session.taskId='next-task';state.session.taskStatus='queued';state.session.taskCompletedAt=0;state.session.taskAcceptedAt=Date.now();
        state.entries.push({seq:9,turnId:'next-task',role:'user',kind:'message',text:'继续下一项检查。',createdAt:Date.now()});renderChat();
      });
      assert.equal(await badge.evaluate(element=>element.closest('.message').dataset.key),'next-task/assistant/stdout');
      assert.match(await badge.textContent(),/排队中/);
      await page.evaluate(()=>{state.session.taskStatus='running';state.session.taskStartedAt=Date.now();renderChat();});
      await page.emulateMedia({reducedMotion:'reduce'});
      assert.equal(await page.locator('.run-state-label').evaluate(element=>getComputedStyle(element).animationName),'none');
      await page.emulateMedia({reducedMotion:'no-preference'});
      await page.evaluate(()=>{
        state.entries.push({seq:10,turnId:'next-task',role:'assistant',kind:'stdout',text:'下一轮已经开始。',createdAt:Date.now()});renderChat();
      });
      assert.equal(await page.locator('.run-status-placeholder').count(),0,'first output upgrades placeholder without breaking controls');
      await page.evaluate(()=>{state.peer=null;renderChat();});
      assert.match(await badge.textContent(),/上次状态/);
      assert.equal(await page.locator('.run-state-label').evaluate(element=>getComputedStyle(element).animationName),'none');
      await page.evaluate(()=>{window.stewardCalls=[];window.openAssistant=async ref=>window.stewardCalls.push(ref);});
      if (width <= 760) await page.locator('#open-devices').click();
      const row=page.locator('.device-node').first().locator('.device-row');
      assert.equal(await page.locator('.device-children > .client-steward').count(),0);
      const openBefore=await row.evaluate(element=>element.parentElement.open);
      await row.locator('.client-steward').click();
      assert.equal(await row.evaluate(element=>element.parentElement.open),openBefore,'steward click must not collapse client');
      await row.locator('.client-steward').focus();await page.keyboard.press('Enter');
      await page.locator('.project-steward').first().click();
      assert.deepEqual(await page.evaluate(()=>window.stewardCalls.map(ref=>ref.scope)),['client','client','project']);
      assert.equal(await page.locator('.device-node').nth(1).locator('.client-steward').isDisabled(),true);
      assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),true,'no page horizontal overflow');
      if (width > 760) {
        if (width < 1200) { await page.locator('#conversations-resizer').focus(); await page.keyboard.press('ArrowRight'); }
        const divider=page.locator('#devices-resizer'),before=await page.locator('#devices-pane').boundingBox();
        await divider.focus();await page.keyboard.press('ArrowRight');
        const resized=await page.locator('#devices-pane').boundingBox();assert(resized.width>before.width+10,'keyboard resize remains usable');
        await page.reload();await page.waitForFunction(()=>state.entries.some(entry=>entry.seq===6) && $('app').style.getPropertyValue('--devices-width').endsWith('px'));
        const restored=await page.locator('#devices-pane').boundingBox();assert(Math.abs(restored.width-resized.width)<2,'saved custom widths survive reload');
        await divider.dblclick();const reset=await page.locator('#devices-pane').boundingBox();assert(Math.abs(reset.width-before.width)<2,'reset uses compact defaults');
      }
      assert.deepEqual(errors,[]);
      await page.close();
    }
  } finally { await browser.close(); }
  console.log(baseline ? 'BASELINE: synthetic desktop/tablet/phone screenshots captured' : 'PASS: inline client steward, message-bottom statuses/timestamps, live shimmer, failure colors, new-turn placeholder, offline/reduced-motion, compact/resizable desktop and phone');
})().catch(error=>{console.error(error);process.exitCode=1;});
