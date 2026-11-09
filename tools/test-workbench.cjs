const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const {chromium} = require(process.env.PLAYWRIGHT_MODULE || "playwright");

(async () => {
  const browser = await chromium.launch({headless:true, channel:process.env.PLAYWRIGHT_CHANNEL || "msedge"});
  try {
    for (const width of [1440, 390]) {
      const page = await browser.newPage({viewport:{width,height:960}}), errors = [];
      page.on("pageerror", error => errors.push(error.message));
      await page.addInitScript(() => { window.setInterval = () => 0; });
      await page.route("https://workbench.test/**", route => {
        const file = new URL(route.request().url()).pathname.slice(1) || "index.html";
        if (file.startsWith("api/")) return route.fulfill({status:401,json:{error:"please sign in"}});
        return route.fulfill({contentType:file.endsWith(".css")?"text/css":file.endsWith(".js")?"text/javascript":"text/html",body:fs.readFileSync(path.join("web",file))});
      });
      await page.goto("https://workbench.test/");
      await page.evaluate(() => {
        state.user={id:"owner"};state.selected={id:"client",name:"测试电脑"};state.serviceId="project";
        state.services=[{id:"project",name:"Project",workspace:"/workspace",provider:"codex"}];
        state.session={id:"current",serviceId:"project",title:"Current",taskId:"task",taskStatus:"running",taskStartedAt:Date.now()-9000};
        state.sessions=[state.session,{id:"reference",serviceId:"project",title:"可引用的会话"}];
        state.peer={ws:{readyState:1}};state.permissionMode="full";
        state.entries=[
          {seq:1,turnId:"task",role:"user",kind:"message",text:"修复问题"},
          {seq:2,turnId:"task",role:"system",kind:"tool",text:"$ go test\n",activity:{id:"call",kind:"command_execution",name:"go test",status:"running",startedAt:Date.now()-5000}},
          {seq:3,turnId:"task",role:"assistant",kind:"stdout",text:"已经找到原因。"},
        ];
        window.requests=[];window.staleDiff=false;
        state.peer.ws.send=wire=>{
          const request=JSON.parse(wire).request;window.requests.push(request);
          let data;
          if(request.type==="workspace.changes") data={git:true,branch:"main",files:[{path:"app.js",index:" ",work:"M"}]};
          if(request.type==="workspace.diff") {
            if(window.staleDiff){window.delayedDiff=request;return;}
            data={path:"app.js",staged:"+staged",unstaged:"+<script>window.PWNED=true</script>"};
          }
          if(request.type==="sessions.children") data={sessions:[{id:"child",serviceId:"project",title:"下游任务",taskStatus:"running"}],nextCursor:0};
          if(request.type==="sessions.get") {
            if (window.failHistory) { queueMicrotask(()=>receiveEvent({id:request.id,type:"error",error:"模拟历史读取失败"})); return; }
            data={session:{id:request.sessionId,serviceId:"project"},entries:[{role:"user",text:"引用的原始要求"},{role:"assistant",text:"引用的结果"}],nextCursor:0};
          }
          if(request.type==="diagnostics.snapshot") data={version:"test-version",permissionMode:"full",activeRuns:1,providers:{codex:true}};
          queueMicrotask(()=>receiveEvent({id:request.id,type:"result",data}));
        };
        refreshServices({services:state.services,permissionMode:"full",runActivities:true,workspaceReview:true,workbenchDiagnostics:true});
        $("auth").hidden=true;$("app").hidden=false;$("prompt").value="保留我的草稿";
        renderChat();
      });
      assert.match(await page.locator(".workbench-status").textContent(), /执行中.*秒/);
      await page.evaluate(async()=>{window.failHistory=true;try{await rpc("sessions.get",{sessionId:"current"})}catch{}});
      assert.match(await page.locator(".workbench-sync").textContent(),/连接正常.*同步失败/);
      await page.evaluate(async()=>{window.failHistory=false;await rpc("sessions.get",{sessionId:"current"})});
      assert.match(await page.locator(".workbench-sync").textContent(),/历史已同步/);
      await page.evaluate(()=>refreshServices({services:state.services,permissionMode:"full"}));
      assert.equal(await page.getByRole("button",{name:"工作区变更",exact:true}).count(),0,"old clients hide unsupported review controls");
      await page.evaluate(()=>refreshServices({services:state.services,permissionMode:"full",runActivities:true,workspaceReview:true,workbenchDiagnostics:true}));
      const bounds = await page.locator("#chat-scroll").boundingBox();
      assert(bounds.height>100,"status bar must leave a usable message viewport");
      await page.locator(".run-details-toggle").click();
      assert.equal(await page.locator(".activity-row").count(),1);
      assert.match(await page.locator(".activity-state").textContent(),/进行中/);
      await page.locator(".quote-response").click();
      const quoted=await page.locator("#prompt").inputValue();
      assert(quoted.includes("保留我的草稿")&&quoted.includes("已经找到原因")&&quoted.includes("session=current"));
      assert.equal(await page.evaluate(()=>window.requests.some(request=>request.type==="tasks.run")),false,"quoting must not send a task");
      await page.getByRole("button",{name:"引用会话",exact:true}).click();
      await page.getByRole("button",{name:"可引用的会话",exact:true}).click();
      assert((await page.locator("#prompt").inputValue()).includes("引用的原始要求"));
      await page.getByRole("button",{name:"工作区变更",exact:true}).click();
      await page.getByRole("button",{name:/app.js/}).click();
      await page.waitForFunction(()=>[...document.querySelectorAll(".diff-content")].some(element=>element.textContent.includes("window.PWNED")));
      assert.equal(await page.evaluate(()=>window.PWNED),undefined,"diff content must remain inert text");
      assert.equal(await page.locator(".diff-content").count(),2);
      await page.screenshot({path:`.local/t3-workbench-20261109/diff-${width}.png`,fullPage:true});
      await page.getByRole("button",{name:"关闭",exact:true}).click();
      await page.getByRole("button",{name:"子任务",exact:true}).click();
      await page.getByText("下游任务",{exact:true}).waitFor();
      await page.getByRole("button",{name:"关闭",exact:true}).click();
      // A delayed review response must not appear after navigating to another session.
      await page.evaluate(()=>window.staleDiff=true);
      await page.getByRole("button",{name:"工作区变更",exact:true}).click();
      await page.getByRole("button",{name:/app.js/}).click();
      await page.waitForFunction(()=>!!window.delayedDiff);
      await page.evaluate(()=>{
        const request=window.delayedDiff;state.session={id:"other",serviceId:"project",title:"Other"};state.entries=[];$("prompt").value="另一个会话的草稿";renderChat();
        receiveEvent({id:request.id,type:"result",data:{path:"app.js",unstaged:"STALE DIFF"}});
      });
      assert.equal(await page.locator(".workbench-dialog").count(),0);
      assert.equal(await page.locator("#prompt").inputValue(),"另一个会话的草稿");
      await page.evaluate(()=>{state.historyLocal=true;state.peer=null;renderChat();});
      assert.match(await page.locator(".workbench-sync").textContent(),/离线/);
      assert.equal(await page.getByRole("button",{name:"工作区变更",exact:true}).isDisabled(),true);
      await page.getByRole("button",{name:"连接检查",exact:true}).click();
      assert(!(await page.locator(".diagnostics-content").textContent()).includes("另一个会话的草稿"));
      await page.getByRole("button",{name:"关闭",exact:true}).click();
      assert.deepEqual(errors,[]);
      await page.close();
    }
    console.log("PASS: desktop/mobile run lifecycle, draft-preserving citations, read-only escaped diff, children, stale response isolation and offline diagnostics");
  } finally { await browser.close(); }
})().catch(error=>{console.error(error);process.exitCode=1;});
