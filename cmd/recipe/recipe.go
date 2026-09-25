// Package recipe implements mise's "recipe" command family.
package recipe

import (
	"context"

	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/urfave/cli/v3"
)

var Command = &cli.Command{
	Name:  "recipe",
	Usage: "Manage an existing Tandoor recipe",
	Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		cfg := config.FromContext(ctx)
		cliutil.WarnIfTandoorVersionUnsupported(ctx, tandoor.FromConfig(cfg))

		return ctx, nil
	},
	Commands: []*cli.Command{
		backupCmd,
		restoreCmd,
		normalizeCmd,
		keywordCmd,
	},
}
