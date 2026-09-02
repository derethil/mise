package tandoor

import (
	"net/http"
	"os"
	"path/filepath"
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

func (s *ClientSuite) TestRecipeBackup() {
	recipe := NewRecipe(s.client, 42)
	err := recipe.Load()
	s.Require().NoError(err)

	err = recipe.Backup()
	s.Require().NoError(err)

	data, err := os.ReadFile(filepath.Join(s.backupDir, "42.json"))
	s.Require().NoError(err)
	s.JSONEq(`{"id": 42, "name": "Tacos"}`, string(data))
}

func (s *ClientSuite) TestRecipeRestore() {
	err := os.WriteFile(filepath.Join(s.backupDir, "42.json"), []byte(`{"id": 42, "name": "Restored Tacos"}`), 0o644)
	s.Require().NoError(err)

	recipe := NewRecipe(s.client, 42)
	err = recipe.Restore()

	s.Require().NoError(err)
	s.Equal(http.MethodPut, s.lastMethod)
	s.Equal("/api/recipe/42/", s.lastPath)
}

func (s *ClientSuite) TestRecipeRestore_MissingBackupFile() {
	recipe := NewRecipe(s.client, 999)

	err := recipe.Restore()

	s.Error(err)
}

func (s *ClientSuite) TestRecipeRestore_InvalidBackupFile() {
	err := os.WriteFile(filepath.Join(s.backupDir, "999.json"), []byte(`not json`), 0o644)
	s.Require().NoError(err)

	recipe := NewRecipe(s.client, 999)
	err = recipe.Restore()

	s.Error(err)
}
