package alias

import (
	"fmt"
	"os"
	"strings"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view <alias-id-or-email>",
	Short: "View alias details",
	Long: `Display detailed information about a specific alias.

Accepts either a numeric alias ID or the full alias email address.
When using an email address, the CLI searches for the matching alias.

Shows the alias email, status, statistics, mailboxes, creation date,
and other metadata.`,
	Example: `  # View alias by ID
  sl alias view 12345

  # View alias by email
  sl alias view my-alias@simplelogin.co

  # View as JSON
  sl alias view 12345 --json

  # Filter with jq
  sl alias view 12345 --json --jq '.mailboxes[].email'`,
	Args: cobra.ExactArgs(1),
	RunE: runView,
}

var (
	viewJSON bool
	viewJQ   string
)

func init() {
	viewCmd.Flags().BoolVar(&viewJSON, "json", false, "Output as JSON")
	viewCmd.Flags().StringVar(&viewJQ, "jq", "", "Apply jq expression to JSON output")
}

func runView(cmd *cobra.Command, args []string) error {
	key, err := auth.GetAPIKey()
	if err != nil {
		return err
	}

	client := api.NewClient(key)
	id, err := client.ResolveAliasID(args[0])
	if err != nil {
		return err
	}

	alias, rawJSON, err := client.GetAlias(id)
	if err != nil {
		return err
	}

	if viewJSON || viewJQ != "" {
		if viewJQ != "" {
			return output.PrintJQ(rawJSON, viewJQ)
		}
		return output.PrintJSON(rawJSON)
	}

	// Detailed view
	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Email:"), alias.Email)
	fmt.Fprintf(os.Stdout, "%s %d\n", output.Bold.Sprint("ID:"), alias.ID)
	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Status:"), output.EnabledStatus(alias.Enabled))
	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Created:"), alias.CreationDate)
	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Name:"), output.StringOrEmpty(alias.Name))
	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Note:"), output.StringOrEmpty(alias.Note))
	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Pinned:"), output.BoolToStatus(alias.Pinned))
	fmt.Fprintf(os.Stdout, "%s %d forwarded, %d blocked, %d replies\n",
		output.Bold.Sprint("Stats:"), alias.NbForward, alias.NbBlock, alias.NbReply)

	if len(alias.Mailboxes) > 0 {
		var emails []string
		for _, m := range alias.Mailboxes {
			emails = append(emails, m.Email)
		}
		fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Mailboxes:"), strings.Join(emails, ", "))
	}

	return nil
}
