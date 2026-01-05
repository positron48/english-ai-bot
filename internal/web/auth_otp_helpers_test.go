package web

import (
	"testing"
)

func TestNormalizeOTPCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal code",
			input:    "123456",
			expected: "123456",
		},
		{
			name:     "Code with spaces",
			input:    "123 456",
			expected: "123456",
		},
		{
			name:     "Code with dashes",
			input:    "123-456",
			expected: "123456",
		},
		{
			name:     "Code with mixed separators",
			input:    "123 45-6",
			expected: "123456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeOTPCode(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMaskCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "6 digit code",
			input:    "123456",
			expected: "1****6",
		},
		{
			name:     "4 digit code",
			input:    "1234",
			expected: "1**4",
		},
		{
			name:     "Short code",
			input:    "12",
			expected: "**",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskCode(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
