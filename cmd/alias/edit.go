package alias

import (
	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <alias-id-or-email>",
	Short: "Edit alias properties",
	Long: `Update properties of an existing alias.

You can modify the alias name, note, assigned mailboxes, and pinned status.
Only the specified fields are updated; unspecified fields remain unchanged.

Accepts either a numeric alias ID or the full alias email address.`,
	Example: `  # Set a note on an alias
  sl alias edit 12345 --note "Used for newsletters"

  # Pin an alias
  sl alias edit 12345 --pin

  # Change display name and mailbox
  sl alias edit my-alias@simplelogin.co --name "My Alias" --mailbox 456

  # Unpin an alias
  sl alias edit 12345 --unpin

  # Edit and return updated alias as JSON
  sl alias edit 12345 --note "test" --json`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

var (
	editName    string
	editNote    string
	editMailbox []int
	editPin     bool
	editUnpin   bool
	editJSON    bool
	editJQ      string
)

func init() {
	editCmd.Flags().StringVar(&editName, "name", "", "Set display name")
	editCmd.Flags().StringVar(&editNote, "note", "", "Set note")
	editCmd.Flags().IntSliceVar(&editMailbox, "mailbox", nil, "Set mailbox IDs")
	editCmd.Flags().BoolVar(&editPin, "pin", false, "Pin the alias")
	editCmd.Flags().BoolVar(&editUnpin, "unpin", false, "Unpin the alias")
	editCmd.Flags().BoolVar(&editJSON, "json", false, "Output updated alias as JSON")
	editCmd.Flags().StringVar(&editJQ, "jq", "", "Apply jq expression to JSON output")
}

func runEdit(cmd *cobra.Command, args []string) error {
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

	req := &api.UpdateAliasRequest{}
	hasChanges := false

	if cmd.Flags().Changed("name") {
		req.Name = &editName
		hasChanges = true
	}
	if cmd.Flags().Changed("note") {
		req.Note = &editNote
		hasChanges = true
	}
	if cmd.Flags().Changed("mailbox") {
		req.MailboxIDs = editMailbox
		hasChanges = true
	}
	if editPin {
		pinTrue := true
		req.Pinned = &pinTrue
		hasChanges = true
	}
	if editUnpin {
		pinFalse := false
		req.Pinned = &pinFalse
		hasChanges = true
	}

	if !hasChanges {
		output.PrintWarning("No changes specified")
		return nil
	}

	if err := client.UpdateAlias(id, req); err != nil {
		output.PrintError("%v", err)
		return err
	}

	output.PrintSuccess("Alias updated")

	if editJQ != "" || editJSON {
		_, rawJSON, err := client.GetAlias(id)
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
