//go:build test

package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

const gapUserID int64 = 900001

func gapRouter(cfg *config.Config) *Router {
	if cfg == nil {
		cfg = &config.Config{WebApp: config.WebAppConfig{JWTSecret: "gap-test"}}
	}
	return &Router{
		mux:    http.NewServeMux(),
		logger: zap.NewNop(),
		config: cfg,
	}
}

func gapPronSvc(t *testing.T, tag string) *service.PronunciationService {
	t.Helper()
	cfg := config.TTSConfig{
		Enabled:        true,
		Provider:       "dictionary",
		AudioDir:       t.TempDir(),
		PublicBasePath: "/media/" + tag,
	}
	return service.NewPronunciationService(cfg, config.DefaultLearningConfig(), nil, zap.NewNop())
}

func TestRouterGapSetGrammarServices(t *testing.T) {
	router := gapRouter(nil)
	defaultSvc := &service.GrammarService{}
	enSvc := &service.GrammarService{}
	esSvc := &service.GrammarService{}

	router.SetGrammarService(defaultSvc)
	router.SetGrammarServices(map[string]*service.GrammarService{
		"en": enSvc,
		"es": esSvc,
	})

	if router.grammarServices["en"] != enSvc || router.grammarServices["es"] != esSvc {
		t.Fatal("SetGrammarServices did not register per-bundle services")
	}
}

func TestRouterGapGrammarServiceForRequest(t *testing.T) {
	defaultSvc := &service.GrammarService{}
	enSvc := &service.GrammarService{}
	esSvc := &service.GrammarService{}

	tests := []struct {
		name       string
		setup      func(*Router)
		query      string
		want       *service.GrammarService
	}{
		{
			name: "no multi-bundle map returns default",
			setup: func(r *Router) {
				r.grammarService = defaultSvc
			},
			want: defaultSvc,
		},
		{
			name: "course_code query selects en bundle",
			setup: func(r *Router) {
				r.grammarService = defaultSvc
				r.grammarServices = map[string]*service.GrammarService{"en": enSvc, "es": esSvc}
			},
			query: "?course_code=en_ru",
			want:  enSvc,
		},
		{
			name: "course_code query selects es bundle",
			setup: func(r *Router) {
				r.grammarService = defaultSvc
				r.grammarServices = map[string]*service.GrammarService{"en": enSvc, "es": esSvc}
			},
			query: "?course_code=es_ru",
			want:  esSvc,
		},
		{
			name: "unknown bundle falls back to default",
			setup: func(r *Router) {
				r.grammarService = defaultSvc
				r.grammarServices = map[string]*service.GrammarService{"en": enSvc}
			},
			query: "?course_code=fr_ru",
			want:  defaultSvc,
		},
		{
			name: "empty course resolves via currentCourseCodeForUser then fallback",
			setup: func(r *Router) {
				r.grammarService = defaultSvc
				r.grammarServices = map[string]*service.GrammarService{"en": enSvc}
			},
			want: defaultSvc,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gapRouter(nil)
			tt.setup(router)
			req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/categories"+tt.query, nil)
			got := router.grammarServiceForRequest(req, gapUserID)
			if got != tt.want {
				t.Fatalf("grammarServiceForRequest() = %p, want %p", got, tt.want)
			}
		})
	}
}

func TestRouterGapSetPronunciationServices(t *testing.T) {
	router := gapRouter(nil)
	primary := gapPronSvc(t, "primary")
	en := gapPronSvc(t, "en")
	es := gapPronSvc(t, "es")
	router.SetPronunciationService(primary)
	router.SetPronunciationServices(map[string]*service.PronunciationService{
		"en": en,
		"es": es,
	})

	if router.pronunciationServices["en"] != en || router.pronunciationServices["es"] != es {
		t.Fatal("SetPronunciationServices did not register per-bundle services")
	}
}

