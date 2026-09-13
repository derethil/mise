// Package recipe implements mise's "recipe" command family.
package recipe

import "github.com/urfave/cli/v3"

var Command = &cli.Command{
	Name:  "recipe",
	Usage: "Manage a Tandoor recipe",
	Commands: []*cli.Command{
		backupCmd,
		restoreCmd,
		normalizeCmd,
		keywordCmd,
	},
}
