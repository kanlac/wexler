package models

import (
	"fmt"
	"strings"
)

// SourceConfig represents the complete source configuration
type SourceConfig struct {
	Memory    *MemoryConfig            `yaml:"memory" json:"memory"`
	Subagents map[string]*SubagentConfig `yaml:"subagents" json:"subagents"`
}

// MemoryConfig represents memory configuration parsed from memory.mdc
// Only the WEXLER section content is extracted and used
type MemoryConfig struct {
	Content      string `yaml:"content" json:"content"`             // raw file content
	WexlerMemory string `yaml:"wexler_memory" json:"wexler_memory"` // WEXLER section content only
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
		Memory:    NewMemoryConfig(),
		Subagents: make(map[string]*SubagentConfig),
	}
}

// NewMemoryConfig creates a new empty memory configuration
func NewMemoryConfig() *MemoryConfig {
	return &MemoryConfig{
		Content:      "",
		WexlerMemory: "",
	}
}

// NewSubagentConfig creates a new subagent configuration
func NewSubagentConfig(name, content string) *SubagentConfig {
	return &SubagentConfig{
		Name:    name,
		Content: content,
	}
}

// ParseMemoryContent parses markdown content and extracts only the WEXLER section
func (m *MemoryConfig) ParseMemoryContent(content string) error {
	m.Content = content
	
	if strings.TrimSpace(content) == "" {
		m.WexlerMemory = ""
		return nil // Empty file is valid
	}
	
	// Parse content and extract WEXLER section (first one wins)
	lines := strings.Split(content, "\n")
	var currentSection string
	var currentContent []string
	var wexlerContent string
	var foundWexler bool
	
	for _, line := range lines {
		// Check for markdown header
		if strings.HasPrefix(line, "# ") {
			// Save WEXLER section if we found it and haven't saved one yet
			if currentSection == "WEXLER" && !foundWexler {
				wexlerContent = strings.Join(currentContent, "\n")
				foundWexler = true
			}
			
			// Start new section
			currentSection = strings.TrimPrefix(line, "# ")
			currentSection = strings.TrimSpace(currentSection)
			currentContent = []string{}
		} else if currentSection != "" {
			// Add line to current section
			currentContent = append(currentContent, line)
		}
	}
	
	// Save final WEXLER section if it was the last one and we haven't found one yet
	if currentSection == "WEXLER" && !foundWexler {
		wexlerContent = strings.Join(currentContent, "\n")
	}
	
	m.WexlerMemory = strings.TrimSpace(wexlerContent)
	return nil
}

// GetWexlerMemory returns the WEXLER section content
func (m *MemoryConfig) GetWexlerMemory() string {
	return m.WexlerMemory
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
			Content:      s.Memory.Content,
			WexlerMemory: s.Memory.WexlerMemory,
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