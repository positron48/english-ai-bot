package readingcms

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type coverStageDef struct {
	ID    string
	Label string
}

var coverStageDefs = []coverStageDef{
	{ID: "prepare", Label: "Подготовка"},
	{ID: "llm", Label: "Промпт (LLM)"},
	{ID: "comfyui", Label: "Картинка (ComfyUI)"},
	{ID: "resize", Label: "WebP thumb + hero"},
	{ID: "save", Label: "Сохранение"},
}

var coverStageSpan = map[string]struct {
	Base int
	Span int
}{
	"prepare": {0, 3},
	"llm":     {3, 42},
	"comfyui": {45, 40},
	"resize":  {85, 8},
	"save":    {93, 7},
}

var (
	reLLMResponseSec  = regexp.MustCompile(`LLM: response in ([0-9.]+)s`)
	reComfyWaitSec    = regexp.MustCompile(`still waiting… (\d+)s`)
	reComfyPromptID   = regexp.MustCompile(`ComfyUI: prompt_id=`)
)

// CoverStageView is one pipeline step for the CMS progress UI.
type CoverStageView struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Status string `json:"status"`
}

// CoverProgressView is the live cover generation state for polling.
type CoverProgressView struct {
	Log        string           `json:"log"`
	Running    bool             `json:"running"`
	Done       bool             `json:"done"`
	Error      string           `json:"error"`
	Percent    int              `json:"percent"`
	Stage      string           `json:"stage"`
	StageLabel string           `json:"stage_label"`
	Stages     []CoverStageView `json:"stages"`
}

type coverProgressState struct {
	mu            sync.Mutex
	lines         []string
	stages        []CoverStageView
	running       bool
	done          bool
	errMsg        string
	started       time.Time
	activeID      string
	activeLabel   string
	stageDetail   string
	stageMilestone float64
}

func coverProgressKey(courseCode, textID string) string {
	return strings.ToLower(strings.TrimSpace(courseCode)) + ":" + strings.TrimSpace(textID)
}

func newCoverStages() []CoverStageView {
	out := make([]CoverStageView, len(coverStageDefs))
	for i, def := range coverStageDefs {
		out[i] = CoverStageView{ID: def.ID, Label: def.Label, Status: "pending"}
	}
	return out
}

func (s *coverProgressState) initStages() {
	s.stages = newCoverStages()
}

func stageIndex(stages []CoverStageView, id string) int {
	for i, st := range stages {
		if st.ID == id {
			return i
		}
	}
	return -1
}

func (s *coverProgressState) activateStage(id, label string) {
	idx := stageIndex(s.stages, id)
	if idx < 0 {
		return
	}
	if s.activeID != id {
		s.stageMilestone = 0.08
		s.stageDetail = ""
	}
	for i := range s.stages {
		switch {
		case i < idx && s.stages[i].Status != "error":
			s.stages[i].Status = "done"
		case i == idx:
			s.stages[i].Status = "active"
			if strings.TrimSpace(label) != "" {
				s.stages[i].Label = label
			}
		case i > idx && s.stages[i].Status == "active":
			s.stages[i].Status = "pending"
		}
	}
	s.activeID = s.stages[idx].ID
	s.activeLabel = s.stages[idx].Label
}

func (s *coverProgressState) setMilestone(m float64, detail string) {
	if m > s.stageMilestone {
		s.stageMilestone = m
	}
	if strings.TrimSpace(detail) != "" {
		s.stageDetail = detail
	}
}

func parseCoverStageLine(line string) (id, label string, ok bool) {
	for _, prefix := range []string{"] stage ", "] stage\t"} {
		if idx := strings.Index(line, prefix); idx >= 0 {
			payload := strings.TrimSpace(line[idx+len(prefix):])
			parts := strings.SplitN(payload, "|", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), true
			}
		}
	}
	if idx := strings.Index(line, "[reading-cover stage]"); idx >= 0 {
		payload := strings.TrimSpace(line[idx+len("[reading-cover stage]"):])
		parts := strings.SplitN(payload, "|", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), true
		}
	}
	return "", "", false
}

