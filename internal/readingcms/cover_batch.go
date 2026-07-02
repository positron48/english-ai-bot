package readingcms

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	coverBatchPhasePrompts     = "prompts"
	coverBatchPhaseStoppingLLM = "stopping_llm"
	coverBatchPhaseImages      = "images"
)

// CoverBatchProgressView is live batch cover generation state for polling.
type CoverBatchProgressView struct {
	BatchID         string            `json:"batch_id"`
	CourseCode      string            `json:"course_code"`
	Phase           string            `json:"phase"`
	Running         bool              `json:"running"`
	Done            bool              `json:"done"`
	Error           string            `json:"error"`
	Total           int               `json:"total"`
	Current         int               `json:"current"`
	Completed       int               `json:"completed"`
	Skipped         int               `json:"skipped"`
	Failed          int               `json:"failed"`
	Remaining       int               `json:"remaining"`
	ETASeconds      int               `json:"eta_seconds"`
	Percent         int               `json:"percent"`
	CurrentTextID   string            `json:"current_text_id"`
	CurrentTitle    string            `json:"current_text_title"`
	Log             string            `json:"log"`
	CurrentProgress CoverProgressView `json:"current_progress"`
}

type coverBatchState struct {
	mu            sync.Mutex
	batchID       string
	courseCode    string
	phase         string
	total         int
	current       int
	pCompleted    int
	pSkipped      int
	pFailed       int
	iCompleted    int
	iSkipped      int
	iFailed       int
	currentTextID string
	currentTitle  string
	skipPrompts   bool
	startedAt     time.Time
	lines         []string
	running       bool
	done          bool
	errMsg        string
}

func newBatchID(courseCode string) string {
	return strings.ToLower(strings.TrimSpace(courseCode)) + ":" + fmt.Sprintf("%d", time.Now().UnixNano())
}

func (s *coverBatchState) appendLine(line string) {
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lines = append(s.lines, line)
}

func (s *coverBatchState) appendHeader(phase string, n, total int, textID, title string) {
	s.appendLine(fmt.Sprintf("[batch:%s] === %d/%d — %s — %s ===", phase, n, total, textID, title))
}

func (s *coverBatchState) setPhase(phase string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.phase = phase
}

func (s *coverBatchState) setCurrent(n int, textID, title string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.current = n
	s.currentTextID = textID
	s.currentTitle = title
}

func (s *coverBatchState) finish(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	s.done = true
	if err != nil {
		s.errMsg = err.Error()
	}
}

func batchPhaseDone(completed, skipped, failed int) int {
	return completed + skipped + failed
}

func batchPercent(total, doneOps int, done bool, errMsg string) int {
	if total <= 0 {
		if done && errMsg == "" {
			return 100
		}
		return 0
	}
	if done && errMsg == "" {
		return 100
	}
	if doneOps < 0 {
		doneOps = 0
	}
	p := int(float64(doneOps) / float64(total) * 100)
	if p > 99 && !done {
		return 99
	}
	if p > 100 {
		return 100
	}
	return p
}

func batchETA(total, doneOps int, startedAt time.Time, running, done bool) (remaining, etaSeconds int) {
	if total <= 0 {
		return 0, 0
	}
	if doneOps < 0 {
		doneOps = 0
	}
	if doneOps > total {
		doneOps = total
	}
	remaining = total - doneOps
	if remaining < 0 {
		remaining = 0
	}
	if done || !running || doneOps <= 0 || startedAt.IsZero() {
		return remaining, 0
	}
	elapsed := time.Since(startedAt).Seconds()
	if elapsed <= 0 {
		return remaining, 0
	}
	left := float64(remaining) / (float64(doneOps) / elapsed)
	if left < 0 {
		left = 0
	}
	return remaining, int(left + 0.5)
}

