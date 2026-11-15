package cn.sunnystring.remoteagent;

import android.Manifest;
import android.app.Activity;
import android.app.AlertDialog;
import android.content.Intent;
import android.content.SharedPreferences;
import android.content.pm.PackageManager;
import android.graphics.Color;
import android.net.Uri;
import android.os.Bundle;
import android.view.View;
import android.view.Gravity;
import android.webkit.CookieManager;
import android.webkit.PermissionRequest;
import android.webkit.ValueCallback;
import android.webkit.WebChromeClient;
import android.webkit.WebResourceError;
import android.webkit.WebResourceRequest;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.widget.Button;
import android.widget.EditText;
import android.widget.FrameLayout;
import android.widget.LinearLayout;
import android.widget.ProgressBar;
import android.widget.TextView;
import android.widget.Toast;
import java.net.URI;
import org.json.JSONObject;

public final class MainActivity extends Activity {
    static final String DEFAULT_SITE = "https://codex.sunny-string.cn/agents/";
    private static final int FILE_REQUEST = 11, AUDIO_REQUEST = 12;
    private WebView web;
    private ProgressBar progress;
    private SharedPreferences preferences;
    private ValueCallback<Uri[]> fileCallback;
    private PermissionRequest microphone;
    private AppUpdater updater;
    private String site;
    private LinearLayout layout, connectionError;
    private boolean dark = true, updateBusy;
    private String updateStatus = "启动时自动检查更新";

