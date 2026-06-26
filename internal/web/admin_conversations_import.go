package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// conversationImportPayload is the JSON shape accepted by the admin import: one scenario plus its
// full task list. course_code in the body wins over the query param when present.
type conversationImportPayload struct {
	CourseCode string `json:"course_code"`
	scenarioPayload
	Tasks []taskPayload `json:"tasks"`
}

// handleAdminConversationImport upserts a single scenario (by course+code) and replaces its tasks.
// @Summary      Admin: импорт сценария общения из JSON
// @Tags         Admin
// @Router       /api/admin/conversations/import [post]
func (r *Router) handleAdminConversationImport(w http.ResponseWriter, req *http.Request) {
	if r.conversationRepo == nil {
		http.Error(w, "Conversation repository is not available", http.StatusServiceUnavailable)
		return
	}
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var p conversationImportPayload
	if err := json.NewDecoder(req.Body).Decode(&p); err != nil {
		http.Error(w, "Невалидный JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	courseCode := strings.TrimSpace(strings.ToLower(p.CourseCode))
	if courseCode == "" {
		courseCode = r.adminConversationCourseCode(req)
	}

	in, err := p.toInput(courseCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tasks := make([]repository.AdminTaskInput, 0, len(p.Tasks))
	for i := range p.Tasks {
		ti, terr := p.Tasks[i].toInput()
		if terr != nil {
			http.Error(w, "Задача #"+strconv.Itoa(i+1)+": "+terr.Error(), http.StatusBadRequest)
			return
		}
		tasks = append(tasks, ti)
	}
	if in.IsQuest && len(tasks) == 0 {
		http.Error(w, "Квест должен содержать хотя бы одну задачу (tasks)", http.StatusBadRequest)
		return
	}

	ctx := req.Context()
	id, created, err := r.conversationRepo.ImportScenario(ctx, in, tasks)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateTaskCode) {
			http.Error(w, "Дублирующийся code задачи в одном сценарии", http.StatusConflict)
			return
		}
		r.logger.Error("admin import scenario", zap.Error(err))
		http.Error(w, "Не удалось импортировать: "+err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{
		"id":          id,
		"created":     created,
		"task_count":  len(tasks),
		"course_code": courseCode,
		"success":     true,
	})
}

// handleAdminConversationPromptTemplate returns a ready-to-use prompt for an LLM to generate a
// scenario JSON in exactly the shape the import endpoint accepts.
// @Summary      Admin: промпт для генерации сценария общения
// @Tags         Admin
// @Router       /api/admin/conversations/prompt-template [get]
func (r *Router) handleAdminConversationPromptTemplate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	courseCode := r.adminConversationCourseCode(req)
	targetLang := ""
	if r.conversationRepo != nil {
		if courseID, err := r.conversationRepo.CourseIDByCode(req.Context(), courseCode); err == nil {
			targetLang = r.courseTargetLang(req.Context(), courseID)
		}
	}
	writeJSON(w, map[string]interface{}{
		"course_code": courseCode,
		"prompt":      conversationGenerationPrompt(courseCode, targetLang),
	})
}

// conversationGenerationPrompt builds the scenario-generation prompt. It describes the JSON
// structure precisely but keeps the creative brief loose (general phrasing, no hard limits).
func conversationGenerationPrompt(courseCode, targetLang string) string {
	lang := targetLang
	if lang == "" {
		lang = "the course target language"
	}
	var b strings.Builder
	b.WriteString("Ты — методист языкового приложения. Сгенерируй ОДИН сценарий ролевого диалога с NPC ")
	b.WriteString("для курса «" + courseCode + "». NPC и весь диалог идут на языке: " + lang + ". ")
	b.WriteString("Названия и формулировки для ученика (title, task.title) — на русском; ")
	b.WriteString("инструкции для модели (npc_persona, scene_setup, completion_criteria) — на английском.\n\n")

	b.WriteString("Это сценарий из «района общения» в городе. У каждого NPC может быть цепочка ")
	b.WriteString("последовательных сценариев, открывающихся друг за другом. Чтобы связать сценарий в цепочку, ")
	b.WriteString("задай общий npc_code и prerequisite_code (code предыдущего сценария этого NPC).\n\n")

	b.WriteString("Верни СТРОГО валидный JSON (без markdown, без комментариев) такой структуры:\n\n")
	b.WriteString(`{
  "code": "string — уникальный машинный код сценария в рамках курса, snake_case",
  "title": "string — название для ученика (RU)",
  "cefr_level": "string — уровень: A0/A1/A2/B1/B2/C1",
  "place_type": "string — тип места: cafe, shop, police_station, pharmacy, hotel, ...",
  "npc_name": "string — имя NPC",
  "npc_code": "string — код NPC для группировки цепочки сценариев (пусто = одиночный)",
  "prerequisite_code": "string — code предыдущего сценария, который надо пройти первым (пусто = доступен сразу)",
  "npc_persona": "string (EN) — характер и манера речи NPC, инструкция для модели",
  "scene_setup": "string (EN) — завязка сцены, что происходит, инструкция для модели",
  "is_quest": true,
  "max_turns": 30,
  "token_budget": 40000,
  "sort_order": 0,
  "status": "draft",
  "tasks": [
    {
      "code": "string — код задачи, уникальный в сценарии, snake_case",
      "title": "string — название задачи для ученика (RU)",
      "completion_criteria": "string (EN) — когда задача считается выполненной, инструкция для модели",
      "is_required": true,
      "sort_order": 0
    }
  ]
}`)
	b.WriteString("\n\nПравила:\n")
	b.WriteString("- Если is_quest=true — добавь 3–6 задач, по порядку (sort_order), первая обычно «поздороваться», последняя необязательная «попрощаться».\n")
	b.WriteString("- Если is_quest=false — это свободная беседа, tasks оставь пустым массивом [].\n")
	b.WriteString("- Реплики ученика и NPC соответствуют уровню cefr_level: чем ниже уровень, тем проще язык.\n")
	b.WriteString("- completion_criteria формулируй так, чтобы модель могла объективно определить выполнение по сообщению ученика.\n")
	b.WriteString("- Будь содержательным и реалистичным, без искусственных ограничений по теме.\n")
	return b.String()
}
