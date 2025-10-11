package cli

import (
	"errors"
	"fmt"
	"os"

	"mindful/src/models"

	"github.com/spf13/cobra"
)

func newBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Render mindful/out artefacts from project sources",
		RunE:  runBuild,
	}
	return cmd
}

func runBuild(cmd *cobra.Command, args []string) error {
	ctx, err := NewProjectContext()
	if err != nil {
		return err
	}
	defer ctx.Close()

	artifacts, err := executeBuild(ctx)
	if err != nil {
		return err
	}

	if verboseFlag {
		subagentCount := 0
		if artifacts != nil {
			subagentCount = len(artifacts.Subagents)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "mindful/out refreshed (subagents: %d)\n", subagentCount)
	}

	return nil
}

func executeBuild(ctx *ProjectContext) (*models.BuildArtifacts, error) {
	if ctx == nil {
		return nil, errors.New("project context cannot be nil")
	}

	teamSource, err := ctx.ResolveTeamSource()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve team source: %w", err)
	}

	artifacts, err := ctx.SourceManager.LoadArtifacts(teamSource, ctx.ProjectPath)
	if err != nil {
		return nil, err
	}
	if artifacts == nil {
		artifacts = &models.BuildArtifacts{}
	}

	mcpContent, err := loadMCPContent(ctx)
	if err != nil {
		return nil, err
	}
	artifacts.MCPContent = mcpContent

	if err := ctx.WriteArtifacts(artifacts); err != nil {
		return nil, err
	}

	return artifacts, nil
}

func loadMCPContent(ctx *ProjectContext) ([]byte, error) {
	storageManager, err := ctx.GetStorageManager()
	if err != nil {
		// Treat absence of storage as a non-fatal error when the directory does not exist yet.
		var pathErr *os.PathError
		if errors.As(err, &pathErr) && os.IsNotExist(pathErr.Err) {
			return nil, nil
		}
		return nil, err
	}

	records, err := storageManager.ListMCP()
	if err != nil {
		return nil, fmt.Errorf("failed to list MCP configurations: %w", err)
	}

	if len(records) == 0 {
		return nil, nil
	}

	cfg := models.NewMCPConfig()
	for name, encoded := range records {
		cfg.Servers[name] = encoded
	}

	data, err := cfg.ToMCPJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to render MCP configuration: %w", err)
	}

	return data, nil
}
