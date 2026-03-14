package export

import (
	"github.com/spf13/cobra"
)

// Cmd is the export parent command.
var Cmd = &cobra.Command{
	Use:   "export",
	Short: "Export account data",
	Long: `Export your SimpleLogin account data.

You can export all account data as JSON or export just your aliases
as CSV. Data is printed to stdout by default, or saved to a file
with --output.`,
}

func init() {
	Cmd.AddCommand(dataCmd)
	Cmd.AddCommand(aliasesCmd)
}
