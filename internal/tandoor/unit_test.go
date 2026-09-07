package tandoor

import "net/http"

func (s *ClientSuite) TestUnitsSearch() {
	s.response = map[string]any{
		"results": []map[string]any{
			{"id": 1, "name": "cup", "plural_name": "cups"},
		},
	}

	units, err := s.client.Units.SearchUnits(s.T().Context(), "cup", "page_size", "200")

	s.Require().NoError(err)
	s.Equal([]Unit{{ID: 1, Name: "cup", PluralName: "cups"}}, units)
	s.Equal(http.MethodGet, s.lastMethod)
	s.Equal("/api/unit/", s.lastPath)
	s.Equal("page_size=200&query=cup", s.lastQuery)
}

func (s *ClientSuite) TestUnitsSearch_HTTPError() {
	s.status = http.StatusForbidden

	_, err := s.client.Units.SearchUnits(s.T().Context(), "cup")

	s.Error(err)
}

func (s *ClientSuite) TestUnitsSearch_InvalidPayload() {
	s.response = map[string]any{"results": "not an array"}

	_, err := s.client.Units.SearchUnits(s.T().Context(), "cup")

	s.Error(err)
}
