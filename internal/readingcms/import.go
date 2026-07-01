package readingcms

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ImportPlainText runs the reading generator with --input-text so LLM structures
// the pasted source, optionally TTS, then auto-approves to audio_ready.
func (s *Service) ImportPlainText(ctx context.Context, req ImportTextRequest) (*DraftMeta, *TextDocument, error) {
	level, err := normalizeLevel(req.Level)
	if err != nil {
		return nil, nil, err
	}
	course, err := s.paths.Course(req.CourseCode)
	if err != nil {
		return nil, nil, err
	}
	format := normalizeFormat(req.Format)
	text := strings.TrimSpace(req.Text)
	if text == "" {
		return nil, nil, fmt.Errorf("text is required")
	}
	withAudio := withAudioDefault(req.WithAudio)

	meta, err := s.runReadingScript(ctx, scriptRunOptions{
		course:    course,
		level:     level,
		format:    format,
		title:     strings.TrimSpace(req.Title),
		withAudio: withAudio,
		inputText: text,
		origin:    OriginManualText,
		jobLog:    "structured from plain text via LLM",
	})
	if err != nil {
		return nil, nil, err
	}

	return s.finalizeImportedDraft(meta, importFinalizeOptions{
		withAudio:   withAudio,
		autoPublish: req.AutoPublish,
		syncBundle:  req.SyncBundle,
		readyLog:    "structured + audio ready",
		partialLog:  "structured via LLM; audio incomplete — retry Generate audio",
		noAudioLog:  "structured via LLM; run Generate audio before publish",
	})
}

// ImportJSON ingests ready-made JSON via generate-reading-text.py --input-json (no LLM),
// then TTS and optional auto-publish.
func (s *Service) ImportJSON(ctx context.Context, req ImportJSONRequest) (*DraftMeta, *TextDocument, error) {
	if len(req.Document) == 0 {
		return nil, nil, fmt.Errorf("document is required")
	}
	course, err := s.paths.Course(req.CourseCode)
	if err != nil {
		return nil, nil, err
	}
	level, format, title, err := levelFormatTitleFromJSON(req.Document, req.Level, req.Format, req.Title)
	if err != nil {
		return nil, nil, err
	}
	withAudio := withAudioDefault(req.WithAudio)

	meta, err := s.runReadingScript(ctx, scriptRunOptions{
		course:    course,
		level:     level,
		format:    format,
		title:     title,
		withAudio: withAudio,
		inputJSON: append([]byte(nil), req.Document...),
		origin:    OriginInputJSON,
		jobLog:    "imported from JSON (no LLM)",
	})
	if err != nil {
		return nil, nil, err
	}

	return s.finalizeImportedDraft(meta, importFinalizeOptions{
		withAudio:   withAudio,
		autoPublish: req.AutoPublish,
		syncBundle:  req.SyncBundle,
		readyLog:    "JSON imported + audio ready",
		partialLog:  "JSON imported; audio incomplete — retry Generate audio",
		noAudioLog:  "JSON imported; run Generate audio before publish",
	})
}

func (s *Service) ImportJSONBatch(ctx context.Context, req ImportJSONBatchRequest) (*ImportJSONBatchResponse, error) {
	docs, parseErr := parseImportJSONDocuments(req.DocumentsText)
	if len(docs) == 0 {
		if parseErr != nil {
			return nil, parseErr
		}
		return nil, fmt.Errorf("no JSON documents found")
	}

	resp := &ImportJSONBatchResponse{
		Results: make([]ImportJSONBatchResult, 0, len(docs)+1),
	}
	for i, doc := range docs {
		meta, textDoc, err := s.ImportJSON(ctx, ImportJSONRequest{
			CourseCode:  req.CourseCode,
			Level:       req.Level,
			Format:      req.Format,
			Title:       req.Title,
			Document:    doc,
			WithAudio:   req.WithAudio,
			AutoPublish: req.AutoPublish,
			SyncBundle:  req.SyncBundle,
		})
		item := ImportJSONBatchResult{Index: i}
		if err != nil {
			item.Error = err.Error()
			resp.Failed++
		} else {
			item.Draft = meta
			item.Document = textDoc
			resp.Succeeded++
		}
		resp.Results = append(resp.Results, item)
	}
	if parseErr != nil {
		resp.Results = append(resp.Results, ImportJSONBatchResult{
			Index: len(docs),
			Error: parseErr.Error(),
		})
		resp.Failed++
	}
	resp.Total = resp.Succeeded + resp.Failed
	return resp, nil
}

func parseImportJSONDocuments(input string) ([]json.RawMessage, error) {
	text := strings.TrimSpace(input)
	if text == "" {
		return nil, fmt.Errorf("documents_text is required")
	}

	dec := json.NewDecoder(strings.NewReader(text))
	var docs []json.RawMessage
	for {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			if errors.Is(err, io.EOF) {
				return docs, nil
			}
			if len(docs) == 0 {
				return nil, err
			}
			return docs, err
		}
		raw = bytesTrimSpace(raw)
		if len(raw) == 0 {
			continue
		}
		if raw[0] == '[' {
			var arr []json.RawMessage
			if err := json.Unmarshal(raw, &arr); err != nil {
				if len(docs) == 0 {
					return nil, err
				}
				return docs, err
			}
			for _, item := range arr {
				item = bytesTrimSpace(item)
				if len(item) > 0 {
					docs = append(docs, append(json.RawMessage(nil), item...))
				}
			}
			continue
		}
		docs = append(docs, append(json.RawMessage(nil), raw...))
	}
}

func bytesTrimSpace(b []byte) []byte {
	return bytes.TrimSpace(b)
}

func hasNarrator(generated map[string]interface{}) bool {
	segs, _ := generated["segments"].([]interface{})
	for _, raw := range segs {
		seg, _ := raw.(map[string]interface{})
		if str(seg["speaker_id"]) == "narrator" {
			return true
		}
	}
	return false
}

func str(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmt.Sprintf("%v", t)
	}
}

func strDefault(v interface{}, def string) string {
	s := strings.TrimSpace(str(v))
	if s == "" {
		return def
	}
	return s
}

func defaultVoiceID(targetLang, speakerID string) string {
	if speakerID == "narrator" {
		return targetLang + "_narrator"
	}
	if speakerID == "speaker_b" {
		return targetLang + "_male_1"
	}
	return targetLang + "_female_1"
}
