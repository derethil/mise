package tandoor

import "context"

type FoodService struct{ ResourceService[Food] }

type Food struct {
	ID         int    `json:"id" jsonschema_description:"The food's unique identifier in Tandoor."`
	Name       string `json:"name" jsonschema_description:"The food's name."`
	PluralName string `json:"plural_name" jsonschema_description:"The plural form of the food's name, if one is set."`
}

func (s *FoodService) SearchFoods(ctx context.Context, query string, kv ...string) ([]Food, error) {
	return s.Search(ctx, query, kv...)
}
