package model

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/derethil/mise/internal/ai/providers"
	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/urfave/cli/v3"
)

var pullCmd = &cli.Command{
	Name:  "pull",
	Usage: "Pull configured models from the Ollama model registry",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		cfg := config.FromContext(ctx)

		models, err := selectedModels(cmd, cfg)
		if err != nil {
			return err
		}

		provider, err := ollamaProvider(cfg.Providers[providers.ProviderOllama], models)
		if err != nil {
			return err
		}

		for _, model := range models {
			if model.ref.Provider != providers.ProviderOllama {
				continue
			}

			if err := pullModel(ctx, provider, model.ref); err != nil {
				return err
			}
		}

		return nil
	},
}

func pullModel(ctx context.Context, provider *providers.OllamaProvider, model providers.ModelRef) error {
	pulled := false
	onProgress := cliutil.PrintProgress()

	err := provider.Ensure(ctx, model, cliutil.AutoConfirm, func(p providers.PullProgress) error {
		pulled = true
		return onProgress(cliutil.Progress{Label: model.String(), Status: p.Status, Total: p.Total, Completed: p.Completed})
	})
	if err != nil {
		return err
	}

	if pulled {
		fmt.Println()
	} else {
		slog.InfoContext(ctx, fmt.Sprintf("%s: already available", model), slog.String("model", model.String()))
	}

	return nil
}
