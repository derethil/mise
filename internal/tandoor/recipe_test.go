package tandoor

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRecipe(t *testing.T) {
	t.Run("valid recipe", func(t *testing.T) {
		raw := []byte(`{"id": 42, "name": "Tacos"}`)

		recipe, err := parseRecipe(raw)

		require.NoError(t, err)
		assert.Equal(t, 42, recipe.ID)
		assert.JSONEq(t, string(raw), string(recipe.JSON()))
	})

	t.Run("missing id", func(t *testing.T) {
		_, err := parseRecipe([]byte(`{"name": "Tacos"}`))

		assert.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := parseRecipe([]byte(`not json`))

		assert.Error(t, err)
	})
}

func (s *ClientSuite) TestRecipeLoad() {
	recipe := NewRecipe(s.client, 42)
	err := recipe.Load()

	s.Require().NoError(err)
	s.Equal(42, recipe.ID)
	s.Equal(http.MethodGet, s.lastMethod)
	s.Equal("/api/recipe/42/", s.lastPath)
	s.Equal("Bearer test-token", s.lastAuth)
}

func (s *ClientSuite) TestRecipeLoad_HTTPError() {
	s.status = http.StatusForbidden

	recipe := NewRecipe(s.client, 42)
	err := recipe.Load()

	s.Error(err)
}

func (s *ClientSuite) TestRecipeUpdate() {
	recipe := NewRecipe(s.client, 42)

	err := recipe.Update([]byte(`{"id": 42, "name": "Restored Tacos"}`))

	s.Require().NoError(err)
	s.Equal(http.MethodPut, s.lastMethod)
	s.Equal("/api/recipe/42/", s.lastPath)
}

func (s *ClientSuite) TestRecipeUpdate_InvalidJSON() {
	recipe := NewRecipe(s.client, 42)

	err := recipe.Update([]byte(`not json`))

	s.Error(err)
	s.Empty(s.lastMethod, "should not send a request")
}

func (s *ClientSuite) TestRecipeUpdate_HTTPError() {
	s.status = http.StatusForbidden

	recipe := NewRecipe(s.client, 42)
	err := recipe.Update([]byte(`{"id": 42, "name": "Tacos"}`))

	s.Error(err)
}
