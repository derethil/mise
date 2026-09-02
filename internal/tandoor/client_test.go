package tandoor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ClientSuite struct {
	suite.Suite

	server    *httptest.Server
	client    *Client
	backupDir string

	lastMethod string
	lastPath   string
	lastAuth   string
	status     int
	response   map[string]any
}

func (s *ClientSuite) SetupTest() {
	s.status = http.StatusOK
	s.response = map[string]any{"id": 42, "name": "Tacos"}

	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.lastMethod = r.Method
		s.lastPath = r.URL.Path
		s.lastAuth = r.Header.Get("Authorization")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(s.status)
		_ = json.NewEncoder(w).Encode(s.response)
	}))

	s.backupDir = s.T().TempDir()
	s.client = NewClient(s.server.URL, "test-token", s.backupDir)
}

func (s *ClientSuite) TearDownTest() {
	s.server.Close()
}

func (s *ClientSuite) TestGetRecipeByID() {
	recipe, err := s.client.GetRecipeByID(42)

	s.Require().NoError(err)
	s.Equal(42, recipe.ID)
	s.Equal(http.MethodGet, s.lastMethod)
	s.Equal("/api/recipe/42/", s.lastPath)
	s.Equal("Bearer test-token", s.lastAuth)
}

func (s *ClientSuite) TestGetRecipeByID_HTTPError() {
	s.status = http.StatusForbidden

	_, err := s.client.GetRecipeByID(42)

	s.Error(err)
}

func (s *ClientSuite) TestBackupRecipe() {
	recipe, err := s.client.GetRecipeByID(42)
	s.Require().NoError(err)

	err = s.client.BackupRecipe(recipe)
	s.Require().NoError(err)

	data, err := os.ReadFile(filepath.Join(s.backupDir, "42.json"))
	s.Require().NoError(err)
	s.JSONEq(`{"id": 42, "name": "Tacos"}`, string(data))
}

func (s *ClientSuite) TestRestoreRecipe() {
	err := os.WriteFile(filepath.Join(s.backupDir, "42.json"), []byte(`{"id": 42, "name": "Restored Tacos"}`), 0o644)
	s.Require().NoError(err)

	err = s.client.RestoreRecipe(42)

	s.Require().NoError(err)
	s.Equal(http.MethodPut, s.lastMethod)
	s.Equal("/api/recipe/42/", s.lastPath)
}

func (s *ClientSuite) TestRestoreRecipe_MissingBackupFile() {
	err := s.client.RestoreRecipe(999)

	s.Error(err)
}

func (s *ClientSuite) TestRestoreRecipe_InvalidBackupFile() {
	err := os.WriteFile(filepath.Join(s.backupDir, "999.json"), []byte(`not json`), 0o644)
	s.Require().NoError(err)

	err = s.client.RestoreRecipe(999)

	s.Error(err)
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}
