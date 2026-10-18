package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"
)

func journeyChatFixture(t *testing.T, telegram int64, mode dailyjourney.Mode) (*Router, *dailyJourneyChatAdapter, int64, string, int64) {
	return journeyChatCourseFixture(t, telegram, mode, "es_ru")
}
func journeyChatCourseFixture(t *testing.T, telegram int64, mode dailyjourney.Mode, course string) (*Router, *dailyJourneyChatAdapter, int64, string, int64) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	user, err := repository.NewUserRepository(db, zap.NewNop()).GetOrCreateUser(telegram)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE users SET subscription_tier='pro' WHERE id=?`, user.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.NewCourseRepository(db, zap.NewNop()).EnsureUserCourse(context.Background(), user.ID, course); err != nil {
		t.Fatal(err)
	}
	var courseID, districtID int64
	if err := db.QueryRow(`SELECT c.id,d.id FROM courses c JOIN districts d ON d.course_id=c.id WHERE c.code=? AND d.level_code='A0'`, course).Scan(&courseID, &districtID); err != nil {
		t.Fatal(err)
	}
	code := fmt.Sprintf("journey-chat-%d", telegram)
	var id int64
	if mode == dailyjourney.Picture {
		err = db.QueryRow(`INSERT INTO picture_quests(course_id,district_id,code,cefr_level,title,image_url,image_description,max_turns,token_budget,status) VALUES(?, ?,?,'A0','Picture','/test.jpg','PRIVATE PICTURE GROUND TRUTH',2,100000,'active') RETURNING id`, courseID, districtID, code).Scan(&id)
	} else {
		err = db.QueryRow(`INSERT INTO conversation_scenarios(course_id,district_id,code,place_type,cefr_level,title,npc_name,npc_persona,scene_setup,is_quest,max_turns,token_budget,status) VALUES(?,?,?,'cafe','A0','Cafe','Marta','PRIVATE NPC PERSONA','At a cafe',true,2,100000,'active') RETURNING id`, courseID, districtID, code).Scan(&id)
	}
	if err != nil {
		t.Fatal(err)
	}
	tasks, fk := "conversation_tasks", "scenario_id"
	if mode == dailyjourney.Picture {
		tasks = "picture_quest_tasks"
		fk = "quest_id"
	}
	for i, required := range []bool{true, false} {
		taskCode := "required"
		if !required {
			taskCode = "optional"
		}
		if _, err := db.Exec(`INSERT INTO `+tasks+`(`+fk+`,code,sort_order,is_required,title,completion_criteria) VALUES(?,?,?,?,?,'PRIVATE TASK CRITERIA')`, id, taskCode, i, required, taskCode); err != nil {
			t.Fatal(err)
		}
	}
	r := NewRouter(zap.NewNop(), &config.Config{}, db, nil, nil, nil, nil)
	adapter := &dailyJourneyChatAdapter{router: r, mode: mode}
	return r, adapter, user.ID, course, id
}
func TestDailyJourneyChatBoundSessionsAndAtomicResult(t *testing.T) {
	for i, mode := range []dailyjourney.Mode{dailyjourney.Picture, dailyjourney.Conversation} {
		t.Run(string(mode), func(t *testing.T) {
			r, adapter, user, course, contentID := journeyChatFixture(t, int64(808701+i), mode)
			db := r.db
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			var uc int64
			if err := db.QueryRow(`SELECT uc.id FROM user_courses uc JOIN courses c ON c.id=uc.course_id WHERE uc.user_id=? AND c.code=?`, user, course).Scan(&uc); err != nil {
				t.Fatal(err)
			}
			var normalID int64
			if mode == dailyjourney.Picture {
				s, _, err := r.pictureQuestRepo.StartSession(ctx, uc, contentID)
				if err != nil {
					t.Fatal(err)
				}
				normalID = s.ID
			} else {
				s, _, err := r.conversationRepo.StartSession(ctx, uc, contentID)
				if err != nil {
					t.Fatal(err)
				}
				normalID = s.ID
			}
			var calls atomic.Int32
			adapter.run = func(ctx context.Context, in ai.ConversationTurnSplitInput) (*ai.ChatTurnResult, error) {
				calls.Add(1)
				var n int
				if err := db.QueryRowContext(ctx, `SELECT 1`).Scan(&n); err != nil {
					return nil, err
				}
				if !in.EvaluateQuest {
					return &ai.ChatTurnResult{VisibleContent: "Hola", PromptTokens: 1, CompletionTokens: 1}, nil
				}
				return &ai.ChatTurnResult{VisibleContent: "Muy bien", CompletedTaskCodes: []string{"required", "unknown"}, AllDoneSignal: true, PromptTokens: 2, CompletionTokens: 3}, nil
			}
			svc := service.NewDailyJourneyService(db, map[dailyjourney.Mode]service.DailyJourneyAdapter{mode: adapter})
			j, err := svc.Start(ctx, user, course)
			if err != nil {
				t.Fatal(err)
			}
			oldMax := db.Stats().MaxOpenConnections
			db.SetMaxOpenConns(1)
			t.Cleanup(func() { db.SetMaxOpenConns(oldMax) })
			act := func(key, kind string, payload string) {
				t.Helper()
				j, err = svc.Act(ctx, user, course, j.ID, j.Current().ID, dailyjourney.Action{ClientActionID: key, Type: kind, ExpectedRevision: j.Revision, Payload: json.RawMessage(payload)})
				if err != nil {
					t.Fatal(key, err)
				}
			}
			act("start", "start", `{}`)
			var state journeyChatState
			_ = json.Unmarshal(j.Current().State, &state)
			sessionID := state.SessionID
			if sessionID == normalID || len(state.Messages) != 1 || state.Turns != 0 {
				t.Fatal("ordinary session reused", state)
			}
			act("resume", "start", `{}`)
			if calls.Load() != 1 {
				t.Fatal("resume called provider")
			}
			view, err := adapter.Exercise(ctx, user, j, j.Current())
			if err != nil {
				t.Fatal(err)
			}
			public, _ := json.Marshal(view)
			if strings.Contains(string(public), "PRIVATE") {
				t.Fatal("hidden criteria/prompt leaked")
			}
			// A complete real subject write followed by failure must roll back messages,
			// task progress, exercise/events/statistics and the journey together.
			step := j.Current()
			action := dailyjourney.Action{ClientActionID: "rollback", Type: "message", ExpectedRevision: j.Revision, Payload: json.RawMessage(`{"text":"Hola"}`)}
			repo := repository.NewDailyJourneyRepository(db)
			reservation, err := repo.ReserveAI(ctx, user, course, j.ID, step.ID, action, time.Minute, func(tx *sql.Tx, j *dailyjourney.Journey, s *dailyjourney.Step, _ json.RawMessage) (json.RawMessage, error) {
				return adapter.PrepareAI(ctx, tx, user, j, s, action)
			})
			if err != nil {
				t.Fatal(err)
			}
			result, err := adapter.RunAI(ctx, reservation.Input)
			if err != nil {
				t.Fatal(err)
			}
			forced := errors.New("rollback chat subject")
			_, err = repo.CompleteAI(ctx, user, course, j.ID, step.ID, action, reservation.Token, func(tx *sql.Tx, j *dailyjourney.Journey, s *dailyjourney.Step) error {
				if err := adapter.ApplyAI(ctx, tx, user, j, s, action, result); err != nil {
					return err
				}
				return forced
			})
			if !errors.Is(err, forced) {
				t.Fatal(err)
			}
			if err := repo.FailAI(ctx, user, course, j.ID, action.ClientActionID, reservation.Token); err != nil {
				t.Fatal(err)
			}
			table := "conversation_messages"
			if mode == dailyjourney.Picture {
				table = "picture_quest_messages"
			}
			var n int
			if err := db.QueryRow(`SELECT COUNT(*) FROM `+table+` WHERE session_id=?`, sessionID).Scan(&n); err != nil || n != 1 {
				t.Fatal("messages survived rollback", n, err)
			}
			if err := db.QueryRow(`SELECT COUNT(*) FROM exercise_attempts WHERE user_course_id=?`, uc).Scan(&n); err != nil || n != 0 {
				t.Fatal("credit survived rollback", n, err)
			}
			before := calls.Load()
			action.ClientActionID = "message"
			var wg sync.WaitGroup
			errs := make(chan error, 8)
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func() { defer wg.Done(); _, err := svc.Act(ctx, user, course, j.ID, step.ID, action); errs <- err }()
			}
			wg.Wait()
			close(errs)
			for err := range errs {
				if err != nil && !errors.Is(err, dailyjourney.ErrProcessing) {
					t.Fatal(err)
				}
			}
			if calls.Load() != before+1 {
				t.Fatal("duplicate provider turn")
			}
			j, err = svc.Get(ctx, user, course, j.ID)
			if err != nil {
				t.Fatal(err)
			}
			_ = json.Unmarshal(j.Current().State, &state)
			if !state.Done || !state.Passed || len(state.Completed) != 1 || len(state.Messages) != 3 {
				t.Fatal("task/budget result", state)
			}
			if j.Current().Status != "active" {
				t.Fatal("breakdown auto-dismissed")
			}
			if err := db.QueryRow(`SELECT COUNT(*) FROM exercise_attempts WHERE user_course_id=?`, uc).Scan(&n); err != nil || n != 1 {
				t.Fatal("quest credited twice", n, err)
			}
			if mode == dailyjourney.Picture {
				s, err := r.pictureQuestRepo.GetOpenSession(ctx, uc, contentID)
				if err != nil || s.ID != normalID {
					t.Fatal("ordinary active session lost", err)
				}
				hidden, err := r.pictureQuestRepo.GetSession(ctx, sessionID, uc)
				if err != nil || hidden != nil {
					t.Fatal("ordinary endpoint exposes journey session", err)
				}
				codes, err := r.pictureQuestRepo.CompletedQuestCodes(ctx, uc, func() int64 {
					var c int64
					_ = db.QueryRow(`SELECT course_id FROM picture_quests WHERE id=?`, contentID).Scan(&c)
					return c
				}())
				if err != nil || len(codes) != 0 {
					t.Fatal("mandatory pass became 100%", codes, err)
				}
			} else {
				hidden, err := r.conversationRepo.GetSessionForUser(ctx, sessionID, user)
				if err != nil || hidden != nil {
					t.Fatal("ordinary endpoint exposes journey session", err)
				}
			}
			act("ack", "ack_feedback", `{}`)
			if j.Status != "completed" {
				t.Fatal(j.Status)
			}
		})
	}
}
func TestDailyJourneyChatBudgetAndTierGuards(t *testing.T) {
	r, adapter, user, course, _ := journeyChatFixture(t, 808703, dailyjourney.Picture)
	ctx := context.Background()
	adapter.run = func(context.Context, ai.ConversationTurnSplitInput) (*ai.ChatTurnResult, error) {
		return &ai.ChatTurnResult{VisibleContent: "Hola", CompletionTokens: 1}, nil
	}
	if _, err := r.db.Exec(`UPDATE users SET subscription_tier='free' WHERE id=?`, user); err != nil {
		t.Fatal(err)
	}
	items, err := adapter.Candidates(ctx, r.db, user, course)
	if err != nil || len(items) != 0 {
		t.Fatal("free access", items, err)
	}
	if _, err := r.db.Exec(`UPDATE users SET subscription_tier='pro' WHERE id=?`, user); err != nil {
		t.Fatal(err)
	}
	svc := service.NewDailyJourneyService(r.db, map[dailyjourney.Mode]service.DailyJourneyAdapter{dailyjourney.Picture: adapter})
	j, err := svc.Start(ctx, user, course)
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"start", "message", "message"} {
		payload := json.RawMessage(`{}`)
		if kind == "message" {
			payload = json.RawMessage(`{"text":"Hola"}`)
		}
		j, err = svc.Act(ctx, user, course, j.ID, j.Current().ID, dailyjourney.Action{ClientActionID: fmt.Sprint(kind, j.Revision), Type: kind, ExpectedRevision: j.Revision, Payload: payload})
		if err != nil {
			t.Fatal(err)
		}
	}
	var state journeyChatState
	_ = json.Unmarshal(j.Current().State, &state)
	if !state.Done || state.Passed || !state.Exhausted {
		t.Fatal("budget masquerades as pass", state)
	}
	j, err = svc.Act(ctx, user, course, j.ID, j.Current().ID, dailyjourney.Action{ClientActionID: "finish", Type: "ack_feedback", ExpectedRevision: j.Revision, Payload: json.RawMessage(`{}`)})
	if err != nil || j.Status != "completed" {
		t.Fatal(j, err)
	}
	var n int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM exercise_attempts ea JOIN user_courses uc ON uc.id=ea.user_course_id WHERE uc.user_id=?`, user).Scan(&n); err != nil || n != 0 {
		t.Fatal("unsuccessful attempt granted quest pass", n, err)
	}
}

