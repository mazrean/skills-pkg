package cli

import (
	"github.com/alecthomas/kong"
)

// InstallCmd represents the install command
type InstallCmd struct {
	Skills []string `arg:"" optional:"" help:"Skill names to install (empty for all)"`
}

// Run executes the install command
func (c *InstallCmd) Run(ctx *kong.Context) error {
	// Implementation will be added in task 7.4
	return nil
}
