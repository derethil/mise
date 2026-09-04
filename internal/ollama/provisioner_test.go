package ollama

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ProvisionerSuite struct {
	suite.Suite

	server      *httptest.Server
	provisioner *Provisioner

	models          []map[string]any
	pullStatuses    []string
	lastPullModel   string
	lastDeleteModel string
}

func (s *ProvisionerSuite) SetupTest() {
	s.models = []map[string]any{
		{
			"name":        "qwen2.5:7b",
			"modified_at": time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			"size":        int64(123),
			"details": map[string]any{
				"family":             "qwen2",
				"parameter_size":     "7B",
				"quantization_level": "Q4_0",
			},
		},
	}
	s.pullStatuses = []string{"pulling manifest", "success"}
	s.lastPullModel = ""
	s.lastDeleteModel = ""

	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/tags":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"models": s.models})

		case r.Method == http.MethodPost && r.URL.Path == "/api/pull":
			var req struct {
				Model string `json:"model"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			s.lastPullModel = req.Model

			w.Header().Set("Content-Type", "application/x-ndjson")
			enc := json.NewEncoder(w)
			for _, status := range s.pullStatuses {
				_ = enc.Encode(map[string]any{"status": status})
			}

		case r.Method == http.MethodDelete && r.URL.Path == "/api/delete":
			var req struct {
				Model string `json:"model"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			s.lastDeleteModel = req.Model

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	var err error
	s.provisioner, err = NewProvisioner(s.server.URL)
	s.Require().NoError(err)
}

func (s *ProvisionerSuite) TearDownTest() {
	s.server.Close()
}

func TestProvisionerSuite(t *testing.T) {
	suite.Run(t, new(ProvisionerSuite))
}

func (s *ProvisionerSuite) TestNewProvisionerRequiresBaseURL() {
	_, err := NewProvisioner("")

	s.Require().Error(err)
	s.ErrorIs(err, config.ErrInvalidConfig)
}

func (s *ProvisionerSuite) TestModels() {
	models, err := s.provisioner.models(s.T().Context())

	s.Require().NoError(err)
	s.Require().Len(models, 1)
	s.Equal("qwen2.5:7b", models[0].Name)
	s.Equal("qwen2", models[0].Family)
	s.Equal("7B", models[0].ParameterSize)
	s.Equal("Q4_0", models[0].QuantizationLevel)
}

func (s *ProvisionerSuite) TestHasModelTrue() {
	has, err := s.provisioner.hasModel(s.T().Context(), ai.ModelRef{Provider: ai.ProviderOllama, Name: "qwen2.5", Tag: "7b"})

	s.Require().NoError(err)
	s.True(has)
}

func (s *ProvisionerSuite) TestHasModelFalse() {
	has, err := s.provisioner.hasModel(s.T().Context(), ai.ModelRef{Provider: ai.ProviderOllama, Name: "llama3", Tag: "8b"})

	s.Require().NoError(err)
	s.False(has)
}

func (s *ProvisionerSuite) TestPullModelReportsProgress() {
	var statuses []string
	err := s.provisioner.pullModel(s.T().Context(), ai.ModelRef{Provider: ai.ProviderOllama, Name: "llama3", Tag: "8b"}, func(p PullProgress) error {
		statuses = append(statuses, p.Status)
		return nil
	})

	s.Require().NoError(err)
	s.Equal("llama3:8b", s.lastPullModel)
	s.Equal(s.pullStatuses, statuses)
}

func (s *ProvisionerSuite) TestStatusesMarksInstalledModel() {
	statuses, err := s.provisioner.Statuses(s.T().Context(), []ai.ModelRef{
		{Provider: ai.ProviderOllama, Name: "qwen2.5", Tag: "7b"},
		{Provider: ai.ProviderOllama, Name: "llama3", Tag: "8b"},
	})

	s.Require().NoError(err)
	s.Require().Len(statuses, 2)

	s.NotNil(statuses[0].Info)
	s.Equal("qwen2.5:7b", statuses[0].Info.Name)

	s.Nil(statuses[1].Info)
}

func (s *ProvisionerSuite) TestStatusesSkipsNonOllamaModels() {
	statuses, err := s.provisioner.Statuses(s.T().Context(), []ai.ModelRef{
		{Provider: "openai", Name: "gpt-4"},
	})

	s.Require().NoError(err)
	s.Require().Len(statuses, 1)
	s.Nil(statuses[0].Info)
}

func (s *ProvisionerSuite) TestEnsureSkipsWhenAlreadyInstalled() {
	called := false
	err := s.provisioner.Ensure(s.T().Context(), ai.ModelRef{Provider: ai.ProviderOllama, Name: "qwen2.5", Tag: "7b"},
		func(string) (bool, error) { called = true; return true, nil },
		func(PullProgress) error { return nil },
	)

	s.Require().NoError(err)
	s.False(called, "should not ask for confirmation when the model is already installed")
}

func (s *ProvisionerSuite) TestEnsurePullsWhenConfirmed() {
	err := s.provisioner.Ensure(s.T().Context(), ai.ModelRef{Provider: ai.ProviderOllama, Name: "llama3", Tag: "8b"},
		func(string) (bool, error) { return true, nil },
		func(PullProgress) error { return nil },
	)

	s.Require().NoError(err)
	s.Equal("llama3:8b", s.lastPullModel)
}

func (s *ProvisionerSuite) TestEnsureReturnsErrorWhenDeclined() {
	err := s.provisioner.Ensure(s.T().Context(), ai.ModelRef{Provider: ai.ProviderOllama, Name: "llama3", Tag: "8b"},
		func(string) (bool, error) { return false, nil },
		func(PullProgress) error { return nil },
	)

	s.Require().Error(err)
	s.ErrorIs(err, ErrPullDeclined)
	s.Empty(s.lastPullModel, "should not send a pull request")
}

func (s *ProvisionerSuite) TestDeleteModel() {
	err := s.provisioner.deleteModel(s.T().Context(), "qwen2.5:7b")

	s.Require().NoError(err)
	s.Equal("qwen2.5:7b", s.lastDeleteModel)
}

func (s *ProvisionerSuite) TestClearSkipsWhenNothingStale() {
	called := false
	deleted, err := s.provisioner.Clear(s.T().Context(),
		[]ai.ModelRef{{Provider: ai.ProviderOllama, Name: "qwen2.5", Tag: "7b"}},
		func(string) (bool, error) { called = true; return true, nil },
	)

	s.Require().NoError(err)
	s.Empty(deleted)
	s.False(called, "should not ask for confirmation when nothing is stale")
	s.Empty(s.lastDeleteModel)
}

func (s *ProvisionerSuite) TestClearDeletesStaleModelsWhenConfirmed() {
	deleted, err := s.provisioner.Clear(s.T().Context(), nil,
		func(string) (bool, error) { return true, nil },
	)

	s.Require().NoError(err)
	s.Require().Len(deleted, 1)
	s.Equal("qwen2.5:7b", deleted[0].Name)
	s.Equal("qwen2.5:7b", s.lastDeleteModel)
}

func (s *ProvisionerSuite) TestClearReturnsErrorWhenDeclined() {
	deleted, err := s.provisioner.Clear(s.T().Context(), nil,
		func(string) (bool, error) { return false, nil },
	)

	s.Require().Error(err)
	s.ErrorIs(err, ErrClearDeclined)
	s.Empty(deleted)
	s.Empty(s.lastDeleteModel, "should not send a delete request")
}

func TestModelName(t *testing.T) {
	t.Run("ollama model without tag defaults to latest", func(t *testing.T) {
		name := ModelName(ai.ModelRef{Provider: ai.ProviderOllama, Name: "qwen2.5"})
		assert.Equal(t, "qwen2.5:latest", name)
	})

	t.Run("non-ollama model falls back to String", func(t *testing.T) {
		ref := ai.ModelRef{Provider: "openai", Name: "gpt-4"}
		assert.Equal(t, ref.String(), ModelName(ref))
	})
}
