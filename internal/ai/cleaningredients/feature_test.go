package cleaningredients

import (
	"testing"

	miseai "github.com/derethil/mise/internal/ai"
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
