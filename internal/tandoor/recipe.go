package tandoor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tidwall/gjson"
)

type RecipeService struct {
	client *Client
}

type Recipe struct {
	ID   int
	Name string

	raw []byte
}

func (r *Recipe) JSON() []byte {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, r.raw, "", "    "); err != nil {
		return r.raw
	}
	return pretty.Bytes()
}

func (s *RecipeService) Get(ctx context.Context, id int) (*Recipe, error) {
	body, err := s.client.Request(ctx, http.MethodGet, fmt.Sprintf("recipe/%d/", id), nil)
	if err != nil {
		return nil, err
	}

	return parseRecipe(body)
}

func (s *RecipeService) GetAllRecipeIDs(ctx context.Context) ([]int, error) {
	type recipeStub struct {
		ID int `json:"id"`
	}

	endpoint := constructURL("recipe/", map[string]string{"page_size": "100"})
	stubs, err := RequestAllPages[recipeStub](ctx, s.client, endpoint)
	if err != nil {
		return nil, err
	}

	ids := make([]int, len(stubs))
	for i, stub := range stubs {
		ids[i] = stub.ID
	}

	return ids, nil
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

	name := gjson.GetBytes(raw, "name")

	return &Recipe{ID: int(id.Int()), Name: name.String(), raw: raw}, nil
}
