package web

import (
	"testing"
	"time"
)

// Note: parseDateTime is not exported, so we test it indirectly through handleVocab
// This test file documents the expected behavior

func TestParseDateTime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		checkTime func(*testing.T, *time.Time)
	}{
		{
			name:    "Valid time string",
			input:   "2006-01-02 15:04:05",
			wantErr: false,
			checkTime: func(t *testing.T, tm *time.Time) {
				if tm == nil {
					t.Error("Expected non-nil time")
					return
				}
				if tm.Year() != 2006 {
					t.Errorf("Expected year 2006, got %d", tm.Year())
				}
			},
		},
		{
			name:    "Empty string",
			input:   "",
			wantErr: false,
			checkTime: func(t *testing.T, tm *time.Time) {
				if tm != nil {
					t.Error("Expected nil time for empty string")
				}
			},
		},
		{
			name:    "Invalid time string",
			input:   "invalid",
			wantErr: true,
			checkTime: func(t *testing.T, tm *time.Time) {
				// Should be nil on error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDateTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDateTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkTime != nil {
				tt.checkTime(t, result)
			}
		})
	}
}
