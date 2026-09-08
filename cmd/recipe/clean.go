package recipe

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/ai/cleaningredients"
	"github.com/derethil/mise/internal/backup"
	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/urfave/cli/v3"
)

var cleanCmd = &cli.Command{
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
			return cliutil.TandoorUserError(err)
		}

		feature, model, err := cliutil.LoadFeature[*cleaningredients.Feature](ctx, cmd, ai.Deps{Tandoor: tclient})
		if err != nil {
			return cliutil.AIUserError(err, model.String())
		}

		onProgress := cliutil.PrintProgress()
		cleaned, err := feature.CleanRecipe(ctx, recipe, func(p cleaningredients.Progress) error {
			return onProgress(cliutil.Progress{
				Label:     recipe.Name,
				Status:    "cleaning ingredients",
				Total:     int64(p.Total),
				Completed: int64(p.Completed),
			})
		})
		if err != nil {
			return cliutil.TandoorUserError(cliutil.AIUserError(err, model.String()))
		}

		updated, changes, err := cleaned.Apply(recipe.JSON())
		if errors.Is(err, ai.ErrMalformedResponse) {
			return cliutil.ErrWithUserMessage(err, "The model's corrections didn't match the recipe. Try running the command again.")
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

		if err := backupAndUpdate(ctx, tclient, cfg.Tandoor.BackupDir, cfg.Backup.Keep, recipe.ID, recipe.JSON(), updated); err != nil {
			return err
		}

		return nil
	},
}

func backupAndUpdate(ctx context.Context, client *tandoor.Client, dir string, keep int, id int, before, updated []byte) error {
	entry, err := backup.NewStore(dir, keep).Save(id, before)
	if err != nil {
		return cliutil.ErrWithUserMessage(err, "Could not write a backup to %s. Check that the directory is writable.", dir)
	}

	slog.InfoContext(ctx,
		fmt.Sprintf("Pre-patched recipe backed up to %s", entry.Path),
		slog.Int("recipe_id", id),
		slog.String("backup_path", entry.Path),
	)

	return cliutil.TandoorUserError(client.Recipes.Update(ctx, id, updated))
}
