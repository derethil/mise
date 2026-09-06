// Package cmd defines the CLI commands for the mise application.
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/logging"
	"github.com/urfave/cli/v3"
)

// version is set via -ldflags at release build time (see .goreleaser.yaml).
var version = "dev"

type GlobalFlag string

const (
	GlobalFlagModel   GlobalFlag = "model"
	GlobalFlagVerbose GlobalFlag = "verbose"
)

var globalFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    string(GlobalFlagModel),
		Usage:   "Override the AI model to use for this command",
		Aliases: []string{"m"},
	},
	&cli.BoolFlag{
		Name:    string(GlobalFlagVerbose),
		Usage:   "Enable verbose (debug) logging",
		Aliases: []string{"v"},
	},
}

var rootCmd = &cli.Command{
	Name:    "mise",
	Usage:   "mise is a CLI for managing Tandoor recipes",
	Version: version,
	Flags:   append(config.Flags(), globalFlags...),
	Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		if err := logging.Init(cmd.Bool(string(GlobalFlagVerbose))); err != nil {
			return ctx, err
		}

		slog.DebugContext(ctx, "command started")

		cfg, err := config.Load(cmd)
		if err != nil {
			return ctx, err
		}

		slog.DebugContext(ctx, "config loaded", slog.Any("config", cfg))

		return config.NewContext(ctx, cfg), nil
	},
	Commands: []*cli.Command{
		recipeCmd,
		modelsCmd,
	},
	EnableShellCompletion: true,
}

func Execute() {
	args := os.Args[1:]
	ctx := logging.NewInvocation(context.Background(), commandPath(rootCmd, args), args)

	if err := rootCmd.Run(ctx, os.Args); err != nil {
		slog.ErrorContext(ctx, err.Error())
		fmt.Fprintln(os.Stderr, "Error:", userMessage(err))
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

func resolveFlag(cmd *cli.Command, flag GlobalFlag, fallback string) string {
	if v := cmd.String(string(flag)); v != "" {
		return v
	}

	return fallback
}
