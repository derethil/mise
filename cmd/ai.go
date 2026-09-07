package cmd

import (
	"context"

	"github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/config"

	"github.com/urfave/cli/v3"
)

func aiFeature[T ai.Feature](ctx context.Context, cmd *cli.Command, deps ai.Deps) (T, ai.ModelRef, error) {
	var zero T

	cfg := config.FromContext(ctx)

	model, err := ai.ParseModel(resolveFlag(cmd, GlobalFlagModel, cfg.Models.Small))
	if err != nil {
		return zero, model, err
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
