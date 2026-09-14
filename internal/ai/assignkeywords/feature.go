// Package assignkeywords implements the miseai.Feature for assigning keywords to a recipe
// according to a keyword schema supplied by the user.
package assignkeywords

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	miseai "github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

func init() {
	miseai.RegisterFeature(func() miseai.Feature { return &Feature{} })
}

const (
	assignMaxAttempts = 2
	assignMaxTurns    = 10

	maxVocabulary = 250 // Fall back to searchKeywords tool if the vocabulary is larger than this.

	maxPerCategory = 4
	maxKeywords    = 10
)

var assignKeywordsConfig = miseai.GenerateConfig{
	Temperature: new(0.2),
	Reasoning:   new(false),
}

type AssignKeywordsInput struct {
	Category           string   `json:"category" jsonschema_description:"The name of the single schema category being assigned on this call."`
	Schema             string   `json:"schema" jsonschema_description:"The part of the user's keyword schema covering this category, describing exactly which keywords to assign."`
	Vocabulary         []string `json:"vocabulary" jsonschema_description:"Keywords that already exist in this Tandoor instance."`
	VocabularyComplete bool     `json:"vocabulary_complete" jsonschema_description:"Whether the vocabulary lists every keyword in the instance, or only a sample."`

	Name            string   `json:"name" jsonschema_description:"The recipe's title."`
	Description     string   `json:"description" jsonschema_description:"The recipe's description, often empty."`
	CurrentKeywords []string `json:"current_keywords" jsonschema_description:"Keywords already assigned to this recipe."`
	Ingredients     []string `json:"ingredients" jsonschema_description:"The recipe's ingredients, by name."`
	Steps           []string `json:"steps" jsonschema_description:"The recipe's instructions, in order."`
}

type AssignedKeyword struct {
	Category string `json:"-"`

	Reason  string `json:"reason" jsonschema_description:"One short sentence working out whether this keyword applies to this recipe, citing the ingredient, technique, or step that decides it. Write this before deciding anything else."`
	Applies bool   `json:"applies" jsonschema_description:"Whether the reason above concluded that the keyword does apply. False discards the entry, so use it whenever working through the reason ruled the keyword out."`
	Name    string `json:"name" jsonschema_description:"The keyword to assign, exactly as it should appear in Tandoor. When an existing keyword covers this, copy its name verbatim, including its capitalization."`
}

type AssignedKeywords []AssignedKeyword

type Feature struct {
	Schema string

	tandoor *tandoor.Client

	opts   []ai.PromptExecuteOption
	prompt *ai.DataPrompt[AssignKeywordsInput, AssignedKeywords]
	flow   *core.Flow[AssignKeywordsInput, AssignedKeywords, struct{}]
}

type Progress struct {
	Status    string
	Completed int
	Total     int
}

type ProgressFunc func(Progress) error

type progressKey struct{}

func withProgress(ctx context.Context, fn ProgressFunc) context.Context {
	if fn == nil {
		return ctx
	}

	return context.WithValue(ctx, progressKey{}, fn)
}

func reportProgress(ctx context.Context, status string, completed, total int) error {
	fn, ok := ctx.Value(progressKey{}).(ProgressFunc)
	if !ok {
		return nil
	}

	return fn(Progress{Status: status, Completed: completed, Total: total})
}

func (f *Feature) Register(r miseai.Registry) error {
	genkit.DefineSchemaFor[AssignKeywordsInput](r.Genkit)
	genkit.DefineSchemaFor[AssignedKeyword](r.Genkit)
	genkit.DefineSchemaFor[AssignedKeywords](r.Genkit)

	f.prompt = genkit.LookupDataPrompt[AssignKeywordsInput, AssignedKeywords](r.Genkit, "assign_keywords")
	if f.prompt == nil {
		return fmt.Errorf("%w: assign_keywords", miseai.ErrPromptNotFound)
	}

	f.tandoor = r.Tandoor

	f.opts = r.PromptOptions(assignKeywordsConfig,
		ai.WithTools(miseai.SearchKeywordsTool(r)),
	)

	f.flow = genkit.DefineFlow(r.Genkit, "assignRecipeKeywords", f.assign)

	return nil
}

