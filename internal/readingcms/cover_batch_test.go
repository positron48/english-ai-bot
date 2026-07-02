package readingcms

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBatchPhaseDone(t *testing.T) {
	if batchPhaseDone(2, 1, 1) != 4 {
		t.Fatalf("got %d", batchPhaseDone(2, 1, 1))
	}
}

func TestBatchPercent(t *testing.T) {
	cases := []struct {
		name    string
		total   int
		doneOps int
		done    bool
		errMsg  string
		want    int
	}{
		{"empty", 0, 0, false, "", 0},
		{"done", 5, 5, true, "", 100},
		{"one of four ignores current operation", 4, 1, false, "", 25},
		{"two of three", 3, 2, false, "", 66},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := batchPercent(tc.total, tc.doneOps, tc.done, tc.errMsg)
			if got != tc.want {
				t.Fatalf("got %d want %d", got, tc.want)
			}
		})
	}
}

func TestCoverBatchStateView(t *testing.T) {
	svc := &Service{}
	st := &coverBatchState{
		batchID:    "es_ru:123",
		courseCode: "es_ru",
		phase:      coverBatchPhasePrompts,
		total:      2,
		current:    1,
		pCompleted: 1,
		running:    true,
	}
	st.appendLine("line one")
	st.appendHeader("prompts", 1, 2, "t1", "Title")
	st.setPhase(coverBatchPhaseImages)
	st.setCurrent(2, "t2", "Second")

	view := st.view(svc)
	if view.BatchID != "es_ru:123" || view.Total != 2 || view.Completed != 1 {
		t.Fatalf("view=%+v", view)
	}
	if !strings.Contains(view.Log, "line one") {
		t.Fatalf("log=%q", view.Log)
	}
	if view.Phase != coverBatchPhaseImages {
		t.Fatalf("phase=%q", view.Phase)
	}

	st.finish(nil)
	doneView := st.view(svc)
	if !doneView.Done || doneView.Running {
		t.Fatalf("done view=%+v", doneView)
	}
	if doneView.Percent != 100 {
		t.Fatalf("percent=%d", doneView.Percent)
	}
}

func TestCoverBatchStateFinishWithError(t *testing.T) {
	st := &coverBatchState{running: true}
	st.finish(errInvalid("boom"))
	if st.errMsg == "" || st.running || !st.done {
		t.Fatalf("state=%+v", st)
	}
}

func TestNewBatchID(t *testing.T) {
	id := newBatchID("ES_RU")
	if !strings.HasPrefix(id, "es_ru:") {
		t.Fatalf("id=%q", id)
	}
}

