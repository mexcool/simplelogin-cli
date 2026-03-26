package alias

import (
	"fmt"
	"os"

	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/mexcool/simplelogin-cli/internal/auth"
	"github.com/mexcool/simplelogin-cli/internal/output"
	"github.com/spf13/cobra"
)

var optionsCmd = &cobra.Command{
	Use:   "options",
	Short: "Show alias creation options (suffixes, domains)",
	Long: `Display the available options for creating a new alias.

This includes whether the account can create aliases, the suggested
prefix, and all available suffixes. Each suffix is annotated with
"(custom)" or "(premium)" where applicable.

Use --hostname to get options tailored for a specific website.`,
	Example: `  # Show available alias creation options
  sl alias options

  # Show options for a specific hostname
  sl alias options --hostname github.com

  # Output as JSON
  sl alias options --json

  # Filter with jq
  sl alias options --json --jq '.suffixes[].suffix'`,
	RunE: runOptions,
}

var (
	optionsHostname string
	optionsJSON     bool
	optionsJQ       string
)

func init() {
	optionsCmd.Flags().StringVar(&optionsHostname, "hostname", "", "Hostname to get tailored options for")
	optionsCmd.Flags().BoolVar(&optionsJSON, "json", false, "Output as JSON")
	optionsCmd.Flags().StringVar(&optionsJQ, "jq", "", "Apply jq expression to JSON output")
}

func runOptions(cmd *cobra.Command, args []string) error {
	key, err := auth.GetAPIKey()
	if err != nil {
		return err
	}

	client := api.NewClient(key, auth.GetAPIBase())

	opts, rawJSON, err := client.GetAliasOptions(optionsHostname)
	if err != nil {
		return err
	}

	if optionsJSON || optionsJQ != "" {
		if optionsJQ != "" {
			return output.PrintJQ(rawJSON, optionsJQ)
		}
		return output.PrintJSON(rawJSON)
	}

	// Human-readable output
	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Can create:"), output.BoolToStatus(opts.CanCreate))
	fmt.Fprintf(os.Stdout, "%s %s\n", output.Bold.Sprint("Prefix suggestion:"), opts.Prefixes)

	if len(opts.Suffixes) == 0 {
		fmt.Fprintln(os.Stdout, "\nNo suffixes available.")
		return nil
	}

	fmt.Fprintf(os.Stdout, "\n%s\n", output.Bold.Sprint("Suffixes:"))
	for _, s := range opts.Suffixes {
		tags := formatSuffixTags(s)
		fmt.Fprintf(os.Stdout, "  %s%s\n", s.Suffix, tags)
	}

	return nil
}

func formatSuffixTags(s api.SuffixOption) string {
	var tags string
	if s.IsCustom {
		tags += " (custom)"
	}
	if s.IsPremium {
		tags += " (premium)"
	}
	return tags
}
