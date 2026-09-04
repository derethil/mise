// Package ai provides an interface for interacting with AI services.
package ai

import (
	"context"
	"log/slog"

	"github.com/derethil/mise/internal/config"
	"github.com/firebase/genkit/go/core/logger"
	"github.com/firebase/genkit/go/genkit"
)

type Client struct {
	g *genkit.Genkit
}

func New(ctx context.Context, providers config.ProvidersConfig, model ModelRef, extra ...ModelRef) (*Client, error) {
	models := append([]ModelRef{model}, extra...)

	plugins, err := getProviderPlugins(providers, models...)
	if err != nil {
		return nil, err
	}

	logger.SetLevel(slog.LevelWarn)

	return &Client{
		g: genkit.Init(ctx,
			genkit.WithPlugins(plugins...),
			genkit.WithDefaultModel(model.String()),
		),
	}, nil
}
