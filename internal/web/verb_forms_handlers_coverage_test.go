package web

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupVerbFormsCoverageRouter(t *testing.T, db *sql.DB, training config.TrainingConfig, telegramID int64) (*Router, *repository.UserRepository, int64) {
	t.Helper()
	if telegramID == 0 {
		telegramID = 900001
	}
	logger := zap.NewNop()
	cfg := &config.Config{
		Learning: config.LearningConfig{
			TargetLang:      "es",
			NativeLang:      "ru",
			GrammarBundleID: "es",
		},
		Training: training,
		WebApp:   config.WebAppConfig{JWTSecret: "verb-forms-coverage-secret"},
	}
	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	userRepo := repository.NewUserRepository(db, logger)
	user, err := userRepo.GetOrCreateUser(telegramID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	return router, userRepo, user.ID
}

type verbCoverageSeed struct {
	lemma           string
	cardType        string
	promptJSON      string
	answerJSON      string
	distractorsJSON string
	srsReps         int
	srsState        string
}

func seedVerbFormsCoverageCard(t *testing.T, db *sql.DB, userID int64, opts ...func(*verbCoverageSeed)) (wordCardID, uvcID int64) {
	t.Helper()
	s := &verbCoverageSeed{
		lemma:           "hablar",
		cardType:        models.VerbCardTypeCloze,
		promptJSON:      `{"lemma":"hablar","mood":"indicativo","tense":"presente","person":"1","number":"singular","ru_gloss":"говорить","example_translation":"Я говорю."}`,
		answerJSON:      `{"surface_form":"hablo"}`,
		distractorsJSON: `["hablas","habla","hablamos"]`,
		srsReps:         0,
		srsState:        "new",
	}
	for _, fn := range opts {
		fn(s)
	}
	repo := repository.NewVerbFormsRepository(db, zap.NewNop())
	var wcID int64
	if err := db.QueryRow(`INSERT INTO word_cards (word, definition, pos) VALUES ($1, 'говорить', 'verb') RETURNING id`, s.lemma).
		Scan(&wcID); err != nil {
		t.Fatalf("insert word_card: %v", err)
	}
	var trainingCardID int64
	if err := db.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos)
		VALUES ($1, $2, 0, 'говорить', 'to speak', 'verb') RETURNING id`, wcID, s.lemma).Scan(&trainingCardID); err != nil {
		t.Fatalf("insert training_card: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state) VALUES ($1, $2, 'es_ru', 'review')`,
		userID, trainingCardID); err != nil {
		t.Fatalf("insert user_card: %v", err)
	}
	lemmaID, err := repo.UpsertVerbLemma(s.lemma, "es", "test", "v1", "chk", `{"ru":{"gloss":"говорить"}}`)
	if err != nil {
		t.Fatalf("UpsertVerbLemma: %v", err)
	}
	formID, err := repo.UpsertVerbForm(&models.VerbFormDict{
		VerbLemmaID: lemmaID, Mood: "indicativo", Tense: "presente",
		Person: "1", Number: "singular", SurfaceForm: "hablo",
	})
	if err != nil {
		t.Fatalf("UpsertVerbForm: %v", err)
	}
	if err := repo.LinkWordCardToLemma(wcID, lemmaID, 1.0, "test"); err != nil {
		t.Fatalf("LinkWordCardToLemma: %v", err)
	}
	vtcID, err := repo.UpsertVerbTrainingCard(&models.VerbTrainingCard{
		WordCardID: wcID, VerbFormDictID: formID, CardType: s.cardType,
		PromptJSON: s.promptJSON, AnswerJSON: s.answerJSON, DistractorsJSON: s.distractorsJSON,
	})
	if err != nil {
		t.Fatalf("UpsertVerbTrainingCard: %v", err)
	}
	uvc, err := repo.GetOrCreateUserVerbCard(userID, vtcID)
	if err != nil {
		t.Fatalf("GetOrCreateUserVerbCard: %v", err)
	}
	if _, err := db.Exec(`UPDATE user_verb_cards SET reps=$1, state=$2 WHERE id=$3`, s.srsReps, s.srsState, uvc); err != nil {
		t.Fatalf("update srs: %v", err)
	}
	return wcID, uvc
}

