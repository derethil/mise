package cleaningredients

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/tidwall/gjson"
)

type projectedRecipe struct {
	Ingredients []string `json:"ingredients"`
}

func projectRecipe(raw []byte) projectedRecipe {
	projected := projectedRecipe{Ingredients: []string{}}

	gjson.GetBytes(raw, "steps").ForEach(func(_, step gjson.Result) bool {
		step.Get("ingredients").ForEach(func(_, ingredient gjson.Result) bool {
			projected.Ingredients = append(projected.Ingredients, projectIngredient(ingredient))
			return true
		})

		return true
	})

	return projected
}

func projectIngredient(ingredient gjson.Result) string {
	parts := make([]string, 0, 4)
	if amount := ingredient.Get("amount").Float(); amount != 0 {
		parts = append(parts, strconv.FormatFloat(amount, 'g', -1, 64))
	}

	for _, path := range []string{"unit.name", "food.name", "note"} {
		if part := cleaned(ingredient.Get(path).String()); part != "" {
			parts = append(parts, part)
		}
	}

	return strings.Join(parts, " ")
}

func cleaned(text string) string {
	stripped := strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, text)

	return strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(stripped), "•‣▪·*-–— \t "))
}
