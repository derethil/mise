package recipe

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/derethil/mise/internal/backup"
	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/stretchr/testify/suite"
)

type NormalizeSuite struct {
	suite.Suite

	server *httptest.Server
	client *tandoor.Client

	lastMethod string
	lastPath   string
	response   map[string]any
}

func (s *NormalizeSuite) SetupTest() {
	s.lastMethod = ""
	s.lastPath = ""
	s.response = map[string]any{"id": 42, "name": "Tacos"}

	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.lastMethod = r.Method
		s.lastPath = r.URL.Path

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(s.response)
	}))

	s.client = tandoor.NewClient(s.server.URL, "test-token")
}

func (s *NormalizeSuite) TearDownTest() {
	s.server.Close()
}

func TestNormalizeSuite(t *testing.T) {
	suite.Run(t, new(NormalizeSuite))
}

func (s *NormalizeSuite) TestNormalizeCmd_RequiresIDOrAll() {
	err := normalizeCmd.Run(context.Background(), []string{"normalize"})

	s.Require().Error(err)
	s.Equal("Provide a recipe id, or pass --all, --failed, or --new.", cliutil.UserMessage(err))
}

func (s *NormalizeSuite) TestNormalizeCmd_RejectsAllAndFailedTogether() {
	err := normalizeCmd.Run(context.Background(), []string{"normalize", "--all", "--failed"})

	s.Require().Error(err)
	s.Equal("Pass only one of --all, --failed, or --new.", cliutil.UserMessage(err))
}

func (s *NormalizeSuite) TestSaveRecipe() {
	recipe, err := s.client.Recipes.Get(s.T().Context(), 42)
	s.Require().NoError(err)

	dir := s.T().TempDir()
	store := backup.NewStore(dir, 0)

	err = saveRecipe(s.T().Context(), s.client, store, dir, recipe, []byte(`{"id":42,"name":"Updated Tacos"}`))
	s.Require().NoError(err)

	s.Equal(http.MethodPut, s.lastMethod)
	s.Equal("/api/recipe/42/", s.lastPath)

	entries, err := store.List(42)
	s.Require().NoError(err)
	s.Len(entries, 1, "should back up the pre-patched recipe before updating")
}

func (s *NormalizeSuite) TestSaveRecipe_BackupFailureSkipsUpdate() {
	recipe, err := s.client.Recipes.Get(s.T().Context(), 42)
	s.Require().NoError(err)

	dir := s.T().TempDir()
	badDir := filepath.Join(dir, "not-a-directory")
	s.Require().NoError(os.WriteFile(badDir, []byte("x"), 0o644))
	store := backup.NewStore(badDir, 0)

	s.lastMethod = ""
	err = saveRecipe(s.T().Context(), s.client, store, badDir, recipe, []byte(`{}`))

	s.Error(err)
	s.Empty(s.lastMethod, "should not update the recipe when the backup fails")
}
