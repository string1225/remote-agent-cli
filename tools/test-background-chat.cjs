const assert = require('node:assert/strict'), fs = require('node:fs'), vm = require('node:vm');
const source = fs.readFileSync('web/app.js','utf8');
class Element {
  constructor(tag='div', text='', cls='') { this.tag=tag;this.textContent=text;this.className=cls;this.dataset={};this.children=[];this.classList={toggle(){}};this.scrollHeight=500;this.scrollTop=400;this.clientHeight=100; }
  append(...nodes) { for(const node of nodes){if(node.parent)node.parent.children.splice(node.parent.children.indexOf(node),1);this.children.push(node);node.parent=this;} }
  querySelector(cls){return this.children.find(n=>n.className===cls.slice(1));}
  remove(){this.parent.children.splice(this.parent.children.indexOf(this),1);}
  replaceChildren(){this.children=[];}
}
const elements=new Map(), $=id=>{if(!elements.has(id))elements.set(id,new Element());return elements.get(id);};
const state={session:{id:'im-web'},entries:[],mode:'web',permissionMode:'full',peer:{ws:{readyState:1}},submitting:new Set()};
const sandbox={state,$,window:{},node:(...args)=>new Element(...args),service:()=>({provider:'codex'}),sessionTitle:()=>"Web",updateSessionContext(){},renderMarkdown:(el,text)=>{el.textContent=text;},Map,Set,console};
vm.createContext(sandbox);
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
const text=$('messages').children.map(el=>el.querySelector('.message-text').textContent).join('\n');
assert(text.includes('$ Get-Content SKILL.md'));assert(!text.includes('PRIVATE'));assert(!text.includes('raw diagnostic'));
vm.runInContext(source.slice(source.indexOf('function setRunning('),source.indexOf('function resetWorkspace(')),sandbox);
state.submitting.add('another-session');sandbox.setRunning(null);
assert.equal($('send-run').disabled,false,'another session must not block sending');
assert.equal($('new-session').disabled,false,'background turns must not block creating sessions');
state.session.taskStatus='running';state.session.taskId='task';sandbox.setRunning(null);
assert.equal($('send-run').disabled,true);assert.equal($('cancel-run').hidden,false);
assert(source.includes('await rpc("tasks.run"'));
console.log('PASS: IM first-entry backfill reorders existing DOM; user cards remain distinct; old skill output is hidden; only the viewed session blocks sending');
