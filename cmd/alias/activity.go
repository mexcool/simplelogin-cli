package alias

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var activityCmd = &cobra.Command{
	Use:   "activity <alias-id-or-email>",
	Short: "View alias activity log",
	Long: `Display the activity log for an alias, showing forwards, blocks, and replies.

Each activity entry shows the action type, timestamp, sender, and recipient.
Use --all to fetch all activities across all pages.

Accepts either a numeric alias ID or the full alias email address.`,
	Example: `  # View recent activity for an alias
  sl alias activity 12345

  # View all activity
  sl alias activity 12345 --all

  # Output as JSON
  sl alias activity 12345 --json

  # Filter with jq
  sl alias activity 12345 --all --json --jq '.activities[] | select(.action == "forward")'`,
	Args: cobra.ExactArgs(1),
	RunE: runActivity,
}

var (
	activityPage int
	activityAll  bool
	activityJSON bool
	activityJQ   string
)

func init() {
	activityCmd.Flags().IntVar(&activityPage, "page", 1, "Page number (1-indexed)")
	activityCmd.Flags().BoolVar(&activityAll, "all", false, "Fetch all pages")
	activityCmd.Flags().BoolVar(&activityJSON, "json", false, "Output as JSON")
	activityCmd.Flags().StringVar(&activityJQ, "jq", "", "Apply jq expression to JSON output")
}

func runActivity(cmd *cobra.Command, args []string) error {
	key, err := auth.GetAPIKey()
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	client := api.NewClient(key)
	id, err := client.ResolveAliasID(args[0])
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	if activityAll {
		activities, err := client.GetAllAliasActivities(id)
		if err != nil {
			output.PrintError("%v", err)
			return err
		}

		if activityJSON || activityJQ != "" {
			data, _ := json.Marshal(map[string]interface{}{"activities": activities})
			if activityJQ != "" {
				return output.PrintJQ(data, activityJQ)
			}
			return output.PrintJSON(data)
		}

		printActivityTable(activities)
		return nil
	}

	activities, rawJSON, err := client.GetAliasActivities(id, activityPage-1)
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	if activityJSON || activityJQ != "" {
		if activityJQ != "" {
			return output.PrintJQ(rawJSON, activityJQ)
		}
		return output.PrintJSON(rawJSON)
	}

	printActivityTable(activities)
	return nil
}

func printActivityTable(activities []api.Activity) {
	if len(activities) == 0 {
		output.PrintWarning("No activities found")
		return
	}

	table := output.NewTable(os.Stdout, []string{"Action", "From", "To", "Time"})
	for _, a := range activities {
		ts := time.Unix(a.Timestamp, 0).Format("2006-01-02 15:04")
		table.Append([]string{
			a.Action,
			output.Truncate(a.From, 40),
			output.Truncate(a.To, 40),
			ts,
		})
	}
	table.Render()
	fmt.Fprintf(os.Stderr, "\n%d activities shown\n", len(activities))
}
