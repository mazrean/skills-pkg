package cli

import (
	"github.com/alecthomas/kong"
)

// InitCmd represents the init command
type InitCmd struct {
	InstallDir []string `help:"Custom install directory (can be specified multiple times)" short:"d"`
	Agent      string   `help:"Agent name (e.g., 'claude') to use default directory" short:"a"`
}

// Run executes the init command
func (c *InitCmd) Run(ctx *kong.Context) error {
	// Implementation will be added in task 7.2
	return nil
}
