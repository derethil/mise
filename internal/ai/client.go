// Package ai provides an interface for interacting with AI services.
package ai

import (
	"context"
	"log/slog"
	"strings"

	"github.com/derethil/mise/internal/config"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
)

type Client struct {
	g *genkit.Genkit
}

func New(ctx context.Context, providers config.ProvidersConfig, model ModelRef, extra ...ModelRef) (*Client, error) {
	models := append([]ModelRef{model}, extra...)

	slog.DebugContext(ctx, "initializing ai client", slog.String("model", model.String()))

	plugins, err := getProviderPlugins(ctx, providers, models...)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "using provider plugins", slog.String("plugins", pluginNames(plugins)))

	return &Client{
		g: genkit.Init(ctx,
			genkit.WithPlugins(plugins...),
			genkit.WithDefaultModel(model.String()),
		),
	}, nil
}

func pluginNames(plugins []api.Plugin) string {
	var names []string

	for _, p := range plugins {
		names = append(names, p.Name())
	}

	return strings.Join(names, ", ")
}
