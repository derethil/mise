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
		&cli.IntArg{Name: "id"},
	},
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "dry-run",
			Usage:   "Log the proposed changes without writing them",
			Aliases: []string{"d"},
		},
		&cli.BoolFlag{
			Name:    "all",
			Usage:   "Run on all recipes in the Tandoor instance. Overrides id argument.",
			Aliases: []string{"a"},
		},
		&cli.BoolFlag{
			Name:    "ignore-cleaned",
			Usage:   "Skip recipes that already have a backup, since that indicates they've already been cleaned",
			Aliases: []string{"i"},
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if cmd.NArg() == 0 && !cmd.Bool("all") {
			return cliutil.ErrWithUserMessage(cliutil.ErrIncorrectUsage, "Provide a recipe id or pass --all.")
		}

		cfg := config.FromContext(ctx)
		tclient := tandoor.FromConfig(cfg)

		feature, model, err := cliutil.LoadFeature[*cleaningredients.Feature](ctx, cmd, ai.Deps{Tandoor: tclient})
		if err != nil {
			return cliutil.AIUserError(err, model.String())
		}

		dryRun := cmd.Bool("dry-run")
		ignoreCleaned := cmd.Bool("ignore-cleaned")

		if !cmd.Bool("all") {
			return cleanRecipe(ctx, tclient, feature, model, cfg, cmd.IntArg("id"), dryRun, ignoreCleaned)
		}

		ids, err := tclient.Recipes.GetAllRecipeIDs(ctx)
		if err != nil {
			return cliutil.TandoorUserError(err)
		}

		var errs []error
		for _, id := range ids {
			if err := cleanRecipe(ctx, tclient, feature, model, cfg, id, dryRun, ignoreCleaned); err != nil {
				slog.ErrorContext(ctx, err.Error(), slog.Int("recipe_id", id))
				errs = append(errs, err)
			}
		}

		return errors.Join(errs...)
	},
}

func cleanRecipe(ctx context.Context, tclient *tandoor.Client, feature *cleaningredients.Feature, model ai.ModelRef, cfg config.Config, id int, dryRun, ignoreCleaned bool) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("recipe %d: %w", id, err)
		}
	}()

	store := backup.NewStore(cfg.Tandoor.BackupDir, cfg.Backup.Keep)

	if ignoreCleaned {
		skip, err := alreadyCleaned(store, id)
		if err != nil {
			return err
		}
		if skip {
			slog.InfoContext(ctx, "Skipping already-cleaned recipe", slog.Int("recipe_id", id))
			return nil
		}
	}

	recipe, err := tclient.Recipes.Get(ctx, id)
	if err != nil {
		return cliutil.TandoorUserError(err)
	}

	updated, changes, err := runCleaning(ctx, feature, model, recipe)
	if err != nil {
		return err
	}

	printChanges(ctx, changes)

	if dryRun {
		slog.InfoContext(ctx, "Dry run complete. No changes were written.")
		return nil
	}

	return saveRecipe(ctx, tclient, store, cfg.Tandoor.BackupDir, recipe, updated)
}

func alreadyCleaned(store *backup.Store, id int) (bool, error) {
	entries, err := store.List(id)
	if err != nil {
		return false, err
	}

	return len(entries) > 0, nil
}

func runCleaning(ctx context.Context, feature *cleaningredients.Feature, model ai.ModelRef, recipe *tandoor.Recipe) ([]byte, []cleaningredients.Change, error) {
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
		return nil, nil, cliutil.TandoorUserError(cliutil.AIUserError(err, model.String()))
	}

	updated, changes, err := cleaned.Apply(recipe.JSON())
	if errors.Is(err, ai.ErrMalformedResponse) {
		return nil, nil, cliutil.ErrWithUserMessage(err, "The model's corrections didn't match the recipe. Try running the command again.")
	}
	if err != nil {
		return nil, nil, err
	}

	return updated, changes, nil
}

func printChanges(ctx context.Context, changes []cleaningredients.Change) {
	fmt.Println("\nChanges:")
	for _, change := range changes {
		fmt.Println(change)
	}

	slog.DebugContext(ctx, "clean recipe finished with changes", slog.Any("changes", changes))
}

func saveRecipe(ctx context.Context, tclient *tandoor.Client, store *backup.Store, backupDir string, recipe *tandoor.Recipe, updated []byte) error {
	entry, err := store.Save(recipe.ID, recipe.JSON())
	if err != nil {
		return cliutil.ErrWithUserMessage(err, "Could not write a backup to %s. Check that the directory is writable.", backupDir)
	}

	slog.InfoContext(ctx,
		fmt.Sprintf("Pre-patched recipe backed up to %s", entry.Path),
		slog.Int("recipe_id", recipe.ID),
		slog.String("backup_path", entry.Path),
	)

	return cliutil.TandoorUserError(tclient.Recipes.Update(ctx, recipe.ID, updated))
}
