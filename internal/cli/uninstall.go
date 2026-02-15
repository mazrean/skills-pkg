package cli

import (
	"github.com/alecthomas/kong"
)

// UninstallCmd represents the uninstall command
type UninstallCmd struct {
	SkillName string `arg:"" help:"Skill name to uninstall"`
}

// Run executes the uninstall command
func (c *UninstallCmd) Run(ctx *kong.Context) error {
	// Implementation will be added in task 7.7
	return nil
}
