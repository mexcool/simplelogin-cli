package auth

import (
	"fmt"
	"os"

	"github.com/mexcool/simplelogin-cli/internal/api"
	intauth "github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	Long: `Display the current authentication status, including the authenticated
user's name and email, and the source of the API key.

This is useful to verify which account is active and how the CLI is
obtaining the API key (environment variable, 1Password, or config file).`,
	Example: `  # Check authentication status
  sl auth status`,
	RunE: runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	key, err := intauth.GetAPIKey()
	if err != nil {
		output.PrintError("Not authenticated: %v", err)
		return err
	}

	// Determine key source
	source := "config file"
	if os.Getenv("SIMPLELOGIN_API_KEY") != "" {
		source = "SIMPLELOGIN_API_KEY env var"
	} else if os.Getenv("SL_API_KEY") != "" {
		source = "SL_API_KEY env var"
	} else if intauth.GetOPRef() != "" {
		source = "1Password (" + intauth.GetOPRef() + ")"
	}

	client := api.NewClient(key)
	info, _, err := client.GetUserInfo()
	if err != nil {
		output.PrintError("Failed to validate key: %v", err)
		return err
	}

	fmt.Fprintf(os.Stderr, "Authenticated as: %s (%s)\n", output.Bold.Sprint(info.Name), info.Email)
	fmt.Fprintf(os.Stderr, "Key source:        %s\n", source)
	fmt.Fprintf(os.Stderr, "Key:               %s\n", intauth.MaskKey(key))

	premium := "no"
	if info.IsPremium {
		premium = output.Green.Sprint("yes")
	}
	fmt.Fprintf(os.Stderr, "Premium:           %s\n", premium)

	return nil
}
