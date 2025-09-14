package unit

import (
	"mindful/src/models"
	"testing"
)

func TestMemoryConfig_ParseMemoryContent(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantMindful string
		wantErr     bool
	}{
		{
			name:        "Empty content",
			content:     "",
			wantMindful: "",
			wantErr:     false,
		},
		{
			name:        "Only whitespace",
			content:     "   \n\t  \n  ",
			wantMindful: "",
			wantErr:     false,
		},
		{
			name: "MINDFUL section only",
			content: `# MINDFUL
This is the MINDFUL section content.

## Workflow
Use TDD approach.

## Code Style
Follow Go conventions.`,
			wantMindful: `This is the MINDFUL section content.

## Workflow
Use TDD approach.

## Code Style
Follow Go conventions.`,
			wantErr: false,
		},
		{
			name: "MINDFUL section with other sections",
			content: `# Introduction
This is the introduction.

# MINDFUL
Main memory content here.

## Guidelines
Follow these guidelines.

# Other Section
This should be ignored.`,
			wantMindful: `Main memory content here.

## Guidelines
Follow these guidelines.`,
			wantErr: false,
		},
		{
			name: "No MINDFUL section",
			content: `# Introduction
Some content here.

# Other Section
More content.`,
			wantMindful: "",
			wantErr:     false,
		},
		{
			name: "MINDFUL section at the end",
			content: `# Introduction
Introduction content.

# MINDFUL
<--- Main memory of all coding agents. Managed by mindful. DO NOT EDIT outside our mindful source directory. --->

## Workflow
Prefer running single tests for performance.

## Code Style
Use Go conventions and direct framework usage.

## Project Context
This is the project context and instructions for AI assistants.`,
			wantMindful: `<--- Main memory of all coding agents. Managed by mindful. DO NOT EDIT outside our mindful source directory. --->

## Workflow
Prefer running single tests for performance.

## Code Style
Use Go conventions and direct framework usage.

## Project Context
This is the project context and instructions for AI assistants.`,
			wantErr: false,
		},
		{
			name: "Multiple MINDFUL sections (first one wins)",
			content: `# MINDFUL
First MINDFUL section.

# Other
Other content.

# MINDFUL
Second MINDFUL section.`,
			wantMindful: "First MINDFUL section.",
			wantErr:     false,
		},
		{
			name: "MINDFUL section with trailing whitespace",
			content: `# MINDFUL
Content with spaces.   

More content.  
`,
			wantMindful: `Content with spaces.   

More content.`,
			wantErr: false,
		},
		{
			name: "Case sensitive MINDFUL",
			content: `# mindful
Lowercase should not match.

# MINDFUL
Uppercase should match.

# Mindful
Mixed case should not match.`,
			wantMindful: "Uppercase should match.",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := models.NewMemoryConfig()
			err := m.ParseMemoryContent(tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMemoryContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if m.GetMindfulMemory() != tt.wantMindful {
				t.Errorf("ParseMemoryContent() MindfulMemory = %q, want %q", m.GetMindfulMemory(), tt.wantMindful)
			}

			// Verify that original content is preserved
			if m.Content != tt.content {
				t.Errorf("ParseMemoryContent() Content = %q, want %q", m.Content, tt.content)
			}
		})
	}
}
