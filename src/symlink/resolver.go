package symlink

import (
	"path/filepath"

	"mindful/src/models"
)

// Resolver centralises path calculations for symlink management.
type Resolver struct {
	projectPath string
	mindfulDir  string
	outDir      string
}

// NewResolver constructs a resolver rooted at the given project path.
func NewResolver(projectPath string) *Resolver {
	mindfulDir := filepath.Join(projectPath, models.DefaultMindfulDirName)
	return &Resolver{
		projectPath: projectPath,
		mindfulDir:  mindfulDir,
		outDir:      filepath.Join(mindfulDir, models.DefaultOutDirName),
	}
}

// ProjectPath returns the project root.
func (r *Resolver) ProjectPath() string {
	return r.projectPath
}

// MindfulDir returns the mindful/ directory within the project.
func (r *Resolver) MindfulDir() string {
	return r.mindfulDir
}

// OutDir returns the mindful/out directory.
func (r *Resolver) OutDir() string {
	return r.outDir
}

// SubagentDir returns mindful/out/subagents.
func (r *Resolver) SubagentDir() string {
	return filepath.Join(r.outDir, "subagents")
}

// MemoryArtifact returns mindful/out/memory.md.
func (r *Resolver) MemoryArtifact() string {
	return filepath.Join(r.outDir, "memory.md")
}

// MCPArtifact returns mindful/out/mcp.json.
func (r *Resolver) MCPArtifact() string {
	return filepath.Join(r.outDir, "mcp.json")
}

// ResolveLink resolves a configured link path to both absolute and project-relative forms.
func (r *Resolver) ResolveLink(linkPath string) (string, string) {
	if filepath.IsAbs(linkPath) {
		abs := filepath.Clean(linkPath)
		return abs, r.RelativeToProject(abs)
	}

	abs := filepath.Join(r.projectPath, linkPath)
	return filepath.Clean(abs), filepath.Clean(linkPath)
}

// ResolveTarget returns the absolute path to the target artefact.
func (r *Resolver) ResolveTarget(targetPath string) string {
	if filepath.IsAbs(targetPath) {
		return filepath.Clean(targetPath)
	}
	return filepath.Clean(filepath.Join(r.projectPath, targetPath))
}

// RelativeToProject converts an absolute path into a project-relative one if possible.
func (r *Resolver) RelativeToProject(path string) string {
	rel, err := filepath.Rel(r.projectPath, path)
	if err != nil {
		return filepath.Clean(path)
	}
	return filepath.Clean(rel)
}