func TestRouterGapPronunciationServiceForRequest(t *testing.T) {
	primary := gapPronSvc(t, "primary")
	en := gapPronSvc(t, "en")
	es := gapPronSvc(t, "es")

	tests := []struct {
		name      string
		setup     func(*Router)
		query     string
		wantBase  string
	}{
		{
			name: "no multi-bundle map returns primary",
			setup: func(r *Router) {
				r.pronunciationService = primary
			},
			wantBase: "/media/primary",
		},
		{
			name: "course_code query selects en bundle",
			setup: func(r *Router) {
				r.pronunciationService = primary
				r.pronunciationServices = map[string]pronunciationServiceInterface{"en": en, "es": es}
			},
			query:    "?course_code=en_ru",
			wantBase: "/media/en",
		},
		{
			name: "course_code query selects es bundle",
			setup: func(r *Router) {
				r.pronunciationService = primary
				r.pronunciationServices = map[string]pronunciationServiceInterface{"en": en, "es": es}
			},
			query:    "?course_code=es_ru",
			wantBase: "/media/es",
		},
		{
			name: "unknown bundle falls back to primary",
			setup: func(r *Router) {
				r.pronunciationService = primary
				r.pronunciationServices = map[string]pronunciationServiceInterface{"en": en}
			},
			query:    "?course_code=fr_ru",
			wantBase: "/media/primary",
		},
		{
			name: "empty course uses currentCourseCodeForUser then fallback",
			setup: func(r *Router) {
				r.pronunciationService = primary
				r.pronunciationServices = map[string]pronunciationServiceInterface{"en": en}
			},
			wantBase: "/media/primary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gapRouter(nil)
			tt.setup(router)
			req := httptest.NewRequest(http.MethodGet, "/api/tts/word"+tt.query, nil)
			got := router.pronunciationServiceForRequest(req, gapUserID)
			if got.PublicBasePath() != tt.wantBase {
				t.Fatalf("PublicBasePath() = %q, want %q", got.PublicBasePath(), tt.wantBase)
			}
		})
	}
}

func TestRouterGapPronunciationServiceForLang(t *testing.T) {
	primary := &mockPronunciationService{publicBase: "primary"}
	en := &mockPronunciationService{publicBase: "en"}
	es := &mockPronunciationService{publicBase: "es"}
	byLang := &mockPronunciationService{publicBase: "lang-key"}

	router := gapRouter(nil)
	router.pronunciationService = primary
	router.pronunciationServices = map[string]pronunciationServiceInterface{
		"en": en,
		"es": es,
		"xx": byLang,
	}

	tests := []struct {
		name     string
		target   string
		wantBase string
	}{
		{name: "empty target falls back", target: "", wantBase: "primary"},
		{name: "whitespace target falls back", target: "  ", wantBase: "primary"},
		{name: "bundle id en", target: "en", wantBase: "en"},
		{name: "course code es_ru maps to es", target: "es_ru", wantBase: "es"},
		{name: "uppercase normalized", target: "EN", wantBase: "en"},
		{name: "direct map key when bundle id empty", target: "xx", wantBase: "lang-key"},
		{name: "unknown falls back", target: "de", wantBase: "primary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := router.pronunciationServiceForLang(tt.target)
			if got.PublicBasePath() != tt.wantBase {
				t.Fatalf("PublicBasePath() = %q, want %q", got.PublicBasePath(), tt.wantBase)
			}
		})
	}

	t.Run("course code falls back to full target lang key", func(t *testing.T) {
		router := gapRouter(nil)
		router.pronunciationService = primary
		router.pronunciationServices = map[string]pronunciationServiceInterface{
			"es_ru": &mockPronunciationService{publicBase: "es-ru-key"},
		}
		got := router.pronunciationServiceForLang("es_ru")
		if got.PublicBasePath() != "es-ru-key" {
			t.Fatalf("PublicBasePath() = %q, want es-ru-key", got.PublicBasePath())
		}
	})

	t.Run("empty bundle id uses trimmed target lang key", func(t *testing.T) {
		router := gapRouter(nil)
		router.pronunciationService = primary
		router.pronunciationServices = map[string]pronunciationServiceInterface{
			"_ru": &mockPronunciationService{publicBase: "underscore-key"},
		}
		got := router.pronunciationServiceForLang("_ru")
		if got.PublicBasePath() != "underscore-key" {
			t.Fatalf("PublicBasePath() = %q, want underscore-key", got.PublicBasePath())
		}
	})
}

