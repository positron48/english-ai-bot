package utils

import (
	"strings"
	"testing"

	"tgbot-skeleton/internal/models"
)

func TestConvertMarkdownToTelegram(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Headers",
			input:    "# Main Title\n## Subtitle\n### Section",
			expected: "**Main Title**\n**Subtitle**\n**Section**",
		},
		{
			name:     "Bold text with underscores",
			input:    "This is __bold__ text",
			expected: "This is **bold** text",
		},
		{
			name:     "Italic text",
			input:    "This is *italic* text",
			expected: "This is _italic_ text",
		},
		{
			name:     "Code blocks",
			input:    "```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```",
			expected: "```\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```",
		},
		{
			name:     "Inline code",
			input:    "Use `fmt.Println()` function",
			expected: "Use `fmt.Println()` function",
		},
		{
			name:     "Unordered lists",
			input:    "- First item\n- Second item\n* Third item",
			expected: "• First item\n• Second item\n• Third item",
		},
		{
			name:     "Ordered lists",
			input:    "1. First item\n2. Second item\n3. Third item",
			expected: "1. First item\n2. Second item\n3. Third item",
		},
		{
			name:     "Simple mixed formatting",
			input:    "# Title\n\nThis is __bold__ and *italic* text.\n\n- Item 1\n- Item 2",
			expected: "**Title**\n\nThis is **bold** and _italic_ text.\n\n• Item 1\n• Item 2",
		},
		{
			name:     "Links",
			input:    "Visit [Google](https://google.com) for search",
			expected: "Visit [Google](https://google.com) for search",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertMarkdownToTelegram(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertMarkdownToTelegram() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEscapeTelegramMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic escaping",
			input:    "This has _underscores_ and *asterisks*",
			expected: "This has \\_underscores\\_ and \\*asterisks\\*",
		},
		{
			name:     "Special characters",
			input:    "Text with [brackets] and (parentheses)",
			expected: "Text with \\[brackets\\] and \\(parentheses\\)",
		},
		{
			name:     "Code-like text",
			input:    "Use `backticks` and ```code blocks```",
			expected: "Use \\`backticks\\` and \\`\\`\\`code blocks\\`\\`\\`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeTelegramMarkdown(tt.input)
			if result != tt.expected {
				t.Errorf("EscapeTelegramMarkdown() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRenderWordCardMarkdown(t *testing.T) {
	pos := "noun"
	transcription := "/spaɪ/"
	definitionRU := "шпион"
	displayEN := "spy"

	card := &models.WordCard{
		Word:          "spy",
		Definition:    "",
		POS:           &pos,
		Transcription: &transcription,
		DefinitionRU: &definitionRU,
		DisplayEN:     &displayEN,
	}

	examples := []models.WordInfoExample{
		{
			ExampleEN: "He is a spy",
			GlossRU:   "Он шпион",
		},
		{
			ExampleEN: "The spy was caught",
			GlossRU:   "Шпиона поймали",
		},
	}

	result := RenderWordCardMarkdown(card, examples, nil)
	
	if !strings.Contains(result, "spy") {
		t.Error("Result should contain word 'spy'")
	}
	if !strings.Contains(result, "noun") {
		t.Error("Result should contain POS 'noun'")
	}
	if !strings.Contains(result, definitionRU) {
		t.Error("Result should contain Russian definition")
	}
	if !strings.Contains(result, "He is a spy") {
		t.Error("Result should contain example")
	}
}

func TestRenderWordCardMarkdown_WithVerbForms(t *testing.T) {
	pos := "verb"
	transcription := "/raɪt/"
	definitionRU := "писать"
	displayEN := "to write"

	card := &models.WordCard{
		Word:          "write",
		Definition:    "",
		POS:           &pos,
		Transcription: &transcription,
		DefinitionRU: &definitionRU,
		DisplayEN:     &displayEN,
	}

	examples := []models.WordInfoExample{
		{
			ExampleEN: "I write letters",
			GlossRU:   "Я пишу письма",
		},
	}

	verbForms := &models.WordInfoVerbForms{
		V1:         "write",
		V2:         "wrote",
		V3:         "written",
		Gerund:     "writing",
		ThirdPerson: "writes",
	}

	result := RenderWordCardMarkdown(card, examples, verbForms)
	
	if !strings.Contains(result, "to write") {
		t.Error("Result should contain display word 'to write'")
	}
	if !strings.Contains(result, "Verb Forms") {
		t.Error("Result should contain 'Verb Forms' section")
	}
	if !strings.Contains(result, "write") {
		t.Error("Result should contain V1")
	}
	if !strings.Contains(result, "wrote") {
		t.Error("Result should contain V2")
	}
	if !strings.Contains(result, "written") {
		t.Error("Result should contain V3")
	}
}
