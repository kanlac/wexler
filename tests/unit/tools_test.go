package unit

import (
	"mindful/src/models"
	"mindful/src/tools"
	"strings"
	"testing"
)

func TestCursorFileGeneration(t *testing.T) {
	tests := []struct {
		name     string
		config   *tools.ToolConfig
		expected map[string]string // file path -> expected content pattern
	}{
		{
			name: "cursor memory file with frontmatter",
			config: &tools.ToolConfig{
				ToolName: "cursor",
				Memory: &models.MemoryConfig{
					MindfulMemory: "This is test memory content.",
				},
			},
			expected: map[string]string{
				".cursor/rules/general.mindful.mdc": "---\ndescription: General Memories\nglobs:\nalwaysApply: true\n---\n\nThis is test memory content.",
			},
		},
		{
			name: "cursor subagent with title",
			config: &tools.ToolConfig{
				ToolName: "cursor",
				Subagents: []*models.SubagentConfig{
					{
						Name:    "test-agent",
						Content: "# Test Agent\n\nThis is a test subagent with a title.",
					},
				},
			},
			expected: map[string]string{
				".cursor/rules/test-agent.mindful.mdc": "---\ndescription: Test Agent\nglobs:\nalwaysApply: true\n---\n\n# Test Agent\n\nThis is a test subagent with a title.",
			},
		},
		{
			name: "cursor subagent without title",
			config: &tools.ToolConfig{
				ToolName: "cursor",
				Subagents: []*models.SubagentConfig{
					{
						Name:    "plain-agent",
						Content: "This is a test subagent without any title.",
					},
				},
			},
			expected: map[string]string{
				".cursor/rules/plain-agent.mindful.mdc": "---\ndescription: plain-agent\nglobs:\nalwaysApply: true\n---\n\nThis is a test subagent without any title.",
			},
		},
		{
			name: "cursor complete configuration",
			config: &tools.ToolConfig{
				ToolName: "cursor",
				Memory: &models.MemoryConfig{
					MindfulMemory: "General memory content.",
				},
				Subagents: []*models.SubagentConfig{
					{
						Name:    "frontend",
						Content: "# Frontend Agent\n\nHandles React components.",
					},
					{
						Name:    "backend",
						Content: "Handles Go backend logic.",
					},
				},
				MCP: &models.MCPConfig{
					Servers: map[string]string{
						"context7": "eyJjb21tYW5kIjoicHl0aG9uIn0=", // base64 encoded {"command":"python"}
					},
				},
			},
			expected: map[string]string{
				".cursor/rules/general.mindful.mdc":  "---\ndescription: General Memories\nglobs:\nalwaysApply: true\n---\n\nGeneral memory content.",
				".cursor/rules/frontend.mindful.mdc": "---\ndescription: Frontend Agent\nglobs:\nalwaysApply: true\n---\n\n# Frontend Agent\n\nHandles React components.",
				".cursor/rules/backend.mindful.mdc":  "---\ndescription: backend\nglobs:\nalwaysApply: true\n---\n\nHandles Go backend logic.",
				".cursor/mcp.json":                   "{\n  \"mcpServers\": {\n    \"context7\": {\n      \"command\": \"python\"\n    }\n  }\n}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := tools.NewAdapter("cursor")
			if err != nil {
				t.Fatalf("Failed to create cursor adapter: %v", err)
			}

			files, err := adapter.Generate(tt.config)
			if err != nil {
				t.Fatalf("Failed to generate cursor files: %v", err)
			}

			// Check that we have the expected number of files
			if len(files) != len(tt.expected) {
				t.Errorf("Expected %d files, got %d", len(tt.expected), len(files))
			}

			// Check each expected file
			fileMap := make(map[string]string)
			for _, file := range files {
				fileMap[file.Path] = file.Content
			}

			for expectedPath, expectedContent := range tt.expected {
				actualContent, exists := fileMap[expectedPath]
				if !exists {
					t.Errorf("Expected file %s was not generated", expectedPath)
					continue
				}

				if actualContent != expectedContent {
					t.Errorf("File %s content mismatch:\nExpected:\n%s\n\nActual:\n%s",
						expectedPath, expectedContent, actualContent)
				}
			}
		})
	}
}

