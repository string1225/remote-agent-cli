const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {chromium} = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const root = path.resolve(process.env.NAVIGATION_ROOT || '.');

(async () => {
  const browser = await chromium.launch({headless:true, channel:process.env.PLAYWRIGHT_CHANNEL || 'msedge'});
  try {
    const page = await browser.newPage({viewport:{width:1360,height:850}}), errors = [];
    page.on('pageerror', error => errors.push(error.message));
    let releaseCatalog;
    const catalog = new Promise(resolve => { releaseCatalog = resolve; });
    await page.route('https://workspace.test/agents/**', async route => {
      const url = new URL(route.request().url()), relative = url.pathname.slice('/agents/'.length);
      if (relative === 'api/me') return route.fulfill({json:{id:'owner',username:'Fixture'}});
      if (relative === 'api/agents') { await catalog; return route.fulfill({json:await page.evaluate(() => fixture.agents)}); }
      let file = path.join(root,'web',relative || 'index.html');
      if (!fs.existsSync(file)) return route.fulfill({status:404,body:'missing fixture asset'});
      let body = fs.readFileSync(file);
      if (file.endsWith('index.html')) body = Buffer.from(body.toString().replace(/<script src="\.\/(?!vendor\/|markdown\.js|session-cache\.js|app\.js)[^"]+" defer><\/script>/g,''));
      const contentType = file.endsWith('.js') ? 'text/javascript' : file.endsWith('.css') ? 'text/css' : file.endsWith('.svg') ? 'image/svg+xml' : 'text/html';
      return route.fulfill({contentType,body});
    });
    await page.addInitScript(() => {
      const project = (id,name) => ({id,name,provider:'codex',workspace:'/'+name});
      const session = (id,serviceId) => ({id,serviceId,title:id,updatedAt:'2026-11-03T00:00:00Z'});
      const entries = name => Array.from({length:35},(_,i) => ({seq:i+1,turnId:name+i,role:i%2?'assistant':'user',kind:'stdout',text:name+' message '+i+'\n'+('Long fixture transcript. '.repeat(8))}));
      const a = {id:'a',name:'Computer A',os:'windows',online:true,services:[project('p1','Project One'),project('p2','Project Two')]};
      const b = {id:'b',name:'Computer B',os:'linux',online:true,services:[project('p3','Project Three')]};
      const meta = agent => ({services:agent.services,providers:{codex:{available:true}},permissionMode:'full',backgroundRuns:true,attachments:false});
      window.fixture = {agents:[a,b],sockets:[],requests:[],pending:[],delay:true,session,entries,meta};
      const prefix = 'remote-agent:workspace:v2:/agents/:';
      const write = (kind,scope,data) => localStorage.setItem(prefix+JSON.stringify([kind,...scope]),JSON.stringify({version:2,at:Date.now(),data}));
      write('agents',['owner'],[a,b]);
      write('selection',['owner'],{client:'a',project:'p1',session:'s1'});
      for (const [agent,pid,sid] of [[a,'p1','s1'],[a,'p2','s2'],[b,'p3','s3']]) {
        write('client',['owner',agent.id],meta(agent));
        write('sessions',['owner',agent.id,pid],[session(sid,pid)]);
        write('project-selection',['owner',agent.id,pid],{session:sid});
        write('history',['owner',agent.id,pid,sid],{session:session(sid,pid),entries:entries(sid)});
      }
      window.WebSocket = class {
        static OPEN = 1;
        constructor(url) { this.client=url.match(/agents\/([^/]+)\/connect/)[1];this.readyState=0;this.bufferedAmount=0;fixture.sockets.push(this); }
        emit(message) { this.onmessage?.({data:JSON.stringify(message)}); }
        open() { this.readyState=1;this.onopen?.();this.emit({type:'ready',transport:'relay-v1'}); }
        reply(request,data) { this.emit({type:'event',event:{id:request.id,type:'result',data}}); }
        send(raw) {
          const message=JSON.parse(raw);if(message.type==='ping'){this.emit({type:'pong'});return;}
          const request=message.request;fixture.requests.push({client:this.client,...request});
          if(request.type==='services.list'){this.reply(request,meta(fixture.agents.find(a=>a.id===this.client)));return;}
          const respond=()=>{
            if(request.type==='sessions.list')this.reply(request,{sessions:[session(({p1:'s1',p2:'s2',p3:'s3'})[request.serviceId],request.serviceId)],nextCursor:0});
            else if(request.type==='sessions.get')this.reply(request,{session:session(request.sessionId,({s1:'p1',s2:'p2',s3:'p3'})[request.sessionId]),entries:request.cursor?[]:entries('fresh '+request.sessionId),nextCursor:0});
          };
          if(fixture.delay)fixture.pending.push({client:this.client,request,respond});else respond();
        }
        close() { this.readyState=3;this.onclose?.(); }
      };
    });
    await page.goto('https://workspace.test/agents/',{waitUntil:'domcontentloaded'});
    await page.waitForFunction(() => document.querySelector('#messages').textContent.includes('s1 message 34'));
    await page.evaluate(()=>{
      window.emptyRenders=[];
      new MutationObserver(()=>{if(!document.querySelector('#messages').textContent.trim())emptyRenders.push('empty');})
        .observe(document.querySelector('#messages'),{subtree:true,childList:true,characterData:true});
    });
    const atBottom = () => page.locator('#chat-scroll').evaluate(el => el.scrollHeight-el.scrollTop-el.clientHeight<3);
    assert(await atBottom(),'cached session must start at latest message before relay/catalog respond');
    assert.equal(await page.locator('#send-run').isDisabled(),true,'sending must wait for authentication');
    await page.locator('.project-button').filter({hasText:'Project Two'}).click();
    await page.waitForFunction(() => document.querySelector('#messages').textContent.includes('s2 message 34'));
    assert(await atBottom(),'cached project must show its last session at the bottom');
    assert.equal(await page.locator('.conversation-card.active strong').textContent(),'s2');
    assert.equal(await page.evaluate(()=>fixture.requests.length),0,'navigation must not wait for relay readiness');
    await page.locator('.device-node').filter({hasText:'Computer B'}).locator('summary').click();
    await page.locator('.project-button').filter({hasText:'Project Three'}).click();
    await page.waitForFunction(() => document.querySelector('#messages').textContent.includes('s3 message 34'));
    await page.evaluate(()=>fixture.sockets.at(-1).open());
    await page.waitForFunction(()=>fixture.pending.some(p=>p.request.type==='sessions.get'));
    assert((await page.locator('#messages').textContent()).includes('s3 message 34'),'cached messages vanished during background refresh');
    // A user reading older content while the full history refreshes keeps their position.
    await page.locator('#chat-scroll').evaluate(el=>{el.scrollTop=0;el.dispatchEvent(new Event('scroll'));});
    await page.evaluate(()=>fixture.pending.splice(0).forEach(p=>p.respond()));
    await page.waitForFunction(()=>document.querySelector('#messages').textContent.includes('fresh s3 message 34'));
    assert.equal(await page.locator('#chat-scroll').evaluate(el=>el.scrollTop),0,'background history pulled a reader away from older messages');
    await page.locator('.conversation-card.active').click();
    assert(await atBottom(),'explicit session selection must return to the latest message');
    await page.evaluate(()=>fixture.pending.splice(0).forEach(p=>p.respond()));
    await page.locator('.project-button').filter({hasText:'Project One'}).click();
    await page.waitForFunction(()=>document.querySelector('#messages').textContent.includes('s1 message 34'));
    await page.evaluate(()=>fixture.sockets.at(-1).open());
    await page.waitForFunction(()=>fixture.pending.some(p=>p.request.sessionId==='s1'));
    await page.locator('.project-button').filter({hasText:'Project Two'}).click();
    await page.waitForFunction(()=>document.querySelector('#messages').textContent.includes('s2 message 34'));
    await page.evaluate(()=>fixture.pending.filter(p=>p.request.serviceId==='p1'||p.request.sessionId==='s1').forEach(p=>p.respond()));
    assert((await page.locator('#messages').textContent()).includes('s2 message 34'),'late project response replaced the current session');
    await page.evaluate(()=>fixture.pending.splice(0).filter(p=>p.request.serviceId!=='p1'&&p.request.sessionId!=='s1').forEach(p=>p.respond()));
    await page.waitForFunction(()=>document.querySelector('#messages').textContent.includes('fresh s2 message 34'));
    releaseCatalog();
    // Closing a relay retains its selected project, messages and editable draft.
    await page.locator('#prompt').fill('unsent fixture draft');
    await page.evaluate(()=>fixture.sockets.at(-1).close());
    await page.waitForFunction(()=>document.querySelector('#connection-status').textContent.includes('已断开'));
    assert((await page.locator('#messages').textContent()).includes('fresh s2 message 34'));
    assert.equal(await page.locator('#prompt').inputValue(),'unsent fixture draft');
    assert.equal(await page.locator('#send-run').isDisabled(),true);
    // Full loaded history remains available in this page, beyond the disk cache limit.
    assert.deepEqual(await page.evaluate(()=>{
      const rows=Array.from({length:400},(_,i)=>({seq:i+1,text:'entry '+i})),scope=['owner','a','p1','long'];
      WorkspaceCache.history(scope,{id:'long',serviceId:'p1'},rows);
      const live=WorkspaceCache.read('history',scope).entries;
      const disk=JSON.parse(localStorage.getItem('remote-agent:workspace:v2:/agents/:'+JSON.stringify(['history',...scope]))).data.entries;
      return [live.length,live[0].seq,disk.length];
    }),[400,1,250]);
    await page.setViewportSize({width:390,height:844});
    await page.locator('#open-conversations').click();
    await page.locator('.conversation-card.active').click();
    assert(await atBottom(),'mobile/offline session must start at latest message');
    assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);
    assert.deepEqual(await page.evaluate(()=>emptyRenders),[],'switching among loaded views must never render an empty transcript');
    assert.deepEqual(errors,[]);
    console.log('PASS: cached startup before catalog/relay, instant project/client switches, background reconciliation, late response isolation, latest-message scroll, reader position, disconnect drafts, full memory history, mobile/offline');
  } finally { await browser.close(); }
})().catch(error=>{console.error(error);process.exitCode=1;});