func (s *coverProgressState) applyLineForStages(line string) {
	if id, label, ok := parseCoverStageLine(line); ok {
		if id == "done" {
			s.stageMilestone = 1
			s.stageDetail = "Готово"
			s.completeAll()
			return
		}
		s.activateStage(id, label)
		return
	}

	lower := strings.ToLower(line)

	switch {
	case strings.Contains(lower, "started:") || strings.Contains(lower, "started cover generation"):
		s.activateStage("prepare", "Подготовка")
		s.setMilestone(0.5, "Старт пайплайна")
	case strings.Contains(line, "===") && strings.Contains(lower, "reading-cover"):
		s.activateStage("prepare", "Подготовка")
		s.setMilestone(0.9, "Чтение текста")
	case strings.Contains(lower, "step 1/4"):
		if strings.Contains(lower, "skipped llm") {
			if idx := stageIndex(s.stages, "prepare"); idx >= 0 {
				s.stages[idx].Status = "done"
			}
			if idx := stageIndex(s.stages, "llm"); idx >= 0 {
				s.stages[idx].Status = "done"
				s.stages[idx].Label = "Промпт (без LLM)"
			}
			s.activeID = "llm"
			s.activeLabel = "Промпт (без LLM)"
			s.setMilestone(1, "Готовый промпт")
		} else {
			s.activateStage("llm", "Промпт (LLM)")
			s.setMilestone(0.05, "Генерация сцены для обложки")
		}
	case strings.Contains(lower, "llm: requesting"):
		s.activateStage("llm", "Промпт (LLM)")
		s.setMilestone(0.1, "Запрос к локальной LLM…")
	case strings.Contains(line, "[reading-llm"):
		s.activateStage("llm", "Промпт (LLM)")
		switch {
		case strings.Contains(lower, "занят (tcp)") || strings.Contains(lower, "проверяем llama"):
			s.setMilestone(0.12, "Проверка llama.cpp…")
		case strings.Contains(lower, "запускаем start_cmd"):
			s.setMilestone(0.2, "Запуск llama.cpp (START_CMD)…")
		case strings.Contains(lower, "ready after start_cmd") || strings.Contains(lower, "ready после"):
			s.setMilestone(0.35, "llama.cpp готов")
		case strings.Contains(lower, "chat/completions") || strings.Contains(lower, "→ http"):
			s.setMilestone(0.5, "Генерация промпта (LLM)…")
		}
	case strings.Contains(lower, "comfyui_prompt|"):
		s.activateStage("comfyui", "Картинка (ComfyUI)")
		s.setMilestone(0.12, "Промпт отправлен в ComfyUI")
	case strings.Contains(lower, "▶ comfyui prompt"):
		s.activateStage("comfyui", "Картинка (ComfyUI)")
		s.setMilestone(0.1, "Промпт для ComfyUI готов")
	case strings.Contains(lower, "llm: response in"):
		s.activateStage("llm", "Промпт (LLM)")
		sec := ""
		if m := reLLMResponseSec.FindStringSubmatch(line); len(m) == 2 {
			sec = m[1] + "s"
		}
		if sec != "" {
			s.setMilestone(0.82, "Ответ LLM получен за "+sec)
		} else {
			s.setMilestone(0.82, "Ответ LLM получен")
		}
	case strings.Contains(lower, "llm: scene parsed") || strings.Contains(lower, "llm: final prompt"):
		s.activateStage("llm", "Промпт (LLM)")
		s.setMilestone(0.95, "Промпт для картинки готов")
	case strings.Contains(lower, "step 2/4"):
		s.activateStage("comfyui", "Картинка (ComfyUI)")
		s.setMilestone(0.05, "Запуск ComfyUI txt2img")
	case strings.Contains(lower, "comfyui: queueing"):
		s.activateStage("comfyui", "Картинка (ComfyUI)")
		s.setMilestone(0.15, "Постановка в очередь ComfyUI…")
	case reComfyPromptID.MatchString(line):
		s.activateStage("comfyui", "Картинка (ComfyUI)")
		s.setMilestone(0.22, "ComfyUI: рендер в работе…")
	case strings.Contains(lower, "comfyui: still waiting"):
		s.activateStage("comfyui", "Картинка (ComfyUI)")
		elapsed := 0
		if m := reComfyWaitSec.FindStringSubmatch(line); len(m) == 2 {
			elapsed, _ = strconv.Atoi(m[1])
		}
		milestone := 0.25 + float64(elapsed)/100.0
		if milestone > 0.95 {
			milestone = 0.95
		}
		detail := "ComfyUI: рендер…"
		if elapsed > 0 {
			detail = fmt.Sprintf("ComfyUI: рендер… %ds", elapsed)
		}
		s.setMilestone(milestone, detail)
	case strings.Contains(lower, "comfyui: render finished") || strings.Contains(lower, "comfyui: downloading"):
		s.activateStage("comfyui", "Картинка (ComfyUI)")
		s.setMilestone(0.96, "Скачивание PNG из ComfyUI…")
	case strings.Contains(lower, "comfyui: saved"):
		s.activateStage("comfyui", "Картинка (ComfyUI)")
		s.setMilestone(1, "PNG сохранён")
	case strings.Contains(lower, "step 3/4"):
		s.activateStage("resize", "WebP thumb + hero")
		s.setMilestone(0.2, "Конвертация в WebP…")
	case strings.Contains(lower, "thumb:") || strings.Contains(lower, "hero:"):
		s.activateStage("resize", "WebP thumb + hero")
		s.setMilestone(0.85, "WebP файлы записаны")
	case strings.Contains(lower, "step 4/4"):
		s.activateStage("save", "Сохранение")
		s.setMilestone(0.3, "Обновление JSON текста…")
	case strings.Contains(lower, "] done"):
		s.stageMilestone = 1
		s.stageDetail = "Готово"
		s.completeAll()
	}
}

