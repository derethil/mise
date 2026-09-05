// Package ai provides an interface for interacting with AI services.
package ai

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/derethil/mise/internal/config"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
)

var ErrModelMissingTools = errors.New("model does not support tool calling")

type Client struct {
	g *genkit.Genkit
}

func NewGenkitClient(ctx context.Context, providers config.ProvidersConfig, model ModelRef, extra ...ModelRef) (*Client, error) {
	models := append([]ModelRef{model}, extra...)

	slog.DebugContext(ctx, "initializing ai client", slog.String("model", model.String()))

	plugins, err := getProviderPlugins(ctx, providers, models...)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "using provider plugins", slog.String("plugins", pluginNames(plugins)))

	g := genkit.Init(ctx,
		genkit.WithPlugins(plugins...),
		genkit.WithDefaultModel(model.String()),
	)

	for _, m := range models {
		if !supportsTools(g, m.String()) {
			return nil, fmt.Errorf("%w: %s", ErrModelMissingTools, m.String())
		}
	}

	return &Client{g: g}, nil
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
