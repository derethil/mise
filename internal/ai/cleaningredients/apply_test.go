package cleaningredients

import (
	"testing"

	"github.com/derethil/mise/internal/ai"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

type ApplySuite struct {
	suite.Suite
}

func TestApplySuite(t *testing.T) {
	suite.Run(t, new(ApplySuite))
}

func (s *ApplySuite) recipe() []byte {
	return []byte(`{
		"id": 1,
		"steps": [
			{"ingredients": [
				{"original_text": "2 cups chopped onion", "order": 0},
				{"original_text": "some junk row", "order": 1},
				{"original_text": "For the sauce", "order": 2}
			]}
		]
	}`)
}

func (s *ApplySuite) TestApply_PatchesIngredientFields() {
	c := &CleanedRecipe{Ingredients: []CleanedRow{
		{Kind: KindIngredient, Amount: 2, Unit: "cup", Food: "onion", Note: "chopped"},
		{Kind: KindJunk},
		{Kind: KindHeader, Note: "For the sauce"},
	}}

	out, _, err := c.Apply(s.recipe())
	s.Require().NoError(err)

	ingredients := gjson.GetBytes(out, "steps.0.ingredients").Array()
	s.Require().Len(ingredients, 2, "junk row should be dropped")

	first := ingredients[0]
	s.Equal(false, first.Get("is_header").Bool())
	s.Equal(false, first.Get("no_amount").Bool())
	s.Equal(2.0, first.Get("amount").Float())
	s.Equal("cup", first.Get("unit.name").String())
	s.Equal("onion", first.Get("food.name").String())
	s.Equal("chopped", first.Get("note").String())
	s.Equal(int64(0), first.Get("order").Int())

	second := ingredients[1]
	s.True(second.Get("is_header").Bool())
	s.Equal("For the sauce", second.Get("note").String())
	s.True(second.Get("food").Type == gjson.Null, "header rows should not carry a food")
	s.Equal(int64(1), second.Get("order").Int(), "order is renumbered after dropping the junk row")
}

func (s *ApplySuite) TestApply_NoAmountWhenZero() {
	c := &CleanedRecipe{Ingredients: []CleanedRow{
		{Kind: KindIngredient, Food: "salt"},
		{Kind: KindJunk},
		{Kind: KindJunk},
	}}

	out, _, err := c.Apply(s.recipe())
	s.Require().NoError(err)

	ingredient := gjson.GetBytes(out, "steps.0.ingredients.0")
	s.True(ingredient.Get("no_amount").Bool())
	s.Equal(0.0, ingredient.Get("amount").Float())
	s.True(ingredient.Get("unit").Type == gjson.Null)
}

func (s *ApplySuite) TestApply_RecordsChangeForEveryOriginalRow() {
	c := &CleanedRecipe{Ingredients: []CleanedRow{
		{Kind: KindIngredient, Food: "onion"},
		{Kind: KindJunk},
		{Kind: KindHeader, Note: "For the sauce"},
	}}

	_, changes, err := c.Apply(s.recipe())
	s.Require().NoError(err)

	s.Require().Len(changes, 3, "a change is recorded even for dropped junk rows")
	s.Equal("2 cups chopped onion", changes[0].Before)
	s.Equal("some junk row", changes[1].Before)
	s.Equal(KindJunk, changes[1].Row.Kind)
}

func (s *ApplySuite) TestApply_ErrorsWhenCorrectionCountMismatches() {
	c := &CleanedRecipe{Ingredients: []CleanedRow{{Kind: KindIngredient, Food: "onion"}}}

	_, _, err := c.Apply(s.recipe())

	s.ErrorIs(err, ai.ErrMalformedResponse)
}

func (s *ApplySuite) TestApply_EmptyRecipe() {
	c := &CleanedRecipe{}

	out, changes, err := c.Apply([]byte(`{"id": 1, "steps": []}`))

	s.Require().NoError(err)
	s.Empty(changes)
	s.Equal(`{"id": 1, "steps": []}`, string(out))
}
