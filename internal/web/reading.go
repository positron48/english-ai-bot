package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	"tgbot-skeleton/internal/grammarbundle"
	"tgbot-skeleton/internal/readingbundle"
	"tgbot-skeleton/internal/readingsync"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

type readingIndex struct {
	Version     string                     `json:"version"`
	GeneratedAt string                     `json:"generated_at"`
	Categories  map[string]readingCategory `json:"categories"`
	Texts       map[string]string          `json:"texts"`
}

type readingCategory struct {
	ID                string            `json:"id"`
	Title             string            `json:"title"`
	TitleTranslations map[string]string `json:"title_translations,omitempty"`
	Level             string            `json:"level"`
	Order             int               `json:"order"`
	TextIDs           []string          `json:"text_ids"`
}

type readingTextDoc struct {
	ID                string                 `json:"id"`
	CategoryID        string                 `json:"category_id"`
	Title             string                 `json:"title"`
	TitleTranslations map[string]string      `json:"title_translations,omitempty"`
	Level             string                 `json:"level"`
	TargetLanguage    string                 `json:"target_language"`
	ReadingPassage    map[string]interface{} `json:"reading_passage"`
}

// SyncReadingCatalogFromBundle loads reading content from the grammar bundle into the database.
func (r *Router) SyncReadingCatalogFromBundle(ctx context.Context) error {
	if r == nil || r.readingCatalogRepo == nil {
		return nil
	}
	return readingsync.SyncFromBundle(ctx, r.config, r.readingCatalogRepo, r.logger)
}

func docFromRepo(d *repository.ReadingTextDocument) *readingTextDoc {
	if d == nil {
		return nil
	}
	return &readingTextDoc{
		ID:                d.ID,
		CategoryID:        d.CategoryID,
		Title:             d.Title,
		TitleTranslations: d.TitleTranslations,
		Level:             d.Level,
		TargetLanguage:    d.TargetLanguage,
		ReadingPassage:    d.ReadingPassage,
	}
}

func snapshotToReadingIndex(snap *repository.ReadingCatalogSnapshot) *readingIndex {
	idx := &readingIndex{
		Version:     snap.Version,
		GeneratedAt: snap.GeneratedAt,
		Categories:  make(map[string]readingCategory, len(snap.Categories)),
		Texts:       make(map[string]string),
	}
	for id, c := range snap.Categories {
		idx.Categories[id] = readingCategory{
			ID:                c.CategoryID,
			Title:             c.Title,
			TitleTranslations: c.TitleTranslations,
			Level:             c.Level,
			Order:             c.Order,
			TextIDs:           append([]string(nil), c.TextIDs...),
		}
		for _, tid := range c.TextIDs {
			idx.Texts[tid] = ""
		}
	}
	return idx
}

