package ai

import (
	"testing"

	"github.com/derethil/mise/internal/config"
	"github.com/stretchr/testify/suite"
)

type ProvidersSuite struct {
	suite.Suite

	providers config.ProvidersConfig
}

func (s *ProvidersSuite) SetupTest() {
	s.providers = config.ProvidersConfig{
		Ollama: config.ProviderConfig{BaseURL: "http://localhost:11434"},
	}
}

func TestProvidersSuite(t *testing.T) {
	suite.Run(t, new(ProvidersSuite))
}

func (s *ProvidersSuite) TestBuildsAPluginPerDistinctProvider() {
	plugins, err := getProviderPlugins(s.providers,
		ModelRef{Provider: ProviderOllama, Name: "qwen2.5"},
		ModelRef{Provider: ProviderOllama, Name: "llama3"},
	)

	s.Require().NoError(err)
	s.Len(plugins, 1)
}

func (s *ProvidersSuite) TestUnsupportedProvider() {
	_, err := getProviderPlugins(s.providers, ModelRef{Provider: "openai", Name: "gpt-4"})

	s.Require().Error(err)
	s.ErrorIs(err, config.ErrInvalidConfig)
}

func (s *ProvidersSuite) TestProviderNotConfigured() {
	_, err := getProviderPlugins(config.ProvidersConfig{}, ModelRef{Provider: ProviderOllama, Name: "qwen2.5"})

	s.Require().Error(err)
	s.ErrorIs(err, config.ErrInvalidConfig)
}

func (s *ProvidersSuite) TestFactoryErrorPropagates() {
	_, err := getProviderPlugins(config.ProvidersConfig{Ollama: config.ProviderConfig{}}, ModelRef{Provider: ProviderOllama, Name: "qwen2.5"})

	s.Error(err)
}
