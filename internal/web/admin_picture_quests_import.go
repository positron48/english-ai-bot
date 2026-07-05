package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// pictureQuestImportPayload is the JSON shape accepted by the admin import: one picture quest plus
// its full task list. course_code in the body wins over the query param when present.
type pictureQuestImportPayload struct {
	CourseCode string `json:"course_code"`
	pictureQuestPayload
	Tasks []taskPayload `json:"tasks"`
}

// handleAdminPictureQuestImport upserts one picture quest or a JSON array of quests (by course+code)
// and replaces tasks for each. course_code in each body wins over the query param when present.
// @Summary      Admin: импорт квеста «опиши картинку» из JSON
// @Tags         Admin
// @Router       /api/admin/picture-quests/import [post]
func (r *Router) handleAdminPictureQuestImport(w http.ResponseWriter, req *http.Request) {
	if r.pictureQuestRepo == nil {
		http.Error(w, "Picture quest repository is not available", http.StatusServiceUnavailable)
		return
	}
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Не удалось прочитать тело запроса", http.StatusBadRequest)
		return
	}
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		http.Error(w, "Пустое тело запроса", http.StatusBadRequest)
		return
	}

	defaultCourse := r.adminConversationCourseCode(req)
	ctx := req.Context()

	if body[0] == '[' {
		var payloads []pictureQuestImportPayload
		if err := json.Unmarshal(body, &payloads); err != nil {
			http.Error(w, "Невалидный JSON-массив: "+err.Error(), http.StatusBadRequest)
			return
		}
		if len(payloads) == 0 {
			http.Error(w, "Массив квестов пуст", http.StatusBadRequest)
			return
		}
		results := make([]map[string]interface{}, 0, len(payloads))
		createdCount := 0
		for i := range payloads {
			res, ierr := r.importPictureQuestPayload(ctx, &payloads[i], defaultCourse, i+1)
			if ierr != nil {
				if httpErr, ok := ierr.(*pictureQuestImportHTTPError); ok {
					http.Error(w, httpErr.msg, httpErr.status)
					return
				}
				r.logger.Error("admin import picture quest batch", zap.Int("index", i+1), zap.Error(ierr))
				http.Error(w, "Квест #"+strconv.Itoa(i+1)+": "+ierr.Error(), http.StatusBadRequest)
				return
			}
			if res.created {
				createdCount++
			}
			results = append(results, res.response)
		}
		writeJSON(w, map[string]interface{}{
			"success":       true,
			"imported":      len(results),
			"created_count": createdCount,
			"updated_count": len(results) - createdCount,
			"results":       results,
		})
		return
	}

	p := pictureQuestImportPayload{}
	if err := json.Unmarshal(body, &p); err != nil {
		http.Error(w, "Невалидный JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	res, err := r.importPictureQuestPayload(ctx, &p, defaultCourse, 0)
	if err != nil {
		if httpErr, ok := err.(*pictureQuestImportHTTPError); ok {
			http.Error(w, httpErr.msg, httpErr.status)
			return
		}
		r.logger.Error("admin import picture quest", zap.Error(err))
		http.Error(w, "Не удалось импортировать: "+err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, res.response)
}

type pictureQuestImportHTTPError struct {
	msg    string
	status int
}

func (e *pictureQuestImportHTTPError) Error() string { return e.msg }

type pictureQuestImportResult struct {
	created  bool
	response map[string]interface{}
}

func (r *Router) importPictureQuestPayload(ctx context.Context, p *pictureQuestImportPayload, defaultCourse string, questIndex int) (*pictureQuestImportResult, error) {
	prefix := ""
	if questIndex > 0 {
		prefix = "Квест #" + strconv.Itoa(questIndex) + ": "
	}

	courseCode := strings.TrimSpace(strings.ToLower(p.CourseCode))
	if courseCode == "" {
		courseCode = defaultCourse
	}

	in, err := p.toInput(courseCode)
	if err != nil {
		return nil, &pictureQuestImportHTTPError{msg: prefix + err.Error(), status: http.StatusBadRequest}
	}

	tasks := make([]repository.AdminPictureTaskInput, 0, len(p.Tasks))
	for i := range p.Tasks {
		ti, terr := p.Tasks[i].toPictureInput()
		if terr != nil {
			return nil, &pictureQuestImportHTTPError{
				msg:    prefix + "Задача #" + strconv.Itoa(i+1) + ": " + terr.Error(),
				status: http.StatusBadRequest,
			}
		}
		tasks = append(tasks, ti)
	}
	if len(tasks) == 0 {
		return nil, &pictureQuestImportHTTPError{
			msg:    prefix + "Квест должен содержать хотя бы одну задачу (tasks)",
			status: http.StatusBadRequest,
		}
	}

	id, created, err := r.pictureQuestRepo.ImportQuest(ctx, in, tasks)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicatePictureTaskCode) {
			return nil, &pictureQuestImportHTTPError{
				msg:    prefix + "Дублирующийся code задачи в одном квесте",
				status: http.StatusConflict,
			}
		}
		return nil, err
	}
	return &pictureQuestImportResult{
		created: created,
		response: map[string]interface{}{
			"id":          id,
			"created":     created,
			"code":        in.Code,
			"task_count":  len(tasks),
			"course_code": courseCode,
			"success":     true,
		},
	}, nil
}

