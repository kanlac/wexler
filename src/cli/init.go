package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"wexler/src/models"

	"github.com/spf13/cobra"
)

var (
	initSourcePath string
	initName       string
	initVersion    string
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [PROJECT_NAME]",
		Short: "Initialize Wexler in project directory",
		Long: `Initialize Wexler configuration management in the current directory.

Creates a wexler.yaml configuration file and sets up the basic project structure
including source directory for AI configurations and MCP settings.`,
		Example: `  # Initialize with default settings
  wexler init

  # Initialize with custom project name
  wexler init my-project

  # Initialize with custom source directory
  wexler init --source=/usr/ai-configs`,
		Args: cobra.MaximumNArgs(1),
		RunE: runInit,
	}

	cmd.Flags().StringVar(&initSourcePath, "source", models.DefaultWexlerSource, "source directory for AI configurations")
	cmd.Flags().StringVar(&initName, "name", "", "project name (default: directory name)")
	cmd.Flags().StringVar(&initVersion, "version", "1.0.0", "project version")

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	ctx, err := NewAppContext()
	if err != nil {
		return err
	}
	defer ctx.CloseResources()

	projectPath := ctx.ProjectPath

	// Determine project name
	projectName := initName
	if len(args) > 0 {
		projectName = args[0]
	}
	if projectName == "" {
		projectName = filepath.Base(projectPath)
	}

	if verbose {
		fmt.Printf("Initializing Wexler project '%s' in %s\n", projectName, projectPath)
	}

	// Check if already initialized
	configPath := filepath.Join(projectPath, "wexler.yaml")
	if _, err := os.Stat(configPath); err == nil && !force {
		return fmt.Errorf("wexler.yaml already exists (use --force to overwrite)")
	}

	// Create project configuration
	config := &models.ProjectConfig{
		Name:       projectName,
		Version:    initVersion,
		SourcePath: initSourcePath,
		Tools: map[string]string{
			"claude": "enabled",
			"cursor": "enabled",
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Save project configuration
	if err := ctx.ConfigManager.SaveProject(config); err != nil {
		return fmt.Errorf("failed to save project configuration: %w", err)
	}

	// Create global wexler source directory structure
	sourcePath, err := config.GetAbsoluteSourcePath()
	if err != nil {
		return fmt.Errorf("failed to resolve source path: %w", err)
	}
	
	if err := os.MkdirAll(sourcePath, 0755); err != nil {
		return fmt.Errorf("failed to create source directory: %w", err)
	}

	// Create subagent directory
	subagentPath := filepath.Join(sourcePath, "subagent")
	if err := os.MkdirAll(subagentPath, 0755); err != nil {
		return fmt.Errorf("failed to create subagent directory: %w", err)
	}

	// Create sample memory.mdc file
	memoryPath := filepath.Join(sourcePath, "memory.mdc")
	if _, err := os.Stat(memoryPath); os.IsNotExist(err) {
		sampleMemory := `# Workflow
Prefer running single tests for performance.

# Code Style
Use Go conventions and direct framework usage.

# Project Context
This is the project context and instructions for AI assistants.`

		if err := os.WriteFile(memoryPath, []byte(sampleMemory), 0644); err != nil {
			return fmt.Errorf("failed to create sample memory file: %w", err)
		}
	}

	// Create sample subagent file
	plannerPath := filepath.Join(subagentPath, "planner.mdc")
	if _, err := os.Stat(plannerPath); os.IsNotExist(err) {
		samplePlanner := `Use this agent when the user asks for planning or task breakdown.

Focus on:
- Breaking down complex tasks into manageable steps
- Identifying dependencies and requirements
- Creating clear, actionable plans`

		if err := os.WriteFile(plannerPath, []byte(samplePlanner), 0644); err != nil {
			return fmt.Errorf("failed to create sample planner file: %w", err)
		}
	}

	fmt.Printf("✅ Wexler initialized successfully!\n\n")
	fmt.Printf("Project: %s (v%s)\n", projectName, initVersion)
	fmt.Printf("Source directory: %s\n", initSourcePath)
	fmt.Printf("Actual source path: %s\n", sourcePath)
	fmt.Printf("\n")
	fmt.Printf("Next steps:\n")
	fmt.Printf("  1. Edit %s/memory.mdc with your AI instructions\n", sourcePath)
	fmt.Printf("  2. Add subagent files to %s/subagent/\n", sourcePath)
	fmt.Printf("  3. Run 'wexler apply --tool=claude' to apply configurations\n")

	return nil
}
