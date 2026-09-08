package cliutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/urfave/cli/v3"
)

type FlagsSuite struct {
	suite.Suite
}

func TestFlagsSuite(t *testing.T) {
	suite.Run(t, new(FlagsSuite))
}

func (s *FlagsSuite) resolveFlag(fallback string, args ...string) string {
	var result string

	cmd := &cli.Command{
		Name: "mise",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: string(GlobalFlagModel)},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			result = ResolveFlag(cmd, GlobalFlagModel, fallback)
			return nil
		},
	}

	s.Require().NoError(cmd.Run(context.Background(), append([]string{"mise"}, args...)))

	return result
}

func (s *FlagsSuite) TestResolveFlagReturnsFallbackWhenUnset() {
	s.Equal("ollama/qwen2.5:7b", s.resolveFlag("ollama/qwen2.5:7b"))
}

func (s *FlagsSuite) TestResolveFlagReturnsFlagValueWhenSet() {
	s.Equal("ollama/llama3:8b", s.resolveFlag("ollama/qwen2.5:7b", "--model", "ollama/llama3:8b"))
}