func seedPendingVerbLemma(t *testing.T, db *sql.DB, lemma string) int64 {
	t.Helper()
	repo := repository.NewVerbFormsRepository(db, zap.NewNop())
	var wordCardID int64
	if err := db.QueryRow(`INSERT INTO word_cards (word, definition, pos) VALUES ($1, 'def', 'verb') RETURNING id`, lemma).
		Scan(&wordCardID); err != nil {
		t.Fatalf("insert word_card: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos)
		VALUES ($1, $2, 0, 'def', 'def', 'verb')`, wordCardID, lemma); err != nil {
		t.Fatalf("insert training_card: %v", err)
	}
	lemmaID, err := repo.UpsertVerbLemma(lemma, "es", "test", "v1", "chk", `{}`)
	if err != nil {
		t.Fatalf("UpsertVerbLemma: %v", err)
	}
	if err := repo.LinkWordCardToLemma(wordCardID, lemmaID, 1.0, "test"); err != nil {
		t.Fatalf("LinkWordCardToLemma: %v", err)
	}
	return wordCardID
}

func putVerbSession(t *testing.T, userID int64, state *webVerbTrainingState) {
	t.Helper()
	webVerbSessionsMu.Lock()
	webVerbSessions[userID] = state
	webVerbSessionsMu.Unlock()
	t.Cleanup(func() {
		webVerbSessionsMu.Lock()
		delete(webVerbSessions, userID)
		webVerbSessionsMu.Unlock()
	})
}

func TestVerbFormsHandlersCoverage(t *testing.T) {
	t.Run("nilRouterGuards", func(t *testing.T) {
		var nilRouter *Router
		if nilRouter.verbFormsEnabled() {
			t.Fatal("nil router must not enable verb forms")
		}
		if nilRouter.verbFormsEnabledForUser(context.Background(), 1) {
			t.Fatal("nil router must not enable verb forms for user")
		}
		r := &Router{}
		if r.verbFormsEnabled() {
			t.Fatal("router without config must not enable verb forms")
		}
		okRouter := NewRouter(zap.NewNop(), &config.Config{
			Learning: config.LearningConfig{TargetLang: "es"},
			Training: config.TrainingConfig{SpanishVerbFormsEnabled: true},
		}, nil, nil, nil, nil, nil)
		if !okRouter.verbFormsEnabled() {
			t.Fatal("expected verb forms enabled for es config")
		}
	})

	t.Run("helperFunctions", func(t *testing.T) {
		if !isVerbClozeCardType(" cloze_form ") {
			t.Fatal("expected cloze card type")
		}
		if isVerbClozeCardType("form_recall") {
			t.Fatal("recall is not cloze")
		}

		if promptString(nil, "k") != "" {
			t.Fatal("nil map")
		}
		m := map[string]interface{}{
			"s": " hello ", "f": float64(3), "n": json.Number("7"), "b": true, "nil": nil,
		}
		if promptString(m, "s") != "hello" {
			t.Fatalf("string: %q", promptString(m, "s"))
		}
		if promptString(m, "f") != "3" {
			t.Fatalf("float64: %q", promptString(m, "f"))
		}
		if promptString(m, "n") != "7" {
			t.Fatalf("json.Number: %q", promptString(m, "n"))
		}
		if promptString(m, "b") != "true" {
			t.Fatalf("default: %q", promptString(m, "b"))
		}
		if promptString(m, "nil") != "" || promptString(m, "missing") != "" {
			t.Fatal("nil/missing key")
		}

		if stringSliceFromPrompt(nil, "x") != nil {
			t.Fatal("nil map slice")
		}
		ss := stringSliceFromPrompt(map[string]interface{}{
			"a": []string{" one ", "", "two"},
			"b": []interface{}{" 3 ", 4, ""},
			"c": "nope",
		}, "a")
		if len(ss) != 2 || ss[0] != "one" || ss[1] != "two" {
			t.Fatalf("[]string: %v", ss)
		}
		si := stringSliceFromPrompt(map[string]interface{}{"b": []interface{}{"x", 1}}, "b")
		if len(si) != 2 || si[0] != "x" || si[1] != "1" {
			t.Fatalf("[]interface{}: %v", si)
		}
		if stringSliceFromPrompt(map[string]interface{}{"c": "nope"}, "c") != nil {
			t.Fatal("default type")
		}

		p1 := map[string]interface{}{"lemma": "hablar", "mood": "indicativo", "tense": "presente", "person": "1", "number": "singular", "ru_gloss": "говорить"}
		hydrateVerbClozePrompt(p1, "hablo", 42)
		if strings.TrimSpace(promptString(p1, "question")) == "" {
			t.Fatal("expected generated question")
		}
		p2 := map[string]interface{}{"question": "Yo ___ español.", "lemma": "hablar", "mood": "indicativo", "tense": "presente", "ru_gloss": "говорить"}
		hydrateVerbClozePrompt(p2, "hablo", 99)
		if strings.TrimSpace(promptString(p2, "example_translation")) == "" {
			t.Fatal("expected literary translation")
		}
		hydrateVerbClozePrompt(nil, "x", 1)

		ensureVerbClozeQuestionLine(nil)
		eq := map[string]interface{}{"lemma": "hablar", "person": "1", "number": "singular", "mood": "indicativo", "tense": "presente"}
		ensureVerbClozeQuestionLine(eq)
		if strings.TrimSpace(promptString(eq, "question")) == "" {
			t.Fatal("expected cloze question line")
		}
		ensureVerbClozeQuestionLine(map[string]interface{}{"question": "already"})
		ensureVerbClozeQuestionLine(map[string]interface{}{})

		if _, ok := parseWordCardIDForVerbForms("/api/vocab/42/other"); ok {
			t.Fatal("wrong suffix")
		}
		if _, ok := parseWordCardIDForVerbForms("/api/vocab/0/verb-forms"); ok {
			t.Fatal("zero id should fail")
		}
		if _, ok := parseWordCardIDForVerbForms("/api/vocab/-1/verb-forms"); ok {
			t.Fatal("negative id should fail")
		}
		if _, ok := parseWordCardIDForVerbForms("/api/other/42/verb-forms"); ok {
			t.Fatal("wrong prefix")
		}
	})

	t.Run("getUserVerbScopes", func(t *testing.T) {
		db := testutil.SetupTestDB(t)
		router, userRepo, userID := setupVerbFormsCoverageRouter(t, db, config.TrainingConfig{SpanishVerbFormsEnabled: true}, 900001)

		enRouter := NewRouter(zap.NewNop(), &config.Config{
			Learning: config.LearningConfig{TargetLang: "en", GrammarBundleID: "en"},
		}, db, nil, nil, nil, nil)
		if scopes := enRouter.getUserVerbScopes(context.Background(), userID); scopes != nil {
			t.Fatalf("en target should return nil, got %v", scopes)
		}

		noRepo := NewRouter(zap.NewNop(), &config.Config{
			Learning: config.LearningConfig{TargetLang: "es", GrammarBundleID: "no-such-bundle"},
		}, db, nil, nil, nil, nil)
		if scopes := noRepo.getUserVerbScopes(context.Background(), userID); len(scopes) == 0 {
			t.Fatal("nil userRepo should fall back to defaults")
		}

		wrongRepo := NewRouter(zap.NewNop(), &config.Config{
			Learning: config.LearningConfig{TargetLang: "es", GrammarBundleID: "no-such-bundle"},
		}, db, nil, nil, nil, nil)
		wrongRepo.userRepo = struct{}{}
		if scopes := wrongRepo.getUserVerbScopes(context.Background(), userID); len(scopes) == 0 {
			t.Fatal("wrong userRepo type should fall back to defaults")
		}

		if scopes := router.getUserVerbScopes(context.Background(), 999999999); len(scopes) == 0 {
			t.Fatal("missing user should fall back to defaults")
		}

		if err := userRepo.UpdateUserSettings(userID, `{bad`); err != nil {
			t.Fatalf("UpdateUserSettings: %v", err)
		}
		badJSONRouter := NewRouter(zap.NewNop(), &config.Config{
			Learning: config.LearningConfig{TargetLang: "es", GrammarBundleID: "no-such-bundle"},
		}, db, nil, nil, nil, nil)
		badJSONRouter.userRepo = userRepo
		if scopes := badJSONRouter.getUserVerbScopes(context.Background(), userID); len(scopes) == 0 {
			t.Fatal("bad settings JSON should fall back to defaults")
		}

		settings := models.UserSettings{EnabledVerbScopes: []string{"es.presente.indicativo", "es.grammar.foo"}}
		settingsJSON, _ := json.Marshal(settings)
		if err := userRepo.UpdateUserSettings(userID, string(settingsJSON)); err != nil {
			t.Fatalf("UpdateUserSettings: %v", err)
		}
		settingsRouter := NewRouter(zap.NewNop(), &config.Config{
			Learning: config.LearningConfig{TargetLang: "es", GrammarBundleID: "no-such-bundle"},
		}, db, nil, nil, nil, nil)
		settingsRouter.userRepo = userRepo
		scopes := settingsRouter.getUserVerbScopes(context.Background(), userID)
		found := false
		for _, s := range scopes {
			if s == "es.presente.indicativo" {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected settings scopes, got %v", scopes)
		}

		contentRepo, err := repository.NewGrammarContentRepositoryForLearning(config.LearningConfig{TargetLang: "es", GrammarBundleID: "es"}, zap.NewNop())
		if err != nil {
			t.Fatalf("content repo: %v", err)
		}
		publishRepo := repository.NewGrammarPublishRepository(db, zap.NewNop())
		attemptRepo := repository.NewGrammarAttemptRepository(db, zap.NewNop())
		gs := service.NewGrammarService(contentRepo, publishRepo, attemptRepo, config.LearningConfig{TargetLang: "es", GrammarBundleID: "es"}, zap.NewNop())
		router.SetGrammarService(gs)
		sectionID := "es.grammar.past_preterito_indefinido"
		chapterID := "es.grammar.past_preterito_indefinido.regular_indefinido_endings"
		if err := gs.PublishRepo.SetPublished("section", sectionID, true, nil); err != nil {
			t.Fatalf("publish section: %v", err)
		}
		if err := gs.PublishRepo.SetPublished("chapter", chapterID, true, nil); err != nil {
			t.Fatalf("publish chapter: %v", err)
		}
		if err := attemptRepo.SavePlacementTestResult(userID, 80, 10, []string{sectionID}); err != nil {
			t.Fatalf("placement: %v", err)
		}
		gateScopes := router.getUserVerbScopes(context.Background(), userID)
		hasPreterito := false
		for _, s := range gateScopes {
			if s == "es.preterito_indefinido.indicativo" {
				hasPreterito = true
			}
		}
		if !hasPreterito {
			t.Fatalf("expected chapter unlock scope, got %v", gateScopes)
		}

		gateRouter := NewRouter(zap.NewNop(), &config.Config{
			Learning: config.LearningConfig{TargetLang: "es", GrammarBundleID: "es"},
		}, db, nil, nil, nil, nil)
		if scopes := gateRouter.getUserVerbScopes(context.Background(), userID); len(scopes) == 0 {
			t.Fatal("gates without grammar service should still unlock defaults")
		}

		freshUser, err := userRepo.GetOrCreateUser(900024)
		if err != nil {
			t.Fatalf("fresh user: %v", err)
		}
		emptySettingsRouter := NewRouter(zap.NewNop(), &config.Config{
			Learning: config.LearningConfig{TargetLang: "es", GrammarBundleID: "no-such-bundle"},
		}, db, nil, nil, nil, nil)
		emptySettingsRouter.userRepo = userRepo
		if scopes := emptySettingsRouter.getUserVerbScopes(context.Background(), freshUser.ID); len(scopes) == 0 {
			t.Fatal("fresh user empty settings should use defaults")
		}

		badGS := service.NewGrammarService(contentRepo, publishRepo, repository.NewGrammarAttemptRepository(newBrokenDB(t), zap.NewNop()), config.LearningConfig{TargetLang: "es", GrammarBundleID: "es"}, zap.NewNop())
		router.SetGrammarService(badGS)
		if scopes := router.getUserVerbScopes(context.Background(), userID); len(scopes) == 0 {
			t.Fatal("CanAccessChapter errors should still yield always-unlocked scopes")
		}

		noPlaceUser, err := userRepo.GetOrCreateUser(900027)
		if err != nil {
			t.Fatalf("user: %v", err)
		}
		noPlaceRouter := NewRouter(zap.NewNop(), &config.Config{
			Learning: config.LearningConfig{TargetLang: "es", GrammarBundleID: "es"},
		}, db, nil, nil, nil, nil)
		noPlaceGS := service.NewGrammarService(contentRepo, publishRepo, repository.NewGrammarAttemptRepository(newBrokenDB(t), zap.NewNop()), config.LearningConfig{TargetLang: "es", GrammarBundleID: "es"}, zap.NewNop())
		noPlaceRouter.SetGrammarService(noPlaceGS)
		_ = noPlaceRouter.getUserVerbScopes(context.Background(), noPlaceUser.ID)
	})

	t.Run("ensureVerbFormUserCards", func(t *testing.T) {
		db := testutil.SetupTestDB(t)
		router, _, userID := setupVerbFormsCoverageRouter(t, db, config.TrainingConfig{SpanishVerbFormsEnabled: true}, 900002)
		seedVerbFormsCoverageCard(t, db, userID)
		router.ensureVerbFormUserCardsAfterVocab(userID)

		enRouter := NewRouter(zap.NewNop(), &config.Config{
			Learning: config.LearningConfig{TargetLang: "en"},
			Training: config.TrainingConfig{SpanishVerbFormsEnabled: true},
		}, db, nil, nil, nil, nil)
		enRouter.ensureVerbFormUserCardsForUser(context.Background(), userID)

		broken := NewRouter(zap.NewNop(), &config.Config{
			Learning: config.LearningConfig{TargetLang: "es", GrammarBundleID: "es"},
			Training: config.TrainingConfig{SpanishVerbFormsEnabled: true},
		}, newBrokenDB(t), nil, nil, nil, nil)
		broken.ensureVerbFormUserCardsForUser(context.Background(), userID)
	})

	t.Run("lemmaFormsHandler", func(t *testing.T) {
		db := testutil.SetupTestDB(t)
		router, _, userID := setupVerbFormsCoverageRouter(t, db, config.TrainingConfig{SpanishVerbFormsEnabled: true}, 900003)
		seedVerbFormsCoverageCard(t, db, userID)

		cases := []struct {
			name   string
			method string
			url    string
			uid    int64
			code   int
		}{
			{"method", http.MethodPost, "/api/verb-training/forms-by-lemma?lemma=hablar", userID, http.StatusMethodNotAllowed},
			{"unauth", http.MethodGet, "/api/verb-training/forms-by-lemma?lemma=hablar", 0, http.StatusUnauthorized},
			{"missing", http.MethodGet, "/api/verb-training/forms-by-lemma", userID, http.StatusBadRequest},
			{"notfound", http.MethodGet, "/api/verb-training/forms-by-lemma?lemma=zzzunknown", userID, http.StatusNotFound},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				req := httptest.NewRequest(tc.method, tc.url, nil)
				req = verbFormsUserContext(req, tc.uid)
				rr := httptest.NewRecorder()
				router.handleVerbTrainingLemmaForms(rr, req)
				if rr.Code != tc.code {
					t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
				}
			})
		}

		enRouter := NewRouter(zap.NewNop(), &config.Config{
			Learning: config.LearningConfig{TargetLang: "en"},
			Training: config.TrainingConfig{SpanishVerbFormsEnabled: true},
		}, db, nil, nil, nil, nil)
		req := httptest.NewRequest(http.MethodGet, "/api/verb-training/forms-by-lemma?lemma=hablar", nil)
		req = verbFormsUserContext(req, userID)
		rr := httptest.NewRecorder()
		enRouter.handleVerbTrainingLemmaForms(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("disabled status=%d", rr.Code)
		}

		broken := NewRouter(zap.NewNop(), &config.Config{
			Learning: config.LearningConfig{TargetLang: "es", GrammarBundleID: "es"},
			Training: config.TrainingConfig{SpanishVerbFormsEnabled: true},
		}, newBrokenDB(t), nil, nil, nil, nil)
		req2 := httptest.NewRequest(http.MethodGet, "/api/verb-training/forms-by-lemma?lemma=hablar", nil)
		req2 = verbFormsUserContext(req2, userID)
		rr2 := httptest.NewRecorder()
		broken.handleVerbTrainingLemmaForms(rr2, req2)
		if rr2.Code != http.StatusInternalServerError {
			t.Fatalf("broken db status=%d", rr2.Code)
		}
	})

	t.Run("vocabVerbFormsHandler", func(t *testing.T) {
		db := testutil.SetupTestDB(t)
		router, _, userID := setupVerbFormsCoverageRouter(t, db, config.TrainingConfig{SpanishVerbFormsEnabled: true}, 900004)
		wordCardID, _ := seedVerbFormsCoverageCard(t, db, userID)

		req := httptest.NewRequest(http.MethodPost, "/api/vocab/x/verb-forms", nil)
		rr := httptest.NewRecorder()
		router.handleVocabVerbForms(rr, req, userID, wordCardID)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("method status=%d", rr.Code)
		}

		enRouter := NewRouter(zap.NewNop(), &config.Config{
			Learning: config.LearningConfig{TargetLang: "en"},
			Training: config.TrainingConfig{SpanishVerbFormsEnabled: true},
		}, db, nil, nil, nil, nil)
		req2 := httptest.NewRequest(http.MethodGet, "/api/vocab/x/verb-forms", nil)
		rr2 := httptest.NewRecorder()
		enRouter.handleVocabVerbForms(rr2, req2, userID, wordCardID)
		if rr2.Code != http.StatusNotFound {
			t.Fatalf("disabled status=%d", rr2.Code)
		}

		broken := NewRouter(zap.NewNop(), &config.Config{
			Learning: config.LearningConfig{TargetLang: "es"},
			Training: config.TrainingConfig{SpanishVerbFormsEnabled: true},
		}, newBrokenDB(t), nil, nil, nil, nil)
		req3 := httptest.NewRequest(http.MethodGet, "/api/vocab/x/verb-forms", nil)
		rr3 := httptest.NewRecorder()
		broken.handleVocabVerbForms(rr3, req3, userID, wordCardID)
		if rr3.Code != http.StatusInternalServerError {
			t.Fatalf("broken db status=%d", rr3.Code)
		}
	})

	t.Run("trainingStartAndCurrent", func(t *testing.T) {
		db := testutil.SetupTestDB(t)
		training := config.TrainingConfig{
			SpanishVerbFormsEnabled:     true,
			VerbFormsMaxCards:           5,
			VerbFormsMaxNew:             3,
			VerbFormsTypedMinReps:       0,
			VerbFormsTypedChancePercent: 100,
		}
		router, _, userID := setupVerbFormsCoverageRouter(t, db, training, 900005)
		_, uvcID := seedVerbFormsCoverageCard(t, db, userID, func(s *verbCoverageSeed) {
			s.srsReps = 3
			s.srsState = "review"
			s.distractorsJSON = `["hablas","habla"]`
		})

		req := httptest.NewRequest(http.MethodGet, "/api/verb-training/start", nil)
		req = verbFormsUserContext(req, userID)
		rr := httptest.NewRecorder()
		router.handleVerbTrainingStart(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("method status=%d", rr.Code)
		}

		reqUnauth := httptest.NewRequest(http.MethodPost, "/api/verb-training/start", nil)
		rrUnauth := httptest.NewRecorder()
		router.handleVerbTrainingStart(rrUnauth, reqUnauth)
		if rrUnauth.Code != http.StatusUnauthorized {
			t.Fatalf("unauth status=%d", rrUnauth.Code)
		}

		disabled := NewRouter(zap.NewNop(), &config.Config{
			Learning: config.LearningConfig{TargetLang: "es"},
			Training: config.TrainingConfig{SpanishVerbFormsEnabled: false},
		}, db, nil, nil, nil, nil)
		reqDis := httptest.NewRequest(http.MethodPost, "/api/verb-training/start", nil)
		reqDis = verbFormsUserContext(reqDis, userID)
		rrDis := httptest.NewRecorder()
		disabled.handleVerbTrainingStart(rrDis, reqDis)
		if rrDis.Code != http.StatusForbidden {
			t.Fatalf("disabled status=%d", rrDis.Code)
		}

		emptyUserRouter, _, emptyUID := setupVerbFormsCoverageRouter(t, db, training, 900006)
		reqNoCards := httptest.NewRequest(http.MethodPost, "/api/verb-training/start", nil)
		reqNoCards = verbFormsUserContext(reqNoCards, emptyUID)
		rrNoCards := httptest.NewRecorder()
		emptyUserRouter.handleVerbTrainingStart(rrNoCards, reqNoCards)
		if rrNoCards.Code != http.StatusBadRequest {
			t.Fatalf("no cards status=%d body=%s", rrNoCards.Code, rrNoCards.Body.String())
		}

		startReq := httptest.NewRequest(http.MethodPost, "/api/verb-training/start", nil)
		startReq = verbFormsUserContext(startReq, userID)
		startRR := httptest.NewRecorder()
		router.handleVerbTrainingStart(startRR, startReq)
		if startRR.Code != http.StatusOK {
			t.Fatalf("start status=%d body=%s", startRR.Code, startRR.Body.String())
		}
		var card map[string]interface{}
		if err := json.Unmarshal(startRR.Body.Bytes(), &card); err != nil {
			t.Fatal(err)
		}
		if card["input_mode"] != "typed" {
			t.Fatalf("expected typed mode, got %v", card["input_mode"])
		}

		curReq := httptest.NewRequest(http.MethodGet, "/api/verb-training/current", nil)
		curReq = verbFormsUserContext(curReq, userID)
		curRR := httptest.NewRecorder()
		router.handleVerbTrainingCurrent(curRR, curReq)
		if curRR.Code != http.StatusOK {
			t.Fatalf("current status=%d", curRR.Code)
		}

		curBadMethod := httptest.NewRequest(http.MethodPost, "/api/verb-training/current", nil)
		curBadMethod = verbFormsUserContext(curBadMethod, userID)
		curBadRR := httptest.NewRecorder()
		router.handleVerbTrainingCurrent(curBadRR, curBadMethod)
		if curBadRR.Code != http.StatusMethodNotAllowed {
			t.Fatalf("current method status=%d", curBadRR.Code)
		}

		curUnauth := httptest.NewRequest(http.MethodGet, "/api/verb-training/current", nil)
		curUnauthRR := httptest.NewRecorder()
		router.handleVerbTrainingCurrent(curUnauthRR, curUnauth)
		if curUnauthRR.Code != http.StatusUnauthorized {
			t.Fatalf("current unauth status=%d", curUnauthRR.Code)
		}

		noSessUID := int64(900021)
		noSessReq := httptest.NewRequest(http.MethodGet, "/api/verb-training/current", nil)
		noSessReq = verbFormsUserContext(noSessReq, noSessUID)
		noSessRR := httptest.NewRecorder()
		router.handleVerbTrainingCurrent(noSessRR, noSessReq)
		if noSessRR.Code != http.StatusNotFound {
			t.Fatalf("current no session status=%d", noSessRR.Code)
		}

		learningTypedRouter, _, learningTypedUID := setupVerbFormsCoverageRouter(t, db, config.TrainingConfig{
			SpanishVerbFormsEnabled: true, VerbFormsMaxCards: 5, VerbFormsMaxNew: 3,
			VerbFormsTypedChancePercent: 100,
		}, 900022)
		_, learningUVC := seedVerbFormsCoverageCard(t, db, learningTypedUID, func(s *verbCoverageSeed) {
			s.lemma = "salir"
			s.srsReps = 5
			s.srsState = "learning"
			s.distractorsJSON = `["a","b"]`
		})
		learningState := &webVerbTrainingState{
			UserID: learningTypedUID, SessionID: 50, Index: 0,
			Queue: []repository.VerbQueueCard{{
				UserVerbCardID: learningUVC, CardType: models.VerbCardTypeCloze,
				PromptJSON: `{"lemma":"salir","mood":"indicativo","tense":"presente","person":"1","number":"singular","question":"Yo ___ ."}`,
				AnswerJSON: `{"surface_form":"salgo"}`, DistractorsJSON: `["a","b"]`,
			}},
		}
		learningRR := httptest.NewRecorder()
		learningTypedRouter.writeCurrentVerbCard(learningRR, learningState)
		if learningRR.Code != http.StatusOK {
			t.Fatalf("learning srs card status=%d", learningRR.Code)
		}

		negChanceRouter, _, negUID := setupVerbFormsCoverageRouter(t, db, config.TrainingConfig{
			SpanishVerbFormsEnabled: true, VerbFormsTypedChancePercent: -3,
		}, 900025)
		_, negUVC := seedVerbFormsCoverageCard(t, db, negUID, func(s *verbCoverageSeed) {
			s.lemma = "poner"
			s.srsReps = 4
			s.srsState = "review"
			s.distractorsJSON = ``
		})
		negState := &webVerbTrainingState{
			UserID: negUID, SessionID: 51, Index: 0,
			Queue: []repository.VerbQueueCard{{
				UserVerbCardID: negUVC, CardType: models.VerbCardTypeCloze,
				PromptJSON: `{"lemma":"poner"}`, AnswerJSON: `{"surface_form":"pongo"}`, DistractorsJSON: ``,
			}},
		}
		negRR := httptest.NewRecorder()
		negChanceRouter.writeCurrentVerbCard(negRR, negState)
		if negRR.Code != http.StatusOK {
			t.Fatalf("neg chance card status=%d", negRR.Code)
		}

		brokenSRSRouter, _, brokenSRSUID := setupVerbFormsCoverageRouter(t, db, config.TrainingConfig{SpanishVerbFormsEnabled: true}, 900026)
		_, brokenSRSUVC := seedVerbFormsCoverageCard(t, db, brokenSRSUID, func(s *verbCoverageSeed) { s.lemma = "venir" })
		brokenSRSState := &webVerbTrainingState{
			UserID: brokenSRSUID, SessionID: 52, Index: 0,
			Queue: []repository.VerbQueueCard{{
				UserVerbCardID: brokenSRSUVC, CardType: models.VerbCardTypeCloze,
				PromptJSON: `{"lemma":"venir"}`, AnswerJSON: `{"surface_form":"vengo"}`, DistractorsJSON: `["a","b"]`,
			}},
		}
		brokenSRSRouter.db = newBrokenDB(t)
		brokenSRSRR := httptest.NewRecorder()
		brokenSRSRouter.writeCurrentVerbCard(brokenSRSRR, brokenSRSState)
		if brokenSRSRR.Code != http.StatusOK {
			t.Fatalf("broken srs card status=%d", brokenSRSRR.Code)
		}

		finishedState := &webVerbTrainingState{
			UserID: userID, SessionID: 1, Index: 1,
			Queue: []repository.VerbQueueCard{{UserVerbCardID: uvcID, CardType: models.VerbCardTypeCloze, PromptJSON: `{}`, AnswerJSON: `{"surface_form":"hablo"}`}},
		}
		putVerbSession(t, userID, finishedState)
		finRR := httptest.NewRecorder()
		router.writeCurrentVerbCard(finRR, finishedState)
		if finRR.Code != http.StatusOK {
			t.Fatalf("finish card status=%d body=%s", finRR.Code, finRR.Body.String())
		}
		var finBody map[string]interface{}
		if err := json.Unmarshal(finRR.Body.Bytes(), &finBody); err != nil {
			t.Fatal(err)
		}
		if finBody["finished"] != true {
			t.Fatalf("expected finished response: %v", finBody)
		}

		okFinishRouter, _, okFinishUID := setupVerbFormsCoverageRouter(t, db, config.TrainingConfig{SpanishVerbFormsEnabled: true}, 900020)
		_, okUVC := seedVerbFormsCoverageCard(t, db, okFinishUID, func(s *verbCoverageSeed) { s.lemma = "leer" })
		okRepo := repository.NewVerbFormsRepository(db, zap.NewNop())
		okSessID, _ := okRepo.StartVerbSession(okFinishUID, 1, `{}`)
		_ = okRepo.CreateVerbReviewEvent(okSessID, okFinishUID, okUVC, true, 5)
		okFinishState := &webVerbTrainingState{UserID: okFinishUID, SessionID: okSessID, Index: 1, Queue: []repository.VerbQueueCard{{}}}
		okFinishRR := httptest.NewRecorder()
		okFinishRouter.finishVerbTrainingSessionResponse(okFinishRR, okFinishState)
		if okFinishRR.Code != http.StatusOK {
			t.Fatalf("ok finish status=%d", okFinishRR.Code)
		}

		typedFallbackState := &webVerbTrainingState{
			UserID: userID, SessionID: 3, Index: 0,
			Queue: []repository.VerbQueueCard{{
				UserVerbCardID: uvcID, CardType: models.VerbCardTypeCloze,
				PromptJSON: `{bad`, AnswerJSON: `{"surface_form":"hablo"}`,
				DistractorsJSON: `["onlyone"]`,
			}},
		}
		fallbackRR := httptest.NewRecorder()
		router.writeCurrentVerbCard(fallbackRR, typedFallbackState)
		if fallbackRR.Code != http.StatusOK {
			t.Fatalf("fallback card status=%d", fallbackRR.Code)
		}

		choiceRouter, _, choiceUID := setupVerbFormsCoverageRouter(t, db, config.TrainingConfig{
			SpanishVerbFormsEnabled:     true,
			VerbFormsMaxCards:           5,
			VerbFormsMaxNew:             3,
			VerbFormsTypedChancePercent: 0,
		}, 900007)
		seedVerbFormsCoverageCard(t, db, choiceUID, func(s *verbCoverageSeed) { s.lemma = "comer" })
		startChoice := httptest.NewRequest(http.MethodPost, "/api/verb-training/start", nil)
		startChoice = verbFormsUserContext(startChoice, choiceUID)
		choiceRR := httptest.NewRecorder()
		choiceRouter.handleVerbTrainingStart(choiceRR, startChoice)
		if choiceRR.Code != http.StatusOK {
			t.Fatalf("choice start status=%d", choiceRR.Code)
		}
		var choiceCard map[string]interface{}
		_ = json.Unmarshal(choiceRR.Body.Bytes(), &choiceCard)
		if choiceCard["input_mode"] != "choice" {
			t.Fatalf("expected choice mode, got %v", choiceCard["input_mode"])
		}

		recallRouter, _, recallUID := setupVerbFormsCoverageRouter(t, db, config.TrainingConfig{
			SpanishVerbFormsEnabled: true, VerbFormsMaxCards: 5, VerbFormsMaxNew: 3,
		}, 900008)
		_, recallUVC := seedVerbFormsCoverageCard(t, db, recallUID, func(s *verbCoverageSeed) {
			s.lemma = "deber"
			s.cardType = models.VerbCardTypeRecall
			s.promptJSON = `{"lemma":"deber"}`
			s.distractorsJSON = ``
		})
		recallState := &webVerbTrainingState{
			UserID: recallUID, SessionID: 99, Index: 0,
			Queue: []repository.VerbQueueCard{{
				UserVerbCardID: recallUVC, CardType: models.VerbCardTypeRecall,
				PromptJSON: `{"lemma":"deber"}`, AnswerJSON: `{"surface_form":"debo"}`,
				DistractorsJSON: `[]`,
			}},
		}
		putVerbSession(t, recallUID, recallState)
		recallRR := httptest.NewRecorder()
		recallRouter.writeCurrentVerbCard(recallRR, recallState)
		if recallRR.Code != http.StatusOK {
			t.Fatalf("recall card status=%d", recallRR.Code)
		}

		badSRSState := &webVerbTrainingState{
			UserID: userID, SessionID: 2, Index: 0,
			Queue: []repository.VerbQueueCard{{
				UserVerbCardID: 999999999, CardType: models.VerbCardTypeCloze,
				PromptJSON: `{"lemma":"hablar","mood":"indicativo","tense":"presente","person":"1","number":"singular"}`,
				AnswerJSON: `{"surface_form":"hablo"}`, DistractorsJSON: `["a","b"]`,
			}},
		}
		srsRR := httptest.NewRecorder()
		router.writeCurrentVerbCard(srsRR, badSRSState)
		if srsRR.Code != http.StatusOK {
			t.Fatalf("bad srs status=%d", srsRR.Code)
		}

		clampRouter, _, clampUID := setupVerbFormsCoverageRouter(t, db, config.TrainingConfig{
			SpanishVerbFormsEnabled: true, VerbFormsMaxCards: 5, VerbFormsMaxNew: 3,
			VerbFormsTypedMinReps: -1, VerbFormsTypedChancePercent: 150,
		}, 900009)
		seedVerbFormsCoverageCard(t, db, clampUID, func(s *verbCoverageSeed) {
			s.lemma = "beber"
			s.srsReps = 5
			s.srsState = "review"
			s.distractorsJSON = `["x"]`
		})
		clampStart := httptest.NewRequest(http.MethodPost, "/api/verb-training/start", nil)
		clampStart = verbFormsUserContext(clampStart, clampUID)
		clampRR := httptest.NewRecorder()
		clampRouter.handleVerbTrainingStart(clampRR, clampStart)
		if clampRR.Code != http.StatusOK {
			t.Fatalf("clamp start status=%d", clampRR.Code)
		}

		brokenFinish := NewRouter(zap.NewNop(), &config.Config{
			Learning: config.LearningConfig{TargetLang: "es"},
			Training: config.TrainingConfig{SpanishVerbFormsEnabled: true},
		}, newBrokenDB(t), nil, nil, nil, nil)
		finishState := &webVerbTrainingState{UserID: 900009, SessionID: 404, Index: 1, Queue: []repository.VerbQueueCard{{}}}
		finishRR2 := httptest.NewRecorder()
		brokenFinish.finishVerbTrainingSessionResponse(finishRR2, finishState)
		if finishRR2.Code != http.StatusOK {
			t.Fatalf("broken finish status=%d", finishRR2.Code)
		}
	})

	t.Run("trainingAnswer", func(t *testing.T) {
		db := testutil.SetupTestDB(t)
		router, _, userID := setupVerbFormsCoverageRouter(t, db, config.TrainingConfig{
			SpanishVerbFormsEnabled: true, VerbFormsMaxCards: 5, VerbFormsMaxNew: 3,
			WrongAnswerDelaySeconds: 4,
		}, 900010)
		_, uvcID := seedVerbFormsCoverageCard(t, db, userID)

		vrepo := repository.NewVerbFormsRepository(db, zap.NewNop())
		sessionID, err := vrepo.StartVerbSession(userID, 1, `{}`)
		if err != nil {
			t.Fatalf("StartVerbSession: %v", err)
		}

		ansMethod := httptest.NewRequest(http.MethodGet, "/api/verb-training/answer", nil)
		ansMethod = verbFormsUserContext(ansMethod, userID)
		ansMethodRR := httptest.NewRecorder()
		router.handleVerbTrainingAnswer(ansMethodRR, ansMethod)
		if ansMethodRR.Code != http.StatusMethodNotAllowed {
			t.Fatalf("method status=%d", ansMethodRR.Code)
		}

		ansUnauth := httptest.NewRequest(http.MethodPost, "/api/verb-training/answer", nil)
		ansUnauthRR := httptest.NewRecorder()
		router.handleVerbTrainingAnswer(ansUnauthRR, ansUnauth)
		if ansUnauthRR.Code != http.StatusUnauthorized {
			t.Fatalf("unauth status=%d", ansUnauthRR.Code)
		}

		webVerbSessionsMu.Lock()
		delete(webVerbSessions, userID)
		webVerbSessionsMu.Unlock()
		noSess := httptest.NewRequest(http.MethodPost, "/api/verb-training/answer", bytes.NewReader([]byte(`{}`)))
		noSess = verbFormsUserContext(noSess, userID)
		noSessRR := httptest.NewRecorder()
		router.handleVerbTrainingAnswer(noSessRR, noSess)
		if noSessRR.Code != http.StatusBadRequest {
			t.Fatalf("no session status=%d", noSessRR.Code)
		}

		state := &webVerbTrainingState{
			UserID: userID, SessionID: sessionID, Index: 0,
			Queue: []repository.VerbQueueCard{{
				UserVerbCardID: uvcID, CardType: models.VerbCardTypeCloze,
				PromptJSON: `{}`, AnswerJSON: `{"surface_form":"hablo"}`,
			}},
		}
		putVerbSession(t, userID, state)

		badBody := httptest.NewRequest(http.MethodPost, "/api/verb-training/answer", strings.NewReader(`{`))
		badBody = verbFormsUserContext(badBody, userID)
		badBodyRR := httptest.NewRecorder()
		router.handleVerbTrainingAnswer(badBodyRR, badBody)
		if badBodyRR.Code != http.StatusBadRequest {
			t.Fatalf("bad body status=%d", badBodyRR.Code)
		}

		mismatchBody, _ := json.Marshal(map[string]interface{}{"user_verb_card_id": uvcID + 1, "answer": "hablo"})
		mismatch := httptest.NewRequest(http.MethodPost, "/api/verb-training/answer", bytes.NewReader(mismatchBody))
		mismatch = verbFormsUserContext(mismatch, userID)
		mismatchRR := httptest.NewRecorder()
		router.handleVerbTrainingAnswer(mismatchRR, mismatch)
		if mismatchRR.Code != http.StatusBadRequest {
			t.Fatalf("mismatch status=%d", mismatchRR.Code)
		}

		wrongBody, _ := json.Marshal(map[string]interface{}{"user_verb_card_id": uvcID, "answer": "wrong"})
		wrongReq := httptest.NewRequest(http.MethodPost, "/api/verb-training/answer", bytes.NewReader(wrongBody))
		wrongReq = verbFormsUserContext(wrongReq, userID)
		wrongRR := httptest.NewRecorder()
		router.handleVerbTrainingAnswer(wrongRR, wrongReq)
		if wrongRR.Code != http.StatusOK {
			t.Fatalf("wrong answer status=%d", wrongRR.Code)
		}
		var wrongFB map[string]interface{}
		_ = json.Unmarshal(wrongRR.Body.Bytes(), &wrongFB)
		if wrongFB["is_correct"] != false || wrongFB["delay_seconds"] == nil {
			t.Fatalf("expected wrong feedback with delay: %v", wrongFB)
		}

		skipSessionID, _ := vrepo.StartVerbSession(userID, 1, `{}`)
		state2 := &webVerbTrainingState{
			UserID: userID, SessionID: skipSessionID, Index: 0,
			Queue: []repository.VerbQueueCard{{
				UserVerbCardID: uvcID, CardType: models.VerbCardTypeCloze,
				PromptJSON: `{}`, AnswerJSON: `{"surface_form":"hablo"}`,
			}},
		}
		putVerbSession(t, userID, state2)
		skipBody, _ := json.Marshal(map[string]interface{}{"user_verb_card_id": uvcID, "skip": true})
		skipReq := httptest.NewRequest(http.MethodPost, "/api/verb-training/answer", bytes.NewReader(skipBody))
		skipReq = verbFormsUserContext(skipReq, userID)
		skipRR := httptest.NewRecorder()
		router.handleVerbTrainingAnswer(skipRR, skipReq)
		if skipRR.Code != http.StatusOK {
			t.Fatalf("skip status=%d", skipRR.Code)
		}

		completeSessionID, _ := vrepo.StartVerbSession(userID, 1, `{}`)
		completeState := &webVerbTrainingState{
			UserID: userID, SessionID: completeSessionID, Index: 0,
			Queue: []repository.VerbQueueCard{{
				UserVerbCardID: uvcID, CardType: models.VerbCardTypeCloze,
				PromptJSON: `{}`, AnswerJSON: `{"surface_form":"hablo"}`,
			}},
		}
		putVerbSession(t, userID, completeState)
		okBody, _ := json.Marshal(map[string]interface{}{"answer": "hablo"})
		okReq := httptest.NewRequest(http.MethodPost, "/api/verb-training/answer", bytes.NewReader(okBody))
		okReq = verbFormsUserContext(okReq, userID)
		okRR := httptest.NewRecorder()
		router.handleVerbTrainingAnswer(okRR, okReq)
		if okRR.Code != http.StatusOK {
			t.Fatalf("complete status=%d body=%s", okRR.Code, okRR.Body.String())
		}
		var completeFB map[string]interface{}
		_ = json.Unmarshal(okRR.Body.Bytes(), &completeFB)
		if completeFB["complete"] != true {
			t.Fatalf("expected complete session: %v", completeFB)
		}

		zeroIDSession, _ := vrepo.StartVerbSession(userID, 1, `{}`)
		zeroIDState := &webVerbTrainingState{
			UserID: userID, SessionID: zeroIDSession, Index: 0,
			Queue: []repository.VerbQueueCard{{
				UserVerbCardID: uvcID, CardType: models.VerbCardTypeCloze,
				PromptJSON: `{}`, AnswerJSON: `{"surface_form":"hablo"}`,
			}},
		}
		putVerbSession(t, userID, zeroIDState)
		zeroBody, _ := json.Marshal(map[string]interface{}{"user_verb_card_id": 0, "answer": "hablo"})
		zeroReq := httptest.NewRequest(http.MethodPost, "/api/verb-training/answer", bytes.NewReader(zeroBody))
		zeroReq = verbFormsUserContext(zeroReq, userID)
		zeroRR := httptest.NewRecorder()
		router.handleVerbTrainingAnswer(zeroRR, zeroReq)
		if zeroRR.Code != http.StatusOK {
			t.Fatalf("zero card id status=%d", zeroRR.Code)
		}

		_, uvcB := seedVerbFormsCoverageCard(t, db, userID, func(s *verbCoverageSeed) { s.lemma = "traer" })
		nextSession, _ := vrepo.StartVerbSession(userID, 2, `{}`)
		nextState := &webVerbTrainingState{
			UserID: userID, SessionID: nextSession, Index: 0,
			Queue: []repository.VerbQueueCard{
				{UserVerbCardID: uvcID, CardType: models.VerbCardTypeCloze, PromptJSON: `{}`, AnswerJSON: `{"surface_form":"hablo"}`},
				{UserVerbCardID: uvcB, CardType: models.VerbCardTypeCloze, PromptJSON: `{}`, AnswerJSON: `{"surface_form":"traigo"}`},
			},
		}
		putVerbSession(t, userID, nextState)
		nextBody, _ := json.Marshal(map[string]interface{}{"answer": "hablo"})
		nextReq := httptest.NewRequest(http.MethodPost, "/api/verb-training/answer", bytes.NewReader(nextBody))
		nextReq = verbFormsUserContext(nextReq, userID)
		nextRR := httptest.NewRecorder()
		router.handleVerbTrainingAnswer(nextRR, nextReq)
		if nextRR.Code != http.StatusOK {
			t.Fatalf("next card status=%d body=%s", nextRR.Code, nextRR.Body.String())
		}
		var nextFB map[string]interface{}
		_ = json.Unmarshal(nextRR.Body.Bytes(), &nextFB)
		if nextFB["next"] != true || nextFB["complete"] == true {
			t.Fatalf("expected partial session: %v", nextFB)
		}

		exhausted := &webVerbTrainingState{
			UserID: userID, SessionID: sessionID, Index: 1,
			Queue: []repository.VerbQueueCard{{
				UserVerbCardID: uvcID, CardType: models.VerbCardTypeCloze,
				PromptJSON: `{}`, AnswerJSON: `{"surface_form":"hablo"}`,
			}},
		}
		putVerbSession(t, userID, exhausted)
		exReq := httptest.NewRequest(http.MethodPost, "/api/verb-training/answer", bytes.NewReader([]byte(`{"answer":"hablo"}`)))
		exReq = verbFormsUserContext(exReq, userID)
		exRR := httptest.NewRecorder()
		router.handleVerbTrainingAnswer(exRR, exReq)
		if exRR.Code != http.StatusBadRequest {
			t.Fatalf("exhausted session status=%d", exRR.Code)
		}

		brokenRouter, _, brokenUID := setupVerbFormsCoverageRouter(t, db, config.TrainingConfig{SpanishVerbFormsEnabled: true}, 900011)
		_, brokenUVC := seedVerbFormsCoverageCard(t, db, brokenUID, func(s *verbCoverageSeed) { s.lemma = "vivir" })
		brokenSessionID, _ := vrepo.StartVerbSession(brokenUID, 1, `{}`)
		brokenState := &webVerbTrainingState{
			UserID: brokenUID, SessionID: brokenSessionID, Index: 0,
			Queue: []repository.VerbQueueCard{{
				UserVerbCardID: brokenUVC, CardType: models.VerbCardTypeCloze,
				PromptJSON: `{}`, AnswerJSON: `{"surface_form":"vivo"}`,
			}},
		}
		putVerbSession(t, brokenUID, brokenState)
		brokenRouter.db = newBrokenDB(t)
		gradeBody, _ := json.Marshal(map[string]interface{}{"answer": "vivo"})
		gradeReq := httptest.NewRequest(http.MethodPost, "/api/verb-training/answer", bytes.NewReader(gradeBody))
		gradeReq = verbFormsUserContext(gradeReq, brokenUID)
		gradeRR := httptest.NewRecorder()
		brokenRouter.handleVerbTrainingAnswer(gradeRR, gradeReq)
		if gradeRR.Code != http.StatusInternalServerError {
			t.Fatalf("grade error status=%d", gradeRR.Code)
		}
	})

	t.Run("upcomingHandler", func(t *testing.T) {
		db := testutil.SetupTestDB(t)
		router, _, userID := setupVerbFormsCoverageRouter(t, db, config.TrainingConfig{
			SpanishVerbFormsEnabled: true, VerbFormsMaxCards: 10, VerbFormsMaxNew: 5,
		}, 900012)
		seedVerbFormsCoverageCard(t, db, userID)

		reqMethod := httptest.NewRequest(http.MethodPost, "/api/verb-training/upcoming", nil)
		reqMethod = verbFormsUserContext(reqMethod, userID)
		rrMethod := httptest.NewRecorder()
		router.handleVerbTrainingUpcoming(rrMethod, reqMethod)
		if rrMethod.Code != http.StatusMethodNotAllowed {
			t.Fatalf("method status=%d", rrMethod.Code)
		}

		reqUnauth := httptest.NewRequest(http.MethodGet, "/api/verb-training/upcoming", nil)
		rrUnauth := httptest.NewRecorder()
		router.handleVerbTrainingUpcoming(rrUnauth, reqUnauth)
		if rrUnauth.Code != http.StatusUnauthorized {
			t.Fatalf("unauth status=%d", rrUnauth.Code)
		}

		emptyPoolRouter, _, emptyPoolUID := setupVerbFormsCoverageRouter(t, db, config.TrainingConfig{
			SpanishVerbFormsEnabled: true, VerbFormsMaxCards: 10, VerbFormsMaxNew: 5,
		}, 900023)
		reqEmpty := httptest.NewRequest(http.MethodGet, "/api/verb-training/upcoming", nil)
		reqEmpty = verbFormsUserContext(reqEmpty, emptyPoolUID)
		rrEmpty := httptest.NewRecorder()
		emptyPoolRouter.handleVerbTrainingUpcoming(rrEmpty, reqEmpty)
		if rrEmpty.Code != http.StatusOK {
			t.Fatalf("empty pool status=%d", rrEmpty.Code)
		}
		var emptyBody map[string]interface{}
		_ = json.Unmarshal(rrEmpty.Body.Bytes(), &emptyBody)
		if emptyBody["pool_ready"] != false {
			t.Fatalf("expected pool_ready false: %v", emptyBody)
		}

		disabledUpcoming := NewRouter(zap.NewNop(), &config.Config{
			Learning: config.LearningConfig{TargetLang: "en"},
			Training: config.TrainingConfig{SpanishVerbFormsEnabled: true},
		}, db, nil, nil, nil, nil)
		reqDisabled := httptest.NewRequest(http.MethodGet, "/api/verb-training/upcoming", nil)
		reqDisabled = verbFormsUserContext(reqDisabled, userID)
		rrDisabled := httptest.NewRecorder()
		disabledUpcoming.handleVerbTrainingUpcoming(rrDisabled, reqDisabled)
		if rrDisabled.Code != http.StatusForbidden {
			t.Fatalf("disabled upcoming status=%d", rrDisabled.Code)
		}

		reqOK := httptest.NewRequest(http.MethodGet, "/api/verb-training/upcoming", nil)
		reqOK = verbFormsUserContext(reqOK, userID)
		rrOK := httptest.NewRecorder()
		router.handleVerbTrainingUpcoming(rrOK, reqOK)
		if rrOK.Code != http.StatusOK {
			t.Fatalf("ok status=%d body=%s", rrOK.Code, rrOK.Body.String())
		}

		broken := NewRouter(zap.NewNop(), &config.Config{
			Learning: config.LearningConfig{TargetLang: "es", GrammarBundleID: "es"},
			Training: config.TrainingConfig{SpanishVerbFormsEnabled: true, VerbFormsMaxCards: 10, VerbFormsMaxNew: 5},
		}, newBrokenDB(t), nil, nil, nil, nil)
		reqBroken := httptest.NewRequest(http.MethodGet, "/api/verb-training/upcoming", nil)
		reqBroken = verbFormsUserContext(reqBroken, userID)
		rrBroken := httptest.NewRecorder()
		broken.handleVerbTrainingUpcoming(rrBroken, reqBroken)
		if rrBroken.Code != http.StatusInternalServerError {
			t.Fatalf("broken db status=%d", rrBroken.Code)
		}
	})

	t.Run("internalPendingHandler", func(t *testing.T) {
		db := testutil.SetupTestDB(t)
		router, _, _ := setupVerbFormsCoverageRouter(t, db, config.TrainingConfig{SpanishVerbFormsEnabled: true}, 900013)
		router.internalServiceTokens = map[string]string{"default": "svc-tok"}
		seedPendingVerbLemma(t, db, "cantar")

		reqMethod := httptest.NewRequest(http.MethodPost, "/api/internal/verb-training/pending", nil)
		rrMethod := httptest.NewRecorder()
		router.handleInternalVerbTrainingPending(rrMethod, reqMethod)
		if rrMethod.Code != http.StatusMethodNotAllowed {
			t.Fatalf("method status=%d", rrMethod.Code)
		}

		reqUnauth := httptest.NewRequest(http.MethodGet, "/api/internal/verb-training/pending", nil)
		rrUnauth := httptest.NewRecorder()
		router.handleInternalVerbTrainingPending(rrUnauth, reqUnauth)
		if rrUnauth.Code != http.StatusUnauthorized {
			t.Fatalf("unauth status=%d", rrUnauth.Code)
		}

		for _, q := range []string{"?limit=0", "?limit=x", "?cursor=-1", "?cursor=bad"} {
			req := httptest.NewRequest(http.MethodGet, "/api/internal/verb-training/pending"+q, nil)
			req.Header.Set("X-Service-Token", "svc-tok")
			rr := httptest.NewRecorder()
			router.handleInternalVerbTrainingPending(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("query %s status=%d", q, rr.Code)
			}
		}

		reqGap := httptest.NewRequest(http.MethodGet, "/api/internal/verb-training/pending?forms_gap_only=false", nil)
		reqGap.Header.Set("X-Service-Token", "svc-tok")
		rrGap := httptest.NewRecorder()
		router.handleInternalVerbTrainingPending(rrGap, reqGap)
		if rrGap.Code != http.StatusOK {
			t.Fatalf("forms_gap_only=false status=%d", rrGap.Code)
		}

		reqOK := httptest.NewRequest(http.MethodGet, "/api/internal/verb-training/pending?limit=10&all=1&forms_gap_only=0", nil)
		reqOK.Header.Set("X-Service-Token", "svc-tok")
		rrOK := httptest.NewRecorder()
		router.handleInternalVerbTrainingPending(rrOK, reqOK)
		if rrOK.Code != http.StatusOK {
			t.Fatalf("ok status=%d body=%s", rrOK.Code, rrOK.Body.String())
		}
		var body map[string]interface{}
		if err := json.Unmarshal(rrOK.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body["count"].(float64) < 1 {
			t.Fatalf("expected pending items: %v", body)
		}

		broken := NewRouter(zap.NewNop(), &config.Config{
			Learning: config.LearningConfig{TargetLang: "es"},
		}, newBrokenDB(t), nil, nil, nil, nil)
		broken.internalServiceTokens = map[string]string{"default": "svc-tok"}
		reqBroken := httptest.NewRequest(http.MethodGet, "/api/internal/verb-training/pending", nil)
		reqBroken.Header.Set("X-Service-Token", "svc-tok")
		rrBroken := httptest.NewRecorder()
		broken.handleInternalVerbTrainingPending(rrBroken, reqBroken)
		if rrBroken.Code != http.StatusInternalServerError {
			t.Fatalf("broken db status=%d", rrBroken.Code)
		}

		reqEmptyList := httptest.NewRequest(http.MethodGet, "/api/internal/verb-training/pending?cursor=999999999", nil)
		reqEmptyList.Header.Set("X-Service-Token", "svc-tok")
		rrEmptyList := httptest.NewRecorder()
		router.handleInternalVerbTrainingPending(rrEmptyList, reqEmptyList)
		if rrEmptyList.Code != http.StatusOK {
			t.Fatalf("empty list status=%d", rrEmptyList.Code)
		}
	})

	t.Run("writeVerbFormsGroupedResponse", func(t *testing.T) {
		rr := httptest.NewRecorder()
		writeVerbFormsGroupedResponse(rr, 42, []repository.VerbFormViewRow{
			{Mood: "indicativo", Tense: "presente", SurfaceForm: "hablo"},
			{Mood: "indicativo", Tense: "preterito", SurfaceForm: "hable"},
		})
		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
	})
}

func TestVerbFormsHandlersCoverage_ParseWordCardID(t *testing.T) {
	if _, ok := parseWordCardIDForVerbForms("/api/vocab/12/verb-forms"); !ok {
		t.Fatal("valid path should parse")
	}
	if id, ok := parseWordCardIDForVerbForms("/api/vocab/12/verb-forms"); !ok || id != 12 {
		t.Fatalf("id=%d ok=%v", id, ok)
	}
}

func TestVerbFormsHandlersCoverage_WriteDisabled(t *testing.T) {
	db := testutil.SetupTestDB(t)
	router, _, _ := setupVerbFormsCoverageRouter(t, db, config.TrainingConfig{SpanishVerbFormsEnabled: true}, 900014)
	rr := httptest.NewRecorder()
	router.writeVerbTrainingDisabled(rr)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestVerbFormsHandlersCoverage_LemmaFormsOK(t *testing.T) {
	db := testutil.SetupTestDB(t)
	router, _, userID := setupVerbFormsCoverageRouter(t, db, config.TrainingConfig{SpanishVerbFormsEnabled: true}, 900015)
	seedVerbFormsCoverageCard(t, db, userID)
	req := httptest.NewRequest(http.MethodGet, "/api/verb-training/forms-by-lemma?lemma=hablar", nil)
	req = verbFormsUserContext(req, userID)
	rr := httptest.NewRecorder()
	router.handleVerbTrainingLemmaForms(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestVerbFormsHandlersCoverage_VocabFormsOK(t *testing.T) {
	db := testutil.SetupTestDB(t)
	router, _, userID := setupVerbFormsCoverageRouter(t, db, config.TrainingConfig{SpanishVerbFormsEnabled: true}, 900016)
	wordCardID, _ := seedVerbFormsCoverageCard(t, db, userID)
	req := httptest.NewRequest(http.MethodGet, "/api/vocab/"+strconv.FormatInt(wordCardID, 10)+"/verb-forms", nil)
	rr := httptest.NewRecorder()
	router.handleVocabVerbForms(rr, req, userID, wordCardID)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}