func TestDailyJourneyChatRevalidatesAccessAndContent(t *testing.T) {
	for i, mode := range []dailyjourney.Mode{dailyjourney.Picture, dailyjourney.Conversation} {
		t.Run(string(mode), func(t *testing.T) {
			r, adapter, user, course, contentID := journeyChatFixture(t, int64(808704+i), mode)
			ctx := context.Background()
			adapter.run = func(context.Context, ai.ConversationTurnSplitInput) (*ai.ChatTurnResult, error) {
				return &ai.ChatTurnResult{VisibleContent: "Hola"}, nil
			}
			svc := service.NewDailyJourneyService(r.db, map[dailyjourney.Mode]service.DailyJourneyAdapter{mode: adapter})
			j, err := svc.Start(ctx, user, course)
			if err != nil {
				t.Fatal(err)
			}
			step := j.Current()
			if err := adapter.Validate(ctx, r.db, user, course, step); err != nil {
				t.Fatal(err)
			}
			if _, err := r.db.Exec(`UPDATE users SET subscription_tier='free' WHERE id=?`, user); err != nil {
				t.Fatal(err)
			}
			if err := adapter.Validate(ctx, r.db, user, course, step); !errors.Is(err, service.ErrDailyJourneyContentUnavailable) {
				t.Fatal("revoked tier accepted", err)
			}
			if _, err := r.db.Exec(`UPDATE users SET subscription_tier='pro' WHERE id=?`, user); err != nil {
				t.Fatal(err)
			}
			table := "conversation_scenarios"
			if mode == dailyjourney.Picture {
				table = "picture_quests"
			}
			if _, err := r.db.Exec(`UPDATE `+table+` SET title='Changed content' WHERE id=?`, contentID); err != nil {
				t.Fatal(err)
			}
			if err := adapter.Validate(ctx, r.db, user, course, step); !errors.Is(err, service.ErrDailyJourneyContentUnavailable) {
				t.Fatal("changed snapshot accepted", err)
			}
		})
	}
}

