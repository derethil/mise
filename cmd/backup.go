package cmd

import (
	"context"

	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/urfave/cli/v3"
)

var backupCmd = &cli.Command{
	Name:  "backup",
	Usage: "Backup a recipe to the local backup directory",
	Arguments: []cli.Argument{
		&cli.IntArg{Name: "recipe_id", Required: true},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		id := cmd.IntArg("recipe_id")

		cfg := config.FromContext(ctx)

		client := tandoor.NewClient(cfg.Tandoor.BaseURL, cfg.Tandoor.Token, cfg.Tandoor.BackupDir)
		recipe := tandoor.NewRecipe(client, id)
		if err := recipe.Load(); err != nil {
			return err
		}

		return recipe.Backup()
	},
}
