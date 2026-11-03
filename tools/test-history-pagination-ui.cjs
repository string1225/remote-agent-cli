const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const {chromium} = require(process.env.PLAYWRIGHT_MODULE || "playwright");

(async () => {
  const browser = await chromium.launch({headless:true, channel:process.env.PLAYWRIGHT_CHANNEL || "msedge"});
  try {
    for (const width of [1440, 390]) {
      const page = await browser.newPage({viewport:{width, height:900}}), errors = [];
      page.on("pageerror", error => errors.push(error.message));
      await page.addInitScript(() => { window.setInterval = () => 0; });
      await page.route("https://history.test/**", route => {
        const file = new URL(route.request().url()).pathname.slice(1) || "index.html";
        if (file.startsWith("api/")) return route.fulfill({status:401, json:{error:"please sign in"}});
        const filename = file === "app.js" && process.env.UI_TEST_APP_SOURCE ? process.env.UI_TEST_APP_SOURCE : path.join("web", file);
        return route.fulfill({contentType:file.endsWith(".css") ? "text/css" : file.endsWith(".js") ? "text/javascript" : "text/html", body:fs.readFileSync(filename)});
      });
      await page.goto("https://history.test/");
      await page.waitForFunction(() => document.getElementById("notice").textContent === "");
      await page.evaluate(async () => {
        state.user = {id:"history-user"};
        state.selected = {id:"client", name:"测试客户端"};
        state.services = [{id:"project", name:"历史分页", provider:"codex", workspace:"C:/test"}];
        state.serviceId = "project";
        state.peer = {ws:{readyState:WebSocket.OPEN}};
        state.permissionMode = "full";
        state.providers = {codex:{available:true}};
        $("auth").hidden = true;
        $("app").hidden = false;
        window.historyCalls = [];
        window.historyRows = Array.from({length:7}, (_, index) => [
          {seq:index * 3 + 1, turnId:String(index), role:"user", kind:"message", text:"第 " + (index + 1) + " 轮请求"},
          {seq:index * 3 + 2, turnId:String(index), role:"system", kind:"tool", text:"$ pwd"},
          {seq:index * 3 + 3, turnId:String(index), role:"assistant", kind:"stdout", text:("第 " + (index + 1) + " 轮回复\n\n").repeat(20)},
        ]).flat();
        rpc = async (type, fields) => {
          historyCalls.push({type, ...fields});
          const end = fields.before || 22, start = Math.max(1, end - 6);
          return {session:{id:"chat", serviceId:"project", title:"会话历史分页"}, entries:historyRows.filter(row => row.seq >= start && row.seq < end), olderCursor:start > 1 ? start : 0, historyEnd:end, nextCursor:0};
        };
        await loadSession({id:"chat", serviceId:"project", title:"会话历史分页"});
        $("prompt").value = "保留的草稿";
      });
      const more = page.locator("#load-more-history");
      assert.equal(await page.locator(".message.user").count(), 2);
      assert.equal(await page.evaluate(() => historyCalls.length), 1);
      await page.locator("#chat-scroll").evaluate(scroll => { scroll.scrollTop = 0; });
      const box = await more.boundingBox(), scrollBox = await page.locator("#chat-scroll").boundingBox();
      assert(Math.abs(box.x + box.width / 2 - (scrollBox.x + scrollBox.width / 2)) < 12, "load-more is centered");
      const anchor = page.locator('.message.user[data-key="5/user/message"]');
      const top = (await anchor.boundingBox()).y;
      await more.click();
      await page.waitForFunction(() => !state.loadingOlder && state.entries.length === 12);
      assert.equal(await page.locator(".message.user").count(), 4);
      assert(Math.abs((await anchor.boundingBox()).y - top) < 2, "prepending preserves the visible message");
      assert.equal(await page.locator("#prompt").inputValue(), "保留的草稿");
      for (let index = 0; index < 2; index++) {
        await page.locator("#chat-scroll").evaluate(scroll => { scroll.scrollTop = 0; });
        await more.click();
        await page.waitForFunction(() => !state.loadingOlder);
      }
      assert.equal(await page.locator(".message.user").count(), 7);
      assert.equal(await more.isVisible(), false);
      assert.deepEqual(await page.locator(".message.user .message-text").allTextContents(), Array.from({length:7}, (_, index) => "第 " + (index + 1) + " 轮请求"));
      assert.equal(await page.evaluate(() => historyCalls.length), 4);
      assert.equal(await page.evaluate(() => document.documentElement.scrollWidth), width);
      assert.deepEqual(errors, []);
      if (process.env.UI_TEST_SCREENSHOT_DIR) {
        fs.mkdirSync(process.env.UI_TEST_SCREENSHOT_DIR, {recursive:true});
        await page.screenshot({path:path.join(process.env.UI_TEST_SCREENSHOT_DIR, `history-${width}.png`)});
      }
      await page.close();
    }
    console.log("PASS: desktop/mobile two-turn history, centered load-more, stable scroll, complete ordering, preserved drafts and exhausted history");
  } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode = 1; });
