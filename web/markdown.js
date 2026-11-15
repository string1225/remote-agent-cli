/* Local GFM renderer. Agent output is untrusted; sanitize every render. */
(() => {
  const escape = value => String(value).replace(/[&<>"']/g, ch => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[ch]));
  const renderer = new marked.Renderer();
  renderer.html = ({text}) => escape(text);
  renderer.image = ({text}) => escape(text || "图片");
  function localPath(value) {
    let path = String(value || "").trim();
    if (/^file:\/\//i.test(path)) {
      try {
        const url = new URL(path);
        path = decodeURIComponent(url.pathname);
        if (url.hostname && url.hostname !== "localhost") path = "\\\\" + url.hostname + path.replace(/\//g, "\\");
        else if (/^\/[a-z]:\//i.test(path)) path = path.slice(1);
      } catch { return ""; }
    } else if (!/^[a-z]:[\\/]/i.test(path) && !/^\\\\[^\\]+\\[^\\]+/.test(path)) return "";
    if (/[\u0000-\u001f\u007f]/.test(path)) return "";
    return /^[a-z]:[\\/]/i.test(path) ? path.replace(/\//g, "\\") : path;
  }
  // Capture file destinations before sanitization removes their href. Never
  // navigate to a file/drive URL or run a command from assistant output.
  const webLink = renderer.link;
  renderer.link = function(token) {
    const path = localPath(token.href);
    if (!path) return webLink.call(this, token);
    return `<button type="button" class="local-file-link" data-local-path="${escape(path)}" aria-haspopup="dialog" title="查看客户端文件路径">${this.parser.parseInline(token.tokens)}</button>`;
  };
  let fileDialog, filePath, fileHint, fileStatus, fileCopy, fileTrigger, fileRevision = 0;
  function showLocalFile(path, trigger) {
    if (!fileDialog) {
      fileDialog = document.createElement("dialog");
      fileDialog.id = "local-file-dialog";
      fileDialog.setAttribute("aria-labelledby", "local-file-title");
      fileDialog.setAttribute("aria-describedby", "local-file-description");
      fileDialog.innerHTML = '<div class="dialog-head"><h2 id="local-file-title">客户端文件</h2><button type="button" class="icon-button" aria-label="关闭文件路径">×</button></div><p id="local-file-description">文件保存在生成它的客户端电脑上。浏览器无法直接打开，请在文件所在的电脑上使用以下路径。</p><label>文件路径<textarea readonly rows="3" spellcheck="false"></textarea></label><p class="small muted local-file-hint"></p><button type="button" class="primary local-file-copy">复制路径</button><p class="small" role="status" aria-live="polite"></p>';
      filePath = fileDialog.querySelector("textarea");
      fileHint = fileDialog.querySelector(".local-file-hint");
      fileStatus = fileDialog.querySelector("[role=status]");
      fileCopy = fileDialog.querySelector(".local-file-copy");
      fileDialog.querySelector(".icon-button").onclick = () => fileDialog.close();
      fileDialog.addEventListener("close", () => { fileRevision++; if (fileTrigger?.isConnected) fileTrigger.focus(); });
      filePath.onclick = () => filePath.select();
      fileCopy.onclick = async () => {
        const revision = fileRevision, value = filePath.value;
        fileCopy.disabled = true;
        try {
          await navigator.clipboard.writeText(value);
          if (revision === fileRevision) fileStatus.textContent = "路径已复制，请在文件所在的客户端电脑上打开。";
        } catch {
          if (revision === fileRevision) {
            fileStatus.textContent = "无法访问剪贴板，请手动复制上方已选中的路径。";
            filePath.focus(); filePath.select();
          }
        } finally { if (revision === fileRevision) fileCopy.disabled = false; }
      };
      document.body.append(fileDialog);
    }
    fileRevision++; fileTrigger = trigger;
    filePath.value = path; fileStatus.textContent = ""; fileCopy.disabled = false;
    fileHint.textContent = /^[a-z]:[\\/]|^\\\\/i.test(path)
      ? "在该客户端的 Windows 资源管理器地址栏粘贴路径，然后按 Enter。"
      : "在该客户端的文件管理器中使用此路径，打开对应文件或目录。";
    fileDialog.showModal(); fileCopy.focus();
  }
  window.renderMarkdown = (element, source) => {
    if (element.renderedSource === source) return;
    element.renderedSource = source;
    element.classList.add("markdown-body");
    element.innerHTML = DOMPurify.sanitize(marked.parse(source, {renderer, gfm:true, breaks:true}), {
      USE_PROFILES:{html:true}, FORBID_TAGS:["img","video","audio","source","iframe","input","form","style"], FORBID_ATTR:["style"],
    });
    for (const link of element.querySelectorAll("a")) {
      const href = link.getAttribute("href");
      if (!href) continue;
      try {
        const url = new URL(href, location.href);
        if (!["http:","https:","mailto:"].includes(url.protocol)) { link.removeAttribute("href"); continue; }
        link.target = "_blank"; link.rel = "noopener noreferrer";
      } catch { link.removeAttribute("href"); }
    }
    // A later reply may give the same path as inline code instead of a link.
    for (const code of element.querySelectorAll("code")) {
      if (code.closest("pre,a,button")) continue;
      const path = localPath(code.textContent);
      if (!path) continue;
      const button = document.createElement("button");
      button.type = "button"; button.className = "local-file-link";
      button.dataset.localPath = path; button.setAttribute("aria-haspopup", "dialog");
      button.title = "查看客户端文件路径"; code.replaceWith(button); button.append(code);
    }
    for (const button of element.querySelectorAll("button.local-file-link")) {
      const path = button.dataset.localPath;
      button.onclick = () => showLocalFile(path, button);
    }
    for (const table of element.querySelectorAll("table")) {
      const wrap = document.createElement("div"); wrap.className = "markdown-table"; table.replaceWith(wrap); wrap.append(table);
    }
    for (const pre of element.querySelectorAll("pre")) {
      const button = document.createElement("button"); button.type = "button"; button.className = "code-copy"; button.textContent = "复制代码";
      button.onclick = async () => {
        try { await navigator.clipboard.writeText(pre.querySelector("code")?.textContent || ""); button.textContent = "已复制"; }
        catch { button.textContent = "请选中代码复制"; }
      };
      pre.append(button);
    }
  };
})();
