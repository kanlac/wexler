package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Show instructions for migrating existing files to the symlink workflow",
		RunE:  runImport,
	}
	return cmd
}

func runImport(cmd *cobra.Command, args []string) error {
	fmt.Fprintln(cmd.OutOrStdout(), "Import is no longer required. Use 'mindful build' followed by 'mindful apply' to refresh symlinks.")
	return nil
}
