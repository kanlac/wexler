package integration

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"mindful/src/cli"
	"mindful/src/config"
	"mindful/src/models"
	"mindful/src/symlink"
)

func TestBuildAndApplyPipeline(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation on Windows requires special privileges")
	}

	tempDir := t.TempDir()
	teamDir := filepath.Join(tempDir, "team")
	projectDir := filepath.Join(tempDir, "project")

	if err := os.MkdirAll(filepath.Join(teamDir, "subagents"), 0o755); err != nil {
		t.Fatalf("team setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(teamDir, "memory.mdc"), []byte("Team memory"), 0o644); err != nil {
		t.Fatalf("team memory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(teamDir, "subagents", "shared.mdc"), []byte("Team subagent"), 0o644); err != nil {
		t.Fatalf("team subagent: %v", err)
	}

	mindfulDir := filepath.Join(projectDir, "mindful", "project-subagents")
	if err := os.MkdirAll(mindfulDir, 0o755); err != nil {
		t.Fatalf("project mindful setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(filepath.Dir(mindfulDir), "project-memory.mdc"), []byte("Project memory"), 0o644); err != nil {
		t.Fatalf("project memory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(mindfulDir, "researcher.mdc"), []byte("Project subagent"), 0o644); err != nil {
		t.Fatalf("project subagent: %v", err)
	}

	cfg := &models.ProjectConfig{
		Name:               "demo",
		Version:            "1.0.0",
		Source:             teamDir,
		EnableCodingAgents: []string{"claude"},
	}

	cfgManager := config.NewManager()
	if err := cfgManager.SaveProject(projectDir, cfg); err != nil {
		t.Fatalf("save project config: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(cwd)

	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	ctx, err := cli.NewProjectContext()
	if err != nil {
		t.Fatalf("NewProjectContext: %v", err)
	}
	defer ctx.Close()

	teamSource, err := ctx.ResolveTeamSource()
	if err != nil {
		t.Fatalf("resolve team source: %v", err)
	}
	artifacts, err := ctx.SourceManager.LoadArtifacts(teamSource, ctx.ProjectPath)
	if err != nil {
		t.Fatalf("LoadArtifacts: %v", err)
	}
	if err := ctx.WriteArtifacts(artifacts); err != nil {
		t.Fatalf("WriteArtifacts: %v", err)
	}

	// Provide a dummy MCP configuration so the default mapping can create links.
	mcpPath := filepath.Join(projectDir, "mindful", "out", "mcp.json")
	if err := os.WriteFile(mcpPath, []byte("{}"), 0o644); err != nil {
		t.Fatalf("write mcp.json: %v", err)
	}

	// Ensure artefacts were written
	if _, err := os.Stat(filepath.Join(projectDir, "mindful", "out", "memory.md")); err != nil {
		t.Fatalf("memory artifact missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, "mindful", "out", "subagents", "researcher.mdc")); err != nil {
		t.Fatalf("subagent artifact missing: %v", err)
	}

	manager, err := symlink.NewManager(projectDir, nil)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	if err := manager.CreateSymlinks("claude"); err != nil {
		t.Fatalf("CreateSymlinks: %v", err)
	}

	if err := manager.ValidateSymlinks("claude"); err != nil {
		t.Fatalf("ValidateSymlinks: %v", err)
	}
}
