const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {chromium} = require(process.env.PLAYWRIGHT_MODULE || 'playwright');

(async () => {
  const browser = await chromium.launch({headless:true,channel:process.env.PLAYWRIGHT_CHANNEL || 'msedge'});
  try {
    for (const [width,height,theme] of [[375,812,'dark'],[812,375,'light'],[390,844,'light'],[1024,768,'dark']]) {
      const page = await browser.newPage({viewport:{width,height},userAgent:'RemoteAgentAndroid/1.0 RemoteAgentNative/1',reducedMotion:'reduce',isMobile:true,hasTouch:true});
      const errors = [];
      page.on('pageerror',error => errors.push(error.message));
      await page.addInitScript(theme => {
        localStorage.setItem('remote-agent.theme',theme);
        window.setInterval = () => 0;
        window.phoneHost = {version:'1.0-test',site:'https://phone.test/',autoUpdate:true,updateBusy:false,updateStatus:'当前已是最新版本'};
        window.phoneCommands = [];
        window.phoneNativeCommand = (action,params) => {
          window.phoneCommands.push({action,params});
          if (action === 'auto-update') window.phoneHost.autoUpdate = params.enabled;
          if (action === 'update') {window.phoneHost.updateBusy = true;window.phoneHost.updateStatus = '正在下载并校验…';}
          window.dispatchEvent(new CustomEvent('remote-agent-android-state',{detail:{...window.phoneHost}}));
        };
      },theme);
      await page.route('https://phone.test/**',route => {
        const file = new URL(route.request().url()).pathname.slice(1) || 'index.html';
        if (file === 'api/me') return route.fulfill({json:{id:'phone-owner',username:'手机测试'}});
        if (file === 'api/agents') return route.fulfill({json:[]});
        if (file === 'api/default-host') return route.fulfill({json:{}});
        if (file.startsWith('api/')) return route.fulfill({status:409,json:{error:'模拟环境'}});
        let body = fs.readFileSync(path.join('web',file));
        // Exercise the real settings UI with a simulated native transport; no updates or site changes.
        if (file === 'android-app.js') body = body.toString().replace('location.href = `remoteagent://${action}${query ? `?${query}` : ""}`;', 'window.phoneNativeCommand(action, params);');
        return route.fulfill({contentType:file.endsWith('.css')?'text/css':file.endsWith('.js')?'text/javascript':'text/html',body});
      });
      await page.goto('https://phone.test/');
      await page.locator('#app').waitFor({state:'visible'});
      const header = page.locator('.chat-header');
      if (process.env.UI_TEST_SCREENSHOT_DIR) {
        fs.mkdirSync(process.env.UI_TEST_SCREENSHOT_DIR,{recursive:true});
        await page.screenshot({path:path.join(process.env.UI_TEST_SCREENSHOT_DIR,`phone-workbench-${width}.png`)});
      }
      assert.equal((await header.boundingBox()).y,0,'workbench begins at the top without a second toolbar');
      assert.equal(await page.locator('#new-agent').isEnabled(),false);
      const entry = page.getByRole('button',{name:'手机设置',exact:true}).filter({has:page.locator('svg')});
      const target = await entry.boundingBox();
      assert(target.width >= 48 && target.height >= 48,'phone settings have an Android-size touch target');
      await entry.click();
      const dialog = page.getByRole('dialog',{name:'手机设置',exact:true});
      await dialog.waitFor();
      await page.waitForFunction(()=>document.getElementById('phone-site').value==='https://phone.test/');
      assert.equal(await dialog.locator('#phone-version').textContent(),'版本 1.0-test');
      assert.equal(await dialog.locator('#phone-auto-update').isChecked(),true);
      if (process.env.UI_TEST_SCREENSHOT_DIR) await page.screenshot({path:path.join(process.env.UI_TEST_SCREENSHOT_DIR,`phone-settings-initial-${width}.png`)});
      await dialog.getByRole('switch',{name:'自动检查更新'}).click();
      assert.equal(await dialog.locator('#phone-auto-update').isChecked(),false,'host acknowledgement updates the saved switch');
      await dialog.locator('#phone-check-update').click();
      assert.equal(await dialog.locator('#phone-check-update').isEnabled(),false,'duplicate update requests are blocked while downloading');
      assert.match(await dialog.locator('[role=status]').textContent(),/下载并校验/);
      await dialog.locator('#phone-site').fill('http://invalid.test/');
      await dialog.locator('#phone-save-site').click();
      assert.equal(await dialog.locator('#phone-site-error').isVisible(),true);
      assert.equal(await page.evaluate(()=>window.phoneCommands.filter(c=>c.action==='site').length),0);
      await dialog.locator('#phone-site').fill('https://next.test/agents/');
      await dialog.locator('#phone-save-site').click();
      assert.deepEqual(await page.evaluate(()=>window.phoneCommands.find(c=>c.action==='site').params),{url:'https://next.test/agents/'});
      assert.equal(await dialog.locator('#phone-site').inputValue(),'https://next.test/agents/','background update events preserve edits');
      await dialog.locator('#phone-appearance').click();
      await page.getByRole('dialog',{name:'界面设置'}).waitFor();
      await page.getByRole('radio',{name:theme==='dark'?'明亮主题':'暗黑主题'}).check();
      assert.equal(await page.evaluate(()=>window.phoneCommands.at(-1).action),'theme');
      assert.equal(await page.evaluate(()=>window.remoteAgentAndroidBack()),true);
      assert.equal(await dialog.isVisible(),true,'back closes nested appearance settings first');
      assert.equal(await page.locator('#appearance-dialog').isVisible(),false);
      const bounds = await dialog.boundingBox();
      assert(bounds.x>=0 && bounds.y>=0 && bounds.x+bounds.width<=width+1 && bounds.y+bounds.height<=height+1,'settings fit portrait and landscape');
      assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth),width,'no horizontal page overflow');
      await dialog.locator('#phone-reload').click();
      assert.equal(await page.evaluate(()=>window.phoneCommands.at(-1).action),'reload');
      if (process.env.UI_TEST_SCREENSHOT_DIR) {
        fs.mkdirSync(process.env.UI_TEST_SCREENSHOT_DIR,{recursive:true});
        await dialog.locator('.dialog-scroll-body').evaluate(el=>el.scrollTop=0);
        await page.screenshot({path:path.join(process.env.UI_TEST_SCREENSHOT_DIR,`phone-settings-${width}.png`)});
      }
      assert.equal(await page.evaluate(()=>window.remoteAgentAndroidBack()),true);
      assert.equal(await dialog.isVisible(),false);
      if (width <= 760) {
        await page.locator('#open-devices').click();
        await page.locator('#account-toggle').click();
        assert.equal(await page.evaluate(()=>window.remoteAgentAndroidBack()),true);
        assert.equal(await page.locator('#account-popover').isVisible(),false);
        assert.equal(await page.locator('#devices-pane').evaluate(el=>el.classList.contains('drawer-open')),true);
        assert.equal(await page.evaluate(()=>window.remoteAgentAndroidBack()),true);
        assert.equal(await page.locator('#drawer-backdrop').isVisible(),false);
      }
      assert.equal(await page.evaluate(()=>window.remoteAgentAndroidBack()),false,'host handles navigation only after overlays close');
      await page.evaluate(()=>{document.getElementById('app').hidden=true;document.getElementById('auth').hidden=false;});
      await page.locator('.phone-auth-settings').click();
      assert.equal(await dialog.isVisible(),true,'site settings are available before login');
      assert.deepEqual(errors,[]);
      await page.close();
    }
    console.log('PASS: integrated Android settings, host state/actions, validation, drafts, updates, nested back, login access and portrait/landscape themes');
  } finally {await browser.close();}
})().catch(error=>{console.error(error);process.exitCode=1;});
