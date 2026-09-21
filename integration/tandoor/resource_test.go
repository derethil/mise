//go:build integration

package tandoor_test

import (
	"context"
	"strings"

	"github.com/derethil/mise/internal/tandoor"
)

const noMatchQuery = "zzz-definitely-nothing"

type resourceCase[T any] struct {
	label   string
	query   string
	partial string
	wantID  int

	list   func(context.Context, ...string) ([]T, error)
	search func(context.Context, string, ...string) ([]T, error)

	id     func(T) int
	name   func(T) string
	plural func(T) string

	wantPlural string
}

func assertResource[T any](s *TandoorSuite, testCase resourceCase[T]) {
	s.Run("ListIncludesFixture", func() {
		listed, err := testCase.list(s.T().Context())
		s.Require().NoError(err)

		s.Contains(resourceIDs(listed, testCase.id), testCase.wantID)
	})

	s.Run("SearchByNameFindsFixture", func() {
		s.Contains(resourceIDs(searchResource(s, testCase, testCase.query), testCase.id), testCase.wantID)
	})

	s.Run("SearchByInfixFindsFixture", func() {
		s.Contains(resourceIDs(searchResource(s, testCase, testCase.partial), testCase.id), testCase.wantID)
	})

	s.Run("SearchIsCaseInsensitive", func() {
		found := searchResource(s, testCase, strings.ToUpper(testCase.query))

		s.Contains(resourceIDs(found, testCase.id), testCase.wantID)
	})

	s.Run("SearchWithoutMatchesReturnsNothing", func() {
		s.Empty(searchResource(s, testCase, noMatchQuery))
	})

	s.Run("FieldsDecode", func() {
		found := searchResource(s, testCase, testCase.query)
		s.Require().Contains(resourceIDs(found, testCase.id), testCase.wantID,
			"the %s fixture was not in the search results, so its fields cannot be checked", testCase.label)

		fixture := resourceByID(found, testCase.id, testCase.wantID)

		s.Equal(testCase.query, testCase.name(fixture))

		if testCase.plural != nil {
			s.Equal(testCase.wantPlural, testCase.plural(fixture))
		}
	})
}

func searchResource[T any](s *TandoorSuite, testCase resourceCase[T], query string) []T {
	found, err := testCase.search(s.T().Context(), query)
	s.Require().NoError(err, "search %ss for %q", testCase.label, query)

	return found
}

func resourceIDs[T any](items []T, id func(T) int) []int {
	ids := make([]int, 0, len(items))
	for _, item := range items {
		ids = append(ids, id(item))
	}

	return ids
}

func resourceByID[T any](items []T, id func(T) int, want int) T {
	var found T

	for _, item := range items {
		if id(item) == want {
			return item
		}
	}

	return found
}

func assertSingleResource[T any](
	s *TandoorSuite,
	label string,
	search func(context.Context, string, ...string) ([]T, error),
	name func(T) string,
	want string,
) T {
	found, err := search(s.T().Context(), want)
	s.Require().NoError(err, "search %ss for %q", label, want)

	var matches []T
	for _, item := range found {
		if name(item) == want {
			matches = append(matches, item)
		}
	}

	s.Require().Len(matches, 1,
		"expected exactly one %s named %q, found %d among %d search hits",
		label, want, len(matches), len(found))

	return matches[0]
}

func keywordID(keyword tandoor.Keyword) int      { return keyword.ID }
func keywordName(keyword tandoor.Keyword) string { return keyword.Name }
func unitID(unit tandoor.Unit) int               { return unit.ID }
func unitName(unit tandoor.Unit) string          { return unit.Name }
func foodID(food tandoor.Food) int               { return food.ID }
func foodName(food tandoor.Food) string          { return food.Name }

func (s *TandoorSuite) TestKeywords() {
	assertResource(s, resourceCase[tandoor.Keyword]{
		label:   "keyword",
		query:   s.fixtures.keyword.Name,
		partial: "ise-keywor",
		wantID:  s.fixtures.keyword.ID,
		list:    s.client.Keywords.ListKeywords,
		search:  s.client.Keywords.SearchKeywords,
		id:      keywordID,
		name:    keywordName,
	})
}

func (s *TandoorSuite) TestUnits() {
	assertResource(s, resourceCase[tandoor.Unit]{
		label:   "unit",
		query:   s.fixtures.unit.Name,
		partial: "ise-uni",
		wantID:  s.fixtures.unit.ID,
		list:    s.client.Units.All,
		search:  s.client.Units.SearchUnits,
		id:      unitID,
		name:    unitName,
	})
}

func (s *TandoorSuite) TestFoods() {
	assertResource(s, resourceCase[tandoor.Food]{
		label:      "food",
		query:      s.fixtures.food.Name,
		partial:    "ise-foo",
		wantID:     s.fixtures.food.ID,
		list:       s.client.Foods.All,
		search:     s.client.Foods.SearchFoods,
		id:         foodID,
		name:       foodName,
		plural:     func(food tandoor.Food) string { return food.PluralName },
		wantPlural: s.fixtures.food.PluralName,
	})
}

func (s *TandoorSuite) TestUnitPluralName() {
	s.Run("MirrorsNameInsteadOfStoringIt", func() {
		unit := assertSingleResource(s, "unit", s.client.Units.SearchUnits, unitName, s.fixtures.unit.Name)

		s.NotEqual(sentUnitPlural, unit.PluralName,
			"tandoor started storing unit plural names, so mise can rely on the field")
		s.Equal(unit.Name, unit.PluralName, "tandoor no longer mirrors the name into plural_name")
	})
}

func (s *TandoorSuite) TestSearchIsNotPaginated() {
	s.Run("ReturnsOnlyTheFirstPage", func() {
		all, err := s.client.Keywords.SearchKeywords(s.T().Context(), "mise")
		s.Require().NoError(err)
		s.Require().Greater(len(all), paginationPageSize, "not enough keywords match to truncate a page")

		page, err := s.client.Keywords.SearchKeywords(s.T().Context(), "mise", "page_size", "2")
		s.Require().NoError(err)

		s.Len(page, paginationPageSize,
			"ResourceService.Search does not follow next links, so it returns a single page")
	})
}
