package cmd

import (
	"context"
	"fmt"

	"github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/backup"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/urfave/cli/v3"
)

var recipeCmd = &cli.Command{
	Name:  "recipe",
	Usage: "Manage a Tandoor recipe",
	Commands: []*cli.Command{
		recipeBackupCmd,
		recipeRestoreCmd,
		recipeCleanCmd,
	},
}

var recipeBackupCmd = &cli.Command{
	Name:  "backup",
	Usage: "Backup a recipe to the local backup directory",
	Arguments: []cli.Argument{
		&cli.IntArg{Name: "recipe_id", Required: true},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		id := cmd.IntArg("recipe_id")

		cfg := config.FromContext(ctx)

		client := tandoor.FromConfig(cfg)
		recipe, err := client.Recipes.Get(ctx, id)
		if err != nil {
			return err
		}

		entry, err := backup.NewStore(cfg.Tandoor.BackupDir).Save(id, recipe.JSON())
		if err != nil {
			return fmt.Errorf("recipe %d: %w", id, err)
		}

		fmt.Println(entry.Path)
		return nil
	},
}

var recipeRestoreCmd = &cli.Command{
	Name:  "restore",
	Usage: "Restore a recipe from a local backup",
	Arguments: []cli.Argument{
		&cli.IntArg{Name: "recipe_id", Required: true},
		&cli.IntArg{Name: "n", Required: false},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		id := cmd.IntArg("recipe_id")
		n := cmd.IntArg("n")

		cfg := config.FromContext(ctx)

		data, err := backup.NewStore(cfg.Tandoor.BackupDir).Load(id, n)
		if err != nil {
			return fmt.Errorf("recipe %d: %w", id, err)
		}

		client := tandoor.FromConfig(cfg)
		return client.Recipes.Update(ctx, id, data)
	},
}

var recipeCleanCmd = &cli.Command{
	Name:  "clean",
	Usage: "Clean up ingredients in a recipe using AI",
	Arguments: []cli.Argument{
		&cli.IntArg{Name: "recipe_id", Required: true},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		cfg := config.FromContext(ctx)
		model := resolveFlag(cmd, GlobalFlagModel, cfg.Models.Small)

		_, err := ai.NewAIClient(ctx, cfg, model)
		if err != nil {
			return err
		}

		fmt.Println("Cleaning ingredients for recipe", cmd.IntArg("recipe_id"), "using model", model)

		return nil
	},
}
