package account

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

// Cmd is the account parent command.
var Cmd = &cobra.Command{
	Use:   "account",
	Short: "View account information",
	Long: `View your SimpleLogin account information and statistics.

Shows your user profile, subscription status, and usage statistics
including total aliases, forwards, blocks, and replies.`,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "View account status and stats",
	Long: `Display your account information and usage statistics.

Shows your name, email, premium status, and cumulative statistics
for all your aliases including total forwards, blocks, and replies.`,
	Example: `  # View account status
  sl account status

  # View as JSON
  sl account status --json`,
	RunE: runStatus,
}

var statusJSON bool

func init() {
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output as JSON")
	Cmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	key, err := auth.GetAPIKey()
	if err != nil {
		return err
	}

	client := api.NewClient(key)
	info, infoJSON, err := client.GetUserInfo()
	if err != nil {
		return err
	}

	stats, statsJSON, err := client.GetStats()
	if err != nil {
		return err
	}

	if statusJSON {
		// Merge info and stats into one JSON object
		var combined map[string]interface{}
		json.Unmarshal(infoJSON, &combined)
		var statsMap map[string]interface{}
		json.Unmarshal(statsJSON, &statsMap)
		for k, v := range statsMap {
			combined[k] = v
		}
		data, _ := json.MarshalIndent(combined, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	premium := "no"
	if info.IsPremium {
		premium = output.Green.Sprint("yes")
	}

	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Name:"), info.Name)
	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Email:"), info.Email)
	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Premium:"), premium)

	if info.InTrial {
		fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Trial:"), output.Yellow.Sprint("active"))
	}

	fmt.Fprintln(os.Stdout)
	fmt.Fprintf(os.Stdout, "%s\n", output.Bold.Sprint("Statistics:"))
	fmt.Fprintf(os.Stdout, "  Aliases:    %d\n", stats.NbAlias)
	fmt.Fprintf(os.Stdout, "  Forwarded:  %d\n", stats.NbForward)
	fmt.Fprintf(os.Stdout, "  Blocked:    %d\n", stats.NbBlock)
	fmt.Fprintf(os.Stdout, "  Replies:    %d\n", stats.NbReply)
	return nil
}
