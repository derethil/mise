package model

import (
	"github.com/derethil/mise/internal/ai/providers"
	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/urfave/cli/v3"
)

type labeledModel struct {
	label string
	ref   providers.ModelRef
}

func selectedModels(cmd *cli.Command, cfg config.Config) ([]labeledModel, error) {
	if override := cmd.String(string(cliutil.GlobalFlagModel)); override != "" {
		model, err := providers.ParseModel(override)
		if err != nil {
			return nil, err
		}

		return []labeledModel{{label: "override", ref: model}}, nil
	}

	small, err := providers.ParseModel(cfg.Models.Small)
	if err != nil {
		return nil, err
	}

	large, err := providers.ParseModel(cfg.Models.Large)
	if err != nil {
		return nil, err
	}

	return []labeledModel{
		{label: "small", ref: small},
		{label: "large", ref: large},
	}, nil
}

func ollamaProvider(cfg config.ProviderConfig, models []labeledModel) (*providers.OllamaProvider, error) {
	for _, model := range models {
		if model.ref.Provider == providers.ProviderOllama {
			return providers.NewOllamaProvider(cfg)
		}
	}

	return nil, cliutil.ErrWithUserMessage(
		cliutil.ErrIncorrectUsage,
		"No Ollama models are selected. Configure an Ollama model in models.small or models.large, or select one with --model ollama/<model>.",
	)
}

func modelRefs(models []labeledModel) []providers.ModelRef {
	refs := make([]providers.ModelRef, len(models))
	for i, model := range models {
		refs[i] = model.ref
	}

	return refs
}
