const assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm'),test=require('node:test');
const source=fs.readFileSync((process.env.PROJECT_SOURCE_ROOT||'.')+'/web/app.js','utf8');
function harness(pages){
 const elements=new Map(),cache=new Map(),requests=[];
 const $=id=>{if(!elements.has(id))elements.set(id,{value:'',scrollHeight:100,scrollTop:0});return elements.get(id)};
 const key=(kind,scope)=>JSON.stringify([kind,scope]);
 const state={user:{id:'owner'},selected:{id:'client'},peer:{},services:[{id:'p',name:'Project',provider:'codex'}],serviceId:'old',sessions:[],entries:[],readRevision:0,listRevision:0,mode:'web'};
 const context={state,$,WorkspaceCache:{read:(kind,scope)=>cache.get(key(kind,scope)),write:(kind,scope,value)=>cache.set(key(kind,scope),value)},
  closeSayso:null,window:{chatAttachments:{restore(){}}},saveCurrentDraft(){},cacheCurrentSession(){},setRunning(){},renderChat(){},renderSessions(){},renderDevices(){},
  rpc:async(type,fields)=>{requests.push({type,...fields});return pages(fields)},
  loadSession:async(session,options={})=>{if(options.refreshOnly)return;state.session=session;state.readRevision++;if(!options.automatic)state.preferLatestWeb=false;},
  service:()=>state.services.find(s=>s.id===state.serviceId),Promise,Map};
 vm.createContext(context);vm.runInContext(source.slice(source.indexOf('function latestWebSession('),source.indexOf('let sessionSignature;')),context);
 return {...context,cache,requests,key};
}
const im={id:'im',kind:'im',serviceId:'p',updatedAt:1000},old={id:'old-web',kind:'web',serviceId:'p',updatedAt:10},latest={id:'latest-web',kind:'web',serviceId:'p',updatedAt:30};
test('project clicks choose newest active Web session ahead of remembered IM and archived Web rows',async()=>{
 const h=harness(()=>({sessions:[],nextCursor:0}));
 h.cache.set(h.key('sessions',['owner','client','p']),[im,old,{...latest,id:'archived',archivedAt:1,updatedAt:2000},latest]);
 h.cache.set(h.key('project-selection',['owner','client','p']),{session:'im'});
 await h.selectProject('p',{localOnly:true});assert.equal(h.state.session.id,'latest-web');assert.equal(h.requests.length,0);
});
test('IM-first pagination finds the latest Web session on the next page',async()=>{
 const h=harness(fields=>fields.cursor?{sessions:[latest,old],nextCursor:0}:{sessions:[im],nextCursor:50});
 await h.selectProject('p');assert.equal(h.state.session.id,'latest-web');assert.deepEqual(h.requests.map(r=>r.cursor),[0,50]);
});
test('fresh metadata replaces an older cached Web selection with the most recently updated Web session',async()=>{
 const h=harness(()=>({sessions:[im,latest,old],nextCursor:0}));
 h.cache.set(h.key('sessions',['owner','client','p']),[im,old]);
 await h.selectProject('p');assert.equal(h.state.session.id,'latest-web');
});
test('an explicit IM selection made while the list loads cannot be replaced by automatic Web selection',async()=>{
 let release;const pending=new Promise(resolve=>release=resolve),h=harness(()=>pending);
 const opening=h.selectProject('p');await h.loadSession(im);release({sessions:[im,latest],nextCursor:0});await opening;
 assert.equal(h.state.session.id,'im');
});
test('a project with only IM sessions does not automatically open an IM conversation',async()=>{
 const h=harness(()=>({sessions:[im],nextCursor:0}));await h.selectProject('p');assert.equal(h.state.session,null);
});
test('IM titles use channel and peer display names and hide raw IDs',()=>{
 const context={service:()=>({channel:{name:'研发助理'}})};vm.createContext(context);
 vm.runInContext(source.slice(source.indexOf('function sessionTitle('),source.indexOf('function mergeEntries(')),context);
 assert.equal(context.sessionTitle({...im,channelPeerName:'小林'}),'研发助理 · 小林');
 assert.equal(context.sessionTitle({...im,channelPeerName:'oc_fixture',channelChatId:'oc_fixture',channelChatType:'user'}),'研发助理 · 私聊');
 assert.equal(context.sessionTitle({...im,channelPeerName:'ou_fixture',channelChatType:'group'}),'研发助理 · 群聊');
 assert.equal(context.sessionTitle({...im,customTitle:true,title:'我起的名字'}),'我起的名字');
});
