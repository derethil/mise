package ai

import (
	"context"
	"testing"

	"github.com/firebase/genkit/go/ai"
	"github.com/stretchr/testify/suite"
)

type MiddlewareSuite struct {
	suite.Suite
}

func TestMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareSuite))
}

func (s *MiddlewareSuite) wrapModel(text string) string {
	hooks, err := stripThinkArtifacts.New(s.T().Context())
	s.Require().NoError(err)
	s.Require().NotNil(hooks.WrapModel)

	next := func(context.Context, *ai.ModelParams) (*ai.ModelResponse, error) {
		return &ai.ModelResponse{Message: &ai.Message{Content: []*ai.Part{ai.NewTextPart(text)}}}, nil
	}

	resp, err := hooks.WrapModel(s.T().Context(), &ai.ModelParams{}, next)
	s.Require().NoError(err)

	return resp.Message.Content[0].Text
}

func (s *MiddlewareSuite) TestStripsThinkArtifacts() {
	s.Equal(`{"food": "garlic cloves"}`, s.wrapModel(`/no_think {"food": "garlic cloves"}`))
	s.Equal(`{"kind": "junk"}`, s.wrapModel(`(/think) {"kind": "junk"}`))
	s.Equal(`{"unit": "cup"}`, s.wrapModel(`{"unit": "cup"} /NO_THINK`))
}

func (s *MiddlewareSuite) TestLeavesCleanTextAlone() {
	s.Equal(`{"food": "chicken thighs"}`, s.wrapModel(`{"food": "chicken thighs"}`))
}

func (s *MiddlewareSuite) TestOllamaMiddlewareAlwaysStripsThinkArtifacts() {
	s.Len(ollamaMiddleware(GenerateConfig{}), 1)
	s.Len(ollamaMiddleware(GenerateConfig{Reasoning: new(true)}), 1)
	s.Len(ollamaMiddleware(GenerateConfig{Reasoning: new(false)}), 1)
}