func (f *Feature) AssignKeywords(ctx context.Context, recipe *tandoor.Recipe, onProgress ProgressFunc) (AssignedKeywords, []tandoor.Keyword, error) {
	ctx = withProgress(ctx, onProgress)

	sections := splitSchema(f.Schema)
	if len(sections) == 0 {
		return nil, nil, fmt.Errorf("%w: the keyword schema is empty", config.ErrInvalidConfig)
	}

	if err := reportProgress(ctx, "reading keywords", 0, len(sections)); err != nil {
		return nil, nil, err
	}

	vocabulary, err := f.loadVocabulary(ctx)
	if err != nil {
		return nil, nil, err
	}

	projected := projectRecipe(recipe.JSON())

	assigned, err := f.assignSections(ctx, sections, vocabulary, projected)
	if err != nil {
		return nil, nil, err
	}

	assigned = dedupe(assigned)

	if len(assigned) == 0 {
		slog.WarnContext(ctx,
			fmt.Sprintf("no keywords were assigned to recipe %d", recipe.ID),
			slog.Int("recipe_id", recipe.ID),
			slog.String("recipe_name", recipe.Name),
		)
	}

	return assigned, vocabulary, nil
}

func (f *Feature) assignSections(ctx context.Context, sections []section, vocabulary []tandoor.Keyword, projected projectedRecipe) (AssignedKeywords, error) {
	var assigned AssignedKeywords

	for i, sec := range sections {
		found, err := f.assignSection(ctx, sec, vocabulary, projected)
		if err != nil {
			return nil, err
		}
		assigned = append(assigned, found...)

		if err := reportProgress(ctx, "assigning keywords", i+1, len(sections)); err != nil {
			return nil, err
		}
	}

	return assigned, nil
}

func (f *Feature) assignSection(ctx context.Context, sec section, vocabulary []tandoor.Keyword, projected projectedRecipe) (AssignedKeywords, error) {
	input := sectionInput(sec, vocabulary, projected)

	found, err := f.flow.Run(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", sec.label(), err)
	}

	for i := range found {
		found[i].Category = sec.Title
	}

	return found, nil
}

func sectionInput(sec section, vocabulary []tandoor.Keyword, projected projectedRecipe) AssignKeywordsInput {
	input := AssignKeywordsInput{
		Category:           sec.Title,
		Schema:             sec.Body,
		Vocabulary:         names(vocabulary),
		VocabularyComplete: len(vocabulary) <= maxVocabulary,
		Name:               projected.Name,
		Description:        projected.Description,
		CurrentKeywords:    projected.CurrentKeywords,
		Ingredients:        projected.Ingredients,
		Steps:              projected.Steps,
	}
	if !input.VocabularyComplete {
		input.Vocabulary = input.Vocabulary[:maxVocabulary]
	}

	return input
}

func (f *Feature) assign(ctx context.Context, input AssignKeywordsInput) (AssignedKeywords, error) {
	opts := f.opts
	if !input.VocabularyComplete {
		opts = append(f.opts, ai.WithMaxTurns(assignMaxTurns))
	}

	var assigned AssignedKeywords
	var err error

	for attempt := 1; attempt <= assignMaxAttempts; attempt++ {
		assigned, _, err = f.prompt.Execute(ctx, input, opts...)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("assignKeywords: %w", err)
	}

	return normalize(assigned)
}

func (f *Feature) loadVocabulary(ctx context.Context) ([]tandoor.Keyword, error) {
	vocabulary, err := f.tandoor.Keywords.ListKeywords(ctx)
	if err != nil {
		return nil, err
	}

	if vocabulary == nil {
		vocabulary = []tandoor.Keyword{}
	}

	return vocabulary, nil
}

func normalize(assigned AssignedKeywords) (AssignedKeywords, error) {
	seen := map[string]bool{}
	normalized := make(AssignedKeywords, 0, len(assigned))

	for _, keyword := range assigned {
		keyword.Name = cleaned(keyword.Name)
		keyword.Category = cleaned(keyword.Category)
		keyword.Reason = cleaned(keyword.Reason)

		if !keyword.Applies || keyword.Name == "" || seen[strings.ToLower(keyword.Name)] {
			continue
		}

		seen[strings.ToLower(keyword.Name)] = true
		normalized = append(normalized, keyword)

		if len(normalized) == maxPerCategory {
			break
		}
	}

	return normalized, nil
}

func dedupe(assigned AssignedKeywords) AssignedKeywords {
	seen := map[string]bool{}
	out := make(AssignedKeywords, 0, len(assigned))

	for _, keyword := range assigned {
		key := strings.ToLower(keyword.Name)
		if seen[key] {
			continue
		}

		seen[key] = true
		out = append(out, keyword)

		if len(out) == maxKeywords {
			break
		}
	}

	return out
}

func names(keywords []tandoor.Keyword) []string {
	out := make([]string, 0, len(keywords))
	for _, keyword := range keywords {
		out = append(out, keyword.Name)
	}

	return out
}
