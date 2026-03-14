package auth

import (
	"github.com/spf13/cobra"
)

// Cmd is the auth parent command.
var Cmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long: `Manage authentication for the SimpleLogin CLI.

The CLI supports three authentication methods:
  1. Environment variables (SIMPLELOGIN_API_KEY or SL_API_KEY)
  2. Direct API key stored in config file
  3. 1Password integration via the op CLI

Use "sl auth login" to configure authentication, "sl auth status" to
verify your current session, and "sl auth logout" to clear stored credentials.`,
}

func init() {
	Cmd.AddCommand(loginCmd)
	Cmd.AddCommand(logoutCmd)
	Cmd.AddCommand(statusCmd)
}
