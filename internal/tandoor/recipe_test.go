package tandoor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRecipe(t *testing.T) {
	t.Run("valid recipe", func(t *testing.T) {
		raw := []byte(`{"id": 42, "name": "Tacos"}`)

		recipe, err := ParseRecipe(raw)

		require.NoError(t, err)
		assert.Equal(t, 42, recipe.ID)
		assert.JSONEq(t, string(raw), string(recipe.JSON()))
	})

	t.Run("missing id", func(t *testing.T) {
		_, err := ParseRecipe([]byte(`{"name": "Tacos"}`))

		assert.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := ParseRecipe([]byte(`not json`))

		assert.Error(t, err)
	})
}
