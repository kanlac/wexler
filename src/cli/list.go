package cli

import (
	"fmt"
	"sort"
	"strings"

	"mindful/src/models"
	"mindful/src/symlink"

	"github.com/spf13/cobra"
)

var listTool string

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show the current state of managed symlinks",
		RunE:  runList,
	}
	cmd.Flags().StringVarP(&listTool, "tool", "t", "", "show symlinks for a specific tool")
	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	ctx, err := NewProjectContext()
	if err != nil {
		return err
	}
	defer ctx.Close()

	manager, err := symlink.NewManager(ctx.ProjectPath, nil)
	if err != nil {
		return err
	}

	tools := collectListTools(ctx, listTool)
	if len(tools) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no symlink mappings available")
		return nil
	}

	for _, tool := range tools {
		infos, err := manager.ListSymlinks(tool)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "%s: %v\n", tool, err)
			continue
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%s:\n", tool)
		if len(infos) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "  (no symlinks configured)")
			continue
		}

		for _, info := range infos {
			status := renderSymlinkStatus(info)
			fmt.Fprintf(cmd.OutOrStdout(), "  %-8s %s -> %s\n", status, info.LinkPath, info.TargetPath)
		}
	}

	return nil
}

func collectListTools(ctx *ProjectContext, selected string) []string {
	cfg, err := symlink.DefaultConfig()
	if err != nil {
		tools := ctx.ProjectConfig.GetEnabledTools()
		sort.Strings(tools)
		return tools
	}
	names := cfg.ToolNames()

	if strings.TrimSpace(selected) != "" {
		filter := strings.TrimSpace(selected)
		if len(names) == 0 {
			return []string{filter}
		}
		for _, name := range names {
			if name == filter {
				return []string{filter}
			}
		}
		return []string{filter}
	}

	if len(names) == 0 {
		tools := ctx.ProjectConfig.GetEnabledTools()
		sort.Strings(tools)
		return tools
	}

	sort.Strings(names)
	return names
}

func renderSymlinkStatus(info models.SymlinkInfo) string {
	if info.IsValid {
		return "ok"
	}
	return "missing"
}
