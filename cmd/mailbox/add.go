package mailbox

import (
	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <email>",
	Short: "Add a new mailbox",
	Long: `Add a new mailbox to your SimpleLogin account.

A verification email will be sent to the provided address. You must
verify the mailbox before it can be used to receive forwarded emails.`,
	Example: `  # Add a new mailbox
  sl mailbox add my-other-email@example.com`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

func runAdd(cmd *cobra.Command, args []string) error {
	key, err := auth.GetAPIKey()
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	client := api.NewClient(key)
	_, _, err = client.CreateMailbox(args[0])
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	output.PrintSuccess("Mailbox added. Check %s for a verification email.", args[0])
	return nil
}
