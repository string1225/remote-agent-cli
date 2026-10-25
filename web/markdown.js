/* Local GFM renderer. Agent output is untrusted; sanitize every render. */
(() => {
  const escape = value => String(value).replace(/[&<>"']/g, ch => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[ch]));
  const renderer = new marked.Renderer();
  renderer.html = ({text}) => escape(text);
  renderer.image = ({text}) => escape(text || "图片");
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
