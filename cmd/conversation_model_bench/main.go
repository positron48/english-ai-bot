// Command conversation_model_bench compares LLMs on the Linglow NPC conversation task.
// It replays a scripted Spanish A0 "order coffee" quest against each candidate model via an
// OpenRouter-compatible endpoint and scores: control-JSON validity, INCREMENTAL task marking
// (the key weakness of gpt-4o-mini), error corrections, and Spanish-only replies — plus token
// cost pulled live from OpenRouter pricing.
//
// Usage:
//
//	AI_URL=https://openrouter.ai/api/v1 AI_API_KEY=sk-or-... \
//	  go run ./cmd/conversation_model_bench -models "openai/gpt-4o-mini,google/gemini-2.0-flash-001"
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
	"unicode"

	"go.uber.org/zap"

	"tgbot-skeleton/internal/ai"
)

type task struct {
	code, criteria string
	required       bool
}

// scriptedTurn is one learner message plus what we expect the model to do that turn.
type scriptedTurn struct {
	user            string
	expectTaskByNow string // task code that should be marked complete by this turn at the latest
	expectCorrection bool  // the message contains a deliberate Spanish mistake
}

func main() {
	modelsCSV := flag.String("models", "", "comma-separated model ids (default: built-in candidate set)")
	flag.Parse()

	url := os.Getenv("AI_URL")
	key := os.Getenv("AI_API_KEY")
	if url == "" || key == "" {
		fmt.Fprintln(os.Stderr, "set AI_URL and AI_API_KEY (OpenRouter)")
		os.Exit(1)
	}

	models := defaultModels
	if strings.TrimSpace(*modelsCSV) != "" {
		models = nil
		for _, m := range strings.Split(*modelsCSV, ",") {
			if m = strings.TrimSpace(m); m != "" {
				models = append(models, m)
			}
		}
	}

	prices := fetchPricing(url, key) // id -> (promptUSDPerTok, completionUSDPerTok)

	basePrompt, err := os.ReadFile("prompts/conversation-ru-es.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read es prompt (run from repo root): %v\n", err)
		os.Exit(1)
	}

	tasks := []task{
		{"greet", "The user greets the barista in the target language (any common greeting).", true},
		{"order", "The user orders a coffee with milk (e.g. cafe con leche / coffee with milk).", true},
		{"sugar", "The user asks for two sugars (e.g. dos azucares / two sugars) for the coffee.", true},
		{"thank", "The user thanks the barista and/or says goodbye.", false},
	}
	script := []scriptedTurn{
		{user: "Hola, buenas", expectTaskByNow: "greet"},
		{user: "quiero un cafe con leche", expectTaskByNow: "order", expectCorrection: true}, // cafe -> café
		{user: "y dos azucar por favor", expectTaskByNow: "sugar", expectCorrection: true},   // azucar -> azúcares
		{user: "muchas gracias, adios", expectTaskByNow: "thank"},
	}

	logger := zap.NewNop()
	var results []scorecard
	for _, model := range models {
		fmt.Fprintf(os.Stderr, "▶ %s …\n", model)
		svc := ai.NewServiceWithTimeout(url, model, key, "", 90*time.Second, logger)
		sc := runOne(svc, model, string(basePrompt), tasks, script)
		if p, ok := prices[model]; ok {
			sc.costUSD = float64(sc.promptTokens)*p[0] + float64(sc.completionTokens)*p[1]
			sc.hasPrice = true
		}
		results = append(results, sc)
	}

	printTable(results)
}

func runOne(svc *ai.Service, model, base string, tasks []task, script []scriptedTurn) scorecard {
	sc := scorecard{model: model}
	ctx := context.Background()
	completed := map[string]bool{}
	var history []ai.Message

	for i, st := range script {
		sys := buildSystemPrompt(base, tasks, completed)
		res, err := svc.ConversationTurn(ctx, sys, history, st.user, 500, model)
		if err != nil {
			sc.errors++
			sc.notes = append(sc.notes, fmt.Sprintf("turn %d error: %v", i+1, truncate(err.Error(), 80)))
			continue
		}
		sc.promptTokens += res.PromptTokens
		sc.completionTokens += res.CompletionTokens

		hadControl := strings.Contains(res.Raw, ai.ControlSentinel)
		if hadControl {
			sc.controlOK++
		}
		if res.VisibleContent == "" {
			sc.notes = append(sc.notes, fmt.Sprintf("turn %d: empty reply", i+1))
		}
		if leaksControl(res.VisibleContent) {
			sc.leaks++
		}
		if looksEnglish(res.VisibleContent) {
			sc.englishLeak++
		}

		for _, c := range res.CompletedTaskCodes {
			completed[strings.TrimSpace(c)] = true
		}
		// Incremental credit: the task expected by THIS turn must be marked done by now.
		if st.expectTaskByNow != "" {
			if completed[st.expectTaskByNow] {
				sc.incrementalHits++
			} else {
				sc.notes = append(sc.notes, fmt.Sprintf("turn %d: %q not marked yet", i+1, st.expectTaskByNow))
			}
		}
		if st.expectCorrection {
			sc.correctionExpected++
			if len(res.Corrections) > 0 {
				sc.correctionHits++
			}
		}

		history = append(history, ai.Message{Role: "user", Content: st.user})
		history = append(history, ai.Message{Role: "assistant", Content: res.VisibleContent})
	}
	sc.turns = len(script)
	return sc
}

