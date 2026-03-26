package main

import (
	"os"
	"runtime/debug"

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
		os.Exit(1)
	}
}
