package alias

import (
	"github.com/spf13/cobra"
)

var disableCmd = &cobra.Command{
	Use:   "disable <alias-id-or-email>",
	Short: "Disable an alias (idempotent)",
	Long: `Ensure an alias is disabled. If the alias is already disabled, this is a
no-op and exits successfully. This is safer than "toggle" for scripting
and agentic use because it is idempotent — calling it twice has the same
effect as calling it once.

When disabled, the alias will reject all incoming emails. When enabled,
emails will be forwarded normally.

Accepts either a numeric alias ID or the full alias email address.`,
	Example: `  # Disable alias by ID
  sl alias disable 12345

  # Disable alias by email (idempotent — safe to call repeatedly)
  sl alias disable my-alias@simplelogin.co

  # Disable and get result as JSON
  sl alias disable 12345 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runDisable,
}

var (
	disableJSON bool
	disableJQ   string
)

func init() {
	disableCmd.Flags().BoolVar(&disableJSON, "json", false, "Output as JSON")
	disableCmd.Flags().StringVar(&disableJQ, "jq", "", "Apply jq expression to JSON output")
}

func runDisable(cmd *cobra.Command, args []string) error {
	return setAliasState(args[0], false, disableJSON, disableJQ)
}
