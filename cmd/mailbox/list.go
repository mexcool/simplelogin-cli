package mailbox

import (
	"fmt"
	"os"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List mailboxes",
	Long: `List all mailboxes associated with your SimpleLogin account.

Shows the mailbox ID, email, verification status, default status,
and the number of aliases assigned to each mailbox.`,
	Example: `  # List all mailboxes
  sl mailbox list

  # List as JSON
  sl mailbox list --json

  # Get default mailbox email
  sl mailbox list --json --jq '.mailboxes[] | select(.default) | .email'`,
	RunE: runList,
}

var (
	listJSON bool
	listJQ   string
)

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")
	listCmd.Flags().StringVar(&listJQ, "jq", "", "Apply jq expression to JSON output")
}

func runList(cmd *cobra.Command, args []string) error {
	key, err := auth.GetAPIKey()
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	client := api.NewClient(key)
	mailboxes, rawJSON, err := client.ListMailboxes()
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	if listJSON || listJQ != "" {
		if listJQ != "" {
			return output.PrintJQ(rawJSON, listJQ)
		}
		return output.PrintJSON(rawJSON)
	}

	table := output.NewTable(os.Stdout, []string{"ID", "Email", "Verified", "Default", "Aliases"})
	for _, m := range mailboxes {
		table.Append([]string{
			fmt.Sprintf("%d", m.ID),
			m.Email,
			output.BoolToStatus(m.Verified),
			output.BoolToStatus(m.Default),
			fmt.Sprintf("%d", m.NbAlias),
		})
	}
	table.Render()
	return nil
}
