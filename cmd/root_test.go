package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/urfave/cli/v3"
)

type RootSuite struct {
	suite.Suite
}

func TestRootSuite(t *testing.T) {
	suite.Run(t, new(RootSuite))
}

func (s *RootSuite) resolveFlag(fallback string, args ...string) string {
	var result string

	cmd := &cli.Command{
		Name:  "mise",
		Flags: globalFlags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			result = resolveFlag(cmd, GlobalFlagModel, fallback)
			return nil
		},
	}

	s.Require().NoError(cmd.Run(context.Background(), append([]string{"mise"}, args...)))

	return result
}

func (s *RootSuite) TestResolveFlagReturnsFallbackWhenUnset() {
	s.Equal("ollama/qwen2.5:7b", s.resolveFlag("ollama/qwen2.5:7b"))
}

func (s *RootSuite) TestResolveFlagReturnsFlagValueWhenSet() {
	s.Equal("ollama/llama3:8b", s.resolveFlag("ollama/qwen2.5:7b", "--model", "ollama/llama3:8b"))
}

func (s *RootSuite) TestCommandPathResolvesNestedCommand() {
	s.Equal("recipe clean", commandPath(rootCmd, []string{"recipe", "clean", "5"}))
}

func (s *RootSuite) TestCommandPathTopLevelOnly() {
	s.Equal("recipe", commandPath(rootCmd, []string{"recipe"}))
}

func (s *RootSuite) TestCommandPathSkipsLeadingFlags() {
	s.Equal("recipe clean", commandPath(rootCmd, []string{"-v", "recipe", "clean", "5"}))
}

func (s *RootSuite) TestCommandPathStopsAtUnknownCommand() {
	s.Equal("", commandPath(rootCmd, []string{"nonexistent"}))
}
