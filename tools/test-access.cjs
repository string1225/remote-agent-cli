const assert = require('node:assert/strict');
const path = require('node:path');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
(async () => {
  const browser = await chromium.launch({headless:true,channel:process.env.PLAYWRIGHT_CHANNEL || 'msedge'});
  try {
    const page = await browser.newPage({viewport:{width:1100,height:850}}), errors=[];
    page.on('pageerror',e=>errors.push(e.message));
    await page.route('https://access.test/**',r=>r.fulfill({contentType:'text/html',body:'<!doctype html><html><head></head><body><nav class="platform-nav"></nav><div id="device-menu" hidden><button id="device-menu-delete">删除</button></div></body></html>'}));
    await page.goto('https://access.test/');
    await page.evaluate(()=>{
      window.$=id=>document.getElementById(id);
      window.node=(tag,text,cls)=>{const n=document.createElement(tag);if(text!==undefined)n.textContent=text;if(cls)n.className=cls;return n};
      window.state={user:{id:'owner'},agents:[{id:'current',name:'测试电脑',online:true},{id:'old',name:'旧客户端',online:true}],selected:{id:'current'}};
      window.basePath='';window.openDeviceMenu=()=>{};window.openProjectMenu=()=>{};window.closeDrawers=()=>{};
      window.fixture={requests:[],servers:[],accounts:[],status:{},pendingAdd:null};
      window.WebSocket=class {
        static OPEN=1;
        constructor(url){this.client=url.includes('/old/')?'old':'current';this.readyState=1;setTimeout(()=>this.onmessage?.({data:JSON.stringify({type:'ready',transport:'relay-v1'})}),2)}
        reply(request,data){setTimeout(()=>this.onmessage?.({data:JSON.stringify({type:'event',event:{id:request.id,type:'result',data}})}),2)}
        send(raw){const request=JSON.parse(raw).request;if(!request)return;fixture.requests.push(request);if(request.type==='authenticate'){this.reply(request,{credentialManagement:this.client!=='old'});return}
          const body=request.body;let result;
          switch(body.action){
            case 'servers.list':result=fixture.servers;break;
            case 'servers.add':fixture.pendingAdd=()=>{const server={name:body.name,host:body.host,user:body.user,port:body.port,fingerprint:'SHA256:fixture',configPath:'C:/private/fixture/config'};fixture.servers=[server];this.reply(request,server)};return;
            case 'servers.test':result=fixture.servers[0];break;
            case 'github.accounts':result=fixture.accounts;break;
            case 'github.start':fixture.status={status:'waiting',code:'ABCD-1234',url:'https://github.com/login/device',expiresAt:Date.now()+900000,message:'请在 GitHub 页面确认授权'};result=fixture.status;break;
            case 'github.status':result=fixture.status;break;
            case 'github.cancel':fixture.status={status:'cancelled',message:'授权已取消'};result=true;break;
          }
          this.reply(request,result);
        }
        close(){this.readyState=3;this.onclose?.()}
      };
    });
    await page.addStyleTag({path:path.resolve('web/style.css')});
    await page.addStyleTag({path:path.resolve('web/access.css')});
    await page.addScriptTag({path:path.resolve('web/access.js')});
    await page.getByRole('button',{name:'⌘ 服务器与 GitHub'}).click();
    await page.waitForFunction(()=>!$('access-content').hidden);
    const form=page.locator('#access-server-form');
    await form.locator('[name=name]').fill('fixture');await form.locator('[name=host]').fill('127.0.0.1');await form.locator('[name=user]').fill('fixture');await form.locator('[name=password]').fill('mock-password');
    await form.getByRole('button').click();
    assert.equal(await form.locator('[name=password]').inputValue(),'','password not cleared immediately');
    assert.equal(await page.evaluate(()=>Object.values(localStorage).some(v=>v.includes('mock-password'))),false);
    await page.evaluate(()=>fixture.pendingAdd());
    await page.waitForFunction(()=>$('access-servers').textContent.includes('SHA256:fixture'));
    await page.locator('#access-github-form button').click();
    await page.waitForFunction(()=>$('access-github-state').textContent.includes('ABCD-1234'));
    await page.locator('#access-dialog .icon-button').click();
    await page.getByRole('button',{name:'⌘ 服务器与 GitHub'}).click();
    await page.waitForFunction(()=>$('access-github-state').textContent.includes('ABCD-1234'));
    await page.locator('#access-github-cancel').click();
    await page.waitForFunction(()=>$('access-github-state').textContent.includes('授权已取消'));
    await page.locator('#access-client').selectOption('old');
    await page.waitForFunction(()=>$('access-message').textContent.includes('请先升级'));
    assert.equal(await page.locator('#access-content').isHidden(),true);
    await page.locator('#access-client').selectOption('current');await page.waitForFunction(()=>!$('access-content').hidden);
    await page.setViewportSize({width:390,height:844});
    await page.screenshot({path:'.local/qa/access-mobile.png'});
    assert.equal(await page.evaluate(()=>$('access-dialog').scrollWidth<=$('access-dialog').clientWidth+1),true,'mobile dialog overflow');
    await form.locator('[name=password]').fill('discard-on-signout');
    await page.evaluate(()=>platformSignedOut());
    assert.equal(await form.locator('[name=password]').inputValue(),'');
    assert.equal(await page.locator('#access-dialog').isVisible(),false);
    assert.equal(await page.evaluate(()=>fixture.requests.every(r=>['authenticate','access.call'].includes(r.type))),true,'credentials were sent as model tasks');
    assert.deepEqual(errors,[]);
    console.log('PASS: isolated credential forms, immediate password clearing, persistent authorization, cancellation, old client gating, mobile layout and sign-out');
  } finally {await browser.close()}
})().catch(error=>{console.error(error);process.exit(1)});
