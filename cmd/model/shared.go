package model

import (
	"github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/urfave/cli/v3"
)

type labeledModel struct {
	label string
	ref   ai.ModelRef
}

func selectedModels(cmd *cli.Command, cfg config.Config) ([]labeledModel, error) {
	if override := cmd.String(string(cliutil.GlobalFlagModel)); override != "" {
		model, err := ai.ParseModel(override)
		if err != nil {
			return nil, err
		}

		return []labeledModel{{label: "override", ref: model}}, nil
	}

	small, err := ai.ParseModel(cfg.Models.Small)
	if err != nil {
		return nil, err
	}

	large, err := ai.ParseModel(cfg.Models.Large)
	if err != nil {
		return nil, err
	}

	return []labeledModel{
		{label: "small", ref: small},
		{label: "large", ref: large},
	}, nil
}

func modelRefs(models []labeledModel) []ai.ModelRef {
	refs := make([]ai.ModelRef, len(models))
	for i, model := range models {
		refs[i] = model.ref
	}

	return refs
}
