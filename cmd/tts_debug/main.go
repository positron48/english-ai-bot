package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/logger"
	"tgbot-skeleton/internal/service"
)

func main() {
	var timeout time.Duration
	flag.DurationVar(&timeout, "timeout", 45*time.Second, "request timeout per word")
	flag.Parse()
	words := flag.Args()
	if len(words) == 0 {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/tts_debug -- [word ...]")
		os.Exit(2)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}
	log, err := logger.New(cfg.Logging.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync() //nolint:errcheck

	tts := service.NewPronunciationService(cfg.TTS, nil, log)
	for _, word := range words {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		res, err := tts.DebugFetch(ctx, word)
		cancel()
		if err != nil {
			fmt.Fprintf(os.Stderr, "word=%q error=%v\n", word, err)
			continue
		}
		b, _ := json.Marshal(res)
		fmt.Println(string(b))
	}
}
