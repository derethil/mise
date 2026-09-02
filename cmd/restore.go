package cmd

import (
	"context"

	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/urfave/cli/v3"
)

var restoreCmd = &cli.Command{
	Name:  "restore",
	Usage: "Restore a recipe from the local backup directory",
	Arguments: []cli.Argument{
		&cli.IntArg{Name: "recipe_id", Required: true},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		id := cmd.IntArg("recipe_id")
		cfg := config.FromContext(ctx)
		client := tandoor.NewClient(cfg.Tandoor.BaseURL, cfg.Tandoor.Token, config.BackupDir)
		recipe := tandoor.NewRecipe(client, id)
		return recipe.Restore()
	},
}
