package cleaningredients

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ReportSuite struct {
	suite.Suite
}

func TestReportSuite(t *testing.T) {
	suite.Run(t, new(ReportSuite))
}

func (s *ReportSuite) TestString_Junk() {
	c := Change{Before: "some junk row", Row: CleanedRow{Kind: KindJunk}}

	s.Equal("  [drop]   some junk row", c.String())
}

func (s *ReportSuite) TestString_Header() {
	c := Change{Before: "For the sauce", Row: CleanedRow{Kind: KindHeader, Note: "For the sauce"}}

	s.Equal("  [header] For the sauce", c.String())
}

func (s *ReportSuite) TestString_Ingredient() {
	c := Change{
		Before: "2 cups chopped onion",
		Row:    CleanedRow{Kind: KindIngredient, Amount: 2, Unit: "cup", Food: "onion", Note: "chopped"},
	}

	s.Equal("  2 cups chopped onion -> amount: 2 | unit: cup | food: onion | note: chopped", c.String())
}

func (s *ReportSuite) TestFormatIngredient_OmitsZeroAndEmptyFields() {
	row := CleanedRow{Kind: KindIngredient, Food: "salt"}

	s.Equal("food: salt", formatIngredient(row))
}
