package ai

import (
	"github.com/derethil/mise/internal/tandoor"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type GetTandoorFoodsInput struct {
	Search string `json:"search" jsonschema_description:"The search term to use when searching for foods in Tandoor. Ranked by fuzzy similarity, so pass the food name as parsed - no need to strip parenthetical, brand, or regional/varietal qualifiers. Results are ranked candidates, not guaranteed matches - use the top one whenever it's the same real-world food (even if less precise or differently worded than your own parse), to avoid creating a duplicate."`
}

func SearchFoodsTool(r Registry) ai.Tool {
	return genkit.DefineTool(r.Genkit, "searchFoods",
		"Searches for existing foods in Tandoor.",
		func(ctx *ai.ToolContext, input GetTandoorFoodsInput) ([]tandoor.Food, error) {
			return r.Tandoor.Foods.SearchFoods(ctx,
				input.Search,
				"page_size", "5",
				"simple", "true",
			)
		},
	)
}

type GetTandoorUnitsInput struct {
	Search string `json:"search" jsonschema_description:"The search term to use when searching for units in Tandoor"`
}

func SearchUnitsTool(r Registry) ai.Tool {
	return genkit.DefineTool(r.Genkit, "searchUnits",
		"Searches for existing units in Tandoor.",
		func(ctx *ai.ToolContext, input GetTandoorUnitsInput) ([]tandoor.Unit, error) {
			return r.Tandoor.Units.SearchUnits(ctx,
				input.Search,
				"page_size", "5",
			)
		},
	)
}

type SearchKeywordsInput struct {
	Search string `json:"search" jsonschema_description:"The search term to use when searching for keywords in Tandoor. Ranked by fuzzy similarity, so pass the keyword as you would write it. Results are ranked candidates, not guaranteed matches - reuse the top one whenever it means the same thing as the keyword you were about to create, to avoid ending up with two keywords that differ only in wording."`
}

func SearchKeywordsTool(r Registry) ai.Tool {
	return genkit.DefineTool(r.Genkit, "searchKeywords",
		"Searches for existing keywords in Tandoor.",
		func(ctx *ai.ToolContext, input SearchKeywordsInput) ([]tandoor.Keyword, error) {
			return r.Tandoor.Keywords.SearchKeywords(ctx,
				input.Search,
				"page_size", "5",
			)
		},
	)
}
