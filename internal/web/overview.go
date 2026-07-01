package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"

	"go.uber.org/zap"
)

// overviewPart describes one sub-endpoint to fold into an aggregated screen response: the JSON
// key it lands under, the existing handler that produces it, and an optional query override
// (e.g. the daily-route limit) applied to a cloned request.
type overviewPart struct {
	key     string
	handler http.HandlerFunc
	query   map[string]string
}

// runOverviewParts executes each part's existing handler in-process against its own recorder,
// in parallel, and assembles the successful (HTTP 200) results into a single JSON object keyed
// by part.key. This collapses a screen's fan-out of network requests into one round trip while
// reusing every existing handler verbatim (course-code resolution, grammar-service selection,
// error handling, and the shared Redis course-map cache all apply unchanged).
//
// A part that fails (non-200) is simply omitted from the response; the frontend already tolerates
// missing sections (each individual call had its own .catch()).
func (r *Router) runOverviewParts(w http.ResponseWriter, req *http.Request, parts []overviewPart) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	results := make([]json.RawMessage, len(parts))
	var wg sync.WaitGroup
	for i, p := range parts {
		wg.Add(1)
		go func(i int, p overviewPart) {
			defer wg.Done()
			// One misbehaving sub-handler must not take down the whole aggregate request; a
			// panicking part is simply omitted (same effect as a non-200 result).
			defer func() {
				if rec := recover(); rec != nil {
					r.logger.Error("overview part panicked", zap.String("part", p.key), zap.Any("recover", rec))
				}
			}()
			subReq := req
			if len(p.query) > 0 {
				subReq = req.Clone(req.Context())
				q := subReq.URL.Query()
				for k, v := range p.query {
					q.Set(k, v)
				}
				subReq.URL.RawQuery = q.Encode()
			}
			rec := httptest.NewRecorder()
			p.handler(rec, subReq)
			if rec.Code == http.StatusOK && rec.Body.Len() > 0 {
				results[i] = json.RawMessage(append([]byte(nil), rec.Body.Bytes()...))
			}
		}(i, p)
	}
	wg.Wait()

	out := make(map[string]json.RawMessage, len(parts))
	for i, p := range parts {
		if results[i] != nil {
			out[p.key] = results[i]
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(out); err != nil {
		r.logger.Error("failed to encode overview response", zap.Error(err))
	}
}

// handleDashboardOverview aggregates everything the dashboard screen needs into one response.
func (r *Router) handleDashboardOverview(w http.ResponseWriter, req *http.Request) {
	r.runOverviewParts(w, req, []overviewPart{
		{key: "dashboard", handler: r.handleDashboard},
		{key: "progress", handler: r.handleLinglowProgress},
		{key: "daily_route", handler: r.handleLinglowDailyRoute, query: map[string]string{"limit": "8"}},
		{key: "continue_chapter", handler: r.handleLearningGrammarContinueChapter},
		{key: "sentence_today", handler: r.handleSentenceTrainingToday},
	})
}

// handleCityOverview aggregates everything the city-map screen needs into one response.
func (r *Router) handleCityOverview(w http.ResponseWriter, req *http.Request) {
	r.runOverviewParts(w, req, []overviewPart{
		{key: "grammar_categories", handler: r.handleLearningGrammarCategories},
		{key: "progress", handler: r.handleLinglowProgress},
		{key: "course_map", handler: r.handleCourseMap},
		{key: "word_levels", handler: r.handleLinglowWordLevelProgress},
	})
}

// handleLearningOverview aggregates everything the practice screen needs into one response.
func (r *Router) handleLearningOverview(w http.ResponseWriter, req *http.Request) {
	r.runOverviewParts(w, req, []overviewPart{
		{key: "continue_chapter", handler: r.handleLearningGrammarContinueChapter},
		{key: "settings", handler: r.handleSettings},
		{key: "verb_upcoming", handler: r.handleVerbTrainingUpcoming},
		{key: "vocab_summary", handler: r.handleVocabSummary},
		{key: "sentence_today", handler: r.handleSentenceTrainingToday},
	})
}

// handleProgressOverview aggregates everything the progress screen needs into one response.
func (r *Router) handleProgressOverview(w http.ResponseWriter, req *http.Request) {
	r.runOverviewParts(w, req, []overviewPart{
		{key: "stats", handler: r.handleLinglowStats},
		{key: "progress", handler: r.handleLinglowProgress},
		{key: "dashboard", handler: r.handleDashboard},
		{key: "history", handler: r.handleLinglowHistory, query: map[string]string{"days": "7"}},
	})
}
