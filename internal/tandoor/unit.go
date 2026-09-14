package tandoor

import "context"

type UnitService struct{ ResourceService[Unit] }

type Unit struct {
	ID         int    `json:"id" jsonschema_description:"The unit's unique identifier in Tandoor."`
	Name       string `json:"name" jsonschema_description:"The unit's name."`
	PluralName string `json:"plural_name" jsonschema_description:"The plural form of the unit's name, if one is set."`
}

func (s *UnitService) SearchUnits(ctx context.Context, query string, kv ...string) ([]Unit, error) {
	return s.Search(ctx, query, kv...)
}
