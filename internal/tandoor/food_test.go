package tandoor

import "net/http"

func (s *ClientSuite) TestFoodsSearch() {
	s.response = map[string]any{
		"results": []map[string]any{
			{"id": 1, "name": "Chicken Thigh", "plural_name": "Chicken Thighs"},
		},
	}

	foods, err := s.client.Foods.SearchFoods(s.T().Context(), "chicken")

	s.Require().NoError(err)
	s.Equal([]Food{{ID: 1, Name: "Chicken Thigh", PluralName: "Chicken Thighs"}}, foods)
	s.Equal(http.MethodGet, s.lastMethod)
	s.Equal("/api/food/", s.lastPath)
	s.Equal("simple=true&page_size=200&query=chicken", s.lastQuery)
}

func (s *ClientSuite) TestFoodsSearch_EscapesQuery() {
	s.response = map[string]any{"results": []map[string]any{}}

	_, err := s.client.Foods.SearchFoods(s.T().Context(), "chicken thigh")

	s.Require().NoError(err)
	s.Contains(s.lastQuery, "query=chicken+thigh")
}

func (s *ClientSuite) TestFoodsSearch_HTTPError() {
	s.status = http.StatusForbidden

	_, err := s.client.Foods.SearchFoods(s.T().Context(), "chicken")

	s.Error(err)
}

func (s *ClientSuite) TestFoodsSearch_InvalidPayload() {
	s.response = map[string]any{"results": "not an array"}

	_, err := s.client.Foods.SearchFoods(s.T().Context(), "chicken")

	s.Error(err)
}
