package model

import (
	"context"
	"testing"

	"github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/cliutil"
	"github.com/derethil/mise/internal/config"
	"github.com/stretchr/testify/suite"
	"github.com/urfave/cli/v3"
)

type SharedSuite struct {
	suite.Suite

	cfg config.Config
}

func (s *SharedSuite) SetupTest() {
	s.cfg = config.Config{
		Models: config.ModelsConfig{
			Small: "ollama/qwen2.5:7b",
			Large: "ollama/qwen2.5:14b",
		},
	}
}

func TestSharedSuite(t *testing.T) {
	suite.Run(t, new(SharedSuite))
}

func (s *SharedSuite) selectedModels(args ...string) []labeledModel {
	var (
		models []labeledModel
		err    error
	)

	cmd := &cli.Command{
		Name: "mise",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: string(cliutil.GlobalFlagModel)},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			models, err = selectedModels(cmd, s.cfg)
			return err
		},
	}

	runErr := cmd.Run(context.Background(), append([]string{"mise"}, args...))
	s.Require().NoError(runErr)
	s.Require().NoError(err)

	return models
}

func (s *SharedSuite) TestReturnsConfiguredSmallAndLargeModels() {
	models := s.selectedModels()

	s.Require().Len(models, 2)
	s.Equal(labeledModel{label: "small", ref: ai.ModelRef{Provider: "ollama", Name: "qwen2.5", Tag: "7b"}}, models[0])
	s.Equal(labeledModel{label: "large", ref: ai.ModelRef{Provider: "ollama", Name: "qwen2.5", Tag: "14b"}}, models[1])
}

func (s *SharedSuite) TestModelFlagOverridesConfig() {
	models := s.selectedModels("--model", "ollama/llama3:8b")

	s.Require().Len(models, 1)
	s.Equal(labeledModel{label: "override", ref: ai.ModelRef{Provider: "ollama", Name: "llama3", Tag: "8b"}}, models[0])
}

func (s *SharedSuite) TestInvalidConfiguredModel() {
	s.cfg.Models.Small = "not-a-valid-model"

	var err error
	cmd := &cli.Command{
		Name: "mise",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: string(cliutil.GlobalFlagModel)},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			_, err = selectedModels(cmd, s.cfg)
			return nil
		},
	}
	s.Require().NoError(cmd.Run(context.Background(), []string{"mise"}))

	s.Error(err)
}

func (s *SharedSuite) TestModelRefs() {
	refs := modelRefs([]labeledModel{
		{label: "small", ref: ai.ModelRef{Provider: "ollama", Name: "qwen2.5"}},
		{label: "large", ref: ai.ModelRef{Provider: "ollama", Name: "llama3"}},
	})

	s.Equal([]ai.ModelRef{
		{Provider: "ollama", Name: "qwen2.5"},
		{Provider: "ollama", Name: "llama3"},
	}, refs)
}
