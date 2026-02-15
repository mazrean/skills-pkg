package main

import (
	"os"

	"github.com/alecthomas/kong"
)

// CLI represents the command-line interface structure
var CLI struct {
	// Subcommands will be added here in future tasks
}

// Version information (will be injected by GoReleaser via ldflags)
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	ctx := kong.Parse(&CLI,
		kong.Name("skills-pkg"),
		kong.Description("Agent Skills package manager for Claude Code and Codex CLI"),
		kong.UsageOnError(),
		kong.Vars{
			"version": version,
		},
	)

	// Execute the selected command
	err := ctx.Run()

	// Handle exit codes according to requirements 12.5 and 12.6
	if err != nil {
		// Non-zero exit code for errors (requirement 12.6)
		os.Exit(1)
	}
	// Zero exit code for success (requirement 12.5)
	os.Exit(0)
}
