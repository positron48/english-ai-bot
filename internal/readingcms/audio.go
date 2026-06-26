package readingcms

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GenerateAudio renders missing segment MP3 files into staging.
func (s *Service) GenerateAudio(ctx context.Context, textID string) (*DraftMeta, error) {
	meta, ok, err := s.store.GetMeta(textID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("draft not found: %s", textID)
	}
	doc, err := s.store.GetDocument(textID)
	if err != nil {
		return nil, err
	}
	stagingRoot := s.paths.StagingDir(textID)
	course, err := s.paths.Course(meta.CourseCode)
	if err != nil {
		return nil, err
	}
	segs := passageSegments(doc.ReadingPassage)
	var logLines []string
	for i, seg := range segs {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		text := strings.TrimSpace(str(seg["text"]))
		if text == "" {
			continue
		}
		voiceID := strDefault(seg["voice_id"], defaultVoiceID(course.TargetLang, str(seg["speaker_id"])))
		rel := str(seg["audio_rel_path"])
		if rel == "" {
			speakerID := str(seg["speaker_id"])
			rel = fmt.Sprintf("assets/reading/%s/seg_%02d_%s.mp3", textID, i+1, speakerID)
			seg["audio_rel_path"] = rel
		}
		outPath := filepath.Join(stagingRoot, filepath.FromSlash(rel))
		if st, err := os.Stat(outPath); err == nil && st.Size() > 0 {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return nil, err
		}
		script := filepath.Join(s.paths.RepoRoot, "scripts", "tts-reading-segment.sh")
		cmd := exec.CommandContext(ctx, "bash", script,
			"--voice-id", voiceID,
			"--text", text,
			"--output", outPath,
		)
		cmd.Env = s.paths.ScriptEnv(meta.CourseCode)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			logLines = append(logLines, fmt.Sprintf("seg %d: %s", i+1, strings.TrimSpace(stderr.String())))
		}
	}
	doc.ReadingPassage["segments"] = segs
	total, withAudio, audioSt := AudioStats(doc, stagingRoot)
	meta.SegmentsTotal = total
	meta.SegmentsWithAudio = withAudio
	meta.AudioStatus = audioSt
	if len(logLines) > 0 {
		meta.LastJobLog = strings.Join(logLines, "\n")
	} else {
		meta.LastJobLog = fmt.Sprintf("audio generated (%d/%d)", withAudio, total)
	}
	if audioSt == AudioReady && meta.Status == StatusApproved {
		meta.Status = StatusAudioReady
	}
	if err := s.store.SaveDraft(meta, doc); err != nil {
		return nil, err
	}
	return &meta, nil
}
