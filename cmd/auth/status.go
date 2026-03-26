package auth

import (
	"encoding/json"
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
  sl auth status

  # Output as JSON
  sl auth status --json

  # Filter JSON output with jq
  sl auth status --json --jq '.email'`,
	RunE: runStatus,
}

var (
	statusJSON bool
	statusJQ   string
)

func init() {
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output as JSON")
	statusCmd.Flags().StringVar(&statusJQ, "jq", "", "Apply jq expression to JSON output")
}

func runStatus(cmd *cobra.Command, args []string) error {
	key, err := intauth.GetAPIKey()
	if err != nil {
		return fmt.Errorf("not authenticated: %w (run 'sl auth login' to authenticate)", err)
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

	client := api.NewClient(key, intauth.GetAPIBase())
	info, _, err := client.GetUserInfo()
	if err != nil {
		return fmt.Errorf("failed to validate API key: %w", err)
	}

	apiBase := intauth.GetAPIBase()

	if statusJSON || statusJQ != "" {
		data := make(map[string]interface{})
		data["name"] = info.Name
		data["email"] = info.Email
		data["is_premium"] = info.IsPremium
		data["in_trial"] = info.InTrial
		data["key_source"] = source
		data["key"] = intauth.MaskKey(key)
		data["api_url"] = apiBase
		jsonBytes, _ := json.Marshal(data)
		if statusJQ != "" {
			return output.PrintJQ(jsonBytes, statusJQ)
		}
		return output.PrintJSON(jsonBytes)
	}

	fmt.Fprintf(os.Stderr, "Authenticated as: %s (%s)\n", output.Bold.Sprint(info.Name), info.Email)
	fmt.Fprintf(os.Stderr, "API URL:           %s\n", apiBase)
	fmt.Fprintf(os.Stderr, "Key source:        %s\n", source)
	fmt.Fprintf(os.Stderr, "Key:               %s\n", intauth.MaskKey(key))

	premium := "no"
	if info.IsPremium {
		premium = output.Green.Sprint("yes")
	}
	fmt.Fprintf(os.Stderr, "Premium:           %s\n", premium)

	return nil
}
