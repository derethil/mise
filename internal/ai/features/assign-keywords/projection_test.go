package assignkeywords

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProjectRecipe_Basic(t *testing.T) {
	raw := []byte(`{
		"name": "Tacos",
		"description": "Weeknight tacos",
		"keywords": [{"name": "Dinner"}, {"name": "Quick"}],
		"steps": [
			{"instruction": "Cook the beef.", "ingredients": [
				{"food": {"name": "Beef"}},
				{"is_header": true, "food": {"name": "For the tacos"}}
			]},
			{"instruction": "Warm the tortillas.", "ingredients": [
				{"food": {"name": "beef"}},
				{"food": {"name": "Tortillas"}}
			]}
		]
	}`)

	projected := projectRecipe(raw)

	assert.Equal(t, "Tacos", projected.Name)
	assert.Equal(t, "Weeknight tacos", projected.Description)
	assert.Equal(t, []string{"Dinner", "Quick"}, projected.CurrentKeywords)
	assert.Equal(t, []string{"Beef", "Tortillas"}, projected.Ingredients, "header rows and case-insensitive dupes are dropped")
	assert.Equal(t, []string{"Cook the beef.", "Warm the tortillas."}, projected.Steps)
}

func TestProjectRecipe_MissingFieldsStillProducesEmptySlicesNotNil(t *testing.T) {
	projected := projectRecipe([]byte(`{}`))

	assert.Empty(t, projected.CurrentKeywords)
	assert.Empty(t, projected.Ingredients)
	assert.Empty(t, projected.Steps)
	assert.NotNil(t, projected.CurrentKeywords, "must serialize as [] rather than null for genkit's schema validation")
	assert.NotNil(t, projected.Ingredients)
	assert.NotNil(t, projected.Steps)
}

func TestProjectRecipe_StepsRespectCombinedCharBudget(t *testing.T) {
	first := strings.Repeat("a", maxStepChars-10)
	second := strings.Repeat("b", 100)

	raw := []byte(`{"steps": [
		{"instruction": "` + first + `"},
		{"instruction": "` + second + `"}
	]}`)

	projected := projectRecipe(raw)

	assert.Len(t, projected.Steps, 2)
	assert.Equal(t, first, projected.Steps[0])
	assert.Equal(t, 10, len(projected.Steps[1])-len("…"), "second step is truncated to what's left of the budget")
}

func TestCleaned_StripsFormatCharactersAndTrims(t *testing.T) {
	assert.Equal(t, "onion", cleaned("  \u200bonion\u200b  "), "zero-width space is a Unicode format character")
	assert.Equal(t, "", cleaned("   "))
}

func TestTruncate_LeavesShortTextUnchanged(t *testing.T) {
	assert.Equal(t, "short", truncate("short", 100))
}

func TestTruncate_CutsAtRuneBoundaryNotMidRune(t *testing.T) {
	text := "a" + "é" // 'a' (1 byte) + 'é' (2 bytes, U+00E9)

	// limit=2 lands on the second (continuation) byte of 'é'; truncate must
	// back up to the start of the previous rune instead of splitting it.
	assert.Equal(t, "a…", truncate(text, 2))
}