    @Override public void onCreate(Bundle state) {
        super.onCreate(state);
        preferences = getSharedPreferences("remote-agent", MODE_PRIVATE);
        site = preferences.getString("site", DEFAULT_SITE);
        layout = new LinearLayout(this); layout.setOrientation(LinearLayout.VERTICAL);
        // Account for status/navigation bars and the keyboard on target SDK 36.
        layout.setOnApplyWindowInsetsListener((v, insets) -> {
            if (android.os.Build.VERSION.SDK_INT >= 30) {
                android.graphics.Insets bars = insets.getInsets(android.view.WindowInsets.Type.systemBars() | android.view.WindowInsets.Type.ime());
                v.setPadding(bars.left, bars.top, bars.right, bars.bottom);
            } else { v.setPadding(insets.getSystemWindowInsetLeft(), insets.getSystemWindowInsetTop(), insets.getSystemWindowInsetRight(), insets.getSystemWindowInsetBottom()); }
            return insets;
        });
        progress = new ProgressBar(this, null, android.R.attr.progressBarStyleHorizontal); progress.setMax(100);
        progress.setVisibility(View.GONE);
        web = new WebView(this);
        WebSettings config = web.getSettings(); config.setJavaScriptEnabled(true); config.setDomStorageEnabled(true);
        config.setAllowFileAccess(false); config.setAllowContentAccess(false);
        config.setMixedContentMode(WebSettings.MIXED_CONTENT_NEVER_ALLOW);
        config.setMediaPlaybackRequiresUserGesture(true);
        config.setUserAgentString(config.getUserAgentString() + " RemoteAgentAndroid/" + BuildConfig.VERSION_NAME + " RemoteAgentNative/1");
        CookieManager.getInstance().setAcceptCookie(true); CookieManager.getInstance().setAcceptThirdPartyCookies(web, false);
        WebView.setWebContentsDebuggingEnabled(BuildConfig.DEBUG);
        web.setWebViewClient(new WebViewClient() {
            @Override public boolean shouldOverrideUrlLoading(WebView view, WebResourceRequest request) {
                Uri uri = request.getUrl();
                if ("remoteagent".equals(uri.getScheme())) {
                    if (request.isForMainFrame() && trusted(view.getUrl())) handleAppAction(uri);
                    return true;
                }
                if (trusted(uri.toString())) return false;
                if (request.isForMainFrame() && "https".equals(uri.getScheme())) {
                    try { startActivity(new Intent(Intent.ACTION_VIEW, uri)); } catch (Exception e) { toast("没有可打开此链接的应用"); }
                }
                return true;
            }
            @Override public void onReceivedError(WebView view, WebResourceRequest request, WebResourceError error) {
                if (request.isForMainFrame()) showConnectionError();
            }
            @Override public void onReceivedHttpError(WebView view, WebResourceRequest request, android.webkit.WebResourceResponse response) {
                if (request.isForMainFrame() && response.getStatusCode() >= 400) showConnectionError();
            }
            @Override public void onPageStarted(WebView view, String url, android.graphics.Bitmap icon) {
                connectionError.setVisibility(View.GONE); web.setVisibility(View.VISIBLE);
            }
            @Override public void onPageFinished(WebView view, String url) {
                CookieManager.getInstance().flush(); publishAppState();
                if (trusted(url) && web.getVisibility() == View.VISIBLE) view.evaluateJavascript("!!document.documentElement && document.documentElement.dataset.theme === 'light'", value -> applyTheme(!"true".equals(value)));
            }
        });
        web.setWebChromeClient(new WebChromeClient() {
            @Override public void onProgressChanged(WebView view, int value) { progress.setProgress(value); progress.setVisibility(value == 100 ? View.GONE : View.VISIBLE); }
            @Override public boolean onShowFileChooser(WebView view, ValueCallback<Uri[]> callback, FileChooserParams params) {
                if (!trusted(view.getUrl())) return false;
                if (fileCallback != null) fileCallback.onReceiveValue(null);
                fileCallback = callback;
                Intent intent = new Intent(Intent.ACTION_OPEN_DOCUMENT); intent.setType("*/*"); intent.addCategory(Intent.CATEGORY_OPENABLE);
                intent.putExtra(Intent.EXTRA_ALLOW_MULTIPLE, true);
                try { startActivityForResult(intent, FILE_REQUEST); } catch (Exception e) { fileCallback.onReceiveValue(null); fileCallback = null; toast("无法打开文件选择器"); }
                return true;
            }
            @Override public void onPermissionRequest(PermissionRequest request) {
                runOnUiThread(() -> {
                    if (!trusted(web.getUrl()) || !trusted(request.getOrigin().toString())) { request.deny(); return; }
                    boolean audio = false;
                    for (String resource : request.getResources()) if (PermissionRequest.RESOURCE_AUDIO_CAPTURE.equals(resource)) audio = true;
                    if (!audio) { request.deny(); return; }
                    if (microphone != null) microphone.deny();
                    microphone = request;
                    if (checkSelfPermission(Manifest.permission.RECORD_AUDIO) == PackageManager.PERMISSION_GRANTED) grantMicrophone();
                    else requestPermissions(new String[]{Manifest.permission.RECORD_AUDIO}, AUDIO_REQUEST);
                });
            }
            @Override public void onPermissionRequestCanceled(PermissionRequest request) { if (microphone == request) microphone = null; }
        });
        FrameLayout content = new FrameLayout(this);
        content.addView(web, new FrameLayout.LayoutParams(-1, -1));
        connectionError = new LinearLayout(this); connectionError.setOrientation(LinearLayout.VERTICAL); connectionError.setGravity(Gravity.CENTER);
        int space = Math.round(24 * getResources().getDisplayMetrics().density); connectionError.setPadding(space, space, space, space);
        TextView message = new TextView(this); message.setText("暂时无法连接工作台\n检查网络后重试，或修改手机的站点设置。"); message.setTextSize(16); message.setGravity(Gravity.CENTER);
        Button retry = new Button(this); retry.setText("重新连接"); retry.setOnClickListener(v -> web.loadUrl(site));
        Button settings = new Button(this); settings.setText("手机设置"); settings.setOnClickListener(v -> showSettings());
        connectionError.addView(message); connectionError.addView(retry); connectionError.addView(settings);
        connectionError.setVisibility(View.GONE); content.addView(connectionError, new FrameLayout.LayoutParams(-1, -1));
        content.addView(progress, new FrameLayout.LayoutParams(-1, Math.max(2, Math.round(2 * getResources().getDisplayMetrics().density)), Gravity.TOP));
        layout.addView(content, new LinearLayout.LayoutParams(-1, -1));
        setContentView(layout); applyTheme(preferences.getBoolean("darkTheme", true)); updater = new AppUpdater(this);
        if (android.os.Build.VERSION.SDK_INT >= 33) getOnBackInvokedDispatcher().registerOnBackInvokedCallback(android.window.OnBackInvokedDispatcher.PRIORITY_DEFAULT, this::goBack);
        if (state == null || web.restoreState(state) == null) web.loadUrl(site);
        if (preferences.getBoolean("autoUpdate", true)) updater.check(false);
    }

