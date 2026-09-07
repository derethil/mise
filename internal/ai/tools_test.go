package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/derethil/mise/internal/tandoor"
	"github.com/firebase/genkit/go/genkit"
	"github.com/stretchr/testify/suite"
)

type ToolsSuite struct {
	suite.Suite

	server   *httptest.Server
	response map[string]any
	registry Registry
}

func (s *ToolsSuite) SetupTest() {
	s.response = map[string]any{"results": []map[string]any{}}

	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(s.response)
	}))

	s.registry = Registry{
		Genkit: genkit.Init(s.T().Context()),
		Deps:   Deps{Tandoor: tandoor.NewClient(s.server.URL+"/api", "test-token")},
	}
}

func (s *ToolsSuite) TearDownTest() {
	s.server.Close()
}

func TestToolsSuite(t *testing.T) {
	suite.Run(t, new(ToolsSuite))
}

func (s *ToolsSuite) TestTandoorSearchFoodTool() {
	s.response = map[string]any{
		"results": []map[string]any{
			{"id": 1, "name": "Chicken Thigh", "plural_name": "Chicken Thighs"},
		},
	}

	tool := SearchFoodsTool(s.registry)

	s.Equal("searchFoods", tool.Name())

	output, err := tool.RunRaw(s.T().Context(), map[string]any{"search": "chicken"})
	s.Require().NoError(err)

	data, err := json.Marshal(output)
	s.Require().NoError(err)
	s.JSONEq(`[{"id": 1, "name": "Chicken Thigh", "plural_name": "Chicken Thighs"}]`, string(data))
}

func (s *ToolsSuite) TestTandoorSearchFoodTool_Error() {
	s.server.Close()
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	s.registry.Tandoor = tandoor.NewClient(s.server.URL+"/api", "test-token")

	tool := SearchFoodsTool(s.registry)

	_, err := tool.RunRaw(s.T().Context(), map[string]any{"search": "chicken"})
	s.Error(err)
}

func (s *ToolsSuite) TestTandoorSearchUnitTool() {
	s.response = map[string]any{
		"results": []map[string]any{
			{"id": 1, "name": "cup", "plural_name": "cups"},
		},
	}

	tool := SearchUnitsTool(s.registry)

	s.Equal("searchUnits", tool.Name())

	output, err := tool.RunRaw(s.T().Context(), map[string]any{"search": "cup"})
	s.Require().NoError(err)

	data, err := json.Marshal(output)
	s.Require().NoError(err)
	s.JSONEq(`[{"id": 1, "name": "cup", "plural_name": "cups"}]`, string(data))
}
