package alias

import (
	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <alias-id-or-email>",
	Short: "Delete an alias",
	Long: `Permanently delete a SimpleLogin alias.

This action is irreversible. The alias will stop forwarding emails
immediately. You will be prompted for confirmation unless --yes is provided.

Accepts either a numeric alias ID or the full alias email address.`,
	Example: `  # Delete alias by ID (with confirmation prompt)
  sl alias delete 12345

  # Delete alias by email without confirmation
  sl alias delete my-alias@simplelogin.co --yes`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

var deleteYes bool

func init() {
	deleteCmd.Flags().BoolVar(&deleteYes, "yes", false, "Skip confirmation prompt")
}

func runDelete(cmd *cobra.Command, args []string) error {
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

	if !deleteYes {
		if !output.ConfirmAction("Delete alias " + args[0] + "?") {
			output.PrintWarning("Cancelled")
			return nil
		}
	}

	if err := client.DeleteAlias(id); err != nil {
		output.PrintError("%v", err)
		return err
	}

	output.PrintSuccess("Alias deleted")
	return nil
}