    private void handleAppAction(Uri uri) {
        String action = uri.getHost();
        if ("settings".equals(action)) showSettings(); // Older web deployments remain usable.
        else if ("state".equals(action)) publishAppState();
        else if ("update".equals(action)) updater.check(true);
        else if ("auto-update".equals(action)) {
            String enabled = uri.getQueryParameter("enabled");
            if (!"true".equals(enabled) && !"false".equals(enabled)) return;
            preferences.edit().putBoolean("autoUpdate", Boolean.parseBoolean(enabled)).apply(); publishAppState();
        } else if ("site".equals(action)) changeSite(uri.getQueryParameter("url"));
        else if ("reload".equals(action)) web.reload();
        else if ("theme".equals(action)) {
            String value = uri.getQueryParameter("value");
            if ("light".equals(value) || "dark".equals(value)) applyTheme("dark".equals(value));
        }
    }
    private void publishAppState() {
        if (isFinishing() || isDestroyed() || !trusted(web.getUrl())) return;
        try {
            JSONObject state = new JSONObject().put("version", BuildConfig.VERSION_NAME).put("site", site)
                .put("autoUpdate", preferences.getBoolean("autoUpdate", true)).put("updateBusy", updateBusy).put("updateStatus", updateStatus);
            web.evaluateJavascript("window.dispatchEvent(new CustomEvent('remote-agent-android-state',{detail:" + state + "}))", null);
        } catch (org.json.JSONException ignored) { /* Values above are primitive and never non-finite. */ }
    }
    void updateStatus(String message, boolean busy) {
        runOnUiThread(() -> { updateStatus = message; updateBusy = busy; publishAppState(); });
    }
    void updateFinished() { runOnUiThread(() -> { updateBusy = false; publishAppState(); }); }
    @SuppressWarnings("deprecation")
    private void applyTheme(boolean useDark) {
        dark = useDark;
        preferences.edit().putBoolean("darkTheme", dark).apply();
        int background = Color.parseColor(dark ? "#11151d" : "#f5f7fb");
        layout.setBackgroundColor(background); web.setBackgroundColor(background); connectionError.setBackgroundColor(background);
        ((TextView) connectionError.getChildAt(0)).setTextColor(Color.parseColor(dark ? "#dfe5ef" : "#26382d"));
        getWindow().setStatusBarColor(background); getWindow().setNavigationBarColor(background);
        if (android.os.Build.VERSION.SDK_INT >= 30) {
            int light = android.view.WindowInsetsController.APPEARANCE_LIGHT_STATUS_BARS | android.view.WindowInsetsController.APPEARANCE_LIGHT_NAVIGATION_BARS;
            getWindow().getInsetsController().setSystemBarsAppearance(dark ? 0 : light, light);
        } else {
            int light = View.SYSTEM_UI_FLAG_LIGHT_STATUS_BAR | View.SYSTEM_UI_FLAG_LIGHT_NAVIGATION_BAR;
            getWindow().getDecorView().setSystemUiVisibility(dark ? 0 : light);
        }
    }
    private void showConnectionError() { progress.setVisibility(View.GONE); web.setVisibility(View.GONE); connectionError.setVisibility(View.VISIBLE); }
    AlertDialog.Builder alert() {
        return new AlertDialog.Builder(this, dark ? android.R.style.Theme_Material_Dialog_Alert : android.R.style.Theme_Material_Light_Dialog_Alert);
    }

