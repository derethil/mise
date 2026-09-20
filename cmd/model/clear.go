package model

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/derethil/mise/internal/ai/providers"
	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/urfave/cli/v3"
)

var clearCmd = &cli.Command{
	Name:  "clear",
	Usage: "Delete Ollama models that are not configured for use by mise",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "yes",
			Aliases: []string{"y"},
			Usage:   "Skip the confirmation prompt",
		},
		&cli.BoolFlag{
			Name:    "all",
			Aliases: []string{"a"},
			Usage:   "Delete all Ollama models, even those configured for use by mise",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		cfg := config.FromContext(ctx)

		models, err := selectedModels(cmd, cfg)
		if err != nil {
			return err
		}

		providerCfg, _ := cfg.Providers.Get(providers.ProviderOllama)
		provider, err := ollamaProvider(providerCfg, models)
		if err != nil {
			return err
		}

		confirmFunc := cliutil.Confirm
		if cmd.Bool("yes") {
			confirmFunc = cliutil.AutoConfirm
		}

		var keep []providers.ModelRef
		if !cmd.Bool("all") {
			keep = modelRefs(models)
		}

		deleted, err := provider.Clear(ctx, keep, confirmFunc)
		if errors.Is(err, providers.ErrClearDeclined) {
			return nil
		}
		if err != nil {
			return err
		}

		if len(deleted) == 0 {
			slog.InfoContext(ctx, "No unused models to delete")
			return nil
		}

		for _, model := range deleted {
			slog.InfoContext(ctx, fmt.Sprintf("deleted %s", model.Name), slog.String("model", model.Name))
		}

		return nil
	},
}
