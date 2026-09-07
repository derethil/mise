// Package cleaningredients implements the miseai.Feature for cleaning up recipe ingredients.
package cleaningredients

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	miseai "github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

func init() {
	miseai.RegisterFeature(func() miseai.Feature { return &Feature{} })
}

const (
	cleanBatchMaxAttempts = 2
	ingredientBatchSize   = 2
)

var cleanIngredientsConfig = miseai.GenerateConfig{
	Temperature: new(0.1),
	Reasoning:   new(false),
}

type CleanedRow struct {
	Kind   string  `json:"kind" jsonschema:"enum=ingredient,enum=header,enum=junk" jsonschema_description:"Whether the row is a real ingredient, a section heading to keep, or junk to delete."`
	Amount float64 `json:"amount" jsonschema_description:"The quantity as a number, taken from the text. A range becomes its midpoint. Only 0 when the text gives no quantity at all, or the row is a header or junk."`
	Unit   string  `json:"unit" jsonschema_description:"The unit of measurement, empty when the row has no real unit, or is a header or junk."`
	Food   string  `json:"food" jsonschema_description:"The ingredient itself, with no amount, unit, or preparation in it. Empty for header and junk rows."`
	Note   string  `json:"note" jsonschema_description:"Preparation and anything else the line said. Empty for header and junk rows."`
}

type CleanIngredientBatchInput struct {
	Rows []string `json:"rows" jsonschema_description:"Each element is one ingredient row's original_text, in original order."`
}

type CleanedRowBatch struct {
	Rows []CleanedRow `json:"rows" jsonschema_description:"One cleaned row per input row, in the same order and count as the input rows."`
}

type Feature struct {
	opts   []ai.PromptExecuteOption
	prompt *ai.DataPrompt[CleanIngredientBatchInput, *CleanedRowBatch]
	flow   *core.Flow[projectedRecipe, *CleanedRecipe, struct{}]
}

func (c *Feature) Register(r miseai.Registry) error {
	genkit.DefineSchemaFor[CleanedRow](r.Genkit)
	genkit.DefineSchemaFor[CleanIngredientBatchInput](r.Genkit)
	genkit.DefineSchemaFor[CleanedRowBatch](r.Genkit)

	c.prompt = genkit.LookupDataPrompt[CleanIngredientBatchInput, *CleanedRowBatch](r.Genkit, "clean_ingredients")
	if c.prompt == nil {
		return fmt.Errorf("%w: clean_ingredients", miseai.ErrPromptNotFound)
	}

	c.opts = r.PromptOptions(cleanIngredientsConfig,
		ai.WithTools(
			miseai.SearchFoodsTool(r),
			miseai.SearchUnitsTool(r),
		),
		ai.WithMaxTurns(2*ingredientBatchSize+2),
	)

	c.flow = genkit.DefineFlow(r.Genkit, "cleanRecipeIngredients", c.cleanRecipe)

	return nil
}

func (c *Feature) CleanRecipe(ctx context.Context, recipe *tandoor.Recipe) (*CleanedRecipe, error) {
	cleaned, err := c.flow.Run(ctx, projectRecipe(recipe.JSON()))
	if err != nil {
		return nil, fmt.Errorf("recipe %d: %w", recipe.ID, err)
	}

	return cleaned, nil
}

func (c *Feature) cleanRecipe(ctx context.Context, projected projectedRecipe) (*CleanedRecipe, error) {
	if len(projected.Ingredients) == 0 {
		return &CleanedRecipe{}, nil
	}

	cleaned := &CleanedRecipe{Ingredients: make([]CleanedRow, 0, len(projected.Ingredients))}

	for start := 0; start < len(projected.Ingredients); start += ingredientBatchSize {
		end := min(start+ingredientBatchSize, len(projected.Ingredients))

		fixes, err := c.cleanRowBatch(ctx, projected.Ingredients[start:end])
		if err != nil {
			return nil, fmt.Errorf("rows %d-%d: %w", start, end-1, err)
		}

		cleaned.Ingredients = append(cleaned.Ingredients, fixes...)

		slog.InfoContext(ctx, "cleaned ingredient batch", "remaining", len(projected.Ingredients)-end)
	}

	return cleaned, nil
}

func (c *Feature) cleanRowBatch(ctx context.Context, rows []string) ([]CleanedRow, error) {
	input := CleanIngredientBatchInput{Rows: rows}

	var batch *CleanedRowBatch
	var err error

	for attempt := 1; attempt <= cleanBatchMaxAttempts; attempt++ {
		batch, _, err = c.prompt.Execute(ctx, input, c.opts...)
		if err != nil {
			return nil, fmt.Errorf("cleanRowBatch: %w", err)
		}
	}

	normalized, err := normalizeRows(batch, rows)
	if err != nil {
		return nil, err
	}

	return normalized, nil
}

func normalizeRows(batch *CleanedRowBatch, rows []string) ([]CleanedRow, error) {
	if batch == nil {
		return nil, errors.New("cleanRowBatch: model returned nil batch")
	}

	if len(batch.Rows) != len(rows) {
		return nil, fmt.Errorf("cleanRowBatch: want %d rows, got %d", len(rows), len(batch.Rows))
	}

	result := make([]CleanedRow, len(rows))
	for i, row := range rows {
		result[i] = normalizeRow(row, batch.Rows[i])
	}
	return result, nil
}

func normalizeRow(text string, cleaned CleanedRow) CleanedRow {
	switch strings.ToLower(cleaned.Kind) {
	case KindHeader:
		return CleanedRow{Kind: KindHeader, Note: text}
	case KindJunk:
		return CleanedRow{Kind: KindJunk}
	default:
		cleaned.Kind = KindIngredient
		return cleaned
	}
}
