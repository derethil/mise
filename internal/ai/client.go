// Package ai provides an interface for interacting with AI services.
package ai

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/derethil/mise/internal/config"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/middleware"
)

//go:embed prompts
var promptFS embed.FS

type Client struct {
	features map[reflect.Type]Feature
}

func NewGenkitClient(ctx context.Context, providers config.ProvidersConfig, deps Deps, models ...ModelRef) (*Client, error) {
	if len(models) == 0 {
		return nil, ErrNoModels
	}

	slog.DebugContext(ctx, "initializing ai client", slog.String("model", models[0].String()))

	plugins, err := getProviderPlugins(ctx, providers, models...)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "using provider plugins", slog.String("plugins", pluginNames(plugins)))

	g := genkit.Init(ctx,
		genkit.WithPlugins(append(plugins, &middleware.Middleware{})...),
		genkit.WithDefaultModel(models[0].String()),
		genkit.WithPromptFS(promptFS),
	)

	for _, m := range models {
		if !supportsTools(g, m.String()) {
			return nil, fmt.Errorf("%w: %s", ErrModelMissingTools, m.String())
		}
	}

	client := &Client{features: make(map[reflect.Type]Feature, len(featureFactories))}
	err = client.RegisterFeatures(ctx, g, models, deps)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) RegisterFeatures(ctx context.Context, g *genkit.Genkit, models []ModelRef, deps Deps) error {
	registry := Registry{Genkit: g, Provider: models[0].Provider, Deps: deps}

	for _, newFeature := range featureFactories {
		feature := newFeature()
		if err := feature.Register(registry); err != nil {
			return fmt.Errorf("register %T: %w", feature, err)
		}

		c.features[reflect.TypeOf(feature)] = feature
	}

	slog.DebugContext(ctx, "registered ai features", slog.Int("features", len(c.features)))

	return nil

}

func supportsTools(g *genkit.Genkit, name string) bool {
	m, ok := genkit.LookupModel(g, name).(*ai.ModelAction)
	if !ok || m == nil {
		return false
	}

	model, _ := m.Desc().Metadata["model"].(map[string]any)
	supports, _ := model["supports"].(map[string]any)
	tools, _ := supports["tools"].(bool)
	return tools
}

func pluginNames(plugins []api.Plugin) string {
	var names []string

	for _, p := range plugins {
		names = append(names, p.Name())
	}

	return strings.Join(names, ", ")
}
