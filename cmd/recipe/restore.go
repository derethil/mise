package recipe

import (
	"context"

	"github.com/derethil/mise/internal/backup"
	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/urfave/cli/v3"
)

var restoreCmd = &cli.Command{
	Name:  "restore",
	Usage: "Restore a recipe from a local backup",
	Arguments: []cli.Argument{
		&cli.IntArg{Name: "recipe_id", Required: true},
	},
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:    "index",
			Aliases: []string{"i"},
			Usage:   "Which backup to restore (0 = most recent)",
			Value:   0,
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		id := cmd.IntArg("recipe_id")
		n := cmd.Int("index")

		cfg := config.FromContext(ctx)

		data, err := backup.NewStore(cfg.Tandoor.BackupDir, cfg.Backup.Keep).Load(id, n)
		if err != nil {
			return cliutil.ErrWithUserMessage(err, "Could not load backup for recipe %d. Check that the backup directory is correct and contains backups for this recipe.", id)
		}

		client := tandoor.FromConfig(cfg)

		return cliutil.TandoorUserError(client.Recipes.Update(ctx, id, data))
	},
}
