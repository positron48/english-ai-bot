package service

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type speakingAudioConverter func(ctx context.Context, audio []byte, format string) ([]byte, string, error)

func prepareSpeakingAudioForModel(ctx context.Context, audio []byte, format string, convert speakingAudioConverter) ([]byte, string, error) {
	format = normalizeAudioFormat(format)
	switch format {
	case "mp3", "wav":
		return audio, format, nil
	}
	if convert == nil {
		convert = defaultSpeakingAudioConverter
	}
	return convert(ctx, audio, format)
}

func defaultSpeakingAudioConverter(ctx context.Context, audio []byte, _ string) ([]byte, string, error) {
	mp3, err := convertSpeakingAudioToMP3(ctx, audio)
	if err != nil {
		return nil, "", err
	}
	return mp3, "mp3", nil
}

// convertSpeakingAudioToMP3 transcodes browser recordings (webm/ogg/…) to mp3 for OpenRouter audio input.
func convertSpeakingAudioToMP3(ctx context.Context, audio []byte) ([]byte, error) {
	if len(audio) == 0 {
		return nil, fmt.Errorf("empty audio")
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-loglevel", "error",
		"-i", "pipe:0",
		"-ac", "1",
		"-ar", "16000",
		"-f", "mp3",
		"pipe:1")
	cmd.Stdin = bytes.NewReader(audio)
	var out bytes.Buffer
	cmd.Stdout = &out
	var errOut bytes.Buffer
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg webm to mp3: %w (stderr: %s)", err, strings.TrimSpace(errOut.String()))
	}
	if out.Len() == 0 {
		return nil, fmt.Errorf("ffmpeg webm to mp3: empty output")
	}
	return out.Bytes(), nil
}
