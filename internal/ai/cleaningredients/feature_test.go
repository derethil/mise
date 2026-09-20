package cleaningredients

import (
	"os"
	"testing"

	miseai "github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/ai/providers"
	"github.com/derethil/mise/internal/ai/testutil"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/firebase/genkit/go/core/status"
	"github.com/firebase/genkit/go/genkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type FeatureSuite struct {
	suite.Suite
}

func TestFeatureSuite(t *testing.T) {
	suite.Run(t, new(FeatureSuite))
}

func (s *FeatureSuite) TestNormalizeRow_Header_DiscardsModelNoteAndKeepsOriginalText() {
	row := normalizeRow("For the sauce", CleanedRow{Kind: "header", Note: "model made this up", Amount: 3})

	s.Equal(CleanedRow{Kind: KindHeader, Note: "For the sauce"}, row)
}

func (s *FeatureSuite) TestNormalizeRow_Junk_DiscardsAllFields() {
	row := normalizeRow("some junk row", CleanedRow{Kind: "JUNK", Food: "onion", Amount: 1})

	s.Equal(CleanedRow{Kind: KindJunk}, row)
}

func (s *FeatureSuite) TestNormalizeRow_Ingredient_IsCaseInsensitiveAndPassesThrough() {
	row := normalizeRow("2 cups onion", CleanedRow{Kind: "Ingredient", Amount: 2, Unit: "cup", Food: "onion"})

	s.Equal(CleanedRow{Kind: KindIngredient, Amount: 2, Unit: "cup", Food: "onion"}, row)
}

func (s *FeatureSuite) TestNormalizeRow_UnrecognizedKindDefaultsToIngredient() {
	row := normalizeRow("2 cups onion", CleanedRow{Kind: "garbage", Food: "onion"})

	s.Equal(KindIngredient, row.Kind)
	s.Equal("onion", row.Food)
}

func (s *FeatureSuite) TestNormalizeRows_MapsEachRowInOrder() {
	batch := &CleanedRowBatch{Rows: []CleanedRow{
		{Kind: KindIngredient, Food: "onion"},
		{Kind: KindJunk},
	}}

	rows, err := normalizeRows(batch, []string{"2 cups onion", "some junk row"})

	s.Require().NoError(err)
	s.Equal([]CleanedRow{
		{Kind: KindIngredient, Food: "onion"},
		{Kind: KindJunk},
	}, rows)
}

func (s *FeatureSuite) TestNormalizeRows_NilBatchIsMalformed() {
	_, err := normalizeRows(nil, []string{"2 cups onion"})

	s.ErrorIs(err, miseai.ErrMalformedResponse)
}

func (s *FeatureSuite) TestNormalizeRows_CountMismatchIsMalformed() {
	batch := &CleanedRowBatch{Rows: []CleanedRow{{Kind: KindIngredient}}}

	_, err := normalizeRows(batch, []string{"2 cups onion", "1 egg"})

	s.ErrorIs(err, miseai.ErrMalformedResponse)
}

func TestCleanFlow_UsesFakeModelOutputForEachBatch(t *testing.T) {
	g := genkit.Init(t.Context(), genkit.WithPromptFS(os.DirFS("..")))
	testutil.DefineModel(g, "test/model",
		testutil.JSONResponse(CleanedRowBatch{Rows: []CleanedRow{
			{Kind: KindIngredient, Amount: 2, Unit: "cup", Food: "onion"},
			{Kind: KindHeader, Note: "ignored by normalization"},
		}}),
		testutil.JSONResponse(CleanedRowBatch{Rows: []CleanedRow{
			{Kind: KindJunk, Food: "ignored by normalization"},
		}}),
	)

	feature := &Feature{BatchSize: 2}
	require.NoError(t, feature.Register(miseai.Registry{
		Genkit: g,
		Model:  providers.ModelRef{Provider: "test", Name: "model"},
	}))

	got, err := feature.flow.Run(t.Context(), projectedRecipe{
		Ingredients: []string{"2 cups onion", "For the sauce", "Save this recipe"},
	})

	require.NoError(t, err)
	assert.Equal(t, []CleanedRow{
		{Kind: KindIngredient, Amount: 2, Unit: "cup", Food: "onion"},
		{Kind: KindHeader, Note: "For the sauce"},
		{Kind: KindJunk},
	}, got.Ingredients)
}

func TestCleanFlow_AnnotatesModelFailureWithBatchRange(t *testing.T) {
	g := genkit.Init(t.Context(), genkit.WithPromptFS(os.DirFS("..")))
	modelErr := status.Errorf(status.ErrInvalidArgument, "model unavailable")
	testutil.DefineModel(g, "test/model", testutil.ErrorResponse(modelErr), testutil.ErrorResponse(modelErr))

	feature := &Feature{}
	require.NoError(t, feature.Register(miseai.Registry{
		Genkit: g,
		Model:  providers.ModelRef{Provider: "test", Name: "model"},
	}))

	_, err := feature.flow.Run(t.Context(), projectedRecipe{Ingredients: []string{"1 onion"}})

	require.ErrorIs(t, err, status.ErrInvalidArgument)
	require.ErrorContains(t, err, "rows 0-0")
	require.ErrorContains(t, err, "cleanRowBatch")
}

func TestCleanFlow_RetriesAfterClassifiedModelFailure(t *testing.T) {
	g := genkit.Init(t.Context(), genkit.WithPromptFS(os.DirFS("..")))
	testutil.DefineModel(g, "test/model",
		testutil.ErrorResponse(status.Errorf(status.ErrInvalidArgument, "bad request")),
		testutil.JSONResponse(CleanedRowBatch{Rows: []CleanedRow{{Kind: KindIngredient, Food: "onion"}}}),
	)

	feature := &Feature{}
	require.NoError(t, feature.Register(miseai.Registry{
		Genkit: g,
		Model:  providers.ModelRef{Provider: "test", Name: "model"},
	}))

	got, err := feature.flow.Run(t.Context(), projectedRecipe{Ingredients: []string{"1 onion"}})

	require.NoError(t, err)
	assert.Equal(t, []CleanedRow{{Kind: KindIngredient, Food: "onion"}}, got.Ingredients)
}

func TestCleanFlow_MalformedBatchIncludesBatchRange(t *testing.T) {
	g := genkit.Init(t.Context(), genkit.WithPromptFS(os.DirFS("..")))
	testutil.DefineModel(g, "test/model", testutil.JSONResponse(CleanedRowBatch{
		Rows: []CleanedRow{{Kind: KindIngredient, Food: "onion"}},
	}))

	feature := &Feature{}
	require.NoError(t, feature.Register(miseai.Registry{
		Genkit: g,
		Model:  providers.ModelRef{Provider: "test", Name: "model"},
	}))

	_, err := feature.flow.Run(t.Context(), projectedRecipe{Ingredients: []string{"1 onion", "1 garlic clove"}})

	require.ErrorIs(t, err, miseai.ErrMalformedResponse)
	require.ErrorContains(t, err, "rows 0-1")
}

func TestCleanRecipe_EmptyRecipeSkipsModel(t *testing.T) {
	got, err := (&Feature{}).CleanRecipe(t.Context(), &tandoor.Recipe{ID: 1, Name: "Empty"}, nil)

	require.NoError(t, err)
	assert.NotNil(t, got.Ingredients)
	assert.Empty(t, got.Ingredients)
}
