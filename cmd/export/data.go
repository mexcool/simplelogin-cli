package export

import (
	"fmt"
	"os"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var dataCmd = &cobra.Command{
	Use:   "data",
	Short: "Export all account data as JSON",
	Long: `Export all your SimpleLogin account data as JSON.

This includes all aliases, contacts, mailboxes, and other account data.
By default, data is printed to stdout. Use --output to save to a file.`,
	Example: `  # Export data to stdout
  sl export data

  # Export data to a file
  sl export data --output backup.json

  # Pipe to jq for processing
  sl export data | jq '.aliases | length'`,
	RunE: runData,
}

var dataOutput string

func init() {
	dataCmd.Flags().StringVar(&dataOutput, "output", "", "Output file path")
}

func runData(cmd *cobra.Command, args []string) error {
	key, err := auth.GetAPIKey()
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	client := api.NewClient(key)
	data, err := client.ExportData()
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	if dataOutput != "" {
		if err := os.WriteFile(dataOutput, data, 0600); err != nil {
			output.PrintError("Failed to write file: %v", err)
			return err
		}
		output.PrintSuccess("Data exported to %s", dataOutput)
		return nil
	}

	fmt.Println(string(data))
	return nil
}
