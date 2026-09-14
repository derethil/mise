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
	s.Equal("Provide a recipe id, or pass --all or --failed.", cliutil.UserMessage(err))
}

func (s *KeywordSuite) TestKeywordCmd_RejectsAllAndFailedTogether() {
	err := keywordCmd.Run(context.Background(), []string{"keyword", "--all", "--failed"})

	s.Require().Error(err)
	s.Equal("Pass either --all or --failed, not both.", cliutil.UserMessage(err))
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

func (s *KeywordSuite) recipeWithKeywords(names ...string) *tandoor.Recipe {
	keywords := make([]map[string]any, len(names))
	for i, name := range names {
		keywords[i] = map[string]any{"id": i + 1, "name": name}
	}
	s.response = map[string]any{"id": 42, "name": "Tacos", "keywords": keywords}

	recipe, err := s.client.Recipes.Get(s.T().Context(), 42)
	s.Require().NoError(err)

	return recipe
}

func (s *KeywordSuite) TestIsAlreadyTagged_NoKeywords() {
	recipe := s.recipeWithKeywords()

	s.False(isAlreadyTagged(recipe, nil))
}

func (s *KeywordSuite) TestIsAlreadyTagged_OnlyIgnoredKeywordsPresent() {
	recipe := s.recipeWithKeywords("Uncategorized")

	s.False(isAlreadyTagged(recipe, []string{"uncategorized"}), "an ignored keyword doesn't count as tagged")
}

func (s *KeywordSuite) TestIsAlreadyTagged_HasNonIgnoredKeyword() {
	recipe := s.recipeWithKeywords("Uncategorized", "Dinner")

	s.True(isAlreadyTagged(recipe, []string{"uncategorized"}))
}

func (s *KeywordSuite) TestIsAlreadyTagged_IgnoreMatchIsCaseAndWhitespaceInsensitive() {
	recipe := s.recipeWithKeywords("  Uncategorized  ")

	s.False(isAlreadyTagged(recipe, []string{"UNCATEGORIZED"}))
}