func (s *coverBatchState) view(svc *Service) CoverBatchProgressView {
	s.mu.Lock()
	lines := append([]string(nil), s.lines...)
	running := s.running
	done := s.done
	errMsg := s.errMsg
	batchID := s.batchID
	courseCode := s.courseCode
	phase := s.phase
	total := s.total
	current := s.current
	pCompleted := s.pCompleted
	pSkipped := s.pSkipped
	pFailed := s.pFailed
	iCompleted := s.iCompleted
	iSkipped := s.iSkipped
	iFailed := s.iFailed
	currentTextID := s.currentTextID
	currentTitle := s.currentTitle
	startedAt := s.startedAt
	s.mu.Unlock()

	var cur CoverProgressView
	if currentTextID != "" {
		if v, ok := svc.CoverProgress(courseCode, currentTextID); ok {
			cur = v
		}
	}
	phaseDone1 := batchPhaseDone(pCompleted, pSkipped, pFailed)
	phaseDone2 := batchPhaseDone(iCompleted, iSkipped, iFailed)
	doneOps := phaseDone1 + phaseDone2
	percent := batchPercent(total, doneOps, done, errMsg)
	remaining, etaSeconds := batchETA(total, doneOps, startedAt, running, done)

	return CoverBatchProgressView{
		BatchID:         batchID,
		CourseCode:      courseCode,
		Phase:           phase,
		Running:         running,
		Done:            done,
		Error:           errMsg,
		Total:           total,
		Current:         current,
		Completed:       pCompleted + iCompleted,
		Skipped:         pSkipped + iSkipped,
		Failed:          pFailed + iFailed,
		Remaining:       remaining,
		ETASeconds:      etaSeconds,
		Percent:         percent,
		CurrentTextID:   currentTextID,
		CurrentTitle:    currentTitle,
		Log:             strings.Join(lines, "\n"),
		CurrentProgress: cur,
	}
}

func (s *Service) planCoverBatchTexts(courseCode, level string, limit int) ([]string, error) {
	course, err := s.paths.Course(courseCode)
	if err != nil {
		return nil, err
	}
	idx, err := loadReadingIndex(course.GrammarDir)
	if err != nil {
		return nil, err
	}
	textIDs := make([]string, 0, len(idx.Texts))
	for id := range idx.Texts {
		textIDs = append(textIDs, id)
	}
	sort.Strings(textIDs)
	if limit > 0 && len(textIDs) > limit {
		textIDs = textIDs[:limit]
	}
	level = strings.ToUpper(strings.TrimSpace(level))
	if level == "" {
		return textIDs, nil
	}
	filtered := make([]string, 0, len(textIDs))
	for _, textID := range textIDs {
		doc, err := readTextFile(course.GrammarDir, idx, textID)
		if err != nil {
			continue
		}
		if strings.ToUpper(doc.Level) != level {
			continue
		}
		filtered = append(filtered, textID)
	}
	return filtered, nil
}

type coverBatchPlan struct {
	course    Course
	idx       *readingIndex
	promptIDs []string
	imageIDs  []string
	totalOps  int
}

func (s *Service) planCoverBatchWork(req CoverBatchRequest) (*coverBatchPlan, error) {
	course, err := s.paths.Course(req.CourseCode)
	if err != nil {
		return nil, err
	}
	idx, err := loadReadingIndex(course.GrammarDir)
	if err != nil {
		return nil, err
	}
	textIDs, err := s.planCoverBatchTexts(course.Code, req.Level, req.Limit)
	if err != nil {
		return nil, err
	}
	plan := &coverBatchPlan{course: course, idx: idx}
	for _, textID := range textIDs {
		doc, err := readTextFile(course.GrammarDir, idx, textID)
		if err != nil {
			continue
		}
		ready := CoverStats(doc, course.GrammarDir) == CoverReady
		prompt := strings.TrimSpace(doc.CoverImagePrompt)
		if req.SkipPrompts {
			if req.Force {
				if prompt != "" {
					plan.imageIDs = append(plan.imageIDs, textID)
				}
				continue
			}
			if !ready && prompt != "" {
				plan.imageIDs = append(plan.imageIDs, textID)
			}
			continue
		}
		if req.Force {
			plan.promptIDs = append(plan.promptIDs, textID)
			plan.imageIDs = append(plan.imageIDs, textID)
			continue
		}
		if ready {
			continue
		}
		if prompt == "" {
			plan.promptIDs = append(plan.promptIDs, textID)
		}
		plan.imageIDs = append(plan.imageIDs, textID)
	}
	plan.totalOps = len(plan.promptIDs) + len(plan.imageIDs)
	return plan, nil
}

