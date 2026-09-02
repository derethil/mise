package tandoor

import (
	"fmt"

	"github.com/tidwall/gjson"
)

type Recipe struct {
	ID int `json:"id"`

	raw []byte
}

func ParseRecipe(raw []byte) (Recipe, error) {
	if !gjson.ValidBytes(raw) {
		return Recipe{}, fmt.Errorf("invalid recipe JSON")
	}

	id := gjson.GetBytes(raw, "id")
	if !id.Exists() {
		return Recipe{}, fmt.Errorf("recipe is missing an id")
	}

	return Recipe{ID: int(id.Int()), raw: raw}, nil
}

func (r Recipe) JSON() []byte {
	return r.raw
}
