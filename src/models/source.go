package models

import (
	"fmt"
)

// SourceConfig represents the complete source configuration
type SourceConfig struct {
	Memory    *MemoryConfig              `yaml:"memory" json:"memory"`
	Subagents map[string]*SubagentConfig `yaml:"subagents" json:"subagents"`
}

// MemoryConfig represents memory configuration parsed from memory.mdc files
// Supports dual-scope structure: team (external source) and project (local mindful/)
type MemoryConfig struct {
	TeamContent        string `yaml:"team_content" json:"team_content"`               // external source/memory.mdc content
	ProjectContent     string `yaml:"project_content" json:"project_content"`         // project mindful/memory.mdc content
	HasTeam            bool   `yaml:"has_team" json:"has_team"`                       // team-level file exists
	HasProject         bool   `yaml:"has_project" json:"has_project"`                 // project-level file exists
	TeamSourcePath     string `yaml:"team_source_path" json:"team_source_path"`       // team config source path (for marking)
	ProjectSourcePath  string `yaml:"project_source_path" json:"project_source_path"` // project config source path (for marking)
}

// SubagentConfig represents subagent configuration from subagent/*.mdc files
// Subagent files are applied as complete file replacements (no section parsing)
type SubagentConfig struct {
	Name    string `yaml:"name" json:"name"`       // subagent name (derived from filename)
	Content string `yaml:"content" json:"content"` // complete file content
}

// NewSourceConfig creates a new empty source configuration
func NewSourceConfig() *SourceConfig {
	return &SourceConfig{
		Memory:    nil, // Will be set only if memory files exist
		Subagents: make(map[string]*SubagentConfig),
	}
}

// NewMemoryConfig creates a new empty memory configuration
func NewMemoryConfig() *MemoryConfig {
	return &MemoryConfig{
		TeamContent:       "",
		ProjectContent:    "",
		HasTeam:           false,
		HasProject:        false,
		TeamSourcePath:    "",
		ProjectSourcePath: "",
	}
}

// NewSubagentConfig creates a new subagent configuration
func NewSubagentConfig(name, content string) *SubagentConfig {
	return &SubagentConfig{
		Name:    name,
		Content: content,
	}
}


// Validate checks if the memory configuration is valid
func (m *MemoryConfig) Validate() error {
	if m == nil {
		return fmt.Errorf("memory config is nil")
	}
	// Memory config can be empty, so no validation needed
	return nil
}

// Validate checks if the subagent configuration is valid
func (s *SubagentConfig) Validate() error {
	if s == nil {
		return fmt.Errorf("subagent config is nil")
	}

	if s.Name == "" {
		return fmt.Errorf("subagent name cannot be empty")
	}

	// Content can be empty for subagents
	return nil
}

// AddSubagent adds a subagent configuration
func (s *SourceConfig) AddSubagent(name, content string) {
	if s.Subagents == nil {
		s.Subagents = make(map[string]*SubagentConfig)
	}
	s.Subagents[name] = NewSubagentConfig(name, content)
}

// GetSubagent returns a subagent configuration by name
func (s *SourceConfig) GetSubagent(name string) (*SubagentConfig, bool) {
	if s.Subagents == nil {
		return nil, false
	}
	subagent, exists := s.Subagents[name]
	return subagent, exists
}

// RemoveSubagent removes a subagent configuration
func (s *SourceConfig) RemoveSubagent(name string) {
	if s.Subagents != nil {
		delete(s.Subagents, name)
	}
}

// ListSubagents returns all subagent names
func (s *SourceConfig) ListSubagents() []string {
	if s.Subagents == nil {
		return []string{}
	}

	names := make([]string, 0, len(s.Subagents))
	for name := range s.Subagents {
		names = append(names, name)
	}
	return names
}

// Validate checks if the source configuration is valid
func (s *SourceConfig) Validate() error {
	if s == nil {
		return fmt.Errorf("source config is nil")
	}

	// Validate memory config
	if s.Memory != nil {
		if err := s.Memory.Validate(); err != nil {
			return fmt.Errorf("memory config validation failed: %w", err)
		}
	}

	// Validate subagent configs
	for name, subagent := range s.Subagents {
		if err := subagent.Validate(); err != nil {
			return fmt.Errorf("subagent %s validation failed: %w", name, err)
		}
	}

	return nil
}

// Clone creates a deep copy of the source configuration
func (s *SourceConfig) Clone() *SourceConfig {
	if s == nil {
		return nil
	}

	clone := NewSourceConfig()

	// Clone memory config
	if s.Memory != nil {
		clone.Memory = &MemoryConfig{
			TeamContent:        s.Memory.TeamContent,
			ProjectContent:     s.Memory.ProjectContent,
			HasTeam:            s.Memory.HasTeam,
			HasProject:         s.Memory.HasProject,
			TeamSourcePath:     s.Memory.TeamSourcePath,
			ProjectSourcePath:  s.Memory.ProjectSourcePath,
		}
	}

	// Clone subagent configs
	for name, subagent := range s.Subagents {
		clone.Subagents[name] = &SubagentConfig{
			Name:    subagent.Name,
			Content: subagent.Content,
		}
	}

	return clone
}
