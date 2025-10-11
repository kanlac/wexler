package symlink

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"mindful/src/models"
)

// Manager orchestrates planning, creation, validation, and cleanup of symlinks.
type Manager struct {
	projectPath string
	outPath     string
	config      *models.SymlinkConfig
	resolver    *Resolver
}

// NewManager constructs a new Manager for a project.
func NewManager(projectPath string, config *models.SymlinkConfig) (*Manager, error) {
	if config == nil {
		var err error
		config, err = DefaultConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load default symlink configuration: %w", err)
		}
	}

	resolver := NewResolver(projectPath)

	return &Manager{
		projectPath: projectPath,
		outPath:     resolver.OutDir(),
		config:      config,
		resolver:    resolver,
	}, nil
}

// PlanSymlinks computes the desired symlink state for a tool without mutating the filesystem.
func (m *Manager) PlanSymlinks(toolName string) ([]models.SymlinkInfo, error) {
	plans, err := m.plan(toolName, true)
	if err != nil {
		return nil, err
	}

	infos := make([]models.SymlinkInfo, 0, len(plans))
	for _, plan := range plans {
		infos = append(infos, plan.info)
	}

	return infos, nil
}

// ListSymlinks reports the current state of symlinks for a tool.
func (m *Manager) ListSymlinks(toolName string) ([]models.SymlinkInfo, error) {
	plans, err := m.plan(toolName, false)
	if err != nil {
		return nil, err
	}

	infos := make([]models.SymlinkInfo, 0, len(plans))
	for _, plan := range plans {
		infos = append(infos, plan.info)
	}

	return infos, nil
}

