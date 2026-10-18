package web

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"
)

// Opt-in localhost-only QA harness. Uses a disposable PostgreSQL container, real
// HTTP/auth/subject adapters and deterministic AI. Never connects to production.
func TestDailyJourneyBrowserQA(t *testing.T) {
	if os.Getenv("DAILY_JOURNEY_BROWSER_QA") != "1" {
		t.Skip("opt-in browser harness")
	}
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	exec := func(query string, args ...any) { t.Helper(); _, err := db.Exec(query, args...); must(err) }
	users := repository.NewUserRepository(db, logger)
	user, err := users.GetOrCreateUser(809000)
	must(err)
	exec(`UPDATE users SET subscription_tier='pro_plus',settings_json='{"autoplay_pronunciation":false,"sounds_enabled":false,"spell_mode_enabled":false,"type_mode_enabled":true,"type_mastering_threshold":0,"options_delay_seconds":0}' WHERE id=?`, user.ID)
	courses := repository.NewCourseRepository(db, logger)
	_, err = courses.EnsureUserCourse(ctx, user.ID, "es_ru")
	must(err)
	var courseID, districtID int64
	must(db.QueryRow(`SELECT c.id,d.id FROM courses c JOIN districts d ON d.course_id=c.id WHERE c.code='es_ru' AND d.level_code='A0'`).Scan(&courseID, &districtID))
	lc := config.DefaultLearningConfig()
	lc.TargetLang = "es"
	lc.GrammarBundleID = "es"
	cfg := &config.Config{Learning: lc, WebApp: config.WebAppConfig{JWTSecret: "disposable-local-journey-qa-only", JWTTTLHours: 1, RefreshTTLHours: 1}, Training: config.TrainingConfig{SpanishVerbFormsEnabled: true, VerbFormsMaxCards: 10, VerbFormsMaxNew: 10}}
	cfg.Linglow.DailyJourneyEnabled = true
	cards := repository.NewTrainingCardRepository(db, logger)
	userCards := repository.NewUserCardRepository(db, logger)
	for i, word := range []string{"casa", "gato", "perro", "libro", "mesa"} {
		var id int64
		must(db.QueryRow(`INSERT INTO word_cards(word,definition,course_code) VALUES(?,'{}','es_ru') RETURNING id`, word).Scan(&id))
		tc, err := cards.CreateTrainingCard(&models.TrainingCard{WordCardID: id, WordEN: word, WordRU: []string{"дом", "кот", "собака", "книга", "стол"}[i], DistractorsEN: `["uno","dos","tres"]`, DistractorsRU: `["один","два","три"]`})
		must(err)
		direction := models.DirectionENtoRU
		if os.Getenv("DAILY_JOURNEY_QA_WORD_INPUT") != "" {
			direction = models.DirectionRUtoEN
		}
		_, err = userCards.CreateUserCard(&models.UserCard{UserID: user.ID, TrainingCardID: tc, Direction: direction, State: models.StateReview, EF: 2.5, Reps: 3})
		must(err)
	}
	if mode := os.Getenv("DAILY_JOURNEY_QA_WORD_INPUT"); mode != "" {
		exec(`UPDATE users SET settings_json=? WHERE id=?`, fmt.Sprintf(`{"autoplay_pronunciation":false,"spell_mode_enabled":%t,"type_mode_enabled":%t,"spell_mastering_threshold":0,"type_mastering_threshold":0,"options_delay_seconds":0}`, mode == "spell", mode == "type"), user.ID)
	}
	if os.Getenv("DAILY_JOURNEY_QA_ADD_WORDS") == "1" {
		sets := repository.NewWordSetRepository(db, logger)
		level := "A0"
		setID, err := sets.CreateWordSet(&models.WordSet{CourseCode: "es_ru", Title: "QA colours", IsPublished: true, LevelCode: &level})
		must(err)
		ids := []int64{}
		for i, word := range []string{"azul", "rojo", "verde"} {
			var id int64
			must(db.QueryRow(`INSERT INTO word_cards(word,definition,course_code) VALUES(?,'{}','es_ru') RETURNING id`, word).Scan(&id))
			_, err := cards.CreateTrainingCard(&models.TrainingCard{WordCardID: id, WordEN: word, WordRU: []string{"синий", "красный", "зелёный"}[i], DistractorsEN: `["uno","dos","tres"]`, DistractorsRU: `["один","два","три"]`})
			must(err)
			ids = append(ids, id)
		}
		must(sets.SetWordSetItems(setID, ids))
	}
	training := service.NewTrainingService(userCards, cards, repository.NewSessionRepository(db, logger), repository.NewUserWordMasteringRepository(db, logger), lc, logger)
	r := NewRouter(logger, cfg, db, training, nil, nil, nil)
	r.SetDependencies(users, nil, nil, nil, "qa-only")
	r.SetOTPRepo(repository.NewWebOTPRepository(db, logger))
	cat := repository.NewReadingCatalogRepository(db)
	must(cat.ReplaceCatalog("qa", "", []repository.ReadingCategoryUpsert{{CategoryID: "qa-a0", Title: "QA", Level: "A0", TextIDs: []string{"qa-story"}}}, []repository.ReadingTextUpsert{{TextID: "qa-story", CategoryID: "qa-a0", Title: "Un día", Level: "A0", TargetLanguage: "es", ReadingPassageJSON: `{"title":"Un día","segments":[{"segment_id":"qa-1","speaker_id":"narrator","tokens":[{"token_idx":0,"surface":"Tengo"},{"token_idx":1,"surface":"un"},{"token_idx":2,"surface":"gato."}]}],"comprehension_questions":[{"id":"q1","prompt":"¿Tengo un gato?","type":"true_false","correct_answer":"true"}]}`}}))
	exec(`INSERT INTO grammar_content_bundle_meta(bundle_id,app_code,native_lang,target_lang,version,source_hash,sections_json,index_json) VALUES('es','es','ru','es','qa','qa',?,?) ON CONFLICT(bundle_id) DO UPDATE SET sections_json=excluded.sections_json,index_json=excluded.index_json`, `{"sections":[{"section_id":"es.qa","level":"A0","chapter_ids":["es.qa.one","es.qa.two"]}]}`, `{"chapters":{"es.qa.one":"one.json","es.qa.two":"two.json"}}`)
	pub := repository.NewGrammarPublishRepository(db, logger)
	must(pub.SetPublished("section", "es.qa", true, nil))
	for _, chapter := range []string{"es.qa.one", "es.qa.two"} {
		raw := fmt.Sprintf(`{"id":%q,"section_id":"es.qa","title":"QA grammar","target_language":"es","blocks":[{"id":"b1","type":"theory","theory":{"content_md":"**Soy** means I am."}},{"id":"quiz","type":"quiz_inline","quiz_inline":{"question_ids":["q1"]}}],"question_bank":{"questions":[{"id":"q1","type":"fill_blank","prompt":"Yo ___ estudiante.","correct_answer":"soy","chapter_id":%q,"theory_block_id":"b1"}]},"chapter_test":{"pool_question_ids":["q1"],"num_questions":1}}`, chapter, chapter)
		exec(`INSERT INTO grammar_content_chapters(bundle_id,chapter_id,section_id,title,level,ui_language,target_language,raw_json,source_hash) VALUES('es',?,'es.qa','QA grammar','A0','ru','es',?,'qa')`, chapter, raw)
		must(pub.SetPublished("chapter", chapter, true, nil))
	}
	exec(`INSERT INTO grammar_training_content_questions(bundle_id,question_id,chapter_id,theory_block_id,raw_json,source_hash) VALUES('es','qa-srs','es.qa.one','b1',?,'qa')`, `{"id":"qa-srs","chapter_id":"es.qa.one","theory_block_id":"b1","type":"fill_blank","prompt":"Yo ___ estudiante.","correct_answer":"soy"}`)
	grammar := service.NewGrammarService(repository.NewGrammarContentRepositoryFromDB(db, "es", logger), pub, repository.NewGrammarAttemptRepository(db, logger), lc, logger)
	grammar.SetSRSRepository(repository.NewGrammarSRSRepository(db, logger))
	grammar.SetTrainingPackRepository(repository.NewGrammarTrainingPackRepositoryFromDB(db, "es", logger))
	must(grammar.AttemptRepo.UpdateProgress(user.ID, "es.qa.one", 100, true))
	r.SetGrammarServices(map[string]*service.GrammarService{"es": grammar})
	_, err = courses.MapLegacyContent(ctx, "es_ru", "es")
	must(err)
	var picture, npc int64
	must(db.QueryRow(`INSERT INTO picture_quests(course_id,district_id,code,cefr_level,title,image_url,image_description,max_turns,token_budget,status) VALUES(?,?,'qa-picture','A0','Describe the cat','/app/src/assets/linglow/art/lumi-map.png','A cat in a house',3,10000,'active') RETURNING id`, courseID, districtID).Scan(&picture))
	exec(`INSERT INTO picture_quest_tasks(quest_id,code,is_required,title,completion_criteria) VALUES(?,'required',true,'Describe the cat','Any sentence about the cat')`, picture)
	must(db.QueryRow(`INSERT INTO conversation_scenarios(course_id,district_id,code,place_type,cefr_level,title,npc_name,npc_persona,scene_setup,is_quest,max_turns,token_budget,status) VALUES(?,?,'qa-npc','cafe','A0','At a cafe','Marta','A friendly barista','Order a coffee',true,3,10000,'active') RETURNING id`, courseID, districtID).Scan(&npc))
	exec(`INSERT INTO conversation_tasks(scenario_id,code,is_required,title,completion_criteria) VALUES(?,'required',true,'Order a coffee','Any coffee request')`, npc)
	var failed atomic.Bool
	fake := func(_ context.Context, input ai.ConversationTurnSplitInput) (*ai.ChatTurnResult, error) {
		if len(input.History) == 0 {
			return &ai.ChatTurnResult{VisibleContent: "¡Hola! ¿Qué ves aquí?", CompletionTokens: 2}, nil
		}
		if os.Getenv("DAILY_JOURNEY_QA_AI_FAIL_ONCE") == "1" && failed.CompareAndSwap(false, true) {
			return nil, errors.New("intentional QA provider failure")
		}
		return &ai.ChatTurnResult{VisibleContent: "¡Muy bien! Gracias.", CompletedTaskCodes: []string{"required"}, CompletionTokens: 2}, nil
	}
	resolve := func(course string) *service.GrammarService {
		if course == "es_ru" {
			return grammar
		}
		return nil
	}
	adapters := map[dailyjourney.Mode]service.DailyJourneyAdapter{dailyjourney.Words: service.NewDailyJourneyWordsAdapter(db, logger, lc), dailyjourney.Reading: service.NewDailyJourneyReadingAdapter(db), dailyjourney.Grammar: service.NewDailyJourneyGrammarAdapter(db, resolve), dailyjourney.GrammarSRS: service.NewDailyJourneyGrammarSRSAdapter(db, resolve), dailyjourney.VerbForms: service.NewDailyJourneyVerbsAdapter(db, cfg.Training, lc, logger), dailyjourney.Picture: &dailyJourneyChatAdapter{router: r, mode: dailyjourney.Picture, run: fake}, dailyjourney.Conversation: &dailyJourneyChatAdapter{router: r, mode: dailyjourney.Conversation, run: fake}}
	verbs := repository.NewVerbFormsRepository(db, logger)
	var verbWord int64
	must(db.QueryRow(`INSERT INTO word_cards(word,definition,pos,course_code) VALUES('qaviajar','{}','verb','es_ru') RETURNING id`).Scan(&verbWord))
	exec(`INSERT INTO user_word_knowledge(user_id,word_card_id,status) VALUES(?,?,'known')`, user.ID, verbWord)
	lemma, err := verbs.UpsertVerbLemma("qaviajar", "es", "qa", "1", "qa", "{}")
	must(err)
	must(verbs.LinkWordCardToLemma(verbWord, lemma, 1, "qa"))
	form, err := verbs.UpsertVerbForm(&models.VerbFormDict{VerbLemmaID: lemma, Mood: "indicativo", Tense: "presente", Person: "3", Number: "singular", SurfaceForm: "viaja"})
	must(err)
	_, err = verbs.UpsertVerbTrainingCard(&models.VerbTrainingCard{WordCardID: verbWord, VerbFormDictID: form, CardType: models.VerbCardTypeCloze, PromptJSON: `{"content_version":4,"practice_eligible":true,"lemma":"qaviajar","question":"Él ____.","example_translation":"Он путешествует.","person":"3","number":"singular","tense":"presente","mood":"indicativo","rule":{"id":"qa","regular":false}}`, AnswerJSON: `{"surface_form":"viaja"}`, DistractorsJSON: `[]`})
	must(err)
	r.SetDailyJourneyService(service.NewDailyJourneyService(db, adapters))
	// Explicit QA itinerary intentionally includes every seeded mode, bypassing only random selection.
	_, err = repository.NewDailyJourneyRepository(db).Start(ctx, user.ID, "es_ru", time.Now(), func(tx *sql.Tx) ([]dailyjourney.Step, error) {
		steps := []dailyjourney.Step{}
		modes := []dailyjourney.Mode{dailyjourney.Words, dailyjourney.Reading, dailyjourney.Grammar, dailyjourney.GrammarSRS, dailyjourney.VerbForms, dailyjourney.Picture, dailyjourney.Conversation}
		if selected := os.Getenv("DAILY_JOURNEY_QA_MODES"); selected != "" {
			modes = nil
			for _, mode := range strings.Split(selected, ",") {
				modes = append(modes, dailyjourney.Mode(mode))
			}
		}
		for _, mode := range modes {
			items, err := adapters[mode].Candidates(ctx, tx, user.ID, "es_ru")
			if err != nil {
				return nil, err
			}
			if len(items) == 0 {
				return nil, fmt.Errorf("QA seed missing %s", mode)
			}
			item := items[0]
			item.Status = "pending"
			item.Required = mode == dailyjourney.Words || mode == dailyjourney.Reading
			steps = append(steps, item)
		}
		return steps, nil
	})
	must(err)
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/auth/request_otp" {
			if req.Method != "POST" {
				http.Error(w, "POST required", http.StatusMethodNotAllowed)
				return
			}
			_ = req.ParseForm()
			if req.FormValue("username") != "journey-qa" {
				http.Error(w, "Use journey-qa on this disposable QA server", 400)
				return
			}
			hash := sha256.Sum256([]byte("123456"))
			_, err := db.Exec(`INSERT INTO web_otps(user_id,code_hash,expires_at) VALUES(?,?,CURRENT_TIMESTAMP+INTERVAL '30 minutes')`, user.ID, hex.EncodeToString(hash[:]))
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			writeJSON(w, map[string]any{"success": true, "user_id": user.ID, "message": "Disposable QA OTP: 123456"})
			return
		}
		r.ServeHTTP(w, req)
	})
	listener, err := net.Listen("tcp", "127.0.0.1:8184")
	must(err)
	server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	defer server.Close()
	t.Log("QA ready: localhost:8184; login journey-qa / 123456; disposable PostgreSQL, deterministic AI")
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	select {
	case err := <-done:
		if err != http.ErrServerClosed {
			t.Fatal(err)
		}
	case <-time.After(20 * time.Minute):
	}
}
