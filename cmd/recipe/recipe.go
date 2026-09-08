// Package recipe implements mise's "recipe" command family: backup,
// restore, and AI-assisted ingredient cleanup for Tandoor recipes.
package recipe

import "github.com/urfave/cli/v3"

var Command = &cli.Command{
	Name:  "recipe",
	Usage: "Manage a Tandoor recipe",
	Commands: []*cli.Command{
		backupCmd,
		restoreCmd,
		cleanCmd,
	},
}
