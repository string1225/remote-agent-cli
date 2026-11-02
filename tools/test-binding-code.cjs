const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {chromium} = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
(async () => {
  const browser = await chromium.launch({headless:true,channel:process.env.PLAYWRIGHT_CHANNEL || 'msedge'});
  try {
    for (const width of [1440,390]) {
      const page = await browser.newPage({viewport:{width,height:1000}}), errors=[];
      page.on('pageerror', error => errors.push(error.message));
      await page.addInitScript(() => { window.setInterval=()=>0; });
      const device={id:'pending',name:'测试电脑',os:'windows',enrolled:false,online:false,services:[]};
      let code='fixture-initial-code',readCount=0;
      const binding=()=>({agent:device,bindingCode:code,expiresAt:Math.floor(Date.now()/1000)+900,windows:'legacy-command',windowsInstall:'msiexec BINDING_CODE='+code,windowsMsi:'/downloads/x64.msi',windowsArm64Msi:'/downloads/arm64.msi'});
      await page.route('https://binding.test/**', route => {
        const request=route.request(),file=new URL(request.url()).pathname.slice(1)||'index.html';
        if(file==='api/me')return route.fulfill({json:{id:'owner',username:'测试账号'}});
        if(file==='api/agents')return route.fulfill({json:request.method()==='POST'?binding():[device]});
        if(file==='api/agents/pending/binding-code') {
          readCount++;if(request.postDataJSON().renew)code='fixture-renewed-code';
          return route.fulfill({json:binding()});
        }
        if(file.startsWith('api/'))return route.fulfill({json:[]});
        const filename=file==='app.js'&&process.env.UI_TEST_APP_SOURCE?process.env.UI_TEST_APP_SOURCE:path.join('web',file);
        return route.fulfill({contentType:file.endsWith('.js')?'text/javascript':file.endsWith('.css')?'text/css':file.endsWith('.svg')?'image/svg+xml':'text/html',body:fs.readFileSync(filename)});
      });
      await page.goto('https://binding.test/');
      await page.locator('#app').waitFor({state:'visible'});
      await page.evaluate(()=>{showEnroll();document.querySelector('#enroll-form [name=name]').value='测试电脑';document.querySelector('#enroll-form [name=os]').value='windows';});
      await page.locator('#enroll-form button[type=submit]').click();
      await page.waitForFunction(()=>document.querySelector('#device-binding-code').value==='fixture-initial-code');
      assert((await page.locator('#install-command').inputValue()).includes('BINDING_CODE=fixture-initial-code'));
      await page.locator('#enroll-dialog .close-dialog').click();
      await page.waitForFunction(()=>document.querySelector('#device-binding-code').value==='');
      assert.equal(await page.locator('#device-binding-code').inputValue(),'');
      if(width<760)await page.locator('#open-devices').click();
      await page.locator('.device-menu-button').click();
      await page.getByRole('menuitem',{name:'查看绑定码',exact:true}).click();
      await page.waitForFunction(()=>document.querySelector('#device-binding-code').value==='fixture-initial-code');
      await page.locator('#renew-device-binding-code').click();
      await page.waitForFunction(()=>document.querySelector('#device-binding-code').value==='fixture-renewed-code');
      assert.equal(readCount,2);
      assert((await page.locator('#install-command').inputValue()).includes('fixture-renewed-code'));
      assert(!(await page.evaluate(()=>JSON.stringify(localStorage))).includes('fixture-'));
      assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth),width);
      assert.deepEqual(errors,[]);
      await page.close();
    }
    console.log('PASS: desktop/mobile binding code, installation command prefill, reopen, renewal, secret clearing and no cache');
  } finally {await browser.close();}
})().catch(error=>{console.error(error);process.exitCode=1;});
