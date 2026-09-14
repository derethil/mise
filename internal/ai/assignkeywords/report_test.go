package assignkeywords

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangeString_NameOnly(t *testing.T) {
	c := Change{Name: "Dinner", Action: ActionKept}

	assert.Equal(t, "  [kept  ] Dinner", c.String())
}

func TestChangeString_WithCategory(t *testing.T) {
	c := Change{Name: "Vegetarian", Category: "Diet", Action: ActionNew}

	assert.Equal(t, "  [new   ] Vegetarian (Diet)", c.String())
}

func TestChangeString_WithCategoryAndReason(t *testing.T) {
	c := Change{Name: "Vegetarian", Category: "Diet", Reason: "no meat found", Action: ActionNew}

	assert.Equal(t, "  [new   ] Vegetarian (Diet) — no meat found", c.String())
}

func TestChangeString_RemovedHasNoCategoryOrReason(t *testing.T) {
	c := Change{Name: "Quick", Action: ActionRemoved}

	assert.Equal(t, "  [remove] Quick", c.String())
}
