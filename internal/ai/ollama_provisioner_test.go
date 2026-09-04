package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/derethil/mise/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type OllamaProvisionerSuite struct {
	suite.Suite

	server      *httptest.Server
	provisioner *OllamaProvisioner

	models        []map[string]any
	pullStatuses  []string
	lastPullModel string
}

func (s *OllamaProvisionerSuite) SetupTest() {
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

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	var err error
	s.provisioner, err = NewOllamaProvisioner(s.server.URL)
	s.Require().NoError(err)
}

func (s *OllamaProvisionerSuite) TearDownTest() {
	s.server.Close()
}

func TestOllamaProvisionerSuite(t *testing.T) {
	suite.Run(t, new(OllamaProvisionerSuite))
}

func (s *OllamaProvisionerSuite) TestNewOllamaProvisionerRequiresBaseURL() {
	_, err := NewOllamaProvisioner("")

	s.Require().Error(err)
	s.ErrorIs(err, config.ErrInvalidConfig)
}

func (s *OllamaProvisionerSuite) TestModels() {
	models, err := s.provisioner.Models(s.T().Context())

	s.Require().NoError(err)
	s.Require().Len(models, 1)
	s.Equal("qwen2.5:7b", models[0].Name)
	s.Equal("qwen2", models[0].Family)
	s.Equal("7B", models[0].ParameterSize)
	s.Equal("Q4_0", models[0].QuantizationLevel)
}

func (s *OllamaProvisionerSuite) TestHasModelTrue() {
	has, err := s.provisioner.HasModel(s.T().Context(), ModelRef{Provider: ProviderOllama, Name: "qwen2.5", Tag: "7b"})

	s.Require().NoError(err)
	s.True(has)
}

func (s *OllamaProvisionerSuite) TestHasModelFalse() {
	has, err := s.provisioner.HasModel(s.T().Context(), ModelRef{Provider: ProviderOllama, Name: "llama3", Tag: "8b"})

	s.Require().NoError(err)
	s.False(has)
}

func (s *OllamaProvisionerSuite) TestPullModelReportsProgress() {
	var statuses []string
	err := s.provisioner.PullModel(s.T().Context(), ModelRef{Provider: ProviderOllama, Name: "llama3", Tag: "8b"}, func(p PullProgress) error {
		statuses = append(statuses, p.Status)
		return nil
	})

	s.Require().NoError(err)
	s.Equal("llama3:8b", s.lastPullModel)
	s.Equal(s.pullStatuses, statuses)
}

func (s *OllamaProvisionerSuite) TestStatusesMarksInstalledModel() {
	statuses, err := s.provisioner.Statuses(s.T().Context(), []ModelRef{
		{Provider: ProviderOllama, Name: "qwen2.5", Tag: "7b"},
		{Provider: ProviderOllama, Name: "llama3", Tag: "8b"},
	})

	s.Require().NoError(err)
	s.Require().Len(statuses, 2)

	s.NotNil(statuses[0].Info)
	s.Equal("qwen2.5:7b", statuses[0].Info.Name)

	s.Nil(statuses[1].Info)
}

func (s *OllamaProvisionerSuite) TestStatusesSkipsNonOllamaModels() {
	statuses, err := s.provisioner.Statuses(s.T().Context(), []ModelRef{
		{Provider: "openai", Name: "gpt-4"},
	})

	s.Require().NoError(err)
	s.Require().Len(statuses, 1)
	s.Nil(statuses[0].Info)
}

func (s *OllamaProvisionerSuite) TestEnsureSkipsWhenAlreadyInstalled() {
	called := false
	err := s.provisioner.Ensure(s.T().Context(), ModelRef{Provider: ProviderOllama, Name: "qwen2.5", Tag: "7b"},
		func(string) (bool, error) { called = true; return true, nil },
		func(PullProgress) error { return nil },
	)

	s.Require().NoError(err)
	s.False(called, "should not ask for confirmation when the model is already installed")
}

func (s *OllamaProvisionerSuite) TestEnsurePullsWhenConfirmed() {
	err := s.provisioner.Ensure(s.T().Context(), ModelRef{Provider: ProviderOllama, Name: "llama3", Tag: "8b"},
		func(string) (bool, error) { return true, nil },
		func(PullProgress) error { return nil },
	)

	s.Require().NoError(err)
	s.Equal("llama3:8b", s.lastPullModel)
}

func (s *OllamaProvisionerSuite) TestEnsureReturnsErrorWhenDeclined() {
	err := s.provisioner.Ensure(s.T().Context(), ModelRef{Provider: ProviderOllama, Name: "llama3", Tag: "8b"},
		func(string) (bool, error) { return false, nil },
		func(PullProgress) error { return nil },
	)

	s.Require().Error(err)
	s.ErrorIs(err, ErrPullDeclined)
	s.Empty(s.lastPullModel, "should not send a pull request")
}

func TestOllamaModelName(t *testing.T) {
	t.Run("ollama model without tag defaults to latest", func(t *testing.T) {
		name := OllamaModelName(ModelRef{Provider: ProviderOllama, Name: "qwen2.5"})
		assert.Equal(t, "qwen2.5:latest", name)
	})

	t.Run("non-ollama model falls back to String", func(t *testing.T) {
		ref := ModelRef{Provider: "openai", Name: "gpt-4"}
		assert.Equal(t, ref.String(), OllamaModelName(ref))
	})
}
