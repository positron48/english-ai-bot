package logger

import (
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		wantError bool
	}{
		{
			name:      "Debug level",
			level:     "debug",
			wantError: false,
		},
		{
			name:      "Info level",
			level:     "info",
			wantError: false,
		},
		{
			name:      "Warn level",
			level:     "warn",
			wantError: false,
		},
		{
			name:      "Error level",
			level:     "error",
			wantError: false,
		},
		{
			name:      "Invalid level (defaults to info)",
			level:     "invalid",
			wantError: false,
		},
		{
			name:      "Empty level (defaults to info)",
			level:     "",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.level)
			if (err != nil) != tt.wantError {
				t.Errorf("New() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if logger == nil && !tt.wantError {
				t.Error("New() returned nil logger")
			}
		})
	}
}
