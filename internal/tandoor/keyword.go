package tandoor

import "context"

type KeywordService struct{ ResourceService[Keyword] }

type Keyword struct {
	ID   int    `json:"id" jsonschema_description:"The keyword's unique identifier in Tandoor."`
	Name string `json:"name" jsonschema_description:"The keyword's name."`
}

func (s *KeywordService) SearchKeywords(ctx context.Context, query string, kv ...string) ([]Keyword, error) {
	return s.Search(ctx, query, kv...)
}
