package source

import (
	"fmt"
	"strings"
)

// ParseMarkdownSections parses markdown content into sections based on headers
func ParseMarkdownSections(content string) (map[string]string, error) {
	if content == "" {
		return make(map[string]string), nil
	}

	sections := make(map[string]string)
	lines := strings.Split(content, "\n")

	var currentSection string
	var currentContent []string

	for _, line := range lines {
		// Check for markdown header (# Header Name)
		if strings.HasPrefix(line, "# ") {
			// Save previous section if exists
			if currentSection != "" {
				content := strings.Join(currentContent, "\n")
				// Trim trailing whitespace but preserve leading whitespace
				sections[currentSection] = strings.TrimRight(content, " \t\n\r")
			}

			// Start new section
			currentSection = strings.TrimPrefix(line, "# ")
			currentSection = strings.TrimSpace(currentSection)
			currentContent = []string{}
		} else if currentSection != "" {
			// Add line to current section content
			currentContent = append(currentContent, line)
		}
		// Lines before any section header are ignored
	}

	// Save final section
	if currentSection != "" {
		content := strings.Join(currentContent, "\n")
		// Trim trailing whitespace but preserve leading whitespace
		sections[currentSection] = strings.TrimRight(content, " \t\n\r")
	}

	return sections, nil
}

// ParseMindfulMemory parses memory.mdc and returns only the MINDFUL section content
func ParseMindfulMemory(content string) string {
	if content == "" {
		return ""
	}

	sections, err := ParseMarkdownSections(content)
	if err != nil {
		return ""
	}

	mindfulContent, exists := sections["MINDFUL"]
	if !exists {
		return ""
	}

	return mindfulContent
}

// ReconstructMarkdown reconstructs markdown content from sections
func ReconstructMarkdown(sections map[string]string) string {
	if len(sections) == 0 {
		return ""
	}

	var parts []string
	for sectionName, content := range sections {
		if sectionName == "" || content == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("# %s\n%s", sectionName, content))
	}

	return strings.Join(parts, "\n\n")
}

// ExtractMetadata extracts metadata from subagent file content
func ExtractMetadata(content string) map[string]string {
	metadata := make(map[string]string)

	lines := strings.Split(content, "\n")

	// Look for metadata in the first few lines
	for i, line := range lines {
		if i > 10 { // Only check first 10 lines
			break
		}

		line = strings.TrimSpace(line)

		// Look for key: value patterns in comments
		if strings.HasPrefix(line, "<!--") && strings.HasSuffix(line, "-->") {
			comment := strings.TrimSpace(line[4 : len(line)-3])

			if strings.Contains(comment, ":") {
				parts := strings.SplitN(comment, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					metadata[key] = value
				}
			}
		}
	}

	return metadata
}

// SanitizeContent sanitizes content for safe processing
func SanitizeContent(content string) string {
	// Remove null bytes
	content = strings.ReplaceAll(content, "\x00", "")

	// Normalize line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	return content
}

// ValidateSubagentContent validates subagent file content
func ValidateSubagentContent(content string, name string) error {
	if name == "" {
		return fmt.Errorf("subagent name cannot be empty")
	}

	// Content can be empty for subagents
	if content == "" {
		return nil
	}

	// Sanitize first
	content = SanitizeContent(content)

	// Check for reasonable content length
	if len(content) > 1024*1024 { // 1MB limit
		return fmt.Errorf("subagent content too large (max 1MB)")
	}

	return nil
}