// CreateSymlinks ensures all declared symlinks exist and point to mindful/out artifacts.
func (m *Manager) CreateSymlinks(toolName string) error {
	plans, err := m.plan(toolName, true)
	if err != nil {
		return err
	}

	var errs []error
	for _, plan := range plans {
		if err := m.ensureSymlink(plan); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// UpdateSymlinks is an alias for CreateSymlinks, retained for API completeness.
func (m *Manager) UpdateSymlinks(toolName string) error {
	return m.CreateSymlinks(toolName)
}

// CleanupSymlinks removes the symlinks declared for a tool.
func (m *Manager) CleanupSymlinks(toolName string) error {
	plans, err := m.plan(toolName, false)
	if err != nil {
		return err
	}

	var errs []error
	for _, plan := range plans {
		if err := m.removeSymlink(plan); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// ValidateSymlinks returns an error when any symlink is missing or points to the wrong target.
func (m *Manager) ValidateSymlinks(toolName string) error {
	plans, err := m.plan(toolName, true)
	if err != nil {
		return err
	}

	var invalid []string
	for _, plan := range plans {
		if !plan.info.IsValid {
			invalid = append(invalid, plan.info.LinkPath)
		}
	}

	if len(invalid) > 0 {
		return fmt.Errorf("invalid symlinks detected for %s: %s", toolName, strings.Join(invalid, ", "))
	}
	return nil
}

// plan builds the desired symlink state.
func (m *Manager) plan(toolName string, verifyTargets bool) ([]*plannedLink, error) {
	if strings.TrimSpace(toolName) == "" {
		return nil, fmt.Errorf("tool name cannot be empty")
	}

	toolConfig, ok := m.config.ToolConfig(toolName)
	if !ok || toolConfig == nil || toolConfig.IsEmpty() {
		return nil, fmt.Errorf("no symlink configuration for tool %q", toolName)
	}

	planner := newPlanner(m.resolver, toolConfig)
	return planner.buildPlans(verifyTargets)
}

func (m *Manager) ensureSymlink(plan *plannedLink) error {
	if plan == nil {
		return nil
	}

	// Quick exit when the existing symlink is already correct.
	if plan.info.IsValid {
		return nil
	}

	// Refuse to overwrite an existing non-symlink to avoid destroying user files.
	if stat, err := os.Lstat(plan.linkAbs); err == nil {
		if stat.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf(
				"cannot create symlink at %s: a regular file or directory already exists. "+
					"Please back up and remove it before rerunning mindful apply",
				plan.info.LinkPath,
			)
		}
	}

	if err := os.MkdirAll(filepath.Dir(plan.linkAbs), 0o755); err != nil {
		return fmt.Errorf("failed to prepare directory for %s: %w", plan.linkAbs, err)
	}

	if err := m.clearExistingPath(plan.linkAbs); err != nil {
		return err
	}

	target := plan.targetAbs
	if !filepath.IsAbs(target) {
		target = filepath.Join(m.projectPath, target)
	}

	linkDir := filepath.Dir(plan.linkAbs)
	relativeTarget, err := filepath.Rel(linkDir, target)
	if err != nil {
		relativeTarget = target
	}

	if runtime.GOOS == "windows" && plan.info.IsDirectory {
		// On Windows we need to hint directory links; os.Symlink handles this via the target existing as a directory.
	}

	if err := os.Symlink(relativeTarget, plan.linkAbs); err != nil {
		return fmt.Errorf("failed to create symlink %s -> %s: %w", plan.linkAbs, relativeTarget, err)
	}

	return nil
}

func (m *Manager) removeSymlink(plan *plannedLink) error {
	if plan == nil {
		return nil
	}

	info, err := os.Lstat(plan.linkAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to stat %s: %w", plan.linkAbs, err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		// Skip non-symlink paths to avoid accidental data loss.
		return nil
	}

	if err := os.Remove(plan.linkAbs); err != nil {
		return fmt.Errorf("failed to remove symlink %s: %w", plan.linkAbs, err)
	}
	return nil
}

func (m *Manager) clearExistingPath(linkPath string) error {
	info, err := os.Lstat(linkPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to inspect %s: %w", linkPath, err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		if err := os.Remove(linkPath); err != nil {
			return fmt.Errorf("failed to replace existing symlink %s: %w", linkPath, err)
		}
		return nil
	}

	return fmt.Errorf("cannot replace %s: existing path is not a symlink", linkPath)
}

// plannedLink keeps bookkeeping information for symlink operations.
type plannedLink struct {
	info      models.SymlinkInfo
	linkAbs   string
	targetAbs string
}

// planner transforms tool configuration into executable plans.
type planner struct {
	config   *models.ToolSymlinkConfig
	resolver *Resolver
}

func newPlanner(resolver *Resolver, config *models.ToolSymlinkConfig) *planner {
	return &planner{
		config:   config,
		resolver: resolver,
	}
}

func (p *planner) buildPlans(verifyTargets bool) ([]*plannedLink, error) {
	var plans []*plannedLink

	if plan, err := p.planMemory(verifyTargets); err != nil {
		return nil, err
	} else if plan != nil {
		plans = append(plans, plan)
	}

	if subPlans, err := p.planSubagents(verifyTargets); err != nil {
		return nil, err
	} else {
		plans = append(plans, subPlans...)
	}

	if plan, err := p.planMCP(verifyTargets); err != nil {
		return nil, err
	} else if plan != nil {
		plans = append(plans, plan)
	}

	return plans, nil
}

func (p *planner) planMemory(verify bool) (*plannedLink, error) {
	if p.config == nil || strings.TrimSpace(p.config.Memory) == "" {
		return nil, nil
	}
	return p.planSingle(p.config.Memory, p.resolver.MemoryArtifact(), verify)
}

func (p *planner) planMCP(verify bool) (*plannedLink, error) {
	if p.config == nil || strings.TrimSpace(p.config.MCP) == "" {
		return nil, nil
	}
	return p.planSingle(p.config.MCP, p.resolver.MCPArtifact(), verify)
}

func (p *planner) planSubagents(verify bool) ([]*plannedLink, error) {
	template := strings.TrimSpace(p.config.Subagents)
	if template == "" {
		return nil, nil
	}

	entries, err := os.ReadDir(p.resolver.SubagentDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read subagents directory: %w", err)
	}

	var plans []*plannedLink
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		linkPath := strings.ReplaceAll(template, SubagentPlaceholder, name)
		target := filepath.Join(p.resolver.SubagentDir(), entry.Name())

		plan, err := p.planSingle(linkPath, target, verify)
		if err != nil {
			return nil, err
		}
		if plan != nil {
			plans = append(plans, plan)
		}
	}

	return plans, nil
}

func (p *planner) planSingle(linkTemplate, target string, verify bool) (*plannedLink, error) {
	linkAbs, linkRel := p.resolver.ResolveLink(linkTemplate)
	targetAbs := p.resolver.ResolveTarget(target)
	targetRel := p.resolver.RelativeToProject(targetAbs)

	var isDir bool
	if verify {
		info, err := os.Stat(targetAbs)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("target %s does not exist", targetAbs)
			}
			return nil, fmt.Errorf("failed to stat target %s: %w", targetAbs, err)
		}
		isDir = info.IsDir()
	}

	info := models.SymlinkInfo{
		LinkPath:    linkRel,
		TargetPath:  targetRel,
		IsDirectory: isDir,
		IsValid:     false,
	}

	exists, isValid, err := p.evaluateExistingLink(linkAbs, targetAbs)
	if err != nil {
		return nil, err
	}

	if exists && isValid {
		info.IsValid = true
	}

	return &plannedLink{
		info:      info,
		linkAbs:   linkAbs,
		targetAbs: targetAbs,
	}, nil
}

func (p *planner) evaluateExistingLink(linkAbs, targetAbs string) (bool, bool, error) {
	stat, err := os.Lstat(linkAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return false, false, nil
		}
		return false, false, fmt.Errorf("failed to inspect %s: %w", linkAbs, err)
	}

	if stat.Mode()&os.ModeSymlink == 0 {
		return true, false, nil
	}

	dest, err := os.Readlink(linkAbs)
	if err != nil {
		return true, false, fmt.Errorf("failed to read symlink %s: %w", linkAbs, err)
	}

	if !filepath.IsAbs(dest) {
		dest = filepath.Join(filepath.Dir(linkAbs), dest)
	}

	dest = filepath.Clean(dest)
	targetAbs = filepath.Clean(targetAbs)

	if pathsEqual(dest, targetAbs) {
		return true, true, nil
	}

	return true, false, nil
}

func pathsEqual(a, b string) bool {
	a = filepath.Clean(a)
	b = filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}
