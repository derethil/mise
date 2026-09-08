package cmd

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type RootSuite struct {
	suite.Suite
}

func TestRootSuite(t *testing.T) {
	suite.Run(t, new(RootSuite))
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
