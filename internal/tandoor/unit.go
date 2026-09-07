package tandoor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type UnitService struct {
	client *Client
}

type Unit struct {
	ID         int    `json:"id" jsonschema_description:"The unit's unique identifier in Tandoor."`
	Name       string `json:"name" jsonschema_description:"The unit's name."`
	PluralName string `json:"plural_name" jsonschema_description:"The plural form of the unit's name, if one is set."`
}

type unitSearchResponse struct {
	Results []Unit `json:"results"`
}

func (s *UnitService) SearchUnits(ctx context.Context, query string, kv ...string) ([]Unit, error) {
	params := constructParams(map[string]string{"query": query}, kv...)
	endpoint := constructURL("unit/", params)

	body, err := s.client.Request(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var resp unitSearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("invalid unit search response: %w", err)
	}

	return resp.Results, nil
}
