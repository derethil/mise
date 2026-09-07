package cleaningredients

import (
	"fmt"
	"strings"

	"github.com/derethil/mise/internal/ai"
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

func (c *CleanedRecipe) Apply(rawBefore []byte) ([]byte, []Change, error) {
	if err := c.validate(rawBefore); err != nil {
		return nil, nil, err
	}

	out := rawBefore
	index := 0
	changes := make([]Change, 0, len(c.Ingredients))

	for position, step := range gjson.GetBytes(rawBefore, "steps").Array() {
		var kept []string
		for _, ingredient := range step.Get("ingredients").Array() {
			patchInput := c.Ingredients[index]
			index++

			before := projectIngredient(ingredient)
			changes = append(changes, Change{Before: before, Row: patchInput})

			if patchInput.Kind == KindJunk {
				continue
			}

			patched, err := patchIngredient(ingredient.Raw, patchInput)
			if err != nil {
				return nil, nil, err
			}

			patched, err = sjson.Set(patched, "order", len(kept))
			if err != nil {
				return nil, nil, err
			}

			kept = append(kept, patched)
		}

		var err error

		out, err = sjson.SetRawBytes(out, fmt.Sprintf("steps.%d.ingredients", position), []byte("["+strings.Join(kept, ",")+"]"))
		if err != nil {
			return nil, nil, err
		}
	}

	return out, changes, nil
}

func (c *CleanedRecipe) validate(raw []byte) error {
	rows := len(projectRecipe(raw).Ingredients)
	if len(c.Ingredients) != rows {
		return fmt.Errorf("%w: have %d corrections for a recipe with %d rows", ai.ErrMalformedResponse, len(c.Ingredients), rows)
	}

	return nil
}

func patchIngredient(raw string, fix CleanedRow) (string, error) {
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
