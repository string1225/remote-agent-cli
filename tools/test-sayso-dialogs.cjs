const assert=require('node:assert/strict'), fs=require('node:fs'), path=require('node:path');
const {chromium}=require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const web=path.resolve(process.env.UI_WEB_ROOT || 'web');
(async()=>{
 const browser=await chromium.launch({headless:true,channel:process.env.PLAYWRIGHT_CHANNEL || 'msedge'});
 try {
  for(const width of [1440,390]) for(const theme of ['light','dark']) {
   const page=await browser.newPage({viewport:{width,height:800}}), errors=[];
   page.on('pageerror',e=>errors.push(e.message));
   await page.route('https://sayso-dialog.test/**',route=>{
    const file=new URL(route.request().url()).pathname.slice(1);
    if(!file)return route.fulfill({contentType:'text/html',body:'<html><body style="margin:0"><div id="host" style="height:100dvh"></div></body></html>'});
    return route.fulfill({contentType:file.endsWith('.js')?'text/javascript':'text/css',body:fs.readFileSync(path.join(web,file))});
   });
   await page.goto('https://sayso-dialog.test/');
   await page.evaluate(async theme=>{
    window.setInterval=()=>0; document.documentElement.dataset.theme=theme;
    const {mount}=await import('/sayso/sayso.js');
    const data={projects:[{id:'p',name:'模拟项目',rootPath:'/fixture',color:'green',createdAt:'2026-11-15T00:00:00Z'}],sessions:[{id:'s',projectId:'p',title:'模拟讨论',status:'paused',startedAt:'2026-11-15T00:00:00Z',recordings:[],transcript:[]}],actions:[{id:'a',sessionId:'s',projectId:'p',title:'模拟 Action',detail:'说明',source:'讨论',priority:'medium',status:'ready',createdAt:'2026-11-15T00:00:00Z',updatedAt:'2026-11-15T00:00:00Z'}],runs:[],settings:{analysisProvider:'local',analysisModel:'gpt-5',analysisIntervalMinutes:2,executor:'codex',executionModel:'gpt-5'}};
    window.mountFixture=ceo=>mount(document.getElementById('host'),async (_,url)=>url==='/api/state'?data:url==='/api/assistants'?{assistants:[{scope:'project',clientId:'c',projectId:'p',name:'模拟管家',online:true}]}:{},()=>{},'模拟客户端',ceo?{projectId:'p',sessionId:'s',ceo:true}:undefined);
    window.dispose=window.mountFixture(false);
   },theme);
   await page.locator('.system-settings').waitFor({state:'attached'});
   await page.evaluate(()=>document.getElementById('host').shadowRoot.querySelector('.system-settings').click());
   for(const mode of ['settings','dispatch']) {
    if(mode==='dispatch') {
     await page.evaluate(()=>{window.dispose();window.dispose=window.mountFixture(true);});
     await page.locator('.dispatch-button').waitFor({state:'attached'});
     await page.evaluate(()=>document.getElementById('host').shadowRoot.querySelector('.dispatch-button').click());
    }
    await page.locator('.modal-body').waitFor({state:'visible'});
    await page.evaluate(()=>{const root=document.getElementById('host').shadowRoot; const filler=document.createElement('div');filler.style.height='1800px';root.querySelector('.modal-body').append(filler);});
    await page.waitForTimeout(250);
    const result=await page.evaluate(async()=>{
     const root=document.getElementById('host').shadowRoot,head=root.querySelector('.modal-heading,.modal-header'),body=root.querySelector('.modal-body');
     const before=head.getBoundingClientRect();body.scrollTop=900;await new Promise(requestAnimationFrame);const after=head.getBoundingClientRect(),box=root.querySelector('.modal').getBoundingClientRect();
     return {delta:Math.abs(before.top-after.top),scroll:body.scrollTop,bottom:box.bottom,height:innerHeight,top:box.top,head:after.top};
    });
    assert(result.delta<1 && result.scroll>0 && result.bottom<=result.height && result.top>=0 && result.head>=result.top,JSON.stringify(result));
    await page.locator('.modal').getByRole('button',{name:'关闭',exact:true}).click();
    await page.locator('.modal').waitFor({state:'detached'});
   }
   assert.deepEqual(errors,[]);await page.close();console.log(`PASS: SaySo ${width}/${theme}, common and dispatch modal headings stay fixed`);
  }
 } finally {await browser.close();}
})().catch(e=>{console.error(e);process.exitCode=1});
