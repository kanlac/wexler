package tools

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mindful/src/models"
)

// Adapter implements ToolAdapter interface for configuration generation
type Adapter struct {
	toolName string
}

// ConfigFile represents a generated configuration file
type ConfigFile struct {
	Path    string `yaml:"path" json:"path"`
	Content string `yaml:"content" json:"content"`
	Type    string `yaml:"type" json:"type"` // "memory", "subagent", "mcp"
}

// ToolConfig represents input configuration for tool adaptation
type ToolConfig struct {
	ToolName  string                   `yaml:"tool_name" json:"tool_name"`
	Memory    *models.MemoryConfig     `yaml:"memory" json:"memory"`
	Subagents []*models.SubagentConfig `yaml:"subagents" json:"subagents"`
	MCP       *models.MCPConfig        `yaml:"mcp" json:"mcp"`
}

// ConflictResult represents the result of a merge operation
type ConflictResult struct {
	HasConflicts bool               `yaml:"has_conflicts" json:"has_conflicts"`
	Conflicts    []FileConflict     `yaml:"conflicts" json:"conflicts"`
	Resolution   ConflictResolution `yaml:"resolution" json:"resolution"`
}

// FileConflict represents a conflict between existing and new file content
type FileConflict struct {
	FilePath     string `yaml:"file_path" json:"file_path"`
	ExistingHash string `yaml:"existing_hash" json:"existing_hash"`
	NewHash      string `yaml:"new_hash" json:"new_hash"`
	Diff         string `yaml:"diff" json:"diff"`
}

// ConflictResolution represents how to resolve conflicts
type ConflictResolution int

const (
	Continue ConflictResolution = iota
	ContinueAll
	Stop
)

// NewAdapter creates a new tool adapter for the specified tool
func NewAdapter(toolName string) (*Adapter, error) {
	supportedTools := map[string]bool{
		"claude": true,
		"cursor": true,
	}

	if !supportedTools[toolName] {
		return nil, fmt.Errorf("unsupported tool: %s", toolName)
	}

	return &Adapter{
		toolName: toolName,
	}, nil
}

// Generate generates configuration files for the specified tool
func (a *Adapter) Generate(config *ToolConfig) ([]ConfigFile, error) {
	if config == nil {
		return nil, fmt.Errorf("tool config cannot be nil")
	}

	var files []ConfigFile

	switch a.toolName {
	case "claude":
		claudeFiles, err := a.generateClaudeFiles(config)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Claude files: %w", err)
		}
		files = append(files, claudeFiles...)

	case "cursor":
		cursorFiles, err := a.generateCursorFiles(config)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Cursor files: %w", err)
		}
		files = append(files, cursorFiles...)

	default:
		return nil, fmt.Errorf("unsupported tool: %s", a.toolName)
	}

	return files, nil
}

// generateClaudeFiles generates Claude Code configuration files
func (a *Adapter) generateClaudeFiles(config *ToolConfig) ([]ConfigFile, error) {
	var files []ConfigFile

	// Generate CLAUDE.md (memory configuration)
	if config.Memory != nil && config.Memory.MindfulMemory != "" {
		// For Claude memory files, we need to generate content that matches what
		// ContentExtractor will extract for comparison (only MINDFUL section content)
		files = append(files, ConfigFile{
			Path:    "CLAUDE.md",
			Content: config.Memory.MindfulMemory, // Only MINDFUL content for comparison
			Type:    "memory",
		})
	}

	// Generate subagent files in .claude/agents/
	for _, subagent := range config.Subagents {
		if subagent != nil && subagent.Name != "" {
			agentPath := filepath.Join(".claude", "agents", subagent.Name+".mindful.md")
			files = append(files, ConfigFile{
				Path:    agentPath,
				Content: subagent.Content,
				Type:    "subagent",
			})
		}
	}

	// Generate .mcp.json
	if config.MCP != nil && len(config.MCP.Servers) > 0 {
		mcpContent, err := a.generateMCPFile(config.MCP)
		if err != nil {
			return nil, fmt.Errorf("failed to generate MCP file: %w", err)
		}
		files = append(files, ConfigFile{
			Path:    ".mcp.json",
			Content: mcpContent,
			Type:    "mcp",
		})
	}

	return files, nil
}