func TestRouterGapParseTTSTokens(t *testing.T) {
	logger := zap.NewNop()

	t.Run("empty raw", func(t *testing.T) {
		got := parseTTSTokens("", logger)
		if len(got) != 0 {
			t.Fatalf("expected empty map, got %v", got)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		got := parseTTSTokens("{not-json", logger)
		if len(got) != 0 {
			t.Fatalf("expected empty map on parse error, got %v", got)
		}
	})

	t.Run("valid json skips empty keys and normalizes", func(t *testing.T) {
		raw := `{" EN ":" token-en ", "": "x", "es": "", "default":" tok "}`
		got := parseTTSTokens(raw, logger)
		if got["en"] != "token-en" {
			t.Fatalf("en token = %q", got["en"])
		}
		if got["default"] != "tok" {
			t.Fatalf("default token = %q", got["default"])
		}
		if _, ok := got["es"]; ok {
			t.Fatal("empty value should be skipped")
		}
	})

	t.Run("via NewRouter", func(t *testing.T) {
		cfg := &config.Config{
			TTS: config.TTSConfig{
				InternalTokensJSON: `{"en":"router-en"}`,
			},
		}
		router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)
		if router.internalTTSTokens["en"] != "router-en" {
			t.Fatalf("internalTTSTokens = %v", router.internalTTSTokens)
		}
	})
}

func TestRouterGapParseInternalServiceTokens(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name string
		cfg  config.Config
		want map[string]string
	}{
		{
			name: "json tokens only",
			cfg: config.Config{
				WebApp: config.WebAppConfig{
					InternalServiceTokensJSON: `{"en":"en-token"}`,
				},
			},
			want: map[string]string{"en": "en-token"},
		},
		{
			name: "complaints token sets default when missing",
			cfg: config.Config{
				WebApp: config.WebAppConfig{
					ComplaintsServiceToken: "complaint-secret",
				},
			},
			want: map[string]string{"default": "complaint-secret"},
		},
		{
			name: "complaints token sets complaints key when default exists",
			cfg: config.Config{
				WebApp: config.WebAppConfig{
					InternalServiceTokensJSON: `{"default":"existing"}`,
					ComplaintsServiceToken:    "complaint-secret",
				},
			},
			want: map[string]string{
				"default":    "existing",
				"complaints": "complaint-secret",
			},
		},
		{
			name: "null json tokens with complaints uses default branch",
			cfg: config.Config{
				WebApp: config.WebAppConfig{
					InternalServiceTokensJSON: "null",
					ComplaintsServiceToken:    "complaint-secret",
				},
			},
			want: map[string]string{"default": "complaint-secret"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseInternalServiceTokens(&tt.cfg, logger)
			for k, v := range tt.want {
				if got[k] != v {
					t.Fatalf("token[%q] = %q, want %q (full map=%v)", k, got[k], v, got)
				}
			}
		})
	}
}

func TestRouterGapSplitCSVNonEmpty(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{in: "", want: nil},
		{in: "  ", want: nil},
		{in: "a,b,c", want: []string{"a", "b", "c"}},
		{in: " a , , b ", want: []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := splitCSVNonEmpty(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("splitCSVNonEmpty(%q) = %v, want %v", tt.in, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("splitCSVNonEmpty(%q) = %v, want %v", tt.in, got, tt.want)
				}
			}
		})
	}
}

