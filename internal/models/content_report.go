package models

import "time"

type ContentReportStatus string

const (
	ContentReportStatusActive   ContentReportStatus = "active"
	ContentReportStatusResolved ContentReportStatus = "resolved"
)

type ContentReport struct {
	ID                   int64
	UserID               int64
	SourceType           string
	Status               ContentReportStatus
	Word                 string
	TranslationDirection string
	WordCardID           *int64
	TrainingCardID       *int64
	UserCardID           *int64
	WordCategory         string
	GrammarChapterID     string
	TheoryBlockID        string
	GrammarQuestionID    string
	CommentText          string
	PayloadJSON          string
	ResolvedAt           *time.Time
	ResolvedByUserID     *int64
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
