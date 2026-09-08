package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	miseai "github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/urfave/cli/v3"
)

// Initializes the AI client and idles, for use with genkit's dev tooling
// Run with `genkit start -- mise genkit`
var genkitDevCmd = &cli.Command{
	Name:   "genkit",
	Usage:  "Initialize the AI client and idle, for use with genkit's dev tooling",
	Hidden: true,
	Action: func(ctx context.Context, cmd *cli.Command) error {
		cfg := config.FromContext(ctx)

		if version != "dev" {
			slog.InfoContext(ctx, "warmup command is only intended for development use.")
			return nil
		}

		models, err := genkitDevModels(cfg, cmd)
		if err != nil {
			return err
		}

		tclient := tandoor.FromConfig(cfg)

		if _, err := miseai.NewGenkitClient(ctx, cfg.Providers, miseai.Deps{Tandoor: tclient}, models...); err != nil {
			return err
		}

		slog.InfoContext(ctx, "AI client ready, flows registered. Press Ctrl+C to exit.")

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop

		return nil
	},
}

func genkitDevModels(cfg config.Config, cmd *cli.Command) ([]miseai.ModelRef, error) {
	names := []string{cfg.Models.Small, cfg.Models.Large}
	if override := cmd.String(string(cliutil.GlobalFlagModel)); override != "" {
		names = append(names, override)
	}

	seen := make(map[string]bool, len(names))
	var models []miseai.ModelRef

	for _, name := range names {
		if seen[name] {
			continue
		}

		seen[name] = true

		model, err := miseai.ParseModel(name)
		if err != nil {
			return nil, err
		}

		models = append(models, model)
	}

	return models, nil
}
