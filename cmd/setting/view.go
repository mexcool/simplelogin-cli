package setting

import (
	"fmt"
	"os"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "View current settings",
	Long: `Display your current SimpleLogin account settings.

Shows the alias generator mode, notification preference, default domain
for random aliases, sender format, and random alias suffix type.`,
	Example: `  # View settings
  sl setting view

  # View as JSON
  sl setting view --json

  # Filter JSON output with jq
  sl setting view --json --jq '.alias_generator'`,
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
	settings, rawJSON, err := client.GetSettings()
	if err != nil {
		return err
	}

	if viewJSON || viewJQ != "" {
		if viewJQ != "" {
			return output.PrintJQ(rawJSON, viewJQ)
		}
		return output.PrintJSON(rawJSON)
	}

	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Alias Generator:"), settings.AliasGenerator)
	fmt.Fprintf(os.Stdout, "%s %v\n", output.Bold.Sprint("Notifications:"), output.BoolToStatus(settings.Notification))
	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Default Domain:"), settings.RandomAliasDefaultDomain)
	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Sender Format:"), settings.SenderFormat)
	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Random Suffix:"), settings.RandomAliasSuffix)
	return nil
}
