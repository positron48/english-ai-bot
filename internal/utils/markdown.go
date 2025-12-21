package utils

import (
	"regexp"
	"strings"
)

// ConvertMarkdownToTelegram converts Markdown formatting to Telegram's format
func ConvertMarkdownToTelegram(text string) string {
	// Convert headers first
	text = convertHeaders(text)

	// Convert code blocks (before other formatting)
	text = convertCodeBlocks(text)

	// Convert lists (before italic to avoid conflicts)
	text = convertLists(text)

	// Convert bold text
	text = convertBold(text)

	// Convert italic text
	text = convertItalic(text)

	// Convert links (basic support)
	text = convertLinks(text)

	// Clean up extra whitespace
	text = strings.TrimSpace(text)

	return text
}

// convertHeaders converts Markdown headers to Telegram format
func convertHeaders(text string) string {
	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// H1: # Header -> **Header**
		if strings.HasPrefix(trimmed, "# ") {
			header := strings.TrimPrefix(trimmed, "# ")
			result = append(result, "**"+header+"**")
			continue
		}

		// H2: ## Header -> **Header**
		if strings.HasPrefix(trimmed, "## ") {
			header := strings.TrimPrefix(trimmed, "## ")
			result = append(result, "**"+header+"**")
			continue
		}

		// H3: ### Header -> **Header**
		if strings.HasPrefix(trimmed, "### ") {
			header := strings.TrimPrefix(trimmed, "### ")
			result = append(result, "**"+header+"**")
			continue
		}

		// H4-H6: #### Header -> **Header**
		if strings.HasPrefix(trimmed, "#### ") || strings.HasPrefix(trimmed, "##### ") || strings.HasPrefix(trimmed, "###### ") {
			header := strings.TrimPrefix(trimmed, "#")
			header = strings.TrimPrefix(header, "#")
			header = strings.TrimPrefix(header, "#")
			header = strings.TrimPrefix(header, "#")
			header = strings.TrimPrefix(header, "#")
			header = strings.TrimSpace(header)
			result = append(result, "**"+header+"**")
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// convertBold converts Markdown bold to Telegram format
func convertBold(text string) string {
	// Convert __text__ to **text**
	text = strings.ReplaceAll(text, "__", "**")
	return text
}

// convertItalic converts Markdown italic to Telegram format
func convertItalic(text string) string {
	// Convert *text* to _text_ (Telegram format)
	// Simple approach: find single asterisks that are not part of double asterisks
	// We'll use a simple string replacement approach

	// First, protect existing **bold** by replacing with a temporary marker
	text = strings.ReplaceAll(text, "**", "___BOLD_MARKER___")

	// Now convert single asterisks to underscores
	text = strings.ReplaceAll(text, "*", "_")

	// Restore bold markers
	text = strings.ReplaceAll(text, "___BOLD_MARKER___", "**")

	return text
}

// convertCodeBlocks converts Markdown code blocks to Telegram format
func convertCodeBlocks(text string) string {
	// Convert ```language\ncode\n``` to ```\ncode\n```
	// Use a more flexible regex that handles multiline content
	text = regexp.MustCompile("```\\w*\\n([\\s\\S]*?)\\n```").ReplaceAllString(text, "```\n$1\n```")
	return text
}

// convertLists converts Markdown lists to Telegram format
func convertLists(text string) string {
	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Convert unordered lists: - item -> • item
		if strings.HasPrefix(trimmed, "- ") {
			item := strings.TrimPrefix(trimmed, "- ")
			result = append(result, "• "+item)
			continue
		}

		// Convert unordered lists: * item -> • item
		if strings.HasPrefix(trimmed, "* ") {
			item := strings.TrimPrefix(trimmed, "* ")
			result = append(result, "• "+item)
			continue
		}

		// Convert ordered lists: 1. item -> 1. item (keep as is)
		if regexp.MustCompile(`^\d+\.\s`).MatchString(trimmed) {
			result = append(result, line)
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// convertLinks converts Markdown links to Telegram format
func convertLinks(text string) string {
	// Convert [text](url) to [text](url) (already Telegram format)
	return text
}

// EscapeTelegramMarkdown escapes special characters for Telegram
func EscapeTelegramMarkdown(text string) string {
	// Escape special characters that have meaning in Telegram Markdown
	text = strings.ReplaceAll(text, "_", "\\_")
	text = strings.ReplaceAll(text, "*", "\\*")
	text = strings.ReplaceAll(text, "[", "\\[")
	text = strings.ReplaceAll(text, "]", "\\]")
	text = strings.ReplaceAll(text, "(", "\\(")
	text = strings.ReplaceAll(text, ")", "\\)")
	text = strings.ReplaceAll(text, "~", "\\~")
	text = strings.ReplaceAll(text, "`", "\\`")
	text = strings.ReplaceAll(text, ">", "\\>")
	text = strings.ReplaceAll(text, "#", "\\#")
	text = strings.ReplaceAll(text, "+", "\\+")
	text = strings.ReplaceAll(text, "-", "\\-")
	text = strings.ReplaceAll(text, "=", "\\=")
	text = strings.ReplaceAll(text, "|", "\\|")
	text = strings.ReplaceAll(text, "{", "\\{")
	text = strings.ReplaceAll(text, "}", "\\}")
	text = strings.ReplaceAll(text, ".", "\\.")
	text = strings.ReplaceAll(text, "!", "\\!")

	return text
}

// ConvertMarkdownToHTML converts Markdown formatting to HTML
func ConvertMarkdownToHTML(text string) string {
	// Escape HTML first
	text = escapeHTML(text)
	
	// Convert headers
	text = convertHeadersToHTML(text)
	
	// Convert code blocks
	text = convertCodeBlocksToHTML(text)
	
	// Convert lists
	text = convertListsToHTML(text)
	
	// Convert bold
	text = convertBoldToHTML(text)
	
	// Convert italic
	text = convertItalicToHTML(text)
	
	// Convert links
	text = convertLinksToHTML(text)
	
	// Convert line breaks
	text = strings.ReplaceAll(text, "\n", "<br>")
	
	return text
}

// escapeHTML escapes HTML special characters
func escapeHTML(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	text = strings.ReplaceAll(text, "\"", "&quot;")
	text = strings.ReplaceAll(text, "'", "&#39;")
	return text
}

// convertHeadersToHTML converts Markdown headers to HTML
func convertHeadersToHTML(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if strings.HasPrefix(trimmed, "# ") {
			header := strings.TrimPrefix(trimmed, "# ")
			result = append(result, "<h1>"+header+"</h1>")
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			header := strings.TrimPrefix(trimmed, "## ")
			result = append(result, "<h2>"+header+"</h2>")
			continue
		}
		if strings.HasPrefix(trimmed, "### ") {
			header := strings.TrimPrefix(trimmed, "### ")
			result = append(result, "<h3>"+header+"</h3>")
			continue
		}
		
		result = append(result, line)
	}
	
	return strings.Join(result, "\n")
}

// convertBoldToHTML converts Markdown bold to HTML
func convertBoldToHTML(text string) string {
	// Convert **text** to <strong>text</strong>
	text = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(text, "<strong>$1</strong>")
	text = regexp.MustCompile(`__([^_]+)__`).ReplaceAllString(text, "<strong>$1</strong>")
	return text
}

// convertItalicToHTML converts Markdown italic to HTML
func convertItalicToHTML(text string) string {
	// Convert *text* to <em>text</em> (but not **text**)
	// First protect **bold**
	text = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(text, "___BOLD___$1___BOLD___")
	text = regexp.MustCompile(`\*([^*]+)\*`).ReplaceAllString(text, "<em>$1</em>")
	text = strings.ReplaceAll(text, "___BOLD___", "**")
	return text
}

// convertCodeBlocksToHTML converts Markdown code blocks to HTML
func convertCodeBlocksToHTML(text string) string {
	// Convert ```code``` to <pre><code>code</code></pre>
	text = regexp.MustCompile("```\\w*\\n([\\s\\S]*?)\\n```").ReplaceAllString(text, "<pre><code>$1</code></pre>")
	// Convert `code` to <code>code</code>
	text = regexp.MustCompile("`([^`]+)`").ReplaceAllString(text, "<code>$1</code>")
	return text
}

// convertListsToHTML converts Markdown lists to HTML
func convertListsToHTML(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	inList := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			if !inList {
				result = append(result, "<ul>")
				inList = true
			}
			item := strings.TrimPrefix(trimmed, "- ")
			item = strings.TrimPrefix(item, "* ")
			result = append(result, "<li>"+item+"</li>")
			continue
		}
		
		if regexp.MustCompile(`^\d+\.\s`).MatchString(trimmed) {
			if !inList {
				result = append(result, "<ol>")
				inList = true
			}
			item := regexp.MustCompile(`^\d+\.\s`).ReplaceAllString(trimmed, "")
			result = append(result, "<li>"+item+"</li>")
			continue
		}
		
		if inList {
			result = append(result, "</ul>")
			inList = false
		}
		
		result = append(result, line)
	}
	
	if inList {
		result = append(result, "</ul>")
	}
	
	return strings.Join(result, "\n")
}

// convertLinksToHTML converts Markdown links to HTML
func convertLinksToHTML(text string) string {
	// Convert [text](url) to <a href="url">text</a>
	text = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`).ReplaceAllString(text, `<a href="$2">$1</a>`)
	return text
}
