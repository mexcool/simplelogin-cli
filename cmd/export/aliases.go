package export

import (
	"fmt"
	"os"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var aliasesCmd = &cobra.Command{
	Use:   "aliases",
	Short: "Export aliases as CSV",
	Long: `Export all your SimpleLogin aliases as CSV.

The CSV includes alias email, creation date, status, and other metadata.
By default, data is printed to stdout. Use --output to save to a file.`,
	Example: `  # Export aliases to stdout
  sl export aliases

  # Export aliases to a file
  sl export aliases --output aliases.csv`,
	RunE: runAliases,
}

var aliasesOutput string

func init() {
	aliasesCmd.Flags().StringVar(&aliasesOutput, "output", "", "Output file path")
}

func runAliases(cmd *cobra.Command, args []string) error {
	key, err := auth.GetAPIKey()
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	client := api.NewClient(key)
	data, err := client.ExportAliases()
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	if aliasesOutput != "" {
		if err := os.WriteFile(aliasesOutput, data, 0600); err != nil {
			output.PrintError("Failed to write file: %v", err)
			return err
		}
		output.PrintSuccess("Aliases exported to %s", aliasesOutput)
		return nil
	}

	fmt.Println(string(data))
	return nil
}
