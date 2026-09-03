package web

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http/httptest"
	"strings"
	"testing"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/placement"
	"tgbot-skeleton/internal/placementbundle"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"
	"time"
)

func setupPlacementHTTP(t *testing.T) (*Router, int64, int64) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	log := zap.NewNop()
	cfg := &config.Config{Learning: config.DefaultLearningConfig()}
	r := NewRouter(log, cfg, db, nil, nil, nil, nil)
	u := repository.NewUserRepository(db, log)
	one, e := u.GetOrCreateUser(88001)
	if e != nil {
		t.Fatal(e)
	}
	two, e := u.GetOrCreateUser(88002)
	if e != nil {
		t.Fatal(e)
	}
	services := map[string]*service.GrammarService{}
	for _, lang := range []string{"en", "es"} {
		lc := config.DefaultLearningConfig()
		lc.GrammarBundleID = lang
		lc.TargetLang = lang
		content, e := repository.NewGrammarContentRepositoryForLearning(lc, log)
		if e != nil {
			t.Fatal(e)
		}
		pub := repository.NewGrammarPublishRepository(db, log)
		sections, e := content.GetSections()
		if e != nil {
			t.Fatal(e)
		}
		for _, sec := range sections.Sections {
			if e = pub.SetPublished("section", sec.SectionID, true, nil); e != nil {
				t.Fatal(e)
			}
			for _, ch := range sec.ChapterIDs {
				if e = pub.SetPublished("chapter", ch, true, nil); e != nil {
					t.Fatal(e)
				}
			}
		}
		services[lang] = service.NewGrammarService(content, pub, repository.NewGrammarAttemptRepository(db, log), lc, log)
	}
	r.SetGrammarService(services["en"])
	r.SetGrammarServices(services)
	return r, one.ID, two.ID
}
func placementRequest(t *testing.T, r *Router, user int64, method, path string, body interface{}, status int) *service.PlacementSessionView {
	t.Helper()
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(raw))
	req = setUserIDInContext(req, user)
	w := httptest.NewRecorder()
	switch {
	case path == "/api/learning/placement/sessions":
		r.handlePlacementSessions(w, req)
	case strings.HasPrefix(path, "/api/learning/placement/results"):
		r.handlePlacementResults(w, req)
	default:
		r.handlePlacementSession(w, req)
	}
	if w.Code != status {
		t.Fatalf("%s %s got%d want%d: %s", method, path, w.Code, status, w.Body.String())
	}
	if status != 200 {
		return nil
	}
	var v service.PlacementSessionView
	if e := json.Unmarshal(w.Body.Bytes(), &v); e != nil {
		t.Fatal(e)
	}
	if v.Result == nil {
		for _, hidden := range []string{"correct_answer", "explanation", "family_id", "reserve", "skill_id"} {
			if strings.Contains(w.Body.String(), `"`+hidden+`"`) {
				t.Fatalf("private field %s leaked", hidden)
			}
		}
	}
	return &v
}
func startPlacement(t *testing.T, r *Router, user int64, course, key string, fresh bool) *service.PlacementSessionView {
	return placementRequest(t, r, user, "POST", "/api/learning/placement/sessions", map[string]interface{}{"course_code": course, "idempotency_key": key, "new_attempt": fresh}, 200)
}
func answerPlacement(t *testing.T, r *Router, user int64, v *service.PlacementSessionView, correct bool) *service.PlacementSessionView {
	t.Helper()
	stored, e := repository.NewPlacementAttemptRepository(r.db).Get(context.Background(), user, v.ID)
	if e != nil {
		t.Fatal(e)
	}
	keys := map[string]string{}
	for _, q := range append(stored.Snapshot.Items, stored.Snapshot.Reserve...) {
		keys[q.ID] = q.CorrectAnswer
	}
	for index := 0; index < len(v.Questions); index++ {
		q := v.Questions[index]
		a := ""
		if correct {
			a = keys[q.ID]
		}
		v = placementRequest(t, r, user, "POST", "/api/learning/placement/sessions/"+v.ID+"/answers", map[string]string{"question_id": q.ID, "answer": a}, 200)
	}
	return v
}
func TestPlacementHTTPIsolationRetakesAndAccess(t *testing.T) {
	r, user, other := setupPlacementHTTP(t)
	placementRequest(t, r, 0, "POST", "/api/learning/placement/sessions", map[string]string{}, 401)
	v := startPlacement(t, r, user, "es_ru", "first-attempt-key", false)
	if len(v.Questions) != 30 {
		t.Fatal("base count")
	}
	same := startPlacement(t, r, user, "es_ru", "first-attempt-key", false)
	if same.ID != v.ID {
		t.Fatal("idempotency")
	}
	same = startPlacement(t, r, user, "es_ru", "different-key-resume", false)
	if same.ID != v.ID {
		t.Fatal("resume")
	}
	path := "/api/learning/placement/sessions/" + v.ID
	placementRequest(t, r, other, "GET", path, nil, 404)
	placementRequest(t, r, user, "POST", path+"/finish", nil, 409)
	placementRequest(t, r, user, "POST", path+"/answers", map[string]string{"question_id": "forged", "answer": "a"}, 400)
	placementRequest(t, r, user, "POST", path+"/answers", map[string]interface{}{"question_id": v.Questions[0].ID, "answer": nil}, 400)
	answerPlacement(t, r, user, v, true)
	v = placementRequest(t, r, user, "POST", path+"/finish", nil, 200)
	if v.Result.Level != "C1" || v.Result.Total != 30 || len(v.Result.OpenedSections) == 0 {
		t.Fatalf("bad result %+v", v.Result)
	}
	repo := repository.NewGrammarAttemptRepository(r.db, zap.NewNop())
	access, e := repo.ForCourse("es_ru").GetPlacementTestResult(user)
	if e != nil || access == nil {
		t.Fatal(e)
	}
	opened := len(access.OpenedSections)
	english, e := repo.ForCourse("en_ru").GetPlacementTestResult(user)
	if e != nil || english != nil {
		t.Fatal("language access leaked")
	}
	old, e := repository.NewPlacementAttemptRepository(r.db).Get(context.Background(), user, v.ID)
	if e != nil {
		t.Fatal(e)
	}
	second := startPlacement(t, r, user, "es_ru", "second-attempt-key", true)
	latest, e := repository.NewPlacementAttemptRepository(r.db).Get(context.Background(), user, second.ID)
	if e != nil {
		t.Fatal(e)
	}
	families := map[string]bool{}
	for _, q := range old.Snapshot.Items {
		families[q.FamilyID] = true
	}
	for _, q := range latest.Snapshot.Items {
		if families[q.FamilyID] {
			t.Fatal("immediate family repeat")
		}
	}
	second = answerPlacement(t, r, user, second, false)
	second = placementRequest(t, r, user, "POST", "/api/learning/placement/sessions/"+second.ID+"/finish", nil, 200)
	if second.Result.Level != "below_a1" {
		t.Fatal("unknown answers")
	}
	access, e = repo.ForCourse("es_ru").GetPlacementTestResult(user)
	if e != nil || len(access.OpenedSections) != opened {
		t.Fatal("weak retake removed access")
	}
	// Retrying finish is harmless even after an explicit admin reset.
	if e = repo.ForCourse("es_ru").DeletePlacementTestResult(user); e != nil {
		t.Fatal(e)
	}
	placementRequest(t, r, user, "POST", path+"/finish", nil, 200)
	access, e = repo.ForCourse("es_ru").GetPlacementTestResult(user)
	if e != nil || access != nil {
		t.Fatal("finish replay restored reset access")
	}
}
func TestPlacementPinnedDBVersionAndClarification(t *testing.T) {
	r, user, _ := setupPlacementHTTP(t)
	r.config.Learning.ContentSource = "db"
	ctx := context.Background()
	bank, e := placementbundle.Load("en_ru")
	if e != nil {
		t.Fatal(e)
	}
	if e = repository.ImportPlacementBank(ctx, r.db, bank); e != nil {
		t.Fatal(e)
	}
	v := startPlacement(t, r, user, "en_ru", "version-one-key", false)
	stored, e := repository.NewPlacementAttemptRepository(r.db).Get(ctx, user, v.ID)
	if e != nil {
		t.Fatal(e)
	}
	changed := *bank
	changed.Version = "test-new-version"
	changed.Items = append(changed.Items[:0:0], bank.Items...)
	for i := range changed.Items {
		changed.Items[i].CorrectAnswer = changed.Items[i].Choices[(i+1)%len(changed.Items[i].Choices)].ID
	}
	if e = repository.ImportPlacementBank(ctx, r.db, &changed); e != nil {
		t.Fatal(e)
	}
	correctByLevel := map[string]int{"A1": 6, "A2": 6, "B1": 4}
	seen := map[string]int{}
	for _, q := range stored.Snapshot.Items {
		a := ""
		if seen[q.Level] < correctByLevel[q.Level] {
			a = q.CorrectAnswer
		}
		seen[q.Level]++
		v = placementRequest(t, r, user, "POST", "/api/learning/placement/sessions/"+v.ID+"/answers", map[string]string{"question_id": q.ID, "answer": a}, 200)
	}
	if len(v.Questions) != 36 || !v.Clarifying || !v.BaseClosed {
		t.Fatal("missing clarification")
	}
	placementRequest(t, r, user, "POST", "/api/learning/placement/sessions/"+v.ID+"/answers", map[string]string{"question_id": stored.Snapshot.Items[0].ID, "answer": ""}, 409)
	// Keys come from the original unseen reserve, despite active bank replacement.
	for _, q := range stored.Snapshot.Reserve {
		if q.Level == "B1" {
			v = placementRequest(t, r, user, "POST", "/api/learning/placement/sessions/"+v.ID+"/answers", map[string]string{"question_id": q.ID, "answer": q.CorrectAnswer}, 200)
		}
	}
	v = placementRequest(t, r, user, "POST", "/api/learning/placement/sessions/"+v.ID+"/finish", nil, 200)
	if v.Result.Level != "B1" || v.Result.Correct != 22 || v.BankVersion != bank.Version {
		t.Fatalf("pinned grading wrong: %s/%d", v.Result.Level, v.Result.Correct)
	}
	var versions int
	if e = r.db.QueryRow(`SELECT COUNT(*) FROM placement_banks WHERE course_code='en_ru'`).Scan(&versions); e != nil || versions != 2 {
		t.Fatal("old version removed", e)
	}
	for i := 0; i < 2; i++ {
		v = startPlacement(t, r, user, "en_ru", fmt.Sprintf("next-version-key-%d", i), true)
		if v.BankVersion != changed.Version {
			t.Fatal("new start not on active bank")
		}
	}
}

