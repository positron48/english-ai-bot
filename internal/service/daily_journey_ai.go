package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"tgbot-skeleton/internal/dailyjourney"
	"time"
)

// AI adapters have a strict three-phase contract. PrepareAI only reads through
// tx; RunAI owns provider I/O with no transaction; ApplyAI writes only via tx.
// The private input is durable for retry and must not be exposed as an exercise.
type DailyJourneyAIAdapter interface {
	DailyJourneyAdapter
	RequiresAI(dailyjourney.Action) bool
	PrepareAI(context.Context, *sql.Tx, int64, *dailyjourney.Journey, *dailyjourney.Step, dailyjourney.Action) (json.RawMessage, error)
	RunAI(context.Context, json.RawMessage) (json.RawMessage, error)
	ApplyAI(context.Context, *sql.Tx, int64, *dailyjourney.Journey, *dailyjourney.Step, dailyjourney.Action, json.RawMessage) error
}

func (s *DailyJourneyService) actAI(ctx context.Context, user int64, course, id, stepID string, action dailyjourney.Action, adapter DailyJourneyAIAdapter) (*dailyjourney.Journey, error) {
	reservation, err := s.repo.ReserveAI(ctx, user, course, id, stepID, action, 3*time.Minute, func(tx *sql.Tx, j *dailyjourney.Journey, step *dailyjourney.Step, previous json.RawMessage) (json.RawMessage, error) {
		if j.AlgorithmVersion != dailyjourney.AlgorithmVersion {
			return nil, dailyjourney.ErrConflict
		}
		if err := adapter.Validate(ctx, tx, user, course, step); err != nil {
			return nil, err
		}
		if action.Type != "start" && step.Status != "active" {
			return nil, dailyjourney.ErrInvalidAction
		}
		if len(previous) > 0 {
			return previous, nil
		}
		return adapter.PrepareAI(ctx, tx, user, j, step, action)
	})
	if err != nil {
		if errors.Is(err, ErrDailyJourneyContentUnavailable) {
			return s.repo.Apply(ctx, user, course, id, stepID, action, func(tx *sql.Tx, j *dailyjourney.Journey, step *dailyjourney.Step) error {
				if accessErr := adapter.Validate(ctx, tx, user, course, step); !errors.Is(accessErr, ErrDailyJourneyContentUnavailable) {
					if accessErr != nil {
						return accessErr
					}
					return dailyjourney.ErrConflict
				}
				step.Status = "unavailable"
				reason := "Задание больше недоступно"
				step.Result = &reason
				return nil
			})
		}
		return nil, err
	}
	if reservation.Result != nil {
		return reservation.Result, nil
	}
	providerCtx, cancel := context.WithTimeout(ctx, 150*time.Second)
	result, err := adapter.RunAI(providerCtx, reservation.Input)
	cancel()
	if err == nil && (!json.Valid(result) || len(result) > 2<<20) {
		err = fmt.Errorf("invalid daily journey AI result")
	}
	if err != nil {
		// HTTP cancellation must not strand a lease when failure is already known.
		cleanupCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		cleanupErr := s.repo.FailAI(cleanupCtx, user, course, id, action.ClientActionID, reservation.Token)
		return nil, errors.Join(fmt.Errorf("%w: %w", dailyjourney.ErrAIProvider, err), cleanupErr)
	}
	journey, completeErr := s.repo.CompleteAI(ctx, user, course, id, stepID, action, reservation.Token, func(tx *sql.Tx, j *dailyjourney.Journey, step *dailyjourney.Step) error {
		if j.AlgorithmVersion != dailyjourney.AlgorithmVersion {
			return dailyjourney.ErrConflict
		}
		if err := adapter.Validate(ctx, tx, user, course, step); err != nil {
			if errors.Is(err, ErrDailyJourneyContentUnavailable) {
				step.Status = "unavailable"
				reason := "Задание больше недоступно"
				step.Result = &reason
				return nil
			}
			return err
		}
		if err := adapter.ApplyAI(ctx, tx, user, j, step, action, result); err != nil {
			return err
		}
		if action.Type == "start" {
			step.Status = "active"
		}
		return nil
	})
	if completeErr != nil {
		cleanupCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		cleanupErr := s.repo.FailAI(cleanupCtx, user, course, id, action.ClientActionID, reservation.Token)
		return nil, errors.Join(completeErr, cleanupErr)
	}
	return journey, nil
}
