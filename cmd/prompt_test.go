package cmd

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type PromptSuite struct {
	suite.Suite
}

func TestPromptSuite(t *testing.T) {
	suite.Run(t, new(PromptSuite))
}

func (s *PromptSuite) TestConfirmErrorsWhenNotInteractive() {
	// go test's stdin is never a terminal, so confirm should refuse to prompt.
	ok, err := confirm("proceed?")

	s.False(ok)
	s.ErrorIs(err, errNotInteractive)
}

func (s *PromptSuite) TestAutoConfirmAlwaysAccepts() {
	ok, err := autoConfirm("proceed?")

	s.Require().NoError(err)
	s.True(ok)
}

func (s *PromptSuite) TestPrintProgressWithTotal() {
	onProgress := printProgress()

	err := onProgress(progress{Label: "qwen2.5", Status: "pulling", Total: 200, Completed: 100})

	s.Require().NoError(err)
}

func (s *PromptSuite) TestPrintProgressWithoutTotal() {
	onProgress := printProgress()

	err := onProgress(progress{Label: "qwen2.5", Status: "verifying"})

	s.Require().NoError(err)
}

func (s *PromptSuite) TestPrintProgressTracksStatusChanges() {
	onProgress := printProgress()

	s.Require().NoError(onProgress(progress{Label: "qwen2.5", Status: "pulling", Total: 10, Completed: 1}))
	s.Require().NoError(onProgress(progress{Label: "qwen2.5", Status: "pulling", Total: 10, Completed: 5}))
	s.Require().NoError(onProgress(progress{Label: "qwen2.5", Status: "verifying"}))
}
