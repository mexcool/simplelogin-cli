package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mexcool/simplelogin-cli/internal/api"
	intauth "github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with SimpleLogin",
	Long: `Store your SimpleLogin API key for CLI access.

There are three ways to authenticate:

1. Direct API key:
   Provide your API key directly. It will be stored in
   $XDG_CONFIG_HOME/simplelogin/config.yml (defaults to ~/.config/simplelogin/config.yml).

2. 1Password integration:
   Store a reference to your API key in 1Password. The CLI will
   use the "op" CLI to retrieve the key on each request. This is
   the most secure option as the key is never stored on disk.

3. Interactive:
   If no flags are provided, you will be prompted to enter your
   API key interactively.

You can obtain an API key from your SimpleLogin dashboard at
https://app.simplelogin.io/dashboard/setting#api-key`,
	Example: `  # Login with API key directly
  sl auth login --key sl_xxxxxxxxxxxxx

  # Login with 1Password integration
  sl auth login --1password --vault Personal --item "SimpleLogin API Key"

  # Login interactively (will prompt for key)
  sl auth login`,
	RunE: runLogin,
}

var (
	loginKey       string
	login1Password bool
	loginVault     string
	loginItem      string
)

func init() {
	loginCmd.Flags().StringVar(&loginKey, "key", "", "API key to store (note: value will appear in shell history and ps output; prefer interactive or --1password)")
	loginCmd.Flags().BoolVar(&login1Password, "1password", false, "Use 1Password integration")
	loginCmd.Flags().StringVar(&loginVault, "vault", "", "1Password vault name")
	loginCmd.Flags().StringVar(&loginItem, "item", "", "1Password item name")
}

func runLogin(cmd *cobra.Command, args []string) error {
	if login1Password {
		if loginVault == "" || loginItem == "" {
			return fmt.Errorf("--vault and --item are required with --1password")
		}

		if err := intauth.SaveOPRef(loginVault, loginItem); err != nil {
			return fmt.Errorf("failed to save 1Password reference: %w", err)
		}

		// Validate by trying to get the key and calling the API
		key, err := intauth.GetAPIKey()
		if err != nil {
			output.PrintWarning("1Password reference saved, but could not validate: %v", err)
			output.PrintWarning("Make sure the 'op' CLI is installed and you are signed in.")
			return nil
		}

		client := api.NewClient(key)
		info, _, err := client.GetUserInfo()
		if err != nil {
			output.PrintWarning("1Password reference saved, but API validation failed: %v", err)
			return nil
		}

		output.PrintSuccess("Authenticated as %s (%s) via 1Password", info.Name, info.Email)
		return nil
	}

	key := loginKey
	if key == "" {
		if !output.IsInteractive() {
			return fmt.Errorf("no API key provided; use: sl auth login --key <api-key>")
		}
		// Interactive mode
		fmt.Fprint(os.Stderr, "Enter your SimpleLogin API key: ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			key = strings.TrimSpace(scanner.Text())
		}
		if key == "" {
			return fmt.Errorf("no API key provided")
		}
	}

	// Validate key by calling the API
	client := api.NewClient(key)
	info, _, err := client.GetUserInfo()
	if err != nil {
		return fmt.Errorf("invalid API key: %w", err)
	}

	// Save the key
	if err := intauth.SaveAPIKey(key); err != nil {
		return fmt.Errorf("failed to save API key: %w", err)
	}

	output.PrintSuccess("Authenticated as %s (%s)", info.Name, info.Email)
	output.PrintSuccess("API key saved to %s", intauth.ConfigPath())
	return nil
}
