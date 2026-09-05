package ai

import (
	"context"
	"testing"

	"github.com/derethil/mise/internal/config"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	"github.com/stretchr/testify/suite"
)

type fakeToolPlugin struct {
	name   string
	action api.Action
}

func (p *fakeToolPlugin) Name() string { return p.name }

func (p *fakeToolPlugin) Init(context.Context) []api.Action {
	return []api.Action{p.action}
}

func stubModelFunc(_ context.Context, _ *ai.ModelRequest, _ any, _ ai.ModelStreamCallback) (*ai.ModelResponse, error) {
	return &ai.ModelResponse{}, nil
}

func newFakeModel(name string, supports *ai.ModelSupports) api.Action {
	return ai.NewModelAction(name, &ai.ModelOptions{Supports: supports}, stubModelFunc)
}

func (s *ClientSuite) withFakeOllamaModel(modelName string, supports *ai.ModelSupports) {
	action := newFakeModel(api.NewName(ProviderOllama, modelName), supports)

	original := providerFactories[ProviderOllama]
	providerFactories[ProviderOllama] = func(config.ProviderConfig) (api.Plugin, error) {
		return &fakeToolPlugin{name: ProviderOllama, action: action}, nil
	}

	s.T().Cleanup(func() {
		providerFactories[ProviderOllama] = original
	})
}

type ClientSuite struct {
	suite.Suite

	providers config.ProvidersConfig
}

func (s *ClientSuite) SetupTest() {
	s.providers = config.ProvidersConfig{
		Ollama: config.ProviderConfig{BaseURL: "http://localhost:11434"},
	}
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}

func (s *ClientSuite) TestNewGenkitClientSucceedsWhenModelSupportsTools() {
	s.withFakeOllamaModel("with-tools", &ai.ModelSupports{Tools: true})

	client, err := NewGenkitClient(s.T().Context(), s.providers, ModelRef{Provider: ProviderOllama, Name: "with-tools"})

	s.Require().NoError(err)
	s.NotNil(client)
}

func (s *ClientSuite) TestNewGenkitClientErrorsWhenModelDoesNotSupportTools() {
	s.withFakeOllamaModel("no-tools", &ai.ModelSupports{Tools: false})

	client, err := NewGenkitClient(s.T().Context(), s.providers, ModelRef{Provider: ProviderOllama, Name: "no-tools"})

	s.Require().Error(err)
	s.ErrorIs(err, ErrModelMissingTools)
	s.Nil(client)
}

func (s *ClientSuite) TestNewGenkitClientChecksExtraModelsToo() {
	s.withFakeOllamaModel("no-tools", &ai.ModelSupports{Tools: false})

	_, err := NewGenkitClient(s.T().Context(), s.providers,
		ModelRef{Provider: ProviderOllama, Name: "no-tools"},
		ModelRef{Provider: ProviderOllama, Name: "no-tools"},
	)

	s.Require().Error(err)
	s.ErrorIs(err, ErrModelMissingTools)
}

func (s *ClientSuite) TestNewGenkitClientPropagatesProviderConfigErrors() {
	_, err := NewGenkitClient(s.T().Context(), config.ProvidersConfig{}, ModelRef{Provider: ProviderOllama, Name: "qwen2.5"})

	s.Require().Error(err)
	s.ErrorIs(err, config.ErrInvalidConfig)
}

func (s *ClientSuite) TestSupportsToolsWhenModelSupportsThem() {
	g := genkit.Init(s.T().Context())
	genkit.DefineModelAction(g, "test/with-tools", &ai.ModelOptions{Supports: &ai.ModelSupports{Tools: true}}, stubModelFunc)

	s.True(supportsTools(g, "test/with-tools"))
}

func (s *ClientSuite) TestSupportsToolsWhenModelDoesNotSupportThem() {
	g := genkit.Init(s.T().Context())
	genkit.DefineModelAction(g, "test/no-tools", &ai.ModelOptions{Supports: &ai.ModelSupports{Tools: false}}, stubModelFunc)

	s.False(supportsTools(g, "test/no-tools"))
}

func (s *ClientSuite) TestSupportsToolsWhenModelIsNotRegistered() {
	g := genkit.Init(s.T().Context())

	s.False(supportsTools(g, "test/missing"))
}
