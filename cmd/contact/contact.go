package contact

import (
	"github.com/spf13/cobra"
)

// Cmd is the contact parent command.
var Cmd = &cobra.Command{
	Use:   "contact",
	Short: "Manage alias contacts",
	Long: `Manage contacts for your SimpleLogin aliases.

Contacts represent email addresses that have interacted with your aliases.
You can list contacts for an alias, add new contacts to create reverse
aliases (for sending email as your alias), block/unblock contacts, and
delete contacts.

When you add a contact, SimpleLogin creates a "reverse alias" that you
can use to send emails from your alias to that contact.`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(toggleCmd)
	Cmd.AddCommand(blockCmd)
	Cmd.AddCommand(unblockCmd)
}
