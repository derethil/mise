package cleaningredients

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

type ProjectionSuite struct {
	suite.Suite
}

func TestProjectionSuite(t *testing.T) {
	suite.Run(t, new(ProjectionSuite))
}

func (s *ProjectionSuite) TestProjectIngredient_IgnoresOriginalText() {
	raw := []byte(`{
		"original_text": "2 cups chopped onion",
		"amount": 1, "unit": {"name": "cup"}, "food": {"name": "onion"}, "note": "chopped"
	}`)

	s.Equal("1 cup onion chopped", projectIngredient(gjson.ParseBytes(raw)))
}

func (s *ProjectionSuite) TestProjectIngredient_StripsBulletsAndWhitespace() {
	raw := []byte(`{"food": {"name": "  • onion  "}}`)

	s.Equal("onion", projectIngredient(gjson.ParseBytes(raw)))
}

func (s *ProjectionSuite) TestProjectIngredient_ConstructsFromStructuredFields() {
	raw := []byte(`{
		"amount": 2, "unit": {"name": "cup"}, "food": {"name": "onion"}, "note": "chopped"
	}`)

	s.Equal("2 cup onion chopped", projectIngredient(gjson.ParseBytes(raw)))
}

func (s *ProjectionSuite) TestProjectIngredient_OmitsEmptyParts() {
	raw := []byte(`{"amount": 0, "food": {"name": "salt"}}`)

	s.Equal("salt", projectIngredient(gjson.ParseBytes(raw)))
}

func (s *ProjectionSuite) TestProjectRecipe_FlattensAllStepsInOrder() {
	raw := []byte(`{
		"steps": [
			{"ingredients": [{"food": {"name": "flour"}}, {"food": {"name": "egg"}}]},
			{"ingredients": [{"food": {"name": "salt"}}]}
		]
	}`)

	projected := projectRecipe(raw)

	s.Equal([]string{"flour", "egg", "salt"}, projected.Ingredients)
}

func (s *ProjectionSuite) TestProjectRecipe_NoSteps() {
	projected := projectRecipe([]byte(`{"steps": []}`))

	s.Empty(projected.Ingredients)
	s.NotNil(projected.Ingredients, "must serialize as [] rather than null for genkit's schema validation")
}

func (s *ProjectionSuite) TestProjectRecipe_MissingStepsField() {
	projected := projectRecipe([]byte(`{}`))

	s.Empty(projected.Ingredients)
	s.NotNil(projected.Ingredients, "must serialize as [] rather than null for genkit's schema validation")
}
