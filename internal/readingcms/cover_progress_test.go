package readingcms

import (
	"strings"
	"testing"
)

func TestCoverProgressRegisterAndFinish(t *testing.T) {
	svc := &Service{}
	st := svc.coverProgressRegister("es_ru", "demo_text")
	st.appendLine("[reading-cover 12:00:00] stage llm|Промпт (LLM)")
	st.appendLine("[reading-cover 12:00:01] stage comfyui|Картинка (ComfyUI)")
	st.finish(nil)

	view, ok := svc.CoverProgress("es_ru", "demo_text")
	if !ok {
		t.Fatal("expected progress state")
	}
	if view.Running {
		t.Fatal("expected not running after finish")
	}
	if !view.Done {
		t.Fatal("expected done")
	}
	if view.Error != "" {
		t.Fatalf("unexpected error: %q", view.Error)
	}
	if view.Percent != 100 {
		t.Fatalf("expected percent 100, got %d", view.Percent)
	}
	if !strings.Contains(view.Log, "started:") {
		t.Fatalf("unexpected log: %q", view.Log)
	}
}

func TestCoverProgressLLMTimeline(t *testing.T) {
	st := &coverProgressState{}
	st.initStages()
	st.activateStage("prepare", "Подготовка")

	lines := []string{
		"started cover generation: course=es_ru text_id=demo",
		"[reading-cover 21:59:18] step 1/4: LLM cover scene prompt (local llama.cpp)",
		"[reading-cover 21:59:18] LLM: requesting scene description from local llama.cpp…",
		"[reading-llm] port 8090 already in use — ждём готовности llama.cpp",
		"[reading-llm] llama.cpp ready after START_CMD (3s)",
		"[reading-llm] → http://127.0.0.1:8090/v1/chat/completions model=qwen3:30b",
		"[reading-cover 22:00:54] LLM: response in 95.1s (53 chars)",
		"[reading-cover 22:00:54] prompt ready (87 chars): casual watercolor",
	}
	for _, line := range lines {
		st.appendLine(line)
	}
	view := st.view()
	if view.Stage != "llm" {
		t.Fatalf("expected llm stage, got %q", view.Stage)
	}
	if view.Percent < 35 || view.Percent > 50 {
		t.Fatalf("expected llm progress ~40%%, got %d (%s)", view.Percent, view.StageLabel)
	}
	if !strings.Contains(view.StageLabel, "промпт") && !strings.Contains(view.StageLabel, "Промпт") && !strings.Contains(view.StageLabel, "LLM") {
		t.Fatalf("unexpected stage label: %q", view.StageLabel)
	}
}

func TestCoverProgressComfyUITimeline(t *testing.T) {
	st := &coverProgressState{}
	st.initStages()
	st.appendLine("[reading-cover 22:00:54] step 2/4: ComfyUI txt2img -> cover_raw.png")
	st.appendLine("[reading-cover 22:00:54] ComfyUI: queueing workflow…")
	st.appendLine("[reading-cover 22:00:54] ComfyUI: prompt_id=abc")
	st.appendLine("[reading-cover 22:01:55] ComfyUI: still waiting… 61s elapsed")

	view := st.view()
	if view.Stage != "comfyui" {
		t.Fatalf("expected comfyui stage, got %q", view.Stage)
	}
	if view.Percent < 70 || view.Percent > 85 {
		t.Fatalf("expected comfyui progress ~75%%, got %d", view.Percent)
	}
	if !strings.Contains(view.StageLabel, "61s") {
		t.Fatalf("expected elapsed in label, got %q", view.StageLabel)
	}
}

func TestCoverProgressStageFromStepLine(t *testing.T) {
	st := &coverProgressState{}
	st.initStages()
	st.activateStage("prepare", "Подготовка")
	st.appendLine("[reading-cover 12:00:00] step 2/4: ComfyUI txt2img")
	if st.activeID != "comfyui" {
		t.Fatalf("expected comfyui active, got %q", st.activeID)
	}
}

func TestCoverProgressSkipLLMFullPipeline(t *testing.T) {
	st := &coverProgressState{}
	st.initStages()
	st.activateStage("prepare", "Подготовка")

	lines := []string{
		"[reading-cover 09:48:11] started: course=es_ru text_id=demo",
		"[reading-cover 09:48:11] === demo — Tapas (prompt only, no LLM) ===",
		"[reading-cover 09:48:11] stage prepare|Подготовка",
		"[reading-cover 09:48:11] step 1/4: skipped LLM — using provided image prompt",
		"[reading-cover 09:48:11] stage comfyui|Картинка (ComfyUI)",
		"▶ ComfyUI prompt (95 chars)",
		"[reading-cover 09:48:11] step 2/4: ComfyUI txt2img -> cover_raw.png",
		"[reading-cover 09:49:09] ComfyUI: saved cover_raw.png (1271687 bytes)",
		"[reading-cover 09:49:09] stage resize|WebP thumb + hero",
		"[reading-cover 09:49:09] step 3/4: resize WebP thumb + hero",
		"[reading-cover 09:49:09]   thumb: /tmp/cover_thumb.webp (29586 bytes)",
		"[reading-cover 09:49:09] stage save|Сохранение",
		"[reading-cover 09:49:09] step 4/4: update text JSON",
		"[reading-cover 09:49:09] stage done|Готово",
		"[reading-cover 09:49:09] done",
	}
	for _, line := range lines {
		st.appendLine(line)
	}
	st.finish(nil)

	view := st.view()
	if !view.Done {
		t.Fatal("expected done")
	}
	if view.Percent != 100 {
		t.Fatalf("expected percent 100, got %d", view.Percent)
	}
	for _, stg := range view.Stages {
		if stg.Status != "done" {
			t.Fatalf("stage %q expected done, got %q", stg.ID, stg.Status)
		}
	}
}