// StartCoverBatch plans texts and runs cover generation in the background.
func (s *Service) StartCoverBatch(req CoverBatchRequest) (batchID string, total int, err error) {
	courseCode := strings.ToLower(strings.TrimSpace(req.CourseCode))
	if courseCode == "" {
		return "", 0, errInvalid("course_code required")
	}
	req.CourseCode = courseCode
	plan, err := s.planCoverBatchWork(req)
	if err != nil {
		return "", 0, err
	}
	if plan.totalOps == 0 {
		return "", 0, errInvalid("no texts to process")
	}
	batchID = newBatchID(courseCode)
	st := &coverBatchState{
		batchID:     batchID,
		courseCode:  courseCode,
		phase:       coverBatchPhasePrompts,
		total:       plan.totalOps,
		skipPrompts: req.SkipPrompts,
		startedAt:   time.Now(),
		running:     true,
	}
	if req.SkipPrompts {
		st.phase = coverBatchPhaseImages
		st.appendLine(fmt.Sprintf("[batch] started: course=%s total=%d force=%v level=%s skip_prompts=true (images from saved prompts)",
			courseCode, plan.totalOps, req.Force, strings.TrimSpace(req.Level)))
	} else {
		st.appendLine(fmt.Sprintf("[batch] started: course=%s total=%d prompts=%d images=%d force=%v level=%s",
			courseCode, plan.totalOps, len(plan.promptIDs), len(plan.imageIDs), req.Force, strings.TrimSpace(req.Level)))
	}
	s.coverBatchJobs.Store(batchID, st)

	go s.runCoverBatch(context.Background(), st, req, plan)
	return batchID, plan.totalOps, nil
}

func stopLocalLLM(repoRoot string) error {
	script := filepath.Join(repoRoot, "scripts", "stop-local-llm.sh")
	if _, err := os.Stat(script); err != nil {
		return err
	}
	cmd := exec.Command("bash", script)
	cmd.Dir = repoRoot
	return cmd.Run()
}