func TestDailyJourneyChatCourseTierMatrix(t *testing.T) {
	for ci, course := range []string{"en_ru", "es_ru"} {
		for mi, mode := range []dailyjourney.Mode{dailyjourney.Picture, dailyjourney.Conversation} {
			t.Run(course+"/"+string(mode), func(t *testing.T) {
				r, a, user, _, id := journeyChatCourseFixture(t, int64(808720+ci*2+mi), mode, course)
				a.run = func(context.Context, ai.ConversationTurnSplitInput) (*ai.ChatTurnResult, error) {
					return &ai.ChatTurnResult{VisibleContent: "Hello"}, nil
				}
				ctx := context.Background()
				code := fmt.Sprintf("journey-chat-%d", 808720+ci*2+mi)
				for _, tier := range []string{"free", "pro", "pro_plus"} {
					if _, err := r.db.Exec(`UPDATE users SET subscription_tier=? WHERE id=?`, tier, user); err != nil {
						t.Fatal(err)
					}
					items, err := a.Candidates(ctx, r.db, user, course)
					if err != nil {
						t.Fatal(err)
					}
					found := false
					for _, item := range items {
						var ref journeyChatRef
						_ = json.Unmarshal(item.ContentRef, &ref)
						if ref.Code == code {
							found = true
						}
					}
					if found != (tier != "free") {
						t.Fatalf("%s access=%v", tier, found)
					}
				}
				table := "conversation_scenarios"
				if mode == dailyjourney.Picture {
					table = "picture_quests"
				}
				if _, err := r.db.Exec(`UPDATE `+table+` SET cefr_level='C2' WHERE id=?`, id); err != nil {
					t.Fatal(err)
				}
				_, err := a.content(ctx, r.db, user, course, journeyChatRef{Code: code}, false)
				if !errors.Is(err, service.ErrDailyJourneyContentUnavailable) {
					t.Fatal("closed level accepted", err)
				}
			})
		}
	}
}

