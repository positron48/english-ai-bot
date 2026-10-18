package web

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/service"
)

// Opt-in smoke against the developer's configured provider. Only synthetic
// content is sent; subject writes stay in the disposable test PostgreSQL.
func TestDailyJourneyChatRealProvider(t *testing.T) {
	if os.Getenv("DAILY_JOURNEY_REAL_AI") != "1" {
		t.Skip("external provider smoke is opt-in")
	}
	env, err := godotenv.Read("../../.env")
	if err != nil {
		t.Fatal("local AI configuration unavailable")
	}
	providerName := env["AI_PROVIDER"]
	if override := os.Getenv("DAILY_JOURNEY_REAL_AI_PROVIDER"); override != "" {
		if override != "openrouter" && override != "polza" {
			t.Fatal("unknown test provider")
		}
		providerName = override
	}
	endpoint, key := env["AI_URL"], env["AI_API_KEY"]
	if providerName == "polza" {
		endpoint, key = env["POLZA_AI_URL"], env["POLZA_AI_API_KEY"]
		if endpoint == "" {
			endpoint = "https://polza.ai/api/v1"
		}
	}
	if endpoint == "" || key == "" {
		t.Fatal("configured provider URL/key missing")
	}
	model := env["AI_CONVERSATION_MODEL"]
	if model == "" {
		model = env["AI_MODEL"]
	}
	if override := os.Getenv("DAILY_JOURNEY_REAL_AI_MODEL"); override != "" {
		model = override
	}
	for ci, qaCourse := range []string{"en_ru", "es_ru"} {
		for i, mode := range []dailyjourney.Mode{dailyjourney.Picture, dailyjourney.Conversation} {
			t.Run(qaCourse+"/"+string(mode), func(t *testing.T) {
				r, adapter, user, course, id := journeyChatCourseFixture(t, int64(809010+ci*2+i), mode, qaCourse)
				proxy := env["OPENROUTER_SOCKS5_PROXY"]
				if providerName == "polza" {
					proxy = ""
				}
				provider := ai.NewServiceWithTimeoutAndSocks5Proxy(endpoint, model, key, "", 90*time.Second, proxy, zap.NewNop())
				load := func(name string) string {
					t.Helper()
					prompt, err := ai.LoadRenderedPromptFile("../../prompts/"+name+"-ru-"+strings.Split(course, "_")[0]+".txt", "ru", strings.Split(course, "_")[0], course)
					if err != nil {
						t.Fatal(err)
					}
					return prompt
				}
				provider.SetConversationQuestPromptForCourse(course, load("conversation-quest"))
				provider.SetConversationCorrectionPromptForCourse(course, load("conversation-correction"))
				provider.SetConversationNPCPromptForCourse(course, load("conversation-npc"))
				provider.SetPictureQuestPromptForCourse(course, load("picture-quest"))
				provider.SetPictureLumiPromptForCourse(course, load("picture-lumi"))
				r.aiService = provider
				fk, tasks := "scenario_id", "conversation_tasks"
				text := "Hola, quiero un café, por favor."
				if mode == dailyjourney.Picture {
					fk, tasks = "quest_id", "picture_quest_tasks"
					text = "Veo un gato en una casa."
					if _, err := r.db.Exec(`UPDATE picture_quests SET image_description='A cat sitting inside a house' WHERE id=?`, id); err != nil {
						t.Fatal(err)
					}
				}
				if _, err := r.db.Exec(`UPDATE `+tasks+` SET completion_criteria=? WHERE `+fk+`=?`, "Learner describes the cat or politely orders a coffee. Optional task: learner says goodbye.", id); err != nil {
					t.Fatal(err)
				}
				svc := service.NewDailyJourneyService(r.db, map[dailyjourney.Mode]service.DailyJourneyAdapter{mode: adapter})
				ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
				defer cancel()
				j, err := svc.Start(ctx, user, course)
				if err != nil {
					t.Fatal(err)
				}
				for n, kind := range []string{"start", "message"} {
					payload := json.RawMessage(`{}`)
					if kind == "message" {
						if qaCourse == "en_ru" {
							if mode == dailyjourney.Picture {
								text = "I see a cat in a house."
							} else {
								text = "Hello, I would like a coffee, please."
							}
						}
						payload, _ = json.Marshal(map[string]string{"text": text})
					}
					j, err = svc.Act(ctx, user, course, j.ID, j.Current().ID, dailyjourney.Action{ClientActionID: kind, Type: kind, ExpectedRevision: j.Revision, Payload: payload})
					if err != nil {
						t.Fatal(strings.ReplaceAll(err.Error(), key, "[redacted]"))
					}
					var state journeyChatState
					if err := json.Unmarshal(j.Current().State, &state); err != nil {
						t.Fatal(err)
					}
					if len(state.Messages) != 1+n*2 || strings.TrimSpace(state.Messages[len(state.Messages)-1].Content) == "" {
						t.Fatal("provider response not persisted")
					}
					t.Logf("%s: confirmed messages=%d turns=%d tokens=%d", kind, len(state.Messages), state.Turns, state.Tokens)
				}
			})
		}
	}
}
