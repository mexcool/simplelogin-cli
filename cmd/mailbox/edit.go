package mailbox

import (
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
  sl mailbox edit 456 --cancel-change`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

var (
	editDefault     bool
	editEmail       string
	editCancelChange bool
)

func init() {
	editCmd.Flags().BoolVar(&editDefault, "default", false, "Set as default mailbox")
	editCmd.Flags().StringVar(&editEmail, "email", "", "Change mailbox email")
	editCmd.Flags().BoolVar(&editCancelChange, "cancel-change", false, "Cancel pending email change")
}

func runEdit(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid mailbox ID: %s", args[0])
	}

	key, err := auth.GetAPIKey()
	if err != nil {
		output.PrintError("%v", err)
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
		output.PrintError("%v", err)
		return err
	}

	output.PrintSuccess("Mailbox updated")
	return nil
}
