package cliutil

import (
	"context"

	"github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/ai/providers"
	"github.com/derethil/mise/internal/config"

	"github.com/urfave/cli/v3"
)

func LoadFeature[T ai.Feature](ctx context.Context, cmd *cli.Command, size config.ModelSize, deps ai.Deps) (T, providers.ModelRef, func(), error) {
	var zero T

	cfg := config.FromContext(ctx)

	model, err := providers.ParseModel(ResolveFlag(cmd, GlobalFlagModel, cfg.Models.Get(size)))
	if err != nil {
		return zero, model, nil, err
	}

	cleanup := func() {}

	if model.Provider == providers.ProviderOllama {
		providerConfig := cfg.Providers[providers.ProviderOllama]
		ollamaCtx, cancelOllama := context.WithCancel(ctx)

		if err := providers.EnsureOllama(ollamaCtx, providerConfig, cfg.StartOllama || cmd.Bool("start-ollama")); err != nil {
			cancelOllama()
			return zero, model, nil, err
		}

		cleanup = cancelOllama
		ctx = ollamaCtx
	}

	client, err := ai.NewGenkitClient(ctx, cfg.Providers, deps, model)
	if err != nil {
		cleanup()
		return zero, model, nil, err
	}

	feature, err := ai.FeatureOf[T](client)
	if err != nil {
		cleanup()
		return zero, model, nil, err
	}

	return feature, model, cleanup, nil
}
