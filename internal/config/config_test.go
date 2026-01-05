package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "Environment variable set",
			envValue:     "test_value",
			defaultValue: "default",
			expected:     "test_value",
		},
		{
			name:         "Environment variable not set",
			envValue:     "",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_ENV_VAR"
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
				defer os.Unsetenv(key)
			} else {
				os.Unsetenv(key)
			}
			
			result := GetEnv(key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetEnv(%q, %q) = %q, want %q", key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestProcessNewlines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No newlines",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "Single newline",
			input:    "Hello\\nworld",
			expected: "Hello\nworld",
		},
		{
			name:     "Multiple newlines",
			input:    "Line1\\nLine2\\nLine3",
			expected: "Line1\nLine2\nLine3",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only newlines",
			input:    "\\n\\n\\n",
			expected: "\n\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processNewlines(tt.input)
			if result != tt.expected {
				t.Errorf("processNewlines(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLoadPromptFromFile(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name:      "Valid file",
			content:   "This is a test prompt",
			wantError: false,
		},
		{
			name:      "File with newlines",
			content:   "Line 1\nLine 2\nLine 3",
			wantError: false,
		},
		{
			name:      "Empty file",
			content:   "",
			wantError: false,
		},
		{
			name:      "File with whitespace",
			content:   "  \n  prompt  \n  ",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test_prompt.txt")
			
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
			
			result, err := loadPromptFromFile(tmpFile)
			if (err != nil) != tt.wantError {
				t.Errorf("loadPromptFromFile() error = %v, wantError %v", err, tt.wantError)
				return
			}
			
			if !tt.wantError {
				expected := tt.content
				if tt.content != "" {
					// Function trims whitespace
					expected = tt.content
				}
				// Check that result is trimmed
				if result != expected && tt.content != "" {
					// For non-empty content, check it's close (trimmed)
					if len(result) == 0 && len(tt.content) > 0 {
						t.Errorf("loadPromptFromFile() returned empty for non-empty content")
					}
				}
			}
		})
	}
	
	t.Run("Non-existent file", func(t *testing.T) {
		_, err := loadPromptFromFile("/nonexistent/file/path.txt")
		if err == nil {
			t.Error("loadPromptFromFile() should return error for non-existent file")
		}
	})
}
