// One-off: generate TTS samples (OpenRouter/OpenAI /audio/speech) into tmp/ for listening.
// Run from repo root: go run ./cmd/tts_samples
// Requires .env with TTS_API_KEY, TTS_MODEL (e.g. tts-1 or gpt-4o-mini-tts), TTS_BASE_URL (e.g. https://api.openai.com/v1).
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	loadEnv(".env")
	key := strings.TrimSpace(os.Getenv("TTS_API_KEY"))
	model := strings.TrimSpace(os.Getenv("TTS_MODEL"))
	baseURL := strings.TrimSpace(os.Getenv("TTS_BASE_URL"))
	voice := strings.TrimSpace(os.Getenv("TTS_VOICE"))
	if key == "" || model == "" {
		return fmt.Errorf("set TTS_API_KEY and TTS_MODEL in .env")
	}
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if voice == "" {
		voice = "alloy"
	}

	outDir := "tmp"
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	words := []string{"hello", "lettuce", "world"}
	client := &http.Client{Timeout: 30 * time.Second}

	for _, word := range words {
		mp3, err := fetchSpeech(client, baseURL, model, voice, key, word)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %q: %v\n", word, err)
			continue
		}
		path := filepath.Join(outDir, "openrouter_"+word+".mp3")
		if err := os.WriteFile(path, mp3, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		fmt.Printf("wrote %s\n", path)
	}
	return nil
}

func fetchSpeech(client *http.Client, baseURL, model, voice, apiKey, word string) ([]byte, error) {
	body := map[string]interface{}{
		"input":           word,
		"model":           model,
		"voice":           voice,
		"response_format": "mp3",
	}
	bodyBytes, _ := json.Marshal(body)
	endpoint := strings.TrimRight(baseURL, "/") + "/audio/speech"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(b))
	}
	return io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
}

func loadEnv(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		i := strings.Index(line, "=")
		if i <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:i])
		if strings.HasPrefix(key, "TTS_") {
			val := strings.TrimSpace(line[i+1:])
			val = strings.Trim(val, "\"'")
			_ = os.Setenv(key, val)
		}
	}
}
