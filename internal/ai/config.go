package ai

import "github.com/firebase/genkit/go/plugins/ollama"

type GenerateConfig struct {
	Temperature *float64
	TopP        *float64
	Reasoning   *bool
}

var providerConfigFactories = map[string]func(GenerateConfig) any{
	ProviderOllama: ollamaConfig,
}

func ollamaConfig(cfg GenerateConfig) any {
	out := &ollama.GenerateContentConfig{
		Temperature: cfg.Temperature,
		TopP:        cfg.TopP,
	}

	if cfg.Reasoning != nil {
		out.Think = ollama.ThinkEnabled(*cfg.Reasoning)
	}

	return out
}
