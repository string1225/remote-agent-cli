const assert=require('node:assert/strict');
const fs=require('node:fs');
const path=require('node:path');
const {chromium}=require(process.env.PLAYWRIGHT_MODULE||'playwright');
const root=path.resolve(process.env.CHANNEL_UI_ROOT||'.');
(async()=>{
 const browser=await chromium.launch({headless:true,channel:process.env.PLAYWRIGHT_CHANNEL||'msedge'});
 try{
  const page=await browser.newPage({viewport:{width:1200,height:900}}),errors=[];
  page.on('pageerror',e=>errors.push(e.message));
  await page.route('https://channel.test/**',route=>{
   const name=new URL(route.request().url()).pathname.slice(1)||'index.html';
   const file=path.join(root,'web',name);if(!fs.existsSync(file))return route.fulfill({status:404,body:''});
   let body=fs.readFileSync(file);if(name==='index.html')body=Buffer.from(body.toString().replace(/<script[^>]*><\/script>/g,''));
   return route.fulfill({body,contentType:name.endsWith('.js')?'text/javascript':name.endsWith('.css')?'text/css':'text/html'});
  });
  await page.goto('https://channel.test/');
  await page.evaluate(()=>{
   window.$=id=>document.getElementById(id);
   window.state={permissionMode:'full',peer:{},services:[{id:'p',name:'remote-agent-cli',channel:{name:'研发助理'}}]};
   window.fixture={config:{configured:true,provider:'feishu',name:'研发助理',appId:'cli_example_app',state:'connected',secretConfigured:true},calls:[]};
   window.service=()=>state.services[0];window.connectDevice=async()=>{};window.selectProject=async()=>{};window.closeDrawers=()=>{};
   window.refreshServices=()=>{};window.refreshSessions=async()=>{};window.notice=text=>{fixture.notice=text};
   window.rpc=async(type,body)=>{fixture.calls.push({type,body});return type==='channels.get'?fixture.config:{}};
   $('auth').hidden=true;$('app').hidden=false;$('account-name').textContent='设计预览';
   $('device-list').innerHTML='<details class="device-node selected" open><summary class="device-row"><span class="device-name">工作电脑</span><span class="device-status">在线</span></summary><div class="device-children"><button class="project-button active">remote-agent-cli</button></div></details>';
   $('device-empty').hidden=true;$('device-count').textContent='1';$('project-label').textContent='Codex / remote-agent-cli';
   $('conversation-list').innerHTML='<button class="conversation-card pinned"><strong>SaySo</strong><p>讨论、录音与 Action</p></button><button class="conversation-card im-session"><strong>研发助理 · 小林</strong><p>飞书私聊</p></button><button class="conversation-card active"><strong>优化渠道配置页面</strong><p>Web 会话</p></button>';
   $('conversation-empty').hidden=true;$('chat-title').textContent='优化渠道配置页面';
   document.querySelectorAll('.close-dialog').forEach(button=>button.onclick=()=>button.closest('dialog').close());
  });
  await page.addScriptTag({path:path.join(root,'web/channels.js')});
  const open=()=>page.evaluate(()=>openChannelSetup({id:'client'},state.services[0]));
  await open();
  assert.equal(await page.locator('#channel-form [name=name]').inputValue(),'研发助理');
  assert.equal(await page.locator('#channel-form [name=appSecret]').inputValue(),'');
  assert.equal(await page.locator('#channel-form [name=appSecret]').getAttribute('required'),null);
  assert.equal(await page.locator('#channel-robot-code').isHidden(),true);
  assert.equal(await page.locator('#channel-dialog').getByText('所有人都可以聊').count(),0);
  assert.equal(await page.locator('#channel-dialog').getByText('消息接收事件',{exact:false}).count(),0);
  const checkButtons=async()=>{
   const rects=await page.locator('.channel-actions button').evaluateAll(buttons=>buttons.map(b=>{const r=b.getBoundingClientRect();return{x:r.x,y:r.y,width:r.width,height:r.height}}));
   assert.equal(rects.length,2);assert(Math.abs(rects[0].width-rects[1].width)<1);assert.equal(rects[0].y,rects[1].y);assert(rects[0].height>=44);
   assert(await page.locator('#channel-dialog').evaluate(el=>el.scrollWidth<=el.clientWidth));
  };
  await checkButtons();
  const out=process.env.CHANNEL_PREVIEW_DIR||'.local/channel-redesign/previews';fs.mkdirSync(out,{recursive:true});
  await page.screenshot({path:path.join(out,'channel-desktop.png')});
  await page.locator('#channel-dialog').screenshot({path:path.join(out,'channel-settings.png')});
  await page.setViewportSize({width:390,height:844});await checkButtons();
  await page.screenshot({path:path.join(out,'channel-mobile.png')});
  await page.locator('#channel-form [name=provider]').selectOption('dingtalk');
  assert.equal(await page.locator('#channel-robot-code').isVisible(),true);
  await page.locator('#channel-dialog .close-dialog').click();
  await page.evaluate(()=>{fixture.config.name='';fixture.config.conversations=[{name:'研发助理'}]});
  await open();assert.equal(await page.locator('#channel-form [name=name]').inputValue(),'研发助理','existing public channel metadata was not restored');
  await page.locator('#channel-form [name=name]').fill('研发值班');
  await page.locator('#channel-form [name=appSecret]').fill('fixture-secret');
  assert.deepEqual(await page.locator('#channel-form').evaluate(form=>Array.from(form.elements).filter(e=>e.willValidate&&!e.validity.valid).map(e=>({name:e.name,message:e.validationMessage,value:e.value}))),[]);
  await page.locator('#channel-form button[type=submit]').click();
  assert.equal(await page.evaluate(()=>fixture.notice||''),'');
  assert.deepEqual(errors,[]);
  await page.waitForFunction(()=>fixture.calls.some(c=>c.type==='channels.configure'));
  assert.equal(await page.evaluate(()=>fixture.calls.find(c=>c.type==='channels.configure').body.body.name),'研发值班');
  assert.equal(await page.evaluate(()=>Object.values(localStorage).some(v=>v.includes('fixture-secret'))),false);
  assert.deepEqual(errors,[]);
  console.log('PASS: saved channel names, credential isolation, clean form, equal side-by-side actions at desktop/mobile, provider fields, configure payload and rendered previews');
 }finally{await browser.close()}
})().catch(error=>{console.error(error);process.exitCode=1});
