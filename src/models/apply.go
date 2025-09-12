package models

import (
	"fmt"
	"time"
)

// ConflictResolution represents user's choice when conflicts are detected
type ConflictResolution int

const (
	// Continue - resolve this conflict and continue with remaining files
	Continue ConflictResolution = iota
	// ContinueAll - resolve all conflicts automatically without prompting
	ContinueAll
	// Stop - halt the operation, preserving changes made so far
	Stop
)

// String returns the string representation of ConflictResolution
func (cr ConflictResolution) String() string {
	switch cr {
	case Continue:
		return "Continue"
	case ContinueAll:
		return "Continue All"
	case Stop:
		return "Stop"
	default:
		return "Unknown"
	}
}

// ApplyProgress tracks the progress of an apply operation
type ApplyProgress struct {
	Total     int       `yaml:"total" json:"total"`         // Total number of operations
	Completed int       `yaml:"completed" json:"completed"` // Number of completed operations
	Current   string    `yaml:"current" json:"current"`     // Currently processing item
	Status    string    `yaml:"status" json:"status"`       // "running", "paused", "completed", "failed"
	StartTime time.Time `yaml:"start_time" json:"start_time"`
	EndTime   time.Time `yaml:"end_time" json:"end_time"`
}

// NewApplyProgress creates a new progress tracker
func NewApplyProgress(total int) *ApplyProgress {
	return &ApplyProgress{
		Total:     total,
		Completed: 0,
		Current:   "",
		Status:    "running",
		StartTime: time.Now(),
	}
}

// UpdateProgress updates the current progress
func (p *ApplyProgress) UpdateProgress(completed int, current string) {
	p.Completed = completed
	p.Current = current
}

// Complete marks the progress as completed
func (p *ApplyProgress) Complete() {
	p.Completed = p.Total
	p.Current = ""
	p.Status = "completed"
	p.EndTime = time.Now()
}

// Fail marks the progress as failed
func (p *ApplyProgress) Fail(current string) {
	p.Current = current
	p.Status = "failed"
	p.EndTime = time.Now()
}

// Pause marks the progress as paused
func (p *ApplyProgress) Pause() {
	p.Status = "paused"
}

// Resume marks the progress as running again
func (p *ApplyProgress) Resume() {
	p.Status = "running"
}

// IsComplete returns true if the progress is completed
func (p *ApplyProgress) IsComplete() bool {
	return p.Status == "completed" || p.Completed >= p.Total
}

// IsFailed returns true if the progress is in failed state
func (p *ApplyProgress) IsFailed() bool {
	return p.Status == "failed"
}

// IsPaused returns true if the progress is paused
func (p *ApplyProgress) IsPaused() bool {
	return p.Status == "paused"
}

// GetPercentage returns the completion percentage
func (p *ApplyProgress) GetPercentage() float64 {
	if p.Total == 0 {
		return 100.0
	}
	return float64(p.Completed) / float64(p.Total) * 100.0
}

// GetDuration returns the duration of the operation
func (p *ApplyProgress) GetDuration() time.Duration {
	if p.EndTime.IsZero() {
		return time.Since(p.StartTime)
	}
	return p.EndTime.Sub(p.StartTime)
}

// FileConflict represents a conflict between existing and new file content
type FileConflict struct {
	FilePath     string `yaml:"file_path" json:"file_path"`         // Path to the conflicting file
	ExistingHash string `yaml:"existing_hash" json:"existing_hash"` // Hash of existing content
	NewHash      string `yaml:"new_hash" json:"new_hash"`           // Hash of new content
	Diff         string `yaml:"diff" json:"diff"`                   // Unified diff of changes
	FileType     string `yaml:"file_type" json:"file_type"`         // "memory", "subagent", "mcp"
}

// NewFileConflict creates a new file conflict
func NewFileConflict(filePath, existingHash, newHash, diff, fileType string) *FileConflict {
	return &FileConflict{
		FilePath:     filePath,
		ExistingHash: existingHash,
		NewHash:      newHash,
		Diff:         diff,
		FileType:     fileType,
	}
}

// ConflictResult represents the result of conflict detection
type ConflictResult struct {
	HasConflicts bool              `yaml:"has_conflicts" json:"has_conflicts"`
	Conflicts    []*FileConflict   `yaml:"conflicts" json:"conflicts"`
	Resolution   ConflictResolution `yaml:"resolution" json:"resolution"`
}

// NewConflictResult creates a new conflict result
func NewConflictResult() *ConflictResult {
	return &ConflictResult{
		HasConflicts: false,
		Conflicts:    []*FileConflict{},
		Resolution:   Stop, // Default to stopping on conflicts
	}
}

// AddConflict adds a file conflict to the result
func (cr *ConflictResult) AddConflict(conflict *FileConflict) {
	cr.Conflicts = append(cr.Conflicts, conflict)
	cr.HasConflicts = true
}

