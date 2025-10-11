package models

// BuildArtifacts represents the rendered outputs that should be written into mindful/out.
type BuildArtifacts struct {
	Memory     *MemoryArtifact     // Unified memory document for all tools
	Subagents  []*SubagentArtifact // Collection of rendered subagent files
	MCPContent []byte              // Serialized MCP configuration (optional)
}

// MemoryArtifact contains the text content of the unified memory file.
type MemoryArtifact struct {
	Content     string   // The final memory document text
	SourcePaths []string // Source files that contributed to the content
}

// SubagentArtifact captures the rendered content for a single subagent.
type SubagentArtifact struct {
	Name       string // Logical name of the subagent (e.g. researcher)
	FileName   string // File name to use on disk (e.g. researcher.mdc)
	Content    string // Rendered file contents
	SourcePath string // Originating file path (useful for diagnostics)
}
