package mailbox

import (
	"fmt"
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

You will be prompted for confirmation unless --yes is provided.`,
	Example: `  # Delete a mailbox (with confirmation)
  sl mailbox delete 456

  # Delete and transfer aliases to another mailbox
  sl mailbox delete 456 --transfer-to 123

  # Delete without confirmation
  sl mailbox delete 456 --yes`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

var (
	deleteTransferTo int
	deleteYes        bool
)

func init() {
	deleteCmd.Flags().IntVar(&deleteTransferTo, "transfer-to", 0, "Transfer aliases to this mailbox ID")
	deleteCmd.Flags().BoolVar(&deleteYes, "yes", false, "Skip confirmation prompt")
}

func runDelete(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid mailbox ID: %s", args[0])
	}

	key, err := auth.GetAPIKey()
	if err != nil {
		output.PrintError("%v", err)
		return err
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

	client := api.NewClient(key)
	if err := client.DeleteMailbox(id, transferTo); err != nil {
		output.PrintError("%v", err)
		return err
	}

	output.PrintSuccess("Mailbox deleted")
	return nil
}
