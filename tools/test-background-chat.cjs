const assert = require('node:assert/strict'), fs = require('node:fs'), vm = require('node:vm');
const source = fs.readFileSync(process.env.UI_TEST_APP_SOURCE || 'web/app.js','utf8');
class Element {
  constructor(tag='div', text='', cls='') { this.tag=tag;this.textContent=text;this.className=cls;this.dataset={};this.children=[];this.classList={toggle(){}};this.scrollHeight=500;this.scrollTop=400;this.clientHeight=100; }
  append(...nodes) { for(const node of nodes){if(node.parent)node.parent.children.splice(node.parent.children.indexOf(node),1);this.children.push(node);node.parent=this;} }
  querySelector(cls){for(const child of this.children){if(child.className===cls.slice(1))return child;const found=child.querySelector(cls);if(found)return found;}}
  setAttribute(name,value){this[name]=value;}
  addEventListener(){}
  remove(){this.parent.children.splice(this.parent.children.indexOf(this),1);}
  replaceChildren(){this.children=[];}
}
const elements=new Map(), $=id=>{if(!elements.has(id))elements.set(id,new Element());return elements.get(id);};
const state={session:{id:'im-web'},entries:[],mode:'web',permissionMode:'full',peer:{ws:{readyState:1}},submitting:new Set()};
const sandbox={state,$,window:{},node:(...args)=>new Element(...args),service:()=>({provider:'codex'}),sessionTitle:()=>"Web",updateSessionContext(){},renderMarkdown:(el,text)=>{el.textContent=text;},ResizeObserver:class{observe(){}},Map,Set,console};
vm.createContext(sandbox);
vm.runInContext(source.slice(source.indexOf('function setRunning('),source.indexOf('function resetWorkspace(')),sandbox);
vm.runInContext(source.slice(source.indexOf('function scrollChatToLatest('),source.indexOf('function sessionTitle(')),sandbox);
vm.runInContext(source.slice(source.indexOf('function toolInvocation('),source.indexOf('function initPaneResizers(')),sandbox);
state.entries=[{seq:2,turnId:'im-task',role:'assistant',kind:'stdout',text:'answer'}];sandbox.renderChat();
const assistant=$('messages').children[0];
state.entries.unshift({seq:1,turnId:'im-task',role:'user',kind:'message',text:'first IM-created prompt'});
sandbox.renderChat();
assert.equal($('messages').children[0].className,'message user');
assert.equal($('messages').children[1],assistant,'existing DOM answer should move after the backfilled user entry');
state.entries.push({seq:3,turnId:'im-task',role:'system',kind:'tool',text:'$ Get-Content SKILL.md\nPRIVATE SKILL CONTENT'}, {seq:4,turnId:'im-task',role:'system',kind:'tool',text:'CONTINUED PRIVATE SKILL CONTENT'}, {seq:5,turnId:'im-task',role:'system',kind:'stderr',text:'raw diagnostic'}, {seq:6,turnId:'next',role:'user',kind:'message',text:'follow-up'});
sandbox.renderChat();
assert.equal($('messages').children.at(-1).className,'message user');
const text=$('messages').children.map(el=>el.querySelector('.message-text').textContent + (el.querySelector('.run-details')?.textContent || '')).join('\n');
assert(text.includes('$ Get-Content SKILL.md'));assert(!text.includes('PRIVATE'));assert(!text.includes('raw diagnostic'));
assert.equal($('messages').children.length,3,'tool calls belong to the assistant reply');
const toggle=assistant.querySelector('.run-details-toggle'),details=assistant.querySelector('.run-details');
assert.equal(toggle.hidden,false);assert.equal(details.hidden,true);assert.equal(toggle['aria-expanded'],'false');
assert.equal(toggle['aria-controls'],details.id);
assert.equal(assistant.children[0].children[0],toggle,'toggle sits beside the assistant label');
assert.equal(assistant.children[1],details,'details sit between label and reply');
toggle.onclick();assert.equal(details.hidden,false);assert.equal(toggle['aria-expanded'],'true');
sandbox.renderChat();assert.equal(details.hidden,false,'poll renders preserve expansion');
toggle.onclick();assert.equal(details.hidden,true);
state.entries.push({seq:7,turnId:'next',role:'system',kind:'tool',text:'$ pwd\nPRIVATE OUTPUT'});
sandbox.renderChat();
const nextAssistant=$('messages').children.at(-1);
assert.equal(nextAssistant.className,'message assistant','tools before an answer still have a toggle');
state.entries.push({seq:8,turnId:'next',role:'assistant',kind:'stdout',text:'second answer'});
sandbox.renderChat();
assert.equal($('messages').children.at(-1),nextAssistant,'answer reuses the tool placeholder');
assert.equal(nextAssistant.querySelector('.run-details').textContent,'$ pwd');
assert.equal(details.textContent,'$ Get-Content SKILL.md','different turns keep separate tool records');
state.submitting.add('another-session');sandbox.setRunning(null);
assert.equal($('send-run').disabled,false,'another session must not block sending');
assert.equal($('new-session').disabled,false,'background turns must not block creating sessions');
state.session.taskStatus='running';state.session.taskId='task';sandbox.setRunning(null);
assert.equal($('send-run').disabled,true);assert.equal($('cancel-run').hidden,false);
state.session.taskStatus='done';sandbox.renderChat();
assert.equal($('send-run').disabled,false,'rendering a completed reply refreshes the send button');
assert.equal($('cancel-run').hidden,true);
assert(source.includes('await rpc("tasks.run"'));
console.log('PASS: IM first-entry backfill reorders existing DOM; user cards remain distinct; old skill output is hidden; only the viewed session blocks sending');
