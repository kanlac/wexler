package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"wexler/src/models"
)

var (
	applyTool   string
	applySource string
)

func newApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply configurations from source to AI tools",
		Long: `Apply source configurations to the specified AI tool(s).

Reads memory.mdc and subagent/*.mdc files from the source directory and generates
appropriate configuration files for the target AI tool. Handles conflict detection
and provides interactive resolution options.`,
		Example: `  # Apply to Claude Code
  wexler apply --tool=claude

  # Apply to Cursor
  wexler apply --tool=cursor

  # Apply to all tools (dry run)
  wexler apply --dry-run

  # Apply with custom source directory
  wexler apply --tool=claude --source=./custom-source`,
		RunE: runApply,
	}

	cmd.Flags().StringVarP(&applyTool, "tool", "t", "", "target tool (claude, cursor, or 'all')")
	cmd.Flags().StringVar(&applySource, "source", "", "source directory (default: from wexler.yaml)")

	return cmd
}

func runApply(cmd *cobra.Command, args []string) error {
	ctx, err := NewAppContext()
	if err != nil {
		return err
	}
	defer ctx.CloseResources()

	// Load project configuration
	projectConfig, err := ctx.ConfigManager.LoadProject(ctx.ProjectPath)
	if err != nil {
		return fmt.Errorf("failed to load project configuration: %w", err)
	}

	// Determine source path
	sourcePath := applySource
	if sourcePath == "" {
		sourcePath = projectConfig.GetAbsoluteSourcePath(ctx.ProjectPath)
	} else {
		sourcePath = filepath.Join(ctx.ProjectPath, sourcePath)
	}

	if verbose {
		fmt.Printf("Loading source configurations from %s\n", sourcePath)
	}

	// Load source configuration
	sourceConfig, err := ctx.SourceManager.LoadSource(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to load source configuration: %w", err)
	}

	// Load MCP configuration from storage
	storageManager, err := ctx.GetStorageManager()
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	mcpConfigs, err := storageManager.ListMCP()
	if err != nil {
		return fmt.Errorf("failed to load MCP configurations: %w", err)
	}

	mcpConfig := models.NewMCPConfig()
	for serverName, config := range mcpConfigs {
		mcpConfig.Servers[serverName] = config
	}

	// Determine target tools
	var tools []string
	if applyTool == "" || applyTool == "all" {
		// Apply to all enabled tools
		for toolName, status := range projectConfig.Tools {
			if status == "enabled" {
				tools = append(tools, toolName)
			}
		}
	} else {
		tools = strings.Split(applyTool, ",")
	}

	if len(tools) == 0 {
		return fmt.Errorf("no tools specified or enabled")
	}

	// Apply to each tool
	for _, toolName := range tools {
		toolName = strings.TrimSpace(toolName)
		
		if verbose {
			fmt.Printf("Applying configuration to %s...\n", toolName)
		}

		// Create apply configuration
		applyConfig := &models.ApplyConfig{
			ProjectPath: ctx.ProjectPath,
			ToolName:    toolName,
			Source:      sourceConfig,
			MCP:         mcpConfig,
			DryRun:      dryRun,
			Force:       force,
		}

		// Check for conflicts first (unless force is enabled)
		if !force {
			conflicts, err := ctx.ApplyManager.DetectConflicts(applyConfig)
			if err != nil {
				return fmt.Errorf("failed to detect conflicts for %s: %w", toolName, err)
			}

			if len(conflicts) > 0 {
				fmt.Printf("⚠️  Found %d conflict(s) for %s:\n", len(conflicts), toolName)
				for _, conflict := range conflicts {
					fmt.Printf("  - %s (%s)\n", conflict.FilePath, conflict.FileType)
				}

				if !dryRun {
					// In real implementation, this would prompt user for resolution
					// For now, we'll stop on conflicts
					return fmt.Errorf("conflicts detected for %s (use --force to overwrite)", toolName)
				}
			}
		}

		// Apply the configuration
		result, err := ctx.ApplyManager.ApplyConfig(applyConfig)
		if err != nil {
			return fmt.Errorf("failed to apply configuration to %s: %w", toolName, err)
		}

		// Display results
		if result.Success {
			fmt.Printf("✅ Successfully applied configuration to %s\n", toolName)
			if verbose {
				fmt.Printf("   Files written: %d\n", len(result.FilesWritten))
				fmt.Printf("   Files skipped: %d\n", len(result.FilesSkipped))
				if len(result.FilesWritten) > 0 {
					fmt.Printf("   Written files:\n")
					for _, file := range result.FilesWritten {
						fmt.Printf("     - %s\n", file)
					}
				}
			}
		} else {
			fmt.Printf("❌ Failed to apply configuration to %s: %s\n", toolName, result.Error)
		}
	}

	if dryRun {
		fmt.Printf("\n🔍 Dry run completed - no files were modified\n")
	}

	return nil
}