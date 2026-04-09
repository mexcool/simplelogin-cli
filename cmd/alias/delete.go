package alias

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete <alias-id-or-email>",
	Aliases: []string{"rm"},
	Short:   "Delete an alias",
	Long: `Permanently delete a SimpleLogin alias.

This action is irreversible. The alias will stop forwarding emails
immediately. You will be prompted for confirmation unless --yes is provided.

Use --dry-run to preview what would be deleted without actually deleting.

Accepts either a numeric alias ID or the full alias email address.`,
	Example: `  # Delete alias by ID (with confirmation prompt)
  sl alias delete 12345

  # Delete alias by email without confirmation
  sl alias delete my-alias@simplelogin.co --yes

  # Preview what would be deleted
  sl alias delete 12345 --dry-run

  # Delete and get JSON confirmation
  sl alias delete 12345 --yes --json

  # Filter JSON output with jq
  sl alias delete 12345 --yes --json --jq '.id'`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

var (
	deleteYes    bool
	deleteJSON   bool
	deleteJQ     string
	deleteDryRun bool
)

func init() {
	deleteCmd.Flags().BoolVar(&deleteYes, "yes", false, "Skip confirmation prompt")
	deleteCmd.Flags().BoolVar(&deleteJSON, "json", false, "Output as JSON")
	deleteCmd.Flags().StringVar(&deleteJQ, "jq", "", "Apply jq expression to JSON output")
	deleteCmd.Flags().BoolVar(&deleteDryRun, "dry-run", false, "Preview without deleting")
}

func runDelete(cmd *cobra.Command, args []string) error {
	key, err := auth.GetAPIKey()
	if err != nil {
		return err
	}

	client := api.NewClient(key, auth.GetAPIBase())
	id, err := client.ResolveAliasID(args[0])
	if err != nil {
		return err
	}

	if deleteDryRun {
		alias, rawJSON, err := client.GetAlias(id)
		if err != nil {
			return fmt.Errorf("failed to fetch alias for preview: %w", err)
		}
		if deleteJSON || deleteJQ != "" {
			if deleteJQ != "" {
				_ = output.PrintJQ(rawJSON, deleteJQ)
			} else {
				_ = output.PrintJSON(rawJSON)
			}
		} else {
			fmt.Fprintln(os.Stdout, "Would delete alias:")
			fmt.Fprintf(os.Stdout, "  ID:     %d\n", alias.ID)
			fmt.Fprintf(os.Stdout, "  Email:  %s\n", alias.Email)
			fmt.Fprintf(os.Stdout, "  Status: %s\n", output.EnabledStatus(alias.Enabled))
			fmt.Fprintf(os.Stdout, "  Note:   %s\n", output.StringOrEmpty(alias.Note))
		}
		output.PrintWarning("No changes made. (dry-run)")
		return nil
	}

	if !deleteYes {
		if !output.ConfirmAction("Delete alias " + args[0] + "?") {
			return fmt.Errorf("operation cancelled")
		}
	}

	if err := client.DeleteAlias(id); err != nil {
		return err
	}

	if deleteJSON || deleteJQ != "" {
		data, _ := json.Marshal(map[string]interface{}{"deleted": true, "id": id})
		if deleteJQ != "" {
			return output.PrintJQ(data, deleteJQ)
		}
		return output.PrintJSON(data)
	}
	output.PrintSuccess("Alias deleted")
	output.PrintHint("sl alias list")
	fmt.Println(id)
	return nil
}
