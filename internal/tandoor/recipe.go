package tandoor

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tidwall/gjson"
)

type Recipe struct {
	ID int `json:"id"`

	raw    []byte
	loaded bool
	client *Client
}

func NewRecipe(client *Client, id int) Recipe {
	return Recipe{
		ID:     id,
		raw:    nil,
		loaded: false,
		client: client,
	}
}

func (r *Recipe) Load(ctx context.Context) error {
	body, err := r.client.Request(ctx, http.MethodGet, fmt.Sprintf("recipe/%d/", r.ID), nil)
	if err != nil {
		return err
	}

	parsed, err := parseRecipe(body)
	if err != nil {
		return err
	}

	r.ID = parsed.ID
	r.raw = parsed.raw
	r.loaded = true

	return nil
}

func (r Recipe) JSON() []byte {
	return r.raw
}

func (r Recipe) Update(ctx context.Context, raw []byte) error {
	parsed, err := parseRecipe(raw)
	if err != nil {
		return fmt.Errorf("cannot update recipe %d: %w", r.ID, err)
	}

	_, err = r.client.Request(ctx, http.MethodPut, fmt.Sprintf("recipe/%d/", r.ID), parsed.JSON())
	return err
}

func parseRecipe(raw []byte) (Recipe, error) {
	if !gjson.ValidBytes(raw) {
		return Recipe{}, fmt.Errorf("invalid recipe JSON")
	}

	id := gjson.GetBytes(raw, "id")
	if !id.Exists() {
		return Recipe{}, fmt.Errorf("recipe is missing an id")
	}

	return Recipe{ID: int(id.Int()), raw: raw}, nil
}
