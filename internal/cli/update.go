package cli

import (
	"github.com/alecthomas/kong"
)

// UpdateCmd represents the update command
type UpdateCmd struct {
	Skills []string `arg:"" optional:"" help:"Skill names to update (empty for all)"`
}

// Run executes the update command
func (c *UpdateCmd) Run(ctx *kong.Context) error {
	// Implementation will be added in task 7.5
	return nil
}
