package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mexcool/simplelogin-cli/cmd/account"
	"github.com/mexcool/simplelogin-cli/cmd/alias"
	cmdauth "github.com/mexcool/simplelogin-cli/cmd/auth"
	"github.com/mexcool/simplelogin-cli/cmd/contact"
	"github.com/mexcool/simplelogin-cli/cmd/domain"
	"github.com/mexcool/simplelogin-cli/cmd/export"
	"github.com/mexcool/simplelogin-cli/cmd/mailbox"
	"github.com/mexcool/simplelogin-cli/cmd/setting"
	"github.com/mexcool/simplelogin-cli/internal/api"
	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "sl",
	Short: "SimpleLogin CLI - manage email aliases from the terminal",
	Long: `SimpleLogin CLI (sl) is a command-line interface for the SimpleLogin
email alias service. It allows you to create, manage, and monitor email
aliases, contacts, mailboxes, and domains directly from your terminal.

Authentication:
  The CLI looks for your API key in this order:
  1. SIMPLELOGIN_API_KEY environment variable
  2. SL_API_KEY environment variable
  3. 1Password CLI (if configured via sl auth login --1password)
  4. Config file ($XDG_CONFIG_HOME/simplelogin/config.yml, defaults to ~/.config/simplelogin/config.yml)

  Run "sl auth login" to get started.

Output:
  By default, commands display colored table output. Use --json for
  machine-readable JSON output, or --jq to filter JSON with jq expressions.`,
	SilenceUsage: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose || os.Getenv("SL_VERBOSE") == "1" || os.Getenv("SL_DEBUG") == "1" {
			api.Verbose = true
		}
	},
}

// SetVersionInfo sets the version string shown by --version, including
// optional build metadata (commit hash and build date).
func SetVersionInfo(v, commit, date string) {
	display := v
	if commit != "" || date != "" {
		parts := make([]string, 0, 2)
		if commit != "" {
			parts = append(parts, commit)
		}
		if date != "" {
			parts = append(parts, date)
		}
		display = fmt.Sprintf("%s (%s)", v, strings.Join(parts, ", "))
	}
	rootCmd.Version = display
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// RootCmd returns the root cobra.Command, used by tooling such as man page generation.
func RootCmd() *cobra.Command {
	return rootCmd
}

var completionCmd = &cobra.Command{
	Use:       "completion [bash|zsh|fish|powershell]",
	Short:     "Generate shell completion scripts",
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Example: `  # Bash (add to ~/.bashrc)
  source <(sl completion bash)

  # Zsh (add to ~/.zshrc)
  source <(sl completion zsh)

  # Fish
  sl completion fish | source`,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletionV2(os.Stdout, true)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Log HTTP requests to stderr (method, URL, status, latency)")

	rootCmd.AddCommand(cmdauth.Cmd)
	rootCmd.AddCommand(alias.Cmd)
	rootCmd.AddCommand(contact.Cmd)
	rootCmd.AddCommand(mailbox.Cmd)
	rootCmd.AddCommand(domain.Cmd)
	rootCmd.AddCommand(setting.Cmd)
	rootCmd.AddCommand(account.Cmd)
	rootCmd.AddCommand(export.Cmd)
	rootCmd.AddCommand(completionCmd)
}