func TestRouterGapHandleWebAppManifest(t *testing.T) {
	tests := []struct {
		name       string
		cfg        config.Config
		method     string
		wantStatus int
		wantName   string
		wantIcon   string
	}{
		{
			name:       "english default",
			cfg:        config.Config{Learning: config.LearningConfig{AppCode: "english", TargetLang: "en"}},
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
			wantName:   "Qantrix English",
			wantIcon:   "/app/icons/english-512.png",
		},
		{
			name:       "spanish by app code",
			cfg:        config.Config{Learning: config.LearningConfig{AppCode: "spanish", TargetLang: "en"}},
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
			wantName:   "Qantrix Spanish",
			wantIcon:   "/app/icons/spanish-512.png",
		},
		{
			name:       "spanish by target lang",
			cfg:        config.Config{Learning: config.LearningConfig{AppCode: "english", TargetLang: "es"}},
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
			wantName:   "Qantrix Spanish",
			wantIcon:   "/app/icons/spanish-512.png",
		},
		{
			name:       "method not allowed",
			cfg:        config.Config{Learning: config.LearningConfig{AppCode: "english"}},
			method:     http.MethodPost,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gapRouter(&tt.cfg)
			req := httptest.NewRequest(tt.method, "/app/manifest.webmanifest", nil)
			w := httptest.NewRecorder()
			router.handleWebAppManifest(w, req)
			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantStatus != http.StatusOK {
				return
			}
			var body map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode manifest: %v", err)
			}
			if body["name"] != tt.wantName {
				t.Fatalf("name = %v, want %q", body["name"], tt.wantName)
			}
			icons, ok := body["icons"].([]interface{})
			if !ok || len(icons) == 0 {
				t.Fatalf("icons = %v", body["icons"])
			}
			icon0, ok := icons[0].(map[string]interface{})
			if !ok || icon0["src"] != tt.wantIcon {
				t.Fatalf("icon src = %v, want %q", icon0["src"], tt.wantIcon)
			}
		})
	}
}

