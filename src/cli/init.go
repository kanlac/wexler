package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"mindful/src/config"
	"mindful/src/models"

	"github.com/spf13/cobra"
)

var initForce bool

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialise a Mindful project in the current directory",
		RunE:  runInit,
	}

	cmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing mindful.yaml if present")

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to determine working directory: %w", err)
	}

	mindfulDir := filepath.Join(projectPath, models.DefaultMindfulDirName)
	configPath := filepath.Join(mindfulDir, "mindful.yaml")

	if _, err := os.Stat(configPath); err == nil && !initForce {
		return fmt.Errorf("mindful.yaml already exists; re-run with --force to overwrite")
	}

	if err := os.MkdirAll(mindfulDir, 0o755); err != nil {
		return fmt.Errorf("failed to create mindful directory: %w", err)
	}

	projectName := filepath.Base(projectPath)
	projectConfig := &models.ProjectConfig{
		Name:               projectName,
		Version:            "1.0.0",
		Source:             "~/.mindful",
		EnableCodingAgents: []string{"claude", "cursor", "codex"},
	}

	manager := config.NewManager()
	if err := manager.SaveProject(projectPath, projectConfig); err != nil {
		return fmt.Errorf("failed to write mindful.yaml: %w", err)
	}

	projectMemoryPath := filepath.Join(mindfulDir, "project-memory.mdc")
	if _, err := os.Stat(projectMemoryPath); os.IsNotExist(err) || initForce {
		memoryTemplate := "# Project Memory\n\nDescribe your project-specific context here.\n"
		if err := os.WriteFile(projectMemoryPath, []byte(memoryTemplate), 0o644); err != nil {
			return fmt.Errorf("failed to create %s: %w", projectMemoryPath, err)
		}
	}

	projectSubagentDir := filepath.Join(mindfulDir, "project-subagents")
	if err := os.MkdirAll(projectSubagentDir, 0o755); err != nil {
		return fmt.Errorf("failed to create %s: %w", projectSubagentDir, err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Mindful project initialised in %s\n", mindfulDir)
	return nil
}
