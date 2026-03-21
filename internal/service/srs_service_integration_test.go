package service

import (
	"database/sql"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/testutil"
)

func setupSRSServiceTestDB(t *testing.T) (*sql.DB, *repository.UserRepository, *repository.UserCardRepository) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)

	return db, userRepo, userCardRepo
}

func TestSRSService_GradeCard_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo := setupSRSServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(111)

	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "grade", "to grade")
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (1, $1, $2, $3, $4)",
		"grade", 0, "оценить", "to grade")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create a user card
	now := time.Now()
	card := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
		Reps:           0,
		IntervalDays:   0,
		LearningStep:   0,
		LapseCount:     0,
		NextDueAt:      &now,
	}
	id, err := userCardRepo.CreateUserCard(card)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	// Get the card
	userCard, err := userCardRepo.GetUserCard(id)
	if err != nil {
		t.Fatalf("Failed to get user card: %v", err)
	}

	// Create SRS service
	service := NewSRSService(userCardRepo, config.DefaultLearningConfig(), logger)

	// Grade the card with correct answer
	attemptData := models.AttemptData{
		Correct:      true,
		EarlyReveal:  false,
		AnswerTimeMS: 3000, // Normal speed
	}

	err = service.GradeCard(userCard, attemptData)
	if err != nil {
		t.Fatalf("GradeCard() error = %v", err)
	}

	// Verify the card was updated
	updated, err := userCardRepo.GetUserCard(id)
	if err != nil {
		t.Fatalf("Failed to get updated user card: %v", err)
	}
	if updated.State != models.StateLearning {
		t.Errorf("Expected State %v, got %v", models.StateLearning, updated.State)
	}
	if updated.LastQuality == nil {
		t.Error("LastQuality should be set after grading")
	}
}

func TestSRSService_GradeCard_WrongAnswer(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo := setupSRSServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(222)

	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "wrong", "wrong")
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (1, $1, $2, $3, $4)",
		"wrong", 0, "неправильно", "wrong")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create a user card in review state
	now := time.Now()
	card := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		Reps:           5,
		IntervalDays:   10,
		NextDueAt:      &now,
	}
	id, err := userCardRepo.CreateUserCard(card)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	// Get the card
	userCard, err := userCardRepo.GetUserCard(id)
	if err != nil {
		t.Fatalf("Failed to get user card: %v", err)
	}

	// Create SRS service
	service := NewSRSService(userCardRepo, config.DefaultLearningConfig(), logger)

	// Grade the card with wrong answer
	attemptData := models.AttemptData{
		Correct: false,
	}

	err = service.GradeCard(userCard, attemptData)
	if err != nil {
		t.Fatalf("GradeCard() error = %v", err)
	}

	// Verify the card uses gentle approach (stays in review, interval reduced)
	updated, err := userCardRepo.GetUserCard(id)
	if err != nil {
		t.Fatalf("Failed to get updated user card: %v", err)
	}
	// Should stay in review for first error
	if updated.State != models.StateReview {
		t.Errorf("Expected State %v, got %v (should stay in review for first error)", models.StateReview, updated.State)
	}
	// Interval should be reduced (10 / 2 = 5)
	if updated.IntervalDays != 5 {
		t.Errorf("Expected IntervalDays 5 (10/2), got %d", updated.IntervalDays)
	}
	// Reps should be preserved
	if updated.Reps != 5 {
		t.Errorf("Expected Reps 5 (preserved), got %d", updated.Reps)
	}
	if updated.LapseCount != 1 {
		t.Errorf("Expected LapseCount 1, got %d", updated.LapseCount)
	}
	// EF should be reduced
	if updated.EF >= 2.0 {
		t.Errorf("Expected EF < 2.0, got %f", updated.EF)
	}
}

// TestSRSService_GradeCard_NextDueAtPersisted проверяет, что next_due_at после GradeCard
// сохраняется и читается из Postgres корректно (не обнуляется и не «всегда 24h»).
func TestSRSService_GradeCard_NextDueAtPersisted(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo := setupSRSServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(333)

	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "persist", "to persist")
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (1, $1, $2, $3, $4)",
		"persist", 0, "сохранять", "to persist")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	now := time.Now()
	card := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		Reps:           2,
		IntervalDays:   6,
		NextDueAt:      &now,
	}
	id, err := userCardRepo.CreateUserCard(card)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	userCard, err := userCardRepo.GetUserCard(id)
	if err != nil {
		t.Fatalf("Failed to get user card: %v", err)
	}

	svc := NewSRSService(userCardRepo, config.DefaultLearningConfig(), logger)
	attemptData := models.AttemptData{Correct: true, EarlyReveal: false, AnswerTimeMS: 3000}
	if err := svc.GradeCard(userCard, attemptData); err != nil {
		t.Fatalf("GradeCard() error = %v", err)
	}

	// Перечитываем из БД — next_due_at должен быть далёкой датой (интервал вырос), не нулём и не now+24h
	updated, err := userCardRepo.GetUserCard(id)
	if err != nil {
		t.Fatalf("Failed to get updated user card: %v", err)
	}
	if updated.NextDueAt == nil {
		t.Fatal("NextDueAt must be set after correct answer (persisted from Postgres)")
	}
	intervalHours := updated.NextDueAt.Sub(now).Hours()
	if intervalHours < 24 {
		t.Errorf("after correct answer next_due_at should be > 24h ahead, got %.1fh (possible Postgres read bug)", intervalHours)
	}
	// При Reps=3 и IntervalDays=6 новый интервал = ceil(6*EF) дней, т.е. много дней
	if updated.IntervalDays < 2 {
		t.Errorf("IntervalDays should grow after correct answer, got %d", updated.IntervalDays)
	}
}

