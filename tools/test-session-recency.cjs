const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {chromium} = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const web = path.resolve(process.env.UI_WEB_ROOT || 'web');

(async () => {
  const browser = await chromium.launch({headless:true, channel:process.env.PLAYWRIGHT_CHANNEL || 'msedge'});
  try {
    for (const width of [1440,390]) {
      const page = await browser.newPage({viewport:{width,height:960}}), errors=[];
      page.on('pageerror',error=>errors.push(error.message));
      await page.clock.install({time:new Date('2026-11-15T12:00:00Z')});
      await page.addInitScript(()=>{window.setInterval=()=>0;});
      await page.route('https://recency.test/**',route=>{
        const name=new URL(route.request().url()).pathname.slice(1)||'index.html';
        if(name==='api/me')return route.fulfill({json:{id:'owner',username:'Recency fixture'}});
        if(name==='api/agents')return route.fulfill({json:[]});
        if(name==='api/default-host')return route.fulfill({json:{agentId:'',online:false}});
        if(name==='api/assistants')return route.fulfill({json:{assistants:[]}});
        if(name.startsWith('api/'))return route.fulfill({json:{}});
        const file=path.join(web,name);
        return route.fulfill({contentType:name.endsWith('.js')?'text/javascript':name.endsWith('.css')?'text/css':name.endsWith('.svg')?'image/svg+xml':'text/html',body:fs.readFileSync(file)});
      });
      await page.goto('https://recency.test/');
      await page.locator('#app').waitFor({state:'visible'});
      await page.evaluate(()=>{
        const now=Date.now(), day=86400000;
        const row=(id,age,extra={})=>({id,serviceId:'project',title:id,createdAt:now-20*day,updatedAt:now-age,lastMessageAt:now-age,taskStatus:'done',...extra});
        window.fixture={calls:[],sessions:[
          row('recent',day), row('old-web',8*day), row('old-im',9*day,{kind:'im',channelName:'旧群聊'}),
          row('exact-seven-days',7*day), row('renamed-today',10*day,{updatedAt:now}),
          row('old-running',10*day,{taskStatus:'running'}),row('old-queued',10*day,{taskStatus:'queued'}),
          row('pending-edit',10*day,{queuedMessages:[{id:'q',status:'editing'}]}),
          row('empty-old',10*day,{lastMessageAt:0,createdAt:now-8*day}),
          row('unknown-time',day,{lastMessageAt:'invalid'}),row('future-time',-day),
          row('legacy-old',8*day,{lastMessageAt:undefined}),
          row('expires-soon',7*day-60000),row('manual-archive',10*day,{archivedAt:1}),
          row('deleted',day,{deletedAt:1})
        ]};
        rpc=async(type,fields={})=>{
          fixture.calls.push([type,structuredClone(fields)]);
          if(type==='sessions.list'){
            if(fields.archived)return {sessions:[fixture.sessions.find(s=>s.id==='manual-archive')],nextCursor:0};
            if(fields.cursor===50)return {sessions:[row('page-two-old',9*day),row('page-two-recent',day)],nextCursor:0};
            return {sessions:fixture.sessions.filter(s=>!s.archivedAt&&!s.deletedAt&&s.title.includes(fields.search||'')),nextCursor:0};
          }
          if(type==='sessions.get')return {session:fixture.sessions.find(s=>s.id===fields.sessionId),entries:[{seq:1,role:'assistant',kind:'stdout',text:'Synthetic retained history',createdAt:now-8*day}],nextCursor:0};
          throw Error('Unexpected request '+type);
        };
        state.user={id:'owner'};state.selected={id:'client',online:true};state.peer={ws:{readyState:WebSocket.OPEN}};
        state.services=[{id:'project',name:'Recency fixture',provider:'codex'}];state.serviceId='project';state.mode='web';
        state.sessionManagement=true;state.sessions=structuredClone(fixture.sessions);state.session=state.sessions[0];renderSessions();
        $('project-label').textContent='CODEX / Recency fixture';$('notice').hidden=true;
      });
      if(width<760)await page.locator('#open-conversations').click();
      const old=page.locator('.older-sessions'),summary=old.locator('summary');
      const direct=id=>page.locator('#conversation-list>.conversation-card').filter({hasText:id});
      const inside=id=>old.locator('.conversation-card').filter({hasText:id});
      assert.match(await summary.textContent(),/7天前的会话 · 6/);
      assert(await inside('old-web').isHidden());
      assert.equal(await direct('renamed-today').count(),0,'renaming must not revive old messages');
      for(const id of ['recent','old-running','old-queued','pending-edit','unknown-time','future-time','expires-soon'])assert.equal(await direct(id).count(),1,id);
      assert.equal(await page.locator('#conversation-count').textContent(),'8','count excludes older, archived and deleted rows, includes SaySo');
      assert.equal(await page.evaluate(()=>fixture.calls.length),0,'folding must not issue history or mutation requests');
      await summary.focus();await summary.press('Enter');
      assert(await inside('old-web').isVisible());
      await page.evaluate(()=>renderSessions());
      assert(await inside('old-web').isVisible(),'poll rerenders preserve disclosure state');
      await page.evaluate(()=>{document.documentElement.dataset.theme='dark';});
      if(process.env.UI_TEST_SCREENSHOT_DIR){fs.mkdirSync(process.env.UI_TEST_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:path.join(process.env.UI_TEST_SCREENSHOT_DIR,`older-open-${width}.png`),animations:'disabled'});}
      await summary.click();
      await page.evaluate(()=>loadSession(fixture.sessions.find(s=>s.id==='old-web')));
      assert.equal(await page.evaluate(()=>state.session.id),'old-web','direct access to an old session remains valid');
      assert.match(await page.locator('#messages').textContent(),/Synthetic retained history/);
      assert(await inside('old-web').isHidden(),'reading history does not change message age');
      await page.evaluate(()=>{const s=fixture.sessions.find(s=>s.id==='old-web');s.lastMessageAt=s.updatedAt=Date.now();state.sessions=state.sessions.map(row=>row.id===s.id?{...s}:row);renderSessions();});
      assert.equal(await direct('old-web').count(),1,'new message returns the same session to the main list');
      await page.evaluate(()=>{$('session-search').value='renamed';});
      await page.evaluate(()=>refreshSessions());
      assert.equal(await direct('renamed-today').count(),1,'search includes normally hidden sessions');
      assert.equal(await old.count(),0);
      await page.evaluate(()=>{$('session-search').value='';});await page.evaluate(()=>refreshSessions());
      assert(await inside('renamed-today').isHidden(),'clearing search restores the default collapsed list');
      await page.clock.runFor(60000);
      assert.equal(await direct('expires-soon').count(),0,'crossing seven days hides without a new message or network refresh');
      assert.equal(await inside('expires-soon').count(),1);
      await page.evaluate(()=>{state.listCursor=50;renderSessions();});
      await page.locator('#more-sessions').click();
      assert.equal(await direct('page-two-recent').count(),1,'pagination is retained');
      assert.equal(await inside('page-two-old').count(),1);
      await summary.click();
      await page.evaluate(()=>{state.selected={id:'other-client',online:false};state.peer=null;renderSessions();});
      assert(await inside('renamed-today').isHidden(),'client switch resets old-session disclosure');
      await summary.click();
      await page.evaluate(()=>{state.user={id:'other-account'};renderSessions();});
      assert(await inside('renamed-today').isHidden(),'account switch resets disclosure');
      await page.evaluate(()=>{state.sessions=state.sessions.filter(s=>s.id==='renamed-today');renderSessions();});
      assert.match(await page.locator('#conversation-empty').textContent(),/旧会话可在下方展开/);
      assert(await page.locator('#conversation-empty').isVisible());
      assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);
      assert.deepEqual(errors,[]);
      await page.close();
    }
    console.log('PASS: seven-day message cutoff, rename isolation, legacy/empty timestamps, active queues, searchable collapsed history, same-session revival, local expiry, pagination, account/client isolation, desktop/mobile');
  }finally{await browser.close();}
})().catch(error=>{console.error(error);process.exitCode=1;});
