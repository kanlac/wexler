package apply

import (
	"fmt"
	"os"
	"strings"
)

// ContentExtractor defines the interface for extracting existing content from files
type ContentExtractor interface {
	// ExtractExistingContent extracts the relevant existing content from a file
	// Returns empty string if file doesn't exist or relevant content not found
	ExtractExistingContent(filePath, toolName, fileType string) (string, error)
}

// DefaultContentExtractor implements ContentExtractor with tool-specific logic
type DefaultContentExtractor struct{}

// NewContentExtractor creates a new DefaultContentExtractor
func NewContentExtractor() ContentExtractor {
	return &DefaultContentExtractor{}
}

// ExtractExistingContent extracts existing content based on tool and file type
func (e *DefaultContentExtractor) ExtractExistingContent(filePath, toolName, fileType string) (string, error) {
	// Check if file exists
	if !e.fileExists(filePath) {
		return "", nil // File doesn't exist, return empty content
	}

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	content := string(data)

	// Apply tool and file type specific extraction logic
	switch toolName {
	case "claude":
		return e.extractClaudeContent(content, filePath, fileType)
	case "cursor":
		return e.extractCursorContent(content, filePath, fileType)
	default:
		// Default behavior: return entire file content
		return content, nil
	}
}

// extractClaudeContent extracts content for Claude Code tool
func (e *DefaultContentExtractor) extractClaudeContent(content, filePath, fileType string) (string, error) {
	switch fileType {
	case "memory":
		// For CLAUDE.md, extract only WEXLER section content
		if strings.HasSuffix(filePath, "CLAUDE.md") {
			return e.extractWexlerSection(content), nil
		}
		return content, nil
	case "subagent":
		// For subagent files in .claude/agents/, return entire content
		return content, nil
	case "mcp":
		// For .mcp.json, return entire content
		return content, nil
	default:
		return content, nil
	}
}

// extractCursorContent extracts content for Cursor tool
func (e *DefaultContentExtractor) extractCursorContent(content, filePath, fileType string) (string, error) {
	switch fileType {
	case "memory":
		// For .cursor/rules/general.wexler.mdc, return entire content
		return content, nil
	case "subagent":
		// For .cursor/rules/*.wexler.mdc, return entire content
		return content, nil
	case "mcp":
		// For .cursor/mcp.json, return entire content
		return content, nil
	default:
		return content, nil
	}
}

// extractWexlerSection extracts content under WEXLER level-1 heading
func (e *DefaultContentExtractor) extractWexlerSection(content string) string {
	if content == "" {
		return ""
	}

	lines := strings.Split(content, "\n")
	var wexlerContent []string
	var inWexlerSection bool
	var foundWexler bool

	for _, line := range lines {
		// Check for level-1 heading
		if strings.HasPrefix(line, "# ") {
			sectionName := strings.TrimSpace(strings.TrimPrefix(line, "# "))
			if strings.EqualFold(sectionName, "WEXLER") {
				inWexlerSection = true
				foundWexler = true
				continue // Skip the heading line itself
			} else if inWexlerSection {
				// Found another level-1 heading, exit WEXLER section
				break
			}
		} else if inWexlerSection {
			// We're in the WEXLER section, collect content
			wexlerContent = append(wexlerContent, line)
		}
	}

	if !foundWexler {
		return "" // No WEXLER section found, return empty
	}

	// Join content and trim trailing whitespace
	result := strings.Join(wexlerContent, "\n")
	return strings.TrimRight(result, "\n\t ")
}

// fileExists checks if a file exists
func (e *DefaultContentExtractor) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// MemoryContentExtractor is a specialized extractor for memory configurations
type MemoryContentExtractor struct {
	*DefaultContentExtractor
}

// NewMemoryContentExtractor creates a memory-specific content extractor
func NewMemoryContentExtractor() ContentExtractor {
	return &MemoryContentExtractor{
		DefaultContentExtractor: &DefaultContentExtractor{},
	}
}

// ExtractExistingContent specialized for memory configurations
func (e *MemoryContentExtractor) ExtractExistingContent(filePath, toolName, fileType string) (string, error) {
	if fileType != "memory" {
		return e.DefaultContentExtractor.ExtractExistingContent(filePath, toolName, fileType)
	}

	// Enhanced memory-specific extraction logic can be added here
	return e.DefaultContentExtractor.ExtractExistingContent(filePath, toolName, fileType)
}

// SubagentContentExtractor is a specialized extractor for subagent configurations
type SubagentContentExtractor struct {
	*DefaultContentExtractor
}

// NewSubagentContentExtractor creates a subagent-specific content extractor
func NewSubagentContentExtractor() ContentExtractor {
	return &SubagentContentExtractor{
		DefaultContentExtractor: &DefaultContentExtractor{},
	}
}

// ExtractExistingContent specialized for subagent configurations
func (e *SubagentContentExtractor) ExtractExistingContent(filePath, toolName, fileType string) (string, error) {
	if fileType != "subagent" {
		return e.DefaultContentExtractor.ExtractExistingContent(filePath, toolName, fileType)
	}

	// Enhanced subagent-specific extraction logic can be added here
	return e.DefaultContentExtractor.ExtractExistingContent(filePath, toolName, fileType)
}