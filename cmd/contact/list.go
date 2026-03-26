package contact

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list <alias-id-or-email>",
	Aliases: []string{"ls"},
	Short:   "List contacts for an alias",
	Long: `List all contacts associated with a specific alias.

Shows the contact email, reverse alias address (for sending mail as the alias),
creation date, and blocking status. Use --all to fetch all pages.

Accepts either a numeric alias ID or the full alias email address.`,
	Example: `  # List contacts for an alias
  sl contact list 12345

  # List all contacts
  sl contact list my-alias@simplelogin.co --all

  # Output as JSON
  sl contact list 12345 --json

  # Get reverse aliases with jq
  sl contact list 12345 --all --json --jq '.contacts[] | {contact, reverse_alias}'`,
	Args: cobra.ExactArgs(1),
	RunE: runList,
}

var (
	listPage int
	listAll  bool
	listJSON bool
	listJQ   string
)

func init() {
	listCmd.Flags().IntVar(&listPage, "page", 1, "Page number (1-indexed)")
	listCmd.Flags().BoolVar(&listAll, "all", false, "Fetch all pages")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")
	listCmd.Flags().StringVar(&listJQ, "jq", "", "Apply jq expression to JSON output")
}

func runList(cmd *cobra.Command, args []string) error {
	key, err := auth.GetAPIKey()
	if err != nil {
		return err
	}

	client := api.NewClient(key)
	aliasID, err := client.ResolveAliasID(args[0])
	if err != nil {
		return err
	}

	if listAll {
		contacts, err := client.GetAllAliasContacts(aliasID)
		if err != nil {
			return err
		}

		if listJSON || listJQ != "" {
			data, _ := json.Marshal(map[string]interface{}{"contacts": contacts})
			if listJQ != "" {
				return output.PrintJQ(data, listJQ)
			}
			return output.PrintJSON(data)
		}

		printContactTable(contacts)
		return nil
	}

	contacts, rawJSON, err := client.GetAliasContacts(aliasID, listPage-1)
	if err != nil {
		return err
	}

	if listJSON || listJQ != "" {
		if listJQ != "" {
			return output.PrintJQ(rawJSON, listJQ)
		}
		return output.PrintJSON(rawJSON)
	}

	printContactTable(contacts)
	return nil
}

func printContactTable(contacts []api.Contact) {
	if len(contacts) == 0 {
		output.PrintWarning("No contacts found")
		return
	}

	table := output.NewTable(os.Stdout, []string{"ID", "Contact", "Reverse Alias", "Blocked", "Created"})
	for _, c := range contacts {
		table.Append([]string{
			fmt.Sprintf("%d", c.ID),
			c.Contact,
			output.Truncate(c.ReverseAliasAddress, 50),
			output.BoolToStatus(c.BlockForward),
			c.CreationDate,
		})
	}
	table.Render()
}
