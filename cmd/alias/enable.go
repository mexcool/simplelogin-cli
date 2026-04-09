package alias

import (
	"fmt"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var enableCmd = &cobra.Command{
	Use:   "enable <alias-id-or-email>",
	Short: "Enable an alias (idempotent)",
	Long: `Ensure an alias is enabled. If the alias is already enabled, this is a
no-op and exits successfully. This is safer than "toggle" for scripting
and agentic use because it is idempotent — calling it twice has the same
effect as calling it once.

Accepts either a numeric alias ID or the full alias email address.`,
	Example: `  # Enable alias by ID
  sl alias enable 12345

  # Enable alias by email (idempotent — safe to call repeatedly)
  sl alias enable my-alias@simplelogin.co

  # Enable and get result as JSON
  sl alias enable 12345 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runEnable,
}

var (
	enableJSON bool
	enableJQ   string
)

func init() {
	enableCmd.Flags().BoolVar(&enableJSON, "json", false, "Output as JSON")
	enableCmd.Flags().StringVar(&enableJQ, "jq", "", "Apply jq expression to JSON output")
}

func runEnable(cmd *cobra.Command, args []string) error {
	return setAliasState(args[0], true, enableJSON, enableJQ)
}

// setAliasState ensures an alias is in the desired enabled/disabled state.
// It reads the current state first and only toggles if needed (idempotent).
func setAliasState(idOrEmail string, wantEnabled bool, jsonFlag bool, jqExpr string) error {
	key, err := auth.GetAPIKey()
	if err != nil {
		return err
	}

	client := api.NewClient(key, auth.GetAPIBase())
	id, err := client.ResolveAliasID(idOrEmail)
	if err != nil {
		return err
	}

	alias, rawJSON, err := client.GetAlias(id)
	if err != nil {
		return err
	}

	verb := "enabled"
	if !wantEnabled {
		verb = "disabled"
	}

	if alias.Enabled == wantEnabled {
		// Already in the desired state — idempotent success.
		if jqExpr != "" {
			return output.PrintJQ(rawJSON, jqExpr)
		}
		if jsonFlag {
			return output.PrintJSON(rawJSON)
		}
		output.PrintSuccess("Alias %s is already %s", idOrEmail, verb)
		output.PrintHint("sl alias view %s", idOrEmail)
		fmt.Printf("enabled=%v\n", alias.Enabled)
		return nil
	}

	// Toggle to reach the desired state.
	enabled, rawJSON, err := client.ToggleAlias(id)
	if err != nil {
		return err
	}

	if jqExpr != "" {
		return output.PrintJQ(rawJSON, jqExpr)
	}
	if jsonFlag {
		return output.PrintJSON(rawJSON)
	}

	if enabled {
		output.PrintSuccess("Alias %s is now enabled", idOrEmail)
	} else {
		output.PrintWarning("Alias %s is now disabled", idOrEmail)
	}
	output.PrintHint("sl alias view %s", idOrEmail)
	fmt.Printf("enabled=%v\n", enabled)
	return nil
}
