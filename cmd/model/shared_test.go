package model

import (
	"context"
	"testing"

	"github.com/derethil/mise/internal/ai/providers"
	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/config/section"
	"github.com/stretchr/testify/suite"
	"github.com/urfave/cli/v3"
)

type SharedSuite struct {
	suite.Suite

	cfg config.Config
}

func (s *SharedSuite) SetupTest() {
	s.cfg = config.Config{
		Models: section.ModelsConfig{
			Small: "ollama/qwen2.5:7b",
			Large: "ollama/qwen2.5:14b",
		},
	}
}

func TestSharedSuite(t *testing.T) {
	suite.Run(t, new(SharedSuite))
}

func (s *SharedSuite) selectedModels(args ...string) ([]labeledModel, error) {
	s.T().Helper()
	var models []labeledModel

	cmd := &cli.Command{
		Name: "mise",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: string(cliutil.GlobalFlagModel)},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			var err error
			models, err = selectedModels(cmd, s.cfg)
			return err
		},
	}

	err := cmd.Run(context.Background(), append([]string{"mise"}, args...))
	return models, err
}

func (s *SharedSuite) TestReturnsConfiguredSmallAndLargeModels() {
	models, err := s.selectedModels()

	s.Require().NoError(err)
	s.Equal([]labeledModel{
		{label: "small", ref: providers.ModelRef{Provider: "ollama", Name: "qwen2.5", Tag: "7b"}},
		{label: "large", ref: providers.ModelRef{Provider: "ollama", Name: "qwen2.5", Tag: "14b"}},
	}, models)
}

func (s *SharedSuite) TestModelFlagOverridesConfig() {
	models, err := s.selectedModels("--model", "ollama/llama3:8b")

	s.Require().NoError(err)
	s.Equal([]labeledModel{
		{label: "override", ref: providers.ModelRef{Provider: "ollama", Name: "llama3", Tag: "8b"}},
	}, models)
}

func (s *SharedSuite) TestInvalidConfiguredModel() {
	s.cfg.Models.Small = "not-a-valid-model"

	_, err := s.selectedModels()

	s.Error(err)
}

func (s *SharedSuite) TestModelRefs() {
	refs := modelRefs([]labeledModel{
		{label: "small", ref: providers.ModelRef{Provider: "ollama", Name: "qwen2.5"}},
		{label: "large", ref: providers.ModelRef{Provider: "ollama", Name: "llama3"}},
	})

	s.Equal([]providers.ModelRef{
		{Provider: "ollama", Name: "qwen2.5"},
		{Provider: "ollama", Name: "llama3"},
	}, refs)
}

func (s *SharedSuite) TestOllamaProviderRequiresSelectedOllamaModel() {
	for _, tc := range []struct {
		name   string
		models []labeledModel
	}{
		{"none", nil},
		{"configured", []labeledModel{
			{label: "small", ref: providers.ModelRef{Provider: "openai", Name: "small"}},
			{label: "large", ref: providers.ModelRef{Provider: "anthropic", Name: "large"}},
		}},
		{"override", []labeledModel{
			{label: "override", ref: providers.ModelRef{Provider: "openai", Name: "remote"}},
		}},
	} {
		s.Run(tc.name, func() {
			_, err := ollamaProvider(section.ProviderConfig{}, tc.models)

			s.Require().ErrorIs(err, cliutil.ErrIncorrectUsage)
			s.Contains(cliutil.UserMessage(err), "No Ollama models are selected.")
		})
	}
}

func (s *SharedSuite) TestOllamaProviderAllowsMixedModels() {
	provider, err := ollamaProvider(section.ProviderConfig{BaseURL: "http://localhost:11434"}, []labeledModel{
		{label: "small", ref: providers.ModelRef{Provider: "openai", Name: "remote"}},
		{label: "large", ref: providers.ModelRef{Provider: "ollama", Name: "local"}},
	})

	s.Require().NoError(err)
	s.NotNil(provider)
}
