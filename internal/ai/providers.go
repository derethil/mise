package ai

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/derethil/mise/internal/config"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/plugins/ollama"
)

const ProviderOllama = "ollama"

var providerFactories = map[string]func(config.ProviderConfig) (api.Plugin, error){
	ProviderOllama: getOllamaProviderPlugin,
}

func getProviderPlugins(ctx context.Context, providers config.ProvidersConfig, models ...ModelRef) ([]api.Plugin, error) {
	seen := make(map[string]bool, len(models))
	plugins := make([]api.Plugin, 0, len(models))

	for _, model := range models {
		if seen[model.Provider] {
			continue
		}
		seen[model.Provider] = true

		factory, ok := providerFactories[model.Provider]
		if !ok {
			return nil, fmt.Errorf("%w: unsupported provider: %s", config.ErrInvalidConfig, model.Provider)
		}

		cfg, ok := providers.Get(model.Provider)
		if !ok {
			return nil, fmt.Errorf("%w: provider %s is not configured", config.ErrInvalidConfig, model.Provider)
		}

		plugin, err := factory(cfg)
		if err != nil {
			return nil, fmt.Errorf("provider %s: %w", model.Provider, err)
		}

		slog.DebugContext(ctx, "provider plugin ready", slog.String("provider", model.Provider), slog.String("base_url", cfg.BaseURL))

		plugins = append(plugins, plugin)
	}

	return plugins, nil
}

func getOllamaProviderPlugin(cfg config.ProviderConfig) (api.Plugin, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("%w: base_url is not set", config.ErrInvalidConfig)
	}

	return &ollama.Ollama{
		ServerAddress: cfg.BaseURL,
		Timeout:       cfg.Timeout,
	}, nil
}
