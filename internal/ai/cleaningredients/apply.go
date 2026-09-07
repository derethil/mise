package cleaningredients

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	KindIngredient = "ingredient"
	KindHeader     = "header"
	KindJunk       = "junk"
)

type CleanedRecipe struct {
	Ingredients []CleanedRow `json:"ingredients"`
}

func (c *CleanedRecipe) Apply(ctx context.Context, raw []byte) ([]byte, error) {
	if err := c.validate(raw); err != nil {
		return nil, err
	}

	out := raw
	index := 0
	for position, step := range gjson.GetBytes(raw, "steps").Array() {
		var kept []string
		for _, ingredient := range step.Get("ingredients").Array() {
			fix := c.Ingredients[index]
			index++

			before := projectIngredient(ingredient)

			if fix.Kind == KindJunk {
				slog.InfoContext(ctx, "dropped ingredient", slog.String("before", before))
				continue
			}

			slog.InfoContext(ctx, "cleaned ingredient",
				slog.String("before", before),
				slog.String("kind", fix.Kind),
				slog.Float64("amount", fix.Amount),
				slog.String("unit", fix.Unit),
				slog.String("food", fix.Food),
				slog.String("note", fix.Note),
			)

			patched, err := applyIngredientFix(ingredient.Raw, fix)
			if err != nil {
				return nil, err
			}

			patched, err = sjson.Set(patched, "order", len(kept))
			if err != nil {
				return nil, err
			}

			kept = append(kept, patched)
		}

		var err error

		out, err = sjson.SetRawBytes(out, fmt.Sprintf("steps.%d.ingredients", position), []byte("["+strings.Join(kept, ",")+"]"))
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

func (c *CleanedRecipe) validate(raw []byte) error {
	rows := len(projectRecipe(raw).Ingredients)
	if len(c.Ingredients) != rows {
		return fmt.Errorf("have %d corrections for a recipe with %d rows", len(c.Ingredients), rows)
	}

	return nil
}

func applyIngredientFix(raw string, fix CleanedRow) (string, error) {
	if fix.Kind == KindHeader {
		fix = CleanedRow{Kind: KindHeader, Note: fix.Note}
	}

	out := raw
	var err error

	set := func(path string, value any) {
		if err == nil {
			out, err = sjson.Set(out, path, value)
		}
	}

	set("is_header", fix.Kind == KindHeader)
	set("no_amount", fix.Amount == 0)
	set("amount", fix.Amount)
	set("note", fix.Note)
	set("food", named(fix.Food))
	set("unit", named(fix.Unit))

	return out, err
}

// Replaces the whole object rather than patching to force Tandoor to create it new
// or use the existing one if it fully matches.
func named(name string) any {
	if name == "" {
		return nil
	}

	return map[string]string{"name": name}
}
