const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {chromium} = require(process.env.PLAYWRIGHT_MODULE || 'playwright');

(async () => {
  const browser = await chromium.launch({headless:true,channel:process.env.PLAYWRIGHT_CHANNEL || 'msedge'});
  try {
    for (const width of [1440,390]) {
      const page = await browser.newPage({viewport:{width,height:900},userAgent:width === 390 ? 'RemoteAgentAndroid/1.0' : undefined});
      const errors = []; let logout = 0;
      page.on('pageerror',error => errors.push(error.message));
      await page.addInitScript(() => {
        window.setInterval = () => 0;
      });
      await page.route('https://account.test/**', route => {
        const file = new URL(route.request().url()).pathname.slice(1) || 'index.html';
        if (file === 'api/me') return route.fulfill({json:{id:'owner',username:'测试用户',admin:false}});
        if (file === 'api/agents') return route.fulfill({json:[]});
        if (file === 'api/default-host') return route.fulfill({json:{agentId:'host',name:'工作电脑',online:true}});
        if (file === 'api/codex') return route.fulfill({json:{enabled:false}});
        if (file === 'api/tokens') return route.fulfill({json:[]});
        if (file === 'api/logout') {logout++;return route.fulfill({json:{ok:true}});}
        if (file.startsWith('api/')) return route.fulfill({status:409,json:{error:'模拟环境'}});
        let body = fs.readFileSync(path.join('web',file));
        if (file === 'android-app.js') body = body.toString().replace('location.href = "remoteagent://settings";', 'window.androidSettingsOpened = true;');
        return route.fulfill({contentType:file.endsWith('.css')?'text/css':file.endsWith('.js')?'text/javascript':'text/html',body});
      });
      await page.goto('https://account.test/'); await page.locator('#app').waitFor({state:'visible'});
      if (width === 390) await page.locator('#open-devices').click();
      assert.equal(await page.locator('.platform-nav button').count(),1,'only todos remain above the username');
      assert.equal(await page.locator('.platform-nav #todos-entry').count(),1);
      const help = page.getByRole('button',{name:'Codex 接入与帮助',exact:true});
      const brand = await page.locator('.brand-heading .wordmark').boundingBox(), helpBounds = await help.boundingBox();
      assert(helpBounds.x >= brand.x + brand.width,'help sits to the right of the brand');
      assert.equal(await page.locator('#account-popover').isVisible(),false);
      const trigger = page.locator('#account-toggle');
      await trigger.click();
      const menu = page.getByRole('menu',{name:'账号菜单'}); await menu.waitFor();
      assert.equal(await trigger.getAttribute('aria-expanded'),'true');
      assert.equal(await menu.getByRole('menuitem',{name:/账号管理/}).count(),0,'non-admins cannot see admin controls');
      assert.equal(await menu.getByRole('menuitem',{name:/我的代理 Token/}).count(),1);
      assert.equal(await menu.getByRole('menuitem',{name:/服务器与 GitHub/}).count(),1);
      assert.equal(await menu.getByRole('menuitem',{name:/接入与帮助/}).count(),0);
      const bounds = await menu.boundingBox();
      assert(bounds.x >= 0 && bounds.x + bounds.width <= width && bounds.y >= 0,'popover stays inside the viewport');
      if (process.env.UI_TEST_SCREENSHOT_DIR) {
        fs.mkdirSync(process.env.UI_TEST_SCREENSHOT_DIR,{recursive:true});
        await page.screenshot({path:path.join(process.env.UI_TEST_SCREENSHOT_DIR,`account-menu-${width}.png`)});
      }
      await page.keyboard.press('End');
      assert.equal(await page.evaluate(()=>document.activeElement.id),'logout');
      await page.keyboard.press('Escape'); await menu.waitFor({state:'hidden'});
      assert.equal(await page.evaluate(()=>document.activeElement.id),'account-toggle');
      await page.keyboard.press('ArrowUp'); await menu.waitFor();
      await page.waitForFunction(()=>document.activeElement.id === 'logout');
      await page.keyboard.press('Escape');
      await trigger.click();
      await menu.getByRole('menuitem',{name:/账号设置/}).click();
      await page.locator('#platform-dialog').waitFor({state:'visible'});
      assert.equal(await menu.isVisible(),false,'account actions dismiss the popover before opening a dialog');
      assert.equal(await page.locator('#platform-account').isVisible(),true);
      await page.evaluate(()=>document.getElementById('platform-dialog').close());
      if (width === 390) await page.locator('#open-devices').click();
      await help.click();
      await page.locator('#platform-setup').waitFor({state:'visible'});
      await page.evaluate(()=>document.getElementById('platform-dialog').close());
      if (width === 390) await page.locator('#open-devices').click();
      await trigger.click();
      await page.locator('.device-tools .eyebrow').click();
      await menu.waitFor({state:'hidden'});
      await page.evaluate(()=>{state.user.admin=true;window.platformSignedIn(state.user)});
      await trigger.click();
      assert.equal(await menu.getByRole('menuitem',{name:/账号管理/}).count(),1,'existing admin role still controls visibility');
      if (width === 390) {
        await menu.getByRole('menuitem',{name:/安卓更新与设置/}).click();
        assert.equal(await page.evaluate(()=>window.androidSettingsOpened),true,'Android settings retain their handoff handler');
        await trigger.click();
      }
      await menu.getByRole('menuitem',{name:'退出登录',exact:true}).click();
      await page.locator('#auth').waitFor({state:'visible'});
      assert.equal(logout,1,'logout retains the existing authenticated action');
      assert.equal(await menu.isVisible(),false);
      assert.deepEqual(errors,[]);
      await page.close();
    }
    console.log('PASS: desktop/Android compact sidebar, help, account actions, role visibility, keyboard/light dismiss, update and logout');
  } finally {await browser.close();}
})().catch(error=>{console.error(error);process.exitCode=1;});
