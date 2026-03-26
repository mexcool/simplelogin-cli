package setting

import (
	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Update account settings",
	Long: `Update your SimpleLogin account settings.

Available options:
  --generator: Alias generation mode ("word" or "uuid")
  --sender-format: How sender addresses appear ("AT", "A", "NAME_ONLY", "AT_ONLY", "NO_NAME")
  --notifications: Email notifications ("on" or "off")
  --default-domain: Default domain for random aliases
  --suffix-type: Random alias suffix type ("word" or "random_string")`,
	Example: `  # Change alias generator to UUID mode
  sl setting edit --generator uuid

  # Set sender format
  sl setting edit --sender-format NAME_ONLY

  # Disable notifications
  sl setting edit --notifications off

  # Set default domain and suffix type
  sl setting edit --default-domain simplelogin.co --suffix-type word

  # Edit and return updated settings as JSON
  sl setting edit --generator uuid --json`,
	RunE: runEdit,
}

var (
	editGenerator     string
	editSenderFormat  string
	editNotifications string
	editDefaultDomain string
	editSuffixType    string
	editJSON          bool
	editJQ            string
)

func init() {
	editCmd.Flags().StringVar(&editGenerator, "generator", "", "Alias generator: word or uuid")
	editCmd.Flags().StringVar(&editSenderFormat, "sender-format", "", "Sender format: AT, A, NAME_ONLY, AT_ONLY, NO_NAME")
	editCmd.Flags().StringVar(&editNotifications, "notifications", "", "Notifications: on or off")
	editCmd.Flags().StringVar(&editDefaultDomain, "default-domain", "", "Default domain for random aliases")
	editCmd.Flags().StringVar(&editSuffixType, "suffix-type", "", "Suffix type: word or random_string")
	editCmd.Flags().BoolVar(&editJSON, "json", false, "Output updated settings as JSON")
	editCmd.Flags().StringVar(&editJQ, "jq", "", "Apply jq expression to JSON output")
}

func runEdit(cmd *cobra.Command, args []string) error {
	key, err := auth.GetAPIKey()
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	req := &api.UpdateSettingsRequest{}
	hasChanges := false

	if cmd.Flags().Changed("generator") {
		req.AliasGenerator = &editGenerator
		hasChanges = true
	}
	if cmd.Flags().Changed("sender-format") {
		req.SenderFormat = &editSenderFormat
		hasChanges = true
	}
	if cmd.Flags().Changed("notifications") {
		val := editNotifications == "on"
		req.Notification = &val
		hasChanges = true
	}
	if cmd.Flags().Changed("default-domain") {
		req.RandomAliasDefaultDomain = &editDefaultDomain
		hasChanges = true
	}
	if cmd.Flags().Changed("suffix-type") {
		req.RandomAliasSuffix = &editSuffixType
		hasChanges = true
	}

	if !hasChanges {
		output.PrintWarning("No changes specified")
		return nil
	}

	client := api.NewClient(key)
	if err := client.UpdateSettings(req); err != nil {
		output.PrintError("%v", err)
		return err
	}

	output.PrintSuccess("Settings updated")

	if editJQ != "" || editJSON {
		_, rawJSON, err := client.GetSettings()
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
