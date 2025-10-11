package cli

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	projectPathFlag string
	verboseFlag     bool
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "mindful",
	Short: "Mindful keeps AI assistant configurations in sync",
	Long: `Mindful builds project configuration artefacts once and shares them across tools via symlinks.

Use "mindful build" to render mindful/out, and "mindful apply" to link the artefacts to enabled tools.`,
}

// Execute runs the CLI.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(applyProjectFlag)

	rootCmd.PersistentFlags().StringVar(&projectPathFlag, "project", "", "path to the mindful project (defaults to current directory)")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "enable verbose output")

	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newBuildCmd())
	rootCmd.AddCommand(newApplyCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newImportCmd())
	rootCmd.AddCommand(newVersionCmd())
}

func applyProjectFlag() {
	if projectPathFlag == "" {
		return
	}

	if err := os.Chdir(projectPathFlag); err != nil {
		// Fallback silently so command execution can report a meaningful error.
		return
	}

	projectPathFlag, _ = filepath.Abs(".")
}
