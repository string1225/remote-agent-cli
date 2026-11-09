const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {chromium} = require(process.env.PLAYWRIGHT_MODULE || 'playwright');

(async () => {
  const browser = await chromium.launch({headless:true,channel:process.env.PLAYWRIGHT_CHANNEL || 'msedge'});
  try {
    for (const width of [1440,390]) {
      const context = await browser.newContext({viewport:{width,height:960},reducedMotion:'reduce',userAgent:width===390?'RemoteAgentAndroid/1.0':undefined});
      const page = await context.newPage(), errors = [], calls = [];
      const routeInfo = {clientId:'client',projectId:'project',sessionId:'persistent-session-id'};
      let failUpdate = false;
      const rows = [
        {id:'decision',title:'<img src=x onerror=alert(1)>选择发布窗口',body:'**目标**：确认发布窗口。\n[测试](javascript:alert(1))',options:[{label:'A · 周五',description:'保留周末观察时间'},{label:'B · 周一',description:'安排工作日观察'}],status:'open',version:1,createdAt:4,...routeInfo},
        {id:'notice',title:'构建已经完成',body:'只需确认已读，也可以补充下一条指令。',status:'open',version:1,createdAt:3,...routeInfo},
        {id:'history',title:'历史事项',body:'已处理的事项',status:'completed',version:1,createdAt:2},
        {id:'removed',title:'删除的事项',body:'保留恢复数据',status:'cancelled',version:1,createdAt:1},
      ];
      page.on('pageerror',error=>errors.push(error.message));
      await context.addInitScript(()=>{window.setInterval=()=>0;window.open=(...args)=>{window.openedTodoWindow=args;return null;};});
      await context.route('https://todo.test/**',async route=>{
        const url=new URL(route.request().url()), file=url.pathname.slice(1)||'index.html';
        if(file==='api/me')return route.fulfill({json:{id:'owner',username:'测试用户'}});
        if(file==='api/agents')return route.fulfill({json:[{id:'client',name:'工作电脑',os:'windows',enrolled:true,online:false,services:[{id:'project',name:'项目',provider:'codex'}]}]});
        if(file==='api/default-host')return route.fulfill({json:{agentId:'client',name:'工作电脑',online:false}});
        if(file==='api/todos'){
          if(route.request().method()==='GET')return route.fulfill({json:{todos:rows.filter(row=>(!url.searchParams.get('status')||row.status===url.searchParams.get('status'))&&(!url.searchParams.get('search')||(row.title+row.body).includes(url.searchParams.get('search'))))}});
          const command=route.request().postDataJSON();calls.push(command);
          if(command.operation==='create')return route.fulfill({json:{id:'new',...command,status:'open',version:1}});
          const row=rows.find(row=>row.id===command.id);
          assert.equal(command.version,row.version,'always submit the displayed version');
          if(command.operation==='update'&&failUpdate){failUpdate=false;return route.fulfill({status:409,json:{error:'模拟保存冲突'}});}
          row.version++;
          if(command.operation==='update'){row.title=command.title;row.body=command.body;}
          if(command.operation==='complete')row.status='completed';
          if(command.operation==='delete')row.status='cancelled';
          if(command.operation==='reopen')row.status='open';
          if(command.operation==='reply'){row.status='completed';row.reply={id:command.requestId,text:command.message,state:'delivered'};}
          return route.fulfill({json:row});
        }
        if(file.startsWith('api/'))return route.fulfill({status:409,json:{error:'模拟环境'}});
        return route.fulfill({contentType:file.endsWith('.js')?'text/javascript':file.endsWith('.css')?'text/css':file.endsWith('.svg')?'image/svg+xml':'text/html',body:fs.readFileSync(path.join('web',file))});
      });
      await page.goto('https://todo.test/');await page.locator('#app').waitFor({state:'visible'});
      if(width===390)await page.locator('#open-devices').click();
      await page.locator('#todos-entry').click();
      const dialog=page.locator('#todos-dialog'), decision=dialog.locator('[data-todo-id=decision]'), notice=dialog.locator('[data-todo-id=notice]');
      await decision.waitFor();
      const filters=dialog.getByRole('group',{name:'待办状态'});
      assert.equal(await filters.getByRole('button').count(),3);
      assert.equal(await filters.getByRole('button',{name:'待处理',exact:true}).getAttribute('aria-pressed'),'true');
      assert.equal(await filters.getByRole('button',{name:'已处理',exact:true}).getAttribute('aria-pressed'),'false');
      assert.equal(await dialog.locator('.todo-card').count(),2);
      assert.equal(await dialog.getByRole('button',{name:'刷新待办'}).locator('svg').count(),1);
      assert.equal(await decision.locator('h3 img').count(),0);
      assert.equal(await decision.locator('a[href^="javascript:"]').count(),0);
      assert.equal(await decision.getByRole('radio').count(),2);
      assert.equal(await notice.getByRole('radio').count(),0,'notifications do not invent choices');
      await dialog.locator('.todo-host-name').click();
      await page.getByRole('dialog',{name:'选择默认主机'}).waitFor();
      await page.getByRole('dialog',{name:'选择默认主机'}).getByRole('button',{name:'取消',exact:true}).click();
      await decision.getByRole('button',{name:'在新窗口打开会话'}).click();
      let opened;
      if (width===390) {
        const viewer=page.getByRole('dialog',{name:'来源会话',exact:true});await viewer.waitFor();
        opened=[await viewer.locator('iframe').getAttribute('src')];
        await viewer.frameLocator('iframe').locator('#app').waitFor({state:'visible'});
        await page.waitForFunction(()=>{
          const frame=document.querySelector('.todo-session-dialog iframe');
          return frame?.contentWindow?.eval('state.session?.id')==='persistent-session-id';
        });
        await viewer.getByRole('button',{name:'关闭来源会话',exact:true}).click();
      } else {
        opened=await page.evaluate(()=>window.openedTodoWindow);
        assert(opened[2].includes('popup=yes')&&opened[2].includes('noopener'));
      }
      const link=new URL(opened[0]);
      assert.equal(link.searchParams.get('client'),'client');assert.equal(link.searchParams.get('project'),'project');assert.equal(link.searchParams.get('session'),'persistent-session-id');
      assert.equal(await dialog.isVisible(),true,'opening a session leaves the list open');
      const bounds=await decision.evaluate(el=>({card:el.getBoundingClientRect().toJSON(),tools:el.querySelector('.todo-card-tools').getBoundingClientRect().toJSON(),options:[...el.querySelectorAll('.todo-option')].map(option=>option.getBoundingClientRect().toJSON())}));
      assert(bounds.tools.right<=bounds.card.right&&bounds.tools.top<bounds.options[0].top,'icons sit at the upper right');
      assert(bounds.options[1].top>=bounds.options[0].bottom,'choices stack vertically');
      await decision.getByRole('button',{name:'编辑待办'}).click();
      assert.equal(await page.locator('.todo-editor').count(),0,'edit inside the list card');
      await decision.getByRole('textbox',{name:'待办标题',exact:true}).fill('已更新的发布窗口');
      await decision.getByRole('textbox',{name:'待办详情',exact:true}).fill('**更新说明**\n保留两个选项');
      failUpdate=true;await decision.getByRole('button',{name:'保存待办'}).click();
      await dialog.locator('.todo-error').filter({hasText:'模拟保存冲突'}).waitFor();
      assert.equal(await decision.getByRole('textbox',{name:'待办标题',exact:true}).inputValue(),'已更新的发布窗口','failed saves preserve edited text');
      await decision.getByRole('button',{name:'保存待办'}).click();
      await decision.getByRole('heading',{name:'已更新的发布窗口',exact:true}).waitFor();
      await decision.getByRole('radio',{name:/A · 周五/}).check();
      await decision.getByRole('textbox',{name:'补充指令'}).fill('先检查线上健康');
      if(process.env.UI_TEST_SCREENSHOT_DIR){fs.mkdirSync(process.env.UI_TEST_SCREENSHOT_DIR,{recursive:true});await dialog.screenshot({path:path.join(process.env.UI_TEST_SCREENSHOT_DIR,`todos-decisions-${width}.png`)});}
      await decision.getByRole('button',{name:'发送回复',exact:true}).click();await decision.waitFor({state:'detached'});
      assert.equal(rows[0].status,'completed');
      assert.match(calls.find(c=>c.operation==='reply').message,/A · 周五：保留周末观察时间.*\n\n补充：先检查线上健康/);
      await notice.getByRole('button',{name:'标记完成',exact:true}).click();await notice.waitFor({state:'detached'});
      assert.equal(calls.filter(c=>c.operation==='reply').length,1,'marking complete never sends a reply');
      await dialog.getByRole('heading',{name:'暂无待处理任务'}).waitFor();
      await filters.getByRole('button',{name:'已处理',exact:true}).click();
      assert.equal(await filters.getByRole('button',{name:'待处理',exact:true}).getAttribute('aria-pressed'),'true','status buttons are multi-select');
      await notice.waitFor();await notice.getByRole('button',{name:'重新打开'}).click();
      await filters.getByRole('button',{name:'已处理',exact:true}).click();
      await notice.waitFor();await notice.getByRole('button',{name:'删除',exact:true}).click();await notice.waitFor({state:'detached'});
      assert.equal(calls.filter(c=>c.operation==='reply').length,1,'deletion never sends a reply');
      await filters.getByRole('button',{name:'已删除',exact:true}).click();await notice.waitFor();
      await notice.getByRole('button',{name:'恢复',exact:true}).click();await notice.waitFor();
      await filters.getByRole('button',{name:'已删除',exact:true}).click();
      await page.waitForFunction(()=>document.querySelector('.todo-list').getAttribute('aria-busy')==='false');
      await notice.getByRole('textbox',{name:'补充指令'}).fill('继续检查构建产物');
      await dialog.getByRole('button',{name:'关闭待办',exact:true}).click();
      if(width===390)await page.locator('#open-devices').click();
      await page.locator('#todos-entry').click();await notice.waitFor();
      assert.equal(await notice.getByRole('textbox',{name:'补充指令'}).inputValue(),'继续检查构建产物','reply drafts survive closing the dialog');
      await notice.getByRole('button',{name:'发送回复',exact:true}).click();await notice.waitFor({state:'detached'});
      assert.equal(calls.filter(c=>c.operation==='reply').at(-1).message,'继续检查构建产物','notification replies accept arbitrary instructions');
      await filters.getByRole('button',{name:'已处理',exact:true}).click();await decision.waitFor();
      await dialog.getByRole('searchbox',{name:'搜索待办'}).fill('没有匹配项');
      await dialog.getByRole('heading',{name:'没有匹配的待办'}).waitFor();
      assert.equal(await dialog.evaluate(el=>el.scrollWidth<=el.clientWidth),true,'phone layout has no horizontal overflow');
      const popup=await context.newPage();await popup.goto(opened[0]);
      await popup.waitForFunction(()=>state.selected?.id==='client'&&state.serviceId==='project'&&state.session?.id==='persistent-session-id');
      await popup.close();
      assert.deepEqual(errors,[]);await context.close();
    }
    console.log('PASS: desktop/mobile choices, free replies, inline edits/failure drafts, independent complete/delete, multiselect, host picker and isolated query navigation');
  }finally{await browser.close();}
})().catch(error=>{console.error(error);process.exitCode=1;});
