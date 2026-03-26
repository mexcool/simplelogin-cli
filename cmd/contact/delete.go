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
	Use:   "delete <contact-id>",
	Short: "Delete a contact",
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
  sl contact delete 456 --yes --json`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

var (
	deleteYes    bool
	deleteJSON   bool
	deleteDryRun bool
)

func init() {
	deleteCmd.Flags().BoolVar(&deleteYes, "yes", false, "Skip confirmation prompt")
	deleteCmd.Flags().BoolVar(&deleteJSON, "json", false, "Output as JSON")
	deleteCmd.Flags().BoolVar(&deleteDryRun, "dry-run", false, "Preview without deleting")
}

func runDelete(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid contact ID: %s", args[0])
	}

	key, err := auth.GetAPIKey()
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	if deleteDryRun {
		if deleteJSON {
			data, _ := json.Marshal(map[string]interface{}{"id": id})
			fmt.Println(string(data))
		} else {
			fmt.Printf("Would delete contact %d\n", id)
		}
		output.PrintWarning("No changes made. (dry-run)")
		return nil
	}

	if !deleteYes {
		if !output.ConfirmAction(fmt.Sprintf("Delete contact %d?", id)) {
			output.PrintWarning("Cancelled")
			return nil
		}
	}

	client := api.NewClient(key)
	if err := client.DeleteContact(id); err != nil {
		output.PrintError("%v", err)
		return err
	}

	if deleteJSON {
		data, _ := json.Marshal(map[string]interface{}{"deleted": true, "id": id})
		fmt.Println(string(data))
		return nil
	}
	output.PrintSuccess("Contact deleted")
	fmt.Println(id)
	return nil
}
