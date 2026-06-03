package ru.qantrix.embedded;

import android.annotation.SuppressLint;
import android.app.Activity;
import android.net.Uri;
import android.os.Bundle;
import android.webkit.CookieManager;
import android.webkit.WebResourceRequest;
import android.webkit.WebResourceResponse;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;

import androidx.webkit.WebViewAssetLoader;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;

public class MainActivity extends Activity {
    private WebView webView;
    private WebViewAssetLoader assetLoader;

    @Override
    @SuppressLint("SetJavaScriptEnabled")
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        assetLoader = new WebViewAssetLoader.Builder()
                .setDomain(BuildConfig.APP_HOST)
                .addPathHandler("/app/", new WebViewAssetLoader.AssetsPathHandler(this, "public/app/"))
                .build();

        webView = new WebView(this);
        setContentView(webView);

        WebSettings settings = webView.getSettings();
        settings.setJavaScriptEnabled(true);
        settings.setDomStorageEnabled(true);
        settings.setDatabaseEnabled(true);
        settings.setMediaPlaybackRequiresUserGesture(false);
        settings.setAllowFileAccess(false);
        settings.setAllowContentAccess(false);

        CookieManager.getInstance().setAcceptCookie(true);
        CookieManager.getInstance().setAcceptThirdPartyCookies(webView, true);

        webView.setWebViewClient(new EmbeddedWebViewClient());
        webView.loadUrl(BuildConfig.START_URL);
    }

    @Override
    public void onBackPressed() {
        if (webView != null && webView.canGoBack()) {
            webView.goBack();
            return;
        }
        super.onBackPressed();
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

            return assetLoader.shouldInterceptRequest(uri);
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

        private WebResourceResponse textResponse(int statusCode, String reasonPhrase, String body) {
            ByteArrayInputStream stream = new ByteArrayInputStream(body.getBytes(StandardCharsets.UTF_8));
            WebResourceResponse response = new WebResourceResponse("text/plain", "UTF-8", stream);
            response.setStatusCodeAndReasonPhrase(statusCode, reasonPhrase);
            return response;
        }
    }
}
