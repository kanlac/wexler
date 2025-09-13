package apply

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"wexler/src/models"
	"wexler/src/tools"
)

// Manager implements ApplyManager interface for configuration application
type Manager struct{
	contentExtractor ContentExtractor
}

// NewManager creates a new ApplyManager instance
func NewManager() *Manager {
	return &Manager{
		contentExtractor: NewContentExtractor(),
	}
}

// NewManagerWithExtractor creates a new ApplyManager instance with custom content extractor
func NewManagerWithExtractor(extractor ContentExtractor) *Manager {
	return &Manager{
		contentExtractor: extractor,
	}
}

// ApplyConfig applies source configuration to the target tool
func (m *Manager) ApplyConfig(config *models.ApplyConfig) (*models.ApplyResult, error) {
	if config == nil {
		return nil, fmt.Errorf("apply config cannot be nil")
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid apply config: %w", err)
	}

	result := models.NewApplyResult()
	
	// Create tool adapter
	adapter, err := tools.NewAdapter(config.ToolName)
	if err != nil {
		result.SetError(fmt.Errorf("failed to create tool adapter: %w", err))
		return result, err
	}

	// Convert to tool config
	toolConfig := m.convertToToolConfig(config)
	
	// Generate configuration files
	files, err := adapter.Generate(toolConfig)
	if err != nil {
		result.SetError(fmt.Errorf("failed to generate configuration files: %w", err))
		return result, err
	}

	// Set up progress tracking
	result.Progress = models.NewApplyProgress(len(files))

	// Process each file
	for i, file := range files {
		result.Progress.UpdateProgress(i, file.Path)
		
		targetPath := filepath.Join(config.ProjectPath, file.Path)
		
		if config.DryRun {
			// In dry run mode, just track what would be written
			result.AddSkippedFile(file.Path)
		} else {
			// Check for conflicts
			if m.fileExists(targetPath) && !config.Force {
				// Extract existing content using tool and file type specific logic
				existingContent, err := m.contentExtractor.ExtractExistingContent(targetPath, config.ToolName, file.Type)
				if err != nil {
					result.SetError(fmt.Errorf("failed to extract existing content from %s: %w", targetPath, err))
					return result, err
				}

				if existingContent != file.Content {
					// Create conflict
					conflict := m.createConflict(file.Path, existingContent, file.Content, file.Type)
					result.AddConflict(conflict)
					result.AddSkippedFile(file.Path)
					continue
				}
			}

			// Write the file with appropriate content transformation
			actualContent, err := m.getActualWriteContent(file, config.ToolName)
			if err != nil {
				result.SetError(fmt.Errorf("failed to generate write content for %s: %w", targetPath, err))
				return result, err
			}
			
			if err := m.writeFile(targetPath, actualContent); err != nil {
				result.SetError(fmt.Errorf("failed to write file %s: %w", targetPath, err))
				return result, err
			}
			result.AddWrittenFile(file.Path)
		}
	}

	// Check if we have conflicts - but don't fail the operation
	if len(result.Conflicts) > 0 {
		// Mark as successful but with conflicts detected
		result.Success = true
		if result.Error == "" {
			result.Error = fmt.Sprintf("%d conflicts detected but proceeding", len(result.Conflicts))
		}
	} else {
		result.SetSuccess()
	}

	return result, nil
}

// DetectConflicts detects potential conflicts without applying changes
func (m *Manager) DetectConflicts(config *models.ApplyConfig) ([]*models.FileConflict, error) {
	if config == nil {
		return nil, fmt.Errorf("apply config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid apply config: %w", err)
	}

	var conflicts []*models.FileConflict

	// Create tool adapter
	adapter, err := tools.NewAdapter(config.ToolName)
	if err != nil {
		return nil, fmt.Errorf("failed to create tool adapter: %w", err)
	}

	// Convert to tool config
	toolConfig := m.convertToToolConfig(config)

	// Generate configuration files
	files, err := adapter.Generate(toolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate configuration files: %w", err)
	}

	// Check each file for conflicts
	for _, file := range files {
		targetPath := filepath.Join(config.ProjectPath, file.Path)
		
		if m.fileExists(targetPath) {
			// Extract existing content using tool and file type specific logic
			existingContent, err := m.contentExtractor.ExtractExistingContent(targetPath, config.ToolName, file.Type)
			if err != nil {
				return nil, fmt.Errorf("failed to extract existing content from %s: %w", targetPath, err)
			}

			if existingContent != file.Content {
				conflict := m.createConflict(file.Path, existingContent, file.Content, file.Type)
				conflicts = append(conflicts, conflict)
			}
		}
	}

	return conflicts, nil
}

