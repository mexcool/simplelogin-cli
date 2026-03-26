package alias

import (
	"fmt"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var toggleCmd = &cobra.Command{
	Use:   "toggle <alias-id-or-email>",
	Short: "Enable or disable an alias",
	Long: `Toggle an alias between enabled and disabled states.

When disabled, the alias will reject all incoming emails. When enabled,
emails will be forwarded normally. The command reports the new state
after toggling.

Accepts either a numeric alias ID or the full alias email address.`,
	Example: `  # Toggle alias by ID
  sl alias toggle 12345

  # Toggle alias by email
  sl alias toggle my-alias@simplelogin.co

  # Toggle and get result as JSON
  sl alias toggle 12345 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runToggle,
}

var (
	toggleJSON bool
	toggleJQ   string
)

func init() {
	toggleCmd.Flags().BoolVar(&toggleJSON, "json", false, "Output as JSON")
	toggleCmd.Flags().StringVar(&toggleJQ, "jq", "", "Apply jq expression to JSON output")
}

func runToggle(cmd *cobra.Command, args []string) error {
	key, err := auth.GetAPIKey()
	if err != nil {
		return err
	}

	client := api.NewClient(key)
	id, err := client.ResolveAliasID(args[0])
	if err != nil {
		return err
	}

	enabled, rawJSON, err := client.ToggleAlias(id)
	if err != nil {
		return err
	}

	if toggleJQ != "" {
		return output.PrintJQ(rawJSON, toggleJQ)
	}
	if toggleJSON {
		return output.PrintJSON(rawJSON)
	}

	if enabled {
		output.PrintSuccess("Alias %s is now enabled", args[0])
	} else {
		output.PrintWarning("Alias %s is now disabled", args[0])
	}
	fmt.Printf("enabled=%v\n", enabled)
	return nil
}
