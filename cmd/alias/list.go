package alias

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List aliases",
	Long: `List your SimpleLogin email aliases with optional filtering.

By default, lists the first page of aliases (20 per page). Use --all
to fetch all aliases across all pages. Filter by status with --enabled,
--disabled, or --pinned. Search with --query to match alias emails.

Results are displayed as a colored table by default. Use --json for
machine-readable output or --jq to filter the JSON response.`,
	Example: `  # List all aliases
  sl alias list --all

  # List only enabled aliases
  sl alias list --enabled

  # Search for aliases containing "github"
  sl alias list --query github

  # List as JSON and filter with jq
  sl alias list --all --json --jq '.aliases[] | .email'

  # List page 2 (1-indexed)
  sl alias list --page 2`,
	RunE: runList,
}

var (
	listEnabled  bool
	listDisabled bool
	listPinned   bool
	listQuery    string
	listPage     int
	listAll      bool
	listJSON     bool
	listJQ       string
)

func init() {
	listCmd.Flags().BoolVar(&listEnabled, "enabled", false, "Show only enabled aliases")
	listCmd.Flags().BoolVar(&listDisabled, "disabled", false, "Show only disabled aliases")
	listCmd.Flags().BoolVar(&listPinned, "pinned", false, "Show only pinned aliases")
	listCmd.Flags().StringVar(&listQuery, "query", "", "Search query to filter aliases")
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

	client := api.NewClient(key, auth.GetAPIBase())

	if listAll {
		aliases, err := client.ListAllAliases(listPinned, listDisabled, listEnabled, listQuery)
		if err != nil {
			return err
		}

		if listJSON || listJQ != "" {
			data, _ := json.Marshal(map[string]interface{}{"aliases": aliases})
			if listJQ != "" {
				return output.PrintJQ(data, listJQ)
			}
			return output.PrintJSON(data)
		}

		printAliasTable(aliases)
		output.PrintSuccess("\nTotal: %d aliases", len(aliases))
		return nil
	}

	aliases, rawJSON, err := client.ListAliases(listPage-1, listPinned, listDisabled, listEnabled, listQuery)
	if err != nil {
		return err
	}

	if listJSON || listJQ != "" {
		if listJQ != "" {
			return output.PrintJQ(rawJSON, listJQ)
		}
		return output.PrintJSON(rawJSON)
	}

	printAliasTable(aliases)
	output.PrintSuccess("\nPage %d, %d aliases shown", listPage, len(aliases))
	return nil
}

func printAliasTable(aliases []api.Alias) {
	table := output.NewTable(os.Stdout, []string{"ID", "Email", "Status", "Fwd", "Block", "Reply", "Pinned", "Note"})
	for _, a := range aliases {
		note := output.StringOrEmpty(a.Note)
		note = output.Truncate(note, 30)
		table.Append([]string{
			strconv.Itoa(a.ID),
			a.Email,
			output.EnabledStatus(a.Enabled),
			fmt.Sprintf("%d", a.NbForward),
			fmt.Sprintf("%d", a.NbBlock),
			fmt.Sprintf("%d", a.NbReply),
			output.BoolToStatus(a.Pinned),
			note,
		})
	}
	table.Render()
}
