package domain

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
	Short:   "List custom domains",
	Long: `List all custom domains registered with your SimpleLogin account.

Shows domain name, verification status, catch-all status, number of aliases,
and random prefix generation setting.`,
	Example: `  # List all custom domains
  sl domain list

  # List as JSON
  sl domain list --json

  # Get domain names
  sl domain list --json --jq '.custom_domains[].domain_name'`,
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
	listCmd.Flags().StringVar(&listFields, "fields", "", "Comma-separated columns to show (e.g. id,domain,verified)")
}

func runList(cmd *cobra.Command, args []string) error {
	key, err := auth.GetAPIKey()
	if err != nil {
		return err
	}

	client := api.NewClient(key, auth.GetAPIBase())
	domains, rawJSON, err := client.ListCustomDomains()
	if err != nil {
		return err
	}

	if listJSON || listJQ != "" {
		if listJQ != "" {
			return output.PrintJQ(rawJSON, listJQ)
		}
		return output.PrintJSON(rawJSON)
	}

	if len(domains) == 0 {
		output.PrintWarning("No custom domains found")
		return nil
	}

	headers := []string{"ID", "Domain", "Verified", "Catch-All", "Aliases", "Random Prefix"}
	indices := output.SelectColumns(headers, listFields)
	table := output.NewTable(os.Stdout, output.FilterRow(headers, indices))
	for _, d := range domains {
		row := []string{
			fmt.Sprintf("%d", d.ID),
			d.DomainName,
			output.BoolToStatus(d.Verified),
			output.BoolToStatus(d.CatchAll),
			fmt.Sprintf("%d", d.NbAlias),
			output.BoolToStatus(d.RandomPrefixGeneration),
		}
		table.Append(output.FilterRow(row, indices))
	}
	table.Render()
	return nil
}
