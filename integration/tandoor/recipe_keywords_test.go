//go:build integration

package tandoor_test

import (
	"github.com/tidwall/gjson"

	assignkeywords "github.com/derethil/mise/internal/ai/features/assign-keywords"
	"github.com/derethil/mise/internal/tandoor"
)

func (s *TandoorSuite) applyAssignKeywords(fixture recipeFixture, assigned assignkeywords.AssignedKeywords, opts assignkeywords.ApplyOptions) updatedRecipe {
	ctx := s.T().Context()

	vocabulary, err := s.client.Keywords.ListKeywords(ctx)
	s.Require().NoError(err, "load the keyword vocabulary")

	return s.updateRecipe(ctx, fixture, func(raw []byte) []byte {
		out, _, err := assigned.Apply(raw, vocabulary, opts)
		s.Require().NoError(err, "assign-keywords could not build a body")

		return out
	})
}

func (s *TandoorSuite) TestKeywordEdit() {
	known := s.fixtures.paginated[0]

	edited := s.applyAssignKeywords(
		s.createRecipe(s.T().Context(), "mise-recipe-keywords"),
		assignkeywords.AssignedKeywords{
			{Name: known.Name, Applies: true},
			{Name: createdKeyword, Applies: true},
		},
		assignkeywords.ApplyOptions{},
	)

	names := jsonStrings(edited.after, "keywords.#.name")

	linkedKeywordID := func(name string) gjson.Result {
		result := gjson.GetBytes(edited.after, `keywords.#(name=="`+name+`").id`)
		s.Require().True(result.Exists(), "the recipe has no keyword named %q", name)
		return result
	}

	s.Run("CarriesAllThreeShapes", func() {
		s.ElementsMatch([]string{s.fixtures.keyword.Name, known.Name, createdKeyword}, names)
	})

	s.Run("KeepsKeywordAlreadyOnRecipe", func() {
		s.EqualValues(s.fixtures.keyword.ID, linkedKeywordID(s.fixtures.keyword.Name).Int())
	})

	s.Run("LinksKnownKeywordByID", func() {
		s.EqualValues(known.ID, linkedKeywordID(known.Name).Int())
	})

	s.Run("CreatesUnknownKeyword", func() {
		created := assertSingleResource(s, "keyword", s.client.Keywords.SearchKeywords, keywordName, createdKeyword)
		s.Require().NotZero(created.ID)
		s.EqualValues(created.ID, linkedKeywordID(createdKeyword).Int())
	})

	s.Run("DoesNotDuplicateExistingKeyword", func() {
		assertSingleResource(s, "keyword", s.client.Keywords.SearchKeywords, keywordName, s.fixtures.keyword.Name)
	})
}

func (s *TandoorSuite) TestKeywordReplaceClearsAll() {
	edited := s.applyAssignKeywords(
		s.createRecipe(s.T().Context(), "mise-recipe-keywords-cleared"),
		assignkeywords.AssignedKeywords{},
		assignkeywords.ApplyOptions{Replace: true},
	)

	s.Run("StartsWithAKeyword", func() {
		s.Require().NotEmpty(jsonStrings(edited.before, "keywords.#.name"))
	})

	s.Run("RemovesEveryKeyword", func() {
		keywords := gjson.GetBytes(edited.after, "keywords")
		s.Require().True(keywords.Exists(), "the recipe lost its keywords key entirely")
		s.Empty(keywords.Array())
	})

	s.Run("LeavesTheKeywordItself", func() {
		var found tandoor.Keyword = assertSingleResource(s, "keyword", s.client.Keywords.SearchKeywords, keywordName, s.fixtures.keyword.Name)

		s.Equal(s.fixtures.keyword.ID, found.ID)
	})
}