// handleAdminPictureQuestPromptTemplate returns a ready-to-use prompt for an LLM to generate a
// picture quest JSON in exactly the shape the import endpoint accepts.
// @Summary      Admin: промпт для генерации квеста «опиши картинку»
// @Tags         Admin
// @Router       /api/admin/picture-quests/prompt-template [get]
func (r *Router) handleAdminPictureQuestPromptTemplate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	courseCode := r.adminConversationCourseCode(req)
	targetLang := ""
	if r.pictureQuestRepo != nil {
		if courseID, err := r.pictureQuestRepo.CourseIDByCode(req.Context(), courseCode); err == nil {
			targetLang = r.courseTargetLang(req.Context(), courseID)
		}
	}
	writeJSON(w, map[string]interface{}{
		"course_code": courseCode,
		"prompt":      pictureQuestGenerationPrompt(courseCode, targetLang),
	})
}

// pictureQuestGenerationPrompt builds the quest-generation prompt. It describes the JSON structure
// precisely and stresses that image_description must enumerate every describable fact.
func pictureQuestGenerationPrompt(courseCode, targetLang string) string {
	lang := targetLang
	if lang == "" {
		lang = "the course target language"
	}
	var b strings.Builder
	b.WriteString("Ты — методист языкового приложения. Сгенерируй ОДИН квест «опиши картинку» ")
	b.WriteString("для курса «" + courseCode + "». Ученик смотрит на картинку и описывает её на языке: " + lang + ", ")
	b.WriteString("общаясь с маскотом Lumi. Модель НЕ видит картинку — она судит по image_description, ")
	b.WriteString("поэтому image_description должен ПЕРЕЧИСЛЯТЬ ВСЕ описываемые факты картинки: объекты, ")
	b.WriteString("их цвета, количество, расположение, действия, погоду, время суток и т.п.\n\n")
	b.WriteString("Названия и формулировки для ученика (title, task.title) — на русском; ")
	b.WriteString("инструкции для модели (image_description, completion_criteria) — на английском.\n\n")

	b.WriteString("Верни СТРОГО валидный JSON (без markdown, без комментариев) такой структуры:\n\n")
	b.WriteString(`{
  "code": "string — уникальный машинный код квеста в рамках курса, snake_case",
  "title": "string — название для ученика (RU)",
  "cefr_level": "string — уровень: A0/A1/A2/B1/B2/C1",
  "image_url": "string — URL загруженной картинки (можно оставить пустым и заполнить в админке)",
  "image_description": "string (EN) — полное описание картинки для модели: ВСЕ объекты, цвета, количество, расположение, действия",
  "max_turns": 30,
  "token_budget": 40000,
  "sort_order": 0,
  "status": "draft",
  "tasks": [
    {
      "code": "string — код задачи, уникальный в квесте, snake_case",
      "title": "string — название задачи для ученика (RU)",
      "completion_criteria": "string (EN) — когда задача считается выполненной, инструкция для модели",
      "is_required": true,
      "sort_order": 0
    }
  ]
}`)
	b.WriteString("\n\nПравила:\n")
	b.WriteString("- Добавь 3–6 задач по порядку (sort_order): обязательные — назвать главные объекты/факты картинки, 1–2 необязательные — бонусные детали (за 100% выполнение ученик получает звёздочку).\n")
	b.WriteString("- Каждая completion_criteria должна опираться ТОЛЬКО на факты, которые есть в image_description, и быть объективно проверяемой по сообщению ученика.\n")
	b.WriteString("- Сложность фактов и лексики соответствует cefr_level: чем ниже уровень, тем проще объекты и описания.\n")
	return b.String()
}
