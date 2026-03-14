package setting

import (
	"os"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var domainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "List available domains for alias creation",
	Long: `List all domains available for creating new aliases.

This includes both SimpleLogin's shared domains and any custom domains
you have registered. Shows whether each domain is a custom domain.`,
	Example: `  # List available domains
  sl setting domains

  # List as JSON
  sl setting domains --json

  # Get just domain names
  sl setting domains --json --jq '.[].domain'`,
	RunE: runDomains,
}

var (
	domainsJSON bool
	domainsJQ   string
)

func init() {
	domainsCmd.Flags().BoolVar(&domainsJSON, "json", false, "Output as JSON")
	domainsCmd.Flags().StringVar(&domainsJQ, "jq", "", "Apply jq expression to JSON output")
}

func runDomains(cmd *cobra.Command, args []string) error {
	key, err := auth.GetAPIKey()
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	client := api.NewClient(key)
	domains, rawJSON, err := client.GetAvailableDomains()
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	if domainsJSON || domainsJQ != "" {
		if domainsJQ != "" {
			return output.PrintJQ(rawJSON, domainsJQ)
		}
		return output.PrintJSON(rawJSON)
	}

	table := output.NewTable(os.Stdout, []string{"Domain", "Custom"})
	for _, d := range domains {
		table.Append([]string{d.Domain, output.BoolToStatus(d.IsCustom)})
	}
	table.Render()
	return nil
}