func TestDailyJourneyNPCPrerequisiteCooldownAndPassedGuards(t *testing.T) {
	r, a, user, course, id := journeyChatFixture(t, 808730, dailyjourney.Conversation)
	a.run = func(context.Context, ai.ConversationTurnSplitInput) (*ai.ChatTurnResult, error) {
		return &ai.ChatTurnResult{VisibleContent: "Hello"}, nil
	}
	ctx := context.Background()
	code := "journey-chat-808730"
	var prior, uc int64
	if err := r.db.QueryRow(`SELECT uc.id FROM user_courses uc JOIN courses c ON c.id=uc.course_id WHERE uc.user_id=? AND c.code=?`, user, course).Scan(&uc); err != nil {
		t.Fatal(err)
	}
	if err := r.db.QueryRow(`INSERT INTO conversation_scenarios(course_id,district_id,code,place_type,cefr_level,title,npc_name,npc_persona,scene_setup,is_quest,max_turns,token_budget,status) SELECT course_id,district_id,'matrix-prior','cafe','A0','Prior','Marta','Persona','Scene',true,2,10000,'active' FROM conversation_scenarios WHERE id=? RETURNING id`, id).Scan(&prior); err != nil {
		t.Fatal(err)
	}
	if _, err := r.db.Exec(`UPDATE conversation_scenarios SET prerequisite_code='matrix-prior' WHERE id=?`, id); err != nil {
		t.Fatal(err)
	}
	check := func(allowed bool) {
		t.Helper()
		_, err := a.content(ctx, r.db, user, course, journeyChatRef{Code: code}, false)
		if allowed && err != nil {
			t.Fatal(err)
		}
		if !allowed && !errors.Is(err, service.ErrDailyJourneyContentUnavailable) {
			t.Fatal("locked NPC accepted", err)
		}
	}
	check(false)
	session, _, err := r.conversationRepo.StartSession(ctx, uc, prior)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.db.Exec(`UPDATE conversation_sessions SET status='completed',completed_at=CURRENT_TIMESTAMP WHERE id=?`, session.ID); err != nil {
		t.Fatal(err)
	}
	check(false)
	if _, err := r.db.Exec(`UPDATE conversation_sessions SET completed_at=CURRENT_TIMESTAMP-INTERVAL '2 hours' WHERE id=?`, session.ID); err != nil {
		t.Fatal(err)
	}
	check(true)
	own, _, err := r.conversationRepo.StartSession(ctx, uc, id)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.db.Exec(`UPDATE conversation_sessions SET status='completed',completed_at=CURRENT_TIMESTAMP WHERE id=?`, own.ID); err != nil {
		t.Fatal(err)
	}
	check(false)
}

