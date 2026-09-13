package tandoor

import "net/http"

func (s *ClientSuite) TestKeywordsSearch() {
	s.response = map[string]any{
		"results": []map[string]any{
			{"id": 1, "name": "quick"},
		},
	}

	keywords, err := s.client.Keywords.SearchKeywords(s.T().Context(), "quick", "page_size", "200")

	s.Require().NoError(err)
	s.Equal([]Keyword{{ID: 1, Name: "quick"}}, keywords)
	s.Equal(http.MethodGet, s.lastMethod)
	s.Equal("/api/keyword/", s.lastPath)
	s.Equal("page_size=200&query=quick", s.lastQuery)
}

func (s *ClientSuite) TestKeywordsSearch_HTTPError() {
	s.status = http.StatusForbidden

	_, err := s.client.Keywords.SearchKeywords(s.T().Context(), "quick")

	s.Error(err)
}

func (s *ClientSuite) TestKeywordsSearch_InvalidPayload() {
	s.response = map[string]any{"results": "not an array"}

	_, err := s.client.Keywords.SearchKeywords(s.T().Context(), "quick")

	s.Error(err)
}
