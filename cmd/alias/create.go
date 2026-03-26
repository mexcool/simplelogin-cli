package alias

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new alias",
	Long: `Create a new SimpleLogin email alias.

By default, creates a custom alias. You must provide a --prefix and
select a --suffix from the available options. Use "sl alias create"
without flags to see available suffixes interactively.

Use --random to create a random alias instead, which requires no
prefix or suffix selection.

Optionally attach a note, display name, or assign to specific mailboxes.`,
	Example: `  # Create a random alias
  sl alias create --random

  # Create a random alias with a note
  sl alias create --random --note "Used for newsletters"

  # Create a custom alias (interactive suffix selection)
  sl alias create --prefix myalias

  # Create a custom alias with specific suffix
  sl alias create --prefix myalias --suffix 0 --mailbox 123

  # Output new alias as JSON
  sl alias create --random --json`,
	RunE: runCreate,
}

var (
	createPrefix  string
	createRandom  bool
	createSuffix  string
	createMailbox []int
	createNote    string
	createName    string
	createJSON    bool
)

func init() {
	createCmd.Flags().StringVar(&createPrefix, "prefix", "", "Alias prefix (for custom alias)")
	createCmd.Flags().BoolVar(&createRandom, "random", false, "Create a random alias")
	createCmd.Flags().StringVar(&createSuffix, "suffix", "", "Suffix index (from available options)")
	createCmd.Flags().IntSliceVar(&createMailbox, "mailbox", nil, "Mailbox IDs to assign")
	createCmd.Flags().StringVar(&createNote, "note", "", "Note for the alias")
	createCmd.Flags().StringVar(&createName, "name", "", "Display name for the alias")
	createCmd.Flags().BoolVar(&createJSON, "json", false, "Output as JSON")
}

func runCreate(cmd *cobra.Command, args []string) error {
	key, err := auth.GetAPIKey()
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	client := api.NewClient(key)

	if createRandom {
		alias, rawJSON, err := client.CreateRandomAlias(createNote)
		if err != nil {
			output.PrintError("%v", err)
			return err
		}

		if createJSON {
			return output.PrintJSON(rawJSON)
		}

		output.PrintSuccess("Created alias: %s", alias.Email)
		fmt.Println(alias.Email)
		return nil
	}

	// Custom alias creation
	if createPrefix == "" {
		return fmt.Errorf("--prefix is required for custom alias creation (or use --random)")
	}

	// Get available options
	opts, _, err := client.GetAliasOptions()
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	if len(opts.Suffixes) == 0 {
		output.PrintError("No suffixes available. You may need a premium account.")
		return fmt.Errorf("no suffixes available")
	}

	var selectedSuffix api.SuffixOption

	if createSuffix == "" {
		if !output.IsInteractive() {
			var hints []string
			for i, s := range opts.Suffixes {
				hints = append(hints, fmt.Sprintf("%d: %s%s", i, createPrefix, s.Suffix))
			}
			err := fmt.Errorf("--suffix is required in non-interactive mode. Available suffixes:\n  %s", strings.Join(hints, "\n  "))
			output.PrintError("%v", err)
			return err
		}
		// Interactive suffix selection
		fmt.Fprintln(os.Stderr, "Available suffixes:")
		for i, s := range opts.Suffixes {
			fmt.Fprintf(os.Stderr, "  [%d] %s%s\n", i, createPrefix, s.Suffix)
		}
		fmt.Fprint(os.Stderr, "\nSelect suffix number: ")
		var input string
		fmt.Scanln(&input)
		idx, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil || idx < 0 || idx >= len(opts.Suffixes) {
			return fmt.Errorf("invalid suffix selection")
		}
		selectedSuffix = opts.Suffixes[idx]
	} else {
		idx, err := strconv.Atoi(createSuffix)
		if err != nil || idx < 0 || idx >= len(opts.Suffixes) {
			return fmt.Errorf("invalid suffix index: %s (must be 0-%d)", createSuffix, len(opts.Suffixes)-1)
		}
		selectedSuffix = opts.Suffixes[idx]
	}

	// Get mailbox IDs
	mailboxIDs := createMailbox
	if len(mailboxIDs) == 0 {
		// Use default mailbox
		mailboxes, _, err := client.ListMailboxes()
		if err != nil {
			output.PrintError("Failed to get mailboxes: %v", err)
			return err
		}
		for _, m := range mailboxes {
			if m.Default {
				mailboxIDs = []int{m.ID}
				break
			}
		}
		if len(mailboxIDs) == 0 && len(mailboxes) > 0 {
			mailboxIDs = []int{mailboxes[0].ID}
		}
	}

	req := &api.CreateCustomAliasRequest{
		AliasPrefix:  createPrefix,
		SignedSuffix: selectedSuffix.SignedSuffix,
		MailboxIDs:   mailboxIDs,
		Note:         createNote,
		Name:         createName,
	}

	alias, rawJSON, err := client.CreateCustomAlias(req)
	if err != nil {
		output.PrintError("%v", err)
		return err
	}

	if createJSON {
		return output.PrintJSON(rawJSON)
	}

	output.PrintSuccess("Created alias: %s", alias.Email)
	fmt.Println(alias.Email)
	return nil
}
