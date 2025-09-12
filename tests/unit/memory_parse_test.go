package unit

import (
	"testing"
	"wexler/src/models"
)

func TestMemoryConfig_ParseMemoryContent(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		wantWexler     string
		wantErr        bool
	}{
		{
			name:       "Empty content",
			content:    "",
			wantWexler: "",
			wantErr:    false,
		},
		{
			name:       "Only whitespace",
			content:    "   \n\t  \n  ",
			wantWexler: "",
			wantErr:    false,
		},
		{
			name: "WEXLER section only",
			content: `# WEXLER
This is the WEXLER section content.

## Workflow
Use TDD approach.

## Code Style
Follow Go conventions.`,
			wantWexler: `This is the WEXLER section content.

## Workflow
Use TDD approach.

## Code Style
Follow Go conventions.`,
			wantErr: false,
		},
		{
			name: "WEXLER section with other sections",
			content: `# Introduction
This is the introduction.

# WEXLER
Main memory content here.

## Guidelines
Follow these guidelines.

# Other Section
This should be ignored.`,
			wantWexler: `Main memory content here.

## Guidelines
Follow these guidelines.`,
			wantErr: false,
		},
		{
			name: "No WEXLER section",
			content: `# Introduction
Some content here.

# Other Section
More content.`,
			wantWexler: "",
			wantErr:    false,
		},
		{
			name: "WEXLER section at the end",
			content: `# Introduction
Introduction content.

# WEXLER
<--- Main memory of all coding agents. Managed by wexler. DO NOT EDIT outside our wexler source directory. --->

## Workflow
Prefer running single tests for performance.

## Code Style
Use Go conventions and direct framework usage.

## Project Context
This is the project context and instructions for AI assistants.`,
			wantWexler: `<--- Main memory of all coding agents. Managed by wexler. DO NOT EDIT outside our wexler source directory. --->

## Workflow
Prefer running single tests for performance.

## Code Style
Use Go conventions and direct framework usage.

## Project Context
This is the project context and instructions for AI assistants.`,
			wantErr: false,
		},
		{
			name: "Multiple WEXLER sections (first one wins)",
			content: `# WEXLER
First WEXLER section.

# Other
Other content.

# WEXLER
Second WEXLER section.`,
			wantWexler: "First WEXLER section.",
			wantErr:    false,
		},
		{
			name: "WEXLER section with trailing whitespace",
			content: `# WEXLER
Content with spaces.   

More content.  
`,
			wantWexler: `Content with spaces.   

More content.`,
			wantErr: false,
		},
		{
			name: "Case sensitive WEXLER",
			content: `# wexler
Lowercase should not match.

# WEXLER
Uppercase should match.

# Wexler
Mixed case should not match.`,
			wantWexler: "Uppercase should match.",
			wantErr:    false,
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
			
			if m.GetWexlerMemory() != tt.wantWexler {
				t.Errorf("ParseMemoryContent() WexlerMemory = %q, want %q", m.GetWexlerMemory(), tt.wantWexler)
			}
			
			// Verify that original content is preserved
			if m.Content != tt.content {
				t.Errorf("ParseMemoryContent() Content = %q, want %q", m.Content, tt.content)
			}
		})
	}
}