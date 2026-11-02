const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {chromium} = require(process.env.PLAYWRIGHT_MODULE || 'playwright');

(async () => {
  const browser = await chromium.launch({headless:true,channel:process.env.PLAYWRIGHT_CHANNEL || 'chrome'});
  try {
    for (const width of [1440,390]) {
      const page = await browser.newPage({viewport:{width,height:900}}), errors=[];
      page.setDefaultTimeout(8000);
      page.on('pageerror', e=>errors.push(e.message));
      await page.addInitScript(()=>{window.setInterval=()=>0;});
      await page.route('https://directory.test/**', route=>{
        const file=new URL(route.request().url()).pathname.slice(1)||'index.html';
        if(file.startsWith('api/'))return route.fulfill({status:401,json:{error:'please sign in'}});
        return route.fulfill({contentType:file.endsWith('.js')?'text/javascript':file.endsWith('.css')?'text/css':'text/html',body:fs.readFileSync(path.join('web',file))});
      });
      await page.goto('https://directory.test/');
      await page.evaluate(()=>{
        window.fixture={calls:[],fail:false,delay:false};
        const capabilities=()=>({services:[],providers:{codex:{available:true}},allowedRoots:['C:\\Code'],permissionMode:'full',directorySearch:true,directoryCreate:true});
        rpc=async(type,fields={})=>{
          fixture.calls.push({type,...fields});
          if(type==='directories.search')return {directories:[{name:'Code',path:'C:\\Code'}],base:fields.path||fields.search||'',parent:'',separator:'\\'};
          if(type==='directories.create'){
            if(fixture.fail){fixture.fail=false;throw Error('没有权限新建文件夹');}
            const result={name:fields.title,path:fields.path+'\\'+fields.title,created:true};
            if(fixture.delay)return new Promise(resolve=>{fixture.finish=()=>resolve(result)});
            return result;
          }
          if(type==='services.add')return {id:'new-project',...fields.service,session:{id:'first-session'}};
          if(type==='services.list')return capabilities();
          throw Error('unexpected '+type);
        };
        selectProject=async id=>{fixture.project=id};loadSession=async session=>{fixture.session=session};
        state.user={id:'tester'};state.selected={id:'client',name:'测试电脑'};state.peer={ws:{readyState:WebSocket.OPEN}};
        refreshServices(capabilities());$('auth').hidden=true;$('app').hidden=false;showService();
      });
      const service=page.locator('#service-dialog'), folder=page.locator('#directory-create-dialog');
      await service.getByLabel('服务名称').fill('新的项目');
      await page.locator('#workspace-input').focus();
      await service.getByRole('button',{name:'浏览子目录 Code'}).click();
      await page.locator('#directory-create').click();
      await folder.getByLabel('文件夹名称').fill('新项目');
      await folder.getByRole('button',{name:'取消',exact:true}).click();
      assert.equal(await page.locator('#workspace-input').inputValue(),'C:\\Code');
      await page.locator('#workspace-input').focus();
      await page.locator('#directory-create').click();
      await folder.getByLabel('文件夹名称').fill('新项目');
      await page.evaluate(()=>{fixture.fail=true});
      await folder.getByRole('button',{name:'创建并选择'}).click();
      await folder.getByRole('alert').filter({hasText:'没有权限'}).waitFor();
      assert.equal(await folder.getByLabel('文件夹名称').inputValue(),'新项目','failure preserves name for retry');
      if(process.env.UI_TEST_SCREENSHOT_DIR){fs.mkdirSync(process.env.UI_TEST_SCREENSHOT_DIR,{recursive:true});await page.screenshot({path:path.join(process.env.UI_TEST_SCREENSHOT_DIR,`directory-create-${width}.png`)});}
      await folder.getByLabel('文件夹名称').press('Enter');
      await folder.waitFor({state:'hidden'});
      assert.equal(await page.locator('#workspace-input').inputValue(),'C:\\Code\\新项目');
      assert.equal(await page.evaluate(()=>fixture.calls.filter(c=>c.type==='services.add').length),0,'creating folder must not submit project form');
      await page.locator('#service-submit').click();await service.waitFor({state:'hidden'});
      assert.equal(await page.evaluate(()=>fixture.calls.find(c=>c.type==='services.add').service.workspace),'C:\\Code\\新项目');
      assert.equal(await page.evaluate(()=>fixture.session.id),'first-session');
      await page.evaluate(()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve))));
      // Late responses after reconnecting must not overwrite the new client's form.
      await page.evaluate(()=>{showService();fixture.delay=true});
      await page.locator('#workspace-input').fill('C:\\Code');await page.locator('#directory-create').click();
      await folder.getByLabel('文件夹名称').fill('迟到的目录');await folder.getByRole('button',{name:'创建并选择'}).click();
      await page.waitForFunction(()=>!!fixture.finish);
      await page.evaluate(()=>{state.peer={ws:{readyState:WebSocket.OPEN}};$('workspace-input').value='D:\\Other';fixture.finish()});
      await page.waitForTimeout(50);
      assert.equal(await page.locator('#workspace-input').inputValue(),'D:\\Other');
      assert.deepEqual(errors,[]);await page.close();
    }
    console.log('PASS: desktop/mobile folder creation, cancellation, retry, automatic project path, Enter isolation and stale-client response guard');
  } finally {await browser.close();}
})().catch(e=>{console.error(e);process.exitCode=1});
