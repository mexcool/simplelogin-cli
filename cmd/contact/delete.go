package contact

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete <contact-id>",
	Aliases: []string{"rm"},
	Short:   "Delete a contact",
	Long: `Delete a contact by its ID.

This removes the contact and its reverse alias. You will be prompted
for confirmation unless --yes is provided.

Use --dry-run to preview what would be deleted without actually deleting.

Use "sl contact list <alias>" to find contact IDs.`,
	Example: `  # Delete a contact (with confirmation)
  sl contact delete 456

  # Delete without confirmation
  sl contact delete 456 --yes

  # Preview what would be deleted
  sl contact delete 456 --dry-run

  # Delete and get JSON confirmation
  sl contact delete 456 --yes --json

  # Filter JSON output with jq
  sl contact delete 456 --yes --json --jq '.id'`,
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
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid contact ID %q: expected a numeric ID (use 'sl contact list <alias>' to find IDs)", args[0])
	}

	key, err := auth.GetAPIKey()
	if err != nil {
		return err
	}

	if deleteDryRun {
		if deleteJSON || deleteJQ != "" {
			data, _ := json.Marshal(map[string]interface{}{"id": id})
			if deleteJQ != "" {
				_ = output.PrintJQ(data, deleteJQ)
			} else {
				_ = output.PrintJSON(data)
			}
		} else {
			fmt.Printf("Would delete contact %d\n", id)
		}
		output.PrintWarning("No changes made. (dry-run)")
		return nil
	}

	if !deleteYes {
		if !output.ConfirmAction(fmt.Sprintf("Delete contact %d?", id)) {
			return fmt.Errorf("operation cancelled")
		}
	}

	client := api.NewClient(key, auth.GetAPIBase())
	if err := client.DeleteContact(id); err != nil {
		return err
	}

	if deleteJSON || deleteJQ != "" {
		data, _ := json.Marshal(map[string]interface{}{"deleted": true, "id": id})
		if deleteJQ != "" {
			return output.PrintJQ(data, deleteJQ)
		}
		return output.PrintJSON(data)
	}
	output.PrintSuccess("Contact deleted")
	output.PrintHint("sl contact list <alias-id>")
	fmt.Println(id)
	return nil
}
