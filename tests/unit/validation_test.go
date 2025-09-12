package unit

import (
	"testing"
	"wexler/src/config"
	"wexler/src/models"
)

func TestProjectConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *models.ProjectConfig
		wantErr bool
	}{
		{
			name: "valid complete configuration",
			config: &models.ProjectConfig{
				Name:        "test-project",
				Version:     "1.0.0",
				SourcePath:  "source",
				Tools: map[string]string{
					"claude": "enabled",
					"cursor": "disabled",
				},
			},
			wantErr: false,
		},
		{
			name: "missing project name",
			config: &models.ProjectConfig{
				Version:     "1.0.0",
				SourcePath:  "source",
				Tools:       map[string]string{"claude": "enabled"},
			},
			wantErr: true,
		},
		{
			name: "invalid version format",
			config: &models.ProjectConfig{
				Name:        "test-project",
				Version:     "invalid-version",
				SourcePath:  "source",
				Tools:       map[string]string{"claude": "enabled"},
			},
			wantErr: true,
		},
		{
			name: "missing source path",
			config: &models.ProjectConfig{
				Name:        "test-project",
				Version:     "1.0.0",
				Tools:       map[string]string{"claude": "enabled"},
			},
			wantErr: true,
		},
		{
			name: "invalid tool status",
			config: &models.ProjectConfig{
				Name:        "test-project",
				Version:     "1.0.0",
				SourcePath:  "source",
				Tools: map[string]string{
					"claude": "maybe",
				},
			},
			wantErr: true,
		},
		{
			name: "unsupported tool",
			config: &models.ProjectConfig{
				Name:        "test-project",
				Version:     "1.0.0",
				SourcePath:  "source",
				Tools: map[string]string{
					"unsupported-tool": "enabled",
				},
			},
			wantErr: true,
		},
		{
			name:    "nil configuration",
			config:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := config.NewManager()
			err := manager.ValidateProject(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProjectNameValidation(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		wantErr     bool
	}{
		{
			name:        "valid simple name",
			projectName: "my-project",
			wantErr:     false,
		},
		{
			name:        "valid name with numbers",
			projectName: "project-123",
			wantErr:     false,
		},
		{
			name:        "valid name with underscores",
			projectName: "my_project_v2",
			wantErr:     false,
		},
		{
			name:        "empty name",
			projectName: "",
			wantErr:     true,
		},
		{
			name:        "name too long",
			projectName: "this-is-a-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-long-project-name-that-exceeds-the-limit",
			wantErr:     true,
		},
		{
			name:        "name with invalid characters",
			projectName: "project/with/slashes",
			wantErr:     true,
		},
		{
			name:        "name with spaces",
			projectName: "project with spaces",
			wantErr:     false, // Spaces should be allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateProjectName(tt.projectName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProjectName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVersionValidation(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{
			name:    "valid semantic version",
			version: "1.0.0",
			wantErr: false,
		},
		{
			name:    "valid version with patch",
			version: "2.1.3",
			wantErr: false,
		},
		{
			name:    "valid version with large numbers",
			version: "10.20.30",
			wantErr: false,
		},
		{
			name:    "empty version",
			version: "",
			wantErr: true,
		},
		{
			name:    "invalid format missing patch",
			version: "1.0",
			wantErr: true,
		},
		{
			name:    "invalid format too many parts",
			version: "1.0.0.0",
			wantErr: true,
		},
		{
			name:    "invalid format with text",
			version: "v1.0.0",
			wantErr: true,
		},
		{
			name:    "invalid format with empty part",
			version: "1..0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestToolConfigurationValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *models.ProjectConfig
		wantErr bool
	}{
		{
			name: "valid tool configuration",
			config: &models.ProjectConfig{
				Name:        "test",
				Version:     "1.0.0",
				SourcePath:  "source",
				Tools: map[string]string{
					"claude": "enabled",
					"cursor": "disabled",
				},
			},
			wantErr: false,
		},
		{
			name: "empty tools map",
			config: &models.ProjectConfig{
				Name:        "test",
				Version:     "1.0.0",
				SourcePath:  "source",
				Tools:       map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "nil tools map",
			config: &models.ProjectConfig{
				Name:        "test",
				Version:     "1.0.0",
				SourcePath:  "source",
				Tools:       nil,
			},
			wantErr: false,
		},
		{
			name: "unsupported tool",
			config: &models.ProjectConfig{
				Name:        "test",
				Version:     "1.0.0",
				SourcePath:  "source",
				Tools: map[string]string{
					"invalid-tool": "enabled",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid tool status",
			config: &models.ProjectConfig{
				Name:        "test",
				Version:     "1.0.0",
				SourcePath:  "source",
				Tools: map[string]string{
					"claude": "maybe",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.ValidateToolConfiguration(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToolConfiguration() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}