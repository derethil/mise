//go:build integration

package tandoor_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tidwall/sjson"

	"github.com/derethil/mise/internal/tandoor"
)

const (
	paginationRecords = 5

	createdKeyword = "mise-created-keyword"
	createdUnit    = "mise-created-unit"
	createdFood    = "mise-created-food"
	renamedRecipe  = "mise-recipe-renamed"
	sentUnitPlural = "mise-units"
)

type recipeFixture struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type updatedRecipe struct {
	id     int
	before []byte
	after  []byte
}

type fixtures struct {
	keyword tandoor.Keyword
	unit    tandoor.Unit
	food    tandoor.Food
	recipe  recipeFixture

	paginated []tandoor.Keyword
}

func (s *TandoorSuite) createFixtures(ctx context.Context) {
	s.fixtures.keyword = createResource[tandoor.Keyword](s, ctx, "keyword", map[string]any{
		"name": "mise-keyword",
	})

	s.fixtures.unit = createResource[tandoor.Unit](s, ctx, "unit", map[string]any{
		"name":        "mise-unit",
		"plural_name": sentUnitPlural,
	})

	s.fixtures.food = createResource[tandoor.Food](s, ctx, "food", map[string]any{
		"name":        "mise-food",
		"plural_name": "mise-foods",
	})

	createResource[tandoor.Food](s, ctx, "food", map[string]any{"name": "xx-mise-food-xx"})
	createResource[tandoor.Unit](s, ctx, "unit", map[string]any{"name": "xx-mise-unit-xx"})
	createResource[tandoor.Keyword](s, ctx, "keyword", map[string]any{"name": "xx-mise-keyword-xx"})

	s.fixtures.recipe = s.createRecipe(ctx, "mise-recipe")
	s.fixtures.paginated = s.createPaginationRecords(ctx)
}

func (s *TandoorSuite) applyRecipeUpdate(ctx context.Context, name string, mutate func([]byte) []byte) updatedRecipe {
	return s.updateRecipe(ctx, s.createRecipe(ctx, name), mutate)
}

func (s *TandoorSuite) updateRecipe(ctx context.Context, fixture recipeFixture, mutate func([]byte) []byte) updatedRecipe {
	before, err := s.client.Recipes.Get(ctx, fixture.ID)
	s.Require().NoError(err, "fetch recipe %d", fixture.ID)

	err = s.client.Recipes.Update(ctx, fixture.ID, mutate(before.JSON()))
	s.Require().NoError(err, "tandoor rejected the body mise produced for recipe %d", fixture.ID)

	after, err := s.client.Recipes.Get(ctx, fixture.ID)
	s.Require().NoError(err, "refetch recipe %d", fixture.ID)

	return updatedRecipe{id: fixture.ID, before: before.JSON(), after: after.JSON()}
}

func (s *TandoorSuite) createPaginationRecords(ctx context.Context) []tandoor.Keyword {
	records := make([]tandoor.Keyword, 0, paginationRecords)

	for index := range paginationRecords {
		records = append(records, createResource[tandoor.Keyword](s, ctx, "keyword", map[string]any{
			"name": fmt.Sprintf("mise-page-%d", index),
		}))
	}

	return records
}

func (s *TandoorSuite) createRecipe(ctx context.Context, name string) recipeFixture {
	return s.createRecipeWithSteps(ctx, name, 1)
}

func (s *TandoorSuite) createRecipeWithSteps(ctx context.Context, name string, ingredientsPerStep ...int) recipeFixture {
	steps := make([]map[string]any, 0, len(ingredientsPerStep))

	for position, count := range ingredientsPerStep {
		ingredients := make([]map[string]any, 0, count)

		for range count {
			ingredients = append(ingredients, map[string]any{
				"amount": 1,
				"food":   map[string]any{"id": s.fixtures.food.ID, "name": s.fixtures.food.Name},
				"unit":   map[string]any{"id": s.fixtures.unit.ID, "name": s.fixtures.unit.Name},
			})
		}

		steps = append(steps, map[string]any{
			"instruction": fmt.Sprintf("Step %d.", position+1),
			"ingredients": ingredients,
		})
	}

	return createResource[recipeFixture](s, ctx, "recipe", map[string]any{
		"name":         name,
		"working_time": 0,
		"waiting_time": 0,
		"servings":     1,
		"keywords":     []map[string]any{{"id": s.fixtures.keyword.ID, "name": s.fixtures.keyword.Name}},
		"steps":        steps,
	})
}

func (s *TandoorSuite) createNastyRecipe(ctx context.Context, name string) recipeFixture {
	return createResource[recipeFixture](s, ctx, "recipe", map[string]any{
		"name":         name,
		"description":  "Crème brûlée — keep the accents, punctuation, and context.",
		"source_url":   "https://example.invalid/recipes/creme-brulee?servings=4",
		"internal":     true,
		"working_time": 25,
		"waiting_time": 90,
		"servings":     4,
		"keywords":     []map[string]any{{"id": s.fixtures.keyword.ID, "name": s.fixtures.keyword.Name}},
		"steps": []map[string]any{{
			"instruction": "Whisk gently; do not boil.",
			"ingredients": []map[string]any{
				{
					"amount": 0.5,
					"food":   map[string]any{"id": s.fixtures.food.ID, "name": s.fixtures.food.Name},
					"unit":   map[string]any{"id": s.fixtures.unit.ID, "name": s.fixtures.unit.Name},
					"note":   "room temperature",
				},
				{
					"amount":    0,
					"food":      nil,
					"unit":      nil,
					"is_header": true,
					"note":      "For the caramel",
				},
			},
		}},
	})
}

func (s *TandoorSuite) setJSON(raw []byte, path string, value any) []byte {
	out, err := sjson.SetBytes(raw, path, value)
	s.Require().NoError(err, "set %s", path)

	return out
}

func (s *TandoorSuite) setRawJSON(raw []byte, path, value string) []byte {
	out, err := sjson.SetRawBytes(raw, path, []byte(value))
	s.Require().NoError(err, "set %s", path)

	return out
}

func createResource[T any](s *TandoorSuite, ctx context.Context, resource string, body any) T {
	var created T

	s.request(ctx, http.MethodPost, "/api/"+resource+"/", body, &created)

	return created
}