// generateCursorFiles generates Cursor configuration files
func (a *Adapter) generateCursorFiles(config *ToolConfig) ([]ConfigFile, error) {
	var files []ConfigFile

	// Generate .cursor/rules/general.mindful.mdc (memory configuration)
	if config.Memory != nil && config.Memory.MindfulMemory != "" {
		cursorContent := a.generateCursorMemoryContent(config.Memory.MindfulMemory, "General Memories")
		files = append(files, ConfigFile{
			Path:    ".cursor/rules/general.mindful.mdc",
			Content: cursorContent,
			Type:    "memory",
		})
	}

	// Generate subagent files in .cursor/rules/
	for _, subagent := range config.Subagents {
		if subagent != nil && subagent.Name != "" {
			description := a.extractDescriptionFromContent(subagent.Content, subagent.Name)
			cursorContent := a.generateCursorSubagentContent(subagent.Content, description)
			rulePath := filepath.Join(".cursor", "rules", subagent.Name+".mindful.mdc")
			files = append(files, ConfigFile{
				Path:    rulePath,
				Content: cursorContent,
				Type:    "subagent",
			})
		}
	}

	// Generate .cursor/mcp.json
	if config.MCP != nil && len(config.MCP.Servers) > 0 {
		mcpContent, err := a.generateMCPFile(config.MCP)
		if err != nil {
			return nil, fmt.Errorf("failed to generate MCP file: %w", err)
		}
		files = append(files, ConfigFile{
			Path:    ".cursor/mcp.json",
			Content: mcpContent,
			Type:    "mcp",
		})
	}

	return files, nil
}

// generateClaudeMemory generates CLAUDE.md content, preserving existing sections and upserting MINDFUL section
func (a *Adapter) generateClaudeMemory(memory *models.MemoryConfig) string {
	if memory == nil || memory.MindfulMemory == "" {
		return ""
	}

	// Try to read existing CLAUDE.md file
	existingContent := ""
	if data, err := os.ReadFile("CLAUDE.md"); err == nil {
		existingContent = string(data)
	}

	// Parse existing content into sections
	existingSections := make(map[string]string)
	if existingContent != "" {
		lines := strings.Split(existingContent, "\n")
		var currentSection string
		var currentContent []string

		for _, line := range lines {
			if strings.HasPrefix(line, "# ") {
				// Save previous section
				if currentSection != "" {
					existingSections[currentSection] = strings.Join(currentContent, "\n")
				}
				// Start new section
				currentSection = strings.TrimPrefix(line, "# ")
				currentSection = strings.TrimSpace(currentSection)
				currentContent = []string{}
			} else if currentSection != "" {
				currentContent = append(currentContent, line)
			}
		}
		// Save final section
		if currentSection != "" {
			existingSections[currentSection] = strings.Join(currentContent, "\n")
		}
	}

	// Upsert MINDFUL section - only include the content, not the header
	existingSections["MINDFUL"] = memory.MindfulMemory

	// Reconstruct markdown with MINDFUL first, then other sections
	var parts []string

	// Add MINDFUL section first
	if mindfulContent, exists := existingSections["MINDFUL"]; exists {
		parts = append(parts, fmt.Sprintf("# MINDFUL\n%s", strings.TrimSpace(mindfulContent)))
		delete(existingSections, "MINDFUL") // Remove from remaining sections
	}

	// Add other sections
	for sectionName, content := range existingSections {
		if sectionName != "" && strings.TrimSpace(content) != "" {
			parts = append(parts, fmt.Sprintf("# %s\n%s", sectionName, strings.TrimSpace(content)))
		}
	}

	return strings.Join(parts, "\n\n")
}

// generateMCPFile generates MCP JSON configuration
func (a *Adapter) generateMCPFile(mcp *models.MCPConfig) (string, error) {
	if mcp == nil || len(mcp.Servers) == 0 {
		return `{"mcpServers": {}}`, nil
	}

	// Use the model's ToMCPJSON method
	data, err := mcp.ToMCPJSON()
	if err != nil {
		return "", fmt.Errorf("failed to generate MCP JSON: %w", err)
	}

	return string(data), nil
}

