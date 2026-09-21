//go:build integration

package tandoor_test

import (
	"fmt"

	"github.com/tidwall/gjson"

	"github.com/derethil/mise/internal/tandoor"
)

// Fields that a rename must leave untouched.
var preservedRecipePaths = []string{
	"servings", "working_time", "waiting_time",
	"keywords.#", "keywords.0.id",
	"steps.#", "steps.0.instruction",
	"steps.0.ingredients.#",
	"steps.0.ingredients.0.amount",
	"steps.0.ingredients.0.food.id",
	"steps.0.ingredients.0.unit.id",
}

func (s *TandoorSuite) TestRecipeRead() {
	s.Run("ListsFixtureID", func() {
		ids, err := s.client.Recipes.GetAllRecipeIDs(s.T().Context())
		s.Require().NoError(err)

		s.Contains(ids, s.fixtures.recipe.ID)
	})

	s.Run("ReturnsName", func() {
		recipe := s.getRecipe(s.fixtures.recipe.ID)

		s.Equal(s.fixtures.recipe.Name, recipe.Name)
	})

	s.Run("IncludesFixtureRelations", func() {
		raw := s.getRecipe(s.fixtures.recipe.ID).JSON()

		s.EqualValues(s.fixtures.keyword.ID, gjson.GetBytes(raw, "keywords.0.id").Int())
		s.EqualValues(s.fixtures.food.ID, gjson.GetBytes(raw, "steps.0.ingredients.0.food.id").Int())
		s.EqualValues(s.fixtures.unit.ID, gjson.GetBytes(raw, "steps.0.ingredients.0.unit.id").Int())
	})

	s.Run("ReportsMissingRecipeAsNotFound", func() {
		_, err := s.client.Recipes.Get(s.T().Context(), 999999)

		s.ErrorIs(err, tandoor.ErrTandoorNotFound)
	})
}

func (s *TandoorSuite) TestRecipeUpdate() {
	var edited updatedRecipe

	s.Run("AcceptsRenameBody", func() {
		edited = s.applyRecipeUpdate(s.T().Context(), "mise-recipe-rename", func(raw []byte) []byte {
			return s.setJSON(raw, "name", renamedRecipe)
		})
	})

	s.Run("PersistsRename", func() {
		s.Equal(renamedRecipe, gjson.GetBytes(edited.after, "name").String())
	})

	s.Run("PreservesOtherFields", func() {
		for _, path := range preservedRecipePaths {
			s.Require().NotEmpty(gjson.GetBytes(edited.before, path).String(),
				"%s was absent before the update, so comparing it proves nothing", path)

			s.Equal(
				gjson.GetBytes(edited.before, path).String(),
				gjson.GetBytes(edited.after, path).String(),
				"%s did not survive the update round trip", path,
			)
		}
	})
}

func (s *TandoorSuite) TestRecipeUpdatePreservesNastyRecipe() {
	fixture := s.createNastyRecipe(s.T().Context(), "mise-recipe-nasty")
	edited := s.updateRecipe(s.T().Context(), fixture, func(raw []byte) []byte {
		return s.setJSON(raw, "name", "mise-recipe-nasty-renamed")
	})

	s.Run("PreservesMetadataAndIngredientDetails", func() {
		paths := []string{
			"description", "source_url", "internal", "servings", "working_time", "waiting_time",
			"keywords.0.id", "steps.0.instruction",
			"steps.0.ingredients.0.amount", "steps.0.ingredients.0.note",
			"steps.0.ingredients.0.food.id", "steps.0.ingredients.0.unit.id",
			"steps.0.ingredients.0.checked", "steps.0.ingredients.0.conversions",
			"steps.0.ingredients.0.original_text",
			"steps.0.ingredients.1.is_header", "steps.0.ingredients.1.note",
			"steps.0.ingredients.1.food", "steps.0.ingredients.1.unit",
		}

		for _, path := range paths {
			before := gjson.GetBytes(edited.before, path)
			s.Require().True(before.Exists(), "%s was absent before the update", path)

			after := gjson.GetBytes(edited.after, path)
			s.Require().True(after.Exists(), "%s was absent after the update", path)
			s.Equal(before.Raw, after.Raw, "%s did not survive the update round trip", path)
		}
	})
}

func (s *TandoorSuite) getRecipe(id int) *tandoor.Recipe {
	recipe, err := s.client.Recipes.Get(s.T().Context(), id)
	s.Require().NoError(err, "fetch recipe %d", id)

	return recipe
}

func (s *TandoorSuite) TestRecipeUpdateRejectsBadBody() {
	ctx := s.T().Context()

	fixture := s.createRecipe(ctx, "mise-recipe-rejected")

	recipe, err := s.client.Recipes.Get(ctx, fixture.ID)
	s.Require().NoError(err)

	trimmed := s.setRawJSON(recipe.JSON(), "keywords",
		fmt.Sprintf(`[{"id":%d,"label":%q}]`, s.fixtures.keyword.ID, s.fixtures.keyword.Name))

	err = s.client.Recipes.Update(ctx, fixture.ID, trimmed)

	s.Run("ReportsRequestFailed", func() {
		s.ErrorIs(err, tandoor.ErrTandoorRequestFailed)
	})

	s.Run("LeavesTheRecipeUnchanged", func() {
		after, getErr := s.client.Recipes.Get(ctx, fixture.ID)
		s.Require().NoError(getErr)

		s.ElementsMatch(
			jsonStrings(recipe.JSON(), "keywords.#.name"),
			jsonStrings(after.JSON(), "keywords.#.name"),
		)
	})
}
