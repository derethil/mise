package ai

import (
	"fmt"
	"strings"

	"github.com/derethil/mise/internal/config"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/plugins/ollama"
)

var providerFactories = map[string]func(config.ProvidersConfig) api.Plugin{
	"ollama": func(cfg config.ProvidersConfig) api.Plugin {
		return &ollama.Ollama{
			ServerAddress: cfg.Ollama.BaseURL,
		}
	},
}

func getProviderPlugins(cfg config.ProvidersConfig, models ...string) ([]api.Plugin, error) {
	providers, err := uniqueProviders(models)
	if err != nil {
		return nil, err
	}

	plugins := make([]api.Plugin, 0, len(providers))
	for _, provider := range providers {
		plugins = append(plugins, providerFactories[provider](cfg))
	}

	return plugins, nil
}

func uniqueProviders(models []string) ([]string, error) {
	seen := make(map[string]bool, len(models))
	providers := make([]string, 0, len(models))

	for _, model := range models {
		provider, _, err := parseModel(model)
		if err != nil {
			return nil, fmt.Errorf("failed to parse model %s: %w", model, err)
		}

		if seen[provider] {
			continue
		}

		seen[provider] = true
		providers = append(providers, provider)
	}

	return providers, nil
}

func parseModel(model string) (string, string, error) {
	provider, modelName, ok := strings.Cut(model, "/")
	if !ok {
		return "", "", fmt.Errorf("invalid model format, expected provider/model: %s", model)
	}

	if _, ok := providerFactories[provider]; !ok {
		return "", "", fmt.Errorf("unsupported provider: %s", provider)
	}

	return provider, modelName, nil
}
