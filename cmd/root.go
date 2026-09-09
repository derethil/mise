// Package cmd defines the CLI commands for the mise application.
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/derethil/mise/cmd/model"
	"github.com/derethil/mise/cmd/recipe"
	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/logging"
	"github.com/urfave/cli/v3"
)

// version is set via -ldflags at release build time (see .goreleaser.yaml).
var version = "dev"

var globalFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    string(cliutil.GlobalFlagModel),
		Usage:   "Override the AI model to use for this command",
		Aliases: []string{"m"},
	},
	&cli.BoolFlag{
		Name:    string(cliutil.GlobalFlagVerbose),
		Usage:   "Enable verbose (debug) logging; repeat (-vv) for even more verbose output",
		Aliases: []string{"v"},
	},
	&cli.StringFlag{
		Name:    string(cliutil.GlobalFlagConfig),
		Usage:   "Path to the configuration file",
		Aliases: []string{"c"},
	},
}

var rootCmd = &cli.Command{
	Name:                   "mise",
	Usage:                  "mise is a CLI for managing Tandoor recipes",
	Version:                version,
	Flags:                  append(config.Flags(), globalFlags...),
	UseShortOptionHandling: true,
	Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		ctx, err := logging.Init(ctx, cmd.Count(string(cliutil.GlobalFlagVerbose)))
		if err != nil {
			return ctx, err
		}

		slog.DebugContext(ctx, "command started")

		configPath := cliutil.ResolveFlag(cmd, cliutil.GlobalFlagConfig, config.DefaultConfigPath())
		cfg, err := config.Load(cmd, configPath)
		if err != nil {
			return ctx, err
		}

		slog.DebugContext(ctx, "config loaded", slog.Any("config", cfg))

		return config.NewContext(ctx, cfg), nil
	},
	Commands: []*cli.Command{
		genkitDevCmd,
		configureCmd,
		logsCmd,
		recipe.Command,
		model.Command,
	},
	EnableShellCompletion: true,
}

func Execute() {
	args := os.Args[1:]

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ctx = logging.NewInvocation(ctx, commandPath(rootCmd, args), args)

	if err := rootCmd.Run(ctx, os.Args); err != nil {
		slog.ErrorContext(ctx, err.Error())

		message := cliutil.UserMessage(err)
		if ctx.Err() != nil {
			message = "Cancelled."
		}

		fmt.Fprintln(os.Stderr, "Error:", message)
		os.Exit(1)
	}

	slog.DebugContext(ctx, "command finished")
}

func commandPath(cmd *cli.Command, args []string) string {
	var parts []string

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}

		next := cmd.Command(arg)
		if next == nil {
			break
		}

		parts = append(parts, arg)
		cmd = next
	}

	return strings.Join(parts, " ")
}