func TestPlacementResumeSurvivesBankOutageAndLostResponse(t *testing.T) {
	r, user, _ := setupPlacementHTTP(t)
	r.config.Learning.ContentSource = "db"
	bank, err := placementbundle.Load("en_ru")
	if err != nil {
		t.Fatal(err)
	}
	if err = repository.ImportPlacementBank(context.Background(), r.db, bank); err != nil {
		t.Fatal(err)
	}
	v := startPlacement(t, r, user, "en_ru", "outage-start-key", false)
	if _, err = r.db.Exec(`UPDATE placement_banks SET active=false`); err != nil {
		t.Fatal(err)
	}
	resumed := startPlacement(t, r, user, "en_ru", "other-device-key", false)
	if resumed.ID != v.ID {
		t.Fatal("bank outage lost pinned session")
	}
	v = answerPlacement(t, r, user, v, false)
	v = placementRequest(t, r, user, "POST", "/api/learning/placement/sessions/"+v.ID+"/finish", nil, 200)
	retry := startPlacement(t, r, user, "en_ru", "other-device-key", false)
	if retry.ID != v.ID || retry.Status != "completed" {
		t.Fatal("retry of lost resume response created another attempt")
	}
	placementRequest(t, r, user, "POST", "/api/learning/placement/sessions", map[string]interface{}{"course_code": "en_ru", "idempotency_key": "new-after-outage-key", "new_attempt": true}, 503)
}

