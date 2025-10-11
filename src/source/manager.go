package source

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"mindful/src/models"
)

// Manager loads configuration sources and renders unified build artefacts.
type Manager struct{}

// NewManager creates a new Manager instance.
func NewManager() *Manager {
	return &Manager{}
}

// LoadArtifacts loads memory, subagents, and other assets from the team source and project directories.
func (m *Manager) LoadArtifacts(teamSourcePath, projectPath string) (*models.BuildArtifacts, error) {
	if projectPath == "" {
		return nil, fmt.Errorf("project path cannot be empty")
	}

	mindfulDir := filepath.Join(projectPath, models.DefaultMindfulDirName)

	memory, err := m.buildMemoryArtifact(teamSourcePath, mindfulDir)
	if err != nil {
		return nil, err
	}

	subagents, err := m.buildSubagentArtifacts(teamSourcePath, mindfulDir)
	if err != nil {
		return nil, err
	}

	artifacts := &models.BuildArtifacts{
		Memory:    memory,
		Subagents: subagents,
	}

	return artifacts, nil
}

func (m *Manager) buildMemoryArtifact(teamSourcePath, mindfulDir string) (*models.MemoryArtifact, error) {
	var segments []string
	var sources []string

	if teamSourcePath != "" {
		if content, sourcePath, err := m.readOptionalFile(teamSourcePath, []string{"memory.md", "memory.mdc"}); err != nil {
			return nil, fmt.Errorf("failed to read team memory: %w", err)
		} else if content != "" {
			segments = append(segments, annotateContent("team", sourcePath, content))
			sources = append(sources, sourcePath)
		}
	}

	if content, sourcePath, err := m.readOptionalFile(mindfulDir, []string{"project-memory.mdc", "project-memory.md", "memory.mdc"}); err != nil {
		return nil, fmt.Errorf("failed to read project memory: %w", err)
	} else if content != "" {
		segments = append(segments, annotateContent("project", sourcePath, content))
		sources = append(sources, sourcePath)
	}

	if len(segments) == 0 {
		return nil, nil
	}

	return &models.MemoryArtifact{
		Content:     strings.Join(segments, "\n\n"),
		SourcePaths: sources,
	}, nil
}

func (m *Manager) buildSubagentArtifacts(teamSourcePath, mindfulDir string) ([]*models.SubagentArtifact, error) {
	results := make(map[string]*models.SubagentArtifact)

	// Load team subagents first
	if teamSourcePath != "" {
		if err := m.mergeSubagentDir(results, filepath.Join(teamSourcePath, "subagents"), "team"); err != nil {
			return nil, err
		}
	}

	// Project overrides
	projectDirs := []string{
		filepath.Join(mindfulDir, "project-subagents"),
		filepath.Join(mindfulDir, "subagents"), // legacy fallback
	}
	for _, dir := range projectDirs {
		if err := m.mergeSubagentDir(results, dir, "project"); err != nil {
			return nil, err
		}
	}

	if len(results) == 0 {
		return nil, nil
	}

	names := make([]string, 0, len(results))
	for name := range results {
		names = append(names, name)
	}
	sort.Strings(names)

	var artifacts []*models.SubagentArtifact
	for _, name := range names {
		artifacts = append(artifacts, results[name])
	}

	return artifacts, nil
}

func (m *Manager) mergeSubagentDir(target map[string]*models.SubagentArtifact, dir string, scope string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read subagent directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		if name == "" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read subagent file %s: %w", path, err)
		}

		content := normalizeContent(string(data))

		target[name] = &models.SubagentArtifact{
			Name:       name,
			FileName:   entry.Name(),
			Content:    annotateContent(scope, path, content),
			SourcePath: path,
		}
	}

	return nil
}

func (m *Manager) readOptionalFile(basePath string, filenames []string) (string, string, error) {
	for _, name := range filenames {
		path := filepath.Join(basePath, name)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", "", err
		}
		content := normalizeContent(string(data))
		if strings.TrimSpace(content) == "" {
			continue
		}
		return content, path, nil
	}
	return "", "", nil
}

func normalizeContent(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	return strings.TrimSpace(content)
}

func annotateContent(scope, sourcePath, content string) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("<!-- scope:%s source:%s -->\n", scope, sourcePath))
	builder.WriteString(strings.TrimSpace(content))
	return builder.String()
}
