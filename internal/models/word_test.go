package models

import (
	"encoding/json"
	"testing"
)

func TestErrorField_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ErrorField
	}{
		{
			name:     "boolean true",
			input:    "true",
			expected: ErrorField{IsError: true, Message: "true"},
		},
		{
			name:     "boolean false",
			input:    "false",
			expected: ErrorField{IsError: false, Message: ""},
		},
		{
			name:     "string true",
			input:    `"true"`,
			expected: ErrorField{IsError: true, Message: "true"},
		},
		{
			name:     "string false",
			input:    `"false"`,
			expected: ErrorField{IsError: false, Message: "false"},
		},
		{
			name:     "string error message",
			input:    `"some error message"`,
			expected: ErrorField{IsError: true, Message: "some error message"},
		},
		{
			name:     "empty string",
			input:    `""`,
			expected: ErrorField{IsError: false, Message: ""},
		},
		{
			name:     "string null",
			input:    `"null"`,
			expected: ErrorField{IsError: false, Message: "null"},
		},
		{
			name:     "invalid JSON",
			input:    `{invalid}`,
			expected: ErrorField{IsError: false, Message: ""},
		},
		{
			name:     "number",
			input:    `123`,
			expected: ErrorField{IsError: false, Message: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e ErrorField
			err := json.Unmarshal([]byte(tt.input), &e)
			if err != nil && tt.name != "invalid JSON" {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}

			if e.IsError != tt.expected.IsError {
				t.Errorf("IsError = %v, want %v", e.IsError, tt.expected.IsError)
			}
			if e.Message != tt.expected.Message {
				t.Errorf("Message = %q, want %q", e.Message, tt.expected.Message)
			}
		})
	}
}

func TestErrorField_IsTrue(t *testing.T) {
	tests := []struct {
		name     string
		field    ErrorField
		expected bool
	}{
		{
			name:     "IsError true",
			field:    ErrorField{IsError: true, Message: ""},
			expected: true,
		},
		{
			name:     "IsError false, message true",
			field:    ErrorField{IsError: false, Message: "true"},
			expected: true,
		},
		{
			name:     "IsError false, message TRUE",
			field:    ErrorField{IsError: false, Message: "TRUE"},
			expected: true,
		},
		{
			name:     "IsError false, message True",
			field:    ErrorField{IsError: false, Message: "True"},
			expected: true,
		},
		{
			name:     "IsError false, message with spaces",
			field:    ErrorField{IsError: false, Message: "  true  "},
			expected: true,
		},
		{
			name:     "IsError false, message false",
			field:    ErrorField{IsError: false, Message: "false"},
			expected: false,
		},
		{
			name:     "IsError false, empty message",
			field:    ErrorField{IsError: false, Message: ""},
			expected: false,
		},
		{
			name:     "IsError false, other message",
			field:    ErrorField{IsError: false, Message: "some error"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.field.IsTrue()
			if result != tt.expected {
				t.Errorf("IsTrue() = %v, want %v", result, tt.expected)
			}
		})
	}
}
