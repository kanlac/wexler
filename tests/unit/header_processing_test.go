package unit

import (
	"strings"
	"testing"

	"mindful/src/tools/common"
)

func TestProcessMemoryContent_NoHeaders(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		scope    string
		source   string
		expected string
	}{
		{
			name:    "simple content without headers",
			content: "This is simple configuration content.\n\nMore details here.",
			scope:   "team",
			source:  "/team/memory.mdc",
			expected: "# Mindful (scope:team)\n<!-- Source: /team/memory.mdc -->\n\nThis is simple configuration content.\n\nMore details here.",
		},
		{
			name:    "content with level-2 headers only",
			content: "## Code Style\nUse Go conventions.\n\n## Workflow\nPrefer TDD approach.",
			scope:   "project",
			source:  "mindful/memory.mdc",
			expected: "# Mindful (scope:project)\n<!-- Source: mindful/memory.mdc -->\n\n## Code Style\nUse Go conventions.\n\n## Workflow\nPrefer TDD approach.",
		},
		{
			name:     "empty content",
			content:  "",
			scope:    "team",
			source:   "/team/memory.mdc",
			expected: "",
		},
		{
			name:     "whitespace only content",
			content:  "   \n\t  \n  ",
			scope:    "team",
			source:   "/team/memory.mdc",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := common.ProcessMemoryContent(tt.content, tt.scope, tt.source)
			if result != tt.expected {
				t.Errorf("ProcessMemoryContent() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestProcessMemoryContent_WithHeaders(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		scope    string
		source   string
		expected string
	}{
		{
			name:    "single level-1 header",
			content: "# Team Development Guidelines\n\n## Code Style\nUse Go conventions.",
			scope:   "team",
			source:  "/team/memory.mdc",
			expected: "# Team Development Guidelines -- Mindful (scope:team)\n<!-- Source: /team/memory.mdc -->\n\n## Code Style\nUse Go conventions.",
		},
		{
			name:    "multiple level-1 headers",
			content: "# Code Standards\nGo conventions here.\n\n# Deployment Process\nUse Docker containers.\n\n## Details\nMore info.",
			scope:   "project",
			source:  "mindful/memory.mdc",
			expected: "# Code Standards -- Mindful (scope:project)\n<!-- Source: mindful/memory.mdc -->\n\nGo conventions here.\n\n# Deployment Process -- Mindful (scope:project)\n<!-- Source: mindful/memory.mdc -->\n\nUse Docker containers.\n\n## Details\nMore info.",
		},
		{
			name:    "level-1 header with complex content",
			content: "# 团队配置规范\n\n这是团队的配置规范内容。\n\n## 代码风格\n使用 Go 约定。\n\n### 具体要求\n- 使用 gofmt\n- 遵循命名约定",
			scope:   "team",
			source:  "/external/source/memory.mdc",
			expected: "# 团队配置规范 -- Mindful (scope:team)\n<!-- Source: /external/source/memory.mdc -->\n\n这是团队的配置规范内容。\n\n## 代码风格\n使用 Go 约定。\n\n### 具体要求\n- 使用 gofmt\n- 遵循命名约定",
		},
		{
			name:    "level-1 header at the end",
			content: "Some introduction text.\n\n## Section A\nContent for section A.\n\n# Main Configuration\nThis is the main config.",
			scope:   "team",
			source:  "/team/memory.mdc",
			expected: "Some introduction text.\n\n## Section A\nContent for section A.\n\n# Main Configuration -- Mindful (scope:team)\n<!-- Source: /team/memory.mdc -->\n\nThis is the main config.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := common.ProcessMemoryContent(tt.content, tt.scope, tt.source)
			if result != tt.expected {
				t.Errorf("ProcessMemoryContent() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestProcessMemoryContent_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		scope    string
		source   string
		expected string
	}{
		{
			name:    "header with just hash symbol",
			content: "#\nSome content here.",
			scope:   "team",
			source:  "/team/memory.mdc",
			expected: "# Mindful (scope:team)\n<!-- Source: /team/memory.mdc -->\n\n#\nSome content here.",
		},
		{
			name:    "header with only spaces after hash",
			content: "#   \nSome content here.",
			scope:   "team",
			source:  "/team/memory.mdc",
			expected: "# Mindful (scope:team)\n<!-- Source: /team/memory.mdc -->\n\n#   \nSome content here.",
		},
		{
			name:    "fake header in middle of line",
			content: "This line contains # symbol but is not a header.\n\n# Real Header\nReal header content.",
			scope:   "project",
			source:  "mindful/memory.mdc",
			expected: "This line contains # symbol but is not a header.\n\n# Real Header -- Mindful (scope:project)\n<!-- Source: mindful/memory.mdc -->\n\nReal header content.",
		},
		{
			name:    "mixed level headers",
			content: "## Level 2\nContent.\n\n# Level 1\nMore content.\n\n### Level 3\nEven more content.",
			scope:   "team",
			source:  "/team/memory.mdc",
			expected: "## Level 2\nContent.\n\n# Level 1 -- Mindful (scope:team)\n<!-- Source: /team/memory.mdc -->\n\nMore content.\n\n### Level 3\nEven more content.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := common.ProcessMemoryContent(tt.content, tt.scope, tt.source)
			if result != tt.expected {
				t.Errorf("ProcessMemoryContent() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestProcessMemoryContent_Integration(t *testing.T) {
	// Test that simulates real dual-scope scenario
	teamContent := "# Team Standards\n\nUse Go conventions for all projects.\n\n## Testing\nAlways write tests first."
	projectContent := "## Project Setup\nThis project uses Docker.\n\n## Local Development\nRun with make dev."

	teamResult := common.ProcessMemoryContent(teamContent, "team", "/external/source/memory.mdc")
	projectResult := common.ProcessMemoryContent(projectContent, "project", "mindful/memory.mdc")

	// Verify team result (has level-1 header)
	expectedTeam := "# Team Standards -- Mindful (scope:team)\n<!-- Source: /external/source/memory.mdc -->\n\nUse Go conventions for all projects.\n\n## Testing\nAlways write tests first."
	if teamResult != expectedTeam {
		t.Errorf("Team result = %q, want %q", teamResult, expectedTeam)
	}

	// Verify project result (no level-1 header)
	expectedProject := "# Mindful (scope:project)\n<!-- Source: mindful/memory.mdc -->\n\n## Project Setup\nThis project uses Docker.\n\n## Local Development\nRun with make dev."
	if projectResult != expectedProject {
		t.Errorf("Project result = %q, want %q", projectResult, expectedProject)
	}

	// Verify combined result structure
	combined := teamResult + "\n\n" + projectResult
	lines := strings.Split(combined, "\n")

	// Should have proper structure
	if !strings.Contains(lines[0], "Team Standards -- Mindful (scope:team)") {
		t.Error("Combined result missing team header with suffix")
	}
	if !strings.Contains(combined, "# Mindful (scope:project)") {
		t.Error("Combined result missing project header")
	}
}