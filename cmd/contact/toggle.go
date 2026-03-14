package contact

import (
	"fmt"
	"strconv"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var toggleCmd = &cobra.Command{
	Use:   "toggle <contact-id>",
	Short: "Block or unblock a contact",
	Long: `Toggle the block status of a contact.

When a contact is blocked, emails from that address to your alias will
be rejected. When unblocked, emails flow normally.

Use "sl contact list <alias>" to find contact IDs.`,
	Example: `  # Toggle block status of a contact
  sl contact toggle 456`,
	Args: cobra.ExactArgs(1),
	RunE: runToggle,
}

func runToggle(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid contact ID: %s", args[0])
	}

	key, err := auth.GetAPIKey()
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	client := api.NewClient(key)
	blocked, err := client.ToggleContact(id)
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	if blocked {
		output.PrintWarning("Contact %d is now blocked", id)
	} else {
		output.PrintSuccess("Contact %d is now unblocked", id)
	}
	return nil
}
