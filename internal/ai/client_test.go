package ai

import (
	"context"
	"testing"

	"github.com/derethil/mise/internal/ai/providers"
	"github.com/derethil/mise/internal/config"
	"github.com/derethil/mise/internal/config/section"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	"github.com/stretchr/testify/suite"
)

type fakeToolPlugin struct {
	name    string
	actions []api.Action
}

func (p *fakeToolPlugin) Name() string { return p.name }

func (p *fakeToolPlugin) Init(context.Context) []api.Action {
	return p.actions
}

func stubModelFunc(_ context.Context, _ *ai.ModelRequest, _ any, _ ai.ModelStreamCallback) (*ai.ModelResponse, error) {
	return &ai.ModelResponse{}, nil
}

func newFakeModel(name string, supports *ai.ModelSupports) api.Action {
	return ai.NewModelAction(name, &ai.ModelOptions{Supports: supports}, stubModelFunc)
}

func testProvider(name string, actions ...api.Action) *providers.Provider {
	return &providers.Provider{
		Name:   name,
		Plugin: &fakeToolPlugin{name: name, actions: actions},
	}
}

type ClientSuite struct {
	suite.Suite
}

func (s *ClientSuite) newClient(provider *providers.Provider, names ...string) (*Client, error) {
	models := make([]providers.ModelRef, len(names))
	for i, name := range names {
		models[i] = providers.ModelRef{Provider: provider.Name, Name: name}
	}

	return newGenkitClient(s.T().Context(), []*providers.Provider{provider}, Deps{}, models...)
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}

func (s *ClientSuite) TestNewGenkitClientSucceedsWhenModelSupportsTools() {
	provider := testProvider("remote", newFakeModel("remote/model", &ai.ModelSupports{Tools: true}))

	_, err := s.newClient(provider, "model")

	s.NoError(err)
}

func (s *ClientSuite) TestNewGenkitClientErrorsWhenModelDoesNotSupportTools() {
	provider := testProvider("remote", newFakeModel("remote/model", &ai.ModelSupports{}))

	_, err := s.newClient(provider, "model")

	s.ErrorIs(err, ErrModelMissingTools)
}

func (s *ClientSuite) TestNewGenkitClientChecksExtraModelsToo() {
	provider := testProvider("remote",
		newFakeModel("remote/with-tools", &ai.ModelSupports{Tools: true}),
		newFakeModel("remote/no-tools", &ai.ModelSupports{}),
	)

	_, err := s.newClient(provider, "with-tools", "no-tools")

	s.ErrorIs(err, ErrModelMissingTools)
	s.ErrorContains(err, "remote/no-tools")
}

func (s *ClientSuite) TestNewGenkitClientPropagatesProviderConfigErrors() {
	model := providers.ModelRef{Provider: providers.ProviderOllama, Name: "qwen2.5"}

	_, err := NewGenkitClient(s.T().Context(), section.ProvidersConfig{}, Deps{}, model)

	s.ErrorIs(err, config.ErrInvalidConfig)
}

type providerTestFeature struct {
	registry Registry
}

func (f *providerTestFeature) Register(r Registry) error {
	f.registry = r
	return nil
}

func (s *ClientSuite) TestPassesSelectedProviderToFeatures() {
	feature := &providerTestFeature{}
	provider := testProvider("remote", newFakeModel("remote/model", &ai.ModelSupports{Tools: true}))

	original := featureFactories
	featureFactories = []func() Feature{func() Feature { return feature }}
	s.T().Cleanup(func() { featureFactories = original })

	_, err := s.newClient(provider, "model")
	s.Require().NoError(err)

	s.Same(provider, feature.registry.Providers[provider.Name])
	s.Equal(providers.ModelRef{Provider: "remote", Name: "model"}, feature.registry.Model)
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