// buildSystemPrompt mirrors web.buildConversationSystemPrompt for the cafe_order_coffee scenario.
func buildSystemPrompt(base string, tasks []task, completed map[string]bool) string {
	var b strings.Builder
	b.WriteString(strings.ReplaceAll(base, "\\n", "\n"))
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "SCENE\n- You are %s, %s.\n- Setting: %s\n- Target language code: %s. CEFR level: %s.\n",
		"Mara", "a warm, patient cafe barista who keeps sentences short and simple",
		"The learner walks into a cozy neighborhood cafe. You are behind the counter, ready to take their order.",
		"es", "A0")
	b.WriteString("\nTASKS the learner should accomplish (do NOT read this list aloud):\n")
	for _, t := range tasks {
		req := "required"
		if !t.required {
			req = "optional"
		}
		status := "not done yet"
		if completed[t.code] {
			status = "ALREADY DONE"
		}
		fmt.Fprintf(&b, "- [%s] (%s, %s) %s\n", t.code, req, status, t.criteria)
	}
	b.WriteString("\nUse the task codes in square brackets in completed_task_codes. Only mark a task when the learner has genuinely done it in their own words. Do not re-mark tasks already done.\n")
	b.WriteString("IMPORTANT: mark each task in completed_task_codes on the SAME turn the learner accomplishes it — as soon as it happens. Never wait until the end of the conversation to report several tasks at once. Each reply must report any tasks newly satisfied by the learner's latest message.\n")
	return b.String()
}

type scorecard struct {
	model              string
	turns              int
	controlOK          int
	incrementalHits    int
	correctionExpected int
	correctionHits     int
	leaks              int
	englishLeak        int
	errors             int
	promptTokens       int
	completionTokens   int
	costUSD            float64
	hasPrice           bool
	notes              []string
}

func printTable(rows []scorecard) {
	sort.Slice(rows, func(i, j int) bool { return score(rows[i]) > score(rows[j]) })
	fmt.Println()
	fmt.Printf("%-42s %5s %7s %7s %6s %6s %7s %9s %s\n",
		"model", "ctrl", "tasks", "corr", "leak", "EN", "err", "cost/4t", "score")
	fmt.Println(strings.Repeat("-", 110))
	for _, r := range rows {
		cost := "n/a"
		if r.hasPrice {
			cost = fmt.Sprintf("$%.5f", r.costUSD)
		}
		fmt.Printf("%-42s %2d/%-2d %5d/4 %4d/%-2d %6d %6d %7d %9s %5.1f\n",
			r.model, r.controlOK, r.turns, r.incrementalHits, r.correctionHits, r.correctionExpected,
			r.leaks, r.englishLeak, r.errors, cost, score(r))
	}
	fmt.Println("\nLegend: ctrl=valid control blocks, tasks=incremental task hits/4, corr=corrections found/expected,")
	fmt.Println("leak=visible reply leaked control JSON, EN=replies that looked English, cost/4t=USD for this 4-turn run.")
	for _, r := range rows {
		if len(r.notes) > 0 {
			fmt.Printf("\n%s:\n", r.model)
			for _, n := range r.notes {
				fmt.Printf("  - %s\n", n)
			}
		}
	}
}

// score is a quality heuristic in [0,10]: control validity, incremental tasks, corrections,
// minus penalties for leaking control JSON or replying in English.
func score(r scorecard) float64 {
	if r.turns == 0 {
		return 0
	}
	ctrl := float64(r.controlOK) / float64(r.turns)
	tasks := float64(r.incrementalHits) / float64(r.turns)
	corr := 0.0
	if r.correctionExpected > 0 {
		corr = float64(r.correctionHits) / float64(r.correctionExpected)
	}
	s := 4*tasks + 3*ctrl + 3*corr
	s -= float64(r.leaks) * 1.0
	s -= float64(r.englishLeak) * 1.5
	s -= float64(r.errors) * 2.0
	if s < 0 {
		s = 0
	}
	return s
}

func leaksControl(s string) bool { return strings.Contains(s, ai.ControlSentinel) || strings.Contains(s, "completed_task_codes") }

func looksEnglish(s string) bool {
	s = strings.ToLower(s)
	for _, w := range []string{" the ", " you ", " coffee", " please", " thank you", " milk", " sugar"} {
		if strings.Contains(s, w) {
			return true
		}
	}
	return false
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// fetchPricing pulls per-token prices from the OpenRouter /models endpoint.
func fetchPricing(baseURL, key string) map[string][2]float64 {
	out := map[string][2]float64{}
	req, _ := http.NewRequest("GET", strings.TrimRight(baseURL, "/")+"/models", nil)
	req.Header.Set("Authorization", "Bearer "+key)
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return out
	}
	defer resp.Body.Close()
	var body struct {
		Data []struct {
			ID      string `json:"id"`
			Pricing struct {
				Prompt     string `json:"prompt"`
				Completion string `json:"completion"`
			} `json:"pricing"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return out
	}
	for _, m := range body.Data {
		out[m.ID] = [2]float64{parseFloat(m.Pricing.Prompt), parseFloat(m.Pricing.Completion)}
	}
	return out
}

func parseFloat(s string) float64 {
	var f float64
	_, _ = fmt.Sscanf(strings.TrimSpace(s), "%g", &f)
	return f
}

var _ = unicode.IsLetter

var defaultModels = []string{
	"openai/gpt-4o-mini",
	"openai/gpt-4.1-mini",
	"openai/gpt-4.1-nano",
	"google/gemini-2.0-flash-001",
	"google/gemini-2.5-flash",
	"google/gemini-2.5-flash-lite",
	"anthropic/claude-3.5-haiku",
	"meta-llama/llama-3.3-70b-instruct",
	"qwen/qwen-2.5-72b-instruct",
	"mistralai/mistral-small-3.2-24b-instruct",
	"deepseek/deepseek-chat-v3.1",
}