func TestPlanCoverBatchTexts(t *testing.T) {
	svc, _ := setupCoverService(t)
	textA := "batch_text_a"
	textB := "batch_text_b"
	seedPublishedDoc(t, svc, textA, &TextDocument{
		ID: textA, Title: "A", Level: "A2", TargetLanguage: "es",
		ReadingPassage: map[string]interface{}{"segments": []interface{}{}},
	})
	seedPublishedDoc(t, svc, textB, &TextDocument{
		ID: textB, Title: "B", Level: "B1", TargetLanguage: "es",
		ReadingPassage: map[string]interface{}{"segments": []interface{}{}},
	})

	ids, err := svc.planCoverBatchTexts("es_ru", "A2", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != textA {
		t.Fatalf("ids=%v", ids)
	}

	all, err := svc.planCoverBatchTexts("es_ru", "", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Fatalf("limit ids=%v", all)
	}
}

func TestPlanCoverBatchWorkSkipsReadyAndSplitsPhases(t *testing.T) {
	svc, _ := setupCoverService(t)
	readyID := "batch_ready"
	promptID := "batch_prompt_existing"
	missingID := "batch_prompt_missing"
	for _, tc := range []struct {
		id     string
		title  string
		prompt string
		ready  bool
	}{
		{readyID, "Ready", "ready prompt", true},
		{promptID, "Prompt", "saved prompt", false},
		{missingID, "Missing", "", false},
	} {
		doc := &TextDocument{
			ID: tc.id, Title: tc.title, Level: "A2", TargetLanguage: "es",
			CoverImagePrompt: tc.prompt,
			ReadingPassage:   map[string]interface{}{"segments": []interface{}{}},
		}
		if tc.ready {
			doc.CoverThumbRelPath = "assets/reading/" + tc.id + "/cover_thumb.webp"
			doc.CoverHeroRelPath = "assets/reading/" + tc.id + "/cover_hero.webp"
		}
		seedPublishedDoc(t, svc, tc.id, doc)
		if tc.ready {
			course, err := svc.paths.Course("es_ru")
			if err != nil {
				t.Fatal(err)
			}
			dir := filepath.Join(course.GrammarDir, "assets", "reading", tc.id)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"cover_thumb.webp", "cover_hero.webp"} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte("mock"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
		}
	}

	plan, err := svc.planCoverBatchWork(CoverBatchRequest{CourseCode: "es_ru", Level: "A2"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(plan.promptIDs, ",") != missingID {
		t.Fatalf("promptIDs=%v", plan.promptIDs)
	}
	if strings.Join(plan.imageIDs, ",") != promptID+","+missingID {
		t.Fatalf("imageIDs=%v", plan.imageIDs)
	}
	if plan.totalOps != 3 {
		t.Fatalf("totalOps=%d", plan.totalOps)
	}

	plan, err = svc.planCoverBatchWork(CoverBatchRequest{CourseCode: "es_ru", Level: "A2", SkipPrompts: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.promptIDs) != 0 || strings.Join(plan.imageIDs, ",") != promptID || plan.totalOps != 1 {
		t.Fatalf("promptIDs=%v imageIDs=%v totalOps=%d", plan.promptIDs, plan.imageIDs, plan.totalOps)
	}
}

func TestStartCoverBatchPromptPhaseMock(t *testing.T) {
	svc, _ := setupCoverService(t)
	textID := "batch_prompt_only"
	seedPublishedDoc(t, svc, textID, &TextDocument{
		ID: textID, Title: "Batch", Level: "A2", TargetLanguage: "es",
		ReadingPassage: map[string]interface{}{
			"segments": []map[string]interface{}{{"text": "Hola"}},
		},
	})

	batchID, total, err := svc.StartCoverBatch(CoverBatchRequest{CourseCode: "es_ru", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || batchID == "" {
		t.Fatalf("batchID=%q total=%d", batchID, total)
	}

	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		view, ok := svc.CoverBatchProgress(batchID)
		if !ok {
			t.Fatal("batch job missing")
		}
		if view.Done {
			if view.Phase == coverBatchPhaseImages || view.Phase == coverBatchPhaseStoppingLLM || view.Phase == "" {
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	view, ok := svc.CoverBatchProgress(batchID)
	if !ok {
		t.Fatal("batch progress missing")
	}
	if view.Total != 2 {
		t.Fatalf("total=%d", view.Total)
	}
	if !strings.Contains(view.Log, "[batch] started") {
		t.Fatalf("log=%q", view.Log)
	}

	course, err := svc.paths.Course("es_ru")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := readTextFile(course.GrammarDir, mustLoadIndex(t, course.GrammarDir), textID)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(doc.CoverImagePrompt) == "" {
		t.Fatal("expected prompt after batch phase 1")
	}
}

func TestStartCoverBatchSkipPromptsMock(t *testing.T) {
	svc, _ := setupCoverService(t)
	textID := "batch_skip_prompts"
	seedPublishedDoc(t, svc, textID, &TextDocument{
		ID: textID, Title: "Batch Images", Level: "A2", TargetLanguage: "es",
		CoverImagePrompt: "saved watercolor prompt",
		ReadingPassage: map[string]interface{}{
			"segments": []map[string]interface{}{{"text": "Hola"}},
		},
	})

	batchID, total, err := svc.StartCoverBatch(CoverBatchRequest{CourseCode: "es_ru", Limit: 1, SkipPrompts: true})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || batchID == "" {
		t.Fatalf("batchID=%q total=%d", batchID, total)
	}

	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		view, ok := svc.CoverBatchProgress(batchID)
		if !ok {
			t.Fatal("batch job missing")
		}
		if view.Done {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	view, ok := svc.CoverBatchProgress(batchID)
	if !ok {
		t.Fatal("batch progress missing")
	}
	if !view.Done || view.Error != "" {
		t.Fatalf("done=%v error=%q log=%q", view.Done, view.Error, view.Log)
	}
	if view.Completed != 1 || view.Percent != 100 {
		t.Fatalf("completed=%d percent=%d", view.Completed, view.Percent)
	}
	if !strings.Contains(view.Log, "skip_prompts=true") || strings.Contains(view.Log, "phase 1/2") {
		t.Fatalf("log=%q", view.Log)
	}

	course, err := svc.paths.Course("es_ru")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := readTextFile(course.GrammarDir, mustLoadIndex(t, course.GrammarDir), textID)
	if err != nil {
		t.Fatal(err)
	}
	if CoverStats(doc, course.GrammarDir) != CoverReady {
		t.Fatalf("expected ready cover, got %s", CoverStats(doc, course.GrammarDir))
	}
}

func TestStartCoverBatchValidation(t *testing.T) {
	svc, _ := setupCoverService(t)
	if _, _, err := svc.StartCoverBatch(CoverBatchRequest{}); err == nil {
		t.Fatal("expected course_code required")
	}
	if _, _, err := svc.StartCoverBatch(CoverBatchRequest{CourseCode: "es_ru"}); err == nil {
		t.Fatal("expected no texts error")
	}
}

func TestGenerateCoverBatchWrapper(t *testing.T) {
	svc, _ := setupCoverService(t)
	textID := "batch_wrap"
	seedPublishedDoc(t, svc, textID, &TextDocument{
		ID: textID, Title: "Wrap", Level: "A2", TargetLanguage: "es",
		ReadingPassage: map[string]interface{}{"segments": []interface{}{}},
	})
	total, err := svc.GenerateCoverBatch(context.Background(), CoverBatchRequest{CourseCode: "es_ru", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Fatalf("total=%d", total)
	}
}

func TestCoverBatchProgressMissing(t *testing.T) {
	svc := &Service{}
	if _, ok := svc.CoverBatchProgress("missing"); ok {
		t.Fatal("expected missing batch")
	}
}
