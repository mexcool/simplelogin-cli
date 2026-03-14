package setting

import (
	"github.com/spf13/cobra"
)

// Cmd is the setting parent command.
var Cmd = &cobra.Command{
	Use:   "setting",
	Short: "Manage account settings",
	Long: `View and update your SimpleLogin account settings.

Settings control default behavior for alias creation, email sender
format, notification preferences, and more.`,
}

func init() {
	Cmd.AddCommand(viewCmd)
	Cmd.AddCommand(editCmd)
	Cmd.AddCommand(domainsCmd)
}
