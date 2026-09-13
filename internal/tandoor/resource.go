package tandoor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ResourceService[T any] struct {
	client   *Client
	endpoint string
}

func (s *ResourceService[T]) Search(ctx context.Context, query string, kv ...string) ([]T, error) {
	params := constructParams(map[string]string{"query": query}, kv...)
	endpoint := constructURL(s.endpoint, params)

	body, err := s.client.Request(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Results []T `json:"results"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("invalid search response: %w", err)
	}

	return resp.Results, nil
}

func (s *ResourceService[T]) All(ctx context.Context, kv ...string) ([]T, error) {
	params := constructParams(map[string]string{"page_size": "100"}, kv...)
	endpoint := constructURL(s.endpoint, params)

	return RequestAllPages[T](ctx, s.client, endpoint)
}