func (s *Service) runCoverBatch(ctx context.Context, st *coverBatchState, req CoverBatchRequest, plan *coverBatchPlan) {
	var runErr error
	defer func() { st.finish(runErr) }()

	if plan == nil || plan.totalOps == 0 {
		runErr = errInvalid("no texts to process")
		st.appendLine("[batch] error: " + runErr.Error())
		return
	}
	course := plan.course
	idx := plan.idx

	sink := func(line string) { st.appendLine(line) }
	op := 0

	if req.SkipPrompts {
		st.appendLine("[batch] skipping prompt phase; using saved cover_image_prompt values")
	} else {
		// Phase 1: LLM prompts only.
		st.setPhase(coverBatchPhasePrompts)
		st.appendLine(fmt.Sprintf("[batch] phase 1/2: generating and saving cover prompts (LLM), total=%d", len(plan.promptIDs)))
		for i, textID := range plan.promptIDs {
			doc, err := readTextFile(course.GrammarDir, idx, textID)
			if err != nil {
				st.mu.Lock()
				st.pFailed++
				st.mu.Unlock()
				st.appendLine(fmt.Sprintf("[batch:prompts] !! %s: read failed: %v", textID, err))
				continue
			}
			title := strings.TrimSpace(doc.Title)
			if title == "" {
				title = textID
			}
			op++
			st.setCurrent(op, textID, title)
			st.appendHeader("prompts", i+1, len(plan.promptIDs), textID, title)

			jobLog, err := s.runCoverPromptScript(ctx, course.Code, course.GrammarDir, textID, req.Force, sink)
			if err != nil {
				st.mu.Lock()
				st.pFailed++
				st.mu.Unlock()
				st.appendLine(fmt.Sprintf("[batch:prompts] !! %s failed: %v", textID, err))
				if strings.TrimSpace(jobLog) != "" {
					st.appendLine(jobLog)
				}
				continue
			}
			st.mu.Lock()
			st.pCompleted++
			st.mu.Unlock()
			st.appendLine(fmt.Sprintf("[batch:prompts] -> ok %s", textID))
			updated, rerr := readTextFile(course.GrammarDir, idx, textID)
			if rerr == nil {
				_ = s.syncPublishedCoverToDraft(textID, updated, course.GrammarDir)
			}
			s.coverProgressUnregister(course.Code, textID)
		}

		st.setPhase(coverBatchPhaseStoppingLLM)
		st.setCurrent(0, "", "")
		st.appendLine("[batch] stopping local llama.cpp to free memory before ComfyUI phase…")
		if err := stopLocalLLM(s.paths.RepoRoot); err != nil {
			st.appendLine("[batch] warning: stop-local-llm: " + err.Error())
		} else {
			st.appendLine("[batch] llama.cpp stopped")
		}
	}

	// Phase 2: ComfyUI images from saved prompts.
	st.setPhase(coverBatchPhaseImages)
	if req.SkipPrompts {
		st.appendLine(fmt.Sprintf("[batch] phase 1/1: generating images from saved prompts (ComfyUI), total=%d", len(plan.imageIDs)))
	} else {
		st.appendLine(fmt.Sprintf("[batch] phase 2/2: generating images from saved prompts (ComfyUI), total=%d", len(plan.imageIDs)))
	}
	for i, textID := range plan.imageIDs {
		doc, err := readTextFile(course.GrammarDir, idx, textID)
		if err != nil {
			st.mu.Lock()
			st.iFailed++
			st.mu.Unlock()
			st.appendLine(fmt.Sprintf("[batch:images] !! %s: read failed: %v", textID, err))
			continue
		}
		title := strings.TrimSpace(doc.Title)
		if title == "" {
			title = textID
		}
		op++
		st.setCurrent(op, textID, title)
		st.appendHeader("images", i+1, len(plan.imageIDs), textID, title)

		prompt := strings.TrimSpace(doc.CoverImagePrompt)
		if prompt == "" {
			st.mu.Lock()
			st.iFailed++
			st.mu.Unlock()
			st.appendLine(fmt.Sprintf("[batch:images] !! %s: no cover_image_prompt", textID))
			continue
		}

		jobLog, err := s.runCoverScript(ctx, course.Code, course.GrammarDir, textID, true, prompt, sink)
		if err != nil {
			st.mu.Lock()
			st.iFailed++
			st.mu.Unlock()
			st.appendLine(fmt.Sprintf("[batch:images] !! %s failed: %v", textID, err))
			if strings.TrimSpace(jobLog) != "" {
				st.appendLine(jobLog)
			}
			continue
		}
		st.mu.Lock()
		st.iCompleted++
		st.mu.Unlock()
		st.appendLine(fmt.Sprintf("[batch:images] -> ok %s", textID))
		updated, rerr := readTextFile(course.GrammarDir, idx, textID)
		if rerr == nil {
			_ = s.syncPublishedCoverToDraft(textID, updated, course.GrammarDir)
		}
		s.coverProgressUnregister(course.Code, textID)
	}

	st.mu.Lock()
	pc, ps, pf := st.pCompleted, st.pSkipped, st.pFailed
	ic, isk, iff := st.iCompleted, st.iSkipped, st.iFailed
	st.mu.Unlock()
	st.appendLine(fmt.Sprintf("[batch] finished: prompts ok=%d skip=%d fail=%d; images ok=%d skip=%d fail=%d",
		pc, ps, pf, ic, isk, iff))
	if pf+iff > 0 {
		runErr = fmt.Errorf("%d cover step(s) failed", pf+iff)
	}
}

// CoverBatchProgress returns live batch job state.
func (s *Service) CoverBatchProgress(batchID string) (CoverBatchProgressView, bool) {
	v, ok := s.coverBatchJobs.Load(strings.TrimSpace(batchID))
	if !ok {
		return CoverBatchProgressView{}, false
	}
	return v.(*coverBatchState).view(s), true
}