func TestDailyJourneyChatSkipClosesBoundSession(t *testing.T) {
	for i, mode := range []dailyjourney.Mode{dailyjourney.Picture, dailyjourney.Conversation} {
		t.Run(string(mode), func(t *testing.T) {
			r, a, user, course, _ := journeyChatFixture(t, int64(808731+i), mode)
			a.run = func(context.Context, ai.ConversationTurnSplitInput) (*ai.ChatTurnResult, error) {
				return &ai.ChatTurnResult{VisibleContent: "Hello"}, nil
			}
			svc := service.NewDailyJourneyService(r.db, map[dailyjourney.Mode]service.DailyJourneyAdapter{mode: a})
			ctx := context.Background()
			j, err := svc.Start(ctx, user, course)
			if err != nil {
				t.Fatal(err)
			}
			id := j.Current().ID
			j, err = svc.Act(ctx, user, course, j.ID, id, dailyjourney.Action{ClientActionID: "open", ExpectedRevision: j.Revision, Type: "start", Payload: json.RawMessage(`{}`)})
			if err != nil {
				t.Fatal(err)
			}
			_, err = svc.Act(ctx, user, course, j.ID, id, dailyjourney.Action{ClientActionID: "skip", ExpectedRevision: j.Revision, Type: "skip", Payload: json.RawMessage(`{}`)})
			if err != nil {
				t.Fatal(err)
			}
			table := "conversation_sessions"
			if mode == dailyjourney.Picture {
				table = "picture_quest_sessions"
			}
			var status string
			if err := r.db.QueryRow(`SELECT status FROM `+table+` WHERE daily_journey_step_id=?`, id).Scan(&status); err != nil || status != "abandoned" {
				t.Fatal("skip left live session", status, err)
			}
			var credit int
			if err := r.db.QueryRow(`SELECT COUNT(*) FROM exercise_attempts ea JOIN user_courses uc ON uc.id=ea.user_course_id WHERE uc.user_id=?`, user).Scan(&credit); err != nil || credit != 0 {
				t.Fatal("skip credited quest", credit, err)
			}
		})
	}
}
