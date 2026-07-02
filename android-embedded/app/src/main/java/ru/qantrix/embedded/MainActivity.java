package ru.qantrix.embedded;

import android.Manifest;
import android.annotation.SuppressLint;
import android.app.Activity;
import android.app.DownloadManager;
import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.content.IntentFilter;
import android.content.pm.PackageManager;
import android.database.Cursor;
import android.graphics.Color;
import android.net.Uri;
import android.os.Build;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.speech.RecognitionListener;
import android.speech.RecognizerIntent;
import android.speech.SpeechRecognizer;
import android.webkit.JavascriptInterface;
import android.webkit.CookieManager;
import android.webkit.WebResourceRequest;
import android.webkit.WebResourceResponse;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;

import androidx.core.content.FileProvider;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.net.HttpURLConnection;
import java.net.URL;
import java.net.URLConnection;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;

public class MainActivity extends Activity {
    private static final String LATEST_RELEASE_API =
            "https://api.github.com/repos/positron48/english-ai-bot/releases/latest";

    private static final int REQ_RECORD_AUDIO = 4201;

    private WebView webView;
    private int lastInsetTop = 0;
    private int lastInsetBottom = 0;
    private long updateDownloadId = -1;
    private SpeechRecognizer speechRecognizer;
    private String pendingSpeechLang;
    private BroadcastReceiver downloadReceiver;
    private final Handler downloadProgressHandler = new Handler(Looper.getMainLooper());
    private final Runnable downloadProgressPoller = new Runnable() {
        @Override
        public void run() {
            if (updateDownloadId == -1) {
                return;
            }
            pollUpdateDownloadProgress();
            if (updateDownloadId != -1) {
                downloadProgressHandler.postDelayed(this, 1000);
            }
        }
    };

    @Override
    @SuppressLint("SetJavaScriptEnabled")
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        webView = new WebView(this);
        setContentView(webView);

        WebSettings settings = webView.getSettings();
        settings.setJavaScriptEnabled(true);
        settings.setDomStorageEnabled(true);
        settings.setDatabaseEnabled(true);
        settings.setMediaPlaybackRequiresUserGesture(false);
        settings.setAllowFileAccess(false);
        settings.setAllowContentAccess(false);
        settings.setUserAgentString(settings.getUserAgentString() + " QantrixEmbeddedApp");

        CookieManager.getInstance().setAcceptCookie(true);
        CookieManager.getInstance().setAcceptThirdPartyCookies(webView, true);

        webView.addJavascriptInterface(new AndroidBridge(), "QantrixAndroid");
        webView.setWebViewClient(new EmbeddedWebViewClient());

        // The WebView draws edge-to-edge under the system bars on some devices (e.g. Samsung)
        // without reliably exposing env(safe-area-inset-*) to CSS. Read the real insets and
        // push them to the web layer as CSS px so the layout can pad the status/nav bars.
        final float density = getResources().getDisplayMetrics().density;
        webView.setOnApplyWindowInsetsListener((v, insets) -> {
            lastInsetTop = Math.round(insets.getSystemWindowInsetTop() / density);
            lastInsetBottom = Math.round(insets.getSystemWindowInsetBottom() / density);
            pushSafeAreaInsets();
            return insets;
        });

