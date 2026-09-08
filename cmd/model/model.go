// Package model implements mise's "models" command family: listing,
// pulling, and clearing Ollama models.
package model

import "github.com/urfave/cli/v3"

var Command = &cli.Command{
	Name:  "models",
	Usage: "Manage your configured models and their availability",
	Commands: []*cli.Command{
		listCmd,
		pullCmd,
		clearCmd,
	},
}
