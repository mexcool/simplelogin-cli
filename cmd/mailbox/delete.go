package mailbox

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <mailbox-id>",
	Short: "Delete a mailbox",
	Long: `Delete a mailbox from your SimpleLogin account.

You cannot delete the default mailbox. If the mailbox has aliases,
you can optionally transfer them to another mailbox using --transfer-to.

Use --dry-run to preview what would be deleted without actually deleting.

You will be prompted for confirmation unless --yes is provided.`,
	Example: `  # Delete a mailbox (with confirmation)
  sl mailbox delete 456

  # Delete and transfer aliases to another mailbox
  sl mailbox delete 456 --transfer-to 123

  # Delete without confirmation
  sl mailbox delete 456 --yes

  # Preview what would be deleted
  sl mailbox delete 456 --dry-run

  # Delete and get JSON confirmation
  sl mailbox delete 456 --yes --json

  # Filter JSON output with jq
  sl mailbox delete 456 --yes --json --jq '.id'`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

var (
	deleteTransferTo int
	deleteYes        bool
	deleteJSON       bool
	deleteJQ         string
	deleteDryRun     bool
)

func init() {
	deleteCmd.Flags().IntVar(&deleteTransferTo, "transfer-to", 0, "Transfer aliases to this mailbox ID")
	deleteCmd.Flags().BoolVar(&deleteYes, "yes", false, "Skip confirmation prompt")
	deleteCmd.Flags().BoolVar(&deleteJSON, "json", false, "Output as JSON")
	deleteCmd.Flags().StringVar(&deleteJQ, "jq", "", "Apply jq expression to JSON output")
	deleteCmd.Flags().BoolVar(&deleteDryRun, "dry-run", false, "Preview without deleting")
}

func runDelete(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid mailbox ID %q: expected a numeric ID (use 'sl mailbox list' to find IDs)", args[0])
	}

	key, err := auth.GetAPIKey()
	if err != nil {
		return err
	}

	client := api.NewClient(key)

	if deleteDryRun {
		mailboxes, _, err := client.ListMailboxes()
		if err != nil {
			return fmt.Errorf("failed to fetch mailboxes for preview: %w", err)
		}
		var found *api.Mailbox
		for i := range mailboxes {
			if mailboxes[i].ID == id {
				found = &mailboxes[i]
				break
			}
		}
		if found == nil {
			return fmt.Errorf("mailbox %d not found (use 'sl mailbox list' to see available mailboxes)", id)
		}
		if deleteJSON || deleteJQ != "" {
			data, _ := json.Marshal(found)
			if deleteJQ != "" {
				_ = output.PrintJQ(data, deleteJQ)
			} else {
				_ = output.PrintJSON(data)
			}
		} else {
			fmt.Fprintln(os.Stdout, "Would delete mailbox:")
			fmt.Fprintf(os.Stdout, "  ID:      %d\n", found.ID)
			fmt.Fprintf(os.Stdout, "  Email:   %s\n", found.Email)
			fmt.Fprintf(os.Stdout, "  Aliases: %d\n", found.NbAlias)
			fmt.Fprintf(os.Stdout, "  Default: %v\n", found.Default)
		}
		output.PrintWarning("No changes made. (dry-run)")
		return nil
	}

	if !deleteYes {
		if !output.ConfirmAction(fmt.Sprintf("Delete mailbox %d?", id)) {
			output.PrintWarning("Cancelled")
			return nil
		}
	}

	var transferTo *int
	if cmd.Flags().Changed("transfer-to") {
		transferTo = &deleteTransferTo
	}

	if err := client.DeleteMailbox(id, transferTo); err != nil {
		return err
	}

	if deleteJSON || deleteJQ != "" {
		data, _ := json.Marshal(map[string]interface{}{"deleted": true, "id": id})
		if deleteJQ != "" {
			return output.PrintJQ(data, deleteJQ)
		}
		return output.PrintJSON(data)
	}
	output.PrintSuccess("Mailbox deleted")
	fmt.Println(id)
	return nil
}
