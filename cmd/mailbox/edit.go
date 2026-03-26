package mailbox

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <mailbox-id>",
	Short: "Edit mailbox settings",
	Long: `Update settings for a mailbox.

You can set a mailbox as the default, change its email address, or
cancel a pending email change. Only specified fields are updated.

When changing the email, a verification email is sent to the new address.
Use --cancel-change to cancel a pending email change.`,
	Example: `  # Set a mailbox as default
  sl mailbox edit 456 --default

  # Change mailbox email
  sl mailbox edit 456 --email new-email@example.com

  # Cancel a pending email change
  sl mailbox edit 456 --cancel-change

  # Edit and return updated mailbox as JSON
  sl mailbox edit 456 --default --json`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

var (
	editDefault      bool
	editEmail        string
	editCancelChange bool
	editJSON         bool
	editJQ           string
)

func init() {
	editCmd.Flags().BoolVar(&editDefault, "default", false, "Set as default mailbox")
	editCmd.Flags().StringVar(&editEmail, "email", "", "Change mailbox email")
	editCmd.Flags().BoolVar(&editCancelChange, "cancel-change", false, "Cancel pending email change")
	editCmd.Flags().BoolVar(&editJSON, "json", false, "Output updated mailbox as JSON")
	editCmd.Flags().StringVar(&editJQ, "jq", "", "Apply jq expression to JSON output")
}

func runEdit(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid mailbox ID %q: expected a numeric ID (use 'sl mailbox list' to find IDs)", args[0])
	}

	key, err := auth.GetAPIKey()
	if err != nil {
		return err
	}

	req := &api.UpdateMailboxRequest{}
	hasChanges := false

	if editDefault {
		t := true
		req.Default = &t
		hasChanges = true
	}
	if cmd.Flags().Changed("email") {
		req.Email = &editEmail
		hasChanges = true
	}
	if editCancelChange {
		t := true
		req.CancelEmailChange = &t
		hasChanges = true
	}

	if !hasChanges {
		output.PrintWarning("No changes specified")
		return nil
	}

	client := api.NewClient(key)
	if err := client.UpdateMailbox(id, req); err != nil {
		return err
	}

	output.PrintSuccess("Mailbox updated")

	if editJQ != "" || editJSON {
		mailboxes, _, err := client.ListMailboxes()
		if err != nil {
			output.PrintWarning("Updated, but failed to fetch updated state: %v", err)
			return nil
		}
		var found *api.Mailbox
		for i := range mailboxes {
			if mailboxes[i].ID == id {
				found = &mailboxes[i]
				break
			}
		}
		if found == nil {
			output.PrintWarning("Updated, but mailbox ID %d not found in list", id)
			return nil
		}
		data, err := json.Marshal(found)
		if err != nil {
			return fmt.Errorf("failed to marshal mailbox: %w", err)
		}
		if editJQ != "" {
			return output.PrintJQ(data, editJQ)
		}
		return output.PrintJSON(data)
	}
	return nil
}
