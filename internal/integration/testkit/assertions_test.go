package testkit

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIntVal(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
		want int
	}{
		{"int", 42, 42},
		{"int64", int64(100), 100},
		{"float64", float64(7.9), 7},
		{"float64_whole", float64(12.0), 12},
		{"nil_other", "x", 0},
		{"bool", true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intVal(tt.v)
			if got != tt.want {
				t.Errorf("intVal(%v) = %d, want %d", tt.v, got, tt.want)
			}
		})
	}
}

func TestAssertDashboardCounters(t *testing.T) {
	tests := []struct {
		name         string
		code         int
		body         string
		wantDue      int
		wantNew      int
		wantLearning int
		wantReview   int
		shouldFail   bool
	}{
		{
			name:         "ok",
			code:         http.StatusOK,
			body:         `{"due_count":1,"new_count":2,"learning_count":3,"review_count":4}`,
			wantDue:      1,
			wantNew:      2,
			wantLearning: 3,
			wantReview:   4,
			shouldFail:   false,
		},
		{
			name:         "wrong_status",
			code:         http.StatusNotFound,
			body:         `{}`,
			wantDue:      0,
			wantNew:      0,
			wantLearning: 0,
			wantReview:   0,
			shouldFail:   true,
		},
		{
			name:         "zero_counts",
			code:         http.StatusOK,
			body:         `{"due_count":0,"new_count":0,"learning_count":0,"review_count":0}`,
			wantDue:      0,
			wantNew:      0,
			wantLearning: 0,
			wantReview:   0,
			shouldFail:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			w.WriteHeader(tt.code)
			_, _ = w.Write([]byte(tt.body))

			fakeT := &testing.T{}
			AssertDashboardCounters(fakeT, w, tt.wantDue, tt.wantNew, tt.wantLearning, tt.wantReview)
			failed := fakeT.Failed()
			if failed != tt.shouldFail {
				t.Errorf("AssertDashboardCounters: expected fail=%v, got fail=%v", tt.shouldFail, failed)
			}
		})
	}
}

func TestAssertDashboardCounters_ValidBody(t *testing.T) {
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"due_count":10,"new_count":20,"learning_count":30,"review_count":40}`))
	AssertDashboardCounters(t, w, 10, 20, 30, 40)
}
