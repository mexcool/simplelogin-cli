// gen-man generates man pages for the sl CLI using cobra/doc.
// Usage: go run ./cmd/gen-man [output-dir]
// Default output directory is "man/".
package main

import (
	"log"
	"os"
	"time"

	"github.com/mexcool/simplelogin-cli/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func main() {
	outDir := "man"
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatalf("failed to create output directory %q: %v", outDir, err)
	}

	// Build a minimal root command that mirrors the real one, with a fixed
	// date so the generated pages are deterministic.
	header := &doc.GenManHeader{
		Title:   "SL",
		Section: "1",
		Date:    func() *time.Time { t := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC); return &t }(),
		Source:  "SimpleLogin CLI",
		Manual:  "SimpleLogin CLI Manual",
	}

	// Execute is exported; use the exported root via a tiny shim so we can
	// pass a *cobra.Command to GenManTree.
	root := buildRoot()

	if err := doc.GenManTree(root, header, outDir); err != nil {
		log.Fatalf("failed to generate man pages: %v", err)
	}

	log.Printf("man pages written to %s/", outDir)
}

// buildRoot returns a *cobra.Command that has the same sub-command tree as the
// real sl binary.  We call cmd.RootCmd() which we expose below.
func buildRoot() *cobra.Command {
	return cmd.RootCmd()
}
