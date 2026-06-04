package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

type offlineChapterManifest struct {
	ChapterID         string            `json:"chapter_id"`
	Title             string            `json:"title"`
	TitleTranslations map[string]string `json:"title_translations,omitempty"`
	TitleShort        string            `json:"title_short,omitempty"`
	Description       string            `json:"description,omitempty"`
	Level             string            `json:"level,omitempty"`
	Order             int               `json:"order"`
	EstimatedMinutes  int               `json:"estimated_minutes,omitempty"`
	BestScore         int               `json:"best_score"`
	Passed            bool              `json:"passed"`
	CanAccess         bool              `json:"can_access"`
	DownloadURL       string            `json:"download_url"`
	ApproxBytes       int               `json:"approx_bytes"`
}

type offlineSectionManifest struct {
	SectionID          string                   `json:"section_id"`
	Title              string                   `json:"title"`
	TitleTranslations  map[string]string        `json:"title_translations,omitempty"`
	Level              string                   `json:"level"`
	Order              int                      `json:"order"`
	PublishedChapters  int                      `json:"published_chapters"`
	PassedChapters     int                      `json:"passed_chapters"`
	TotalChapters      int                      `json:"total_chapters"`
	ProgressPercentage int                      `json:"progress_percentage"`
	CanAccess          bool                     `json:"can_access"`
	CategoryTestScore  *int                     `json:"category_test_score,omitempty"`
	Chapters           []offlineChapterManifest `json:"chapters"`
}

