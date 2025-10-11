package unit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mindful/src/source"
)

func TestLoadArtifactsCombinesTeamAndProject(t *testing.T) {
	tempDir := t.TempDir()
	teamDir := filepath.Join(tempDir, "team")
	projectDir := filepath.Join(tempDir, "project")

	if err := os.MkdirAll(teamDir, 0o755); err != nil {
		t.Fatalf("team dir: %v", err)
	}
	mindfulDir := filepath.Join(projectDir, "mindful")
	if err := os.MkdirAll(filepath.Join(mindfulDir, "project-subagents"), 0o755); err != nil {
		t.Fatalf("mindful dir: %v", err)
	}

	teamMemory := "# Team Notes\nTeam scope content"
	if err := os.WriteFile(filepath.Join(teamDir, "memory.mdc"), []byte(teamMemory), 0o644); err != nil {
		t.Fatalf("write team memory: %v", err)
	}

	projectMemory := "# Project Notes\nProject scope content"
	if err := os.WriteFile(filepath.Join(mindfulDir, "project-memory.mdc"), []byte(projectMemory), 0o644); err != nil {
		t.Fatalf("write project memory: %v", err)
	}

	// Team subagent overridden by project
	if err := os.MkdirAll(filepath.Join(teamDir, "subagents"), 0o755); err != nil {
		t.Fatalf("team subagents dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(teamDir, "subagents", "researcher.mdc"), []byte("Team researcher"), 0o644); err != nil {
		t.Fatalf("write team subagent: %v", err)
	}

	if err := os.WriteFile(filepath.Join(mindfulDir, "project-subagents", "researcher.mdc"), []byte("Project researcher"), 0o644); err != nil {
		t.Fatalf("write project subagent: %v", err)
	}

	mgr := source.NewManager()
	artifacts, err := mgr.LoadArtifacts(teamDir, projectDir)
	if err != nil {
		t.Fatalf("LoadArtifacts error: %v", err)
	}

	if artifacts.Memory == nil {
		t.Fatalf("expected memory artifact")
	}

	memory := artifacts.Memory.Content
	if !strings.Contains(memory, "scope:team") || !strings.Contains(memory, "Team scope content") {
		t.Errorf("memory should include team section, got %q", memory)
	}
	if !strings.Contains(memory, "scope:project") || !strings.Contains(memory, "Project scope content") {
		t.Errorf("memory should include project section, got %q", memory)
	}

	if len(artifacts.Subagents) != 1 {
		t.Fatalf("expected 1 subagent, got %d", len(artifacts.Subagents))
	}

	subagent := artifacts.Subagents[0]
	if subagent.Name != "researcher" {
		t.Fatalf("unexpected subagent name %q", subagent.Name)
	}
	if !strings.Contains(subagent.Content, "scope:project") {
		t.Errorf("subagent should use project scope annotation: %q", subagent.Content)
	}
	if !strings.Contains(subagent.Content, "Project researcher") {
		t.Errorf("subagent content mismatch: %q", subagent.Content)
	}
}
