package assignkeywords

import (
	"testing"

	"github.com/derethil/mise/internal/tandoor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalize_DropsEntriesThatDoNotApply(t *testing.T) {
	out, err := normalize(AssignedKeywords{
		{Name: "Vegetarian", Applies: true},
		{Name: "Vegan", Applies: false},
	})

	require.NoError(t, err)
	assert.Equal(t, AssignedKeywords{{Name: "Vegetarian", Applies: true}}, out)
}

func TestNormalize_DropsEmptyNames(t *testing.T) {
	out, err := normalize(AssignedKeywords{{Name: "  ", Applies: true}})

	require.NoError(t, err)
	assert.Empty(t, out)
}

func TestNormalize_DedupesCaseInsensitivelyKeepingFirst(t *testing.T) {
	out, err := normalize(AssignedKeywords{
		{Name: "Dinner", Applies: true, Reason: "first"},
		{Name: "dinner", Applies: true, Reason: "second"},
	})

	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, "first", out[0].Reason)
}

func TestNormalize_TrimsAndStripsFormatCharactersFromFields(t *testing.T) {
	out, err := normalize(AssignedKeywords{{Name: "  Dinner  ", Reason: " because ", Applies: true}})

	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, "Dinner", out[0].Name)
	assert.Equal(t, "because", out[0].Reason)
}

func TestNormalize_CapsAtMaxPerCategory(t *testing.T) {
	var assigned AssignedKeywords
	for i := range maxPerCategory + 2 {
		assigned = append(assigned, AssignedKeyword{Name: string(rune('A' + i)), Applies: true})
	}

	out, err := normalize(assigned)

	require.NoError(t, err)
	assert.Len(t, out, maxPerCategory)
	assert.Equal(t, "A", out[0].Name, "earlier entries are kept over later ones")
}

func TestDedupe_DedupesAcrossCategoriesCaseInsensitively(t *testing.T) {
	out := dedupe(AssignedKeywords{
		{Name: "Dinner", Category: "Meal"},
		{Name: "dinner", Category: "Other"},
		{Name: "Quick", Category: "Time"},
	})

	assert.Equal(t, AssignedKeywords{
		{Name: "Dinner", Category: "Meal"},
		{Name: "Quick", Category: "Time"},
	}, out)
}

func TestDedupe_CapsAtMaxKeywords(t *testing.T) {
	var assigned AssignedKeywords
	for i := range maxKeywords + 2 {
		assigned = append(assigned, AssignedKeyword{Name: string(rune('A' + i))})
	}

	out := dedupe(assigned)

	assert.Len(t, out, maxKeywords)
}

func TestNames(t *testing.T) {
	out := names([]tandoor.Keyword{{ID: 1, Name: "Dinner"}, {ID: 2, Name: "Quick"}})

	assert.Equal(t, []string{"Dinner", "Quick"}, out)
}

func TestNames_Empty(t *testing.T) {
	out := names(nil)

	assert.NotNil(t, out, "must serialize as [] rather than null for genkit's schema validation")
	assert.Empty(t, out)
}
