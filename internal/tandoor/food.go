package tandoor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type FoodService struct {
	client *Client
}

type Food struct {
	ID         int    `json:"id" jsonschema_description:"The food's unique identifier in Tandoor."`
	Name       string `json:"name" jsonschema_description:"The food's name."`
	PluralName string `json:"plural_name" jsonschema_description:"The plural form of the food's name, if one is set."`
}

type foodSearchResponse struct {
	Results []Food `json:"results"`
}

func (s *FoodService) SearchFoods(ctx context.Context, query string, kv ...string) ([]Food, error) {
	params := constructParams(map[string]string{"query": query}, kv...)
	endpoint := constructURL("food/", params)

	body, err := s.client.Request(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var resp foodSearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("invalid food search response: %w", err)
	}

	return resp.Results, nil
}