func (s *coverProgressState) completeAll() {
	for i := range s.stages {
		if s.stages[i].Status != "error" {
			s.stages[i].Status = "done"
		}
	}
	s.activeID = "done"
	s.activeLabel = "Готово"
	s.stageMilestone = 1
}

func (s *coverProgressState) markActiveError() {
	for i := range s.stages {
		if s.stages[i].Status == "active" {
			s.stages[i].Status = "error"
			return
		}
	}
}

func (s *coverProgressState) percent() int {
	if s.done && s.errMsg == "" {
		return 100
	}
	if s.activeID == "done" {
		return 100
	}
	span, ok := coverStageSpan[s.activeID]
	if !ok {
		return 0
	}
	m := s.stageMilestone
	if m < 0 {
		m = 0
	}
	if m > 1 {
		m = 1
	}
	p := span.Base + int(float64(span.Span)*m)
	if p > 99 && !s.done {
		return 99
	}
	return p
}

func (s *coverProgressState) displayLabel() string {
	if strings.TrimSpace(s.stageDetail) != "" {
		return s.stageDetail
	}
	return s.activeLabel
}

func (s *coverProgressState) appendLine(line string) {
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lines = append(s.lines, line)
	s.applyLineForStages(line)
}

func (s *coverProgressState) finish(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	s.done = true
	if err != nil {
		s.errMsg = err.Error()
		s.markActiveError()
		return
	}
	s.completeAll()
}

func (s *coverProgressState) view() CoverProgressView {
	s.mu.Lock()
	defer s.mu.Unlock()
	stages := make([]CoverStageView, len(s.stages))
	copy(stages, s.stages)
	return CoverProgressView{
		Log:        strings.Join(s.lines, "\n"),
		Running:    s.running,
		Done:       s.done,
		Error:      s.errMsg,
		Percent:    s.percent(),
		Stage:      s.activeID,
		StageLabel: s.displayLabel(),
		Stages:     stages,
	}
}

func (s *Service) coverProgressRegister(courseCode, textID string) *coverProgressState {
	st := &coverProgressState{running: true, started: time.Now().UTC()}
	st.initStages()
	st.activateStage("prepare", "Подготовка")
	st.appendLine(fmt.Sprintf("[reading-cover %s] started: course=%s text_id=%s",
		time.Now().Format("15:04:05"), courseCode, textID))
	s.coverJobs.Store(coverProgressKey(courseCode, textID), st)
	return st
}

func (s *Service) coverProgressUnregister(courseCode, textID string) {
	s.coverJobs.Delete(coverProgressKey(courseCode, textID))
}

func (s *Service) CoverProgress(courseCode, textID string) (CoverProgressView, bool) {
	v, ok := s.coverJobs.Load(coverProgressKey(courseCode, textID))
	if !ok {
		return CoverProgressView{}, false
	}
	return v.(*coverProgressState).view(), true
}
