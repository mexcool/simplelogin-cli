package auth

import (
	"github.com/mexcool/simplelogin-cli/internal/api"
	intauth "github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	Long: `Remove stored API key and 1Password references from the config file.

This command clears stored credentials from $XDG_CONFIG_HOME/simplelogin/config.yml (defaults to ~/.config/simplelogin/config.yml).
It also attempts to invalidate the API session on the server.

Note: This does not affect environment variables (SIMPLELOGIN_API_KEY, SL_API_KEY).
To fully log out when using environment variables, unset them manually.`,
	Example: `  # Logout and clear stored credentials
  sl auth logout`,
	RunE: runLogout,
}

func runLogout(cmd *cobra.Command, args []string) error {
	// Try to logout from server
	key, err := intauth.GetAPIKey()
	if err == nil {
		client := api.NewClient(key)
		_ = client.Logout()
	}

	// Clear local config
	if err := intauth.ClearConfig(); err != nil {
		output.PrintError("Failed to clear config: %v", err)
		return err
	}

	output.PrintSuccess("Logged out successfully")
	return nil
}
