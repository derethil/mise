package assignkeywords

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/tidwall/gjson"
)

const maxStepChars = 6000

type projectedRecipe struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	CurrentKeywords []string `json:"current_keywords"`
	Ingredients     []string `json:"ingredients"`
	Steps           []string `json:"steps"`
}

func projectRecipe(raw []byte) projectedRecipe {
	projected := projectedRecipe{
		Name:            cleaned(gjson.GetBytes(raw, "name").String()),
		Description:     cleaned(gjson.GetBytes(raw, "description").String()),
		CurrentKeywords: []string{},
		Ingredients:     []string{},
		Steps:           []string{},
	}

	gjson.GetBytes(raw, "keywords").ForEach(func(_, keyword gjson.Result) bool {
		if name := cleaned(keyword.Get("name").String()); name != "" {
			projected.CurrentKeywords = append(projected.CurrentKeywords, name)
		}
		return true
	})

	seen := map[string]bool{}
	budget := maxStepChars

	gjson.GetBytes(raw, "steps").ForEach(func(_, step gjson.Result) bool {
		step.Get("ingredients").ForEach(func(_, ingredient gjson.Result) bool {
			if ingredient.Get("is_header").Bool() {
				return true
			}

			food := cleaned(ingredient.Get("food.name").String())
			if food == "" || seen[strings.ToLower(food)] {
				return true
			}

			seen[strings.ToLower(food)] = true
			projected.Ingredients = append(projected.Ingredients, food)
			return true
		})

		instruction := cleaned(step.Get("instruction").String())
		if instruction == "" || budget <= 0 {
			return true
		}

		instruction = truncate(instruction, budget)
		budget -= len(instruction)

		projected.Steps = append(projected.Steps, instruction)
		return true
	})

	return projected
}

func truncate(text string, limit int) string {
	if len(text) <= limit {
		return text
	}

	for limit > 0 && !utf8.RuneStart(text[limit]) {
		limit--
	}

	return text[:limit] + "…"
}

func cleaned(text string) string {
	stripped := strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, text)

	return strings.TrimSpace(stripped)
}
