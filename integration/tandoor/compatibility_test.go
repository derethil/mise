//go:build integration

package tandoor_test

import (
	"github.com/tidwall/gjson"

	"github.com/derethil/mise/internal/tandoor"
)

func (s *TandoorSuite) TestVersion() {
	version, err := s.client.Version(s.T().Context())
	s.Require().NoError(err)
	s.Equal(s.version, version, "the running server is not the requested image")
}

func (s *TandoorSuite) TestConnection() {
	s.Run("SucceedsWithGeneratedToken", func() {
		s.NoError(s.client.TestConnection(s.T().Context()))
	})

	s.Run("RejectsInvalidToken", func() {
		client := tandoor.NewClient(s.baseURL, "not-a-valid-token")
		s.ErrorIs(client.TestConnection(s.T().Context()), tandoor.ErrTandoorUnauthorized)
	})
}

func jsonStrings(raw []byte, path string) []string {
	results := gjson.GetBytes(raw, path).Array()

	values := make([]string, 0, len(results))
	for _, result := range results {
		values = append(values, result.String())
	}

	return values
}
