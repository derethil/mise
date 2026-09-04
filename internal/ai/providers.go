package ai

import (
	"fmt"

	"github.com/derethil/mise/internal/config"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/plugins/ollama"
)

const ProviderOllama = "ollama"

var providerFactories = map[string]func(config.ProviderConfig) (api.Plugin, error){
	ProviderOllama: func(cfg config.ProviderConfig) (api.Plugin, error) {
		if cfg.BaseURL == "" {
			return nil, fmt.Errorf("%w: base_url is not set", config.ErrInvalidConfig)
		}

		return &ollama.Ollama{ServerAddress: cfg.BaseURL}, nil
	},
}

func getProviderPlugins(providers config.ProvidersConfig, models ...ModelRef) ([]api.Plugin, error) {
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

		plugins = append(plugins, plugin)
	}

	return plugins, nil
}
