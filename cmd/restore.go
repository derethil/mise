package cmd

import (
	"context"
	"fmt"

	"github.com/derethil/mise/internal/backup"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/urfave/cli/v3"
)

var restoreCmd = &cli.Command{
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

		client := tandoor.NewClient(cfg.Tandoor.BaseURL, cfg.Tandoor.Token)
		return client.Recipes.Update(ctx, id, data)
	},
}
