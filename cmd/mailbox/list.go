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
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List mailboxes",
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
	listJSON   bool
	listJQ     string
	listFields string
)

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")
	listCmd.Flags().StringVar(&listJQ, "jq", "", "Apply jq expression to JSON output")
	listCmd.Flags().StringVar(&listFields, "fields", "", "Comma-separated columns to show (e.g. id,email,verified)")
}

func runList(cmd *cobra.Command, args []string) error {
	key, err := auth.GetAPIKey()
	if err != nil {
		return err
	}

	client := api.NewClient(key, auth.GetAPIBase())
	mailboxes, rawJSON, err := client.ListMailboxes()
	if err != nil {
		return err
	}

	if listJSON || listJQ != "" {
		if listJQ != "" {
			return output.PrintJQ(rawJSON, listJQ)
		}
		return output.PrintJSON(rawJSON)
	}

	if len(mailboxes) == 0 {
		output.PrintWarning("No mailboxes found")
		output.PrintHint("sl mailbox add <email>")
		return nil
	}

	headers := []string{"ID", "Email", "Verified", "Default", "Aliases"}
	indices := output.SelectColumns(headers, listFields)
	table := output.NewTable(os.Stdout, output.FilterRow(headers, indices))
	for _, m := range mailboxes {
		row := []string{
			fmt.Sprintf("%d", m.ID),
			m.Email,
			output.BoolToStatus(m.Verified),
			output.BoolToStatus(m.Default),
			fmt.Sprintf("%d", m.NbAlias),
		}
		table.Append(output.FilterRow(row, indices))
	}
	table.Render()
	return nil
}