// ResolveConflicts applies the specified resolution to conflicts
func (m *Manager) ResolveConflicts(conflicts []*models.FileConflict, resolution models.ConflictResolution) error {
	if len(conflicts) == 0 {
		return nil // No conflicts to resolve
	}

	switch resolution {
	case models.Continue:
		// Continue processing - this is usually handled by the caller
		return nil
	case models.ContinueAll:
		// Continue with all conflicts - this is usually handled by the caller
		return nil
	case models.Stop:
		// Stop processing - this is the default behavior
		return fmt.Errorf("operation stopped due to %d conflicts", len(conflicts))
	default:
		return fmt.Errorf("unknown conflict resolution: %v", resolution)
	}
}

// convertToToolConfig converts ApplyConfig to ToolConfig
func (m *Manager) convertToToolConfig(config *models.ApplyConfig) *tools.ToolConfig {
	toolConfig := &tools.ToolConfig{
		ToolName:  config.ToolName,
		Memory:    config.Source.Memory,
		MCP:       config.MCP,
	}

	// Convert subagents map to slice
	if config.Source.Subagents != nil {
		subagents := make([]*models.SubagentConfig, 0, len(config.Source.Subagents))
		for _, subagent := range config.Source.Subagents {
			subagents = append(subagents, subagent)
		}
		toolConfig.Subagents = subagents
	}

	return toolConfig
}

// fileExists checks if a file exists
func (m *Manager) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// writeFile writes content to a file, creating directories as needed
func (m *Manager) writeFile(path, content string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the file
	return os.WriteFile(path, []byte(content), 0644)
}

// createConflict creates a file conflict object
func (m *Manager) createConflict(filePath, existingContent, newContent, fileType string) *models.FileConflict {
	existingHash := m.hashContent(existingContent)
	newHash := m.hashContent(newContent)
	diff := m.generateDiff(existingContent, newContent)

	return models.NewFileConflict(filePath, existingHash, newHash, diff, fileType)
}

// hashContent generates a simple hash for content
func (m *Manager) hashContent(content string) string {
	// Simple hash - in production, you might use SHA256
	return fmt.Sprintf("%x", len(content)^0xDEADBEEF)
}

// generateDiff generates a simple diff representation
func (m *Manager) generateDiff(existing, new string) string {
	return fmt.Sprintf("-%d lines, +%d lines", 
		len(existing), len(new))
}

// getActualWriteContent gets the actual content to write to file
// For some file types like CLAUDE.md, this involves merging with existing content
func (m *Manager) getActualWriteContent(file tools.ConfigFile, toolName string) (string, error) {
	switch toolName {
	case "claude":
		return m.getClaudeWriteContent(file)
	case "cursor":
		return m.getCursorWriteContent(file)
	default:
		return file.Content, nil
	}
}

// getClaudeWriteContent gets write content for Claude tool
func (m *Manager) getClaudeWriteContent(file tools.ConfigFile) (string, error) {
	if file.Type == "memory" && file.Path == "CLAUDE.md" {
		// For CLAUDE.md, we need to generate full content with WEXLER section merged
		return m.generateClaudeMemoryContent(file.Content)
	}
	return file.Content, nil
}

// getCursorWriteContent gets write content for Cursor tool
func (m *Manager) getCursorWriteContent(file tools.ConfigFile) (string, error) {
	// Cursor files are written as-is
	return file.Content, nil
}

// generateClaudeMemoryContent generates full CLAUDE.md content with WEXLER section
func (m *Manager) generateClaudeMemoryContent(wexlerContent string) (string, error) {
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

	// Upsert WEXLER section
	existingSections["WEXLER"] = wexlerContent

	// Reconstruct markdown with WEXLER first, then other sections
	var parts []string
	
	// Add WEXLER section first
	if wexlerContent, exists := existingSections["WEXLER"]; exists && strings.TrimSpace(wexlerContent) != "" {
		parts = append(parts, fmt.Sprintf("# WEXLER\n%s", strings.TrimSpace(wexlerContent)))
		delete(existingSections, "WEXLER") // Remove from remaining sections
	}

	// Add other sections
	for sectionName, content := range existingSections {
		if sectionName != "" && strings.TrimSpace(content) != "" {
			parts = append(parts, fmt.Sprintf("# %s\n%s", sectionName, strings.TrimSpace(content)))
		}
	}

	return strings.Join(parts, "\n\n"), nil
}