package alias

import (
	"github.com/spf13/cobra"
)

// Cmd is the alias parent command.
var Cmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage email aliases",
	Long: `Create, list, and manage SimpleLogin email aliases.

Aliases are the core feature of SimpleLogin. Each alias forwards
emails to your real mailbox while hiding your actual email address.
You can create random aliases or custom aliases with your preferred
prefix and domain.

Most commands accept either an alias ID (integer) or the full alias
email address as an identifier.`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(viewCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(toggleCmd)
	Cmd.AddCommand(editCmd)
	Cmd.AddCommand(activityCmd)
}