func TestPlacementUnsupportedPolicyExpiresAndCanRestart(t *testing.T) {
	r, user, _ := setupPlacementHTTP(t)
	v := startPlacement(t, r, user, "es_ru", "old-policy-key", false)
	if _, err := r.db.Exec(`UPDATE placement_sessions SET policy_version='retired-policy' WHERE id=?`, v.ID); err != nil {
		t.Fatal(err)
	}
	path := "/api/learning/placement/sessions/" + v.ID
	placementRequest(t, r, user, "GET", path, nil, 410)
	placementRequest(t, r, user, "POST", path+"/answers", map[string]string{"question_id": v.Questions[0].ID, "answer": ""}, 410)
	placementRequest(t, r, user, "POST", "/api/learning/placement/sessions", map[string]string{"course_code": "es_ru", "idempotency_key": "old-policy-key"}, 410)
	fresh := startPlacement(t, r, user, "es_ru", "replacement-policy-key", false)
	if fresh.ID == v.ID {
		t.Fatal("unsupported policy was resumed")
	}
}

func TestPlacementPublishedLinksFollowCurrentParentState(t *testing.T) {
	r, user, _ := setupPlacementHTTP(t)
	v := startPlacement(t, r, user, "es_ru", "published-links-key", false)
	v = answerPlacement(t, r, user, v, false)
	path := "/api/learning/placement/sessions/" + v.ID
	v = placementRequest(t, r, user, "POST", path+"/finish", nil, 200)
	if len(v.Result.RecommendedSkills) == 0 {
		t.Fatal("fixture requires recommendations")
	}
	g := r.placementGrammar("es_ru")
	sections, err := g.ContentRepo.GetSections()
	if err != nil {
		t.Fatal(err)
	}
	for _, sec := range sections.Sections {
		if err = g.PublishRepo.SetPublished("section", sec.SectionID, false, nil); err != nil {
			t.Fatal(err)
		}
	}
	check := func(v *service.PlacementSessionView) {
		t.Helper()
		if len(v.Result.RecommendedSkills) != 0 {
			t.Fatal("unpublished parent recommendation leaked")
		}
		for _, q := range v.Result.Review {
			if len(q.ChapterIDs) != 0 {
				t.Fatal("unpublished parent chapter link leaked")
			}
		}
	}
	check(placementRequest(t, r, user, "GET", path, nil, 200))
	check(placementRequest(t, r, user, "POST", path+"/finish", nil, 200))
	check(startPlacement(t, r, user, "es_ru", "published-links-key", false))
	req := setUserIDInContext(httptest.NewRequest("GET", "/api/learning/placement/results?course_code=es_ru", nil), user)
	w := httptest.NewRecorder()
	r.handlePlacementResults(w, req)
	if w.Code != 200 {
		t.Fatalf("history: %s", w.Body.String())
	}
	var history struct {
		Sessions []*service.PlacementSessionView `json:"sessions"`
	}
	if err = json.Unmarshal(w.Body.Bytes(), &history); err != nil {
		t.Fatal(err)
	}
	if len(history.Sessions) != 1 {
		t.Fatal("history missing result")
	}
	check(history.Sessions[0])
	// Republishing the parent but withdrawing the chapter must also hide links.
	blocked := map[string]bool{}
	stored, err := repository.NewPlacementAttemptRepository(r.db).Get(context.Background(), user, v.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range stored.Result.Review {
		for _, id := range q.ChapterIDs {
			blocked[id] = true
		}
	}
	for _, sec := range sections.Sections {
		if err = g.PublishRepo.SetPublished("section", sec.SectionID, true, nil); err != nil {
			t.Fatal(err)
		}
	}
	for id := range blocked {
		if err = g.PublishRepo.SetPublished("chapter", id, false, nil); err != nil {
			t.Fatal(err)
		}
	}
	check(placementRequest(t, r, user, "GET", path, nil, 200))
}

func TestPlacementConcurrentFinishAndStartKeepCompletedHistory(t *testing.T) {
	r, user, _ := setupPlacementHTTP(t)
	v := startPlacement(t, r, user, "es_ru", "concurrent-first-key", false)
	repo := repository.NewPlacementAttemptRepository(r.db)
	stored, err := repo.Get(context.Background(), user, v.ID)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	locked, release := make(chan struct{}), make(chan struct{})
	finished := make(chan error, 1)
	go func() {
		_, err := repo.Update(ctx, user, v.ID, func(s *repository.PlacementSession, tx *sql.Tx) error {
			close(locked)
			<-release
			// This is Finish's transaction boundary: both result and access commit together.
			s.Status = "completed"
			s.Result = &placement.Result{Level: "A1"}
			return repository.SaveDiagnosticPlacementAccessTx(ctx, tx, s.UserCourseID, 80, 30, []string{"es.grammar.a1"})
		})
		finished <- err
	}()
	<-locked
	type outcome struct {
		s   *repository.PlacementSession
		err error
	}
	started := make(chan outcome, 1)
	go func() {
		s, err := repo.Start(ctx, user, stored.UserCourseID, "es_ru", "concurrent-next-key", true, func(map[string]int) (string, placement.Snapshot, error) {
			return stored.BankVersion, stored.Snapshot, nil
		})
		started <- outcome{s, err}
	}()
	// Give Start time to reach the held enrollment; it must wait for Finish.
	select {
	case premature := <-started:
		close(release)
		t.Fatalf("Start did not serialize with Finish: %+v", premature)
	case <-time.After(100 * time.Millisecond):
	}
	// It must never overwrite completion or deadlock with the access FK.
	close(release)
	if err = <-finished; err != nil {
		t.Fatal(err)
	}
	next := <-started
	if next.err != nil {
		t.Fatal(next.err)
	}
	previous, err := repo.Get(ctx, user, v.ID)
	if err != nil {
		t.Fatal(err)
	}
	if previous.Status != "completed" || next.s.Status != "active" || next.s.ID == v.ID {
		t.Fatal("concurrent start corrupted completion")
	}
	var resultCount int
	if err = r.db.QueryRow(`SELECT COUNT(*) FROM grammar_placement_access WHERE user_course_id=? AND cleared=false`, stored.UserCourseID).Scan(&resultCount); err != nil || resultCount != 1 {
		t.Fatalf("completion access was lost: %d %v", resultCount, err)
	}
}

func TestPlacementResultLinksRespectCurrentCourseAccess(t *testing.T) {
	r, user, _ := setupPlacementHTTP(t)
	v := startPlacement(t, r, user, "en_ru", "available-links-key", false)
	stored, err := repository.NewPlacementAttemptRepository(r.db).Get(context.Background(), user, v.ID)
	if err != nil {
		t.Fatal(err)
	}
	path := "/api/learning/placement/sessions/" + v.ID
	for _, q := range stored.Snapshot.Items {
		answer := ""
		if q.Level == "A1" {
			answer = q.CorrectAnswer
		}
		placementRequest(t, r, user, "POST", path+"/answers", map[string]string{"question_id": q.ID, "answer": answer}, 200)
	}
	v = placementRequest(t, r, user, "POST", path+"/finish", nil, 200)
	if v.Result.Level != "A1" || len(v.Result.RecommendedSkills) == 0 || len(v.AvailableChapterIDs) == 0 {
		t.Fatal("fixture requires A1 access and A2 recommendations")
	}
	checkRecommendations := func(v *service.PlacementSessionView, want bool) {
		t.Helper()
		available := map[string]bool{}
		for _, id := range v.AvailableChapterIDs {
			available[id] = true
		}
		for _, sk := range v.Result.RecommendedSkills {
			found := false
			for _, id := range sk.ChapterIDs {
				found = found || available[id]
			}
			if found != want {
				t.Fatalf("available=%v want=%v for %s", found, want, sk.ID)
			}
		}
	}
	checkRecommendations(v, false)
	if err = r.placementGrammar("en_ru").AdminSetGrammarPlacementLevel(context.Background(), user, "C1"); err != nil {
		t.Fatal(err)
	}
	updated := placementRequest(t, r, user, "GET", path, nil, 200)
	checkRecommendations(updated, true)
	if updated.Result.Level != "A1" {
		t.Fatal("current access changed historical score")
	}
}
