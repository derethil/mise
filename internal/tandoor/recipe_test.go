package tandoor

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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

func (s *ClientSuite) TestRecipesGet() {
	recipe, err := s.client.Recipes.Get(s.T().Context(), 42)

	s.Require().NoError(err)
	s.Equal(42, recipe.ID)
	s.JSONEq(`{"id": 42, "name": "Tacos"}`, string(recipe.JSON()))
	s.Equal(http.MethodGet, s.lastMethod)
	s.Equal("/api/recipe/42/", s.lastPath)
	s.Equal("Bearer test-token", s.lastAuth)
}

func (s *ClientSuite) TestRecipesGet_HTTPError() {
	s.status = http.StatusForbidden

	_, err := s.client.Recipes.Get(s.T().Context(), 42)

	s.Error(err)
}

func (s *ClientSuite) TestRecipesGet_InvalidPayload() {
	s.response = map[string]any{"name": "Tacos"}

	_, err := s.client.Recipes.Get(s.T().Context(), 42)

	s.Error(err)
}

func (s *ClientSuite) TestRecipesGetAllRecipeIDs() {
	s.server.Close()
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.lastPath = r.URL.Path
		s.lastQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"id":1},{"id":2},{"id":3}],"next":null}`))
	}))
	s.client = NewClient(s.server.URL+"/api", "test-token")

	ids, err := s.client.Recipes.GetAllRecipeIDs(s.T().Context())

	s.Require().NoError(err)
	s.Equal([]int{1, 2, 3}, ids)
	s.Equal("/api/recipe/", s.lastPath)
	s.Equal("page_size=100", s.lastQuery)
}

func (s *ClientSuite) TestRecipesGetAllRecipeIDs_MultiplePages() {
	s.server.Close()

	var page2URL string
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("page") == "2" {
			_, _ = w.Write([]byte(`{"results":[{"id":3}],"next":null}`))
			return
		}
		_, _ = fmt.Fprintf(w, `{"results":[{"id":1},{"id":2}],"next":%q}`, page2URL)
	}))
	page2URL = s.server.URL + "/api/recipe/?page_size=100&page=2"
	s.client = NewClient(s.server.URL+"/api", "test-token")

	ids, err := s.client.Recipes.GetAllRecipeIDs(s.T().Context())

	s.Require().NoError(err)
	s.Equal([]int{1, 2, 3}, ids)
}

func (s *ClientSuite) TestRecipesGetAllRecipeIDs_HTTPError() {
	s.status = http.StatusForbidden

	_, err := s.client.Recipes.GetAllRecipeIDs(s.T().Context())

	s.Error(err)
}

func (s *ClientSuite) TestRecipesUpdate() {
	err := s.client.Recipes.Update(s.T().Context(), 42, []byte(`{"id": 42, "name": "Restored Tacos"}`))

	s.Require().NoError(err)
	s.Equal(http.MethodPut, s.lastMethod)
	s.Equal("/api/recipe/42/", s.lastPath)
}

func (s *ClientSuite) TestRecipesUpdate_InvalidJSON() {
	err := s.client.Recipes.Update(s.T().Context(), 42, []byte(`not json`))

	s.Error(err)
	s.Empty(s.lastMethod, "should not send a request")
}

func (s *ClientSuite) TestRecipesUpdate_HTTPError() {
	s.status = http.StatusForbidden

	err := s.client.Recipes.Update(s.T().Context(), 42, []byte(`{"id": 42, "name": "Tacos"}`))

	s.Error(err)
}