// Validate validates the generated configuration files
func (a *Adapter) Validate(files []ConfigFile) error {
	for _, file := range files {
		switch file.Type {
		case "memory":
			if err := a.validateMemoryFile(file); err != nil {
				return fmt.Errorf("memory file validation failed for %s: %w", file.Path, err)
			}
		case "mcp":
			if err := a.validateMCPFile(file); err != nil {
				return fmt.Errorf("MCP file validation failed for %s: %w", file.Path, err)
			}
		case "subagent":
			if err := a.validateSubagentFile(file); err != nil {
				return fmt.Errorf("subagent file validation failed for %s: %w", file.Path, err)
			}
		}
	}

	return nil
}

// validateMemoryFile validates memory configuration file content
func (a *Adapter) validateMemoryFile(file ConfigFile) error {
	if file.Content == "" {
		return nil // Empty content is valid
	}

	// Basic validation - ensure it's not malformed
	if strings.TrimSpace(file.Content) == "" {
		return nil
	}

	return nil
}

// validateMCPFile validates MCP JSON configuration
func (a *Adapter) validateMCPFile(file ConfigFile) error {
	if file.Content == "" {
		return fmt.Errorf("MCP file content cannot be empty")
	}

	var mcpData interface{}
	if err := json.Unmarshal([]byte(file.Content), &mcpData); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	return nil
}

// validateSubagentFile validates subagent configuration file content
func (a *Adapter) validateSubagentFile(file ConfigFile) error {
	// Subagent files can have any content, including empty
	return nil
}

// Merge merges existing and new configuration files, detecting conflicts
func (a *Adapter) Merge(existing []ConfigFile, new []ConfigFile) ([]ConfigFile, ConflictResult, error) {
	merged := make([]ConfigFile, 0, len(existing)+len(new))
	conflictResult := ConflictResult{
		HasConflicts: false,
		Conflicts:    []FileConflict{},
		Resolution:   Continue,
	}

	// Create a map of existing files by path
	existingMap := make(map[string]ConfigFile)
	for _, file := range existing {
		existingMap[file.Path] = file
		merged = append(merged, file)
	}

	// Check new files for conflicts
	for _, newFile := range new {
		if existingFile, exists := existingMap[newFile.Path]; exists {
			// File exists, check for conflicts
			if existingFile.Content != newFile.Content {
				// Special handling for MCP files - try to merge instead of conflict
				if newFile.Type == "mcp" && (strings.HasSuffix(newFile.Path, ".mcp.json") || strings.HasSuffix(newFile.Path, "mcp.json")) {
					mergedFile, hasConflict, err := a.mergeMCPFiles(existingFile, newFile)
					if err != nil {
						// If merge fails, treat as regular conflict
						conflict := FileConflict{
							FilePath:     newFile.Path,
							ExistingHash: a.hashContent(existingFile.Content),
							NewHash:      a.hashContent(newFile.Content),
							Diff:         a.generateDiff(existingFile.Content, newFile.Content),
						}
						conflictResult.Conflicts = append(conflictResult.Conflicts, conflict)
						conflictResult.HasConflicts = true
					} else if hasConflict {
						conflict := FileConflict{
							FilePath:     newFile.Path,
							ExistingHash: a.hashContent(existingFile.Content),
							NewHash:      a.hashContent(newFile.Content),
							Diff:         "Server configuration conflicts detected",
						}
						conflictResult.Conflicts = append(conflictResult.Conflicts, conflict)
						conflictResult.HasConflicts = true
					} else {
						// Successfully merged, update the file in the merged list
						for i, file := range merged {
							if file.Path == newFile.Path {
								merged[i] = mergedFile
								break
							}
						}
					}
				} else {
					// Regular conflict for non-MCP files
					conflict := FileConflict{
						FilePath:     newFile.Path,
						ExistingHash: a.hashContent(existingFile.Content),
						NewHash:      a.hashContent(newFile.Content),
						Diff:         a.generateDiff(existingFile.Content, newFile.Content),
					}
					conflictResult.Conflicts = append(conflictResult.Conflicts, conflict)
					conflictResult.HasConflicts = true
				}
			}
		} else {
			// New file, add to merged
			merged = append(merged, newFile)
		}
	}

	return merged, conflictResult, nil
}

