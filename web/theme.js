/* Synchronous head script: apply the saved theme before the first paint. */
(() => {
  const key = "remote-agent.theme";
  const normalize = value => value === "light" ? "light" : "dark";
  let current = "dark", dialog, message;
  function apply(value) {
    current = normalize(value);
    document.documentElement.dataset.theme = current;
    document.documentElement.style.colorScheme = current;
    document.querySelector('meta[name="theme-color"]')?.setAttribute("content", current === "dark" ? "#11151d" : "#f5f7fb");
    dialog?.querySelectorAll('input[name="theme"]').forEach(input => { input.checked = input.value === current; });
    window.dispatchEvent(new Event("remote-agent-theme"));
  }
  try { current = normalize(localStorage.getItem(key)); } catch {}
  apply(current);
  window.addEventListener("storage", event => { if (event.key === key || event.key === null) apply(event.newValue); });
  window.openAppearanceSettings = () => {
    if (!dialog) {
      dialog = document.createElement("dialog"); dialog.id = "appearance-dialog";
      dialog.className = "appearance-dialog"; dialog.setAttribute("aria-labelledby", "appearance-title");
      dialog.innerHTML = '<div class="dialog-head"><h2 id="appearance-title">界面设置</h2><button type="button" class="icon-button" aria-label="关闭界面设置">×</button></div><fieldset class="theme-options"><legend>主题</legend><label><input type="radio" name="theme" value="light" aria-label="明亮主题"><span><i aria-hidden="true">☀</i><strong>明亮主题</strong></span></label><label><input type="radio" name="theme" value="dark" aria-label="暗黑主题"><span><i aria-hidden="true">☾</i><strong>暗黑主题</strong></span></label></fieldset><p class="muted small">默认暗黑主题，选项保存在当前浏览器。</p><p class="theme-status" role="status"></p>';
      dialog.querySelector("button").onclick = () => dialog.close();
      message = dialog.querySelector(".theme-status");
      dialog.addEventListener("change", event => {
        if (!event.target.matches('input[name="theme"]')) return;
        apply(event.target.value);
        try { localStorage.setItem(key, current); message.textContent = "已保存"; }
        catch { message.textContent = "本页已切换；浏览器存储不可用，刷新后将使用默认主题。"; }
      });
      document.body.append(dialog);
    }
    apply(current); message.textContent = ""; dialog.showModal();
  };
})();
