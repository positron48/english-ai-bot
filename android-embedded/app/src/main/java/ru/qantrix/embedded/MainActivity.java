package ru.qantrix.embedded;

import android.annotation.SuppressLint;
import android.app.Activity;
import android.graphics.Color;
import android.net.Uri;
import android.os.Bundle;
import android.webkit.JavascriptInterface;
import android.webkit.CookieManager;
import android.webkit.WebResourceRequest;
import android.webkit.WebResourceResponse;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.net.URLConnection;
import java.nio.charset.StandardCharsets;

public class MainActivity extends Activity {
    private WebView webView;

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
        loadBundledApp();
    }

    @Override
    public void onBackPressed() {
        if (webView != null && webView.canGoBack()) {
            webView.goBack();
            return;
        }
        super.onBackPressed();
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
    }

    private final class EmbeddedWebViewClient extends WebViewClient {
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
