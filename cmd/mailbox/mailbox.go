package mailbox

import (
	"github.com/spf13/cobra"
)

// Cmd is the mailbox parent command.
var Cmd = &cobra.Command{
	Use:   "mailbox",
	Short: "Manage mailboxes",
	Long: `Manage your SimpleLogin mailboxes.

Mailboxes are your real email addresses that receive forwarded emails
from aliases. You can have multiple mailboxes and assign different
aliases to different mailboxes.

Each mailbox must be verified before it can receive forwarded emails.
One mailbox is marked as the default, which is used for new aliases
when no mailbox is explicitly specified.`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(editCmd)
}
