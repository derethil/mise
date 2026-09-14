package assignkeywords

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitSchema_Empty(t *testing.T) {
	assert.Nil(t, splitSchema(""))
	assert.Nil(t, splitSchema("   \n  \n"), "whitespace-only schema has nothing to assign")
}

func TestSplitSchema_NoHeadingsIsOneUntitledSection(t *testing.T) {
	sections := splitSchema("Assign whatever you think fits.")

	assert.Equal(t, []section{{Body: "Assign whatever you think fits."}}, sections)
}

func TestSplitSchema_SingleHeading(t *testing.T) {
	sections := splitSchema("## Cuisine\nAssign the broad cuisine.\n")

	assert.Equal(t, []section{{Title: "Cuisine", Body: "Assign the broad cuisine."}}, sections)
}

func TestSplitSchema_PreambleIsSharedAcrossEverySection(t *testing.T) {
	schema := "Shared intro.\n\n## A\nBody A\n\n## B\nBody B\n"

	sections := splitSchema(schema)

	assert.Equal(t, []section{
		{Title: "A", Body: "Shared intro.\n\nBody A"},
		{Title: "B", Body: "Shared intro.\n\nBody B"},
	}, sections)
}

func TestSplitSchema_NoPreambleLeavesBodyUntouched(t *testing.T) {
	sections := splitSchema("## A\nBody A\n")

	assert.Equal(t, "Body A", sections[0].Body)
}

func TestSplitSchema_SubheadingsStayInsideTheirSection(t *testing.T) {
	sections := splitSchema("## A\nBody\n### Not a new section\nMore body\n")

	assert.Len(t, sections, 1, "### is not the '## ' section marker")
	assert.Equal(t, "Body\n### Not a new section\nMore body", sections[0].Body)
}

func TestSectionLabel(t *testing.T) {
	assert.Equal(t, "keywords", section{}.label())
	assert.Equal(t, "Cuisine", section{Title: "Cuisine"}.label())
}
