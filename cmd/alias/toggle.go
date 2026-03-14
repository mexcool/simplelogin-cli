package alias

import (
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
  sl alias toggle my-alias@simplelogin.co`,
	Args: cobra.ExactArgs(1),
	RunE: runToggle,
}

func runToggle(cmd *cobra.Command, args []string) error {
	key, err := auth.GetAPIKey()
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	client := api.NewClient(key)
	id, err := client.ResolveAliasID(args[0])
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	enabled, err := client.ToggleAlias(id)
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	if enabled {
		output.PrintSuccess("Alias %s is now enabled", args[0])
	} else {
		output.PrintWarning("Alias %s is now disabled", args[0])
	}
	return nil
}
