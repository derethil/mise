package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/ai/cleaningredients"
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
			return tandoorUserError(err)
		}

		entry, err := backup.NewStore(cfg.Tandoor.BackupDir).Save(id, recipe.JSON())
		if err != nil {
			return errWithUserMessage(err, "Could not write a backup to %s. Check that the directory is writable.", cfg.Tandoor.BackupDir)
		}

		slog.InfoContext(ctx, entry.Path, slog.Int("recipe_id", id), slog.String("backup_path", entry.Path))

		return nil
	},
}

var recipeRestoreCmd = &cli.Command{
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

		data, err := backup.NewStore(cfg.Tandoor.BackupDir).Load(id, n)
		if err != nil {
			return errWithUserMessage(err, "Could not load backup for recipe %d. Check that the backup directory is correct and contains backups for this recipe.", id)
		}

		client := tandoor.FromConfig(cfg)

		return tandoorUserError(client.Recipes.Update(ctx, id, data))
	},
}

var recipeCleanCmd = &cli.Command{
	Name:  "clean",
	Usage: "Clean up ingredients in a recipe using AI",
	Arguments: []cli.Argument{
		&cli.IntArg{Name: "recipe_id", Required: true},
	},
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "Log the proposed changes without writing them",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) (err error) {
		id := cmd.IntArg("recipe_id")
		defer func() {
			if err != nil {
				err = fmt.Errorf("recipe %d: %w", id, err)
			}
		}()

		cfg := config.FromContext(ctx)

		tclient := tandoor.FromConfig(cfg)
		recipe, err := tclient.Recipes.Get(ctx, id)
		if err != nil {
			return tandoorUserError(err)
		}

		feature, model, err := aiFeature[*cleaningredients.Feature](ctx, cmd, ai.Deps{Tandoor: tclient})
		if err != nil {
			return aiUserError(err, model.String())
		}

		onProgress := printProgress()
		cleaned, err := feature.CleanRecipe(ctx, recipe, func(p cleaningredients.Progress) error {
			return onProgress(progress{
				Label:     recipe.Name,
				Status:    "cleaning ingredients",
				Total:     int64(p.Total),
				Completed: int64(p.Completed),
			})
		})
		if err != nil {
			return tandoorUserError(aiUserError(err, model.String()))
		}

		updated, changes, err := cleaned.Apply(recipe.JSON())
		if errors.Is(err, ai.ErrMalformedResponse) {
			return errWithUserMessage(err, "The model's corrections didn't match the recipe. Try running the command again.")
		}
		if err != nil {
			return err
		}

		fmt.Println("\nChanges:")
		for _, change := range changes {
			fmt.Println(change)
		}

		slog.DebugContext(ctx, "clean recipe finished with changes", slog.Any("changes", changes))

		if cmd.Bool("dry-run") {
			slog.InfoContext(ctx, "Dry run complete. No changes were written.")
			return nil
		}

		if err := backupAndUpdate(ctx, tclient, cfg.Tandoor.BackupDir, recipe.ID, recipe.JSON(), updated); err != nil {
			return err
		}

		return nil
	},
}

func backupAndUpdate(ctx context.Context, client *tandoor.Client, dir string, id int, before, updated []byte) error {
	entry, err := backup.NewStore(dir).Save(id, before)
	if err != nil {
		return errWithUserMessage(err, "Could not write a backup to %s. Check that the directory is writable.", dir)
	}

	slog.InfoContext(ctx,
		fmt.Sprintf("Pre-patched recipe backed up to %s", entry.Path),
		slog.Int("recipe_id", id),
		slog.String("backup_path", entry.Path),
	)

	return tandoorUserError(client.Recipes.Update(ctx, id, updated))
}
