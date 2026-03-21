package service

import (
	"reflect"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

// English regression: SRS state transitions must not depend on LearningConfig pair (only logs use it).
func TestSRSService_updateCardState_SameOutcomeForRUENAndRUES(t *testing.T) {
	logger := zap.NewNop()
	lcEN := config.LearningConfig{
		Pair: "ru-en", NativeLang: "ru", TargetLang: "en",
		AppCode: "english", GrammarBundleID: "en",
	}
	lcES := config.LearningConfig{
		Pair: "ru-es", NativeLang: "ru", TargetLang: "es",
		AppCode: "spanish", GrammarBundleID: "es",
	}

	sEn := &SRSService{learning: lcEN, logger: logger}
	sEs := &SRSService{learning: lcES, logger: logger}

	now := time.Date(2025, 3, 22, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name    string
		card    *models.UserCard
		quality models.Quality
	}{
		{
			name: "new correct",
			card: &models.UserCard{
				State: models.StateNew, EF: models.InitialEF, Direction: models.DirectionENtoRU,
			},
			quality: models.QualityGood,
		},
		{
			name: "review wrong",
			card: &models.UserCard{
				State: models.StateReview, EF: 2.1, Reps: 3, IntervalDays: 10,
				Direction: models.DirectionENtoRU,
			},
			quality: models.QualityWrong,
		},
		{
			name: "learning correct",
			card: &models.UserCard{
				State: models.StateLearning, EF: 2.0, LearningStep: 0, Direction: models.DirectionENtoRU,
			},
			quality: models.QualityGood,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := cloneUserCard(tc.card)
			b := cloneUserCard(tc.card)
			sEn.updateCardState(a, tc.quality, now)
			sEs.updateCardState(b, tc.quality, now)
			if !userCardsEqualForSRS(a, b) {
				t.Fatalf("cards differ:\n%+v\nvs\n%+v", a, b)
			}
		})
	}
}

func userCardsEqualForSRS(a, b *models.UserCard) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return reflect.DeepEqual(a, b)
}