// GetConflictCount returns the number of conflicts
func (cr *ConflictResult) GetConflictCount() int {
	return len(cr.Conflicts)
}

// GetConflictsByType returns conflicts filtered by file type
func (cr *ConflictResult) GetConflictsByType(fileType string) []*FileConflict {
	var filtered []*FileConflict
	for _, conflict := range cr.Conflicts {
		if conflict.FileType == fileType {
			filtered = append(filtered, conflict)
		}
	}
	return filtered
}

// ApplyConfig represents configuration for an apply operation
type ApplyConfig struct {
	ProjectPath string        `yaml:"project_path" json:"project_path"` // Root path of the project
	ToolName    string        `yaml:"tool_name" json:"tool_name"`       // Target tool (claude, cursor)
	Source      *SourceConfig `yaml:"source" json:"source"`             // Source configuration to apply
	MCP         *MCPConfig    `yaml:"mcp" json:"mcp"`                   // MCP configuration to apply
	DryRun      bool          `yaml:"dry_run" json:"dry_run"`           // If true, don't actually write files
	Force       bool          `yaml:"force" json:"force"`               // If true, overwrite without prompting
}

// NewApplyConfig creates a new apply configuration
func NewApplyConfig(projectPath, toolName string) *ApplyConfig {
	return &ApplyConfig{
		ProjectPath: projectPath,
		ToolName:    toolName,
		Source:      NewSourceConfig(),
		MCP:         NewMCPConfig(),
		DryRun:      false,
		Force:       false,
	}
}

// Validate checks if the apply configuration is valid
func (ac *ApplyConfig) Validate() error {
	if ac == nil {
		return fmt.Errorf("apply config is nil")
	}
	
	if ac.ProjectPath == "" {
		return fmt.Errorf("project path cannot be empty")
	}
	
	if ac.ToolName == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	
	// Validate source config if present
	if ac.Source != nil {
		if err := ac.Source.Validate(); err != nil {
			return fmt.Errorf("source config validation failed: %w", err)
		}
	}
	
	// Validate MCP config if present
	if ac.MCP != nil {
		if err := ac.MCP.Validate(); err != nil {
			return fmt.Errorf("MCP config validation failed: %w", err)
		}
	}
	
	return nil
}

// ApplyResult represents the result of an apply operation
type ApplyResult struct {
	Success      bool             `yaml:"success" json:"success"`           // Overall operation success
	FilesWritten []string         `yaml:"files_written" json:"files_written"` // List of files that were written
	FilesSkipped []string         `yaml:"files_skipped" json:"files_skipped"` // List of files that were skipped
	Conflicts    []*FileConflict  `yaml:"conflicts" json:"conflicts"`       // Conflicts encountered
	Progress     *ApplyProgress   `yaml:"progress" json:"progress"`         // Progress information
	Error        string           `yaml:"error,omitempty" json:"error,omitempty"` // Error message if failed
}

// NewApplyResult creates a new apply result
func NewApplyResult() *ApplyResult {
	return &ApplyResult{
		Success:      false,
		FilesWritten: []string{},
		FilesSkipped: []string{},
		Conflicts:    []*FileConflict{},
		Progress:     nil,
		Error:        "",
	}
}

// AddWrittenFile adds a file to the written files list
func (ar *ApplyResult) AddWrittenFile(filePath string) {
	ar.FilesWritten = append(ar.FilesWritten, filePath)
}

// AddSkippedFile adds a file to the skipped files list
func (ar *ApplyResult) AddSkippedFile(filePath string) {
	ar.FilesSkipped = append(ar.FilesSkipped, filePath)
}

// AddConflict adds a conflict to the result
func (ar *ApplyResult) AddConflict(conflict *FileConflict) {
	ar.Conflicts = append(ar.Conflicts, conflict)
}

// SetError sets the error message and marks the result as failed
func (ar *ApplyResult) SetError(err error) {
	ar.Success = false
	ar.Error = err.Error()
	if ar.Progress != nil {
		ar.Progress.Fail(ar.Progress.Current)
	}
}

// SetSuccess marks the result as successful
func (ar *ApplyResult) SetSuccess() {
	ar.Success = true
	ar.Error = ""
	if ar.Progress != nil {
		ar.Progress.Complete()
	}
}

// GetSummary returns a human-readable summary of the apply result
func (ar *ApplyResult) GetSummary() string {
	if !ar.Success {
		return fmt.Sprintf("Apply failed: %s", ar.Error)
	}
	
	return fmt.Sprintf("Apply successful: %d files written, %d files skipped, %d conflicts resolved",
		len(ar.FilesWritten), len(ar.FilesSkipped), len(ar.Conflicts))
}