    private boolean trusted(String url) {
        try { URI candidate = new URI(url), configured = new URI(site);
            return "https".equals(candidate.getScheme()) && candidate.getHost() != null && candidate.getHost().equalsIgnoreCase(configured.getHost()) && effectivePort(candidate) == effectivePort(configured);
        } catch (Exception e) { return false; }
    }
    private static int effectivePort(URI uri) { return uri.getPort() < 0 ? 443 : uri.getPort(); }
    private void grantMicrophone() {
        if (microphone != null) { if (trusted(web.getUrl()) && trusted(microphone.getOrigin().toString())) microphone.grant(new String[]{PermissionRequest.RESOURCE_AUDIO_CAPTURE}); else microphone.deny(); microphone = null; }
    }
    @Override public void onRequestPermissionsResult(int request, String[] permissions, int[] results) {
        super.onRequestPermissionsResult(request, permissions, results);
        if (request == AUDIO_REQUEST) { if (results.length > 0 && results[0] == PackageManager.PERMISSION_GRANTED) grantMicrophone(); else if (microphone != null) { microphone.deny(); microphone = null; } }
    }
    @Override protected void onActivityResult(int request, int result, Intent data) {
        super.onActivityResult(request, result, data);
        if (request == FILE_REQUEST && fileCallback != null) {
            Uri[] files = null;
            if (result == RESULT_OK && data != null) {
                if (data.getClipData() != null) { int size = Math.min(data.getClipData().getItemCount(), 6); files = new Uri[size]; for (int i = 0; i < size; i++) files[i] = data.getClipData().getItemAt(i).getUri(); }
                else if (data.getData() != null) files = new Uri[]{data.getData()};
            }
            fileCallback.onReceiveValue(files); fileCallback = null;
        }
    }
    public void showSettings() {
        alert().setTitle("Remote Agent " + BuildConfig.VERSION_NAME)
            .setItems(new String[]{"检查更新", preferences.getBoolean("autoUpdate", true) ? "自动更新检查：已开启" : "自动更新检查：已关闭", "修改工作台站点", "重新加载工作台"}, (d, which) -> {
                if (which == 0) updater.check(true);
                else if (which == 1) { boolean enabled = !preferences.getBoolean("autoUpdate", true); preferences.edit().putBoolean("autoUpdate", enabled).apply(); toast(enabled ? "启动时自动检查并下载更新，安装由系统确认" : "已关闭自动检查，可手动更新"); }
                else if (which == 2) changeSite(null);
                else web.loadUrl(site);
            }).setNegativeButton("关闭", null).show();
    }
    private void changeSite(String proposed) {
        if (proposed != null) {
            try {
                String chosen = normalizeSite(proposed);
                if (chosen.equals(site)) { publishAppState(); return; }
                alert().setTitle("切换工作台站点？").setMessage(chosen + "\n\n将清除原站点登录状态和缓存，需要重新登录。")
                    .setPositiveButton("切换站点", (d, which) -> saveSite(chosen)).setNegativeButton("取消", null).show();
            } catch (Exception e) { toast("请输入完整 HTTPS 站点地址，不包含账号、查询参数或片段"); }
            return;
        }
        EditText text = new EditText(this); text.setSingleLine(true); text.setText(site);
        AlertDialog dialog = alert().setTitle("工作台 HTTPS 地址").setMessage("例如 https://codex.sunny-string.cn/agents/\n更换站点会清除原站点登录状态和缓存。").setView(text).setPositiveButton("保存", null).setNegativeButton("取消", null).create();
        dialog.setOnShowListener(v -> dialog.getButton(AlertDialog.BUTTON_POSITIVE).setOnClickListener(button -> {
            try {
                saveSite(normalizeSite(text.getText().toString())); dialog.dismiss();
            } catch (Exception e) { text.setError("请输入完整 HTTPS 站点地址，不包含账号、查询参数或片段"); }
        })); dialog.show();
    }
    private static String normalizeSite(String value) throws Exception {
        URI uri = new URI(value.trim());
        if (!"https".equals(uri.getScheme()) || uri.getHost() == null || uri.getRawUserInfo() != null || uri.getQuery() != null || uri.getFragment() != null) throw new IllegalArgumentException();
        String next = uri.toString(); return next.endsWith("/") ? next : next + "/";
    }
    private void saveSite(String chosen) {
        web.stopLoading(); CookieManager.getInstance().removeAllCookies(done -> {
            android.webkit.WebStorage.getInstance().deleteAllData(); web.clearCache(true); web.clearHistory();
            site = chosen; preferences.edit().putString("site", site).apply(); CookieManager.getInstance().flush(); web.loadUrl(site);
        });
    }
    private void goBack() {
        if (trusted(web.getUrl()) && web.getVisibility() == View.VISIBLE) {
            web.evaluateJavascript("typeof window.remoteAgentAndroidBack === 'function' && window.remoteAgentAndroidBack()", consumed -> { if (!"true".equals(consumed)) navigateBack(); });
        } else navigateBack();
    }
    private void navigateBack() { if (web.canGoBack()) web.goBack(); else alert().setMessage("退出工作台？电脑上的任务会继续运行。").setPositiveButton("退出", (d, i) -> finish()).setNegativeButton("取消", null).show(); }
    // API 33+ uses OnBackInvokedDispatcher above; this fallback supports API 26–32.
    @android.annotation.SuppressLint("GestureBackNavigation")
    @Override public void onBackPressed() { goBack(); }
    @Override protected void onSaveInstanceState(Bundle state) { web.saveState(state); super.onSaveInstanceState(state); }
    @Override protected void onPause() { CookieManager.getInstance().flush(); super.onPause(); }
    @Override protected void onDestroy() { if (fileCallback != null) fileCallback.onReceiveValue(null); if (microphone != null) microphone.deny(); updater.close(); web.destroy(); super.onDestroy(); }
    void toast(String message) { runOnUiThread(() -> { if (!isFinishing() && !isDestroyed()) Toast.makeText(this, message, Toast.LENGTH_LONG).show(); }); }
}