func TestExtractDescriptionFromContent(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		fallbackName string
		expected     string
	}{
		{
			name:         "content with level 1 header",
			content:      "# My Agent\n\nThis is my agent description.",
			fallbackName: "fallback",
			expected:     "My Agent",
		},
		{
			name:         "content with multiple headers",
			content:      "# First Header\n\nSome content.\n\n## Second Header\n\nMore content.",
			fallbackName: "fallback",
			expected:     "First Header",
		},
		{
			name:         "content without level 1 header",
			content:      "## Second Level Header\n\nSome content.",
			fallbackName: "my-agent",
			expected:     "my-agent",
		},
		{
			name:         "empty content",
			content:      "",
			fallbackName: "empty-agent",
			expected:     "empty-agent",
		},
		{
			name:         "content with whitespace only header",
			content:      "#   \n\nSome content.",
			fallbackName: "whitespace-agent",
			expected:     "whitespace-agent",
		},
		{
			name:         "content with header containing special characters",
			content:      "# Agent with Special-Characters & Symbols!\n\nDescription.",
			fallbackName: "fallback",
			expected:     "Agent with Special-Characters & Symbols!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := tools.NewAdapter("cursor")
			if err != nil {
				t.Fatalf("Failed to create adapter: %v", err)
			}

			// We need to access the private method, so we'll test it indirectly
			// by checking the generated subagent file content
			config := &tools.ToolConfig{
				ToolName: "cursor",
				Subagents: []*models.SubagentConfig{
					{
						Name:    tt.fallbackName,
						Content: tt.content,
					},
				},
			}

			files, err := adapter.Generate(config)
			if err != nil {
				t.Fatalf("Failed to generate files: %v", err)
			}

			if len(files) != 1 {
				t.Fatalf("Expected 1 file, got %d", len(files))
			}

			content := files[0].Content
			expectedDescLine := "description: " + tt.expected
			if !strings.Contains(content, expectedDescLine) {
				t.Errorf("Expected description '%s' not found in content:\n%s",
					expectedDescLine, content)
			}
		})
	}
}

func TestCursorMemoryContentGeneration(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expected    string
		expectFiles int
	}{
		{
			name:        "basic memory content",
			content:     "This is memory content.",
			expectFiles: 1,
			expected: `---
description: General Memories
globs:
alwaysApply: true
---

This is memory content.`,
		},
		{
			name:        "memory content with whitespace",
			content:     "   \n\nContent with leading/trailing whitespace.\n\n   ",
			expectFiles: 1,
			expected: `---
description: General Memories
globs:
alwaysApply: true
---

Content with leading/trailing whitespace.`,
		},
		{
			name:        "empty memory content",
			content:     "",
			expectFiles: 0, // Empty content doesn't generate files
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := tools.NewAdapter("cursor")
			if err != nil {
				t.Fatalf("Failed to create adapter: %v", err)
			}

			config := &tools.ToolConfig{
				ToolName: "cursor",
				Memory: &models.MemoryConfig{
					MindfulMemory: tt.content,
				},
			}

			files, err := adapter.Generate(config)
			if err != nil {
				t.Fatalf("Failed to generate files: %v", err)
			}

			if len(files) != tt.expectFiles {
				t.Fatalf("Expected %d files, got %d", tt.expectFiles, len(files))
			}

			if tt.expectFiles > 0 && files[0].Content != tt.expected {
				t.Errorf("Content mismatch:\nExpected:\n%s\n\nActual:\n%s",
					tt.expected, files[0].Content)
			}
		})
	}
}
