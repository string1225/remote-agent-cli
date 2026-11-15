const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {chromium} = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const web = path.resolve(process.env.UI_WEB_ROOT || 'web');

(async () => {
  const browser = await chromium.launch({headless:true,channel:process.env.PLAYWRIGHT_CHANNEL || 'msedge'});
  try {
    for (const width of [1440,390]) for (const theme of ['light','dark']) {
      const page = await browser.newPage({viewport:{width,height:800}}), errors=[];
      let copies=0, delay=false, release;
      page.on('pageerror',error=>errors.push(error.message));
      await page.addInitScript(theme=>{
        window.setInterval=()=>0; localStorage.setItem('remote-agent.theme',theme);
        window.copied=[];
        Object.defineProperty(navigator,'clipboard',{value:{writeText:async text=>{
          if(window.failCopy) throw Error('fixture denied'); window.copied.push(text);
        }}});
      },theme);
      await page.route('https://onboarding.test/**',async route=>{
        const file=new URL(route.request().url()).pathname.slice(1)||'index.html';
        if(file==='api/me') return route.fulfill({json:{id:'admin',username:'管理员',admin:true}});
        if(file==='api/admin/users') return route.fulfill({json:[{id:'member',username:'新用户',deviceCount:0},{id:'disabled',username:'停用账号',disabled:true,deviceCount:0}]});
        if(file==='api/admin/users/member/endpoint-guide') {
          copies++; if(delay) await new Promise(resolve=>{release=resolve;});
          return route.fulfill({json:{token:'ra_codex_fixture-guide-token',tokenId:'guide-token'}});
        }
        if(file.startsWith('api/')) return route.fulfill({json:[]});
        const filename=path.join(web,file);
        if(!fs.existsSync(filename)) return route.fulfill({status:404,body:''});
        return route.fulfill({contentType:file.endsWith('.css')?'text/css':file.endsWith('.js')?'text/javascript':file.endsWith('.svg')?'image/svg+xml':'text/html',body:fs.readFileSync(filename)});
      });
      await page.goto('https://onboarding.test/');
      await page.locator('#app').waitFor({state:'visible'});
      // Every existing static/dynamically registered dialog, plus the dynamic
      // form-only and workbench variants. No real account/session data is used.
      await page.evaluate(()=>{
        window.openAppearanceSettings(); document.querySelector('#appearance-dialog').close();
        for(const cls of ['todo-host-dialog','todo-editor','work-reply-dialog','workbench-dialog','todo-session-dialog']) {
          const d=document.createElement('dialog');d.className=cls;d.id='fixture-'+cls;
          d.innerHTML=cls==='workbench-dialog'?'<div class="workbench-dialog-heading"><h3>工作区</h3><button type="button">关闭</button></div><div class="workbench-dialog-body">内容</div>':'<form><h2>测试标题</h2><input value="未提交草稿"><button type="button">取消</button></form>';
          d.querySelector('button').onclick=()=>d.close();document.body.append(d);
        }
      });
      await page.waitForFunction(()=>[...document.querySelectorAll('dialog')].every(d=>d.classList.contains('fixed-dialog-header')));
      const ids=await page.locator('dialog').evaluateAll(ds=>ds.map(d=>d.id));
      for(const id of ids) {
        const result=await page.evaluate(async id=>{
          const d=document.getElementById(id), body=d.querySelector('.dialog-scroll-body'), head=d.querySelector('.fixed-dialog-heading');
          const filler=document.createElement('div');filler.style.height='1800px';filler.style.flexShrink='0';filler.textContent='滚动测试';body.append(filler);
          d.showModal();await new Promise(requestAnimationFrame);
          const before=head.getBoundingClientRect(); body.scrollTop=800;await new Promise(requestAnimationFrame);
          const after=head.getBoundingClientRect(),box=d.getBoundingClientRect();
          const r={id,delta:Math.abs(before.top-after.top),scroll:body.scrollTop,headTop:after.top,dialogTop:box.top,bottom:box.bottom,viewport:innerHeight,overflow:document.documentElement.scrollWidth>innerWidth};
          const closed=new Promise(resolve=>d.addEventListener('close',resolve,{once:true}));d.close();await closed;filler.remove();return r;
        },id);
        assert(result.delta<1,JSON.stringify(result));
        if(id!=='fixture-todo-session-dialog') assert(result.scroll>0,JSON.stringify(result));
        assert(result.headTop>=result.dialogTop && result.bottom<=result.viewport+1,JSON.stringify(result));
        assert(!result.overflow,JSON.stringify(result));
      }
      await page.evaluate(()=>{const d=document.getElementById('fixture-work-reply-dialog');d.showModal();});
      await page.locator('#fixture-work-reply-dialog > .fixed-dialog-heading button').click();
      assert.equal(await page.locator('#fixture-work-reply-dialog input').inputValue(),'未提交草稿');
      // Endpoint copy contains the supplied token and no web-login/password step.
      await page.evaluate(()=>document.getElementById('admin-entry').click());
      const endpoint=page.getByRole('button',{name:'为 新用户 复制 Endpoint / Codex 指引',exact:true});
      await endpoint.click();
      await page.waitForFunction(()=>window.copied.length===1);
      const guide=await page.evaluate(()=>window.copied[0]);
      assert(guide.includes('API Token：ra_codex_fixture-guide-token'));
      assert(guide.includes('下载安装客户端') && guide.includes('默认不勾选'));
      assert(!guide.includes('初始密码') && !guide.includes('登录后进入') && !guide.includes('用浏览器打开'));
      assert(guide.includes('不新增远程设备'));
      assert(await page.getByRole('button',{name:'为 停用账号 复制 Endpoint / Codex 指引',exact:true}).isDisabled());
      // Clipboard denied: manual fallback clears when the dialog closes.
      await page.evaluate(()=>{window.failCopy=true;});await endpoint.click();
      await page.waitForFunction(()=>document.getElementById('admin-guide-text').value.includes('ra_codex_fixture'));
      assert.equal(copies,2);
      await page.evaluate(()=>document.getElementById('platform-dialog').close());
      await page.waitForFunction(()=>document.getElementById('admin-guide-text').value==='');
      // A delayed response must never repopulate a closed/changed account view.
      delay=true;
      await page.evaluate(()=>document.getElementById('admin-entry').click());
      await endpoint.click();
      await page.waitForTimeout(50);
      assert(release,'guide request not started');
      await page.evaluate(()=>document.getElementById('platform-dialog').close());
      release();await page.waitForTimeout(100);
      assert.equal(await page.locator('#admin-guide-text').inputValue(),'');
      assert.equal(await page.evaluate(()=>window.copied.length),1);
      assert.deepEqual(errors,[]);
      if(process.env.UI_TEST_SCREENSHOT_DIR) {
        fs.mkdirSync(process.env.UI_TEST_SCREENSHOT_DIR,{recursive:true});
        await page.evaluate(()=>document.getElementById('admin-entry').click());
        await page.screenshot({path:path.join(process.env.UI_TEST_SCREENSHOT_DIR,`dialogs-${width}-${theme}.png`)});
      }
      await page.close();
      console.log(`PASS: ${width}/${theme}, ${ids.length} fixed headings, token guide, fallback/close/late response`);
    }
  } finally {await browser.close();}
})().catch(error=>{console.error(error);process.exitCode=1;});
