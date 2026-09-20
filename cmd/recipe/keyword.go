package recipe

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/ai/features/assign-keywords"
	"github.com/derethil/mise/internal/ai/providers"
	"github.com/derethil/mise/internal/backup"
	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/runs"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/urfave/cli/v3"
)

var keywordCmd = &cli.Command{
	Name:  "keyword",
	Usage: "Assign keywords to a recipe using AI, following a keyword schema you provide",
	Arguments: []cli.Argument{
		&cli.IntArg{Name: "id"},
	},
	Flags: append(config.FlagsForCommand("recipe keyword"), []cli.Flag{
		&cli.BoolFlag{
			Name:    "dry-run",
			Usage:   "Log the proposed keywords without writing them",
			Aliases: []string{"d"},
		},
		&cli.BoolFlag{
			Name:  "replace",
			Usage: "Replace the recipe's keywords instead of adding to them",
		},
		&cli.BoolFlag{
			Name:  "all",
			Usage: "Run on all recipes in the Tandoor instance. Overrides id argument.",
		},
		&cli.BoolFlag{
			Name:  "failed",
			Usage: "Re-run only the recipes that failed on a previous run",
		},
		&cli.BoolFlag{
			Name:  "new",
			Usage: "Run on all recipes that haven't been keyworded before",
		},
	}...),
	Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		configPath := cliutil.ResolveFlag(cmd, cliutil.GlobalFlagConfig, config.DefaultConfigPath())
		cfg, err := config.Load(cmd, configPath)
		if err != nil {
			return ctx, err
		}

		return config.NewContext(ctx, cfg), nil
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if err := validateSelector(cmd); err != nil {
			return err
		}

		cfg := config.FromContext(ctx)
		tclient := tandoor.FromConfig(cfg)

		schema, err := readSchema(cfg.Keywords.SchemaFile)
		if err != nil {
			return err
		}

		feature, model, err := cliutil.LoadFeature[*assignkeywords.Feature](ctx, cmd, config.ModelLarge, ai.Deps{Tandoor: tclient})
		if err != nil {
			return cliutil.AIUserError(err, model.String())
		}
		feature.Schema = schema

		cache, err := runs.OpenDefault("keyword")
		if err != nil {
			return cliutil.ErrWithUserMessage(err, "Could not open the run state.")
		}
		defer cache.Close()

		opts := keywordOptions{
			dryRun:  cmd.Bool("dry-run"),
			replace: cmd.Bool("replace"),
			ignore:  cfg.Keywords.Ignore,
		}

		run := func(ctx context.Context, id int) error {
			return keywordRecipe(ctx, tclient, feature, model, cfg, id, opts, cache)
		}

		if !isBulkRun(cmd) {
			return runSingle(ctx, cache, cmd.IntArg("id"), run)
		}

		return runBulk(ctx, cmd, tclient, cache, run)
	},
}

type keywordOptions struct {
	dryRun  bool
	replace bool
	ignore  []string
}

func readSchema(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", cliutil.ErrWithUserMessage(err,
			"Could not read the keyword schema at %s. Write the keywords you want assigned and the rules for assigning them to that file.",
			path,
		)
	}

	schema := strings.TrimSpace(string(data))
	if schema == "" {
		return "", cliutil.ErrWithUserMessage(
			fmt.Errorf("keyword schema at %s is empty", path),
			"The keyword schema at %s is empty. Describe the keywords you want assigned and the rules for assigning them.",
			path,
		)
	}

	return schema, nil
}

func keywordRecipe(ctx context.Context, tclient *tandoor.Client, feature *assignkeywords.Feature, model providers.ModelRef, cfg config.Config, id int, opts keywordOptions, cache *runs.Store) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("recipe %d: %w", id, err)
		}
	}()

	recipe, err := tclient.Recipes.Get(ctx, id)
	if err != nil {
		return cliutil.TandoorUserError(err)
	}

	updated, changes, err := runAssignment(ctx, feature, model, recipe, assignkeywords.ApplyOptions{
		Replace: opts.replace,
		Protect: opts.ignore,
	})
	if err != nil {
		return err
	}

	printKeywordChanges(ctx, recipe, changes)

	if opts.dryRun {
		slog.InfoContext(ctx, "Dry run complete. No changes were written.")
		return nil
	}

	if !assignkeywords.HasChanges(changes) {
		cache.MarkOK(ctx, id)
		return nil
	}

	store := backup.NewStore(cfg.Tandoor.BackupDir, cfg.Backup.Keep)
	if err := saveRecipe(ctx, tclient, store, cfg.Tandoor.BackupDir, recipe, updated); err != nil {
		return err
	}

	cache.MarkOK(ctx, id)
	return nil
}

func runAssignment(ctx context.Context, feature *assignkeywords.Feature, model providers.ModelRef, recipe *tandoor.Recipe, applyOpts assignkeywords.ApplyOptions) ([]byte, []assignkeywords.Change, error) {
	onProgress := cliutil.PrintProgress()
	assigned, vocabulary, err := feature.AssignKeywords(ctx, recipe, func(p assignkeywords.Progress) error {
		return onProgress(cliutil.Progress{
			Label:     recipe.Name,
			Status:    p.Status,
			Total:     int64(p.Total),
			Completed: int64(p.Completed),
		})
	})
	if err != nil {
		return nil, nil, cliutil.TandoorUserError(cliutil.AIUserError(err, model.String()))
	}

	updated, changes, err := assigned.Apply(recipe.JSON(), vocabulary, applyOpts)
	if err != nil {
		return nil, nil, err
	}

	return updated, changes, nil
}

func printKeywordChanges(ctx context.Context, recipe *tandoor.Recipe, changes []assignkeywords.Change) {
	slog.DebugContext(ctx, "keyword assignment finished with changes", slog.Any("changes", changes))

	fmt.Printf("\n%s:\n", recipe.Name)
	if len(changes) == 0 {
		fmt.Println("  (no keywords)")
		return
	}

	for _, change := range changes {
		fmt.Println(change)
	}
}
