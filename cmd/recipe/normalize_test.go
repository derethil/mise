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
	s.Equal("Provide a recipe id, or pass --all or --failed.", cliutil.UserMessage(err))
}

func (s *NormalizeSuite) TestNormalizeCmd_RejectsAllAndFailedTogether() {
	err := normalizeCmd.Run(context.Background(), []string{"normalize", "--all", "--failed"})

	s.Require().Error(err)
	s.Equal("Pass either --all or --failed, not both.", cliutil.UserMessage(err))
}

func (s *NormalizeSuite) TestFailedIDs_RoundTrip() {
	path := filepath.Join(s.T().TempDir(), "failed.json")

	ids, err := loadFailedIDs(path)
	s.Require().NoError(err, "missing file should not be an error")
	s.Empty(ids)

	s.Require().NoError(saveFailedIDs(path, []int{3, 7, 9}))

	ids, err = loadFailedIDs(path)
	s.Require().NoError(err)
	s.Equal([]int{3, 7, 9}, ids)
}

func (s *NormalizeSuite) TestFailedIDs_SaveEmptyClearsFile() {
	path := filepath.Join(s.T().TempDir(), "failed.json")

	s.Require().NoError(saveFailedIDs(path, []int{3, 7}))
	s.Require().NoError(saveFailedIDs(path, nil))

	ids, err := loadFailedIDs(path)
	s.Require().NoError(err)
	s.Empty(ids)
}

func (s *NormalizeSuite) TestAlreadyNormalized_NoBackups() {
	store := backup.NewStore(s.T().TempDir(), 0)

	skip, err := alreadyNormalized(store, 42)

	s.Require().NoError(err)
	s.False(skip)
}

func (s *NormalizeSuite) TestAlreadyNormalized_HasBackup() {
	store := backup.NewStore(s.T().TempDir(), 0)
	_, err := store.Save(42, []byte(`{"id":42}`))
	s.Require().NoError(err)

	skip, err := alreadyNormalized(store, 42)

	s.Require().NoError(err)
	s.True(skip)
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
