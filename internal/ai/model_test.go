package ai

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ModelSuite struct {
	suite.Suite
}

func TestModelSuite(t *testing.T) {
	suite.Run(t, new(ModelSuite))
}

func (s *ModelSuite) TestStringWithoutTag() {
	ref := ModelRef{Provider: "ollama", Name: "qwen2.5"}

	s.Equal("ollama/qwen2.5", ref.String())
}

func (s *ModelSuite) TestStringWithTag() {
	ref := ModelRef{Provider: "ollama", Name: "qwen2.5", Tag: "7b"}

	s.Equal("ollama/qwen2.5:7b", ref.String())
}

func (s *ModelSuite) TestParseModel() {
	ref, err := ParseModel("ollama/qwen2.5:7b")

	s.Require().NoError(err)
	s.Equal(ModelRef{Provider: "ollama", Name: "qwen2.5", Tag: "7b"}, ref)
}

func (s *ModelSuite) TestParseModelWithoutTag() {
	ref, err := ParseModel("ollama/qwen2.5")

	s.Require().NoError(err)
	s.Equal(ModelRef{Provider: "ollama", Name: "qwen2.5"}, ref)
}

func (s *ModelSuite) TestParseModelMissingProviderSeparator() {
	_, err := ParseModel("qwen2.5")

	s.Error(err)
}

func (s *ModelSuite) TestParseModelUnsupportedProvider() {
	_, err := ParseModel("openai/gpt-4")

	s.Error(err)
}
