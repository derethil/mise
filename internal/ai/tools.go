package ai

import (
	"github.com/derethil/mise/internal/tandoor"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type GetTandoorFoodsInput struct {
	Search string `json:"search" jsonschema_description:"The search term to use when searching for foods in Tandoor."`
}

func (c *Client) TandoorGetFoodTool(t *tandoor.Client) ai.Tool {
	return genkit.DefineTool(c.g, "getFoods",
		"Searches for existing foods in Tandoor.",
		func(ctx *ai.ToolContext, input GetTandoorFoodsInput) ([]tandoor.Food, error) {
			return t.Foods.SearchFoods(ctx, input.Search)
		},
	)
}

type GetTandoorUnitsInput struct {
	Search string `json:"search" jsonschema_description:"The search term to use when searching for units in Tandoor."`
}

func (c *Client) TandoorGetUnitTool(t *tandoor.Client) ai.Tool {
	return genkit.DefineTool(c.g, "getUnits",
		"Searches for existing units in Tandoor.",
		func(ctx *ai.ToolContext, input GetTandoorUnitsInput) ([]tandoor.Unit, error) {
			return t.Units.SearchUnits(ctx, input.Search)
		},
	)
}
