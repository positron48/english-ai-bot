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

func TestWordInfoResponse_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name          string
		json          string
		wantLemma     string
		wantError     bool
		wantErrorMsg  string
		wantVerbForms bool
		wantExamples  int
	}{
		{
			name:          "success response with verb forms",
			json:          `{"input_word":"spied","lemma":"spy","pos":"verb","transcription":"/spaɪ/","definition_ru":"шпионить","examples":[{"example_en":"He was spying on them.","gloss_ru":"Он за ними шпионил."}],"verb_forms":{"v1":"spy","v2":"spied","v3":"spied","gerund":"spying","third_person":"spies"}}`,
			wantLemma:     "spy",
			wantError:     false,
			wantVerbForms: true,
			wantExamples:  1,
		},
		{
			name:         "error response with hint",
			json:         `{"error":"Not an English word","hint":"Try a different word.","input_word":"xyz","lemma":"","pos":"","transcription":"","definition_ru":""}`,
			wantLemma:    "",
			wantError:    true,
			wantErrorMsg: "Not an English word",
		},
		{
			name:         "error as boolean true",
			json:         `{"error":true,"input_word":"x","lemma":"","pos":"","transcription":"","definition_ru":""}`,
			wantLemma:    "",
			wantError:    true,
			wantErrorMsg: "true",
		},
		{
			name:         "definition_native only (neutral wire)",
			json:         `{"input_word":"hola","lemma":"hola","pos":"interjection","transcription":"","definition_native":"привет","examples":[]}`,
			wantLemma:    "hola",
			wantError:    false,
			wantExamples: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp WordInfoResponse
			if err := json.Unmarshal([]byte(tt.json), &resp); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if resp.Lemma != tt.wantLemma {
				t.Errorf("Lemma = %q, want %q", resp.Lemma, tt.wantLemma)
			}
			if resp.Error.IsError != tt.wantError {
				t.Errorf("Error.IsError = %v, want %v", resp.Error.IsError, tt.wantError)
			}
			if tt.wantErrorMsg != "" && resp.Error.Message != tt.wantErrorMsg {
				t.Errorf("Error.Message = %q, want %q", resp.Error.Message, tt.wantErrorMsg)
			}
			if tt.wantVerbForms && resp.VerbForms == nil {
				t.Error("VerbForms = nil, want non-nil")
			}
			if tt.wantExamples > 0 && len(resp.Examples) != tt.wantExamples {
				t.Errorf("len(Examples) = %d, want %d", len(resp.Examples), tt.wantExamples)
			}
			if tt.name == "definition_native only (neutral wire)" && resp.DefinitionRU != "привет" {
				t.Errorf("DefinitionRU = %q, want %q", resp.DefinitionRU, "привет")
			}
		})
	}
}
