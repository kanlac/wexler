package symlink

import (
	_ "embed"
	"sync"

	"gopkg.in/yaml.v3"

	"mindful/src/models"
)

var (
	defaultConfigOnce sync.Once
	defaultConfig     *models.SymlinkConfig
	defaultConfigErr  error
)

//go:embed coding-agent-config-file-mapping.yaml
var defaultConfigYAML []byte

// DefaultConfig returns the shared tool → symlink mapping shipped with Mindful.
func DefaultConfig() (*models.SymlinkConfig, error) {
	defaultConfigOnce.Do(func() {
		var raw map[string]*models.ToolSymlinkConfig
		defaultConfigErr = yaml.Unmarshal(defaultConfigYAML, &raw)
		if defaultConfigErr != nil {
			return
		}
		defaultConfig = models.NewSymlinkConfig(raw)
	})
	return defaultConfig, defaultConfigErr
}
