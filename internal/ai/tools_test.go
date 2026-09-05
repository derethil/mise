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
	client   *Client
	tandoor  *tandoor.Client
}

func (s *ToolsSuite) SetupTest() {
	s.response = map[string]any{"results": []map[string]any{}}

	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(s.response)
	}))

	s.tandoor = tandoor.NewClient(s.server.URL+"/api", "test-token")
	s.client = &Client{g: genkit.Init(s.T().Context())}
}

func (s *ToolsSuite) TearDownTest() {
	s.server.Close()
}

func TestToolsSuite(t *testing.T) {
	suite.Run(t, new(ToolsSuite))
}

func (s *ToolsSuite) TestTandoorGetFoodTool() {
	s.response = map[string]any{
		"results": []map[string]any{
			{"id": 1, "name": "Chicken Thigh", "plural_name": "Chicken Thighs"},
		},
	}

	tool := s.client.TandoorGetFoodTool(s.tandoor)

	s.Equal("getFoods", tool.Name())

	output, err := tool.RunRaw(s.T().Context(), map[string]any{"search": "chicken"})
	s.Require().NoError(err)

	data, err := json.Marshal(output)
	s.Require().NoError(err)
	s.JSONEq(`[{"id": 1, "name": "Chicken Thigh", "plural_name": "Chicken Thighs"}]`, string(data))
}

func (s *ToolsSuite) TestTandoorGetFoodTool_Error() {
	s.server.Close()
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	s.tandoor = tandoor.NewClient(s.server.URL+"/api", "test-token")

	tool := s.client.TandoorGetFoodTool(s.tandoor)

	_, err := tool.RunRaw(s.T().Context(), map[string]any{"search": "chicken"})
	s.Error(err)
}

func (s *ToolsSuite) TestTandoorGetUnitTool() {
	s.response = map[string]any{
		"results": []map[string]any{
			{"id": 1, "name": "cup", "plural_name": "cups"},
		},
	}

	tool := s.client.TandoorGetUnitTool(s.tandoor)

	s.Equal("getUnits", tool.Name())

	output, err := tool.RunRaw(s.T().Context(), map[string]any{"search": "cup"})
	s.Require().NoError(err)

	data, err := json.Marshal(output)
	s.Require().NoError(err)
	s.JSONEq(`[{"id": 1, "name": "cup", "plural_name": "cups"}]`, string(data))
}
