package recipe

import (
	"context"
	"log/slog"

	"github.com/derethil/mise/internal/backup"
	"github.com/derethil/mise/internal/cliutil"
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

		client := tandoor.FromConfig(cfg)
		recipe, err := client.Recipes.Get(ctx, id)
		if err != nil {
			return cliutil.TandoorUserError(err)
		}

		entry, err := backup.NewStore(cfg.Tandoor.BackupDir, cfg.Backup.Keep).Save(id, recipe.JSON())
		if err != nil {
			return cliutil.ErrWithUserMessage(err, "Could not write a backup to %s. Check that the directory is writable.", cfg.Tandoor.BackupDir)
		}

		slog.InfoContext(ctx, entry.Path, slog.Int("recipe_id", id), slog.String("backup_path", entry.Path))

		return nil
	},
}
