// Package ai provides an interface for interacting with AI services.
package ai

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/derethil/mise/internal/ai/providers"
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

func NewGenkitClient(ctx context.Context, cfg config.ProvidersConfig, deps Deps, models ...providers.ModelRef) (*Client, error) {
	if len(models) == 0 {
		return nil, ErrNoModels
	}

	slog.DebugContext(ctx, "initializing ai client", slog.String("model", models[0].String()))

	configured, err := providers.ForModels(ctx, cfg, models...)
	if err != nil {
		return nil, err
	}

	return newGenkitClient(ctx, configured, deps, models...)
}

func newGenkitClient(ctx context.Context, configured []*providers.Provider, deps Deps, models ...providers.ModelRef) (*Client, error) {
	plugins := make([]api.Plugin, len(configured))
	for i, provider := range configured {
		plugins[i] = provider.Plugin
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
	err := client.RegisterFeatures(ctx, g, configured, models[0], deps)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) RegisterFeatures(ctx context.Context, g *genkit.Genkit, configured []*providers.Provider, model providers.ModelRef, deps Deps) error {
	providerByName := make(map[string]*providers.Provider, len(configured))
	for _, provider := range configured {
		providerByName[provider.Name] = provider
	}

	registry := Registry{Genkit: g, Providers: providerByName, Model: model, Deps: deps}

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
