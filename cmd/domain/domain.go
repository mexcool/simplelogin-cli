package domain

import (
	"github.com/spf13/cobra"
)

// Cmd is the domain parent command.
var Cmd = &cobra.Command{
	Use:   "domain",
	Short: "Manage custom domains",
	Long: `Manage custom domains registered with your SimpleLogin account.

Custom domains allow you to create aliases using your own domain name.
You can configure catch-all, random prefix generation, and manage
which mailboxes receive emails for the domain.

Custom domains require a premium SimpleLogin subscription.`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(editCmd)
	Cmd.AddCommand(trashCmd)
}
