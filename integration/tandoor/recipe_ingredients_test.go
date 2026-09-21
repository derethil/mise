//go:build integration

package tandoor_test

import (
	"fmt"

	"github.com/tidwall/gjson"

	cleaningredients "github.com/derethil/mise/internal/ai/features/clean-ingredients"
)

func (s *TandoorSuite) applyCleanIngredients(fixture recipeFixture, rows ...cleaningredients.CleanedRow) updatedRecipe {
	cleaned := cleaningredients.CleanedRecipe{Ingredients: rows}

	return s.updateRecipe(s.T().Context(), fixture, func(raw []byte) []byte {
		out, _, err := cleaned.Apply(raw)
		s.Require().NoError(err, "clean-ingredients could not build a body")

		return out
	})
}

func (s *TandoorSuite) TestIngredientEdit() {
	fixture := s.createRecipeWithSteps(s.T().Context(), "mise-recipe-ingredients", 4)

	edited := s.applyCleanIngredients(fixture,
		cleaningredients.CleanedRow{Kind: cleaningredients.KindIngredient, Amount: 2, Food: s.fixtures.food.Name, Unit: createdUnit, Note: "reuse food, create unit"},
		cleaningredients.CleanedRow{Kind: cleaningredients.KindIngredient, Amount: 3, Food: createdFood, Unit: s.fixtures.unit.Name, Note: "create food, reuse unit"},
		cleaningredients.CleanedRow{Kind: cleaningredients.KindIngredient, Amount: 1, Food: s.fixtures.food.Name, Note: ""},
		cleaningredients.CleanedRow{Kind: cleaningredients.KindHeader, Note: "For the sauce"},
	)

	value := func(index int, path string) gjson.Result {
		result := gjson.GetBytes(edited.after, fmt.Sprintf("steps.0.ingredients.%d.%s", index, path))
		s.Require().True(result.Exists(), "steps.0.ingredients.%d.%s is missing", index, path)
		return result
	}

	s.Run("KeepsEveryRow", func() {
		s.EqualValues(4, gjson.GetBytes(edited.after, "steps.0.ingredients.#").Int())
	})

	s.Run("PatchesRowsInPlace", func() {
		for index := range 4 {
			s.EqualValues(
				gjson.GetBytes(edited.before, fmt.Sprintf("steps.0.ingredients.%d.id", index)).Int(),
				value(index, "id").Int(),
				"row %d was recreated instead of patched", index,
			)
		}
	})

	s.Run("PersistsAmounts", func() {
		s.EqualValues(2, value(0, "amount").Float())
		s.EqualValues(3, value(1, "amount").Float())
	})

	s.Run("PersistsNotes", func() {
		s.Equal("reuse food, create unit", value(0, "note").String())
		s.Equal("create food, reuse unit", value(1, "note").String())
		s.Equal("", value(2, "note").String())
	})

	s.Run("PersistsOrder", func() {
		for index := range 4 {
			s.EqualValues(index, value(index, "order").Int())
		}
	})

	s.Run("ReusesExistingFood", func() {
		s.EqualValues(s.fixtures.food.ID, value(0, "food.id").Int())
	})

	s.Run("ReusesExistingUnit", func() {
		s.EqualValues(s.fixtures.unit.ID, value(1, "unit.id").Int())
	})

	s.Run("CreatesUnknownFood", func() {
		s.Equal(createdFood, value(1, "food.name").String())
		s.NotZero(assertSingleResource(s, "food", s.client.Foods.SearchFoods, foodName, createdFood).ID)
	})

	s.Run("CreatesUnknownUnit", func() {
		s.Equal(createdUnit, value(0, "unit.name").String())
		s.NotZero(assertSingleResource(s, "unit", s.client.Units.SearchUnits, unitName, createdUnit).ID)
	})

	s.Run("DoesNotDuplicateExistingFood", func() {
		assertSingleResource(s, "food", s.client.Foods.SearchFoods, foodName, s.fixtures.food.Name)
	})

	s.Run("DoesNotDuplicateExistingUnit", func() {
		assertSingleResource(s, "unit", s.client.Units.SearchUnits, unitName, s.fixtures.unit.Name)
	})

	s.Run("KeepsEmptyUnitNull", func() {
		s.Equal(gjson.Null, value(2, "unit").Type)
	})

	s.Run("KeepsHeaderRow", func() {
		s.True(value(3, "is_header").Bool())
		s.Equal("For the sauce", value(3, "note").String())
		s.Equal(gjson.Null, value(3, "food").Type)
		s.Equal(gjson.Null, value(3, "unit").Type)
		s.False(value(2, "is_header").Bool(), "a normal row was marked as a header")
	})
}

func (s *TandoorSuite) TestIngredientJunkRemoval() {
	fixture := s.createRecipeWithSteps(s.T().Context(), "mise-recipe-junk", 2)

	edited := s.applyCleanIngredients(fixture,
		cleaningredients.CleanedRow{Kind: cleaningredients.KindJunk},
		cleaningredients.CleanedRow{Kind: cleaningredients.KindJunk},
	)

	s.Run("EmptiesTheStep", func() {
		ingredients := gjson.GetBytes(edited.after, "steps.0.ingredients")
		s.Require().True(ingredients.Exists(), "the step lost its ingredients key entirely")

		s.Empty(ingredients.Array())
	})
}

func (s *TandoorSuite) TestIngredientMultiStep() {
	fixture := s.createRecipeWithSteps(s.T().Context(), "mise-recipe-multistep", 1, 1)

	edited := s.applyCleanIngredients(fixture,
		cleaningredients.CleanedRow{Kind: cleaningredients.KindIngredient, Amount: 7, Food: s.fixtures.food.Name, Note: "first step"},
		cleaningredients.CleanedRow{Kind: cleaningredients.KindIngredient, Amount: 9, Food: s.fixtures.food.Name, Note: "second step"},
	)

	s.Run("PatchesEveryStep", func() {
		s.Require().EqualValues(2, gjson.GetBytes(edited.after, "steps.#").Int())

		s.EqualValues(7, gjson.GetBytes(edited.after, "steps.0.ingredients.0.amount").Float())
		s.EqualValues(9, gjson.GetBytes(edited.after, "steps.1.ingredients.0.amount").Float())
	})

	s.Run("DoesNotCrossStepBoundaries", func() {
		s.Equal("first step", gjson.GetBytes(edited.after, "steps.0.ingredients.0.note").String())
		s.Equal("second step", gjson.GetBytes(edited.after, "steps.1.ingredients.0.note").String())
	})
}

func (s *TandoorSuite) TestIngredientWithoutFood() {
	fixture := s.createRecipeWithSteps(s.T().Context(), "mise-recipe-no-food", 1)

	edited := s.applyCleanIngredients(fixture,
		cleaningredients.CleanedRow{Kind: cleaningredients.KindIngredient, Amount: 4, Note: "food the model could not name"},
	)

	s.Run("AcceptsNullFoodOnANormalRow", func() {
		row := gjson.GetBytes(edited.after, "steps.0.ingredients.0")
		s.Require().True(row.Exists(), "the row was dropped entirely")

		s.Equal(gjson.Null, row.Get("food").Type)
		s.False(row.Get("is_header").Bool(), "the row was turned into a header")
		s.EqualValues(4, row.Get("amount").Float())
	})
}
