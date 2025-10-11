package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const mindfulVersion = "0.0.0-dev"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show the Mindful CLI version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), mindfulVersion)
		},
	}
}
