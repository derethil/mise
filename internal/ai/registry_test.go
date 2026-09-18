package ai

import (
	"context"
	"errors"
	"testing"

	"github.com/derethil/mise/internal/ai/providers"
	genai "github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/middleware"
	"github.com/stretchr/testify/suite"
)

type RegistrySuite struct {
	suite.Suite

	registry Registry
	prompt   genai.Prompt
	request  *genai.ModelRequest

	modelCalls      int
	middlewareCalls int
	failOnce        bool
}

func TestRegistrySuite(t *testing.T) {
	suite.Run(t, new(RegistrySuite))
}

func (s *RegistrySuite) SetupTest() {
	s.modelCalls, s.middlewareCalls = 0, 0
	s.failOnce = false
	s.request = nil

	g := genkit.Init(s.T().Context(), genkit.WithPlugins(&middleware.Middleware{}))
	model := genkit.DefineModelAction(g, "remote/model", &genai.ModelOptions{}, s.generate)
	s.prompt = genkit.DefinePrompt(g, "test", genai.WithPrompt("Hello"), genai.WithModel(model))

	s.registry = Registry{
		Genkit: g,
		Model:  providers.ModelRef{Provider: "remote", Name: "model"},
		Providers: map[string]*providers.Provider{"remote": {
			Name: "remote",
			GenerateConfig: func(cfg providers.GenerateConfig) any {
				return map[string]any{"remote_reasoning": *cfg.Reasoning}
			},
			Middleware: func(providers.GenerateConfig) []genai.Middleware {
				return []genai.Middleware{genai.MiddlewareFunc(s.hooks)}
			},
		}},
	}
}

func (s *RegistrySuite) generate(_ context.Context, req *genai.ModelRequest, _ any, _ genai.ModelStreamCallback) (*genai.ModelResponse, error) {
	s.request = req
	s.modelCalls++

	if s.failOnce && s.modelCalls == 1 {
		return nil, errors.New("temporary failure")
	}

	message := &genai.Message{
		Role:    genai.RoleModel,
		Content: []*genai.Part{genai.NewTextPart("/no_think remote response")},
	}

	return &genai.ModelResponse{Message: message}, nil
}

func (s *RegistrySuite) hooks(context.Context) (*genai.Hooks, error) {
	return &genai.Hooks{WrapModel: s.wrapModel}, nil
}

func (s *RegistrySuite) wrapModel(ctx context.Context, params *genai.ModelParams, next genai.ModelNext) (*genai.ModelResponse, error) {
	s.middlewareCalls++

	return next(ctx, params)
}

func (s *RegistrySuite) execute(cfg providers.GenerateConfig) string {
	opts := s.registry.PromptOptions(cfg)
	response, err := s.prompt.Execute(s.T().Context(), opts...)
	s.Require().NoError(err)

	return response.Text()
}

func (s *RegistrySuite) TestEmptyConfigSkipsProviderOptions() {
	s.execute(providers.GenerateConfig{})

	s.Nil(s.request.Config)
	s.Zero(s.middlewareCalls)
}

func (s *RegistrySuite) TestTranslatesExplicitFalseReasoning() {
	s.execute(providers.GenerateConfig{Reasoning: new(false)})

	s.Equal(map[string]any{"remote_reasoning": false}, s.request.Config)
}

func (s *RegistrySuite) TestRetriesWithEmptyConfig() {
	s.failOnce = true

	s.execute(providers.GenerateConfig{})

	s.Equal(2, s.modelCalls)
}

func (s *RegistrySuite) TestProviderMiddlewareRunsInsideRetry() {
	s.failOnce = true

	s.execute(providers.GenerateConfig{Reasoning: new(false)})

	s.Equal(2, s.middlewareCalls)
}

func (s *RegistrySuite) TestDoesNotApplyOllamaMiddlewareToOtherProviders() {
	s.registry.Providers["remote"] = &providers.Provider{Name: "remote"}

	text := s.execute(providers.GenerateConfig{Reasoning: new(false)})

	s.Equal("/no_think remote response", text)
}

func (s *RegistrySuite) TestUsesRequestedModelsProvider() {
	called := false
	s.registry.Providers["local"] = &providers.Provider{
		GenerateConfig: func(providers.GenerateConfig) any {
			called = true
			return nil
		},
	}

	model := providers.ModelRef{Provider: "local", Name: "model"}
	s.registry.PromptOptionsFor(model, providers.GenerateConfig{Reasoning: new(false)})

	s.True(called)
}
