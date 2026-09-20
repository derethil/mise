package recipe

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/stretchr/testify/suite"
)

type KeywordSuite struct {
	suite.Suite

	server   *httptest.Server
	client   *tandoor.Client
	response map[string]any
}

func (s *KeywordSuite) SetupTest() {
	s.response = map[string]any{"id": 42, "name": "Tacos"}

	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(s.response)
	}))

	s.client = tandoor.NewClient(s.server.URL, "test-token")
}

func (s *KeywordSuite) TearDownTest() {
	s.server.Close()
}

func TestKeywordSuite(t *testing.T) {
	suite.Run(t, new(KeywordSuite))
}

func (s *KeywordSuite) TestKeywordCmd_RequiresIDOrAll() {
	err := keywordCmd.Run(context.Background(), []string{"keyword"})

	s.Require().Error(err)
	s.Equal("Provide a recipe id, or pass --all, --failed, or --new.", cliutil.UserMessage(err))
}

func (s *KeywordSuite) TestKeywordCmd_RejectsAllAndFailedTogether() {
	err := keywordCmd.Run(context.Background(), []string{"keyword", "--all", "--failed"})

	s.Require().Error(err)
	s.Equal("Pass only one of --all, --failed, or --new.", cliutil.UserMessage(err))
}

func (s *KeywordSuite) TestReadSchema_MissingFile() {
	path := filepath.Join(s.T().TempDir(), "missing.md")

	_, err := readSchema(path)

	s.Require().Error(err)
	s.Contains(cliutil.UserMessage(err), path)
}

func (s *KeywordSuite) TestReadSchema_EmptyFileIsAnError() {
	path := filepath.Join(s.T().TempDir(), "schema.md")
	s.Require().NoError(os.WriteFile(path, []byte("   \n"), 0o644))

	_, err := readSchema(path)

	s.Require().Error(err)
	s.Contains(cliutil.UserMessage(err), "empty")
}

func (s *KeywordSuite) TestReadSchema_ReturnsTrimmedContent() {
	path := filepath.Join(s.T().TempDir(), "schema.md")
	s.Require().NoError(os.WriteFile(path, []byte("\n## Cuisine\nAssign it.\n\n"), 0o644))

	schema, err := readSchema(path)

	s.Require().NoError(err)
	s.Equal("## Cuisine\nAssign it.", schema)
}