// TestSRSService_GradeCard_ReviewWithNonZeroData runs several correct answers on a card
// that starts in review with non-zero interval/reps (realistic "healthy" state).
// Asserts that interval and next_due_at grow and are persisted after each answer.
func TestSRSService_GradeCard_ReviewWithNonZeroData(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo := setupSRSServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(444)

	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "grow", "to grow")
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (1, $1, $2, $3, $4)",
		"grow", 0, "расти", "to grow")
	if err != nil {
		t.Fatalf("create training card: %v", err)
	}

	now := time.Now()
	card := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		Reps:           2,
		IntervalDays:   6,
		NextDueAt:      &now,
	}
	id, err := userCardRepo.CreateUserCard(card)
	if err != nil {
		t.Fatalf("create user card: %v", err)
	}

	svc := NewSRSService(userCardRepo, config.DefaultLearningConfig(), logger)
	attempt := models.AttemptData{Correct: true, EarlyReveal: false, AnswerTimeMS: 3000}

	// First correct: Reps 2->3, newInterval = ceil(6*EF) = 12, next_due = now+12d
	uc, err := userCardRepo.GetUserCard(id)
	if err != nil || uc == nil {
		t.Fatalf("get user card: %v", err)
	}
	if err := svc.GradeCard(uc, attempt); err != nil {
		t.Fatalf("GradeCard 1: %v", err)
	}
	after1, _ := userCardRepo.GetUserCard(id)
	if after1.NextDueAt == nil {
		t.Fatal("after first correct: NextDueAt must be set")
	}
	interval1 := after1.IntervalDays
	due1 := after1.NextDueAt.Sub(now).Hours() / 24

	// Second correct: newInterval = ceil(12*EF), next_due = now2 + ...
	uc2, _ := userCardRepo.GetUserCard(id)
	if err := svc.GradeCard(uc2, attempt); err != nil {
		t.Fatalf("GradeCard 2: %v", err)
	}
	after2, _ := userCardRepo.GetUserCard(id)
	if after2.NextDueAt == nil {
		t.Fatal("after second correct: NextDueAt must be set")
	}
	interval2 := after2.IntervalDays

	if interval2 <= interval1 {
		t.Errorf("interval should grow: after1=%d after2=%d", interval1, interval2)
	}
	if due1 < 1 {
		t.Errorf("after first correct next_due_at should be at least 1 day ahead, got %.2f days", due1)
	}
}

// TestSRSService_GradeCard_ReviewWithZeroReps simulates "migrated" or broken state:
// state=review but reps=0 (and optionally interval_days=0). After one correct we get interval 1;
// after second correct we get reps=2 and interval = ceil(1*EF) >= 2. So interval should grow.
func TestSRSService_GradeCard_ReviewWithZeroReps(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo := setupSRSServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(555)

	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "broken", "broken")
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (1, $1, $2, $3, $4)",
		"broken", 0, "сломан", "broken")
	if err != nil {
		t.Fatalf("create training card: %v", err)
	}

	now := time.Now()
	card := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		Reps:           0, // broken: review with 0 reps (e.g. after bad migration)
		IntervalDays:   0, // broken
		NextDueAt:      &now,
	}
	id, err := userCardRepo.CreateUserCard(card)
	if err != nil {
		t.Fatalf("create user card: %v", err)
	}

	svc := NewSRSService(userCardRepo, config.DefaultLearningConfig(), logger)
	attempt := models.AttemptData{Correct: true, EarlyReveal: false, AnswerTimeMS: 3000}

	// First correct: Reps 0 -> 1, newInterval = 1 (handleReview Reps==0 branch), next_due = now+1d
	uc, _ := userCardRepo.GetUserCard(id)
	if err := svc.GradeCard(uc, attempt); err != nil {
		t.Fatalf("GradeCard 1: %v", err)
	}
	after1, _ := userCardRepo.GetUserCard(id)
	if after1.IntervalDays != 1 {
		t.Errorf("after first correct (reps was 0): expected interval_days 1, got %d", after1.IntervalDays)
	}

	// Second correct: Reps 1 -> 2, newInterval = 6 (handleReview Reps==1 branch), next_due = now+6d
	uc2, _ := userCardRepo.GetUserCard(id)
	if err := svc.GradeCard(uc2, attempt); err != nil {
		t.Fatalf("GradeCard 2: %v", err)
	}
	after2, _ := userCardRepo.GetUserCard(id)
	if after2.IntervalDays != 6 {
		t.Errorf("after second correct (reps was 1): expected interval_days 6, got %d", after2.IntervalDays)
	}
	if after2.NextDueAt == nil {
		t.Fatal("NextDueAt must be set")
	}
	hoursAhead := time.Until(*after2.NextDueAt).Hours()
	if hoursAhead < 5*24 { // at least ~5 days (6 days - small tolerance)
		t.Errorf("next_due_at should be ~6 days ahead, got %.1f hours", hoursAhead)
	}
}

