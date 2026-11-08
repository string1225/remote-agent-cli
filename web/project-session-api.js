/* Project credentials live on the server; never store them in browser caches. */
let revision = 0, section;

function clear() {
  revision++;
  if (!section) return;
  for (const input of section.querySelectorAll('input, textarea')) input.value = '';
  section.querySelector('[data-token]').type = 'password';
  section.querySelector('[data-reveal]').textContent = '查看 Token';
  section.querySelector('[data-status]').textContent = '';
  for (const button of section.querySelectorAll('button')) button.disabled = true;
}

function build() {
  if (section) return;
  section = document.createElement('details');
  section.className = 'project-session-api';
  section.innerHTML = `<summary>外部系统 API</summary>
    <p class="small muted">通过项目专用 Token 创建会话，可附带消息并选择是否立即执行。目标客户端需在线且已升级。</p>
    <p class="small" data-status role="status"></p>
    <label>创建会话接口<input data-endpoint readonly aria-label="创建会话接口"></label>
    <button type="button" data-copy-url>复制接口地址</button>
    <label>项目 Token<input data-token type="password" readonly autocomplete="off" aria-label="项目 Token"></label>
    <div class="project-api-actions"><button type="button" data-generate>生成 Token</button><button type="button" data-reveal>查看 Token</button><button type="button" data-copy>复制 Token</button><button type="button" data-revoke>撤销 Token</button></div>
    <form data-custom><label>替换为自定义 Token<input name="token" type="password" required minlength="32" maxlength="256" pattern="[A-Za-z0-9_-]{32,256}" autocomplete="new-password" placeholder="32–256 位字母、数字、- 或 _"></label><button type="submit">保存自定义 Token</button></form>
    <label>调用示例<textarea data-example readonly rows="6" aria-label="调用示例"></textarea></label>
    <p class="small muted">execute 默认 false：消息保存为已暂停，可在工作台继续或编辑；true：创建后立即后台执行。重置或撤销后旧 Token 立即失效。</p>`;
  const dialog = document.getElementById('project-settings-dialog');
  for (const button of section.querySelectorAll('button')) button.classList.add('secondary');
  dialog.append(section);
  dialog.addEventListener('close', clear);
  window.addEventListener('remote-agent-disconnected', clear);
}

export async function open(agentID, projectID, valid) {
  build(); clear();
  const generation = revision, current = () => generation === revision && valid();
  const path = `/api/agents/${encodeURIComponent(agentID)}/projects/${encodeURIComponent(projectID)}`;
  const apiPath = path + '/session-token';
  const endpoint = location.origin + new URL('.', location.href).pathname.replace(/\/$/, '') + path + '/sessions';
  const token = section.querySelector('[data-token]'), status = section.querySelector('[data-status]');
  let configured = false, busy = true;
  const buttons = [...section.querySelectorAll('button')];
  function render() {
    if (!current()) return;
    for (const button of buttons) button.disabled = busy;
    for (const name of ['reveal', 'copy', 'revoke']) section.querySelector(`[data-${name}]`).disabled = busy || !configured;
    section.querySelector('[data-generate]').textContent = configured ? '重置 Token' : '生成 Token';
  }
  async function request(method, suffix = '', body) {
    const response = await fetch(new URL('.', location.href).pathname.replace(/\/$/, '') + apiPath + suffix, {
      method, credentials:'same-origin', headers:{'Content-Type':'application/json'}, body:body === undefined ? undefined : JSON.stringify(body), cache:'no-store',
    });
    const data = await response.json();
    if (!response.ok) throw Error(data.error || '项目 Token 操作失败');
    return data;
  }
  async function act(action) {
    if (!current() || busy) return;
    busy = true; render();
    try { await action(); }
    catch (error) { if (current()) status.textContent = error.message; }
    finally { if (current()) { busy = false; render(); } }
  }
  async function copy(value, field, text) {
    if (!current()) return;
    try { await navigator.clipboard.writeText(value); if (current()) status.textContent = text; }
    catch { if (current()) { field.focus(); field.select(); status.textContent = '请按 Ctrl / ⌘ + C 复制'; } }
  }
  async function secret() {
    if (token.value) return token.value;
    const data = await request('POST', '/reveal', {});
    if (!current()) return '';
    token.value = data.token;
    return data.token;
  }
  section.querySelector('[data-endpoint]').value = endpoint;
  section.querySelector('[data-example]').value = [
    `curl -X POST '${endpoint}'`,
    "  -H 'Authorization: Bearer YOUR_PROJECT_TOKEN'",
    "  -H 'Content-Type: application/json'",
    `  --data '{"title":"外部系统任务","message":"请处理这条消息","execute":true}'`,
  ].join(' \\\n');
  section.querySelector('[data-copy-url]').onclick = () => act(() => copy(endpoint, section.querySelector('[data-endpoint]'), '接口地址已复制'));
  section.querySelector('[data-generate]').onclick = () => act(async () => {
    if (configured && !confirm('重置后，其他系统当前使用的项目 Token 将立即失效。继续重置？')) return;
    const data = await request('PUT', '', {});
    if (current()) { configured = true; token.value = data.token; status.textContent = 'Token 已生成，可复制到其他系统'; }
  });
  section.querySelector('[data-reveal]').onclick = () => act(async () => {
    if (token.type === 'text') { token.type = 'password'; section.querySelector('[data-reveal]').textContent = '查看 Token'; return; }
    await secret();
    if (current()) { token.type = 'text'; section.querySelector('[data-reveal]').textContent = '隐藏 Token'; }
  });
  section.querySelector('[data-copy]').onclick = () => act(async () => { const value = await secret(); if (current()) await copy(value, token, 'Token 已复制'); });
  section.querySelector('[data-revoke]').onclick = () => act(async () => {
    if (!confirm('撤销后，其他系统将无法再用此 Token 创建会话。继续撤销？')) return;
    await request('DELETE');
    if (current()) { configured = false; token.value = ''; token.type = 'password'; section.querySelector('[data-reveal]').textContent = '查看 Token'; status.textContent = 'Token 已撤销'; }
  });
  section.querySelector('[data-custom]').onsubmit = event => {
    event.preventDefault();
    const value = event.currentTarget.elements.token.value;
    act(async () => {
      if (configured && !confirm('替换后旧 Token 将立即失效。继续替换？')) return;
      const data = await request('PUT', '', {token:value});
      if (current()) { configured = true; token.value = data.token; section.querySelector('[name=token]').value = ''; status.textContent = '项目 Token 已保存'; }
    });
  };
  status.textContent = '正在读取项目 Token…'; render();
  try {
    const data = await request('GET');
    if (current()) { configured = data.configured; status.textContent = configured ? '已配置 · 仅允许在本项目创建会话' : '尚未配置 Token，外部接口未启用'; }
  } catch (error) { if (current()) status.textContent = error.message; }
  finally { if (current()) { busy = false; render(); } }
}
