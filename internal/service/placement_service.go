package service

import (
	"context"
	cryptorand "crypto/rand"
	"database/sql"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"
	"tgbot-skeleton/internal/placement"
	"tgbot-skeleton/internal/repository"
)

type PlacementService struct {
	content  *repository.PlacementContentRepository
	attempts *repository.PlacementAttemptRepository
	courses  *repository.CourseRepository
}

func NewPlacementService(db *sql.DB, source string) *PlacementService {
	return &PlacementService{
		content: repository.NewPlacementContentRepository(db, source), attempts: repository.NewPlacementAttemptRepository(db), courses: repository.NewCourseRepository(db, nil),
	}
}

type PlacementSessionView struct {
	ID                  string               `json:"id"`
	CourseCode          string               `json:"course_code"`
	Status              string               `json:"status"`
	BankVersion         string               `json:"bank_version"`
	PolicyVersion       string               `json:"policy_version"`
	Questions           []placement.Question `json:"questions"`
	Answers             map[string]string    `json:"answers"`
	BaseClosed          bool                 `json:"base_closed"`
	Clarifying          bool                 `json:"clarifying"`
	AvailableChapterIDs []string             `json:"available_chapter_ids,omitempty"`
	Result              *placement.Result    `json:"result,omitempty"`
}

func placementView(s *repository.PlacementSession) *PlacementSessionView {
	questions := make([]placement.Question, 0, len(s.Snapshot.Items))
	for _, q := range s.Snapshot.Items {
		questions = append(questions, q.Public())
	}
	return &PlacementSessionView{ID: s.ID, CourseCode: s.CourseCode, Status: s.Status, BankVersion: s.BankVersion, PolicyVersion: s.PolicyVersion, Questions: questions, Answers: s.Snapshot.Answers, BaseClosed: s.Snapshot.BaseClosed, Clarifying: s.Snapshot.ClarificationLevel != "", Result: s.Result}
}
func (s *PlacementService) Start(ctx context.Context, user int64, course, key string, newAttempt bool) (*PlacementSessionView, error) {
	if len(key) < 8 || len(key) > 100 || strings.TrimSpace(key) != key {
		return nil, repository.ErrPlacementAnswer
	}
	uc, err := s.courses.EnsureUserCourse(ctx, user, course)
	if err != nil {
		return nil, err
	}
	session, err := s.attempts.Start(ctx, user, uc.ID, course, key, newAttempt, func(recent map[string]int) (string, placement.Snapshot, error) {
		bank, err := s.content.Load(ctx, course)
		if err != nil {
			return "", placement.Snapshot{}, fmt.Errorf("placement unavailable: %w", err)
		}
		var seed [8]byte
		if _, err := cryptorand.Read(seed[:]); err != nil {
			return "", placement.Snapshot{}, err
		}
		rng := rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(seed[:]))))
		items, repeated, err := placement.Select(bank, placement.Levels, recent, nil, rng)
		if err != nil {
			return "", placement.Snapshot{}, err
		}
		reserve, _, err := placement.Select(bank, placement.Levels, recent, items, rng)
		if err != nil {
			return "", placement.Snapshot{}, err
		}
		return bank.Version, placement.Snapshot{Items: items, Reserve: reserve, Skills: bank.Skills, Answers: map[string]string{}, RepeatedFamilies: repeated}, nil
	})
	if err != nil {
		return nil, err
	}
	return placementView(session), nil
}
func (s *PlacementService) Get(ctx context.Context, user int64, id string) (*PlacementSessionView, error) {
	v, e := s.attempts.Get(ctx, user, id)
	if e != nil {
		return nil, e
	}
	return placementView(v), nil
}
func (s *PlacementService) Answer(ctx context.Context, user int64, id, qid string, answer *string) (*PlacementSessionView, error) {
	if answer == nil {
		return nil, repository.ErrPlacementAnswer
	}
	session, err := s.attempts.Update(ctx, user, id, func(v *repository.PlacementSession, _ *sql.Tx) error {
		if v.Status != "active" {
			if old, ok := v.Snapshot.Answers[qid]; ok && old == *answer {
				return nil
			}
			return repository.ErrPlacementConflict
		}
		if v.PolicyVersion != placement.PolicyVersion {
			return repository.ErrPlacementExpired
		}
		index := -1
		for i, q := range v.Snapshot.Items {
			if q.ID == qid {
				index = i
				break
			}
		}
		if index < 0 {
			return repository.ErrPlacementAnswer
		}
		if v.Snapshot.BaseClosed && index < 30 {
			if old, ok := v.Snapshot.Answers[qid]; ok && old == *answer {
				return nil
			}
			return repository.ErrPlacementConflict
		}
		valid := *answer == ""
		for _, c := range v.Snapshot.Items[index].Choices {
			if c.ID == *answer {
				valid = true
			}
		}
		if !valid {
			return repository.ErrPlacementAnswer
		}
		v.Snapshot.Answers[qid] = *answer
		if len(v.Snapshot.Answers) == 30 && !v.Snapshot.BaseClosed {
			v.Snapshot.BaseClosed = true
			level := placement.Clarify(&v.Snapshot)
			if level != "" {
				v.Snapshot.ClarificationLevel = level
				count := 0
				for _, q := range v.Snapshot.Reserve {
					if q.Level == level {
						v.Snapshot.Items = append(v.Snapshot.Items, q)
						count++
					}
				}
				if count != 6 {
					return fmt.Errorf("invalid pinned clarification reserve")
				}
			}
			v.Snapshot.Reserve = nil
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return placementView(session), nil
}
func (s *PlacementService) Finish(ctx context.Context, user int64, id string, grammar func(string) *GrammarService) (*PlacementSessionView, error) {
	session, err := s.attempts.Update(ctx, user, id, func(v *repository.PlacementSession, tx *sql.Tx) error {
		if v.Status == "completed" {
			return nil
		}
		if v.PolicyVersion != placement.PolicyVersion {
			return repository.ErrPlacementExpired
		}
		if !v.Snapshot.BaseClosed || len(v.Snapshot.Answers) != len(v.Snapshot.Items) {
			return repository.ErrPlacementConflict
		}
		r := placement.Grade(&v.Snapshot)
		g := grammar(v.CourseCode)
		if g == nil {
			return fmt.Errorf("course grammar unavailable")
		}
		if r.Level != "below_a1" {
			var e error
			r.OpenedSections, e = g.OpenPublishedSectionsThroughLevel(r.Level)
			if e != nil {
				return e
			}
		}
		score := 0
		if r.Total > 0 {
			score = r.Correct * 100 / r.Total
		}
		if err := repository.SaveDiagnosticPlacementAccessTx(ctx, tx, v.UserCourseID, score, r.Total, r.OpenedSections); err != nil {
			return err
		}
		v.Result = &r
		v.Status = "completed"
		return nil
	})
	if err != nil {
		return nil, err
	}
	return placementView(session), nil
}
func (s *PlacementService) History(ctx context.Context, user int64, course string) ([]*PlacementSessionView, error) {
	history, err := s.attempts.History(ctx, user, course)
	if err != nil {
		return nil, err
	}
	out := []*PlacementSessionView{}
	for _, h := range history {
		v := placementView(h)
		v.Questions = nil
		v.Answers = nil
		out = append(out, v)
	}
	return out, nil
}
