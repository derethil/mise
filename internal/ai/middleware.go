package ai

import (
	"context"
	"regexp"
	"strings"

	"github.com/firebase/genkit/go/ai"
)

var providerMiddlewareFactories = map[string]func(GenerateConfig) []ai.Middleware{
	ProviderOllama: ollamaMiddleware,
}

func ollamaMiddleware(cfg GenerateConfig) []ai.Middleware {
	return []ai.Middleware{stripThinkArtifacts}
}

// some Ollama-served Qwen3 builds leak /think and /no_think into generated text instead of consuming as config
var thinkArtifactRegex = regexp.MustCompile(`(?i)\(?\s*/(?:no_)?think\s*\)?`)

var stripThinkArtifacts = ai.MiddlewareFunc(func(context.Context) (*ai.Hooks, error) {
	return &ai.Hooks{
		WrapModel: func(ctx context.Context, params *ai.ModelParams, next ai.ModelNext) (*ai.ModelResponse, error) {
			resp, err := next(ctx, params)
			if err != nil || resp == nil || resp.Message == nil {
				return resp, err
			}

			for _, part := range resp.Message.Content {
				if part.IsText() {
					part.Text = strings.TrimSpace(thinkArtifactRegex.ReplaceAllString(part.Text, ""))
				}
			}

			return resp, nil
		},
	}, nil
})
