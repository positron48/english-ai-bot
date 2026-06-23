package readingcms

import (
	"encoding/json"
	"strings"
)

type importFinalizeOptions struct {
	withAudio   bool
	autoPublish bool
	syncBundle  bool
	readyLog    string
	partialLog  string
	noAudioLog  string
}

func (s *Service) finalizeImportedDraft(meta *DraftMeta, opts importFinalizeOptions) (*DraftMeta, *TextDocument, error) {
	doc, err := s.store.GetDocument(meta.TextID)
	if err != nil {
		return nil, nil, err
	}

	stagingRoot := s.paths.StagingDir(meta.TextID)
	total, withAudioCount, audioSt := AudioStats(doc, stagingRoot)
	meta.SegmentsTotal = total
	meta.SegmentsWithAudio = withAudioCount
	meta.AudioStatus = audioSt
	switch {
	case audioSt == AudioReady:
		meta.Status = StatusAudioReady
		meta.LastJobLog = opts.readyLog
	case opts.withAudio:
		meta.Status = StatusApproved
		meta.LastJobLog = opts.partialLog
	default:
		meta.Status = StatusApproved
		meta.LastJobLog = opts.noAudioLog
	}
	if err := s.store.SaveDraft(*meta, doc); err != nil {
		return nil, nil, err
	}

	if opts.autoPublish && meta.Status == StatusAudioReady {
		published, err := s.Publish(meta.TextID, opts.syncBundle)
		if err != nil {
			return meta, doc, err
		}
		meta = published
		doc, err = s.store.GetDocument(meta.TextID)
		if err != nil {
			return meta, nil, err
		}
	}
	return meta, doc, nil
}

func withAudioDefault(v *bool) bool {
	if v == nil {
		return true
	}
	return *v
}

func levelFormatTitleFromJSON(raw []byte, reqLevel, reqFormat, reqTitle string) (level, format, title string, err error) {
	var generated map[string]interface{}
	if err := json.Unmarshal(raw, &generated); err != nil {
		return "", "", "", err
	}
	level = str(generated["level"])
	if strings.TrimSpace(level) == "" {
		level = reqLevel
	}
	level, err = normalizeLevel(level)
	if err != nil {
		level = "A2"
	}
	format = normalizeFormat(reqFormat)
	if hasNarrator(generated) {
		format = "narrative"
	}
	title = str(generated["title_short"])
	if strings.TrimSpace(title) == "" {
		title = str(generated["title"])
	}
	if strings.TrimSpace(title) == "" {
		title = strings.TrimSpace(reqTitle)
	}
	if title == "" {
		title = "Imported JSON"
	}
	return level, format, title, nil
}
