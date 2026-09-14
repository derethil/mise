package recipe

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/urfave/cli/v3"
)

func failedIDsPath(command string) string {
	return filepath.Join(config.DataDir, fmt.Sprintf("recipe-%s-failed.json", command))
}

func runBulk(ctx context.Context, cmd *cli.Command, tclient *tandoor.Client, command string, perRecipe func(context.Context, int) error) error {
	path := failedIDsPath(command)

	ids, err := recipeIDsToRun(ctx, cmd, tclient, path)
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		slog.InfoContext(ctx, "No recipes to process")
		return nil
	}

	var failed []int
	var errs []error

	for _, id := range ids {
		if err := perRecipe(ctx, id); err != nil {
			slog.ErrorContext(ctx, err.Error(), slog.Int("recipe_id", id))
			errs = append(errs, err)
			failed = append(failed, id)

			if saveErr := saveFailedIDs(path, failed); saveErr != nil {
				slog.WarnContext(ctx, "Could not write failed-recipe-ids file", slog.String("path", path), slog.Any("error", saveErr))
			}
		}
	}

	return errors.Join(errs...)
}

func recipeIDsToRun(ctx context.Context, cmd *cli.Command, tclient *tandoor.Client, path string) ([]int, error) {
	if cmd.Bool("failed") {
		ids, err := loadFailedIDs(path)
		if err != nil {
			return nil, cliutil.ErrWithUserMessage(err, "Could not read the failed-recipe-ids file at %s.", path)
		}

		return ids, nil
	}

	ids, err := tclient.Recipes.GetAllRecipeIDs(ctx)
	if err != nil {
		return nil, cliutil.TandoorUserError(err)
	}

	return ids, nil
}