func (r *Router) handleLearningGrammarOfflineManifest(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sectionsData, err := r.grammarService.ContentRepo.GetSections()
	if err != nil {
		r.logger.Error("failed to get offline grammar sections", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	versionHash, err := r.grammarService.ContentRepo.BundleVersionHash()
	if err != nil {
		r.logger.Error("failed to hash offline grammar bundle", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	index, err := r.grammarService.ContentRepo.GetIndex()
	if err != nil {
		r.logger.Error("failed to get offline grammar index", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	publishedSections, err := r.grammarService.PublishRepo.GetPublishedItemsByType("section")
	if err != nil {
		r.logger.Error("failed to get published grammar sections", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	publishedChapters, err := r.grammarService.PublishRepo.GetPublishedItemsByType("chapter")
	if err != nil {
		r.logger.Error("failed to get published grammar chapters", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	sections := make([]offlineSectionManifest, 0)
	totalBytes := 0
	totalChapters := 0
	for i := range sectionsData.Sections {
		section := &sectionsData.Sections[i]
		sectionItem, ok := publishedSections[section.SectionID]
		if !ok || !sectionItem.IsPublished {
			continue
		}

		title := section.Title
		if sectionItem.Name != nil && *sectionItem.Name != "" {
			title = *sectionItem.Name
		}

		canAccess, errAccess := r.grammarService.CanAccessSection(req.Context(), userID, section.SectionID)
		if errAccess != nil {
			r.logger.Warn("failed to check offline section access", zap.String("section_id", section.SectionID), zap.Error(errAccess))
		}
		bestCategoryScore, errScore := r.grammarService.AttemptRepo.GetCategoryTestBestScore(userID, section.SectionID)
		var categoryTestScore *int
		if errScore == nil && bestCategoryScore > 0 {
			categoryTestScore = &bestCategoryScore
		}

		chapters := make([]offlineChapterManifest, 0)
		passedChapters := 0
		scoreSum := 0
		for _, chapterID := range section.ChapterIDs {
			chapterItem, ok := publishedChapters[chapterID]
			if !ok || !chapterItem.IsPublished {
				continue
			}
			chapter, err := r.grammarService.ContentRepo.GetChapter(chapterID)
			if err != nil {
				r.logger.Warn("failed to load offline grammar chapter", zap.String("chapter_id", chapterID), zap.Error(err))
				continue
			}
			raw, err := r.grammarService.ContentRepo.GetChapterRawJSON(chapterID)
			if err != nil {
				r.logger.Warn("failed to size offline grammar chapter", zap.String("chapter_id", chapterID), zap.Error(err))
			}
			chapterTitle := chapter.Title
			if chapterItem.Name != nil && *chapterItem.Name != "" {
				chapterTitle = *chapterItem.Name
			}
			progress, _ := r.grammarService.AttemptRepo.GetChapterProgress(userID, chapterID)
			if progress.Passed {
				passedChapters++
			}
			scoreSum += progress.BestScore
			chapterAccess, _ := r.grammarService.CanAccessChapter(req.Context(), userID, chapterID)
			approxBytes := len(raw)
			totalBytes += approxBytes
			totalChapters++
			chapters = append(chapters, offlineChapterManifest{
				ChapterID:         chapter.ID,
				Title:             chapterTitle,
				TitleTranslations: chapter.TitleTranslations,
				TitleShort:        chapter.TitleShort,
				Description:       chapter.Description,
				Level:             chapter.Level,
				Order:             chapter.Order,
				EstimatedMinutes:  chapter.EstimatedMinutes,
				BestScore:         progress.BestScore,
				Passed:            progress.Passed,
				CanAccess:         chapterAccess,
				DownloadURL:       "/api/learning/grammar/offline/chapters/" + chapter.ID,
				ApproxBytes:       approxBytes,
			})
		}

		progressPct := 0
		if len(chapters) > 0 {
			progressPct = scoreSum / len(chapters)
		}
		sections = append(sections, offlineSectionManifest{
			SectionID:          section.SectionID,
			Title:              title,
			TitleTranslations:  section.TitleTranslations,
			Level:              section.Level,
			Order:              section.Order,
			PublishedChapters:  len(chapters),
			PassedChapters:     passedChapters,
			TotalChapters:      len(section.ChapterIDs),
			ProgressPercentage: progressPct,
			CanAccess:          canAccess,
			CategoryTestScore:  categoryTestScore,
			Chapters:           chapters,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"app_code":        r.config.Learning.AppCode,
		"bundle_id":       r.config.Learning.GrammarBundleID,
		"native_lang":     r.config.Learning.NativeLang,
		"target_lang":     r.config.Learning.TargetLang,
		"course_version":  index.Version,
		"generated_at":    index.GeneratedAt,
		"version_hash":    versionHash,
		"approx_bytes":    totalBytes,
		"total_chapters":  totalChapters,
		"downloaded_from": r.config.WebApp.PublicURL,
		"sections":        sections,
		"training_pack": map[string]interface{}{
			"download_url": "/api/learning/grammar/offline/training-pack",
		},
	})
}

func (r *Router) handleLearningGrammarOfflineTrainingPack(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	questions, err := r.grammarService.GetOfflineGrammarTrainingQuestions(req.Context(), userID)
	if err != nil {
		r.logger.Error("failed to get offline grammar training pack", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"bundle_id":  r.config.Learning.GrammarBundleID,
		"language":   r.config.Learning.TargetLang,
		"questions":  questions,
		"total":      len(questions),
		"downloaded": false,
	})
}

func (r *Router) handleLearningGrammarOfflineChapter(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	chapterID := strings.TrimPrefix(req.URL.Path, "/api/learning/grammar/offline/chapters/")
	chapterID = strings.Trim(chapterID, "/")
	if chapterID == "" {
		http.Error(w, "chapter_id required", http.StatusBadRequest)
		return
	}
	if ok, err := r.grammarService.PublishRepo.IsPublished("chapter", chapterID); err != nil || !ok {
		http.Error(w, "Chapter not found", http.StatusNotFound)
		return
	}
	chapter, err := r.grammarService.ContentRepo.GetChapter(chapterID)
	if err != nil {
		r.logger.Error("failed to get offline grammar chapter", zap.String("chapter_id", chapterID), zap.Error(err))
		http.Error(w, "Chapter not found", http.StatusNotFound)
		return
	}
	if ok, err := r.grammarService.PublishRepo.IsPublished("section", chapter.SectionID); err != nil || !ok {
		http.Error(w, "Chapter not found", http.StatusNotFound)
		return
	}
	item, _ := r.grammarService.PublishRepo.GetPublishedItem("chapter", chapterID)
	title := chapter.Title
	if item.Name != nil && *item.Name != "" {
		title = *item.Name
	}
	resp := map[string]interface{}{
		"chapter":            chapter,
		"title":              title,
		"title_translations": chapter.TitleTranslations,
	}
	if sec, err := r.grammarService.GetSectionBySectionID(req.Context(), chapter.SectionID); err == nil && sec != nil {
		resp["section"] = map[string]interface{}{
			"section_id":         sec.SectionID,
			"title":              sec.Title,
			"title_translations": sec.TitleTranslations,
			"level":              sec.Level,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

type offlineSyncAttempt struct {
	ClientAttemptID string               `json:"client_attempt_id"`
	Scope           string               `json:"scope"`
	ScopeID         string               `json:"scope_id"`
	Answers         []service.AnswerItem `json:"answers"`
	CourseVersion   string               `json:"course_version,omitempty"`
}

func (r *Router) handleLearningGrammarOfflineSyncAttempts(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Attempts []offlineSyncAttempt `json:"attempts"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	results := make([]map[string]interface{}, 0, len(body.Attempts))
	for _, attempt := range body.Attempts {
		clientID := strings.TrimSpace(attempt.ClientAttemptID)
		if clientID == "" || attempt.Scope == "" || attempt.ScopeID == "" {
			results = append(results, map[string]interface{}{
				"client_attempt_id": clientID,
				"synced":            false,
				"error":             "client_attempt_id, scope and scope_id are required",
			})
			continue
		}
		exists, err := r.grammarService.AttemptRepo.HasClientAttempt(userID, clientID)
		if err != nil {
			results = append(results, map[string]interface{}{
				"client_attempt_id": clientID,
				"synced":            false,
				"error":             err.Error(),
			})
			continue
		}
		if exists {
			results = append(results, map[string]interface{}{
				"client_attempt_id": clientID,
				"synced":            true,
				"duplicate":         true,
			})
			continue
		}
		result, err := r.grammarService.SubmitTestWithClientAttemptID(req.Context(), userID, attempt.Scope, attempt.ScopeID, attempt.Answers, clientID)
		if err != nil {
			r.logger.Warn("failed to sync offline grammar attempt",
				zap.String("client_attempt_id", clientID),
				zap.String("scope", attempt.Scope),
				zap.String("scope_id", attempt.ScopeID),
				zap.Error(err))
			results = append(results, map[string]interface{}{
				"client_attempt_id": clientID,
				"synced":            false,
				"error":             err.Error(),
			})
			continue
		}
		r.recordLinglowGrammarTestAttempt(req, userID, attempt.Scope, attempt.ScopeID, clientID, attempt.Answers, result)
		results = append(results, map[string]interface{}{
			"client_attempt_id": clientID,
			"synced":            true,
			"result":            result,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
	})
}

type offlineSyncTrainingAttempt struct {
	ClientAttemptID string      `json:"client_attempt_id"`
	QuestionID      string      `json:"question_id"`
	Answer          interface{} `json:"answer"`
}

func (r *Router) handleLearningGrammarOfflineSyncTrainingAttempts(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Attempts []offlineSyncTrainingAttempt `json:"attempts"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	results := make([]map[string]interface{}, 0, len(body.Attempts))
	for _, attempt := range body.Attempts {
		clientID := strings.TrimSpace(attempt.ClientAttemptID)
		if clientID == "" || strings.TrimSpace(attempt.QuestionID) == "" {
			results = append(results, map[string]interface{}{"client_attempt_id": clientID, "synced": false, "error": "client_attempt_id and question_id are required"})
			continue
		}
		if r.grammarService.SRSRepo != nil {
			exists, err := r.grammarService.SRSRepo.HasClientAttempt(userID, clientID)
			if err != nil {
				results = append(results, map[string]interface{}{"client_attempt_id": clientID, "synced": false, "error": err.Error()})
				continue
			}
			if exists {
				results = append(results, map[string]interface{}{"client_attempt_id": clientID, "synced": true, "duplicate": true})
				continue
			}
		}
		result, err := r.grammarService.SubmitGrammarSrsAnswerWithClientAttemptID(req.Context(), userID, attempt.QuestionID, attempt.Answer, clientID)
		if err != nil {
			results = append(results, map[string]interface{}{"client_attempt_id": clientID, "synced": false, "error": err.Error()})
			continue
		}
		results = append(results, map[string]interface{}{"client_attempt_id": clientID, "synced": true, "result": result})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": results})
}
