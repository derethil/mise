package cliutil

import (
	"context"

	"github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/ai/providers"
	"github.com/derethil/mise/internal/config"

	"github.com/urfave/cli/v3"
)

func LoadFeature[T ai.Feature](ctx context.Context, cmd *cli.Command, size config.ModelSize, deps ai.Deps) (T, providers.ModelRef, error) {
	var zero T

	cfg := config.FromContext(ctx)

	model, err := providers.ParseModel(ResolveFlag(cmd, GlobalFlagModel, cfg.Models.Get(size)))
	if err != nil {
		return zero, model, err
	}

	if model.Provider == providers.ProviderOllama {
		providerConfig := cfg.Providers[providers.ProviderOllama]
		if err := providers.EnsureOllama(ctx, providerConfig, cfg.StartOllama || cmd.Bool("start-ollama")); err != nil {
			return zero, model, err
		}
	}

	client, err := ai.NewGenkitClient(ctx, cfg.Providers, deps, model)
	if err != nil {
		return zero, model, err
	}

	feature, err := ai.FeatureOf[T](client)
	if err != nil {
		return zero, model, err
	}

	return feature, model, nil
}
