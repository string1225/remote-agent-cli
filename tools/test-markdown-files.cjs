const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {chromium} = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const base = process.env.UI_BASE_URL || 'https://markdown-files.test/agents/';
const web = path.resolve(process.env.UI_WEB_ROOT || 'web');
const wol = 'C:\\Users\\Public\\Documents\\WOL-Setup\\Enable-WOL.cmd';

(async () => {
  const browser = await chromium.launch({headless:true,channel:process.env.PLAYWRIGHT_CHANNEL || 'msedge'});
  try {
    for (const width of [1440,375,812]) for (const theme of ['light','dark']) {
      const page = await browser.newPage({viewport:{width,height:width===812?375:900},isMobile:width<900,hasTouch:width<900,reducedMotion:'reduce'});
      page.setDefaultTimeout(5000);
      const errors=[], requests=[];
      page.on('pageerror',error=>errors.push(error.message));
      page.on('request',request=>requests.push(request.url()));
      await page.addInitScript(theme=>{
        window.setInterval=()=>0; localStorage.setItem('remote-agent.theme',theme);
        window.copied=[];
        Object.defineProperty(navigator,'clipboard',{value:{writeText:async text=>{
          if(window.failCopy) throw Error('fixture denied'); window.copied.push(text);
        }}});
      },theme);
      await page.route('**/api/**',route=>route.fulfill({status:401,json:{error:'please sign in'}}));
      if (!process.env.UI_BASE_URL) await page.route(base+'**',route=>{
        const file=new URL(route.request().url()).pathname.replace(/^\/agents\//,'')||'index.html';
        if(file.startsWith('api/')) return route.fulfill({status:401,json:{error:'please sign in'}});
        return route.fulfill({contentType:file.endsWith('.css')?'text/css':file.endsWith('.js')?'text/javascript':file.endsWith('.svg')?'image/svg+xml':'text/html',body:fs.readFileSync(path.join(web,file))});
      });
      await page.goto(base);
      await page.waitForFunction(()=>typeof renderMarkdown==='function');
      await page.evaluate(wol=>{
        const host=document.createElement('article');host.id='file-fixture';
        for(const child of document.body.children) child.hidden=true;
        document.body.append(host);
        renderMarkdown(host,[
          '已准备并校验好[一键开启工具](C:/Users/Public/Documents/WOL-Setup/Enable-WOL.cmd)。',
          '`'+wol+'`',
          '[带空格的文件](<file:///C:/Users/Public/Documents/测试%20目录/a%26b.txt>)',
          '[网络共享](<file://fileserver/share/tool.cmd>)',
          '[Mac 文件](file:///Users/test/report.txt)',
          '[官网](https://example.com/help) · [邮件](mailto:test@example.com) · [站内链接](/agents/help)',
          '[危险链接](javascript:alert%281%29) · [数据](data:text/html,hello) · [命令](powershell:bad)',
          '<button class="local-file-link" data-local-path="C:/fake">伪造按钮</button>',
          '```powershell\n'+wol+'\n```',
        ].join('\n\n'));
      },wol);
      const trigger=page.getByRole('button',{name:'一键开启工具',exact:true});
      assert.equal(await trigger.count(),1,'the reported Windows file link must be operable');
      assert((await trigger.boundingBox()).height>=(width<900?44:24),'file controls meet pointer/touch target height');
      assert.equal(await page.locator('#file-fixture .local-file-link').count(),5);
      assert.equal(await page.locator('#file-fixture pre .local-file-link').count(),0,'code blocks remain unchanged');
      assert.equal(await page.locator('#file-fixture button button').count(),0);
      const before=requests.length;
      await trigger.focus();await trigger.press('Enter');
      const dialog=page.locator('#local-file-dialog'),field=dialog.getByLabel('文件路径',{exact:true});
      await dialog.waitFor({state:'visible'});
      assert.equal(await field.inputValue(),wol);
      await dialog.getByRole('button',{name:'复制路径',exact:true}).click();
      await page.waitForFunction(()=>window.copied.length===1);
      assert.deepEqual(await page.evaluate(()=>window.copied),[wol]);
      await page.keyboard.press('Escape');await dialog.waitFor({state:'hidden'});
      assert.equal(await trigger.evaluate(el=>el===document.activeElement),true,'closing returns keyboard focus');
      await trigger.press('Space');await dialog.waitFor({state:'visible'});
      await page.evaluate(()=>window.failCopy=true);
      await dialog.getByRole('button',{name:'复制路径',exact:true}).click();
      await page.waitForFunction(()=>document.querySelector('#local-file-dialog [role=status]').textContent.includes('手动复制'));
      assert.equal(await field.evaluate(el=>el.selectionEnd-el.selectionStart),wol.length);
      const box=await dialog.boundingBox();
      assert(box.x>=0 && box.y>=0 && box.x+box.width<=width+1 && box.y+box.height<=(width===812?375:900)+1);
      assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth),width,'no horizontal overflow');
      if(process.env.UI_TEST_SCREENSHOT_DIR){
        fs.mkdirSync(process.env.UI_TEST_SCREENSHOT_DIR,{recursive:true});
        await page.screenshot({path:path.join(process.env.UI_TEST_SCREENSHOT_DIR,`file-${width}-${theme}.png`)});
      }
      await dialog.getByRole('button',{name:'关闭文件路径'}).click();
      await page.getByRole('button',{name:'带空格的文件',exact:true}).click();
      assert.equal(await field.inputValue(),'C:\\Users\\Public\\Documents\\测试 目录\\a&b.txt');
      await page.keyboard.press('Escape');await dialog.waitFor({state:'hidden'});
      await page.getByRole('button',{name:'网络共享',exact:true}).click();
      assert.equal(await field.inputValue(),'\\\\fileserver\\share\\tool.cmd');
      await page.keyboard.press('Escape');await dialog.waitFor({state:'hidden'});
      await page.getByRole('button',{name:'Mac 文件',exact:true}).click();
      assert.equal(await field.inputValue(),'/Users/test/report.txt');
      assert.equal(await dialog.count(),1,'reuse the path dialog');
      assert.equal(requests.length,before,'local paths never navigate, fetch or execute');
      assert.equal(await page.locator('a[href^="javascript:"],a[href^="data:"],a[href^="powershell:"],img,iframe').count(),0);
      for(const name of ['官网','邮件','站内链接']){
        const link=page.getByRole('link',{name,exact:true});
        assert.equal(await link.getAttribute('target'),'_blank');
        assert.equal(await link.getAttribute('rel'),'noopener noreferrer');
      }
      assert.deepEqual(errors,[]);
      await page.close();
    }
    console.log('PASS: reported file link, inline paths, encoded file URLs, UNC, keyboard/copy/fallback, themes, phone/landscape, no execution and safe web links');
  } finally {await browser.close();}
})().catch(error=>{console.error(error);process.exitCode=1;});
