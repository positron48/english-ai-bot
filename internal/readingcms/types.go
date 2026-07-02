package readingcms

import (
	"encoding/json"
	"time"
)

const (
	StatusDraft      = "draft"
	StatusApproved   = "approved"
	StatusAudioReady = "audio_ready"
	StatusPublished  = "published"
	StatusRejected   = "rejected"

	AudioNone    = "none"
	AudioPartial = "partial"
	AudioReady   = "ready"

	CoverNone   = "none"
	CoverPrompt = "prompt"
	CoverReady  = "ready"

	OriginLLM          = "llm"
	OriginManualText   = "manual_text"
	OriginInputJSON    = "input_json"
	OriginCourseImport = "course_import"
)

// DraftMeta is stored in drafts/index.json.
type DraftMeta struct {
	TextID            string     `json:"text_id"`
	CourseCode        string     `json:"course_code"`
	Title             string     `json:"title"`
	Level             string     `json:"level"`
	Format            string     `json:"format"`
	TargetLanguage    string     `json:"target_language"`
	Status            string     `json:"status"`
	Origin            string     `json:"origin"`
	AudioStatus       string     `json:"audio_status"`
	CoverStatus       string     `json:"cover_status"`
	CoverThumbRelPath string     `json:"cover_thumb_rel_path,omitempty"`
	CoverImagePrompt  string     `json:"cover_image_prompt,omitempty"`
	SegmentsTotal     int        `json:"segments_total"`
	SegmentsWithAudio int        `json:"segments_with_audio"`
	LastJobLog        string     `json:"last_job_log,omitempty"`
	PublishedAt       *time.Time `json:"published_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// TextDocument matches bundle reading text JSON on disk.
type TextDocument struct {
	ID                string                 `json:"id"`
	CategoryID        string                 `json:"category_id"`
	Title             string                 `json:"title"`
	TitleTranslations map[string]string      `json:"title_translations,omitempty"`
	Level             string                 `json:"level"`
	TargetLanguage    string                 `json:"target_language"`
	CoverThumbRelPath string                 `json:"cover_thumb_rel_path,omitempty"`
	CoverHeroRelPath  string                 `json:"cover_hero_rel_path,omitempty"`
	CoverImagePrompt  string                 `json:"cover_image_prompt,omitempty"`
	ReadingPassage    map[string]interface{} `json:"reading_passage"`
}

// PublishedItem is a text already present in course reading catalog.
type PublishedItem struct {
	TextID            string     `json:"text_id"`
	CourseCode        string     `json:"course_code"`
	Title             string     `json:"title"`
	Level             string     `json:"level"`
	TargetLanguage    string     `json:"target_language"`
	CategoryID        string     `json:"category_id"`
	SegmentsCount     int        `json:"segments_count"`
	SegmentsWithAudio int        `json:"segments_with_audio"`
	AudioStatus       string     `json:"audio_status"`
	AudioReady        bool       `json:"audio_ready"`
	CoverStatus       string     `json:"cover_status"`
	CoverThumbRelPath string     `json:"cover_thumb_rel_path,omitempty"`
	CoverHeroRelPath  string     `json:"cover_hero_rel_path,omitempty"`
	CoverImagePrompt  string     `json:"cover_image_prompt,omitempty"`
	CoverGeneratedAt  *time.Time `json:"cover_generated_at,omitempty"`
	InCMS             bool       `json:"in_cms"`
	GitStatus         string     `json:"git_status,omitempty"`
	IsNewUncommitted  bool       `json:"is_new_uncommitted,omitempty"`
}

// PublishedCoverRequest generates a cover for one published course text.
type PublishedCoverRequest struct {
	CourseCode string `json:"course_code"`
	TextID     string `json:"text_id"`
	Force      bool   `json:"force"`
	Prompt     string `json:"prompt,omitempty"`
	SkipLLM    bool   `json:"skip_llm"`
}

// DraftCoverRequest generates a cover for a draft.
type DraftCoverRequest struct {
	Force   bool   `json:"force"`
	Prompt  string `json:"prompt,omitempty"`
	SkipLLM bool   `json:"skip_llm"`
}

// GenerateRequest is batch LLM generation input.
type GenerateRequest struct {
	CourseCode string `json:"course_code"`
	Level      string `json:"level"`
	Count      int    `json:"count"`
	Format     string `json:"format"`
	Title      string `json:"title,omitempty"`
	WithAudio  bool   `json:"with_audio"`
}

// ImportTextRequest is plain-text import transformed by LLM (same pipeline as generate).
type ImportTextRequest struct {
	CourseCode  string `json:"course_code"`
	Level       string `json:"level"`
	Format      string `json:"format"`
	Title       string `json:"title"`
	Text        string `json:"text"`
	WithAudio   *bool  `json:"with_audio"`
	AutoPublish bool   `json:"auto_publish"`
	SyncBundle  bool   `json:"sync_bundle"`
}

// ImportJSONRequest imports LLM-ready JSON without calling the LLM (TTS + publish optional).
type ImportJSONRequest struct {
	CourseCode  string          `json:"course_code"`
	Level       string          `json:"level"`
	Format      string          `json:"format"`
	Title       string          `json:"title"`
	Document    json.RawMessage `json:"document"`
	WithAudio   *bool           `json:"with_audio"`
	AutoPublish bool            `json:"auto_publish"`
	SyncBundle  bool            `json:"sync_bundle"`
}

// ImportJSONBatchRequest imports several LLM-ready JSON documents from pasted text.
type ImportJSONBatchRequest struct {
	CourseCode    string `json:"course_code"`
	Level         string `json:"level"`
	Format        string `json:"format"`
	Title         string `json:"title"`
	DocumentsText string `json:"documents_text"`
	WithAudio     *bool  `json:"with_audio"`
	AutoPublish   bool   `json:"auto_publish"`
	SyncBundle    bool   `json:"sync_bundle"`
}

// ImportJSONBatchResult reports one item in a batch import.
type ImportJSONBatchResult struct {
	Index    int           `json:"index"`
	Draft    *DraftMeta    `json:"draft,omitempty"`
	Document *TextDocument `json:"document,omitempty"`
	Error    string        `json:"error,omitempty"`
}

// ImportJSONBatchResponse summarizes a batch import.
type ImportJSONBatchResponse struct {
	Total     int                     `json:"total"`
	Succeeded int                     `json:"succeeded"`
	Failed    int                     `json:"failed"`
	Results   []ImportJSONBatchResult `json:"results"`
}

// PromptRequest selects course-specific reading LLM prompt text.
type PromptRequest struct {
	CourseCode string `json:"course_code"`
	Level      string `json:"level"`
	Format     string `json:"format"`
	Title      string `json:"title"`
	Kind       string `json:"kind"` // generate | transform
	SourceText string `json:"source_text,omitempty"`
}

// CoverBatchRequest triggers cover generation for published texts in a course.
type CoverBatchRequest struct {
	CourseCode  string `json:"course_code"`
	Level       string `json:"level,omitempty"`
	Force       bool   `json:"force"`
	SkipPrompts bool   `json:"skip_prompts,omitempty"`
	Limit       int    `json:"limit,omitempty"`
}
