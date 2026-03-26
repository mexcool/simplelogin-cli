package contact

import (
	"github.com/spf13/cobra"
)

var unblockCmd = &cobra.Command{
	Use:   "unblock <contact-id>",
	Short: "Unblock a contact (idempotent)",
	Long: `Ensure a contact is unblocked. If the contact is already unblocked, this
is a no-op and exits successfully. This is safer than "toggle" for
scripting and agentic use because it is idempotent — calling it twice
has the same effect as calling it once.

When unblocked, emails from the contact to your alias flow normally.

Use "sl contact list <alias>" to find contact IDs.`,
	Example: `  # Unblock a contact
  sl contact unblock 456

  # Unblock and get result as JSON
  sl contact unblock 456 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runUnblock,
}

var (
	unblockJSON bool
	unblockJQ   string
)

func init() {
	unblockCmd.Flags().BoolVar(&unblockJSON, "json", false, "Output as JSON")
	unblockCmd.Flags().StringVar(&unblockJQ, "jq", "", "Apply jq expression to JSON output")
}

func runUnblock(cmd *cobra.Command, args []string) error {
	return setContactBlockState(args[0], false, unblockJSON, unblockJQ)
}
