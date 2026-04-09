package domain

import (
	"github.com/spf13/cobra"
)

// Cmd is the domain parent command.
// Running "sl domain" with no subcommand lists domains (content-first).
var Cmd = &cobra.Command{
	Use:   "domain",
	Short: "Manage custom domains",
	Long: `Manage custom domains registered with your SimpleLogin account.

Custom domains allow you to create aliases using your own domain name.
You can configure catch-all, random prefix generation, and manage
which mailboxes receive emails for the domain.

Custom domains require a premium SimpleLogin subscription.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList(cmd, args)
	},
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(viewCmd)
	Cmd.AddCommand(editCmd)
	Cmd.AddCommand(trashCmd)
}
