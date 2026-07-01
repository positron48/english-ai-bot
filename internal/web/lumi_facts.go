package web

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// handleLumiFact returns the fact of the day for the given context.
// @Summary      Факт дня от Lumi
// @Description  Возвращает факт дня для пары (курс, контекст, локаль) с ротацией «самый давно не показанный — следующий»
// @Tags         Linglow
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Success      204  "Фактов нет"
// @Router       /api/linglow/lumi-fact [get]
func (r *Router) handleLumiFact(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	courseCode := strings.TrimSpace(req.URL.Query().Get("course_code"))
	if courseCode == "" {
		courseCode = r.currentCourseCodeForUser(req.Context(), userID)
	}
	locale := strings.TrimSpace(req.URL.Query().Get("locale"))
	if locale == "" {
		locale = "ru"
	}
	factContext := req.URL.Query().Get("context")

	repo := repository.NewLumiFactRepository(r.db)
	fact, err := repo.GetDailyFact(req.Context(), courseCode, factContext, locale)
	if err != nil {
		r.logger.Error("failed to get lumi fact", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if fact == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writeJSON(w, map[string]interface{}{
		"id":      fact.ID,
		"body":    fact.Body,
		"context": fact.Context,
	})
}

// handleAdminLumiFacts serves GET (list), POST (bulk add) and PUT (update one).
// @Summary      Управление Lumi-фактами
// @Tags         Admin
// @Router       /api/admin/lumi-facts [get]
func (r *Router) handleAdminLumiFacts(w http.ResponseWriter, req *http.Request) {
	repo := repository.NewLumiFactRepository(r.db)
	switch req.Method {
	case http.MethodGet:
		q := req.URL.Query()
		limit, _ := strconv.Atoi(q.Get("limit"))
		offset, _ := strconv.Atoi(q.Get("offset"))
		facts, total, err := repo.List(req.Context(), repository.LumiFactFilter{
			CourseCode: strings.TrimSpace(q.Get("course_code")),
			Context:    strings.TrimSpace(q.Get("context")),
			Locale:     strings.TrimSpace(q.Get("locale")),
			Status:     strings.TrimSpace(q.Get("status")),
			Limit:      limit,
			Offset:     offset,
		})
		if err != nil {
			r.logger.Error("failed to list lumi facts", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]interface{}{"facts": facts, "total": total})
	case http.MethodPost:
		facts, err := parseLumiFactsImport(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if len(facts) == 0 {
			http.Error(w, "facts are empty", http.StatusBadRequest)
			return
		}
		inserted, err := repo.BulkInsertFacts(req.Context(), facts, getUserIDFromContext(req.Context()))
		if err != nil {
			r.logger.Error("failed to bulk insert lumi facts", zap.Error(err))
			if errors.Is(err, repository.ErrInvalidLumiFact) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]interface{}{"inserted": inserted})
	case http.MethodPut:
		var fact repository.LumiFact
		if err := json.NewDecoder(req.Body).Decode(&fact); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if err := repo.Update(req.Context(), fact); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, map[string]interface{}{"success": true})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAdminLumiFactsPromptTemplate returns a ready-to-use prompt for generating import JSON.
// @Summary      Admin: промпт для генерации Lumi-фактов
// @Tags         Admin
// @Router       /api/admin/lumi-facts/prompt-template [get]
func (r *Router) handleAdminLumiFactsPromptTemplate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	courses, _ := r.lumiFactPromptCourses(req.Context())
	writeJSON(w, map[string]interface{}{"prompt": lumiFactsGenerationPrompt(courses)})
}

func splitFactBlocks(text string) []string {
	parts := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n\n")
	out := []string{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseLumiFactsImport(body io.Reader) ([]repository.LumiFact, error) {
	raw, err := io.ReadAll(body)
	if err != nil {
		return nil, errors.New("failed to read request body")
	}
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil, errors.New("request body is empty")
	}
	if raw[0] == '[' {
		var facts []repository.LumiFact
		if err := json.Unmarshal(raw, &facts); err != nil {
			return nil, errors.New("Invalid JSON: " + err.Error())
		}
		return facts, nil
	}

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, errors.New("Invalid JSON: " + err.Error())
	}
	if factsRaw, ok := envelope["facts"]; ok {
		var facts []repository.LumiFact
		if err := json.Unmarshal(factsRaw, &facts); err != nil {
			return nil, errors.New("Invalid facts JSON: " + err.Error())
		}
		return facts, nil
	}

	var textBody struct {
		CourseCode string `json:"course_code"`
		Context    string `json:"context"`
		Locale     string `json:"locale"`
		Text       string `json:"text"`
	}
	if err := json.Unmarshal(raw, &textBody); err != nil {
		return nil, errors.New("Invalid JSON: " + err.Error())
	}
	blocks := splitFactBlocks(textBody.Text)
	if len(blocks) == 0 {
		return nil, errors.New("text is empty")
	}
	factContext := repository.NormalizeLumiContext(textBody.Context)
	locale := strings.TrimSpace(strings.ToLower(textBody.Locale))
	if locale == "" {
		locale = "ru"
	}
	facts := make([]repository.LumiFact, 0, len(blocks))
	for _, block := range blocks {
		facts = append(facts, repository.LumiFact{
			CourseCode: strings.TrimSpace(strings.ToLower(textBody.CourseCode)),
			Context:    factContext,
			Locale:     locale,
			Body:       block,
			Status:     "active",
		})
	}
	return facts, nil
}

type lumiFactPromptCourse struct {
	Code           string
	Title          string
	TargetLanguage string
	CityName       string
}

func (r *Router) lumiFactPromptCourses(ctx context.Context) ([]lumiFactPromptCourse, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT code, title, target_lang, city_name
		FROM courses
		WHERE status = 'active'
		ORDER BY code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []lumiFactPromptCourse
	for rows.Next() {
		var c lumiFactPromptCourse
		if err := rows.Scan(&c.Code, &c.Title, &c.TargetLanguage, &c.CityName); err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}
	return courses, rows.Err()
}

func lumiFactsGenerationPrompt(courses []lumiFactPromptCourse) string {
	var b strings.Builder
	b.WriteString("Ты — методист языкового приложения Linglow. Сгенерируй пакет Lumi Facts для админского JSON-импорта.\n\n")
	b.WriteString("Верни СТРОГО валидный JSON без markdown, без комментариев и без текста вокруг. Формат:\n\n")
	b.WriteString(`{
  "facts": [
    {
      "course_code": "es_ru",
      "context": "grammar",
      "locale": "ru",
      "body": "Короткий, точный и интересный факт на русском."
    }
  ]
}`)
	b.WriteString("\n\nДоступные course_code:\n")
	if len(courses) == 0 {
		b.WriteString("- en_ru — English for Russian speakers\n")
		b.WriteString("- es_ru — Spanish for Russian speakers\n")
	} else {
		for _, c := range courses {
			b.WriteString("- " + c.Code)
			if c.Title != "" {
				b.WriteString(" — " + c.Title)
			}
			if c.TargetLanguage != "" || c.CityName != "" {
				b.WriteString(" (")
				parts := []string{}
				if c.TargetLanguage != "" {
					parts = append(parts, "target_lang: "+c.TargetLanguage)
				}
				if c.CityName != "" {
					parts = append(parts, "city: "+c.CityName)
				}
				b.WriteString(strings.Join(parts, ", "))
				b.WriteString(")")
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("- Пустой course_code \"\" означает факт для всех курсов; используй редко, только если факт универсален.\n\n")

	b.WriteString("Доступные context и темы:\n")
	b.WriteString("- general: общие экраны и дашборд; история языка, этимология, рекорды, культурные факты.\n")
	b.WriteString("- grammar: экраны грамматики; правила, исключения, сравнения с русским, частые ошибки.\n")
	b.WriteString("- reading: экраны чтения; техники чтения, жанры, уровни текстов, авторы, работа с контекстом.\n")
	b.WriteString("- practice: практические упражнения; evidence-based техники обучения, память, интервальные повторения, привычки.\n")
	b.WriteString("- progress: экран прогресса; CEFR, плато, измеримый прогресс, мотивация на основе исследований.\n")
	b.WriteString("- city: карта города Linglow; география, история, архитектура и культура городов носителей языка.\n\n")

	b.WriteString("Правила:\n")
	b.WriteString("- locale всегда \"ru\".\n")
	b.WriteString("- body пиши на русском, 1–2 предложения, без списков внутри факта.\n")
	b.WriteString("- Каждый факт должен быть конкретным, проверяемым и небанальным; не используй общие мотивационные фразы.\n")
	b.WriteString("- Не дублируй смысл внутри ответа; избегай повторов известных seed-фактов вроде общих утверждений про 500 млн носителей испанского или OK как самое узнаваемое слово.\n")
	b.WriteString("- Для grammar фактов давай практическую языковую пользу, а не только энциклопедию.\n")
	b.WriteString("- Для city фактов связывай город/культуру с языком курса.\n")
	b.WriteString("- Не добавляй поля кроме course_code, context, locale, body.\n")
	return b.String()
}
