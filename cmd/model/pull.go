package model

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/ollama"
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

		provisioner, err := ollama.NewProvisioner(cfg.Providers.Ollama.BaseURL)
		if err != nil {
			return err
		}

		for _, model := range models {
			if model.ref.Provider != ai.ProviderOllama {
				continue
			}

			if err := pullModel(ctx, provisioner, model.ref); err != nil {
				return err
			}
		}

		return nil
	},
}

func pullModel(ctx context.Context, provisioner *ollama.Provisioner, model ai.ModelRef) error {
	pulled := false
	onProgress := cliutil.PrintProgress()

	err := provisioner.Ensure(ctx, model, cliutil.AutoConfirm, func(p ollama.PullProgress) error {
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