func (r *Router) handleLearningReadingCategories(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if getUserIDFromContext(req.Context()) == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idx, err := r.readReadingIndex()
	if err != nil {
		r.logger.Error("failed to read reading index", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	type categoryResponse struct {
		CategoryID        string            `json:"category_id"`
		Title             string            `json:"title"`
		TitleTranslations map[string]string `json:"title_translations,omitempty"`
		Level             string            `json:"level"`
		Order             int               `json:"order"`
		TextCount         int               `json:"text_count"`
	}
	out := make([]categoryResponse, 0, len(idx.Categories))
	for id, cat := range idx.Categories {
		title := strings.TrimSpace(cat.Title)
		if title == "" {
			title = id
		}
		out = append(out, categoryResponse{
			CategoryID:        id,
			Title:             title,
			TitleTranslations: cat.TitleTranslations,
			Level:             cat.Level,
			Order:             cat.Order,
			TextCount:         len(cat.TextIDs),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Order == out[j].Order {
			return out[i].CategoryID < out[j].CategoryID
		}
		return out[i].Order < out[j].Order
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"categories": out})
}

func (r *Router) handleLearningReadingCategoryTexts(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	path := strings.TrimPrefix(req.URL.Path, "/api/learning/reading/categories/")
	path = strings.Trim(path, "/")
	if !strings.HasSuffix(path, "/texts") {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	categoryID := strings.Trim(strings.TrimSuffix(path, "/texts"), "/")
	if categoryID == "" {
		http.Error(w, "category_id required", http.StatusBadRequest)
		return
	}

	idx, err := r.readReadingIndex()
	if err != nil {
		r.logger.Error("failed to read reading index", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	cat, ok := idx.Categories[categoryID]
	if !ok {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	progressRepo := repository.NewReadingTextProgressRepository(r.db)
	type textResponse struct {
		TextID            string            `json:"text_id"`
		CategoryID        string            `json:"category_id"`
		Title             string            `json:"title"`
		TitleTranslations map[string]string `json:"title_translations,omitempty"`
		Level             string            `json:"level"`
		TargetLanguage    string            `json:"target_language"`
		IsRead            bool              `json:"is_read"`
	}
	out := make([]textResponse, 0, len(cat.TextIDs))
	for _, textID := range cat.TextIDs {
		doc, err := r.readReadingText(idx, textID)
		if err != nil {
			r.logger.Warn("failed to read reading text", zap.String("text_id", textID), zap.Error(err))
			continue
		}
		progress, err := progressRepo.Get(userID, textID)
		if err != nil {
			r.logger.Warn("failed to get reading progress", zap.Int64("user_id", userID), zap.String("text_id", textID), zap.Error(err))
		}
		out = append(out, textResponse{
			TextID:            doc.ID,
			CategoryID:        doc.CategoryID,
			Title:             doc.Title,
			TitleTranslations: doc.TitleTranslations,
			Level:             doc.Level,
			TargetLanguage:    doc.TargetLanguage,
			IsRead:            progress != nil,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"category": map[string]interface{}{
			"category_id":        categoryID,
			"title":              cat.Title,
			"title_translations": cat.TitleTranslations,
			"level":              cat.Level,
		},
		"texts": out,
	})
}

func (r *Router) handleLearningReadingTexts(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		r.handleLearningReadingTextGet(w, req)
		return
	case http.MethodPost:
		r.handleLearningReadingTextMarkRead(w, req)
		return
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (r *Router) handleLearningReadingTextGet(w http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	textID, ok := readingTextIDFromPath(req.URL.Path, false)
	if !ok {
		http.Error(w, "text_id required", http.StatusBadRequest)
		return
	}

	idx, err := r.readReadingIndex()
	if err != nil {
		r.logger.Error("failed to read reading index", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	doc, err := r.readReadingText(idx, textID)
	if err != nil {
		http.Error(w, "Text not found", http.StatusNotFound)
		return
	}

	block := map[string]interface{}{
		"id":    "reading_passage_" + textID,
		"type":  "reading_passage",
		"title": doc.Title,
		"reading_passage": func() map[string]interface{} {
			if doc.ReadingPassage == nil {
				return map[string]interface{}{}
			}
			return doc.ReadingPassage
		}(),
	}

	progressRepo := repository.NewReadingTextProgressRepository(r.db)
	progress, err := progressRepo.Get(userID, textID)
	if err != nil {
		r.logger.Warn("failed to get reading progress", zap.Int64("user_id", userID), zap.String("text_id", textID), zap.Error(err))
	}
	readingProgress := map[string]interface{}{"is_read": progress != nil}
	if progress != nil {
		readingProgress["read_at"] = progress.ReadAt.Format("2006-01-02T15:04:05Z")
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"text_id":            doc.ID,
		"category_id":        doc.CategoryID,
		"title":              doc.Title,
		"title_translations": doc.TitleTranslations,
		"level":              doc.Level,
		"target_language":    doc.TargetLanguage,
		"block":              block,
		"reading_progress":   readingProgress,
	})
}

func (r *Router) handleLearningReadingTextMarkRead(w http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	textID, ok := readingTextIDFromPath(req.URL.Path, true)
	if !ok {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	progressRepo := repository.NewReadingTextProgressRepository(r.db)
	if err := progressRepo.MarkRead(userID, textID); err != nil {
		r.logger.Error("failed to mark reading text as read", zap.Int64("user_id", userID), zap.String("text_id", textID), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if r.linglowEventRepo != nil {
		if _, err := r.linglowEventRepo.RecordReadingCompleted(req.Context(), repository.ReadingCompletedInput{
			UserID:     userID,
			CourseCode: r.currentCourseCodeForUser(req.Context(), userID),
			ChapterID:  textID,
		}); err != nil {
			r.logger.Warn("failed to record linglow reading event", zap.Int64("user_id", userID), zap.String("text_id", textID), zap.Error(err))
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"text_id": textID,
	})
}

func readingTextIDFromPath(path string, expectMarkRead bool) (string, bool) {
	trimmed := strings.Trim(strings.TrimPrefix(path, "/api/learning/reading/texts/"), "/")
	if trimmed == "" {
		return "", false
	}
	if expectMarkRead {
		if !strings.HasSuffix(trimmed, "/mark-read") {
			return "", false
		}
		textID := strings.Trim(strings.TrimSuffix(trimmed, "/mark-read"), "/")
		return textID, textID != ""
	}
	if strings.Contains(trimmed, "/") {
		return "", false
	}
	return trimmed, true
}

func (r *Router) handleReadingWordLookup(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	lemma := strings.TrimSpace(req.URL.Query().Get("lemma"))
	if lemma == "" {
		http.Error(w, "lemma required", http.StatusBadRequest)
		return
	}

	canonicalLemma, wordCardID, found, err := r.findReadingWordCardByInput(lemma)
	if err != nil {
		r.logger.Error("reading word lookup: failed to resolve word card", zap.String("lemma", lemma), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !found {
		if ws := r.getReadingWordService(); ws != nil && ws.IsSingleWord(lemma) {
			// Reuse the same path as chat single-word lookup: DB + word_forms + LLM fallback.
			if _, err := ws.GetWordDefinition(req.Context(), userID, lemma); err != nil {
				r.logger.Warn("reading word lookup: word service fallback failed", zap.String("lemma", lemma), zap.Error(err))
			}
			canonicalLemma, wordCardID, found, err = r.findReadingWordCardByInput(lemma)
			if err != nil {
				r.logger.Error("reading word lookup: failed to resolve word card after fallback", zap.String("lemma", lemma), zap.Error(err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}
		if !found {
			http.Error(w, "Word not found", http.StatusNotFound)
			return
		}
	}

	userWordKnowledgeRepo := repository.NewUserWordKnowledgeRepository(r.db, r.logger)
	isKnown, err := userWordKnowledgeRepo.IsKnown(userID, wordCardID)
	if err != nil {
		r.logger.Warn("reading word lookup: failed to check known status", zap.Int64("user_id", userID), zap.Int64("word_card_id", wordCardID), zap.Error(err))
	}
	if !isKnown {
		wordSetService := r.getWordSetService()
		if err := wordSetService.EnsureTrainingCardsExist(req.Context(), wordCardID); err != nil {
			r.logger.Warn("reading word lookup: failed to ensure training cards", zap.Int64("word_card_id", wordCardID), zap.Error(err))
		}
		if err := wordSetService.EnsureUserCardsForWord(userID, wordCardID); err != nil {
			r.logger.Error("reading word lookup: failed to create user cards", zap.Int64("user_id", userID), zap.Int64("word_card_id", wordCardID), zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	r.handleVocabWordCards(w, req, userID, canonicalLemma)
}

type readingWordLookupService interface {
	IsSingleWord(text string) bool
	GetWordDefinition(ctx context.Context, userID int64, word string) (string, error)
}

func (r *Router) getReadingWordService() readingWordLookupService {
	if r == nil || r.wordService == nil {
		return nil
	}
	ws, ok := r.wordService.(readingWordLookupService)
	if !ok {
		return nil
	}
	return ws
}

func (r *Router) findReadingWordCardByInput(input string) (string, int64, bool, error) {
	normalized := strings.TrimSpace(strings.ToLower(input))
	if normalized == "" {
		return "", 0, false, nil
	}
	var canonicalLemma string
	var wordCardID int64
	err := r.db.QueryRow(`
SELECT wc.word, wc.id
FROM word_forms wf
JOIN word_cards wc ON wc.id = wf.word_card_id
WHERE LOWER(wf.form) = LOWER(?)
LIMIT 1`, normalized).Scan(&canonicalLemma, &wordCardID)
	if err == nil {
		return canonicalLemma, wordCardID, true, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", 0, false, err
	}
	err = r.db.QueryRow(`SELECT word, id FROM word_cards WHERE LOWER(word) = LOWER(?)`, normalized).Scan(&canonicalLemma, &wordCardID)
	if err == nil {
		return canonicalLemma, wordCardID, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, false, nil
	}
	return "", 0, false, err
}

// BootstrapReadingWordCards pre-creates reading vocabulary word cards after deploy/startup.
// It scans all reading tokens and ensures a word card exists (DB lookup -> word_forms -> LLM if missing).
func (r *Router) BootstrapReadingWordCards(ctx context.Context) {
	if r == nil || r.db == nil {
		return
	}
	if r.courseRepo != nil && r.courseRepo.HasMultipleActiveCourses(ctx) {
		r.logger.Info("reading bootstrap skipped for multi-course database")
		return
	}
	idx, err := r.readReadingIndex()
	if err != nil {
		r.logger.Warn("reading bootstrap: failed to read index", zap.Error(err))
		return
	}
	if len(idx.Texts) == 0 {
		return
	}

	wordSetService := r.getWordSetService()
	ws := r.getReadingWordService()
	seen := make(map[string]struct{}, 1024)
	total := 0
	created := 0

	for textID := range idx.Texts {
		select {
		case <-ctx.Done():
			r.logger.Info("reading bootstrap stopped by context", zap.Int("processed_tokens", total), zap.Int("created_cards", created))
			return
		default:
		}
		doc, err := r.readReadingText(idx, textID)
		if err != nil {
			r.logger.Warn("reading bootstrap: failed to read text", zap.String("text_id", textID), zap.Error(err))
			continue
		}
		for _, token := range readingBootstrapTokens(doc) {
			n := strings.ToLower(strings.TrimSpace(token))
			if n == "" {
				continue
			}
			if _, ok := seen[n]; ok {
				continue
			}
			seen[n] = struct{}{}
			total++

			if _, _, found, err := r.findReadingWordCardByInput(n); err == nil && found {
				continue
			}
			if ws != nil && !ws.IsSingleWord(n) {
				continue
			}
			if _, err := wordSetService.EnsureWordCardExists(ctx, n); err != nil {
				r.logger.Warn("reading bootstrap: failed to ensure word card", zap.String("token", n), zap.Error(err))
				continue
			}
			created++
		}
	}

	r.logger.Info("reading bootstrap completed",
		zap.Int("unique_tokens", total),
		zap.Int("created_cards", created),
	)
}

func readingBootstrapTokens(doc *readingTextDoc) []string {
	out := make([]string, 0, 64)
	if doc == nil || doc.ReadingPassage == nil {
		return out
	}
	segs, ok := doc.ReadingPassage["segments"].([]interface{})
	if !ok {
		return out
	}
	for _, segRaw := range segs {
		seg, ok := segRaw.(map[string]interface{})
		if !ok {
			continue
		}
		tokens, ok := seg["tokens"].([]interface{})
		if !ok {
			continue
		}
		for _, tokenRaw := range tokens {
			tok, ok := tokenRaw.(map[string]interface{})
			if !ok {
				continue
			}
			clickable, _ := tok["clickable"].(bool)
			if !clickable {
				continue
			}
			if lemma, _ := tok["lemma"].(string); strings.TrimSpace(lemma) != "" {
				out = append(out, lemma)
			}
			if surface, _ := tok["surface"].(string); strings.TrimSpace(surface) != "" {
				out = append(out, surface)
			}
		}
	}
	return out
}

func (r *Router) handleLearningReadingAudio(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	audioPath := strings.TrimSpace(req.URL.Query().Get("path"))
	if audioPath == "" {
		http.Error(w, "path required", http.StatusBadRequest)
		return
	}
	if strings.Contains(audioPath, "..") || strings.HasPrefix(audioPath, "/") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	audioPath = filepath.Clean(audioPath)
	var bundleFS fs.FS
	var err error
	if courseCode := strings.TrimSpace(req.URL.Query().Get("course_code")); courseCode != "" {
		bundleID := grammarBundleForCourse(courseCode)
		bundleFS, err = grammarbundle.BundleFS(bundleID)
	}
	if bundleFS == nil {
		bundleFS, err = readingbundle.BundleFS(r.config)
	}
	if err != nil {
		r.logger.Error("failed to select grammar bundle filesystem", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	data, err := fs.ReadFile(bundleFS, audioPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		r.logger.Error("failed to read grammar audio asset", zap.String("path", audioPath), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(data)
}

func (r *Router) readReadingText(idx *readingIndex, textID string) (*readingTextDoc, error) {
	if r.readingCatalogRepo != nil {
		doc, ok, err := r.readingCatalogRepo.GetTextDocument(textID)
		if err != nil {
			return nil, err
		}
		if ok {
			return docFromRepo(doc), nil
		}
	}
	if idx == nil {
		return nil, fmt.Errorf("text not found: %s", textID)
	}
	relPath, ok := idx.Texts[textID]
	if !ok {
		return nil, fmt.Errorf("text not found: %s", textID)
	}
	if strings.TrimSpace(relPath) == "" {
		return nil, fmt.Errorf("text not found: %s", textID)
	}
	if strings.Contains(relPath, "..") || strings.HasPrefix(relPath, "/") {
		return nil, fmt.Errorf("invalid reading text path")
	}
	bundleFS, err := readingbundle.BundleFS(r.config)
	if err != nil {
		return nil, err
	}
	data, err := fs.ReadFile(bundleFS, filepath.Join("reading", filepath.Clean(relPath)))
	if err != nil {
		return nil, err
	}
	var doc readingTextDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if strings.TrimSpace(doc.ID) == "" {
		doc.ID = textID
	}
	return &doc, nil
}

func (r *Router) readReadingIndex() (*readingIndex, error) {
	if r.readingCatalogRepo != nil {
		n, err := r.readingCatalogRepo.CountCategories()
		if err != nil {
			r.logger.Warn("reading index: db category count failed, falling back to bundle", zap.Error(err))
		} else if n > 0 {
			snap, err := r.readingCatalogRepo.LoadSnapshot()
			if err != nil {
				r.logger.Warn("reading index: db snapshot failed, falling back to bundle", zap.Error(err))
			} else {
				return snapshotToReadingIndex(snap), nil
			}
		}
	}
	return r.readReadingIndexFromBundleFS()
}

func (r *Router) readReadingIndexFromBundleFS() (*readingIndex, error) {
	bundleFS, err := readingbundle.BundleFS(r.config)
	if err != nil {
		return nil, err
	}
	data, err := fs.ReadFile(bundleFS, "reading/index.json")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &readingIndex{
				Version:     "1.0.0",
				GeneratedAt: "",
				Categories:  map[string]readingCategory{},
				Texts:       map[string]string{},
			}, nil
		}
		return nil, fmt.Errorf("read reading/index.json: %w", err)
	}
	var idx readingIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("parse reading/index.json: %w", err)
	}
	if idx.Categories == nil {
		idx.Categories = map[string]readingCategory{}
	}
	if idx.Texts == nil {
		idx.Texts = map[string]string{}
	}
	return &idx, nil
}
