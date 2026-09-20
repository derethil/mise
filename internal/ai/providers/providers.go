// Package providers provides a unified interface for interacting with different AI model providers.
package providers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/derethil/mise/internal/config"
	genai "github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/api"
)

const ProviderOllama = "ollama"

type Provider struct {
	Name           string
	Config         config.ProviderConfig
	Plugin         api.Plugin
	GenerateConfig func(GenerateConfig) any
	Middleware     func(GenerateConfig) []genai.Middleware
}

// NOTE: Provider environment variables and CLI flags are generated from
// config.defaultConfig.Providers. Add a provider key there when registering a
// factory here to enable those config options for a new provider.
var providerFactories = map[string]func(config.ProviderConfig) (*Provider, error){
	ProviderOllama: func(cfg config.ProviderConfig) (*Provider, error) {
		p, err := NewOllamaProvider(cfg)
		if err != nil {
			return nil, err
		}
		return &p.Provider, nil
	},
}

func NewProvider(name string, cfg config.ProviderConfig) (*Provider, error) {
	factory, ok := providerFactories[name]
	if !ok {
		return nil, fmt.Errorf("%w: unsupported provider: %s", config.ErrInvalidConfig, name)
	}
	p, err := factory(cfg)
	if err != nil {
		return nil, fmt.Errorf("provider %s: %w", name, err)
	}
	return p, nil
}

func SupportedProviders() []string {
	providers := make([]string, 0, len(providerFactories))
	for provider := range providerFactories {
		providers = append(providers, provider)
	}
	return providers
}

func ForModels(ctx context.Context, providers config.ProvidersConfig, models ...ModelRef) ([]*Provider, error) {
	seen := make(map[string]bool, len(models))
	configured := make([]*Provider, 0, len(models))

	for _, model := range models {
		if seen[model.Provider] {
			continue
		}
		seen[model.Provider] = true

		_, ok := providerFactories[model.Provider]
		if !ok {
			return nil, fmt.Errorf("%w: unsupported provider: %s", config.ErrInvalidConfig, model.Provider)
		}

		cfg, ok := providers.Get(model.Provider)
		if !ok {
			return nil, fmt.Errorf("%w: provider %s is not configured", config.ErrInvalidConfig, model.Provider)
		}

		provider, err := NewProvider(model.Provider, cfg)
		if err != nil {
			return nil, err
		}

		slog.DebugContext(ctx, "provider plugin ready", slog.String("provider", model.Provider), slog.String("base_url", cfg.BaseURL))

		configured = append(configured, provider)
	}

	return configured, nil
}

func EnsureProvider(ctx context.Context, name string, cfg config.ProvidersConfig, start bool) error {
	switch name {
	case ProviderOllama:
		providerCfg, _ := cfg.Get(name)
		return EnsureOllama(ctx, providerCfg, cfg.Ollama.Autostart || start)
	default:
		return nil
	}
}
