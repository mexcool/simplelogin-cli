package contact

import (
	"fmt"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add <alias-id-or-email> <contact-email>",
	Aliases: []string{"create"},
	Short:   "Add a contact to an alias",
	Long: `Add a contact email to an alias, creating a reverse alias.

The reverse alias can be used to send emails from your alias to the
contact. When you send an email to the reverse alias address, SimpleLogin
will forward it to the contact from your alias email.

This is useful for initiating conversations from your alias without
revealing your real email address.`,
	Example: `  # Add a contact to an alias
  sl contact add 12345 friend@example.com

  # Add a contact using alias email
  sl contact add my-alias@simplelogin.co friend@example.com

  # Get reverse alias as JSON
  sl contact add 12345 friend@example.com --json`,
	Args: cobra.ExactArgs(2),
	RunE: runAdd,
}

var addJSON bool

func init() {
	addCmd.Flags().BoolVar(&addJSON, "json", false, "Output as JSON")
}

func runAdd(cmd *cobra.Command, args []string) error {
	key, err := auth.GetAPIKey()
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	client := api.NewClient(key)
	aliasID, err := client.ResolveAliasID(args[0])
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	contact, rawJSON, err := client.CreateContact(aliasID, args[1])
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	if addJSON {
		return output.PrintJSON(rawJSON)
	}

	if contact.Existed {
		output.PrintWarning("Contact already exists")
	} else {
		output.PrintSuccess("Contact added")
	}
	fmt.Printf("Reverse alias: %s\n", contact.ReverseAliasAddress)
	return nil
}
