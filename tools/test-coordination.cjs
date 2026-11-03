const assert=require('node:assert/strict'),fs=require('node:fs'),path=require('node:path');
const {chromium}=require(process.env.PLAYWRIGHT_MODULE||'playwright');
(async()=>{
 const browser=await chromium.launch({headless:true,channel:process.env.PLAYWRIGHT_CHANNEL||'msedge'});
 try{for(const width of [1440,390]){
  const page=await browser.newPage({viewport:{width,height:900}}),errors=[],replies=[],dispatches=[];
  page.on('pageerror',e=>errors.push(e.message));
  await page.addInitScript(()=>{window.setInterval=()=>0});
  const baseProject={id:'project',name:'研发项目',provider:'codex',workspace:'/projects/test'};
  const ceo={id:'ceo',name:'CEO',provider:'codex',workspace:'/ceo',scope:'account'};
  const assistants=[{scope:'account',clientId:'client',name:'CEO',online:true},{scope:'client',clientId:'client',name:'电脑 · 管家',online:true},{scope:'project',clientId:'client',projectId:'project',name:'研发项目 · 管家',online:true}];
  const report={id:'report',workId:'work',replyHandle:'handle',status:'ready',summary:'已验证报告 <img src=x onerror=alert(1)>',origin:{scope:'project',clientId:'client',projectId:'project-steward',sessionId:'leaf-session',name:'研发项目管家'},related:[{workId:'deep-work',replyHandle:'leaf-handle',origin:{clientId:'client',projectId:'project-steward',sessionId:'deep-session',name:'叶子管家'}}]};
  const services=[baseProject,ceo,{...ceo,id:'client-steward',scope:'client',name:'客户端管家'},{...baseProject,id:'project-steward',scope:'project',name:'项目管家'}];
  const sessions=new Map(services.map(s=>[s.id,{id:'session-'+s.id,serviceId:s.id,title:s.name+' 对话'}]));
  let actionStatus='ready';
  await page.route('https://hierarchy.test/**',async route=>{
   const file=new URL(route.request().url()).pathname.slice(1)||'index.html';
   if(file==='api/me')return route.fulfill({json:{id:'user',username:'Test'}});
   if(file==='api/agents')return route.fulfill({json:[{id:'client',name:'电脑',os:'windows',online:true,enrolled:true,services} ]});
   if(file==='api/default-host')return route.fulfill({json:{agentId:'client',name:'电脑',online:true}});
   if(file==='api/todos')return route.fulfill({json:{todos:[]}});
   if(file==='api/assistants')return route.fulfill({json:{assistants}});
   if(file==='api/coordination'){
    const c=route.request().postDataJSON();
    if(c.operation==='resolve'){const id=c.target.scope==='account'?'ceo':c.target.scope==='client'?'client-steward':'project-steward';return route.fulfill({json:{...c.target,clientId:'client',projectId:id,service:services.find(s=>s.id===id)}})}
    if(c.operation==='reply'){replies.push(c);return route.fulfill({json:{target:report.origin,taskId:'reply-task'}})}
   }
   const target=path.join('web',file);if(!fs.existsSync(target))return route.fulfill({status:404,body:''});
   return route.fulfill({contentType:file.endsWith('.js')?'text/javascript':file.endsWith('.css')?'text/css':file.endsWith('.svg')?'image/svg+xml':'text/html',body:fs.readFileSync(target)});
  });
  await page.routeWebSocket('**/api/agents/*/connect',ws=>{
   ws.onMessage(message=>{
    const m=JSON.parse(message);if(m.type==='ping'){ws.send(JSON.stringify({type:'pong'}));return}
    const r=m.request;if(!r)return;let data={};
    if(r.type==='services.list')data={services,providers:{codex:{available:true}},permissionMode:'full',sayso:true,taskQueue:true,backgroundRuns:true,clientVersion:'test'};
    if(r.type==='sessions.list')data={sessions:[sessions.get(r.serviceId)],nextCursor:0};
    if(r.type==='sessions.get')data={session:[...sessions.values()].find(s=>s.id===r.sessionId)||{id:r.sessionId,serviceId:'project-steward',title:'leaf'},entries:[{seq:1,kind:'report',role:'assistant',text:report.summary,report}],nextCursor:0};
    if(r.type==='sayso.open')data={id:'so-project',pinnedSessionId:'discussion'};
    if(r.type==='sayso.call'){
     const b=r;
     if(b.path==='/api/state')data={projects:[{id:'so-project',name:'CEO',rootPath:'/ceo'}],sessions:[{id:'discussion',projectId:'so-project',title:'产品讨论',status:'paused',startedAt:new Date().toISOString(),recordings:[],transcript:[],analysisTraces:[]}],actions:[{id:'action',projectId:'so-project',sessionId:'discussion',title:'发布报告',detail:'验证结果',status:actionStatus,priority:'medium'}],runs:[],settings:{analysisProvider:'local',analysisModel:'default',executor:'codex',executionModel:'default'}};
     if(b.path==='/api/actions/action/dispatch'){dispatches.push(b.body);actionStatus='running';data={target:b.body.target}}
    }
    ws.send(JSON.stringify({type:'event',event:{id:r.id,type:'result',data}}));
   });
   ws.send(JSON.stringify({type:'ready',transport:'relay-v1'}));
  });
  await page.goto('https://hierarchy.test/');await page.locator('#app').waitFor({state:'visible'});
  if(width===1440)await page.locator('#ceo-open').click();
  await page.locator('.work-report').waitFor();assert.match(await page.locator('#project-label').textContent(),/CEO/);
  assert.equal(await page.locator('.device-tree .project-button').count(),1,'system projects stay out of ordinary tree');
  assert.equal(await page.locator('.work-report img').count(),0,'report sanitization');
  await page.getByRole('button',{name:'直接回复 叶子管家',exact:true}).click();
  const dialog=page.locator('.work-reply-dialog');await dialog.getByRole('textbox').fill('继续原任务');await dialog.getByRole('button',{name:'发送到下游'}).click();await dialog.waitFor({state:'detached'});
  assert.equal(replies[0].replyHandle,'leaf-handle');assert.equal(replies[0].prompt,'继续原任务');assert(replies[0].requestId);
  if(width===390)await page.locator('#open-devices').click();
  await page.locator('#ceo-sayso').click();
  if(width===390)await page.getByRole('button',{name:/^Action ·/}).click();
  await page.getByRole('button',{name:'派给管家',exact:true}).click();
  const picker=page.getByRole('dialog',{name:'选择执行管家'});await picker.getByRole('combobox').selectOption('1');await picker.getByRole('button',{name:'确认派工',exact:true}).click();await picker.waitFor({state:'detached'});
  assert.equal(dispatches[0].target.scope,'project');assert.equal(dispatches[0].target.projectId,'project');
  assert.deepEqual(errors,[]);
  if(process.env.UI_TEST_SCREENSHOT_DIR){fs.mkdirSync(process.env.UI_TEST_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:path.join(process.env.UI_TEST_SCREENSHOT_DIR,`hierarchy-${width}.png`)})}
  await page.close();
 }}finally{await browser.close()}
 console.log('PASS: desktop/phone CEO, hidden system projects, sanitized result cards, direct leaf reply and SaySo target dispatch');
})().catch(e=>{console.error(e);process.exitCode=1});
