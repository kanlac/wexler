package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"mindful/src/models"

	"github.com/spf13/cobra"
)

var (
	importTool string
)

func newImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import existing AI tool configurations to central storage",
		Long: `Import existing AI tool configurations into Mindful's central storage.

Scans the project directory for existing AI tool configuration files (like .mcp.json)
and imports them into Mindful's secure storage for unified management.`,
		Example: `  # Import from Claude Code
  mindful import --tool=claude

  # Import from Cursor
  mindful import --tool=cursor

  # Import from all tools with dry run
  mindful import --dry-run`,
		RunE: runImport,
	}

	cmd.Flags().StringVarP(&importTool, "tool", "t", "", "source tool to import from (claude, cursor)")

	return cmd
}

func runImport(cmd *cobra.Command, args []string) error {
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

	// Initialize storage manager
	storageManager, err := ctx.GetStorageManager()
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Determine which tools to import from
	tools := []string{}
	if importTool != "" {
		tools = append(tools, importTool)
	} else {
		// Import from all enabled tools
		for toolName, status := range projectConfig.Tools {
			if status == "enabled" {
				tools = append(tools, toolName)
			}
		}
	}

	importCount := 0

	for _, toolName := range tools {
		if verbose {
			fmt.Printf("Importing configurations from %s...\n", toolName)
		}

		// Define tool-specific paths
		var mcpPaths []string
		switch toolName {
		case "claude":
			mcpPaths = []string{".mcp.json"}
		case "cursor":
			mcpPaths = []string{".cursor/mcp.json"}
		default:
			fmt.Printf("⚠️  Unknown tool: %s\n", toolName)
			continue
		}

		// Import MCP configurations
		for _, mcpPath := range mcpPaths {
			fullPath := filepath.Join(ctx.ProjectPath, mcpPath)

			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				if verbose {
					fmt.Printf("   No MCP config found at %s\n", mcpPath)
				}
				continue
			}

			if verbose {
				fmt.Printf("   Found MCP config at %s\n", mcpPath)
			}

			// Read MCP configuration
			data, err := os.ReadFile(fullPath)
			if err != nil {
				fmt.Printf("   ⚠️  Failed to read %s: %v\n", mcpPath, err)
				continue
			}

			// Parse and import MCP servers
			mcpConfig, err := models.FromMCPJSON(data)
			if err != nil {
				fmt.Printf("   ⚠️  Failed to parse %s: %v\n", mcpPath, err)
				continue
			}

			// Store each server configuration
			for _, serverName := range mcpConfig.ListServers() {
				if dryRun {
					fmt.Printf("   Would import MCP server: %s\n", serverName)
					importCount++
				} else {
					encoded := mcpConfig.Servers[serverName]
					if err := storageManager.StoreMCP(serverName, encoded); err != nil {
						fmt.Printf("   ⚠️  Failed to import server %s: %v\n", serverName, err)
						continue
					}
					fmt.Printf("   ✅ Imported MCP server: %s\n", serverName)
					importCount++
				}
			}
		}

		// Import memory configurations (reverse apply)
		// This would read existing tool configs and extract memory content
		// For now, we'll just report what would be imported
		switch toolName {
		case "claude":
			claudePath := filepath.Join(ctx.ProjectPath, "CLAUDE.md")
			if _, err := os.Stat(claudePath); err == nil {
				if dryRun {
					fmt.Printf("   Would import memory config from CLAUDE.md\n")
				} else {
					fmt.Printf("   ℹ️  Found CLAUDE.md (manual import to source/memory.mdc recommended)\n")
				}
			}
		case "cursor":
			cursorPath := filepath.Join(ctx.ProjectPath, ".cursor/rules")
			if _, err := os.Stat(cursorPath); err == nil {
				if dryRun {
					fmt.Printf("   Would import memory config from .cursor/rules/\n")
				} else {
					fmt.Printf("   ℹ️  Found Cursor rules (manual import to source/ recommended)\n")
				}
			}
		}
	}

	if importCount > 0 {
		if dryRun {
			fmt.Printf("\n🔍 Dry run completed - would import %d configuration(s)\n", importCount)
		} else {
			fmt.Printf("\n✅ Successfully imported %d configuration(s)\n", importCount)
			fmt.Printf("\nNext steps:\n")
			fmt.Printf("  1. Review imported configurations with 'mindful list'\n")
			fmt.Printf("  2. Apply configurations with 'mindful apply'\n")
		}
	} else {
		fmt.Printf("ℹ️  No configurations found to import\n")
	}

	return nil
}
