package contact

import (
	"fmt"
	"strconv"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var blockCmd = &cobra.Command{
	Use:   "block <contact-id>",
	Short: "Block a contact (idempotent)",
	Long: `Ensure a contact is blocked. If the contact is already blocked, this is a
no-op and exits successfully. This is safer than "toggle" for scripting
and agentic use because it is idempotent — calling it twice has the same
effect as calling it once.

When a contact is blocked, emails from that address to your alias will
be rejected.

Use "sl contact list <alias>" to find contact IDs.`,
	Example: `  # Block a contact
  sl contact block 456

  # Block and get result as JSON
  sl contact block 456 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runBlock,
}

var (
	blockJSON bool
	blockJQ   string
)

func init() {
	blockCmd.Flags().BoolVar(&blockJSON, "json", false, "Output as JSON")
	blockCmd.Flags().StringVar(&blockJQ, "jq", "", "Apply jq expression to JSON output")
}

func runBlock(cmd *cobra.Command, args []string) error {
	return setContactBlockState(args[0], true, blockJSON, blockJQ)
}

// setContactBlockState ensures a contact is in the desired blocked/unblocked state.
//
// The SimpleLogin API does not expose a "get single contact" endpoint, so we
// cannot read the current state before toggling. Instead we call toggle and
// check the resulting state. If the result is not what we want, we toggle once
// more. This means at most two API calls to guarantee the desired state, and is
// still idempotent from the caller's perspective.
func setContactBlockState(idStr string, wantBlocked bool, jsonFlag bool, jqExpr string) error {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return fmt.Errorf("invalid contact ID %q: expected a numeric ID (use 'sl contact list <alias>' to find IDs)", idStr)
	}

	key, err := auth.GetAPIKey()
	if err != nil {
		return err
	}

	client := api.NewClient(key, auth.GetAPIBase())

	blocked, rawJSON, err := client.ToggleContact(id)
	if err != nil {
		return err
	}

	// If the first toggle landed us in the wrong state, toggle again.
	if blocked != wantBlocked {
		blocked, rawJSON, err = client.ToggleContact(id)
		if err != nil {
			return err
		}
	}

	verb := "blocked"
	if !wantBlocked {
		verb = "unblocked"
	}

	if jqExpr != "" {
		return output.PrintJQ(rawJSON, jqExpr)
	}
	if jsonFlag {
		return output.PrintJSON(rawJSON)
	}

	if blocked {
		output.PrintWarning("Contact %d is now %s", id, verb)
	} else {
		output.PrintSuccess("Contact %d is now %s", id, verb)
	}
	fmt.Printf("blocked=%v\n", blocked)
	return nil
}
