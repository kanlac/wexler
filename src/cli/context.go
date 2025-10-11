package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mindful/src/config"
	"mindful/src/models"
	"mindful/src/source"
	"mindful/src/storage"
)

// ProjectContext aggregates shared services for CLI commands.
type ProjectContext struct {
	ProjectPath    string
	ConfigManager  *config.Manager
	SourceManager  *source.Manager
	StorageManager *storage.Manager
	ProjectConfig  *models.ProjectConfig
}

// NewProjectContext loads project configuration and initialises managers.
func NewProjectContext() (*ProjectContext, error) {
	projectPath, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to determine working directory: %w", err)
	}

	configManager := config.NewManager()
	projectConfig, err := configManager.LoadProject(projectPath)
	if err != nil {
		return nil, err
	}

	ctx := &ProjectContext{
		ProjectPath:   projectPath,
		ConfigManager: configManager,
		SourceManager: source.NewManager(),
		ProjectConfig: projectConfig,
	}

	return ctx, nil
}

// Close releases any resources held by the context.
func (c *ProjectContext) Close() error {
	if c.StorageManager != nil {
		return c.StorageManager.Close()
	}
	return nil
}

// GetStorageManager lazily initialises the storage manager.
func (c *ProjectContext) GetStorageManager() (*storage.Manager, error) {
	if c.StorageManager != nil {
		return c.StorageManager, nil
	}

	dbPath, err := c.ProjectConfig.GetDatabasePath(c.ProjectPath)
	if err != nil {
		return nil, err
	}

	manager, err := storage.NewManager(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open storage at %s: %w", dbPath, err)
	}

	c.StorageManager = manager
	return manager, nil
}

// ResolveTeamSource resolves the configured team source path.
func (c *ProjectContext) ResolveTeamSource() (string, error) {
	return c.ProjectConfig.ResolveSourceRoot(c.ProjectPath)
}

// ResolveOutDir returns mindful/out for the project.
func (c *ProjectContext) ResolveOutDir() string {
	return c.ProjectConfig.ResolveOutDir(c.ProjectPath)
}

// ResolveMindfulDir returns the mindful directory for the project.
func (c *ProjectContext) ResolveMindfulDir() string {
	return c.ProjectConfig.ResolveMindfulDir(c.ProjectPath)
}

// HasMindfulStructure reports whether mindful/ exists.
func (c *ProjectContext) HasMindfulStructure() bool {
	_, err := os.Stat(c.ResolveMindfulDir())
	return err == nil
}

// EnsureMindfulStructure creates mindful/ if missing.
func (c *ProjectContext) EnsureMindfulStructure() error {
	return os.MkdirAll(c.ResolveMindfulDir(), 0o755)
}

// WriteArtifacts writes build artefacts to mindful/out.
func (c *ProjectContext) WriteArtifacts(artifacts *models.BuildArtifacts) error {
	outDir := c.ResolveOutDir()

	if err := os.RemoveAll(outDir); err != nil {
		return fmt.Errorf("failed to clean %s: %w", outDir, err)
	}

	if err := os.MkdirAll(filepath.Join(outDir, "subagents"), 0o755); err != nil {
		return fmt.Errorf("failed to prepare output directories: %w", err)
	}

	if artifacts != nil && artifacts.Memory != nil && strings.TrimSpace(artifacts.Memory.Content) != "" {
		memoryPath := filepath.Join(outDir, "memory.md")
		if err := os.WriteFile(memoryPath, []byte(artifacts.Memory.Content+"\n"), 0o644); err != nil {
			return fmt.Errorf("failed to write %s: %w", memoryPath, err)
		}
	}

	if artifacts != nil {
		for _, subagent := range artifacts.Subagents {
			if subagent == nil || subagent.Content == "" {
				continue
			}
			filename := subagent.FileName
			if filename == "" {
				filename = subagent.Name + ".mdc"
			}
			path := filepath.Join(outDir, "subagents", filename)
			if err := os.WriteFile(path, []byte(subagent.Content+"\n"), 0o644); err != nil {
				return fmt.Errorf("failed to write subagent %s: %w", path, err)
			}
		}

		if len(artifacts.MCPContent) > 0 {
			mcpPath := filepath.Join(outDir, "mcp.json")
			if err := os.WriteFile(mcpPath, artifacts.MCPContent, 0o644); err != nil {
				return fmt.Errorf("failed to write %s: %w", mcpPath, err)
			}
		}
	}

	return nil
}
