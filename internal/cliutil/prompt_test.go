package cliutil

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
	// go test's stdin is never a terminal, so Confirm should refuse to prompt.
	ok, err := Confirm("proceed?")

	s.False(ok)
	s.ErrorIs(err, ErrNotInteractive)
}

func (s *PromptSuite) TestAutoConfirmAlwaysAccepts() {
	ok, err := AutoConfirm("proceed?")

	s.Require().NoError(err)
	s.True(ok)
}

func (s *PromptSuite) TestPrintProgressWithTotal() {
	onProgress := PrintProgress()

	err := onProgress(Progress{Label: "qwen2.5", Status: "pulling", Total: 200, Completed: 100})

	s.Require().NoError(err)
}

func (s *PromptSuite) TestPrintProgressWithoutTotal() {
	onProgress := PrintProgress()

	err := onProgress(Progress{Label: "qwen2.5", Status: "verifying"})

	s.Require().NoError(err)
}

func (s *PromptSuite) TestPrintProgressTracksStatusChanges() {
	onProgress := PrintProgress()

	s.Require().NoError(onProgress(Progress{Label: "qwen2.5", Status: "pulling", Total: 10, Completed: 1}))
	s.Require().NoError(onProgress(Progress{Label: "qwen2.5", Status: "pulling", Total: 10, Completed: 5}))
	s.Require().NoError(onProgress(Progress{Label: "qwen2.5", Status: "verifying"}))
}
