package cli

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"mindful/src/models"
	"mindful/src/symlink"

	"github.com/spf13/cobra"
)

var (
	applyTools     string
	applySkipBuild bool
	applyDryRun    bool
)

func newApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Build artefacts and create symlinks for enabled tools",
		RunE:  runApply,
	}

	cmd.Flags().StringVarP(&applyTools, "tool", "t", "", "comma separated list of tools to target (defaults to enabled tools)")
	cmd.Flags().BoolVar(&applySkipBuild, "skip-build", false, "skip automatic build before applying symlinks")
	cmd.Flags().BoolVar(&applyDryRun, "dry-run", false, "plan symlink changes without modifying the filesystem")

	return cmd
}

func runApply(cmd *cobra.Command, args []string) error {
	ctx, err := NewProjectContext()
	if err != nil {
		return err
	}
	defer ctx.Close()

	if !applySkipBuild {
		if _, err := executeBuild(ctx); err != nil {
			return fmt.Errorf("build failed: %w", err)
		}
	}

	tools, err := resolveTargetTools(ctx.ProjectConfig, applyTools)
	if err != nil {
		return err
	}

	manager, err := symlink.NewManager(ctx.ProjectPath, nil)
	if err != nil {
		return err
	}

	var toolErrs []error
	for _, tool := range tools {
		if applyDryRun {
			if err := reportPlannedSymlinks(cmd, manager, tool); err != nil {
				toolErrs = append(toolErrs, err)
			}
			continue
		}

		if err := manager.CreateSymlinks(tool); err != nil {
			toolErrs = append(toolErrs, fmt.Errorf("%s: %w", tool, err))
			fmt.Fprintf(cmd.ErrOrStderr(), "✗ %s: %v\n", tool, err)
			continue
		}

		fmt.Fprintf(cmd.OutOrStdout(), "✓ %s symlinks updated\n", tool)
	}

	return errors.Join(toolErrs...)
}

func resolveTargetTools(cfg *models.ProjectConfig, selection string) ([]string, error) {
	var tools []string
	if strings.TrimSpace(selection) != "" {
		for _, part := range strings.Split(selection, ",") {
			name := strings.TrimSpace(part)
			if name != "" {
				tools = append(tools, name)
			}
		}
	} else {
		tools = cfg.GetEnabledTools()
	}

	if len(tools) == 0 {
		return nil, fmt.Errorf("no tools specified or enabled in mindful.yaml")
	}

	unique := make(map[string]struct{})
	for _, tool := range tools {
		unique[tool] = struct{}{}
	}

	final := make([]string, 0, len(unique))
	for tool := range unique {
		final = append(final, tool)
	}
	sort.Strings(final)

	return final, nil
}

func reportPlannedSymlinks(cmd *cobra.Command, manager *symlink.Manager, tool string) error {
	infos, err := manager.PlanSymlinks(tool)
	if err != nil {
		return fmt.Errorf("dry-run failed for %s: %w", tool, err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%s:\n", tool)
	if len(infos) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "  (no symlinks configured)")
		return nil
	}

	for _, info := range infos {
		status := "create"
		if info.IsValid {
			status = "ok"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  %-7s %s -> %s\n", status, info.LinkPath, info.TargetPath)
	}

	return nil
}
