package domain

import (
	"fmt"
	"os"
	"strconv"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var trashCmd = &cobra.Command{
	Use:     "trash <domain-id>",
	Aliases: []string{"deleted"},
	Short:   "View deleted aliases for a domain",
	Long: `List aliases that have been deleted from a custom domain.

Shows the alias email and deletion date. These aliases can potentially
be recovered by contacting SimpleLogin support.`,
	Example: `  # View deleted aliases for a domain
  sl domain trash 123

  # Output as JSON
  sl domain trash 123 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runTrash,
}

var (
	trashJSON bool
	trashJQ   string
)

func init() {
	trashCmd.Flags().BoolVar(&trashJSON, "json", false, "Output as JSON")
	trashCmd.Flags().StringVar(&trashJQ, "jq", "", "Apply jq expression to JSON output")
}

func runTrash(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid domain ID %q: expected a numeric ID (use 'sl domain list' to find IDs)", args[0])
	}

	key, err := auth.GetAPIKey()
	if err != nil {
		return err
	}

	client := api.NewClient(key, auth.GetAPIBase())
	aliases, rawJSON, err := client.GetDomainTrash(id)
	if err != nil {
		return err
	}

	if trashJSON || trashJQ != "" {
		if trashJQ != "" {
			return output.PrintJQ(rawJSON, trashJQ)
		}
		return output.PrintJSON(rawJSON)
	}

	if len(aliases) == 0 {
		output.PrintWarning("No deleted aliases found")
		return nil
	}

	table := output.NewTable(os.Stdout, []string{"Alias", "Deleted"})
	for _, a := range aliases {
		table.Append([]string{a.Alias, a.DeletionDate})
	}
	table.Render()
	return nil
}
