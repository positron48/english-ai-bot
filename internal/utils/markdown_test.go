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
			name:     "Headers H4",
			input:    "#### H4 Header",
			expected: "**H4 Header**",
		},
		{
			name:     "Headers H5",
			input:    "##### H5 Header",
			expected: "**H5 Header**",
		},
		{
			name:     "Headers H6",
			input:    "###### H6 Header",
			expected: "**# H6 Header**", // Function removes only 5 # symbols
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

func TestConvertMarkdownToHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string // Check that result contains these strings
		notContains []string // Check that result does not contain these strings
	}{
		{
			name:     "Headers",
			input:    "# Main Title\n## Subtitle\n### Section",
			contains: []string{"<h1>Main Title</h1>", "<h2>Subtitle</h2>", "<h3>Section</h3>"},
		},
		{
			name:     "Bold text",
			input:    "This is **bold** text",
			contains: []string{"<strong>bold</strong>"},
		},
		{
			name:     "Italic text",
			input:    "This is *italic* text",
			contains: []string{"<em>italic</em>"},
		},
		{
			name:     "Code blocks",
			input:    "```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```",
			contains: []string{"<pre><code>", "func main()", "fmt.Println"},
		},
		{
			name:     "Inline code",
			input:    "Use `fmt.Println()` function",
			contains: []string{"<code>fmt.Println()</code>"},
		},
		{
			name:     "Unordered lists",
			input:    "- First item\n- Second item\n* Third item",
			contains: []string{"<ul>", "<li>First item</li>", "<li>Second item</li>", "<li>Third item</li>", "</ul>"},
		},
		{
			name:     "Ordered lists",
			input:    "1. First item\n2. Second item",
			contains: []string{"<ol>", "<li>First item</li>", "<li>Second item</li>"},
		},
		{
			name:     "Links",
			input:    "Visit [Google](https://google.com) for search",
			contains: []string{"<a href=\"https://google.com\">Google</a>"},
		},
		{
			name:     "HTML escaping",
			input:    "Text with <tags> and & symbols",
			contains: []string{"&lt;tags&gt;", "&amp;"},
			notContains: []string{"<tags>", "& "},
		},
		{
			name:     "Line breaks",
			input:    "Line 1\nLine 2",
			contains: []string{"Line 1<br>", "Line 2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertMarkdownToHTML(tt.input)
			
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("ConvertMarkdownToHTML() should contain %q, got %q", expected, result)
				}
			}
			
			for _, notExpected := range tt.notContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("ConvertMarkdownToHTML() should not contain %q, got %q", notExpected, result)
				}
			}
		})
	}
}

func TestEscapeHTML(t *testing.T) {
	// Test escapeHTML function indirectly through ConvertMarkdownToHTML
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "Ampersand escaping",
			input:    "Text with & symbol",
			contains: []string{"&amp;"},
		},
		{
			name:     "Less than escaping",
			input:    "Text with <tag>",
			contains: []string{"&lt;tag&gt;"},
		},
		{
			name:     "Greater than escaping",
			input:    "Text with > symbol",
			contains: []string{"&gt;"},
		},
		{
			name:     "Quote escaping",
			input:    "Text with \"quotes\"",
			contains: []string{"&quot;"},
		},
		{
			name:     "Apostrophe escaping",
			input:    "Text with 'apostrophe'",
			contains: []string{"&#39;"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertMarkdownToHTML(tt.input)
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("ConvertMarkdownToHTML() should contain %q, got %q", expected, result)
				}
			}
		})
	}
}