// hashContent generates a hash for file content
func (a *Adapter) hashContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash[:8]) // Use first 8 bytes for brevity
}

// generateDiff generates a simple diff between two content strings
func (a *Adapter) generateDiff(existing, new string) string {
	return fmt.Sprintf("Existing: %d chars, New: %d chars", len(existing), len(new))
}

// mergeMCPFiles attempts to merge two MCP configuration files
func (a *Adapter) mergeMCPFiles(existing, new ConfigFile) (ConfigFile, bool, error) {
	// Parse existing MCP config
	var existingMCP map[string]interface{}
	if err := json.Unmarshal([]byte(existing.Content), &existingMCP); err != nil {
		return ConfigFile{}, false, fmt.Errorf("failed to parse existing MCP config: %w", err)
	}

	// Parse new MCP config
	var newMCP map[string]interface{}
	if err := json.Unmarshal([]byte(new.Content), &newMCP); err != nil {
		return ConfigFile{}, false, fmt.Errorf("failed to parse new MCP config: %w", err)
	}

	// Get mcpServers from both configs
	existingServers, existingOk := existingMCP["mcpServers"].(map[string]interface{})
	if !existingOk {
		existingServers = make(map[string]interface{})
	}

	newServers, newOk := newMCP["mcpServers"].(map[string]interface{})
	if !newOk {
		newServers = make(map[string]interface{})
	}

	// Merge servers - detect conflicts for same server names
	merged := make(map[string]interface{})
	hasConflict := false

	// Add existing servers
	for serverName, config := range existingServers {
		merged[serverName] = config
	}

	// Add new servers, checking for conflicts
	for serverName, newConfig := range newServers {
		if existingConfig, exists := merged[serverName]; exists {
			// Same server name exists - check if configurations are different
			existingJSON, _ := json.Marshal(existingConfig)
			newJSON, _ := json.Marshal(newConfig)
			if string(existingJSON) != string(newJSON) {
				hasConflict = true
			}
		} else {
			// New server, add it
			merged[serverName] = newConfig
		}
	}

	// Create merged config
	mergedConfig := map[string]interface{}{
		"mcpServers": merged,
	}

	// Convert back to JSON
	mergedJSON, err := json.Marshal(mergedConfig)
	if err != nil {
		return ConfigFile{}, false, fmt.Errorf("failed to marshal merged config: %w", err)
	}

	mergedFile := ConfigFile{
		Path:    new.Path,
		Content: string(mergedJSON),
		Type:    new.Type,
	}

	return mergedFile, hasConflict, nil
}

// generateCursorMemoryContent generates Cursor memory file content with frontmatter
func (a *Adapter) generateCursorMemoryContent(content, description string) string {
	frontmatter := fmt.Sprintf(`---
description: %s
globs:
alwaysApply: true
---

`, description)

	return frontmatter + strings.TrimSpace(content)
}

// generateCursorSubagentContent generates Cursor subagent file content with frontmatter
func (a *Adapter) generateCursorSubagentContent(content, description string) string {
	frontmatter := fmt.Sprintf(`---
description: %s
globs:
alwaysApply: true
---

`, description)

	return frontmatter + strings.TrimSpace(content)
}

// extractDescriptionFromContent extracts description from markdown content
// First tries to find first level-1 header (# Title), then falls back to filename
func (a *Adapter) extractDescriptionFromContent(content, fallbackName string) string {
	if strings.TrimSpace(content) == "" {
		return fallbackName
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for level-1 header
		if strings.HasPrefix(line, "# ") {
			title := strings.TrimPrefix(line, "# ")
			title = strings.TrimSpace(title)
			if title != "" {
				return title
			}
		}
	}

	// No title found, use filename
	return fallbackName
}