        loadBundledApp();
    }

    private void pushSafeAreaInsets() {
        if (webView == null) {
            return;
        }
        final String js = "window.__setSafeAreaInsets && window.__setSafeAreaInsets("
                + lastInsetTop + "," + lastInsetBottom + ")";
        webView.post(() -> webView.evaluateJavascript(js, null));
    }

    @Override
    public void onBackPressed() {
        if (webView != null && webView.canGoBack()) {
            webView.goBack();
            return;
        }
        super.onBackPressed();
    }

    @Override
    protected void onDestroy() {
        releaseSpeechRecognizer();
        downloadProgressHandler.removeCallbacks(downloadProgressPoller);
        if (downloadReceiver != null) {
            try {
                unregisterReceiver(downloadReceiver);
            } catch (IllegalArgumentException ignored) {
                // Receiver was not registered; nothing to clean up.
            }
            downloadReceiver = null;
        }
        super.onDestroy();
    }

    private void loadBundledApp() {
        try {
            String html = readAssetText("public/app/index.html");
            webView.loadDataWithBaseURL(BuildConfig.START_URL, html, "text/html", "UTF-8", BuildConfig.START_URL);
        } catch (IOException ignored) {
            webView.loadUrl(BuildConfig.START_URL);
        }
    }

    private String readAssetText(String assetPath) throws IOException {
        try (InputStream input = getAssets().open(assetPath);
             ByteArrayOutputStream output = new ByteArrayOutputStream()) {
            byte[] buffer = new byte[8192];
            int read;
            while ((read = input.read(buffer)) != -1) {
                output.write(buffer, 0, read);
            }
            return output.toString(StandardCharsets.UTF_8.name());
        }
    }

    private final class AndroidBridge {
        @JavascriptInterface
        public void setSystemBarsColor(String color) {
            runOnUiThread(() -> {
                try {
                    int parsed = Color.parseColor(color);
                    getWindow().setStatusBarColor(parsed);
                    getWindow().setNavigationBarColor(parsed);
                } catch (IllegalArgumentException ignored) {
                    // Ignore invalid CSS colors from the web layer.
                }
            });
        }

        @JavascriptInterface
        public String getAppVersion() {
            return BuildConfig.VERSION_NAME;
        }

        @JavascriptInterface
        public void checkLatestVersion() {
            new Thread(MainActivity.this::fetchLatestRelease).start();
        }

        @JavascriptInterface
        public void startUpdateDownload(String apkUrl) {
            runOnUiThread(() -> beginUpdateDownload(apkUrl));
        }

        @JavascriptInterface
        public void cancelUpdateDownload() {
            runOnUiThread(MainActivity.this::cancelUpdateDownload);
        }

        @JavascriptInterface
        public boolean speechRecognitionAvailable() {
            return SpeechRecognizer.isRecognitionAvailable(MainActivity.this);
        }

        @JavascriptInterface
        public void startSpeechRecognition(String lang) {
            runOnUiThread(() -> beginSpeechRecognition(lang));
        }

        @JavascriptInterface
        public void stopSpeechRecognition() {
            runOnUiThread(MainActivity.this::stopSpeechRecognitionInternal);
        }
    }

    // --- Native speech recognition -----------------------------------------
    // The Android System WebView does not expose the Web Speech API, so we bridge
    // to android.speech.SpeechRecognizer and report results back to the web layer
    // via window.__onSpeechResult / window.__onSpeechState.

    private void beginSpeechRecognition(String lang) {
        pendingSpeechLang = lang;
        if (checkSelfPermission(Manifest.permission.RECORD_AUDIO) != PackageManager.PERMISSION_GRANTED) {
            requestPermissions(new String[]{Manifest.permission.RECORD_AUDIO}, REQ_RECORD_AUDIO);
            return;
        }
        launchSpeechRecognizer(lang);
    }

    @Override
    public void onRequestPermissionsResult(int requestCode, String[] permissions, int[] grantResults) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults);
        if (requestCode == REQ_RECORD_AUDIO) {
            if (grantResults.length > 0 && grantResults[0] == PackageManager.PERMISSION_GRANTED) {
                launchSpeechRecognizer(pendingSpeechLang);
            } else {
                reportSpeechResult(null, "permission_denied");
            }
        }
    }

    private void launchSpeechRecognizer(String lang) {
        releaseSpeechRecognizer();
        if (!SpeechRecognizer.isRecognitionAvailable(this)) {
            reportSpeechResult(null, "unavailable");
            return;
        }
        speechRecognizer = SpeechRecognizer.createSpeechRecognizer(this);
        speechRecognizer.setRecognitionListener(new RecognitionListener() {
            @Override public void onReadyForSpeech(Bundle params) {
                evalJs("window.__onSpeechState && window.__onSpeechState('listening')");
            }
            @Override public void onBeginningOfSpeech() { }
            @Override public void onRmsChanged(float rmsdB) { }
            @Override public void onBufferReceived(byte[] buffer) { }
            @Override public void onEndOfSpeech() { }
            @Override public void onError(int error) {
                reportSpeechResult(null, "error_" + error);
            }
            @Override public void onResults(Bundle results) {
                ArrayList<String> matches =
                        results.getStringArrayList(SpeechRecognizer.RESULTS_RECOGNITION);
                String text = (matches != null && !matches.isEmpty()) ? matches.get(0) : "";
                reportSpeechResult(text, null);
            }
            @Override public void onPartialResults(Bundle partialResults) { }
            @Override public void onEvent(int eventType, Bundle params) { }
        });

        Intent intent = new Intent(RecognizerIntent.ACTION_RECOGNIZE_SPEECH);
        intent.putExtra(RecognizerIntent.EXTRA_LANGUAGE_MODEL,
                RecognizerIntent.LANGUAGE_MODEL_FREE_FORM);
        if (lang != null && !lang.isEmpty()) {
            intent.putExtra(RecognizerIntent.EXTRA_LANGUAGE, lang);
        }
        intent.putExtra(RecognizerIntent.EXTRA_MAX_RESULTS, 1);
        try {
            speechRecognizer.startListening(intent);
        } catch (Exception e) {
            reportSpeechResult(null, e.getClass().getSimpleName());
        }
    }

    private void stopSpeechRecognitionInternal() {
        if (speechRecognizer != null) {
            try {
                speechRecognizer.stopListening();
            } catch (Exception ignored) {
                // Best effort; results/error listener still fires.
            }
        }
    }

    private void releaseSpeechRecognizer() {
        if (speechRecognizer != null) {
            try {
                speechRecognizer.destroy();
            } catch (Exception ignored) {
                // Ignore teardown failures.
            }
            speechRecognizer = null;
        }
    }

    private void reportSpeechResult(String text, String error) {
        StringBuilder sb = new StringBuilder();
        sb.append("window.__onSpeechResult && window.__onSpeechResult({");
        boolean hasField = false;
        if (text != null) {
            sb.append("\"transcript\":\"").append(jsEscape(text)).append("\"");
            hasField = true;
        }
        if (error != null) {
            if (hasField) {
                sb.append(",");
            }
            sb.append("\"error\":\"").append(jsEscape(error)).append("\"");
        }
        sb.append("})");
        evalJs(sb.toString());
        evalJs("window.__onSpeechState && window.__onSpeechState('ended')");
        releaseSpeechRecognizer();
    }

    // Queries the GitHub "latest release" API and reports {latestVersion, apkUrl, error}
    // back to the web layer. Runs on a background thread (network on the UI thread is forbidden).
    private void fetchLatestRelease() {
        HttpURLConnection connection = null;
        try {
            URL url = new URL(LATEST_RELEASE_API);
            connection = (HttpURLConnection) url.openConnection();
            connection.setRequestMethod("GET");
            connection.setRequestProperty("Accept", "application/vnd.github+json");
            connection.setRequestProperty("User-Agent", "QantrixLinglowApp");
            connection.setConnectTimeout(15000);
            connection.setReadTimeout(15000);

            int status = connection.getResponseCode();
            if (status != HttpURLConnection.HTTP_OK) {
                reportUpdateCheckError("HTTP " + status);
                return;
            }

            String body = readStream(connection.getInputStream());
            JSONObject release = new JSONObject(body);
            String tag = release.optString("tag_name", "");

            String apkUrl = "";
            JSONArray assets = release.optJSONArray("assets");
            if (assets != null) {
                for (int i = 0; i < assets.length(); i++) {
                    JSONObject asset = assets.optJSONObject(i);
                    if (asset == null) {
                        continue;
                    }
                    String name = asset.optString("name", "");
                    if (name.startsWith("qantrix-linglow-") && name.endsWith(".apk")) {
                        apkUrl = asset.optString("browser_download_url", "");
                        break;
                    }
                }
            }

            if (tag.isEmpty() || apkUrl.isEmpty()) {
                reportUpdateCheckError("no_apk_asset");
                return;
            }

            String js = "window.__onUpdateCheckResult && window.__onUpdateCheckResult({"
                    + "\"latestVersion\":\"" + jsEscape(tag) + "\","
                    + "\"apkUrl\":\"" + jsEscape(apkUrl) + "\"})";
            evalJs(js);
        } catch (Exception e) {
            reportUpdateCheckError(e.getClass().getSimpleName());
        } finally {
            if (connection != null) {
                connection.disconnect();
            }
        }
    }

    private void reportUpdateCheckError(String message) {
        evalJs("window.__onUpdateCheckResult && window.__onUpdateCheckResult({\"error\":\""
                + jsEscape(message) + "\"})");
    }

    private void beginUpdateDownload(String apkUrl) {
        try {
            registerDownloadReceiver();

            File targetDir = getExternalFilesDir(android.os.Environment.DIRECTORY_DOWNLOADS);
            File apkFile = new File(targetDir, "linglow-update.apk");
            if (apkFile.exists()) {
                apkFile.delete();
            }

            DownloadManager.Request request = new DownloadManager.Request(Uri.parse(apkUrl));
            request.setMimeType("application/vnd.android.package-archive");
            request.setNotificationVisibility(
                    DownloadManager.Request.VISIBILITY_VISIBLE_NOTIFY_COMPLETED);
            request.setDestinationInExternalFilesDir(
                    this, android.os.Environment.DIRECTORY_DOWNLOADS, "linglow-update.apk");

            DownloadManager manager = (DownloadManager) getSystemService(DOWNLOAD_SERVICE);
            updateDownloadId = manager.enqueue(request);
            reportDownloadState("downloading", null);
            downloadProgressHandler.removeCallbacks(downloadProgressPoller);
            downloadProgressHandler.post(downloadProgressPoller);
        } catch (Exception e) {
            reportDownloadState("error", e.getClass().getSimpleName());
        }
    }

    private void cancelUpdateDownload() {
        if (updateDownloadId == -1) {
            return;
        }
        DownloadManager manager = (DownloadManager) getSystemService(DOWNLOAD_SERVICE);
        manager.remove(updateDownloadId);
        updateDownloadId = -1;
        downloadProgressHandler.removeCallbacks(downloadProgressPoller);
    }

    private void registerDownloadReceiver() {
        if (downloadReceiver != null) {
            return;
        }
        downloadReceiver = new BroadcastReceiver() {
            @Override
            public void onReceive(Context context, Intent intent) {
                long id = intent.getLongExtra(DownloadManager.EXTRA_DOWNLOAD_ID, -1);
                if (id != updateDownloadId) {
                    return;
                }
                onUpdateDownloadComplete(id);
            }
        };
        IntentFilter filter = new IntentFilter(DownloadManager.ACTION_DOWNLOAD_COMPLETE);
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            registerReceiver(downloadReceiver, filter, Context.RECEIVER_EXPORTED);
        } else {
            registerReceiver(downloadReceiver, filter);
        }
    }

    private void onUpdateDownloadComplete(long id) {
        downloadProgressHandler.removeCallbacks(downloadProgressPoller);
        DownloadManager manager = (DownloadManager) getSystemService(DOWNLOAD_SERVICE);
        DownloadManager.Query query = new DownloadManager.Query().setFilterById(id);
        try (Cursor cursor = manager.query(query)) {
            if (cursor == null || !cursor.moveToFirst()) {
                reportDownloadState("error", "query_failed");
                return;
            }
            int statusCol = cursor.getColumnIndex(DownloadManager.COLUMN_STATUS);
            int status = cursor.getInt(statusCol);
            if (status != DownloadManager.STATUS_SUCCESSFUL) {
                updateDownloadId = -1;
                reportDownloadState("error", "status_" + status);
                return;
            }
        }

        File apkFile = new File(
                getExternalFilesDir(android.os.Environment.DIRECTORY_DOWNLOADS),
                "linglow-update.apk");
        Uri contentUri = FileProvider.getUriForFile(
                this, getPackageName() + ".fileprovider", apkFile);

        Intent install = new Intent(Intent.ACTION_VIEW);
        install.setDataAndType(contentUri, "application/vnd.android.package-archive");
        install.addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION);
        install.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
        reportDownloadState("installing", null);
        updateDownloadId = -1;
        startActivity(install);
    }

    private void reportDownloadState(String state, String error) {
        reportDownloadState(state, error, -1, -1);
    }

    private void pollUpdateDownloadProgress() {
        DownloadManager manager = (DownloadManager) getSystemService(DOWNLOAD_SERVICE);
        DownloadManager.Query query = new DownloadManager.Query().setFilterById(updateDownloadId);
        try (Cursor cursor = manager.query(query)) {
            if (cursor == null || !cursor.moveToFirst()) {
                return;
            }
            int status = cursor.getInt(cursor.getColumnIndexOrThrow(DownloadManager.COLUMN_STATUS));
            if (status == DownloadManager.STATUS_FAILED) {
                updateDownloadId = -1;
                downloadProgressHandler.removeCallbacks(downloadProgressPoller);
                reportDownloadState("error", "status_" + status);
                return;
            }
            long downloaded = cursor.getLong(cursor.getColumnIndexOrThrow(DownloadManager.COLUMN_BYTES_DOWNLOADED_SO_FAR));
            long total = cursor.getLong(cursor.getColumnIndexOrThrow(DownloadManager.COLUMN_TOTAL_SIZE_BYTES));
            reportDownloadState("downloading", null, downloaded, total);
        } catch (Exception ignored) {
            // Progress is best-effort; completion receiver still reports final state.
        }
    }

    private void reportDownloadState(String state, String error, long bytesDownloaded, long bytesTotal) {
        StringBuilder sb = new StringBuilder();
        sb.append("window.__onUpdateDownload && window.__onUpdateDownload({\"state\":\"")
                .append(jsEscape(state)).append("\"");
        if (error != null) {
            sb.append(",\"error\":\"").append(jsEscape(error)).append("\"");
        }
        if (bytesDownloaded >= 0) {
            sb.append(",\"bytesDownloaded\":").append(bytesDownloaded);
        }
        if (bytesTotal > 0) {
            sb.append(",\"bytesTotal\":").append(bytesTotal);
            int progress = Math.max(0, Math.min(100, Math.round((bytesDownloaded * 100f) / bytesTotal)));
            sb.append(",\"progress\":").append(progress);
        }
        sb.append("})");
        evalJs(sb.toString());
    }

    private void evalJs(String js) {
        if (webView == null) {
            return;
        }
        webView.post(() -> webView.evaluateJavascript(js, null));
    }

    private static String jsEscape(String value) {
        if (value == null) {
            return "";
        }
        return value.replace("\\", "\\\\").replace("\"", "\\\"")
                .replace("\n", "\\n").replace("\r", "\\r");
    }

    private static String readStream(InputStream input) throws IOException {
        try (ByteArrayOutputStream output = new ByteArrayOutputStream()) {
            byte[] buffer = new byte[8192];
            int read;
            while ((read = input.read(buffer)) != -1) {
                output.write(buffer, 0, read);
            }
            return output.toString(StandardCharsets.UTF_8.name());
        }
    }

    private final class EmbeddedWebViewClient extends WebViewClient {
        @Override
        public void onPageFinished(WebView view, String url) {
            super.onPageFinished(view, url);
            // Re-push insets once the web layer is ready to receive them.
            pushSafeAreaInsets();
        }

        @Override
        public WebResourceResponse shouldInterceptRequest(WebView view, WebResourceRequest request) {
            return intercept(request.getUrl());
        }

        @Override
        public WebResourceResponse shouldInterceptRequest(WebView view, String url) {
            return intercept(Uri.parse(url));
        }

        private WebResourceResponse intercept(Uri uri) {
            if (!"https".equals(uri.getScheme()) || !BuildConfig.APP_HOST.equals(uri.getHost())) {
                return null;
            }

            String path = uri.getPath();
            if (path == null) {
                return null;
            }

            if (path.equals("/sw.js")) {
                return textResponse(404, "Not Found", "Service worker is disabled in embedded APK.");
            }

            if (!path.startsWith("/app")) {
                return null;
            }

            if (isSpaRoute(path)) {
                return openAsset("public/app/index.html", "text/html");
            }

            return openAppAsset(path);
        }

        private boolean isSpaRoute(String path) {
            if (path.equals("/app") || path.equals("/app/")) {
                return true;
            }
            String lastSegment = path.substring(path.lastIndexOf('/') + 1);
            return !lastSegment.contains(".");
        }

        private WebResourceResponse openAsset(String assetPath, String mimeType) {
            try {
                InputStream stream = getAssets().open(assetPath);
                return new WebResourceResponse(mimeType, "UTF-8", stream);
            } catch (IOException ignored) {
                return null;
            }
        }

        private WebResourceResponse openAppAsset(String path) {
            String relativePath = path.startsWith("/app/") ? path.substring("/app/".length()) : path;
            String mimeType = URLConnection.guessContentTypeFromName(relativePath);
            if (mimeType == null) {
                mimeType = "application/octet-stream";
            }
            return openAsset("public/app/" + relativePath, mimeType);
        }

        private WebResourceResponse textResponse(int statusCode, String reasonPhrase, String body) {
            ByteArrayInputStream stream = new ByteArrayInputStream(body.getBytes(StandardCharsets.UTF_8));
            WebResourceResponse response = new WebResourceResponse("text/plain", "UTF-8", stream);
            response.setStatusCodeAndReasonPhrase(statusCode, reasonPhrase);
            return response;
        }
    }
}
