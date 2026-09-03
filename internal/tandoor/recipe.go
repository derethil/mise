package tandoor

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tidwall/gjson"
)

type RecipeService struct {
	client *Client
}

type Recipe struct {
	ID int

	raw []byte
}

func (r *Recipe) JSON() []byte {
	return r.raw
}

func (s *RecipeService) Get(ctx context.Context, id int) (*Recipe, error) {
	body, err := s.client.Request(ctx, http.MethodGet, fmt.Sprintf("recipe/%d/", id), nil)
	if err != nil {
		return nil, err
	}

	return parseRecipe(body)
}

func (s *RecipeService) Update(ctx context.Context, id int, raw []byte) error {
	recipe, err := parseRecipe(raw)
	if err != nil {
		return fmt.Errorf("cannot update recipe %d: %w", id, err)
	}

	_, err = s.client.Request(ctx, http.MethodPut, fmt.Sprintf("recipe/%d/", id), recipe.JSON())
	return err
}

func parseRecipe(raw []byte) (*Recipe, error) {
	if !gjson.ValidBytes(raw) {
		return nil, fmt.Errorf("invalid recipe JSON")
	}

	id := gjson.GetBytes(raw, "id")
	if !id.Exists() {
		return nil, fmt.Errorf("recipe is missing an id")
	}

	return &Recipe{ID: int(id.Int()), raw: raw}, nil
}
