package tandoor

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func (s *ClientSuite) TestVersionExtractsParenthesizedVersion() {
	s.response = map[string]any{"info": map[string]any{"version": "0.0.0 (2.6.13)"}}

	version, err := s.client.Version(s.T().Context())

	s.Require().NoError(err)
	s.Equal("2.6.13", version)
	s.Equal("/openapi/", s.lastPath)
	s.Equal("format=json", s.lastQuery)
}

func (s *ClientSuite) TestVersionFallsBackToRawStringWithoutParens() {
	s.response = map[string]any{"info": map[string]any{"version": "1.2.3"}}

	version, err := s.client.Version(s.T().Context())

	s.Require().NoError(err)
	s.Equal("1.2.3", version)
}

func (s *ClientSuite) TestVersion_MissingInfoVersion() {
	s.response = map[string]any{"info": map[string]any{}}

	_, err := s.client.Version(s.T().Context())

	s.Error(err)
}

func (s *ClientSuite) TestVersion_InvalidJSON() {
	s.server.Close()
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not json"))
	}))
	s.client = NewClient(s.server.URL, "test-token")

	_, err := s.client.Version(s.T().Context())

	s.Error(err)
}

func (s *ClientSuite) TestVersion_HTTPError() {
	s.status = http.StatusForbidden

	_, err := s.client.Version(s.T().Context())

	s.Error(err)
}

func (s *ClientSuite) TestVersionSupported_WithinRange() {
	s.response = map[string]any{"info": map[string]any{"version": fmt.Sprintf("0.0.0 (%s)", MinSupportedVersion)}}

	version, supported, err := s.client.VersionSupported(s.T().Context())

	s.Require().NoError(err)
	s.True(supported)
	s.Equal(MinSupportedVersion, version)
}

func (s *ClientSuite) TestVersionSupported_BelowMin() {
	s.response = map[string]any{"info": map[string]any{"version": "0.0.0 (1.0.0)"}}

	_, supported, err := s.client.VersionSupported(s.T().Context())

	s.Require().NoError(err)
	s.False(supported)
}

func (s *ClientSuite) TestVersionSupported_AboveMax() {
	s.response = map[string]any{"info": map[string]any{"version": "0.0.0 (99.0.0)"}}

	_, supported, err := s.client.VersionSupported(s.T().Context())

	s.Require().NoError(err)
	s.False(supported)
}

func (s *ClientSuite) TestVersionSupported_PropagatesError() {
	s.server.Close()

	_, supported, err := s.client.VersionSupported(s.T().Context())

	s.Error(err)
	s.False(supported)
}

func (s *ClientSuite) TestCompareVersions() {
	cases := []struct {
		a, b string
		want int
	}{
		{"2.6.13", "2.6.13", 0},
		{"2.6.12", "2.6.13", -1},
		{"2.6.13", "2.6.12", 1},
		{"2.7.0", "2.6.13", 1},
		{"1.9.9", "2.0.0", -1},
		{"2.6", "2.6.0", 0},
		{"2.6.1", "2.6", 1},
	}

	for _, c := range cases {
		s.Run(fmt.Sprintf("%s_vs_%s", c.a, c.b), func() {
			s.Equal(c.want, compareVersions(c.a, c.b))
		})
	}
}
