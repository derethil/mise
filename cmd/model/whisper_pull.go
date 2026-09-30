package model

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/config/section"
	"github.com/derethil/mise/internal/whisper"
	"github.com/urfave/cli/v3"
)

var whisperPullCmd = &cli.Command{
	Name:     "pull",
	Usage:    "Download the whisper.cpp model configured by models.whisper",
	Metadata: cliutil.GlobalFlagMetadata(),
	Action: func(ctx context.Context, cmd *cli.Command) error {
		cfg := config.FromContext(ctx)

		model := cfg.Models.Whisper
		if model == "" {
			return cliutil.ErrWithUserMessage(cliutil.ErrIncorrectUsage,
				"models.whisper is not set. Configure a model (%s) and try again.", section.WhisperModelTiers)
		}

		onProgress := cliutil.PrintProgress()
		downloaded, err := whisper.Pull(ctx, model, func(p whisper.PullProgress) error {
			return onProgress(cliutil.Progress{Label: model, Status: "downloading", Total: p.Total, Completed: p.Completed})
		})

		if errors.Is(err, whisper.ErrInvalidModel) {
			return cliutil.ErrWithUserMessage(err,
				"%q is not a valid whisper model. Choose one of %s (append .en for English-only).", model, section.WhisperModelTiers)
		}
		if err != nil {
			return err
		}

		if downloaded {
			fmt.Println()
			slog.InfoContext(ctx, fmt.Sprintf("downloaded whisper model %s", model), slog.String("model", model))
		} else {
			slog.InfoContext(ctx, fmt.Sprintf("%s: already available", model), slog.String("model", model))
		}

		return nil
	},
}
