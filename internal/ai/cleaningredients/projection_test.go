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

func (s *ProjectionSuite) TestProjectIngredient_PrefersOriginalText() {
	raw := []byte(`{
		"original_text": "2 cups chopped onion",
		"amount": 1, "unit": {"name": "cup"}, "food": {"name": "onion"}, "note": "chopped"
	}`)

	s.Equal("2 cups chopped onion", projectIngredient(gjson.ParseBytes(raw)))
}

func (s *ProjectionSuite) TestProjectIngredient_StripsBulletsAndWhitespace() {
	raw := []byte(`{"original_text": "  • 2 cups onion  "}`)

	s.Equal("2 cups onion", projectIngredient(gjson.ParseBytes(raw)))
}

func (s *ProjectionSuite) TestProjectIngredient_FallsBackToStructuredFields() {
	raw := []byte(`{
		"original_text": "",
		"amount": 2, "unit": {"name": "cup"}, "food": {"name": "onion"}, "note": "chopped"
	}`)

	s.Equal("2 cup onion chopped", projectIngredient(gjson.ParseBytes(raw)))
}

func (s *ProjectionSuite) TestProjectIngredient_FallbackOmitsEmptyParts() {
	raw := []byte(`{"original_text": "", "amount": 0, "food": {"name": "salt"}}`)

	s.Equal("salt", projectIngredient(gjson.ParseBytes(raw)))
}

func (s *ProjectionSuite) TestProjectRecipe_FlattensAllStepsInOrder() {
	raw := []byte(`{
		"steps": [
			{"ingredients": [{"original_text": "2 cups flour"}, {"original_text": "1 egg"}]},
			{"ingredients": [{"original_text": "pinch of salt"}]}
		]
	}`)

	projected := projectRecipe(raw)

	s.Equal([]string{"2 cups flour", "1 egg", "pinch of salt"}, projected.Ingredients)
}

func (s *ProjectionSuite) TestProjectRecipe_NoSteps() {
	projected := projectRecipe([]byte(`{"steps": []}`))

	s.Empty(projected.Ingredients)
}
