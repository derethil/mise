// Package importcmd implements mise's "import" command
package importcmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/video"
	"github.com/derethil/mise/internal/whisper"
	"github.com/urfave/cli/v3"
)

const (
	flagCookiesFile        = "cookies-file"
	flagCookiesFromBrowser = "cookies-from-browser"
	flagDryRun             = "dry-run"
)

var Command = &cli.Command{
	Name:     "import",
	Usage:    "Import a recipe from various social media video sources",
	Metadata: cliutil.GlobalFlagMetadata(cliutil.ProviderOptions, cliutil.TandoorOptions),
	Arguments: []cli.Argument{
		&cli.StringArg{
			Name:     "url",
			Required: true,
		},
	},
	Flags: append([]cli.Flag{
		&cli.BoolFlag{
			Name:     flagDryRun,
			Usage:    "Fetch and print video metadata without downloading",
			Category: "ACTION OPTIONS",
		},
	}, config.FlagsForCommand("import")...),
	ArgValidator: func(ctx context.Context, cmd *cli.Command) error {
		validators := []func(*cli.Command) error{
			validateUrl,
			validateCookiesFromFile,
			validateExclusiveCookies,
		}

		for _, validator := range validators {
			if err := validator(cmd); err != nil {
				return err
			}
		}

		return nil
	},
	Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		configPath := cliutil.ResolveFlag(cmd, cliutil.GlobalFlagConfig, config.DefaultConfigPath())
		cfg, err := config.Load(cmd, configPath)
		if err != nil {
			return ctx, err
		}

		return config.NewContext(ctx, cfg), nil
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		cfg := config.FromContext(ctx)

		extraction, err := video.Probe(ctx, cmd.StringArg("url"), cfg.Extract, progressPrinter())
		if err != nil {
			return cliutil.VideoUserError(err)
		}

		slog.InfoContext(ctx, fmt.Sprintf("Running import for video: %s", extraction.Source.Title))

		if cmd.Bool(flagDryRun) {
			slog.InfoContext(ctx, "Dry run complete, no import performed.")
			return nil
		}

		media, err := extraction.Extract(ctx)
		if err != nil {
			return cliutil.VideoUserError(err)
		}

		transript, err := transcribe(ctx, cfg, *media)
		if errors.Is(err, whisper.ErrPullDeclined) {
			return nil
		}
		if err != nil {
			return err
		}

		fmt.Println(transript)

		return nil
	},
}

func progressPrinter() video.ProgressFunc {
	printProgress := cliutil.PrintProgress()

	return func(p video.Progress) {
		_ = printProgress(cliutil.Progress{
			Label:     "yt-dlp",
			Status:    p.Status,
			Total:     p.Total,
			Completed: p.Completed,
		})
	}
}
