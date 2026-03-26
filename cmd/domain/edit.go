package domain

import (
	"fmt"
	"strconv"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <domain-id>",
	Short: "Edit custom domain settings",
	Long: `Update settings for a custom domain.

You can enable or disable catch-all, toggle random prefix generation,
and set a display name. Only specified fields are updated.

Catch-all means any email sent to your domain (even to non-existing aliases)
will be forwarded. Random prefix generation adds a random prefix to
auto-created aliases on this domain.`,
	Example: `  # Enable catch-all for a domain
  sl domain edit 123 --catch-all

  # Disable catch-all
  sl domain edit 123 --no-catch-all

  # Enable random prefix generation
  sl domain edit 123 --random-prefix

  # Set a display name
  sl domain edit 123 --name "My Domain"

  # Assign specific mailboxes to a domain
  sl domain edit 123 --mailbox 1 --mailbox 2

  # Edit and return updated domain as JSON
  sl domain edit 123 --catch-all --json`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

var (
	editCatchAll       bool
	editNoCatchAll     bool
	editRandomPrefix   bool
	editNoRandomPrefix bool
	editName           string
	editMailbox        []int
	editJSON           bool
	editJQ             string
)

func init() {
	editCmd.Flags().BoolVar(&editCatchAll, "catch-all", false, "Enable catch-all")
	editCmd.Flags().BoolVar(&editNoCatchAll, "no-catch-all", false, "Disable catch-all")
	editCmd.Flags().BoolVar(&editRandomPrefix, "random-prefix", false, "Enable random prefix generation")
	editCmd.Flags().BoolVar(&editNoRandomPrefix, "no-random-prefix", false, "Disable random prefix generation")
	editCmd.Flags().StringVar(&editName, "name", "", "Set display name")
	editCmd.Flags().IntSliceVar(&editMailbox, "mailbox", nil, "Mailbox IDs to assign (can be repeated)")
	editCmd.Flags().BoolVar(&editJSON, "json", false, "Output updated domain as JSON")
	editCmd.Flags().StringVar(&editJQ, "jq", "", "Apply jq expression to JSON output")
}

func runEdit(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid domain ID %q: expected a numeric ID (use 'sl domain list' to find IDs)", args[0])
	}

	key, err := auth.GetAPIKey()
	if err != nil {
		return err
	}

	req := &api.UpdateDomainRequest{}
	hasChanges := false

	if editCatchAll {
		t := true
		req.CatchAll = &t
		hasChanges = true
	}
	if editNoCatchAll {
		f := false
		req.CatchAll = &f
		hasChanges = true
	}
	if editRandomPrefix {
		t := true
		req.RandomPrefixGeneration = &t
		hasChanges = true
	}
	if editNoRandomPrefix {
		f := false
		req.RandomPrefixGeneration = &f
		hasChanges = true
	}
	if cmd.Flags().Changed("name") {
		req.Name = &editName
		hasChanges = true
	}
	if len(editMailbox) > 0 {
		req.MailboxIDs = editMailbox
		hasChanges = true
	}

	if !hasChanges {
		output.PrintWarning("No changes specified")
		return nil
	}

	client := api.NewClient(key, auth.GetAPIBase())
	if err := client.UpdateCustomDomain(id, req); err != nil {
		return err
	}

	output.PrintSuccess("Domain updated")

	if editJQ != "" || editJSON {
		_, rawJSON, err := client.GetCustomDomain(id)
		if err != nil {
			output.PrintWarning("Updated, but failed to fetch updated state: %v", err)
			return nil
		}
		if editJQ != "" {
			return output.PrintJQ(rawJSON, editJQ)
		}
		return output.PrintJSON(rawJSON)
	}
	return nil
}
