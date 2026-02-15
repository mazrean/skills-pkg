package cli

import (
	"github.com/alecthomas/kong"
)

// AddCmd represents the add command
type AddCmd struct {
	Name           string `arg:"" help:"Skill name"`
	Source         string `required:"" help:"Source type (git, npm, go-module)"`
	URL            string `required:"" help:"Source URL (Git URL, npm package name, or Go module path)"`
	Version        string `default:"latest" help:"Version (tag, commit hash, or semantic version)"`
	PackageManager string `help:"Package manager (npm, go-module) if applicable"`
}

// Run executes the add command
func (c *AddCmd) Run(ctx *kong.Context) error {
	// Implementation will be added in task 7.3
	return nil
}
