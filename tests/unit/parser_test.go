package unit

import (
	"mindful/src/source"
	"reflect"
	"testing"
)

func TestParseMarkdownSections(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    map[string]string
		wantErr bool
	}{
		{
			name: "single section",
			content: `# Workflow
Prefer running single tests for performance.`,
			want: map[string]string{
				"Workflow": "Prefer running single tests for performance.",
			},
			wantErr: false,
		},
		{
			name: "multiple sections",
			content: `# Workflow
Prefer running single tests.

# Code Style
Use Go conventions.

# Context
This is the project context.`,
			want: map[string]string{
				"Workflow":   "Prefer running single tests.",
				"Code Style": "Use Go conventions.",
				"Context":    "This is the project context.",
			},
			wantErr: false,
		},
		{
			name: "sections with empty lines",
			content: `# Section 1

Content with empty line above.

# Section 2


Content with multiple empty lines above.`,
			want: map[string]string{
				"Section 1": "\nContent with empty line above.",
				"Section 2": "\n\nContent with multiple empty lines above.",
			},
			wantErr: false,
		},
		{
			name: "content before first header",
			content: `This content should be ignored.

# First Section
This content should be included.`,
			want: map[string]string{
				"First Section": "This content should be included.",
			},
			wantErr: false,
		},
		{
			name:    "empty content",
			content: "",
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name:    "only whitespace",
			content: "   \n\t  \n  ",
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name: "no sections just content",
			content: `This is some content without any headers.
It should result in an empty map.`,
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name: "section with no content",
			content: `# Empty Section
# Another Section
Some content here.`,
			want: map[string]string{
				"Empty Section":   "",
				"Another Section": "Some content here.",
			},
			wantErr: false,
		},
		{
			name: "sections with special characters",
			content: `# Section with Symbols !@#$%
Content for special section.

# 数字和中文
Chinese content.

# Section-With-Dashes_And_Underscores
Mixed content.`,
			want: map[string]string{
				"Section with Symbols !@#$%":          "Content for special section.",
				"数字和中文":                               "Chinese content.",
				"Section-With-Dashes_And_Underscores": "Mixed content.",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := source.ParseMarkdownSections(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMarkdownSections() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseMarkdownSections() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReconstructMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		sections map[string]string
		want     string
	}{
		{
			name: "single section",
			sections: map[string]string{
				"Workflow": "Prefer running single tests.",
			},
			want: "# Workflow\nPrefer running single tests.",
		},
		{
			name:     "empty sections",
			sections: map[string]string{},
			want:     "",
		},
		{
			name: "multiple sections",
			sections: map[string]string{
				"Workflow":   "Prefer running single tests.",
				"Code Style": "Use Go conventions.",
			},
			// Note: map iteration order is not guaranteed, so we need to check both possibilities
		},
		{
			name: "sections with empty content",
			sections: map[string]string{
				"Empty Section": "",
				"Full Section":  "Some content",
			},
		},
		{
			name: "sections with empty names should be skipped",
			sections: map[string]string{
				"":              "This should be skipped",
				"Valid Section": "This should be included",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := source.ReconstructMarkdown(tt.sections)

			// For single section or empty, we can do exact match
			if len(tt.sections) <= 1 {
				if got != tt.want {
					t.Errorf("ReconstructMarkdown() = %q, want %q", got, tt.want)
				}
				return
			}

			// For multiple sections, verify all sections are present
			for sectionName, content := range tt.sections {
				if sectionName == "" || content == "" {
					continue // These should be skipped
				}
				expectedSection := "# " + sectionName + "\n" + content
				if !contains(got, expectedSection) {
					t.Errorf("ReconstructMarkdown() missing section %q in result %q", expectedSection, got)
				}
			}
		})
	}
}

func TestSanitizeContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "clean content",
			content: "This is clean content.",
			want:    "This is clean content.",
		},
		{
			name:    "content with null bytes",
			content: "Content\x00with\x00null\x00bytes",
			want:    "Contentwithnullbytes",
		},
		{
			name:    "content with windows line endings",
			content: "Line 1\r\nLine 2\r\nLine 3",
			want:    "Line 1\nLine 2\nLine 3",
		},
		{
			name:    "content with mac line endings",
			content: "Line 1\rLine 2\rLine 3",
			want:    "Line 1\nLine 2\nLine 3",
		},
		{
			name:    "mixed line endings",
			content: "Line 1\r\nLine 2\rLine 3\nLine 4",
			want:    "Line 1\nLine 2\nLine 3\nLine 4",
		},
		{
			name:    "empty content",
			content: "",
			want:    "",
		},
		{
			name:    "content with both issues",
			content: "Line 1\x00\r\nLine 2\x00\rLine 3",
			want:    "Line 1\nLine 2\nLine 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := source.SanitizeContent(tt.content)
			if got != tt.want {
				t.Errorf("SanitizeContent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateSubagentContent(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		agentName string
		wantErr   bool
	}{
		{
			name:      "valid subagent content",
			content:   "This is valid subagent content with instructions.",
			agentName: "planner",
			wantErr:   false,
		},
		{
			name:      "empty content is valid for subagents",
			content:   "",
			agentName: "empty-agent",
			wantErr:   false,
		},
		{
			name:      "empty agent name",
			content:   "Some content",
			agentName: "",
			wantErr:   true,
		},
		{
			name:      "very long content should be rejected",
			content:   generateLongString(2 * 1024 * 1024), // 2MB
			agentName: "large-agent",
			wantErr:   true,
		},
		{
			name:      "content with special characters",
			content:   "Content with émojis 🎉 and spëcial characters åäö",
			agentName: "special-agent",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := source.ValidateSubagentContent(tt.content, tt.agentName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSubagentContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Helper function to generate a long string for testing
func generateLongString(length int) string {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = 'a'
	}
	return string(result)
}
