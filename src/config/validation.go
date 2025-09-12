package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"wexler/src/models"
)

// ValidateProjectStructure validates the project directory structure
func ValidateProjectStructure(projectPath string, config *models.ProjectConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate source path exists or can be created
	sourcePath, err := config.GetAbsoluteSourcePath()
	if err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}
	if err := validatePath(sourcePath, "source directory"); err != nil {
		return err
	}

	return nil
}

// ValidateToolConfiguration validates tool-specific configuration
func ValidateToolConfiguration(config *models.ProjectConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	supportedTools := map[string]bool{
		"claude": true,
		"cursor": true,
	}

	for toolName, status := range config.Tools {
		// Validate tool name
		if !supportedTools[toolName] {
			return fmt.Errorf("unsupported tool: %s", toolName)
		}

		// Validate tool status
		validStatuses := map[string]bool{
			"enabled":  true,
			"disabled": true,
		}
		if !validStatuses[status] {
			return fmt.Errorf("invalid status '%s' for tool '%s', must be 'enabled' or 'disabled'", status, toolName)
		}
	}

	return nil
}

// ValidateProjectName validates the project name follows conventions
func ValidateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	if len(name) > 100 {
		return fmt.Errorf("project name too long (max 100 characters)")
	}

	// Check for invalid characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("project name cannot contain '%s'", char)
		}
	}

	return nil
}

// ValidateVersion validates the version string
func ValidateVersion(version string) error {
	if version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	// Basic semantic version validation
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return fmt.Errorf("version must be in format 'x.y.z' (got '%s')", version)
	}

	for i, part := range parts {
		if part == "" {
			return fmt.Errorf("version part %d cannot be empty", i+1)
		}
		// Validate that each part is numeric
		if _, err := strconv.Atoi(part); err != nil {
			return fmt.Errorf("version part '%s' is not a valid number", part)
		}
	}

	return nil
}

// validatePath validates that a path exists or can be created
func validatePath(path, description string) error {
	if path == "" {
		return fmt.Errorf("%s path cannot be empty", description)
	}

	// Check if path exists
	if info, err := os.Stat(path); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s path '%s' exists but is not a directory", description, path)
		}
		return nil
	}

	// Try to create the path
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("cannot create %s at '%s': %w", description, path, err)
	}

	// Clean up the test directory we just created
	os.RemoveAll(path)

	return nil
}