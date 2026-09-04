package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/ollama"
	"github.com/ollama/ollama/format"
	"github.com/urfave/cli/v3"
)

var modelsCmd = &cli.Command{
	Name:  "models",
	Usage: "Manage your configured models and their availability",
	Commands: []*cli.Command{
		{
			Name:  "list",
			Usage: "Print Mise's configured Ollama models and their availability",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				cfg := config.FromContext(ctx)

				models, err := selectedModels(cmd, cfg)
				if err != nil {
					return err
				}

				provisioner, err := ollama.NewProvisioner(cfg.Providers.Ollama.BaseURL)
				if err != nil {
					return err
				}

				statuses, err := provisioner.Statuses(ctx, modelRefs(models))
				if err != nil {
					return err
				}

				for i, model := range models {
					printModelStatus(model.label, statuses[i])
				}

				return nil
			},
		},
		{
			Name:  "pull",
			Usage: "Pull configured models from the Ollama model registry",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				cfg := config.FromContext(ctx)

				models, err := selectedModels(cmd, cfg)
				if err != nil {
					return err
				}

				provisioner, err := ollama.NewProvisioner(cfg.Providers.Ollama.BaseURL)
				if err != nil {
					return err
				}

				for _, model := range models {
					if model.ref.Provider != ai.ProviderOllama {
						continue
					}

					if err := pullModel(ctx, provisioner, model.ref); err != nil {
						return err
					}
				}

				return nil
			},
		},
		{
			Name:  "clear",
			Usage: "Delete Ollama models that are not configured for use by mise",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "yes",
					Aliases: []string{"y"},
					Usage:   "Skip the confirmation prompt",
				},
			},
			Action: func(ctx context.Context, cmd *cli.Command) error {
				cfg := config.FromContext(ctx)

				models, err := selectedModels(cmd, cfg)
				if err != nil {
					return err
				}

				provisioner, err := ollama.NewProvisioner(cfg.Providers.Ollama.BaseURL)
				if err != nil {
					return err
				}

				confirmFunc := confirm
				if cmd.Bool("yes") {
					confirmFunc = autoConfirm
				}

				deleted, err := provisioner.Clear(ctx, modelRefs(models), confirmFunc)
				if errors.Is(err, ollama.ErrClearDeclined) {
					return nil
				}
				if err != nil {
					return err
				}

				if len(deleted) == 0 {
					fmt.Println("No unused models to delete")
					return nil
				}

				for _, model := range deleted {
					fmt.Printf("deleted %s\n", model.Name)
				}

				return nil
			},
		},
	},
}

type labeledModel struct {
	label string
	ref   ai.ModelRef
}

func selectedModels(cmd *cli.Command, cfg config.Config) ([]labeledModel, error) {
	if override := cmd.String(string(GlobalFlagModel)); override != "" {
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

func pullModel(ctx context.Context, provisioner *ollama.Provisioner, model ai.ModelRef) error {
	pulled := false
	onProgress := printProgress()

	err := provisioner.Ensure(ctx, model, autoConfirm, func(p ollama.PullProgress) error {
		pulled = true
		return onProgress(progress{Label: model.String(), Status: p.Status, Total: p.Total, Completed: p.Completed})
	})
	if err != nil {
		return err
	}

	if pulled {
		fmt.Println()
	} else {
		fmt.Printf("%s: already available\n", model)
	}

	return nil
}

func printModelStatus(label string, s ollama.ModelStatus) {
	if s.Model.Provider != ai.ProviderOllama {
		fmt.Printf("[%s] %s: availability not tracked (%s is not an Ollama model)\n", label, s.Model, s.Model.Provider)
		return
	}

	if s.Info == nil {
		fmt.Printf("[%s] %s: missing\n", label, s.Model)
		return
	}

	fmt.Printf("[%s] %s: available (%s, %s, %s, %s, updated %s)\n",
		label,
		s.Model,
		format.HumanBytes(s.Info.Size),
		s.Info.Family,
		s.Info.ParameterSize,
		s.Info.QuantizationLevel,
		format.HumanTimeLower(s.Info.ModifiedAt, "unknown"),
	)

	if len(s.Info.Capabilities) > 0 {
		fmt.Printf("  capabilities: %s\n", strings.Join(s.Info.Capabilities, ", "))
	}
}
