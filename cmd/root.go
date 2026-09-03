// Package cmd defines the CLI commands for the mise application.
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/derethil/mise/internal/config"
	"github.com/urfave/cli/v3"
)

// version is set via -ldflags at release build time (see .goreleaser.yaml).
var version = "dev"

var rootCmd = &cli.Command{
	Name:    "mise",
	Usage:   "mise is a CLI for managing Tandoor recipes",
	Version: version,
	Flags:   config.Flags(),
	Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		cfg, err := config.Load(cmd)
		if err != nil {
			return ctx, err
		}
		return config.NewContext(ctx, cfg), nil
	},
	Commands: []*cli.Command{
		backupCmd,
		restoreCmd,
	},
	EnableShellCompletion: true,
}

func Execute() {
	if err := rootCmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
