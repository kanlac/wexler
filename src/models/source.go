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
// Memory files use section-based parsing with markdown headers
type MemoryConfig struct {
	Sections map[string]string `yaml:"sections" json:"sections"` // section name -> content
	Content  string            `yaml:"content" json:"content"`   // raw file content
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
		Sections: make(map[string]string),
		Content:  "",
	}
}

// NewSubagentConfig creates a new subagent configuration
func NewSubagentConfig(name, content string) *SubagentConfig {
	return &SubagentConfig{
		Name:    name,
		Content: content,
	}
}

// ParseMemoryContent parses markdown content into sections
// Sections are identified by markdown headers (# Header Name)
func (m *MemoryConfig) ParseMemoryContent(content string) error {
	m.Content = content
	m.Sections = make(map[string]string)
	
	if strings.TrimSpace(content) == "" {
		return nil // Empty file is valid
	}
	
	lines := strings.Split(content, "\n")
	var currentSection string
	var currentContent []string
	
	for _, line := range lines {
		// Check for markdown header
		if strings.HasPrefix(line, "# ") {
			// Save previous section if exists
			if currentSection != "" {
				m.Sections[currentSection] = strings.Join(currentContent, "\n")
			}
			
			// Start new section
			currentSection = strings.TrimPrefix(line, "# ")
			currentContent = []string{}
		} else if currentSection != "" {
			// Add line to current section
			currentContent = append(currentContent, line)
		}
		// Lines before any section header are ignored
	}
	
	// Save final section
	if currentSection != "" {
		m.Sections[currentSection] = strings.Join(currentContent, "\n")
	}
	
	return nil
}

// GetSection returns the content of a specific section
func (m *MemoryConfig) GetSection(sectionName string) (string, bool) {
	if m.Sections == nil {
		return "", false
	}
	content, exists := m.Sections[sectionName]
	return strings.TrimSpace(content), exists
}

// SetSection sets the content of a specific section
func (m *MemoryConfig) SetSection(sectionName, content string) {
	if m.Sections == nil {
		m.Sections = make(map[string]string)
	}
	m.Sections[sectionName] = content
}

// RemoveSection removes a section from the memory configuration
func (m *MemoryConfig) RemoveSection(sectionName string) {
	if m.Sections != nil {
		delete(m.Sections, sectionName)
	}
}

// ListSections returns all section names
func (m *MemoryConfig) ListSections() []string {
	if m.Sections == nil {
		return []string{}
	}
	
	sections := make([]string, 0, len(m.Sections))
	for sectionName := range m.Sections {
		sections = append(sections, sectionName)
	}
	return sections
}

// ToMarkdown reconstructs the markdown content from sections
func (m *MemoryConfig) ToMarkdown() string {
	if len(m.Sections) == 0 {
		return m.Content // Return original if no sections parsed
	}
	
	var parts []string
	for sectionName, content := range m.Sections {
		parts = append(parts, fmt.Sprintf("# %s\n%s", sectionName, content))
	}
	
	return strings.Join(parts, "\n\n")
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
			Content:  s.Memory.Content,
			Sections: make(map[string]string),
		}
		for name, content := range s.Memory.Sections {
			clone.Memory.Sections[name] = content
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