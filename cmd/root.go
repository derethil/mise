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

	importcmd "github.com/derethil/mise/cmd/import"
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
		Name:     string(cliutil.GlobalFlagModel),
		Usage:    "Override the AI model to use for this command",
		Aliases:  []string{"m"},
		Category: "GENERAL OPTIONS",
	},
	&cli.BoolFlag{
		Name:     string(cliutil.GlobalFlagVerbose),
		Usage:    "Enable verbose (debug) logging; repeat (-vv) for even more verbose output",
		Aliases:  []string{"v"},
		Category: "GENERAL OPTIONS",
	},
	&cli.StringFlag{
		Name:        string(cliutil.GlobalFlagConfig),
		Usage:       "Path to the configuration file",
		Aliases:     []string{"c"},
		DefaultText: config.DefaultConfigPath(),
		Category:    "GENERAL OPTIONS",
	},
}

var rootFlags = append(config.Flags(), globalFlags...)

var rootCmd = &cli.Command{
	Name:                   "mise",
	Usage:                  "mise is a CLI for managing Tandoor recipes",
	Version:                version,
	Flags:                  rootFlags,
	Metadata:               map[string]any{"globalFlagCategories": newGlobalFlagCategories(rootFlags)},
	UseShortOptionHandling: true,
	Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {

		// Re-init after parsing flags to handle -v and -vv
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
		recipe.Command,
		importcmd.Command,
		model.Command,
		configureCmd,
		logsCmd,
		genkitDevCmd,
	},
	EnableShellCompletion: true,
}

func Execute() {
	args := os.Args[1:]

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ctx, err := logging.Init(ctx, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx = logging.NewInvocation(ctx, commandPath(rootCmd, args), args)

	if err := rootCmd.Run(ctx, os.Args); err != nil {
		message := cliutil.UserMessage(err)
		if ctx.Err() != nil {
			message = "\nCancelled."
		}

		slog.ErrorContext(ctx, message)
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
