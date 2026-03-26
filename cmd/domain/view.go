package domain

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view <domain-id>",
	Short: "View custom domain details",
	Long: `Display detailed information about a specific custom domain.

Shows the domain name, verification status, catch-all and random prefix
settings, alias count, and linked mailboxes.`,
	Example: `  # View domain by ID
  sl domain view 123

  # View as JSON
  sl domain view 123 --json

  # Filter with jq
  sl domain view 123 --json --jq '.mailboxes[].email'`,
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
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid domain ID %q: expected a numeric ID (use 'sl domain list' to find IDs)", args[0])
	}

	key, err := auth.GetAPIKey()
	if err != nil {
		return err
	}

	client := api.NewClient(key, auth.GetAPIBase())
	domain, rawJSON, err := client.GetCustomDomain(id)
	if err != nil {
		return err
	}

	if viewJSON || viewJQ != "" {
		if viewJQ != "" {
			return output.PrintJQ(rawJSON, viewJQ)
		}
		return output.PrintJSON(rawJSON)
	}

	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Domain:"), domain.DomainName)
	fmt.Fprintf(os.Stdout, "%s %d\n", output.Bold.Sprint("ID:"), domain.ID)
	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Name:"), output.StringOrEmpty(domain.Name))
	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Created:"), domain.CreationDate)
	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Verified:"), output.BoolToStatus(domain.Verified))
	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Catch-all:"), output.EnabledStatus(domain.CatchAll))
	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Random prefix generation:"), output.EnabledStatus(domain.RandomPrefixGeneration))
	fmt.Fprintf(os.Stdout, "%s %d\n", output.Bold.Sprint("Aliases:"), domain.NbAlias)

	if len(domain.Mailboxes) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output.Bold.Sprint("Mailboxes:"))
		var entries []string
		for _, m := range domain.Mailboxes {
			entries = append(entries, fmt.Sprintf("  %s (ID %d)", m.Email, m.ID))
		}
		fmt.Fprintln(os.Stdout, strings.Join(entries, "\n"))
	}

	return nil
}
