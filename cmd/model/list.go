package model

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/ollama"
	"github.com/ollama/ollama/format"
	"github.com/urfave/cli/v3"
)

var listCmd = &cli.Command{
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
			printModelStatus(ctx, model.label, statuses[i])
		}

		return nil
	},
}

func printModelStatus(ctx context.Context, label string, s ollama.ModelStatus) {
	if s.Model.Provider != ai.ProviderOllama {
		slog.InfoContext(ctx, fmt.Sprintf("[%s] %s: availability not tracked (%s is not an Ollama model)", label, s.Model, s.Model.Provider),
			slog.String("label", label), slog.String("model", s.Model.String()), slog.String("provider", string(s.Model.Provider)))
		return
	}

	if s.Info == nil {
		slog.WarnContext(ctx, fmt.Sprintf("[%s] %s: missing", label, s.Model),
			slog.String("label", label), slog.String("model", s.Model.String()))
		return
	}

	message := fmt.Sprintf("[%s] %s: available (%s, %s, %s, %s, updated %s)",
		label,
		s.Model,
		format.HumanBytes(s.Info.Size),
		s.Info.Family,
		s.Info.ParameterSize,
		s.Info.QuantizationLevel,
		format.HumanTimeLower(s.Info.ModifiedAt, "unknown"),
	)

	if len(s.Info.Capabilities) > 0 {
		message += fmt.Sprintf("\n  capabilities: %s", strings.Join(s.Info.Capabilities, ", "))
	}

	slog.InfoContext(ctx, message, slog.String("label", label), slog.String("model", s.Model.String()), slog.Int64("size_bytes", s.Info.Size))
}