// TestSRSService_GradeCard_SameAlgorithmForSpanishLearningConfig — English regression: SRS поля после GradeCard
// совпадают для ru-en и ru-es LearningConfig (пара влияет только на логи; время может отличаться на миллисекунды).
func TestSRSService_GradeCard_SameAlgorithmForSpanishLearningConfig(t *testing.T) {
	logger := zap.NewNop()
	db, userRepo, userCardRepo := setupSRSServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(555)

	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2), ($3, $4)", "pairA", "a", "pairB", "b")
	_, err := db.Exec(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES
		(1, 'pairA', 0, 'а', 'a'),
		(2, 'pairB', 0, 'б', 'b')`)
	if err != nil {
		t.Fatalf("training cards: %v", err)
	}

	now := time.Now()
	base := models.UserCard{
		UserID:       user.ID,
		Direction:    models.DirectionENtoRU,
		State:        models.StateNew,
		EF:           models.InitialEF,
		Reps:         0,
		IntervalDays: 0,
		LearningStep: 0,
		LapseCount:   0,
		NextDueAt:    &now,
	}
	base.TrainingCardID = 1
	id1, err := userCardRepo.CreateUserCard(&base)
	if err != nil {
		t.Fatal(err)
	}
	base.TrainingCardID = 2
	id2, err := userCardRepo.CreateUserCard(&base)
	if err != nil {
		t.Fatal(err)
	}

	lcES := config.LearningConfig{
		Pair: "ru-es", NativeLang: "ru", TargetLang: "es",
		AppCode: "spanish", GrammarBundleID: "es",
	}
	if err := config.ValidateLearningConfig(lcES); err != nil {
		t.Fatal(err)
	}

	svcEN := NewSRSService(userCardRepo, config.DefaultLearningConfig(), logger)
	svcES := NewSRSService(userCardRepo, lcES, logger)
	attempt := models.AttemptData{Correct: true, EarlyReveal: false, AnswerTimeMS: 3000}

	uc1, _ := userCardRepo.GetUserCard(id1)
	uc2, _ := userCardRepo.GetUserCard(id2)
	if err := svcEN.GradeCard(uc1, attempt); err != nil {
		t.Fatalf("GradeCard EN: %v", err)
	}
	if err := svcES.GradeCard(uc2, attempt); err != nil {
		t.Fatalf("GradeCard ES learning config: %v", err)
	}

	after1, _ := userCardRepo.GetUserCard(id1)
	after2, _ := userCardRepo.GetUserCard(id2)

	if after1.State != after2.State || after1.EF != after2.EF || after1.Reps != after2.Reps ||
		after1.IntervalDays != after2.IntervalDays || after1.LearningStep != after2.LearningStep ||
		after1.LapseCount != after2.LapseCount {
		t.Fatalf("SRS core fields differ: EN=%+v ES=%+v", after1, after2)
	}
	if (after1.LastQuality == nil) != (after2.LastQuality == nil) {
		t.Fatal("LastQuality presence mismatch")
	}
	if after1.LastQuality != nil && *after1.LastQuality != *after2.LastQuality {
		t.Fatalf("LastQuality: %v vs %v", *after1.LastQuality, *after2.LastQuality)
	}
	if after1.NextDueAt == nil || after2.NextDueAt == nil {
		t.Fatal("NextDueAt must be set")
	}
	skew := after1.NextDueAt.Sub(*after2.NextDueAt)
	if skew < 0 {
		skew = -skew
	}
	if skew > 2*time.Second {
		t.Fatalf("NextDueAt skew too large: %v vs %v (skew %v)", after1.NextDueAt, after2.NextDueAt, skew)
	}
}
