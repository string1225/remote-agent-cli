package cn.sunnystring.remoteagent;

import android.app.AlertDialog;
import android.content.Intent;
import android.content.pm.PackageInfo;
import android.content.pm.PackageManager;
import android.content.pm.Signature;
import android.net.Uri;
import android.os.Build;
import android.provider.Settings;
import java.io.ByteArrayOutputStream;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.InputStream;
import java.net.HttpURLConnection;
import java.net.URL;
import java.nio.file.Files;
import java.nio.file.StandardCopyOption;
import java.security.MessageDigest;
import java.util.Arrays;
import java.util.HashSet;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.atomic.AtomicBoolean;
import org.json.JSONObject;

// Updates pass through gh-proxy but are accepted only if their package, version,
// SHA-256 and signing certificate match this installation's release identity.
final class AppUpdater {
    private final MainActivity activity;
    private final ExecutorService worker = Executors.newSingleThreadExecutor();
    private final AtomicBoolean busy = new AtomicBoolean(false);
    private volatile boolean closed;
    private volatile HttpURLConnection active;
    AppUpdater(MainActivity activity) { this.activity = activity; }
    void close() { closed = true; HttpURLConnection connection = active; if (connection != null) connection.disconnect(); worker.shutdownNow(); }
    void check(boolean manual) {
        if (closed) return;
        if (!busy.compareAndSet(false, true)) { if (manual) activity.toast("正在检查或下载更新，请稍候"); return; }
        if (manual) activity.toast("正在通过 gh-proxy 检查更新…");
        worker.execute(() -> {
            File partial = null;
            try {
                JSONObject update = new JSONObject(new String(fetch(BuildConfig.UPDATE_URL + "?t=" + System.currentTimeMillis(), 64 * 1024), java.nio.charset.StandardCharsets.UTF_8));
                int version = update.getInt("versionCode");
                if (version <= BuildConfig.VERSION_CODE) { if (manual) activity.toast("当前已是最新版本"); return; }
                String hash = update.getString("sha256").toLowerCase(java.util.Locale.ROOT);
                if (!hash.matches("[a-f0-9]{64}")) throw new Exception("更新清单校验值无效");
                String url = update.getString("apkUrl");
                if (!url.startsWith("https://gh-proxy.com/https://github.com/string1225/remote-agent-cli/releases/download/")) throw new Exception("更新下载地址不受支持");
                File directory = new File(activity.getFilesDir(), "updates");
                if (!directory.isDirectory() && !directory.mkdirs()) throw new Exception("无法创建更新目录");
                partial = new File(directory, "update.part"); File apk = new File(directory, "update.apk");
                activity.toast("发现新版本，正在下载并校验 APK…");
                HttpURLConnection connection = connect(url);
                try (InputStream input = connection.getInputStream(); FileOutputStream output = new FileOutputStream(partial)) {
                    byte[] buffer = new byte[32768]; long count = 0; int n;
                    while ((n = input.read(buffer)) != -1) { if (closed || Thread.currentThread().isInterrupted()) throw new Exception("更新已中止"); count += n; if (count > 100L * 1024 * 1024) throw new Exception("更新包过大"); output.write(buffer, 0, n); }
                    output.getFD().sync();
                } finally { connection.disconnect(); active = null; }
                if (!hash.equals(sha256(partial))) throw new Exception("APK SHA-256 校验失败");
                verifyPackage(partial, version);
                Files.move(partial.toPath(), apk.toPath(), StandardCopyOption.REPLACE_EXISTING); partial = null;
                String versionName = update.optString("versionName", String.valueOf(version));
                activity.runOnUiThread(() -> {
                    if (closed || activity.isFinishing() || activity.isDestroyed()) return;
                    new AlertDialog.Builder(activity).setTitle("更新已就绪 · " + versionName).setMessage("已通过校验。安装将保留登录状态与本地草稿，电脑上的任务继续运行。")
                        .setPositiveButton("安装更新", (d, which) -> install(apk)).setNegativeButton("稍后", null).show();
                });
            } catch (Exception e) { if (manual || !closed) activity.toast("更新失败：" + (e.getMessage() == null ? "请检查网络后重试" : e.getMessage())); }
            finally { if (partial != null) partial.delete(); busy.set(false); }
        });
    }
    private byte[] fetch(String url, int limit) throws Exception {
        HttpURLConnection connection = connect(url);
        try (InputStream input = connection.getInputStream(); ByteArrayOutputStream output = new ByteArrayOutputStream()) {
            byte[] buffer = new byte[8192]; int n;
            while ((n = input.read(buffer)) != -1) { if (closed) throw new Exception("更新已中止"); if (output.size() + n > limit) throw new Exception("更新清单过大"); output.write(buffer, 0, n); }
            return output.toByteArray();
        } finally { connection.disconnect(); active = null; }
    }
    private HttpURLConnection connect(String address) throws Exception {
        URL url = new URL(address);
        for (int redirects = 0; redirects < 6; redirects++) {
            if (closed || !"https".equals(url.getProtocol()) || url.getUserInfo() != null) throw new Exception("无效更新连接");
            HttpURLConnection connection = (HttpURLConnection) url.openConnection(); active = connection;
            connection.setConnectTimeout(20000); connection.setReadTimeout(45000); connection.setInstanceFollowRedirects(false);
            connection.setRequestProperty("Cache-Control", "no-cache"); connection.setRequestProperty("User-Agent", "RemoteAgentAndroid/" + BuildConfig.VERSION_NAME);
            int status = connection.getResponseCode();
            if (status >= 300 && status < 400) { String location = connection.getHeaderField("Location"); connection.disconnect(); if (location == null) throw new Exception("更新重定向无地址"); url = new URL(url, location); continue; }
            if (status != 200) { connection.disconnect(); throw new Exception("更新服务返回 HTTP " + status); }
            return connection;
        }
        throw new Exception("更新重定向过多");
    }
    static String sha256(File file) throws Exception {
        MessageDigest digest = MessageDigest.getInstance("SHA-256");
        try (InputStream input = new FileInputStream(file)) { byte[] buffer = new byte[32768]; int n; while ((n = input.read(buffer)) != -1) digest.update(buffer, 0, n); }
        StringBuilder text = new StringBuilder(); for (byte value : digest.digest()) text.append(String.format(java.util.Locale.ROOT, "%02x", value & 255)); return text.toString();
    }
    @SuppressWarnings("deprecation") private void verifyPackage(File apk, int version) throws Exception {
        PackageManager pm = activity.getPackageManager();
        int flags = Build.VERSION.SDK_INT >= 28 ? PackageManager.GET_SIGNING_CERTIFICATES : PackageManager.GET_SIGNATURES;
        PackageInfo candidate = pm.getPackageArchiveInfo(apk.getAbsolutePath(), flags), installed = pm.getPackageInfo(activity.getPackageName(), flags);
        if (candidate == null || !activity.getPackageName().equals(candidate.packageName) || (Build.VERSION.SDK_INT >= 28 ? candidate.getLongVersionCode() : candidate.versionCode) != version || version <= BuildConfig.VERSION_CODE) throw new Exception("APK 包名或版本不匹配");
        Signature[] current = signatures(installed), incoming = signatures(candidate);
        if (current == null || incoming == null || current.length == 0 || incoming.length != current.length || !new HashSet<>(Arrays.asList(current)).equals(new HashSet<>(Arrays.asList(incoming)))) throw new Exception("APK 签名与已安装版本不匹配");
    }
    @SuppressWarnings("deprecation") private static Signature[] signatures(PackageInfo info) {
        return Build.VERSION.SDK_INT >= 28 ? (info.signingInfo == null ? null : info.signingInfo.getApkContentsSigners()) : info.signatures;
    }
    private void install(File apk) {
        if (!activity.getPackageManager().canRequestPackageInstalls()) {
            new AlertDialog.Builder(activity).setMessage("请允许 Remote Agent 安装更新，返回后再次检查更新并点安装。")
                .setPositiveButton("打开系统设置", (d, which) -> activity.startActivity(new Intent(Settings.ACTION_MANAGE_UNKNOWN_APP_SOURCES, Uri.parse("package:" + activity.getPackageName())))).setNegativeButton("取消", null).show(); return;
        }
        try {
            Uri uri = Uri.parse("content://" + activity.getPackageName() + ".updates/update.apk");
            Intent intent = new Intent(Intent.ACTION_VIEW).setDataAndType(uri, "application/vnd.android.package-archive").addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION);
            activity.startActivity(intent);
        } catch (Exception e) { activity.toast("无法打开系统安装器，请手动下载 Release APK 安装"); }
    }
}
