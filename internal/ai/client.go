// Package ai provides an interface for interacting with AI services.
package ai

import (
	"context"
	"fmt"

	"github.com/derethil/mise/internal/config"
	"github.com/firebase/genkit/go/genkit"
)

type AIClient struct {
	g     *genkit.Genkit
	model string
}

type AIClientOption func(*AIClient)

func NewAIClient(ctx context.Context, cfg config.Config, model string, opts ...AIClientOption) (*AIClient, error) {
	c := &AIClient{model: model}
	for _, opt := range opts {
		opt(c)
	}

	plugins, err := getProviderPlugins(cfg.Providers, c.model)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}

	c.g = genkit.Init(ctx,
		genkit.WithPlugins(plugins...),
		genkit.WithDefaultModel(c.model),
	)

	return c, nil
}
