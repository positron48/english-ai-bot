package models

import "time"

// User represents a bot user
type User struct {
	ID                    int64
	TelegramID            int64
	TelegramUsername      string
	Timezone              string
	PreferredTrainingTime string
	SettingsJSON          string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// UserSettings represents user preferences for training
type UserSettings struct {
	MaxCardsPerSession      int    `json:"max_cards_per_session"`
	MaxNewPerSession        int    `json:"max_new_per_session"`
	EnableENtoRU            bool   `json:"enable_en_to_ru"`
	DirectionRatio          int    `json:"direction_ratio"`                      // 50 means 50/50, 60 means 60/40 RU->EN
	NotificationFrequency   string `json:"notification_frequency"`               // "never", "daily", or number of days as string (e.g., "3")
	LastNotificationDate    string `json:"last_notification_date,omitempty"`     // ISO date format: "2006-01-02"
	Language                string `json:"language,omitempty"`                   // "en" or "ru"
	OptionsDelaySeconds     *int   `json:"options_delay_seconds,omitempty"`      // 0-10, delay before showing answer options; nil = default 5
	WrongAnswerDelaySeconds *int   `json:"wrong_answer_delay_seconds,omitempty"` // 0-10, delay before next card after wrong answer; nil = default 5
	SpellModeEnabled        *bool  `json:"spell_mode_enabled,omitempty"`         // enable "compose word from letters" in training; nil = true
	SpellMasteringThreshold *int   `json:"spell_mastering_threshold,omitempty"`  // min mastering_score (0-100) for spell; nil = 50
	TypeModeEnabled         *bool  `json:"type_mode_enabled,omitempty"`          // enable "type the word" (no letters) in training; nil = true
	TypeMasteringThreshold  *int   `json:"type_mastering_threshold,omitempty"`   // min mastering_score (0-100) for type challenge; nil = 70
}
