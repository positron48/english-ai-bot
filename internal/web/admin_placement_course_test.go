package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestAdminPlacement_SelectedCourseAndProgress(t *testing.T) {
	db := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	lc := config.DefaultLearningConfig()
	router := NewRouter(logger, &config.Config{Learning: lc}, db, nil, nil, nil, nil)
	users := repository.NewUserRepository(db, logger)
	admin, err := users.GetOrCreateUser(980001)
	if err != nil {
		t.Fatal(err)
	}
	student, err := users.GetOrCreateUser(980002)
	if err != nil {
		t.Fatal(err)
	}
	publish := repository.NewGrammarPublishRepository(db, logger)
	attempts := repository.NewGrammarAttemptRepository(db, logger)
	services := map[string]*service.GrammarService{}
	for _, language := range []string{"en", "es"} {
		courseConfig := lc
		courseConfig.GrammarBundleID = language
		content, err := repository.NewGrammarContentRepositoryForLearning(courseConfig, logger)
		if err != nil {
			t.Fatal(err)
		}
		services[language] = service.NewGrammarService(content, publish, attempts, courseConfig, logger)
		sections, err := content.GetSections()
		if err != nil {
			t.Fatal(err)
		}
		for _, section := range sections.Sections {
			if err := publish.SetPublished("section", section.SectionID, true, nil); err != nil {
				t.Fatal(err)
			}
			if err := publish.SetPublished("chapter", section.ChapterIDs[0], true, nil); err != nil {
				t.Fatal(err)
			}
		}
	}
	router.grammarService = services["en"]
	router.SetGrammarServices(services)
	setLevel := func(course, level string) {
		t.Helper()
		req := httptest.NewRequest(http.MethodPut, "/api/admin/users/2/grammar-placement?course_code="+course, strings.NewReader(fmt.Sprintf(`{"level":%q}`, level)))
		req = setUserIDInContext(req, admin.ID)
		response := httptest.NewRecorder()
		router.handleAdminUserGrammarPlacement(response, req, student.ID)
		if response.Code != http.StatusOK {
			t.Fatalf("set %s: %d %s", course, response.Code, response.Body.String())
		}
	}
	getLevel := func(course string) string {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/api/admin/users?course_code="+course, nil)
		req = setUserIDInContext(req, admin.ID)
		response := httptest.NewRecorder()
		router.handleAdminUsers(response, req)
		if response.Code != http.StatusOK {
			t.Fatalf("list %s: %d %s", course, response.Code, response.Body.String())
		}
		var body struct {
			Users []struct {
				ID        int64 `json:"id"`
				Placement *struct {
					Level  string `json:"level"`
					Course string `json:"course_code"`
				} `json:"grammar_placement"`
			} `json:"users"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		for _, user := range body.Users {
			if user.ID == student.ID {
				if user.Placement == nil {
					return ""
				}
				if user.Placement.Course != course {
					t.Fatalf("wrong course in summary: %+v", user.Placement)
				}
				return user.Placement.Level
			}
		}
		t.Fatal("student missing")
		return ""
	}
	setLevel("en_ru", "A2")
	setLevel("es_ru", "B2")
	if en, es := getLevel("en_ru"), getLevel("es_ru"); en != "A2" || es != "B2" {
		t.Fatalf("course summaries: en=%s es=%s", en, es)
	}
	stats, err := services["es"].GetGrammarStatistics(t.Context(), student.ID)
	if err != nil || stats.PassedChapters != 0 || stats.CourseCompletionPct != 0 || stats.ConfirmedLevel != "" {
		t.Fatalf("admin access manufactured learning progress: %+v, %v", stats, err)
	}
	setLevel("es_ru", "")
	if en, es := getLevel("en_ru"), getLevel("es_ru"); en != "A2" || es != "" {
		t.Fatalf("course reset: en=%s es=%s", en, es)
	}
}
