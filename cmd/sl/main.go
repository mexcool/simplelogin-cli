package main

import (
	"os"
	"runtime/debug"
	"strings"

	"github.com/mexcool/simplelogin-cli/cmd"
)

// These variables are set at build time via ldflags.
var (
	version = "dev"
	commit  = ""
	date    = ""
)

func main() {
	// Fallback: when installed via `go install`, ldflags are not set.
	// Use runtime/debug.ReadBuildInfo to recover version and VCS metadata.
	// We only override version from build info when no ldflags were injected
	// at all (commit is empty), to avoid replacing an intentional "dev" value.
	if info, ok := debug.ReadBuildInfo(); ok {
		if version == "dev" && commit == "" && info.Main.Version != "" {
			version = info.Main.Version
		}
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				if commit == "" && len(s.Value) >= 7 {
					commit = s.Value[:7]
				}
			case "vcs.time":
				if date == "" {
					date = s.Value
				}
			}
		}
	}

	cmd.SetVersionInfo(version, commit, date)
	if err := cmd.Execute(); err != nil {
		// Exit code 2 for usage/validation errors (POSIX EX_USAGE convention).
		// This lets scripts distinguish "I called the CLI wrong" from
		// "the API returned an error" without parsing stderr.
		exitCode := 1
		errMsg := err.Error()
		if strings.Contains(errMsg, "unknown flag") ||
			strings.Contains(errMsg, "unknown shorthand flag") ||
			strings.Contains(errMsg, "unknown command") ||
			strings.Contains(errMsg, "accepts") || // "accepts N arg(s)"
			strings.Contains(errMsg, "required flag") ||
			strings.Contains(errMsg, "invalid argument") {
			exitCode = 2
		}
		os.Exit(exitCode)
	}
}
