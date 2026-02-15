package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/mazrean/skills-pkg/internal/cli"
)

// CLI represents the command-line interface structure
var CLI struct {
	Init      cli.InitCmd      `cmd:"" help:"Initialize project with .skillspkg.toml"`
	Add       cli.AddCmd       `cmd:"" help:"Add a skill to configuration"`
	Install   cli.InstallCmd   `cmd:"" help:"Install skills from configuration"`
	Update    cli.UpdateCmd    `cmd:"" help:"Update skills to latest versions"`
	List      cli.ListCmd      `cmd:"" help:"List installed skills"`
	Uninstall cli.UninstallCmd `cmd:"" help:"Uninstall skills"`
	Verify    cli.VerifyCmd    `cmd:"" help:"Verify skill integrity with hash"`

	Verbose bool `help:"Enable verbose logging" short:"v" env:"SKILLSPKG_VERBOSE"`
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
