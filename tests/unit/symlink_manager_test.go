package unit

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"mindful/src/models"
	"mindful/src/symlink"
)

func TestSymlinkManagerCreateAndValidate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation on Windows requires special privileges")
	}

	projectDir := t.TempDir()
	mindfulOut := filepath.Join(projectDir, "mindful", "out")
	if err := os.MkdirAll(filepath.Join(mindfulOut, "subagents"), 0o755); err != nil {
		t.Fatalf("create out dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(mindfulOut, "memory.md"), []byte("memory"), 0o644); err != nil {
		t.Fatalf("write memory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(mindfulOut, "mcp.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("write mcp: %v", err)
	}
	if err := os.WriteFile(filepath.Join(mindfulOut, "subagents", "researcher.mdc"), []byte("agent"), 0o644); err != nil {
		t.Fatalf("write subagent: %v", err)
	}

	config := models.NewSymlinkConfig(map[string]*models.ToolSymlinkConfig{
		"claude": {
			Memory:    "CLAUDE.md",
			Subagents: ".claude/{name}.mdc",
			MCP:       ".mcp.json",
		},
	})

	manager, err := symlink.NewManager(projectDir, config)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	plans, err := manager.PlanSymlinks("claude")
	if err != nil {
		t.Fatalf("PlanSymlinks error: %v", err)
	}
	if len(plans) != 3 {
		t.Fatalf("expected 3 symlinks, got %d", len(plans))
	}

	if err := manager.CreateSymlinks("claude"); err != nil {
		t.Fatalf("CreateSymlinks error: %v", err)
	}

	for _, link := range []string{"CLAUDE.md", ".mcp.json", filepath.Join(".claude", "researcher.mdc")} {
		path := filepath.Join(projectDir, link)
		if info, err := os.Lstat(path); err != nil {
			t.Fatalf("expected symlink %s: %v", path, err)
		} else if info.Mode()&os.ModeSymlink == 0 {
			t.Fatalf("expected %s to be symlink", path)
		}
	}

	if err := manager.ValidateSymlinks("claude"); err != nil {
		t.Fatalf("ValidateSymlinks error: %v", err)
	}

	if err := manager.CleanupSymlinks("claude"); err != nil {
		t.Fatalf("CleanupSymlinks error: %v", err)
	}

	if _, err := os.Lstat(filepath.Join(projectDir, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Fatalf("expected CLAUDE.md to be removed, err=%v", err)
	}
}
