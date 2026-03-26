package mailbox

import (
	"fmt"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add <email>",
	Aliases: []string{"create"},
	Short:   "Add a new mailbox",
	Long: `Add a new mailbox to your SimpleLogin account.

A verification email will be sent to the provided address. You must
verify the mailbox before it can be used to receive forwarded emails.`,
	Example: `  # Add a new mailbox
  sl mailbox add my-other-email@example.com

  # Add and output as JSON
  sl mailbox add my-other-email@example.com --json

  # Filter JSON output with jq
  sl mailbox add my-other-email@example.com --json --jq '.id'`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

var (
	addJSON bool
	addJQ   string
)

func init() {
	addCmd.Flags().BoolVar(&addJSON, "json", false, "Output as JSON")
	addCmd.Flags().StringVar(&addJQ, "jq", "", "Apply jq expression to JSON output")
}

func runAdd(cmd *cobra.Command, args []string) error {
	key, err := auth.GetAPIKey()
	if err != nil {
		return err
	}

	client := api.NewClient(key, auth.GetAPIBase())
	mailbox, rawJSON, err := client.CreateMailbox(args[0])
	if err != nil {
		return err
	}

	if addJSON || addJQ != "" {
		if addJQ != "" {
			return output.PrintJQ(rawJSON, addJQ)
		}
		return output.PrintJSON(rawJSON)
	}

	output.PrintSuccess("Mailbox added. Check %s for a verification email.", args[0])
	fmt.Println(mailbox.ID)
	return nil
}
