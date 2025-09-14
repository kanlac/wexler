package cli

import (
	"fmt"
	"os"

	"mindful/src/apply"
	"mindful/src/config"
	"mindful/src/source"
	"mindful/src/storage"
	"mindful/src/tools"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	verbose   bool
	dryRun    bool
	configDir string
	force     bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mindful",
	Short: "AI Configuration Management Tool",
	Long: `Mindful is a unified AI assistant configuration management tool that maintains 
a single source of configuration truth across multiple AI tools (Claude Code, Cursor).

It prevents configuration fragmentation and ensures team consistency by syncing
configurations from a central source directory to various AI tools.`,
	Example: `  # Initialize Mindful in current directory
  mindful init

  # Apply configurations to Claude Code
  mindful apply --tool=claude

  # Import existing configurations
  mindful import --tool=cursor

  # List all managed configurations
  mindful list`,
}

// Application context holds shared managers
type AppContext struct {
	ConfigManager  *config.Manager
	SourceManager  *source.Manager
	StorageManager *storage.Manager
	ApplyManager   *apply.Manager
	ProjectPath    string
	StoragePath    string
}

// NewAppContext creates a new application context with all managers
func NewAppContext() (*AppContext, error) {
	projectPath, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Load project configuration to get the correct storage path
	configManager := config.NewManager()
	projectConfig, err := configManager.LoadProject(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load project configuration: %w", err)
	}

	// Get database path from project configuration
	storagePath, err := projectConfig.GetDatabasePath()
	if err != nil {
		return nil, fmt.Errorf("failed to determine storage path: %w", err)
	}

	return &AppContext{
		ConfigManager: configManager,
		SourceManager: source.NewManager(),
		ApplyManager:  apply.NewManager(),
		ProjectPath:   projectPath,
		StoragePath:   storagePath,
	}, nil
}

// GetStorageManager initializes storage manager with proper path
func (ctx *AppContext) GetStorageManager() (*storage.Manager, error) {
	if ctx.StorageManager == nil {
		mgr, err := storage.NewManager(ctx.StoragePath)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize storage: %w", err)
		}
		ctx.StorageManager = mgr
	}
	return ctx.StorageManager, nil
}

// GetToolAdapter creates a tool adapter for the specified tool
func (ctx *AppContext) GetToolAdapter(toolName string) (*tools.Adapter, error) {
	return tools.NewAdapter(toolName)
}

// CloseResources closes all open resources
func (ctx *AppContext) CloseResources() error {
	if ctx.StorageManager != nil {
		return ctx.StorageManager.Close()
	}
	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would be done without making changes")
	rootCmd.PersistentFlags().StringVar(&configDir, "config", "", "config directory (default is current directory)")
	rootCmd.PersistentFlags().BoolVar(&force, "force", false, "force operation without confirmation prompts")

	// Add subcommands
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newApplyCmd())
	rootCmd.AddCommand(newImportCmd())
	rootCmd.AddCommand(newListCmd())
}

func initConfig() {
	if configDir != "" {
		// Use config directory from flag
		os.Chdir(configDir)
	}
}
