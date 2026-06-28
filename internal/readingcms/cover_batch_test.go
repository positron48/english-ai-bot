package readingcms

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestBatchPhaseDone(t *testing.T) {
	if batchPhaseDone(2, 1, 1) != 4 {
		t.Fatalf("got %d", batchPhaseDone(2, 1, 1))
	}
}

func TestBatchPercentTwoPhase(t *testing.T) {
	cases := []struct {
		name           string
		total          int
		phaseDone1     int
		phaseDone2     int
		current        int
		currentPercent int
		running        bool
		done           bool
		errMsg         string
		want           int
	}{
		{"empty", 0, 0, 0, 0, 0, false, false, "", 0},
		{"done", 5, 5, 5, 0, 0, false, true, "", 100},
		{"half first phase one at 50", 4, 1, 0, 2, 50, true, false, "", 18},
		{"first phase done", 3, 3, 0, 0, 0, false, false, "", 50},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := batchPercentTwoPhase(tc.total, tc.phaseDone1, tc.phaseDone2, tc.current, tc.currentPercent, tc.running, tc.done, tc.errMsg)
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
	if view.Total != 1 {
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
	if total != 1 {
		t.Fatalf("total=%d", total)
	}
}

func TestCoverBatchProgressMissing(t *testing.T) {
	svc := &Service{}
	if _, ok := svc.CoverBatchProgress("missing"); ok {
		t.Fatal("expected missing batch")
	}
}