func TestRouterGapHandleAndroidAssetLinks(t *testing.T) {
	t.Run("get with fingerprints", func(t *testing.T) {
		cfg := &config.Config{
			WebApp: config.WebAppConfig{
				AndroidCertFingerprints: " AA:BB , , CC:DD ",
			},
		}
		router := gapRouter(cfg)
		req := httptest.NewRequest(http.MethodGet, "/.well-known/assetlinks.json", nil)
		w := httptest.NewRecorder()
		router.handleAndroidAssetLinks(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		var statements []map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &statements); err != nil {
			t.Fatalf("decode assetlinks: %v", err)
		}
		if len(statements) != 3 {
			t.Fatalf("expected 3 package statements, got %d", len(statements))
		}
		target, ok := statements[0]["target"].(map[string]interface{})
		if !ok {
			t.Fatal("missing target")
		}
		fps, ok := target["sha256_cert_fingerprints"].([]interface{})
		if !ok || len(fps) != 2 {
			t.Fatalf("fingerprints = %v", target["sha256_cert_fingerprints"])
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		router := gapRouter(nil)
		req := httptest.NewRequest(http.MethodPut, "/.well-known/assetlinks.json", nil)
		w := httptest.NewRecorder()
		router.handleAndroidAssetLinks(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want 405", w.Code)
		}
	})
}

func TestRouterGapSetupWebappRoutes(t *testing.T) {
	savedFS := webappFS
	defer func() { webappFS = savedFS }()

	t.Run("registers manifest and assetlinks handlers", func(t *testing.T) {
		webappFS = fstest.MapFS{}
		router := gapRouter(&config.Config{
			Learning: config.LearningConfig{AppCode: "english"},
			WebApp:   config.WebAppConfig{AndroidCertFingerprints: "FP1"},
		})
		router.setupWebappRoutes()

		req := httptest.NewRequest(http.MethodGet, "/app/manifest.webmanifest", nil)
		w := httptest.NewRecorder()
		router.mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("manifest status = %d", w.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/.well-known/assetlinks.json", nil)
		w = httptest.NewRecorder()
		router.mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("assetlinks status = %d", w.Code)
		}
	})

	t.Run("embedded sw.js success and method guard", func(t *testing.T) {
		webappFS = fstest.MapFS{
			"dist/index.html": &fstest.MapFile{Data: []byte("<html>ok</html>")},
			"dist/sw.js":      &fstest.MapFile{Data: []byte("// service worker")},
		}
		router := gapRouter(nil)
		router.setupWebappRoutes()

		req := httptest.NewRequest(http.MethodGet, "/sw.js", nil)
		w := httptest.NewRecorder()
		router.mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("GET /sw.js status = %d", w.Code)
		}
		if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "javascript") {
			t.Fatalf("Content-Type = %q", ct)
		}
		if !strings.Contains(w.Body.String(), "service worker") {
			t.Fatalf("body = %q", w.Body.String())
		}

		req = httptest.NewRequest(http.MethodPost, "/sw.js", nil)
		w = httptest.NewRecorder()
		router.mux.ServeHTTP(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("POST /sw.js status = %d, want 405", w.Code)
		}
	})

	t.Run("embedded sw.js missing file", func(t *testing.T) {
		webappFS = fstest.MapFS{
			"dist/index.html": &fstest.MapFile{Data: []byte("<html>ok</html>")},
		}
		router := gapRouter(nil)
		router.setupWebappRoutes()

		req := httptest.NewRequest(http.MethodGet, "/sw.js", nil)
		w := httptest.NewRecorder()
		router.mux.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("GET /sw.js status = %d, want 404", w.Code)
		}
	})

	t.Run("admin entry missing admin.html", func(t *testing.T) {
		webappFS = fstest.MapFS{
			"dist/index.html": &fstest.MapFile{Data: []byte("<html>index</html>")},
		}
		router := gapRouter(nil)
		router.setupWebappRoutes()

		req := httptest.NewRequest(http.MethodGet, "/app/admin", nil)
		w := httptest.NewRecorder()
		router.mux.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("/app/admin status = %d, want 404", w.Code)
		}
	})

	t.Run("embedded assets and icons routes", func(t *testing.T) {
		webappFS = fstest.MapFS{
			"dist/index.html":      &fstest.MapFile{Data: []byte("<html>index</html>")},
			"dist/admin.html":      &fstest.MapFile{Data: []byte("<html>admin</html>")},
			"dist/assets/main.js":  &fstest.MapFile{Data: []byte("console.log('x')")},
			"dist/icons/icon.png":  &fstest.MapFile{Data: []byte{1, 2, 3}},
		}
		router := gapRouter(nil)
		router.setupWebappRoutes()

		for _, path := range []string{"/app/assets/main.js", "/app/icons/icon.png"} {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			router.mux.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("%s status = %d", path, w.Code)
			}
		}

		req := httptest.NewRequest(http.MethodGet, "/app", nil)
		w := httptest.NewRecorder()
		router.mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "index") {
			t.Fatalf("/app status = %d body = %q", w.Code, w.Body.String())
		}

		req = httptest.NewRequest(http.MethodGet, "/app/admin/users", nil)
		w = httptest.NewRecorder()
		router.mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "admin") {
			t.Fatalf("/app/admin/users status = %d body = %q", w.Code, w.Body.String())
		}
	})

	t.Run("embedded index read failure on /app", func(t *testing.T) {
		webappFS = &countingIndexFS{
			inner: fstest.MapFS{
				"dist/index.html": &fstest.MapFile{Data: []byte("<html>index</html>")},
			},
			failAfter: 1,
			count:     new(int),
		}
		router := gapRouter(nil)
		router.setupWebappRoutes()

		req := httptest.NewRequest(http.MethodGet, "/app", nil)
		w := httptest.NewRecorder()
		router.mux.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("/app status = %d, want 404", w.Code)
		}
	})
}
