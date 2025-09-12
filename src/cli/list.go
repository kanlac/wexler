package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	listFormat string
	listMCP    bool
	listTools  bool
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all managed configurations",
		Long: `List all configurations managed by Wexler.

Shows project configuration, source files, MCP server configurations,
and tool status in various output formats.`,
		Example: `  # List all configurations
  wexler list

  # List only MCP configurations
  wexler list --mcp

  # List only tool configurations
  wexler list --tools

  # Output in JSON format
  wexler list --format=json`,
		RunE: runList,
	}

	cmd.Flags().StringVar(&listFormat, "format", "table", "output format (table, json, yaml)")
	cmd.Flags().BoolVar(&listMCP, "mcp", false, "list only MCP configurations")
	cmd.Flags().BoolVar(&listTools, "tools", false, "list only tool configurations")

	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
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

	// Initialize tabwriter for table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Show project information (unless filtering)
	if !listMCP && !listTools {
		fmt.Fprintf(w, "PROJECT CONFIGURATION\n")
		fmt.Fprintf(w, "Name:\t%s\n", projectConfig.Name)
		fmt.Fprintf(w, "Version:\t%s\n", projectConfig.Version)
		fmt.Fprintf(w, "Source Path:\t%s\n", projectConfig.SourcePath)
		fmt.Fprintf(w, "\n")
	}

	// Show source files (unless filtering for MCP/tools only)
	if !listMCP && !listTools {
		sourcePath, err := projectConfig.GetAbsoluteSourcePath()
		if err != nil {
			fmt.Printf("⚠️  Failed to resolve source path: %v\n", err)
		} else {
			sourceFiles, err := ctx.SourceManager.ListSourceFiles(sourcePath)
			if err != nil {
				fmt.Printf("⚠️  Failed to list source files: %v\n", err)
			} else {
				fmt.Fprintf(w, "SOURCE FILES\n")
				if len(sourceFiles) == 0 {
					fmt.Fprintf(w, "(none)\n")
				} else {
					for _, file := range sourceFiles {
						// Make path relative to source path
						relPath := strings.TrimPrefix(file, sourcePath)
						relPath = strings.TrimPrefix(relPath, "/")
						fmt.Fprintf(w, "%s\n", relPath)
					}
				}
				fmt.Fprintf(w, "\n")
			}
		}
	}

	// Show MCP configurations
	if !listTools {
		storageManager, err := ctx.GetStorageManager()
		if err != nil {
			fmt.Printf("⚠️  Failed to initialize storage: %v\n", err)
		} else {
			mcpConfigs, err := storageManager.ListMCP()
			if err != nil {
				fmt.Printf("⚠️  Failed to list MCP configurations: %v\n", err)
			} else {
				fmt.Fprintf(w, "MCP SERVERS\n")
				if len(mcpConfigs) == 0 {
					fmt.Fprintf(w, "(none)\n")
				} else {
					fmt.Fprintf(w, "Name\tStatus\n")
					fmt.Fprintf(w, "----\t------\n")
					for serverName := range mcpConfigs {
						fmt.Fprintf(w, "%s\tstored\n", serverName)
					}
				}
				fmt.Fprintf(w, "\n")
			}
		}
	}

	// Show tool configurations
	if !listMCP {
		fmt.Fprintf(w, "AI TOOLS\n")
		fmt.Fprintf(w, "Tool\tStatus\tConfiguration\n")
		fmt.Fprintf(w, "----\t------\t-------------\n")
		for toolName, status := range projectConfig.Tools {
			configStatus := "not applied"
			
			// Check if tool has been configured
			switch toolName {
			case "claude":
				if _, err := os.Stat("CLAUDE.md"); err == nil {
					configStatus = "configured"
				}
			case "cursor":
				if _, err := os.Stat(".cursor/rules"); err == nil {
					configStatus = "configured"
				}
			}
			
			fmt.Fprintf(w, "%s\t%s\t%s\n", toolName, status, configStatus)
		}
		fmt.Fprintf(w, "\n")
	}

	// Flush tabwriter
	w.Flush()

	// Show summary (unless filtering)
	if !listMCP && !listTools {
		enabledTools := 0
		for _, status := range projectConfig.Tools {
			if status == "enabled" {
				enabledTools++
			}
		}

		fmt.Printf("Summary: %d enabled tool(s), project '%s' v%s\n", 
			enabledTools, projectConfig.Name, projectConfig.Version)
	}

	return nil
}