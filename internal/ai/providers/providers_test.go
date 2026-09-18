package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/derethil/mise/internal/config"
	"github.com/firebase/genkit/go/plugins/ollama"
	"github.com/stretchr/testify/suite"
)

type ProvidersSuite struct {
	suite.Suite

	providers config.ProvidersConfig
}

func (s *ProvidersSuite) SetupTest() {
	s.providers = config.ProvidersConfig{Ollama: config.ProviderConfig{BaseURL: "http://localhost:11434"}}
}

func TestProvidersSuite(t *testing.T) {
	suite.Run(t, new(ProvidersSuite))
}

func (s *ProvidersSuite) TestBuildsAPluginPerDistinctProvider() {
	plugins, err := ForModels(context.Background(), s.providers,
		ModelRef{Provider: ProviderOllama, Name: "qwen2.5"},
		ModelRef{Provider: ProviderOllama, Name: "llama3"},
	)

	s.Require().NoError(err)
	s.Len(plugins, 1)
}

func (s *ProvidersSuite) TestUnsupportedProvider() {
	_, err := ForModels(context.Background(), s.providers, ModelRef{Provider: "openai", Name: "gpt-4"})

	s.ErrorIs(err, config.ErrInvalidConfig)
}

func (s *ProvidersSuite) TestProviderNotConfigured() {
	_, err := ForModels(context.Background(), config.ProvidersConfig{}, ModelRef{Provider: ProviderOllama, Name: "qwen2.5"})

	s.ErrorIs(err, config.ErrInvalidConfig)
}

func (s *ProvidersSuite) TestFactoryErrorPropagates() {
	s.providers.Ollama.BaseURL = "://invalid"

	_, err := ForModels(context.Background(), s.providers, ModelRef{Provider: ProviderOllama, Name: "qwen2.5"})

	s.Error(err)
}

func (s *ProvidersSuite) TestConstructsConfiguredProviderWithoutRequests() {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests.Add(1)
	}))
	defer server.Close()

	cfg := config.ProviderConfig{BaseURL: server.URL, Timeout: 42}
	provider, err := NewProvider(ProviderOllama, cfg)

	s.Require().NoError(err)
	s.Equal(&ollama.Ollama{ServerAddress: server.URL, Timeout: 42}, provider.Plugin)
	s.Zero(requests.Load())
}

func (s *ProvidersSuite) TestRejectsUnknownProvider() {
	_, err := NewProvider("unknown", config.ProviderConfig{})

	s.ErrorIs(err, config.ErrInvalidConfig)
}

func (s *ProvidersSuite) TestRejectsInvalidURL() {
	_, err := NewProvider(ProviderOllama, config.ProviderConfig{BaseURL: "://invalid"})

	s.Error(err)
}

func (s *ProvidersSuite) TestSamplingConfigTranslation() {
	provider, err := NewProvider(ProviderOllama, s.providers.Ollama)
	s.Require().NoError(err)

	translated := provider.GenerateConfig(GenerateConfig{Temperature: new(0.2), TopP: new(0.8)})

	s.Equal(&ollama.GenerateContentConfig{Temperature: new(0.2), TopP: new(0.8)}, translated)
}

func (s *ProvidersSuite) TestReasoningConfigTranslation() {
	provider, err := NewProvider(ProviderOllama, s.providers.Ollama)
	s.Require().NoError(err)

	for _, tc := range []struct {
		name      string
		reasoning *bool
		want      *ollama.GenerateContentConfig
	}{
		{"unset", nil, &ollama.GenerateContentConfig{}},
		{"disabled", new(false), &ollama.GenerateContentConfig{Think: ollama.ThinkEnabled(false)}},
		{"enabled", new(true), &ollama.GenerateContentConfig{Think: ollama.ThinkEnabled(true)}},
	} {
		s.Run(tc.name, func() {
			translated := provider.GenerateConfig(GenerateConfig{Reasoning: tc.reasoning})

			s.Equal(tc.want, translated)
		})
	}
}
