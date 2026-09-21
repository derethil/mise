//go:build integration

package tandoor_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func (s *TandoorSuite) request(ctx context.Context, method, path string, body, into any) {
	var payload io.Reader

	if body != nil {
		encoded, err := json.Marshal(body)
		s.Require().NoError(err, "encode the %s %s body", method, path)

		payload = bytes.NewReader(encoded)
	}

	request, err := http.NewRequestWithContext(ctx, method, s.baseURL+path, payload)
	s.Require().NoError(err, "build %s %s", method, path)

	request.Header.Set("Content-Type", "application/json")

	if s.token != "" {
		request.Header.Set("Authorization", "Bearer "+s.token)
	}

	response, err := http.DefaultClient.Do(request)
	s.Require().NoError(err, "%s %s", method, path)
	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	s.Require().NoError(err, "read the %s %s response", method, path)

	s.Require().Less(
		response.StatusCode, http.StatusMultipleChoices,
		"%s %s returned %d: %s", method, path, response.StatusCode, content,
	)

	if into != nil {
		s.Require().NoError(json.Unmarshal(content, into), "decode the %s %s response: %s", method, path, content)
	}
}
