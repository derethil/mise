package recipe

import (
	"context"
	"errors"
	"log/slog"

	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/runs"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/urfave/cli/v3"
)

func validateSelector(cmd *cli.Command) error {
	selected := 0
	for _, name := range []string{"all", "failed", "new"} {
		if cmd.Bool(name) {
			selected++
		}
	}

	if selected > 1 {
		return cliutil.ErrWithUserMessage(cliutil.ErrIncorrectUsage, "Pass only one of --all, --failed, or --new.")
	}
	if cmd.IntArg("id") == 0 && selected == 0 {
		return cliutil.ErrWithUserMessage(cliutil.ErrIncorrectUsage, "Provide a recipe id, or pass --all, --failed, or --new.")
	}

	return nil
}

func isBulkRun(cmd *cli.Command) bool {
	return cmd.Bool("all") || cmd.Bool("failed") || cmd.Bool("new")
}

func runSingle(ctx context.Context, cache *runs.Store, id int, perRecipe func(context.Context, int) error) error {
	if err := perRecipe(ctx, id); err != nil {
		cache.MarkFailure(ctx, id)
		return err
	}

	return nil
}

func runBulk(ctx context.Context, cmd *cli.Command, tclient *tandoor.Client, cache *runs.Store, perRecipe func(context.Context, int) error) error {
	ids, err := recipeIDsToRun(ctx, cmd, tclient, cache)
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		slog.InfoContext(ctx, "No recipes to process")
		return nil
	}

	var errs []error

	for _, id := range ids {
		if err := perRecipe(ctx, id); err != nil {
			slog.ErrorContext(ctx, err.Error(), slog.Int("recipe_id", id))
			errs = append(errs, err)
			cache.MarkFailure(ctx, id)
		}
	}

	return errors.Join(errs...)
}

func recipeIDsToRun(ctx context.Context, cmd *cli.Command, tclient *tandoor.Client, cache *runs.Store) ([]int, error) {
	if cmd.Bool("failed") {
		return cache.Failed(), nil
	}

	ids, err := tclient.Recipes.GetAllRecipeIDs(ctx)
	if err != nil {
		return nil, cliutil.TandoorUserError(err)
	}

	if cmd.Bool("new") {
		ids = withoutCached(ids, cache)
	}

	return ids, nil
}

func withoutCached(ids []int, cache *runs.Store) []int {
	remaining := ids[:0]
	for _, id := range ids {
		if !cache.Has(id) {
			remaining = append(remaining, id)
		}
	}

	return remaining
}
