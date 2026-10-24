const base='/openai/admin';
const $=id=>document.getElementById(id);
const message=(text,error=false)=>{$('message').textContent=text;$('message').classList.toggle('error',error);};
function loggedIn(value){$('login-panel').hidden=value;$('dashboard').hidden=!value;$('logout').hidden=!value;}
async function api(path,body){
 const options={credentials:'same-origin',headers:{'X-Sunny-Admin':'1'}};
 if(body!==undefined){options.method='POST';options.headers['Content-Type']='application/json';options.body=JSON.stringify(body);}
 const response=await fetch(base+'/api/'+path,options);
 const data=await response.json();
 if(!response.ok){if(response.status===401)loggedIn(false);throw new Error(data.error?.message||'操作未完成，请稍后重试。');}
 return data;
}
async function copy(text,done){try{await navigator.clipboard.writeText(text);message(done);}catch{message('复制未完成，请手动选中内容复制。',true);}}
async function refresh(){
 const data=await api('devices');loggedIn(true);
 const tbody=$('devices');tbody.replaceChildren();
 $('active-count').textContent=data.devices.filter(x=>x.enabled).length;
 $('revoked-count').textContent=data.devices.filter(x=>!x.enabled).length;
 $('empty').hidden=data.devices.length!==0;
 for(const device of data.devices){
  const row=document.createElement('tr');
  const name=document.createElement('td');name.textContent=device.name;name.className='device-name';
  const status=document.createElement('td');const badge=document.createElement('span');badge.className='badge'+(device.enabled?'':' revoked');badge.textContent=device.enabled?'可用':'已撤销';status.append(badge);
  const date=document.createElement('td');date.className='date';date.textContent=device.created_at?new Date(device.created_at*1000).toLocaleString('zh-CN',{hour12:false}):'—';
  const actions=document.createElement('td');const action=document.createElement('button');action.className='danger';action.textContent=device.enabled?'撤销 Token':'移除记录';
  action.addEventListener('click',async()=>{
   const prompt=device.enabled?`撤销「${device.name}」的 Token？这台设备将无法继续发起新请求。`:`移除「${device.name}」的已撤销记录？`;
   if(!confirm(prompt))return;action.disabled=true;
   try{await api('devices/'+(device.enabled?'revoke':'delete'),{name:device.name});await refresh();message(device.enabled?'Token 已撤销。':'设备记录已移除。');}catch(error){message(error.message,true);}finally{action.disabled=false;}
  });actions.append(action);row.append(name,status,date,actions);tbody.append(row);
 }
}
async function status(){try{const data=await api('status');const badge=$('upstream-status');badge.textContent=data.logged_in?'服务器已登录':'服务器需要重新登录';badge.className='status '+(data.logged_in?'ok':'error');}catch{$('upstream-status').textContent='暂时无法检查登录状态';$('upstream-status').className='status error';}}
$('login-form').addEventListener('submit',async event=>{event.preventDefault();const button=event.target.querySelector('button');button.disabled=true;try{await api('login',{password:$('password').value});$('password').value='';await refresh();status();message('');}catch(error){message(error.message,true);}finally{button.disabled=false;}});
$('create-form').addEventListener('submit',async event=>{event.preventDefault();const button=event.target.querySelector('button');button.disabled=true;try{const data=await api('devices',{name:$('device-name').value.trim()});$('new-device-label').textContent='设备：'+data.name;$('new-token').value=data.token;$('device-name').value='';$('token-dialog').showModal();await refresh();message('设备 Token 已创建。');}catch(error){message(error.message,true);}finally{button.disabled=false;}});
$('logout').addEventListener('click',async()=>{try{await api('logout',{});loggedIn(false);message('已退出管理。');}catch(error){message(error.message,true);}});
$('refresh').addEventListener('click',()=>refresh().then(()=>{status();message('设备列表已更新。');}).catch(error=>message(error.message,true)));
$('copy-endpoint').addEventListener('click',()=>copy($('endpoint').textContent,'接入地址已复制。'));
$('copy-config').addEventListener('click',()=>copy($('config').textContent,'Codex 配置已复制。'));
$('copy-token').addEventListener('click',()=>copy($('new-token').value,'设备 Token 已复制，请保存到对应电脑。'));
$('close-dialog').addEventListener('click',()=>$('token-dialog').close());
$('token-dialog').addEventListener('close',()=>{$('new-token').value='';$('new-device-label').textContent='';});
refresh().then(status).catch(()=>loggedIn(false));
